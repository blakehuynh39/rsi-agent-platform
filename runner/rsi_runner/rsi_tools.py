from __future__ import annotations

from copy import deepcopy
from dataclasses import dataclass, field
import json
import time
from typing import Any, Iterable, Protocol
from urllib import error as urlerror
from urllib import request as urlrequest

from .json_types import JsonObject, JsonToolFunctionSchema, JsonToolWrapperSchema


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
        "description": "Read-only Kubernetes pod and event inspection for a namespace and target selector.",
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
        "description": "Read-only Kubernetes pod log lookup for a namespace and target or pod name.",
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
        "description": "Read-only Kubernetes event lookup for a namespace and target selector.",
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
        "description": "Read-only RSI deployment and runtime-configuration facts, including live Kubernetes deployment state when available.",
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
}


def _string_list(value: Any) -> list[str]:
    if not isinstance(value, list):
        return []
    return [str(item).strip() for item in value if str(item).strip()]


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


def tool_schema_wrappers(names: Iterable[str]) -> list[JsonToolWrapperSchema]:
    wrappers: list[JsonToolWrapperSchema] = []
    for name in names:
        schema = _TOOL_SCHEMAS.get(name)
        if schema is None:
            continue
        wrapped = deepcopy(schema)
        wrapped["name"] = tool_transport_name(name)
        wrappers.append({"type": "function", "function": wrapped})
    return wrappers


def transport_tool_schema(name: str) -> JsonToolFunctionSchema:
    schema = _TOOL_SCHEMAS.get(name)
    if schema is None:
        raise KeyError(name)
    wrapped = deepcopy(schema)
    wrapped["name"] = tool_transport_name(name)
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
    execution_mode: str = ""
    attempt_id: str = ""
    workspace_id: str = ""
    timeout_seconds: int = 30
    selected_context_surfaces: list[JsonObject] = field(default_factory=list)
    tool_calls: list[JsonObject] = field(default_factory=list)
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

    def handle_tool_call(self, name: str, args: JsonObject) -> str:
        transport_name = tool_transport_name(name)
        canonical_name = canonical_tool_name(name)
        payload = self._default_payload(canonical_name)
        payload.update(args or {})
        self._record_tool_selection(canonical_name, payload)
        call_started_at = time.time()
        tool_call_id = self._next_tool_call_id(canonical_name)
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

    def diagnostics(self) -> JsonObject:
        return {
            "candidate_read_surfaces": self._candidate_read_surfaces(),
            "selected_context_surfaces": list(self.selected_context_surfaces),
            "tool_calls": deepcopy(self.tool_calls),
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

    def _default_payload(self, name: str) -> JsonObject:
        payload: JsonObject = {}
        if self.trace_id:
            payload["trace_id"] = self.trace_id
        if name in {"repo.context", "repo.read_file", "repo.search", "github.repo_activity", "github.repo_context"} and self.task_repo:
            payload["repo"] = self.task_repo
        if name in {"repo.read_file", "repo.search"} and self.task_repo_ref:
            payload["ref"] = self.task_repo_ref
        if name == "repo.context":
            payload["question"] = self.task_prompt
        if name == "knowledge.context":
            payload["question"] = self.task_prompt
            payload["topic"] = self.task_prompt or self.task_context_summary
            payload["scope_id"] = self.task_repo
        if name == "slack.history":
            surface = self._default_slack_surface()
            channel_id = surface.get("channel_id", "")
            thread_ts = surface.get("thread_ts", "")
            if channel_id:
                payload["channel_id"] = channel_id
            if thread_ts:
                payload["thread_ts"] = thread_ts
            if self.task_prompt:
                payload["question"] = self.task_prompt
        if name == "slack.search":
            channel_ids = self._default_slack_channel_ids()
            if channel_ids:
                payload["channel_ids"] = channel_ids
            if self.task_prompt:
                payload["query"] = self.task_prompt
        if name == "github.repo_activity":
            since, until = self._activity_window_from_context_refs()
            if since:
                payload["since"] = since
            if until:
                payload["until"] = until
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
                f"Available tools: {readonly_names}. GitHub mutation, Slack posting, and unmanaged shell access remain blocked."
            )
        return "\n\n".join(part for part in parts if part)

    def __getattr__(self, name: str) -> Any:
        if self._base_manager is None:
            raise AttributeError(name)
        return getattr(self._base_manager, name)
