from __future__ import annotations

import json
import os
from pathlib import Path
import re
import shlex
import shutil
import time
from typing import Any


JsonObject = dict[str, Any]

PR_REVIEW_VERDICT_MARKER = "RSI_PR_REVIEW_VERDICT"
_BLOCKING_SUBAGENT_EXIT_REASONS = {"max_iterations", "timeout", "interrupted", "error", "failed"}
_APPROVAL_DELIVERY_TOOL_NAMES = {
    "conversations_add_message",
    "post_message",
    "send_message",
    "slack_send_message",
    "slack_reply",
    "slack.reply",
    "rsi_slack.message_post",
    "rsi_slack.report_post",
    "rsi_slack_report_post",
    "rsi_slack_message_post",
}


def render_pr_review_context_sections(payload: JsonObject) -> list[str]:
    parts: list[str] = []
    if payload.get("pr_review_approval_gate"):
        parts.append(
            "PR review approval gate: GitHub PR approvals and Slack approval reports require a fresh delegate_task "
            f"result from this same session whose summary contains {PR_REVIEW_VERDICT_MARKER} JSON with "
            "approval_safe=true, blocking_findings=0, and verdict=approve. Missing or partial subagent verdicts "
            "are not approvable."
        )
    review_workspace_root = str(payload.get("pr_review_workspace_root", "") or "").strip()
    if review_workspace_root:
        parts.append(
            "PR review isolated workspace root: "
            + review_workspace_root
            + ". Put any temporary PR-review clones or worktrees under this directory; RSI cleans it at session end."
        )
    return parts


def pre_tool_call_guard(
    runtime_root: Path,
    tool_name: str,
    args: JsonObject | None = None,
    *,
    task_id: str = "",
    session_id: str = "",
    **kwargs: Any,
) -> JsonObject | None:
    safe_args = args if isinstance(args, dict) else {}
    command = _extract_terminal_command(tool_name, safe_args)
    is_gh_approval, command_prs = _extract_gh_pr_review_approval_prs(command)
    is_approval_delivery_tool = _is_approval_delivery_tool(tool_name)
    delivery_text = _approval_delivery_text(safe_args) if is_approval_delivery_tool else ""
    is_approval_delivery = (
        is_approval_delivery_tool
        and _text_claims_pr_approval(delivery_text)
        and _looks_like_pr_review_context({}, delivery_text)
    )

    resolved_session_id = _session_id_from_hook(task_id=task_id, session_id=session_id, **kwargs)
    if not resolved_session_id:
        if is_gh_approval:
            return _block_pr_approval(
                runtime_root,
                "",
                "github_cli",
                tool_name,
                command_prs,
                "approval attempted without a session id, so RSI cannot verify the PR-review gate context or fresh-subagent verdict",
            )
        if is_approval_delivery:
            return _block_pr_approval(
                runtime_root,
                "",
                "slack_delivery",
                tool_name,
                _extract_pr_numbers(delivery_text),
                "Slack approval delivery attempted without a session id, so RSI cannot verify the PR-review gate context or fresh-subagent verdict",
            )
        return None

    payload = _load_context(runtime_root, resolved_session_id)
    if "pr_review_approval_gate" not in payload:
        if is_gh_approval:
            return _block_pr_approval(
                runtime_root,
                resolved_session_id,
                "github_cli",
                tool_name,
                command_prs,
                "approval attempted without staged PR-review gate context, so RSI cannot verify the fresh-subagent verdict",
            )
        if is_approval_delivery:
            return _block_pr_approval(
                runtime_root,
                resolved_session_id,
                "slack_delivery",
                tool_name,
                _extract_pr_numbers(delivery_text),
                "Slack approval delivery attempted without staged PR-review gate context, so RSI cannot verify the fresh-subagent verdict",
            )
        return None
    if not payload.get("pr_review_approval_gate"):
        return None

    if is_gh_approval:
        if not command_prs:
            return _block_pr_approval(
                runtime_root,
                resolved_session_id,
                "github_cli",
                tool_name,
                [],
                "gh pr review --approve did not include a PR number, so RSI cannot match it to a fresh-subagent verdict",
            )
        for pr_number in command_prs:
            ok, reason = _clean_pr_review_verdict_for_pr(runtime_root, resolved_session_id, pr_number)
            if not ok:
                return _block_pr_approval(runtime_root, resolved_session_id, "github_cli", tool_name, command_prs, reason)
        return None

    if is_approval_delivery_tool:
        text = delivery_text
        if _text_claims_pr_approval(text) and _looks_like_pr_review_context(payload, text):
            pr_numbers = _extract_pr_numbers(text)
            if not pr_numbers:
                pr_numbers = _extract_pr_numbers(_pr_review_context_text(payload))
            if not pr_numbers:
                return _block_pr_approval(
                    runtime_root,
                    resolved_session_id,
                    "slack_delivery",
                    tool_name,
                    [],
                    "Slack approval delivery did not identify a PR number, so RSI cannot match it to a fresh-subagent verdict",
                )
            for pr_number in pr_numbers:
                ok, reason = _clean_pr_review_verdict_for_pr(runtime_root, resolved_session_id, pr_number)
                if not ok:
                    return _block_pr_approval(
                        runtime_root,
                        resolved_session_id,
                        "slack_delivery",
                        tool_name,
                        pr_numbers,
                        reason,
                    )
    return None


