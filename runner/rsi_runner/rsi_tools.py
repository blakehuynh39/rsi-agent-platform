from __future__ import annotations

from copy import deepcopy
from dataclasses import dataclass, field
import json
from pathlib import Path
import time
from typing import Any, Iterable, Protocol
from urllib import error as urlerror
from urllib import parse as urlparse
from urllib import request as urlrequest

from .json_types import JsonObject, JsonToolFunctionSchema, JsonToolWrapperSchema
from .observability import ObservationEmitter
from .slack_uploads import prepare_local_slack_upload_payload


READ_ONLY_HONCHO_TOOLS = frozenset({"honcho_profile", "honcho_search", "honcho_context"})
BLOCKED_HONCHO_TOOLS = frozenset({"honcho_conclude"})


_READ_ONLY_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "repo.context": {
        "name": "repo.context",
        "description": "Read-only repository context lookup for the current repo or a named repo.",
        "parameters": {
            "type": "object",
            "properties": {
                "repo": {"type": "string"},
                "question": {"type": "string"},
            },
        },
    },
    "repo.read_file": {
        "name": "repo.read_file",
        "description": "Read a file from the governed repository at a specific ref.",
        "parameters": {
            "type": "object",
            "properties": {
                "repo": {"type": "string"},
                "path": {"type": "string"},
                "ref": {"type": "string"},
            },
            "required": ["path"],
        },
    },
    "repo.search": {
        "name": "repo.search",
        "description": "Search for text in the governed repository using the provider-backed code search surface.",
        "parameters": {
            "type": "object",
            "properties": {
                "repo": {"type": "string"},
                "pattern": {"type": "string"},
                "path": {"type": "string"},
                "ref": {"type": "string"},
            },
            "required": ["pattern"],
        },
    },
    "knowledge.context": {
        "name": "knowledge.context",
        "description": "Read-only RSI knowledge lookup with canonical and working knowledge provenance.",
        "parameters": {
            "type": "object",
            "properties": {
                "topic": {"type": "string"},
                "question": {"type": "string"},
                "scope_id": {"type": "string"},
            },
        },
    },
    "slack.history": {
        "name": "slack.history",
        "description": "Read-only Slack channel or thread history lookup for the bound conversation or an allowed channel.",
        "parameters": {
            "type": "object",
            "properties": {
                "channel_id": {"type": "string"},
                "thread_ts": {"type": "string"},
                "scope": {"type": "string"},
                "question": {"type": "string"},
                "oldest": {"type": "string"},
                "latest": {"type": "string"},
                "limit": {"type": "integer"},
            },
        },
    },
    "slack.search": {
        "name": "slack.search",
        "description": "Read-only Slack message search bounded to allowed channels and a time window.",
        "parameters": {
            "type": "object",
            "properties": {
                "query": {"type": "string"},
                "channel_ids": {"type": "array", "items": {"type": "string"}},
                "since": {"type": "string"},
                "until": {"type": "string"},
                "limit": {"type": "integer"},
            },
        },
    },
    "slack.upload_file": {
        "name": "slack.upload_file",
        "description": "Upload a generated file into the bound Slack thread or an allowed channel using inline text, base64 content, or a local runner file path / file:// artifact reference.",
        "parameters": {
            "type": "object",
            "properties": {
                "channel_id": {"type": "string"},
                "thread_ts": {"type": "string"},
                "filename": {"type": "string"},
                "title": {"type": "string"},
                "content": {"type": "string"},
                "content_base64": {"type": "string"},
                "path": {"type": "string"},
                "artifact_ref": {"type": "string"},
                "initial_comment": {"type": "string"},
                "alt_txt": {"type": "string"},
                "snippet_type": {"type": "string"},
            },
        },
    },
    "github.repo_activity": {
        "name": "github.repo_activity",
        "description": "Read-only GitHub activity window lookup for commits and pull requests.",
        "parameters": {
            "type": "object",
            "properties": {
                "repo": {"type": "string"},
                "since": {"type": "string"},
                "until": {"type": "string"},
            },
        },
    },
    "github.repo_context": {
        "name": "github.repo_context",
        "description": "Read-only GitHub repository metadata and default branch context.",
        "parameters": {
            "type": "object",
            "properties": {
                "repo": {"type": "string"},
            },
        },
    },
}

