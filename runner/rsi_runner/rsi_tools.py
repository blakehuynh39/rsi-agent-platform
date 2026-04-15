from __future__ import annotations

from dataclasses import dataclass
import json
from typing import Any, Iterable
from urllib import error as urlerror
from urllib import request as urlrequest

from .json_types import JsonObject


READ_ONLY_HONCHO_TOOLS = frozenset({"honcho_profile", "honcho_search", "honcho_context"})
BLOCKED_HONCHO_TOOLS = frozenset({"honcho_conclude"})


_READ_ONLY_TOOL_SCHEMAS: dict[str, JsonObject] = {
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

_WORKSPACE_TOOL_SCHEMAS: dict[str, JsonObject] = {
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

READ_ONLY_RSI_TOOL_NAMES = tuple(sorted(_READ_ONLY_TOOL_SCHEMAS.keys()))
WORKSPACE_RSI_TOOL_NAMES = tuple(sorted(_WORKSPACE_TOOL_SCHEMAS.keys()))
IMPLEMENT_RSI_TOOL_NAMES = tuple(sorted(_TOOL_SCHEMAS.keys()))


def tool_schema_wrappers(names: Iterable[str]) -> list[JsonObject]:
    wrappers: list[JsonObject] = []
    for name in names:
        schema = _TOOL_SCHEMAS.get(name)
        if schema is None:
            continue
        wrappers.append({"type": "function", "function": schema})
    return wrappers


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


@dataclass
class ReadOnlyToolBinding:
    base_url: str
    allowed_tool_names: list[str]
    task_repo: str
    task_prompt: str
    task_context_summary: str
    trace_id: str
    session_scope_kind: str
    session_scope_id: str
    context_refs: list[JsonObject]
    execution_mode: str = ""
    attempt_id: str = ""
    workspace_id: str = ""
    timeout_seconds: int = 30

    def has_tool(self, name: str) -> bool:
        return name in _TOOL_SCHEMAS and name in set(self.allowed_tool_names)

    def tool_names(self) -> list[str]:
        return list(normalize_tool_names(self.allowed_tool_names))

    def handle_tool_call(self, name: str, args: JsonObject) -> str:
        payload = self._default_payload(name)
        payload.update(args or {})
        req = urlrequest.Request(
            f"{self.base_url.rstrip('/')}/api/tools/{name}/execute",
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            with urlrequest.urlopen(req, timeout=self.timeout_seconds) as resp:
                body = resp.read().decode("utf-8")
        except urlerror.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")
            return json.dumps(
                {
                    "tool_name": name,
                    "status": "error",
                    "error": f"tool gateway returned {exc.code}: {detail}",
                }
            )
        except (urlerror.URLError, TimeoutError, ConnectionError, OSError) as exc:
            return json.dumps(
                {
                    "tool_name": name,
                    "status": "error",
                    "error": f"tool gateway request failed: {exc}",
                }
            )

        try:
            parsed = json.loads(body)
        except json.JSONDecodeError:
            return json.dumps(
                {
                    "tool_name": name,
                    "status": "error",
                    "error": "tool gateway returned invalid JSON",
                    "body": body[:8000],
                }
            )
        return json.dumps(
            {
                "tool_name": name,
                "status": parsed.get("status", ""),
                "available": parsed.get("available", True),
                "summary": parsed.get("summary", ""),
                "provider": parsed.get("provider", ""),
                "provider_ref": parsed.get("provider_ref", ""),
                "approval_state": parsed.get("approval_state", ""),
                "output": parsed.get("output", {}),
                "raw_artifact_refs": parsed.get("raw_artifact_refs", []),
            }
        )

    def _default_payload(self, name: str) -> JsonObject:
        payload: JsonObject = {}
        if self.trace_id:
            payload["trace_id"] = self.trace_id
        if name in {"repo.context", "github.repo_activity", "github.repo_context"} and self.task_repo:
            payload["repo"] = self.task_repo
        if name == "repo.context":
            payload["question"] = self.task_prompt
        if name == "knowledge.context":
            payload["question"] = self.task_prompt
            payload["topic"] = self.task_prompt or self.task_context_summary
            payload["scope_id"] = self.task_repo
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

    def _attempt_id_from_context_refs(self) -> str:
        for item in self.context_refs:
            if str(item.get("kind", "")).strip() != "change_attempt":
                continue
            ref = str(item.get("ref", "")).strip()
            if ref:
                return ref
        return ""


class CompositeToolProvider:
    def __init__(self, base_manager: Any, readonly_tools: ReadOnlyToolBinding) -> None:
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
