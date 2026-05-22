from __future__ import annotations

import concurrent.futures
from dataclasses import dataclass, replace
import base64
import hashlib
import hmac
import json
import logging
import os
from pathlib import Path
import re
import shlex
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

from .config import ModelProfile, RunnerConfig
from .file_utils import _atomic_write_json
from .hermes_adapter import HermesAdapter
from .hermes_agent_adapter import validate_hermes_contract
from .hermes_mcp_adapter import HermesTaskScopedMCPAdapter, TaskScopedMCPRegistration
from .execution_contract import (
    EXECUTION_CONTRACT_VERSION,
    RUNNER_PLANNER_MODE,
    HermesCompanyComputer,
)
from .observability import ObservationEmitter, execution_observation_id
from .rsi_tools import (
    HERMES_ARTIFACT_TOOLSET,
    HERMES_DB_READ_TOOLSET,
    HERMES_RSI_KNOWLEDGE_TOOLSET,
    HERMES_RSI_NOTION_TOOLSET,
    HERMES_RSI_OBSERVABILITY_TOOLSET,
    HERMES_RSI_SENTRY_TOOLSET,
    HERMES_RSI_SLACK_TOOLSET,
    canonical_tool_name,
    normalize_tool_names,
)
from .session_manager import MemoryTracker, SessionContext, SessionManager, stable_session_id

ROLE_TASK_TYPES = {
    "prod": {"general", "workflow", "prod"},
    "proactive": {"general", "workflow", "proactive"},
    "eval": {"general", "eval"},
    "proposal": {"general", "proposal", "repo-change"},
}

logger = logging.getLogger(__name__)

DEFAULT_HERMES_NATIVE_TOOLSETS = ("terminal", "file", "company_knowledge")
RSI_NATIVE_TOOLSETS = (
    HERMES_RSI_SLACK_TOOLSET,
    HERMES_RSI_NOTION_TOOLSET,
    HERMES_RSI_KNOWLEDGE_TOOLSET,
    HERMES_RSI_SENTRY_TOOLSET,
)
PARTIAL_COMPLETION_TERMINATION_REASONS = frozenset(
    {
        "task_timeout",
        "inactivity_timeout",
        "iteration_budget_exhausted",
        "output_token_budget_exhausted",
    }
)
DIRECT_DELIVERY_SUCCESS_STATUSES = frozenset({"posted", "sent", "uploaded", "completed", "ok", "success", "shared"})
SELF_REVIEW_SECRET_ENV_ALLOWLIST = frozenset(
    {
        "OPENROUTER_API_KEY",
        "RSI_OPENROUTER_API_KEY",
        "DEEPSEEK_API_KEY",
        "RSI_DEEPSEEK_API_KEY",
        "OPENAI_API_KEY",
        "LLM_OPENAI_API_KEY",
        "ANTHROPIC_API_KEY",
        "ANTHROPIC_TOKEN",
        "CLAUDE_CODE_OAUTH_TOKEN",
        "GOOGLE_API_KEY",
        "GEMINI_API_KEY",
        "LM_API_KEY",
        "COPILOT_GITHUB_TOKEN",
        "GH_TOKEN",
        "GITHUB_TOKEN",
        "GLM_API_KEY",
        "ZAI_API_KEY",
        "Z_AI_API_KEY",
        "KIMI_API_KEY",
        "KIMI_CODING_API_KEY",
        "STEPFUN_API_KEY",
        "ARCEEAI_API_KEY",
        "GMI_API_KEY",
        "MINIMAX_API_KEY",
        "DASHSCOPE_API_KEY",
        "HERMES_STATE_POSTGRES_URL",
        "RSI_POSTGRES_URL",
        "DATABASE_URL",
    }
)
NATIVE_WORKER_SOURCE_CREDENTIAL_DENYLIST = frozenset(
    {
        "SLACK_BOT_TOKEN",
        "RSI_SLACK_BOT_TOKEN",
        "NOTION_TOKEN",
        "NOTION_API_KEY",
        "SENTRY_AUTH_TOKEN",
        "RSI_SENTRY_AUTH_TOKEN",
        "RSI_NATIVE_TOOLS_CLIENT_TOKEN",
        "RSI_DB_READ_CLIENT_TOKEN",
        "RSI_DB_READ_RELAY_TOKEN",
    }
)
SELF_REVIEW_STATE_ENV_ALLOWLIST = frozenset(
    {
        "HERMES_STATE_BACKEND",
        "HERMES_STATE_POSTGRES_SCHEMA",
        "HERMES_STATE_POSTGRES_POOL_SIZE",
        "HERMES_STATE_SEARCH_MODE",
        "HERMES_STATE_EMBEDDINGS_ENABLED",
        "HERMES_STATE_EMBEDDING_MODEL",
        "HERMES_STATE_EMBEDDING_DIMENSIONS",
        "HERMES_STATE_EMBEDDING_BATCH_SIZE",
        "HERMES_STATE_EMBEDDING_RECONCILE_ON_START",
        "LLM_EMBEDDING_MODEL",
        "LLM_OPENAI_BASE_URL",
    }
)
GROUNDED_EVIDENCE_TOOL_NAMES = frozenset(
    {
        "repo.read_file",
        "repo.context",
        "repo.search",
        "github.repo_context",
        "github.repo_activity",
        "kubernetes.inspect",
        "kubernetes.events",
        "knowledge.context",
        "rsi.runtime_deployment_facts",
        "slack.history",
    }
)
_LIFECYCLE_TOOL_NAME_ALIASES = {
    "github_repo_activity": "github.repo_activity",
    "github_repo_context": "github.repo_context",
    "knowledge_context": "knowledge.context",
    "kubernetes_events": "kubernetes.events",
    "kubernetes_inspect": "kubernetes.inspect",
    "kubernetes_logs": "kubernetes.logs",
    "repo_context": "repo.context",
    "repo_read_file": "repo.read_file",
    "repo_search": "repo.search",
    "rsi_runtime_deployment_facts": "rsi.runtime_deployment_facts",
    "slack_history": "slack.history",
}
ARTIFACT_RENDER_NATIVE_TOOL_NAMES = frozenset({"write_file", "read_file", "search_files", "skill_view"})
_NATIVE_EXECUTOR_RESULT_MARKER = "RSI_EXECUTOR_RESULT::"
_NATIVE_EXECUTOR_OUTPUT_CHUNK_CHARS = 8 * 1024
_NATIVE_ENVELOPE_WAIT_SECONDS = 2.0
_NATIVE_ENVELOPE_REQUIRED_FIELDS = (
    "contract_version",
    "producer",
    "producer_version",
    "created_at",
    "facts_source",
    "execution_id",
    "operation_id",
    "trace_id",
    "workflow_id",
    "phase_runs",
    "ledger_events",
    "artifacts",
    "deliveries",
    "completion",
    "final_response",
)
_NATIVE_ENVELOPE_FAILURE_CLASSES = frozenset(
    {
        "plugin_execution_envelope_missing",
        "plugin_execution_envelope_mismatch",
        "plugin_execution_envelope_invalid",
        "native_envelope_plugin_unavailable",
        "native_workflow_preflight_failed",
    }
)
_SENSITIVE_ENV_KEY_FRAGMENTS = (
    "authorization",
    "api-key",
    "api_key",
    "apikey",
    "client-id",
    "client_id",
    "token",
    "secret",
    "private-key",
    "private_key",
    "password",
)
_SENSITIVE_ENV_KEY_PATTERN = "|".join(re.escape(fragment) for fragment in _SENSITIVE_ENV_KEY_FRAGMENTS)
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
        lambda _match: "[redacted-api-key]",
    ),
    (
        re.compile(
            rf"(?im)^(\s*(?:export\s+)?[A-Za-z0-9_.-]*(?:{_SENSITIVE_ENV_KEY_PATTERN})[A-Za-z0-9_.-]*\s*[:=]\s*)(?!Bearer\s)(['\"]?)([^\s'\",}}{{\]]{{4,}})(\2)",
        ),
        lambda match: f"{match.group(1)}{match.group(2)}[redacted]{match.group(4)}",
    ),
    (
        re.compile(
            rf"(?is)(\"name\"\s*:\s*\"[^\"]*(?:{_SENSITIVE_ENV_KEY_PATTERN})[^\"]*\"\s*,\s*\"value\"\s*:\s*\")([^\"]+)(\")",
        ),
        lambda match: f"{match.group(1)}[redacted]{match.group(3)}",
    ),
    (
        re.compile(
            rf"(?is)(\"value\"\s*:\s*\")([^\"]+)(\"\s*,\s*\"name\"\s*:\s*\"[^\"]*(?:{_SENSITIVE_ENV_KEY_PATTERN})[^\"]*\")",
        ),
        lambda match: f"{match.group(1)}[redacted]{match.group(3)}",
    ),
    (
        re.compile(
            rf"(?is)(\"[A-Za-z0-9_.-]*(?:{_SENSITIVE_ENV_KEY_PATTERN})[A-Za-z0-9_.-]*\"\s*:\s*\")([^\"]+)(\")",
        ),
        lambda match: f"{match.group(1)}[redacted]{match.group(3)}",
    ),
)
_BENIGN_MCP_TOOLSET_WARNING = re.compile(
    r"Warning: Unknown toolsets:\s*\n(?:mcp-[^\n]*(?:\n|$))+\n*",
    re.MULTILINE,
)
_MARKDOWN_TABLE_SEPARATOR_REGEX = re.compile(r"^\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?$")
_PROVIDER_RUNTIME_ERROR_MARKERS = (
    "api call failed",
    "permissiondeniederror",
    "authenticationerror",
    "rate limit",
    "rate limited after",
    "key limit exceeded",
    "monthly limit",
    "insufficient_quota",
    "quota exceeded",
    "provider error",
)


def _json_object_or_empty(value: JsonValue | None) -> JsonObject:
    if isinstance(value, dict):
        return value
    return {}


def _env_map(name: str) -> dict[str, str]:
    raw = str(os.getenv(name, "") or "").strip()
    if not raw:
        return {}
    out: dict[str, str] = {}
    for part in raw.split(","):
        item = part.strip()
        if not item:
            continue
        key, sep, value = item.partition("=")
        if not sep:
            continue
        key = key.strip()
        value = value.strip()
        if key and value:
            out[key] = value
    return out


def _normalize_private_key(raw: str) -> str:
    value = raw.strip()
    if "\\n" in value and "\n" not in value:
        value = value.replace("\\n", "\n")
    if "BEGIN" in value:
        return value
    try:
        decoded = base64.b64decode(value).decode("utf-8")
    except Exception:
        return value
    return decoded if "BEGIN" in decoded else value


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


def _truncate_raw_string(value: str, limit: int) -> str:
    if len(value) <= limit:
        return value
    return value[: max(0, limit - 1)] + "…"


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


def _redact_json_value(value: Any, *, secret_values: list[str], limit: int) -> Any:
    if isinstance(value, str):
        return _truncate_raw_string(_redact_subprocess_output(value, secret_values=secret_values), limit)
    if isinstance(value, dict):
        name_value = str(value.get("name") or value.get("key") or "").strip()
        redact_value_field = bool(name_value and _is_sensitive_env_key(name_value))
        redacted: dict[str, Any] = {}
        for key, item in value.items():
            key_text = str(key)
            if _is_sensitive_env_key(key_text) or (redact_value_field and key_text == "value"):
                redacted[key_text] = "[redacted]"
                continue
            redacted[key_text] = _redact_json_value(item, secret_values=secret_values, limit=limit)
        return redacted
    if isinstance(value, list):
        return [_redact_json_value(item, secret_values=secret_values, limit=limit) for item in value]
    return value


def _suppress_benign_subprocess_output(text: str) -> str:
    return _BENIGN_MCP_TOOLSET_WARNING.sub("", str(text or ""))


class _NativeLifecycleTailer:
    def __init__(
        self,
        *,
        path: Path,
        phase: str,
        observer: ObservationEmitter,
        secret_values: list[str],
        emit_event: Any,
        activity_callback: Any | None = None,
        start_at_end: bool = False,
    ) -> None:
        self._path = path
        self._phase = phase
        self._observer = observer
        self._secret_values = secret_values
        self._emit_event = emit_event
        self._activity_callback = activity_callback
        self._stop = threading.Event()
        self._thread: threading.Thread | None = None
        self._offset = 0
        self._buffer = ""
        self.emitted = 0
        if start_at_end:
            try:
                self._offset = self._path.stat().st_size
            except OSError:
                self._offset = 0

    def start(self) -> None:
        self._thread = threading.Thread(target=self._run, name="rsi-hermes-lifecycle-tailer", daemon=True)
        self._thread.start()

    def stop(self) -> None:
        self._stop.set()
        if self._thread is not None:
            self._thread.join(timeout=5)
            if not self._thread.is_alive():
                self.drain()

    def drain(self) -> None:
        self._read_available()
        if self._buffer.strip():
            self._emit_line(self._buffer)
        self._buffer = ""

    def _run(self) -> None:
        while not self._stop.is_set():
            self._read_available()
            self._stop.wait(0.25)
        self._read_available()

    def _read_available(self) -> None:
        try:
            stat = self._path.stat()
        except OSError:
            return
        if stat.st_size < self._offset:
            self._offset = 0
            self._buffer = ""
        if stat.st_size == self._offset:
            return
        try:
            with self._path.open("rb") as handle:
                handle.seek(self._offset)
                data = handle.read()
                self._offset = handle.tell()
        except OSError:
            return
        if not data:
            return
        self._buffer += data.decode("utf-8", errors="replace")
        parts = self._buffer.split("\n")
        self._buffer = parts.pop() if not self._buffer.endswith("\n") else ""
        for line in parts:
            self._emit_line(line)

    def _emit_line(self, line: str) -> None:
        text = line.strip()
        if not text:
            return
        try:
            item = json.loads(text)
        except json.JSONDecodeError:
            return
        if not isinstance(item, dict):
            return
        if self._activity_callback is not None:
            event_type = str(item.get("event_type") or item.get("event") or "lifecycle").strip()
            status = str(item.get("status") or "").strip()
            self._activity_callback(f"lifecycle:{event_type}:{status}" if status else f"lifecycle:{event_type}")
        self._emit_event(
            self._observer,
            self._phase,
            item,
            secret_values=self._secret_values,
        )
        self.emitted += 1


def _float_or_zero(value: JsonValue | None) -> float:
    if isinstance(value, Number):
        return float(value)
    try:
        return float(str(value).strip())
    except (AttributeError, TypeError, ValueError):
        return 0.0


def _int_or_zero(value: JsonValue | None) -> int:
    if isinstance(value, bool):
        return 0
    if isinstance(value, Number):
        return int(value)
    try:
        return int(str(value).strip())
    except (AttributeError, TypeError, ValueError):
        return 0


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


def _normalize_skill_identifiers(value: JsonValue | None) -> list[str]:
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
    normalized["session_title"] = _string_or_json(payload.get("session_title"))
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


def _unwrap_json_code_fence(text: str) -> str:
    value = str(text or "").strip()
    if not value.startswith("```"):
        return value
    lines = value.splitlines()
    if len(lines) < 3 or lines[-1].strip() != "```":
        return value
    opener = lines[0].strip().lower()
    if opener not in {"```", "```json"}:
        return value
    return "\n".join(lines[1:-1]).strip()


STRUCTURED_OUTPUT_CANDIDATE_KEYS = {
    "session_title",
    "reply_draft",
    "final_answer",
    "proposed_actions",
    "reply_delivery",
    "artifact_render_briefs",
    "produced_artifacts",
    "change_plan",
    "retry_assessment",
}


def _looks_like_structured_output(payload: JsonObject) -> bool:
    return any(key in payload for key in STRUCTURED_OUTPUT_CANDIDATE_KEYS)


def _json_code_fence_payloads(text: str) -> list[str]:
    payloads: list[str] = []
    for match in re.finditer(r"```(?:json)?[ \t]*\r?\n(.*?)\r?\n```", text or "", flags=re.IGNORECASE | re.DOTALL):
        payload = match.group(1).strip()
        if payload:
            payloads.append(payload)
    return payloads


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


def _absolute_path_without_resolving(value: str) -> Path:
    path = Path(str(value or "")).expanduser()
    if path.is_absolute():
        return path
    return Path.cwd() / path


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


def _percent_encode_gateway_part(value: str) -> str:
    return urlparse.quote((value or "").strip(), safe="")


def _derive_root_message_ts(task: "RunnerTaskRequest") -> str:
    channel_id = (task.channel_id or "").strip()
    for entry in reversed(task.recent_conversation_entries or []):
        entry_channel = _string_or_json(entry.get("channel_id"))
        if channel_id and entry_channel and entry_channel != channel_id:
            continue
        message_ts = first_non_empty(
            _string_or_json(entry.get("message_ts")),
            _string_or_json(entry.get("ts")),
            _string_or_json(entry.get("event_ts")),
        )
        thread_ts = _string_or_json(entry.get("thread_ts"))
        if message_ts and (not thread_ts or thread_ts == message_ts):
            return message_ts
    return ""


def canonical_gateway_session_key(task: "RunnerTaskRequest", role: str) -> str:
    role_part = (role or "").strip().lower() or "unknown"
    channel_id = (task.channel_id or "").strip()
    if channel_id:
        thread_key = first_non_empty(task.thread_ts, task.message_ts, _derive_root_message_ts(task), "channel")
        return f"rsi:{role_part}:slack:{channel_id}:{thread_key.strip()}"
    scope_kind = (task.session_scope_kind or "role").strip().lower() or "role"
    scope_id = _percent_encode_gateway_part((task.session_scope_id or role_part).strip() or role_part)
    return f"rsi:{role_part}:scope:{scope_kind}:{scope_id}"


def _normalize_skill_identifier(value: JsonValue | None) -> str:
    normalized = _normalize_skill_identifiers([_string_or_json(value)])
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


def _reasoning_config_for_profile(profile: ModelProfile) -> JsonObject:
    if str(profile.thinking_mode or "").strip().lower() == "disabled":
        return {"enabled": False}
    parsed = parse_reasoning_effort(profile.reasoning_effort)
    if parsed:
        return parsed
    return {"enabled": True, "effort": "medium"}


def _runtime_api_key_for_provider(provider: str) -> str:
    provider = str(provider or "").strip().lower()
    if provider == "deepseek":
        return first_non_empty(os.getenv("RSI_DEEPSEEK_API_KEY"), os.getenv("DEEPSEEK_API_KEY"))
    if provider == "openrouter":
        return first_non_empty(os.getenv("RSI_OPENROUTER_API_KEY"), os.getenv("OPENROUTER_API_KEY"))
    return ""


