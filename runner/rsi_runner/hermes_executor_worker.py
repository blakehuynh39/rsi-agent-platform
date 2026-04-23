from __future__ import annotations

import json
import logging
import os
from pathlib import Path
import tempfile
import sys
from typing import Any

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
        runtime = _runtime_override(payload)
        if not cli._init_agent(
            model_override=_string(payload.get("model")) or None,
            runtime_override=runtime,
            route_label="rsi_executor",
        ):
            raise RuntimeError("HermesCLI failed to initialize an agent.")
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
            output = {
                "ok": False,
                "error": str(exc),
                "mcp_cleanup_errors": list(cleanup_result.errors),
                "mcp_cleanup_status": cleanup_result.status,
                "session_id": "",
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
