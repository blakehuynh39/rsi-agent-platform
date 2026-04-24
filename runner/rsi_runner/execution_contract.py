from __future__ import annotations

from dataclasses import dataclass
import hashlib
from pathlib import Path
import time
from typing import Any

from .json_types import JsonObject


EXECUTION_CONTRACT_VERSION = "execution-envelope/v1"
RUNNER_PLANNER_MODE = "runner_first"

READ_CONTEXT = "read_context"
WORKSPACE_READ = "workspace_read"
WORKSPACE_WRITE = "workspace_write"
ARTIFACT_WRITE = "artifact_write"
SLACK_READ = "slack_read"
SLACK_SEND = "slack_send"
SLACK_UPLOAD = "slack_upload"
MEMORY_READ = "memory_read"
MEMORY_WRITE = "memory_write"
PLATFORM_MUTATION_REQUEST = "platform_mutation_request"

ALLOWED_CAPABILITIES = frozenset(
    {
        READ_CONTEXT,
        WORKSPACE_READ,
        WORKSPACE_WRITE,
        ARTIFACT_WRITE,
        SLACK_READ,
        SLACK_SEND,
        SLACK_UPLOAD,
        MEMORY_READ,
        MEMORY_WRITE,
        PLATFORM_MUTATION_REQUEST,
    }
)

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