_READ_ONLY_TOOL_SCHEMAS.update({
    "sentry.lookup": {
        "name": "sentry.lookup",
        "description": "Read-only Sentry issue lookup for a service, alert, or query.",
        "parameters": {
            "type": "object",
            "properties": {
                "service": {"type": "string"},
                "alert": {"type": "string"},
                "query": {"type": "string"},
            },
        },
    },
    "kubernetes.inspect": {
        "name": "kubernetes.inspect",
        "description": "Read-only Kubernetes pod and event inspection within the configured namespace scope. Omit namespace to search all allowed namespaces.",
        "parameters": {
            "type": "object",
            "properties": {
                "namespace": {"type": "string"},
                "target": {"type": "string"},
                "service": {"type": "string"},
            },
        },
    },
    "kubernetes.logs": {
        "name": "kubernetes.logs",
        "description": "Read-only Kubernetes pod log lookup within the configured namespace scope. Omit namespace to search all allowed namespaces.",
        "parameters": {
            "type": "object",
            "properties": {
                "namespace": {"type": "string"},
                "target": {"type": "string"},
                "pod_name": {"type": "string"},
                "container": {"type": "string"},
            },
        },
    },
    "kubernetes.events": {
        "name": "kubernetes.events",
        "description": "Read-only Kubernetes event lookup within the configured namespace scope. Omit namespace to search all allowed namespaces.",
        "parameters": {
            "type": "object",
            "properties": {
                "namespace": {"type": "string"},
                "target": {"type": "string"},
            },
        },
    },
    "cloudflare.inspect": {
        "name": "cloudflare.inspect",
        "description": "Read-only Cloudflare inspection for zones or account-scoped resources.",
        "parameters": {
            "type": "object",
            "properties": {
                "resource": {"type": "string"},
            },
        },
    },
    "rsi.trace_context": {
        "name": "rsi.trace_context",
        "description": "Read-only RSI evidence lookup for a trace, including events, reasoning, tools, evals, and linked proposals.",
        "parameters": {
            "type": "object",
            "properties": {
                "trace_id": {"type": "string"},
            },
        },
    },
    "rsi.workflow_context": {
        "name": "rsi.workflow_context",
        "description": "Read-only RSI workflow context, including workflow state, trace summary, and recent conversation entries.",
        "parameters": {
            "type": "object",
            "properties": {
                "workflow_id": {"type": "string"},
                "trace_id": {"type": "string"},
            },
        },
    },
    "rsi.action_chain": {
        "name": "rsi.action_chain",
        "description": "Read-only RSI action chain lookup for intents, results, and outcomes related to a trace, proposal, or attempt.",
        "parameters": {
            "type": "object",
            "properties": {
                "trace_id": {"type": "string"},
                "proposal_id": {"type": "string"},
                "attempt_id": {"type": "string"},
            },
        },
    },
    "rsi.runner_execution": {
        "name": "rsi.runner_execution",
        "description": "Read-only RSI harness execution lookup for workflow, eval, or proposal runs.",
        "parameters": {
            "type": "object",
            "properties": {
                "trace_id": {"type": "string"},
                "proposal_id": {"type": "string"},
                "role": {"type": "string"},
            },
        },
    },
    "rsi.runtime_config": {
        "name": "rsi.runtime_config",
        "description": "Read-only RSI runtime configuration summary without secrets.",
        "parameters": {
            "type": "object",
            "properties": {},
        },
    },
    "rsi.runtime_health": {
        "name": "rsi.runtime_health",
        "description": "Read-only RSI runtime health summary for runners and Honcho.",
        "parameters": {
            "type": "object",
            "properties": {},
        },
    },
    "rsi.runtime_deployment_facts": {
        "name": "rsi.runtime_deployment_facts",
        "description": "Read-only RSI deployment and runtime-configuration facts, including live Kubernetes deployment state across the configured namespace scope when available.",
        "parameters": {
            "type": "object",
            "properties": {
                "namespace": {"type": "string"},
                "services": {"type": "array", "items": {"type": "string"}},
                "service": {"type": "string"},
            },
        },
    },
    "rsi.proposal_memory": {
        "name": "rsi.proposal_memory",
        "description": "Read-only RSI proposal-memory lookup for a candidate line or proposal.",
        "parameters": {
            "type": "object",
            "properties": {
                "candidate_key": {"type": "string"},
                "proposal_id": {"type": "string"},
            },
        },
    },
    "rsi.candidate_context": {
        "name": "rsi.candidate_context",
        "description": "Read-only RSI improvement-candidate context, including linked proposals and memory.",
        "parameters": {
            "type": "object",
            "properties": {
                "candidate_key": {"type": "string"},
            },
        },
    },
    "rsi.attempt_context": {
        "name": "rsi.attempt_context",
        "description": "Read-only RSI change-attempt context, including validation, PR, action, and outcome state.",
        "parameters": {
            "type": "object",
            "properties": {
                "attempt_id": {"type": "string"},
            },
        },
    },
})


def _string_list(value: Any) -> list[str]:
    if not isinstance(value, list):
        return []
    return [str(item).strip() for item in value if str(item).strip()]


def _string_value(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value.strip()
    return str(value).strip()


def _truncate_text(value: Any, limit: int) -> str:
    text = _string_value(value)
    if len(text) <= limit:
        return text
    return text[: max(0, limit - 1)] + "…"


def first_non_empty(*values: Any) -> str:
    for value in values:
        text = _string_value(value)
        if text:
            return text
    return ""


def _set_non_empty_text(target: JsonObject, key: str, value: Any) -> None:
    text = _string_value(value)
    if text:
        target[key] = text


def _iso_timestamp(value: float) -> str:
    if value <= 0:
        return ""
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(value))

_WORKSPACE_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "workspace.list_files": {
        "name": "workspace.list_files",
        "description": "List files inside the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "path": {"type": "string"},
            },
        },
    },
    "workspace.read_file": {
        "name": "workspace.read_file",
        "description": "Read a file from the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "path": {"type": "string"},
            },
            "required": ["path"],
        },
    },
    "workspace.search": {
        "name": "workspace.search",
        "description": "Search for text within the governed attempt workspace using ripgrep.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "path": {"type": "string"},
                "pattern": {"type": "string"},
            },
            "required": ["pattern"],
        },
    },
    "workspace.git_history": {
        "name": "workspace.git_history",
        "description": "Inspect commit history inside the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "ref": {"type": "string"},
                "path": {"type": "string"},
                "limit": {"type": "integer"},
            },
        },
    },
    "workspace.git_show": {
        "name": "workspace.git_show",
        "description": "Inspect a historical commit or file revision inside the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "ref": {"type": "string"},
                "path": {"type": "string"},
            },
        },
    },
    "workspace.git_search": {
        "name": "workspace.git_search",
        "description": "Search workspace git history for commit messages or content changes.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "ref": {"type": "string"},
                "path": {"type": "string"},
                "pattern": {"type": "string"},
                "search_type": {"type": "string"},
                "limit": {"type": "integer"},
            },
            "required": ["pattern"],
        },
    },
    "workspace.write_file": {
        "name": "workspace.write_file",
        "description": "Write file content inside the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "path": {"type": "string"},
                "content": {"type": "string"},
            },
            "required": ["path", "content"],
        },
    },
    "workspace.apply_patch": {
        "name": "workspace.apply_patch",
        "description": "Apply a unified diff patch inside the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "patch": {"type": "string"},
            },
            "required": ["patch"],
        },
    },
    "workspace.git_status": {
        "name": "workspace.git_status",
        "description": "Inspect git status inside the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
            },
        },
    },
    "workspace.git_diff": {
        "name": "workspace.git_diff",
        "description": "Inspect the current git diff inside the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
            },
        },
    },
    "workspace.run_validation": {
        "name": "workspace.run_validation",
        "description": "Run bounded validation inside the governed attempt workspace.",
        "parameters": {
            "type": "object",
            "properties": {
                "workspace_id": {"type": "string"},
                "attempt_id": {"type": "string"},
                "command": {"type": "string"},
            },
        },
    },
}

_TOOL_SCHEMAS = {**_READ_ONLY_TOOL_SCHEMAS, **_WORKSPACE_TOOL_SCHEMAS}

_TRANSPORT_SAFE_TOOL_CHARS = frozenset("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-")

READ_ONLY_RSI_TOOL_NAMES = tuple(sorted(_READ_ONLY_TOOL_SCHEMAS.keys()))
READ_ONLY_WORKSPACE_RSI_TOOL_NAMES = (
    "workspace.git_history",
    "workspace.git_search",
    "workspace.git_show",
    "workspace.list_files",
    "workspace.read_file",
    "workspace.search",
)
WORKSPACE_RSI_TOOL_NAMES = tuple(sorted(_WORKSPACE_TOOL_SCHEMAS.keys()))
IMPLEMENT_RSI_TOOL_NAMES = tuple(sorted(_TOOL_SCHEMAS.keys()))


def _is_transport_safe_tool_name(name: str) -> bool:
    return bool(name) and all(char in _TRANSPORT_SAFE_TOOL_CHARS for char in name)


