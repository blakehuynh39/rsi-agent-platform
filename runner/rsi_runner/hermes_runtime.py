from __future__ import annotations

import concurrent.futures
from dataclasses import dataclass, replace
import hashlib
import json
import logging
import os
from pathlib import Path
import re
import socket
import subprocess
import sys
import threading
import time
import uuid
from numbers import Number
from typing import Any
from urllib import error as urlerror
from urllib import parse as urlparse
from urllib import request as urlrequest

from .json_types import JsonObject, JsonToolWrapperSchema, JsonValue

from .config import RunnerConfig
from .hermes_adapter import HermesAdapter
from .hermes_agent_adapter import validate_hermes_contract
from .hermes_mcp_adapter import HermesTaskScopedMCPAdapter, TaskScopedMCPRegistration
from .execution_contract import (
    ALLOWED_CAPABILITIES,
    EXECUTION_CONTRACT_VERSION,
    RUNNER_PLANNER_MODE,
    HermesCompanyComputer,
)
from .observability import ObservationEmitter
from .rsi_tools import (
    BLOCKED_HONCHO_TOOLS,
    CompositeToolProvider,
    HERMES_ARTIFACT_TOOLSET,
    HERMES_GOVERNED_READONLY_TOOLSET,
    HERMES_GOVERNED_WORKSPACE_TOOLSET,
    IMPLEMENT_RSI_TOOL_NAMES,
    READ_ONLY_HONCHO_TOOLS,
    READ_ONLY_RSI_TOOL_NAMES,
    READ_ONLY_WORKSPACE_RSI_TOOL_NAMES,
    ReadOnlyToolBinding,
    WORKSPACE_RSI_TOOL_NAMES,
    normalize_tool_names,
    tool_transport_name,
    tool_schema_wrappers,
)
from .session_manager import MemoryTracker, SessionContext, SessionManager, stable_session_id

ROLE_TASK_TYPES = {
    "prod": {"general", "workflow", "prod", "question_gather", "question_reduce", "question_expand"},
    "proactive": {"general", "workflow", "proactive", "question_gather", "question_reduce", "question_expand"},
    "eval": {"general", "eval"},
    "proposal": {"general", "proposal", "repo-change"},
}

logger = logging.getLogger(__name__)

NATIVE_HERMES_DIAGNOSE_TOOLS = frozenset({"todo", "session_search"})
QUESTION_RUN_BOUNDED_STOP_TERMINATION_REASONS = frozenset(
    {
        "task_timeout",
        "iteration_budget_exhausted",
        "output_token_budget_exhausted",
    }
)
ARTIFACT_RENDER_NATIVE_TOOL_NAMES = frozenset({"write_file", "read_file", "search_files", "skill_view"})
ARTIFACT_RENDER_RSI_TOOL_NAMES = frozenset(
    {
        "workspace.list_files",
        "workspace.read_file",
        "workspace.search",
        "workspace.write_file",
    }
)
_NATIVE_EXECUTOR_RESULT_MARKER = "RSI_EXECUTOR_RESULT::"
_NATIVE_EXECUTOR_OUTPUT_CHUNK_CHARS = 8 * 1024
_SENSITIVE_ENV_KEY_FRAGMENTS = (
    "authorization",
    "api_key",
    "apikey",
    "token",
    "secret",
    "private_key",
    "password",
)
_SENSITIVE_OUTPUT_PATTERNS = (
    (
        re.compile(r"(?i)\b(bearer)\s+([A-Za-z0-9._~+/=-]{8,})"),
        lambda match: f"{match.group(1)} [redacted]",
    ),
    (
        re.compile(r"\b(xox[baprs]-[A-Za-z0-9-]{8,})\b"),
        lambda _match: "[redacted-slack-token]",
    ),
    (
        re.compile(r"\b(sk-[A-Za-z0-9_-]{8,})\b"),
        lambda _match: "[redacted-openai-key]",
    ),
)
_BENIGN_MCP_TOOLSET_WARNING = re.compile(
    r"Warning: Unknown toolsets:\s*\n(?:mcp-[^\n]*(?:\n|$))+\n*",
    re.MULTILINE,
)


def _json_object_or_empty(value: JsonValue | None) -> JsonObject:
    if isinstance(value, dict):
        return value
    return {}


