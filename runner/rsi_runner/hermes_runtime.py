from __future__ import annotations

import concurrent.futures
from dataclasses import dataclass, replace
import hashlib
import json
import logging
import os
import re
import socket
import time
from numbers import Number
from typing import Any
from urllib import error as urlerror
from urllib import request as urlrequest

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
    READ_ONLY_WORKSPACE_RSI_TOOL_NAMES,
    ReadOnlyToolBinding,
    WORKSPACE_RSI_TOOL_NAMES,
    normalize_tool_names,
    transport_tool_schema,
    tool_transport_name,
    tool_schema_wrappers,
)
from .session_manager import SessionContext, SessionManager

ROLE_TASK_TYPES = {
    "prod": {"general", "workflow", "prod", "question_expand", "question_reduce"},
    "proactive": {"general", "workflow", "proactive", "question_expand", "question_reduce"},
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
class PartialReducerAttemptResult:
    ok: bool
    response_text: str
    structured_output: JsonObject
    error: str
    provider_response_id: str


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
    mcp_servers: list[JsonObject]

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
            mcp_servers=_json_object_list(task.get("mcp_servers")),
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
        self._slack_mcp_discovery_error = ""
        self._slack_mcp_tool_cache: list[JsonObject] | None = None
        self._slack_mcp_send_tool_name = ""
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
            "slack_mcp_enabled": self._config.slack_mcp_enabled,
            "slack_mcp_configured": self._config.slack_mcp_enabled and self._config.slack_user_token_configured,
            "slack_mcp_available": self._slack_mcp_available(),
            "slack_mcp_server_url": self._config.slack_mcp_server_url,
            "slack_mcp_tool_count": len(self._slack_mcp_tools()),
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

    def _create_agent(
        self,
        task: RunnerTaskRequest,
        context: SessionContext,
        *,
        max_iterations_override: int | None = None,
    ) -> Any:
        configured_max_iterations = max_iterations_override if max_iterations_override and max_iterations_override > 0 else self._max_iterations
        agent_kwargs: JsonObject = {
            "model": self._provider_model,
            "quiet_mode": True,
            "reasoning_config": self._reasoning_config,
            "enabled_toolsets": self._native_toolsets_for_task(task),
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

    def _execute_task_request(
        self,
        task: RunnerTaskRequest,
        tool_policy: ToolPolicy,
        *,
        allow_partial_recovery: bool = True,
        max_iterations_override: int | None = None,
    ) -> HermesExecutionResult:
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
        reasoning_timeout_seconds = self._partial_completion_reasoning_timeout_seconds(task, effective_task_timeout)
        configured_max_iterations = max_iterations_override if max_iterations_override and max_iterations_override > 0 else self._max_iterations
        agent = None
        tracker = None
        try:
            self._stage_task_context(context.session_id, task, tool_policy)
            agent = self._create_agent(task, context, max_iterations_override=max_iterations_override)
            self._attach_tool_policy(agent, task, tool_policy)
            tracker = self._session_manager.attach_tracking(agent, task, context)
            termination_reason, run_result, stop_meta = self._run_with_deadlines(
                agent,
                task,
                context,
                effective_task_timeout,
                effective_inactivity_timeout,
                reasoning_timeout_seconds,
            )
            lifecycle_events = self._adapter.lifecycle_events(context.session_id)
            if termination_reason != "normal_completion":
                finalized = self._session_manager.finalize(context, tracker)
                observed = self._observability_metadata(agent, task, tracker)
                last_activity = _json_object_or_empty(stop_meta.get("last_activity"))
                if termination_reason in {"task_timeout", "iteration_budget_exhausted"} and allow_partial_recovery and self._workflow_partial_completion_eligible(task):
                    return self._finalize_partial_completion(
                        task,
                        tool_policy,
                        finalized,
                        observed,
                        stop_meta,
                        lifecycle_events,
                        termination_reason=termination_reason,
                    )
                if termination_reason == "inactivity_timeout":
                    timeout_kind = string_from_map(stop_meta, "timeout_kind") or termination_reason
                    timeout_message = f"Hermes execution timed out after {effective_task_timeout}s."
                    if timeout_kind == "inactivity_timeout":
                        timeout_message = f"Hermes execution hit inactivity timeout after {effective_inactivity_timeout}s."
                    return HermesExecutionResult(
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
                if termination_reason == "task_timeout":
                    timeout_message = f"Hermes execution timed out after {effective_task_timeout}s."
                    return HermesExecutionResult(
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
                if termination_reason == "iteration_budget_exhausted":
                    return self._partial_completion_failure(
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
        using_native_mcp = self._task_uses_native_mcp(task)
        if task.task_type == "question_reduce" and not using_native_mcp:
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
        fallback_allowlist = sorted(permitted)
        if task.task_type == "question_expand" and requested:
            fallback_allowlist = requested
        if using_native_mcp:
            fallback_allowlist = [name for name in fallback_allowlist if name not in {"slack.history", "slack.search", "slack.reply"} and name not in READ_ONLY_HONCHO_TOOLS]
        effective = normalize_tool_names(requested or fallback_allowlist)
        if using_native_mcp:
            effective = [name for name in effective if name not in {"slack.history", "slack.search", "slack.reply"} and name not in READ_ONLY_HONCHO_TOOLS]
        effective = [name for name in effective if name in permitted]
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
                "task_channel_id": task.channel_id or "",
                "task_thread_ts": task.thread_ts or "",
                "proposal_id": task.session_scope_id if (task.session_scope_kind or "").strip() == "proposal_candidate" else "",
                "attempt_id": task.attempt_id,
                "workspace_id": task.workspace_id,
                "execution_mode": task.execution_mode or "",
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
        if task.task_type not in {"question_expand", "question_reduce"}:
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
            "user_request": first_non_empty(_string_or_json(spec.get("user_request")), _string_or_json(input_ledger.get("user_request")), self._original_task_prompt(task)),
            "reply_target": input_ledger.get("reply_target")
            or {
                "channel_id": task.channel_id or "",
                "thread_ts": task.thread_ts or "",
            },
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
        if task.task_type == "question_expand":
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
                "Return only one JSON object with keys: visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, outcome_hypotheses, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta.",
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
            "input": user_prompt,
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

    def _invoke_direct_json_response(
        self,
        *,
        system_prompt: str,
        user_prompt: str,
        timeout_seconds: int,
        reasoning_effort: str = "low",
    ) -> PartialReducerAttemptResult:
        if self._provider != "openai" or not self._api_key:
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
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error=f"Direct reducer returned {exc.code}: {detail[:2000]}",
                provider_response_id="",
            )
        except (TimeoutError, socket.timeout) as exc:
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error=f"Direct reducer timed out after {max(1, timeout_seconds)}s: {exc}",
                provider_response_id="",
            )
        except (urlerror.URLError, ConnectionError, OSError) as exc:
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
            return PartialReducerAttemptResult(
                ok=False,
                response_text=body[:4000],
                structured_output={},
                error="Direct reducer returned invalid JSON.",
                provider_response_id="",
            )
        if not isinstance(parsed, dict):
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error="Direct reducer returned a non-object JSON payload.",
                provider_response_id="",
            )
        response_text = self._responses_output_text(parsed)
        provider_response_id = _string_or_json(parsed.get("id"))
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

    def _task_uses_native_mcp(self, task: RunnerTaskRequest) -> bool:
        if self._provider != "openai":
            return False
        if task.task_type not in {"workflow", "question_expand", "question_reduce"}:
            return False
        return len(task.mcp_servers) > 0

    def _task_slack_mcp_write_enabled(self, task: RunnerTaskRequest) -> bool:
        for server in task.mcp_servers:
            if _string_or_json(server.get("profile")) == "slack_mcp_reply":
                return True
        return False

    def _direct_task_user_prompt(self, task: RunnerTaskRequest) -> str:
        if task.task_type in {"question_expand", "question_reduce"}:
            return task.prompt
        payload: JsonObject = {
            "user_request": self._original_task_prompt(task),
            "reply_target": {
                "channel_id": task.channel_id or "",
                "thread_ts": task.thread_ts or "",
            },
            "context_summary": task.context_summary or "",
            "recent_conversation_entries": task.recent_conversation_entries,
            "case_summary": task.case_summary or {},
            "prior_trace_refs": task.prior_trace_refs,
            "context_refs": task.context_refs,
            "response_mode": task.response_mode or "",
            "trace_id": task.trace_id or "",
            "workflow_id": task.workflow_id or "",
        }
        return "\n".join(
            [
                "Complete the following Slack-bound RSI workflow.",
                "Use Slack MCP for Slack investigation and governed function tools for repo, GitHub, knowledge, RSI context, and workspace reads.",
                "Return JSON only.",
                json.dumps(payload, ensure_ascii=True, sort_keys=True, indent=2),
            ]
        )

    def _direct_function_tools(self, tool_policy: ToolPolicy) -> list[JsonObject]:
        out: list[JsonObject] = []
        for name in tool_policy.custom_tools:
            schema = transport_tool_schema(name)
            out.append(
                {
                    "type": "function",
                    "name": _string_or_json(schema.get("name")),
                    "description": _string_or_json(schema.get("description")),
                    "parameters": _json_object_or_empty(schema.get("parameters")),
                    "strict": True,
                }
            )
        return out

    def _resolved_task_mcp_servers(self, task: RunnerTaskRequest) -> tuple[list[JsonObject], set[str]]:
        resolved: list[JsonObject] = []
        send_tool_names: set[str] = set()
        for server in task.mcp_servers:
            label = first_non_empty(_string_or_json(server.get("server_label")), "slack")
            profile = _string_or_json(server.get("profile"))
            server_url = first_non_empty(_string_or_json(server.get("server_url")), self._config.slack_mcp_server_url)
            item: JsonObject = {
                "type": "mcp",
                "server_label": label,
                "server_url": server_url,
                "require_approval": server.get("require_approval") or "never",
            }
            headers = _json_object_or_empty(server.get("headers"))
            if headers:
                item["headers"] = headers
            authorization = first_non_empty(_string_or_json(server.get("authorization")), os.getenv("RSI_SLACK_USER_TOKEN"), "")
            if not authorization:
                raise RuntimeError("Slack user token is not configured.")
            item["authorization"] = authorization
            if profile == "slack_mcp_read":
                item["allowed_tools"] = {"read_only": True}
            elif profile == "slack_mcp_reply":
                send_tool_name = self._slack_mcp_send_tool_name_or_error()
                send_tool_names.add(send_tool_name)
                read_tool_names = []
                for tool in self._slack_mcp_tools():
                    annotations = _json_object_or_empty(tool.get("annotations"))
                    if _bool_or_false(annotations.get("readOnlyHint")):
                        read_tool_names.append(_string_or_json(tool.get("name")))
                item["allowed_tools"] = {"tool_names": normalize_tool_names([*read_tool_names, send_tool_name])}
            else:
                allowed_tools = _json_object_or_empty(server.get("allowed_tools"))
                if allowed_tools:
                    item["allowed_tools"] = allowed_tools
            resolved.append(item)
        return resolved, send_tool_names

    def _responses_function_calls(self, payload: JsonObject) -> list[JsonObject]:
        calls: list[JsonObject] = []
        for item in _json_object_list(payload.get("output")):
            item_type = _string_or_json(item.get("type"))
            if item_type not in {"function_call", "custom_tool_call"}:
                continue
            calls.append(item)
        return calls

    def _normalize_mcp_calls(
        self,
        payload: JsonObject,
        *,
        send_tool_names: set[str],
        task: RunnerTaskRequest,
    ) -> tuple[list[JsonObject], JsonObject | None, bool]:
        normalized: list[JsonObject] = []
        reply_delivery: JsonObject | None = None
        reply_attempted = False
        for item in _json_object_list(payload.get("output")):
            if _string_or_json(item.get("type")) != "mcp_call":
                continue
            name = _string_or_json(item.get("name"))
            arguments_text = _string_or_json(item.get("arguments"))
            request_payload = {}
            if arguments_text:
                try:
                    parsed_args = json.loads(arguments_text)
                except json.JSONDecodeError:
                    parsed_args = {}
                if isinstance(parsed_args, dict):
                    request_payload = parsed_args
            status = _string_or_json(item.get("status")) or "completed"
            output_text = _string_or_json(item.get("output"))
            error_text = _string_or_json(item.get("error"))
            summary = first_non_empty(error_text, output_text, status)[:400]
            call_id = first_non_empty(_string_or_json(item.get("id")), _string_or_json(item.get("call_id")), name)
            normalized.append(
                {
                    "id": first_non_empty(_string_or_json(item.get("id")), f"runner-tool-record-{call_id}"),
                    "tool_name": f"slack.mcp.{name}",
                    "tool_call_id": call_id,
                    "request": request_payload,
                    "summary": summary,
                    "approval_state": "not_required",
                    "interpretation_summary": summary,
                    "status": status,
                    "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                }
            )
            if name in send_tool_names:
                reply_attempted = True
                if status in {"completed", "complete", "ok", "success"} and reply_delivery is None:
                    channel_id = first_non_empty(
                        _string_or_json(request_payload.get("channel_id")),
                        _string_or_json(request_payload.get("channel")),
                        _string_or_json(request_payload.get("conversation_id")),
                        task.channel_id or "",
                    )
                    thread_ts = first_non_empty(
                        _string_or_json(request_payload.get("thread_ts")),
                        _string_or_json(request_payload.get("thread")),
                        task.thread_ts or "",
                    )
                    body = first_non_empty(
                        _string_or_json(request_payload.get("body")),
                        _string_or_json(request_payload.get("text")),
                        _string_or_json(request_payload.get("message")),
                        _string_or_json(request_payload.get("content")),
                        _string_or_json(request_payload.get("markdown")),
                    )
                    provider_ref = ""
                    if output_text:
                        try:
                            parsed_output = json.loads(output_text)
                        except json.JSONDecodeError:
                            parsed_output = {}
                        if isinstance(parsed_output, dict):
                            provider_ref = first_non_empty(
                                _string_or_json(parsed_output.get("message_ts")),
                                _string_or_json(parsed_output.get("ts")),
                                _string_or_json(parsed_output.get("id")),
                            )
                            if not body:
                                body = first_non_empty(
                                    _string_or_json(parsed_output.get("body")),
                                    _string_or_json(parsed_output.get("text")),
                                    _string_or_json(parsed_output.get("message")),
                                )
                    reply_delivery = {
                        "channel_id": channel_id,
                        "thread_ts": thread_ts,
                        "body": body,
                        "body_sha1": hashlib.sha1(body.encode("utf-8")).hexdigest() if body else "",
                        "body_excerpt": body[:500],
                        "tool_call_id": call_id,
                        "tool_name": f"slack.mcp.{name}",
                        "provider_ref": provider_ref,
                        "send_status": "posted",
                    }
        return normalized, reply_delivery, reply_attempted

    def _execute_openai_native_task(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> HermesExecutionResult:
        if self._provider != "openai" or not self._api_key:
            return HermesExecutionResult(
                ok=False,
                message="OpenAI Responses API credentials are required for native MCP workflow execution.",
                provider=self._backend,
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "failure_class": "runner_non_ok",
                    "native_mcp_enabled": True,
                },
            )
        if preflight := self._preflight_tool_policy_failure(task, tool_policy):
            return preflight
        try:
            resolved_mcp_servers, send_tool_names = self._resolved_task_mcp_servers(task)
        except RuntimeError as exc:
            return HermesExecutionResult(
                ok=False,
                message=str(exc),
                provider=self._backend,
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "failure_class": "runner_non_ok",
                    "native_mcp_enabled": True,
                    "runner_diagnostics": {
                        "failure_kind": "slack_mcp_unavailable",
                        "provider_error_message": str(exc),
                    },
                },
            )
        binding = ReadOnlyToolBinding(
            base_url=self._config.tool_gateway_base_url or "",
            allowed_tool_names=tool_policy.custom_tools,
            task_repo=task.repo,
            task_repo_ref=task.repo_ref or "",
            task_prompt=task.prompt,
            default_question=str(self._question_default_query_hints(task).get("default_question", "")),
            repo_question=str(self._question_default_query_hints(task).get("repo_question", "")),
            knowledge_topic=str(self._question_default_query_hints(task).get("knowledge_topic", "")),
            knowledge_question=str(self._question_default_query_hints(task).get("knowledge_question", "")),
            slack_history_focus=str(self._question_default_query_hints(task).get("slack_history_focus", "")),
            slack_search_query=str(self._question_default_query_hints(task).get("slack_search_query", "")),
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
        tools = [*self._direct_function_tools(tool_policy), *resolved_mcp_servers]
        timeout_seconds = min(self._effective_task_timeout(task), max(1, self._transport_timeout_seconds - 5))
        user_prompt = self._direct_task_user_prompt(task)
        instructions = first_non_empty(task.system_message or "", "Return valid JSON only.")
        previous_response_id = ""
        current_input: JsonValue = user_prompt
        accumulated_mcp_calls: list[JsonObject] = []
        reply_delivery: JsonObject | None = None
        reply_delivery_attempted = False
        rounds = 0
        while True:
            remaining = timeout_seconds
            payload: JsonObject = {
                "model": self._provider_model,
                "instructions": instructions,
                "input": current_input,
                "parallel_tool_calls": False,
                "tools": tools,
                "max_output_tokens": 4000,
                "text": {
                    "format": {"type": "json_object"},
                    "verbosity": "low",
                },
            }
            if previous_response_id:
                payload["previous_response_id"] = previous_response_id
            if self._reasoning_config.get("enabled", True):
                payload["reasoning"] = {"effort": self._reasoning_effort}
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
                with urlrequest.urlopen(req, timeout=max(1, remaining)) as resp:
                    body = resp.read().decode("utf-8")
            except urlerror.HTTPError as exc:
                detail = exc.read().decode("utf-8", errors="replace")
                return HermesExecutionResult(
                    ok=False,
                    message=f"Native workflow executor returned {exc.code}: {detail[:2000]}",
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        "failure_class": "runner_non_ok",
                        "native_mcp_enabled": True,
                        "reply_delivery_attempted": reply_delivery_attempted,
                        "reply_delivery": reply_delivery or {},
                        "mcp_calls": accumulated_mcp_calls,
                        "tool_calls": self._merge_runtime_values(binding.diagnostics().get("tool_calls"), accumulated_mcp_calls),
                    },
                )
            except (TimeoutError, socket.timeout, urlerror.URLError, ConnectionError, OSError) as exc:
                failure_class = "runner_transport_timeout"
                if reply_delivery_attempted:
                    failure_class = "runner_reply_delivery_uncertain"
                return HermesExecutionResult(
                    ok=False,
                    message=f"Native workflow executor timed out after {max(1, remaining)}s: {exc}",
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        "failure_class": failure_class,
                        "native_mcp_enabled": True,
                        "reply_delivery_attempted": reply_delivery_attempted,
                        "reply_delivery": reply_delivery or {},
                        "mcp_calls": accumulated_mcp_calls,
                        "tool_calls": self._merge_runtime_values(binding.diagnostics().get("tool_calls"), accumulated_mcp_calls),
                    },
                )
            try:
                parsed = json.loads(body)
            except json.JSONDecodeError:
                return HermesExecutionResult(
                    ok=False,
                    message="Native workflow executor returned invalid JSON.",
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        "failure_class": "runner_non_ok",
                        "native_mcp_enabled": True,
                    },
                )
            if not isinstance(parsed, dict):
                return HermesExecutionResult(
                    ok=False,
                    message="Native workflow executor returned a non-object JSON payload.",
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        "failure_class": "runner_non_ok",
                        "native_mcp_enabled": True,
                    },
                )
            previous_response_id = _string_or_json(parsed.get("id"))
            normalized_mcp_calls, latest_reply_delivery, latest_reply_attempted = self._normalize_mcp_calls(
                parsed,
                send_tool_names=send_tool_names,
                task=task,
            )
            accumulated_mcp_calls = self._merge_runtime_values(accumulated_mcp_calls, normalized_mcp_calls)
            reply_delivery_attempted = reply_delivery_attempted or latest_reply_attempted
            if latest_reply_delivery is not None:
                reply_delivery = latest_reply_delivery
            function_calls = self._responses_function_calls(parsed)
            if function_calls:
                if rounds >= self._max_iterations:
                    return HermesExecutionResult(
                        ok=False,
                        message="Native workflow executor exhausted its function-call budget.",
                        provider=self._backend,
                        raw={
                            **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                            "failure_class": "runner_iteration_budget_exhausted",
                            "native_mcp_enabled": True,
                            "reply_delivery_attempted": reply_delivery_attempted,
                            "reply_delivery": reply_delivery or {},
                            "mcp_calls": accumulated_mcp_calls,
                            "tool_calls": self._merge_runtime_values(binding.diagnostics().get("tool_calls"), accumulated_mcp_calls),
                        },
                    )
                outputs: list[JsonObject] = []
                for call in function_calls:
                    name = _string_or_json(call.get("name"))
                    call_id = first_non_empty(_string_or_json(call.get("call_id")), _string_or_json(call.get("id")), name)
                    arguments_text = _string_or_json(call.get("arguments")) or _string_or_json(call.get("input"))
                    arguments = {}
                    if arguments_text:
                        try:
                            parsed_args = json.loads(arguments_text)
                        except json.JSONDecodeError:
                            parsed_args = {}
                        if isinstance(parsed_args, dict):
                            arguments = parsed_args
                    output_text = binding.handle_tool_call(name, arguments)
                    outputs.append(
                        {
                            "type": "function_call_output",
                            "call_id": call_id,
                            "output": output_text,
                        }
                    )
                current_input = outputs
                rounds += 1
                continue
            response_text = self._responses_output_text(parsed)
            if not response_text:
                return HermesExecutionResult(
                    ok=False,
                    message="Native workflow executor returned an empty response.",
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        "failure_class": "runner_non_ok",
                        "native_mcp_enabled": True,
                        "mcp_calls": accumulated_mcp_calls,
                        "tool_calls": self._merge_runtime_values(binding.diagnostics().get("tool_calls"), accumulated_mcp_calls),
                    },
                )
            try:
                structured_output = self._extract_structured_output(response_text)
            except HermesStructuredOutputError as exc:
                failure_class = "runner_structured_output_parse_failure"
                if reply_delivery_attempted:
                    failure_class = "runner_reply_delivery_uncertain"
                return HermesExecutionResult(
                    ok=False,
                    message=str(exc),
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        "failure_class": failure_class,
                        "native_mcp_enabled": True,
                        "reply_delivery_attempted": reply_delivery_attempted,
                        "reply_delivery": reply_delivery or {},
                        "mcp_calls": accumulated_mcp_calls,
                        "tool_calls": self._merge_runtime_values(binding.diagnostics().get("tool_calls"), accumulated_mcp_calls),
                        "structured_output_error": str(exc),
                    },
                )
            observed = binding.diagnostics()
            tool_calls = self._merge_runtime_values(_json_object_list(observed.get("tool_calls")), accumulated_mcp_calls)
            raw: JsonObject = {
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                "native_mcp_enabled": True,
                "mcp_calls": accumulated_mcp_calls,
                "tool_calls": tool_calls,
                "reply_delivery_attempted": reply_delivery_attempted,
                "structured_output": structured_output,
                "provider_response_id": previous_response_id,
                "task_timeout_seconds": self._effective_task_timeout(task),
                "transport_timeout_seconds": self._transport_timeout_seconds,
                "tool_policy_mode": tool_policy.mode,
                "tool_allowlist_effective": tool_policy.effective,
                "tool_transport_allowlist_effective": tool_policy.transport_effective,
                "blocked_tool_names": tool_policy.blocked,
                "runner_diagnostics": {
                    "native_execution_mode": "openai_responses_mcp",
                    "native_mcp_enabled": True,
                    "reply_delivery_attempted": reply_delivery_attempted,
                },
            }
            if reply_delivery is not None:
                raw["reply_delivery"] = reply_delivery
            return HermesExecutionResult(
                ok=True,
                message=json.dumps(structured_output, ensure_ascii=True, sort_keys=True),
                provider=self._backend,
                raw=raw,
            )

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
                "Return only one JSON object with keys: reply_markdown, confidence, completion_verdict, termination_reason, alignment_degraded, alignment_notice.",
                "reply_markdown must be grounded, concise, and ready for Slack posting.",
                "Use completion_verdict=partial when the supplied diagnostics or ledger indicate a bounded stop such as task_timeout or iteration_budget_exhausted.",
                "When alignment evidence is degraded, set alignment_degraded=true and explain the limitation in alignment_notice.",
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
        verdict = "partial" if termination_reason in {"task_timeout", "iteration_budget_exhausted"} else "complete"
        return verdict, termination_reason

    def _execute_question_reduce_task(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> HermesExecutionResult:
        timeout_seconds = min(self._effective_task_timeout(task), max(1, self._transport_timeout_seconds - 5))
        attempt = self._invoke_direct_json_response(
            system_prompt=self._question_reduce_system_prompt(),
            user_prompt=self._question_reduce_user_prompt(task),
            timeout_seconds=timeout_seconds,
            reasoning_effort="medium",
        )
        if not attempt.ok:
            message = first_non_empty(attempt.error, "Question reducer failed.")
            return HermesExecutionResult(
                ok=False,
                message=message,
                provider=self._backend,
                raw={
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
            )
        completion_verdict, termination_reason = self._question_reduce_defaults(task)
        structured_output = dict(attempt.structured_output)
        structured_output["completion_verdict"] = first_non_empty(_string_or_json(structured_output.get("completion_verdict")), completion_verdict)
        structured_output["termination_reason"] = first_non_empty(_string_or_json(structured_output.get("termination_reason")), termination_reason)
        return HermesExecutionResult(
            ok=True,
            message=json.dumps(structured_output, ensure_ascii=True, sort_keys=True),
            provider=self._backend,
            raw={
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
        )

    def _partial_completion_idempotency_key(self, task: RunnerTaskRequest, termination_reason: str) -> str:
        scope = first_non_empty(task.trace_id, task.workflow_id, task.session_scope_id, "workflow")
        return f"partial-{termination_reason}-{scope}"

    def _synthesize_partial_slack_post_action(
        self,
        task: RunnerTaskRequest,
        structured_output: JsonObject,
        termination_reason: str,
    ) -> tuple[JsonObject, bool]:
        if not self._workflow_missing_explicit_reply_action(task, structured_output):
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
    ) -> HermesExecutionResult:
        recovery_timeout_seconds = self._partial_completion_timeout_seconds(task, termination_reason, stop_meta)
        evidence_ledger = self._build_evidence_ledger(task, observed, termination_reason)
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
            return HermesExecutionResult(
                ok=True,
                message=response_text,
                provider=self._backend,
                raw=merged_raw,
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

    def _attach_tool_policy(self, agent: Any, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> None:
        current_tools = list(getattr(agent, "tools", []) or [])
        current_valid = set(getattr(agent, "valid_tool_names", set()) or set())
        setattr(agent, "_rsi_readonly_tool_binding", None)
        native_governed_tools = self._native_governed_tools_enabled(task)
        allowed_names = set(tool_policy.effective)
        filtered_tools = current_tools
        query_hints = self._question_default_query_hints(task)
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
        reasoning_timeout_seconds: int,
    ) -> tuple[str, JsonObject | None, JsonObject]:
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
                    termination_reason = "iteration_budget_exhausted" if self._budget_exhausted(activity) else "normal_completion"
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
                        )
                    if inactivity_timeout_seconds > 0 and idle_seconds >= float(inactivity_timeout_seconds):
                        return self._interrupt_execution(
                            agent,
                            future,
                            "inactivity_timeout",
                            activity,
                            duration_seconds=inactivity_timeout_seconds,
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
    ) -> tuple[str, JsonObject | None, JsonObject]:
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
        return termination_reason, None, meta

    def execute_task(self, task: RunnerTaskRequest) -> HermesExecutionResult:
        if task.task_type not in ROLE_TASK_TYPES.get(self._role, {self._role}):
            return HermesExecutionResult(
                ok=False,
                message=f"Runner role {self._role} cannot execute task type {task.task_type}.",
                provider="policy",
                raw={"role": self._role, "task_type": task.task_type},
            )
        tool_policy = self._resolve_tool_policy(task)
        if self._task_uses_native_mcp(task):
            result = self._execute_openai_native_task(task, tool_policy)
            if result.ok:
                result.raw = {
                    **result.raw,
                    "mcp_servers": task.mcp_servers,
                }
            return result
        if task.task_type == "question_reduce":
            return self._execute_question_reduce_task(task, tool_policy)
        prompt = self._render_task_prompt(task, tool_policy)
        rendered_task = replace(task, prompt=prompt)
        result = self._execute_task_request(rendered_task, tool_policy)
        if not result.ok:
            return result
        completion_verdict = string_from_map(_json_object_or_empty(result.raw), "completion_verdict")
        initial_response = result.message
        repair_attempted = False
        repair_succeeded = False
        action_contract_repair_attempted = False
        action_contract_repair_succeeded = False
        action_contract_repair_error = ""
        action_contract_repair_response = ""
        if completion_verdict == "partial":
            structured_output_payload = _json_object_or_empty(result.raw.get("structured_output"))
            if not structured_output_payload:
                message = "Bounded-stop partial completion did not return structured output."
                return HermesExecutionResult(
                    ok=False,
                    message=message,
                    provider=result.provider,
                    raw={
                        **result.raw,
                        "failure_class": "runner_structured_output_parse_failure",
                        "runner_diagnostics": self._runner_diagnostics(
                            tool_policy,
                            failure_kind="structured_output_parse_failure",
                            provider_error_message=message,
                            repair_attempted=False,
                            repair_succeeded=False,
                            observed=_json_object_or_empty(result.raw.get("runner_diagnostics")),
                        ),
                        "structured_output_error": message,
                        "repair_attempted": False,
                        "repair_succeeded": False,
                    },
                )
            structured_output = _normalize_structured_output(structured_output_payload)
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
                    repair_result = self._execute_task_request(repair_task, repair_tool_policy)
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
            if self._workflow_missing_explicit_reply_action(task, structured_output):
                action_contract_repair_attempted = True
                logger.info(
                    "workflow runner action-contract repair attempted trace_id=%s workflow_id=%s",
                    task.trace_id or "",
                    task.workflow_id or "",
                )
                repair_task = self._build_action_contract_repair_task(rendered_task, structured_output)
                repair_result = self._execute_task_request(repair_task, repair_tool_policy)
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
            "mcp_servers": task.mcp_servers,
            "repair_attempted": repair_attempted or existing_repair_attempted,
            "repair_succeeded": repair_succeeded or existing_repair_succeeded,
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
        if task.task_type == "question_expand":
            question_payload = self._question_task_payload(task)
            parts.append("Expand the evidence ledger with bounded, grounded read-only retrieval.")
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