def _canonical_to_transport_tool_name(name: str) -> str:
    canonical = str(name or "").strip()
    if not canonical:
        raise ValueError("tool name is empty")
    transport = canonical.replace(".", "_")
    if not _is_transport_safe_tool_name(transport):
        raise ValueError(f"tool name {canonical!r} cannot be mapped to an OpenAI-safe transport name")
    return transport


def _build_tool_transport_maps() -> tuple[dict[str, str], dict[str, str]]:
    canonical_to_transport: dict[str, str] = {}
    transport_to_canonical: dict[str, str] = {}
    for canonical in sorted(_TOOL_SCHEMAS):
        transport = _canonical_to_transport_tool_name(canonical)
        existing = transport_to_canonical.get(transport)
        if existing is not None and existing != canonical:
            raise ValueError(f"transport tool name collision for {canonical!r} and {existing!r}: {transport!r}")
        canonical_to_transport[canonical] = transport
        transport_to_canonical[transport] = canonical
    return canonical_to_transport, transport_to_canonical


_CANONICAL_TO_TRANSPORT_TOOL_NAMES, _TRANSPORT_TO_CANONICAL_TOOL_NAMES = _build_tool_transport_maps()

HERMES_GOVERNED_READONLY_TOOLSET = "rsi-governed-readonly"
HERMES_GOVERNED_WORKSPACE_TOOLSET = "rsi-governed-workspace"
HERMES_ARTIFACT_TOOLSET = "rsi-artifacts"

_ARTIFACT_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "artifact.list_files": {
        "name": "artifact.list_files",
        "description": "List files inside the staged Hermes artifact output directory.",
        "parameters": {
            "type": "object",
            "properties": {
                "path": {"type": "string"},
            },
        },
    },
    "artifact.write_file": {
        "name": "artifact.write_file",
        "description": "Write file content inside the staged Hermes artifact output directory.",
        "parameters": {
            "type": "object",
            "properties": {
                "path": {"type": "string"},
                "content": {"type": "string"},
            },
            "required": ["path", "content"],
        },
    },
}


def tool_transport_name(name: str) -> str:
    canonical = str(name or "").strip()
    if not canonical:
        raise ValueError("tool name is empty")
    if canonical in _CANONICAL_TO_TRANSPORT_TOOL_NAMES:
        return _CANONICAL_TO_TRANSPORT_TOOL_NAMES[canonical]
    if "." not in canonical and _is_transport_safe_tool_name(canonical):
        return canonical
    raise ValueError(f"tool name {canonical!r} is not transport-safe")


def canonical_tool_name(name: str) -> str:
    tool = str(name or "").strip()
    if not tool:
        raise ValueError("tool name is empty")
    if tool in _TOOL_SCHEMAS:
        return tool
    if tool in _TRANSPORT_TO_CANONICAL_TOOL_NAMES:
        return _TRANSPORT_TO_CANONICAL_TOOL_NAMES[tool]
    if "." not in tool and _is_transport_safe_tool_name(tool):
        return tool
    raise ValueError(f"tool name {tool!r} is not recognized")


def _nullable_json_schema(value: Any) -> Any:
    if not isinstance(value, dict):
        return value
    out = deepcopy(value)
    schema_type = out.get("type")
    if isinstance(schema_type, str):
        if schema_type != "null":
            out["type"] = [schema_type, "null"]
        return out
    if isinstance(schema_type, list):
        normalized = [item for item in schema_type if isinstance(item, str)]
        if "null" not in normalized:
            out["type"] = [*schema_type, "null"]
        return out
    enum_values = out.get("enum")
    if isinstance(enum_values, list) and None not in enum_values:
        out["enum"] = [*enum_values, None]
    return out


def _strict_json_schema(value: Any) -> Any:
    if isinstance(value, dict):
        out = {key: _strict_json_schema(item) for key, item in value.items()}
        if out.get("type") == "object":
            properties = out.get("properties")
            if isinstance(properties, dict):
                existing_required = {
                    str(item).strip()
                    for item in out.get("required", [])
                    if isinstance(item, str) and str(item).strip()
                }
                ordered_keys = list(properties.keys())
                for key in ordered_keys:
                    if key not in existing_required:
                        properties[key] = _nullable_json_schema(properties[key])
                out["required"] = ordered_keys
            else:
                out["required"] = []
            out["additionalProperties"] = False
        return out
    if isinstance(value, list):
        return [_strict_json_schema(item) for item in value]
    return value


def tool_schema_wrappers(names: Iterable[str]) -> list[JsonToolWrapperSchema]:
    wrappers: list[JsonToolWrapperSchema] = []
    for name in names:
        schema = _TOOL_SCHEMAS.get(name)
        if schema is None:
            continue
        wrapped = deepcopy(schema)
        wrapped["name"] = tool_transport_name(name)
        if "parameters" in wrapped:
            wrapped["parameters"] = _strict_json_schema(wrapped.get("parameters"))
        wrappers.append({"type": "function", "function": wrapped})
    return wrappers


def transport_tool_schema(name: str) -> JsonToolFunctionSchema:
    schema = _TOOL_SCHEMAS.get(name)
    if schema is None:
        raise KeyError(name)
    wrapped = deepcopy(schema)
    wrapped["name"] = tool_transport_name(name)
    if "parameters" in wrapped:
        wrapped["parameters"] = _strict_json_schema(wrapped.get("parameters"))
    return wrapped


def governed_toolset_names() -> dict[str, list[str]]:
    return {
        HERMES_GOVERNED_READONLY_TOOLSET: list(READ_ONLY_RSI_TOOL_NAMES),
        HERMES_GOVERNED_WORKSPACE_TOOLSET: list(WORKSPACE_RSI_TOOL_NAMES),
    }


def governed_toolset_definitions() -> list[JsonObject]:
    definitions: list[JsonObject] = []
    for toolset, names in governed_toolset_names().items():
        for canonical_name in names:
            definitions.append(
                {
                    "canonical_name": canonical_name,
                    "transport_name": tool_transport_name(canonical_name),
                    "toolset": toolset,
                    "schema": transport_tool_schema(canonical_name),
                }
            )
    return definitions


def _artifact_toolset_definitions() -> list[JsonObject]:
    definitions: list[JsonObject] = []
    for canonical_name, schema in _ARTIFACT_TOOL_SCHEMAS.items():
        transport_name = _canonical_to_transport_tool_name(canonical_name)
        wrapped = deepcopy(schema)
        wrapped["name"] = transport_name
        if "parameters" in wrapped:
            wrapped["parameters"] = _strict_json_schema(wrapped.get("parameters"))
        definitions.append(
            {
                "canonical_name": canonical_name,
                "transport_name": transport_name,
                "toolset": HERMES_ARTIFACT_TOOLSET,
                "schema": wrapped,
            }
        )
    return definitions


