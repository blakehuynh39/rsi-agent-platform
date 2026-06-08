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
_PR_REVIEW_MUTATING_GIT_SUBCOMMANDS = {
    "add",
    "apply",
    "checkout",
    "cherry-pick",
    "clean",
    "clone",
    "commit",
    "fetch",
    "merge",
    "mv",
    "pull",
    "rebase",
    "reset",
    "restore",
    "rm",
    "switch",
}
_PR_REVIEW_MUTATING_GIT_STASH_ACTIONS = {"apply", "branch", "clear", "create", "drop", "pop", "push", "save", "store"}
_PR_REVIEW_MUTATING_GIT_WORKTREE_ACTIONS = {"add", "lock", "move", "prune", "remove", "repair", "rm", "unlock"}
_PR_REVIEW_FILE_WRITE_TOOL_NAMES = {
    "edit_file",
    "patch",
    "replace_file",
    "write_file",
    "workspace_write_file",
}
_PATH_ARG_KEYS = {
    "abs_path",
    "cwd",
    "dest",
    "destination",
    "file",
    "file_path",
    "filepath",
    "filename",
    "output",
    "output_path",
    "path",
    "target",
    "target_path",
    "workdir",
}
_PATCH_TEXT_ARG_KEYS = {"diff", "patch", "patch_text", "unified_diff"}


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
    if review_workspace_root and payload.get("pr_review_workspace_guard"):
        parts.append(
            "PR review isolated workspace root: "
            + review_workspace_root
            + ". Put any temporary PR-review clones or worktrees under this directory; RSI cleans it at session end. "
            + "Mutable git and file-write operations for PR review must stay under this root, not in shared /workspace/company checkouts."
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

    workspace_block = _pr_review_workspace_mutation_block(runtime_root, resolved_session_id, payload, tool_name, safe_args, command)
    if workspace_block:
        return workspace_block

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


def _pr_review_workspace_mutation_block(
    runtime_root: Path,
    session_id: str,
    payload: JsonObject,
    tool_name: str,
    args: JsonObject,
    command: str,
) -> JsonObject | None:
    if not payload.get("pr_review_workspace_guard"):
        return None

    target, reason = _safe_pr_review_workspace_root(payload, session_id)
    if target is None:
        if command and _mutating_git_subcommands(command):
            return _block_pr_review_workspace_mutation(
                runtime_root,
                session_id,
                tool_name,
                f"mutable git command blocked because the PR-review workspace root is unavailable: {reason}",
            )
        if _is_pr_review_file_write_tool(tool_name):
            return _block_pr_review_workspace_mutation(
                runtime_root,
                session_id,
                tool_name,
                f"file-write tool blocked because the PR-review workspace root is unavailable: {reason}",
            )
        return None

    if command:
        issue = _mutating_git_workspace_issue(command, target, initial_cwd=_terminal_initial_cwd(tool_name, args))
        if issue:
            return _block_pr_review_workspace_mutation(runtime_root, session_id, tool_name, issue)

    if _is_pr_review_file_write_tool(tool_name):
        issue = _file_write_workspace_issue(args, target)
        if issue:
            return _block_pr_review_workspace_mutation(runtime_root, session_id, tool_name, issue)
    return None


def _block_pr_review_workspace_mutation(runtime_root: Path, session_id: str, tool_name: str, reason: str) -> JsonObject:
    _append_event(
        runtime_root,
        session_id,
        "pr_review.workspace_mutation.blocked",
        {
            "event_type": "pr_review.workspace_mutation.blocked",
            "status": "blocked",
            "tool_name": tool_name,
            "reason": reason,
        },
    )
    return {"action": "block", "message": f"RSI blocked PR-review workspace mutation: {reason}"}


def _is_pr_review_file_write_tool(tool_name: str) -> bool:
    normalized = _normalized_tool_name(tool_name)
    if not normalized:
        return False
    write_names = {_normalized_tool_name(name) for name in _PR_REVIEW_FILE_WRITE_TOOL_NAMES}
    return normalized in write_names or any(normalized.endswith(f"_{name}") for name in write_names)


def _mutating_git_workspace_issue(command: str, workspace_root: Path, *, initial_cwd: Path | None = None) -> str:
    findings = _mutating_git_locations(command, initial_cwd=initial_cwd) + _mutating_gh_locations(command, initial_cwd=initial_cwd)
    for operation, cwd, target_paths in findings:
        if cwd is None:
            return (
                f"{operation} must run with an explicit `cd` or `git -C` under PR-review workspace root "
                f"{workspace_root}"
            )
        if not _path_is_within(cwd, workspace_root):
            return f"{operation} targets {cwd}, outside PR-review workspace root {workspace_root}"
        for target_path in target_paths:
            if target_path is None:
                return f"{operation} target path must resolve under PR-review workspace root {workspace_root}"
            if not _path_is_within(target_path, workspace_root):
                return f"{operation} target path {target_path} is outside PR-review workspace root {workspace_root}"

    if not findings and _mutating_git_subcommands(command):
        return f"mutable git command must run under PR-review workspace root {workspace_root}"
    return ""


def _file_write_workspace_issue(args: JsonObject, workspace_root: Path) -> str:
    base_cwd = _write_tool_base_cwd(args)
    write_paths = _write_tool_paths(args)
    if not write_paths:
        return f"file-write tool must include an explicit path under PR-review workspace root {workspace_root}"

    for raw_path in write_paths:
        resolved = _resolve_shell_path(raw_path, base_cwd)
        if resolved is None:
            return f"file-write path {raw_path!r} must be absolute or relative to an explicit cwd under {workspace_root}"
        if not _path_is_within(resolved, workspace_root):
            return f"file-write path {resolved} is outside PR-review workspace root {workspace_root}"
    return ""


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


def _shell_tokens_with_punctuation(command: str) -> list[str]:
    try:
        lexer = shlex.shlex(command, posix=True, punctuation_chars=True)
        lexer.whitespace_split = True
        return list(lexer)
    except (TypeError, ValueError):
        return _shell_tokens(command)


def _is_gh_token(token: str) -> bool:
    cleaned = str(token or "").strip().strip("\"'")
    if not cleaned:
        return False
    return cleaned.rstrip("/").rsplit("/", 1)[-1] == "gh"


def _is_git_token(token: str) -> bool:
    cleaned = str(token or "").strip().strip("\"'")
    if not cleaned:
        return False
    return cleaned.rstrip("/").rsplit("/", 1)[-1] == "git"


def _is_shell_token(token: str) -> bool:
    cleaned = str(token or "").strip().strip("\"'")
    if not cleaned:
        return False
    return cleaned.rstrip("/").rsplit("/", 1)[-1] in {"bash", "sh", "zsh"}


def _is_command_separator(token: str) -> bool:
    return str(token or "") in {"&&", "||", ";", "|", "&"}


def _path_is_within(path: Path, root: Path) -> bool:
    try:
        path.resolve().relative_to(root.resolve())
        return True
    except ValueError:
        return False


def _resolve_shell_path(raw_path: str, cwd: Path | None) -> Path | None:
    text = str(raw_path or "").strip().strip("\"'")
    if not text or "$" in text:
        return None
    path = Path(text).expanduser()
    if path.is_absolute():
        return path.resolve()
    if cwd is None:
        return None
    return (cwd / path).resolve()


def _command_operands(tokens: list[str], options_with_values: set[str]) -> list[str]:
    operands: list[str] = []
    index = 0
    while index < len(tokens):
        token = tokens[index]
        if token == "--":
            operands.extend(tokens[index + 1 :])
            break
        if token in options_with_values:
            index += 2
            continue
        if any(token.startswith(f"{option}=") for option in options_with_values if option.startswith("--")):
            index += 1
            continue
        if token.startswith("-"):
            index += 1
            continue
        operands.append(token)
        index += 1
    return operands


def _looks_like_path_operand(token: str) -> bool:
    text = str(token or "").strip().strip("\"'")
    if not text or "://" in text:
        return False
    return text.startswith(("/", "~", ".", "a/", "b/")) or "/" in text


def _generic_git_target_paths(remaining: list[str], cwd: Path | None) -> list[Path | None]:
    operands = _command_operands(
        remaining,
        {
            "-b",
            "-B",
            "-c",
            "-C",
            "-m",
            "-S",
            "--author",
            "--branch",
            "--config",
            "--date",
            "--depth",
            "--filter",
            "--message",
            "--reference",
            "--reference-if-able",
            "--shallow-exclude",
            "--shallow-since",
            "--template",
            "--upload-pack",
        },
    )
    return [_resolve_shell_path(operand, cwd) for operand in operands if _looks_like_path_operand(operand)]


def _git_subcommand_is_mutating(subcommand: str, remaining: list[str]) -> bool:
    if subcommand == "stash":
        operands = _command_operands(remaining, set())
        action = operands[0].lower() if operands else "push"
        return action in _PR_REVIEW_MUTATING_GIT_STASH_ACTIONS
    if subcommand == "worktree":
        operands = _command_operands(
            remaining,
            {
                "-b",
                "-B",
                "--lock",
                "--orphan",
                "--reason",
            },
        )
        action = operands[0].lower() if operands else "list"
        return action in _PR_REVIEW_MUTATING_GIT_WORKTREE_ACTIONS
    return subcommand in _PR_REVIEW_MUTATING_GIT_SUBCOMMANDS


def _git_subcommand_target_paths(subcommand: str, remaining: list[str], cwd: Path | None) -> list[Path | None]:
    if subcommand == "clone":
        targets: list[Path | None] = []
        operands = _command_operands(
            remaining,
            {
                "-b",
                "-c",
                "-j",
                "-o",
                "-u",
                "--branch",
                "--config",
                "--depth",
                "--filter",
                "--jobs",
                "--origin",
                "--reference",
                "--reference-if-able",
                "--separate-git-dir",
                "--shallow-exclude",
                "--shallow-since",
                "--template",
                "--upload-pack",
            },
        )
        if len(operands) >= 2:
            targets.append(_resolve_shell_path(operands[-1], cwd))
        if "--separate-git-dir" in remaining:
            index = remaining.index("--separate-git-dir")
            if index + 1 < len(remaining):
                targets.append(_resolve_shell_path(remaining[index + 1], cwd))
        for token in remaining:
            if token.startswith("--separate-git-dir="):
                targets.append(_resolve_shell_path(token.split("=", 1)[1], cwd))
        return targets

    if subcommand == "worktree" and remaining[:1] == ["add"]:
        operands = _command_operands(
            remaining[1:],
            {
                "-b",
                "-B",
                "--lock",
                "--orphan",
                "--reason",
            },
        )
        return [_resolve_shell_path(operands[0], cwd) if operands else None]

    return _generic_git_target_paths(remaining, cwd)


def _parse_git_segment(segment: list[str], cwd: Path | None) -> tuple[str, Path | None, list[Path | None], list[str]]:
    git_cwd = cwd
    target_paths: list[Path | None] = []
    index = 0
    while index < len(segment):
        token = segment[index]
        if not token:
            index += 1
            continue
        if token == "-C":
            if index + 1 >= len(segment):
                return "", git_cwd, target_paths, []
            git_cwd = _resolve_shell_path(segment[index + 1], git_cwd)
            index += 2
            continue
        if token in {"--git-dir", "--work-tree"}:
            if index + 1 < len(segment):
                target_paths.append(_resolve_shell_path(segment[index + 1], git_cwd))
            index += 2
            continue
        if token in {"-c", "--namespace"}:
            index += 2
            continue
        if token.startswith("--git-dir=") or token.startswith("--work-tree="):
            target_paths.append(_resolve_shell_path(token.split("=", 1)[1], git_cwd))
            index += 1
            continue
        if token.startswith("--namespace="):
            index += 1
            continue
        if token.startswith("-"):
            index += 1
            continue
        subcommand = token.lower()
        remaining = segment[index + 1 :]
        return subcommand, git_cwd, [*target_paths, *_git_subcommand_target_paths(subcommand, remaining, git_cwd)], remaining
    return "", git_cwd, target_paths, []


def _mutating_git_locations(command: str, *, initial_cwd: Path | None = None) -> list[tuple[str, Path | None, list[Path | None]]]:
    tokens = _shell_tokens_with_punctuation(command)
    findings: list[tuple[str, Path | None, list[Path | None]]] = []
    cwd = initial_cwd
    index = 0
    while index < len(tokens):
        token = tokens[index]
        if _is_command_separator(token):
            index += 1
            continue
        if token == "cd" and index + 1 < len(tokens):
            cwd = _resolve_shell_path(tokens[index + 1], cwd)
            index += 2
            continue
        if _is_shell_token(token):
            nested_index = index + 1
            while nested_index < len(tokens) and tokens[nested_index].startswith("-"):
                option = tokens[nested_index].lstrip("-")
                is_c_flag = option == "c" or (tokens[nested_index].startswith("-") and not tokens[nested_index].startswith("--") and option.endswith("c"))
                if is_c_flag and nested_index + 1 < len(tokens):
                    findings.extend(_mutating_git_locations(tokens[nested_index + 1], initial_cwd=cwd))
                    index = nested_index + 2
                    break
                nested_index += 1
            else:
                index += 1
            continue
        if _is_git_token(token):
            end = index + 1
            while end < len(tokens) and not _is_command_separator(tokens[end]):
                end += 1
            segment = tokens[index + 1 : end]
            subcommand, git_cwd, target_paths, remaining = _parse_git_segment(segment, cwd)
            if _git_subcommand_is_mutating(subcommand, remaining):
                findings.append((f"git {subcommand}", git_cwd, target_paths))
            index = end
            continue
        index += 1
    return findings


def _gh_repo_clone_targets(segment: list[str], cwd: Path | None) -> list[Path | None]:
    operands = _command_operands(
        segment[2:],
        {
            "-b",
            "-u",
            "--branch",
            "--upstream-remote-name",
        },
    )
    if len(operands) >= 2:
        return [_resolve_shell_path(operands[-1], cwd)]
    return []


def _mutating_gh_locations(command: str, *, initial_cwd: Path | None = None) -> list[tuple[str, Path | None, list[Path | None]]]:
    tokens = _shell_tokens_with_punctuation(command)
    findings: list[tuple[str, Path | None, list[Path | None]]] = []
    cwd = initial_cwd
    index = 0
    while index < len(tokens):
        token = tokens[index]
        if _is_command_separator(token):
            index += 1
            continue
        if token == "cd" and index + 1 < len(tokens):
            cwd = _resolve_shell_path(tokens[index + 1], cwd)
            index += 2
            continue
        if _is_shell_token(token):
            nested_index = index + 1
            while nested_index < len(tokens) and tokens[nested_index].startswith("-"):
                option = tokens[nested_index].lstrip("-")
                is_c_flag = option == "c" or (tokens[nested_index].startswith("-") and not tokens[nested_index].startswith("--") and option.endswith("c"))
                if is_c_flag and nested_index + 1 < len(tokens):
                    findings.extend(_mutating_gh_locations(tokens[nested_index + 1], initial_cwd=cwd))
                    index = nested_index + 2
                    break
                nested_index += 1
            else:
                index += 1
            continue
        if _is_gh_token(token):
            end = index + 1
            while end < len(tokens) and not _is_command_separator(tokens[end]):
                end += 1
            segment = tokens[index + 1 : end]
            if len(segment) >= 2 and segment[0:2] == ["pr", "checkout"]:
                findings.append(("gh pr checkout", cwd, []))
            elif len(segment) >= 2 and segment[0:2] == ["repo", "clone"]:
                findings.append(("gh repo clone", cwd, _gh_repo_clone_targets(segment, cwd)))
            index = end
            continue
        index += 1
    return findings


def _mutating_git_subcommands(command: str) -> list[str]:
    return [
        subcommand
        for subcommand, _cwd, _targets in [*_mutating_git_locations(command), *_mutating_gh_locations(command)]
    ]


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


def _terminal_initial_cwd(tool_name: str, args: JsonObject) -> Path | None:
    normalized = str(tool_name or "").strip()
    if normalized not in {"terminal", "execute_code", "shell", "bash"}:
        return None
    for key in ("cwd", "workdir"):
        value = args.get(key)
        if isinstance(value, str) and value.strip():
            return _resolve_shell_path(value, None)
    return None


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


def _write_tool_base_cwd(args: JsonObject) -> Path | None:
    for key in ("cwd", "workdir"):
        value = args.get(key)
        if isinstance(value, str) and value.strip():
            return _resolve_shell_path(value, None)
    return None


def _clean_patch_path(raw_path: str) -> str:
    text = str(raw_path or "").strip().strip("\"'")
    if "\t" in text:
        text = text.split("\t", 1)[0].strip()
    if text.startswith("a/") or text.startswith("b/"):
        text = text[2:]
    return text


def _paths_from_patch_text(text: str) -> list[str]:
    if "*** " not in text and "+++" not in text and "---" not in text:
        return []
    out: list[str] = []
    for line in text.splitlines():
        match = re.match(r"\s*\*\*\*\s+(?:Add|Update|Delete)\s+File:\s+(.+?)\s*$", line)
        if not match:
            match = re.match(r"\s*(?:---|\+\+\+)\s+(.+?)\s*$", line)
        if not match:
            continue
        path = _clean_patch_path(match.group(1))
        if path and path != "/dev/null":
            out.append(path)
    return out


def _write_tool_paths(args: JsonObject) -> list[str]:
    out: list[str] = []

    def walk(value: Any, key: str = "") -> None:
        normalized_key = str(key or "").strip().lower()
        if isinstance(value, str):
            if normalized_key in _PATH_ARG_KEYS and normalized_key not in {"cwd", "workdir"}:
                out.append(value)
                return
            if normalized_key in _PATCH_TEXT_ARG_KEYS:
                out.extend(_paths_from_patch_text(value))
            return
        if isinstance(value, dict):
            for item_key, item_value in value.items():
                walk(item_value, str(item_key))
            return
        if isinstance(value, list):
            for item in value:
                walk(item, key)

    walk(args)
    deduped: list[str] = []
    seen: set[str] = set()
    for path in out:
        cleaned = _clean_patch_path(path)
        if cleaned and cleaned not in seen:
            seen.add(cleaned)
            deduped.append(cleaned)
    return deduped


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
