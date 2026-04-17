from __future__ import annotations

import concurrent.futures
from dataclasses import dataclass, replace
import json
import logging
import os
import re
import time
from numbers import Number
from typing import Any

from .json_types import JsonObject, JsonToolWrapperSchema, JsonValue

from .config import RunnerConfig
from .hermes_adapter import HermesAdapter
from .rsi_tools import (
    BLOCKED_HONCHO_TOOLS,
    CompositeToolProvider,
    HERMES_GOVERNED_READONLY_TOOLSET,
    HERMES_GOVERNED_WORKSPACE_TOOLSET,
    IMPLEMENT_RSI_TOOL_NAMES,
    READ_ONLY_HONCHO_TOOLS,
    READ_ONLY_RSI_TOOL_NAMES,
    ReadOnlyToolBinding,
    WORKSPACE_RSI_TOOL_NAMES,
    normalize_tool_names,
    tool_transport_name,
    tool_schema_wrappers,
)
from .session_manager import SessionContext, SessionManager

ROLE_TASK_TYPES = {
    "prod": {"general", "workflow", "prod"},
    "proactive": {"general", "proactive"},
    "eval": {"general", "eval"},
    "proposal": {"general", "proposal", "repo-change"},
}

logger = logging.getLogger(__name__)

NATIVE_HERMES_DIAGNOSE_TOOLS = frozenset({"todo", "session_search"})


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


def _normalize_structured_output(payload: JsonObject) -> JsonObject:
    normalized = dict(payload)
    normalized["context_summary"] = _string_or_json(payload.get("context_summary"))
    normalized["reply_draft"] = _string_or_json(payload.get("reply_draft"))
    normalized["final_answer"] = _string_or_json(payload.get("final_answer"))
    normalized["confidence"] = _float_or_zero(payload.get("confidence"))
    normalized["self_critique"] = _string_or_json(payload.get("self_critique"))
    normalized["visible_reasoning"] = _normalize_visible_reasoning(payload.get("visible_reasoning"))
    normalized["proposed_actions"] = _normalize_proposed_actions(payload.get("proposed_actions"))
    normalized["knowledge_drafts"] = _normalize_knowledge_drafts(payload.get("knowledge_drafts"))
    normalized["outcome_hypotheses"] = _normalize_outcome_hypotheses(payload.get("outcome_hypotheses"))
    normalized["change_plan"] = _string_or_json(payload.get("change_plan"))
    normalized["repo_patch"] = _string_or_json(payload.get("repo_patch"))
    normalized["validation_plan"] = _string_or_json(payload.get("validation_plan"))
    normalized["retry_assessment"] = _normalize_retry_assessment(payload.get("retry_assessment"))
    normalized["hypothesis_delta"] = _string_or_json(payload.get("hypothesis_delta"))
    return normalized