def _normalized_tool_name(tool_name: str) -> str:
    return re.sub(r"[^a-z0-9]+", "_", str(tool_name or "").strip().lower()).strip("_")


def _is_approval_delivery_tool(tool_name: str) -> bool:
    normalized = _normalized_tool_name(tool_name)
    if not normalized:
        return False
    approval_names = {_normalized_tool_name(name) for name in _APPROVAL_DELIVERY_TOOL_NAMES}
    return normalized in approval_names or any(normalized.endswith(f"_{name}") for name in approval_names)


def cleanup_pr_review_workspace(runtime_root: Path, session_id: str) -> None:
    payload = _load_context(runtime_root, session_id)
    if not payload or (not payload.get("pr_review_approval_gate") and not payload.get("pr_review_workspace_root")):
        return
    target, reason = _safe_pr_review_workspace_root(payload, session_id)
    if target is None:
        _append_event(
            runtime_root,
            session_id,
            "pr_review.workspace_cleanup.skipped",
            {"event_type": "pr_review.workspace_cleanup.skipped", "status": "skipped", "reason": reason},
        )
        return

    existed = target.exists()
    try:
        if existed:
            shutil.rmtree(target)
    except OSError as exc:
        _append_event(
            runtime_root,
            session_id,
            "pr_review.workspace_cleanup.failed",
            {
                "event_type": "pr_review.workspace_cleanup.failed",
                "status": "failed",
                "path": str(target),
                "error": str(exc),
            },
        )
        return

    _append_event(
        runtime_root,
        session_id,
        "pr_review.workspace_cleanup.completed",
        {
            "event_type": "pr_review.workspace_cleanup.completed",
            "status": "completed",
            "path": str(target),
            "existed": existed,
        },
    )


def _context_path(runtime_root: Path, session_id: str) -> Path:
    explicit_path = os.getenv("RSI_RUNTIME_CONTEXT_PATH", "").strip()
    if explicit_path:
        return Path(explicit_path).expanduser()
    return runtime_root / "context" / f"{session_id}.json"


def _lifecycle_path(runtime_root: Path, session_id: str) -> Path:
    return runtime_root / "lifecycle" / f"{session_id}.jsonl"


def _append_event(runtime_root: Path, session_id: str, event: str, payload: JsonObject) -> None:
    path = _lifecycle_path(runtime_root, session_id)
    path.parent.mkdir(parents=True, exist_ok=True)
    item = {
        "event": event,
        "recorded_at_unix": time.time(),
        **(payload or {}),
    }
    with path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(item, sort_keys=True) + "\n")


