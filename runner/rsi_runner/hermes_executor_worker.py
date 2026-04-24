from __future__ import annotations

import json
import logging
import os
from pathlib import Path
import tempfile
import sys
from typing import Any

from .hermes_agent_adapter import HermesAgentAdapter, HermesContractError
from .hermes_mcp_adapter import HermesTaskScopedMCPAdapter, TaskScopedMCPRegistration, TaskScopedMCPServer


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


def _error_output(
    exc: Exception,
    *,
    payload: dict[str, Any] | None,
    mcp_adapter: HermesTaskScopedMCPAdapter,
    mcp_registration: TaskScopedMCPRegistration,
    adapter: HermesAgentAdapter | None,
) -> dict[str, Any]:
    cleanup_result = mcp_adapter.cleanup_registration(mcp_registration)
    output: dict[str, Any] = {
        "ok": False,
        "error": str(exc),
        "session_id": _string(payload.get("session_id")) if isinstance(payload, dict) else "",
        "artifact_tool_events": adapter.artifact_tool_events() if adapter is not None else [],
        "mcp_cleanup_errors": list(cleanup_result.errors),
        "mcp_cleanup_status": cleanup_result.status,
    }
    if isinstance(exc, HermesContractError):
        output["failure_class"] = "hermes_contract_failed"
        output["contract_status"] = exc.status.to_dict()
    return output


def main() -> None:
    _configure_logging()
    exit_code = 0
    payload: dict[str, Any] | None = None
    adapter: HermesAgentAdapter | None = None
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

        adapter = HermesAgentAdapter(payload)
        output = adapter.execute()
        output["artifact_tool_events"] = adapter.artifact_tool_events()
        cleanup_result = mcp_adapter.cleanup_registration(mcp_registration)
        output["mcp_cleanup_errors"] = list(cleanup_result.errors)
        output["mcp_cleanup_status"] = cleanup_result.status
        _persist_result(output, result_path)
        _result(output)
    except Exception as exc:  # pragma: no cover - exercised via subprocess integration
        logger.exception("native Hermes executor worker failed")
        exit_code = 1
        try:
            result_path = _string(payload.get("result_path")) if isinstance(payload, dict) else ""
            output = _error_output(
                exc,
                payload=payload,
                mcp_adapter=mcp_adapter,
                mcp_registration=mcp_registration,
                adapter=adapter,
            )
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