def _optional_string(value: JsonValue | None) -> str | None:
    if value is None:
        return None
    if isinstance(value, str):
        text = value.strip()
        return text or None
    return str(value)


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
except (ImportError, ModuleNotFoundError):  # pragma: no cover - import depends on external Hermes install
    AIAgent = None

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
class RunnerTaskRequest:
    task_type: str
    repo: str
    repo_ref: str | None
    prompt: str
    system_message: str | None
    allowed_tools: list[str]
    allowed_commands: list[str]
    timeout_seconds: int
    expected_outputs: list[str]
    artifact_destination: str | None
    context_summary: str | None
    rejected_proposal_context: list[JsonObject]
    execution_mode: str | None
    intent: str | None
    trace_id: str | None
    workflow_id: str | None
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

    @classmethod
    def from_payload(cls, payload: JsonObject) -> "RunnerTaskRequest":
        task = _json_object_or_empty(payload.get("task")) or payload
        return cls(
            task_type=_required_string(task.get("task_type"), "general"),
            repo=_required_string(task.get("repo"), "rsi-agent-platform"),
            repo_ref=_optional_string(task.get("repo_ref")),
            prompt=_required_string(_first_non_none(task.get("prompt"), payload.get("prompt")), ""),
            system_message=_optional_string(_first_non_none(task.get("system_message"), payload.get("system_message"))),
            allowed_tools=_string_list(task.get("allowed_tools")),
            allowed_commands=_string_list(task.get("allowed_commands")),
            timeout_seconds=int(task.get("timeout_seconds", 900)),
            expected_outputs=_string_list(task.get("expected_outputs")),
            artifact_destination=_optional_string(task.get("artifact_destination")),
            context_summary=_optional_string(task.get("context_summary")),
            rejected_proposal_context=_json_object_list(task.get("rejected_proposal_context")),
            execution_mode=_optional_string(task.get("execution_mode")),
            intent=_optional_string(task.get("intent")),
            trace_id=_optional_string(task.get("trace_id")),
            workflow_id=_optional_string(task.get("workflow_id")),
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
        self._max_iterations = config.max_iterations
        self._default_task_timeout_seconds = config.task_timeout_seconds
        self._default_inactivity_timeout_seconds = config.inactivity_timeout_seconds
        self._transport_timeout_seconds = config.transport_timeout_seconds
        self._tool_policy_mode = config.tool_policy_mode
        self._configure_runtime()
        self._available = AIAgent is not None and self._runtime_has_credentials() and self._session_manager.available

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

    @property
    def metadata(self) -> JsonObject:
        adapter_meta = self._adapter.metadata
        return {
            "status": "ok" if self.available else "degraded",
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
            "persistence_enabled": self._session_manager.available,
            "session_continuity_status": "ok" if self._session_manager.available else "degraded",
            "hermes_home": self._session_manager.hermes_home,
            "session_db_path": self._session_manager.session_db_path,
            "hermes_version": adapter_meta.version,
            "hermes_pin": adapter_meta.pin,
            "memory_backend": self._config.memory_backend,
            "max_iterations": self._max_iterations,
            "task_timeout_seconds": self._default_task_timeout_seconds,
            "inactivity_timeout_seconds": self._default_inactivity_timeout_seconds,
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "tool_policy_mode": self._tool_policy_mode,
            "tool_allowlist_effective": self._default_policy_allowlist(execution_mode=""),
            "blocked_tool_names": [],
            "context_engine_mode": adapter_meta.context_engine_mode,
            "context_engine_status": adapter_meta.context_engine_status,
            "lifecycle_hook_status": adapter_meta.lifecycle_hook_status,
            "governed_tools_status": adapter_meta.governed_tools_status,
            "honcho_configured": self._config.honcho_api_key_configured or bool(self._config.honcho_base_url),
            "honcho_available": self._session_manager.honcho_available,
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
        return self._execute_task_request(task, self._resolve_tool_policy(task))

    def _native_governed_tools_enabled(self, task: RunnerTaskRequest) -> bool:
        if not self._config.hermes_native_governed_tools_enabled:
            return False
        if self._role != "proposal":
            return False
        return (task.execution_mode or "").strip().lower() in {"diagnose", "implement"}

    def _native_toolsets_for_task(self, task: RunnerTaskRequest) -> list[str]:
        if not self._native_governed_tools_enabled(task):
            return []
        execution_mode = (task.execution_mode or "").strip().lower()
        toolsets = ["todo", "session_search", HERMES_GOVERNED_READONLY_TOOLSET]
        if execution_mode == "implement":
            toolsets.append(HERMES_GOVERNED_WORKSPACE_TOOLSET)
        return toolsets

    def _create_agent(self, task: RunnerTaskRequest, context: SessionContext) -> Any:
        agent_kwargs: JsonObject = {
            "model": self._provider_model,
            "quiet_mode": True,
            "reasoning_config": self._reasoning_config,
            "enabled_toolsets": self._native_toolsets_for_task(task),
            "skip_context_files": True,
            "skip_memory": False,
            "persist_session": True,
            "max_iterations": self._max_iterations,
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

    def _execute_task_request(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> HermesExecutionResult:
        if AIAgent is None:
            return HermesExecutionResult(
                ok=False,
                message="Hermes runtime is not installed in this environment.",
                provider=self._backend,
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
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
        effective_task_timeout = self._effective_task_timeout(task)
        effective_inactivity_timeout = self._effective_inactivity_timeout(effective_task_timeout)
        agent = None
        try:
            self._adapter.stage_task_context(
                context.session_id,
                {
                    "role": self._role,
                    "task_type": task.task_type,
                    "task_repo": task.repo,
                    "task_repo_ref": task.repo_ref or "",
                    "task_prompt": task.prompt,
                    "trace_id": task.trace_id,
                    "workflow_id": task.workflow_id,
                    "task_channel_id": task.channel_id or "",
                    "task_thread_ts": task.thread_ts or "",
                    "proposal_id": task.session_scope_id if (task.session_scope_kind or "").strip() == "proposal_candidate" else "",
                    "attempt_id": task.attempt_id,
                    "workspace_id": task.workspace_id,
                    "execution_mode": task.execution_mode or "",
                    "context_summary": task.context_summary or "",
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
            agent = self._create_agent(task, context)
            self._attach_tool_policy(agent, task, tool_policy)
            tracker = self._session_manager.attach_tracking(agent, task, context)
            interrupted, run_result, run_meta = self._run_with_deadlines(
                agent,
                task,
                context,
                effective_task_timeout,
                effective_inactivity_timeout,
            )
            lifecycle_events = self._adapter.lifecycle_events(context.session_id)
            if interrupted:
                finalized = self._session_manager.finalize(context, tracker)
                observed = self._observability_metadata(agent, task, tracker)
                termination_reason = string_from_map(run_meta, "termination_reason")
                timeout_kind = string_from_map(run_meta, "timeout_kind")
                timeout_message = f"Hermes execution timed out after {effective_task_timeout}s."
                failure_class = "runner_transport_timeout"
                failure_kind = "transport_timeout"
                if timeout_kind == "inactivity_timeout":
                    timeout_message = f"Hermes execution hit inactivity timeout after {effective_inactivity_timeout}s."
                elif termination_reason == "iteration_budget_exhausted":
                    timeout_message = f"Hermes execution exhausted max iterations ({self._max_iterations})."
                    failure_class = "runner_iteration_budget_exhausted"
                    failure_kind = "iteration_budget_exhausted"
                return HermesExecutionResult(
                    ok=False,
                    message=timeout_message,
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        **finalized,
                        **run_meta,
                        "task_timeout_seconds": effective_task_timeout,
                        "inactivity_timeout_seconds": effective_inactivity_timeout,
                        "transport_timeout_seconds": self._transport_timeout_seconds,
                        "max_iterations": self._max_iterations,
                        "tool_policy_mode": tool_policy.mode,
                        "tool_allowlist_effective": tool_policy.effective,
                        "tool_transport_allowlist_effective": tool_policy.transport_effective,
                        "blocked_tool_names": tool_policy.blocked,
                        **observed,
                        "failure_class": failure_class,
                        "runner_diagnostics": self._runner_diagnostics(
                            tool_policy,
                            failure_kind=failure_kind,
                            provider_error_message=timeout_message,
                            timeout_kind=timeout_kind or None,
                            termination_reason=first_non_empty(termination_reason, timeout_kind, "task_timeout"),
                            activity=_json_object_or_empty(run_meta.get("last_activity")),
                            max_iterations_reached=bool(run_meta.get("max_iterations_reached")),
                            session_ready_issues=self._session_manager.ready_issues,
                            repair_attempted=False,
                            repair_succeeded=False,
                            observed=observed,
                        ),
                        "lifecycle_events": lifecycle_events,
                        "termination_reason": first_non_empty(termination_reason, timeout_kind, "task_timeout"),
                    },
                )
            response = str((run_result or {}).get("final_response", "") or "")
        except Exception as exc:
            diagnostics = self._provider_invalid_request_diagnostics(str(exc), tool_policy)
            activity = safe_activity_summary(agent) if agent is not None else {}
            observed = self._observability_metadata(agent, task)
            raw = {
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                "error": str(exc),
                "tool_policy_mode": tool_policy.mode,
                "tool_allowlist_effective": tool_policy.effective,
                "tool_transport_allowlist_effective": tool_policy.transport_effective,
                "tool_transport_map": tool_policy.custom_tool_transport_map,
                "blocked_tool_names": tool_policy.blocked,
                **observed,
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
            return HermesExecutionResult(
                ok=False,
                message=f"Hermes execution failed: {exc}",
                provider=self._backend,
                raw=raw,
            )

        finalized = self._session_manager.finalize(context, tracker)
        lifecycle_events = self._adapter.lifecycle_events(context.session_id)
        observed = self._observability_metadata(agent, task, tracker)
        return HermesExecutionResult(
            ok=True,
            message=response,
            provider=self._backend,
            raw={
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                **finalized,
                "task_timeout_seconds": effective_task_timeout,
                "inactivity_timeout_seconds": effective_inactivity_timeout,
                "transport_timeout_seconds": self._transport_timeout_seconds,
                "max_iterations": self._max_iterations,
                "tool_policy_mode": tool_policy.mode,
                "tool_allowlist_effective": tool_policy.effective,
                "tool_transport_allowlist_effective": tool_policy.transport_effective,
                "tool_transport_map": tool_policy.custom_tool_transport_map,
                "blocked_tool_names": tool_policy.blocked,
                **observed,
                "runner_diagnostics": observed,
                "lifecycle_events": lifecycle_events,
                "termination_reason": "normal_completion",
                "max_iterations_reached": bool(run_meta.get("max_iterations_reached")),
                "harness_profile_id": task.harness_profile_id,
                "effective_overlay_version": task.harness_overlay_version,
            },
        )

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
            "context_engine_mode": adapter_meta.context_engine_mode,
            "context_engine_status": adapter_meta.context_engine_status,
            "lifecycle_hook_status": adapter_meta.lifecycle_hook_status,
            "governed_tools_status": adapter_meta.governed_tools_status,
            "base_url": self._base_url,
            "honcho_base_url": self._config.honcho_base_url or "",
            "honcho_workspace": self._config.honcho_workspace,
            "honcho_environment": self._config.honcho_environment,
            "honcho_environment_effective": self._config.honcho_environment_effective,
            "honcho_recall_mode": self._config.honcho_recall_mode,
            "honcho_write_frequency": self._config.honcho_write_frequency,
            "honcho_session_strategy": self._config.honcho_session_strategy,
            "honcho_ai_peer": self._config.honcho_ai_peer,
            "prompt": prompt,
            "system_message": system_message,
        }

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

    def _observability_metadata(self, agent: Any | None, task: RunnerTaskRequest, tracker: Any | None = None) -> JsonObject:
        observed: JsonObject = {
            "candidate_read_surfaces": self._candidate_read_surfaces_for_task(task),
            "selected_context_surfaces": [],
            "memory_warnings": [],
        }
        binding = getattr(agent, "_rsi_readonly_tool_binding", None) if agent is not None else None
        diagnostics = getattr(binding, "diagnostics", None)
        if callable(diagnostics):
            payload = diagnostics()
            if isinstance(payload, dict):
                observed.update(payload)
        if tracker is not None and hasattr(tracker, "warnings"):
            observed["memory_warnings"] = list(getattr(tracker, "warnings", []) or [])
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
        permitted = set(self._default_policy_allowlist(execution_mode=execution_mode))
        effective = normalize_tool_names(requested or sorted(permitted))
        effective = [name for name in effective if name in permitted]
        blocked = [name for name in requested if name not in permitted]
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

    def _attach_tool_policy(self, agent: Any, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> None:
        current_tools = list(getattr(agent, "tools", []) or [])
        current_valid = set(getattr(agent, "valid_tool_names", set()) or set())
        setattr(agent, "_rsi_readonly_tool_binding", None)
        native_governed_tools = self._native_governed_tools_enabled(task)
        allowed_names = set(tool_policy.effective)
        filtered_tools = current_tools
        if native_governed_tools:
            allowed_transport_names = {
                _transport_name_or_self(name)
                for name in tool_policy.effective
                if name not in BLOCKED_HONCHO_TOOLS
            }
            filtered_tools = [tool for tool in current_tools if tool_name(tool) in allowed_transport_names]
            agent.tools = filtered_tools
            agent.valid_tool_names = {name for name in current_valid if name in allowed_transport_names}
            return
        filtered_tools = [tool for tool in current_tools if tool_name(tool) in allowed_names]
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
                task_channel_id=task.channel_id or "",
                task_thread_ts=task.thread_ts or "",
                task_context_summary=task.context_summary or "",
                trace_id=task.trace_id or "",
                session_scope_kind=task.session_scope_kind or "",
                session_scope_id=task.session_scope_id or "",
                context_refs=task.context_refs,
                execution_mode=task.execution_mode or "",
                attempt_id=task.attempt_id or "",
                workspace_id=task.workspace_id or "",
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
                task_channel_id=task.channel_id or "",
                task_thread_ts=task.thread_ts or "",
                task_context_summary=task.context_summary or "",
                trace_id=task.trace_id or "",
                session_scope_kind=task.session_scope_kind or "",
                session_scope_id=task.session_scope_id or "",
                context_refs=task.context_refs,
                execution_mode=task.execution_mode or "",
                attempt_id=task.attempt_id or "",
                workspace_id=task.workspace_id or "",
            )
            setattr(agent, "_rsi_readonly_tool_binding", readonly_tools)
            agent._memory_manager = CompositeToolProvider(getattr(agent, "_memory_manager", None), readonly_tools)
        effective_names = set(tool_policy.effective)
        current_valid = {name for name in current_valid if name in effective_names and name not in BLOCKED_HONCHO_TOOLS}
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
    ) -> tuple[bool, JsonObject | None, JsonObject]:
        executor = concurrent.futures.ThreadPoolExecutor(max_workers=1)
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
                    return False, result, {
                        "last_activity": activity,
                        "last_tool_invoked": string_from_map(activity, "current_tool"),
                        "max_iterations_reached": _max_iterations_reached(activity),
                    }
                except concurrent.futures.TimeoutError:
                    activity = safe_activity_summary(agent)
                    if _max_iterations_reached(activity):
                        return self._interrupt_execution(
                            agent,
                            future,
                            "iteration_budget_exhausted",
                            int(activity.get("budget_max", 0) or self._max_iterations),
                            activity,
                        )
                    elapsed_seconds = max(0.0, time.monotonic() - started_at)
                    idle_seconds = inactivity_seconds(activity, elapsed_seconds)
                    if elapsed_seconds >= float(timeout_seconds):
                        return self._interrupt_execution(
                            agent,
                            future,
                            "task_timeout",
                            timeout_seconds,
                            activity,
                        )
                    if inactivity_timeout_seconds > 0 and idle_seconds >= float(inactivity_timeout_seconds):
                        return self._interrupt_execution(
                            agent,
                            future,
                            "inactivity_timeout",
                            inactivity_timeout_seconds,
                            activity,
                        )
        finally:
            executor.shutdown(wait=False, cancel_futures=True)

    def _interrupt_execution(
        self,
        agent: Any,
        future: concurrent.futures.Future,
        termination_reason: str,
        threshold_value: int,
        activity: JsonObject,
    ) -> tuple[bool, JsonObject | None, JsonObject]:
        interrupt_message = f"runner {termination_reason} after {threshold_value}s"
        if termination_reason == "iteration_budget_exhausted":
            interrupt_message = f"runner iteration budget exhausted at {threshold_value} iterations"
        agent.interrupt(interrupt_message)
        shutdown_error = ""
        try:
            future.result(timeout=min(5, max(1, threshold_value//10)))
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
            "max_iterations_reached": _max_iterations_reached(latest_activity),
        }
        if termination_reason in {"task_timeout", "inactivity_timeout"}:
            meta["timeout_kind"] = termination_reason
            meta["timed_out_after_seconds"] = threshold_value
        if termination_reason == "iteration_budget_exhausted":
            meta["max_iterations"] = threshold_value
        if shutdown_error:
            meta["shutdown_error"] = shutdown_error
        return True, None, meta

    def execute_task(self, task: RunnerTaskRequest) -> HermesExecutionResult:
        if task.task_type not in ROLE_TASK_TYPES.get(self._role, {self._role}):
            return HermesExecutionResult(
                ok=False,
                message=f"Runner role {self._role} cannot execute task type {task.task_type}.",
                provider="policy",
                raw={"role": self._role, "task_type": task.task_type},
            )
        tool_policy = self._resolve_tool_policy(task)
        prompt = self._render_task_prompt(task, tool_policy)
        rendered_task = replace(task, prompt=prompt)
        result = self._execute_task_request(rendered_task, tool_policy)
        if not result.ok:
            return result
        if invalid_request := self._provider_invalid_request_diagnostics(result.message, tool_policy):
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
        initial_response = result.message
        repair_attempted = False
        repair_succeeded = False
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
                repair_result = self._execute_task_request(repair_task, tool_policy)
                if repair_result.ok:
                    if invalid_request := self._provider_invalid_request_diagnostics(repair_result.message, tool_policy):
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
                                    tool_policy,
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
                            tool_policy,
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
        action_contract_repair_attempted = False
        action_contract_repair_succeeded = False
        action_contract_repair_error = ""
        action_contract_repair_response = ""
        if self._workflow_missing_explicit_reply_action(task, structured_output):
            action_contract_repair_attempted = True
            logger.info(
                "workflow runner action-contract repair attempted trace_id=%s workflow_id=%s",
                task.trace_id or "",
                task.workflow_id or "",
            )
            repair_task = self._build_action_contract_repair_task(rendered_task, structured_output)
            repair_result = self._execute_task_request(repair_task, tool_policy)
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
            "channel_id": task.channel_id,
            "thread_ts": task.thread_ts,
            "repo_allowlist": task.repo_allowlist,
            "tool_allowlist": task.tool_allowlist,
            "tool_policy_mode": tool_policy.mode,
            "tool_allowlist_effective": tool_policy.effective,
            "blocked_tool_names": tool_policy.blocked,
            "response_mode": task.response_mode,
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
            "repair_attempted": repair_attempted,
            "repair_succeeded": repair_succeeded,
            "action_contract_repair_attempted": action_contract_repair_attempted,
            "action_contract_repair_succeeded": action_contract_repair_succeeded,
            "structured_output": structured_output,
        }
        if repair_attempted:
            result.raw["repair_original_response"] = initial_response
        runner_diagnostics = _json_object_or_empty(result.raw.get("runner_diagnostics"))
        runner_diagnostics["candidate_read_surfaces"] = result.raw.get("candidate_read_surfaces", [])
        runner_diagnostics["selected_context_surfaces"] = result.raw.get("selected_context_surfaces", [])
        runner_diagnostics["memory_warnings"] = result.raw.get("memory_warnings", [])
        runner_diagnostics["action_contract_repair_attempted"] = action_contract_repair_attempted
        runner_diagnostics["action_contract_repair_succeeded"] = action_contract_repair_succeeded
        if action_contract_repair_error:
            result.raw["action_contract_repair_error"] = action_contract_repair_error
            runner_diagnostics["action_contract_repair_error"] = action_contract_repair_error
        if action_contract_repair_response:
            result.raw["action_contract_repair_response"] = action_contract_repair_response
            runner_diagnostics["action_contract_repair_response"] = action_contract_repair_response
        result.raw["runner_diagnostics"] = runner_diagnostics
        return result

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
        if task.rejected_proposal_context:
            parts.append(f"Prior rejected/dismissed context: {json.dumps(task.rejected_proposal_context)}")
        if task.response_mode:
            parts.append(f"Response mode: {task.response_mode}")
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
                "Return a JSON object with keys: visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, outcome_hypotheses, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta."
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
        if task.task_type != "workflow":
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
                "Preserve the final_answer and reply_draft unless a correction is required.",
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


def _max_iterations_reached(activity: JsonObject) -> bool:
    try:
        budget_used = int(activity.get("budget_used", 0) or 0)
        budget_max = int(activity.get("budget_max", 0) or 0)
    except (TypeError, ValueError):
        return False
    return budget_max > 0 and budget_used >= budget_max


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
