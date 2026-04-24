from __future__ import annotations

from copy import deepcopy
import hashlib
import json
import logging
import os
from pathlib import Path
import tempfile
import sys
from typing import Any

from .rsi_tools import normalize_tool_names, _strict_json_schema
from .hermes_mcp_adapter import HermesTaskScopedMCPAdapter, TaskScopedMCPRegistration, TaskScopedMCPServer

try:
    from agent.skill_commands import build_preloaded_skills_prompt  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    build_preloaded_skills_prompt = None

try:
    from cli import HermesCLI  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    HermesCLI = None

try:
    from hermes_state import SessionDB  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    SessionDB = None


logger = logging.getLogger(__name__)
_RESULT_PREFIX = "RSI_EXECUTOR_RESULT::"
_ARTIFACT_LIST_FILES_TOOL = "artifact.list_files"
_ARTIFACT_WRITE_FILE_TOOL = "artifact.write_file"
_ARTIFACT_LIST_FILES_TRANSPORT = "artifact_list_files"
_ARTIFACT_WRITE_FILE_TRANSPORT = "artifact_write_file"
_ARTIFACT_TOOL_SCHEMAS = {
    _ARTIFACT_LIST_FILES_TOOL: {
        "name": _ARTIFACT_LIST_FILES_TOOL,
        "description": "List files inside the native executor artifact output directory.",
        "parameters": {
            "type": "object",
            "properties": {
                "path": {"type": "string"},
            },
        },
    },
    _ARTIFACT_WRITE_FILE_TOOL: {
        "name": _ARTIFACT_WRITE_FILE_TOOL,
        "description": "Write file content inside the native executor artifact output directory.",
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


def _json_object(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def _json_list(value: Any) -> list[Any]:
    return value if isinstance(value, list) else []


def _string(value: Any) -> str:
    if value is None:
        return ""
    return str(value).strip()


def _result(payload: dict[str, Any]) -> None:
    sys.stdout.write(_RESULT_PREFIX + json.dumps(payload, ensure_ascii=True, sort_keys=True) + "\n")
    sys.stdout.flush()




def _artifact_tool_schema_wrappers() -> list[dict[str, Any]]:
    wrappers: list[dict[str, Any]] = []
    for canonical_name, transport_name in (
        (_ARTIFACT_LIST_FILES_TOOL, _ARTIFACT_LIST_FILES_TRANSPORT),
        (_ARTIFACT_WRITE_FILE_TOOL, _ARTIFACT_WRITE_FILE_TRANSPORT),
    ):
        schema = _ARTIFACT_TOOL_SCHEMAS[canonical_name]
        wrapped = deepcopy(schema)
        wrapped["name"] = transport_name
        parameters = wrapped.get("parameters")
        if isinstance(parameters, dict):
            wrapped["parameters"] = _strict_json_schema(parameters)
        wrappers.append({"type": "function", "function": wrapped})
    return wrappers


class _OverlayToolProvider:
    def __init__(self, base_manager: Any | None, overlay: Any) -> None:
        self._base_manager = base_manager
        self._overlay = overlay

    def has_tool(self, name: str) -> bool:
        if self._overlay.has_tool(name):
            return True
        if self._base_manager is None:
            return False
        has_tool = getattr(self._base_manager, "has_tool", None)
        return bool(callable(has_tool) and has_tool(name))

    def handle_tool_call(self, name: str, args: dict[str, Any], **kwargs: Any) -> str:
        if self._overlay.has_tool(name):
            return self._overlay.handle_tool_call(name, args)
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
        overlay_prompt = self._overlay.build_system_prompt()
        if overlay_prompt:
            parts.append(str(overlay_prompt).strip())
        return "\n\n".join(part for part in parts if part)

    def __getattr__(self, name: str) -> Any:
        if self._base_manager is None:
            raise AttributeError(name)
        return getattr(self._base_manager, name)


class _LocalArtifactToolBinding:
    def __init__(self, artifact_output_dir: Path) -> None:
        self._artifact_output_dir = artifact_output_dir.expanduser().resolve()
        self._artifact_output_dir.mkdir(parents=True, exist_ok=True)
        self._events: list[dict[str, Any]] = []

    @property
    def events(self) -> list[dict[str, Any]]:
        return list(self._events)

    def has_tool(self, name: str) -> bool:
        return str(name or "").strip() in set(self.tool_names())

    def tool_names(self) -> list[str]:
        return normalize_tool_names(
            [
                _ARTIFACT_LIST_FILES_TRANSPORT,
                _ARTIFACT_WRITE_FILE_TRANSPORT,
            ]
        )

    def build_system_prompt(self) -> str:
        return (
            "Local artifact tools are available for this render. "
            f"Use {_ARTIFACT_LIST_FILES_TRANSPORT} and {_ARTIFACT_WRITE_FILE_TRANSPORT} "
            f"only within {self._artifact_output_dir}."
        )

    def handle_tool_call(self, name: str, args: dict[str, Any], **_kwargs: Any) -> str:
        transport_name = str(name or "").strip()
        if transport_name == _ARTIFACT_LIST_FILES_TRANSPORT:
            return self._handle_list_files(args)
        if transport_name == _ARTIFACT_WRITE_FILE_TRANSPORT:
            return self._handle_write_file(args)
        raise ValueError(f"unknown tool {name}")

    def _record_event(self, event_type: str, status: str, payload: dict[str, Any]) -> None:
        self._events.append(
            {
                "event_type": event_type,
                "status": status,
                "payload": payload,
            }
        )

    def _resolve_path(self, value: Any) -> Path:
        raw = _string(value)
        candidate = Path(raw).expanduser() if raw.startswith("/") else (self._artifact_output_dir / raw)
        resolved = candidate.resolve()
        try:
            resolved.relative_to(self._artifact_output_dir)
        except ValueError as exc:
            raise ValueError("artifact_path_outside_root") from exc
        return resolved

    def _entry(self, path: Path) -> dict[str, Any]:
        stat = path.stat()
        return {
            "name": path.name,
            "path": str(path),
            "is_dir": path.is_dir(),
            "size_bytes": 0 if path.is_dir() else stat.st_size,
        }

    def _handle_list_files(self, args: dict[str, Any]) -> str:
        target = self._resolve_path(args.get("path") or "")
        entries: list[dict[str, Any]] = []
        if target.exists():
            if target.is_dir():
                entries = [self._entry(item) for item in sorted(target.iterdir())]
            else:
                entries = [self._entry(target)]
        payload = {
            "tool_name": _ARTIFACT_LIST_FILES_TOOL,
            "status": "ok",
            "summary": f"Listed {len(entries)} item(s) in the native artifact directory.",
            "output": {
                "artifact_output_dir": str(self._artifact_output_dir),
                "path": str(target),
                "entries": entries,
            },
        }
        return json.dumps(payload, ensure_ascii=True, sort_keys=True)

    def _handle_write_file(self, args: dict[str, Any]) -> str:
        requested_path = _string(args.get("path"))
        started_payload = {
            "tool_name": _ARTIFACT_WRITE_FILE_TOOL,
            "requested_path": requested_path,
            "artifact_output_dir": str(self._artifact_output_dir),
        }
        self._record_event("artifact.write.started", "running", started_payload)
        try:
            target = self._resolve_path(requested_path)
            target.parent.mkdir(parents=True, exist_ok=True)
            content = _string(args.get("content"))
            target.write_text(content, encoding="utf-8")
            content_bytes = content.encode("utf-8")
            file_ref = f"file://{target}"
            completed_payload = {
                "tool_name": _ARTIFACT_WRITE_FILE_TOOL,
                "path": str(target),
                "workspace_path": str(target),
                "file_ref": file_ref,
                "artifact_output_dir": str(self._artifact_output_dir),
                "bytes_written": len(content_bytes),
                "size_bytes": len(content_bytes),
                "sha256": hashlib.sha256(content_bytes).hexdigest(),
                "share_status": "local",
            }
            self._record_event("artifact.write.completed", "completed", completed_payload)
            return json.dumps(
                {
                    "tool_name": _ARTIFACT_WRITE_FILE_TOOL,
                    "status": "ok",
                    "summary": f"Wrote artifact file to {target}.",
                    "output": completed_payload,
                    "raw_artifact_refs": [file_ref],
                },
                ensure_ascii=True,
                sort_keys=True,
            )
        except Exception as exc:
            failed_payload = {
                "tool_name": _ARTIFACT_WRITE_FILE_TOOL,
                "requested_path": requested_path,
                "artifact_output_dir": str(self._artifact_output_dir),
                "error": str(exc),
            }
            self._record_event("artifact.write.failed", "failed", failed_payload)
            raise RuntimeError(str(exc)) from exc


def _persist_result(payload: dict[str, Any], result_path_arg: str) -> Path:
    result_path = Path(result_path_arg).expanduser().resolve()
    result_path.parent.mkdir(parents=True, exist_ok=True)
    fd, temp_path_arg = tempfile.mkstemp(prefix=result_path.name + ".", suffix=".tmp", dir=str(result_path.parent))
    temp_path = Path(temp_path_arg)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            json.dump(payload, handle, ensure_ascii=True, sort_keys=True)
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(temp_path, result_path)
    finally:
        if temp_path.exists():
            temp_path.unlink(missing_ok=True)
    return result_path


def _load_request(path_arg: str) -> dict[str, Any]:
    path = Path(path_arg).expanduser().resolve()
    parsed = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(parsed, dict):
        raise ValueError("executor worker request must be a JSON object")
    return parsed


def _configure_logging() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")


def _load_task_scoped_mcp_registration(payload: dict[str, Any]) -> TaskScopedMCPRegistration:
    servers: list[TaskScopedMCPServer] = []
    for item in _json_list(payload.get("task_scoped_mcp_servers")):
        if not isinstance(item, dict):
            continue
        server_name = _string(item.get("server_name"))
        toolset_alias = _string(item.get("toolset_alias"))
        if not server_name or not toolset_alias:
            continue
        servers.append(
            TaskScopedMCPServer(
                source_label=_string(item.get("source_label")) or server_name,
                profile=_string(item.get("profile")),
                server_name=server_name,
                toolset_alias=toolset_alias,
                included_tool_names=[_string(tool_name) for tool_name in _json_list(item.get("included_tool_names")) if _string(tool_name)],
                hermes_config=_json_object(item.get("hermes_config")),
            )
        )
    return TaskScopedMCPRegistration(servers=servers)


def _prepare_cli_session(payload: dict[str, Any]) -> tuple[object, str]:
    if HermesCLI is None:
        raise RuntimeError("HermesCLI is unavailable in this environment.")
    session_id = _string(payload.get("session_id"))
    if not session_id:
        raise RuntimeError("session_id is required for native Hermes execution.")
    toolsets = [item for item in _json_list(payload.get("toolsets")) if _string(item)]
    cli = HermesCLI(
        model=_string(payload.get("model")) or None,
        toolsets=toolsets or None,
        provider=None,
        api_key=None,
        base_url=None,
        max_turns=max(1, int(payload.get("max_iterations") or 1)),
        verbose=False,
        compact=True,
        resume=None,
        checkpoints=False,
        pass_session_id=False,
    )
    cli.session_id = session_id
    cli.tool_progress_mode = "off"
    cli.conversation_history = [item for item in _json_list(payload.get("conversation_history")) if isinstance(item, dict)]
    cli.system_prompt = _string(payload.get("system_message"))
    if build_preloaded_skills_prompt is not None:
        requested_skills = [_string(item) for item in _json_list(payload.get("requested_skills")) if _string(item)]
        if requested_skills:
            skills_prompt, loaded, missing = build_preloaded_skills_prompt(requested_skills, task_id=session_id)
            if missing:
                raise RuntimeError(f"Unknown skill(s): {', '.join(missing)}")
            if skills_prompt:
                cli.system_prompt = "\n\n".join(part for part in [cli.system_prompt, skills_prompt] if _string(part))
                cli.preloaded_skills = list(loaded)
    return cli, session_id


def _prepare_session_db(payload: dict[str, Any], session_id: str, system_prompt: str) -> Any:
    if SessionDB is None:
        return None
    db = SessionDB()
    parent_session_id = _string(payload.get("parent_session_id")) or None
    model = _string(payload.get("model")) or None
    db.create_session(
        session_id=session_id,
        source="rsi_executor",
        model=model,
        system_prompt=system_prompt or None,
        parent_session_id=parent_session_id,
    )
    db.reopen_session(session_id)
    if system_prompt:
        db.update_system_prompt(session_id, system_prompt)
    return db


def _runtime_override(payload: dict[str, Any]) -> dict[str, Any]:
    runtime = _json_object(payload.get("runtime"))
    return {
        "api_key": _string(runtime.get("api_key")) or None,
        "base_url": _string(runtime.get("base_url")) or None,
        "provider": _string(runtime.get("provider")) or None,
        "api_mode": _string(runtime.get("api_mode")) or None,
        "command": None,
        "args": [],
        "credential_pool": None,
    }


def _initialize_cli_agent(cli: object, payload: dict[str, Any]) -> None:
    runtime = _runtime_override(payload)
    if not cli._init_agent(
        model_override=_string(payload.get("model")) or None,
        runtime_override=runtime,
    ):
        raise RuntimeError("HermesCLI failed to initialize an agent.")


def _attach_local_artifact_tools(cli: object, payload: dict[str, Any]) -> _LocalArtifactToolBinding | None:
    execution_phase = _string(payload.get("execution_phase"))
    artifact_output_dir = _string(payload.get("artifact_output_dir"))
    if execution_phase != "render" or not artifact_output_dir:
        return None
    binding = _LocalArtifactToolBinding(Path(artifact_output_dir))
    agent = getattr(cli, "agent", None)
    if agent is None:
        return binding
    current_tools = list(getattr(agent, "tools", []) or [])
    current_tools.extend(_artifact_tool_schema_wrappers())
    agent.tools = current_tools
    current_valid = set(getattr(agent, "valid_tool_names", set()) or set())
    current_valid.update(binding.tool_names())
    agent.valid_tool_names = current_valid
    agent._memory_manager = _OverlayToolProvider(getattr(agent, "_memory_manager", None), binding)
    return binding


def main() -> None:
    _configure_logging()
    exit_code = 0
    mcp_adapter = HermesTaskScopedMCPAdapter()
    mcp_registration = TaskScopedMCPRegistration()
    try:
        if len(sys.argv) < 2:
            raise RuntimeError("executor worker request path is required")
        payload = _load_request(sys.argv[1])
        result_path = _string(payload.get("result_path"))
        if not result_path:
            raise RuntimeError("result_path is required for native Hermes execution.")
        workdir = _string(payload.get("workdir"))
        if workdir:
            os.chdir(workdir)
        mcp_registration = _load_task_scoped_mcp_registration(payload)
        if mcp_registration.enabled:
            mcp_adapter.register_prepared_servers(mcp_registration)
        cli, session_id = _prepare_cli_session(payload)
        cli._session_db = _prepare_session_db(payload, session_id, _string(cli.system_prompt))
        _initialize_cli_agent(cli, payload)
        artifact_tools = _attach_local_artifact_tools(cli, payload)
        cli.agent.quiet_mode = True
        result = cli.agent.run_conversation(
            user_message=_string(payload.get("prompt")),
            conversation_history=cli.conversation_history,
            task_id=session_id,
        )
        response = ""
        if isinstance(result, dict):
            response = _string(result.get("final_response"))
        if not response:
            response = _string(result)
        output = {
            "ok": not (isinstance(result, dict) and bool(result.get("failed"))),
            "mcp_cleanup_errors": [],
            "mcp_cleanup_status": "not_needed" if not mcp_registration.enabled else "worker_cleanup_pending",
            "response": response,
            "result": result if isinstance(result, dict) else {"value": response},
            "session_id": session_id,
            "artifact_tool_events": artifact_tools.events if artifact_tools is not None else [],
        }
        cleanup_result = mcp_adapter.cleanup_registration(mcp_registration)
        output["mcp_cleanup_errors"] = list(cleanup_result.errors)
        output["mcp_cleanup_status"] = cleanup_result.status
        _persist_result(output, result_path)
        _result(output)
    except Exception as exc:  # pragma: no cover - exercised via subprocess integration
        logger.exception("native Hermes executor worker failed")
        exit_code = 1
        try:
            payload = locals().get("payload")
            result_path = _string(payload.get("result_path")) if isinstance(payload, dict) else ""
            cleanup_result = mcp_adapter.cleanup_registration(mcp_registration)
            artifact_tools = locals().get("artifact_tools")
            output = {
                "ok": False,
                "error": str(exc),
                "session_id": _string(payload.get("session_id")) if isinstance(payload, dict) else "",
                "artifact_tool_events": list(getattr(artifact_tools, "events", []) or []),
                "mcp_cleanup_errors": list(cleanup_result.errors),
                "mcp_cleanup_status": cleanup_result.status,
            }
            if result_path:
                _persist_result(output, result_path)
            _result(output)
        except Exception:
            logger.exception("failed to persist native Hermes executor error result")
    logging.shutdown()
    try:
        sys.stdout.flush()
        sys.stderr.flush()
    except Exception:
        pass
    os._exit(exit_code)


if __name__ == "__main__":
    main()
