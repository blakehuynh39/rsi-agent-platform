from __future__ import annotations

from dataclasses import dataclass
import hashlib
import json
from pathlib import Path
import shutil
import time
from typing import Any

from .json_types import JsonObject


EXECUTION_CONTRACT_VERSION = "execution-envelope/v2"
RUNNER_PLANNER_MODE = "runner_first"

ALLOWED_PHASE_TYPES = frozenset({"plan", "investigate", "operate", "render", "deliver", "reflect"})


def _string(value: Any) -> str:
    if value is None:
        return ""
    return str(value).strip()


def _json_object(value: Any) -> JsonObject:
    return value if isinstance(value, dict) else {}


def _json_object_list(value: Any) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    return [item for item in value if isinstance(item, dict)]


def _string_list(value: Any) -> list[str]:
    if not isinstance(value, list):
        return []
    out: list[str] = []
    for item in value:
        text = _string(item)
        if text:
            out.append(text)
    return out


def _safe_path_segment(value: Any, default: str) -> str:
    text = _string(value)
    out = "".join(ch if ch.isalnum() or ch in {"-", "_", "."} else "-" for ch in text).strip(".-")
    return out[:96] or default


def _json_object_from_string(value: Any) -> JsonObject:
    if isinstance(value, dict):
        return value
    if not isinstance(value, str) or not value.strip():
        return {}
    try:
        parsed = json.loads(value)
    except (TypeError, ValueError, json.JSONDecodeError):
        return {}
    return parsed if isinstance(parsed, dict) else {}