def rsi_plugin_toolset_definitions() -> list[JsonObject]:
    return [*governed_toolset_definitions(), *_artifact_toolset_definitions()]


def normalize_tool_names(values: Iterable[str]) -> list[str]:
    seen: set[str] = set()
    out: list[str] = []
    for value in values:
        name = str(value or "").strip()
        if not name or name in seen:
            continue
        seen.add(name)
        out.append(name)
    return out


class ToolManagerLike(Protocol):
    def has_tool(self, name: str) -> bool:
        ...

    def handle_tool_call(self, name: str, args: JsonObject, **kwargs: Any) -> str:
        ...

    def build_system_prompt(self) -> str:
        ...


@dataclass
class ReadOnlyToolBinding:
    base_url: str
    allowed_tool_names: list[str]
    task_repo: str
    task_repo_ref: str
    task_prompt: str
    task_channel_id: str
    task_thread_ts: str
    task_context_summary: str
    trace_id: str
    session_scope_kind: str
    session_scope_id: str
    context_refs: list[JsonObject]
    default_question: str = ""
    repo_question: str = ""
    knowledge_topic: str = ""
    knowledge_question: str = ""
    slack_history_focus: str = ""
    slack_search_query: str = ""
    execution_mode: str = ""
    execution_phase: str = ""
    attempt_id: str = ""
    workspace_id: str = ""
    kubernetes_read_namespaces: list[str] = field(default_factory=list)
    timeout_seconds: int = 30
    selected_context_surfaces: list[JsonObject] = field(default_factory=list)
    tool_calls: list[JsonObject] = field(default_factory=list)
    evidence_items: list[JsonObject] = field(default_factory=list)
    observer: ObservationEmitter | None = None
    _tool_call_counter: int = 0

    def has_tool(self, name: str) -> bool:
        try:
            canonical = canonical_tool_name(name)
        except ValueError:
            return False
        return canonical in _TOOL_SCHEMAS and canonical in set(self.allowed_tool_names)

    def tool_names(self) -> list[str]:
        out: list[str] = []
        for name in normalize_tool_names(self.allowed_tool_names):
            out.append(tool_transport_name(name))
        return out

    def _observation_phase(self) -> str:
        phase = (self.execution_phase or "").strip().lower()
        if phase:
            return phase
        return "render" if self.execution_mode == "artifact_render" else "investigate"

    def handle_tool_call(self, name: str, args: JsonObject) -> str:
        transport_name = tool_transport_name(name)
        canonical_name = canonical_tool_name(name)
        payload = self._default_payload(canonical_name)
        payload.update({key: value for key, value in (args or {}).items() if value is not None})
        tool_call_id = self._next_tool_call_id(canonical_name)
        call_started_at = time.time()
        if self.observer is not None:
            self.observer.emit(
                phase=self._observation_phase(),
                event_type="tool.call.started",
                status="running",
                payload={
                    "tool_name": canonical_name,
                    "tool_call_id": tool_call_id,
                },
            )
        try:
            payload = self._prepare_tool_payload(canonical_name, payload)
        except (OSError, ValueError) as exc:
            summary = str(exc).strip() or f"{canonical_name} payload preparation failed"
            self._record_tool_call(
                canonical_name=canonical_name,
                tool_call_id=tool_call_id,
                request_payload=payload,
                started_at=call_started_at,
                completed_at=time.time(),
                status="failed",
                summary=summary,
            )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "error",
                    "error": summary,
                }
            )
        self._record_tool_selection(canonical_name, payload)
        req = urlrequest.Request(
            f"{self.base_url.rstrip('/')}/api/tools/{canonical_name}/execute",
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            with urlrequest.urlopen(req, timeout=self.timeout_seconds) as resp:
                body = resp.read().decode("utf-8")
        except urlerror.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")
            summary = f"tool gateway returned {exc.code}: {detail}"
            self._record_tool_call(
                canonical_name=canonical_name,
                tool_call_id=tool_call_id,
                request_payload=payload,
                started_at=call_started_at,
                completed_at=time.time(),
                status="failed",
                summary=summary,
            )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "error",
                    "error": summary,
                }
            )
        except (urlerror.URLError, TimeoutError, ConnectionError, OSError) as exc:
            summary = f"tool gateway request failed: {exc}"
            self._record_tool_call(
                canonical_name=canonical_name,
                tool_call_id=tool_call_id,
                request_payload=payload,
                started_at=call_started_at,
                completed_at=time.time(),
                status="failed",
                summary=summary,
            )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "error",
                    "error": summary,
                }
            )

        try:
            parsed = json.loads(body)
        except json.JSONDecodeError:
            self._record_tool_call(
                canonical_name=canonical_name,
                tool_call_id=tool_call_id,
                request_payload=payload,
                started_at=call_started_at,
                completed_at=time.time(),
                status="failed",
                summary="tool gateway returned invalid JSON",
            )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "error",
                    "error": "tool gateway returned invalid JSON",
                    "body": body[:8000],
                }
            )
        if not isinstance(parsed, dict):
            self._record_tool_call(
                canonical_name=canonical_name,
                tool_call_id=tool_call_id,
                request_payload=payload,
                started_at=call_started_at,
                completed_at=time.time(),
                status="failed",
                summary="tool gateway returned a non-object JSON payload",
            )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "error",
                    "error": "tool gateway returned a non-object JSON payload",
                }
            )
        status = str(parsed.get("status", "") or "completed").strip()
        summary = str(parsed.get("summary", "") or status or canonical_name).strip()
        provider_ref = str(parsed.get("provider_ref", "") or "").strip()
        approval_state = str(parsed.get("approval_state", "") or "").strip()
        raw_artifact_refs = parsed.get("raw_artifact_refs", [])
        self._record_tool_call(
            canonical_name=canonical_name,
            tool_call_id=tool_call_id,
            request_payload=payload,
            started_at=call_started_at,
            completed_at=time.time(),
            status=status or "completed",
            summary=summary,
            provider_ref=provider_ref,
            approval_state=approval_state,
            raw_artifact_refs=raw_artifact_refs,
            output_payload=parsed.get("output", {}),
        )
        return json.dumps(
            {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "status": status,
                "available": parsed.get("available", True),
                "summary": summary,
                "provider": parsed.get("provider", ""),
                "provider_ref": provider_ref,
                "approval_state": approval_state,
                "output": parsed.get("output", {}),
                "raw_artifact_refs": raw_artifact_refs,
                "tool_call_id": tool_call_id,
            }
        )

    def _prepare_tool_payload(self, canonical_name: str, payload: JsonObject) -> JsonObject:
        if canonical_name != "slack.upload_file":
            return payload
        return self._resolve_slack_upload_payload(payload)

    def _resolve_slack_upload_payload(self, payload: JsonObject) -> JsonObject:
        if _string_value(payload.get("content")) or _string_value(payload.get("content_base64")):
            return payload
        resolved_path = self._resolve_local_upload_path(payload)
        if resolved_path is None:
            return payload
        return prepare_local_slack_upload_payload(payload, resolved_path)

    def _resolve_local_upload_path(self, payload: JsonObject) -> Path | None:
        raw_path = first_non_empty(_string_value(payload.get("path")), _string_value(payload.get("artifact_ref")))
        if not raw_path:
            return None
        candidate = raw_path.strip()
        if not candidate:
            return None
        if "://" in candidate:
            parsed = urlparse.urlparse(candidate)
            if parsed.scheme not in {"file", "hermes-file"}:
                return None
            path_text = urlparse.unquote(parsed.path or "")
        else:
            path_text = candidate
        if not path_text:
            raise ValueError("slack.upload_file file artifact_ref is missing a path")
        path = Path(path_text).expanduser()
        if not path.is_absolute():
            path = (Path.cwd() / path).resolve()
        else:
            path = path.resolve()
        if not path.exists():
            raise ValueError(f"slack.upload_file local file does not exist: {path}")
        if not path.is_file():
            raise ValueError(f"slack.upload_file local path is not a file: {path}")
        return path

    def diagnostics(self) -> JsonObject:
        return {
            "candidate_read_surfaces": self._candidate_read_surfaces(),
            "selected_context_surfaces": list(self.selected_context_surfaces),
            "tool_calls": deepcopy(self.tool_calls),
            "evidence_items": deepcopy(self.evidence_items),
        }

    def _next_tool_call_id(self, canonical_name: str) -> str:
        self._tool_call_counter += 1
        return f"{canonical_name}:{self._tool_call_counter}"

    def _record_tool_call(
        self,
        *,
        canonical_name: str,
        tool_call_id: str,
        request_payload: JsonObject,
        started_at: float,
        completed_at: float,
        status: str,
        summary: str,
        provider_ref: str = "",
        approval_state: str = "",
        raw_artifact_refs: Any = None,
        output_payload: Any = None,
    ) -> None:
        self.tool_calls.append(
            {
                "id": f"runner-tool-record-{tool_call_id}",
                "tool_name": canonical_name,
                "tool_call_id": tool_call_id,
                "request": deepcopy(request_payload),
                "summary": (summary or "").strip(),
                "raw_artifact_refs": _string_list(raw_artifact_refs),
                "approval_state": approval_state,
                "interpretation_summary": (summary or "").strip(),
                "status": (status or "").strip(),
                "provider_ref": provider_ref,
                "started_at": _iso_timestamp(started_at),
                "completed_at": _iso_timestamp(completed_at),
                "created_at": _iso_timestamp(completed_at or started_at),
            }
        )
        if self.observer is not None:
            self.observer.emit(
                phase=self._observation_phase(),
                event_type="tool.call.completed",
                status=(status or "").strip(),
                payload={
                    "tool_name": canonical_name,
                    "tool_call_id": tool_call_id,
                    "summary": (summary or "").strip(),
                    "provider_ref": provider_ref,
                    "approval_state": approval_state,
                    "artifact_refs": _string_list(raw_artifact_refs),
                },
            )
        self.evidence_items.extend(
            self._extract_evidence_items(
                canonical_name=canonical_name,
                tool_call_id=tool_call_id,
                request_payload=request_payload,
                output_payload=output_payload,
                summary=summary,
                provider_ref=provider_ref,
            )
        )

    def _extract_evidence_items(
        self,
        *,
        canonical_name: str,
        tool_call_id: str,
        request_payload: JsonObject,
        output_payload: Any,
        summary: str,
        provider_ref: str,
    ) -> list[JsonObject]:
        output = output_payload if isinstance(output_payload, dict) else {}
        extractors = {
            "slack.history": self._slack_history_evidence_items,
            "slack.search": self._slack_search_evidence_items,
            "repo.context": self._repo_context_evidence_items,
            "repo.search": self._repo_search_evidence_items,
            "repo.read_file": self._repo_read_file_evidence_items,
            "github.repo_activity": self._github_repo_activity_evidence_items,
            "github.repo_context": self._github_repo_context_evidence_items,
            "rsi.workflow_context": self._workflow_context_evidence_items,
        }
        extractor = extractors.get(canonical_name)
        items = extractor(request_payload, output) if extractor is not None else []
        if items:
            return items[:6]
        fallback_summary = _truncate_text(summary, 600)
        if not fallback_summary:
            return []
        source_ref = provider_ref or _string_value(request_payload.get("path")) or _string_value(request_payload.get("channel_id"))
        fallback: JsonObject = {
            "kind": "tool_summary",
            "summary": fallback_summary,
            "source_ref": source_ref,
            "tool_name": canonical_name,
            "tool_call_id": tool_call_id,
        }
        return [fallback]

    def _slack_history_evidence_items(self, request_payload: JsonObject, output: JsonObject) -> list[JsonObject]:
        items: list[JsonObject] = []
        channel_id = _string_value(output.get("channel_id")) or _string_value(request_payload.get("channel_id"))
        thread_ts = _string_value(output.get("thread_ts")) or _string_value(request_payload.get("thread_ts"))
        messages = output.get("messages")
        if not isinstance(messages, list):
            return items
        for message in messages[:6]:
            if not isinstance(message, dict):
                continue
            text = _truncate_text(
                first_non_empty(
                    _string_value(message.get("text")),
                    _string_value(message.get("content")),
                    _string_value(message.get("body")),
                ),
                500,
            )
            if not text:
                continue
            ts = _string_value(message.get("ts")) or _string_value(message.get("message_timestamp"))
            source_ref = _string_value(message.get("permalink"))
            if not source_ref and channel_id and ts:
                source_ref = f"slack://{channel_id}/{ts}"
            author = first_non_empty(
                _string_value(message.get("author_name")),
                _string_value(message.get("user")),
                _string_value(message.get("user_id")),
                _string_value(message.get("username")),
            )
            prefix = f"[{author}] " if author else ""
            item: JsonObject = {
                "kind": "slack_message",
                "summary": prefix + text,
                "snippet": text,
                "source_ref": source_ref,
                "tool_name": "slack.history",
                "channel_id": channel_id,
                "thread_ts": _string_value(message.get("thread_ts")) or thread_ts or ts,
                "message_ts": ts,
                "permalink": _string_value(message.get("permalink")),
            }
            _set_non_empty_text(item, "author", author)
            items.append(item)
        return items

    def _slack_search_evidence_items(self, _request_payload: JsonObject, output: JsonObject) -> list[JsonObject]:
        items: list[JsonObject] = []
        messages = output.get("messages")
        if not isinstance(messages, list):
            return items
        for message in messages[:6]:
            if not isinstance(message, dict):
                continue
            text = _truncate_text(first_non_empty(_string_value(message.get("text")), _string_value(message.get("content"))), 500)
            if not text:
                continue
            channel_id = _string_value(message.get("channel_id"))
            ts = _string_value(message.get("ts")) or _string_value(message.get("message_timestamp"))
            source_ref = _string_value(message.get("permalink"))
            if not source_ref and channel_id and ts:
                source_ref = f"slack://{channel_id}/{ts}"
            author = first_non_empty(_string_value(message.get("author_name")), _string_value(message.get("author_user_id")))
            prefix = f"[{author}] " if author else ""
            item: JsonObject = {
                "kind": "slack_search_match",
                "summary": prefix + text,
                "snippet": text,
                "source_ref": source_ref,
                "tool_name": "slack.search",
                "channel_id": channel_id,
                "thread_ts": _string_value(message.get("thread_ts")) or ts,
                "message_ts": ts,
                "permalink": _string_value(message.get("permalink")),
            }
            _set_non_empty_text(item, "author", author)
            items.append(item)
        return items

    def _repo_context_evidence_items(self, request_payload: JsonObject, output: JsonObject) -> list[JsonObject]:
        repo = _string_value(output.get("repo")) or _string_value(request_payload.get("repo"))
        default_branch = _string_value(output.get("default_branch"))
        description = _truncate_text(output.get("description"), 400)
        matches = output.get("matches")
        items: list[JsonObject] = []
        if description:
            summary = f"Repository context for {repo}" if repo else "Repository context"
            if default_branch:
                summary += f" (default branch {default_branch})"
            summary += f": {description}"
            item: JsonObject = {
                "kind": "repo_context",
                "summary": _truncate_text(summary, 700),
                "snippet": description,
                "source_ref": first_non_empty(_string_value(output.get("html_url")), repo),
                "tool_name": "repo.context",
                "repo": repo,
                "default_branch": default_branch,
            }
            items.append(item)
        if not isinstance(matches, list):
            return items
        for match in matches[:4]:
            if not isinstance(match, dict):
                continue
            path = _string_value(match.get("path"))
            snippet = _truncate_text(match.get("snippet"), 500)
            summary = first_non_empty(snippet, f"Repository context match in {path}")
            item = {
                "kind": "repo_context_match",
                "summary": summary,
                "snippet": snippet,
                "source_ref": first_non_empty(_string_value(match.get("html_url")), path),
                "tool_name": "repo.context",
                "path": path,
                "repo": repo,
            }
            items.append(item)
        return items

    def _repo_search_evidence_items(self, request_payload: JsonObject, output: JsonObject) -> list[JsonObject]:
        items: list[JsonObject] = []
        matches = output.get("matches")
        repo = _string_value(output.get("repo")) or _string_value(request_payload.get("repo"))
        pattern = _string_value(output.get("pattern")) or _string_value(request_payload.get("pattern"))
        if not isinstance(matches, list):
            return items
        for match in matches[:6]:
            if not isinstance(match, dict):
                continue
            path = _string_value(match.get("path"))
            snippet = _truncate_text(match.get("snippet"), 500)
            summary = first_non_empty(snippet, f"Search match for {pattern} in {path}")
            items.append(
                {
                    "kind": "repo_search_match",
                    "summary": summary,
                    "snippet": snippet,
                    "source_ref": first_non_empty(_string_value(match.get("html_url")), path),
                    "tool_name": "repo.search",
                    "path": path,
                    "repo": repo,
                }
            )
        return items

    def _repo_read_file_evidence_items(self, request_payload: JsonObject, output: JsonObject) -> list[JsonObject]:
        path = _string_value(output.get("path")) or _string_value(request_payload.get("path"))
        repo = _string_value(output.get("repo")) or _string_value(request_payload.get("repo"))
        ref = _string_value(output.get("ref")) or _string_value(request_payload.get("ref"))
        excerpt = _truncate_text(output.get("content"), 900)
        if not path or not excerpt:
            return []
        return [
            {
                "kind": "repo_file_excerpt",
                "summary": f"{path} ({first_non_empty(ref, 'default branch')}): {excerpt}",
                "snippet": excerpt,
                "source_ref": path,
                "tool_name": "repo.read_file",
                "path": path,
                "repo": repo,
                "ref": ref,
            }
        ]

    def _github_repo_activity_evidence_items(self, request_payload: JsonObject, output: JsonObject) -> list[JsonObject]:
        repo = _string_value(output.get("repo")) or _string_value(request_payload.get("repo"))
        commits = output.get("commits")
        merged = output.get("merged_pull_requests")
        opened = output.get("opened_pull_requests")
        items: list[JsonObject] = []
        if isinstance(commits, list):
            for commit in commits[:2]:
                if not isinstance(commit, dict):
                    continue
                title = _truncate_text(
                    first_non_empty(
                        _string_value(commit.get("message")),
                        _string_value(commit.get("title")),
                        _string_value(commit.get("sha")),
                    ),
                    320,
                )
                if title:
                    item: JsonObject = {
                        "kind": "github_commit",
                        "summary": title,
                        "source_ref": first_non_empty(_string_value(commit.get("url")), repo),
                        "tool_name": "github.repo_activity",
                        "repo": repo,
                        "title": title,
                    }
                    _set_non_empty_text(item, "sha", commit.get("sha"))
                    _set_non_empty_text(item, "author", commit.get("author"))
                    _set_non_empty_text(item, "committed_at", commit.get("committed_at"))
                    _set_non_empty_text(item, "url", commit.get("url"))
                    items.append(item)
        if isinstance(merged, list):
            for pull in merged[:2]:
                if not isinstance(pull, dict):
                    continue
                title = _truncate_text(first_non_empty(_string_value(pull.get("title")), _string_value(pull.get("url"))), 320)
                if title:
                    item = {
                        "kind": "github_pull_request",
                        "summary": f"Merged PR: {title}",
                        "source_ref": first_non_empty(_string_value(pull.get("url")), repo),
                        "tool_name": "github.repo_activity",
                        "repo": repo,
                        "title": title,
                        "state": "merged",
                    }
                    _set_non_empty_text(item, "author", pull.get("author"))
                    _set_non_empty_text(item, "created_at", pull.get("created_at"))
                    _set_non_empty_text(item, "merged_at", pull.get("merged_at"))
                    _set_non_empty_text(item, "url", pull.get("url"))
                    items.append(item)
        if isinstance(opened, list):
            for pull in opened[:1]:
                if not isinstance(pull, dict):
                    continue
                title = _truncate_text(first_non_empty(_string_value(pull.get("title")), _string_value(pull.get("url"))), 320)
                if title:
                    item = {
                        "kind": "github_pull_request",
                        "summary": f"Opened PR: {title}",
                        "source_ref": first_non_empty(_string_value(pull.get("url")), repo),
                        "tool_name": "github.repo_activity",
                        "repo": repo,
                        "title": title,
                        "state": first_non_empty(_string_value(pull.get("state")), "open"),
                    }
                    _set_non_empty_text(item, "author", pull.get("author"))
                    _set_non_empty_text(item, "created_at", pull.get("created_at"))
                    _set_non_empty_text(item, "url", pull.get("url"))
                    items.append(item)
        if items:
            return items
        summary = _truncate_text(output.get("summary"), 900)
        if not summary:
            return []
        return [{"kind": "github_activity_summary", "summary": summary, "source_ref": repo, "tool_name": "github.repo_activity", "repo": repo}]

    def _github_repo_context_evidence_items(self, request_payload: JsonObject, output: JsonObject) -> list[JsonObject]:
        repo = _string_value(request_payload.get("repo")) or _string_value(output.get("name"))
        default_branch = _string_value(output.get("default_branch"))
        description = _truncate_text(output.get("description"), 400)
        summary = f"Repository context for {repo}"
        if default_branch:
            summary += f" (default branch {default_branch})"
        if description:
            summary += f": {description}"
        return [
            {
                "kind": "github_repo_context",
                "summary": _truncate_text(summary, 700),
                "snippet": description,
                "source_ref": first_non_empty(_string_value(output.get("html_url")), repo),
                "tool_name": "github.repo_context",
                "repo": repo,
                "default_branch": default_branch,
            }
        ]

    def _workflow_context_evidence_items(self, request_payload: JsonObject, output: JsonObject) -> list[JsonObject]:
        workflow = output.get("workflow")
        if not isinstance(workflow, dict):
            return []
        workflow_id = _string_value(workflow.get("id")) or _string_value(request_payload.get("workflow_id"))
        items: list[JsonObject] = []
        parts = [f"Workflow {workflow_id}", f"status={_string_value(workflow.get('status')) or 'unknown'}"]
        last_verdict = _string_value(workflow.get("last_verdict"))
        if last_verdict:
            parts.append(f"last_verdict={last_verdict}")
        failure_summary = _truncate_text(workflow.get("failure_summary"), 240)
        if failure_summary:
            parts.append(f"failure={failure_summary}")
        items.append(
            {
                "kind": "workflow_context",
                "summary": "; ".join(part for part in parts if part),
                "source_ref": workflow_id,
                "tool_name": "rsi.workflow_context",
                "workflow_id": workflow_id,
            }
        )
        recent_entries = output.get("recent_conversation_entries")
        if not isinstance(recent_entries, list):
            return items
        for entry in recent_entries[:4]:
            if not isinstance(entry, dict):
                continue
            body = _truncate_text(
                first_non_empty(
                    _string_value(entry.get("body")),
                    _string_value(entry.get("summary")),
                    _string_value(entry.get("content")),
                ),
                500,
            )
            if not body:
                continue
            actor = first_non_empty(_string_value(entry.get("actor_id")), _string_value(entry.get("actor_type")))
            item: JsonObject = {
                "kind": "workflow_conversation_entry",
                "summary": (f"[{actor}] " if actor else "") + body,
                "snippet": body,
                "source_ref": first_non_empty(_string_value(entry.get("id")), workflow_id),
                "tool_name": "rsi.workflow_context",
                "workflow_id": workflow_id,
            }
            _set_non_empty_text(item, "actor", actor)
            _set_non_empty_text(item, "entry_type", entry.get("entry_type"))
            _set_non_empty_text(item, "created_at", entry.get("created_at"))
            items.append(item)
        return items

    def _default_payload(self, name: str) -> JsonObject:
        payload: JsonObject = {}
        default_question = first_non_empty(self.default_question, self.task_prompt)
        repo_question = first_non_empty(self.repo_question, default_question, self.task_context_summary)
        knowledge_topic = first_non_empty(self.knowledge_topic, self.task_context_summary, self.task_repo)
        knowledge_question = first_non_empty(self.knowledge_question, default_question, self.task_context_summary)
        slack_history_focus = first_non_empty(self.slack_history_focus, default_question)
        slack_search_query = first_non_empty(self.slack_search_query, default_question)
        if self.trace_id:
            payload["trace_id"] = self.trace_id
        if name in {"repo.context", "repo.read_file", "repo.search", "github.repo_activity", "github.repo_context"} and self.task_repo:
            payload["repo"] = self.task_repo
        if name in {"repo.read_file", "repo.search"} and self.task_repo_ref:
            payload["ref"] = self.task_repo_ref
        if name == "repo.context" and repo_question:
            payload["question"] = repo_question
        if name == "knowledge.context":
            if knowledge_question:
                payload["question"] = knowledge_question
            if knowledge_topic:
                payload["topic"] = knowledge_topic
            payload["scope_id"] = self.task_repo
        if name == "slack.history":
            surface = self._default_slack_surface()
            channel_id = surface.get("channel_id", "")
            thread_ts = surface.get("thread_ts", "")
            if channel_id:
                payload["channel_id"] = channel_id
            if thread_ts:
                payload["thread_ts"] = thread_ts
            if slack_history_focus:
                payload["question"] = slack_history_focus
        if name == "slack.search":
            channel_ids = self._default_slack_channel_ids()
            if channel_ids:
                payload["channel_ids"] = channel_ids
            if slack_search_query:
                payload["query"] = slack_search_query
        if name == "slack.upload_file":
            surface = self._default_slack_surface()
            channel_id = surface.get("channel_id", "")
            thread_ts = surface.get("thread_ts", "")
            if channel_id:
                payload["channel_id"] = channel_id
            if thread_ts:
                payload["thread_ts"] = thread_ts
        if name == "github.repo_activity":
            since, until = self._activity_window_from_context_refs()
            if since:
                payload["since"] = since
            if until:
                payload["until"] = until
        if name == "rsi.runtime_deployment_facts":
            targets = self._runtime_deployment_targets_from_context_refs()
            if targets:
                payload["services"] = targets
        if name == "rsi.workflow_context":
            payload["trace_id"] = self.trace_id
        if name == "rsi.action_chain":
            payload["trace_id"] = self.trace_id
        if name == "rsi.runner_execution":
            payload["trace_id"] = self.trace_id
        if name == "sentry.lookup":
            payload["alert"] = self.task_context_summary or self.task_prompt
        if name in {"rsi.proposal_memory", "rsi.candidate_context"} and self.session_scope_kind == "proposal_candidate":
            payload["candidate_key"] = self.session_scope_id
        if name == "rsi.attempt_context":
            attempt_id = self._attempt_id_from_context_refs()
            if attempt_id:
                payload["attempt_id"] = attempt_id
        if name in {"kubernetes.inspect", "kubernetes.logs", "kubernetes.events", "rsi.runtime_deployment_facts"}:
            namespaces = _string_list(self.kubernetes_read_namespaces)
            if len(namespaces) == 1:
                payload["namespace"] = namespaces[0]
        if name.startswith("workspace."):
            if self.workspace_id:
                payload["workspace_id"] = self.workspace_id
            if self.attempt_id:
                payload["attempt_id"] = self.attempt_id
        return payload

    def _record_tool_selection(self, name: str, payload: JsonObject) -> None:
        if name not in {"slack.history", "slack.search"}:
            return
        surface: JsonObject = {"tool_name": name}
        if name == "slack.search":
            channel_ids = payload.get("channel_ids")
            if isinstance(channel_ids, list):
                normalized = [str(item).strip() for item in channel_ids if str(item).strip()]
                if normalized:
                    surface["channel_ids"] = normalized
                    surface["channel_id"] = normalized[0]
        else:
            channel_id = str(payload.get("channel_id", "")).strip()
            thread_ts = str(payload.get("thread_ts", "")).strip()
            if channel_id:
                surface["channel_id"] = channel_id
            if thread_ts:
                surface["thread_ts"] = thread_ts
        if len(surface) == 1:
            return
        encoded = json.dumps(surface, sort_keys=True)
        existing = {json.dumps(item, sort_keys=True) for item in self.selected_context_surfaces}
        if encoded not in existing:
            self.selected_context_surfaces.append(surface)

    def _attempt_id_from_context_refs(self) -> str:
        for item in self.context_refs:
            if str(item.get("kind", "")).strip() != "change_attempt":
                continue
            ref = str(item.get("ref", "")).strip()
            if ref:
                return ref
        return ""

    def _candidate_read_surfaces(self) -> list[JsonObject]:
        seen: set[str] = set()
        out: list[JsonObject] = []
        for item in self.context_refs:
            if str(item.get("kind", "")).strip() != "candidate_read_surface":
                continue
            channel_id = str(item.get("channel_id", "")).strip()
            thread_ts = str(item.get("thread_ts", "")).strip()
            ref = str(item.get("ref", "")).strip()
            if not channel_id and not thread_ts and not ref:
                continue
            candidate = {
                "channel_id": channel_id,
                "thread_ts": thread_ts,
                "ref": ref,
                "source": str(item.get("source", "")).strip(),
            }
            encoded = json.dumps(candidate, sort_keys=True)
            if encoded in seen:
                continue
            seen.add(encoded)
            out.append(candidate)
        if self.task_channel_id:
            fallback = {
                "channel_id": self.task_channel_id,
                "thread_ts": self.task_thread_ts,
                "ref": "",
                "source": "task_binding",
            }
            encoded = json.dumps(fallback, sort_keys=True)
            if encoded not in seen:
                out.insert(0, fallback)
        return out

    def _default_slack_surface(self) -> JsonObject:
        candidates = self._candidate_read_surfaces()
        if candidates:
            return candidates[0]
        return {
            "channel_id": self.task_channel_id,
            "thread_ts": self.task_thread_ts,
            "source": "task_binding",
        }

    def _default_slack_channel_ids(self) -> list[str]:
        out: list[str] = []
        seen: set[str] = set()
        for item in self._candidate_read_surfaces():
            channel_id = str(item.get("channel_id", "")).strip()
            if not channel_id or channel_id in seen:
                continue
            seen.add(channel_id)
            out.append(channel_id)
        return out

    def _activity_window_from_context_refs(self) -> tuple[str, str]:
        for item in self.context_refs:
            if str(item.get("kind", "")).strip() != "repo_activity_window":
                continue
            since = str(item.get("since", "")).strip()
            until = str(item.get("until", "")).strip()
            if since or until:
                return since, until
        return "", ""

    def _runtime_deployment_targets_from_context_refs(self) -> list[str]:
        out: list[str] = []
        seen: set[str] = set()
        for item in self.context_refs:
            if str(item.get("kind", "")).strip() != "runtime_deployment_targets":
                continue
            raw_values: list[Any] = []
            services = item.get("services")
            if isinstance(services, list):
                raw_values.extend(services)
            target_ref = str(item.get("target_ref") or item.get("ref") or "").strip()
            if target_ref:
                raw_values.extend(target_ref.split(","))
            for value in raw_values:
                text = str(value or "").strip()
                if not text or text in seen:
                    continue
                seen.add(text)
                out.append(text)
        return out