def _bool(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        return value.strip().lower() in {"1", "true", "yes", "y", "on"}
    return bool(value)


def _now() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def normalize_capability_leases(value: Any) -> list[JsonObject]:
    out: list[JsonObject] = []
    for index, item in enumerate(_json_object_list(value)):
        capability = _string(item.get("capability"))
        lease_id = _string(item.get("lease_id")) or f"lease-{index + 1}-{capability}"
        scope = _json_object(item.get("scope"))
        constraints = _json_object(item.get("constraints"))
        out.append(
            {
                "lease_id": lease_id,
                "capability": capability,
                "scope": scope,
                "constraints": constraints,
                "granted": True if "granted" not in item else _bool(item.get("granted")),
            }
        )
    return out


def default_capability_leases(task: Any) -> list[JsonObject]:
    task_type = _string(getattr(task, "task_type", ""))
    reply_delivery_mode = _string(getattr(task, "reply_delivery_mode", "")).lower()
    requested_artifacts = _json_object_list(getattr(task, "requested_artifacts", []))
    execution_mode = _string(getattr(task, "execution_mode", "")).lower()
    capabilities = [READ_CONTEXT, MEMORY_READ, MEMORY_WRITE]
    if task_type in {"workflow", "prod", "proactive"}:
        capabilities.extend([WORKSPACE_READ, ARTIFACT_WRITE, SLACK_READ, PLATFORM_MUTATION_REQUEST])
        if reply_delivery_mode == "direct":
            capabilities.append(SLACK_SEND)
            capabilities.append(SLACK_UPLOAD)
        elif reply_delivery_mode == "mediated":
            capabilities.append(SLACK_UPLOAD)
    if requested_artifacts:
        capabilities.extend([WORKSPACE_READ, ARTIFACT_WRITE])
    if task_type in {"proposal", "repo-change"}:
        capabilities.extend([WORKSPACE_READ, PLATFORM_MUTATION_REQUEST])
        if execution_mode == "implement":
            capabilities.append(WORKSPACE_WRITE)
    seen: set[str] = set()
    out: list[JsonObject] = []
    for capability in capabilities:
        if capability in seen:
            continue
        seen.add(capability)
        out.append({"lease_id": f"default-{capability}", "capability": capability, "scope": {}, "constraints": {}, "granted": True})
    return out


def granted_capabilities(leases: list[JsonObject]) -> set[str]:
    return {_string(item.get("capability")) for item in leases if _bool(item.get("granted")) and _string(item.get("capability"))}


@dataclass
class ContractValidationResult:
    ok: bool
    errors: list[str]


def validate_capability_leases(leases: list[JsonObject], *, require_present: bool) -> ContractValidationResult:
    errors: list[str] = []
    if require_present and not leases:
        errors.append("CapabilityLease v1 is required for execution-envelope/v1 tasks.")
    for item in leases:
        capability = _string(item.get("capability"))
        if not capability:
            errors.append("CapabilityLease capability is required.")
        elif capability not in ALLOWED_CAPABILITIES:
            errors.append(f"Unknown capability lease {capability}.")
    return ContractValidationResult(ok=not errors, errors=errors)


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


def required_capabilities_for_phase(phase_type: str, task: Any) -> list[str]:
    task_type = _string(getattr(task, "task_type", "")).lower()
    reply_delivery_mode = _string(getattr(task, "reply_delivery_mode", "")).lower()
    if phase_type == "plan":
        return [READ_CONTEXT, MEMORY_READ]
    if phase_type == "investigate":
        caps = [READ_CONTEXT, MEMORY_READ]
        if task_type in {"workflow", "prod", "proactive"}:
            caps.append(SLACK_READ)
            caps.append(WORKSPACE_READ)
        return caps
    if phase_type == "render":
        return [WORKSPACE_READ, ARTIFACT_WRITE]
    if phase_type == "deliver":
        caps = [READ_CONTEXT]
        if reply_delivery_mode == "direct":
            caps.append(SLACK_SEND)
        if _json_object_list(getattr(task, "requested_artifacts", [])) or _json_object_list(getattr(task, "produced_artifacts", [])):
            caps.append(SLACK_UPLOAD)
        return caps
    if phase_type == "reflect":
        return [READ_CONTEXT, MEMORY_READ, MEMORY_WRITE]
    caps = [READ_CONTEXT, MEMORY_READ, MEMORY_WRITE]
    execution_mode = _string(getattr(task, "execution_mode", "")).lower()
    if task_type in {"workflow", "prod", "proactive"}:
        caps.extend([SLACK_READ, WORKSPACE_READ, ARTIFACT_WRITE, PLATFORM_MUTATION_REQUEST])
        if reply_delivery_mode == "direct":
            caps.append(SLACK_SEND)
    if task_type in {"proposal", "repo-change"}:
        caps.extend([WORKSPACE_READ, PLATFORM_MUTATION_REQUEST])
        if execution_mode == "implement":
            caps.append(WORKSPACE_WRITE)
    return list(dict.fromkeys(caps))


def validate_phase_leases(phase_runs: list[JsonObject], leases: list[JsonObject]) -> ContractValidationResult:
    errors: list[str] = []
    granted = granted_capabilities(leases)
    for phase in phase_runs:
        phase_id = _string(phase.get("phase_id")) or "unknown"
        for capability in _string_list(phase.get("required_leases")):
            if capability not in granted:
                errors.append(f"Phase {phase_id} requires unleased capability {capability}.")
    return ContractValidationResult(ok=not errors, errors=errors)


def default_delivery_policy(task: Any) -> JsonObject:
    mode = _string(getattr(task, "reply_delivery_mode", "")).lower()
    channel_id = _string(getattr(task, "channel_id", ""))
    thread_ts = _string(getattr(task, "thread_ts", ""))
    trace_id = _string(getattr(task, "trace_id", ""))
    return {
        "bound_channel_id": channel_id,
        "bound_thread_ts": thread_ts,
        "direct_send_allowed": mode == "direct",
        "upload_allowed": mode in {"direct", "mediated"},
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
    required_leases: list[str],
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
        "required_leases": list(dict.fromkeys(required_leases)),
        "input_refs": list(input_refs or []),
        "output_refs": list(output_refs or []),
        "completion_verdict": completion_verdict,
        "termination_reason": termination_reason,
    }
    if failure_class or failure_reason:
        item["failure"] = {"class": failure_class, "reason": failure_reason}
    return item


def ledger_event(kind: str, *, phase_id: str = "", status: str = "", payload: JsonObject | None = None) -> JsonObject:
    return {
        "event_id": f"ledger-{hashlib.sha1((kind + phase_id + str(time.time_ns())).encode('utf-8')).hexdigest()[:16]}",
        "kind": kind,
        "phase_id": phase_id,
        "status": status,
        "recorded_at": _now(),
        "payload": payload or {},
    }


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

    def task_leases(self, task: Any) -> list[JsonObject]:
        explicit = normalize_capability_leases(getattr(task, "capability_leases", []))
        return explicit or default_capability_leases(task)

    def validate_task(self, task: Any) -> ContractValidationResult:
        explicit_contract = _string(getattr(task, "contract_version", "")) == EXECUTION_CONTRACT_VERSION
        leases = normalize_capability_leases(getattr(task, "capability_leases", []))
        validation = validate_capability_leases(leases, require_present=explicit_contract)
        if not validation.ok:
            return validation
        effective_leases = leases or default_capability_leases(task)
        phase_type = phase_type_for_task(task)
        phases = [
            phase_run(
                phase_id=_string(getattr(task, "execution_phase", "")) or "main",
                phase_type=phase_type,
                status="planned",
                required_leases=required_capabilities_for_phase(phase_type, task),
            )
        ]
        phase_validation = validate_phase_leases(phases, effective_leases)
        if not phase_validation.ok:
            return phase_validation
        return ContractValidationResult(ok=True, errors=[])

    def attach_envelope(self, task: Any, result: Any, *, observer: Any | None = None) -> Any:
        raw = _json_object(getattr(result, "raw", {}))
        if _json_object(raw.get("execution_envelope")):
            return result
        execution_id = _string(getattr(observer, "execution_id", "")) or _string(getattr(task, "execution_id", ""))
        if not execution_id:
            execution_id = _string(raw.get("execution_id")) or _string(raw.get("observation_execution_id")) or f"hexec-{time.time_ns()}"
        structured_output = _json_object(raw.get("structured_output"))
        artifacts = self._artifact_records(task, raw, structured_output, execution_id=execution_id)
        deliveries = self._delivery_records(raw, structured_output, execution_id=execution_id)
        phase_runs = self._phase_runs(task, raw, result_ok=bool(getattr(result, "ok", False)), artifacts=artifacts, deliveries=deliveries)
        leases = self.task_leases(task)
        phase_validation = validate_phase_leases(phase_runs, leases)
        if not phase_validation.ok:
            raw.setdefault("failure_class", "runner_contract_failed")
            raw["runner_diagnostics"] = {
                **_json_object(raw.get("runner_diagnostics")),
                "failure_kind": "runner_contract_failed",
                "contract_errors": list(phase_validation.errors),
            }
        ledger_events = self._ledger_events(raw, phase_runs, artifacts, deliveries)
        memory_events = [event for event in ledger_events if _string(event.get("kind")).startswith("memory.")]
        final_response = _string(structured_output.get("final_answer")) or _string(structured_output.get("reply_draft")) or _string(getattr(result, "message", ""))
        completion_verdict = _string(raw.get("completion_verdict")) or _string(structured_output.get("completion_verdict")) or ("failed" if not getattr(result, "ok", False) else "complete")
        termination_reason = _string(raw.get("termination_reason")) or _string(structured_output.get("termination_reason")) or ("failure" if not getattr(result, "ok", False) else "normal_completion")
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
            "capability_leases": leases,
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
                        "required_leases": _string_list(phase.get("required_leases")),
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
                "ok": bool(getattr(result, "ok", False)) and phase_validation.ok,
            },
            "final_response": final_response,
        }
        if not phase_validation.ok:
            envelope["completion"]["ok"] = False
            envelope["completion"]["termination_reason"] = "runner_contract_failed"
            envelope["completion"]["completion_verdict"] = "failed"
            envelope["ledger_events"].append(
                ledger_event("failure.contract", phase_id="main", status="failed", payload={"errors": list(phase_validation.errors)})
            )
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

    def _artifact_records(self, task: Any, raw: JsonObject, structured_output: JsonObject, *, execution_id: str) -> list[JsonObject]:
        items: list[JsonObject] = []
        items.extend(_json_object_list(structured_output.get("produced_artifacts")))
        items.extend(_json_object_list(raw.get("produced_artifacts")))
        seen_refs: set[str] = set()
        for path in _string_list(raw.get("native_artifact_paths")):
            items.append({"kind": "artifact", "workspace_path": path, "file_ref": f"file://{path}", "artifact_refs": [f"file://{path}"]})
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

    def _phase_runs(self, task: Any, raw: JsonObject, *, result_ok: bool, artifacts: list[JsonObject], deliveries: list[JsonObject]) -> list[JsonObject]:
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
        diagnostics = _json_object(raw.get("runner_diagnostics"))
        if _bool(diagnostics.get("artifact_phase_enabled")) and not explicit_phase:
            render_status = "completed" if artifacts else ("failed" if _string(raw.get("artifact_failure_reason")) else "skipped")
            deliver_status = "completed" if deliveries else ("skipped" if _string(getattr(task, "reply_delivery_mode", "")).lower() != "direct" else "failed")
            deliver_completion_verdict = ""
            if deliver_status == "completed":
                deliver_completion_verdict = _string(diagnostics.get("artifact_delivery_completion_verdict")) or completion_verdict
            return [
                phase_run(
                    phase_id="investigate",
                    phase_type="investigate",
                    status="completed" if result_ok else "failed",
                    required_leases=required_capabilities_for_phase("investigate", task),
                    completion_verdict=_string(diagnostics.get("artifact_investigate_completion_verdict")) or completion_verdict,
                    termination_reason=_string(diagnostics.get("artifact_investigate_termination_reason")) or termination_reason,
                ),
                phase_run(
                    phase_id="render",
                    phase_type="render",
                    status=render_status,
                    required_leases=required_capabilities_for_phase("render", task),
                    output_refs=[ref for artifact in artifacts for ref in _string_list(artifact.get("artifact_refs"))],
                    completion_verdict=completion_verdict if render_status == "completed" else "",
                    termination_reason=_string(raw.get("artifact_failure_reason")) if render_status != "completed" else "normal_completion",
                    failure_class="artifact_render_failed" if render_status == "failed" else "",
                    failure_reason=_string(raw.get("artifact_failure_reason")) if render_status == "failed" else "",
                ),
                phase_run(
                    phase_id="deliver",
                    phase_type="deliver",
                    status=deliver_status,
                    required_leases=required_capabilities_for_phase("deliver", task),
                    input_refs=[ref for artifact in artifacts for ref in _string_list(artifact.get("artifact_refs"))],
                    output_refs=output_refs,
                    completion_verdict=deliver_completion_verdict,
                    termination_reason=_string(diagnostics.get("direct_delivery_phase_failed")) if deliver_status == "failed" else ("normal_completion" if deliver_status == "completed" else "skipped"),
                    failure_class="artifact_delivery_failed" if deliver_status == "failed" else "",
                    failure_reason=_string(diagnostics.get("direct_delivery_phase_failed")) if deliver_status == "failed" else "",
                ),
            ]
        phase_id = explicit_phase or "main"
        return [
            phase_run(
                phase_id=phase_id,
                phase_type=phase_type,
                status=status,
                required_leases=required_capabilities_for_phase(phase_type, task),
                input_refs=[],
                output_refs=output_refs,
                completion_verdict=completion_verdict,
                termination_reason=termination_reason,
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
    ) -> list[JsonObject]:
        events: list[JsonObject] = []
        for phase in phase_runs:
            phase_id = _string(phase.get("phase_id"))
            events.append(ledger_event("phase.started", phase_id=phase_id, status="running", payload={"phase_type": phase.get("phase_type")}))
            events.append(ledger_event("phase.completed", phase_id=phase_id, status=_string(phase.get("status")), payload=phase))
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
            events.append(ledger_event(mapped, phase_id=_string(item.get("phase")) or "main", status=_string(item.get("status")), payload=item))
        for item in _json_object_list(raw.get("artifact_tool_events")):
            kind = _string(item.get("event_type")) or "artifact.tool_event"
            mapped = kind if kind.startswith("artifact.") else f"artifact.{kind}"
            events.append(ledger_event(mapped, phase_id=_string(item.get("phase")) or "render", status=_string(item.get("status")), payload=item))
        for artifact in artifacts:
            events.append(ledger_event("artifact.created", phase_id="render", status="completed", payload=artifact))
            if _string(artifact.get("workspace_path")):
                events.append(ledger_event("file.written", phase_id="render", status="completed", payload=artifact))
        for delivery in deliveries:
            events.append(ledger_event("slack.message.sent", phase_id="deliver", status=_string(delivery.get("send_status")), payload=delivery))
        failure_class = _string(raw.get("failure_class"))
        if failure_class:
            events.append(ledger_event("failure.runner", phase_id="main", status="failed", payload={"failure_class": failure_class, "diagnostics": _json_object(raw.get("runner_diagnostics"))}))
        return events
