from __future__ import annotations

from dataclasses import dataclass, field
import hashlib
import os
import re
from typing import Any, Callable

from .json_types import JsonObject, JsonValue
from .rsi_tools import normalize_tool_names


def _json_object_or_empty(value: JsonValue | None) -> JsonObject:
    if isinstance(value, dict):
        return value
    return {}


def _string_or_empty(value: JsonValue | None) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value.strip()
    return str(value).strip()


def _bool_or_false(value: JsonValue | None) -> bool:
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        lowered = value.strip().lower()
        if lowered in {"1", "true", "t", "yes", "y", "on"}:
            return True
        if lowered in {"0", "false", "f", "no", "n", "off"}:
            return False
    return False


def _string_list(value: JsonValue | None) -> list[str]:
    if not isinstance(value, list):
        return []
    return [str(item).strip() for item in value if str(item).strip()]


def _slug(value: str, *, fallback: str) -> str:
    text = re.sub(r"[^A-Za-z0-9_.-]+", "-", str(value or "").strip().lower()).strip("._-")
    return text or fallback


def _first_non_empty(*values: str) -> str:
    for value in values:
        text = _string_or_empty(value)
        if text:
            return text
    return ""


@dataclass(frozen=True)
class TaskScopedMCPServer:
    source_label: str
    profile: str
    server_name: str
    toolset_alias: str
    included_tool_names: list[str]
    hermes_config: JsonObject


@dataclass(frozen=True)
class TaskScopedMCPCleanupResult:
    status: str
    cleaned_server_names: list[str]
    failed_server_names: list[str]
    errors: list[str]


@dataclass
class TaskScopedMCPRegistration:
    servers: list[TaskScopedMCPServer] = field(default_factory=list)
    cleanup_result: TaskScopedMCPCleanupResult | None = None

    @property
    def enabled_toolsets(self) -> list[str]:
        return [server.toolset_alias for server in self.servers]

    @property
    def server_names(self) -> list[str]:
        return [server.server_name for server in self.servers]

    @property
    def tool_names(self) -> list[str]:
        return normalize_tool_names(
            [tool_name for server in self.servers for tool_name in server.included_tool_names if tool_name]
        )

    @property
    def enabled(self) -> bool:
        return len(self.servers) > 0