def _base_url_host(base_url: str) -> str:
    try:
        return urlparse.urlparse(str(base_url or "").strip()).netloc
    except Exception:
        return ""


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
    operation_id: str | None
    execution_id: str | None
    conversation_id: str | None
    case_id: str | None
    channel_id: str | None
    thread_ts: str | None
    message_ts: str | None
    trigger_event_id: str | None
    recent_conversation_entries: list[JsonObject]
    case_summary: JsonObject | None
    prior_trace_refs: list[JsonObject]
    repo_allowlist: list[str]
    response_mode: str | None
    reply_delivery_mode: str | None
    context_refs: list[JsonObject]
    approval_mode: str | None
    reasoning_verbosity: str | None
    model_override: JsonObject
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
    external_tool_resume: JsonObject
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
            allowed_commands=_string_list(task.get("allowed_commands")),
            timeout_seconds=int(task.get("timeout_seconds", 0) or 0),
            expected_outputs=_string_list(task.get("expected_outputs")),
            artifact_destination=_optional_string(task.get("artifact_destination")),
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
            message_ts=_optional_string(task.get("message_ts")),
            trigger_event_id=_optional_string(task.get("trigger_event_id")),
            recent_conversation_entries=_json_object_list(task.get("recent_conversation_entries")),
            case_summary=_json_object_or_empty(task.get("case_summary")) or None,
            prior_trace_refs=_json_object_list(task.get("prior_trace_refs")),
            repo_allowlist=_string_list(task.get("repo_allowlist")),
            response_mode=_optional_string(task.get("response_mode")),
            reply_delivery_mode=_optional_string(task.get("reply_delivery_mode")),
            context_refs=_json_object_list(task.get("context_refs")),
            approval_mode=_optional_string(task.get("approval_mode")),
            reasoning_verbosity=_optional_string(task.get("reasoning_verbosity")),
            model_override=_json_object_or_empty(task.get("model_override")),
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
            external_tool_resume=_json_object_or_empty(task.get("external_tool_resume")),
            delivery_policy=_json_object_or_empty(task.get("delivery_policy")),
            workspace_policy=_json_object_or_empty(task.get("workspace_policy")),
            approval_policy=_json_object_or_empty(task.get("approval_policy")),
        )


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
        self._model_profiles = {
            "main": ModelProfile(
                name="main",
                provider=config.model_provider,
                configured_model=config.model,
                provider_model=config.provider_model,
                base_url=config.model_base_url,
                api_key_configured=config.model_api_key_configured,
                reasoning_effort=config.reasoning_effort,
                thinking_mode=config.thinking_mode,
                openrouter_provider_routing=dict(config.openrouter_provider_routing or {})
                if config.model_provider == "openrouter"
                else {},
            ),
            "summary": ModelProfile(
                name="summary",
                provider=config.summary_model_provider,
                configured_model=config.summary_model,
                provider_model=config.summary_provider_model,
                base_url=config.summary_model_base_url,
                api_key_configured=config.summary_model_api_key_configured,
                reasoning_effort=config.summary_reasoning_effort,
                thinking_mode=config.summary_thinking_mode,
                openrouter_provider_routing=dict(config.openrouter_provider_routing or {})
                if config.summary_model_provider == "openrouter"
                else {},
            ),
        }
        self._active_model_profile = self._model_profiles["main"]
        self._provider_routing = dict(self._active_model_profile.openrouter_provider_routing or {})
        self._reasoning_config = _reasoning_config_for_profile(self._active_model_profile)
        self._openrouter_configured = False
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
        self._mcp_adapter = HermesTaskScopedMCPAdapter()
        self._started_at_unix = time.time()
        self._executor_recent_results: dict[str, JsonObject] = {}
        self._executor_processes: dict[str, subprocess.Popen[str]] = {}
        self._executor_threads: dict[str, threading.Thread] = {}
        self._self_review_processes: dict[str, subprocess.Popen[str]] = {}
        self._executor_native_session_ids: dict[str, str] = {}
        self._executor_native_session_keys: dict[str, str] = {}
        self._executor_native_started_at_unix: dict[str, float] = {}
        self._executor_native_tracked_pids: dict[str, set[int]] = {}
        self._executor_status_heartbeat_at: dict[str, float] = {}
        self._self_review_draining = False
        self._executor_process_lock = threading.RLock()
        self._executor_cancel_requests: set[str] = set()
        self._company_computer_bootstrap_status = self._configure_company_computer_substrate()
        self._configure_runtime()
        required_toolsets = self._hermes_native_toolsets()
        self._contract_status = validate_hermes_contract(
            expected_pin=config.hermes_pin,
            hermes_home=config.hermes_home,
            session_db=self._session_manager.session_db_ref,
            required_toolsets=required_toolsets,
            require_platform_runtime=config.hermes_executor_enabled,
            require_session_db_ready=False,
        )
        self._available = (
            AIAgent is not None
            and self._runtime_has_credentials()
            and self._session_manager.available
            and bool(self._company_computer_bootstrap_status.get("ok"))
            and self._contract_status.ok
        )

    def _hermes_native_toolsets(self) -> list[str]:
        toolsets: list[str] = []
        if self._config.hermes_native_terminal_enabled:
            toolsets.extend(self._config.hermes_native_toolsets or list(DEFAULT_HERMES_NATIVE_TOOLSETS))
        if self._config.db_read_gateway_configured:
            toolsets.append(HERMES_DB_READ_TOOLSET)
        if self._grafana_observability_configured():
            toolsets.append(HERMES_RSI_OBSERVABILITY_TOOLSET)
        return normalize_tool_names(toolsets)

    def _rsi_native_toolsets_for_task(self, task: RunnerTaskRequest) -> list[str]:
        if self._execution_phase(task) in {"render", "deliver"}:
            return []
        if task.task_type in {"workflow", "prod", "proactive"}:
            return list(RSI_NATIVE_TOOLSETS)
        return []

    def _grafana_observability_configured(self) -> bool:
        return self._config.grafana_observability_configured

    def _configure_company_computer_substrate(self) -> JsonObject:
        status: JsonObject = {
            "ok": True,
            "status": "disabled",
            "errors": [],
            "terminal_enabled": self._config.hermes_native_terminal_enabled,
            "kubernetes_context_enabled": self._config.hermes_kubernetes_context_enabled,
            "prod_kubernetes_context_enabled": self._config.hermes_prod_kubernetes_context_enabled,
            "computer_root": self._config.hermes_computer_root,
            "company_wiki_root": self._config.company_wiki_root,
            "terminal_cwd": self._config.hermes_terminal_cwd,
            "bin_dir": self._config.hermes_company_bin_dir,
            "kubeconfig_path": self._config.hermes_kubeconfig_path
            if (
                self._config.hermes_kubernetes_context_enabled
                or self._config.hermes_prod_kubernetes_context_enabled
            )
            else "",
        }
        if (
            not self._config.hermes_native_terminal_enabled
            and not self._config.hermes_kubernetes_context_enabled
            and not self._config.hermes_prod_kubernetes_context_enabled
        ):
            return status
        try:
            root = Path(self._config.hermes_computer_root).expanduser().resolve()
            root.mkdir(parents=True, exist_ok=True)
            if self._config.hermes_native_terminal_enabled:
                terminal_cwd = Path(self._config.hermes_terminal_cwd).expanduser().resolve()
                bin_dir = Path(self._config.hermes_company_bin_dir).expanduser().resolve()
                terminal_cwd.mkdir(parents=True, exist_ok=True)
                bin_dir.mkdir(parents=True, exist_ok=True)
                removed_legacy_tools = self._remove_legacy_company_bin_tools(bin_dir)
                os.environ["TERMINAL_ENV"] = self._config.hermes_terminal_env
                os.environ["TERMINAL_CWD"] = str(terminal_cwd)
                os.environ["TERMINAL_TIMEOUT"] = str(self._config.hermes_terminal_timeout_seconds)
                os.environ["TERMINAL_LIFETIME_SECONDS"] = str(self._config.hermes_terminal_lifetime_seconds)
                os.environ["TERMINAL_LOCAL_PERSISTENT"] = "true" if self._config.hermes_terminal_local_persistent else "false"
                path_entries = [item for item in os.environ.get("PATH", "").split(os.pathsep) if item]
                if str(bin_dir) not in path_entries:
                    os.environ["PATH"] = os.pathsep.join([str(bin_dir), *path_entries])
                status["grafana_observability"] = self._configure_grafana_cli_environment()
                status["db_read_gateway"] = self._db_read_native_tool_status()
                status.update(
                    {
                        "status": "configured",
                        "terminal_cwd": str(terminal_cwd),
                        "bin_dir": str(bin_dir),
                        "removed_legacy_tools": removed_legacy_tools,
                        "terminal_env": self._config.hermes_terminal_env,
                        "terminal_timeout_seconds": self._config.hermes_terminal_timeout_seconds,
                        "terminal_lifetime_seconds": self._config.hermes_terminal_lifetime_seconds,
                    }
                )
            if self._config.hermes_kubernetes_context_enabled or self._config.hermes_prod_kubernetes_context_enabled:
                kubeconfig_path = self._write_company_kubeconfig()
                os.environ["KUBECONFIG"] = str(kubeconfig_path)
                status["kubeconfig_path"] = str(kubeconfig_path)
                auth_modes: list[str] = []
                if self._config.hermes_kubernetes_context_enabled:
                    auth_modes.append("service_account_token_file")
                if self._config.hermes_prod_kubernetes_context_enabled:
                    auth_modes.append("aws_eks_exec_assume_role")
                    status["prod_kubernetes_context_name"] = self._config.hermes_prod_kubernetes_context_name
                status["kubernetes_auth"] = ",".join(auth_modes)
                status["status"] = "configured"
            self._write_company_computer_manifest(status)
        except Exception as exc:
            errors = [str(exc)]
            status["ok"] = False
            status["status"] = "failed"
            status["errors"] = errors
        return status

    def _remove_legacy_company_bin_tools(self, bin_dir: Path) -> list[str]:
        removed: list[str] = []
        for name in ("rsi-db",):
            path = bin_dir / name
            if not path.exists() and not path.is_symlink():
                continue
            if not path.is_file() and not path.is_symlink():
                raise RuntimeError(f"legacy company bin path {path} exists but is not a file")
            path.unlink()
            removed.append(name)
        return removed

    def _configure_grafana_cli_environment(self) -> JsonObject:
        if not self._grafana_observability_configured():
            return {"configured": False, "tool": "rsi_observability", "toolset": HERMES_RSI_OBSERVABILITY_TOOLSET}
        base_url = os.getenv("RSI_GRAFANA_BASE_URL", "").strip().rstrip("/")
        token = os.getenv("RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN", "").strip()
        metrics_datasource_uid = self._config.grafana_metrics_datasource_uid
        logs_datasource_uid = self._config.grafana_logs_datasource_uid
        os.environ["GRAFANA_SERVER"] = base_url
        os.environ["GRAFANA_TOKEN"] = token
        os.environ["RSI_GRAFANA_METRICS_DATASOURCE_UID"] = metrics_datasource_uid
        os.environ["RSI_GRAFANA_LOGS_DATASOURCE_UID"] = logs_datasource_uid
        return {
            "configured": True,
            "tool": "rsi_observability",
            "toolset": HERMES_RSI_OBSERVABILITY_TOOLSET,
            "transport": "grafana_datasource_proxy",
            "policy_boundary": "grafana_rbac",
            "query_guardrails_enforced": False,
            "server_env": "GRAFANA_SERVER",
            "token_env": "GRAFANA_TOKEN",
            "metrics_datasource_uid": metrics_datasource_uid,
            "logs_datasource_uid": logs_datasource_uid,
        }

    def _db_read_native_tool_status(self) -> JsonObject:
        if not self._config.db_read_gateway_configured:
            return {
                "configured": False,
                "interface": "native_toolset",
                "toolset": HERMES_DB_READ_TOOLSET,
            }
        return {
            "configured": True,
            "interface": "native_toolset",
            "toolset": HERMES_DB_READ_TOOLSET,
            "auth": "execution_scoped_control_plane_token",
            "executes_sql_locally": False,
        }

    def _write_company_kubeconfig(self) -> Path:
        kubeconfig_path = Path(self._config.hermes_kubeconfig_path.split(os.pathsep)[0]).expanduser().resolve()
        kubeconfig_path.parent.mkdir(parents=True, exist_ok=True)
        clusters: list[str] = []
        users: list[str] = []
        contexts: list[str] = []
        current_context = ""
        if self._config.hermes_kubernetes_context_enabled:
            token_path = _absolute_path_without_resolving(self._config.hermes_kubernetes_service_account_token_path)
            ca_path = _absolute_path_without_resolving(self._config.hermes_kubernetes_service_account_ca_path)
            namespace_path = _absolute_path_without_resolving(
                self._config.hermes_kubernetes_service_account_namespace_path
            )
            host = first_non_empty(os.getenv("KUBERNETES_SERVICE_HOST"), "")
            port = first_non_empty(os.getenv("KUBERNETES_SERVICE_PORT"), "443")
            if not host:
                raise RuntimeError("KUBERNETES_SERVICE_HOST is required when RSI_HERMES_KUBERNETES_CONTEXT_ENABLED=true")
            if not token_path.is_file():
                raise RuntimeError(f"Kubernetes service account token file is unavailable: {token_path}")
            if not ca_path.is_file():
                raise RuntimeError(f"Kubernetes service account CA file is unavailable: {ca_path}")
            namespace = "default"
            if namespace_path.is_file():
                namespace = namespace_path.read_text(encoding="utf-8").strip() or namespace
            if ":" in host:
                host = f"[{host}]"
            clusters.extend(
                [
                    "- name: in-cluster",
                    "  cluster:",
                    f"    server: https://{host}:{port}",
                    f"    certificate-authority: {ca_path}",
                ]
            )
            users.extend(
                [
                    "- name: hermes-executor",
                    "  user:",
                    f"    tokenFile: {token_path}",
                ]
            )
            contexts.extend(
                [
                    "- name: hermes-company-computer",
                    "  context:",
                    "    cluster: in-cluster",
                    "    user: hermes-executor",
                    f"    namespace: {namespace}",
                ]
            )
            current_context = "hermes-company-computer"
        if self._config.hermes_prod_kubernetes_context_enabled:
            prod_context = self._prod_kubernetes_context_values()
            clusters.extend(
                [
                    f"- name: {prod_context['cluster_name']}",
                    "  cluster:",
                    f"    server: {prod_context['cluster_server']}",
                    f"    certificate-authority-data: {prod_context['cluster_ca_data']}",
                ]
            )
            users.extend(
                [
                    f"- name: {prod_context['context_name']}",
                    "  user:",
                    "    exec:",
                    "      apiVersion: client.authentication.k8s.io/v1beta1",
                    "      command: aws",
                    "      args:",
                    "      - eks",
                    "      - get-token",
                    "      - --cluster-name",
                    f"      - {prod_context['cluster_name']}",
                    "      - --region",
                    f"      - {prod_context['region']}",
                    "      - --role-arn",
                    f"      - {prod_context['role_arn']}",
                ]
            )
            contexts.extend(
                [
                    f"- name: {prod_context['context_name']}",
                    "  context:",
                    f"    cluster: {prod_context['cluster_name']}",
                    f"    user: {prod_context['context_name']}",
                    f"    namespace: {prod_context['namespace']}",
                ]
            )
            current_context = current_context or prod_context["context_name"]
        if not clusters or not users or not contexts or not current_context:
            raise RuntimeError("At least one Kubernetes context must be enabled before writing kubeconfig")
        content = "\n".join(
            [
                "apiVersion: v1",
                "kind: Config",
                "clusters:",
                *clusters,
                "users:",
                *users,
                "contexts:",
                *contexts,
                f"current-context: {current_context}",
                "",
            ]
        )
        temp_path = kubeconfig_path.with_suffix(kubeconfig_path.suffix + ".tmp")
        temp_path.write_text(content, encoding="utf-8")
        temp_path.chmod(0o600)
        temp_path.replace(kubeconfig_path)
        return kubeconfig_path

    def _prod_kubernetes_context_values(self) -> dict[str, str]:
        required = {
            "RSI_HERMES_PROD_KUBERNETES_CONTEXT_NAME": self._config.hermes_prod_kubernetes_context_name,
            "RSI_HERMES_PROD_KUBERNETES_CLUSTER_NAME": self._config.hermes_prod_kubernetes_cluster_name,
            "RSI_HERMES_PROD_KUBERNETES_CLUSTER_SERVER": self._config.hermes_prod_kubernetes_cluster_server,
            "RSI_HERMES_PROD_KUBERNETES_CLUSTER_CA_DATA": self._config.hermes_prod_kubernetes_cluster_ca_data,
            "RSI_HERMES_PROD_KUBERNETES_ROLE_ARN": self._config.hermes_prod_kubernetes_role_arn,
            "RSI_HERMES_PROD_KUBERNETES_REGION": self._config.hermes_prod_kubernetes_region,
            "RSI_HERMES_PROD_KUBERNETES_NAMESPACE": self._config.hermes_prod_kubernetes_namespace,
        }
        missing = [name for name, value in required.items() if not str(value or "").strip()]
        if missing:
            raise RuntimeError(
                "Production Kubernetes context is enabled but missing required config: " + ", ".join(missing)
            )
        return {
            "context_name": self._config.hermes_prod_kubernetes_context_name,
            "cluster_name": self._config.hermes_prod_kubernetes_cluster_name,
            "cluster_server": self._config.hermes_prod_kubernetes_cluster_server,
            "cluster_ca_data": self._config.hermes_prod_kubernetes_cluster_ca_data,
            "role_arn": self._config.hermes_prod_kubernetes_role_arn,
            "region": self._config.hermes_prod_kubernetes_region,
            "namespace": self._config.hermes_prod_kubernetes_namespace,
        }

    def _write_company_computer_manifest(self, status: JsonObject) -> None:
        manifest_path = Path(self._config.hermes_computer_root).expanduser().resolve() / ".rsi" / "computer.json"
        manifest_path.parent.mkdir(parents=True, exist_ok=True)
        manifest = {
            "company_computer_root": self._config.hermes_computer_root,
            "run_root": self._config.hermes_run_root,
            "artifact_root": self._config.hermes_artifact_root,
            "company_wiki_root": self._config.company_wiki_root,
            "native_terminal_enabled": self._config.hermes_native_terminal_enabled,
            "native_toolsets": self._hermes_native_toolsets(),
            "terminal": {
                "backend": self._config.hermes_terminal_env,
                "cwd": self._config.hermes_terminal_cwd,
                "timeout_seconds": self._config.hermes_terminal_timeout_seconds,
                "lifetime_seconds": self._config.hermes_terminal_lifetime_seconds,
                "bin_dir": self._config.hermes_company_bin_dir,
            },
            "kubernetes_context": {
                "enabled": self._config.hermes_kubernetes_context_enabled,
                "kubeconfig_path": self._config.hermes_kubeconfig_path
                if (
                    self._config.hermes_kubernetes_context_enabled
                    or self._config.hermes_prod_kubernetes_context_enabled
                )
                else "",
                "auth": "service_account_token_file" if self._config.hermes_kubernetes_context_enabled else "",
            },
            "prod_kubernetes_context": {
                "enabled": self._config.hermes_prod_kubernetes_context_enabled,
                "name": self._config.hermes_prod_kubernetes_context_name
                if self._config.hermes_prod_kubernetes_context_enabled
                else "",
                "cluster_name": self._config.hermes_prod_kubernetes_cluster_name
                if self._config.hermes_prod_kubernetes_context_enabled
                else "",
                "region": self._config.hermes_prod_kubernetes_region
                if self._config.hermes_prod_kubernetes_context_enabled
                else "",
                "namespace": self._config.hermes_prod_kubernetes_namespace
                if self._config.hermes_prod_kubernetes_context_enabled
                else "",
                "auth": "aws_eks_exec_assume_role"
                if self._config.hermes_prod_kubernetes_context_enabled
                else "",
            },
            "grafana_observability": dict(
                status.get("grafana_observability")
                or {"configured": False, "tool": "rsi_observability", "toolset": HERMES_RSI_OBSERVABILITY_TOOLSET}
            ),
            "bootstrap_status": status,
        }
        manifest_path.write_text(json.dumps(manifest, ensure_ascii=True, indent=2, sort_keys=True), encoding="utf-8")

    def _configure_runtime(self) -> None:
        profile = self._active_model_profile
        self._provider = profile.provider
        self._provider_hint = profile.provider
        self._provider_model = profile.provider_model
        self._api_mode = ""
        self._base_url = profile.base_url
        self._api_key = _runtime_api_key_for_provider(profile.provider)
        self._provider_routing = dict(profile.openrouter_provider_routing or {}) if profile.provider == "openrouter" else {}
        self._reasoning_config = _reasoning_config_for_profile(profile)
        self._openrouter_configured = bool(_runtime_api_key_for_provider("openrouter"))

    def _runtime_has_credentials(self) -> bool:
        return bool(self._api_key)

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
        observation_sink_configured = bool(self._config.runtime_observation_sink_url)
        return {
            "status": "ok" if self.available and self._session_manager.skills_healthy else "degraded",
            "role": self._role,
            "executor_instance_id": self._config.executor_instance_id,
            "active_execution_count": self.active_execution_count(),
            "backend": self._backend,
            "provider": self._provider,
            "model": self._configured_model,
            "provider_model": self._provider_model,
            "model_profile": self._active_model_profile.name,
            "base_url_host": _base_url_host(self._base_url),
            "thinking_mode": self._active_model_profile.thinking_mode,
            "summary_provider": self._model_profiles["summary"].provider,
            "summary_model": self._model_profiles["summary"].provider_model,
            "summary_thinking_mode": self._model_profiles["summary"].thinking_mode,
            "provider_routing": dict(self._provider_routing) if self._provider == "openrouter" else {},
            "reasoning_effort": self._reasoning_effort,
            "api_mode": self._api_mode,
            "available": self.available,
            "hermes_available": AIAgent is not None,
            "runtime_credentials_configured": self._runtime_has_credentials(),
            "openrouter_configured": self._openrouter_configured,
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
            "direct_delivery_phase_enabled": False,
            "execution_contract_version": EXECUTION_CONTRACT_VERSION,
            "execution_envelope_v1_enabled": self._config.execution_envelope_v1_enabled,
            "execution_ledger_first_projection_enabled": self._config.execution_ledger_first_projection_enabled,
            "company_computer_root": self._config.hermes_computer_root,
            "runner_planner_mode": self._config.runner_planner_mode or RUNNER_PLANNER_MODE,
            "hermes_version": adapter_meta.version,
            "hermes_pin": adapter_meta.pin,
            "hermes_contract_status": self._contract_status.to_dict(),
            "memory_backend": self._config.memory_backend,
            "max_iterations": self._max_iterations,
            "task_timeout_seconds": self._default_task_timeout_seconds,
            "inactivity_timeout_seconds": self._default_inactivity_timeout_seconds,
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "live_stream_status": "configured" if observation_sink_configured else "not_configured",
            "hermes_stream_callback_surfaces": [
                "reasoning_callback",
                "stream_delta_callback",
                "thinking_callback",
                "tool_gen_callback",
                "tool_progress_callback",
                "tool_start_callback",
                "tool_complete_callback",
                "status_callback",
            ],
            "native_max_output_tokens": self._native_max_output_tokens,
            "hermes_executor_enabled": self._config.hermes_executor_enabled,
            "hermes_executor_service_only": self._config.hermes_executor_service_only,
            "hermes_executor_workspace_root": self._config.hermes_executor_workspace_root,
            "hermes_computer_root": self._config.hermes_computer_root,
            "hermes_run_root": self._config.hermes_run_root,
            "hermes_artifact_root": self._config.hermes_artifact_root,
            "company_wiki_root": self._config.company_wiki_root,
            "hermes_native_terminal_enabled": self._config.hermes_native_terminal_enabled,
            "hermes_native_toolsets": self._hermes_native_toolsets(),
            "hermes_terminal_env": self._config.hermes_terminal_env,
            "hermes_terminal_cwd": self._config.hermes_terminal_cwd,
            "hermes_terminal_timeout_seconds": self._config.hermes_terminal_timeout_seconds,
            "hermes_terminal_lifetime_seconds": self._config.hermes_terminal_lifetime_seconds,
            "hermes_company_bin_dir": self._config.hermes_company_bin_dir,
            "hermes_kubernetes_context_enabled": self._config.hermes_kubernetes_context_enabled,
            "hermes_kubeconfig_path": self._config.hermes_kubeconfig_path if (self._config.hermes_kubernetes_context_enabled or self._config.hermes_prod_kubernetes_context_enabled) else "",
            "company_computer_bootstrap_status": dict(self._company_computer_bootstrap_status),
            "context_engine_mode": adapter_meta.context_engine_mode,
            "context_engine_status": adapter_meta.context_engine_status,
            "lifecycle_hook_status": adapter_meta.lifecycle_hook_status,
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

    def probe_metadata(self) -> JsonObject:
        return {
            "status": "ok" if self.available and self._session_manager.skills_healthy else "degraded",
            "role": self._role,
            "executor_instance_id": self._config.executor_instance_id,
            "backend": self._backend,
            "provider": self._provider,
            "model": self._configured_model,
            "provider_model": self._provider_model,
            "api_mode": self._api_mode,
            "available": self.available,
            "hermes_available": AIAgent is not None,
            "runtime_credentials_configured": self._runtime_has_credentials(),
            "openrouter_configured": self._openrouter_configured,
            "persistence_enabled": self._session_manager.available,
            "session_continuity_status": "ok" if self._session_manager.available else "degraded",
            "skills_dir": self._session_manager.skills_dir,
            "bundled_skills_available": self._session_manager.bundled_skills_available,
            "bundled_skills_sync_status": self._session_manager.bundled_skills_sync_status,
            "hermes_config_parity_status": self._hermes_config_parity_status(),
            "observation_sink_status": "configured" if self._config.runtime_observation_sink_url else "not_configured",
            "execution_contract_version": EXECUTION_CONTRACT_VERSION,
            "memory_backend": self._config.memory_backend,
            "max_iterations": self._max_iterations,
            "task_timeout_seconds": self._default_task_timeout_seconds,
            "inactivity_timeout_seconds": self._default_inactivity_timeout_seconds,
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "hermes_executor_enabled": self._config.hermes_executor_enabled,
            "hermes_executor_service_only": self._config.hermes_executor_service_only,
            "executor_started_at_unix": self._started_at_unix,
            "issues": self._session_manager.ready_issues,
        }

    def active_execution_count(self) -> int:
        return int(self.active_execution_snapshot().get("active_execution_count") or 0)

    def request_drain(self) -> None:
        self._self_review_draining = True

    def terminate_self_review_processes(self, timeout_seconds: float = 5.0) -> None:
        deadline = time.monotonic() + max(0.0, float(timeout_seconds or 0))
        with self._executor_process_lock:
            processes = list(self._self_review_processes.items())
        for _review_id, process in processes:
            if process.poll() is None:
                try:
                    process.terminate()
                except OSError:
                    pass
        for _review_id, process in processes:
            remaining = deadline - time.monotonic()
            if process.poll() is not None:
                continue
            if remaining > 0:
                try:
                    process.wait(timeout=remaining)
                    continue
                except subprocess.TimeoutExpired:
                    pass
            try:
                process.kill()
            except OSError:
                pass

    def _process_registry_checkpoint_path(self) -> Path:
        return Path(self._config.hermes_home).expanduser() / "processes.json"

    def _read_process_registry_checkpoint(self) -> list[JsonObject]:
        try:
            parsed = json.loads(self._process_registry_checkpoint_path().read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError):
            return []
        if not isinstance(parsed, list):
            return []
        return [item for item in parsed if isinstance(item, dict)]

    def _proc_stat(self, pid: int) -> JsonObject:
        try:
            raw = (Path("/proc") / str(pid) / "stat").read_text(encoding="utf-8").strip()
            _prefix, suffix = raw.rsplit(")", 1)
            parts = suffix.strip().split()
            if len(parts) < 4:
                return {}
            return {
                "state": parts[0],
                "ppid": int(parts[1]),
                "pgid": int(parts[2]),
                "session_id": int(parts[3]),
            }
        except (OSError, ValueError):
            return {}

    def _pid_is_active(self, pid: int) -> bool:
        if pid <= 0:
            return False
        try:
            os.kill(pid, 0)
        except ProcessLookupError:
            return False
        except PermissionError:
            pass
        state = _string_or_json(self._proc_stat(pid).get("state")).upper()
        return state != "Z"

    def _process_command(self, pid: int) -> str:
        proc_root = Path("/proc") / str(pid)
        try:
            raw = (proc_root / "cmdline").read_bytes()
            command = " ".join(part.decode("utf-8", errors="replace") for part in raw.split(b"\0") if part)
            if command:
                return command[:240]
        except OSError:
            pass
        try:
            return (proc_root / "comm").read_text(encoding="utf-8", errors="replace").strip()[:240]
        except OSError:
            return ""

    def _process_table(self) -> dict[int, JsonObject]:
        proc_root = Path("/proc")
        if not proc_root.exists():
            return {}
        table: dict[int, JsonObject] = {}
        try:
            proc_entries = list(proc_root.iterdir())
        except OSError:
            return {}
        for item in proc_entries:
            if item.name.isdigit():
                pid = int(item.name)
                stat = self._proc_stat(pid)
                if stat:
                    table[pid] = stat
        return table

    def _descendant_process_ids(self, pid: int) -> set[int]:
        if pid <= 0:
            return set()
        table = self._process_table()
        if not table:
            return set()
        children_by_parent: dict[int, list[int]] = {}
        for child_pid, stat in table.items():
            ppid = int(stat.get("ppid") or 0)
            children_by_parent.setdefault(ppid, []).append(child_pid)
        descendants: set[int] = set()
        pending = list(children_by_parent.get(pid, []))
        while pending:
            child_pid = pending.pop()
            if child_pid in descendants:
                continue
            descendants.add(child_pid)
            pending.extend(children_by_parent.get(child_pid, []))
        return descendants

    def _refresh_native_descendant_tracking_locked(self) -> None:
        for execution_id, process in list(self._executor_processes.items()):
            if process.poll() is not None:
                continue
            try:
                pid = int(getattr(process, "pid", 0) or 0)
            except (TypeError, ValueError):
                continue
            if pid <= 0:
                continue
            descendants = self._descendant_process_ids(pid)
            if descendants:
                self._executor_native_tracked_pids.setdefault(execution_id, set()).update(descendants)

    def _native_process_entry(self, *, pid: int, source: str, payload: JsonObject | None = None) -> JsonObject:
        payload = payload or {}
        return {
            "pid": pid,
            "source": source,
            "pid_scope": _string_or_json(payload.get("pid_scope")) or "host",
            "session_id": _string_or_json(payload.get("task_id")),
            "session_key": _string_or_json(payload.get("session_key")),
            "command": _string_or_json(payload.get("command")) or self._process_command(pid),
        }

    def _active_native_processes_by_execution_locked(self) -> dict[str, list[JsonObject]]:
        self._refresh_native_descendant_tracking_locked()
        by_execution: dict[str, dict[int, JsonObject]] = {}
        session_ids = dict(self._executor_native_session_ids)
        session_keys = dict(self._executor_native_session_keys)
        execution_started_at = dict(self._executor_native_started_at_unix)

        if session_ids:
            for entry in self._read_process_registry_checkpoint():
                pid_scope = _string_or_json(entry.get("pid_scope")) or "host"
                if pid_scope != "host":
                    continue
                try:
                    pid = int(entry.get("pid") or 0)
                except (TypeError, ValueError):
                    continue
                if not self._pid_is_active(pid):
                    continue
                task_id = _string_or_json(entry.get("task_id"))
                session_key = _string_or_json(entry.get("session_key"))
                for execution_id, tracked_session_id in session_ids.items():
                    tracked_session_key = session_keys.get(execution_id, "")
                    try:
                        process_started_at = float(entry.get("started_at") or 0)
                    except (TypeError, ValueError):
                        process_started_at = 0.0
                    native_started_at = float(execution_started_at.get(execution_id) or 0)
                    if (
                        process_started_at > 0
                        and native_started_at > 0
                        and process_started_at < native_started_at - 1.0
                    ):
                        continue
                    if (tracked_session_id and task_id == tracked_session_id) or (
                        tracked_session_key and session_key == tracked_session_key
                    ):
                        by_execution.setdefault(execution_id, {})[pid] = self._native_process_entry(
                            pid=pid,
                            source="process_registry",
                            payload=entry,
                        )

        for execution_id, pids in list(self._executor_native_tracked_pids.items()):
            active_pids = {pid for pid in pids if self._pid_is_active(pid)}
            if active_pids:
                self._executor_native_tracked_pids[execution_id] = active_pids
                for pid in active_pids:
                    by_execution.setdefault(execution_id, {}).setdefault(
                        pid,
                        self._native_process_entry(pid=pid, source="descendant"),
                    )
            else:
                self._executor_native_tracked_pids.pop(execution_id, None)

        inactive_execution_ids = [
            execution_id
            for execution_id in set(session_ids) | set(session_keys) | set(execution_started_at)
            if execution_id not in self._executor_threads
            and execution_id not in self._executor_processes
            and not by_execution.get(execution_id)
        ]
        for execution_id in inactive_execution_ids:
            self._executor_native_session_ids.pop(execution_id, None)
            self._executor_native_session_keys.pop(execution_id, None)
            self._executor_native_started_at_unix.pop(execution_id, None)
            self._executor_native_tracked_pids.pop(execution_id, None)

        return {execution_id: list(entries.values()) for execution_id, entries in by_execution.items() if entries}

    def active_execution_snapshot(self, *, include_self_review_queue: bool = True) -> JsonObject:
        with self._executor_process_lock:
            active_thread_ids = sorted(
                execution_id
                for execution_id, thread in self._executor_threads.items()
                if thread.is_alive()
            )
            active_process_ids = sorted(
                execution_id
                for execution_id, process in self._executor_processes.items()
                if process.poll() is None
            )
            finished_review_ids = [
                review_id for review_id, process in self._self_review_processes.items() if process.poll() is not None
            ]
            for review_id in finished_review_ids:
                self._self_review_processes.pop(review_id, None)
            active_self_review_ids = sorted(
                review_id
                for review_id, process in self._self_review_processes.items()
                if process.poll() is None
            )
            native_processes_by_execution = self._active_native_processes_by_execution_locked()
        active_native_subprocess_ids = sorted(native_processes_by_execution)
        active_native_subprocesses = {
            execution_id: sorted(items, key=lambda item: int(item.get("pid") or 0))
            for execution_id, items in native_processes_by_execution.items()
        }
        active_ids = sorted(set(active_thread_ids) | set(active_process_ids) | set(active_native_subprocess_ids))
        review_status = self._self_review_queue_status(reconcile_stale=False) if include_self_review_queue else {}
        local_pending_reviews = int(review_status.get("local_owned_pending_count") or 0) + int(
            review_status.get("local_owned_promotable_count") or 0
        )
        active_self_review_count = len(active_self_review_ids)
        active_native_subprocess_count = sum(len(items) for items in native_processes_by_execution.values())
        return {
            "active_execution_count": len(active_ids) + active_self_review_count + local_pending_reviews,
            "active_main_execution_count": len(set(active_thread_ids) | set(active_process_ids)),
            "active_native_subprocess_execution_count": len(active_native_subprocess_ids),
            "active_native_subprocess_count": active_native_subprocess_count,
            "active_self_review_count": active_self_review_count,
            "local_pending_self_review_count": local_pending_reviews,
            "global_review_blocking_count": int(review_status.get("global_review_blocking_count") or 0),
            "global_skill_review_blocking_count": int(review_status.get("global_skill_review_blocking_count") or 0),
            "promotable_self_review_candidate_count": int(review_status.get("promotable_candidate_count") or 0),
            "active_execution_ids": active_ids,
            "active_native_subprocess_execution_ids": active_native_subprocess_ids,
            "active_native_subprocesses": active_native_subprocesses,
            "active_self_review_ids": active_self_review_ids,
            "active_thread_execution_ids": active_thread_ids,
            "active_process_execution_ids": active_process_ids,
            "executor_instance_id": self._config.executor_instance_id,
            "executor_started_at_unix": self._started_at_unix,
            "self_review_draining": self._self_review_draining,
        }

    def _self_review_config(self) -> Any:
        from self_review_queue import SelfReviewConfig  # type: ignore

        return SelfReviewConfig.from_env(
            executor_instance_id=self._config.executor_instance_id,
            pod_generation=os.getenv("POD_UID") or f"{self._config.executor_instance_id}:{int(self._started_at_unix)}",
            agent_identity=os.getenv("RSI_HERMES_SELF_REVIEW_IDENTITY") or self._config.honcho_ai_peer,
            memory_backend=self._config.memory_backend,
            honcho_workspace=self._config.honcho_workspace,
            honcho_environment=self._config.honcho_environment,
            model=self._active_model_profile.provider_model,
            provider=self._active_model_profile.provider,
            base_url=self._base_url,
            api_mode=self._api_mode,
        )

    def _self_review_worker_env(self) -> dict[str, str]:
        env: dict[str, str] = {
            key: value
            for key in ("PATH", "PYTHONPATH", "VIRTUAL_ENV", "HOME", "USER", "TMPDIR", "LANG", "LC_ALL")
            if (value := os.getenv(key))
        }
        for key in SELF_REVIEW_SECRET_ENV_ALLOWLIST:
            value = os.getenv(key)
            if value:
                env[key] = value
        for key in SELF_REVIEW_STATE_ENV_ALLOWLIST:
            value = os.getenv(key)
            if value:
                env[key] = value
        env.update(
            {
                "HERMES_HOME": self._config.hermes_home,
                "HERMES_STATE_DB_PATH": str(Path(self._config.hermes_home) / "state.db"),
                "RSI_HERMES_EXECUTOR_INSTANCE_ID": self._config.executor_instance_id,
                "POD_UID": os.getenv("POD_UID") or f"{self._config.executor_instance_id}:{int(self._started_at_unix)}",
                "RSI_HERMES_SELF_REVIEW_IDENTITY": os.getenv("RSI_HERMES_SELF_REVIEW_IDENTITY") or self._config.honcho_ai_peer,
                "RSI_MEMORY_BACKEND": self._config.memory_backend,
                "RSI_HONCHO_WORKSPACE": self._config.honcho_workspace,
                "RSI_HONCHO_ENVIRONMENT": self._config.honcho_environment,
                "RSI_MODEL": self._active_model_profile.provider_model,
                "RSI_PROVIDER": self._active_model_profile.provider,
                "RSI_HERMES_API_MODE": self._api_mode,
            }
        )
        for key in (
            "RSI_HERMES_SELF_REVIEW_STALE_AFTER_SECONDS",
            "RSI_HERMES_SELF_REVIEW_RETENTION_DAYS",
            "RSI_HERMES_SELF_REVIEW_MAX_BATCH_ROWS",
            "RSI_HERMES_SELF_REVIEW_MAX_BATCH_TOKENS",
            "RSI_HERMES_SELF_REVIEW_CREDENTIAL_PROFILE",
            "RSI_HERMES_PIN",
        ):
            value = os.getenv(key)
            if value:
                env[key] = value
        if self._base_url:
            env["OPENAI_BASE_URL"] = self._base_url
        if self._config.honcho_base_url:
            env["RSI_HONCHO_BASE_URL"] = self._config.honcho_base_url
        honcho_key = os.getenv("HONCHO_API_KEY")
        if honcho_key:
            env["HONCHO_API_KEY"] = honcho_key
        return env

    def _self_review_queue_status(self, *, reconcile_stale: bool) -> JsonObject:
        try:
            from self_review_queue import review_queue_status  # type: ignore

            result = review_queue_status(self._self_review_config(), reconcile_stale=reconcile_stale)
            return result if isinstance(result, dict) else {}
        except Exception as exc:
            logger.debug("self-review queue status unavailable: %s", exc)
            return {}

    def wait_for_active_executions(self, timeout_seconds: float) -> JsonObject:
        deadline = time.monotonic() + max(0.0, float(timeout_seconds or 0))
        last_snapshot = self.active_execution_snapshot()
        while int(last_snapshot.get("active_execution_count") or 0) > 0:
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                last_snapshot["drain_status"] = "timeout"
                last_snapshot["deadline_unix"] = time.time()
                return last_snapshot
            if self._self_review_draining and int(last_snapshot.get("local_pending_self_review_count") or 0) > 0:
                self._advance_local_self_review_work()
            execution_ids = list(last_snapshot.get("active_execution_ids") or [])
            with self._executor_process_lock:
                threads = [
                    self._executor_threads[execution_id]
                    for execution_id in execution_ids
                    if execution_id in self._executor_threads and self._executor_threads[execution_id].is_alive()
                ]
            if threads:
                threads[0].join(timeout=min(1.0, max(0.0, remaining)))
            else:
                time.sleep(min(1.0, max(0.0, remaining)))
            last_snapshot = self.active_execution_snapshot()
        last_snapshot["drain_status"] = "drained"
        last_snapshot["deadline_unix"] = time.time()
        return last_snapshot

    def _advance_local_self_review_work(self) -> None:
        try:
            from self_review_queue import advance_local_review_queue  # type: ignore

            advance_local_review_queue(self._self_review_config())
        except Exception as exc:
            logger.warning("local self-review drain advance failed: %s", exc)

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
        shared = shared or status in DIRECT_DELIVERY_SUCCESS_STATUSES
        if not shared:
            return artifacts
        return [{**artifact, "share_status": "shared"} for artifact in artifacts]

    def _reply_delivery_succeeded(self, delivery_output: JsonObject) -> bool:
        reply_delivery = _json_object_or_empty(delivery_output.get("reply_delivery"))
        if not reply_delivery:
            return False
        status = _string_or_json(reply_delivery.get("status")).lower()
        if status == "failed":
            return False
        return status in DIRECT_DELIVERY_SUCCESS_STATUSES or bool(
            _string_or_json(reply_delivery.get("provider_ref")) or _string_or_json(reply_delivery.get("message_link"))
        )

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

    def _native_execution_started_payload(self, task: RunnerTaskRequest) -> JsonObject:
        profile = self._model_profile_for_task(task)
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
            "timeout_seconds": self._effective_task_timeout(task),
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "model": profile.provider_model,
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
            render_prompt=False,
            expand_skills=False,
        )

    def _execution_phase(self, task: RunnerTaskRequest) -> str:
        return (task.execution_phase or "").strip().lower() or "main"

    def _task_uses_artifact_phases(self, task: RunnerTaskRequest) -> bool:
        return False

    def _phase_max_iterations_override(self, task: RunnerTaskRequest) -> int | None:
        return None

    def _native_toolsets_for_task(self, task: RunnerTaskRequest, *, extra_toolsets: list[str] | None = None) -> list[str]:
        toolsets: list[str] = []
        execution_phase = self._execution_phase(task)
        execution_mode = (task.execution_mode or "").strip().lower()
        if execution_phase not in {"render", "deliver"} and task.task_type in {"workflow", "proposal"}:
            toolsets.extend(["todo", "session_search"])
        toolsets.extend(self._hermes_native_toolsets())
        toolsets.extend(self._rsi_native_toolsets_for_task(task))
        if execution_phase == "render" or task.task_type in {"workflow", "prod", "proactive"} or execution_mode == "artifact_render":
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
        profile = self._model_profile_for_task(task)
        api_key = _runtime_api_key_for_provider(profile.provider)
        if not api_key:
            raise RuntimeError(self._credentials_error_message_for_profile(profile))
        configured_max_iterations = max_iterations_override if max_iterations_override and max_iterations_override > 0 else self._max_iterations
        agent_kwargs: JsonObject = {
            "model": profile.provider_model,
            "quiet_mode": True,
            "reasoning_config": _reasoning_config_for_profile(profile),
            "enabled_toolsets": enabled_toolsets_override
            if enabled_toolsets_override is not None
            else self._native_toolsets_for_task(task),
            "skip_context_files": True,
            "skip_memory": False,
            "max_iterations": configured_max_iterations,
            "session_id": context.session_id,
            "parent_session_id": context.parent_session_id or None,
            "session_db": self._session_manager.session_db,
            "self_review_mode": "manual",
        }
        if profile.provider:
            agent_kwargs["provider"] = profile.provider
        if self._api_mode:
            agent_kwargs["api_mode"] = self._api_mode
        if profile.base_url:
            agent_kwargs["base_url"] = profile.base_url
        if api_key:
            agent_kwargs["api_key"] = api_key
        self._apply_openrouter_provider_routing(
            agent_kwargs,
            provider=profile.provider,
            provider_routing=profile.openrouter_provider_routing,
        )
        return AIAgent(**agent_kwargs)

    def _model_profile_for_task(self, task: RunnerTaskRequest) -> ModelProfile:
        override = _json_object_or_empty(task.model_override)
        if not override:
            return self._model_profiles["main"]
        unexpected = sorted(key for key in override.keys() if key != "profile")
        if unexpected:
            raise ValueError("model_override only supports the server-side profile selector")
        profile_name = _string_or_json(override.get("profile")).strip().lower()
        if profile_name not in {"main", "summary"}:
            raise ValueError("model_override.profile must be one of: main, summary")
        return self._model_profiles[profile_name]

    def _credentials_error_message_for_profile(self, profile: ModelProfile) -> str:
        if profile.provider == "deepseek":
            return "Hermes DeepSeek runtime selected but RSI_DEEPSEEK_API_KEY / DEEPSEEK_API_KEY is not configured."
        if profile.provider == "openrouter":
            return "Hermes OpenRouter runtime selected but RSI_OPENROUTER_API_KEY / OPENROUTER_API_KEY is not configured."
        return f"Hermes runtime provider {profile.provider or 'unknown'} is missing credentials."

    def _apply_openrouter_provider_routing(
        self,
        agent_kwargs: JsonObject,
        *,
        provider: str | None = None,
        provider_routing: JsonObject | None = None,
    ) -> None:
        provider_name = provider or self._provider
        routing = dict(provider_routing or self._provider_routing or {})
        if provider_name != "openrouter":
            return
        if not routing:
            return
        if isinstance(routing.get("only"), list):
            agent_kwargs["providers_allowed"] = list(routing["only"])  # type: ignore[index]
        if isinstance(routing.get("ignore"), list):
            agent_kwargs["providers_ignored"] = list(routing["ignore"])  # type: ignore[index]
        if isinstance(routing.get("order"), list):
            agent_kwargs["providers_order"] = list(routing["order"])  # type: ignore[index]
        if isinstance(routing.get("sort"), str):
            agent_kwargs["provider_sort"] = routing["sort"]
        if isinstance(routing.get("require_parameters"), bool):
            agent_kwargs["provider_require_parameters"] = routing["require_parameters"]
        if isinstance(routing.get("data_collection"), str):
            agent_kwargs["provider_data_collection"] = routing["data_collection"]

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

    def _prompt_skill_mentions(self, task: RunnerTaskRequest) -> list[str]:
        user_request = self._extract_user_request_text(task.prompt)
        explicit_mentions = self._skill_mentions_from_text(user_request)
        prompt_skills: list[str] = []
        seen_skills: set[str] = set()
        for identifier in explicit_mentions:
            normalized = identifier.replace("_", "-").strip().lower()
            if not normalized or normalized in seen_skills:
                continue
            seen_skills.add(normalized)
            prompt_skills.append(normalized)
        return prompt_skills

    def _expand_task_skills(self, task: RunnerTaskRequest, context: SessionContext) -> tuple[RunnerTaskRequest, JsonObject]:
        diagnostics: JsonObject = {
            "prompt_skills": [],
            "resolved_skills": [],
            "missing_skills": [],
            "skill_injection_mode": "none",
        }
        user_request = self._extract_user_request_text(task.prompt)
        prompt_skills = self._prompt_skill_mentions(task)
        diagnostics["prompt_skills"] = list(prompt_skills)
        if not prompt_skills:
            return task, diagnostics
        if build_skill_invocation_message is None or build_preloaded_skills_prompt is None or resolve_skill_command_key is None:
            diagnostics["missing_skills"] = list(prompt_skills)
            diagnostics["skill_injection_mode"] = "helpers_unavailable"
            return task, diagnostics

        prompt_prefix_parts: list[str] = []
        resolved_skills: list[str] = []
        missing_skills: list[str] = []
        injection_modes: list[str] = []

        remaining_preloads = list(prompt_skills)
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
        *,
        observer: ObservationEmitter | None = None,
        allow_partial_recovery: bool = True,
        max_iterations_override: int | None = None,
        render_prompt: bool = True,
        expand_skills: bool = True,
        enabled_toolsets_override: list[str] | None = None,
        repair_instruction: str | None = None,
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
        try:
            task_model_profile = self._model_profile_for_task(task)
        except ValueError as exc:
            return HermesExecutionResult(
                ok=False,
                message=str(exc),
                provider=self._backend,
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "failure_class": "invalid_model_override",
                    "model_override": task.model_override,
                },
            )
        if not _runtime_api_key_for_provider(task_model_profile.provider):
            return HermesExecutionResult(
                ok=False,
                message=self._credentials_error_message_for_profile(task_model_profile),
                provider=self._backend,
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    "failure_class": "model_profile_credentials_missing",
                    "model_profile": task_model_profile.name,
                    "provider": task_model_profile.provider,
                    "provider_model": task_model_profile.provider_model,
                },
            )
        if not self._session_manager.available:
            return HermesExecutionResult(
                ok=False,
                message="Hermes persistent session runtime is unavailable.",
                provider=self._backend,
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )

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
            "prompt_skills": [],
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
            task = replace(task, prompt=self._render_task_prompt(task))
        effective_task_timeout = self._effective_task_timeout(task)
        effective_inactivity_timeout = self._effective_inactivity_timeout(effective_task_timeout)
        reasoning_timeout_seconds = self._partial_completion_reasoning_timeout_seconds(task, effective_task_timeout)
        configured_max_iterations = max_iterations_override if max_iterations_override and max_iterations_override > 0 else self._max_iterations
        agent = None
        tracker = None
        agentic_mcp_registration = TaskScopedMCPRegistration()
        result: HermesExecutionResult | None = None
        try:
            self._stage_task_context(context.session_id, task)
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
                        "runner_diagnostics": self._runner_diagnostics(
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
                enabled_toolsets_override=enabled_toolsets_override
                if enabled_toolsets_override is not None
                else self._native_toolsets_for_task(
                    task, extra_toolsets=agentic_mcp_registration.enabled_toolsets
                ),
            )
            tracker = self._session_manager.attach_tracking(agent, task, context)
            termination_reason, run_result, stop_meta = self._run_with_deadlines(
                agent,
                task,
                context,
                effective_task_timeout,
                effective_inactivity_timeout,
                reasoning_timeout_seconds,
                observer=observer,
                repair_instruction=repair_instruction,
            )
            lifecycle_events = self._adapter.lifecycle_events(context.session_id)
            if termination_reason != "normal_completion":
                finalized = self._session_manager.finalize(context, tracker)
                observed = self._observability_metadata(
                    agent,
                    task,
                    tracker,
                    skill_diagnostics=skill_diagnostics,
                    observer=observer,
                    lifecycle_events=lifecycle_events,
                )
                last_activity = _json_object_or_empty(stop_meta.get("last_activity"))
                if termination_reason in PARTIAL_COMPLETION_TERMINATION_REASONS and allow_partial_recovery and self._workflow_partial_completion_eligible(task):
                    result = self._finalize_partial_completion(
                        task,
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
                            **observed,
                            **self._workflow_evidence_raw(task, observed, timeout_kind),
                            "failure_class": "runner_transport_timeout",
                            "runner_diagnostics": self._runner_diagnostics(
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
                            **observed,
                            **self._workflow_evidence_raw(task, observed, "task_timeout"),
                            "failure_class": "runner_transport_timeout",
                            "runner_diagnostics": self._runner_diagnostics(
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
            diagnostics = self._provider_invalid_request_diagnostics(str(exc))
            provider_runtime_error = self._provider_runtime_error_diagnostics(str(exc)) if diagnostics is None else None
            activity = safe_activity_summary(agent) if agent is not None else {}
            lifecycle_events = self._adapter.lifecycle_events(context.session_id)
            observed = self._observability_metadata(
                agent,
                task,
                skill_diagnostics=skill_diagnostics,
                observer=observer,
                lifecycle_events=lifecycle_events,
            )
            raw = {
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                "error": str(exc),
                **observed,
                **self._workflow_evidence_raw(task, observed, "exception"),
                "lifecycle_events": lifecycle_events,
            }
            if diagnostics is not None:
                raw["failure_class"] = "runner_invalid_request"
                merged = dict(diagnostics)
                for key, value in observed.items():
                    merged[key] = value
                raw["runner_diagnostics"] = merged
            elif provider_runtime_error is not None:
                raw["failure_class"] = "runner_provider_error"
                merged = dict(provider_runtime_error)
                for key, value in observed.items():
                    merged[key] = value
                raw["runner_diagnostics"] = merged
            else:
                raw["failure_class"] = "runner_non_ok"
                raw["runner_diagnostics"] = self._runner_diagnostics(
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
        observed = self._observability_metadata(
            agent,
            task,
            tracker,
            skill_diagnostics=skill_diagnostics,
            observer=observer,
            lifecycle_events=lifecycle_events,
        )
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
            "model_profile": self._active_model_profile.name,
            "provider_model": self._provider_model,
            "base_url_host": _base_url_host(self._base_url),
            "thinking_mode": self._active_model_profile.thinking_mode,
            "summary_provider": self._model_profiles["summary"].provider,
            "summary_model": self._model_profiles["summary"].provider_model,
            "summary_thinking_mode": self._model_profiles["summary"].thinking_mode,
            "provider_routing": dict(self._provider_routing) if self._provider == "openrouter" else {},
            "reasoning_effort": self._reasoning_effort,
            "reasoning_config": self._reasoning_config,
            "hermes_version": adapter_meta.version,
            "hermes_pin": adapter_meta.pin,
            "hermes_contract_status": self._contract_status.to_dict(),
            "context_engine_mode": adapter_meta.context_engine_mode,
            "context_engine_status": adapter_meta.context_engine_status,
            "lifecycle_hook_status": adapter_meta.lifecycle_hook_status,
            "base_url": self._base_url,
            "hermes_executor_workspace_root": self._config.hermes_executor_workspace_root,
            "hermes_computer_root": self._config.hermes_computer_root,
            "hermes_run_root": self._config.hermes_run_root,
            "hermes_artifact_root": self._config.hermes_artifact_root,
            "company_wiki_root": self._config.company_wiki_root,
            "hermes_native_terminal_enabled": self._config.hermes_native_terminal_enabled,
            "hermes_native_toolsets": self._hermes_native_toolsets(),
            "hermes_terminal_cwd": self._config.hermes_terminal_cwd,
            "hermes_company_bin_dir": self._config.hermes_company_bin_dir,
            "hermes_kubernetes_context_enabled": self._config.hermes_kubernetes_context_enabled,
            "hermes_kubeconfig_path": self._config.hermes_kubeconfig_path if (self._config.hermes_kubernetes_context_enabled or self._config.hermes_prod_kubernetes_context_enabled) else "",
            "company_computer_bootstrap_status": dict(self._company_computer_bootstrap_status),
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

    def start_executor_task(self, task: RunnerTaskRequest) -> JsonObject:
        execution_id = _string_or_json(task.execution_id) or execution_observation_id(
            task.operation_id or "",
            task.trace_id or "",
            task.workflow_id or "",
            "",
        )
        task.execution_id = execution_id
        
        with self._executor_process_lock:
            if execution_id in self._executor_threads or execution_id in self._executor_processes:
                existing = self.executor_status(execution_id)
                return existing
            
            existing = self.executor_status(execution_id)
            existing_status = str(existing.get("status") or "").strip().lower()
            if existing_status in {"accepted", "running", "finalizing", "cancelling", "completed", "failed", "cancelled"}:
                return existing
            
            accepted = {
                "execution_id": execution_id,
                "operation_id": task.operation_id or "",
                "trace_id": task.trace_id or "",
                "workflow_id": task.workflow_id or "",
                "executor_instance_id": self._config.executor_instance_id,
                "executor_started_at_unix": self._started_at_unix,
                "status_file_path": str(self._executor_status_path(execution_id)),
                "status": "accepted",
                "message": "Execution accepted.",
                "phase": self._execution_phase(task),
            }

            def _run() -> None:
                try:
                    self.execute_task(task)
                except Exception as exc:  # pragma: no cover - background fault protection
                    logger.exception("async Hermes execution failed execution_id=%s", execution_id)
                    result = HermesExecutionResult(
                        ok=False,
                        message=f"Async Hermes execution failed: {exc}",
                        provider="hermes-executor",
                        raw={"failure_class": "runner_executor_async_failed", "error": str(exc)},
                    )
                    self._store_executor_result(
                        execution_id,
                        self._executor_final_status(task, result, execution_id=execution_id, status="failed"),
                    )
                finally:
                    with self._executor_process_lock:
                        self._executor_threads.pop(execution_id, None)

            self._store_executor_result(execution_id, accepted)
            thread = threading.Thread(target=_run, name=f"rsi-hermes-exec-{execution_id}", daemon=True)
            self._executor_threads[execution_id] = thread
            thread.start()
            return accepted

    def executor_status(self, execution_id: str) -> JsonObject:
        key = str(execution_id or "").strip()
        if not key:
            return {}
        cached = self._executor_recent_results.get(key)
        if cached:
            cached_status = str(cached.get("status") or "").strip().lower()
            if cached_status in {"running", "accepted", "starting", "finalizing", "cancelling", "cancel_requested"}:
                with self._executor_process_lock:
                    active = key in self._executor_processes or key in self._executor_threads
                if not active:
                    orphaned = self._orphaned_executor_status(
                        key,
                        cached,
                        "Cached executor status was running, but no local execution process is active.",
                    )
                    self._executor_recent_results[key] = orphaned
                    return orphaned
            return self._merge_self_review_status(key, dict(cached))
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
        if str(payload.get("status") or "").strip().lower() in {"running", "accepted", "starting", "finalizing", "cancelling", "cancel_requested"}:
            with self._executor_process_lock:
                active = key in self._executor_processes or key in self._executor_threads
            if not active:
                payload = self._orphaned_executor_status(
                    key,
                    payload,
                    "Persisted executor status was running, but no local execution process is active.",
                )
        payload = self._merge_self_review_status(key, payload)
        self._executor_recent_results[key] = dict(payload)
        return dict(payload)

    def _orphaned_executor_status(self, execution_id: str, payload: JsonObject, message: str) -> JsonObject:
        path = self._executor_status_path(execution_id)
        out = dict(payload)
        out["status"] = "orphaned"
        out["message"] = message
        out["executor_instance_id"] = first_non_empty(
            _string_or_json(out.get("executor_instance_id")),
            self._config.executor_instance_id,
        )
        out["current_executor_instance_id"] = self._config.executor_instance_id
        if "executor_started_at_unix" not in out:
            out["executor_started_at_unix"] = self._started_at_unix
        out["current_executor_started_at_unix"] = self._started_at_unix
        out["status_file_path"] = str(path)
        out["last_observed_status"] = str(payload.get("status") or "").strip()
        if path.exists():
            try:
                stat = path.stat()
                out["status_file_mtime_unix"] = stat.st_mtime
                out["status_file_size_bytes"] = stat.st_size
            except OSError:
                pass
        if "last_observed_ledger_seq" not in out:
            out["last_observed_ledger_seq"] = out.get("last_ledger_seq", "")
        return out

    def _merge_self_review_status(self, execution_id: str, payload: JsonObject) -> JsonObject:
        try:
            from self_review_queue import candidate_status  # type: ignore

            status = candidate_status(self._self_review_config(), execution_id)
        except Exception as exc:
            logger.debug("self-review status merge unavailable execution_id=%s error=%s", execution_id, exc)
            return dict(payload)
        if not isinstance(status, dict) or not status:
            return dict(payload)
        return self._with_self_review_status(payload, status)

    def _with_self_review_status(self, payload: JsonObject, status: JsonObject) -> JsonObject:
        out = dict(payload)
        result = dict(_json_object_or_empty(out.get("result")))
        raw = _json_object_or_empty(result.get("raw"))
        existing = {
            **_json_object_or_empty(raw.get("self_review")),
            **_json_object_or_empty(out.get("self_review")),
        }
        merged = {**existing, **status}
        out["self_review"] = merged
        if raw:
            result["raw"] = {**raw, "self_review": merged}
            out["result"] = result
        return out

    def _persist_self_review_status(self, execution_id: str, final_status: JsonObject, status: JsonObject, result: HermesExecutionResult) -> None:
        if not status:
            return
        result.raw = {**_json_object_or_empty(result.raw), "self_review": {**_json_object_or_empty(result.raw.get("self_review")), **status}}
        self._store_executor_result(execution_id, self._with_self_review_status(final_status, status))

    def _self_review_promotion_status(
        self,
        *,
        candidate: JsonObject,
        delivered: JsonObject | None = None,
        promoted: JsonObject | None = None,
        candidate_status_payload: JsonObject | None = None,
        fallback_status: str = "",
        last_error: str = "",
    ) -> JsonObject:
        promoted = _json_object_or_empty(promoted)
        delivered = _json_object_or_empty(delivered)
        candidate_status_payload = _json_object_or_empty(candidate_status_payload)
        work_created = promoted.get("work_created") if isinstance(promoted.get("work_created"), list) else []
        candidate_state = first_non_empty(
            _string_or_json(candidate_status_payload.get("self_review_candidate_status")),
            _string_or_json(promoted.get("status")),
            _string_or_json(delivered.get("status")),
            fallback_status,
        )
        review_status = _string_or_json(candidate_status_payload.get("self_review_status"))
        if not review_status:
            review_status = "queued" if work_created else candidate_state
        out: JsonObject = {
            "candidate_id": candidate.get("candidate_id") or promoted.get("candidate_id"),
            "execution_id": first_non_empty(_string_or_json(candidate.get("execution_id")), _string_or_json(promoted.get("execution_id"))),
            "agent_identity": first_non_empty(_string_or_json(candidate.get("agent_identity")), _string_or_json(promoted.get("agent_identity"))),
            "gateway_session_key": first_non_empty(_string_or_json(candidate.get("gateway_session_key")), _string_or_json(promoted.get("gateway_session_key"))),
            "cadence_scope_key": first_non_empty(_string_or_json(promoted.get("cadence_scope_key")), _string_or_json(candidate.get("cadence_scope_key"))),
            "self_review_candidate_status": candidate_state,
            "self_review_status": review_status,
            "self_review_enqueue_status": _string_or_json(promoted.get("status")),
            "self_review_last_error": first_non_empty(last_error, _string_or_json(candidate_status_payload.get("self_review_last_error"))),
            "memory_turns_after": promoted.get("memory_turns_after"),
            "skill_iterations_after": promoted.get("skill_iterations_after"),
            "memory_nudge_interval": candidate.get("memory_nudge_interval"),
            "skill_nudge_interval": candidate.get("skill_nudge_interval"),
            "review_memory": promoted.get("review_memory"),
            "review_skills": promoted.get("review_skills"),
            "work_created": work_created,
        }
        if len(candidate_status_payload) > 0:
            out = {**out, **candidate_status_payload}
            for key in ("gateway_session_key", "cadence_scope_key", "memory_turns_after", "skill_iterations_after", "memory_nudge_interval", "skill_nudge_interval", "work_created"):
                if out.get(key) in (None, "", []):
                    out[key] = promoted.get(key) if key in promoted else candidate.get(key)
        return {key: value for key, value in out.items() if value not in (None, "")}

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

    def _result_is_native_strict(self, result: HermesExecutionResult) -> bool:
        return result.provider == "hermes-native-executor" and _bool_or_false(_json_object_or_empty(result.raw).get("native_strict"))

    def _native_strict_has_successful_slack_delivery(self, result: HermesExecutionResult) -> bool:
        raw = _json_object_or_empty(result.raw)
        if self._rsi_native_slack_reply_delivery_succeeded(_json_object_or_empty(raw.get("reply_delivery"))):
            return True
        envelope = _json_object_or_empty(raw.get("execution_envelope"))
        for delivery in _json_object_list(envelope.get("deliveries")):
            if self._rsi_native_slack_reply_delivery_succeeded(delivery):
                return True
        return False

    def _native_strict_should_project_workflow_output(self, task: RunnerTaskRequest, result: HermesExecutionResult) -> bool:
        if not self._result_is_native_strict(result) or not result.ok:
            return False
        if not self._workflow_requires_explicit_reply_action(task):
            return False
        raw = _json_object_or_empty(result.raw)
        if _json_object_or_empty(raw.get("structured_output")):
            return False
        termination_reason = string_from_map(raw, "termination_reason")
        if termination_reason == "external_tool_pending":
            return False
        completion_verdict = string_from_map(raw, "completion_verdict")
        if completion_verdict and completion_verdict not in {"complete", "completed"}:
            return False
        if _bool_or_false(raw.get("suppress_delivery")) or _bool_or_false(raw.get("non_deliverable")):
            return False
        return not self._native_strict_has_successful_slack_delivery(result)

    def _native_runtime_context_path(self, execution_id: str) -> Path:
        return (
            Path(self._config.hermes_home).expanduser()
            / "rsi_runtime"
            / "context"
            / "executions"
            / f"{self._native_execution_slug(execution_id)}.json"
        ).resolve()

    def _native_runtime_envelope_path(self, execution_id: str) -> Path:
        return (
            Path(self._config.hermes_home).expanduser()
            / "rsi_runtime"
            / "envelopes"
            / f"{self._native_execution_slug(execution_id)}.json"
        ).resolve()

    def _db_read_execution_token(self, task: RunnerTaskRequest) -> str:
        secret = os.getenv("RSI_DB_READ_CLIENT_TOKEN", "").strip()
        if not self._config.db_read_gateway_configured or not secret:
            return ""
        now = int(time.time())
        claims: JsonObject = {
            "version": "v1",
            "db_read_query_allowed": True,
            "execution_id": task.execution_id or "",
            "operation_id": task.operation_id or "",
            "conversation_id": task.conversation_id or "",
            "workflow_id": task.workflow_id or "",
            "trace_id": task.trace_id or "",
            "channel_id": task.channel_id or "",
            "thread_ts": task.thread_ts or task.message_ts or _derive_root_message_ts(task) or "",
            "requester": task.user_peer_id or "hermes",
            "iat": now,
            "exp": now + max(60, min(self._effective_task_timeout(task) + 300, 7200)),
        }
        raw_claims = json.dumps(claims, separators=(",", ":"), sort_keys=True).encode("utf-8")
        encoded_claims = base64.urlsafe_b64encode(raw_claims).decode("ascii").rstrip("=")
        signature = hmac.HMAC(secret.encode("utf-8"), encoded_claims.encode("ascii"), hashlib.sha256).digest()
        encoded_signature = base64.urlsafe_b64encode(signature).decode("ascii").rstrip("=")
        return f"v1.{encoded_claims}.{encoded_signature}"

    def _native_runtime_env(self, task: RunnerTaskRequest, *, context_path: Path, envelope_path: Path) -> JsonObject:
        env: JsonObject = {
            "RSI_RUNTIME_CONTEXT_PATH": str(context_path),
            "RSI_RUNTIME_ENVELOPE_PATH": str(envelope_path),
            "RSI_EXECUTION_ID": task.execution_id or "",
            "RSI_OPERATION_ID": task.operation_id or "",
            "RSI_TRACE_ID": task.trace_id or "",
            "RSI_WORKFLOW_ID": task.workflow_id or "",
            "RSI_CONVERSATION_ID": task.conversation_id or "",
            "RSI_SLACK_CHANNEL_ID": task.channel_id or "",
            "RSI_SLACK_THREAD_TS": task.thread_ts or task.message_ts or _derive_root_message_ts(task) or "",
            "RSI_TASK_REQUESTER": task.user_peer_id or "hermes",
        }
        token = self._db_read_execution_token(task)
        if token:
            env["RSI_DB_READ_EXECUTION_TOKEN"] = token
            env["RSI_DB_READ_AUTH_MODE"] = "execution_scoped"
        native_tools_token = self._native_tools_execution_token(task)
        if native_tools_token:
            env["RSI_NATIVE_TOOLS_EXECUTION_TOKEN"] = native_tools_token
        return env

    def _native_tools_execution_token(self, task: RunnerTaskRequest) -> str:
        secret = os.getenv("RSI_NATIVE_TOOLS_CLIENT_TOKEN", "").strip()
        if not secret:
            raise RuntimeError("RSI_NATIVE_TOOLS_CLIENT_TOKEN is required; RSI native tools are the canonical tool path")
        now = int(time.time())
        lifetime = min(max(1, self._effective_task_timeout(task) + 60), 2 * 60 * 60)
        payload: JsonObject = {
            "aud": "rsi-native-tools",
            "iat": now,
            "exp": now + lifetime,
            "execution_id": first_non_empty(task.execution_id, task.trace_id, task.workflow_id, task.session_scope_id, "execution"),
            "operation_id": first_non_empty(task.operation_id, task.execution_id, task.trace_id, "operation"),
            "trace_id": first_non_empty(task.trace_id, task.execution_id, task.workflow_id, "trace"),
            "workflow_id": first_non_empty(task.workflow_id, task.trace_id, task.execution_id, "workflow"),
            "conversation_id": first_non_empty(task.conversation_id, task.session_scope_id, task.channel_id, "conversation"),
            "actor": first_non_empty(task.user_peer_id, task.assistant_peer_id, "hermes"),
            "surfaces": list(self._config.native_tools_surfaces or ["slack", "notion", "knowledge", "sentry"]),
            "slack_channel_id": task.channel_id or "",
            "slack_thread_ts": task.thread_ts or task.message_ts or _derive_root_message_ts(task) or "",
            "slack_delivery_scope": "bound_thread" if task.channel_id else "",
        }
        return _sign_native_tools_execution_token(secret, payload)

    def _legacy_mcp_server_errors(self, task: RunnerTaskRequest) -> list[str]:
        errors: list[str] = []
        for index, server in enumerate(task.mcp_servers or []):
            label = _string_or_json(server.get("server_label")).strip().lower()
            profile = _string_or_json(server.get("profile")).strip().lower()
            profile_tokens = {token for token in re.split(r"[^a-z0-9]+", profile) if token}
            is_reserved_profile = "mcp" in profile_tokens and bool(profile_tokens & {"slack", "notion"})
            if is_reserved_profile or label in {"slack", "notion"}:
                name = profile or label or f"server[{index}]"
                errors.append(f"{name} must use RSI native tools instead")
        return errors

    def _native_worker_session_env(self, task: RunnerTaskRequest) -> dict[str, str]:
        env = {"HERMES_SESSION_KEY": canonical_gateway_session_key(task, self._role)}
        channel_id = (task.channel_id or "").strip()
        if channel_id:
            env["HERMES_SESSION_PLATFORM"] = "slack"
            env["HERMES_SESSION_CHAT_ID"] = channel_id
        thread_ts = (task.thread_ts or "").strip()
        if thread_ts:
            env["HERMES_SESSION_THREAD_ID"] = thread_ts
        return env

    def _strip_native_worker_source_credentials(self, env: dict[str, str]) -> None:
        for key in NATIVE_WORKER_SOURCE_CREDENTIAL_DENYLIST:
            env.pop(key, None)

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
            "context_refs": task.context_refs,
            "contract_version": task.contract_version or EXECUTION_CONTRACT_VERSION,
            "execution_intent": task.execution_intent,
            "external_tool_resume": task.external_tool_resume,
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
        required_toolsets.extend(self._hermes_native_toolsets())
        required_toolsets.extend(self._rsi_native_toolsets_for_task(task))
        return {
            "execution_phase": execution_phase,
            "history_policy": "empty" if execution_phase in {"render", "deliver"} else "session",
            "required_toolsets": required_toolsets,
            "toolsets": list(toolsets),
            "missing_required_toolsets": [item for item in required_toolsets if item not in set(toolsets)],
        }

    def _native_required_final_tool_names(self, task: RunnerTaskRequest) -> list[str]:
        if not self._workflow_requires_explicit_reply_action(task):
            return []
        if self._execution_phase(task) in {"render", "deliver"}:
            return []
        return ["rsi_slack_message_post", "rsi_slack_report_post"]

    def _native_required_final_tool_instruction(self, task: RunnerTaskRequest) -> str:
        if not self._native_required_final_tool_names(task):
            return ""
        return (
            "This RSI workflow is Slack-bound. Deliver the final response by calling "
            "exactly one RSI native Slack tool in the bound channel/thread: "
            "rsi_slack.message_post for simple prose, or rsi_slack.report_post for "
            "rich/tabular output using report_schema_version=1 and structured tables. "
            "Do not use generic send_message, legacy proposed Slack actions, or raw "
            "Markdown pipe tables for tabular Slack output."
        )

    def _github_cli_environment(self, task: RunnerTaskRequest) -> tuple[dict[str, str], JsonObject]:
        status: JsonObject = {
            "configured": False,
            "status": "disabled",
            "reason": "",
            "repo": task.repo,
            "repositories": [],
        }
        if not self._config.hermes_native_terminal_enabled:
            status["reason"] = "native_terminal_disabled"
            return {}, status
        if self._role not in {"prod", "proactive"}:
            status["reason"] = "runner_role_not_eligible"
            return {}, status
        repo = str(task.repo or "").strip()
        if not repo:
            status["reason"] = "missing_repo"
            return {}, status
        repos = normalize_tool_names([repo, *task.repo_allowlist])
        token, app_status = self._github_app_installation_token(repo, repos, scope_repositories=False)
        if not token:
            status.update(app_status)
            return {}, status
        owner = _string_or_json(app_status.get("owner")) or self._github_owner_for_repo(repo)
        tokens_by_owner: dict[str, str] = {owner: token}
        owner_statuses: list[JsonObject] = [self._github_public_token_status(app_status)]
        for extra_owner in self._github_cli_owners_for_task(task):
            if not extra_owner or extra_owner == owner or extra_owner in tokens_by_owner:
                continue
            extra_token, extra_status = self._github_app_installation_token_for_owner(extra_owner, repo=repo)
            owner_statuses.append(self._github_public_token_status(extra_status))
            if not extra_token:
                status.update(extra_status)
                return {}, status
            tokens_by_owner[extra_owner] = extra_token
        status.update(
            {
                "configured": True,
                "status": "configured",
                "reason": "github_app_installation",
                "repo": repo,
                "repositories": _string_list_or_empty(app_status.get("repositories")) or repos,
                "provider_ref": "github_app_installation",
                "owner": owner,
                "owners": sorted(tokens_by_owner),
                "owner_statuses": owner_statuses,
            }
        )
        expires_at = _string_or_json(app_status.get("expires_at"))
        installation_id = _string_or_json(app_status.get("installation_id"))
        if expires_at:
            status["expires_at"] = expires_at
        if installation_id:
            status["installation_id"] = installation_id
        return {
            "GH_TOKEN": token,
            "GITHUB_TOKEN": token,
            "_HERMES_FORCE_GH_TOKEN": token,
            "_HERMES_FORCE_GITHUB_TOKEN": token,
            "_HERMES_GITHUB_DEFAULT_OWNER": owner,
            "_HERMES_GITHUB_TOKEN_MAP_JSON": json.dumps(tokens_by_owner, sort_keys=True),
            "GH_PROMPT_DISABLED": "1",
        }, status

    def _github_cli_owners_for_task(self, task: RunnerTaskRequest) -> list[str]:
        owners: list[str] = []

        def add(owner: str) -> None:
            owner = str(owner or "").strip()
            if owner and owner not in owners:
                owners.append(owner)

        add(self._github_owner_for_repo(task.repo))
        add(str(os.getenv("RSI_GITHUB_OWNER", "") or "").strip())
        repo_owners = _env_map("RSI_GITHUB_REPO_OWNERS")
        for repo in [task.repo, *task.repo_allowlist]:
            add(self._github_owner_for_repo(repo))
            owner_hint, repo_name = self._github_repo_parts(repo)
            add(owner_hint)
            add(repo_owners.get(repo) or repo_owners.get(repo_name) or "")
        for owner in _env_map("RSI_GITHUB_APP_INSTALLATION_IDS"):
            add(owner)
        return owners

    def _github_owner_for_repo(self, repo: str) -> str:
        repo = str(repo or "").strip()
        owner_hint, repo_name = self._github_repo_parts(repo)
        if owner_hint:
            return owner_hint
        repo_owners = _env_map("RSI_GITHUB_REPO_OWNERS")
        owner = repo_owners.get(repo) or repo_owners.get(repo_name) or os.getenv("RSI_GITHUB_OWNER", "")
        return str(owner or "").strip()

    def _github_repo_parts(self, repo: str) -> tuple[str, str]:
        repo = str(repo or "").strip()
        if "/" not in repo:
            return "", repo
        owner, _, name = repo.partition("/")
        return owner.strip(), name.strip()

    def _github_repositories_for_owner(self, owner: str, repos: list[str]) -> list[str]:
        owner = str(owner or "").strip()
        repo_owners = _env_map("RSI_GITHUB_REPO_OWNERS")
        default_owner = str(os.getenv("RSI_GITHUB_OWNER", "") or "").strip()
        out: list[str] = []
        for repo in repos:
            repo = str(repo or "").strip()
            if not repo:
                continue
            owner_hint, repo_name = self._github_repo_parts(repo)
            repo_owner = owner_hint or repo_owners.get(repo) or repo_owners.get(repo_name) or default_owner
            if repo_owner == owner:
                out.append(repo_name)
        return normalize_tool_names(out)

    def _github_repository_guidance(self, task: RunnerTaskRequest) -> str:
        mappings: dict[str, str] = {}
        for repo in normalize_tool_names([task.repo, *task.repo_allowlist]):
            owner = self._github_owner_for_repo(repo)
            _owner_hint, repo_name = self._github_repo_parts(repo)
            if owner and repo_name:
                mappings[repo_name] = f"{owner}/{repo_name}"
        for repo_name, owner in _env_map("RSI_GITHUB_REPO_OWNERS").items():
            repo_name = str(repo_name or "").strip()
            owner = str(owner or "").strip()
            if repo_name and owner and "/" not in repo_name:
                mappings.setdefault(repo_name, f"{owner}/{repo_name}")
        owners = sorted(set(self._github_cli_owners_for_task(task)))
        parts: list[str] = []
        if owners:
            parts.append("configured GitHub App owner(s): " + ", ".join(owners))
        if mappings:
            formatted = ", ".join(f"{repo} -> {qualified}" for repo, qualified in sorted(mappings.items()))
            parts.append("known repo owner mapping(s): " + formatted)
        if not parts:
            return ""
        return (
            "GitHub repository resolution: use owner-qualified repository names for gh and git commands; "
            + "; ".join(parts)
            + "."
        )

    def _github_app_installation_token(self, repo: str, repos: list[str], *, scope_repositories: bool = True) -> tuple[str, JsonObject]:
        owner = self._github_owner_for_repo(repo)
        scoped_repos = self._github_repositories_for_owner(owner, repos) if scope_repositories else []
        return self._github_app_installation_token_for_owner(owner, repo=repo, repositories=scoped_repos if scope_repositories else None)

    def _github_app_installation_token_for_owner(
        self,
        owner: str,
        *,
        repo: str = "",
        repositories: list[str] | None = None,
    ) -> tuple[str, JsonObject]:
        app_id = str(os.getenv("RSI_GITHUB_APP_ID", "") or "").strip()
        private_key = _normalize_private_key(str(os.getenv("RSI_GITHUB_APP_PRIVATE_KEY", "") or ""))
        api_base_url = str(os.getenv("RSI_GITHUB_API_BASE_URL", "https://api.github.com") or "https://api.github.com").rstrip("/")
        owner = str(owner or "").strip()
        installation_ids = _env_map("RSI_GITHUB_APP_INSTALLATION_IDS")
        installation_id = installation_ids.get(owner) or str(os.getenv("RSI_GITHUB_APP_INSTALLATION_ID", "") or "").strip()
        scoped_repos = normalize_tool_names(repositories or [])
        status: JsonObject = {
            "configured": False,
            "status": "failed",
            "reason": "github_app_token_unavailable",
            "repo": repo,
            "repositories": scoped_repos,
            "owner": owner,
            "installation_id": installation_id,
        }
        missing = [
            name
            for name, value in (
                ("RSI_GITHUB_APP_ID", app_id),
                ("RSI_GITHUB_APP_PRIVATE_KEY", private_key),
                ("RSI_GITHUB_APP_INSTALLATION_ID", installation_id),
                ("RSI_GITHUB_OWNER", owner),
            )
            if not value
        ]
        if missing:
            status["reason"] = "missing_github_app_credentials"
            status["missing"] = missing
            return "", status
        try:
            import jwt  # type: ignore
        except ImportError:
            status["reason"] = "missing_pyjwt"
            return "", status
        now = int(time.time())
        try:
            app_jwt = jwt.encode({"iat": now - 60, "exp": now + 540, "iss": app_id}, private_key, algorithm="RS256")
        except Exception as exc:
            status["reason"] = "github_app_jwt_failed"
            status["error"] = str(exc)
            return "", status
        body: JsonObject = {}
        if repositories is not None and scoped_repos:
            body["repositories"] = scoped_repos
        request = urlrequest.Request(
            f"{api_base_url}/app/installations/{installation_id}/access_tokens",
            data=json.dumps(body).encode("utf-8"),
            method="POST",
            headers={
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {app_jwt}",
                "Content-Type": "application/json",
            },
        )
        try:
            with urlrequest.urlopen(request, timeout=30) as response:
                payload = json.loads(response.read().decode("utf-8"))
        except urlerror.HTTPError as exc:
            error_body = exc.read(2048).decode("utf-8", errors="replace")
            status["reason"] = "github_app_token_request_failed"
            status["http_status"] = exc.code
            status["error"] = _redact_subprocess_output(error_body, secret_values=_sensitive_env_values(dict(os.environ)))
            return "", status
        except Exception as exc:
            status["reason"] = "github_app_token_request_failed"
            status["error"] = str(exc)
            return "", status
        token = str(payload.get("token") or "").strip()
        if not token:
            status["reason"] = "github_app_token_response_missing_token"
            return "", status
        status.update(
            {
                "configured": True,
                "status": "configured",
                "reason": "github_app_installation",
                "expires_at": str(payload.get("expires_at") or ""),
            }
        )
        return token, status

    def _github_public_token_status(self, status: JsonObject) -> JsonObject:
        return {
            key: value
            for key, value in status.items()
            if key
            in {
                "configured",
                "status",
                "reason",
                "repo",
                "repositories",
                "owner",
                "installation_id",
                "expires_at",
                "http_status",
                "missing",
            }
        }

    def _write_github_token_store(self, request_dir: Path, env: dict[str, str]) -> Path:
        token_store_path = request_dir / "github-token-store.json"
        try:
            raw_tokens = json.loads(env.get("_HERMES_GITHUB_TOKEN_MAP_JSON", "{}"))
        except json.JSONDecodeError:
            raw_tokens = {}
        tokens = {str(key): str(value) for key, value in raw_tokens.items() if value}
        token_store_path.write_text(
            json.dumps(
                {
                    "default_owner": env.get("_HERMES_GITHUB_DEFAULT_OWNER", ""),
                    "tokens_by_owner": tokens,
                },
                sort_keys=True,
            ),
            encoding="utf-8",
        )
        token_store_path.chmod(0o600)
        return token_store_path

    def _write_github_credential_broker(self, request_dir: Path) -> Path:
        broker_path = request_dir / "rsi-github-credential-broker.py"
        broker_path.write_text(
            r'''#!/usr/bin/env python3
import json
import os
import re
import sys

OWNER_RE = r"[A-Za-z0-9_.-]+"
REPO_RE = r"[A-Za-z0-9_.-]+"


def _load_token_state():
    default_owner = os.environ.get("_HERMES_GITHUB_DEFAULT_OWNER", "").strip().lower()
    tokens = {}
    token_store_path = os.environ.get("_HERMES_GITHUB_TOKEN_MAP_PATH", "").strip()
    if token_store_path:
        try:
            with open(token_store_path, "r", encoding="utf-8") as handle:
                payload = json.load(handle)
            default_owner = str(payload.get("default_owner") or default_owner).strip().lower()
            raw_tokens = payload.get("tokens_by_owner") or {}
            tokens = {str(key).lower(): str(value) for key, value in raw_tokens.items() if value}
        except Exception:
            tokens = {}
    if not tokens:
        try:
            raw_tokens = json.loads(os.environ.get("_HERMES_GITHUB_TOKEN_MAP_JSON", "{}"))
        except json.JSONDecodeError:
            raw_tokens = {}
        tokens = {str(key).lower(): str(value) for key, value in raw_tokens.items() if value}
    return tokens, default_owner


def _known_owner(owner):
    tokens, _default_owner = _load_token_state()
    owner = str(owner or "").strip().lower()
    return owner if owner in tokens else ""


def _owner_from_text(value):
    value = str(value or "").strip().strip("'\"")
    if not value:
        return ""
    patterns = [
        rf"(?:^|[/:])repos/({OWNER_RE})/({REPO_RE})(?:$|[/?#:\s])",
        rf"github\.com[:/]({OWNER_RE})/({REPO_RE})(?:\.git)?(?:$|[/?#:\s])",
        rf"\brepo:({OWNER_RE})/({REPO_RE})(?:$|[,\s])",
        rf"^({OWNER_RE})/({REPO_RE})(?:\.git)?$",
    ]
    for pattern in patterns:
        match = re.search(pattern, value)
        if match:
            owner = _known_owner(match.group(1))
            if owner:
                return owner
    return ""


def _owner_from_gh_args(args):
    for index, arg in enumerate(args):
        if arg in {"--repo", "-R"} and index + 1 < len(args):
            owner = _owner_from_text(args[index + 1])
            if owner:
                return owner
        for prefix in ("--repo=", "-R="):
            if arg.startswith(prefix):
                owner = _owner_from_text(arg[len(prefix):])
                if owner:
                    return owner
    for arg in args:
        owner = _owner_from_text(arg)
        if owner:
            return owner
    return ""


def _token_for_owner(owner):
    tokens, default_owner = _load_token_state()
    owner = str(owner or "").strip().lower()
    token = tokens.get(owner) if owner else ""
    if token:
        return owner, token
    if default_owner and tokens.get(default_owner):
        return default_owner, tokens[default_owner]
    if tokens:
        first_owner = sorted(tokens)[0]
        return first_owner, tokens[first_owner]
    return "", ""


def _git_credential():
    if len(sys.argv) > 2 and sys.argv[2] not in {"get", ""}:
        raise SystemExit(0)
    query = {}
    for line in sys.stdin:
        line = line.rstrip("\n")
        if not line:
            break
        key, _, value = line.partition("=")
        query[key] = value
    if query.get("host") not in {"github.com", "www.github.com"}:
        raise SystemExit(0)
    owner = _owner_from_text(query.get("path", "").lstrip("/"))
    _owner, token = _token_for_owner(owner)
    if token:
        print("username=x-access-token")
        print(f"password={token}")


def _askpass():
    prompt = " ".join(sys.argv[2:])
    if "username" in prompt.lower():
        print("x-access-token")
        return
    owner = _owner_from_text(prompt)
    _owner, token = _token_for_owner(owner)
    if token:
        print(token)


def _find_real_binary(name):
    self_path = os.path.realpath(sys.argv[0])
    shim_dir = os.path.realpath(os.environ.get("_HERMES_GITHUB_SHIM_DIR", ""))
    for directory in os.environ.get("PATH", "").split(os.pathsep):
        if not directory:
            continue
        real_dir = os.path.realpath(directory)
        if shim_dir and real_dir == shim_dir:
            continue
        candidate = os.path.realpath(os.path.join(directory, name))
        if candidate == self_path:
            continue
        if os.path.isfile(candidate) and os.access(candidate, os.X_OK):
            return candidate
    return ""


def _exec_gh():
    args = sys.argv[2:]
    owner = _owner_from_gh_args(args)
    selected_owner, token = _token_for_owner(owner)
    if not token:
        print("rsi github credential broker: no GitHub token is available", file=sys.stderr)
        raise SystemExit(1)
    real_gh = _find_real_binary("gh")
    if not real_gh:
        print("rsi github credential broker: real gh binary not found on PATH", file=sys.stderr)
        raise SystemExit(127)
    env = dict(os.environ)
    env["GH_TOKEN"] = token
    env["GITHUB_TOKEN"] = token
    env["_HERMES_GITHUB_SELECTED_OWNER"] = selected_owner
    os.execve(real_gh, [real_gh, *args], env)


def _print_token():
    owner = ""
    args = sys.argv[2:]
    for index, arg in enumerate(args):
        if arg == "--owner" and index + 1 < len(args):
            owner = args[index + 1]
        elif arg.startswith("--owner="):
            owner = arg.split("=", 1)[1]
        elif "/" in arg:
            owner = _owner_from_text(arg)
    selected_owner, token = _token_for_owner(owner)
    if not token:
        raise SystemExit(1)
    if "--print-owner" in args:
        print(selected_owner)
    else:
        print(token)


def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else ""
    if mode == "git-credential":
        _git_credential()
    elif mode == "askpass":
        _askpass()
    elif mode == "exec-gh":
        _exec_gh()
    elif mode == "token":
        _print_token()
    else:
        print("usage: rsi-github-credential-broker.py {git-credential|askpass|exec-gh|token}", file=sys.stderr)
        raise SystemExit(2)


if __name__ == "__main__":
    main()
''',
            encoding="utf-8",
        )
        broker_path.chmod(0o700)
        return broker_path

    def _write_git_credential_helper(self, request_dir: Path, broker_path: Path | None = None) -> Path:
        helper_path = request_dir / "git-credential-rsi-github.py"
        if broker_path is not None:
            helper_path.write_text(
                "#!/bin/sh\n"
                f"exec {shlex.quote(str(broker_path))} git-credential \"$@\"\n",
                encoding="utf-8",
            )
            helper_path.chmod(0o700)
            return helper_path
        helper_path.write_text(
            "#!/usr/bin/env python3\n"
            "import json\n"
            "import os\n"
            "import sys\n"
            "\n"
            "if len(sys.argv) > 1 and sys.argv[1] not in {'get', ''}:\n"
            "    raise SystemExit(0)\n"
            "query = {}\n"
            "for line in sys.stdin:\n"
            "    line = line.rstrip('\\n')\n"
            "    if not line:\n"
            "        break\n"
            "    key, _, value = line.partition('=')\n"
            "    query[key] = value\n"
            "if query.get('host') not in {'github.com', 'www.github.com'}:\n"
            "    raise SystemExit(0)\n"
            "path = query.get('path', '').lstrip('/')\n"
            "owner = path.split('/', 1)[0].lower() if path else ''\n"
            "try:\n"
            "    raw_tokens = json.loads(os.environ.get('_HERMES_GITHUB_TOKEN_MAP_JSON', '{}'))\n"
            "except json.JSONDecodeError:\n"
            "    raw_tokens = {}\n"
            "tokens = {str(key).lower(): str(value) for key, value in raw_tokens.items() if value}\n"
            "default_owner = os.environ.get('_HERMES_GITHUB_DEFAULT_OWNER', '').lower()\n"
            "token = tokens.get(owner) or tokens.get(default_owner)\n"
            "if token:\n"
            "    print('username=x-access-token')\n"
            "    print(f'password={token}')\n",
            encoding="utf-8",
        )
        helper_path.chmod(0o700)
        return helper_path

    def _append_git_config_env(self, env: dict[str, str], entries: list[tuple[str, str]]) -> None:
        try:
            start = int(env.get("GIT_CONFIG_COUNT", "0") or "0")
        except ValueError:
            start = 0
        for offset, (key, value) in enumerate(entries):
            index = start + offset
            env[f"GIT_CONFIG_KEY_{index}"] = key
            env[f"GIT_CONFIG_VALUE_{index}"] = value
        env["GIT_CONFIG_COUNT"] = str(start + len(entries))

    def _write_git_askpass(self, request_dir: Path, broker_path: Path | None = None) -> Path:
        askpass_path = request_dir / "git-askpass.sh"
        if broker_path is not None:
            askpass_path.write_text(
                "#!/bin/sh\n"
                f"exec {shlex.quote(str(broker_path))} askpass \"$@\"\n",
                encoding="utf-8",
            )
            askpass_path.chmod(0o700)
            return askpass_path
        askpass_path.write_text(
            "#!/bin/sh\n"
            "case \"$1\" in\n"
            "  *Username*) printf '%s\\n' 'x-access-token' ;;\n"
            "  *) printf '%s\\n' \"$GH_TOKEN\" ;;\n"
            "esac\n",
            encoding="utf-8",
        )
        askpass_path.chmod(0o700)
        return askpass_path

    def _write_github_cli_wrapper(self, request_dir: Path, broker_path: Path) -> Path:
        shim_dir = request_dir / "github-bin"
        shim_dir.mkdir(parents=True, exist_ok=True)
        wrapper_path = shim_dir / "gh"
        wrapper_path.write_text(
            "#!/bin/sh\n"
            "_HERMES_GITHUB_SHIM_DIR=${_HERMES_GITHUB_SHIM_DIR:-$(CDPATH= cd -- \"$(dirname -- \"$0\")\" && pwd)}\n"
            "export _HERMES_GITHUB_SHIM_DIR\n"
            f"exec {shlex.quote(str(broker_path))} exec-gh \"$@\"\n",
            encoding="utf-8",
        )
        wrapper_path.chmod(0o700)
        return wrapper_path

    def _write_github_shell_init(self, request_dir: Path, broker_path: Path) -> Path:
        shell_init_path = request_dir / "github-shell-init.sh"
        shell_init_path.write_text(
            "# RSI GitHub capability bridge. Sourced by bash via BASH_ENV before Hermes captures its terminal snapshot.\n"
            "gh() {\n"
            f"  {shlex.quote(str(broker_path))} exec-gh \"$@\"\n"
            "}\n"
            "export -f gh 2>/dev/null || true\n",
            encoding="utf-8",
        )
        shell_init_path.chmod(0o600)
        return shell_init_path

    def _native_executor_completion_meta(self, parsed_result: JsonObject, max_iterations: int) -> JsonObject:
        result_payload = _json_object_or_empty(parsed_result.get("result"))
        explicit_termination_reason = first_non_empty(
            _string_or_json(parsed_result.get("termination_reason")),
            _string_or_json(result_payload.get("termination_reason")),
        )
        explicit_completion_verdict = first_non_empty(
            _string_or_json(parsed_result.get("completion_verdict")),
            _string_or_json(result_payload.get("completion_verdict")),
        )
        timeout_kind = first_non_empty(
            _string_or_json(parsed_result.get("timeout_kind")),
            _string_or_json(result_payload.get("timeout_kind")),
        )
        completed_value = result_payload.get("completed")
        completed_known = isinstance(completed_value, bool)
        completed = completed_value if completed_known else True
        partial = _bool_or_false(result_payload.get("partial")) or _bool_or_false(parsed_result.get("native_result_partial"))
        interrupted = _bool_or_false(result_payload.get("interrupted")) or _bool_or_false(parsed_result.get("native_result_interrupted"))
        api_calls_value = result_payload.get("api_calls")
        api_calls = int(api_calls_value) if isinstance(api_calls_value, Number) and not isinstance(api_calls_value, bool) else 0
        incomplete_without_reason = partial or interrupted or (completed_known and not completed)
        max_iterations_reached = (
            explicit_termination_reason == "iteration_budget_exhausted"
            or (max_iterations > 0 and api_calls >= max_iterations and not completed)
            or _bool_or_false(parsed_result.get("max_iterations_reached"))
        )
        termination_reason = explicit_termination_reason or (
            "iteration_budget_exhausted" if (max_iterations_reached or incomplete_without_reason) else "normal_completion"
        )
        partial_stop = termination_reason in PARTIAL_COMPLETION_TERMINATION_REASONS
        completion_verdict = explicit_completion_verdict or ("partial" if partial_stop or incomplete_without_reason else "complete")
        meta: JsonObject = {
            "termination_reason": termination_reason,
            "completion_verdict": completion_verdict,
            "max_iterations_reached": max_iterations_reached,
            "native_result_completed": completed,
            "native_result_partial": partial or partial_stop or completion_verdict == "partial",
            "native_result_interrupted": interrupted,
            "native_result_api_calls": api_calls,
        }
        if timeout_kind:
            meta["timeout_kind"] = timeout_kind
        return meta

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
        github_cli_credentials: JsonObject,
    ) -> JsonObject:
        profile = self._model_profile_for_task(task)
        return {
            "session_id": context.session_id,
            "parent_session_id": context.parent_session_id,
            "execution_id": task.execution_id or "",
            "operation_id": task.operation_id or "",
            "trace_id": task.trace_id or "",
            "workflow_id": task.workflow_id or "",
            "user_peer_id": context.user_peer_id,
            "assistant_peer_id": context.assistant_peer_id,
            "memory_backend": context.memory_backend,
            "honcho_workspace": self._config.honcho_workspace,
            "honcho_environment": self._config.honcho_environment,
            "chat_id": task.channel_id or "",
            "thread_id": task.thread_ts or "",
            "message_ts": task.message_ts or _derive_root_message_ts(task),
            "gateway_session_key": canonical_gateway_session_key(task, self._role),
            "conversation_history": self._native_executor_conversation_history(task, context),
            "external_tool_resume": task.external_tool_resume,
            "prompt": task.prompt,
            "system_message": task.system_message or "",
            "execution_phase": self._execution_phase(task),
            "toolsets": list(toolsets),
            "phase_contract": phase_contract,
            "required_final_tool_names": self._native_required_final_tool_names(task),
            "required_final_tool_max_attempts": 2,
            "required_final_tool_instruction": self._native_required_final_tool_instruction(task),
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
            "model": profile.provider_model,
            "max_iterations": max_iterations,
            "reasoning_config": _reasoning_config_for_profile(profile),
            "request_overrides": {},
            "workdir": str(workdir),
            "result_path": str(result_path),
            "hermes_home": self._config.hermes_home,
            "artifact_output_dir": str(self._native_artifact_destination(task)),
            "hermes_computer_root": self._config.hermes_computer_root,
            "hermes_run_root": self._config.hermes_run_root,
            "hermes_artifact_root": self._config.hermes_artifact_root,
            "hermes_native_terminal_enabled": self._config.hermes_native_terminal_enabled,
            "hermes_native_toolsets": self._hermes_native_toolsets(),
            "hermes_terminal_cwd": self._config.hermes_terminal_cwd,
            "hermes_company_bin_dir": self._config.hermes_company_bin_dir,
            "hermes_kubernetes_context_enabled": self._config.hermes_kubernetes_context_enabled,
            "hermes_kubeconfig_path": self._config.hermes_kubeconfig_path if (self._config.hermes_kubernetes_context_enabled or self._config.hermes_prod_kubernetes_context_enabled) else "",
            "github_cli_credentials": dict(github_cli_credentials),
            "company_computer_bootstrap_status": dict(self._company_computer_bootstrap_status),
            "run_dir": str(result_path.parent),
            "contract_version": task.contract_version or EXECUTION_CONTRACT_VERSION,
            "execution_intent": task.execution_intent,
            "delivery_policy": task.delivery_policy,
            "workspace_policy": task.workspace_policy,
            "approval_policy": task.approval_policy,
            "runtime": {
                "provider": profile.provider,
                "base_url": profile.base_url,
                "api_mode": self._api_mode,
                "provider_routing": dict(profile.openrouter_provider_routing) if profile.provider == "openrouter" else {},
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
        activity_callback: Any | None = None,
    ) -> None:
        chunk_index = 0
        marker_tail = ""
        marker_window = max(0, len(_NATIVE_EXECUTOR_RESULT_MARKER) - 1)
        try:
            while True:
                chunk = stream.read(_NATIVE_EXECUTOR_OUTPUT_CHUNK_CHARS)
                if not chunk:
                    break
                if activity_callback is not None:
                    activity_callback(f"{stream_name} output")
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
                        event_type="terminal.output",
                        status="streaming",
                        payload={
                            "legacy_event_type": "executor.subprocess.output",
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

    def _native_lifecycle_path(self, session_id: str) -> Path:
        return Path(self._config.hermes_home).expanduser() / "rsi_runtime" / "lifecycle" / f"{session_id}.jsonl"

    def _emit_native_lifecycle_event(
        self,
        observer: ObservationEmitter,
        phase: str,
        item: JsonObject,
        *,
        secret_values: list[str],
    ) -> None:
        event_type = _string_or_json(item.get("event_type")) or _string_or_json(item.get("event"))
        if not event_type:
            return
        payload = _json_object_or_empty(item.get("payload"))
        if not payload:
            payload = {
                key: value
                for key, value in item.items()
                if key not in {"event", "event_type", "status", "recorded_at", "recorded_at_unix", "session_id"}
            }
        redacted_payload = _json_object_or_empty(
            _redact_json_value(
                payload,
                secret_values=secret_values,
                limit=self._config.verbose_trace_log_limit,
            )
        )
        recorded_at_unix = item.get("recorded_at_unix")
        if recorded_at_unix is not None:
            redacted_payload["recorded_at_unix"] = recorded_at_unix
        session_id = _string_or_json(item.get("session_id"))
        if session_id:
            redacted_payload["hermes_session_id"] = session_id
        redacted_payload["source"] = "hermes_lifecycle"
        observer.emit(
            phase=phase,
            event_type=event_type,
            status=_string_or_json(item.get("status")),
            payload=redacted_payload,
        )

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
            _atomic_write_json(path, stored)
        except Exception as exc:
            logger.warning("executor status persist failed execution_id=%s path=%s error=%s", key, path, exc)

    def _attach_executor_status_heartbeat(self, observer: ObservationEmitter, task: RunnerTaskRequest) -> None:
        observer.on_emit = lambda item: self._persist_executor_observation_heartbeat(observer.execution_id, task, item)

    def _persist_executor_observation_heartbeat(self, execution_id: str, task: RunnerTaskRequest, item: JsonObject) -> None:
        key = str(execution_id or "").strip()
        if not key:
            return
        event_type = _string_or_json(item.get("event_type"))
        now = time.time()
        force = not event_type.startswith("model.reasoning.delta") and not event_type.startswith("model.output.delta")
        last_heartbeat_at = self._executor_status_heartbeat_at.get(key, 0.0)
        if not force and now - last_heartbeat_at < 2.0:
            return
        self._executor_status_heartbeat_at[key] = now
        existing = dict(self._executor_recent_results.get(key) or {})
        if not existing:
            path = self._executor_status_path(key)
            try:
                if path.exists():
                    loaded = json.loads(path.read_text(encoding="utf-8"))
                    if isinstance(loaded, dict):
                        existing = loaded
            except Exception as exc:
                logger.debug("executor heartbeat status read failed execution_id=%s error=%s", key, exc)
        existing_status = str(existing.get("status") or "").strip().lower()
        if existing_status in {"completed", "failed", "cancelled"}:
            return
        next_status = existing_status if existing_status in {"accepted", "finalizing", "cancelling", "cancel_requested"} else "running"
        last_observed_seq = item.get("seq")
        heartbeat: JsonObject = {
            **existing,
            "execution_id": key,
            "operation_id": first_non_empty(_string_or_json(existing.get("operation_id")), task.operation_id),
            "trace_id": first_non_empty(_string_or_json(existing.get("trace_id")), task.trace_id),
            "workflow_id": first_non_empty(_string_or_json(existing.get("workflow_id")), task.workflow_id),
            "executor_instance_id": first_non_empty(
                _string_or_json(existing.get("executor_instance_id")),
                self._config.executor_instance_id,
            ),
            "executor_started_at_unix": self._started_at_unix,
            "phase": first_non_empty(_string_or_json(item.get("phase")), self._execution_phase(task)),
            "status": next_status,
            "message": "Execution active; last observation persisted.",
            "last_observed_ledger_seq": "" if last_observed_seq in (None, "") else str(last_observed_seq),
            "last_observed_event_type": event_type,
            "last_observed_event_status": _string_or_json(item.get("status")),
            "last_observed_phase": _string_or_json(item.get("phase")),
            "last_observed_recorded_at": _string_or_json(item.get("recorded_at")),
            "last_observed_at_unix": time.time(),
            "last_observed_invocation_id": _string_or_json(item.get("invocation_id")),
        }
        self._store_executor_result(key, heartbeat)

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
            "executor_instance_id": self._config.executor_instance_id,
            "phase": self._execution_phase(task),
            "status": status or ("completed" if result.ok else "failed"),
            "workspace_root": workspace_root or self._config.hermes_computer_root,
            "session_id": session_id,
            "termination_reason": _string_or_json(raw.get("termination_reason")),
            "completion_verdict": _string_or_json(raw.get("completion_verdict")),
            "message": result.message,
            "result": self._executor_result_payload(result),
        }

    def _finalize_self_review_candidate(
        self,
        task: RunnerTaskRequest,
        result: HermesExecutionResult,
        final_status: JsonObject,
        *,
        execution_id: str,
    ) -> None:
        raw = _json_object_or_empty(result.raw)
        candidate = _json_object_or_empty(raw.get("self_review_candidate"))
        candidate_id = int(candidate.get("candidate_id") or 0)
        if candidate_id <= 0:
            return
        diagnostics = _json_object_or_empty(raw.get("runner_diagnostics"))
        native_validated = _bool_or_false(diagnostics.get("native_envelope_validated"))
        partial_recovered = _bool_or_false(raw.get("partial_recovery_succeeded")) or _bool_or_false(diagnostics.get("partial_completion_succeeded"))
        eligible = (
            bool(result.ok)
            and str(final_status.get("status") or "").strip().lower() == "completed"
            and self._result_is_native_strict(result)
            and not _bool_or_false(raw.get("non_deliverable"))
            and (native_validated or partial_recovered)
        )
        try:
            if not eligible:
                from self_review_queue import mark_candidate_ineligible  # type: ignore

                reason = "result_not_completed"
                if not result.ok:
                    reason = "worker_error"
                elif str(final_status.get("status") or "").strip().lower() != "completed":
                    reason = "result_not_completed"
                elif not self._result_is_native_strict(result):
                    reason = "non_native_strict_fallback"
                elif _bool_or_false(raw.get("non_deliverable")):
                    reason = "non_deliverable_result"
                elif not (native_validated or partial_recovered):
                    reason = "envelope_validation_failed"
                mark_candidate_ineligible(self._self_review_config(), execution_id, reason)
                self._persist_self_review_status(
                    execution_id,
                    final_status,
                    self._self_review_promotion_status(
                        candidate=candidate,
                        fallback_status="ineligible",
                        last_error=reason,
                    ),
                    result,
                )
                return
            from self_review_queue import mark_candidate_delivered, promote_review_candidate  # type: ignore

            result_payload = _json_object_or_empty(final_status.get("result"))
            result_hash = hashlib.sha256(
                json.dumps(result_payload, ensure_ascii=True, sort_keys=True, separators=(",", ":")).encode("utf-8")
            ).hexdigest()
            delivered = mark_candidate_delivered(
                self._self_review_config(),
                execution_id,
                result_hash,
                result_ref=str(self._executor_status_path(execution_id)),
            )
            delivered_status = str(delivered.get("status") or "").strip().lower()
            if delivered_status != "validated":
                logger.warning("self-review delivered marker rejected execution_id=%s status=%s", execution_id, delivered.get("status"))
                self._persist_self_review_status(
                    execution_id,
                    final_status,
                    self._self_review_promotion_status(
                        candidate=candidate,
                        delivered=delivered,
                        fallback_status=delivered_status or "delivery_marker_rejected",
                        last_error=_string_or_json(delivered.get("error")) or "delivery marker rejected",
                    ),
                    result,
                )
                return
            promoted = promote_review_candidate(self._self_review_config(), candidate_id)
            promoted_status = str(promoted.get("status") or "").strip().lower()
            try:
                from self_review_queue import candidate_status  # type: ignore

                status_payload = candidate_status(self._self_review_config(), execution_id)
            except Exception:
                status_payload = {}
            self._persist_self_review_status(
                execution_id,
                final_status,
                self._self_review_promotion_status(
                    candidate=candidate,
                    delivered=delivered,
                    promoted=promoted,
                    candidate_status_payload=status_payload if isinstance(status_payload, dict) else {},
                    fallback_status=promoted_status,
                ),
                result,
            )
            if promoted_status == "enqueued":
                self._start_self_review_worker(candidate_id)
            elif promoted_status in {"skipped", "blocked_by_earlier_candidate"}:
                logger.info(
                    "self-review candidate promotion did not enqueue execution_id=%s candidate_id=%s status=%s",
                    execution_id,
                    candidate_id,
                    promoted.get("status"),
                )
            else:
                logger.warning(
                    "self-review candidate promotion failed to enqueue execution_id=%s candidate_id=%s status=%s",
                    execution_id,
                    candidate_id,
                    promoted.get("status"),
                )
        except Exception as exc:
            logger.warning("self-review candidate finalization failed execution_id=%s candidate_id=%s error=%s", execution_id, candidate_id, exc)

    def _start_self_review_worker(self, candidate_id: int) -> None:
        key = f"candidate-{candidate_id}"
        with self._executor_process_lock:
            existing = self._self_review_processes.get(key)
            if existing is not None and existing.poll() is None:
                return
            cmd = [
                sys.executable,
                "-m",
                "rsi_runner.hermes_self_review_worker",
                "--candidate-id",
                str(candidate_id),
            ]
            if self._self_review_draining:
                cmd.append("--local-only")
            process = subprocess.Popen(
                cmd,
                stdin=subprocess.DEVNULL,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                env=self._self_review_worker_env(),
                text=True,
            )
            self._self_review_processes[key] = process

        def _reap() -> None:
            returncode = process.wait()
            with self._executor_process_lock:
                if self._self_review_processes.get(key) is process:
                    self._self_review_processes.pop(key, None)
            if returncode != 0:
                logger.warning("self-review worker exited candidate_id=%s returncode=%s", candidate_id, returncode)

        threading.Thread(target=_reap, name=f"rsi-hermes-self-review-{candidate_id}", daemon=True).start()

    def _native_strict_failure(
        self,
        task: RunnerTaskRequest,
        *,
        failure_class: str,
        message: str,
        diagnostics: JsonObject | None = None,
    ) -> HermesExecutionResult:
        runner_diagnostics = self._runner_diagnostics(
            failure_kind=failure_class,
            provider_error_message=message,
            termination_reason=failure_class,
            session_ready_issues=self._session_manager.ready_issues,
            repair_attempted=False,
            repair_succeeded=False,
            observed={
                "failure_kind": failure_class,
                "termination_reason": failure_class,
                **_json_object_or_empty(diagnostics),
            },
        )
        return HermesExecutionResult(
            ok=False,
            message=message,
            provider="hermes-native-executor",
            raw={
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                "native_strict": True,
                "failure_class": failure_class,
                "runner_diagnostics": runner_diagnostics,
                "termination_reason": failure_class,
            },
        )

    def _native_cancelled_result(self, task: RunnerTaskRequest, result: HermesExecutionResult) -> HermesExecutionResult:
        raw = dict(_json_object_or_empty(result.raw))
        raw["native_strict"] = True
        raw["non_deliverable"] = True
        raw["failure_class"] = "runner_execution_cancelled"
        raw["termination_reason"] = "runner_execution_cancelled"
        diagnostics = dict(_json_object_or_empty(raw.get("runner_diagnostics")))
        diagnostics["failure_kind"] = "runner_execution_cancelled"
        diagnostics["termination_reason"] = "runner_execution_cancelled"
        diagnostics["non_deliverable"] = True
        raw["runner_diagnostics"] = diagnostics
        return HermesExecutionResult(
            ok=False,
            message="Execution completed after cancellation was requested and is not deliverable.",
            provider=result.provider or "hermes-native-executor",
            raw=raw,
        )

    def _validate_native_workflow_ids(self, task: RunnerTaskRequest) -> list[str]:
        missing: list[str] = []
        for key in ("execution_id", "operation_id", "trace_id", "workflow_id"):
            if not _string_or_json(getattr(task, key, "")):
                missing.append(key)
        return missing

    def _load_native_execution_envelope(
        self,
        task: RunnerTaskRequest,
        *,
        envelope_path: Path,
        context_path: Path,
        wait_seconds: float = _NATIVE_ENVELOPE_WAIT_SECONDS,
    ) -> tuple[JsonObject | None, str, JsonObject]:
        if envelope_path.exists():
            waited_ms = 0
        elif self._contract_status.plugin_status != "ok":
            waited_ms = 0
        else:
            deadline = time.monotonic() + max(0.0, wait_seconds)
            while not envelope_path.exists() and time.monotonic() < deadline:
                time.sleep(0.05)
            waited_ms = int(max(0.0, wait_seconds - max(0.0, deadline - time.monotonic())) * 1000)
        diagnostics: JsonObject = {
            "expected_envelope_path": str(envelope_path),
            "context_path": str(context_path),
            "waited_ms": waited_ms,
            "execution_id": task.execution_id or "",
            "operation_id": task.operation_id or "",
            "trace_id": task.trace_id or "",
            "workflow_id": task.workflow_id or "",
            "plugin_status": self._contract_status.plugin_status,
        }
        if not envelope_path.exists():
            return None, "plugin_execution_envelope_missing", diagnostics
        try:
            parsed = json.loads(envelope_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            diagnostics["error"] = str(exc)
            return None, "plugin_execution_envelope_invalid", diagnostics
        except OSError as exc:
            diagnostics["error"] = str(exc)
            return None, "plugin_execution_envelope_missing", diagnostics
        if not isinstance(parsed, dict):
            diagnostics["error"] = "envelope payload must be a JSON object"
            return None, "plugin_execution_envelope_invalid", diagnostics
        missing_fields = [field for field in _NATIVE_ENVELOPE_REQUIRED_FIELDS if field not in parsed]
        if missing_fields:
            diagnostics["missing_fields"] = missing_fields
            return None, "plugin_execution_envelope_invalid", diagnostics
        for key in ("execution_id", "operation_id", "trace_id", "workflow_id"):
            expected = _string_or_json(getattr(task, key, ""))
            actual = _string_or_json(parsed.get(key))
            if not actual:
                diagnostics["missing_identifier"] = key
                return None, "plugin_execution_envelope_invalid", diagnostics
            if actual != expected:
                diagnostics["mismatched_identifier"] = key
                diagnostics["expected"] = expected
                diagnostics["actual"] = actual
                return None, "plugin_execution_envelope_mismatch", diagnostics
        if _string_or_json(parsed.get("producer")) != "rsi_platform_runtime":
            diagnostics["producer"] = _string_or_json(parsed.get("producer"))
            return None, "plugin_execution_envelope_invalid", diagnostics
        if not isinstance(parsed.get("phase_runs"), list) or not isinstance(parsed.get("ledger_events"), list) or not isinstance(parsed.get("artifacts"), list) or not isinstance(parsed.get("deliveries"), list):
            diagnostics["error"] = "phase_runs, ledger_events, artifacts, and deliveries must be arrays"
            return None, "plugin_execution_envelope_invalid", diagnostics
        completion = _json_object_or_empty(parsed.get("completion"))
        if not completion:
            diagnostics["error"] = "completion must be an object"
            return None, "plugin_execution_envelope_invalid", diagnostics
        if not _bool_or_false(completion.get("ok")):
            diagnostics["completion"] = completion
            return None, "plugin_execution_envelope_invalid", diagnostics
        if not isinstance(parsed.get("final_response"), str) or not parsed.get("final_response", "").strip():
            diagnostics["error"] = "final_response must be a non-empty string for successful native workflow completion"
            return None, "plugin_execution_envelope_invalid", diagnostics
        return parsed, "", diagnostics

    def _execute_native_workflow_task_request(
        self,
        task: RunnerTaskRequest,
        *,
        observer: ObservationEmitter | None = None,
        allow_partial_recovery: bool = True,
        max_iterations_override: int | None = None,
    ) -> HermesExecutionResult:
        try:
            task_model_profile = self._model_profile_for_task(task)
        except ValueError as exc:
            return self._native_strict_failure(
                task,
                failure_class="invalid_model_override",
                message=str(exc),
            )
        if not _runtime_api_key_for_provider(task_model_profile.provider):
            return self._native_strict_failure(
                task,
                failure_class="native_workflow_preflight_failed",
                message=self._credentials_error_message_for_profile(task_model_profile),
            )
        if not self._contract_status.ok:
            return self._native_strict_failure(
                task,
                failure_class="native_envelope_plugin_unavailable",
                message="Hermes native runtime contract failed: " + "; ".join(self._contract_status.errors),
                diagnostics={
                    "contract_errors": list(self._contract_status.errors),
                    "contract_status": self._contract_status.to_dict(),
                },
            )
        missing_ids = self._validate_native_workflow_ids(task)
        if missing_ids:
            return self._native_strict_failure(
                task,
                failure_class="native_workflow_preflight_failed",
                message="Native Hermes workflow execution requires non-empty identifier(s): " + ", ".join(missing_ids),
                diagnostics={"missing_identifiers": missing_ids},
            )
        if not bool(self._company_computer_bootstrap_status.get("ok")):
            errors = _string_list_or_empty(self._company_computer_bootstrap_status.get("errors"))
            message = "Hermes company-computer bootstrap failed: " + "; ".join(errors)
            return self._native_strict_failure(
                task,
                failure_class="native_workflow_preflight_failed",
                message=message,
                diagnostics={"bootstrap_errors": errors},
            )
        if not self._session_manager.available:
            return self._native_strict_failure(
                task,
                failure_class="native_workflow_preflight_failed",
                message="Hermes persistent session runtime is unavailable.",
            )
        context = self._session_manager.prepare(task, load_history=False)
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

        execution_id = task.execution_id or ""
        context_path = self._native_runtime_context_path(execution_id)
        envelope_path = self._native_runtime_envelope_path(execution_id)
        workspace_root = self._stage_native_executor_workspace(task, observer=observer)
        if execution_phase in {"main", "render", "deliver"}:
            task = replace(task, artifact_destination=str(self._native_artifact_destination(task)))
        self._stage_task_context(context.session_id, task, context_path=context_path, envelope_path=envelope_path)

        agentic_mcp_registration = TaskScopedMCPRegistration()
        legacy_mcp_errors = self._legacy_mcp_server_errors(task)
        if legacy_mcp_errors:
            message = "Legacy Slack/Notion MCP servers are disabled for RSI workflows: " + "; ".join(legacy_mcp_errors)
            if observer is not None:
                observer.emit(
                    phase=execution_phase,
                    event_type="mcp.registration",
                    status="failed",
                    payload={"error": message},
                )
            return self._native_strict_failure(
                task,
                failure_class="native_workflow_preflight_failed",
                message=message,
                diagnostics={"preflight_failure_class": "legacy_mcp_disabled", "legacy_mcp_errors": legacy_mcp_errors},
            )
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
            return self._native_strict_failure(
                task,
                failure_class="native_workflow_preflight_failed",
                message=str(exc),
                diagnostics={"preflight_failure_class": "agentic_mcp_registration_failed"},
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

        configured_max_iterations = max_iterations_override if max_iterations_override and max_iterations_override > 0 else self._max_iterations
        toolsets = self._native_toolsets_for_task(task)
        toolsets = normalize_tool_names([*toolsets, *agentic_mcp_registration.enabled_toolsets])
        phase_contract = self._native_executor_phase_contract(task, toolsets)
        missing_phase_toolsets = _string_list_or_empty(phase_contract.get("missing_required_toolsets"))
        if missing_phase_toolsets:
            message = "Native Hermes phase contract failed; missing required toolset(s): " + ", ".join(missing_phase_toolsets)
            return self._native_strict_failure(
                task,
                failure_class="native_workflow_preflight_failed",
                message=message,
                diagnostics={"phase_contract": phase_contract, "missing_required_toolsets": missing_phase_toolsets},
            )
        request_dir = self._native_run_dir(task)
        github_cli_env, github_cli_credentials = self._github_cli_environment(task)
        if github_cli_env:
            gh_config_dir = request_dir / "gh-config"
            gh_config_dir.mkdir(parents=True, exist_ok=True)
            github_cli_env["GH_CONFIG_DIR"] = str(gh_config_dir)
            github_cli_credentials["gh_config_dir"] = str(gh_config_dir)
            token_store_path = self._write_github_token_store(request_dir, github_cli_env)
            github_cli_env["_HERMES_GITHUB_TOKEN_MAP_PATH"] = str(token_store_path)
            github_cli_env.pop("_HERMES_GITHUB_TOKEN_MAP_JSON", None)
            github_cli_credentials["github_token_store_configured"] = True
            github_cli_credentials["github_token_store_path"] = str(token_store_path)
            credential_broker_path = self._write_github_credential_broker(request_dir)
            github_cli_credentials["github_credential_broker_configured"] = True
            github_cli_credentials["github_credential_broker_path"] = str(credential_broker_path)
            credential_helper_path = self._write_git_credential_helper(request_dir, credential_broker_path)
            self._append_git_config_env(
                github_cli_env,
                [
                    ("credential.helper", f"!{shlex.quote(str(credential_helper_path))}"),
                    ("credential.useHttpPath", "true"),
                ],
            )
            github_cli_credentials["git_credential_helper_configured"] = True
            github_cli_credentials["git_credential_helper_path"] = str(credential_helper_path)
            askpass_path = self._write_git_askpass(request_dir, credential_broker_path)
            github_cli_env["GIT_ASKPASS"] = str(askpass_path)
            github_cli_env["GIT_TERMINAL_PROMPT"] = "0"
            github_cli_credentials["git_askpass_configured"] = True
            gh_wrapper_path = self._write_github_cli_wrapper(request_dir, credential_broker_path)
            gh_shim_dir = str(gh_wrapper_path.parent)
            existing_path = github_cli_env.get("PATH") or os.environ.get("PATH", "")
            github_cli_env["PATH"] = f"{gh_shim_dir}{os.pathsep}{existing_path}" if existing_path else gh_shim_dir
            github_cli_env["_HERMES_GITHUB_SHIM_DIR"] = gh_shim_dir
            github_cli_credentials["github_cli_wrapper_configured"] = True
            github_cli_credentials["github_cli_wrapper_path"] = str(gh_wrapper_path)
            shell_init_path = self._write_github_shell_init(request_dir, credential_broker_path)
            github_cli_env["BASH_ENV"] = str(shell_init_path)
            github_cli_credentials["github_shell_init_configured"] = True
            github_cli_credentials["github_shell_init_path"] = str(shell_init_path)
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="github.credentials",
                status=_string_or_json(github_cli_credentials.get("status")) or "disabled",
                payload=github_cli_credentials,
            )
        if _string_or_json(github_cli_credentials.get("status")) == "failed":
            reason = _string_or_json(github_cli_credentials.get("reason")) or "unknown"
            message = f"Native Hermes GitHub App credentials unavailable: {reason}"
            return self._native_strict_failure(
                task,
                failure_class="native_workflow_preflight_failed",
                message=message,
                diagnostics={"preflight_failure_class": "github_app_credentials_unavailable", "github_credentials": github_cli_credentials},
            )
        prompt_skills = self._prompt_skill_mentions(task)
        skill_diagnostics: JsonObject = {
            "prompt_skills": list(prompt_skills),
            "resolved_skills": list(prompt_skills),
            "missing_skills": [],
            "skill_injection_mode": "native_preload" if prompt_skills else "none",
        }
        if observer is not None:
            observer.emit(
                phase=execution_phase,
                event_type="skills.expanded",
                status="completed",
                payload=skill_diagnostics,
            )

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
            github_cli_credentials=github_cli_credentials,
        )
        request_payload["runtime_context_path"] = str(context_path)
        request_payload["runtime_envelope_path"] = str(envelope_path)
        request_file.write_text(json.dumps(request_payload, indent=2, sort_keys=True), encoding="utf-8")
        worker_cmd = [sys.executable, "-m", "rsi_runner.hermes_executor_worker", str(request_file)]
        env_copy = os.environ.copy()
        if self._config.db_read_gateway_configured:
            env_copy.pop("RSI_DB_READ_CLIENT_TOKEN", None)
        env_copy.update(github_cli_env)
        env_copy.update(self._native_worker_session_env(task))
        env_copy.update(self._native_runtime_env(task, context_path=context_path, envelope_path=envelope_path))
        self._strip_native_worker_source_credentials(env_copy)
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
                    "timeout_source": "hermes_subprocess",
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
        cancelled = False

        lifecycle_tailer: _NativeLifecycleTailer | None = None
        if observer is not None:
            lifecycle_tailer = _NativeLifecycleTailer(
                path=self._native_lifecycle_path(context.session_id),
                phase=execution_phase,
                observer=observer,
                secret_values=secret_values,
                emit_event=self._emit_native_lifecycle_event,
                start_at_end=True,
            )
        process = subprocess.Popen(
            worker_cmd,
            cwd=str(workspace_root),
            env=env_copy,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        with self._executor_process_lock:
            self._executor_native_session_ids[execution_id] = context.session_id
            self._executor_native_session_keys[execution_id] = canonical_gateway_session_key(task, self._role)
            self._executor_native_started_at_unix[execution_id] = time.time()
        if lifecycle_tailer is not None:
            lifecycle_tailer.start()
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
        try:
            while True:
                with self._executor_process_lock:
                    self._refresh_native_descendant_tracking_locked()
                    if execution_id in self._executor_cancel_requests:
                        cancelled = True
                        break
                if process.poll() is not None:
                    break
                time.sleep(0.25)
            if cancelled:
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
            if lifecycle_tailer is not None:
                lifecycle_tailer.stop()
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
        existing_status = str(self.executor_status(execution_id).get("status") or "").strip().lower()
        status_to_store = "cancelling" if existing_status in {"cancel_requested", "cancelling"} else "finalizing"
        self._store_executor_result(
            execution_id,
            {
                "execution_id": execution_id,
                "status": status_to_store,
                "operation_id": task.operation_id,
                "trace_id": task.trace_id,
                "workflow_id": task.workflow_id,
                "executor_instance_id": self._config.executor_instance_id,
                "phase": self._execution_phase(task),
                "workspace_root": str(workspace_root),
                "session_id": context.session_id,
                "message": "Hermes subprocess completed; validating native execution envelope.",
            },
        )

        if cancelled and not parsed_result:
            finalized = self._session_manager.finalize(context, tracker)
            lifecycle_events = self._adapter.lifecycle_events(context.session_id)
            observed = self._observability_metadata(
                None,
                task,
                tracker,
                skill_diagnostics=skill_diagnostics,
                observer=observer,
                lifecycle_events=lifecycle_events,
            )
            cleanup_status = "not_needed" if not agentic_mcp_registration.enabled else "worker_unavailable"
            cleanup_errors: list[str] = []
            if observer is not None:
                observer.emit(
                    phase=execution_phase,
                    event_type="model.call.completed",
                    status="cancelled",
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
            termination_reason = "cancelled"
            stop_meta: JsonObject = {
                "termination_reason": termination_reason,
            }
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
            result = HermesExecutionResult(
                ok=False,
                message="Hermes native executor was cancelled.",
                provider="hermes-native-executor",
                raw={
                    **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                    **finalized,
                    **observed,
                    **self._workflow_evidence_raw(task, observed, termination_reason),
                    **stop_meta,
                    "failure_class": "runner_cancelled",
                    "native_strict": True,
                    "native_timeout_source": "hermes_subprocess",
                    "runner_diagnostics": self._runner_diagnostics(
                        failure_kind="cancelled",
                        provider_error_message="Hermes native executor was cancelled.",
                        termination_reason=termination_reason,
                        session_ready_issues=self._session_manager.ready_issues,
                        repair_attempted=False,
                        repair_succeeded=False,
                        observed=observed,
                    ),
                    "lifecycle_events": lifecycle_events,
                    "termination_reason": termination_reason,
                    "native_executor_mode": "subprocess",
                    "github_cli_credentials": github_cli_credentials,
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
                    "executor_instance_id": self._config.executor_instance_id,
                    "phase": self._execution_phase(task),
                    "workspace_root": str(workspace_root),
                    "session_id": context.session_id,
                    "termination_reason": termination_reason,
                    "message": result.message,
                    "result": self._executor_result_payload(result),
                },
            )
            return result
        finalized = self._session_manager.finalize(context, tracker)
        lifecycle_events = self._adapter.lifecycle_events(context.session_id)
        observed = self._observability_metadata(
            None,
            task,
            tracker,
            skill_diagnostics=skill_diagnostics,
            observer=observer,
            lifecycle_events=lifecycle_events,
        )
        parsed_result_loaded = bool(parsed_result) and not parse_error
        completion_meta = self._native_executor_completion_meta(parsed_result, configured_max_iterations) if parsed_result_loaded else {
            "termination_reason": "normal_completion" if completed_returncode == 0 else "exception",
            "completion_verdict": "complete" if completed_returncode == 0 else "",
            "max_iterations_reached": False,
            "native_result_completed": completed_returncode == 0,
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
            "max_iterations": configured_max_iterations,
            "lifecycle_events": lifecycle_events,
            "native_executor_mode": "subprocess",
            "native_strict": True,
            "native_timeout_source": "hermes_subprocess",
            "native_executor_workspace_root": str(workspace_root),
            "native_executor_toolsets": toolsets,
            "native_executor_returncode": completed_returncode,
            "native_executor_stderr": _truncate_string(redacted_completed_stderr, 4000),
            "native_executor_contract_status": _json_object_or_empty(parsed_result.get("contract_status")),
            "self_review_candidate": _json_object_or_empty(parsed_result.get("self_review_candidate")),
            "github_cli_credentials": github_cli_credentials,
            "artifact_tool_events": artifact_tool_events,
            "native_artifact_paths": native_artifact_paths,
            "phase_contract": phase_contract,
            **artifact_event_details,
        }
        if (
            parsed_result_loaded
            and not _bool_or_false(parsed_result.get("ok"))
            and _string_or_json(completion_meta.get("termination_reason")) in PARTIAL_COMPLETION_TERMINATION_REASONS
            and allow_partial_recovery
            and self._workflow_partial_completion_eligible(task)
        ):
            termination_reason = _string_or_json(completion_meta.get("termination_reason"))
            result_payload = _json_object_or_empty(parsed_result.get("result"))
            last_activity = _json_object_or_empty(parsed_result.get("last_activity"))
            if not last_activity:
                last_activity = _json_object_or_empty(result_payload.get("last_activity"))
            stop_meta: JsonObject = {
                "termination_reason": termination_reason,
                "last_activity": last_activity,
                "max_iterations_reached": _bool_or_false(completion_meta.get("max_iterations_reached")),
            }
            for key in ("timeout_kind", "stopped_after_seconds", "task_timeout_seconds", "inactivity_timeout_seconds"):
                value = parsed_result.get(key)
                if value in (None, ""):
                    value = result_payload.get(key)
                if value in (None, ""):
                    value = completion_meta.get(key)
                if value not in (None, ""):
                    stop_meta[key] = value
            if "timeout_kind" not in stop_meta and termination_reason in {"task_timeout", "inactivity_timeout"}:
                stop_meta["timeout_kind"] = termination_reason
            self._store_executor_result(
                execution_id,
                {
                    "execution_id": execution_id,
                    "status": "finalizing",
                    "operation_id": task.operation_id,
                    "trace_id": task.trace_id,
                    "workflow_id": task.workflow_id,
                    "executor_instance_id": self._config.executor_instance_id,
                    "phase": self._execution_phase(task),
                    "workspace_root": str(workspace_root),
                    "session_id": context.session_id,
                    "termination_reason": termination_reason,
                    "completion_verdict": _string_or_json(completion_meta.get("completion_verdict")),
                    "message": "Attempting partial completion recovery from Hermes subprocess result.",
                },
            )
            partial_result = self._finalize_partial_completion(
                task,
                finalized,
                observed,
                stop_meta,
                lifecycle_events,
                termination_reason=termination_reason,
                observer=observer,
            )
            partial_result.raw["self_review_candidate"] = _json_object_or_empty(parsed_result.get("self_review_candidate"))
            self._store_executor_result(
                execution_id,
                {
                    "execution_id": execution_id,
                    "status": "completed" if partial_result.ok else "failed",
                    "operation_id": task.operation_id,
                    "trace_id": task.trace_id,
                    "workflow_id": task.workflow_id,
                    "executor_instance_id": self._config.executor_instance_id,
                    "phase": self._execution_phase(task),
                    "workspace_root": str(workspace_root),
                    "session_id": context.session_id,
                    "termination_reason": termination_reason,
                    "completion_verdict": _string_or_json(_json_object_or_empty(partial_result.raw).get("completion_verdict")),
                    "message": partial_result.message,
                    "result": self._executor_result_payload(partial_result),
                },
            )
            return partial_result
        external_tool_pending = (
            parsed_result_loaded
            and _string_or_json(completion_meta.get("termination_reason")) == "external_tool_pending"
        )
        should_validate_envelope = (completed_returncode == 0 or (_bool_or_false(parsed_result.get("ok")) and parsed_result_loaded)) and (
            not parsed_result_loaded or _bool_or_false(parsed_result.get("ok"))
        )
        if external_tool_pending:
            result = HermesExecutionResult(
                ok=True,
                message="",
                provider="hermes-native-executor",
                raw={
                    **base_raw,
                    **self._workflow_evidence_raw(task, observed, "external_tool_pending"),
                    **completion_meta,
                    "result": _json_object_or_empty(parsed_result.get("result")),
                    "external_tool_pending": _json_object_or_empty(_json_object_or_empty(parsed_result.get("result")).get("external_tool_pending")),
                    "external_tool_pause_id": _string_or_json(parsed_result.get("external_tool_pause_id")) or _string_or_json(_json_object_or_empty(parsed_result.get("result")).get("external_tool_pause_id")),
                    "suppress_delivery": True,
                    "runner_diagnostics": self._runner_diagnostics(
                        failure_kind="",
                        provider_error_message="",
                        termination_reason="external_tool_pending",
                        session_ready_issues=self._session_manager.ready_issues,
                        repair_attempted=False,
                        repair_succeeded=False,
                        observed={**observed, "suppress_delivery": True},
                    ),
                    "harness_profile_id": task.harness_profile_id,
                    "effective_overlay_version": task.harness_overlay_version,
                },
            )
        elif should_validate_envelope:
            envelope, envelope_failure_class, envelope_diagnostics = self._load_native_execution_envelope(
                task,
                envelope_path=envelope_path,
                context_path=context_path,
            )
            if envelope_failure_class:
                result = self._native_strict_failure(
                    task,
                    failure_class=envelope_failure_class,
                    message=f"Hermes native execution envelope failed validation: {envelope_failure_class}.",
                    diagnostics={
                        **envelope_diagnostics,
                        "result_file_present": result_file_present,
                        "result_parse_error": parse_error,
                    },
                )
                result.raw = {**base_raw, **_json_object_or_empty(result.raw)}
            else:
                envelope_completion = _json_object_or_empty((envelope or {}).get("completion"))
                if not parsed_result_loaded:
                    completion_meta = {
                        "termination_reason": _string_or_json(envelope_completion.get("termination_reason")) or "normal_completion",
                        "completion_verdict": _string_or_json(envelope_completion.get("completion_verdict")) or "complete",
                        "max_iterations_reached": False,
                        "native_result_completed": True,
                        "native_result_partial": False,
                        "native_result_interrupted": False,
                        "native_result_api_calls": 0,
                    }
                result = HermesExecutionResult(
                    ok=True,
                    message=_string_or_json(envelope.get("final_response")) if envelope else "",
                    provider="hermes-native-executor",
                    raw={
                        **base_raw,
                        **self._workflow_evidence_raw(task, observed, completion_meta["termination_reason"]),
                        "execution_envelope": envelope or {},
                        "runner_diagnostics": {
                            **observed,
                            "termination_reason": completion_meta["termination_reason"],
                            "max_iterations_reached": _bool_or_false(completion_meta.get("max_iterations_reached")),
                            "completion_verdict": completion_meta["completion_verdict"],
                            "native_result_completed": _bool_or_false(completion_meta.get("native_result_completed")),
                            "native_result_partial": _bool_or_false(completion_meta.get("native_result_partial")),
                            "native_result_interrupted": _bool_or_false(completion_meta.get("native_result_interrupted")),
                            "native_result_api_calls": completion_meta["native_result_api_calls"],
                            "native_envelope_path": str(envelope_path),
                            "native_context_path": str(context_path),
                            "native_envelope_validated": True,
                        },
                        **completion_meta,
                        "harness_profile_id": task.harness_profile_id,
                        "effective_overlay_version": task.harness_overlay_version,
                    },
                )
        elif parse_error or not parsed_result or not _bool_or_false(parsed_result.get("ok")):
            failure_class = _string_or_json(parsed_result.get("failure_class")) or "runner_non_ok"
            failure_kind = "hermes_contract_failed" if failure_class == "hermes_contract_failed" else "execution_error"
            failure_termination_reason = _string_or_json(completion_meta.get("termination_reason")) if parsed_result_loaded else ""
            if not failure_termination_reason or failure_termination_reason == "normal_completion":
                failure_termination_reason = failure_kind
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
                    **self._workflow_evidence_raw(task, observed, failure_termination_reason),
                    **completion_meta,
                    "termination_reason": failure_termination_reason,
                    "failure_class": failure_class,
                    "runner_diagnostics": self._runner_diagnostics(
                        failure_kind=failure_kind,
                        provider_error_message=message,
                        timeout_kind=_string_or_json(completion_meta.get("timeout_kind")) or None,
                        termination_reason=failure_termination_reason,
                        session_ready_issues=self._session_manager.ready_issues,
                        repair_attempted=False,
                        repair_succeeded=False,
                        observed=observed,
                    ),
                    "artifact_failure_reason": message if execution_phase == "render" else "",
                },
            )
        else:
            raise AssertionError(
                f"Hermes native executor reached logically unreachable branch: "
                f"should_validate_envelope={should_validate_envelope}, "
                f"parsed_result_loaded={parsed_result_loaded}, "
                f"parsed_result.ok={_bool_or_false(parsed_result.get('ok')) if parsed_result else None}, "
                f"parse_error={bool(parse_error)}, "
                f"completed_returncode={completed_returncode}"
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
                    "prompt_skills": list(skill_diagnostics.get("prompt_skills") or []),
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
                "executor_instance_id": self._config.executor_instance_id,
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
        }
        latest_activity = _json_object_or_empty(activity)
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
        lifecycle_events: list[JsonObject] | None = None,
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
        if tracker is not None and hasattr(tracker, "warnings"):
            observed["memory_warnings"] = list(getattr(tracker, "warnings", []) or [])
        if skill_diagnostics:
            observed.update(skill_diagnostics)
        if observer is not None:
            observed.update(observer.diagnostics())
        if lifecycle_events:
            lifecycle_observed = self._lifecycle_observability_metadata(task, lifecycle_events)
            if lifecycle_observed.get("tool_calls"):
                observed["tool_calls"] = self._merge_runtime_values(
                    observed.get("tool_calls"),
                    lifecycle_observed.get("tool_calls"),
                )
            if lifecycle_observed.get("evidence_items"):
                observed["evidence_items"] = self._merge_runtime_values(
                    observed.get("evidence_items"),
                    lifecycle_observed.get("evidence_items"),
                )
            if lifecycle_observed.get("selected_context_surfaces"):
                observed["selected_context_surfaces"] = self._merge_runtime_values(
                    observed.get("selected_context_surfaces"),
                    lifecycle_observed.get("selected_context_surfaces"),
                )
            if lifecycle_observed.get("reply_delivery"):
                observed["reply_delivery"] = lifecycle_observed.get("reply_delivery")
        return observed

    def _lifecycle_observability_metadata(self, task: RunnerTaskRequest, lifecycle_events: list[JsonObject]) -> JsonObject:
        tool_calls: list[JsonObject] = []
        evidence_items: list[JsonObject] = []
        selected_context_surfaces: list[JsonObject] = []
        reply_delivery: JsonObject = {}
        started_by_id: dict[str, JsonObject] = {}
        started_by_name: dict[str, list[JsonObject]] = {}
        for index, item in enumerate(lifecycle_events):
            event_name = self._lifecycle_event_name(item)
            if event_name == "reply_delivery":
                candidate = self._lifecycle_reply_delivery(item)
                if candidate:
                    reply_delivery = candidate
                continue
            if event_name not in {"tool_call_started", "tool_call_completed"}:
                continue
            tool_name = self._lifecycle_tool_name(item)
            if not tool_name:
                continue
            tool_call_id = self._lifecycle_tool_call_id(item) or f"{tool_name}:{index + 1}"
            request_payload = self._lifecycle_request_payload(item)
            if event_name == "tool_call_started":
                if request_payload:
                    started_by_id[tool_call_id] = request_payload
                    started_by_name.setdefault(tool_name, []).append(request_payload)
                continue
            if not request_payload:
                request_payload = started_by_id.get(tool_call_id, {})
            if not request_payload and started_by_name.get(tool_name):
                request_payload = started_by_name[tool_name].pop(0)
            output_payload = self._lifecycle_output_payload(item)
            status = first_non_empty(
                _string_or_json(item.get("status")),
                _string_or_json(output_payload.get("status")),
                "completed",
            )
            summary = first_non_empty(
                _string_or_json(item.get("summary")),
                _string_or_json(output_payload.get("summary")),
                status,
                tool_name,
            )
            provider_ref = first_non_empty(
                _string_or_json(item.get("provider_ref")),
                _string_or_json(output_payload.get("provider_ref")),
            )
            raw_artifact_refs = _string_list_or_empty(item.get("raw_artifact_refs") or output_payload.get("raw_artifact_refs"))
            tool_record: JsonObject = {
                "id": f"runner-lifecycle-tool-{hashlib.sha1((tool_call_id + tool_name).encode('utf-8')).hexdigest()[:12]}",
                "tool_name": tool_name,
                "tool_call_id": tool_call_id,
                "request": dict(request_payload),
                "summary": summary,
                "status": status,
                "provider_ref": provider_ref,
                "raw_artifact_refs": raw_artifact_refs,
            }
            if recorded_at := self._lifecycle_recorded_at(item):
                tool_record["completed_at"] = recorded_at
                tool_record["created_at"] = recorded_at
            tool_calls.append(tool_record)
            evidence_items.extend(
                self._native_lifecycle_evidence_items(
                    tool_name=tool_name,
                    tool_call_id=tool_call_id,
                    output_payload=output_payload,
                    summary=summary,
                    provider_ref=provider_ref,
                )
            )
            if tool_name == "slack.history":
                channel_id = first_non_empty(_string_or_json(request_payload.get("channel_id")), task.channel_id)
                thread_ts = first_non_empty(_string_or_json(request_payload.get("thread_ts")), task.thread_ts)
                if channel_id or thread_ts:
                    selected_context_surfaces.append(
                        {
                            "channel_id": channel_id,
                            "thread_ts": thread_ts,
                            "scope": "bound_thread" if thread_ts else "channel",
                            "source": "native_lifecycle",
                        }
                    )
        return {
            "tool_calls": tool_calls,
            "evidence_items": evidence_items,
            "selected_context_surfaces": selected_context_surfaces,
            "reply_delivery": reply_delivery,
        }

    def _lifecycle_reply_delivery(self, item: JsonObject) -> JsonObject:
        delivery = _json_object_or_empty(item.get("reply_delivery"))
        if not delivery:
            delivery = {
                key: value
                for key, value in item.items()
                if key not in {"event", "event_type", "recorded_at", "recorded_at_unix", "payload"}
            }
        payload = _json_object_or_empty(item.get("payload"))
        for key, value in payload.items():
            delivery.setdefault(key, value)
        tool_name = _string_or_json(delivery.get("tool_name"))
        if not tool_name:
            tool_name = self._lifecycle_tool_name(item)
        if tool_name:
            try:
                delivery["tool_name"] = canonical_tool_name(tool_name)
            except ValueError:
                delivery["tool_name"] = tool_name
        if not delivery.get("send_status"):
            delivery["send_status"] = first_non_empty(_string_or_json(item.get("status")), _string_or_json(delivery.get("status")))
        return delivery if delivery else {}

    def _native_lifecycle_evidence_items(
        self,
        *,
        tool_name: str,
        tool_call_id: str,
        output_payload: JsonObject,
        summary: str,
        provider_ref: str,
    ) -> list[JsonObject]:
        items: list[JsonObject] = []
        if tool_name == "repo.search":
            for index, match in enumerate(output_payload.get("matches") if isinstance(output_payload.get("matches"), list) else []):
                if not isinstance(match, dict):
                    continue
                path = _string_or_json(match.get("path"))
                snippet = _string_or_json(match.get("snippet"))
                if not path and not snippet:
                    continue
                items.append(
                    {
                        "id": f"{tool_call_id}:match:{index + 1}",
                        "source": "native_lifecycle",
                        "tool_name": tool_name,
                        "tool_call_id": tool_call_id,
                        "path": path,
                        "snippet": snippet,
                        "summary": summary,
                        "provider_ref": provider_ref,
                    }
                )
        if tool_name == "repo.read_file":
            path = _string_or_json(output_payload.get("path"))
            content = _string_or_json(output_payload.get("content"))
            if path or content:
                items.append(
                    {
                        "id": f"{tool_call_id}:file",
                        "source": "native_lifecycle",
                        "tool_name": tool_name,
                        "tool_call_id": tool_call_id,
                        "path": path,
                        "snippet": content[:4000],
                        "summary": summary,
                        "provider_ref": provider_ref,
                    }
                )
        return items

    def _lifecycle_event_name(self, item: JsonObject) -> str:
        raw = first_non_empty(_string_or_json(item.get("event_type")), _string_or_json(item.get("event"))).replace(".", "_")
        return raw.strip().lower()

    def _lifecycle_recorded_at(self, item: JsonObject) -> str:
        recorded_at = _string_or_json(item.get("recorded_at"))
        if recorded_at:
            return recorded_at
        recorded_at_unix = item.get("recorded_at_unix")
        if isinstance(recorded_at_unix, Number):
            return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(float(recorded_at_unix)))
        return ""

    def _lifecycle_tool_name(self, item: JsonObject) -> str:
        candidates: list[Any] = [
            item.get("tool_name"),
            item.get("name"),
            item.get("transport_tool_name"),
        ]
        kwargs = _json_object_or_empty(item.get("kwargs"))
        candidates.extend([kwargs.get("tool_name"), kwargs.get("name"), kwargs.get("function_name")])
        for key in ("tool_call", "tool", "function"):
            nested = _json_object_or_empty(kwargs.get(key))
            candidates.extend([nested.get("tool_name"), nested.get("name"), nested.get("function_name")])
        args = item.get("args")
        if isinstance(args, list):
            for arg in args:
                if isinstance(arg, str):
                    candidates.append(arg)
                elif isinstance(arg, dict):
                    candidates.extend([arg.get("tool_name"), arg.get("name"), arg.get("function_name")])
                    function = _json_object_or_empty(arg.get("function"))
                    candidates.extend([function.get("tool_name"), function.get("name"), function.get("function_name")])
        for candidate in candidates:
            name = _string_or_json(candidate)
            if not name or name in {"tool_call_started", "tool_call_completed", "tool.call.started", "tool.call.completed"}:
                continue
            if name in GROUNDED_EVIDENCE_TOOL_NAMES:
                return name
            if name in _LIFECYCLE_TOOL_NAME_ALIASES:
                return _LIFECYCLE_TOOL_NAME_ALIASES[name]
            try:
                return canonical_tool_name(name)
            except ValueError:
                continue
        return ""

    def _lifecycle_tool_call_id(self, item: JsonObject) -> str:
        candidates: list[Any] = [item.get("tool_call_id"), item.get("call_id"), item.get("id")]
        kwargs = _json_object_or_empty(item.get("kwargs"))
        candidates.extend([kwargs.get("tool_call_id"), kwargs.get("call_id"), kwargs.get("id")])
        args = item.get("args")
        if isinstance(args, list):
            for arg in args:
                if isinstance(arg, dict):
                    candidates.extend([arg.get("tool_call_id"), arg.get("call_id"), arg.get("id")])
        for candidate in candidates:
            value = _string_or_json(candidate)
            if value:
                return value
        return ""

    def _lifecycle_request_payload(self, item: JsonObject) -> JsonObject:
        for value in (item.get("request_payload"), item.get("request"), item.get("arguments")):
            parsed = self._parse_json_object_maybe(value)
            if parsed:
                return parsed
        kwargs = _json_object_or_empty(item.get("kwargs"))
        for value in (kwargs.get("request_payload"), kwargs.get("request"), kwargs.get("arguments")):
            parsed = self._parse_json_object_maybe(value)
            if parsed:
                return parsed
        args = item.get("args")
        if isinstance(args, list):
            for arg in args:
                parsed = self._parse_json_object_maybe(arg)
                if parsed:
                    for key in ("request_payload", "request", "arguments"):
                        nested = self._parse_json_object_maybe(parsed.get(key))
                        if nested:
                            return nested
        return {}

    def _lifecycle_output_payload(self, item: JsonObject) -> JsonObject:
        for value in (item.get("output"), item.get("result"), item.get("response")):
            parsed = self._parse_json_object_maybe(value)
            if parsed:
                if output := _json_object_or_empty(parsed.get("output")):
                    return output
                if result := self._parse_json_object_maybe(parsed.get("result")):
                    if output := _json_object_or_empty(result.get("output")):
                        return output
                    return result
                return parsed
        kwargs = _json_object_or_empty(item.get("kwargs"))
        for value in (kwargs.get("output"), kwargs.get("result"), kwargs.get("response")):
            parsed = self._parse_json_object_maybe(value)
            if parsed:
                if output := _json_object_or_empty(parsed.get("output")):
                    return output
                return parsed
        args = item.get("args")
        if isinstance(args, list):
            for arg in args:
                parsed = self._parse_json_object_maybe(arg)
                if parsed:
                    if output := _json_object_or_empty(parsed.get("output")):
                        return output
                    if result := self._parse_json_object_maybe(parsed.get("result")):
                        return result
        return {}

    def _provider_invalid_request_diagnostics(self, message: str) -> JsonObject | None:
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

        diagnostics = self._runner_diagnostics(
            failure_kind="invalid_request",
            provider_status_code=status_code or None,
            provider_error_param=provider_error_param or None,
            provider_error_code=provider_error_code or None,
            provider_error_message=provider_error_message,
            repair_attempted=False,
            repair_succeeded=False,
        )
        return diagnostics

    def _provider_runtime_error_message(self, message: str) -> str:
        text = str(message or "").strip()
        if not text:
            return ""
        try:
            parsed = json.loads(_unwrap_json_code_fence(text))
            if isinstance(parsed, dict):
                return ""
        except (json.JSONDecodeError, ValueError):
            pass
        lower = text.lower()
        if not any(marker in lower for marker in _PROVIDER_RUNTIME_ERROR_MARKERS):
            return ""
        return _truncate_string(text, 8000)

    def _provider_runtime_error_from_agent_result(self, result: Any) -> str:
        if not isinstance(result, dict):
            return ""
        final_response = str(result.get("final_response", "") or "")
        if final_response.strip():
            try:
                parsed = json.loads(_unwrap_json_code_fence(final_response))
                if isinstance(parsed, dict):
                    return ""
            except (json.JSONDecodeError, ValueError):
                pass
        candidates: list[str] = []
        for key in ("error", "error_message", "response", "final_response"):
            if value := _string_or_json(result.get(key)):
                candidates.append(value)
        nested_result = _json_object_or_empty(result.get("result"))
        for key in ("error", "error_message", "response", "final_response"):
            if value := _string_or_json(nested_result.get(key)):
                candidates.append(value)
        for candidate in candidates:
            if message := self._provider_runtime_error_message(candidate):
                return message
        return ""

    def _provider_runtime_error_diagnostics(self, message: str) -> JsonObject | None:
        provider_error_message = self._provider_runtime_error_message(message)
        if not provider_error_message:
            return None
        status_code = 0
        status_match = re.search(
            r"(?:http|status code|status)\s*[:= -]*\s*([45]\d{2})",
            provider_error_message,
            flags=re.IGNORECASE,
        )
        if status_match:
            try:
                status_code = int(status_match.group(1))
            except ValueError:
                status_code = 0
        return self._runner_diagnostics(
            failure_kind="provider_runtime_error",
            provider_status_code=status_code or None,
            provider_error_message=provider_error_message,
            repair_attempted=False,
            repair_succeeded=False,
        )

    def _effective_task_timeout(self, task: RunnerTaskRequest) -> int:
        requested = int(task.timeout_seconds or 0)
        candidates = [requested] if requested > 0 else [self._default_task_timeout_seconds]
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

    def _stage_task_context(
        self,
        session_id: str,
        task: RunnerTaskRequest,
        *,
        context_path: Path | None = None,
        envelope_path: Path | None = None,
    ) -> Path:
        query_hints = self._default_query_hints(task)
        payload = {
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
            "task_message_ts": task.message_ts or _derive_root_message_ts(task),
            "gateway_session_key": canonical_gateway_session_key(task, self._role),
            "proposal_id": task.session_scope_id if (task.session_scope_kind or "").strip() == "proposal_candidate" else "",
            "attempt_id": task.attempt_id,
            "workspace_id": task.workspace_id,
            "execution_mode": task.execution_mode or "",
            "execution_phase": self._execution_phase(task),
            "artifact_output_dir": _string_or_json(task.artifact_destination),
            "hermes_computer_root": self._config.hermes_computer_root,
            "hermes_run_root": self._config.hermes_run_root,
            "hermes_artifact_root": self._config.hermes_artifact_root,
            "hermes_native_terminal_enabled": self._config.hermes_native_terminal_enabled,
            "hermes_native_toolsets": self._hermes_native_toolsets(),
            "hermes_terminal_cwd": self._config.hermes_terminal_cwd,
            "hermes_company_bin_dir": self._config.hermes_company_bin_dir,
            "hermes_kubernetes_context_enabled": self._config.hermes_kubernetes_context_enabled,
            "hermes_kubeconfig_path": self._config.hermes_kubeconfig_path if (self._config.hermes_kubernetes_context_enabled or self._config.hermes_prod_kubernetes_context_enabled) else "",
            "context_summary": task.context_summary or "",
            "task_default_question": query_hints.get("default_question", ""),
            "task_repo_question": query_hints.get("repo_question", ""),
            "task_knowledge_topic": query_hints.get("knowledge_topic", ""),
            "task_knowledge_question": query_hints.get("knowledge_question", ""),
            "task_slack_history_focus": query_hints.get("slack_history_focus", ""),
            "task_slack_search_query": query_hints.get("slack_search_query", ""),
            "context_refs": task.context_refs,
            "tool_timeout_seconds": 30,
            "session_scope_kind": task.session_scope_kind or "",
            "session_scope_id": task.session_scope_id or "",
        }
        if context_path is not None:
            payload["rsi_runtime_context_path"] = str(context_path)
        if envelope_path is not None:
            payload["rsi_runtime_envelope_path"] = str(envelope_path)
        return self._adapter.stage_task_context(
            session_id,
            payload,
            context_path=context_path,
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

    def _default_query_hints(self, task: RunnerTaskRequest) -> JsonObject:
        original_prompt = self._original_task_prompt(task)
        repo = _string_or_json(task.repo)
        user_request = original_prompt
        slack_query_parts = [part for part in [repo] if part]
        slack_search_query = " ".join(slack_query_parts) if slack_query_parts else user_request
        history_focus = user_request
        if user_request:
            history_focus = f"Extract the most relevant messages for answering: {user_request}"
        knowledge_topic = first_non_empty(repo, task.context_summary or "", user_request)
        knowledge_question = user_request
        return {
            "default_question": first_non_empty(user_request, original_prompt),
            "repo_question": first_non_empty(user_request, original_prompt),
            "knowledge_topic": knowledge_topic,
            "knowledge_question": first_non_empty(knowledge_question, user_request, original_prompt),
            "slack_history_focus": first_non_empty(history_focus, user_request, original_prompt),
            "slack_search_query": first_non_empty(slack_search_query, user_request, original_prompt),
        }

    def _evidence_salience_terms(self, task: RunnerTaskRequest) -> list[str]:
        values: list[str] = [
            _string_or_json(task.repo),
            _string_or_json(task.context_summary),
            self._original_task_prompt(task),
            _string_or_json(task.execution_intent),
        ]
        stopwords = {
            "about",
            "after",
            "agent",
            "architecture",
            "artifact",
            "backend",
            "before",
            "company",
            "context",
            "current",
            "diagram",
            "draw",
            "from",
            "investigate",
            "latest",
            "please",
            "repo",
            "render",
            "request",
            "source",
            "task",
            "that",
            "their",
            "thread",
            "truth",
            "using",
            "with",
        }
        terms: list[str] = []
        for value in values:
            if not value:
                continue
            lower = value.lower()
            for match in re.findall(r"[a-z0-9][a-z0-9._/-]{2,}", lower):
                cleaned = match.strip("._/-")
                if len(cleaned) < 3 or cleaned in stopwords:
                    continue
                terms.append(cleaned)
                for part in re.split(r"[._/-]+", cleaned):
                    if len(part) >= 4 and part not in stopwords:
                        terms.append(part)
        return [str(item) for item in self._merge_runtime_values(terms)[:48]]

    def _observation_salience_text(self, item: JsonObject) -> str:
        parts: list[str] = []
        for key in (
            "tool_name",
            "kind",
            "summary",
            "snippet",
            "source_ref",
            "provider_ref",
            "path",
            "repo",
            "ref",
            "url",
            "permalink",
            "request",
            "output",
        ):
            text = _string_or_json(item.get(key))
            if text:
                parts.append(text)
        return "\n".join(parts).lower()[:12000]

    def _observation_selection_key(self, item: JsonObject) -> str:
        return "|".join(
            [
                _string_or_json(item.get("tool_call_id")),
                _string_or_json(item.get("id")),
                _string_or_json(item.get("tool_name")),
                _string_or_json(item.get("source_ref")),
                _string_or_json(item.get("path")),
                _string_or_json(item.get("repo")),
                _string_or_json(item.get("permalink")),
                _string_or_json(item.get("url")),
                _string_or_json(item.get("provider_ref")),
                _string_or_json(item.get("summary"))[:160],
            ]
        )

    def _observation_salience_score(self, item: JsonObject, terms: list[str]) -> int:
        text = self._observation_salience_text(item)
        score = 0
        for term in terms:
            if term and term in text:
                score += 25 if any(separator in term for separator in ("-", "/", ".")) else 10
        tool_name = _string_or_json(item.get("tool_name"))
        if tool_name in GROUNDED_EVIDENCE_TOOL_NAMES:
            score += 20
        if _string_or_json(item.get("source_ref")) or _string_or_json(item.get("path")):
            score += 6
        status = _string_or_json(item.get("status")).lower()
        if status and status not in {"completed", "complete", "ok", "success"}:
            score += 12
        if any(marker in text for marker in ("helm", "deployment", "namespace", "pod", "runtime")):
            score += 8
        return score

    def _select_compact_observations(self, items: list[JsonObject], task: RunnerTaskRequest | None, *, limit: int) -> list[JsonObject]:
        if len(items) <= limit:
            return items
        terms = self._evidence_salience_terms(task) if task is not None else []
        selected: dict[str, tuple[int, JsonObject]] = {}

        def add(index: int, item: JsonObject) -> None:
            key = self._observation_selection_key(item)
            if not key.strip("|"):
                key = f"idx:{index}"
            selected.setdefault(key, (index, item))

        head_count = min(4, max(1, limit // 5))
        tail_count = min(8, max(2, limit // 4))
        for index, item in enumerate(items[:head_count]):
            add(index, item)
        tail_start = max(0, len(items) - tail_count)
        for offset, item in enumerate(items[tail_start:], start=tail_start):
            add(offset, item)

        ranked = sorted(
            (
                (self._observation_salience_score(item, terms), index, item)
                for index, item in enumerate(items)
            ),
            key=lambda entry: (-entry[0], entry[1]),
        )
        for score, index, item in ranked:
            if score <= 0:
                break
            add(index, item)
            if len(selected) >= limit:
                break
        return [item for _index, item in sorted(selected.values(), key=lambda entry: entry[0])[:limit]]

    def _compact_tool_calls(
        self, observed: JsonObject, *, task: RunnerTaskRequest | None = None, limit: int = 30
    ) -> list[JsonObject]:
        compact: list[JsonObject] = []
        items = self._select_compact_observations(_json_object_list(observed.get("tool_calls")), task, limit=limit)
        for item in items:
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

    def _compact_evidence_items(
        self, observed: JsonObject, *, task: RunnerTaskRequest | None = None, limit: int = 40
    ) -> list[JsonObject]:
        compact: list[JsonObject] = []
        items = self._select_compact_observations(_json_object_list(observed.get("evidence_items")), task, limit=limit)
        for item in items:
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
        tool_calls = self._compact_tool_calls(observed, task=task)
        evidence_items = self._compact_evidence_items(observed, task=task)
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

    def _workflow_evidence_raw(self, task: RunnerTaskRequest, observed: JsonObject, termination_reason: str) -> JsonObject:
        if task.task_type != "workflow":
            return {}
        return {"evidence_ledger": self._build_evidence_ledger(task, observed, termination_reason)}

    def _partial_reducer_system_prompt(self) -> str:
        return "\n".join(
            [
                "You finalize bounded-stop RSI Slack reply workflows.",
                "Use only the supplied evidence ledger.",
                "Do not call tools. Do not invent evidence. Do not speculate beyond the ledger.",
                "Return only one JSON object with keys: session_title, visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, outcome_hypotheses, produced_artifacts, artifact_failure_reason, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta.",
                "Set session_title to a concise 3-8 word rewrite of the user's Slack question; preserve intent, omit @mentions and filler, and do not summarize your answer.",
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

    def _json_object_input_prompt(self, prompt: str) -> str:
        text = str(prompt or "").strip()
        if "json" in text.lower():
            return text
        prefix = "Return a JSON object only."
        if not text:
            return prefix
        return f"{prefix}\n\n{text}"

    def _invoke_hermes_json_reducer(
        self,
        *,
        system_prompt: str,
        user_prompt: str,
        timeout_seconds: int,
        reasoning_effort: str,
        normalize_workflow_output: bool,
        recorder: NativeExecutionRecorder | None = None,
        operation: str = "hermes_json_reducer",
    ) -> PartialReducerAttemptResult:
        if AIAgent is None:
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error="Hermes reducer is unavailable.",
                provider_response_id="",
            )
        profile = self._model_profiles["main"]
        api_key = _runtime_api_key_for_provider(profile.provider)
        if not api_key:
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error=self._credentials_error_message_for_profile(profile),
                provider_response_id="",
            )
        session_id = f"rsi-{self._role}-{operation}-{int(time.time() * 1000)}"
        agent_kwargs: JsonObject = {
            "model": profile.provider_model,
            "provider": profile.provider,
            "api_key": api_key,
            "quiet_mode": True,
            "reasoning_config": parse_reasoning_effort(reasoning_effort) or {"enabled": True, "effort": reasoning_effort},
            "enabled_toolsets": [],
            "skip_context_files": True,
            "skip_memory": True,
            "max_iterations": 1,
            "session_id": session_id,
            "self_review_mode": "manual",
        }
        if profile.base_url:
            agent_kwargs["base_url"] = profile.base_url
        self._apply_openrouter_provider_routing(
            agent_kwargs,
            provider=profile.provider,
            provider_routing=profile.openrouter_provider_routing,
        )
        if recorder is not None:
            recorder.record(
                "hermes_reducer_request",
                {
                    "operation": operation,
                    "timeout_seconds": timeout_seconds,
                    "model": profile.provider_model,
                    "provider": profile.provider,
                    "provider_routing": dict(profile.openrouter_provider_routing)
                    if profile.provider == "openrouter"
                    else {},
                },
            )
        agent = AIAgent(**agent_kwargs)
        executor = concurrent.futures.ThreadPoolExecutor(max_workers=1)
        future = executor.submit(agent.run_conversation, user_prompt, system_prompt, [], session_id)
        try:
            result = future.result(timeout=max(1, timeout_seconds))
        except concurrent.futures.TimeoutError:
            if hasattr(agent, "interrupt"):
                try:
                    agent.interrupt(f"runner {operation} timeout after {timeout_seconds}s")
                except Exception:
                    pass
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error=f"Hermes reducer timed out after {timeout_seconds}s.",
                provider_response_id="",
            )
        except Exception as exc:
            return PartialReducerAttemptResult(
                ok=False,
                response_text="",
                structured_output={},
                error=str(exc),
                provider_response_id="",
            )
        finally:
            executor.shutdown(wait=False, cancel_futures=True)
        response_text = str((result or {}).get("final_response", "") or "")
        provider_response_id = first_non_empty(
            _string_or_json((result or {}).get("provider_response_id")),
            _string_or_json((result or {}).get("response_id")),
        )
        if provider_error := self._provider_runtime_error_from_agent_result(result):
            return PartialReducerAttemptResult(
                ok=False,
                response_text=response_text,
                structured_output={},
                error=f"Hermes reducer provider call failed: {provider_error}",
                provider_response_id=provider_response_id,
            )
        try:
            parsed = json.loads(response_text)
            if not isinstance(parsed, dict):
                raise ValueError("Hermes reducer returned non-object JSON.")
            structured_output = _normalize_structured_output(parsed) if normalize_workflow_output else parsed
        except Exception as exc:
            return PartialReducerAttemptResult(
                ok=False,
                response_text=response_text,
                structured_output={},
                error=f"Hermes reducer failed to return valid structured output: invalid JSON: {exc}",
                provider_response_id=provider_response_id,
            )
        if recorder is not None:
            recorder.record(
                "hermes_reducer_response",
                {
                    "operation": operation,
                    "provider_response_id": provider_response_id,
                },
            )
        return PartialReducerAttemptResult(
            ok=True,
            response_text=response_text,
            structured_output=structured_output,
            error="",
            provider_response_id=provider_response_id,
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
        return self._invoke_hermes_json_reducer(
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
            normalize_workflow_output=True,
            operation="partial_reducer",
        )

    def _partial_completion_idempotency_key(self, task: RunnerTaskRequest, termination_reason: str) -> str:
        scope = first_non_empty(task.trace_id, task.workflow_id, task.session_scope_id, "workflow")
        return f"partial-{termination_reason}-{scope}"

    def _workflow_reply_delivery_mode(self, task: RunnerTaskRequest) -> str:
        mode = (task.reply_delivery_mode or "").strip().lower()
        if mode in {"direct", "mediated", "none"}:
            return mode
        return "none"

    def _workflow_requires_explicit_reply_action(self, task: RunnerTaskRequest) -> bool:
        return task.task_type == "workflow" and self._workflow_reply_delivery_mode(task) == "mediated"

    def _workflow_allows_fallback_reply_action(self, task: RunnerTaskRequest) -> bool:
        return task.task_type == "workflow" and self._workflow_reply_delivery_mode(task) != "none"

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

    def _contains_markdown_pipe_table_outside_fence(self, text: str) -> bool:
        """Detect markdown pipe tables outside code fences, matching Go validation logic."""
        in_fence = False
        previous_pipe = False
        previous_separator = False
        
        for line in text.split("\n"):
            trimmed = line.strip()
            if trimmed.startswith("```") or trimmed.startswith("~~~"):
                in_fence = not in_fence
                previous_pipe = False
                previous_separator = False
                continue
            if in_fence:
                continue
            has_pipe = trimmed.count("|") >= 2
            is_separator = has_pipe and _MARKDOWN_TABLE_SEPARATOR_REGEX.match(trimmed)
            if (is_separator and previous_pipe) or (previous_separator and has_pipe):
                return True
            previous_pipe = has_pipe and not is_separator
            previous_separator = is_separator
        return False

    def _synthesize_partial_slack_post_action(
        self,
        task: RunnerTaskRequest,
        structured_output: JsonObject,
        termination_reason: str,
    ) -> tuple[JsonObject, bool]:
        if not self._workflow_allows_fallback_reply_action(task):
            return structured_output, False
        for item in _normalize_proposed_actions(structured_output.get("proposed_actions")):
            if _string_or_json(item.get("kind")) in {"slack_post", "slack_report"}:
                return structured_output, False
        if _json_object_or_empty(structured_output.get("reply_delivery")):
            return structured_output, False
        reply_body = first_non_empty(
            _string_or_json(structured_output.get("final_answer")),
            _string_or_json(structured_output.get("reply_draft")),
        )
        if not reply_body:
            return structured_output, False
        if self._contains_markdown_pipe_table_outside_fence(reply_body):
            normalized = dict(structured_output)
            diagnostics = dict(_json_object_or_empty(normalized.get("runner_diagnostics")))
            diagnostics["action_contract_synthesis_skipped"] = "markdown_pipe_table_requires_slack_report"
            normalized["runner_diagnostics"] = diagnostics
            normalized["action_contract_synthesis_error"] = {
                "code": "markdown_pipe_table_requires_slack_report",
                "path": "$.final_answer",
                "message": "Fallback slack_post synthesis skipped because the partial reply contains a Markdown pipe table; use slack_report.tables for structured tabular output.",
            }
            return normalized, False
        actions = [
            dict(item)
            for item in _normalize_proposed_actions(structured_output.get("proposed_actions"))
            if _string_or_json(item.get("kind")) not in {"slack_post", "slack_report"}
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
        diagnostics["partial_finalization_mode"] = "hermes_reducer"
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
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "max_iterations": self._max_iterations,
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
        if self._native_executor_enabled_for_task(task):
            raw["native_strict"] = True
            raw["native_timeout_source"] = "hermes_subprocess"
        else:
            raw["task_timeout_seconds"] = self._effective_task_timeout(task)
            raw["inactivity_timeout_seconds"] = self._effective_inactivity_timeout(self._effective_task_timeout(task))
        result_provider = "hermes-native-executor" if self._native_executor_enabled_for_task(task) else self._backend
        return HermesExecutionResult(
            ok=False,
            message=first_non_empty(recovery_error, message),
            provider=result_provider,
            raw=raw,
        )

    def _partial_completion_failure(
        self,
        task: RunnerTaskRequest,
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
            merged_runner_diagnostics["partial_finalization_mode"] = "hermes_reducer"
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
                "transport_timeout_seconds": self._transport_timeout_seconds,
                "max_iterations": self._max_iterations,
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
            if self._native_executor_enabled_for_task(task):
                merged_raw["native_strict"] = True
                merged_raw["native_timeout_source"] = "hermes_subprocess"
            else:
                merged_raw["task_timeout_seconds"] = self._effective_task_timeout(task)
                merged_raw["inactivity_timeout_seconds"] = self._effective_inactivity_timeout(self._effective_task_timeout(task))
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
            result_provider = "hermes-native-executor" if self._native_executor_enabled_for_task(task) else self._backend
            return HermesExecutionResult(
                ok=True,
                message=response_text,
                provider=result_provider,
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
            finalized,
            observed,
            stop_meta,
            lifecycle_events,
            termination_reason=termination_reason,
            recovery_error=first_non_empty(
                last_error,
                "Hermes bounded-stop reducer could not produce valid structured output.",
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
        if task.task_type == "workflow":
            return names
        return set()

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
        repair_instruction: str | None = None,
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
        if repair_instruction:
            repair_callable = getattr(agent, "repair_with_instructions", None)
            if not callable(repair_callable):
                raise RuntimeError("Hermes AIAgent.repair_with_instructions is required for clean repair mode.")
            future = executor.submit(
                repair_callable,
                instructions=repair_instruction,
                system_message=task.system_message,
                conversation_history=context.conversation_history,
                task_id=context.session_id,
            )
        else:
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
        self._attach_executor_status_heartbeat(observer, task)
        self._store_executor_result(
            observer.execution_id,
            {
                "execution_id": observer.execution_id,
                "status": "running",
                "message": "Execution accepted.",
                "executor_instance_id": self._config.executor_instance_id,
                "phase": self._execution_phase(task),
            },
        )
        if self._config.execution_envelope_v1_enabled:
            contract_validation = self._company_computer.validate_task(task)
            if not contract_validation.ok:
                if self._native_executor_enabled_for_task(task):
                    result = HermesExecutionResult(
                        ok=False,
                        message="Native Hermes workflow preflight failed: " + "; ".join(contract_validation.errors),
                        provider="hermes-native-executor",
                        raw={
                            **self._company_computer.failure_result_raw(task, errors=contract_validation.errors),
                            "native_strict": True,
                            "failure_class": "native_workflow_preflight_failed",
                            "runner_diagnostics": {
                                "failure_kind": "native_workflow_preflight_failed",
                                "termination_reason": "native_workflow_preflight_failed",
                                "contract_errors": list(contract_validation.errors),
                            },
                            "termination_reason": "native_workflow_preflight_failed",
                        },
                    )
                else:
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
            for phase in self._company_computer.planned_phase_runs(task):
                observer.emit(
                    phase=_string_or_json(phase.get("phase_id")),
                    event_type="phase.planned",
                    status="planned",
                    payload=phase,
                )
        result = self._execute_task_internal(task, observer=observer)
        if self._result_is_native_strict(result):
            existing_status = str(self.executor_status(observer.execution_id).get("status") or "").strip().lower()
            if existing_status in {"cancel_requested", "cancelling"} and result.ok:
                result = self._native_cancelled_result(task, result)
            elif not result.ok and _json_object_or_empty(result.raw).get("failure_class") == "runner_cancelled":
                result = self._native_cancelled_result(task, result)
        if self._config.execution_envelope_v1_enabled and not self._result_is_native_strict(result):
            result = self._company_computer.attach_envelope(task, result, observer=observer)
        final_raw = _json_object_or_empty(result.raw)
        non_deliverable_status = "cancelled" if _bool_or_false(final_raw.get("non_deliverable")) else None
        final_status = self._executor_final_status(
            task,
            result,
            execution_id=observer.execution_id,
            status=non_deliverable_status,
        )
        self._store_executor_result(observer.execution_id, final_status)
        self._finalize_self_review_candidate(task, result, final_status, execution_id=observer.execution_id)
        return result

    def _execute_task_internal(self, task: RunnerTaskRequest, *, observer: ObservationEmitter | None = None) -> HermesExecutionResult:
        if task.task_type not in ROLE_TASK_TYPES.get(self._role, {self._role}):
            return HermesExecutionResult(
                ok=False,
                message=f"Runner role {self._role} cannot execute task type {task.task_type}.",
                provider="policy",
                raw={"role": self._role, "task_type": task.task_type},
            )
        if self._task_uses_artifact_phases(task):
            return self._execute_artifact_workflow_task(task, observer=observer)
        if self._native_executor_enabled_for_task(task):
            result = self._execute_native_workflow_task_request(
                task,
                observer=observer,
                max_iterations_override=self._phase_max_iterations_override(task),
            )
        else:
            result = self._execute_task_request(
                task,
                observer=observer,
                max_iterations_override=self._phase_max_iterations_override(task),
            )
        if self._result_is_native_strict(result) and not self._native_strict_should_project_workflow_output(task, result):
            return result
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
        action_contract_repair_attempts = 0
        action_contract_repair_error = ""
        action_contract_repair_errors: list[JsonObject] = []
        action_contract_repair_response = ""
        action_contract_repair_responses: list[str] = []
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
            action_contract_repair_attempts = self._action_contract_repair_attempt_count(
                result.raw,
                partial_runner_diagnostics,
            )
        else:
            if invalid_request := self._provider_invalid_request_diagnostics(result.message):
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
            if provider_runtime_error := self._provider_runtime_error_diagnostics(result.message):
                diagnostics = dict(provider_runtime_error)
                for key, value in _json_object_or_empty(result.raw.get("runner_diagnostics")).items():
                    diagnostics[key] = value
                return HermesExecutionResult(
                    ok=False,
                    message=string_from_map(diagnostics, "provider_error_message") or "Provider call failed.",
                    provider=result.provider,
                    raw={
                        **result.raw,
                        "failure_class": "runner_provider_error",
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
                    repair_instruction = self._build_structured_output_repair_instruction(rendered_task, result.message)
                    repair_result = self._execute_task_request(
                        rendered_task,
                        observer=observer,
                        render_prompt=False,
                        expand_skills=False,
                        enabled_toolsets_override=self._action_contract_repair_toolsets(),
                        repair_instruction=repair_instruction,
                    )
                    if repair_result.ok:
                        if invalid_request := self._provider_invalid_request_diagnostics(repair_result.message):
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
                        if provider_runtime_error := self._provider_runtime_error_diagnostics(repair_result.message):
                            diagnostics = dict(provider_runtime_error)
                            for key, value in _json_object_or_empty(repair_result.raw.get("runner_diagnostics")).items():
                                diagnostics[key] = value
                            diagnostics["repair_attempted"] = True
                            diagnostics["repair_succeeded"] = False
                            return HermesExecutionResult(
                                ok=False,
                                message=string_from_map(diagnostics, "provider_error_message") or "Provider call failed.",
                                provider=repair_result.provider,
                                raw={
                                    **repair_result.raw,
                                    "failure_class": "runner_provider_error",
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
            if reply_delivery:
                structured_output["reply_delivery"] = reply_delivery
                result.raw["reply_delivery"] = reply_delivery
            action_contract_errors = self._workflow_reply_action_contract_errors(task, structured_output, result.raw)
            if action_contract_errors and self._action_contract_repair_attempt_limit() > 0:
                action_contract_repair_attempted = True
                max_action_contract_repair_attempts = self._action_contract_repair_attempt_limit()
                while action_contract_errors and action_contract_repair_attempts < max_action_contract_repair_attempts:
                    action_contract_repair_attempts += 1
                    logger.info(
                        "workflow runner action-contract repair attempted trace_id=%s workflow_id=%s attempt=%s/%s",
                        task.trace_id or "",
                        task.workflow_id or "",
                        action_contract_repair_attempts,
                        max_action_contract_repair_attempts,
                    )
                    repair_instruction = self._build_action_contract_repair_instruction(
                        rendered_task,
                        structured_output,
                        action_contract_errors,
                        attempt=action_contract_repair_attempts,
                        max_attempts=max_action_contract_repair_attempts,
                    )
                    repair_result = self._execute_task_request(
                        rendered_task,
                        observer=observer,
                        render_prompt=False,
                        expand_skills=False,
                        enabled_toolsets_override=self._action_contract_repair_toolsets(),
                        repair_instruction=repair_instruction,
                    )
                    action_contract_repair_response = repair_result.message
                    if repair_result.message:
                        action_contract_repair_responses.append(repair_result.message)
                    if repair_result.ok:
                        try:
                            repaired_output = self._extract_structured_output(repair_result.message)
                        except HermesStructuredOutputError as exc:
                            action_contract_repair_error = str(exc)
                            action_contract_repair_errors.append(
                                {
                                    "attempt": action_contract_repair_attempts,
                                    "code": "structured_output_invalid",
                                    "message": str(exc),
                                }
                            )
                            action_contract_errors = [
                                {
                                    "code": "structured_output_invalid",
                                    "path": "$",
                                    "message": str(exc),
                                }
                            ]
                            continue
                        structured_output = repaired_output
                        repaired_contract_errors = self._workflow_reply_action_contract_errors(task, repaired_output, repair_result.raw)
                        if not repaired_contract_errors:
                            result = repair_result
                            action_contract_repair_succeeded = True
                            action_contract_errors = []
                            logger.info(
                                "workflow runner action-contract repair succeeded trace_id=%s workflow_id=%s attempt=%s",
                                task.trace_id or "",
                                task.workflow_id or "",
                                action_contract_repair_attempts,
                            )
                            break
                        action_contract_errors = repaired_contract_errors
                        action_contract_repair_error = json.dumps(
                            {
                                "code": "final_action_contract_invalid",
                                "message": "Hermes repair response still failed the final Slack action contract.",
                                "attempt": action_contract_repair_attempts,
                                "errors": repaired_contract_errors,
                            },
                            ensure_ascii=True,
                            sort_keys=True,
                        )
                        action_contract_repair_errors.append(
                            {
                                "attempt": action_contract_repair_attempts,
                                "code": "final_action_contract_invalid",
                                "errors": repaired_contract_errors,
                            }
                        )
                    else:
                        action_contract_repair_error = repair_result.message
                        action_contract_repair_errors.append(
                            {
                                "attempt": action_contract_repair_attempts,
                                "code": "repair_request_failed",
                                "message": repair_result.message,
                            }
                        )
                        action_contract_errors = [
                            {
                                "code": "repair_request_failed",
                                "path": "$",
                                "message": repair_result.message,
                            }
                        ]
                        break
        existing_repair_attempted = bool(result.raw.get("repair_attempted"))
        existing_repair_succeeded = bool(result.raw.get("repair_succeeded"))
        result.raw = {
            **result.raw,
            "role": self._role,
            "task_type": task.task_type,
            "repo": task.repo,
            "repo_ref": task.repo_ref,
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
            "message_ts": task.message_ts or _derive_root_message_ts(task),
            "gateway_session_key": canonical_gateway_session_key(task, self._role),
            "repo_allowlist": task.repo_allowlist,
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
            "delivery_policy": task.delivery_policy,
            "workspace_policy": task.workspace_policy,
            "approval_policy": task.approval_policy,
            "repair_attempted": repair_attempted or existing_repair_attempted,
            "repair_succeeded": repair_succeeded or existing_repair_succeeded,
            "action_contract_repair_attempted": action_contract_repair_attempted,
            "action_contract_repair_succeeded": action_contract_repair_succeeded,
            "action_contract_repair_attempts": action_contract_repair_attempts,
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
        runner_diagnostics["action_contract_repair_attempts"] = action_contract_repair_attempts
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
        if action_contract_repair_errors:
            result.raw["action_contract_repair_errors"] = action_contract_repair_errors
            runner_diagnostics["action_contract_repair_errors"] = action_contract_repair_errors
        if action_contract_repair_response:
            result.raw["action_contract_repair_response"] = action_contract_repair_response
            runner_diagnostics["action_contract_repair_response"] = action_contract_repair_response
        if action_contract_repair_responses:
            result.raw["action_contract_repair_responses"] = action_contract_repair_responses
            runner_diagnostics["action_contract_repair_responses"] = action_contract_repair_responses
        result.raw["runner_diagnostics"] = runner_diagnostics
        return result

    def _render_task_prompt(self, task: RunnerTaskRequest) -> str:
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
        parts.append(f"Execution contract: {task.contract_version or EXECUTION_CONTRACT_VERSION}")
        parts.append(f"Runner planner mode: {self._config.runner_planner_mode or RUNNER_PLANNER_MODE}")
        if self._config.hermes_native_terminal_enabled:
            parts.append(
                "Native company-computer toolsets enabled: "
                f"{', '.join(self._hermes_native_toolsets())}. Use the terminal for ordinary CLI work from "
                f"{self._config.hermes_terminal_cwd}; install reusable CLI binaries under {self._config.hermes_company_bin_dir}."
            )
            parts.append(
                "The GitHub CLI is available on the company computer image and is authenticated per request with the configured GitHub App installation. If GitHub App credentials cannot be minted, execution fails before Hermes starts instead of falling back to another token source."
            )
            if self._grafana_observability_configured():
                parts.append(
                    "Grafana read-only observability is available through native `rsi_observability.*` tools in the `rsi-observability` toolset. "
                    "The Grafana service-account token is mounted in the executor environment and is shell-visible; "
                    "Grafana Viewer/RBAC is the read boundary, and the tools are not an RSI policy-enforcement boundary. "
                    "Use `rsi_observability.metrics_query` for Prometheus/Thanos, `rsi_observability.logs_query` for Loki, "
                    "`rsi_observability.dashboards_search`/`dashboard_get` for dashboards, "
                    "`rsi_observability.alert_rules_search`/`alert_rule_get`/`active_alerts` for Grafana alerts, "
                    "and `rsi_observability.datasources` for datasource discovery. "
                    "`rsi_observability.*` is configured with GRAFANA_SERVER, GRAFANA_TOKEN, RSI_GRAFANA_METRICS_DATASOURCE_UID, and RSI_GRAFANA_LOGS_DATASOURCE_UID. "
                    "Do not print tokens and do not call Grafana write APIs. "
                    "Treat explicit environment, cluster, namespace, and pod matchers as required operational guidance, not as a hard security boundary. "
                    "Dashboard edits and imports must be PRs to storyprotocol/story-infra-aws, not live Grafana writes."
                )
            if self._config.db_read_gateway_configured:
                parts.append(
                    "Postgres read requests are available through native `db_read.*` tools in the `rsi-db-read` "
                    "toolset. This is an async, Slack-approved RSI control-plane gateway: the tools never have DB "
                    "DSNs and never execute SQL locally. Use `db_read.sources`, `db_read.schema`, "
                    "`db_read.validate`, `db_read.query`, and `db_read.status`. `db_read.query` pauses the tool "
                    "call for Slack admin approval, then resumes this Hermes session with a sanitized tool result. "
                    "Approval is bound to the exact target and SQL hash; do not bypass this native tool with raw "
                    "control-plane curl or any legacy `rsi-db` executable from agent code."
                )
            repo_guidance = self._github_repository_guidance(task)
            if repo_guidance:
                parts.append(repo_guidance)
            if self._config.hermes_kubernetes_context_enabled:
                parts.append(
                    "Kubernetes read context is available to terminal-native CLIs through KUBECONFIG. Cluster mutations are blocked by policy/RBAC and must become approval intents."
                )
            if self._config.hermes_prod_kubernetes_context_enabled:
                parts.append(
                    "Production Kubernetes topology read access is available through the "
                    f"{self._config.hermes_prod_kubernetes_context_name} kubeconfig context. Use explicit "
                    f"kubectl --context {self._config.hermes_prod_kubernetes_context_name} commands for production reads."
                )
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
        parts.append(f"Timeout seconds: {self._effective_task_timeout(task)}")
        execution_mode = (task.execution_mode or "").strip().lower()
        if self._config.hermes_native_terminal_enabled:
            parts.append("Use the native company-computer toolsets listed in the phase contract plus configured MCP servers.")
        else:
            parts.append("Hermes native terminal tools are not enabled for this runner.")
        parts.append("Eval is read-only. Proposal investigate mode is read-only. Proposal diagnose mode is read-only and must stay grounded in persisted evidence before expanding to repo or log reads. Proposal implement mode may mutate only through native Hermes tools inside the bound workspace; it must not merge code, launch privileged jobs, or post to Slack unless the task contract explicitly allows it.")
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
                "Return a JSON object with keys: session_title, visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, reply_delivery, knowledge_drafts, outcome_hypotheses, produced_artifacts, artifact_failure_reason, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta."
            )
            parts.append(
                "Set session_title to a concise 3-8 word rewrite of the user's Slack question; preserve intent, omit @mentions and filler, and do not summarize your answer."
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
        original_text = text
        text = _unwrap_json_code_fence(text)
        try:
            parsed = json.loads(text)
        except json.JSONDecodeError as exc:
            for payload in reversed(_json_code_fence_payloads(original_text)):
                try:
                    parsed_candidate = json.loads(payload)
                except json.JSONDecodeError:
                    continue
                if isinstance(parsed_candidate, dict) and _looks_like_structured_output(parsed_candidate):
                    return _normalize_structured_output(parsed_candidate)
            raise HermesStructuredOutputError("Hermes execution returned non-JSON output; structured output is required.") from exc
        if not isinstance(parsed, dict):
            raise HermesStructuredOutputError("Hermes execution returned a non-object JSON payload; structured output must be a JSON object.")
        return _normalize_structured_output(parsed)

    def _structured_output_repairable(self, exc: HermesStructuredOutputError) -> bool:
        message = str(exc).lower()
        return "non-json" in message or "non-object json" in message

    def _build_structured_output_repair_instruction(self, task: RunnerTaskRequest, raw_response: str) -> str:
        return "\n".join(
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

    def _workflow_missing_explicit_reply_action(self, task: RunnerTaskRequest, structured_output: JsonObject) -> bool:
        return bool(self._workflow_reply_action_contract_errors(task, structured_output))

    def _workflow_reply_action_contract_errors(self, task: RunnerTaskRequest, structured_output: JsonObject, raw: JsonObject | None = None) -> list[JsonObject]:
        if not self._workflow_requires_explicit_reply_action(task):
            return []
        if self._rsi_native_slack_reply_delivery_succeeded(_json_object_or_empty((raw or {}).get("reply_delivery"))):
            return []
        final_answer = _string_or_json(structured_output.get("final_answer"))
        reply_draft = _string_or_json(structured_output.get("reply_draft"))
        proposed_actions = _normalize_proposed_actions(structured_output.get("proposed_actions"))
        reply_actions = [
            item
            for item in proposed_actions
            if _string_or_json(item.get("kind")) in {"slack_post", "slack_report"}
        ]
        if not final_answer and not reply_draft and not reply_actions:
            return []
        errors: list[JsonObject] = []
        if final_answer or reply_draft:
            errors.append(
                {
                    "code": "missing_native_slack_delivery",
                    "path": "reply_delivery",
                    "message": "Workflow replies must be delivered through rsi_slack.message_post or rsi_slack.report_post before completion.",
                }
            )
        if len(reply_actions) > 1:
            errors.append(
                {
                    "code": "multiple_final_reply_actions",
                    "path": "proposed_actions",
                    "limit": 1,
                    "actual": len(reply_actions),
                    "message": "Workflow replies must include exactly one final Slack action.",
                }
            )
        for index, item in enumerate(reply_actions):
            kind = _string_or_json(item.get("kind"))
            payload = _json_object_or_empty(item.get("request_payload"))
            errors.append(
                {
                    "code": "legacy_proposed_slack_action",
                    "path": f"proposed_actions[{index}]",
                    "message": "Legacy proposed Slack actions are disabled; call rsi_slack.message_post or rsi_slack.report_post instead.",
                }
            )
            if kind == "slack_post":
                body = first_non_empty(
                    _string_or_json(payload.get("final_body")),
                    _string_or_json(payload.get("body")),
                    final_answer,
                    reply_draft,
                    _string_or_json(payload.get("draft_body")),
                )
                if self._contains_markdown_pipe_table_outside_fence(body):
                    errors.append(
                        {
                            "code": "unsupported_markdown_table",
                            "path": f"proposed_actions[{index}].request_payload.body",
                            "message": "slack_post is for simple prose; move tabular output into slack_report.tables.",
                        }
                    )
            elif kind == "slack_report":
                errors.extend(self._validate_slack_report_action_payload(payload, f"proposed_actions[{index}].request_payload"))
        return errors

    def _action_contract_repair_toolsets(self) -> list[str]:
        return normalize_tool_names(["rsi-slack"])

    def _rsi_native_slack_reply_delivery_succeeded(self, delivery: JsonObject) -> bool:
        if not delivery:
            return False
        tool_name = _string_or_json(delivery.get("tool_name"))
        try:
            tool_name = canonical_tool_name(tool_name)
        except ValueError:
            pass
        if tool_name not in {"rsi_slack.message_post", "rsi_slack.report_post"}:
            return False
        status = first_non_empty(
            _string_or_json(delivery.get("send_status")),
            _string_or_json(delivery.get("status")),
        ).lower()
        return status in {"posted", "sent", "succeeded", "success", "completed", "ok"}

    def _action_contract_repair_attempt_limit(self) -> int:
        configured = max(0, int(self._config.workflow_runner_repair_attempts))
        # Delivery contract repair is intentionally bounded: it should only wrap
        # existing output into a valid Slack action, not become another research loop.
        return min(configured, 2)

    def _action_contract_repair_attempt_count(self, raw: JsonObject, diagnostics: JsonObject) -> int:
        return _int_or_zero(
            _first_non_none(
                raw.get("action_contract_repair_attempts"),
                diagnostics.get("action_contract_repair_attempts"),
            )
        )

    def _validate_slack_report_action_payload(self, payload: JsonObject, base_path: str) -> list[JsonObject]:
        errors: list[JsonObject] = []
        if "report" in payload and isinstance(payload["report"], dict):
            payload = payload["report"]
            base_path = f"{base_path}.report"
        version = payload.get("report_schema_version")
        if version != 1:
            errors.append(
                {
                    "code": "unsupported_version",
                    "path": f"{base_path}.report_schema_version",
                    "limit": 1,
                    "actual": version,
                    "message": "report_schema_version must be 1.",
                }
            )
        summary = _string_or_json(payload.get("summary"))
        if not summary:
            errors.append(
                {
                    "code": "required",
                    "path": f"{base_path}.summary",
                    "message": "slack_report summary is required.",
                }
            )
        elif len(summary) > 2800:
            errors.append(
                {
                    "code": "too_long",
                    "path": f"{base_path}.summary",
                    "limit": 2800,
                    "actual": len(summary),
                    "message": "summary is too long.",
                }
            )
        if self._contains_markdown_pipe_table_outside_fence(summary):
            errors.append(
                {
                    "code": "unsupported_markdown_table",
                    "path": f"{base_path}.summary",
                    "message": "Use slack_report.tables instead of raw pipe tables in summary.",
                }
            )
        sections = payload.get("sections")
        if sections is not None and not isinstance(sections, list):
            errors.append({"code": "invalid_type", "path": f"{base_path}.sections", "message": "sections must be a list."})
        elif isinstance(sections, list):
            for section_index, section in enumerate(sections):
                if not isinstance(section, dict):
                    errors.append({"code": "invalid_type", "path": f"{base_path}.sections[{section_index}]", "message": "section must be an object."})
                    continue
                text = _string_or_json(section.get("text"))
                if len(text) > 2800:
                    errors.append(
                        {
                            "code": "too_long",
                            "path": f"{base_path}.sections[{section_index}].text",
                            "limit": 2800,
                            "actual": len(text),
                            "message": "section text is too long.",
                        }
                    )
                if self._contains_markdown_pipe_table_outside_fence(text):
                    errors.append(
                        {
                            "code": "unsupported_markdown_table",
                            "path": f"{base_path}.sections[{section_index}].text",
                            "message": "Use slack_report.tables instead of raw pipe tables in section text.",
                        }
                    )
        tables = payload.get("tables")
        if tables is not None and not isinstance(tables, list):
            errors.append({"code": "invalid_type", "path": f"{base_path}.tables", "message": "tables must be a list."})
        elif isinstance(tables, list):
            for table_index, table in enumerate(tables):
                errors.extend(self._validate_slack_report_table(table, f"{base_path}.tables[{table_index}]"))
        for field_name in ("files", "images"):
            items = payload.get(field_name)
            if items is None:
                continue
            if not isinstance(items, list):
                errors.append({"code": "invalid_type", "path": f"{base_path}.{field_name}", "message": f"{field_name} must be a list."})
                continue
            for item_index, item in enumerate(items):
                if not isinstance(item, dict):
                    errors.append({"code": "invalid_type", "path": f"{base_path}.{field_name}[{item_index}]", "message": "file/image entry must be an object."})
                    continue
                if _string_or_json(item.get("content")) or _string_or_json(item.get("content_base64")):
                    errors.append(
                        {
                            "code": "inline_binary_rejected",
                            "path": f"{base_path}.{field_name}[{item_index}]",
                            "message": "files/images must reference existing RSI artifacts; inline content is rejected.",
                        }
                    )
        return errors

    def _validate_slack_report_table(self, table: Any, base_path: str) -> list[JsonObject]:
        errors: list[JsonObject] = []
        if not isinstance(table, dict):
            return [{"code": "invalid_type", "path": base_path, "message": "table must be an object."}]
        columns = table.get("columns")
        if not isinstance(columns, list) or not columns:
            errors.append({"code": "required", "path": f"{base_path}.columns", "message": "table columns are required."})
            columns = []
        if len(columns) > 20:
            errors.append({"code": "too_many_columns", "path": f"{base_path}.columns", "limit": 20, "actual": len(columns), "message": "table has too many columns."})
        seen: set[str] = set()
        column_keys: list[str] = []
        for column_index, column in enumerate(columns):
            if not isinstance(column, dict):
                errors.append({"code": "invalid_type", "path": f"{base_path}.columns[{column_index}]", "message": "column must be an object."})
                continue
            key = _string_or_json(column.get("key"))
            if not key:
                errors.append({"code": "required", "path": f"{base_path}.columns[{column_index}].key", "message": "column key is required."})
                continue
            if key in seen:
                errors.append({"code": "duplicate", "path": f"{base_path}.columns[{column_index}].key", "actual": key, "message": "column keys must be unique."})
            seen.add(key)
            column_keys.append(key)
        rows = table.get("rows")
        if rows is not None and not isinstance(rows, list):
            errors.append({"code": "invalid_type", "path": f"{base_path}.rows", "message": "rows must be a list."})
            return errors
        for row_index, row in enumerate(rows or []):
            if not isinstance(row, dict):
                errors.append({"code": "invalid_type", "path": f"{base_path}.rows[{row_index}]", "message": "row must be an object."})
                continue
            for key in column_keys:
                value = row.get(key)
                if not (value is None or isinstance(value, (str, bool, Number))):
                    errors.append({"code": "invalid_cell_type", "path": f"{base_path}.rows[{row_index}].{key}", "message": "table cells must be string, number, bool, or null."})
                    continue
                rendered = _string_or_json(value)
                if len(rendered) > 500:
                    errors.append({"code": "cell_too_long", "path": f"{base_path}.rows[{row_index}].{key}", "limit": 500, "actual": len(rendered), "message": "table cell is too long."})
        return errors

    def _build_action_contract_repair_instruction(
        self,
        task: RunnerTaskRequest,
        structured_output: JsonObject,
        errors: list[JsonObject],
        *,
        attempt: int,
        max_attempts: int,
    ) -> str:
        return "\n".join(
            [
                task.prompt,
                "",
                "Repair instruction: your previous structured output failed the final Slack action contract.",
                f"Repair attempt: {attempt} of {max_attempts}.",
                "Do not re-investigate the task. Do not call read/search/terminal/file tools.",
                "Use the previous structured output below as the source of truth for the answer.",
                "First call exactly one RSI native Slack delivery tool in the bound thread: rsi_slack.message_post for simple prose, or rsi_slack.report_post for rich/tabular output with report_schema_version=1, summary, sections, and structured tables/files/images.",
                "Do not emit slack_post or slack_report proposed_actions; legacy proposed Slack actions are disabled.",
                "After the Slack delivery tool succeeds, re-emit the full JSON object with proposed_actions=[] and preserve the final_answer, reply_draft, produced_artifacts, and artifact_failure_reason unless a correction is required.",
                "Do not put Markdown pipe tables in message_post; convert tabular output into report_post tables with columns [{key,label,align?}] and rows [{column_key: scalar}].",
                "Fix these machine-readable validation errors:",
                json.dumps(errors, ensure_ascii=True, sort_keys=True),
                "Return only valid JSON with no markdown or surrounding commentary.",
                "Previous structured output:",
                json.dumps(structured_output, ensure_ascii=True, sort_keys=True),
            ]
        )


def first_non_empty(*values: str | None) -> str:
    for value in values:
        if value and value.strip():
            return value.strip()
    return ""


def _sign_native_tools_execution_token(secret: str, payload: JsonObject) -> str:
    header = {"alg": "HS256", "typ": "JWT"}
    encoded_header = _base64url_json(header)
    encoded_payload = _base64url_json(payload)
    signed = f"{encoded_header}.{encoded_payload}"
    signature = hmac.HMAC(secret.encode("utf-8"), signed.encode("utf-8"), hashlib.sha256).digest()
    encoded_signature = base64.urlsafe_b64encode(signature).decode("ascii").rstrip("=")
    return f"{signed}.{encoded_signature}"


def _base64url_json(payload: JsonObject) -> str:
    raw = json.dumps(payload, ensure_ascii=True, separators=(",", ":"), sort_keys=True).encode("utf-8")
    return base64.urlsafe_b64encode(raw).decode("ascii").rstrip("=")


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