def _bool(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        return value.strip().lower() in {"1", "true", "yes", "y", "on"}
    return bool(value)


def _now() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


@dataclass
class ContractValidationResult:
    ok: bool
    errors: list[str]


def phase_type_for_task(task: Any) -> str:
    phase = _string(getattr(task, "execution_phase", "")).lower()
    if phase == "investigate":
        return "investigate"
    if phase == "render":
        return "render"
    if phase == "deliver":
        return "deliver"
    if phase == "reflect":
        return "reflect"
    task_type = _string(getattr(task, "task_type", "")).lower()
    if task_type in {"eval"}:
        return "reflect"
    if task_type in {"proposal", "repo-change"}:
        return "operate"
    return "operate"


def default_delivery_policy(task: Any) -> JsonObject:
    mode = _string(getattr(task, "reply_delivery_mode", "")).lower()
    channel_id = _string(getattr(task, "channel_id", ""))
    thread_ts = _string(getattr(task, "thread_ts", ""))
    trace_id = _string(getattr(task, "trace_id", ""))
    return {
        "bound_channel_id": channel_id,
        "bound_thread_ts": thread_ts,
        "direct_send_allowed": mode == "direct",
        "upload_allowed": mode == "direct",
        "idempotency_key_base": ":".join(part for part in [channel_id, thread_ts, trace_id] if part),
    }


def default_workspace_policy(task: Any, *, computer_root: str, run_root: str, artifact_root: str) -> JsonObject:
    return {
        "computer_root": computer_root,
        "run_root": run_root,
        "artifact_root": artifact_root,
        "allowed_path_roots": [computer_root, run_root, artifact_root],
    }


def default_approval_policy(task: Any) -> JsonObject:
    return {
        "direct_slack_allowed": _string(getattr(task, "reply_delivery_mode", "")).lower() == "direct",
        "requires_approval": [
            "repo_merge",
            "platform_config",
            "deployment",
            "k8s_mutation",
            "aws_iac",
            "destructive_action",
            "harness_platform_behavior_change",
        ],
        "platform_mutations_execute_directly": False,
    }


def phase_run(
    *,
    phase_id: str,
    phase_type: str,
    status: str,
    input_refs: list[str] | None = None,
    output_refs: list[str] | None = None,
    completion_verdict: str = "",
    termination_reason: str = "",
    failure_class: str = "",
    failure_reason: str = "",
) -> JsonObject:
    item: JsonObject = {
        "phase_id": phase_id,
        "phase_type": phase_type if phase_type in ALLOWED_PHASE_TYPES else "operate",
        "status": status,
        "input_refs": list(input_refs or []),
        "output_refs": list(output_refs or []),
        "completion_verdict": completion_verdict,
        "termination_reason": termination_reason,
    }
    if failure_class or failure_reason:
        item["failure"] = {"class": failure_class, "reason": failure_reason}
    return item


def ledger_event(
    kind: str,
    *,
    phase_id: str = "",
    status: str = "",
    payload: JsonObject | None = None,
    sequence: int | None = None,
    idempotency_key: str = "",
    recorded_at: str = "",
    event_id: str = "",
) -> JsonObject:
    item: JsonObject = {
        "event_id": event_id or f"ledger-{hashlib.sha1((kind + phase_id + str(time.time_ns())).encode('utf-8')).hexdigest()[:16]}",
        "kind": kind,
        "phase_id": phase_id,
        "status": status,
        "recorded_at": recorded_at or _now(),
        "payload": payload or {},
    }
    if sequence is not None:
        item["sequence"] = sequence
    if idempotency_key:
        item["idempotency_key"] = idempotency_key
    return item


def artifact_record(item: JsonObject, *, execution_id: str) -> JsonObject:
    refs = _string_list(item.get("artifact_refs"))
    file_ref = _string(item.get("file_ref")) or (refs[0] if refs else "")
    workspace_path = _string(item.get("workspace_path"))
    if not workspace_path and file_ref.startswith("file://"):
        workspace_path = file_ref.removeprefix("file://")
    record: JsonObject = {
        "kind": _string(item.get("kind")) or "artifact",
        "title": _string(item.get("title")),
        "artifact_refs": refs or ([file_ref] if file_ref else []),
        "workspace_path": workspace_path,
        "file_ref": file_ref or (f"file://{workspace_path}" if workspace_path else ""),
        "size_bytes": item.get("size_bytes", 0),
        "sha256": _string(item.get("sha256")),
        "created_by_execution_id": _string(item.get("created_by_execution_id")) or execution_id,
        "share_status": _string(item.get("share_status")) or _string(item.get("delivery_status")) or "local",
        "failure_reason": _string(item.get("failure_reason")),
    }
    if workspace_path:
        path = Path(workspace_path)
        try:
            if path.is_file():
                record["size_bytes"] = int(path.stat().st_size)
                if not record["sha256"]:
                    digest = hashlib.sha256()
                    with path.open("rb") as handle:
                        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
                            digest.update(chunk)
                    record["sha256"] = digest.hexdigest()
        except OSError:
            pass
    return record


def delivery_record(item: JsonObject, *, execution_id: str) -> JsonObject:
    return {
        "delivery_id": _string(item.get("tool_call_id")) or f"delivery-{hashlib.sha1(execution_id.encode('utf-8')).hexdigest()[:12]}",
        "kind": "slack_message",
        "send_status": _string(item.get("send_status")) or _string(item.get("status")),
        "channel_id": _string(item.get("channel_id")),
        "thread_ts": _string(item.get("thread_ts")),
        "body": _string(item.get("body")),
        "body_sha1": _string(item.get("body_sha1")),
        "body_excerpt": _string(item.get("body_excerpt")),
        "tool_call_id": _string(item.get("tool_call_id")),
        "tool_name": _string(item.get("tool_name")),
        "provider_ref": _string(item.get("provider_ref")),
        "message_link": _string(item.get("message_link")),
        "artifact_refs": _string_list(item.get("artifact_refs")),
        "created_by_execution_id": execution_id,
    }


def structured_output_from_envelope(envelope: JsonObject, fallback: JsonObject | None = None) -> JsonObject:
    out = dict(fallback or {})
    final_response = _string(envelope.get("final_response"))
    if final_response and not _string(out.get("final_answer")):
        out["final_answer"] = final_response
    if "produced_artifacts" not in out:
        out["produced_artifacts"] = list(_json_object_list(envelope.get("artifacts")))
    if "reply_delivery" not in out:
        deliveries = _json_object_list(envelope.get("deliveries"))
        out["reply_delivery"] = deliveries[0] if deliveries else {}
    completion = _json_object(envelope.get("completion"))
    if completion:
        if "completion_verdict" not in out:
            out["completion_verdict"] = _string(completion.get("completion_verdict"))
        if "termination_reason" not in out:
            out["termination_reason"] = _string(completion.get("termination_reason"))
    out.setdefault("visible_reasoning", [])
    out.setdefault("proposed_actions", [])
    out.setdefault("knowledge_drafts", [])
    out.setdefault("outcome_hypotheses", [])
    out.setdefault("artifact_failure_reason", "")
    return out


class HermesCompanyComputer:
    def __init__(self, *, computer_root: str, run_root: str, artifact_root: str, hermes_pin: str) -> None:
        self.computer_root = computer_root
        self.run_root = run_root
        self.artifact_root = artifact_root
        self.hermes_pin = hermes_pin

    def planned_phase_runs(self, task: Any) -> list[JsonObject]:
        phases: list[tuple[str, str]] = []
        task_type = _string(getattr(task, "task_type", "")).lower()
        explicit_phase = _string(getattr(task, "execution_phase", ""))
        reply_delivery_mode = _string(getattr(task, "reply_delivery_mode", "")).lower()
        if explicit_phase:
            phase_type = phase_type_for_task(task)
            phases.append((explicit_phase, phase_type))
        elif task_type in {"eval"}:
            phases.extend([("plan", "plan"), ("reflect", "reflect")])
        elif task_type in {"proposal", "repo-change"}:
            phases.extend([("plan", "plan"), ("operate", "operate"), ("reflect", "reflect")])
        else:
            phases.extend([("plan", "plan"), ("operate", phase_type_for_task(task))])
            if reply_delivery_mode == "direct":
                phases.append(("deliver", "deliver"))
            phases.append(("reflect", "reflect"))
        out: list[JsonObject] = []
        seen: set[str] = set()
        for phase_id, phase_type in phases:
            if phase_id in seen:
                continue
            seen.add(phase_id)
            out.append(
                phase_run(
                    phase_id=phase_id,
                    phase_type=phase_type,
                    status="planned",
                )
            )
        return out

    def validate_task(self, task: Any) -> ContractValidationResult:
        return ContractValidationResult(ok=True, errors=[])

    def attach_envelope(self, task: Any, result: Any, *, observer: Any | None = None) -> Any:
        raw = _json_object(getattr(result, "raw", {}))
        if _json_object(raw.get("execution_envelope")):
            return result
        execution_id = _string(getattr(observer, "execution_id", "")) or _string(getattr(task, "execution_id", ""))
        if not execution_id:
            execution_id = _string(raw.get("execution_id")) or _string(raw.get("observation_execution_id")) or f"hexec-{time.time_ns()}"
        observer_events = self._observer_events(observer)
        structured_output = _json_object(raw.get("structured_output"))
        artifacts = self._artifact_records(task, raw, structured_output, observer_events=observer_events, execution_id=execution_id)
        deliveries = self._delivery_records(raw, structured_output, execution_id=execution_id)
        phase_runs = self._phase_runs(
            task,
            raw,
            result_ok=bool(getattr(result, "ok", False)),
            artifacts=artifacts,
            deliveries=deliveries,
            observer_events=observer_events,
        )
        ledger_events = self._ledger_events(
            raw,
            phase_runs,
            artifacts,
            deliveries,
            observer_events=observer_events,
            execution_id=execution_id,
        )
        memory_events = [event for event in ledger_events if _string(event.get("kind")).startswith("memory.")]
        final_response = _string(structured_output.get("final_answer")) or _string(structured_output.get("reply_draft")) or _string(getattr(result, "message", ""))
        completion_verdict = _string(raw.get("completion_verdict")) or _string(structured_output.get("completion_verdict")) or ("failed" if not getattr(result, "ok", False) else "complete")
        termination_reason = _string(raw.get("termination_reason")) or _string(structured_output.get("termination_reason")) or ("failure" if not getattr(result, "ok", False) else "normal_completion")
        phase_failed = any(_string(phase.get("status")).lower() == "failed" for phase in phase_runs)
        if phase_failed and completion_verdict == "complete":
            completion_verdict = "failed"
            termination_reason = "phase_failed"
        envelope: JsonObject = {
            "contract_version": EXECUTION_CONTRACT_VERSION,
            "execution_id": execution_id,
            "operation_id": _string(getattr(task, "operation_id", "")),
            "trace_id": _string(getattr(task, "trace_id", "")),
            "workflow_id": _string(getattr(task, "workflow_id", "")),
            "session_id": _string(raw.get("hermes_session_id")) or _string(raw.get("session_id")),
            "execution_intent": _json_object(getattr(task, "execution_intent", {}))
            or {
                "task_type": _string(getattr(task, "task_type", "")),
                "intent": _string(getattr(task, "intent", "")),
                "repo": _string(getattr(task, "repo", "")),
                "mode": RUNNER_PLANNER_MODE,
            },
            "delivery_policy": _json_object(getattr(task, "delivery_policy", {})) or default_delivery_policy(task),
            "workspace_policy": _json_object(getattr(task, "workspace_policy", {}))
            or default_workspace_policy(task, computer_root=self.computer_root, run_root=self.run_root, artifact_root=self.artifact_root),
            "approval_policy": _json_object(getattr(task, "approval_policy", {})) or default_approval_policy(task),
            "execution_plan": {
                "planner": "HermesCompanyComputer",
                "mode": RUNNER_PLANNER_MODE,
                "phases": [
                    {
                        "phase_id": _string(phase.get("phase_id")),
                        "phase_type": _string(phase.get("phase_type")),
                    }
                    for phase in phase_runs
                ],
            },
            "phase_runs": phase_runs,
            "ledger_events": ledger_events,
            "artifacts": artifacts,
            "deliveries": deliveries,
            "memory_events": memory_events,
            "completion": {
                "completion_verdict": completion_verdict,
                "termination_reason": termination_reason,
                "partial": completion_verdict == "partial",
                "max_iterations_reached": _bool(raw.get("max_iterations_reached")),
                "ok": bool(getattr(result, "ok", False)) and not phase_failed,
            },
            "final_response": final_response,
        }
        raw["execution_envelope"] = envelope
        raw["structured_output"] = structured_output_from_envelope(envelope, structured_output)
        setattr(result, "raw", raw)
        return result

    def failure_result_raw(self, task: Any, *, errors: list[str]) -> JsonObject:
        return {
            "failure_class": "runner_contract_failed",
            "runner_diagnostics": {
                "failure_kind": "runner_contract_failed",
                "termination_reason": "runner_contract_failed",
                "contract_errors": list(errors),
            },
            "trace_id": _string(getattr(task, "trace_id", "")),
            "workflow_id": _string(getattr(task, "workflow_id", "")),
            "operation_id": _string(getattr(task, "operation_id", "")),
            "structured_output": {
                "visible_reasoning": [],
                "reply_draft": "",
                "final_answer": "",
                "confidence": 0,
                "context_summary": "",
                "self_critique": "Runner task failed the execution contract before model execution.",
                "proposed_actions": [],
                "reply_delivery": {},
                "knowledge_drafts": [],
                "outcome_hypotheses": [],
                "produced_artifacts": [],
                "artifact_failure_reason": "",
                "completion_verdict": "failed",
                "termination_reason": "runner_contract_failed",
            },
        }

    def _artifact_records(
        self,
        task: Any,
        raw: JsonObject,
        structured_output: JsonObject,
        *,
        observer_events: list[JsonObject] | None = None,
        execution_id: str,
    ) -> list[JsonObject]:
        items: list[JsonObject] = []
        items.extend(_json_object_list(structured_output.get("produced_artifacts")))
        items.extend(_json_object_list(raw.get("produced_artifacts")))
        seen_refs: set[str] = set()
        for path in _string_list(raw.get("native_artifact_paths")):
            items.append({"kind": "artifact", "workspace_path": path, "file_ref": f"file://{path}", "artifact_refs": [f"file://{path}"]})
        items.extend(self._generic_file_write_artifacts(task, observer_events or [], execution_id=execution_id))
        out: list[JsonObject] = []
        for item in items:
            record = artifact_record(item, execution_id=execution_id)
            key = _string(record.get("file_ref")) or "|".join(_string_list(record.get("artifact_refs")))
            if key and key in seen_refs:
                continue
            if key:
                seen_refs.add(key)
            out.append(record)
        return out

    def _generic_file_write_artifacts(self, task: Any, observer_events: list[JsonObject], *, execution_id: str) -> list[JsonObject]:
        if not self._task_has_artifact_intent(task):
            return []
        out: list[JsonObject] = []
        for event in observer_events:
            if _string(event.get("event_type")) != "tool.call.completed" or _string(event.get("status")).lower() != "completed":
                continue
            payload = _json_object(event.get("payload"))
            tool_name = _string(payload.get("tool_name")).lower()
            if tool_name not in {"write_file", "workspace_write_file", "workspace.write_file"}:
                continue
            args = _json_object(payload.get("args"))
            source_path = self._resolve_workspace_path(task, args.get("path") or payload.get("path"))
            if source_path is None or not self._path_allowed_for_artifact(task, source_path) or not source_path.is_file():
                continue
            artifact_path = self._canonical_artifact_path(task, source_path)
            source_text = str(source_path)
            if artifact_path != source_path:
                try:
                    artifact_path.parent.mkdir(parents=True, exist_ok=True)
                    if artifact_path.exists() and source_path.resolve() != artifact_path.resolve():
                        artifact_path = self._dedupe_artifact_path(artifact_path)
                    shutil.copy2(source_path, artifact_path)
                except OSError:
                    artifact_path = source_path
            result = _json_object_from_string(payload.get("result"))
            size_bytes = result.get("bytes_written") if isinstance(result.get("bytes_written"), int) else 0
            item: JsonObject = {
                "kind": self._requested_artifact_kind(task),
                "title": artifact_path.name,
                "workspace_path": str(artifact_path),
                "file_ref": f"file://{artifact_path}",
                "artifact_refs": [f"file://{artifact_path}"],
                "size_bytes": size_bytes,
                "created_by_execution_id": execution_id,
                "share_status": "local",
                "source": "generic_write_file",
            }
            if str(artifact_path) != source_text:
                item["source_workspace_path"] = source_text
            out.append(item)
        return out

    def _task_has_artifact_intent(self, task: Any) -> bool:
        intent = _json_object(getattr(task, "execution_intent", {}))
        haystack = " ".join(
            _string(intent.get(key)).lower()
            for key in ("kind", "intent", "user_request", "task_type")
        )
        return any(marker in haystack for marker in ("artifact", "diagram", "architecture"))

    def _requested_artifact_kind(self, task: Any) -> str:
        intent = _json_object(getattr(task, "execution_intent", {}))
        return _string(intent.get("kind")) or "artifact"

    def _workspace_policy(self, task: Any) -> JsonObject:
        return _json_object(getattr(task, "workspace_policy", {})) or default_workspace_policy(
            task,
            computer_root=self.computer_root,
            run_root=self.run_root,
            artifact_root=self.artifact_root,
        )

    def _resolve_workspace_path(self, task: Any, path_value: Any) -> Path | None:
        path_text = _string(path_value)
        if not path_text:
            return None
        path = Path(path_text).expanduser()
        if path.is_absolute():
            return path
        policy = self._workspace_policy(task)
        root = _string(policy.get("computer_root")) or self.computer_root
        return Path(root).expanduser() / path

    def _path_allowed_for_artifact(self, task: Any, path: Path) -> bool:
        try:
            resolved = path.resolve()
        except OSError:
            resolved = path.absolute()
        policy = self._workspace_policy(task)
        roots = _string_list(policy.get("allowed_path_roots")) or [self.computer_root, self.run_root, self.artifact_root]
        for root in roots:
            if not root:
                continue
            try:
                resolved.relative_to(Path(root).expanduser().resolve())
                return True
            except (OSError, ValueError):
                continue
        return False

    def _canonical_artifact_path(self, task: Any, source_path: Path) -> Path:
        try:
            resolved_source = source_path.resolve()
        except OSError:
            resolved_source = source_path.absolute()
        policy = self._workspace_policy(task)
        artifact_root = Path(_string(policy.get("artifact_root")) or self.artifact_root).expanduser()
        try:
            resolved_source.relative_to(artifact_root.resolve())
            return resolved_source
        except (OSError, ValueError):
            pass
        intent = _json_object(getattr(task, "execution_intent", {}))
        repo = _safe_path_segment(getattr(task, "repo", "") or intent.get("repo"), "workspace")
        operation = _safe_path_segment(getattr(task, "operation_id", "") or getattr(task, "trace_id", ""), "manual")
        return artifact_root / repo / time.strftime("%Y-%m-%d", time.gmtime()) / operation / source_path.name

    def _dedupe_artifact_path(self, path: Path) -> Path:
        for index in range(2, 1000):
            candidate = path.with_name(f"{path.stem}-{index}{path.suffix}")
            if not candidate.exists():
                return candidate
        return path.with_name(f"{path.stem}-{time.time_ns()}{path.suffix}")

    def _delivery_records(self, raw: JsonObject, structured_output: JsonObject, *, execution_id: str) -> list[JsonObject]:
        candidates = [_json_object(raw.get("reply_delivery")), _json_object(structured_output.get("reply_delivery"))]
        out: list[JsonObject] = []
        seen: set[str] = set()
        for item in candidates:
            if not item:
                continue
            record = delivery_record(item, execution_id=execution_id)
            key = _string(record.get("delivery_id"))
            if key in seen:
                continue
            seen.add(key)
            out.append(record)
        return out

    def _observer_events(self, observer: Any | None) -> list[JsonObject]:
        if observer is None or not hasattr(observer, "events"):
            return []
        try:
            return _json_object_list(observer.events())
        except Exception:
            return []

    def _phase_runs(
        self,
        task: Any,
        raw: JsonObject,
        *,
        result_ok: bool,
        artifacts: list[JsonObject],
        deliveries: list[JsonObject],
        observer_events: list[JsonObject],
    ) -> list[JsonObject]:
        planned = self.planned_phase_runs(task)
        explicit_phase = _string(getattr(task, "execution_phase", ""))
        phase_type = phase_type_for_task(task)
        status = "completed" if result_ok else "failed"
        completion_verdict = _string(raw.get("completion_verdict")) or ("complete" if result_ok else "failed")
        termination_reason = _string(raw.get("termination_reason")) or ("normal_completion" if result_ok else _string(raw.get("failure_class")) or "failure")
        output_refs: list[str] = []
        for artifact in artifacts:
            output_refs.extend(_string_list(artifact.get("artifact_refs")))
        for delivery in deliveries:
            ref = _string(delivery.get("message_link")) or _string(delivery.get("provider_ref"))
            if ref:
                output_refs.append(ref)
        observed_phase_status: dict[str, JsonObject] = {}
        for event in observer_events:
            if _string(event.get("event_type")) == "phase.completed":
                phase_id = _string(event.get("phase"))
                if phase_id:
                    payload = _json_object(event.get("payload"))
                    observed_phase_status[phase_id] = {
                        "status": _string(event.get("status")) or "completed",
                        "completion_verdict": _string(payload.get("completion_verdict")),
                        "termination_reason": _string(payload.get("termination_reason")),
                    }
        diagnostics = _json_object(raw.get("runner_diagnostics"))
        if _bool(diagnostics.get("artifact_phase_enabled")) and not explicit_phase:
            render_status = "completed" if artifacts else ("failed" if _string(raw.get("artifact_failure_reason")) else "skipped")
            deliver_status = "completed" if deliveries else ("skipped" if _string(getattr(task, "reply_delivery_mode", "")).lower() != "direct" else "failed")
            if deliveries and _string(deliveries[0].get("send_status")).lower() not in {"posted", "sent", "uploaded", "completed", "ok", "success", "shared"}:
                deliver_status = "failed"
            deliver_completion_verdict = ""
            if deliver_status == "completed":
                deliver_completion_verdict = _string(diagnostics.get("artifact_delivery_completion_verdict")) or completion_verdict
            phase_map: dict[str, JsonObject] = {str(item.get("phase_id")): dict(item) for item in planned}
            investigate_observed = observed_phase_status.get("investigate", {})
            phase_map["investigate"] = phase_run(
                phase_id="investigate",
                phase_type="investigate",
                status=_string(investigate_observed.get("status")) or ("completed" if result_ok else "failed"),
                completion_verdict=_string(investigate_observed.get("completion_verdict")) or _string(diagnostics.get("artifact_investigate_completion_verdict")) or completion_verdict,
                termination_reason=_string(investigate_observed.get("termination_reason")) or _string(diagnostics.get("artifact_investigate_termination_reason")) or termination_reason,
            )
            render_observed = observed_phase_status.get("render", {})
            phase_map["render"] = phase_run(
                phase_id="render",
                phase_type="render",
                status=_string(render_observed.get("status")) or render_status,
                output_refs=[ref for artifact in artifacts for ref in _string_list(artifact.get("artifact_refs"))],
                completion_verdict=_string(render_observed.get("completion_verdict")) or (completion_verdict if render_status == "completed" else ""),
                termination_reason=_string(render_observed.get("termination_reason")) or (_string(raw.get("artifact_failure_reason")) if render_status != "completed" else "normal_completion"),
                failure_class="artifact_render_failed" if render_status == "failed" else "",
                failure_reason=_string(raw.get("artifact_failure_reason")) if render_status == "failed" else "",
            )
            deliver_observed = observed_phase_status.get("deliver", {})
            phase_map["deliver"] = phase_run(
                phase_id="deliver",
                phase_type="deliver",
                status=_string(deliver_observed.get("status")) or deliver_status,
                input_refs=[ref for artifact in artifacts for ref in _string_list(artifact.get("artifact_refs"))],
                output_refs=output_refs,
                completion_verdict=_string(deliver_observed.get("completion_verdict")) or deliver_completion_verdict,
                termination_reason=_string(deliver_observed.get("termination_reason")) or (_string(diagnostics.get("direct_delivery_phase_failed")) if deliver_status == "failed" else ("normal_completion" if deliver_status == "completed" else "skipped")),
                failure_class="artifact_delivery_failed" if deliver_status == "failed" else "",
                failure_reason=_string(diagnostics.get("direct_delivery_phase_failed")) if deliver_status == "failed" else "",
            )
            reflect_observed = observed_phase_status.get("reflect", {})
            phase_map["reflect"] = phase_run(
                phase_id="reflect",
                phase_type="reflect",
                status=_string(reflect_observed.get("status")) or ("completed" if result_ok else "failed"),
                completion_verdict=_string(reflect_observed.get("completion_verdict")) or completion_verdict,
                termination_reason=_string(reflect_observed.get("termination_reason")) or termination_reason,
            )
            plan_observed = observed_phase_status.get("plan", {})
            phase_map["plan"] = phase_run(
                phase_id="plan",
                phase_type="plan",
                status=_string(plan_observed.get("status")) or "completed",
                completion_verdict=_string(plan_observed.get("completion_verdict")) or "complete",
                termination_reason=_string(plan_observed.get("termination_reason")) or "phase_graph_created",
            )
            return [
                phase_map[_string(item.get("phase_id"))]
                for item in planned
                if _string(item.get("phase_id")) in phase_map
            ]
        if len(planned) > 1 and not explicit_phase:
            phase_map = {str(item.get("phase_id")): dict(item) for item in planned}
            main_phase_id = "operate" if "operate" in phase_map else ("reflect" if "reflect" in phase_map else _string(planned[-1].get("phase_id")))
            plan_observed = observed_phase_status.get("plan", {})
            phase_map["plan"] = phase_run(
                phase_id="plan",
                phase_type="plan",
                status=_string(plan_observed.get("status")) or "completed",
                completion_verdict=_string(plan_observed.get("completion_verdict")) or "complete",
                termination_reason=_string(plan_observed.get("termination_reason")) or "phase_graph_created",
            )
            main_observed = observed_phase_status.get(main_phase_id, {})
            phase_map[main_phase_id] = phase_run(
                phase_id=main_phase_id,
                phase_type=_string(phase_map[main_phase_id].get("phase_type")) or phase_type,
                status=_string(main_observed.get("status")) or status,
                input_refs=[],
                output_refs=output_refs,
                completion_verdict=_string(main_observed.get("completion_verdict")) or completion_verdict,
                termination_reason=_string(main_observed.get("termination_reason")) or termination_reason,
                failure_class=_string(raw.get("failure_class")) if not result_ok else "",
                failure_reason=(_string(raw.get("structured_output_error")) or _string(raw.get("artifact_failure_reason"))) if not result_ok else "",
            )
            if "deliver" in phase_map:
                delivery_status = "completed" if deliveries else ("skipped" if _string(getattr(task, "reply_delivery_mode", "")).lower() != "direct" else "failed")
                if deliveries and _string(deliveries[0].get("send_status")).lower() not in {"posted", "sent", "uploaded", "completed", "ok", "success", "shared"}:
                    delivery_status = "failed"
                deliver_observed = observed_phase_status.get("deliver", {})
                phase_map["deliver"] = phase_run(
                    phase_id="deliver",
                    phase_type="deliver",
                    status=_string(deliver_observed.get("status")) or delivery_status,
                    input_refs=output_refs,
                    output_refs=output_refs,
                    completion_verdict=_string(deliver_observed.get("completion_verdict")) or (completion_verdict if delivery_status == "completed" else ""),
                    termination_reason=_string(deliver_observed.get("termination_reason")) or ("normal_completion" if delivery_status == "completed" else delivery_status),
                    failure_class="reply_delivery_failed" if delivery_status == "failed" else "",
                    failure_reason="direct delivery was not acknowledged" if delivery_status == "failed" else "",
                )
            if "reflect" in phase_map and main_phase_id != "reflect":
                reflect_observed = observed_phase_status.get("reflect", {})
                phase_map["reflect"] = phase_run(
                    phase_id="reflect",
                    phase_type="reflect",
                    status=_string(reflect_observed.get("status")) or ("completed" if result_ok else "failed"),
                    completion_verdict=_string(reflect_observed.get("completion_verdict")) or completion_verdict,
                    termination_reason=_string(reflect_observed.get("termination_reason")) or termination_reason,
                )
            return [
                phase_map[_string(item.get("phase_id"))]
                for item in planned
                if _string(item.get("phase_id")) in phase_map
            ]
        phase_id = explicit_phase or "main"
        phase_observed = observed_phase_status.get(phase_id, {})
        return [
            phase_run(
                phase_id=phase_id,
                phase_type=phase_type,
                status=_string(phase_observed.get("status")) or status,
                input_refs=[],
                output_refs=output_refs,
                completion_verdict=_string(phase_observed.get("completion_verdict")) or completion_verdict,
                termination_reason=_string(phase_observed.get("termination_reason")) or termination_reason,
                failure_class=_string(raw.get("failure_class")) if not result_ok else "",
                failure_reason=_string(raw.get("structured_output_error")) or _string(raw.get("artifact_failure_reason")) if not result_ok else "",
            )
        ]

    def _ledger_events(
        self,
        raw: JsonObject,
        phase_runs: list[JsonObject],
        artifacts: list[JsonObject],
        deliveries: list[JsonObject],
        *,
        observer_events: list[JsonObject],
        execution_id: str,
    ) -> list[JsonObject]:
        events: list[JsonObject] = []
        sequence = 0
        for phase in phase_runs:
            phase_id = _string(phase.get("phase_id"))
            planned_payload = {
                key: value
                for key, value in phase.items()
                if key not in {"completion_verdict", "failure", "output_refs", "termination_reason"}
            }
            planned_payload["status"] = "planned"
            planned_payload["output_refs"] = []
            sequence += 1
            events.append(ledger_event("phase.planned", phase_id=phase_id, status="planned", payload=planned_payload, sequence=sequence))
        for item in observer_events:
            if _string(item.get("event_type")) == "phase.planned":
                continue
            sequence += 1
            events.append(self._ledger_event_from_observation(item, sequence=sequence, execution_id=execution_id))
        observed_phase_completions = {
            _string(item.get("phase_id"))
            for item in events
            if _string(item.get("kind")) == "phase.completed" and _string(item.get("phase_id"))
        }
        for phase in phase_runs:
            phase_id = _string(phase.get("phase_id"))
            if phase_id in observed_phase_completions:
                continue
            sequence += 1
            events.append(ledger_event("phase.completed", phase_id=phase_id, status=_string(phase.get("status")), payload=phase, sequence=sequence))
        for item in _json_object_list(raw.get("lifecycle_events")):
            kind = _string(item.get("event_type")) or _string(item.get("event")) or "model.lifecycle"
            if "memory" in kind.lower() or "honcho" in kind.lower():
                if "read" in kind.lower() or "recall" in kind.lower() or "search" in kind.lower():
                    mapped = "memory.read"
                elif "write" in kind.lower() or "store" in kind.lower() or "save" in kind.lower():
                    mapped = "memory.write"
                else:
                    mapped = "memory.lifecycle"
            elif "tool" in kind.lower():
                mapped = f"tool.{kind}"
            else:
                mapped = f"model.{kind}"
            sequence += 1
            events.append(ledger_event(mapped, phase_id=_string(item.get("phase")) or "main", status=_string(item.get("status")), payload=item, sequence=sequence))
        for item in _json_object_list(raw.get("artifact_tool_events")):
            kind = _string(item.get("event_type")) or "artifact.tool_event"
            mapped = kind if kind.startswith("artifact.") else f"artifact.{kind}"
            sequence += 1
            events.append(ledger_event(mapped, phase_id=_string(item.get("phase")) or "render", status=_string(item.get("status")), payload=item, sequence=sequence))
        for artifact in artifacts:
            sequence += 1
            events.append(ledger_event("artifact.created", phase_id="render", status="completed", payload=artifact, sequence=sequence))
            if _string(artifact.get("workspace_path")):
                sequence += 1
                events.append(ledger_event("file.written", phase_id="render", status="completed", payload=artifact, sequence=sequence))
        for delivery in deliveries:
            sequence += 1
            events.append(
                ledger_event(
                    "slack.message.sent",
                    phase_id="deliver",
                    status=_string(delivery.get("send_status")),
                    payload=delivery,
                    sequence=sequence,
                    idempotency_key=_string(delivery.get("body_sha1")) or _string(delivery.get("delivery_id")),
                )
            )
        failure_class = _string(raw.get("failure_class"))
        if failure_class:
            sequence += 1
            events.append(ledger_event("failure.runner", phase_id="main", status="failed", payload={"failure_class": failure_class, "diagnostics": _json_object(raw.get("runner_diagnostics"))}, sequence=sequence))
        return events

    def _ledger_event_from_observation(self, item: JsonObject, *, sequence: int, execution_id: str) -> JsonObject:
        event_type = _string(item.get("event_type")) or "model.lifecycle"
        phase_id = _string(item.get("phase")) or "main"
        kind = event_type
        lower = event_type.lower()
        if lower.startswith("direct_delivery."):
            kind = "slack." + event_type
        elif lower.startswith("artifact.pipeline."):
            kind = "phase." + event_type
        elif lower.startswith("artifact.file."):
            kind = "file." + event_type.removeprefix("artifact.file.")
        elif lower.startswith("artifact."):
            kind = event_type
        elif lower.startswith("phase."):
            kind = event_type
        elif lower.startswith("model."):
            kind = event_type
        elif "memory" in lower or "honcho" in lower:
            kind = "memory.lifecycle"
        elif lower.startswith("tool."):
            kind = event_type
        else:
            kind = "model." + event_type
        event_id_seed = "|".join([execution_id, str(item.get("seq") or sequence), kind, phase_id])
        raw_payload = item.get("payload")
        payload = dict(raw_payload) if isinstance(raw_payload, dict) else {}
        return ledger_event(
            kind,
            phase_id=phase_id,
            status=_string(item.get("status")),
            payload=payload,
            sequence=sequence,
            recorded_at=_string(item.get("recorded_at")),
            event_id=f"ledger-{hashlib.sha1(event_id_seed.encode('utf-8')).hexdigest()[:16]}",
        )