def _load_context(runtime_root: Path, session_id: str) -> JsonObject:
    path = _context_path(runtime_root, session_id)
    if not path.exists():
        return {}
    try:
        parsed = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"Invalid RSI context payload for session {session_id}.") from exc
    if not isinstance(parsed, dict):
        raise RuntimeError(f"Invalid RSI context payload for session {session_id}.")
    return parsed


def _session_lifecycle_events(runtime_root: Path, session_id: str) -> list[JsonObject]:
    path = _lifecycle_path(runtime_root, session_id)
    if not path.exists():
        return []
    out: list[JsonObject] = []
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError:
        return []
    for line in lines:
        if not line.strip():
            continue
        try:
            parsed = json.loads(line)
        except json.JSONDecodeError:
            continue
        if isinstance(parsed, dict):
            out.append(parsed)
    return out


def _json_object_from_maybe_text(value: Any) -> JsonObject:
    if isinstance(value, dict):
        return value
    if not isinstance(value, str):
        return {}
    text = value.strip()
    if not text:
        return {}
    try:
        parsed = json.loads(text)
    except json.JSONDecodeError:
        start = text.find("{")
        end = text.rfind("}")
        if start < 0 or end <= start:
            return {}
        try:
            parsed = json.loads(text[start : end + 1])
        except json.JSONDecodeError:
            return {}
    return parsed if isinstance(parsed, dict) else {}


def _event_name(item: JsonObject) -> str:
    return str(item.get("event_type") or item.get("event") or "").strip()


def _event_payload(item: JsonObject) -> JsonObject:
    payload = item.get("payload")
    return payload if isinstance(payload, dict) else {}


def _event_tool_name(item: JsonObject) -> str:
    payload = _event_payload(item)
    return str(payload.get("tool_name") or item.get("tool_name") or "").strip()


def _event_tool_call_id(item: JsonObject) -> str:
    payload = _event_payload(item)
    return str(payload.get("tool_call_id") or item.get("tool_call_id") or "").strip()


def _event_args(item: JsonObject) -> Any:
    payload = _event_payload(item)
    if "args" in payload:
        return payload.get("args")
    if "args" in item:
        return item.get("args")
    if "request_payload" in item:
        return item.get("request_payload")
    return {}


def _event_result(item: JsonObject) -> Any:
    payload = _event_payload(item)
    if "result" in payload:
        return payload.get("result")
    return item.get("result")


def _text_mentions_pr(text: str, pr_number: int) -> bool:
    if pr_number <= 0:
        return False
    escaped = re.escape(str(pr_number))
    patterns = [
        rf"(?i)\bPR\s*#?\s*{escaped}\b",
        rf"(?i)\bpull\s*request\s*#?\s*{escaped}\b",
        rf"(?i)\bpull/\s*{escaped}\b",
        rf"(?<!\d)#\s*{escaped}\b",
    ]
    return any(re.search(pattern, text or "") for pattern in patterns)


def _extract_pr_numbers(text: str) -> list[int]:
    out: list[int] = []
    seen: set[int] = set()
    for pattern in (
        r"(?i)\bPR\s*#?\s*(\d{1,8})\b",
        r"(?i)\bpull\s*request\s*#?\s*(\d{1,8})\b",
        r"(?i)\bpull/\s*(\d{1,8})\b",
        r"(?<!\d)#\s*(\d{1,8})\b",
    ):
        for match in re.finditer(pattern, text or ""):
            try:
                number = int(match.group(1))
            except ValueError:
                continue
            if number > 0 and number not in seen:
                seen.add(number)
                out.append(number)
    return out