class CompositeToolProvider:
    def __init__(self, base_manager: ToolManagerLike | None, readonly_tools: ReadOnlyToolBinding) -> None:
        self._base_manager = base_manager
        self._readonly_tools = readonly_tools

    def has_tool(self, name: str) -> bool:
        if self._readonly_tools.has_tool(name):
            return True
        if self._base_manager is None:
            return False
        has_tool = getattr(self._base_manager, "has_tool", None)
        return bool(callable(has_tool) and has_tool(name))

    def handle_tool_call(self, name: str, args: JsonObject, **kwargs: Any) -> str:
        if self._readonly_tools.has_tool(name):
            return self._readonly_tools.handle_tool_call(name, args)
        if self._base_manager is None:
            raise ValueError(f"unknown tool {name}")
        return self._base_manager.handle_tool_call(name, args, **kwargs)

    def build_system_prompt(self) -> str:
        parts: list[str] = []
        if self._base_manager is not None:
            builder = getattr(self._base_manager, "build_system_prompt", None)
            if callable(builder):
                base_prompt = builder()
                if base_prompt:
                    parts.append(str(base_prompt).strip())
        readonly_names = ", ".join(self._readonly_tools.tool_names())
        if readonly_names:
            parts.append(
                "Additional governed RSI tools are available through the platform tool gateway. "
                f"Available tools: {readonly_names}. GitHub mutations are not exposed as governed RSI tools; when native terminal credentials are available, use gh from the company-computer terminal for explicitly requested issue, PR, or review work."
            )
        return "\n\n".join(part for part in parts if part)

    def __getattr__(self, name: str) -> Any:
        if self._base_manager is None:
            raise AttributeError(name)
        return getattr(self._base_manager, name)