def _json_object_list(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    return [item for item in value if isinstance(item, dict)]


def _string_list(value: JsonValue | None) -> list[str]:
    if not isinstance(value, list):
        return []
    return [str(item) for item in value]


def _string_list_or_empty(value: JsonValue | None) -> list[str]:
    if isinstance(value, list):
        out: list[str] = []
        for item in value:
            text = _string_or_json(item)
            if text:
                out.append(text)
        return out
    text = _string_or_json(value)
    if not text:
        return []
    return [text]


def _string_or_json(value: JsonValue | None) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value.strip()
    if isinstance(value, list):
        if all(not isinstance(item, (dict, list)) for item in value):
            return "\n".join(part for part in (str(item).strip() for item in value) if part)
        return json.dumps(value, ensure_ascii=True, sort_keys=True)
    if isinstance(value, dict):
        return json.dumps(value, ensure_ascii=True, sort_keys=True)
    return str(value).strip()


def _truncate_string(value: Any, limit: int) -> str:
    text = _string_or_json(value)
    if len(text) <= limit:
        return text
    return text[: max(0, limit - 1)] + "…"


def _is_sensitive_env_key(name: str) -> bool:
    lower = str(name or "").strip().lower()
    if not lower:
        return False
    return any(fragment in lower for fragment in _SENSITIVE_ENV_KEY_FRAGMENTS)


def _sensitive_env_values(env: dict[str, str] | None = None) -> list[str]:
    source = env or os.environ
    values: list[str] = []
    for key, value in source.items():
        if not _is_sensitive_env_key(key):
            continue
        secret = str(value or "").strip()
        if len(secret) < 8:
            continue
        values.append(secret)
    return sorted(set(values), key=len, reverse=True)


def _redact_subprocess_output(text: str, *, secret_values: list[str]) -> str:
    redacted = str(text or "")
    for pattern, replacement in _SENSITIVE_OUTPUT_PATTERNS:
        redacted = pattern.sub(replacement, redacted)
    for secret in secret_values:
        redacted = redacted.replace(secret, "[redacted]")
    return redacted


def _suppress_benign_subprocess_output(text: str) -> str:
    return _BENIGN_MCP_TOOLSET_WARNING.sub("", str(text or ""))


def _float_or_zero(value: JsonValue | None) -> float:
    if isinstance(value, Number):
        return float(value)
    try:
        return float(str(value).strip())
    except (AttributeError, TypeError, ValueError):
        return 0.0


def _bool_or_false(value: JsonValue | None) -> bool:
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        text = value.strip().lower()
        if text in {"1", "true", "t", "yes", "y", "on"}:
            return True
        if text in {"0", "false", "f", "no", "n", "off"}:
            return False
    return False


def _normalize_evidence_refs(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    out: list[JsonObject] = []
    for item in value:
        if isinstance(item, dict):
            ref = _string_or_json(item.get("ref"))
            if not ref:
                continue
            normalized: JsonObject = {
                "kind": _string_or_json(item.get("kind")) or "reference",
                "ref": ref,
            }
            summary = _string_or_json(item.get("summary"))
            if summary:
                normalized["summary"] = summary
            out.append(normalized)
            continue
        ref = _string_or_json(item)
        if not ref:
            continue
        out.append({"kind": "reference", "ref": ref})
    return out


def _normalize_visible_reasoning(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    out: list[JsonObject] = []
    for item in value:
        if isinstance(item, dict):
            summary = _string_or_json(item.get("summary")) or json.dumps(item, ensure_ascii=True, sort_keys=True)
            out.append(
                {
                    "step_type": _string_or_json(item.get("step_type")) or "analysis",
                    "summary": summary,
                    "evidence_refs": _normalize_evidence_refs(item.get("evidence_refs")),
                    "alternatives": _string_list_or_empty(item.get("alternatives")),
                    "confidence": _float_or_zero(item.get("confidence")),
                    "decision": _string_or_json(item.get("decision")),
                }
            )
            continue
        summary = _string_or_json(item)
        if not summary:
            continue
        out.append(
            {
                "step_type": "analysis",
                "summary": summary,
                "evidence_refs": [],
                "alternatives": [],
                "confidence": 0.0,
                "decision": "",
            }
        )
    return out


def _normalize_proposed_actions(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    out: list[JsonObject] = []
    for item in value:
        if not isinstance(item, dict):
            continue
        request_payload = item.get("request_payload")
        if not isinstance(request_payload, dict):
            request_payload = {}
        out.append(
            {
                "kind": _string_or_json(item.get("kind")),
                "target_ref": _string_or_json(item.get("target_ref")),
                "request_payload": request_payload,
                "approval_mode": _string_or_json(item.get("approval_mode")),
                "idempotency_key": _string_or_json(item.get("idempotency_key")),
                "rationale": _string_or_json(item.get("rationale")),
                "evidence_refs": _normalize_evidence_refs(item.get("evidence_refs")),
            }
        )
    return out


def _normalize_knowledge_drafts(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    out: list[JsonObject] = []
    for item in value:
        if not isinstance(item, dict):
            continue
        out.append(
            {
                "kind": _string_or_json(item.get("kind")),
                "scope_type": _string_or_json(item.get("scope_type")),
                "scope_id": _string_or_json(item.get("scope_id")),
                "title": _string_or_json(item.get("title")),
                "summary": _string_or_json(item.get("summary")),
                "body": _string_or_json(item.get("body")),
                "confidence": _float_or_zero(item.get("confidence")),
                "fresh_until": _string_or_json(item.get("fresh_until")),
                "evidence_refs": _normalize_evidence_refs(item.get("evidence_refs")),
            }
        )
    return out


def _normalize_outcome_hypotheses(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    out: list[JsonObject] = []
    for item in value:
        if not isinstance(item, dict):
            continue
        out.append(
            {
                "outcome_type": _string_or_json(item.get("outcome_type")),
                "success_condition": _string_or_json(item.get("success_condition")),
                "measurement_ref": _string_or_json(item.get("measurement_ref")),
                "expected_time_horizon": _string_or_json(item.get("expected_time_horizon")),
            }
        )
    return out


def _normalize_retry_assessment(value: JsonValue | None) -> JsonObject:
    if not isinstance(value, dict):
        return {}
    return {
        "failure_class": _string_or_json(value.get("failure_class")),
        "failure_summary": _string_or_json(value.get("failure_summary")),
        "retry_decision": _string_or_json(value.get("retry_decision")),
        "material_hypothesis_change": _bool_or_false(value.get("material_hypothesis_change")),
        "changed_files": _string_list_or_empty(value.get("changed_files")),
    }


def _normalize_requested_artifacts(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    out: list[JsonObject] = []
    for item in value:
        if not isinstance(item, dict):
            continue
        kind = _string_or_json(item.get("kind"))
        if not kind:
            continue
        out.append(
            {
                "kind": kind,
                "description": _string_or_json(item.get("description")),
            }
        )
    return out


def _normalize_requested_skills(value: JsonValue | None) -> list[str]:
    if not isinstance(value, list):
        return []
    out: list[str] = []
    seen: set[str] = set()
    for item in value:
        if not isinstance(item, str):
            continue
        normalized = item.strip().replace("_", "-").lower()
        normalized = normalized.lstrip("/")
        if not normalized or not re.fullmatch(r"[a-z0-9][a-z0-9-]*", normalized):
            continue
        if normalized in seen:
            continue
        seen.add(normalized)
        out.append(normalized)
    return out


def _normalize_produced_artifacts(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    out: list[JsonObject] = []
    for item in value:
        if not isinstance(item, dict):
            continue
        kind = _string_or_json(item.get("kind"))
        refs = _string_list_or_empty(item.get("artifact_refs"))
        workspace_path = _string_or_json(item.get("workspace_path"))
        file_ref = _string_or_json(item.get("file_ref"))
        if not refs and file_ref:
            refs = [file_ref]
        if not refs and workspace_path:
            refs = [f"file://{workspace_path}"]
        if not file_ref and refs:
            file_ref = refs[0]
        if not kind and not refs and not workspace_path:
            continue
        normalized: JsonObject = {"kind": kind, "artifact_refs": refs}
        title = _string_or_json(item.get("title"))
        if title:
            normalized["title"] = title
        delivery_status = _string_or_json(item.get("delivery_status"))
        if delivery_status:
            normalized["delivery_status"] = delivery_status
        failure_reason = _string_or_json(item.get("failure_reason"))
        if failure_reason:
            normalized["failure_reason"] = failure_reason
        if workspace_path:
            normalized["workspace_path"] = workspace_path
        if file_ref:
            normalized["file_ref"] = file_ref
        size_bytes = item.get("size_bytes")
        if isinstance(size_bytes, Number) and not isinstance(size_bytes, bool) and size_bytes >= 0:
            normalized["size_bytes"] = int(size_bytes)
        sha256 = _string_or_json(item.get("sha256"))
        if sha256:
            normalized["sha256"] = sha256
        created_by_execution_id = _string_or_json(item.get("created_by_execution_id"))
        if created_by_execution_id:
            normalized["created_by_execution_id"] = created_by_execution_id
        share_status = _string_or_json(item.get("share_status"))
        if share_status:
            normalized["share_status"] = share_status
        out.append(normalized)
    return out


def _normalize_artifact_render_briefs(value: JsonValue | None) -> list[JsonObject]:
    if not isinstance(value, list):
        return []
    out: list[JsonObject] = []
    for item in value:
        if not isinstance(item, dict):
            continue
        kind = _string_or_json(item.get("kind"))
        if not kind:
            continue
        requested_skill = _string_or_json(_first_non_none(item.get("requested_skill"), item.get("skill")))
        normalized_skill = _normalize_skill_identifier(requested_skill)
        out.append(
            {
                "kind": kind,
                "requested_skill": requested_skill,
                "skill": normalized_skill,
                "title": _string_or_json(item.get("title")),
                "render_prompt": _string_or_json(item.get("render_prompt")),
                "inputs": _json_object_or_empty(item.get("inputs")),
                "output_path_hint": _string_or_json(item.get("output_path_hint")),
            }
        )
    return out


def _normalize_structured_output(payload: JsonObject) -> JsonObject:
    normalized = dict(payload)
    normalized["context_summary"] = _string_or_json(payload.get("context_summary"))
    normalized["reply_draft"] = _string_or_json(payload.get("reply_draft"))
    normalized["final_answer"] = _string_or_json(payload.get("final_answer"))
    normalized["confidence"] = _float_or_zero(payload.get("confidence"))
    normalized["self_critique"] = _string_or_json(payload.get("self_critique"))
    normalized["visible_reasoning"] = _normalize_visible_reasoning(payload.get("visible_reasoning"))
    normalized["proposed_actions"] = _normalize_proposed_actions(payload.get("proposed_actions"))
    reply_delivery = _json_object_or_empty(payload.get("reply_delivery"))
    if reply_delivery:
        normalized["reply_delivery"] = reply_delivery
    else:
        normalized.pop("reply_delivery", None)
    normalized["knowledge_drafts"] = _normalize_knowledge_drafts(payload.get("knowledge_drafts"))
    normalized["outcome_hypotheses"] = _normalize_outcome_hypotheses(payload.get("outcome_hypotheses"))
    normalized["artifact_render_briefs"] = _normalize_artifact_render_briefs(payload.get("artifact_render_briefs"))
    normalized["produced_artifacts"] = _normalize_produced_artifacts(payload.get("produced_artifacts"))
    normalized["artifact_failure_reason"] = _string_or_json(payload.get("artifact_failure_reason"))
    normalized["change_plan"] = _string_or_json(payload.get("change_plan"))
    normalized["repo_patch"] = _string_or_json(payload.get("repo_patch"))
    normalized["validation_plan"] = _string_or_json(payload.get("validation_plan"))
    normalized["retry_assessment"] = _normalize_retry_assessment(payload.get("retry_assessment"))
    normalized["hypothesis_delta"] = _string_or_json(payload.get("hypothesis_delta"))
    return normalized


def _path_from_file_ref(value: str) -> str:
    text = str(value or "").strip()
    if not text:
        return ""
    if "://" not in text:
        return text
    parsed = urlparse.urlparse(text)
    if parsed.scheme not in {"file", "hermes-file"}:
        return ""
    return urlparse.unquote(parsed.path or "")


def _native_artifact_paths_from_events(value: JsonValue | None) -> list[str]:
    out: list[str] = []
    seen: set[str] = set()
    for item in _json_object_list(value):
        if _string_or_json(item.get("event_type")) != "artifact.write.completed":
            continue
        payload = _json_object_or_empty(item.get("payload"))
        path = _string_or_json(payload.get("path"))
        if not path or path in seen:
            continue
        seen.add(path)
        out.append(path)
    return out


def _native_artifact_event_details(value: JsonValue | None) -> JsonObject:
    details: JsonObject = {
        "artifact_output_dir": "",
        "artifact_persistence_tool": "",
        "requested_output_path": "",
        "saved_paths": [],
        "save_error": "",
    }
    saved_paths = _native_artifact_paths_from_events(value)
    if saved_paths:
        details["saved_paths"] = saved_paths
    for item in _json_object_list(value):
        payload = _json_object_or_empty(item.get("payload"))
        event_type = _string_or_json(item.get("event_type"))
        tool_name = _string_or_json(payload.get("tool_name"))
        artifact_output_dir = _string_or_json(payload.get("artifact_output_dir"))
        if artifact_output_dir:
            details["artifact_output_dir"] = artifact_output_dir
        if tool_name:
            details["artifact_persistence_tool"] = tool_name
        requested_path = _string_or_json(payload.get("requested_path"))
        if requested_path:
            details["requested_output_path"] = requested_path
        if event_type == "artifact.write.failed":
            save_error = _string_or_json(payload.get("error"))
            if save_error:
                details["save_error"] = save_error
    return details


def _optional_string(value: JsonValue | None) -> str | None:
    if value is None:
        return None
    if isinstance(value, str):
        text = value.strip()
        return text or None
    return str(value)


def _normalize_skill_identifier(value: JsonValue | None) -> str:
    normalized = _normalize_requested_skills([_string_or_json(value)])
    if normalized:
        return normalized[0]
    return ""


def _required_string(value: JsonValue | None, default: str) -> str:
    if value is None:
        return default
    if isinstance(value, str):
        text = value.strip()
        return text or default
    return str(value)


def _first_non_none(*values: JsonValue | None) -> JsonValue | None:
    for value in values:
        if value is not None:
            return value
    return None


def _transport_tool_policy(custom_tools: list[str], memory_tools: list[str]) -> tuple[list[str], dict[str, str], list[str]]:
    transport_effective = list(memory_tools)
    custom_tool_transport_map: dict[str, str] = {}
    invalid_tool_names: list[str] = []
    seen_transport: dict[str, str] = {}
    for name in custom_tools:
        try:
            transport = tool_transport_name(name)
        except ValueError:
            invalid_tool_names.append(name)
            continue
        existing = seen_transport.get(transport)
        if existing is not None and existing != name:
            invalid_tool_names.extend([existing, name])
            continue
        seen_transport[transport] = name
        custom_tool_transport_map[name] = transport
        transport_effective.append(transport)
    return normalize_tool_names(transport_effective), custom_tool_transport_map, normalize_tool_names(invalid_tool_names)


def _transport_name_or_self(name: str) -> str:
    try:
        return tool_transport_name(name)
    except ValueError:
        return str(name or "").strip()


try:
    from run_agent import AIAgent  # type: ignore
    from hermes_constants import parse_reasoning_effort  # type: ignore
    from agent.skill_commands import (  # type: ignore
        build_preloaded_skills_prompt,
        build_skill_invocation_message,
        resolve_skill_command_key,
    )
except (ImportError, ModuleNotFoundError):  # pragma: no cover - import depends on external Hermes install
    AIAgent = None
    build_preloaded_skills_prompt = None
    build_skill_invocation_message = None
    resolve_skill_command_key = None

    def parse_reasoning_effort(effort: str) -> JsonObject | None:
        level = (effort or "").strip().lower()
        if not level:
            return None
        if level == "none":
            return {"enabled": False}
        if level in {"minimal", "low", "medium", "high", "xhigh"}:
            return {"enabled": True, "effort": level}
        return None


@dataclass
class HermesExecutionResult:
    ok: bool
    message: str
    provider: str
    raw: JsonObject


@dataclass
class PartialReducerAttemptResult:
    ok: bool
    response_text: str
    structured_output: JsonObject
    error: str
    provider_response_id: str


@dataclass
class NativeExecutionRecorder:
    path: Path

    def record(self, event: str, payload: JsonObject | None = None) -> None:
        item = {
            "event": str(event or "").strip(),
            "recorded_at_unix": time.time(),
            **(payload or {}),
        }
        try:
            self.path.parent.mkdir(parents=True, exist_ok=True)
            with self.path.open("a", encoding="utf-8") as handle:
                handle.write(json.dumps(item, ensure_ascii=True, sort_keys=True) + "\n")
        except Exception as exc:
            logger.warning("native execution log append failed path=%s event=%s error=%s", self.path, event, exc)


@dataclass
class RunnerTaskRequest:
    task_type: str
    repo: str
    repo_ref: str | None
    prompt: str
    system_message: str | None
    requested_skills: list[str]
    allowed_tools: list[str]
    allowed_commands: list[str]
    timeout_seconds: int
    expected_outputs: list[str]
    artifact_destination: str | None
    requested_artifacts: list[JsonObject]
    artifact_optional: bool
    context_summary: str | None
    rejected_proposal_context: list[JsonObject]
    execution_mode: str | None
    intent: str | None
    trace_id: str | None
    workflow_id: str | None
    operation_id: str | None
    execution_id: str | None
    conversation_id: str | None
    case_id: str | None
    channel_id: str | None
    thread_ts: str | None
    trigger_event_id: str | None
    recent_conversation_entries: list[JsonObject]
    case_summary: JsonObject | None
    prior_trace_refs: list[JsonObject]
    repo_allowlist: list[str]
    tool_allowlist: list[str]
    response_mode: str | None
    reply_delivery_mode: str | None
    context_refs: list[JsonObject]
    approval_mode: str | None
    reasoning_verbosity: str | None
    session_scope_kind: str | None
    session_scope_id: str | None
    parent_session_scope_kind: str | None
    parent_session_scope_id: str | None
    harness_profile_id: str | None
    harness_overlay_version: str | None
    memory_backend: str | None
    assistant_peer_id: str | None
    user_peer_id: str | None
    attempt_id: str | None
    workspace_id: str | None
    workspace_repo: str | None
    workspace_branch: str | None
    allowed_path_globs: list[str]
    mcp_servers: list[JsonObject]
    execution_phase: str | None
    contract_version: str | None
    execution_intent: JsonObject
    capability_leases: list[JsonObject]
    delivery_policy: JsonObject
    workspace_policy: JsonObject
    approval_policy: JsonObject

    @classmethod
    def from_payload(cls, payload: JsonObject) -> "RunnerTaskRequest":
        task = _json_object_or_empty(payload.get("task")) or payload
        return cls(
            task_type=_required_string(task.get("task_type"), "general"),
            repo=_required_string(task.get("repo"), "rsi-agent-platform"),
            repo_ref=_optional_string(task.get("repo_ref")),
            prompt=_required_string(_first_non_none(task.get("prompt"), payload.get("prompt")), ""),
            system_message=_optional_string(_first_non_none(task.get("system_message"), payload.get("system_message"))),
            requested_skills=_normalize_requested_skills(task.get("requested_skills")),
            allowed_tools=_string_list(task.get("allowed_tools")),
            allowed_commands=_string_list(task.get("allowed_commands")),
            timeout_seconds=int(task.get("timeout_seconds", 900)),
            expected_outputs=_string_list(task.get("expected_outputs")),
            artifact_destination=_optional_string(task.get("artifact_destination")),
            requested_artifacts=_normalize_requested_artifacts(task.get("requested_artifacts")),
            artifact_optional=_bool_or_false(task.get("artifact_optional")),
            context_summary=_optional_string(task.get("context_summary")),
            rejected_proposal_context=_json_object_list(task.get("rejected_proposal_context")),
            execution_mode=_optional_string(task.get("execution_mode")),
            intent=_optional_string(task.get("intent")),
            trace_id=_optional_string(task.get("trace_id")),
            workflow_id=_optional_string(task.get("workflow_id")),
            operation_id=_optional_string(task.get("operation_id")),
            execution_id=_optional_string(task.get("execution_id")),
            conversation_id=_optional_string(task.get("conversation_id")),
            case_id=_optional_string(task.get("case_id")),
            channel_id=_optional_string(task.get("channel_id")),
            thread_ts=_optional_string(task.get("thread_ts")),
            trigger_event_id=_optional_string(task.get("trigger_event_id")),
            recent_conversation_entries=_json_object_list(task.get("recent_conversation_entries")),
            case_summary=_json_object_or_empty(task.get("case_summary")) or None,
            prior_trace_refs=_json_object_list(task.get("prior_trace_refs")),
            repo_allowlist=_string_list(task.get("repo_allowlist")),
            tool_allowlist=_string_list(task.get("tool_allowlist")),
            response_mode=_optional_string(task.get("response_mode")),
            reply_delivery_mode=_optional_string(task.get("reply_delivery_mode")),
            context_refs=_json_object_list(task.get("context_refs")),
            approval_mode=_optional_string(task.get("approval_mode")),
            reasoning_verbosity=_optional_string(task.get("reasoning_verbosity")),
            session_scope_kind=_optional_string(task.get("session_scope_kind")),
            session_scope_id=_optional_string(task.get("session_scope_id")),
            parent_session_scope_kind=_optional_string(task.get("parent_session_scope_kind")),
            parent_session_scope_id=_optional_string(task.get("parent_session_scope_id")),
            harness_profile_id=_optional_string(task.get("harness_profile_id")),
            harness_overlay_version=_optional_string(task.get("harness_overlay_version")),
            memory_backend=_optional_string(task.get("memory_backend")),
            assistant_peer_id=_optional_string(task.get("assistant_peer_id")),
            user_peer_id=_optional_string(task.get("user_peer_id")),
            attempt_id=_optional_string(task.get("attempt_id")),
            workspace_id=_optional_string(task.get("workspace_id")),
            workspace_repo=_optional_string(task.get("workspace_repo")),
            workspace_branch=_optional_string(task.get("workspace_branch")),
            allowed_path_globs=_string_list(task.get("allowed_path_globs")),
            mcp_servers=_json_object_list(task.get("mcp_servers")),
            execution_phase=_optional_string(task.get("execution_phase")),
            contract_version=_optional_string(task.get("contract_version")),
            execution_intent=_json_object_or_empty(task.get("execution_intent")),
            capability_leases=_json_object_list(task.get("capability_leases")),
            delivery_policy=_json_object_or_empty(task.get("delivery_policy")),
            workspace_policy=_json_object_or_empty(task.get("workspace_policy")),
            approval_policy=_json_object_or_empty(task.get("approval_policy")),
        )


@dataclass
class ToolPolicy:
    mode: str
    requested: list[str]
    effective: list[str]
    blocked: list[str]
    memory_tools: list[str]
    custom_tools: list[str]
    transport_effective: list[str]
    custom_tool_transport_map: dict[str, str]


class HermesStructuredOutputError(ValueError):
    pass


class HermesRuntime:
    def __init__(self, config: RunnerConfig) -> None:
        self._config = config
        self._configured_model = config.model
        self._reasoning_effort = config.reasoning_effort
        self._role = config.role
        self._backend = "hermes-aiagent"
        self._provider = "hermes"
        self._api_mode = ""
        self._base_url = ""
        self._api_key = ""
        self._provider_model = config.model
        self._provider_hint = ""
        self._reasoning_config = parse_reasoning_effort(config.reasoning_effort) or {"enabled": True, "effort": "medium"}
        self._openai_configured = False
        self._session_manager = SessionManager(config)
        self._adapter = HermesAdapter(config)
        self._company_computer = HermesCompanyComputer(
            computer_root=config.hermes_computer_root,
            run_root=config.hermes_run_root,
            artifact_root=config.hermes_artifact_root,
            hermes_pin=config.hermes_pin,
        )
        self._max_iterations = config.max_iterations
        self._default_task_timeout_seconds = config.task_timeout_seconds
        self._default_inactivity_timeout_seconds = config.inactivity_timeout_seconds
        self._transport_timeout_seconds = config.transport_timeout_seconds
        self._native_max_output_tokens = config.native_max_output_tokens
        self._tool_policy_mode = config.tool_policy_mode
        self._slack_mcp_discovery_error = ""
        self._slack_mcp_tool_cache: list[JsonObject] | None = None
        self._slack_mcp_send_tool_name = ""
        self._mcp_adapter = HermesTaskScopedMCPAdapter(
            default_slack_server_url=self._config.slack_mcp_server_url,
            slack_read_tool_names_resolver=self._slack_mcp_read_tool_names,
            slack_send_tool_name_resolver=self._slack_mcp_send_tool_name_or_error,
        )
        self._executor_recent_results: dict[str, JsonObject] = {}
        self._executor_processes: dict[str, subprocess.Popen[str]] = {}
        self._executor_process_lock = threading.Lock()
        self._executor_cancel_requests: set[str] = set()
        self._configure_runtime()
        self._contract_status = validate_hermes_contract(
            expected_pin=config.hermes_pin,
            hermes_home=config.hermes_home,
            session_db=self._session_manager.session_db,
            required_toolsets=[HERMES_GOVERNED_READONLY_TOOLSET],
        )
        self._available = (
            AIAgent is not None
            and self._runtime_has_credentials()
            and self._session_manager.available
            and self._contract_status.ok
        )

    def _configure_runtime(self) -> None:
        if self._configured_model.startswith("openai/"):
            self._provider = "openai"
            self._provider_hint = "custom"
            self._provider_model = self._configured_model.split("/", 1)[1]
            self._api_mode = "codex_responses"
            self._base_url = first_non_empty(
                os.getenv("RSI_OPENAI_BASE_URL"),
                os.getenv("OPENAI_BASE_URL"),
                "https://api.openai.com/v1",
            )
            self._api_key = first_non_empty(os.getenv("RSI_OPENAI_API_KEY"), os.getenv("OPENAI_API_KEY"))
            self._openai_configured = bool(self._api_key)
            return

        self._provider = first_non_empty(os.getenv("RSI_HERMES_PROVIDER"), "hermes")
        self._provider_hint = first_non_empty(os.getenv("RSI_HERMES_PROVIDER_HINT"), "")
        self._base_url = first_non_empty(os.getenv("RSI_HERMES_BASE_URL"), "")
        self._api_key = first_non_empty(os.getenv("RSI_HERMES_API_KEY"), "")
        self._api_mode = first_non_empty(os.getenv("RSI_HERMES_API_MODE"), "")

    def _runtime_has_credentials(self) -> bool:
        if self._configured_model.startswith("openai/"):
            return bool(self._api_key)
        return True

    @property
    def available(self) -> bool:
        return self._available

    def _hermes_config_parity_status(self) -> str:
        return str(getattr(self._session_manager, "hermes_config_parity_status", "unknown") or "unknown").strip()

    def _hermes_config_parity_error(self) -> str:
        return str(getattr(self._session_manager, "hermes_config_parity_error", "") or "").strip()

    @property
    def metadata(self) -> JsonObject:
        adapter_meta = self._adapter.metadata
        observation_sink_configured = bool(self._config.tool_gateway_base_url)
        return {
            "status": "ok" if self.available and self._session_manager.skills_healthy else "degraded",
            "role": self._role,
            "backend": self._backend,
            "provider": self._provider,
            "model": self._configured_model,
            "provider_model": self._provider_model,
            "reasoning_effort": self._reasoning_effort,
            "api_mode": self._api_mode,
            "available": self.available,
            "hermes_available": AIAgent is not None,
            "openai_configured": self._openai_configured,
            "slack_mcp_enabled": self._config.slack_mcp_enabled,
            "slack_mcp_configured": self._config.slack_mcp_enabled and self._config.slack_user_token_configured,
            "slack_mcp_available": self._slack_mcp_available(),
            "slack_mcp_server_url": self._config.slack_mcp_server_url,
            "slack_mcp_tool_count": len(self._slack_mcp_tools()),
            "persistence_enabled": self._session_manager.available,
            "session_continuity_status": "ok" if self._session_manager.available else "degraded",
            "hermes_home": self._session_manager.hermes_home,
            "session_db_path": self._session_manager.session_db_path,
            "skills_dir": self._session_manager.skills_dir,
            "bundled_skills_available": self._session_manager.bundled_skills_available,
            "bundled_skills_sync_status": self._session_manager.bundled_skills_sync_status,
            "bundled_skills_sync_error": self._session_manager.bundled_skills_sync_error,
            "hermes_config_parity_status": self._hermes_config_parity_status(),
            "hermes_config_parity_error": self._hermes_config_parity_error(),
            "observation_sink_configured": observation_sink_configured,
            "observation_sink_status": "configured" if observation_sink_configured else "not_configured",
            "direct_delivery_phase_enabled": True,
            "execution_contract_version": EXECUTION_CONTRACT_VERSION,
            "execution_envelope_v1_enabled": self._config.execution_envelope_v1_enabled,
            "company_computer_root": self._config.hermes_computer_root,
            "runner_planner_mode": self._config.runner_planner_mode or RUNNER_PLANNER_MODE,
            "required_capabilities": sorted(ALLOWED_CAPABILITIES),
            "hermes_version": adapter_meta.version,
            "hermes_pin": adapter_meta.pin,
            "hermes_contract_status": self._contract_status.to_dict(),
            "memory_backend": self._config.memory_backend,
            "max_iterations": self._max_iterations,
            "task_timeout_seconds": self._default_task_timeout_seconds,
            "inactivity_timeout_seconds": self._default_inactivity_timeout_seconds,
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "native_max_output_tokens": self._native_max_output_tokens,
            "tool_policy_mode": self._tool_policy_mode,
            "hermes_executor_enabled": self._config.hermes_executor_enabled,
            "hermes_executor_service_only": self._config.hermes_executor_service_only,
            "hermes_executor_workspace_root": self._config.hermes_executor_workspace_root,
            "hermes_computer_root": self._config.hermes_computer_root,
            "hermes_run_root": self._config.hermes_run_root,
            "hermes_artifact_root": self._config.hermes_artifact_root,
            "tool_allowlist_effective": self._default_policy_allowlist(execution_mode=""),
            "blocked_tool_names": [],
            "context_engine_mode": adapter_meta.context_engine_mode,
            "context_engine_status": adapter_meta.context_engine_status,
            "lifecycle_hook_status": adapter_meta.lifecycle_hook_status,
            "governed_tools_status": adapter_meta.governed_tools_status,
            "honcho_configured": self._config.honcho_api_key_configured or bool(self._config.honcho_base_url),
            "honcho_available": self._session_manager.honcho_available,
            "honcho_runtime_status": {
                "configured": self._config.honcho_api_key_configured or bool(self._config.honcho_base_url),
                "available": self._session_manager.honcho_available,
                "base_url": self._config.honcho_base_url or "",
                "workspace": self._config.honcho_workspace,
                "environment": self._config.honcho_environment,
                "environment_effective": self._config.honcho_environment_effective,
            },
            "honcho_base_url": self._config.honcho_base_url or "",
            "honcho_workspace": self._config.honcho_workspace,
            "honcho_environment": self._config.honcho_environment,
            "honcho_environment_effective": self._config.honcho_environment_effective,
            "honcho_recall_mode": self._config.honcho_recall_mode,
            "honcho_write_frequency": self._config.honcho_write_frequency,
            "honcho_session_strategy": self._config.honcho_session_strategy,
            "honcho_ai_peer": self._config.honcho_ai_peer,
            "issues": self._session_manager.ready_issues,
        }

    def _native_execution_log_root(self) -> Path:
        return Path(self._config.hermes_home) / "rsi_runtime" / "native_executions"

    def _native_execution_scope(self, task: RunnerTaskRequest) -> str:
        return first_non_empty(
            task.trace_id,
            task.workflow_id,
            task.session_scope_id,
            task.conversation_id,
            task.channel_id,
            task.task_type,
            "execution",
        )

    def _native_execution_slug(self, value: str) -> str:
        text = re.sub(r"[^A-Za-z0-9_.-]+", "_", str(value or "").strip())
        text = text.strip("._-")
        return text or "unknown"

    def _native_execution_key(self, task: RunnerTaskRequest) -> str:
        key = first_non_empty(task.operation_id, task.trace_id, task.workflow_id)
        if not key:
            key = uuid.uuid4().hex
        return self._native_execution_slug(key)

    def _native_run_dir(self, task: RunnerTaskRequest) -> Path:
        run_dir = (
            Path(self._config.hermes_run_root).expanduser()
            / self._native_execution_key(task)
            / self._native_execution_slug(self._execution_phase(task))
        ).resolve()
        run_dir.mkdir(parents=True, exist_ok=True)
        return run_dir

    def _native_artifact_output_dir(self, task: RunnerTaskRequest) -> Path:
        scope = first_non_empty(task.repo, task.channel_id, task.conversation_id, task.session_scope_id, "company")
        artifact_dir = (
            Path(self._config.hermes_artifact_root).expanduser()
            / self._native_execution_slug(scope)
            / time.strftime("%Y-%m-%d", time.gmtime())
            / self._native_execution_key(task)
        ).resolve()
        artifact_dir.mkdir(parents=True, exist_ok=True)
        return artifact_dir

    def _native_artifact_destination(self, task: RunnerTaskRequest) -> Path:
        destination = _string_or_json(task.artifact_destination)
        if destination:
            return Path(destination).expanduser().resolve()
        return self._native_artifact_output_dir(task)

    def _artifact_record_for_path(self, path_value: str, *, kind: str, title: str, task: RunnerTaskRequest) -> JsonObject:
        path = Path(path_value).expanduser().resolve()
        file_ref = f"file://{path}"
        record: JsonObject = {
            "kind": kind,
            "title": path.name or title,
            "artifact_refs": [file_ref],
            "file_ref": file_ref,
            "workspace_path": str(path),
            "delivery_status": "generated",
            "failure_reason": "",
            "created_by_execution_id": first_non_empty(task.execution_id, task.operation_id, task.trace_id),
            "share_status": "local",
        }
        try:
            stat = path.stat()
            if path.is_file():
                record["size_bytes"] = int(stat.st_size)
                digest = hashlib.sha256()
                with path.open("rb") as handle:
                    for chunk in iter(lambda: handle.read(1024 * 1024), b""):
                        digest.update(chunk)
                record["sha256"] = digest.hexdigest()
        except OSError:
            pass
        return record

    def _enrich_artifact_records(self, artifacts: list[JsonObject], task: RunnerTaskRequest) -> list[JsonObject]:
        out: list[JsonObject] = []
        for artifact in artifacts:
            item = dict(artifact)
            refs = _string_list_or_empty(item.get("artifact_refs"))
            file_ref = _string_or_json(item.get("file_ref")) or (refs[0] if refs else "")
            workspace_path = _string_or_json(item.get("workspace_path"))
            if not workspace_path and file_ref:
                workspace_path = _path_from_file_ref(file_ref)
            if workspace_path:
                metadata = self._artifact_record_for_path(
                    workspace_path,
                    kind=_string_or_json(item.get("kind")),
                    title=_string_or_json(item.get("title")),
                    task=task,
                )
                metadata.update(item)
                if not _string_list_or_empty(metadata.get("artifact_refs")):
                    metadata["artifact_refs"] = refs or [f"file://{workspace_path}"]
                if not _string_or_json(metadata.get("file_ref")):
                    metadata["file_ref"] = file_ref or f"file://{workspace_path}"
                item = metadata
            out.append(item)
        return out

    def _mark_artifacts_shared_if_delivered(self, artifacts: list[JsonObject], delivery_output: JsonObject | None) -> list[JsonObject]:
        if not delivery_output:
            return artifacts
        reply_delivery = _json_object_or_empty(delivery_output.get("reply_delivery"))
        status = _string_or_json(reply_delivery.get("status")).lower()
        shared = bool(_string_or_json(reply_delivery.get("provider_ref")) or _string_or_json(reply_delivery.get("message_link")))
        shared = shared or status in {"posted", "sent", "uploaded", "completed", "ok", "success", "shared"}
        if not shared:
            return artifacts
        return [{**artifact, "share_status": "shared"} for artifact in artifacts]

    def _create_native_execution_recorder(self, task: RunnerTaskRequest, *, label: str) -> NativeExecutionRecorder:
        filename = (
            f"{self._native_execution_slug(task.task_type)}"
            f"-{self._native_execution_slug(label)}"
            f"-{self._native_execution_slug(self._native_execution_scope(task))}"
            f"-{time.time_ns()}.jsonl"
        )
        return NativeExecutionRecorder(self._native_execution_log_root() / filename)

    def _attach_native_execution_log_path(
        self,
        raw: JsonObject,
        recorder: NativeExecutionRecorder | None,
    ) -> JsonObject:
        if recorder is None:
            return raw
        out = dict(raw)
        out["native_execution_log_path"] = str(recorder.path)
        return out

    def _native_execution_started_payload(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> JsonObject:
        mcp_servers = []
        for server in task.mcp_servers:
            if not isinstance(server, dict):
                continue
            mcp_servers.append(
                {
                    "server_label": _string_or_json(server.get("server_label")),
                    "server_url": _string_or_json(server.get("server_url")),
                    "profile": _string_or_json(server.get("profile")),
                    "allowed_tools": server.get("allowed_tools"),
                    "require_approval": _string_or_json(server.get("require_approval")),
                    "headers": _json_object_or_empty(server.get("headers")),
                }
            )
        return {
            "task_type": task.task_type,
            "trace_id": task.trace_id or "",
            "workflow_id": task.workflow_id or "",
            "conversation_id": task.conversation_id or "",
            "channel_id": task.channel_id or "",
            "thread_ts": task.thread_ts or "",
            "session_scope_kind": task.session_scope_kind or "",
            "session_scope_id": task.session_scope_id or "",
            "tool_policy_mode": tool_policy.mode,
            "tool_allowlist_effective": tool_policy.effective,
            "tool_transport_allowlist_effective": tool_policy.transport_effective,
            "blocked_tool_names": tool_policy.blocked,
            "allowed_tools": task.allowed_tools,
            "timeout_seconds": self._effective_task_timeout(task),
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "model": self._provider_model,
            "mcp_servers": mcp_servers,
        }

    def execute(self, prompt: str, system_message: str | None = None) -> HermesExecutionResult:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "general",
                    "repo": "rsi-agent-platform",
                    "prompt": prompt,
                    "system_message": system_message,
                    "session_scope_kind": "adhoc",
                    "session_scope_id": self._role,
                    "memory_backend": self._config.memory_backend,
                    "assistant_peer_id": self._config.honcho_ai_peer,
                }
            }
        )
        return self._execute_task_request(
            task,
            self._resolve_tool_policy(task),
            render_prompt=False,
            expand_skills=False,
        )

    def _native_governed_tools_enabled(self, task: RunnerTaskRequest) -> bool:
        if not self._config.hermes_native_governed_tools_enabled:
            return False
        if self._role != "proposal":
            return False
        return (task.execution_mode or "").strip().lower() in {"diagnose", "implement"}

    def _execution_phase(self, task: RunnerTaskRequest) -> str:
        return (task.execution_phase or "").strip().lower() or "main"

    def _task_uses_artifact_phases(self, task: RunnerTaskRequest) -> bool:
        return (
            task.task_type == "workflow"
            and self._execution_phase(task) == "main"
            and len(task.requested_artifacts) > 0
        )

    def _phase_max_iterations_override(self, task: RunnerTaskRequest) -> int | None:
        phase = self._execution_phase(task)
        if phase == "render":
            return 4
        if phase == "deliver":
            return 3
        return None

    def _artifact_phase_budgets(self, task: RunnerTaskRequest) -> JsonObject:
        total = self._effective_task_timeout(task)
        reserve_for_investigate = max(1, min(60, int(total * 0.45)))
        architecture_diagram_requested = "architecture-diagram" in normalize_tool_names(task.requested_skills)
        desired_render = min(180, max(45, int(total * 0.25)))
        if architecture_diagram_requested:
            desired_render = 300
        desired = {
            "render": desired_render,
            "deliver": min(60, max(30, int(total * 0.1))),
            "reducer_reserve": min(180, max(60, int(total * 0.2))),
        }
        available_for_other = max(0, total - reserve_for_investigate)
        desired_other = sum(desired.values())
        if desired_other <= available_for_other:
            render_budget = desired["render"]
            deliver_budget = desired["deliver"]
            reducer_reserve = desired["reducer_reserve"]
        elif desired_other <= 0 or available_for_other <= 0:
            render_budget = 0
            deliver_budget = 0
            reducer_reserve = 0
        else:
            scale = available_for_other / desired_other
            render_budget = int(desired["render"] * scale)
            deliver_budget = int(desired["deliver"] * scale)
            reducer_reserve = int(desired["reducer_reserve"] * scale)
        investigate_budget = max(1, total - render_budget - deliver_budget - reducer_reserve)
        return {
            "investigate": investigate_budget,
            "render": render_budget,
            "deliver": deliver_budget,
            "reducer_reserve": reducer_reserve,
            "total": total,
        }

    def _native_toolsets_for_task(self, task: RunnerTaskRequest, *, extra_toolsets: list[str] | None = None) -> list[str]:
        toolsets: list[str] = []
        execution_phase = self._execution_phase(task)
        native_governed_tools = self._native_governed_tools_enabled(task)
        execution_mode = (task.execution_mode or "").strip().lower()
        if execution_phase not in {"render", "deliver"} and (
            task.task_type in {"workflow", "question_gather", "question_expand"} or native_governed_tools
        ):
            toolsets.extend(["todo", "session_search"])
        add_workflow_governed_readonly = task.task_type == "workflow" and self._config.tool_gateway_base_url
        add_workflow_governed_workspace = add_workflow_governed_readonly and (
            execution_mode in {"implement", "artifact_render"} or execution_phase == "render"
        )
        if add_workflow_governed_readonly:
            toolsets.append(HERMES_GOVERNED_READONLY_TOOLSET)
        if add_workflow_governed_workspace:
            toolsets.append(HERMES_GOVERNED_WORKSPACE_TOOLSET)
        if native_governed_tools:
            if not add_workflow_governed_readonly:
                toolsets.append(HERMES_GOVERNED_READONLY_TOOLSET)
            if (execution_mode in {"implement", "artifact_render"} or execution_phase == "render") and not add_workflow_governed_workspace:
                toolsets.append(HERMES_GOVERNED_WORKSPACE_TOOLSET)
        if execution_phase == "render" or (
            task.task_type == "workflow"
            and (execution_mode == "artifact_render" or len(task.requested_artifacts) > 0)
        ):
            toolsets.append(HERMES_ARTIFACT_TOOLSET)
        if extra_toolsets:
            toolsets.extend(extra_toolsets)
        return normalize_tool_names(toolsets)

    def _create_agent(
        self,
        task: RunnerTaskRequest,
        context: SessionContext,
        *,
        max_iterations_override: int | None = None,
        enabled_toolsets_override: list[str] | None = None,
    ) -> Any:
        configured_max_iterations = max_iterations_override if max_iterations_override and max_iterations_override > 0 else self._max_iterations
        agent_kwargs: JsonObject = {
            "model": self._provider_model,
            "quiet_mode": True,
            "reasoning_config": self._reasoning_config,
            "enabled_toolsets": enabled_toolsets_override
            if enabled_toolsets_override is not None
            else self._native_toolsets_for_task(task),
            "skip_context_files": True,
            "skip_memory": False,
            "persist_session": True,
            "max_iterations": configured_max_iterations,
            "session_id": context.session_id,
            "parent_session_id": context.parent_session_id or None,
            "session_db": self._session_manager.session_db,
        }
        if self._provider_hint:
            agent_kwargs["provider"] = self._provider_hint
        if self._api_mode:
            agent_kwargs["api_mode"] = self._api_mode
        if self._base_url:
            agent_kwargs["base_url"] = self._base_url
        if self._api_key:
            agent_kwargs["api_key"] = self._api_key
        return AIAgent(**agent_kwargs)

    def _extract_user_request_text(self, prompt: str) -> str:
        text = str(prompt or "")
        if not text:
            return ""
        if text.startswith("User request: "):
            remainder = text[len("User request: "):]
            first_block, _, _ = remainder.partition("\n\n")
            return first_block.strip()
        return text.strip()

    def _skill_mentions_from_text(self, text: str) -> list[str]:
        seen: set[str] = set()
        out: list[str] = []
        for match in re.finditer(r"(?:^|[\s(])/([A-Za-z0-9][A-Za-z0-9_-]*)\b", text or ""):
            identifier = match.group(1).replace("_", "-").strip().lower()
            if not identifier or identifier in seen:
                continue
            seen.add(identifier)
            out.append(identifier)
        return out

    def _skill_runtime_requested(self, task: RunnerTaskRequest) -> list[str]:
        requested = [name.replace("_", "-").strip().lower() for name in task.requested_skills if str(name or "").strip()]
        return normalize_tool_names(requested)

    def _expand_task_skills(self, task: RunnerTaskRequest, context: SessionContext) -> tuple[RunnerTaskRequest, JsonObject]:
        diagnostics: JsonObject = {
            "requested_skills": [],
            "resolved_skills": [],
            "missing_skills": [],
            "skill_injection_mode": "none",
        }
        user_request = self._extract_user_request_text(task.prompt)
        explicit_mentions = self._skill_mentions_from_text(user_request)
        requested_skills = []
        seen_requested: set[str] = set()
        for identifier in [*explicit_mentions, *self._skill_runtime_requested(task)]:
            normalized = identifier.replace("_", "-").strip().lower()
            if not normalized or normalized in seen_requested:
                continue
            seen_requested.add(normalized)
            requested_skills.append(normalized)
        diagnostics["requested_skills"] = list(requested_skills)
        if not requested_skills:
            return task, diagnostics
        if build_skill_invocation_message is None or build_preloaded_skills_prompt is None or resolve_skill_command_key is None:
            diagnostics["missing_skills"] = list(requested_skills)
            diagnostics["skill_injection_mode"] = "helpers_unavailable"
            return task, diagnostics

        prompt_prefix_parts: list[str] = []
        resolved_skills: list[str] = []
        missing_skills: list[str] = []
        injection_modes: list[str] = []

        remaining_preloads = list(requested_skills)
        stripped_request = user_request.lstrip()
        if stripped_request.startswith("/"):
            command, _, user_instruction = stripped_request.partition(" ")
            command_key = resolve_skill_command_key(command.lstrip("/"))
            if command_key:
                invocation_message = build_skill_invocation_message(
                    command_key,
                    user_instruction=user_instruction.strip(),
                    task_id=context.session_id,
                )
                if invocation_message:
                    prompt_prefix_parts.append(invocation_message)
                    resolved_identifier = command_key.lstrip("/").replace("_", "-").strip().lower()
                    resolved_skills.append(resolved_identifier)
                    remaining_preloads = [item for item in remaining_preloads if item != resolved_identifier]
                    injection_modes.append("slash_command")
            else:
                unresolved = command.lstrip("/").replace("_", "-").strip().lower()
                if unresolved:
                    missing_skills.append(unresolved)

        if remaining_preloads:
            preloaded_prompt, loaded_skill_names, missing_identifiers = build_preloaded_skills_prompt(
                remaining_preloads,
                task_id=context.session_id,
            )
            if preloaded_prompt:
                prompt_prefix_parts.append(preloaded_prompt)
            if preloaded_prompt or loaded_skill_names or missing_identifiers:
                injection_modes.append("preloaded")
            for name in loaded_skill_names:
                normalized = str(name or "").replace("_", "-").strip().lower()
                if normalized:
                    resolved_skills.append(normalized)
            for identifier in missing_identifiers:
                normalized = str(identifier or "").replace("_", "-").strip().lower()
                if normalized:
                    missing_skills.append(normalized)

        resolved_skills = normalize_tool_names(resolved_skills)
        missing_skills = normalize_tool_names([item for item in missing_skills if item not in resolved_skills])
        diagnostics["resolved_skills"] = resolved_skills
        diagnostics["missing_skills"] = missing_skills
        diagnostics["skill_injection_mode"] = "+".join(injection_modes) if injection_modes else "none"

        if not prompt_prefix_parts:
            return task, diagnostics
        expanded_prompt = "\n\n".join([*prompt_prefix_parts, task.prompt])
        return replace(task, prompt=expanded_prompt), diagnostics

    def _execute_task_request(
        self,
        task: RunnerTaskRequest,
        tool_policy: ToolPolicy,
        *,
        observer: ObservationEmitter | None = None,
        allow_partial_recovery: bool = True,
        max_iterations_override: int | None = None,
        render_prompt: bool = True,
        expand_skills: bool = True,
    ) -> HermesExecutionResult:
        if AIAgent is None:
            return HermesExecutionResult(
                ok=False,
                message="Hermes runtime is not installed in this environment.",
                provider=self._backend,
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )
        if not self._contract_status.ok:
            return HermesExecutionResult(
                ok=False,
                message="Hermes adapter contract failed: " + "; ".join(self._contract_status.errors),
                provider=self._backend,
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "failure_class": "hermes_contract_failed",
                    "runner_diagnostics": {
                        "failure_kind": "hermes_contract_failed",
                        "termination_reason": "hermes_contract_failed",
                        "contract_errors": list(self._contract_status.errors),
                    },
                },
            )
        if not self._runtime_has_credentials():
            return HermesExecutionResult(
                ok=False,
                message="Hermes OpenAI runtime selected but RSI_OPENAI_API_KEY / OPENAI_API_KEY is not configured.",
                provider=self._backend,
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )
        if not self._session_manager.available:
            return HermesExecutionResult(
                ok=False,
                message="Hermes persistent session runtime is unavailable.",
                provider=self._backend,
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )

        if preflight := self._preflight_tool_policy_failure(task, tool_policy):
            return preflight

        context = self._session_manager.prepare(task)
        execution_phase = self._execution_phase(task)
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="session.prepared",
                status="completed",
                payload={
                    "session_id": context.session_id,
                    "parent_session_id": context.parent_session_id,
                },
            )
            observer.emit(
                phase=execution_phase,
                event_type="hermes.config",
                status=self._hermes_config_parity_status(),
                payload={
                    "skills_dir": self._session_manager.skills_dir,
                    "config_parity_error": self._hermes_config_parity_error(),
                },
            )
            observer.emit(
                phase=execution_phase,
                event_type="skills.sync",
                status=self._session_manager.bundled_skills_sync_status,
                payload={
                    "bundled_skills_available": self._session_manager.bundled_skills_available,
                    "sync_error": self._session_manager.bundled_skills_sync_error,
                },
            )
            observer.emit(phase=execution_phase, event_type="phase.started", status="running")
        skill_diagnostics: JsonObject = {
            "requested_skills": [],
            "resolved_skills": [],
            "missing_skills": [],
            "skill_injection_mode": "none",
        }
        if expand_skills:
            task, skill_diagnostics = self._expand_task_skills(task, context)
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="skills.expanded",
                status="completed",
                payload=skill_diagnostics,
            )
        if render_prompt:
            task = replace(task, prompt=self._render_task_prompt(task, tool_policy))
        effective_task_timeout = self._effective_task_timeout(task)
        effective_inactivity_timeout = self._effective_inactivity_timeout(effective_task_timeout)
        reasoning_timeout_seconds = self._partial_completion_reasoning_timeout_seconds(task, effective_task_timeout)
        configured_max_iterations = max_iterations_override if max_iterations_override and max_iterations_override > 0 else self._max_iterations
        agent = None
        tracker = None
        agentic_mcp_registration = TaskScopedMCPRegistration()
        result: HermesExecutionResult | None = None
        try:
            self._stage_task_context(context.session_id, task, tool_policy)
            try:
                agentic_mcp_registration = self._mcp_adapter.register_task_servers(task)
            except RuntimeError as exc:
                if observer is not None:
                    observer.emit(
                        phase=execution_phase,
                        event_type="mcp.registration",
                        status="failed",
                        payload={"error": str(exc)},
                    )
                result = HermesExecutionResult(
                    ok=False,
                    message=str(exc),
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        "failure_class": "runner_non_ok",
                        "tool_policy_mode": tool_policy.mode,
                        "tool_allowlist_effective": tool_policy.effective,
                        "tool_transport_allowlist_effective": tool_policy.transport_effective,
                        "tool_transport_map": tool_policy.custom_tool_transport_map,
                        "blocked_tool_names": tool_policy.blocked,
                        "runner_diagnostics": self._runner_diagnostics(
                            tool_policy,
                            failure_kind="agentic_mcp_registration_failed",
                            provider_error_message=str(exc),
                            termination_reason="agentic_mcp_registration_failed",
                            session_ready_issues=self._session_manager.ready_issues,
                            repair_attempted=False,
                            repair_succeeded=False,
                        ),
                    },
                )
                return result
            if observer is not None:
                observer.emit(
                    phase=execution_phase,
                    event_type="mcp.registration",
                    status="completed",
                    payload={
                        "enabled": agentic_mcp_registration.enabled,
                        "server_names": agentic_mcp_registration.server_names,
                        "toolsets": agentic_mcp_registration.enabled_toolsets,
                    },
                )
            agent = self._create_agent(
                task,
                context,
                max_iterations_override=max_iterations_override,
                enabled_toolsets_override=self._native_toolsets_for_task(
                    task, extra_toolsets=agentic_mcp_registration.enabled_toolsets
                ),
            )
            self._attach_tool_policy(agent, task, tool_policy, observer=observer)
            tracker = self._session_manager.attach_tracking(agent, task, context)
            termination_reason, run_result, stop_meta = self._run_with_deadlines(
                agent,
                task,
                context,
                effective_task_timeout,
                effective_inactivity_timeout,
                reasoning_timeout_seconds,
                observer=observer,
            )
            lifecycle_events = self._adapter.lifecycle_events(context.session_id)
            if termination_reason != "normal_completion":
                finalized = self._session_manager.finalize(context, tracker)
                observed = self._observability_metadata(agent, task, tracker, skill_diagnostics=skill_diagnostics, observer=observer)
                last_activity = _json_object_or_empty(stop_meta.get("last_activity"))
                if termination_reason in {"task_timeout", "iteration_budget_exhausted"} and allow_partial_recovery and self._workflow_partial_completion_eligible(task):
                    result = self._finalize_partial_completion(
                        task,
                        tool_policy,
                        finalized,
                        observed,
                        stop_meta,
                        lifecycle_events,
                        termination_reason=termination_reason,
                        observer=observer,
                    )
                    return result
                if termination_reason == "inactivity_timeout":
                    timeout_kind = string_from_map(stop_meta, "timeout_kind") or termination_reason
                    timeout_message = f"Hermes execution timed out after {effective_task_timeout}s."
                    if timeout_kind == "inactivity_timeout":
                        timeout_message = f"Hermes execution hit inactivity timeout after {effective_inactivity_timeout}s."
                    result = HermesExecutionResult(
                        ok=False,
                        message=timeout_message,
                        provider=self._backend,
                        raw={
                            **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                            **finalized,
                            **stop_meta,
                            "task_timeout_seconds": effective_task_timeout,
                            "inactivity_timeout_seconds": effective_inactivity_timeout,
                            "transport_timeout_seconds": self._transport_timeout_seconds,
                            "max_iterations": configured_max_iterations,
                            "tool_policy_mode": tool_policy.mode,
                            "tool_allowlist_effective": tool_policy.effective,
                            "tool_transport_allowlist_effective": tool_policy.transport_effective,
                            "blocked_tool_names": tool_policy.blocked,
                            **observed,
                            **self._workflow_evidence_raw(task, observed, timeout_kind),
                            "failure_class": "runner_transport_timeout",
                            "runner_diagnostics": self._runner_diagnostics(
                                tool_policy,
                                failure_kind="transport_timeout",
                                provider_error_message=timeout_message,
                                timeout_kind=timeout_kind,
                                termination_reason=timeout_kind,
                                activity=last_activity,
                                max_iterations_reached=bool(stop_meta.get("max_iterations_reached")),
                                session_ready_issues=self._session_manager.ready_issues,
                                repair_attempted=False,
                                repair_succeeded=False,
                                observed=observed,
                            ),
                            "lifecycle_events": lifecycle_events,
                            "termination_reason": timeout_kind,
                        },
                    )
                    return result
                if termination_reason == "task_timeout":
                    timeout_message = f"Hermes execution timed out after {effective_task_timeout}s."
                    result = HermesExecutionResult(
                        ok=False,
                        message=timeout_message,
                        provider=self._backend,
                        raw={
                            **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                            **finalized,
                            **stop_meta,
                            "task_timeout_seconds": effective_task_timeout,
                            "inactivity_timeout_seconds": effective_inactivity_timeout,
                            "transport_timeout_seconds": self._transport_timeout_seconds,
                            "max_iterations": configured_max_iterations,
                            "tool_policy_mode": tool_policy.mode,
                            "tool_allowlist_effective": tool_policy.effective,
                            "tool_transport_allowlist_effective": tool_policy.transport_effective,
                            "blocked_tool_names": tool_policy.blocked,
                            **observed,
                            **self._workflow_evidence_raw(task, observed, "task_timeout"),
                            "failure_class": "runner_transport_timeout",
                            "runner_diagnostics": self._runner_diagnostics(
                                tool_policy,
                                failure_kind="transport_timeout",
                                provider_error_message=timeout_message,
                                timeout_kind="task_timeout",
                                termination_reason="task_timeout",
                                activity=last_activity,
                                max_iterations_reached=bool(stop_meta.get("max_iterations_reached")),
                                session_ready_issues=self._session_manager.ready_issues,
                                repair_attempted=False,
                                repair_succeeded=False,
                                observed=observed,
                            ),
                            "lifecycle_events": lifecycle_events,
                            "termination_reason": "task_timeout",
                        },
                    )
                    return result
                if termination_reason == "iteration_budget_exhausted":
                    result = self._partial_completion_failure(
                        task,
                        tool_policy,
                        finalized,
                        observed,
                        stop_meta,
                        lifecycle_events,
                        termination_reason=termination_reason,
                        recovery_attempted=False,
                        recovery_succeeded=False,
                    )
                    return result
            response = str((run_result or {}).get("final_response", "") or "")
        except Exception as exc:
            diagnostics = self._provider_invalid_request_diagnostics(str(exc), tool_policy)
            activity = safe_activity_summary(agent) if agent is not None else {}
            observed = self._observability_metadata(agent, task, skill_diagnostics=skill_diagnostics, observer=observer)
            raw = {
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                "error": str(exc),
                "tool_policy_mode": tool_policy.mode,
                "tool_allowlist_effective": tool_policy.effective,
                "tool_transport_allowlist_effective": tool_policy.transport_effective,
                "tool_transport_map": tool_policy.custom_tool_transport_map,
                "blocked_tool_names": tool_policy.blocked,
                **observed,
                **self._workflow_evidence_raw(task, observed, "exception"),
            }
            if diagnostics is not None:
                raw["failure_class"] = "runner_invalid_request"
                merged = dict(diagnostics)
                for key, value in observed.items():
                    merged[key] = value
                raw["runner_diagnostics"] = merged
            else:
                raw["failure_class"] = "runner_non_ok"
                raw["runner_diagnostics"] = self._runner_diagnostics(
                    tool_policy,
                    failure_kind="execution_error",
                    provider_error_message=str(exc),
                    termination_reason="exception",
                    activity=activity,
                    max_iterations_reached=bool(activity.get("budget_used", 0) >= activity.get("budget_max", 0) and activity.get("budget_max", 0) > 0),
                    session_ready_issues=self._session_manager.ready_issues,
                    repair_attempted=False,
                    repair_succeeded=False,
                    observed=observed,
                )
            result = HermesExecutionResult(
                ok=False,
                message=f"Hermes execution failed: {exc}",
                provider=self._backend,
                raw=raw,
            )
            return result
        finally:
            cleanup_result = self._mcp_adapter.cleanup_registration(agentic_mcp_registration)
            if result is not None:
                result.raw = self._attach_agentic_mcp_diagnostics(
                    _json_object_or_empty(result.raw),
                    agentic_mcp_registration,
                    cleanup_status=cleanup_result.status,
                    cleanup_errors=cleanup_result.errors,
                )
            if observer is not None:
                observer.emit(
                    phase=execution_phase,
                    event_type="mcp.cleanup",
                    status=cleanup_result.status,
                    payload={"errors": cleanup_result.errors},
                )
                if result is not None:
                    observer.emit(
                        phase=execution_phase,
                        event_type="phase.completed",
                        status="completed" if result.ok else "failed",
                        payload={
                            "termination_reason": _string_or_json(_json_object_or_empty(result.raw).get("termination_reason")),
                            "completion_verdict": _string_or_json(_json_object_or_empty(result.raw).get("completion_verdict")),
                        },
                    )

        finalized = self._session_manager.finalize(context, tracker)
        lifecycle_events = self._adapter.lifecycle_events(context.session_id)
        observed = self._observability_metadata(agent, task, tracker, skill_diagnostics=skill_diagnostics, observer=observer)
        result = HermesExecutionResult(
            ok=True,
            message=response,
            provider=self._backend,
            raw={
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                **finalized,
                "task_timeout_seconds": effective_task_timeout,
                "inactivity_timeout_seconds": effective_inactivity_timeout,
                "transport_timeout_seconds": self._transport_timeout_seconds,
                "max_iterations": configured_max_iterations,
                "tool_policy_mode": tool_policy.mode,
                "tool_allowlist_effective": tool_policy.effective,
                "tool_transport_allowlist_effective": tool_policy.transport_effective,
                "tool_transport_map": tool_policy.custom_tool_transport_map,
                "blocked_tool_names": tool_policy.blocked,
                **observed,
                **self._workflow_evidence_raw(task, observed, "normal_completion"),
                "runner_diagnostics": {
                    **observed,
                    "termination_reason": "normal_completion",
                    "max_iterations_reached": False,
                    "completion_verdict": "complete",
                },
                "lifecycle_events": lifecycle_events,
                "termination_reason": "normal_completion",
                "max_iterations_reached": False,
                "completion_verdict": "complete",
                "harness_profile_id": task.harness_profile_id,
                "effective_overlay_version": task.harness_overlay_version,
            },
        )
        result.raw = self._attach_agentic_mcp_diagnostics(
            _json_object_or_empty(result.raw),
            agentic_mcp_registration,
            cleanup_status=(agentic_mcp_registration.cleanup_result.status if agentic_mcp_registration.cleanup_result else "not_needed"),
            cleanup_errors=(agentic_mcp_registration.cleanup_result.errors if agentic_mcp_registration.cleanup_result else []),
        )
        return result

    def _base_raw(self, prompt: str = "", system_message: str | None = None) -> JsonObject:
        adapter_meta = self._adapter.metadata
        return {
            "role": self._role,
            "backend": self._backend,
            "provider": self._provider,
            "provider_hint": self._provider_hint,
            "api_mode": self._api_mode,
            "model": self._configured_model,
            "provider_model": self._provider_model,
            "reasoning_effort": self._reasoning_effort,
            "reasoning_config": self._reasoning_config,
            "hermes_version": adapter_meta.version,
            "hermes_pin": adapter_meta.pin,
            "hermes_contract_status": self._contract_status.to_dict(),
            "context_engine_mode": adapter_meta.context_engine_mode,
            "context_engine_status": adapter_meta.context_engine_status,
            "lifecycle_hook_status": adapter_meta.lifecycle_hook_status,
            "governed_tools_status": adapter_meta.governed_tools_status,
            "base_url": self._base_url,
            "hermes_executor_workspace_root": self._config.hermes_executor_workspace_root,
            "hermes_computer_root": self._config.hermes_computer_root,
            "hermes_run_root": self._config.hermes_run_root,
            "hermes_artifact_root": self._config.hermes_artifact_root,
            "honcho_base_url": self._config.honcho_base_url or "",
            "honcho_workspace": self._config.honcho_workspace,
            "honcho_environment": self._config.honcho_environment,
            "honcho_environment_effective": self._config.honcho_environment_effective,
            "honcho_recall_mode": self._config.honcho_recall_mode,
            "honcho_write_frequency": self._config.honcho_write_frequency,
            "honcho_session_strategy": self._config.honcho_session_strategy,
            "honcho_ai_peer": self._config.honcho_ai_peer,
            "skills_dir": self._session_manager.skills_dir,
            "bundled_skills_available": self._session_manager.bundled_skills_available,
            "bundled_skills_sync_status": self._session_manager.bundled_skills_sync_status,
            "bundled_skills_sync_error": self._session_manager.bundled_skills_sync_error,
            "prompt": prompt,
            "system_message": system_message,
        }

    def executor_status(self, execution_id: str) -> JsonObject:
        key = str(execution_id or "").strip()
        if not key:
            return {}
        cached = self._executor_recent_results.get(key)
        if cached:
            return dict(cached)
        path = self._executor_status_path(key)
        try:
            if not path.exists():
                return {}
            payload = json.loads(path.read_text(encoding="utf-8"))
        except Exception as exc:
            logger.warning("executor status read failed execution_id=%s path=%s error=%s", key, path, exc)
            return {}
        if not isinstance(payload, dict):
            return {}
        if str(payload.get("status") or "").strip().lower() in {"running", "accepted", "cancelling"}:
            with self._executor_process_lock:
                active = key in self._executor_processes
            if not active:
                payload = dict(payload)
                payload["status"] = "orphaned"
                payload["message"] = "Persisted executor status was running, but no local execution process is active."
        self._executor_recent_results[key] = dict(payload)
        return dict(payload)

    def cancel_execution(self, execution_id: str) -> JsonObject:
        key = str(execution_id or "").strip()
        if not key:
            return {"error": "execution_id is required"}
        with self._executor_process_lock:
            self._executor_cancel_requests.add(key)
            process = self._executor_processes.get(key)
            if process is not None and process.poll() is None:
                try:
                    process.terminate()
                except OSError:
                    pass
        status = self.executor_status(key)
        if not status:
            status = {"execution_id": key, "status": "cancelling"}
        else:
            status["status"] = "cancelling"
        self._store_executor_result(key, status)
        return status

    def _native_executor_enabled_for_task(self, task: RunnerTaskRequest) -> bool:
        return self._config.hermes_executor_enabled and task.task_type == "workflow"

    def _native_executor_workspace_root(self, task: RunnerTaskRequest) -> Path:
        root = Path(self._config.hermes_computer_root).expanduser().resolve()
        root.mkdir(parents=True, exist_ok=True)
        return root

    def _stage_native_executor_workspace(self, task: RunnerTaskRequest, *, observer: ObservationEmitter | None = None) -> Path:
        root = self._native_executor_workspace_root(task)
        run_dir = self._native_run_dir(task)
        artifacts_dir = self._native_artifact_destination(task)
        inputs_dir = run_dir / "inputs"
        artifacts_dir.mkdir(parents=True, exist_ok=True)
        inputs_dir.mkdir(parents=True, exist_ok=True)
        manifest = {
            "repo": task.repo,
            "repo_ref": task.repo_ref,
            "repo_allowlist": list(task.repo_allowlist),
            "workspace_id": task.workspace_id,
            "workspace_repo": task.workspace_repo,
            "workspace_branch": task.workspace_branch,
            "attempt_id": task.attempt_id,
            "requested_artifacts": task.requested_artifacts,
            "requested_skills": task.requested_skills,
            "context_refs": task.context_refs,
            "contract_version": task.contract_version or EXECUTION_CONTRACT_VERSION,
            "execution_intent": task.execution_intent,
            "capability_leases": task.capability_leases or self._company_computer.task_leases(task),
            "delivery_policy": task.delivery_policy,
            "workspace_policy": task.workspace_policy,
            "approval_policy": task.approval_policy,
        }
        manifest_path = inputs_dir / "task-manifest.json"
        manifest_path.write_text(json.dumps(manifest, indent=2, sort_keys=True), encoding="utf-8")
        if observer is not None:
            observer.emit(
                phase=self._execution_phase(task),
                event_type="workspace.staged",
                status="completed",
                payload={
                    "workspace_root": str(root),
                    "run_dir": str(run_dir),
                    "artifact_output_dir": str(artifacts_dir),
                    "inputs_dir": str(inputs_dir),
                    "manifest_path": str(manifest_path),
                },
            )
        return root

    def _native_executor_conversation_history(self, task: RunnerTaskRequest, context: SessionContext) -> list[JsonObject]:
        if self._execution_phase(task) in {"render", "deliver"}:
            return []
        return [item for item in list(context.conversation_history) if isinstance(item, dict)]

    def _native_executor_phase_contract(self, task: RunnerTaskRequest, toolsets: list[str]) -> JsonObject:
        execution_phase = self._execution_phase(task)
        required_toolsets: list[str] = []
        if execution_phase == "render":
            required_toolsets.append(HERMES_ARTIFACT_TOOLSET)
        elif task.task_type == "workflow" and execution_phase not in {"deliver"} and self._config.tool_gateway_base_url:
            required_toolsets.append(HERMES_GOVERNED_READONLY_TOOLSET)
        return {
            "execution_phase": execution_phase,
            "history_policy": "empty" if execution_phase in {"render", "deliver"} else "session",
            "required_toolsets": required_toolsets,
            "toolsets": list(toolsets),
            "missing_required_toolsets": [item for item in required_toolsets if item not in set(toolsets)],
        }

    def _native_executor_completion_meta(self, parsed_result: JsonObject, max_iterations: int) -> JsonObject:
        result_payload = _json_object_or_empty(parsed_result.get("result"))
        completed_value = result_payload.get("completed")
        completed_known = isinstance(completed_value, bool)
        completed = completed_value if completed_known else True
        partial = _bool_or_false(result_payload.get("partial")) or _bool_or_false(parsed_result.get("native_result_partial"))
        interrupted = _bool_or_false(result_payload.get("interrupted")) or _bool_or_false(parsed_result.get("native_result_interrupted"))
        api_calls_value = result_payload.get("api_calls")
        api_calls = int(api_calls_value) if isinstance(api_calls_value, Number) and not isinstance(api_calls_value, bool) else 0
        max_iterations_reached = (
            partial
            or interrupted
            or (completed_known and not completed)
            or (max_iterations > 0 and api_calls >= max_iterations and not completed)
            or _bool_or_false(parsed_result.get("max_iterations_reached"))
        )
        return {
            "termination_reason": "iteration_budget_exhausted" if max_iterations_reached else "normal_completion",
            "completion_verdict": "partial" if max_iterations_reached else "complete",
            "max_iterations_reached": max_iterations_reached,
            "native_result_completed": completed,
            "native_result_partial": partial,
            "native_result_interrupted": interrupted,
            "native_result_api_calls": api_calls,
        }

    def _native_executor_request_payload(
        self,
        task: RunnerTaskRequest,
        context: SessionContext,
        *,
        toolsets: list[str],
        task_scoped_mcp_registration: TaskScopedMCPRegistration,
        max_iterations: int,
        workdir: Path,
        result_path: Path,
        phase_contract: JsonObject,
    ) -> JsonObject:
        return {
            "session_id": context.session_id,
            "parent_session_id": context.parent_session_id,
            "conversation_history": self._native_executor_conversation_history(task, context),
            "prompt": task.prompt,
            "system_message": task.system_message or "",
            "execution_phase": self._execution_phase(task),
            "requested_skills": list(task.requested_skills),
            "toolsets": list(toolsets),
            "phase_contract": phase_contract,
            "task_scoped_mcp_servers": [
                {
                    "source_label": server.source_label,
                    "profile": server.profile,
                    "server_name": server.server_name,
                    "toolset_alias": server.toolset_alias,
                    "included_tool_names": list(server.included_tool_names),
                    "hermes_config": dict(server.hermes_config),
                }
                for server in task_scoped_mcp_registration.servers
            ],
            "model": self._provider_model,
            "max_iterations": max_iterations,
            "reasoning_config": self._reasoning_config,
            "request_overrides": {},
            "workdir": str(workdir),
            "result_path": str(result_path),
            "hermes_home": self._config.hermes_home,
            "artifact_output_dir": str(self._native_artifact_destination(task)),
            "hermes_computer_root": self._config.hermes_computer_root,
            "hermes_run_root": self._config.hermes_run_root,
            "hermes_artifact_root": self._config.hermes_artifact_root,
            "run_dir": str(result_path.parent),
            "contract_version": task.contract_version or EXECUTION_CONTRACT_VERSION,
            "execution_intent": task.execution_intent,
            "capability_leases": task.capability_leases or self._company_computer.task_leases(task),
            "delivery_policy": task.delivery_policy,
            "workspace_policy": task.workspace_policy,
            "approval_policy": task.approval_policy,
            "runtime": {
                "provider": self._provider_hint or "custom",
                "base_url": self._base_url,
                "api_mode": self._api_mode,
            },
        }

    def _parse_native_executor_stdout(self, stdout: str) -> JsonObject:
        payload = ""
        for line in str(stdout or "").splitlines():
            if line.startswith(_NATIVE_EXECUTOR_RESULT_MARKER):
                payload = line[len(_NATIVE_EXECUTOR_RESULT_MARKER) :].strip()
        if not payload:
            raise ValueError("native Hermes executor did not emit a parseable result payload")
        parsed = json.loads(payload)
        if not isinstance(parsed, dict):
            raise ValueError("native Hermes executor returned a non-object result payload")
        return parsed

    def _load_native_executor_result_file(self, result_path: Path) -> JsonObject:
        parsed = json.loads(result_path.read_text(encoding="utf-8"))
        if not isinstance(parsed, dict):
            raise ValueError("native Hermes executor result file did not contain a JSON object")
        return parsed

    def _read_native_executor_stream(
        self,
        stream: Any,
        *,
        stream_name: str,
        phase: str,
        observer: ObservationEmitter | None,
        chunk_store: list[str],
        secret_values: list[str],
        result_detected: threading.Event,
    ) -> None:
        chunk_index = 0
        marker_tail = ""
        marker_window = max(0, len(_NATIVE_EXECUTOR_RESULT_MARKER) - 1)
        try:
            while True:
                chunk = stream.read(_NATIVE_EXECUTOR_OUTPUT_CHUNK_CHARS)
                if not chunk:
                    break
                chunk_store.append(chunk)
                contains_result_marker = False
                if stream_name == "stdout":
                    combined = marker_tail + chunk
                    contains_result_marker = _NATIVE_EXECUTOR_RESULT_MARKER in combined
                    if contains_result_marker and observer is not None and not result_detected.is_set():
                        result_detected.set()
                        observer.emit(
                            phase=phase,
                            event_type="executor.result.detected",
                            status="completed",
                            payload={
                                "stream": stream_name,
                                "chunk_index": chunk_index,
                            },
                        )
                    marker_tail = combined[-marker_window:] if marker_window else ""
                if observer is not None:
                    redacted_chunk = _suppress_benign_subprocess_output(
                        _redact_subprocess_output(chunk, secret_values=secret_values)
                    )
                    if not redacted_chunk:
                        chunk_index += 1
                        continue
                    observer.emit(
                        phase=phase,
                        event_type="executor.subprocess.output",
                        status="streaming",
                        payload={
                            "stream": stream_name,
                            "chunk_text": redacted_chunk,
                            "chunk_bytes": len(redacted_chunk.encode("utf-8")),
                            "chunk_index": chunk_index,
                            "contains_result_marker": contains_result_marker,
                        },
                    )
                chunk_index += 1
        finally:
            try:
                stream.close()
            except Exception:
                pass

    def _store_executor_result(self, execution_id: str, payload: JsonObject) -> None:
        key = str(execution_id or "").strip()
        if not key:
            return
        stored = dict(payload)
        self._executor_recent_results[key] = stored
        if len(self._executor_recent_results) > 128:
            oldest = next(iter(self._executor_recent_results))
            self._executor_recent_results.pop(oldest, None)
        path = self._executor_status_path(key)
        try:
            path.parent.mkdir(parents=True, exist_ok=True)
            tmp_path = path.with_suffix(path.suffix + ".tmp")
            tmp_path.write_text(json.dumps(stored, ensure_ascii=True, sort_keys=True), encoding="utf-8")
            tmp_path.replace(path)
        except Exception as exc:
            logger.warning("executor status persist failed execution_id=%s path=%s error=%s", key, path, exc)

    def _executor_status_path(self, execution_id: str) -> Path:
        return (
            Path(self._config.hermes_run_root).expanduser()
            / "_executions"
            / f"{self._native_execution_slug(execution_id)}.json"
        ).resolve()

    def _executor_result_payload(self, result: HermesExecutionResult) -> JsonObject:
        return {
            "ok": bool(result.ok),
            "message": result.message,
            "provider": result.provider,
            "raw": dict(result.raw),
        }

    def _executor_final_status(
        self,
        task: RunnerTaskRequest,
        result: HermesExecutionResult,
        *,
        execution_id: str = "",
        status: str | None = None,
        workspace_root: str = "",
        session_id: str = "",
    ) -> JsonObject:
        raw = _json_object_or_empty(result.raw)
        return {
            "execution_id": execution_id or task.execution_id or "",
            "operation_id": task.operation_id or "",
            "trace_id": task.trace_id or "",
            "workflow_id": task.workflow_id or "",
            "phase": self._execution_phase(task),
            "status": status or ("completed" if result.ok else "failed"),
            "workspace_root": workspace_root or self._config.hermes_computer_root,
            "session_id": session_id,
            "termination_reason": _string_or_json(raw.get("termination_reason")),
            "completion_verdict": _string_or_json(raw.get("completion_verdict")),
            "message": result.message,
            "result": self._executor_result_payload(result),
        }

    def _execute_native_workflow_task_request(
        self,
        task: RunnerTaskRequest,
        tool_policy: ToolPolicy,
        *,
        observer: ObservationEmitter | None = None,
        allow_partial_recovery: bool = True,
        max_iterations_override: int | None = None,
    ) -> HermesExecutionResult:
        if not self._runtime_has_credentials():
            return HermesExecutionResult(
                ok=False,
                message="Hermes OpenAI runtime selected but RSI_OPENAI_API_KEY / OPENAI_API_KEY is not configured.",
                provider="hermes-native-executor",
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )
        if not self._contract_status.ok:
            return HermesExecutionResult(
                ok=False,
                message="Hermes adapter contract failed: " + "; ".join(self._contract_status.errors),
                provider="hermes-native-executor",
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "failure_class": "hermes_contract_failed",
                    "runner_diagnostics": {
                        "failure_kind": "hermes_contract_failed",
                        "termination_reason": "hermes_contract_failed",
                        "contract_errors": list(self._contract_status.errors),
                    },
                },
            )
        if not self._session_manager.available:
            return HermesExecutionResult(
                ok=False,
                message="Hermes persistent session runtime is unavailable.",
                provider="hermes-native-executor",
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )
        if preflight := self._preflight_tool_policy_failure(task, tool_policy):
            return preflight

        context = self._session_manager.prepare(task)
        tracker = MemoryTracker()
        execution_phase = self._execution_phase(task)
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="session.prepared",
                status="completed",
                payload={
                    "session_id": context.session_id,
                    "parent_session_id": context.parent_session_id,
                },
            )
            observer.emit(
                phase=execution_phase,
                event_type="hermes.config",
                status=self._hermes_config_parity_status(),
                payload={
                    "skills_dir": self._session_manager.skills_dir,
                    "config_parity_error": self._hermes_config_parity_error(),
                },
            )
            observer.emit(
                phase=execution_phase,
                event_type="skills.sync",
                status=self._session_manager.bundled_skills_sync_status,
                payload={
                    "bundled_skills_available": self._session_manager.bundled_skills_available,
                    "sync_error": self._session_manager.bundled_skills_sync_error,
                },
            )
            observer.emit(phase=execution_phase, event_type="phase.started", status="running")

        workspace_root = self._stage_native_executor_workspace(task, observer=observer)
        if execution_phase in {"main", "render", "deliver"}:
            task = replace(task, artifact_destination=str(self._native_artifact_destination(task)))
        self._stage_task_context(context.session_id, task, tool_policy)

        agentic_mcp_registration = TaskScopedMCPRegistration()
        try:
            agentic_mcp_registration = self._mcp_adapter.plan_task_servers(task)
        except RuntimeError as exc:
            if observer is not None:
                observer.emit(
                    phase=execution_phase,
                    event_type="mcp.registration",
                    status="failed",
                    payload={"error": str(exc)},
                )
            return HermesExecutionResult(
                ok=False,
                message=str(exc),
                provider="hermes-native-executor",
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "failure_class": "runner_non_ok",
                    "runner_diagnostics": self._runner_diagnostics(
                        tool_policy,
                        failure_kind="agentic_mcp_registration_failed",
                        provider_error_message=str(exc),
                        termination_reason="agentic_mcp_registration_failed",
                        session_ready_issues=self._session_manager.ready_issues,
                        repair_attempted=False,
                        repair_succeeded=False,
                    ),
                },
            )
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="mcp.registration",
                status="planned",
                payload={
                    "enabled": agentic_mcp_registration.enabled,
                    "server_names": agentic_mcp_registration.server_names,
                    "toolsets": agentic_mcp_registration.enabled_toolsets,
                    "registration_mode": "worker_subprocess",
                },
            )

        effective_task_timeout = self._effective_task_timeout(task)
        configured_max_iterations = max_iterations_override if max_iterations_override and max_iterations_override > 0 else self._max_iterations
        toolsets = self._native_toolsets_for_task(task)
        if execution_phase == "render":
            toolsets = [
                item
                for item in toolsets
                if item not in {HERMES_GOVERNED_READONLY_TOOLSET, HERMES_GOVERNED_WORKSPACE_TOOLSET}
            ]
        toolsets = normalize_tool_names([*toolsets, *agentic_mcp_registration.enabled_toolsets])
        phase_contract = self._native_executor_phase_contract(task, toolsets)
        missing_phase_toolsets = _string_list_or_empty(phase_contract.get("missing_required_toolsets"))
        if missing_phase_toolsets:
            message = "Native Hermes phase contract failed; missing required toolset(s): " + ", ".join(missing_phase_toolsets)
            return HermesExecutionResult(
                ok=False,
                message=message,
                provider="hermes-native-executor",
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "failure_class": "hermes_phase_contract_failed",
                    "phase_contract": phase_contract,
                    "runner_diagnostics": self._runner_diagnostics(
                        tool_policy,
                        failure_kind="hermes_phase_contract_failed",
                        provider_error_message=message,
                        termination_reason="hermes_phase_contract_failed",
                        session_ready_issues=self._session_manager.ready_issues,
                        repair_attempted=False,
                        repair_succeeded=False,
                    ),
                    "termination_reason": "hermes_phase_contract_failed",
                },
            )
        skill_diagnostics: JsonObject = {
            "requested_skills": list(task.requested_skills),
            "resolved_skills": list(task.requested_skills),
            "missing_skills": [],
            "skill_injection_mode": "native_preload" if task.requested_skills else "none",
        }
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="skills.expanded",
                status="completed",
                payload=skill_diagnostics,
            )

        execution_id = observer.execution_id if observer is not None else first_non_empty(task.execution_id, context.session_id)
        request_dir = self._native_run_dir(task)
        request_file = request_dir / "request.json"
        result_file = request_dir / "result.json"
        request_payload = self._native_executor_request_payload(
            task,
            context,
            toolsets=toolsets,
            task_scoped_mcp_registration=agentic_mcp_registration,
            max_iterations=configured_max_iterations,
            workdir=workspace_root,
            result_path=result_file,
            phase_contract=phase_contract,
        )
        request_file.write_text(json.dumps(request_payload, indent=2, sort_keys=True), encoding="utf-8")
        worker_cmd = [sys.executable, "-m", "rsi_runner.hermes_executor_worker", str(request_file)]
        env_copy = os.environ.copy()
        secret_values = _sensitive_env_values(env_copy)
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="model.call.started",
                status="running",
                payload={
                    "engine": "hermes_aiagent_subprocess",
                    "toolsets": toolsets,
                    "workspace_root": str(workspace_root),
                },
            )
            observer.emit(
                phase=execution_phase,
                event_type="executor.subprocess.started",
                status="running",
                payload={
                    "cmd": worker_cmd[:-1] + ["[request.json]"],
                    "workspace_root": str(workspace_root),
                },
            )
        completed_stdout = ""
        completed_stderr = ""
        completed_returncode = -1
        timed_out = False
        cancelled = False
        process = subprocess.Popen(
            worker_cmd,
            cwd=str(workspace_root),
            env=env_copy,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        stdout_chunks: list[str] = []
        stderr_chunks: list[str] = []
        result_detected = threading.Event()
        reader_threads: list[threading.Thread] = []
        if process.stdout is not None:
            thread = threading.Thread(
                target=self._read_native_executor_stream,
                args=(
                    process.stdout,
                ),
                kwargs={
                    "stream_name": "stdout",
                    "phase": execution_phase,
                    "observer": observer,
                    "chunk_store": stdout_chunks,
                    "secret_values": secret_values,
                    "result_detected": result_detected,
                },
                daemon=True,
            )
            thread.start()
            reader_threads.append(thread)
        if process.stderr is not None:
            thread = threading.Thread(
                target=self._read_native_executor_stream,
                args=(
                    process.stderr,
                ),
                kwargs={
                    "stream_name": "stderr",
                    "phase": execution_phase,
                    "observer": observer,
                    "chunk_store": stderr_chunks,
                    "secret_values": secret_values,
                    "result_detected": result_detected,
                },
                daemon=True,
            )
            thread.start()
            reader_threads.append(thread)
        with self._executor_process_lock:
            self._executor_processes[execution_id] = process
        started_at = time.monotonic()
        try:
            while True:
                with self._executor_process_lock:
                    if execution_id in self._executor_cancel_requests:
                        cancelled = True
                        break
                if process.poll() is not None:
                    break
                if time.monotonic()-started_at > (effective_task_timeout + 5):
                    timed_out = True
                    break
                time.sleep(0.25)
            if cancelled or timed_out:
                try:
                    process.terminate()
                except OSError:
                    pass
                try:
                    process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    try:
                        process.kill()
                    except OSError:
                        pass
                    try:
                        process.wait(timeout=5)
                    except subprocess.TimeoutExpired:
                        pass
            else:
                process.wait()
            completed_returncode = int(process.returncode or 0)
        finally:
            with self._executor_process_lock:
                self._executor_processes.pop(execution_id, None)
                self._executor_cancel_requests.discard(execution_id)
            for thread in reader_threads:
                thread.join(timeout=5)
            completed_stdout = _suppress_benign_subprocess_output("".join(stdout_chunks))
            completed_stderr = _suppress_benign_subprocess_output("".join(stderr_chunks))

        parsed_result: JsonObject = {}
        parse_error = ""
        result_file_present = result_file.exists()
        if result_file_present:
            try:
                parsed_result = self._load_native_executor_result_file(result_file)
                if observer is not None:
                    result_file_bytes: int | None = None
                    try:
                        result_file_bytes = result_file.stat().st_size
                    except OSError:
                        result_file_bytes = None
                    observer.emit(
                        phase=execution_phase,
                        event_type="executor.result.persisted",
                        status="completed",
                        payload={"path": str(result_file), "bytes": result_file_bytes},
                    )
                    observer.emit(
                        phase=execution_phase,
                        event_type="executor.result.loaded",
                        status="completed",
                        payload={"path": str(result_file)},
                    )
            except (OSError, ValueError, json.JSONDecodeError) as exc:
                parse_error = str(exc)
        if not parsed_result and not parse_error:
            try:
                parsed_result = self._parse_native_executor_stdout(completed_stdout)
            except (ValueError, json.JSONDecodeError) as exc:
                parse_error = str(exc)
        redacted_completed_stderr = _redact_subprocess_output(completed_stderr, secret_values=secret_values)
        artifact_tool_events = _json_object_list(parsed_result.get("artifact_tool_events"))
        native_artifact_paths = _native_artifact_paths_from_events(artifact_tool_events)
        artifact_event_details = _native_artifact_event_details(artifact_tool_events)
        if observer is not None:
            for event in artifact_tool_events:
                observer.emit(
                    phase=execution_phase,
                    event_type=_string_or_json(event.get("event_type")),
                    status=_string_or_json(event.get("status")),
                    payload=_json_object_or_empty(event.get("payload")),
                )

        if (timed_out or cancelled) and not parsed_result:
            finalized = self._session_manager.finalize(context, tracker)
            lifecycle_events = self._adapter.lifecycle_events(context.session_id)
            observed = self._observability_metadata(None, task, tracker, skill_diagnostics=skill_diagnostics, observer=observer)
            cleanup_status = "not_needed" if not agentic_mcp_registration.enabled else "worker_unavailable"
            cleanup_errors: list[str] = []
            if observer is not None:
                observer.emit(
                    phase=execution_phase,
                    event_type="model.call.completed",
                    status="cancelled" if cancelled else "task_timeout",
                    payload={
                        "engine": "hermes_aiagent_subprocess",
                        "returncode": completed_returncode,
                        "stderr": _truncate_string(redacted_completed_stderr, 4000),
                    },
                )
                observer.emit(
                    phase=execution_phase,
                    event_type="mcp.cleanup",
                    status=cleanup_status,
                    payload={"errors": cleanup_errors},
                )
            termination_reason = "cancelled" if cancelled else "task_timeout"
            stop_meta: JsonObject = {"termination_reason": termination_reason, "last_activity": {}}
            if observer is not None:
                observer.emit(
                    phase=execution_phase,
                    event_type="executor.subprocess.completed",
                    status=termination_reason,
                    payload={
                        "engine": "hermes_aiagent_subprocess",
                        "returncode": completed_returncode,
                        "termination_reason": termination_reason,
                        "parsed_result_ok": False,
                        "workspace_root": str(workspace_root),
                        "artifact_output_dir": str(self._native_artifact_destination(task)),
                    },
                )
            if not cancelled and allow_partial_recovery and self._workflow_partial_completion_eligible(task):
                return self._finalize_partial_completion(
                    task,
                    tool_policy,
                    finalized,
                    observed,
                    stop_meta,
                    lifecycle_events,
                    termination_reason=termination_reason,
                    observer=observer,
                )
            result = HermesExecutionResult(
                ok=False,
                message="Hermes native executor was cancelled." if cancelled else f"Hermes native executor timed out after {effective_task_timeout}s.",
                provider="hermes-native-executor",
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    **finalized,
                    **observed,
                    **self._workflow_evidence_raw(task, observed, termination_reason),
                    "failure_class": "runner_cancelled" if cancelled else "runner_transport_timeout",
                    "runner_diagnostics": self._runner_diagnostics(
                        tool_policy,
                        failure_kind="cancelled" if cancelled else "transport_timeout",
                        provider_error_message="Hermes native executor was cancelled." if cancelled else f"Hermes native executor timed out after {effective_task_timeout}s.",
                        timeout_kind=termination_reason,
                        termination_reason=termination_reason,
                        session_ready_issues=self._session_manager.ready_issues,
                        repair_attempted=False,
                        repair_succeeded=False,
                        observed=observed,
                    ),
                    "lifecycle_events": lifecycle_events,
                    "termination_reason": termination_reason,
                    "native_executor_mode": "subprocess",
                },
            )
            result.raw = self._attach_agentic_mcp_diagnostics(
                _json_object_or_empty(result.raw),
                agentic_mcp_registration,
                cleanup_status=cleanup_status,
                cleanup_errors=cleanup_errors,
            )
            self._store_executor_result(
                execution_id,
                {
                    "execution_id": execution_id,
                    "status": "cancelled" if cancelled else "failed",
                    "operation_id": task.operation_id,
                    "trace_id": task.trace_id,
                    "workflow_id": task.workflow_id,
                    "phase": self._execution_phase(task),
                    "workspace_root": str(workspace_root),
                    "session_id": context.session_id,
                    "termination_reason": termination_reason,
                    "completion_verdict": "",
                    "message": result.message,
                    "result": self._executor_result_payload(result),
                },
            )
            return result
        finalized = self._session_manager.finalize(context, tracker)
        lifecycle_events = self._adapter.lifecycle_events(context.session_id)
        observed = self._observability_metadata(None, task, tracker, skill_diagnostics=skill_diagnostics, observer=observer)
        parsed_result_loaded = bool(parsed_result) and not parse_error
        completion_meta = self._native_executor_completion_meta(parsed_result, configured_max_iterations) if parsed_result_loaded else {
            "termination_reason": "exception",
            "completion_verdict": "",
            "max_iterations_reached": False,
            "native_result_completed": False,
            "native_result_partial": False,
            "native_result_interrupted": False,
            "native_result_api_calls": 0,
        }
        cleanup_status = _string_or_json(parsed_result.get("mcp_cleanup_status"))
        if not cleanup_status:
            cleanup_status = "not_needed" if not agentic_mcp_registration.enabled else "worker_unreported"
        cleanup_errors = _string_list(parsed_result.get("mcp_cleanup_errors"))
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="model.call.completed",
                status="completed" if parsed_result_loaded else "failed",
                payload={
                    "engine": "hermes_aiagent_subprocess",
                    "returncode": completed_returncode,
                    "stderr": _truncate_string(redacted_completed_stderr, 4000),
                    "termination_reason": completion_meta["termination_reason"],
                    "completion_verdict": completion_meta["completion_verdict"],
                },
            )
            observer.emit(
                phase=execution_phase,
                event_type="executor.subprocess.completed",
                status="completed" if parsed_result_loaded else "failed",
                payload={
                    "engine": "hermes_aiagent_subprocess",
                    "returncode": completed_returncode,
                    "termination_reason": completion_meta["termination_reason"] if parsed_result_loaded else "",
                    "parsed_result_ok": parsed_result_loaded,
                    "workspace_root": str(workspace_root),
                    "artifact_output_dir": str(self._native_artifact_destination(task)),
                },
            )
        base_raw = {
            **self._base_raw(prompt=task.prompt, system_message=task.system_message),
            **finalized,
            **observed,
            "task_timeout_seconds": effective_task_timeout,
            "max_iterations": configured_max_iterations,
            "tool_policy_mode": tool_policy.mode,
            "tool_allowlist_effective": tool_policy.effective,
            "tool_transport_allowlist_effective": tool_policy.transport_effective,
            "tool_transport_map": tool_policy.custom_tool_transport_map,
            "blocked_tool_names": tool_policy.blocked,
            "lifecycle_events": lifecycle_events,
            "native_executor_mode": "subprocess",
            "native_executor_workspace_root": str(workspace_root),
            "native_executor_toolsets": toolsets,
            "native_executor_returncode": completed_returncode,
            "native_executor_stderr": _truncate_string(redacted_completed_stderr, 4000),
            "native_executor_contract_status": _json_object_or_empty(parsed_result.get("contract_status")),
            "artifact_tool_events": artifact_tool_events,
            "native_artifact_paths": native_artifact_paths,
            "phase_contract": phase_contract,
            **artifact_event_details,
        }
        if parse_error or not parsed_result or not _bool_or_false(parsed_result.get("ok")):
            failure_class = _string_or_json(parsed_result.get("failure_class")) or "runner_non_ok"
            failure_kind = "hermes_contract_failed" if failure_class == "hermes_contract_failed" else "execution_error"
            message = first_non_empty(
                _string_or_json(parsed_result.get("error")),
                "native Hermes executor exited without a result file" if not result_file_present else "",
                parse_error,
                _truncate_string(redacted_completed_stderr, 4000),
                "Hermes native executor failed.",
            )
            result = HermesExecutionResult(
                ok=False,
                message=message,
                provider="hermes-native-executor",
                raw={
                    **base_raw,
                    **self._workflow_evidence_raw(task, observed, "exception"),
                    "failure_class": failure_class,
                    "runner_diagnostics": self._runner_diagnostics(
                        tool_policy,
                        failure_kind=failure_kind,
                        provider_error_message=message,
                        termination_reason=failure_kind,
                        session_ready_issues=self._session_manager.ready_issues,
                        repair_attempted=False,
                        repair_succeeded=False,
                        observed=observed,
                    ),
                    "artifact_failure_reason": message if execution_phase == "render" else "",
                    "requested_skills": list(task.requested_skills),
                },
            )
        else:
            result = HermesExecutionResult(
                ok=True,
                message=_string_or_json(parsed_result.get("response")),
                provider="hermes-native-executor",
                raw={
                    **base_raw,
                    **self._workflow_evidence_raw(task, observed, completion_meta["termination_reason"]),
                    "runner_diagnostics": {
                        **observed,
                        "termination_reason": completion_meta["termination_reason"],
                        "max_iterations_reached": _bool_or_false(completion_meta.get("max_iterations_reached")),
                        "completion_verdict": completion_meta["completion_verdict"],
                        "native_result_completed": _bool_or_false(completion_meta.get("native_result_completed")),
                        "native_result_partial": _bool_or_false(completion_meta.get("native_result_partial")),
                        "native_result_interrupted": _bool_or_false(completion_meta.get("native_result_interrupted")),
                        "native_result_api_calls": completion_meta["native_result_api_calls"],
                    },
                    **completion_meta,
                    "harness_profile_id": task.harness_profile_id,
                    "effective_overlay_version": task.harness_overlay_version,
                },
            )
        result.raw = self._attach_agentic_mcp_diagnostics(
            _json_object_or_empty(result.raw),
            agentic_mcp_registration,
            cleanup_status=cleanup_status,
            cleanup_errors=cleanup_errors,
        )
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="mcp.cleanup",
                status=cleanup_status,
                payload={"errors": cleanup_errors},
            )
            observer.emit(
                phase=execution_phase,
                event_type="phase.completed",
                status="completed" if result.ok else "failed",
                payload={
                    "termination_reason": _string_or_json(_json_object_or_empty(result.raw).get("termination_reason")),
                    "completion_verdict": _string_or_json(_json_object_or_empty(result.raw).get("completion_verdict")),
                    "artifact_failure_reason": _string_or_json(_json_object_or_empty(result.raw).get("artifact_failure_reason")),
                    "artifact_output_dir": _string_or_json(_json_object_or_empty(result.raw).get("artifact_output_dir")),
                    "artifact_persistence_tool": _string_or_json(_json_object_or_empty(result.raw).get("artifact_persistence_tool")),
                    "requested_output_path": _string_or_json(_json_object_or_empty(result.raw).get("requested_output_path")),
                    "saved_paths": _string_list_or_empty(_json_object_or_empty(result.raw).get("saved_paths")),
                    "save_error": _string_or_json(_json_object_or_empty(result.raw).get("save_error")),
                    "requested_skills": list(task.requested_skills),
                    "resolved_skills": list(skill_diagnostics.get("resolved_skills") or []),
                },
            )
        self._store_executor_result(
            execution_id,
            {
                "execution_id": execution_id,
                "status": "completed" if result.ok else "failed",
                "operation_id": task.operation_id,
                "trace_id": task.trace_id,
                "workflow_id": task.workflow_id,
                "phase": self._execution_phase(task),
                "workspace_root": str(workspace_root),
                "session_id": context.session_id,
                "termination_reason": _string_or_json(_json_object_or_empty(result.raw).get("termination_reason")),
                "completion_verdict": _string_or_json(_json_object_or_empty(result.raw).get("completion_verdict")),
                "message": result.message,
                "result": self._executor_result_payload(result),
            },
        )
        return result

    def _runner_diagnostics(
        self,
        tool_policy: ToolPolicy,
        *,
        failure_kind: str,
        provider_status_code: int | None = None,
        provider_error_param: str | None = None,
        provider_error_code: str | None = None,
        provider_error_message: str | None = None,
        invalid_tool_names: list[str] | None = None,
        timeout_kind: str | None = None,
        termination_reason: str | None = None,
        activity: JsonObject | None = None,
        max_iterations_reached: bool | None = None,
        session_ready_issues: list[str] | None = None,
        repair_attempted: bool | None = None,
        repair_succeeded: bool | None = None,
        observed: JsonObject | None = None,
    ) -> JsonObject:
        diagnostics: JsonObject = {
            "failure_kind": failure_kind,
            "tool_allowlist_effective": list(tool_policy.effective),
            "tool_transport_allowlist_effective": list(tool_policy.transport_effective),
        }
        latest_activity = _json_object_or_empty(activity)
        if tool_policy.custom_tool_transport_map:
            diagnostics["tool_transport_map"] = dict(tool_policy.custom_tool_transport_map)
        if provider_status_code is not None:
            diagnostics["provider_status_code"] = provider_status_code
        if provider_error_param:
            diagnostics["provider_error_param"] = provider_error_param
        if provider_error_code:
            diagnostics["provider_error_code"] = provider_error_code
        if provider_error_message:
            diagnostics["provider_error_message"] = provider_error_message
        if invalid_tool_names:
            diagnostics["invalid_tool_names"] = normalize_tool_names(invalid_tool_names)
        if timeout_kind:
            diagnostics["timeout_kind"] = timeout_kind
        if termination_reason:
            diagnostics["termination_reason"] = termination_reason
        if "last_activity_desc" in latest_activity:
            diagnostics["last_activity_desc"] = string_from_map(latest_activity, "last_activity_desc")
        if "current_tool" in latest_activity:
            diagnostics["current_tool"] = string_from_map(latest_activity, "current_tool")
        if "api_call_count" in latest_activity:
            diagnostics["api_call_count"] = latest_activity.get("api_call_count")
        if "budget_used" in latest_activity:
            diagnostics["budget_used"] = latest_activity.get("budget_used")
        if "budget_max" in latest_activity:
            diagnostics["budget_max"] = latest_activity.get("budget_max")
        if max_iterations_reached is not None:
            diagnostics["max_iterations_reached"] = max_iterations_reached
        ready_issues = list(session_ready_issues or [])
        if ready_issues:
            diagnostics["session_ready_issues"] = ready_issues
        if repair_attempted is not None:
            diagnostics["repair_attempted"] = repair_attempted
        if repair_succeeded is not None:
            diagnostics["repair_succeeded"] = repair_succeeded
        for key, value in (observed or {}).items():
            diagnostics[key] = value
        return diagnostics

    def _attach_agentic_mcp_diagnostics(
        self,
        raw: JsonObject,
        registration: TaskScopedMCPRegistration,
        *,
        cleanup_status: str,
        cleanup_errors: list[str] | None = None,
    ) -> JsonObject:
        diagnostics = {
            "agentic_mcp_enabled": registration.enabled,
            "agentic_mcp_server_names": registration.server_names,
            "agentic_mcp_toolsets": registration.enabled_toolsets,
            "agentic_mcp_cleanup_status": cleanup_status,
        }
        if cleanup_errors:
            diagnostics["agentic_mcp_cleanup_errors"] = list(cleanup_errors)
        merged = dict(raw)
        merged.update(diagnostics)
        merged["runner_diagnostics"] = {
            **_json_object_or_empty(merged.get("runner_diagnostics")),
            **diagnostics,
        }
        return merged

    def _candidate_read_surfaces_for_task(self, task: RunnerTaskRequest) -> list[JsonObject]:
        surfaces: list[JsonObject] = []
        seen: set[str] = set()
        for item in task.context_refs:
            if str(item.get("kind", "")).strip() != "candidate_read_surface":
                continue
            candidate = {
                "channel_id": str(item.get("channel_id", "")).strip(),
                "thread_ts": str(item.get("thread_ts", "")).strip(),
                "ref": str(item.get("ref", "")).strip(),
                "source": str(item.get("source", "")).strip(),
            }
            if not candidate["channel_id"] and not candidate["thread_ts"] and not candidate["ref"]:
                continue
            encoded = json.dumps(candidate, sort_keys=True)
            if encoded in seen:
                continue
            seen.add(encoded)
            surfaces.append(candidate)
        if task.channel_id:
            fallback = {
                "channel_id": task.channel_id,
                "thread_ts": task.thread_ts or "",
                "ref": "",
                "source": "task_binding",
            }
            encoded = json.dumps(fallback, sort_keys=True)
            if encoded not in seen:
                surfaces.insert(0, fallback)
        return surfaces

    def _observability_metadata(
        self,
        agent: Any | None,
        task: RunnerTaskRequest,
        tracker: Any | None = None,
        *,
        skill_diagnostics: JsonObject | None = None,
        observer: ObservationEmitter | None = None,
    ) -> JsonObject:
        observed: JsonObject = {
            "candidate_read_surfaces": self._candidate_read_surfaces_for_task(task),
            "selected_context_surfaces": [],
            "memory_warnings": [],
            "skills_dir": self._session_manager.skills_dir,
            "bundled_skills_available": self._session_manager.bundled_skills_available,
            "bundled_skills_sync_status": self._session_manager.bundled_skills_sync_status,
        }
        if self._session_manager.bundled_skills_sync_error:
            observed["bundled_skills_sync_error"] = self._session_manager.bundled_skills_sync_error
        binding = getattr(agent, "_rsi_readonly_tool_binding", None) if agent is not None else None
        diagnostics = getattr(binding, "diagnostics", None)
        if callable(diagnostics):
            payload = diagnostics()
            if isinstance(payload, dict):
                observed.update(payload)
        if tracker is not None and hasattr(tracker, "warnings"):
            observed["memory_warnings"] = list(getattr(tracker, "warnings", []) or [])
        if skill_diagnostics:
            observed.update(skill_diagnostics)
        if observer is not None:
            observed.update(observer.diagnostics())
        return observed

    def _preflight_tool_policy_failure(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> HermesExecutionResult | None:
        _, _, invalid_tool_names = _transport_tool_policy(tool_policy.custom_tools, tool_policy.memory_tools)
        if not invalid_tool_names:
            return None
        observed = self._observability_metadata(None, task)
        diagnostics = self._runner_diagnostics(
            tool_policy,
            failure_kind="invalid_request",
            provider_error_param="tools[0].name",
            provider_error_code="invalid_value",
            provider_error_message="Runner tool schema failed provider-safe tool-name preflight validation.",
            invalid_tool_names=invalid_tool_names,
            repair_attempted=False,
            repair_succeeded=False,
            observed=observed,
        )
        return HermesExecutionResult(
            ok=False,
            message="Runner tool schema failed preflight validation for provider-safe tool names.",
            provider=self._backend,
            raw={
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                "tool_policy_mode": tool_policy.mode,
                "tool_allowlist_effective": tool_policy.effective,
                "tool_transport_allowlist_effective": tool_policy.transport_effective,
                "tool_transport_map": tool_policy.custom_tool_transport_map,
                "blocked_tool_names": tool_policy.blocked,
                **observed,
                "failure_class": "runner_invalid_request",
                "runner_diagnostics": diagnostics,
                "repair_attempted": False,
                "repair_succeeded": False,
            },
        )

    def _provider_invalid_request_diagnostics(self, message: str, tool_policy: ToolPolicy) -> JsonObject | None:
        text = str(message or "").strip()
        if not text:
            return None
        lower = text.lower()
        if "invalid_request_error" not in lower and "tools[0].name" not in lower and "invalid 'tools[" not in lower and 'invalid "tools[' not in lower:
            return None

        status_code = 0
        status_match = re.search(r"(?:error code|status code|status)\s*[:= -]+\s*(\d{3})", text, flags=re.IGNORECASE)
        if status_match:
            try:
                status_code = int(status_match.group(1))
            except ValueError:
                status_code = 0
        param_match = re.search(r"['\"]param['\"]\s*:\s*(['\"])(.*?)\1", text)
        code_match = re.search(r"['\"]code['\"]\s*:\s*(['\"])(.*?)\1", text)
        message_match = re.search(r"['\"]message['\"]\s*:\s*(['\"])(.*?)\1", text)
        provider_error_param = param_match.group(2).strip() if param_match else ""
        provider_error_code = code_match.group(2).strip() if code_match else ""
        provider_error_message = message_match.group(2).strip() if message_match else text[:8000]
        if status_code not in {0, 400, 422} and provider_error_param == "":
            return None

        invalid_tool_names: list[str] = []
        if provider_error_param == "tools[0].name" or "tools[0].name" in lower:
            invalid_tool_names = [name for name in tool_policy.custom_tools if "." in name]
            if not invalid_tool_names:
                invalid_tool_names = list(tool_policy.custom_tools)
        diagnostics = self._runner_diagnostics(
            tool_policy,
            failure_kind="invalid_request",
            provider_status_code=status_code or None,
            provider_error_param=provider_error_param or None,
            provider_error_code=provider_error_code or None,
            provider_error_message=provider_error_message,
            invalid_tool_names=invalid_tool_names,
            repair_attempted=False,
            repair_succeeded=False,
        )
        return diagnostics

    def _default_policy_allowlist(self, execution_mode: str) -> list[str]:
        permitted = set(READ_ONLY_HONCHO_TOOLS)
        if self._config.tool_gateway_base_url:
            permitted.update(READ_ONLY_RSI_TOOL_NAMES)
        if self._role == "proposal" and self._config.hermes_native_governed_tools_enabled and execution_mode.strip().lower() == "diagnose":
            permitted.update(NATIVE_HERMES_DIAGNOSE_TOOLS)
        if self._role == "proposal" and execution_mode.strip().lower() == "implement":
            permitted.update(WORKSPACE_RSI_TOOL_NAMES)
        return sorted(permitted)

    def _resolve_tool_policy(self, task: RunnerTaskRequest) -> ToolPolicy:
        requested = normalize_tool_names([*task.allowed_tools, *task.tool_allowlist])
        execution_mode = (task.execution_mode or "").strip().lower()
        execution_phase = self._execution_phase(task)
        if task.task_type == "question_reduce":
            return ToolPolicy(
                mode="no_tools",
                requested=requested,
                effective=[],
                blocked=requested,
                memory_tools=[],
                custom_tools=[],
                transport_effective=[],
                custom_tool_transport_map={},
            )
        permitted = set(self._default_policy_allowlist(execution_mode=execution_mode))
        if self._config.tool_gateway_base_url and (task.workspace_id or task.attempt_id):
            permitted.update(READ_ONLY_WORKSPACE_RSI_TOOL_NAMES)
        if self._config.tool_gateway_base_url and execution_phase == "render":
            permitted.update(READ_ONLY_WORKSPACE_RSI_TOOL_NAMES)
            permitted.update(ARTIFACT_RENDER_RSI_TOOL_NAMES)
        fallback_allowlist = sorted(permitted)
        if task.task_type in {"question_gather", "question_expand"} and requested:
            fallback_allowlist = requested
        effective = normalize_tool_names(requested or fallback_allowlist)
        effective = [name for name in effective if name in permitted]
        if execution_phase == "investigate":
            effective = [name for name in effective if name not in {"slack.reply", "slack.upload_file"}]
        elif execution_phase == "render":
            effective = [name for name in effective if name in ARTIFACT_RENDER_RSI_TOOL_NAMES]
        elif execution_phase == "deliver":
            effective = ["slack.upload_file"] if "slack.upload_file" in requested else []
        blocked = [name for name in requested if name not in permitted and name not in {"slack.history", "slack.search", "slack.reply"}]
        memory_tools = sorted([name for name in effective if name in READ_ONLY_HONCHO_TOOLS])
        custom_tools = sorted([name for name in effective if name in IMPLEMENT_RSI_TOOL_NAMES])
        transport_effective, custom_tool_transport_map, _ = _transport_tool_policy(custom_tools, memory_tools)
        mode = self._tool_policy_mode
        if self._role == "proposal" and execution_mode == "implement":
            mode = "governed_workspace"
        return ToolPolicy(
            mode=mode,
            requested=requested,
            effective=effective,
            blocked=blocked,
            memory_tools=memory_tools,
            custom_tools=custom_tools,
            transport_effective=transport_effective,
            custom_tool_transport_map=custom_tool_transport_map,
        )

    def _effective_task_timeout(self, task: RunnerTaskRequest) -> int:
        requested = int(task.timeout_seconds or 0)
        candidates = [self._default_task_timeout_seconds]
        if requested > 0:
            candidates.append(requested)
        if self._transport_timeout_seconds > 5:
            candidates.append(self._transport_timeout_seconds - 5)
        return max(1, min(value for value in candidates if value > 0))

    def _effective_inactivity_timeout(self, effective_task_timeout: int) -> int:
        return max(1, min(self._default_inactivity_timeout_seconds, effective_task_timeout))

    def _budget_exhausted(self, activity: JsonObject | None) -> bool:
        latest_activity = _json_object_or_empty(activity)
        budget_used = latest_activity.get("budget_used", 0)
        budget_max = latest_activity.get("budget_max", 0)
        return bool(budget_max and budget_used >= budget_max)

    def _workflow_partial_completion_eligible(self, task: RunnerTaskRequest) -> bool:
        if task.task_type != "workflow":
            return False
        if self._role not in {"prod", "proactive"}:
            return False
        return bool((task.channel_id or "").strip() and (task.thread_ts or "").strip())

    def _partial_completion_finalization_reserve_seconds(self, task_timeout_seconds: int) -> int:
        return min(180, max(10, task_timeout_seconds - 30))

    def _partial_completion_reasoning_timeout_seconds(self, task: RunnerTaskRequest, task_timeout_seconds: int) -> int:
        if not self._workflow_partial_completion_eligible(task):
            return task_timeout_seconds
        reserve = self._partial_completion_finalization_reserve_seconds(task_timeout_seconds)
        reasoning_timeout_seconds = task_timeout_seconds - reserve
        if reasoning_timeout_seconds <= 0:
            return task_timeout_seconds
        return reasoning_timeout_seconds

    def _partial_completion_timeout_seconds(
        self,
        task: RunnerTaskRequest,
        termination_reason: str,
        stop_meta: JsonObject,
    ) -> int:
        stopped_after_seconds = int(stop_meta.get("stopped_after_seconds") or 0)
        remaining_task_budget = self._effective_task_timeout(task) - stopped_after_seconds
        if remaining_task_budget <= 0:
            return 0
        remaining_transport_headroom = self._transport_timeout_seconds - stopped_after_seconds - 5
        reserve = self._partial_completion_finalization_reserve_seconds(self._effective_task_timeout(task))
        timeout_seconds = min(remaining_task_budget, remaining_transport_headroom, reserve)
        if termination_reason == "iteration_budget_exhausted":
            timeout_seconds = min(timeout_seconds, reserve)
        return max(0, timeout_seconds)

    def _stage_task_context(self, session_id: str, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> None:
        query_hints = self._question_default_query_hints(task)
        self._adapter.stage_task_context(
            session_id,
            {
                "role": self._role,
                "task_type": task.task_type,
                "task_repo": task.repo,
                "task_repo_ref": task.repo_ref or "",
                "task_prompt": task.prompt,
                "trace_id": task.trace_id,
                "workflow_id": task.workflow_id,
                "operation_id": task.operation_id or "",
                "execution_id": task.execution_id or "",
                "task_channel_id": task.channel_id or "",
                "task_thread_ts": task.thread_ts or "",
                "proposal_id": task.session_scope_id if (task.session_scope_kind or "").strip() == "proposal_candidate" else "",
                "attempt_id": task.attempt_id,
                "workspace_id": task.workspace_id,
                "execution_mode": task.execution_mode or "",
                "execution_phase": self._execution_phase(task),
                "artifact_output_dir": _string_or_json(task.artifact_destination),
                "hermes_computer_root": self._config.hermes_computer_root,
                "hermes_run_root": self._config.hermes_run_root,
                "hermes_artifact_root": self._config.hermes_artifact_root,
                "context_summary": task.context_summary or "",
                "task_default_question": query_hints.get("default_question", ""),
                "task_repo_question": query_hints.get("repo_question", ""),
                "task_knowledge_topic": query_hints.get("knowledge_topic", ""),
                "task_knowledge_question": query_hints.get("knowledge_question", ""),
                "task_slack_history_focus": query_hints.get("slack_history_focus", ""),
                "task_slack_search_query": query_hints.get("slack_search_query", ""),
                "context_refs": task.context_refs,
                "tool_gateway_base_url": self._config.tool_gateway_base_url or "",
                "tool_timeout_seconds": 30,
                "tool_allowlist_effective": tool_policy.effective,
                "tool_transport_allowlist_effective": tool_policy.transport_effective,
                "tool_transport_map": tool_policy.custom_tool_transport_map,
                "blocked_tool_names": tool_policy.blocked,
                "session_scope_kind": task.session_scope_kind or "",
                "session_scope_id": task.session_scope_id or "",
            },
        )

    def _partial_completion_attempt_budgets(self, timeout_seconds: int) -> list[int]:
        total = max(0, int(timeout_seconds or 0))
        if total <= 0:
            return []
        return [total]

    def _merge_runtime_values(self, *lists: Any) -> list[Any]:
        merged: list[Any] = []
        seen: set[str] = set()
        for items in lists:
            if not isinstance(items, list):
                continue
            for item in items:
                encoded = json.dumps(item, ensure_ascii=True, sort_keys=True, default=str)
                if encoded in seen:
                    continue
                seen.add(encoded)
                merged.append(item)
        return merged

    def _original_task_prompt(self, task: RunnerTaskRequest) -> str:
        prompt = _string_or_json(task.prompt)
        marker = "Task prompt:\n"
        if marker in prompt:
            return prompt.rsplit(marker, 1)[1].strip()
        return prompt

    def _question_task_payload(self, task: RunnerTaskRequest) -> JsonObject:
        if task.task_type not in {"question_gather", "question_expand", "question_reduce"}:
            return {}
        try:
            parsed = json.loads(task.prompt)
        except json.JSONDecodeError:
            return {}
        if not isinstance(parsed, dict):
            return {}
        return parsed

    def _question_investigation_spec(self, task: RunnerTaskRequest) -> JsonObject:
        return _json_object_or_empty(self._question_task_payload(task).get("investigation_spec"))

    def _question_input_evidence_ledger(self, task: RunnerTaskRequest) -> JsonObject:
        return _json_object_or_empty(self._question_task_payload(task).get("evidence_ledger"))

    def _question_input_runner_diagnostics(self, task: RunnerTaskRequest) -> JsonObject:
        return _json_object_or_empty(self._question_task_payload(task).get("runner_diagnostics"))

    def _question_default_query_hints(self, task: RunnerTaskRequest) -> JsonObject:
        original_prompt = self._original_task_prompt(task)
        spec = self._question_investigation_spec(task)
        repo = first_non_empty(_string_or_json(spec.get("repo")), task.repo)
        project_key = _string_or_json(spec.get("project_key"))
        user_request = first_non_empty(_string_or_json(spec.get("user_request")), original_prompt)
        slack_query_parts = [part for part in [repo, project_key] if part]
        slack_search_query = " ".join(slack_query_parts) if slack_query_parts else user_request
        history_focus = user_request
        if user_request:
            history_focus = f"Extract the most relevant messages for answering: {user_request}"
        knowledge_topic = first_non_empty(project_key, repo, task.context_summary or "", user_request)
        knowledge_question = user_request
        if project_key:
            knowledge_question = f"What are the current goals, constraints, and expected outcomes for {project_key}?"
        return {
            "default_question": first_non_empty(user_request, original_prompt),
            "repo_question": first_non_empty(user_request, original_prompt),
            "knowledge_topic": knowledge_topic,
            "knowledge_question": first_non_empty(knowledge_question, user_request, original_prompt),
            "slack_history_focus": first_non_empty(history_focus, user_request, original_prompt),
            "slack_search_query": first_non_empty(slack_search_query, user_request, original_prompt),
        }

    def _compact_tool_calls(self, observed: JsonObject, *, limit: int = 12) -> list[JsonObject]:
        compact: list[JsonObject] = []
        for item in _json_object_list(observed.get("tool_calls"))[:limit]:
            tool_name = _string_or_json(item.get("tool_name"))
            if not tool_name:
                continue
            normalized: JsonObject = {"tool_name": tool_name}
            tool_call_id = first_non_empty(_string_or_json(item.get("tool_call_id")), _string_or_json(item.get("id")))
            if tool_call_id:
                normalized["tool_call_id"] = tool_call_id
            request_payload = _json_object_or_empty(item.get("request"))
            if request_payload:
                normalized["request"] = request_payload
            summary = _string_or_json(item.get("summary"))
            if summary:
                normalized["summary"] = summary[:400]
            status = _string_or_json(item.get("status"))
            if status:
                normalized["status"] = status
            provider_ref = _string_or_json(item.get("provider_ref"))
            if provider_ref:
                normalized["provider_ref"] = provider_ref
            started_at = _string_or_json(item.get("started_at"))
            if started_at:
                normalized["started_at"] = started_at
            completed_at = _string_or_json(item.get("completed_at"))
            if completed_at:
                normalized["completed_at"] = completed_at
            raw_artifact_refs = item.get("raw_artifact_refs")
            if isinstance(raw_artifact_refs, list) and raw_artifact_refs:
                normalized["raw_artifact_refs"] = raw_artifact_refs[:4]
            compact.append(normalized)
        return compact

    def _compact_evidence_items(self, observed: JsonObject, *, limit: int = 20) -> list[JsonObject]:
        compact: list[JsonObject] = []
        for item in _json_object_list(observed.get("evidence_items"))[:limit]:
            summary = _string_or_json(item.get("summary")) or _string_or_json(item.get("snippet"))
            source_ref = first_non_empty(
                _string_or_json(item.get("source_ref")),
                _string_or_json(item.get("permalink")),
                _string_or_json(item.get("url")),
                _string_or_json(item.get("path")),
            )
            normalized: JsonObject = {
                "kind": _string_or_json(item.get("kind")) or "evidence",
                "summary": summary[:400],
                "source_ref": source_ref,
                "tool_name": _string_or_json(item.get("tool_name")),
            }
            snippet = _string_or_json(item.get("snippet"))
            if snippet:
                normalized["snippet"] = snippet[:600]
            for key in (
                "author",
                "actor",
                "channel_id",
                "thread_ts",
                "message_ts",
                "path",
                "repo",
                "ref",
                "commit",
                "sha",
                "permalink",
                "workflow_id",
                "entry_type",
                "default_branch",
                "title",
                "state",
                "created_at",
                "merged_at",
                "committed_at",
                "url",
            ):
                value = _string_or_json(item.get(key))
                if value:
                    normalized[key] = value
            compact.append(normalized)
        return compact

    def _evidence_open_questions(self, tool_calls: list[JsonObject], evidence_items: list[JsonObject]) -> list[str]:
        questions: list[str] = []
        if not evidence_items:
            questions.append("No grounded evidence items were captured before the bounded stop.")
        for item in tool_calls:
            status = _string_or_json(item.get("status")).lower()
            if not status or status in {"completed", "complete", "ok", "success"}:
                continue
            tool_name = _string_or_json(item.get("tool_name")) or "tool"
            summary = _string_or_json(item.get("summary")) or status
            questions.append(f"{tool_name}: {summary[:240]}")
        deduped = self._merge_runtime_values(questions)
        return [str(item) for item in deduped[:6]]

    def _build_evidence_ledger(self, task: RunnerTaskRequest, observed: JsonObject, termination_reason: str) -> JsonObject:
        tool_calls = self._compact_tool_calls(observed)
        evidence_items = self._compact_evidence_items(observed)
        ledger: JsonObject = {
            "user_request": self._original_task_prompt(task),
            "reply_target": {
                "channel_id": task.channel_id or "",
                "thread_ts": task.thread_ts or "",
            },
            "termination_reason": termination_reason,
            "tool_calls": tool_calls,
            "evidence_items": evidence_items,
            "open_questions": self._evidence_open_questions(tool_calls, evidence_items),
            "draft_reply_candidates": [],
        }
        if task.requested_artifacts:
            ledger["requested_artifacts"] = list(task.requested_artifacts)
            ledger["artifact_optional"] = task.artifact_optional
        if task.context_summary:
            ledger["context_summary"] = task.context_summary
        if task.trace_id:
            ledger["trace_id"] = task.trace_id
        if task.workflow_id:
            ledger["workflow_id"] = task.workflow_id
        return ledger

    def _build_question_evidence_ledger(self, task: RunnerTaskRequest, observed: JsonObject, termination_reason: str) -> JsonObject:
        spec = self._question_investigation_spec(task)
        input_ledger = dict(self._question_input_evidence_ledger(task))
        tool_calls = self._merge_runtime_values(input_ledger.get("tool_calls"), self._compact_tool_calls(observed))
        evidence_items = self._merge_runtime_values(input_ledger.get("evidence_items"), self._compact_evidence_items(observed))
        open_questions = self._merge_runtime_values(
            input_ledger.get("open_questions"),
            self._evidence_open_questions(_json_object_list(tool_calls), _json_object_list(evidence_items)),
        )
        missing_evidence = self._merge_runtime_values(input_ledger.get("missing_evidence"), input_ledger.get("insufficiency_markers"))
        ledger: JsonObject = {
            "investigation_spec": spec,
            "user_request": first_non_empty(_string_or_json(spec.get("user_request")), _string_or_json(input_ledger.get("user_request")), self._original_task_prompt(task)),
            "reply_target": input_ledger.get("reply_target")
            or {
                "channel_id": task.channel_id or "",
                "thread_ts": task.thread_ts or "",
            },
            "prompt_envelope": _json_object_or_empty(_first_non_none(spec.get("prompt_envelope"), input_ledger.get("prompt_envelope"))),
            "repo": first_non_empty(_string_or_json(spec.get("repo")), _string_or_json(input_ledger.get("repo")), task.repo),
            "project_key": first_non_empty(_string_or_json(spec.get("project_key")), _string_or_json(input_ledger.get("project_key"))),
            "since": first_non_empty(_string_or_json(spec.get("since")), _string_or_json(input_ledger.get("since"))),
            "until": first_non_empty(_string_or_json(spec.get("until")), _string_or_json(input_ledger.get("until"))),
            "alignment_required": _bool_or_false(_first_non_none(spec.get("alignment_required"), input_ledger.get("alignment_required"))),
            "alignment_degraded": _bool_or_false(input_ledger.get("alignment_degraded")),
            "termination_reason": termination_reason,
            "tool_calls": tool_calls,
            "evidence_items": evidence_items,
            "open_questions": open_questions,
            "missing_evidence": missing_evidence,
            "draft_reply_candidates": self._merge_runtime_values(input_ledger.get("draft_reply_candidates")),
        }
        alignment_ledger = _json_object_or_empty(input_ledger.get("alignment_ledger"))
        if alignment_ledger:
            ledger["alignment_ledger"] = alignment_ledger
        return ledger

    def _workflow_evidence_raw(self, task: RunnerTaskRequest, observed: JsonObject, termination_reason: str) -> JsonObject:
        if task.task_type in {"question_gather", "question_expand"}:
            return {"evidence_ledger": self._build_question_evidence_ledger(task, observed, termination_reason)}
        if task.task_type != "workflow":
            return {}
        return {"evidence_ledger": self._build_evidence_ledger(task, observed, termination_reason)}

    def _partial_reducer_system_prompt(self) -> str:
        return "\n".join(
            [
                "You finalize bounded-stop RSI Slack reply workflows.",
                "Use only the supplied evidence ledger.",
                "Do not call tools. Do not invent evidence. Do not speculate beyond the ledger.",
                "Return only one JSON object with keys: visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, outcome_hypotheses, produced_artifacts, artifact_failure_reason, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta.",
                "Produce a grounded best-effort partial reply when the evidence supports one.",
                "If the evidence is incomplete, say so explicitly in final_answer and self_critique instead of guessing.",
                "Keep visible_reasoning concise and grounded in the supplied ledger.",
            ]
        )

    def _partial_reducer_user_prompt(
        self,
        task: RunnerTaskRequest,
        termination_reason: str,
        evidence_ledger: JsonObject,
        *,
        previous_error: str = "",
        previous_response: str = "",
    ) -> str:
        parts = [
            "Produce the final structured output for this bounded-stop Slack workflow reply.",
            f"Bounded stop reason: {termination_reason}",
            "Evidence ledger:",
            json.dumps(evidence_ledger, ensure_ascii=True, sort_keys=True, indent=2),
        ]
        if previous_error:
            parts.extend(
                [
                    "Previous reducer attempt failed.",
                    f"Failure reason: {previous_error}",
                ]
            )
        if previous_response:
            parts.extend(
                [
                    "Previous reducer response excerpt:",
                    _string_or_json(previous_response)[:1200],
                ]
            )
        if task.channel_id and task.thread_ts:
            parts.append(f"Target reply channel/thread: {task.channel_id} / {task.thread_ts}")
        parts.append("Return JSON only.")
        return "\n\n".join(part for part in parts if part)

    def _partial_reducer_request_payload(self, system_prompt: str, user_prompt: str) -> JsonObject:
        payload: JsonObject = {
            "model": self._provider_model,
            "instructions": system_prompt,
            "input": self._json_object_input_prompt(user_prompt),
            "parallel_tool_calls": False,
            "max_output_tokens": 2000,
            "text": {
                "format": {"type": "json_object"},
                "verbosity": "low",
            },
        }
        if self._reasoning_config.get("enabled", True):
            payload["reasoning"] = {"effort": "low"}
        return payload

    def _responses_output_text(self, payload: JsonObject) -> str:
        direct = _string_or_json(payload.get("output_text"))
        if direct:
            return direct
        collected: list[str] = []
        for item in payload.get("output", []):
            if not isinstance(item, dict):
                continue
            item_text = _string_or_json(item.get("text"))
            if item_text:
                collected.append(item_text)
            for content in item.get("content", []):
                if not isinstance(content, dict):
                    continue
                text = content.get("text")
                if isinstance(text, dict):
                    candidate = _string_or_json(text.get("value"))
                else:
                    candidate = _string_or_json(text)
                if candidate:
                    collected.append(candidate)
        return "\n".join(part for part in collected if part)

    def _responses_incomplete_reason(self, payload: JsonObject) -> str:
        details = _json_object_or_empty(payload.get("incomplete_details"))
        return _string_or_json(details.get("reason"))

    def _invoke_direct_json_response(
        self,
        *,
        system_prompt: str,
        user_prompt: str,
        timeout_seconds: int,
        reasoning_effort: str = "low",
        recorder: NativeExecutionRecorder | None = None,
        operation: str = "direct_json_response",
    ) -> PartialReducerAttemptResult:
        if self._provider != "openai" or not self._api_key:
            if recorder is not None:
                recorder.record(
                    "direct_response_failed",
                    {
                        "operation": operation,
                        "error": "Direct JSON reduction requires OpenAI Responses API credentials.",
                    },
                )
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error="Direct JSON reduction requires OpenAI Responses API credentials.",
                provider_response_id="",
            )
        payload = self._partial_reducer_request_payload(system_prompt, user_prompt)
        if reasoning_effort and reasoning_effort != "low":
            payload["reasoning"] = {"effort": reasoning_effort}
        if recorder is not None:
            recorder.record(
                "direct_response_request",
                {
                    "operation": operation,
                    "timeout_seconds": timeout_seconds,
                    "payload": payload,
                },
            )
        req = urlrequest.Request(
            f"{self._base_url.rstrip('/')}/responses",
            data=json.dumps(payload).encode("utf-8"),
            headers={
                "Authorization": f"Bearer {self._api_key}",
                "Content-Type": "application/json",
            },
            method="POST",
        )
        try:
            with urlrequest.urlopen(req, timeout=max(1, timeout_seconds)) as resp:
                body = resp.read().decode("utf-8")
        except urlerror.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")
            if recorder is not None:
                recorder.record(
                    "direct_response_failed",
                    {
                        "operation": operation,
                        "error": f"HTTP {exc.code}: {detail[:4000]}",
                    },
                )
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error=f"Direct reducer returned {exc.code}: {detail[:2000]}",
                provider_response_id="",
            )
        except (TimeoutError, socket.timeout) as exc:
            if recorder is not None:
                recorder.record(
                    "direct_response_failed",
                    {
                        "operation": operation,
                        "error": str(exc),
                    },
                )
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error=f"Direct reducer timed out after {max(1, timeout_seconds)}s: {exc}",
                provider_response_id="",
            )
        except (urlerror.URLError, ConnectionError, OSError) as exc:
            if recorder is not None:
                recorder.record(
                    "direct_response_failed",
                    {
                        "operation": operation,
                        "error": str(exc),
                    },
                )
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error=f"Direct reducer request failed: {exc}",
                provider_response_id="",
            )
        try:
            parsed = json.loads(body)
        except json.JSONDecodeError:
            if recorder is not None:
                recorder.record(
                    "direct_response_invalid_json",
                    {
                        "operation": operation,
                        "body_excerpt": body[:4000],
                    },
                )
            return PartialReducerAttemptResult(
                ok=False,
                response_text=body[:4000],
                structured_output={},
                error="Direct reducer returned invalid JSON.",
                provider_response_id="",
            )
        if not isinstance(parsed, dict):
            if recorder is not None:
                recorder.record(
                    "direct_response_invalid_shape",
                    {
                        "operation": operation,
                        "payload": parsed if isinstance(parsed, list) else {"value": _string_or_json(parsed)},
                    },
                )
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error="Direct reducer returned a non-object JSON payload.",
                provider_response_id="",
            )
        response_text = self._responses_output_text(parsed)
        provider_response_id = _string_or_json(parsed.get("id"))
        if recorder is not None:
            recorder.record(
                "direct_response_response",
                {
                    "operation": operation,
                    "provider_response_id": provider_response_id,
                    "payload": parsed,
                },
            )
        if not response_text:
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error="Direct reducer returned an empty response.",
                provider_response_id=provider_response_id,
            )
        try:
            structured_output = self._extract_structured_output(response_text)
        except HermesStructuredOutputError as exc:
            return PartialReducerAttemptResult(
                ok=False,
                response_text=response_text,
                structured_output={},
                error=str(exc),
                provider_response_id=provider_response_id,
            )
        return PartialReducerAttemptResult(
            ok=True,
            response_text=response_text,
            structured_output=structured_output,
            error="",
            provider_response_id=provider_response_id,
        )

    def _slack_mcp_request(self, method: str, params: JsonObject | None = None, *, notification: bool = False) -> JsonObject:
        if not self._config.slack_mcp_enabled:
            raise RuntimeError("Slack MCP is disabled.")
        token = first_non_empty(os.getenv("RSI_SLACK_USER_TOKEN"), "")
        if not token:
            raise RuntimeError("Slack user token is not configured.")
        request_id = None if notification else method.replace("/", "_")
        payload: JsonObject = {
            "jsonrpc": "2.0",
            "method": method,
            "params": params or {},
        }
        if request_id is not None:
            payload["id"] = request_id
        req = urlrequest.Request(
            self._config.slack_mcp_server_url,
            data=json.dumps(payload).encode("utf-8"),
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
            },
            method="POST",
        )
        try:
            with urlrequest.urlopen(req, timeout=15) as resp:
                body = resp.read().decode("utf-8")
        except urlerror.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")
            raise RuntimeError(f"Slack MCP {method} returned {exc.code}: {detail[:2000]}") from exc
        except (TimeoutError, socket.timeout, urlerror.URLError, ConnectionError, OSError) as exc:
            raise RuntimeError(f"Slack MCP {method} request failed: {exc}") from exc
        if notification:
            return {}
        try:
            parsed = json.loads(body)
        except json.JSONDecodeError as exc:
            raise RuntimeError("Slack MCP returned invalid JSON.") from exc
        if not isinstance(parsed, dict):
            raise RuntimeError("Slack MCP returned a non-object JSON payload.")
        error_payload = _json_object_or_empty(parsed.get("error"))
        if error_payload:
            raise RuntimeError(first_non_empty(_string_or_json(error_payload.get("message")), "Slack MCP returned an error."))
        return _json_object_or_empty(parsed.get("result"))

    def _slack_mcp_tools(self) -> list[JsonObject]:
        if self._slack_mcp_tool_cache is not None:
            return list(self._slack_mcp_tool_cache)
        if not self._config.slack_mcp_enabled:
            self._slack_mcp_tool_cache = []
            return []
        try:
            _ = self._slack_mcp_request(
                "initialize",
                {
                    "protocolVersion": "2025-03-26",
                    "capabilities": {},
                    "clientInfo": {"name": "rsi-agent-platform", "version": "0.1.0"},
                },
            )
            try:
                self._slack_mcp_request("notifications/initialized", {}, notification=True)
            except RuntimeError:
                pass
            result = self._slack_mcp_request("tools/list", {})
            tools = _json_object_list(result.get("tools"))
            self._slack_mcp_tool_cache = tools
            self._slack_mcp_discovery_error = ""
            return list(tools)
        except RuntimeError as exc:
            self._slack_mcp_tool_cache = []
            self._slack_mcp_discovery_error = str(exc)
            return []

    def _slack_mcp_available(self) -> bool:
        if not self._config.slack_mcp_enabled or not self._config.slack_user_token_configured:
            return False
        return len(self._slack_mcp_tools()) > 0

    def _slack_mcp_send_tool_name_or_error(self) -> str:
        if self._slack_mcp_send_tool_name:
            return self._slack_mcp_send_tool_name
        candidates: list[str] = []
        exact_order = [
            "send_message",
            "slack_send_message",
            "conversations_add_message",
            "add_message",
            "post_message",
        ]
        exact_hits = [name for name in exact_order if any(_string_or_json(tool.get("name")) == name for tool in self._slack_mcp_tools())]
        if len(exact_hits) == 1:
            self._slack_mcp_send_tool_name = exact_hits[0]
            return self._slack_mcp_send_tool_name
        for tool in self._slack_mcp_tools():
            name = _string_or_json(tool.get("name"))
            description = _string_or_json(tool.get("description")).lower()
            annotations = _json_object_or_empty(tool.get("annotations"))
            read_only = _bool_or_false(annotations.get("readOnlyHint"))
            lowered = name.lower()
            if read_only:
                continue
            if "canvas" in lowered or "canvas" in description or "draft" in lowered or "draft" in description:
                continue
            if ("send" in lowered or "post" in lowered or "message" in lowered) and ("message" in description or "send" in description or "post" in description):
                candidates.append(name)
        candidates = normalize_tool_names(candidates)
        if len(candidates) != 1:
            raise RuntimeError(f"Slack MCP send-message tool discovery expected exactly one candidate, got {candidates or ['none']}.")
        self._slack_mcp_send_tool_name = candidates[0]
        return self._slack_mcp_send_tool_name

    def _slack_mcp_read_tool_names(self) -> list[str]:
        read_tool_names: list[str] = []
        for tool in self._slack_mcp_tools():
            annotations = _json_object_or_empty(tool.get("annotations"))
            if _bool_or_false(annotations.get("readOnlyHint")):
                read_tool_names.append(_string_or_json(tool.get("name")))
        return normalize_tool_names(read_tool_names)

    def _json_object_input_prompt(self, prompt: str) -> str:
        text = str(prompt or "").strip()
        if "json" in text.lower():
            return text
        prefix = "Return a JSON object only."
        if not text:
            return prefix
        return f"{prefix}\n\n{text}"

    def _invoke_partial_reducer(
        self,
        task: RunnerTaskRequest,
        termination_reason: str,
        evidence_ledger: JsonObject,
        *,
        timeout_seconds: int,
        previous_error: str = "",
        previous_response: str = "",
    ) -> PartialReducerAttemptResult:
        if self._provider != "openai" or not self._api_key:
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error="Direct bounded-stop reduction requires OpenAI Responses API credentials.",
                provider_response_id="",
            )
        return self._invoke_direct_json_response(
            system_prompt=self._partial_reducer_system_prompt(),
            user_prompt=self._partial_reducer_user_prompt(
                task,
                termination_reason,
                evidence_ledger,
                previous_error=previous_error,
                previous_response=previous_response,
            ),
            timeout_seconds=timeout_seconds,
            reasoning_effort="low",
        )

    def _question_reduce_system_prompt(self) -> str:
        return "\n".join(
            [
                "You reduce a read-heavy RSI Slack Q&A evidence ledger into one final reply.",
                "Use only the supplied investigation spec, evidence ledger, and runner diagnostics.",
                "Do not call tools. Do not speculate beyond the supplied evidence.",
                "Return only one JSON object with keys: reply_markdown, confidence, completion_verdict, termination_reason.",
                "reply_markdown must be grounded, concise, and ready for Slack posting.",
                "Use completion_verdict=partial when the supplied diagnostics or ledger indicate a bounded stop such as task_timeout, iteration_budget_exhausted, or output_token_budget_exhausted.",
                "If the evidence is incomplete, say that directly in reply_markdown instead of pretending the evidence was stronger than it was.",
            ]
        )

    def _question_reduce_user_prompt(self, task: RunnerTaskRequest) -> str:
        payload = self._question_task_payload(task)
        return "\n\n".join(
            [
                "Reduce the following read-heavy Slack Q&A workflow into a final JSON reply.",
                json.dumps(payload, ensure_ascii=True, sort_keys=True, indent=2),
                "Return JSON only.",
            ]
        )

    def _question_reduce_defaults(self, task: RunnerTaskRequest) -> tuple[str, str]:
        ledger = self._question_input_evidence_ledger(task)
        diagnostics = self._question_input_runner_diagnostics(task)
        termination_reason = first_non_empty(
            _string_or_json(diagnostics.get("termination_reason")),
            _string_or_json(ledger.get("termination_reason")),
            "normal_completion",
        )
        verdict = "partial" if termination_reason in QUESTION_RUN_BOUNDED_STOP_TERMINATION_REASONS else "complete"
        return verdict, termination_reason

    def _execute_question_reduce_task(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> HermesExecutionResult:
        timeout_seconds = min(self._effective_task_timeout(task), max(1, self._transport_timeout_seconds - 5))
        recorder = self._create_native_execution_recorder(task, label="question_reduce")
        recorder.record(
            "execution_started",
            {
                **self._native_execution_started_payload(task, tool_policy),
                "question_reduce": True,
            },
        )
        attempt = self._invoke_direct_json_response(
            system_prompt=self._question_reduce_system_prompt(),
            user_prompt=self._question_reduce_user_prompt(task),
            timeout_seconds=timeout_seconds,
            reasoning_effort="medium",
            recorder=recorder,
            operation="question_reduce",
        )
        if not attempt.ok:
            message = first_non_empty(attempt.error, "Question reducer failed.")
            recorder.record(
                "execution_failed",
                {
                    "failure_class": "runner_non_ok",
                    "failure_kind": "question_reduce_failed",
                    "error": message,
                },
            )
            return HermesExecutionResult(
                ok=False,
                message=message,
                provider=self._backend,
                raw=self._attach_native_execution_log_path(
                    {
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        "failure_class": "runner_non_ok",
                        "runner_diagnostics": {
                            "failure_kind": "question_reduce_failed",
                            "provider_error_message": message,
                        },
                        "tool_policy_mode": tool_policy.mode,
                        "tool_allowlist_effective": tool_policy.effective,
                        "tool_transport_allowlist_effective": tool_policy.transport_effective,
                        "blocked_tool_names": tool_policy.blocked,
                        "task_timeout_seconds": self._effective_task_timeout(task),
                        "transport_timeout_seconds": self._transport_timeout_seconds,
                        "task_type": task.task_type,
                    },
                    recorder,
                ),
            )
        completion_verdict, termination_reason = self._question_reduce_defaults(task)
        structured_output = dict(attempt.structured_output)
        structured_output["completion_verdict"] = first_non_empty(_string_or_json(structured_output.get("completion_verdict")), completion_verdict)
        structured_output["termination_reason"] = first_non_empty(_string_or_json(structured_output.get("termination_reason")), termination_reason)
        recorder.record(
            "execution_completed",
            {
                "ok": True,
                "completion_verdict": structured_output["completion_verdict"],
                "termination_reason": structured_output["termination_reason"],
                "provider_response_id": attempt.provider_response_id,
            },
        )
        return HermesExecutionResult(
            ok=True,
            message=json.dumps(structured_output, ensure_ascii=True, sort_keys=True),
            provider=self._backend,
            raw=self._attach_native_execution_log_path(
                {
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "task_timeout_seconds": self._effective_task_timeout(task),
                    "transport_timeout_seconds": self._transport_timeout_seconds,
                    "tool_policy_mode": tool_policy.mode,
                    "tool_allowlist_effective": tool_policy.effective,
                    "tool_transport_allowlist_effective": tool_policy.transport_effective,
                    "blocked_tool_names": tool_policy.blocked,
                    "runner_diagnostics": {
                        "completion_verdict": structured_output["completion_verdict"],
                        "termination_reason": structured_output["termination_reason"],
                        "question_reduce_mode": "direct_responses_api",
                    },
                    "completion_verdict": structured_output["completion_verdict"],
                    "termination_reason": structured_output["termination_reason"],
                    "structured_output": structured_output,
                    "task_type": task.task_type,
                    "question_reduce_mode": "direct_responses_api",
                    "provider_response_id": attempt.provider_response_id,
                },
                recorder,
            ),
        )

    def _partial_completion_idempotency_key(self, task: RunnerTaskRequest, termination_reason: str) -> str:
        scope = first_non_empty(task.trace_id, task.workflow_id, task.session_scope_id, "workflow")
        return f"partial-{termination_reason}-{scope}"

    def _workflow_reply_delivery_mode(self, task: RunnerTaskRequest) -> str:
        mode = (task.reply_delivery_mode or "").strip().lower()
        if mode in {"direct", "mediated", "none"}:
            return mode
        return "mediated"

    def _workflow_requires_explicit_reply_action(self, task: RunnerTaskRequest) -> bool:
        return task.task_type == "workflow" and self._workflow_reply_delivery_mode(task) == "mediated"

    def _workflow_allows_fallback_reply_action(self, task: RunnerTaskRequest) -> bool:
        return task.task_type == "workflow" and self._workflow_reply_delivery_mode(task) != "none"

    def _looks_like_slack_send_tool_name(self, tool_name_value: str) -> bool:
        return "slack_send_message" in (tool_name_value or "").strip().lower()

    def _parse_json_object_maybe(self, value: Any) -> JsonObject:
        if isinstance(value, dict):
            return value
        text = _string_or_json(value)
        if not text:
            return {}
        try:
            parsed = json.loads(text)
        except json.JSONDecodeError:
            return {}
        return parsed if isinstance(parsed, dict) else {}

    def _reply_delivery_from_session_delta(
        self,
        task: RunnerTaskRequest,
        structured_output: JsonObject,
        session_messages_delta: list[JsonObject],
    ) -> JsonObject:
        if self._workflow_reply_delivery_mode(task) != "direct":
            return {}
        tool_results: dict[str, JsonObject] = {}
        for item in session_messages_delta:
            if _string_or_json(item.get("role")) != "tool":
                continue
            tool_call_id = first_non_empty(
                _string_or_json(item.get("tool_call_id")),
                _string_or_json(item.get("call_id")),
                _string_or_json(item.get("id")),
            )
            if not tool_call_id:
                continue
            payload = self._parse_json_object_maybe(item.get("content"))
            result_payload = self._parse_json_object_maybe(payload.get("result")) if payload else {}
            merged = dict(payload)
            if result_payload:
                merged["result"] = result_payload
            tool_results[tool_call_id] = merged
        for item in reversed(session_messages_delta):
            if _string_or_json(item.get("role")) != "assistant":
                continue
            for tool_call in reversed(_json_object_list(item.get("tool_calls"))):
                function_payload = _json_object_or_empty(tool_call.get("function"))
                tool_name_value = _string_or_json(function_payload.get("name"))
                if not self._looks_like_slack_send_tool_name(tool_name_value):
                    continue
                request_payload = self._parse_json_object_maybe(function_payload.get("arguments"))
                body = first_non_empty(
                    _string_or_json(request_payload.get("message")),
                    _string_or_json(request_payload.get("text")),
                    _string_or_json(structured_output.get("final_answer")),
                    _string_or_json(structured_output.get("reply_draft")),
                )
                tool_call_id = first_non_empty(
                    _string_or_json(tool_call.get("call_id")),
                    _string_or_json(tool_call.get("id")),
                )
                result_payload = tool_results.get(tool_call_id, {})
                result_data = _json_object_or_empty(result_payload.get("result"))
                message_context = _json_object_or_empty(result_data.get("message_context"))
                provider_ref = first_non_empty(
                    _string_or_json(message_context.get("message_ts")),
                    _string_or_json(result_payload.get("provider_ref")),
                )
                message_link = _string_or_json(result_data.get("message_link"))
                if not provider_ref and not message_link:
                    continue
                artifact_refs = _string_list_or_empty(result_payload.get("raw_artifact_refs"))
                body_sha1 = hashlib.sha1(body.encode("utf-8")).hexdigest() if body else ""
                return {
                    "status": "posted",
                    "channel_id": first_non_empty(_string_or_json(request_payload.get("channel_id")), task.channel_id or ""),
                    "thread_ts": first_non_empty(_string_or_json(request_payload.get("thread_ts")), task.thread_ts or ""),
                    "body": body,
                    "body_sha1": body_sha1,
                    "body_excerpt": body[:280],
                    "tool_call_id": tool_call_id,
                    "tool_name": tool_name_value,
                    "provider_ref": provider_ref,
                    "message_link": message_link,
                    "artifact_refs": artifact_refs,
                }
        return {}

    def _synthesize_partial_slack_post_action(
        self,
        task: RunnerTaskRequest,
        structured_output: JsonObject,
        termination_reason: str,
    ) -> tuple[JsonObject, bool]:
        if not self._workflow_allows_fallback_reply_action(task):
            return structured_output, False
        for item in _normalize_proposed_actions(structured_output.get("proposed_actions")):
            if _string_or_json(item.get("kind")) == "slack_post":
                return structured_output, False
        if _json_object_or_empty(structured_output.get("reply_delivery")):
            return structured_output, False
        reply_body = first_non_empty(
            _string_or_json(structured_output.get("final_answer")),
            _string_or_json(structured_output.get("reply_draft")),
        )
        if not reply_body:
            return structured_output, False
        actions = [
            dict(item)
            for item in _normalize_proposed_actions(structured_output.get("proposed_actions"))
            if _string_or_json(item.get("kind")) != "slack_post"
        ]
        request_payload: JsonObject = {"body": reply_body}
        artifact_refs = []
        for item in _normalize_produced_artifacts(structured_output.get("produced_artifacts")):
            artifact_refs.extend(_string_list_or_empty(item.get("artifact_refs")))
        if artifact_refs:
            request_payload["artifact_refs"] = artifact_refs
        if task.channel_id:
            request_payload["channel_id"] = task.channel_id
        if task.thread_ts:
            request_payload["thread_ts"] = task.thread_ts
        actions.append(
            {
                "kind": "slack_post",
                "target_ref": first_non_empty(task.channel_id, task.thread_ts, task.trace_id),
                "request_payload": request_payload,
                "approval_mode": "not_required",
                "idempotency_key": self._partial_completion_idempotency_key(task, termination_reason),
                "rationale": "Post the grounded partial answer.",
                "evidence_refs": [],
            }
        )
        normalized = dict(structured_output)
        normalized["proposed_actions"] = actions
        return normalized, True

    def _partial_completion_unrecoverable_failure(
        self,
        task: RunnerTaskRequest,
        tool_policy: ToolPolicy,
        finalized: JsonObject,
        observed: JsonObject,
        stop_meta: JsonObject,
        lifecycle_events: list[JsonObject],
        *,
        termination_reason: str,
        recovery_error: str = "",
        recovery_response: str = "",
        partial_finalization_attempted: bool,
        partial_finalization_succeeded: bool,
        partial_finalization_attempts: int = 0,
        partial_finalization_retry_attempted: bool = False,
        partial_finalization_retry_succeeded: bool = False,
        partial_finalization_timeout_seconds: int = 0,
        reducer_attempt_errors: list[str] | None = None,
    ) -> HermesExecutionResult:
        last_activity = _json_object_or_empty(stop_meta.get("last_activity"))
        timeout_kind = string_from_map(stop_meta, "timeout_kind")
        max_iterations_reached = bool(stop_meta.get("max_iterations_reached")) or termination_reason == "iteration_budget_exhausted"
        message = "Hermes could not finalize a partial workflow response before the bounded execution window closed."
        if termination_reason == "task_timeout":
            message = "Hermes hit the workflow time limit and could not finalize a partial workflow response."
        elif termination_reason == "iteration_budget_exhausted":
            message = "Hermes exhausted its iteration budget and could not finalize a partial workflow response."
        diagnostics = self._runner_diagnostics(
            tool_policy,
            failure_kind="partial_completion_unrecoverable",
            provider_error_message=first_non_empty(recovery_error, message),
            timeout_kind=timeout_kind or None,
            termination_reason=termination_reason,
            activity=last_activity,
            max_iterations_reached=max_iterations_reached,
            session_ready_issues=self._session_manager.ready_issues,
            repair_attempted=False,
            repair_succeeded=False,
            observed=observed,
        )
        diagnostics["partial_completion_attempted"] = partial_finalization_attempted
        diagnostics["partial_completion_succeeded"] = partial_finalization_succeeded
        diagnostics["partial_finalization_mode"] = "direct_reducer"
        diagnostics["partial_finalization_attempts"] = partial_finalization_attempts
        diagnostics["partial_finalization_retry_attempted"] = partial_finalization_retry_attempted
        diagnostics["partial_finalization_retry_succeeded"] = partial_finalization_retry_succeeded
        diagnostics["partial_finalization_timeout_seconds"] = partial_finalization_timeout_seconds
        if self._provider_model:
            diagnostics["reducer_model"] = self._provider_model
        if reducer_attempt_errors:
            diagnostics["reducer_attempt_errors"] = list(reducer_attempt_errors)
        raw: JsonObject = {
            **self._base_raw(prompt=task.prompt, system_message=task.system_message),
            **finalized,
            **stop_meta,
            "task_timeout_seconds": self._effective_task_timeout(task),
            "inactivity_timeout_seconds": self._effective_inactivity_timeout(self._effective_task_timeout(task)),
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "max_iterations": self._max_iterations,
            "tool_policy_mode": tool_policy.mode,
            "tool_allowlist_effective": tool_policy.effective,
            "tool_transport_allowlist_effective": tool_policy.transport_effective,
            "blocked_tool_names": tool_policy.blocked,
            **observed,
            **self._workflow_evidence_raw(task, observed, termination_reason),
            "failure_class": "runner_partial_completion_unrecoverable",
            "runner_diagnostics": diagnostics,
            "lifecycle_events": lifecycle_events,
            "termination_reason": termination_reason,
            "max_iterations_reached": max_iterations_reached,
            "partial_recovery_attempted": partial_finalization_attempted,
            "partial_recovery_succeeded": partial_finalization_succeeded,
        }
        if recovery_error:
            raw["recovery_error"] = recovery_error
        if recovery_response:
            raw["recovery_response"] = recovery_response
        if reducer_attempt_errors:
            raw["reducer_attempt_errors"] = list(reducer_attempt_errors)
        return HermesExecutionResult(
            ok=False,
            message=first_non_empty(recovery_error, message),
            provider=self._backend,
            raw=raw,
        )

    def _partial_completion_failure(
        self,
        task: RunnerTaskRequest,
        tool_policy: ToolPolicy,
        finalized: JsonObject,
        observed: JsonObject,
        stop_meta: JsonObject,
        lifecycle_events: list[JsonObject],
        *,
        termination_reason: str,
        recovery_error: str = "",
        recovery_response: str = "",
        recovery_attempted: bool,
        recovery_succeeded: bool,
    ) -> HermesExecutionResult:
        last_activity = _json_object_or_empty(stop_meta.get("last_activity"))
        max_iterations_reached = bool(stop_meta.get("max_iterations_reached")) or termination_reason == "iteration_budget_exhausted"
        message = "Hermes execution exhausted its iteration budget before completing the workflow response."
        failure_class = "runner_iteration_budget_exhausted"
        failure_kind = "iteration_budget_exhausted"
        timeout_kind = ""
        if termination_reason == "task_timeout":
            message = f"Hermes execution timed out after {self._effective_task_timeout(task)}s."
            failure_class = "runner_transport_timeout"
            failure_kind = "transport_timeout"
            timeout_kind = string_from_map(stop_meta, "timeout_kind") or "task_timeout"
        diagnostics = self._runner_diagnostics(
            tool_policy,
            failure_kind=failure_kind,
            provider_error_message=first_non_empty(recovery_error, message),
            timeout_kind=timeout_kind or None,
            termination_reason=termination_reason,
            activity=last_activity,
            max_iterations_reached=max_iterations_reached,
            session_ready_issues=self._session_manager.ready_issues,
            repair_attempted=recovery_attempted,
            repair_succeeded=recovery_succeeded,
            observed=observed,
        )
        diagnostics["recovery_attempted"] = recovery_attempted
        diagnostics["recovery_succeeded"] = recovery_succeeded
        raw: JsonObject = {
            **self._base_raw(prompt=task.prompt, system_message=task.system_message),
            **finalized,
            **stop_meta,
            "task_timeout_seconds": self._effective_task_timeout(task),
            "inactivity_timeout_seconds": self._effective_inactivity_timeout(self._effective_task_timeout(task)),
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "max_iterations": self._max_iterations,
            "tool_policy_mode": tool_policy.mode,
            "tool_allowlist_effective": tool_policy.effective,
            "tool_transport_allowlist_effective": tool_policy.transport_effective,
            "blocked_tool_names": tool_policy.blocked,
            **observed,
            **self._workflow_evidence_raw(task, observed, termination_reason),
            "failure_class": failure_class,
            "runner_diagnostics": diagnostics,
            "lifecycle_events": lifecycle_events,
            "termination_reason": termination_reason,
            "max_iterations_reached": max_iterations_reached,
        }
        if recovery_error:
            raw["recovery_error"] = recovery_error
        if recovery_response:
            raw["recovery_response"] = recovery_response
        return HermesExecutionResult(
            ok=False,
            message=first_non_empty(recovery_error, message),
            provider=self._backend,
            raw=raw,
        )

    def _finalize_partial_completion(
        self,
        task: RunnerTaskRequest,
        tool_policy: ToolPolicy,
        finalized: JsonObject,
        observed: JsonObject,
        stop_meta: JsonObject,
        lifecycle_events: list[JsonObject],
        *,
        termination_reason: str,
        observer: ObservationEmitter | None = None,
    ) -> HermesExecutionResult:
        recovery_timeout_seconds = self._partial_completion_timeout_seconds(task, termination_reason, stop_meta)
        evidence_ledger = self._build_evidence_ledger(task, observed, termination_reason)
        if observer is not None:
            observer.emit(
                phase=self._execution_phase(task),
                event_type="reducer.started",
                status="running",
                payload={
                    "termination_reason": termination_reason,
                    "timeout_seconds": recovery_timeout_seconds,
                },
            )
        if recovery_timeout_seconds <= 0:
            return self._partial_completion_unrecoverable_failure(
                task,
                tool_policy,
                finalized,
                observed,
                stop_meta,
                lifecycle_events,
                termination_reason=termination_reason,
                recovery_error="The partial-completion finalization window expired before recovery could start.",
                partial_finalization_attempted=False,
                partial_finalization_succeeded=False,
            )

        attempt_budgets = self._partial_completion_attempt_budgets(recovery_timeout_seconds)
        if not attempt_budgets:
            return self._partial_completion_unrecoverable_failure(
                task,
                tool_policy,
                finalized,
                observed,
                stop_meta,
                lifecycle_events,
                termination_reason=termination_reason,
                recovery_error="The partial-completion finalization window did not leave enough time for a reducer attempt.",
                partial_finalization_attempted=False,
                partial_finalization_succeeded=False,
            )

        last_error = ""
        last_response = ""
        last_provider_response_id = ""
        attempt_errors: list[str] = []
        last_attempt_timeout_seconds = 0
        for attempt_index, attempt_timeout_seconds in enumerate(attempt_budgets, start=1):
            last_attempt_timeout_seconds = attempt_timeout_seconds
            attempt_result = self._invoke_partial_reducer(
                task,
                termination_reason,
                evidence_ledger,
                timeout_seconds=attempt_timeout_seconds,
                previous_error=last_error if attempt_index > 1 else "",
                previous_response=last_response if attempt_index > 1 else "",
            )
            last_error = attempt_result.error
            last_response = attempt_result.response_text
            last_provider_response_id = attempt_result.provider_response_id
            if not attempt_result.ok:
                if attempt_result.error:
                    attempt_errors.append(attempt_result.error)
                continue
            structured_output, action_contract_synthesized = self._synthesize_partial_slack_post_action(
                task,
                attempt_result.structured_output,
                termination_reason,
            )
            response_text = json.dumps(structured_output, ensure_ascii=True, sort_keys=True)
            merged_runner_diagnostics = dict(_json_object_or_empty(observed))
            if last_provider_response_id:
                merged_runner_diagnostics["partial_finalization_response_id"] = last_provider_response_id
            merged_runner_diagnostics["partial_finalization_mode"] = "direct_reducer"
            merged_runner_diagnostics["partial_finalization_attempts"] = attempt_index
            merged_runner_diagnostics["partial_finalization_retry_attempted"] = attempt_index > 1
            merged_runner_diagnostics["partial_finalization_retry_succeeded"] = attempt_index > 1
            merged_runner_diagnostics["partial_finalization_timeout_seconds"] = attempt_timeout_seconds
            merged_runner_diagnostics["partial_completion_attempted"] = True
            merged_runner_diagnostics["partial_completion_succeeded"] = True
            merged_runner_diagnostics["completion_verdict"] = "partial"
            merged_runner_diagnostics["action_contract_repair_attempted"] = action_contract_synthesized
            merged_runner_diagnostics["action_contract_repair_succeeded"] = action_contract_synthesized
            if self._provider_model:
                merged_runner_diagnostics["reducer_model"] = self._provider_model
            if attempt_errors:
                merged_runner_diagnostics["reducer_attempt_errors"] = list(attempt_errors)
            max_iterations_reached = bool(stop_meta.get("max_iterations_reached")) or termination_reason == "iteration_budget_exhausted"
            for key in ("budget_used", "budget_max", "api_call_count", "current_tool", "last_activity_desc"):
                if key in _json_object_or_empty(stop_meta.get("last_activity")):
                    merged_runner_diagnostics[key] = _json_object_or_empty(stop_meta.get("last_activity")).get(key)
            timeout_kind = string_from_map(stop_meta, "timeout_kind")
            if timeout_kind:
                merged_runner_diagnostics["timeout_kind"] = timeout_kind
            merged_runner_diagnostics["termination_reason"] = termination_reason
            merged_runner_diagnostics["max_iterations_reached"] = max_iterations_reached
            merged_raw: JsonObject = {
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                **finalized,
                **stop_meta,
                "task_timeout_seconds": self._effective_task_timeout(task),
                "inactivity_timeout_seconds": self._effective_inactivity_timeout(self._effective_task_timeout(task)),
                "transport_timeout_seconds": self._transport_timeout_seconds,
                "max_iterations": self._max_iterations,
                "tool_policy_mode": tool_policy.mode,
                "tool_allowlist_effective": tool_policy.effective,
                "tool_transport_allowlist_effective": tool_policy.transport_effective,
                "blocked_tool_names": tool_policy.blocked,
                **observed,
                "evidence_ledger": evidence_ledger,
                "runner_diagnostics": merged_runner_diagnostics,
                "lifecycle_events": lifecycle_events,
                "termination_reason": termination_reason,
                "max_iterations_reached": max_iterations_reached,
                "completion_verdict": "partial",
                "partial_recovery_attempted": True,
                "partial_recovery_succeeded": True,
                "structured_output": structured_output,
            }
            if last_provider_response_id:
                merged_raw["partial_finalization_response_id"] = last_provider_response_id
            if observer is not None:
                observer.emit(
                    phase=self._execution_phase(task),
                    event_type="reducer.completed",
                    status="completed",
                    payload={
                        "attempts": attempt_index,
                        "termination_reason": termination_reason,
                    },
                )
            return HermesExecutionResult(
                ok=True,
                message=response_text,
                provider=self._backend,
                raw=merged_raw,
            )

        if observer is not None:
            observer.emit(
                phase=self._execution_phase(task),
                event_type="reducer.completed",
                status="failed",
                payload={
                    "attempts": len(attempt_budgets),
                    "termination_reason": termination_reason,
                    "errors": attempt_errors,
                },
            )
        return self._partial_completion_unrecoverable_failure(
            task,
            tool_policy,
            finalized,
            observed,
            stop_meta,
            lifecycle_events,
            termination_reason=termination_reason,
            recovery_error=first_non_empty(
                last_error,
                "Direct bounded-stop reducer could not produce valid structured output.",
            ),
            recovery_response=last_response,
            partial_finalization_attempted=True,
            partial_finalization_succeeded=False,
            partial_finalization_attempts=len(attempt_budgets),
            partial_finalization_retry_attempted=len(attempt_budgets) > 1,
            partial_finalization_retry_succeeded=False,
            partial_finalization_timeout_seconds=last_attempt_timeout_seconds,
            reducer_attempt_errors=attempt_errors,
        )

    def _preserved_native_tool_names(self, task: RunnerTaskRequest, current_tools: list[JsonToolWrapperSchema]) -> set[str]:
        execution_phase = self._execution_phase(task)
        names = {name for name in (tool_name(tool) for tool in current_tools) if name}
        if execution_phase == "render":
            return {
                name
                for name in names
                if name in ARTIFACT_RENDER_NATIVE_TOOL_NAMES or name.startswith("mcp_")
            }
        if execution_phase == "deliver":
            return {name for name in names if name.startswith("mcp_")}
        if task.task_type in {"workflow", "question_gather", "question_expand"}:
            return names
        return set()

    def _attach_tool_policy(self, agent: Any, task: RunnerTaskRequest, tool_policy: ToolPolicy, *, observer: ObservationEmitter | None = None) -> None:
        current_tools = list(getattr(agent, "tools", []) or [])
        current_valid = set(getattr(agent, "valid_tool_names", set()) or set())
        setattr(agent, "_rsi_readonly_tool_binding", None)
        native_governed_tools = self._native_governed_tools_enabled(task)
        allowed_names = set(tool_policy.effective)
        preserved_tool_names = self._preserved_native_tool_names(task, current_tools)
        filtered_tools = current_tools
        query_hints = self._question_default_query_hints(task)
        if native_governed_tools:
            allowed_transport_names = {
                _transport_name_or_self(name)
                for name in tool_policy.effective
                if name not in BLOCKED_HONCHO_TOOLS
            }
            allowed_transport_names.update(preserved_tool_names)
            filtered_tools = [tool for tool in current_tools if tool_name(tool) in allowed_transport_names]
            agent.tools = filtered_tools
            agent.valid_tool_names = {name for name in current_valid if name in allowed_transport_names}
            return
        filtered_tools = [tool for tool in current_tools if tool_name(tool) in allowed_names or tool_name(tool) in preserved_tool_names]
        custom_tool_names = [name for name in tool_policy.custom_tools if self._config.tool_gateway_base_url]
        custom_transport_names = [tool_policy.custom_tool_transport_map[name] for name in custom_tool_names if name in tool_policy.custom_tool_transport_map]
        if custom_tool_names:
            filtered_tools.extend(tool_schema_wrappers(custom_tool_names))
            readonly_tools = ReadOnlyToolBinding(
                base_url=self._config.tool_gateway_base_url or "",
                allowed_tool_names=custom_tool_names,
                task_repo=task.repo,
                task_repo_ref=task.repo_ref or "",
                task_prompt=task.prompt,
                default_question=str(query_hints.get("default_question", "")),
                repo_question=str(query_hints.get("repo_question", "")),
                knowledge_topic=str(query_hints.get("knowledge_topic", "")),
                knowledge_question=str(query_hints.get("knowledge_question", "")),
                slack_history_focus=str(query_hints.get("slack_history_focus", "")),
                slack_search_query=str(query_hints.get("slack_search_query", "")),
                task_channel_id=task.channel_id or "",
                task_thread_ts=task.thread_ts or "",
                task_context_summary=task.context_summary or "",
                trace_id=task.trace_id or "",
                session_scope_kind=task.session_scope_kind or "",
                session_scope_id=task.session_scope_id or "",
                context_refs=task.context_refs,
                execution_mode=task.execution_mode or "",
                execution_phase=task.execution_phase or "",
                attempt_id=task.attempt_id or "",
                workspace_id=task.workspace_id or "",
                observer=observer,
            )
            setattr(agent, "_rsi_readonly_tool_binding", readonly_tools)
            agent._memory_manager = CompositeToolProvider(getattr(agent, "_memory_manager", None), readonly_tools)
        elif getattr(agent, "_memory_manager", None) is not None:
            readonly_tools = ReadOnlyToolBinding(
                base_url=self._config.tool_gateway_base_url or "",
                allowed_tool_names=[],
                task_repo=task.repo,
                task_repo_ref=task.repo_ref or "",
                task_prompt=task.prompt,
                default_question=str(query_hints.get("default_question", "")),
                repo_question=str(query_hints.get("repo_question", "")),
                knowledge_topic=str(query_hints.get("knowledge_topic", "")),
                knowledge_question=str(query_hints.get("knowledge_question", "")),
                slack_history_focus=str(query_hints.get("slack_history_focus", "")),
                slack_search_query=str(query_hints.get("slack_search_query", "")),
                task_channel_id=task.channel_id or "",
                task_thread_ts=task.thread_ts or "",
                task_context_summary=task.context_summary or "",
                trace_id=task.trace_id or "",
                session_scope_kind=task.session_scope_kind or "",
                session_scope_id=task.session_scope_id or "",
                context_refs=task.context_refs,
                execution_mode=task.execution_mode or "",
                execution_phase=task.execution_phase or "",
                attempt_id=task.attempt_id or "",
                workspace_id=task.workspace_id or "",
                observer=observer,
            )
            setattr(agent, "_rsi_readonly_tool_binding", readonly_tools)
            agent._memory_manager = CompositeToolProvider(getattr(agent, "_memory_manager", None), readonly_tools)
        effective_names = set(tool_policy.effective)
        current_valid = {
            name
            for name in current_valid
            if (name in effective_names and name not in BLOCKED_HONCHO_TOOLS) or name in preserved_tool_names
        }
        current_valid.update(custom_transport_names)
        agent.tools = filtered_tools
        agent.valid_tool_names = current_valid

    def _run_with_deadlines(
        self,
        agent: Any,
        task: RunnerTaskRequest,
        context: SessionContext,
        timeout_seconds: int,
        inactivity_timeout_seconds: int,
        reasoning_timeout_seconds: int,
        *,
        observer: ObservationEmitter | None = None,
    ) -> tuple[str, JsonObject | None, JsonObject]:
        executor = concurrent.futures.ThreadPoolExecutor(max_workers=1)
        if observer is not None:
            observer.emit(
                phase=self._execution_phase(task),
                event_type="model.call.started",
                status="running",
                payload={
                    "timeout_seconds": timeout_seconds,
                    "reasoning_timeout_seconds": reasoning_timeout_seconds,
                    "inactivity_timeout_seconds": inactivity_timeout_seconds,
                },
            )
        future = executor.submit(
            agent.run_conversation,
            task.prompt,
            task.system_message,
            context.conversation_history,
            context.session_id,
        )
        try:
            started_at = time.monotonic()
            while True:
                try:
                    result = future.result(timeout=0.25)
                    activity = safe_activity_summary(agent)
                    termination_reason = "iteration_budget_exhausted" if self._budget_exhausted(activity) else "normal_completion"
                    if observer is not None:
                        observer.emit(
                            phase=self._execution_phase(task),
                            event_type="model.call.completed",
                            status="completed",
                            payload={
                                "termination_reason": termination_reason,
                                "activity": activity,
                            },
                        )
                    return termination_reason, result, {
                        "last_activity": activity,
                        "last_tool_invoked": string_from_map(activity, "current_tool"),
                        "max_iterations_reached": self._budget_exhausted(activity),
                    }
                except concurrent.futures.TimeoutError:
                    activity = safe_activity_summary(agent)
                    if self._budget_exhausted(activity):
                        return self._interrupt_execution(
                            agent,
                            future,
                            "iteration_budget_exhausted",
                            activity,
                            observer=observer,
                            phase=self._execution_phase(task),
                        )
                    elapsed_seconds = max(0.0, time.monotonic() - started_at)
                    idle_seconds = inactivity_seconds(activity, elapsed_seconds)
                    if elapsed_seconds >= float(reasoning_timeout_seconds):
                        return self._interrupt_execution(
                            agent,
                            future,
                            "task_timeout",
                            activity,
                            duration_seconds=reasoning_timeout_seconds,
                            observer=observer,
                            phase=self._execution_phase(task),
                        )
                    if inactivity_timeout_seconds > 0 and idle_seconds >= float(inactivity_timeout_seconds):
                        return self._interrupt_execution(
                            agent,
                            future,
                            "inactivity_timeout",
                            activity,
                            duration_seconds=inactivity_timeout_seconds,
                            observer=observer,
                            phase=self._execution_phase(task),
                        )
        finally:
            executor.shutdown(wait=False, cancel_futures=True)

    def _interrupt_execution(
        self,
        agent: Any,
        future: concurrent.futures.Future,
        termination_reason: str,
        activity: JsonObject,
        *,
        duration_seconds: int = 0,
        observer: ObservationEmitter | None = None,
        phase: str = "main",
    ) -> tuple[str, JsonObject | None, JsonObject]:
        if observer is not None:
            observer.emit(
                phase=phase,
                event_type="execution.interrupt",
                status=termination_reason,
                payload={
                    "duration_seconds": duration_seconds,
                    "activity": activity,
                },
            )
        if duration_seconds > 0:
            agent.interrupt(f"runner {termination_reason} after {duration_seconds}s")
        else:
            agent.interrupt(f"runner {termination_reason}")
        shutdown_error = ""
        try:
            grace_seconds = min(5, max(1, duration_seconds // 10)) if duration_seconds > 0 else 5
            future.result(timeout=grace_seconds)
        except concurrent.futures.TimeoutError:
            shutdown_error = f"{termination_reason} shutdown did not complete before the grace period elapsed."
        except Exception as exc:
            shutdown_error = str(exc)
        latest_activity = safe_activity_summary(agent) or activity
        meta = {
            "termination_reason": termination_reason,
            "last_activity": latest_activity,
            "last_activity_desc": string_from_map(latest_activity, "last_activity_desc"),
            "last_tool_invoked": string_from_map(latest_activity, "current_tool"),
            "max_iterations_reached": self._budget_exhausted(latest_activity),
        }
        if termination_reason in {"task_timeout", "inactivity_timeout"}:
            meta["timeout_kind"] = termination_reason
        if duration_seconds > 0:
            meta["stopped_after_seconds"] = duration_seconds
        if shutdown_error:
            meta["shutdown_error"] = shutdown_error
        if observer is not None:
            observer.emit(
                phase=phase,
                event_type="model.call.completed",
                status=termination_reason,
                payload=meta,
            )
        return termination_reason, None, meta

    def execute_task(self, task: RunnerTaskRequest) -> HermesExecutionResult:
        observer = ObservationEmitter.create(
            self._config,
            trace_id=task.trace_id or "",
            workflow_id=task.workflow_id or "",
            operation_id=task.operation_id or "",
            role=self._role,
            hermes_session_id=stable_session_id(
                self._role,
                first_non_empty(task.session_scope_kind, "role"),
                first_non_empty(task.session_scope_id, self._role),
            ),
            execution_id=task.execution_id or "",
        )
        self._store_executor_result(
            observer.execution_id,
            {
                "execution_id": observer.execution_id,
                "status": "running",
                "message": "Execution accepted.",
                "phase": self._execution_phase(task),
            },
        )
        if self._config.execution_envelope_v1_enabled:
            contract_validation = self._company_computer.validate_task(task)
            if not contract_validation.ok:
                result = HermesExecutionResult(
                    ok=False,
                    message="Runner execution contract failed: " + "; ".join(contract_validation.errors),
                    provider="runner-contract",
                    raw=self._company_computer.failure_result_raw(task, errors=contract_validation.errors),
                )
                result = self._company_computer.attach_envelope(task, result, observer=observer)
                self._store_executor_result(
                    observer.execution_id,
                    self._executor_final_status(task, result, execution_id=observer.execution_id, status="failed"),
                )
                return result
        result = self._execute_task_internal(task, observer=observer)
        if self._config.execution_envelope_v1_enabled:
            result = self._company_computer.attach_envelope(task, result, observer=observer)
        self._store_executor_result(
            observer.execution_id,
            self._executor_final_status(task, result, execution_id=observer.execution_id),
        )
        return result

    def _execute_task_internal(self, task: RunnerTaskRequest, *, observer: ObservationEmitter | None = None) -> HermesExecutionResult:
        if task.task_type not in ROLE_TASK_TYPES.get(self._role, {self._role}):
            return HermesExecutionResult(
                ok=False,
                message=f"Runner role {self._role} cannot execute task type {task.task_type}.",
                provider="policy",
                raw={"role": self._role, "task_type": task.task_type},
            )
        tool_policy = self._resolve_tool_policy(task)
        if task.task_type == "question_reduce":
            return self._execute_question_reduce_task(task, tool_policy)
        if self._task_uses_artifact_phases(task):
            return self._execute_artifact_workflow_task(task, tool_policy, observer=observer)
        if self._native_executor_enabled_for_task(task):
            result = self._execute_native_workflow_task_request(
                task,
                tool_policy,
                observer=observer,
                max_iterations_override=self._phase_max_iterations_override(task),
            )
        else:
            result = self._execute_task_request(
                task,
                tool_policy,
                observer=observer,
                max_iterations_override=self._phase_max_iterations_override(task),
            )
        if not result.ok:
            return result
        rendered_task = replace(
            task,
            prompt=_string_or_json(result.raw.get("prompt")) or task.prompt,
            system_message=_optional_string(result.raw.get("system_message")) or task.system_message,
        )
        completion_verdict = string_from_map(_json_object_or_empty(result.raw), "completion_verdict")
        initial_completion_verdict = completion_verdict
        initial_termination_reason = string_from_map(_json_object_or_empty(result.raw), "termination_reason")
        initial_max_iterations_reached = _bool_or_false(_json_object_or_empty(result.raw).get("max_iterations_reached"))
        initial_response = result.message
        repair_attempted = False
        repair_succeeded = False
        action_contract_repair_attempted = False
        action_contract_repair_succeeded = False
        action_contract_repair_error = ""
        action_contract_repair_response = ""
        partial_structured_output = _json_object_or_empty(result.raw.get("structured_output"))
        used_partial_structured_output = False
        if completion_verdict == "partial" and partial_structured_output:
            structured_output = _normalize_structured_output(partial_structured_output)
            used_partial_structured_output = True
            partial_runner_diagnostics = _json_object_or_empty(result.raw.get("runner_diagnostics"))
            action_contract_repair_attempted = _bool_or_false(
                result.raw.get("action_contract_repair_attempted")
            ) or _bool_or_false(partial_runner_diagnostics.get("action_contract_repair_attempted"))
            action_contract_repair_succeeded = _bool_or_false(
                result.raw.get("action_contract_repair_succeeded")
            ) or _bool_or_false(partial_runner_diagnostics.get("action_contract_repair_succeeded"))
        else:
            repair_tool_policy = tool_policy
            if invalid_request := self._provider_invalid_request_diagnostics(result.message, repair_tool_policy):
                diagnostics = dict(invalid_request)
                for key, value in _json_object_or_empty(result.raw.get("runner_diagnostics")).items():
                    diagnostics[key] = value
                return HermesExecutionResult(
                    ok=False,
                    message=string_from_map(diagnostics, "provider_error_message") or "Provider rejected the runner request.",
                    provider=result.provider,
                    raw={
                        **result.raw,
                        "failure_class": "runner_invalid_request",
                        "runner_diagnostics": diagnostics,
                        "raw_response": result.message,
                        "repair_attempted": False,
                        "repair_succeeded": False,
                    },
                    )
            try:
                structured_output = self._extract_structured_output(result.message)
            except HermesStructuredOutputError as exc:
                if task.task_type == "workflow" and self._config.workflow_runner_repair_attempts > 0 and self._structured_output_repairable(exc):
                    repair_attempted = True
                    logger.info(
                        "workflow runner structured-output repair attempted trace_id=%s workflow_id=%s",
                        task.trace_id or "",
                        task.workflow_id or "",
                    )
                    repair_task = self._build_structured_output_repair_task(rendered_task, result.message)
                    repair_result = self._execute_task_request(
                        repair_task,
                        repair_tool_policy,
                        observer=observer,
                        render_prompt=False,
                        expand_skills=False,
                    )
                    if repair_result.ok:
                        if invalid_request := self._provider_invalid_request_diagnostics(repair_result.message, repair_tool_policy):
                            diagnostics = dict(invalid_request)
                            for key, value in _json_object_or_empty(repair_result.raw.get("runner_diagnostics")).items():
                                diagnostics[key] = value
                            diagnostics["repair_attempted"] = True
                            diagnostics["repair_succeeded"] = False
                            return HermesExecutionResult(
                                ok=False,
                                message=string_from_map(diagnostics, "provider_error_message") or "Provider rejected the runner request.",
                                provider=repair_result.provider,
                                raw={
                                    **repair_result.raw,
                                    "failure_class": "runner_invalid_request",
                                    "runner_diagnostics": diagnostics,
                                    "raw_response": repair_result.message,
                                    "repair_attempted": True,
                                    "repair_succeeded": False,
                                    "repair_original_response": result.message,
                                },
                            )
                        try:
                            structured_output = self._extract_structured_output(repair_result.message)
                            result = repair_result
                            repair_succeeded = True
                            logger.info(
                                "workflow runner structured-output repair succeeded trace_id=%s workflow_id=%s",
                                task.trace_id or "",
                                task.workflow_id or "",
                            )
                        except HermesStructuredOutputError as repair_exc:
                            logger.warning(
                                "workflow runner structured-output repair failed trace_id=%s workflow_id=%s error=%s",
                                task.trace_id or "",
                                task.workflow_id or "",
                                repair_exc,
                            )
                            return HermesExecutionResult(
                                ok=False,
                                message=str(repair_exc),
                                provider=repair_result.provider,
                                raw={
                                    **repair_result.raw,
                                    "failure_class": "runner_structured_output_parse_failure",
                                    "runner_diagnostics": self._runner_diagnostics(
                                        repair_tool_policy,
                                        failure_kind="structured_output_parse_failure",
                                        provider_error_message=str(repair_exc),
                                        repair_attempted=True,
                                        repair_succeeded=False,
                                        observed=_json_object_or_empty(repair_result.raw.get("runner_diagnostics")),
                                    ),
                                    "structured_output_error": str(repair_exc),
                                    "raw_response": repair_result.message,
                                    "repair_attempted": True,
                                    "repair_succeeded": False,
                                    "repair_original_response": result.message,
                                },
                            )
                    else:
                        logger.warning(
                            "workflow runner structured-output repair failed trace_id=%s workflow_id=%s error=%s",
                            task.trace_id or "",
                            task.workflow_id or "",
                            repair_result.message,
                        )
                        return HermesExecutionResult(
                            ok=False,
                            message=repair_result.message,
                            provider=repair_result.provider,
                            raw={
                                **repair_result.raw,
                                "runner_diagnostics": {
                                    **_json_object_or_empty(repair_result.raw.get("runner_diagnostics")),
                                    "repair_attempted": True,
                                    "repair_succeeded": False,
                                } if repair_result.raw.get("runner_diagnostics") is not None else repair_result.raw.get("runner_diagnostics"),
                                "raw_response": repair_result.message,
                                "repair_attempted": True,
                                "repair_succeeded": False,
                                "repair_original_response": result.message,
                            },
                        )
                else:
                    return HermesExecutionResult(
                        ok=False,
                        message=str(exc),
                        provider=result.provider,
                        raw={
                            **result.raw,
                            "failure_class": "runner_structured_output_parse_failure",
                            "runner_diagnostics": self._runner_diagnostics(
                                repair_tool_policy,
                                failure_kind="structured_output_parse_failure",
                                provider_error_message=str(exc),
                                repair_attempted=False,
                                repair_succeeded=False,
                                observed=_json_object_or_empty(result.raw.get("runner_diagnostics")),
                            ),
                            "structured_output_error": str(exc),
                            "raw_response": result.message,
                            "repair_attempted": False,
                            "repair_succeeded": False,
                        },
                    )
            reply_delivery = _json_object_or_empty(structured_output.get("reply_delivery"))
            if not reply_delivery:
                reply_delivery = self._reply_delivery_from_session_delta(
                    task,
                    structured_output,
                    _json_object_list(result.raw.get("session_messages_delta")),
                )
            if reply_delivery:
                structured_output["reply_delivery"] = reply_delivery
                result.raw["reply_delivery"] = reply_delivery
            if self._workflow_missing_explicit_reply_action(task, structured_output):
                action_contract_repair_attempted = True
                logger.info(
                    "workflow runner action-contract repair attempted trace_id=%s workflow_id=%s",
                    task.trace_id or "",
                    task.workflow_id or "",
                )
                repair_task = self._build_action_contract_repair_task(rendered_task, structured_output)
                repair_result = self._execute_task_request(
                    repair_task,
                    repair_tool_policy,
                    observer=observer,
                    render_prompt=False,
                    expand_skills=False,
                )
                action_contract_repair_response = repair_result.message
                if repair_result.ok:
                    try:
                        repaired_output = self._extract_structured_output(repair_result.message)
                    except HermesStructuredOutputError as exc:
                        action_contract_repair_error = str(exc)
                    else:
                        if not self._workflow_missing_explicit_reply_action(task, repaired_output):
                            structured_output = repaired_output
                            result = repair_result
                            action_contract_repair_succeeded = True
                            logger.info(
                                "workflow runner action-contract repair succeeded trace_id=%s workflow_id=%s",
                                task.trace_id or "",
                                task.workflow_id or "",
                            )
                        else:
                            action_contract_repair_error = "Hermes repair response still omitted the required slack_post action."
                else:
                    action_contract_repair_error = repair_result.message
        existing_repair_attempted = bool(result.raw.get("repair_attempted"))
        existing_repair_succeeded = bool(result.raw.get("repair_succeeded"))
        result.raw = {
            **result.raw,
            "role": self._role,
            "task_type": task.task_type,
            "repo": task.repo,
            "repo_ref": task.repo_ref,
            "allowed_tools": task.allowed_tools,
            "allowed_commands": task.allowed_commands,
            "timeout_seconds": task.timeout_seconds,
            "expected_outputs": task.expected_outputs,
            "artifact_destination": task.artifact_destination,
            "context_summary": task.context_summary,
            "rejected_proposal_context": task.rejected_proposal_context,
            "execution_mode": task.execution_mode,
            "intent": task.intent,
            "trace_id": task.trace_id,
            "workflow_id": task.workflow_id,
            "operation_id": task.operation_id,
            "channel_id": task.channel_id,
            "thread_ts": task.thread_ts,
            "repo_allowlist": task.repo_allowlist,
            "tool_allowlist": task.tool_allowlist,
            "tool_policy_mode": tool_policy.mode,
            "tool_allowlist_effective": tool_policy.effective,
            "blocked_tool_names": tool_policy.blocked,
            "response_mode": task.response_mode,
            "reply_delivery_mode": task.reply_delivery_mode,
            "context_refs": task.context_refs,
            "approval_mode": task.approval_mode,
            "reasoning_verbosity": task.reasoning_verbosity,
            "session_scope_kind": task.session_scope_kind,
            "session_scope_id": task.session_scope_id,
            "parent_session_scope_kind": task.parent_session_scope_kind,
            "parent_session_scope_id": task.parent_session_scope_id,
            "harness_profile_id": task.harness_profile_id,
            "harness_overlay_version": task.harness_overlay_version,
            "memory_backend": task.memory_backend,
            "assistant_peer_id": task.assistant_peer_id,
            "user_peer_id": task.user_peer_id,
            "attempt_id": task.attempt_id,
            "workspace_id": task.workspace_id,
            "workspace_repo": task.workspace_repo,
            "workspace_branch": task.workspace_branch,
            "allowed_path_globs": task.allowed_path_globs,
            "mcp_servers": task.mcp_servers,
            "execution_phase": task.execution_phase,
            "contract_version": task.contract_version or EXECUTION_CONTRACT_VERSION,
            "execution_intent": task.execution_intent,
            "capability_leases": task.capability_leases,
            "delivery_policy": task.delivery_policy,
            "workspace_policy": task.workspace_policy,
            "approval_policy": task.approval_policy,
            "repair_attempted": repair_attempted or existing_repair_attempted,
            "repair_succeeded": repair_succeeded or existing_repair_succeeded,
            "action_contract_repair_attempted": action_contract_repair_attempted,
            "action_contract_repair_succeeded": action_contract_repair_succeeded,
            "structured_output": structured_output,
        }
        if initial_completion_verdict == "partial":
            result.raw["completion_verdict"] = "partial"
            result.raw["termination_reason"] = first_non_empty(initial_termination_reason, _string_or_json(result.raw.get("termination_reason")))
            result.raw["max_iterations_reached"] = initial_max_iterations_reached or initial_termination_reason == "iteration_budget_exhausted"
            if used_partial_structured_output:
                result.raw["model_output_contract"] = "partial_structured_output"
        if repair_attempted:
            result.raw["repair_original_response"] = initial_response
        runner_diagnostics = _json_object_or_empty(result.raw.get("runner_diagnostics"))
        runner_diagnostics["candidate_read_surfaces"] = result.raw.get("candidate_read_surfaces", [])
        runner_diagnostics["selected_context_surfaces"] = result.raw.get("selected_context_surfaces", [])
        runner_diagnostics["memory_warnings"] = result.raw.get("memory_warnings", [])
        runner_diagnostics["action_contract_repair_attempted"] = action_contract_repair_attempted
        runner_diagnostics["action_contract_repair_succeeded"] = action_contract_repair_succeeded
        if initial_completion_verdict == "partial":
            runner_diagnostics["completion_verdict"] = "partial"
            runner_diagnostics["termination_reason"] = result.raw["termination_reason"]
            runner_diagnostics["max_iterations_reached"] = result.raw["max_iterations_reached"]
        if observer is not None:
            for key, value in observer.diagnostics().items():
                runner_diagnostics[key] = value
        if action_contract_repair_error:
            result.raw["action_contract_repair_error"] = action_contract_repair_error
            runner_diagnostics["action_contract_repair_error"] = action_contract_repair_error
        if action_contract_repair_response:
            result.raw["action_contract_repair_response"] = action_contract_repair_response
            runner_diagnostics["action_contract_repair_response"] = action_contract_repair_response
        result.raw["runner_diagnostics"] = runner_diagnostics
        return result

    def _artifact_output_root(self, task: RunnerTaskRequest) -> Path:
        destination = _string_or_json(task.artifact_destination)
        if destination and destination.startswith("/"):
            path = Path(destination).expanduser()
            path.mkdir(parents=True, exist_ok=True)
            return path
        if task.contract_version == EXECUTION_CONTRACT_VERSION or os.getenv("RSI_HERMES_ARTIFACT_ROOT", "").strip():
            return self._native_artifact_output_dir(task)
        scope = first_non_empty(task.trace_id, task.workflow_id, task.session_scope_id, "artifact-workflow")
        path = Path(self._config.hermes_home) / "rsi_runtime" / "artifacts" / scope
        path.mkdir(parents=True, exist_ok=True)
        return path

    def _artifact_output_extension(self, kind: str, skill: str) -> str:
        if kind == "diagram" and skill == "architecture-diagram":
            return ".html"
        if kind == "diagram":
            return ".svg"
        return ".txt"

    def _artifact_output_path(self, task: RunnerTaskRequest, *, kind: str, skill: str, title: str) -> str:
        safe_title = title.replace(" ", "-").lower()
        extension = self._artifact_output_extension(kind, skill)
        return str((self._artifact_output_root(task) / f"{safe_title}{extension}").resolve())

    def _artifact_brief_title(self, task: RunnerTaskRequest, brief: JsonObject, *, index: int) -> str:
        title = _string_or_json(brief.get("title"))
        if title:
            return title
        if index - 1 < len(task.requested_artifacts):
            requested_title = _string_or_json(task.requested_artifacts[index - 1].get("description"))
            if requested_title:
                return requested_title
        kind = _string_or_json(brief.get("kind")) or "artifact"
        return f"{kind}-{index}"

    def _artifact_brief_render_prompt(
        self,
        task: RunnerTaskRequest,
        structured_output: JsonObject,
        brief: JsonObject,
        *,
        title: str,
        index: int,
    ) -> str:
        render_prompt = _string_or_json(brief.get("render_prompt"))
        if render_prompt:
            return render_prompt
        requested_description = ""
        if index - 1 < len(task.requested_artifacts):
            requested_description = _string_or_json(task.requested_artifacts[index - 1].get("description"))
        grounded_final_answer = _string_or_json(structured_output.get("final_answer"))
        grounded_context_summary = _string_or_json(structured_output.get("context_summary"))
        if not requested_description and not grounded_final_answer and not grounded_context_summary:
            return self._extract_user_request_text(task.prompt)
        parts = [
            requested_description,
            f"Render the requested artifact titled {title}.",
            grounded_final_answer,
            grounded_context_summary,
        ]
        synthesized = "\n\n".join(part for part in parts if part)
        return synthesized or self._extract_user_request_text(task.prompt)

    def _artifact_default_skill(self, task: RunnerTaskRequest, brief: JsonObject) -> str:
        requested_brief_skill = _string_or_json(first_non_empty(brief.get("requested_skill"), brief.get("skill")))
        if requested_brief_skill:
            return _normalize_skill_identifier(requested_brief_skill)
        if task.requested_skills:
            return _normalize_skill_identifier(task.requested_skills[0])
        if _string_or_json(brief.get("kind")) == "diagram":
            return "architecture-diagram"
        return ""

    def _artifact_render_skill_diagnostics(self, task: RunnerTaskRequest, brief: JsonObject) -> JsonObject:
        requested_skill = _string_or_json(first_non_empty(brief.get("requested_skill"), brief.get("skill")))
        if not requested_skill and task.requested_skills:
            requested_skill = _string_or_json(task.requested_skills[0])
        resolved_skill = self._artifact_default_skill(task, brief)
        return {
            "requested_skills": [requested_skill] if requested_skill else [],
            "resolved_skills": [resolved_skill] if resolved_skill else [],
        }

    def _artifact_render_briefs(self, task: RunnerTaskRequest, structured_output: JsonObject) -> list[JsonObject]:
        briefs = _normalize_artifact_render_briefs(structured_output.get("artifact_render_briefs"))
        if briefs:
            hydrated: list[JsonObject] = []
            for index, brief in enumerate(briefs, start=1):
                kind = _string_or_json(brief.get("kind"))
                if not kind:
                    continue
                skill = self._artifact_default_skill(task, brief)
                title = self._artifact_brief_title(task, brief, index=index)
                hydrated.append(
                    {
                        "kind": kind,
                        "skill": skill,
                        "requested_skill": _string_or_json(first_non_empty(brief.get("requested_skill"), brief.get("skill"))),
                        "title": title,
                        "render_prompt": self._artifact_brief_render_prompt(
                            task,
                            structured_output,
                            brief,
                            title=title,
                            index=index,
                        ),
                        "inputs": _json_object_or_empty(brief.get("inputs")),
                        "output_path_hint": _string_or_json(brief.get("output_path_hint"))
                        or self._artifact_output_path(task, kind=kind, skill=skill, title=title),
                    }
                )
            return hydrated
        fallback: list[JsonObject] = []
        for index, requested in enumerate(task.requested_artifacts, start=1):
            kind = _string_or_json(requested.get("kind"))
            if not kind:
                continue
            brief = {"kind": kind, "title": _string_or_json(requested.get("description"))}
            skill = self._artifact_default_skill(task, brief)
            title = self._artifact_brief_title(task, brief, index=index)
            fallback.append(
                {
                    "kind": kind,
                    "skill": skill,
                    "requested_skill": _string_or_json(first_non_empty(brief.get("requested_skill"), brief.get("skill"))),
                    "title": title,
                    "render_prompt": self._artifact_brief_render_prompt(
                        task,
                        structured_output,
                        brief,
                        title=title,
                        index=index,
                    ),
                    "inputs": {},
                    "output_path_hint": self._artifact_output_path(task, kind=kind, skill=skill, title=title),
                }
            )
        return fallback

    def _investigate_mcp_servers(self, task: RunnerTaskRequest) -> list[JsonObject]:
        servers: list[JsonObject] = []
        for server in task.mcp_servers:
            item = dict(server)
            if _string_or_json(item.get("profile")) == "slack_mcp_reply":
                item["profile"] = "slack_mcp_read"
            servers.append(item)
        return servers

    def _remove_delivery_tools(self, allowed_tools: list[str]) -> list[str]:
        return [item for item in allowed_tools if item != "slack.upload_file"]

    def _append_system_message(self, task: RunnerTaskRequest, extra: str) -> str | None:
        return "\n\n".join(part for part in [task.system_message or "", extra.strip()] if part.strip()) or None

    def _build_artifact_investigate_task(self, task: RunnerTaskRequest, budgets: JsonObject) -> RunnerTaskRequest:
        prompt = "\n\n".join(
            [
                task.prompt,
                "Artifact workflow contract:",
                "Do not render files, upload files, or send Slack messages in this phase.",
                "Investigate and distill the request into one or more artifact_render_briefs with compact grounded inputs.",
                "Return final_answer, reply_draft, context_summary, artifact_render_briefs, produced_artifacts, and artifact_failure_reason in the structured output.",
            ]
        )
        return replace(
            task,
            prompt=prompt,
            system_message=self._append_system_message(
                task,
                "Investigation phase only. Slack delivery is disabled in this phase. Produce compact artifact_render_briefs instead of generating artifacts now.",
            ),
            allowed_tools=self._remove_delivery_tools(task.allowed_tools),
            requested_artifacts=[],
            mcp_servers=self._investigate_mcp_servers(task),
            reply_delivery_mode="none",
            timeout_seconds=int(budgets.get("investigate") or 0),
            execution_phase="investigate",
        )

    def _build_artifact_render_task(
        self,
        task: RunnerTaskRequest,
        brief: JsonObject,
        investigate_output: JsonObject,
        budgets: JsonObject,
        index: int,
    ) -> RunnerTaskRequest:
        kind = _string_or_json(brief.get("kind"))
        skill = self._artifact_default_skill(task, brief)
        title = self._artifact_brief_title(task, brief, index=index + 1)
        output_path = _string_or_json(brief.get("output_path_hint")) or self._artifact_output_path(
            task, kind=kind, skill=skill, title=title
        )
        skill_diagnostics = self._artifact_render_skill_diagnostics(task, brief)
        requested_skill_input = "none"
        if skill_diagnostics["requested_skills"]:
            requested_skill_input = _string_or_json(skill_diagnostics["requested_skills"][0]) or "none"
        native_render_instructions = "Use file-writing tools to generate the artifact."
        native_render_system_message = "Use file-writing tools and the selected skill to generate the artifact."
        if self._native_executor_enabled_for_task(task):
            native_render_instructions = (
                "Use artifact_list_files and artifact_write_file "
                "to inspect and save files only within the native artifact directory."
            )
            native_render_system_message = (
                "Use artifact_write_file to persist the artifact inside the staged native artifact directory."
            )
        render_prompt = "\n\n".join(
            [
                f"Artifact render phase for {kind}.",
                f"Selected skill: {skill or 'none'}",
                f"Requested skill input: {requested_skill_input}",
                f"Output path: {output_path}",
                f"Grounded context summary: {_string_or_json(investigate_output.get('context_summary'))}",
                f"Grounded final answer: {_string_or_json(investigate_output.get('final_answer'))}",
                f"Render brief: {json.dumps(brief, ensure_ascii=True, sort_keys=True)}",
                f"Generate the artifact only. Save it to the output path. {native_render_instructions} Do not send Slack messages.",
                "Return structured output with produced_artifacts and artifact_failure_reason. produced_artifacts must include file:// artifact refs for saved files.",
            ]
        )
        return replace(
            task,
            prompt=render_prompt,
            system_message=self._append_system_message(
                task,
                f"Render phase only. Do not investigate broadly. Do not send Slack messages. {native_render_system_message}",
            ),
            requested_skills=[skill] if skill else [],
            requested_artifacts=[{"kind": kind, "description": title}],
            allowed_tools=[],
            mcp_servers=[],
            reply_delivery_mode="none",
            timeout_seconds=int(budgets.get("render") or 0),
            context_summary=_string_or_json(investigate_output.get("context_summary")),
            execution_mode="artifact_render",
            execution_phase="render",
        )

    def _build_artifact_delivery_task(
        self,
        task: RunnerTaskRequest,
        investigate_output: JsonObject,
        produced_artifacts: list[JsonObject],
        budgets: JsonObject,
        *,
        artifact_failure_reason: str = "",
    ) -> RunnerTaskRequest:
        artifact_refs: list[str] = []
        for artifact in produced_artifacts:
            artifact_refs.extend(_string_list_or_empty(artifact.get("artifact_refs")))
        has_artifacts = bool(artifact_refs)
        if has_artifacts:
            prompt = "\n\n".join(
                [
                    "Artifact delivery phase.",
                    f"Final reply body: {_string_or_json(investigate_output.get('final_answer')) or _string_or_json(investigate_output.get('reply_draft'))}",
                    f"Produced artifacts: {json.dumps(produced_artifacts, ensure_ascii=True, sort_keys=True)}",
                    "Upload the produced file artifacts to the bound Slack thread using slack.upload_file.",
                    "Pass the final reply body as initial_comment on the upload.",
                    "Do not send a second Slack message after uploading.",
                    "Return reply_delivery plus produced_artifacts. Do not perform repo or knowledge investigation.",
                ]
            )
            system_message = (
                "Delivery phase only. Do not investigate. Upload the produced artifacts once and use the upload initial comment as the single direct reply."
            )
            allowed_tools = ["slack.upload_file"]
            delivery_branch = "artifact_upload"
        else:
            prompt = "\n\n".join(
                [
                    "Artifact delivery phase.",
                    f"Final reply body: {_string_or_json(investigate_output.get('final_answer')) or _string_or_json(investigate_output.get('reply_draft'))}",
                    "No file artifacts were produced.",
                    f"Render failure reason: {artifact_failure_reason or 'none'}",
                    "Send exactly one final Slack reply via Slack MCP.",
                    "Do not call slack.upload_file. Do not perform repo or knowledge investigation.",
                ]
            )
            system_message = "Delivery phase only. Do not investigate. Send exactly one final Slack reply via Slack MCP."
            allowed_tools = []
            delivery_branch = "text_only"
        return replace(
            task,
            prompt=prompt,
            system_message=self._append_system_message(task, system_message),
            requested_skills=[],
            requested_artifacts=[],
            allowed_tools=allowed_tools,
            tool_allowlist=allowed_tools,
            timeout_seconds=int(budgets.get("deliver") or 0),
            context_summary=_string_or_json(investigate_output.get("context_summary")),
            execution_phase="deliver",
            context_refs=[
                *task.context_refs,
                {
                    "kind": "artifact_delivery_branch",
                    "ref": delivery_branch,
                    "summary": f"Artifact delivery branch selected: {delivery_branch}.",
                },
            ],
        )

    def _synthesized_slack_post_action(self, task: RunnerTaskRequest, body: str, produced_artifacts: list[JsonObject]) -> list[JsonObject]:
        if not body:
            return []
        payload: JsonObject = {"body": body}
        artifact_refs: list[str] = []
        for artifact in produced_artifacts:
            artifact_refs.extend(_string_list_or_empty(artifact.get("artifact_refs")))
        if artifact_refs:
            payload["artifact_refs"] = artifact_refs
        if task.channel_id:
            payload["channel_id"] = task.channel_id
        if task.thread_ts:
            payload["thread_ts"] = task.thread_ts
        return [
            {
                "kind": "slack_post",
                "target_ref": first_non_empty(task.channel_id, task.thread_ts, task.trace_id),
                "request_payload": payload,
                "approval_mode": "not_required",
                "idempotency_key": hashlib.sha1(body.encode("utf-8")).hexdigest(),
                "rationale": "Fallback to mediated Slack delivery after direct delivery phase failure.",
                "evidence_refs": [],
            }
        ]

    def _merge_artifact_phase_result(
        self,
        task: RunnerTaskRequest,
        investigate_result: HermesExecutionResult,
        investigate_output: JsonObject,
        produced_artifacts: list[JsonObject],
        artifact_failure_reason: str,
        *,
        delivery_result: HermesExecutionResult | None = None,
        delivery_output: JsonObject | None = None,
        direct_delivery_failed: str = "",
        observer: ObservationEmitter | None = None,
    ) -> HermesExecutionResult:
        final_output = dict(investigate_output)
        final_output["artifact_render_briefs"] = _normalize_artifact_render_briefs(investigate_output.get("artifact_render_briefs"))
        produced_artifacts = self._mark_artifacts_shared_if_delivered(_normalize_produced_artifacts(produced_artifacts), delivery_output)
        final_output["produced_artifacts"] = produced_artifacts
        final_output["artifact_failure_reason"] = artifact_failure_reason
        if delivery_output:
            if _json_object_or_empty(delivery_output.get("reply_delivery")):
                final_output["reply_delivery"] = _json_object_or_empty(delivery_output.get("reply_delivery"))
            if _normalize_proposed_actions(delivery_output.get("proposed_actions")):
                final_output["proposed_actions"] = _normalize_proposed_actions(delivery_output.get("proposed_actions"))
        if direct_delivery_failed and not _json_object_or_empty(final_output.get("reply_delivery")):
            final_output["proposed_actions"] = self._synthesized_slack_post_action(
                task,
                _string_or_json(final_output.get("final_answer")) or _string_or_json(final_output.get("reply_draft")),
                produced_artifacts,
            )
        base_raw = dict(delivery_result.raw if delivery_result is not None else investigate_result.raw)
        runner_diagnostics = _json_object_or_empty(base_raw.get("runner_diagnostics"))
        budgets = self._artifact_phase_budgets(task)
        runner_diagnostics["artifact_phase_budgets"] = budgets
        runner_diagnostics["artifact_phase_enabled"] = True
        runner_diagnostics["artifact_investigate_completion_verdict"] = _string_or_json(
            investigate_result.raw.get("completion_verdict")
        )
        runner_diagnostics["artifact_investigate_termination_reason"] = _string_or_json(
            investigate_result.raw.get("termination_reason")
        )
        if delivery_result is not None:
            runner_diagnostics["artifact_delivery_completion_verdict"] = _string_or_json(
                delivery_result.raw.get("completion_verdict")
            )
        if direct_delivery_failed:
            runner_diagnostics["direct_delivery_phase_failed"] = direct_delivery_failed
        runner_diagnostics["artifact_delivery_branch"] = "artifact_upload" if produced_artifacts else "text_only"
        if artifact_failure_reason:
            runner_diagnostics["artifact_render_failure_reason"] = artifact_failure_reason
        if observer is not None:
            for key, value in observer.diagnostics().items():
                runner_diagnostics[key] = value
        base_raw.update(
            {
                "operation_id": task.operation_id,
                "execution_phase": task.execution_phase,
                "completion_verdict": _string_or_json(base_raw.get("completion_verdict")) or "complete",
                "structured_output": final_output,
                "produced_artifacts": produced_artifacts,
                "artifact_failure_reason": artifact_failure_reason,
                "runner_diagnostics": runner_diagnostics,
            }
        )
        return HermesExecutionResult(
            ok=True,
            message=json.dumps(final_output, ensure_ascii=True, sort_keys=True),
            provider=delivery_result.provider if delivery_result is not None else investigate_result.provider,
            raw=base_raw,
        )

    def _execute_artifact_workflow_task(
        self,
        task: RunnerTaskRequest,
        tool_policy: ToolPolicy,
        *,
        observer: ObservationEmitter | None = None,
    ) -> HermesExecutionResult:
        budgets = self._artifact_phase_budgets(task)
        phase_task = task
        if self._native_executor_enabled_for_task(task):
            artifact_root = self._native_artifact_output_dir(task)
            artifact_root.mkdir(parents=True, exist_ok=True)
            phase_task = replace(task, artifact_destination=str(artifact_root))
        if observer is not None:
            observer.emit(
                phase="main",
                event_type="artifact.pipeline.started",
                status="running",
                payload=budgets,
            )
        investigate_task = self._build_artifact_investigate_task(phase_task, budgets)
        investigate_result = self._execute_task_internal(investigate_task, observer=observer)
        if not investigate_result.ok:
            return investigate_result
        investigate_output = _normalize_structured_output(_json_object_or_empty(investigate_result.raw.get("structured_output")))
        render_briefs = self._artifact_render_briefs(phase_task, investigate_output)
        produced_artifacts: list[JsonObject] = []
        artifact_failure_reasons: list[str] = []
        if not render_briefs:
            artifact_failure_reasons.append(
                first_non_empty(
                    _string_or_json(investigate_output.get("artifact_failure_reason")),
                    "Artifact investigation completed without artifact_render_briefs.",
                )
            )
        for index, brief in enumerate(render_briefs):
            render_task = self._build_artifact_render_task(phase_task, brief, investigate_output, budgets, index)
            render_skill = _normalize_skill_identifier(
                first_non_empty(
                    brief.get("requested_skill"),
                    brief.get("skill"),
                    render_task.requested_skills[0] if render_task.requested_skills else "",
                )
            )
            if not render_skill:
                render_failure = "Artifact render skill identifier is invalid after normalization."
                artifact_failure_reasons.append(render_failure)
                if observer is not None:
                    skill_diagnostics = self._artifact_render_skill_diagnostics(phase_task, brief)
                    observer.emit(
                        phase="render",
                        event_type="phase.completed",
                        status="failed",
                        payload={
                            "completion_verdict": "",
                            "termination_reason": "invalid_render_skill",
                            "artifact_failure_reason": render_failure,
                            "requested_skills": skill_diagnostics["requested_skills"],
                            "resolved_skills": skill_diagnostics["resolved_skills"],
                        },
                    )
                continue
            render_result = self._execute_task_internal(render_task, observer=observer)
            if not render_result.ok:
                artifact_failure_reasons.append(
                    first_non_empty(
                        _string_or_json(_json_object_or_empty(render_result.raw).get("artifact_failure_reason")),
                        render_result.message,
                    )
                )
                continue
            render_output = _normalize_structured_output(_json_object_or_empty(render_result.raw.get("structured_output")))
            rendered_artifacts = self._enrich_artifact_records(_normalize_produced_artifacts(render_output.get("produced_artifacts")), phase_task)
            native_artifact_paths = _string_list_or_empty(render_result.raw.get("native_artifact_paths"))
            if not rendered_artifacts and native_artifact_paths:
                kind = _string_or_json(brief.get("kind"))
                title = self._artifact_brief_title(phase_task, brief, index=index + 1)
                rendered_artifacts = [
                    self._artifact_record_for_path(path, kind=kind, title=title, task=phase_task)
                    for path in native_artifact_paths
                ]
            produced_artifacts.extend(rendered_artifacts)
            for artifact in rendered_artifacts:
                if observer is not None:
                    observer.emit(
                        phase="render",
                        event_type="artifact.file.written",
                        status="completed",
                        payload=artifact,
                    )
            render_failure = _string_or_json(render_output.get("artifact_failure_reason"))
            if render_failure:
                artifact_failure_reasons.append(render_failure)
        produced_artifacts = _normalize_produced_artifacts(produced_artifacts)
        artifact_failure_reason = "; ".join(item for item in artifact_failure_reasons if item)
        reply_delivery_mode = self._workflow_reply_delivery_mode(task)
        if reply_delivery_mode == "direct":
            if observer is not None:
                observer.emit(
                    phase="deliver",
                    event_type="direct_delivery.started",
                    status="running",
                    payload={"delivery_branch": "artifact_upload" if produced_artifacts else "text_only"},
                )
            deliver_task = self._build_artifact_delivery_task(
                phase_task,
                investigate_output,
                produced_artifacts,
                budgets,
                artifact_failure_reason=artifact_failure_reason,
            )
            deliver_result = self._execute_task_internal(deliver_task, observer=observer)
            if deliver_result.ok:
                deliver_output = _normalize_structured_output(_json_object_or_empty(deliver_result.raw.get("structured_output")))
                if observer is not None:
                    observer.emit(
                        phase="deliver",
                        event_type="direct_delivery.completed",
                        status="completed" if _json_object_or_empty(deliver_output.get("reply_delivery")) else "fallback",
                    )
                if _json_object_or_empty(deliver_output.get("reply_delivery")):
                    return self._merge_artifact_phase_result(
                        task,
                        investigate_result,
                        investigate_output,
                        produced_artifacts,
                        artifact_failure_reason,
                        delivery_result=deliver_result,
                        delivery_output=deliver_output,
                        observer=observer,
                    )
                return self._merge_artifact_phase_result(
                    task,
                    investigate_result,
                    investigate_output,
                    produced_artifacts,
                    artifact_failure_reason,
                    delivery_result=deliver_result,
                    delivery_output=deliver_output,
                    direct_delivery_failed="direct delivery phase completed without reply_delivery",
                    observer=observer,
                )
            if observer is not None:
                observer.emit(
                    phase="deliver",
                    event_type="direct_delivery.completed",
                    status="failed",
                    payload={"error": deliver_result.message},
                )
            return self._merge_artifact_phase_result(
                task,
                investigate_result,
                investigate_output,
                produced_artifacts,
                artifact_failure_reason,
                direct_delivery_failed=deliver_result.message,
                observer=observer,
            )
        return self._merge_artifact_phase_result(
            task,
            investigate_result,
            investigate_output,
            produced_artifacts,
            artifact_failure_reason,
            observer=observer,
        )

    def _render_task_prompt(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> str:
        parts = [
            f"Runner role: {self._role}",
            f"Task type: {task.task_type}",
            f"Repository: {task.repo}",
            f"Configured model: {self._configured_model}",
            f"Reasoning effort: {self._reasoning_effort}",
            f"Max iterations: {self._max_iterations}",
            f"Task timeout seconds: {self._effective_task_timeout(task)}",
            f"Inactivity timeout seconds: {self._effective_inactivity_timeout(self._effective_task_timeout(task))}",
            f"Transport timeout seconds: {self._transport_timeout_seconds}",
            f"Tool policy mode: {tool_policy.mode}",
            "Detailed RSI evidence is injected through the Hermes context engine rather than appended inline to this prompt.",
        ]
        if task.task_type in {"question_gather", "question_expand"}:
            question_payload = self._question_task_payload(task)
            parts.append("Gather evidence for the Slack Q&A ledger with bounded, grounded read-only retrieval.")
            parts.append("Return only one JSON object with keys: tool_calls, evidence_items, open_questions, insufficiency_markers, confidence.")
            parts.append("Do not answer the user. Do not emit reply text, actions, or knowledge drafts.")
            parts.append("Use the current evidence ledger to close the most important remaining gaps only.")
            parts.append(f"Question run payload:\n{json.dumps(question_payload, ensure_ascii=True, sort_keys=True, indent=2)}")
            return "\n".join(parts)
        if task.repo_ref:
            parts.append(f"Repository ref: {task.repo_ref}")
        if task.intent:
            parts.append(f"Intent: {task.intent}")
        if task.trace_id:
            parts.append(f"Trace ID: {task.trace_id}")
        if task.workflow_id:
            parts.append(f"Workflow ID: {task.workflow_id}")
        if task.execution_mode:
            parts.append(f"Execution mode: {task.execution_mode}")
        parts.append(f"Execution contract: {task.contract_version or EXECUTION_CONTRACT_VERSION}")
        parts.append(f"Runner planner mode: {self._config.runner_planner_mode or RUNNER_PLANNER_MODE}")
        if task.capability_leases:
            parts.append(f"Capability leases: {json.dumps(task.capability_leases, ensure_ascii=True, sort_keys=True)}")
        if task.delivery_policy:
            parts.append(f"Delivery policy: {json.dumps(task.delivery_policy, ensure_ascii=True, sort_keys=True)}")
        if task.workspace_policy:
            parts.append(f"Workspace policy: {json.dumps(task.workspace_policy, ensure_ascii=True, sort_keys=True)}")
        if task.approval_policy:
            parts.append(f"Approval policy: {json.dumps(task.approval_policy, ensure_ascii=True, sort_keys=True)}")
        if task.context_refs:
            parts.append(f"Context ref count: {len(task.context_refs)}")
        if task.attempt_id:
            parts.append(f"Attempt ID: {task.attempt_id}")
        if task.workspace_id:
            parts.append(f"Workspace ID: {task.workspace_id}")
        if task.workspace_repo:
            parts.append(f"Workspace repo: {task.workspace_repo}")
        if task.workspace_branch:
            parts.append(f"Workspace branch: {task.workspace_branch}")
        if task.allowed_path_globs:
            parts.append(f"Allowed path globs: {', '.join(task.allowed_path_globs)}")
        if task.allowed_tools:
            parts.append(f"Requested allowed tools: {', '.join(task.allowed_tools)}")
        if task.tool_allowlist:
            parts.append(f"Requested tool allowlist: {', '.join(task.tool_allowlist)}")
        if tool_policy.effective:
            parts.append(f"Effective tool allowlist: {', '.join(tool_policy.effective)}")
        if tool_policy.blocked:
            parts.append(f"Blocked tools by policy: {', '.join(tool_policy.blocked)}")
        if task.allowed_commands:
            parts.append(f"Allowed commands: {', '.join(task.allowed_commands)}")
        if task.repo_allowlist:
            parts.append(f"Repo allowlist: {', '.join(task.repo_allowlist)}")
        if task.expected_outputs:
            parts.append(f"Expected outputs: {', '.join(task.expected_outputs)}")
        if task.artifact_destination:
            parts.append(f"Artifact destination: {task.artifact_destination}")
        if task.requested_artifacts:
            parts.append(f"Requested artifacts: {json.dumps(task.requested_artifacts, ensure_ascii=True, sort_keys=True)}")
            if task.artifact_optional:
                parts.append("Artifact production is requested but optional; if an artifact cannot be produced, explain why in artifact_failure_reason and still return the best grounded reply.")
            else:
                parts.append("Artifact production is required when the evidence and tools allow it; if it cannot be produced, explain why in artifact_failure_reason.")
        if task.requested_skills:
            parts.append(f"Requested skills: {', '.join(task.requested_skills)}")
        if task.rejected_proposal_context:
            parts.append(f"Prior rejected/dismissed context: {json.dumps(task.rejected_proposal_context)}")
        if task.response_mode:
            parts.append(f"Response mode: {task.response_mode}")
        if task.reply_delivery_mode:
            parts.append(f"Reply delivery mode: {task.reply_delivery_mode}")
        if task.approval_mode:
            parts.append(f"Approval mode: {task.approval_mode}")
        if task.reasoning_verbosity:
            parts.append(f"Reasoning verbosity: {task.reasoning_verbosity}")
        if task.session_scope_kind:
            parts.append(f"Session scope: {task.session_scope_kind}:{task.session_scope_id}")
        if task.parent_session_scope_kind:
            parts.append(f"Parent session scope: {task.parent_session_scope_kind}:{task.parent_session_scope_id}")
        if task.harness_profile_id:
            parts.append(f"Harness profile: {task.harness_profile_id}")
        if task.harness_overlay_version:
            parts.append(f"Effective harness overlay: {task.harness_overlay_version}")
        if task.memory_backend:
            parts.append(f"Memory backend: {task.memory_backend}")
        parts.append(f"Timeout seconds: {task.timeout_seconds}")
        execution_mode = (task.execution_mode or "").strip().lower()
        parts.append("Use only the effective tool allowlist above. Eval is read-only. Proposal investigate mode is read-only. Proposal diagnose mode is read-only and must stay grounded in persisted evidence before expanding to repo or log reads. Proposal implement mode may mutate only through governed workspace tools inside the bound workspace; it must not mutate GitHub directly, launch jobs, or post to Slack.")
        if execution_mode == "diagnose":
            parts.append(
                "Return a JSON object with keys: status, subsystem, failure_mode, summary, evidence_refs, missing_evidence, recommended_fix, target_surface, validation_plan."
            )
            parts.append(
                "status must be one of: grounded, needs_evidence, closed."
            )
            parts.append(
                "If the evidence does not ground a specific cause, return status=needs_evidence and use missing_evidence instead of guessing."
            )
        else:
            parts.append(
                "Return a JSON object with keys: visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, reply_delivery, knowledge_drafts, outcome_hypotheses, produced_artifacts, artifact_failure_reason, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta."
            )
            parts.append(
                "Each proposed action must include: kind, target_ref, request_payload, approval_mode, idempotency_key, rationale, evidence_refs."
            )
            parts.append(
                "Each knowledge draft must include: kind, scope_type, scope_id, title, summary, body, confidence, fresh_until, evidence_refs."
            )
            parts.append(
                "Each outcome hypothesis must include: outcome_type, success_condition, measurement_ref, expected_time_horizon."
            )
            parts.append(
                "Each produced artifact must include: kind, title, artifact_refs, delivery_status, failure_reason."
            )
        if execution_mode == "implement":
            parts.append(
                "For proposal implement tasks, use the bound workspace tools to inspect, edit, diff, and validate inside the workspace. repo_patch is optional legacy output only; the authoritative patch is the workspace git diff. If local validation succeeds and opening a draft PR is warranted, include exactly one proposed action with kind=draft_pr_open and request_payload containing title, body, branch_name, base_ref, and rationale."
            )
        elif execution_mode != "diagnose":
            parts.append(
                "For proposal or repo-change investigate tasks, change_plan must explain the concrete remediation, repo_patch should contain a unified diff when target_layer is repo_change, validation_plan must name the checks to run, retry_assessment must include failure_class, failure_summary, retry_decision, material_hypothesis_change, and changed_files, and hypothesis_delta must explain what changed from the prior failed attempt."
            )
        parts.append(f"Task prompt:\n{task.prompt}")
        return "\n".join(parts)

    def _extract_structured_output(self, message: str) -> JsonObject:
        text = (message or "").strip()
        if not text:
            raise HermesStructuredOutputError("Hermes execution returned an empty response; structured output is required.")
        try:
            parsed = json.loads(text)
        except json.JSONDecodeError as exc:
            raise HermesStructuredOutputError("Hermes execution returned non-JSON output; structured output is required.") from exc
        if not isinstance(parsed, dict):
            raise HermesStructuredOutputError("Hermes execution returned a non-object JSON payload; structured output must be a JSON object.")
        return _normalize_structured_output(parsed)

    def _structured_output_repairable(self, exc: HermesStructuredOutputError) -> bool:
        message = str(exc).lower()
        return "non-json" in message or "non-object json" in message

    def _build_structured_output_repair_task(self, task: RunnerTaskRequest, raw_response: str) -> RunnerTaskRequest:
        repair_prompt = "\n".join(
            [
                task.prompt,
                "",
                "Repair instruction: your previous response was invalid.",
                "Return only a valid JSON object matching the required schema.",
                "Do not include markdown, explanations, code fences, or any text before or after the JSON object.",
                "Previous invalid response:",
                raw_response,
            ]
        )
        return replace(task, prompt=repair_prompt)

    def _workflow_missing_explicit_reply_action(self, task: RunnerTaskRequest, structured_output: JsonObject) -> bool:
        if not self._workflow_requires_explicit_reply_action(task):
            return False
        final_answer = _string_or_json(structured_output.get("final_answer"))
        reply_draft = _string_or_json(structured_output.get("reply_draft"))
        if not final_answer and not reply_draft:
            return False
        for item in _normalize_proposed_actions(structured_output.get("proposed_actions")):
            if _string_or_json(item.get("kind")) == "slack_post":
                return False
        return True

    def _build_action_contract_repair_task(self, task: RunnerTaskRequest, structured_output: JsonObject) -> RunnerTaskRequest:
        repair_prompt = "\n".join(
            [
                task.prompt,
                "",
                "Repair instruction: your previous structured output included a grounded reply but omitted the required explicit slack_post action.",
                "Re-emit the full JSON object.",
                "Preserve the final_answer, reply_draft, produced_artifacts, and artifact_failure_reason unless a correction is required.",
                "Include exactly one proposed action with kind=slack_post and a request_payload that carries the reply body.",
                "Return only valid JSON with no markdown or surrounding commentary.",
                "Previous structured output:",
                json.dumps(structured_output, ensure_ascii=True, sort_keys=True),
            ]
        )
        return replace(task, prompt=repair_prompt)


def first_non_empty(*values: str | None) -> str:
    for value in values:
        if value and value.strip():
            return value.strip()
    return ""


def tool_name(schema: JsonToolWrapperSchema) -> str:
    function = schema.get("function", {})
    value = function.get("name", "")
    return str(value).strip()


def safe_activity_summary(agent: Any) -> JsonObject:
    getter = getattr(agent, "get_activity_summary", None)
    if not callable(getter):
        return {}
    summary = getter()
    if isinstance(summary, dict):
        return summary
    raise HermesStructuredOutputError("Hermes agent.get_activity_summary() returned a non-dict payload.")
def string_from_map(values: JsonObject, key: str) -> str:
    value = values.get(key, "")
    return str(value or "").strip()


def inactivity_seconds(activity: JsonObject, fallback_elapsed_seconds: float) -> float:
    raw = activity.get("seconds_since_activity")
    if isinstance(raw, Number):
        return float(raw)
    try:
        return float(str(raw).strip())
    except (AttributeError, TypeError, ValueError):
        return max(0.0, fallback_elapsed_seconds)