def _event_delegate_calls(runtime_root: Path, session_id: str) -> list[JsonObject]:
    starts_by_id: dict[str, JsonObject] = {}
    starts_in_order: list[JsonObject] = []
    calls: list[JsonObject] = []
    for item in _session_lifecycle_events(runtime_root, session_id):
        if _event_tool_name(item) != "delegate_task":
            continue
        name = _event_name(item)
        call_id = _event_tool_call_id(item)
        if name in {"tool.call.started", "tool_call_started"}:
            entry = {"tool_call_id": call_id, "args": _event_args(item)}
            starts_in_order.append(entry)
            if call_id:
                starts_by_id[call_id] = entry
            continue
        if name in {"tool.call.completed", "tool_call_completed"}:
            start = starts_by_id.get(call_id) if call_id else (starts_in_order[-1] if starts_in_order else {})
            calls.append(
                {
                    "tool_call_id": call_id,
                    "args": start.get("args") or _event_args(item),
                    "result": _event_result(item),
                }
            )
    return calls


def _delegate_result_for_task(results: list[Any], task_index: int) -> JsonObject:
    for result in results:
        if not isinstance(result, dict):
            continue
        try:
            index = int(result.get("task_index", -1))
        except (TypeError, ValueError):
            index = -1
        if index == task_index:
            return result
    if 0 <= task_index < len(results) and isinstance(results[task_index], dict):
        return results[task_index]
    return {}


def _matching_delegate_results_for_pr(runtime_root: Path, session_id: str, pr_number: int) -> list[JsonObject]:
    matches: list[JsonObject] = []
    for call in _event_delegate_calls(runtime_root, session_id):
        args = _json_object_from_maybe_text(call.get("args"))
        tasks = args.get("tasks")
        if not isinstance(tasks, list):
            tasks = []
        result_payload = _json_object_from_maybe_text(call.get("result"))
        results = result_payload.get("results")
        if not isinstance(results, list):
            results = []
        for index, task in enumerate(tasks):
            task_text = json.dumps(task, ensure_ascii=True, sort_keys=True, default=str)
            if not _text_mentions_pr(task_text, pr_number):
                continue
            result = _delegate_result_for_task(results, index)
            if result:
                matches.append(result)
        if not tasks and results:
            for result in results:
                if not isinstance(result, dict):
                    continue
                result_text = json.dumps(result, ensure_ascii=True, sort_keys=True, default=str)
                if _text_mentions_pr(result_text, pr_number):
                    matches.append(result)
    return matches


def _json_object_after_marker(text: str, marker: str) -> JsonObject:
    if marker not in text:
        return {}
    for line in text.splitlines():
        if marker not in line:
            continue
        candidate = line.split(marker, 1)[1].strip()
        if candidate.startswith(":"):
            candidate = candidate[1:].strip()
        parsed = _json_object_from_maybe_text(candidate)
        if parsed:
            return parsed
    return _json_object_from_maybe_text(text.split(marker, 1)[1])


