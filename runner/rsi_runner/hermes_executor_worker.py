from __future__ import annotations

import json
import logging
import os
from pathlib import Path
import sys
from typing import Any

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


def _load_request(path_arg: str) -> dict[str, Any]:
    path = Path(path_arg).expanduser().resolve()
    parsed = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(parsed, dict):
        raise ValueError("executor worker request must be a JSON object")
    return parsed


def _configure_logging() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")


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
    try:
        if len(sys.argv) < 2:
            raise RuntimeError("executor worker request path is required")
        payload = _load_request(sys.argv[1])
        workdir = _string(payload.get("workdir"))
        if workdir:
            os.chdir(workdir)
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
        _result(
            {
                "ok": not (isinstance(result, dict) and bool(result.get("failed"))),
                "response": response,
                "result": result if isinstance(result, dict) else {"value": response},
                "session_id": session_id,
            }
        )
    except Exception as exc:  # pragma: no cover - exercised via subprocess integration
        logger.exception("native Hermes executor worker failed")
        _result(
            {
                "ok": False,
                "error": str(exc),
                "session_id": "",
            }
        )
        raise SystemExit(1) from exc


if __name__ == "__main__":
    main()