class HermesTaskScopedMCPAdapter:
    def __init__(
        self,
        *,
        default_slack_server_url: str = "",
        slack_read_tool_names_resolver: Callable[[], list[str]] | None = None,
        slack_send_tool_name_resolver: Callable[[], str] | None = None,
    ) -> None:
        self._default_slack_server_url = default_slack_server_url
        self._slack_read_tool_names_resolver = slack_read_tool_names_resolver
        self._slack_send_tool_name_resolver = slack_send_tool_name_resolver

    def register_task_servers(self, task: Any) -> TaskScopedMCPRegistration:
        servers = self._translate_task_servers(task)
        if not servers:
            return TaskScopedMCPRegistration()

        mcp_tool = self._load_hermes_mcp_tool()
        configs = {server.server_name: server.hermes_config for server in servers}
        try:
            registered_names = list(mcp_tool.register_mcp_servers(configs))
        except Exception as exc:  # pragma: no cover - depends on external Hermes runtime
            raise RuntimeError(f"Failed to register Hermes MCP servers for this task: {exc}") from exc

        missing = [
            server_name
            for server_name in configs
            if server_name not in registered_names and server_name not in getattr(mcp_tool, "_servers", {})
        ]
        if missing:
            cleanup_result = self.cleanup_registration(TaskScopedMCPRegistration(servers=servers))
            cleanup_error = f" Cleanup status: {cleanup_result.status}."
            raise RuntimeError(f"Hermes MCP registration did not connect expected servers: {missing}.{cleanup_error}")
        return TaskScopedMCPRegistration(servers=servers)

    def cleanup_registration(self, registration: TaskScopedMCPRegistration) -> TaskScopedMCPCleanupResult:
        if not registration.enabled:
            result = TaskScopedMCPCleanupResult(
                status="not_needed",
                cleaned_server_names=[],
                failed_server_names=[],
                errors=[],
            )
            registration.cleanup_result = result
            return result

        try:
            mcp_tool = self._load_hermes_mcp_tool()
        except RuntimeError as exc:
            result = TaskScopedMCPCleanupResult(
                status="cleanup_runtime_unavailable",
                cleaned_server_names=[],
                failed_server_names=registration.server_names,
                errors=[str(exc)],
            )
            registration.cleanup_result = result
            return result

        cleaned: list[str] = []
        failed: list[str] = []
        errors: list[str] = []
        for server in registration.servers:
            server_task = getattr(mcp_tool, "_servers", {}).get(server.server_name)
            try:
                if server_task is not None:
                    mcp_tool._run_on_mcp_loop(server_task.shutdown())
                cleaned.append(server.server_name)
            except Exception as exc:  # pragma: no cover - depends on external Hermes runtime
                failed.append(server.server_name)
                errors.append(f"{server.server_name}: {exc}")
            finally:
                try:
                    getattr(mcp_tool, "_servers", {}).pop(server.server_name, None)
                except Exception:  # pragma: no cover - defensive
                    pass

        status = "cleaned"
        if failed and cleaned:
            status = "cleanup_partial"
        elif failed:
            status = "cleanup_failed"
        result = TaskScopedMCPCleanupResult(
            status=status,
            cleaned_server_names=cleaned,
            failed_server_names=failed,
            errors=errors,
        )
        registration.cleanup_result = result
        return result

    def _translate_task_servers(self, task: Any) -> list[TaskScopedMCPServer]:
        translated: list[TaskScopedMCPServer] = []
        for index, server in enumerate(getattr(task, "mcp_servers", []) or []):
            if not isinstance(server, dict):
                continue
            translated.append(self._translate_single_server(task, server, index=index))
        return translated

    def _translate_single_server(self, task: Any, server: JsonObject, *, index: int) -> TaskScopedMCPServer:
        label = _first_non_empty(_string_or_empty(server.get("server_label")), "mcp")
        profile = _string_or_empty(server.get("profile"))
        default_server_url = self._default_slack_server_url if profile in {"slack_mcp_read", "slack_mcp_reply"} else ""
        server_url = _first_non_empty(_string_or_empty(server.get("server_url")), default_server_url)
        if not server_url:
            raise RuntimeError(f"MCP server URL is required for server '{label}'.")

        headers = dict(_json_object_or_empty(server.get("headers")))
        header_env_vars = _json_object_or_empty(server.get("header_env_vars"))
        for header_name, env_var in header_env_vars.items():
            if not isinstance(header_name, str) or not isinstance(env_var, str):
                continue
            env_value = os.getenv(env_var, "").strip()
            if not env_value:
                raise RuntimeError(f"MCP header env var {env_var} is not configured.")
            headers[header_name] = env_value

        authorization = _string_or_empty(server.get("authorization"))
        authorization_env_var = _string_or_empty(server.get("authorization_env_var"))
        if not authorization and authorization_env_var:
            authorization = os.getenv(authorization_env_var, "").strip()
            if not authorization:
                raise RuntimeError(f"MCP authorization env var {authorization_env_var} is not configured.")
        if not authorization and profile in {"slack_mcp_read", "slack_mcp_reply"}:
            authorization = os.getenv("RSI_SLACK_USER_TOKEN", "").strip()
            if not authorization:
                raise RuntimeError("Slack user token is not configured.")
        if authorization:
            if not authorization.lower().startswith("bearer "):
                authorization = f"Bearer {authorization}"
            headers["Authorization"] = authorization

        included_tool_names = self._included_tool_names(server, label=label, profile=profile)
        hermes_config: JsonObject = {"url": server_url}
        if headers:
            hermes_config["headers"] = headers
        if included_tool_names:
            hermes_config["tools"] = {"include": included_tool_names}

        server_name = self._task_scoped_server_name(task, label=label, server_url=server_url, index=index)
        return TaskScopedMCPServer(
            source_label=label,
            profile=profile,
            server_name=server_name,
            toolset_alias=f"mcp-{server_name}",
            included_tool_names=included_tool_names,
            hermes_config=hermes_config,
        )

    def _included_tool_names(self, server: JsonObject, *, label: str, profile: str) -> list[str]:
        allowed_tools = _json_object_or_empty(server.get("allowed_tools"))
        explicit_tool_names = normalize_tool_names(_string_list(allowed_tools.get("tool_names")))
        if explicit_tool_names:
            return explicit_tool_names
        if profile == "slack_mcp_read":
            tool_names = normalize_tool_names(self._resolve_slack_read_tool_names())
            if not tool_names:
                raise RuntimeError("Slack MCP read-tool discovery returned no tools.")
            return tool_names
        if profile == "slack_mcp_reply":
            tool_names = normalize_tool_names([*self._resolve_slack_read_tool_names(), self._resolve_slack_send_tool_name()])
            if not tool_names:
                raise RuntimeError("Slack MCP reply-tool discovery returned no tools.")
            return tool_names
        if profile == "notion_mcp_read":
            return ["search", "fetch"]
        if _bool_or_false(allowed_tools.get("read_only")):
            raise RuntimeError(
                f"MCP server '{label}' requested read_only access without explicit tool_names; refusing to expose the full server."
            )
        return []

    def _resolve_slack_read_tool_names(self) -> list[str]:
        if self._slack_read_tool_names_resolver is None:
            raise RuntimeError("Slack MCP read-tool discovery is unavailable in this runner.")
        return normalize_tool_names(self._slack_read_tool_names_resolver())

    def _resolve_slack_send_tool_name(self) -> str:
        if self._slack_send_tool_name_resolver is None:
            raise RuntimeError("Slack MCP send-tool discovery is unavailable in this runner.")
        tool_name = _string_or_empty(self._slack_send_tool_name_resolver())
        if not tool_name:
            raise RuntimeError("Slack MCP send-tool discovery returned an empty tool name.")
        return tool_name

    def _task_scoped_server_name(self, task: Any, *, label: str, server_url: str, index: int) -> str:
        scope = _first_non_empty(
            _string_or_empty(getattr(task, "trace_id", "")),
            _string_or_empty(getattr(task, "workflow_id", "")),
            _string_or_empty(getattr(task, "conversation_id", "")),
            _string_or_empty(getattr(task, "session_scope_id", "")),
            _string_or_empty(getattr(task, "task_type", "")),
            "task",
        )
        digest = hashlib.sha1(f"{scope}|{label}|{server_url}|{index}".encode("utf-8")).hexdigest()[:10]
        return f"rsi-task-{_slug(scope, fallback='task')[:24]}-{index}-{_slug(label, fallback='mcp')[:16]}-{digest}"

    def _load_hermes_mcp_tool(self) -> Any:
        try:
            from tools import mcp_tool  # type: ignore
        except (ImportError, ModuleNotFoundError) as exc:  # pragma: no cover - depends on external Hermes install
            raise RuntimeError("Hermes MCP runtime is not installed in this environment.") from exc
        return mcp_tool