def _bool_value(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        return value.strip().lower() in {"1", "true", "yes", "y", "on", "safe", "approve", "approved"}
    return False


def _int_value(value: Any, default: int = -1) -> int:
    if isinstance(value, bool):
        return default
    try:
        return int(value)
    except (TypeError, ValueError):
        return default


def _subagent_result_allows_pr_approval(result: JsonObject, pr_number: int) -> tuple[bool, str]:
    status = str(result.get("status") or "").strip().lower()
    if status not in {"completed", "success", "ok"}:
        return False, f"fresh subagent status is {status or 'missing'}"
    exit_reason = str(result.get("exit_reason") or "").strip().lower()
    if not exit_reason:
        return False, "fresh subagent exit_reason is missing"
    if exit_reason in _BLOCKING_SUBAGENT_EXIT_REASONS or exit_reason.startswith("max_iterations"):
        return False, f"fresh subagent ended with {exit_reason}"
    if exit_reason not in {"completed", "success", "ok", "normal_completion"}:
        return False, f"fresh subagent exit_reason is {exit_reason}"
    summary = str(result.get("summary") or "").strip()
    verdict = _json_object_after_marker(summary, PR_REVIEW_VERDICT_MARKER)
    if not verdict:
        return False, f"fresh subagent summary is missing {PR_REVIEW_VERDICT_MARKER}"
    verdict_pr = _int_value(verdict.get("pr_number"), default=pr_number)
    if verdict_pr != pr_number:
        return False, f"fresh subagent verdict is for PR #{verdict_pr}, not PR #{pr_number}"
    blocking_findings = _int_value(verdict.get("blocking_findings"), default=-1)
    verdict_text = str(verdict.get("verdict") or "").strip().lower()
    approval_safe = _bool_value(verdict.get("approval_safe") if "approval_safe" in verdict else verdict.get("safe_to_approve"))
    if not approval_safe:
        return False, "fresh subagent verdict says approval_safe is false or missing"
    if blocking_findings != 0:
        return False, f"fresh subagent verdict has blocking_findings={blocking_findings}"
    if verdict_text not in {"approve", "approved", "clean", "no_blocking_findings"}:
        return False, f"fresh subagent verdict is {verdict_text or 'missing'}"
    return True, "fresh subagent verdict allows approval"


def _clean_pr_review_verdict_for_pr(runtime_root: Path, session_id: str, pr_number: int) -> tuple[bool, str]:
    results = _matching_delegate_results_for_pr(runtime_root, session_id, pr_number)
    if not results:
        return False, f"no current-session fresh delegate_task result for PR #{pr_number}"
    return _subagent_result_allows_pr_approval(results[-1], pr_number)


def _shell_tokens(command: str) -> list[str]:
    try:
        return shlex.split(command)
    except ValueError:
        return command.split()


def _is_gh_token(token: str) -> bool:
    cleaned = str(token or "").strip().strip("\"'")
    if not cleaned:
        return False
    return cleaned.rstrip("/").rsplit("/", 1)[-1] == "gh"


def _extract_gh_pr_review_approval_prs(command: str) -> tuple[bool, list[int]]:
    text = str(command or "")
    if not re.search(r"(?i)(?:^|\s)--approve(?:[\s\"']|$)", text):
        return False, []

    numbers: list[int] = []
    seen: set[int] = set()
    saw_approval = False
    tokens = _shell_tokens(text)
    for index in range(0, max(0, len(tokens) - 2)):
        if not (_is_gh_token(tokens[index]) and tokens[index + 1 : index + 3] == ["pr", "review"]):
            continue
        segment: list[str] = []
        for token in tokens[index + 3 :]:
            if token in {"&&", "||", ";", "|"}:
                break
            segment.append(token)
        if "--approve" not in segment:
            continue
        saw_approval = True
        skip_next = False
        for token in segment:
            if skip_next:
                skip_next = False
                continue
            if token in {"--repo", "-R", "--body", "-b", "--body-file", "-F", "--comment"}:
                skip_next = True
                continue
            if token.startswith("-"):
                continue
            if token.isdigit():
                number = int(token)
                if number not in seen:
                    seen.add(number)
                    numbers.append(number)
                break
    for token in tokens:
        if token == text or not re.search(r"(?i)(?:^|[\s\"'])(?:\S*/)?gh\s+pr\s+review\b", token):
            continue
        nested_is_approval, nested_numbers = _extract_gh_pr_review_approval_prs(token)
        if nested_is_approval:
            saw_approval = True
            for number in nested_numbers:
                if number not in seen:
                    seen.add(number)
                    numbers.append(number)
    if not numbers:
        for match in re.finditer(r"(?i)(?:^|[;&|()\s\"'])(?:\S*/)?gh\s+pr\s+review\s+(\d{1,8})\b[^\n;|&]*--approve", text):
            saw_approval = True
            number = int(match.group(1))
            if number not in seen:
                seen.add(number)
                numbers.append(number)
    return saw_approval, numbers


def _extract_terminal_command(tool_name: str, args: JsonObject) -> str:
    normalized = str(tool_name or "").strip()
    if normalized not in {"terminal", "execute_code", "shell", "bash"}:
        return ""
    if not isinstance(args, dict):
        return ""
    for key in ("command", "cmd", "input", "code"):
        value = args.get(key)
        if isinstance(value, str) and value.strip():
            return value
    return ""


def _strings_from_value(value: Any) -> list[str]:
    if isinstance(value, str):
        return [value]
    if isinstance(value, dict):
        out: list[str] = []
        for item in value.values():
            out.extend(_strings_from_value(item))
        return out
    if isinstance(value, list):
        out: list[str] = []
        for item in value:
            out.extend(_strings_from_value(item))
        return out
    return []


def _approval_delivery_text(args: JsonObject) -> str:
    return "\n".join(_strings_from_value(args))


def _text_claims_pr_approval(text: str) -> bool:
    lower = str(text or "").lower()
    if not lower:
        return False
    negative_patterns = [
        r"\bnot\s+approved\b",
        r"\bnot\s+approve\b",
        r"\bcannot\s+approve\b",
        r"\bcan't\s+approve\b",
        r"\bdo\s+not\s+approve\b",
        r"\bnot\s+safe\s+to\s+(?:approve|merge)\b",
        r"\brequest\s+changes\b",
    ]
    if any(re.search(pattern, lower) for pattern in negative_patterns):
        return False
    return bool(
        re.search(
            r"\b(approved|approving|approve|lgtm|looks\s+good\s+to\s+merge|safe\s+to\s+merge|ready\s+to\s+merge|ship\s+it)\b",
            lower,
        )
    )


def _pr_review_context_text(payload: JsonObject, text: str = "") -> str:
    return "\n".join(
        [
            str(text or ""),
            str(payload.get("task_prompt") or ""),
            str(payload.get("context_summary") or ""),
            json.dumps(payload.get("context_refs") or [], ensure_ascii=True, sort_keys=True, default=str),
        ]
    )


def _looks_like_pr_review_context(payload: JsonObject, text: str) -> bool:
    haystack = _pr_review_context_text(payload, text)
    return bool(
        re.search(r"(?i)\b(PR|pull request|re-review|review)\b", haystack)
        and (re.search(r"(?i)\bPR\s*#?\s*\d+\b", haystack) or re.search(r"(?i)github\.com/.+?/pull/\d+", haystack))
    )


def _block_pr_approval(
    runtime_root: Path,
    session_id: str,
    surface: str,
    tool_name: str,
    pr_numbers: list[int],
    reason: str,
) -> JsonObject:
    if session_id:
        _append_event(
            runtime_root,
            session_id,
            "pr_review.approval.blocked",
            {
                "event_type": "pr_review.approval.blocked",
                "status": "blocked",
                "surface": surface,
                "tool_name": tool_name,
                "pr_numbers": pr_numbers,
                "reason": reason,
            },
        )
    return {"action": "block", "message": f"RSI blocked PR approval: {reason}"}


def _session_id_from_hook(task_id: str = "", session_id: str = "", **kwargs: Any) -> str:
    return str(task_id or session_id or kwargs.get("task_id", "") or kwargs.get("session_id", "") or "").strip()


def _safe_pr_review_workspace_root(payload: JsonObject, session_id: str) -> tuple[Path | None, str]:
    run_root_raw = str(payload.get("hermes_run_root") or "").strip()
    target_raw = str(payload.get("pr_review_workspace_root") or "").strip()
    if not run_root_raw:
        return None, "hermes_run_root unavailable"
    run_root = Path(run_root_raw).expanduser().resolve()
    target = Path(target_raw).expanduser().resolve() if target_raw else (run_root / "pr-review-worktrees" / session_id).resolve()
    if target == run_root:
        return None, "refusing to clean hermes_run_root"
    try:
        target.relative_to(run_root)
    except ValueError:
        return None, "refusing to clean PR review workspace outside hermes_run_root"
    return target, ""
