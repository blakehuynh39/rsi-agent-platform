from __future__ import annotations

from dataclasses import dataclass, field
import importlib.metadata
import inspect
import json
import os
from pathlib import Path
import random
import sqlite3
import subprocess
import sys
import tempfile
import time
from typing import Any

from .db_utils import float_env, sqlite_error_is_locked
from .file_utils import _fsync_parent
from .json_types import JsonObject
from .rsi_tools import normalize_tool_names

try:
    from run_agent import AIAgent  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    AIAgent = None

try:
    from hermes_state import SessionDB  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    SessionDB = None

try:
    from hermes_cli.plugins import discover_plugins, get_plugin_manager  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    discover_plugins = None
    get_plugin_manager = None

try:
    from toolsets import validate_toolset  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    validate_toolset = None

try:
    import self_review_contracts as hermes_self_review_contracts  # type: ignore
    import self_review_queue as hermes_self_review_queue  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    hermes_self_review_contracts = None
    hermes_self_review_queue = None


_REQUIRED_AIAgent_INIT_PARAMS = frozenset(
    {
        "api_key",
        "api_mode",
        "base_url",
        "enabled_toolsets",
        "ephemeral_system_prompt",
        "max_iterations",
        "model",
        "parent_session_id",
        "platform",
        "provider",
        "quiet_mode",
        "reasoning_callback",
        "reasoning_config",
        "request_overrides",
        "session_db",
        "session_id",
        "user_id",
        "chat_id",
        "thread_id",
        "gateway_session_key",
        "skip_context_files",
        "skip_memory",
        "self_review_mode",
        "status_callback",
        "stream_delta_callback",
        "thinking_callback",
        "tool_complete_callback",
        "tool_gen_callback",
        "tool_progress_callback",
        "tool_start_callback",
    }
)
_REQUIRED_RUN_CONVERSATION_PARAMS = frozenset({"conversation_history", "task_id", "user_message"})
_REQUIRED_SELF_REVIEW_QUEUE_APIS = frozenset(
    {
        "SelfReviewConfig",
        "advance_local_review_queue",
        "apply_turn_review_candidate",
        "drain_review_queue",
        "mark_candidate_delivered",
        "promote_review_candidate",
        "recover_blocked_work",
        "run_memory_self_review",
        "run_skill_self_review_batch",
        "review_queue_status",
    }
)
_PARTIAL_COMPLETION_TERMINATION_REASONS = frozenset(
    {
        "task_timeout",
        "inactivity_timeout",
        "iteration_budget_exhausted",
        "output_token_budget_exhausted",
    }
)
def _string(value: Any) -> str:
    if value is None:
        return ""
    return str(value).strip()


def _stream_text(value: Any) -> str:
    if value is None:
        return ""
    return str(value)


def _json_object(value: Any) -> JsonObject:
    return value if isinstance(value, dict) else {}


def _json_list(value: Any) -> list[Any]:
    return value if isinstance(value, list) else []


def _json_object_list(value: Any) -> list[JsonObject]:
    return [item for item in _json_list(value) if isinstance(item, dict)]


def _string_list(value: Any) -> list[str]:
    return [_string(item) for item in _json_list(value) if _string(item)]


def _bool(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        text = value.strip().lower()
        if text in {"1", "true", "t", "yes", "y", "on"}:
            return True
        if text in {"0", "false", "f", "no", "n", "off"}:
            return False
    return False


def _int(value: Any) -> int:
    if isinstance(value, bool):
        return 0
    try:
        return int(value)
    except (TypeError, ValueError):
        return 0


def _jsonable(value: Any) -> Any:
    try:
        json.dumps(value)
        return value
    except TypeError:
        if isinstance(value, dict):
            return {str(key): _jsonable(item) for key, item in value.items()}
        if isinstance(value, (list, tuple, set)):
            return [_jsonable(item) for item in value]
        return str(value)


def _bounded_jsonable(value: Any, *, limit: int = 100000) -> Any:
    jsonable = _jsonable(value)
    if isinstance(jsonable, str):
        if len(jsonable) > limit:
            return jsonable[: max(0, limit - 1)] + "…"
        return jsonable
    if isinstance(jsonable, dict):
        return {str(key): _bounded_jsonable(item, limit=limit) for key, item in jsonable.items()}
    if isinstance(jsonable, list):
        return [_bounded_jsonable(item, limit=limit) for item in jsonable]
    return jsonable


def _read_direct_url_commit() -> tuple[str, str]:
    try:
        dist = importlib.metadata.distribution("hermes-agent")
    except importlib.metadata.PackageNotFoundError:
        return "", ""
    version = _string(getattr(dist, "version", ""))
    raw = dist.read_text("direct_url.json") or ""
    if not raw:
        return version, ""
    try:
        parsed = json.loads(raw)
    except json.JSONDecodeError:
        return version, ""
    vcs_info = parsed.get("vcs_info") if isinstance(parsed, dict) else None
    if not isinstance(vcs_info, dict):
        return version, ""
    return version, _string(vcs_info.get("commit_id"))


def _hermes_config_enables_plugin(hermes_home: str, plugin_name: str) -> bool:
    config_path = Path(hermes_home or os.getenv("HERMES_HOME", "~/.hermes")).expanduser() / "config.yaml"
    try:
        text = config_path.read_text(encoding="utf-8")
    except OSError:
        return False
    return plugin_name in text


def _status_text(ok: bool) -> str:
    return "ok" if ok else "failed"


def _runtime_api_key() -> str | None:
    return os.getenv("RSI_OPENROUTER_API_KEY") or os.getenv("OPENROUTER_API_KEY") or None


@dataclass
class HermesContractStatus:
    ok: bool
    expected_pin: str
    installed_commit: str
    hermes_version: str
    api_signature_status: str
    pin_status: str
    plugin_status: str
    required_toolsets: list[str]
    toolset_status: dict[str, str]
    session_db_status: str
    self_review_api_status: str = "unknown"
    errors: list[str] = field(default_factory=list)
    checked_at_unix: float = 0.0

    def to_dict(self) -> JsonObject:
        return {
            "ok": self.ok,
            "expected_pin": self.expected_pin,
            "installed_commit": self.installed_commit,
            "hermes_version": self.hermes_version,
            "api_signature_status": self.api_signature_status,
            "pin_status": self.pin_status,
            "plugin_status": self.plugin_status,
            "required_toolsets": list(self.required_toolsets),
            "toolset_status": dict(self.toolset_status),
            "session_db_status": self.session_db_status,
            "self_review_api_status": self.self_review_api_status,
            "errors": list(self.errors),
            "checked_at_unix": self.checked_at_unix,
        }


class HermesContractError(RuntimeError):
    def __init__(self, status: HermesContractStatus) -> None:
        self.status = status
        super().__init__("; ".join(status.errors) or "Hermes adapter contract failed")


class _LifecycleWriter:
    def __init__(self, hermes_home: str, session_id: str) -> None:
        self._path = Path(hermes_home).expanduser() / "rsi_runtime" / "lifecycle" / f"{session_id}.jsonl"
        self._session_id = session_id

    @property
    def path(self) -> Path:
        return self._path

    def record(self, event: str, payload: JsonObject | None = None) -> None:
        item = {
            "event": _string(event),
            "event_type": _string(event),
            "recorded_at_unix": time.time(),
            "session_id": self._session_id,
            **(payload or {}),
        }
        self._path.parent.mkdir(parents=True, exist_ok=True)
        with self._path.open("a", encoding="utf-8") as handle:
            handle.write(json.dumps(item, ensure_ascii=True, sort_keys=True, default=str) + "\n")


def validate_hermes_contract(
    *,
    expected_pin: str,
    hermes_home: str,
    session_db: Any,
    required_toolsets: list[str] | None = None,
    require_platform_runtime: bool = False,
    require_session_db_ready: bool = True,
) -> HermesContractStatus:
    errors: list[str] = []
    expected = _string(expected_pin)
    version, installed_commit = _read_direct_url_commit()
    api_signature_status = "unknown"
    pin_status = "unknown"
    plugin_status = "unknown"
    self_review_api_status = "unknown"
    session_db_status = "ok" if session_db is not None else "missing"
    if session_db is None:
        errors.append("Hermes SessionDB is unavailable.")
    else:
        session_db_status, session_db_error = _validate_session_db_integrity(
            session_db,
            require_ready=require_session_db_ready,
        )
        if session_db_error:
            errors.append(session_db_error)

    if AIAgent is None:
        api_signature_status = "missing"
        errors.append("run_agent.AIAgent is unavailable.")
    else:
        init_params = set(inspect.signature(AIAgent.__init__).parameters)
        run_params = set(inspect.signature(AIAgent.run_conversation).parameters)
        missing_init = sorted(_REQUIRED_AIAgent_INIT_PARAMS - init_params)
        missing_run = sorted(_REQUIRED_RUN_CONVERSATION_PARAMS - run_params)
        if missing_init or missing_run:
            api_signature_status = "failed"
            if missing_init:
                errors.append("AIAgent.__init__ missing required parameter(s): " + ", ".join(missing_init))
            if missing_run:
                errors.append("AIAgent.run_conversation missing required parameter(s): " + ", ".join(missing_run))
        else:
            api_signature_status = "ok"

    if not require_platform_runtime:
        self_review_api_status = "not_required"
    elif hermes_self_review_contracts is None or hermes_self_review_queue is None:
        self_review_api_status = "missing"
        errors.append("Hermes self-review contract/queue APIs are unavailable.")
    else:
        missing_apis = sorted(
            name for name in _REQUIRED_SELF_REVIEW_QUEUE_APIS if not hasattr(hermes_self_review_queue, name)
        )
        if not hasattr(hermes_self_review_contracts, "SelfReviewObservationV1"):
            missing_apis.append("SelfReviewObservationV1")
        if missing_apis:
            self_review_api_status = "failed"
            errors.append("Hermes self-review APIs missing required symbol(s): " + ", ".join(missing_apis))
        else:
            self_review_api_status = "ok"

    if not expected:
        pin_status = "missing_expected_pin"
        errors.append("RSI_HERMES_PIN is required.")
    elif not installed_commit:
        pin_status = "missing_installed_commit"
        errors.append("Installed hermes-agent direct_url.json did not expose a git commit.")
    elif installed_commit != expected:
        pin_status = "mismatch"
        errors.append(f"Installed hermes-agent commit {installed_commit} does not match RSI_HERMES_PIN {expected}.")
    else:
        pin_status = "ok"

    required_plugins = ["rsi_context_engine", "company_knowledge"]
    if require_platform_runtime:
        required_plugins.append("rsi_platform_runtime")
    plugin_records: dict[str, JsonObject] = {}
    if callable(discover_plugins):
        try:
            discover_plugins(force=True)
            if callable(get_plugin_manager):
                for item in get_plugin_manager().list_plugins():
                    if not isinstance(item, dict):
                        continue
                    name = _string(item.get("name"))
                    if name:
                        plugin_records[name] = item
        except Exception as exc:
            plugin_status = "failed"
            errors.append(f"Hermes plugin discovery failed: {exc}")
        else:
            missing_config: list[str] = []
            missing_loaded: list[str] = []
            invalid_capabilities: list[str] = []
            for plugin_name in required_plugins:
                if not _hermes_config_enables_plugin(hermes_home, plugin_name):
                    missing_config.append(plugin_name)
                    continue
                record = plugin_records.get(plugin_name)
                if not record or not _bool(record.get("enabled")):
                    missing_loaded.append(plugin_name)
                    continue
                if plugin_name == "rsi_platform_runtime":
                    capabilities = _json_object(record.get("capabilities"))
                    if not _bool(capabilities.get("execution_scoped_context_supported")):
                        invalid_capabilities.append(plugin_name)
            if missing_config:
                plugin_status = "config_missing"
                errors.append("Hermes config is missing required plugin(s): " + ", ".join(missing_config) + ".")
            elif missing_loaded:
                plugin_status = "not_loaded"
                errors.append("Hermes did not load required plugin(s): " + ", ".join(missing_loaded) + ".")
            elif invalid_capabilities:
                plugin_status = "capability_missing"
                errors.append("Hermes plugin capability check failed for: " + ", ".join(invalid_capabilities) + ".")
            else:
                plugin_status = "ok"
    else:
        plugin_status = "missing_discovery_api"
        errors.append("Hermes plugin discovery API is unavailable.")

    requested_toolsets = normalize_tool_names(required_toolsets or [])
    toolset_status: dict[str, str] = {}
    if validate_toolset is None and requested_toolsets:
        errors.append("Hermes toolset validation API is unavailable.")
    for toolset in requested_toolsets:
        if validate_toolset is None:
            toolset_status[toolset] = "validate_api_missing"
            continue
        try:
            ok = bool(validate_toolset(toolset))
        except Exception as exc:
            ok = False
            toolset_status[toolset] = f"error: {exc}"
        else:
            toolset_status[toolset] = _status_text(ok)
        if not ok:
            errors.append(f"Required Hermes toolset is unavailable: {toolset}.")

    return HermesContractStatus(
        ok=not errors,
        expected_pin=expected,
        installed_commit=installed_commit,
        hermes_version=version,
        api_signature_status=api_signature_status,
        pin_status=pin_status,
        plugin_status=plugin_status,
        required_toolsets=requested_toolsets,
        toolset_status=toolset_status,
        session_db_status=session_db_status,
        self_review_api_status=self_review_api_status,
        errors=errors,
        checked_at_unix=time.time(),
    )


def _validate_session_db_integrity(session_db: Any, *, require_ready: bool = True) -> tuple[str, str]:
    db_path = getattr(session_db, "db_path", None)
    if not db_path:
        return "ok", ""
    try:
        path = Path(db_path)
    except TypeError:
        return "ok", ""
    if not path.exists():
        if not require_ready:
            return "uninitialized", ""
        return "missing", f"Hermes SessionDB file is missing at {path}."
    if not require_ready and bool(getattr(session_db, "defer_integrity_check", False)):
        return "deferred", ""
    deadline = time.monotonic() + float_env(
        "RSI_HERMES_SESSION_DB_INTEGRITY_RETRY_SECONDS",
        float_env("RSI_HERMES_SESSION_DB_READ_RETRY_SECONDS", 10.0),
    )
    try:
        while True:
            try:
                with sqlite3.connect(str(path), timeout=5.0) as conn:
                    row = conn.execute("PRAGMA quick_check(1)").fetchone()
                    tables = {
                        _string(item[0])
                        for item in conn.execute("SELECT name FROM sqlite_master WHERE type = 'table'").fetchall()
                    }
                break
            except sqlite3.OperationalError as exc:
                if not sqlite_error_is_locked(exc):
                    raise
                if time.monotonic() >= deadline:
                    return "locked", ""
                time.sleep(random.uniform(0.05, 0.25))
    except sqlite3.OperationalError as exc:
        return "corrupt", f"Hermes SessionDB integrity check failed: {exc}"
    except sqlite3.DatabaseError as exc:
        return "corrupt", f"Hermes SessionDB integrity check failed: {exc}"
    except Exception as exc:
        return "failed", f"Hermes SessionDB integrity check failed: {exc}"
    verdict = _string(row[0] if row else "")
    if verdict.lower() == "ok":
        required_tables = {"sessions", "messages"}
        missing_tables = sorted(required_tables - tables)
        if missing_tables:
            if not require_ready:
                return "uninitialized", ""
            return "missing_schema", "Hermes SessionDB schema is missing required table(s): " + ", ".join(missing_tables)
        return "ok", ""
    return "corrupt", f"Hermes SessionDB integrity check failed: {verdict or 'unknown quick_check failure'}"


class HermesAgentAdapter:
    def __init__(self, payload: JsonObject) -> None:
        self._payload = dict(payload)
        self._session_id = _string(payload.get("session_id"))
        self._hermes_home = _string(os.getenv("HERMES_HOME")) or _string(payload.get("hermes_home")) or str(Path.home() / ".hermes")
        self._lifecycle = _LifecycleWriter(self._hermes_home, self._session_id or "unknown")

    @property
    def lifecycle_path(self) -> Path:
        return self._lifecycle.path

    def execute(self) -> JsonObject:
        session_id = self._session_id
        if not session_id:
            raise RuntimeError("session_id is required for native Hermes execution.")
        session_db = self._open_session_db()
        status = validate_hermes_contract(
            expected_pin=_string(os.getenv("RSI_HERMES_PIN")),
            hermes_home=self._hermes_home,
            session_db=session_db,
            required_toolsets=[_string(item) for item in _json_list(self._payload.get("toolsets")) if _string(item)],
            require_platform_runtime=True,
        )
        if not status.ok:
            raise HermesContractError(status)
        system_prompt = self._system_prompt()
        self._prepare_session_db(session_id, session_db, system_prompt)
        agent = self._create_agent(session_id, session_db, system_prompt)
        conversation_history = self._conversation_history(session_db)
        self._lifecycle.record(
            "model.call.started",
            {
                "engine": "hermes_aiagent_subprocess",
                "toolsets": [_string(item) for item in _json_list(self._payload.get("toolsets")) if _string(item)],
                "contract_status": status.to_dict(),
            },
        )
        try:
            resume_payload = _json_object(self._payload.get("external_tool_resume"))
            if resume_payload:
                content = resume_payload.get("content")
                if not isinstance(content, str):
                    content = json.dumps(content if content is not None else {}, ensure_ascii=True, sort_keys=True)
                resume_history = _json_object_list(resume_payload.get("transcript_snapshot")) or conversation_history
                result = agent.resume_with_tool_result(
                    session_id=_string(resume_payload.get("session_id")) or session_id,
                    tool_call_id=_string(resume_payload.get("tool_call_id")),
                    tool_name=_string(resume_payload.get("tool_name")),
                    content=content,
                    conversation_history=resume_history,
                    task_id=session_id,
                    metadata=_json_object(resume_payload.get("metadata")),
                )
            elif repair_instruction := _string(self._payload.get("repair_instruction")):
                repair_callable = getattr(agent, "repair_with_instructions", None)
                if not callable(repair_callable):
                    raise RuntimeError("AIAgent.repair_with_instructions is required for clean repair mode.")
                result = repair_callable(
                    instructions=repair_instruction,
                    conversation_history=conversation_history,
                    task_id=session_id,
                )
            else:
                result = agent.run_conversation(
                    user_message=_string(self._payload.get("prompt")),
                    conversation_history=conversation_history,
                    task_id=session_id,
                )
        except Exception as exc:
            self._lifecycle.record(
                "model.call.failed",
                {
                    "engine": "hermes_aiagent_subprocess",
                    "status": "failed",
                    "error": str(exc),
                },
            )
            raise
        response = ""
        if isinstance(result, dict):
            response = _string(result.get("final_response"))
        if not response:
            response = _string(result)
        completion_meta = self._completion_meta(result)
        reply_delivery = {}
        self_review_candidate = self._write_self_review_candidate(result if isinstance(result, dict) else {})
        safe_result = dict(result) if isinstance(result, dict) else {"value": response}
        safe_result.pop("self_review_observation", None)
        self._lifecycle.record(
            "model.call.completed",
            {
                "engine": "hermes_aiagent_subprocess",
                "status": "completed",
                "termination_reason": completion_meta["termination_reason"],
                "completion_verdict": completion_meta["completion_verdict"],
                "self_review_candidate_status": _string(self_review_candidate.get("candidate_status"))
                or _string(self_review_candidate.get("status")),
            },
        )
        out = {
            "ok": not (isinstance(result, dict) and bool(result.get("failed"))),
            "response": response,
            "result": safe_result,
            "session_id": session_id,
            "contract_status": status.to_dict(),
            "self_review_candidate": self_review_candidate,
            "reply_delivery": reply_delivery,
            **completion_meta,
        }
        if isinstance(result, dict) and bool(result.get("suppress_delivery")):
            out["suppress_delivery"] = True
        if isinstance(result, dict) and _string(result.get("external_tool_pause_id")):
            out["external_tool_pause_id"] = _string(result.get("external_tool_pause_id"))
        return out

    def _write_self_review_candidate(self, result: JsonObject) -> JsonObject:
        observation = _json_object(result.get("self_review_observation"))
        if not observation:
            return {}
        runner_execution_id = _string(self._payload.get("execution_id"))
        if runner_execution_id:
            observation = {**observation, "execution_id": runner_execution_id}
        timeout_seconds = float(os.getenv("RSI_HERMES_SELF_REVIEW_CANDIDATE_WRITE_TIMEOUT_SECONDS") or "2.5")
        payload = {"observation": observation}
        temp_path = ""
        try:
            with tempfile.NamedTemporaryFile("w", encoding="utf-8", suffix=".json", delete=False) as handle:
                temp_path = handle.name
                json.dump(payload, handle, ensure_ascii=True, sort_keys=True)
                handle.flush()
                os.fsync(handle.fileno())
            completed = subprocess.run(
                [
                    sys.executable,
                    "-m",
                    "rsi_runner.hermes_self_review_candidate_writer",
                    temp_path,
                ],
                check=False,
                capture_output=True,
                text=True,
                timeout=max(0.1, timeout_seconds),
            )
            if completed.returncode != 0:
                stderr = _string(completed.stderr)
                try:
                    parsed_error = _json_object(
                        json.loads(completed.stdout.strip().splitlines()[-1])
                        if completed.stdout.strip()
                        else {}
                    )
                except json.JSONDecodeError:
                    parsed_error = {}
                self._lifecycle.record(
                    "self_review.candidate.write_failed",
                    {
                        "status": "failed",
                        "error": stderr[:800] or _string(parsed_error.get("error"))[:800],
                        "ineligible_reason": _string(parsed_error.get("ineligible_reason")),
                    },
                )
                return {
                    "candidate_status": "candidate_write_failed",
                    "status": "candidate_write_failed",
                    "ineligible_reason": _string(parsed_error.get("ineligible_reason")),
                    "error": stderr[:800] or _string(parsed_error.get("error"))[:800],
                }
            parsed = json.loads(completed.stdout.strip().splitlines()[-1] if completed.stdout.strip() else "{}")
            if not isinstance(parsed, dict):
                raise ValueError("candidate writer returned non-object JSON")
            self._lifecycle.record(
                "self_review.candidate.written",
                {
                    "status": _string(parsed.get("candidate_status")) or _string(parsed.get("status")),
                    "candidate_id": parsed.get("candidate_id"),
                    "execution_id": _string(parsed.get("execution_id")),
                    "snapshot_ref": _string(parsed.get("snapshot_ref")),
                    "snapshot_hash": _string(parsed.get("snapshot_hash")),
                    "snapshot_size": parsed.get("snapshot_size"),
                },
            )
            return {
                key: value
                for key, value in parsed.items()
                if key
                in {
                    "candidate_id",
                    "candidate_status",
                    "status",
                    "execution_id",
                    "agent_identity",
                    "snapshot_ref",
                    "snapshot_hash",
                    "snapshot_size",
                    "ineligible_reason",
                    "gateway_session_key",
                    "cadence_scope_key",
                    "memory_turn_delta",
                    "skill_iteration_delta",
                    "skill_iteration_delta_after_last_skill_manage",
                    "memory_nudge_interval",
                    "skill_nudge_interval",
                    "memory_tool_used",
                    "skill_manage_used",
                    "memory_eligible",
                    "skill_eligible",
                }
            }
        except subprocess.TimeoutExpired:
            self._lifecycle.record(
                "self_review.candidate.write_failed",
                {"status": "timeout", "timeout_seconds": timeout_seconds},
            )
            return {"candidate_status": "candidate_write_failed", "status": "candidate_write_failed", "error": "candidate write timeout"}
        except Exception as exc:
            self._lifecycle.record(
                "self_review.candidate.write_failed",
                {"status": "failed", "error": str(exc)[:800]},
            )
            return {"candidate_status": "candidate_write_failed", "status": "candidate_write_failed", "error": str(exc)[:800]}
        finally:
            if temp_path:
                try:
                    Path(temp_path).unlink(missing_ok=True)
                except OSError:
                    pass

    def artifact_tool_events(self) -> list[JsonObject]:
        events: list[JsonObject] = []
        if not self.lifecycle_path.exists():
            return events
        for line in self.lifecycle_path.read_text(encoding="utf-8").splitlines():
            if not line.strip():
                continue
            try:
                parsed = json.loads(line)
            except json.JSONDecodeError:
                continue
            if not isinstance(parsed, dict):
                continue
            event_type = _string(parsed.get("event_type")) or _string(parsed.get("event"))
            if not event_type.startswith("artifact."):
                continue
            payload = _json_object(parsed.get("payload"))
            events.append(
                {
                    "event_type": event_type,
                    "status": _string(parsed.get("status")),
                    "payload": payload,
                }
            )
        return events

    def _open_session_db(self) -> Any:
        if SessionDB is None:
            return None
        db_path = Path(self._hermes_home).expanduser() / "state.db"
        deadline = time.monotonic() + float_env("RSI_HERMES_SESSION_DB_OPEN_RETRY_SECONDS", 20.0)
        attempt = 0
        while True:
            try:
                return SessionDB(db_path=db_path)
            except Exception as exc:
                if not sqlite_error_is_locked(exc) or time.monotonic() >= deadline:
                    raise
                attempt += 1
                if attempt == 1:
                    self._lifecycle.record(
                        "session_db.open_retry",
                        {
                            "status": "retrying",
                            "db_path": str(db_path),
                            "error": str(exc)[:800],
                            "retry_deadline_seconds": float_env("RSI_HERMES_SESSION_DB_OPEN_RETRY_SECONDS", 20.0),
                        },
                    )
                time.sleep(random.uniform(0.05, 0.25))

    def _conversation_history(self, session_db: Any) -> list[JsonObject]:
        payload_history = [item for item in _json_list(self._payload.get("conversation_history")) if isinstance(item, dict)]
        if payload_history:
            return payload_history
        if _string(self._payload.get("execution_phase")) in {"render", "deliver"}:
            return []
        if session_db is None:
            return []
        try:
            return [item for item in list(session_db.get_messages_as_conversation(self._session_id) or []) if isinstance(item, dict)]
        except Exception as exc:
            self._lifecycle.record(
                "session_db.history_read_failed",
                {"status": "warning", "error": str(exc)[:800]},
            )
            return []

    def _prepare_session_db(self, session_id: str, db: Any, system_prompt: str) -> None:
        if db is None:
            return
        parent_session_id = _string(self._payload.get("parent_session_id")) or None
        model = _string(self._payload.get("model")) or None
        if parent_session_id == session_id:
            self._lifecycle.record(
                "session_db.parent_session_ignored",
                {
                    "status": "warning",
                    "reason": "parent_session_id matched session_id",
                    "session_id": session_id,
                },
            )
            parent_session_id = None
        if parent_session_id:
            self._ensure_parent_session(db, parent_session_id, model=model)
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

    def _ensure_parent_session(self, db: Any, parent_session_id: str, *, model: str | None = None) -> None:
        try:
            if hasattr(db, "ensure_session"):
                db.ensure_session(parent_session_id, source="rsi_executor_parent", model=model)
            elif self._session_exists(db, parent_session_id):
                pass
            else:
                db.create_session(parent_session_id, source="rsi_executor_parent", model=model)
        except sqlite3.IntegrityError:
            if self._session_exists(db, parent_session_id):
                self._lifecycle.record(
                    "session_db.parent_session_ensure_raced",
                    {
                        "status": "completed",
                        "parent_session_id": parent_session_id,
                    },
                )
                return
            raise
        except Exception as exc:
            self._lifecycle.record(
                "session_db.parent_session_ensure_failed",
                {
                    "status": "failed",
                    "parent_session_id": parent_session_id,
                    "error": str(exc)[:800],
                },
            )
            raise
        self._lifecycle.record(
            "session_db.parent_session_ensured",
            {
                "status": "completed",
                "parent_session_id": parent_session_id,
            },
        )

    def _session_exists(self, db: Any, session_id: str) -> bool:
        getter = getattr(db, "get_session", None)
        if not callable(getter):
            return False
        try:
            return getter(session_id) is not None
        except Exception as exc:
            self._lifecycle.record(
                "session_db.session_lookup_failed",
                {
                    "status": "warning",
                    "session_id": session_id,
                    "error": str(exc)[:800],
                },
            )
            return False

    def _system_prompt(self) -> str:
        prompt_parts = [_string(self._payload.get("system_message"))]
        return "\n\n".join(part for part in prompt_parts if part)

    def _create_agent(self, session_id: str, session_db: Any, system_prompt: str) -> Any:
        if AIAgent is None:
            raise RuntimeError("run_agent.AIAgent is unavailable.")
        runtime = _json_object(self._payload.get("runtime"))
        runtime_provider = _string(runtime.get("provider")) or "openrouter"
        if runtime_provider != "openrouter":
            raise RuntimeError("Hermes native execution only supports the OpenRouter runtime provider.")
        agent_kwargs: JsonObject = {
            "model": _string(self._payload.get("model")),
            "api_key": _runtime_api_key(),
            "base_url": _string(runtime.get("base_url")) or None,
            "provider": runtime_provider,
            "api_mode": _string(runtime.get("api_mode")) or None,
            "max_iterations": max(1, int(self._payload.get("max_iterations") or 1)),
            "enabled_toolsets": _string_list(self._payload.get("toolsets")),
            "quiet_mode": True,
            "ephemeral_system_prompt": system_prompt or None,
            "reasoning_config": _json_object(self._payload.get("reasoning_config")),
            "request_overrides": _json_object(self._payload.get("request_overrides")),
            "session_id": session_id,
            "parent_session_id": _string(self._payload.get("parent_session_id")) or None,
            "platform": "rsi",
            "user_id": _string(self._payload.get("user_peer_id")) or None,
            "chat_id": _string(self._payload.get("chat_id")) or None,
            "thread_id": _string(self._payload.get("thread_id")) or None,
            "gateway_session_key": _string(self._payload.get("gateway_session_key")) or None,
            "session_db": session_db,
            "skip_context_files": True,
            "skip_memory": False,
            "self_review_mode": "manual",
            "reasoning_callback": self._reasoning_callback,
            "stream_delta_callback": self._stream_delta_callback,
            "thinking_callback": self._thinking_callback,
            "tool_gen_callback": self._tool_generation_callback,
            "tool_progress_callback": self._tool_progress_callback,
            "tool_start_callback": self._tool_start_callback,
            "tool_complete_callback": self._tool_complete_callback,
            "status_callback": self._status_callback,
        }
        provider_routing = _json_object(runtime.get("provider_routing"))
        if provider_routing:
            allowed = _string_list(provider_routing.get("only"))
            ignored = _string_list(provider_routing.get("ignore"))
            order = _string_list(provider_routing.get("order"))
            if allowed:
                agent_kwargs["providers_allowed"] = allowed
            if ignored:
                agent_kwargs["providers_ignored"] = ignored
            if order:
                agent_kwargs["providers_order"] = order
            if _string(provider_routing.get("sort")):
                agent_kwargs["provider_sort"] = _string(provider_routing.get("sort"))
            if "require_parameters" in provider_routing:
                agent_kwargs["provider_require_parameters"] = _bool(provider_routing.get("require_parameters"))
            if _string(provider_routing.get("data_collection")):
                agent_kwargs["provider_data_collection"] = _string(provider_routing.get("data_collection"))
        return AIAgent(**agent_kwargs)

    def _completion_meta(self, result: Any) -> JsonObject:
        if not isinstance(result, dict):
            return {
                "termination_reason": "normal_completion",
                "completion_verdict": "complete",
                "max_iterations_reached": False,
                "native_result_completed": True,
                "native_result_partial": False,
                "native_result_interrupted": False,
                "native_result_api_calls": 0,
            }
        completed_value = result.get("completed")
        completed_known = isinstance(completed_value, bool)
        completed = completed_value if completed_known else True
        partial = _bool(result.get("partial"))
        interrupted = _bool(result.get("interrupted"))
        api_calls = _int(result.get("api_calls"))
        max_iterations = _int(self._payload.get("max_iterations"))
        explicit_termination_reason = _string(result.get("termination_reason"))
        explicit_completion_verdict = _string(result.get("completion_verdict"))
        timeout_kind = _string(result.get("timeout_kind"))
        incomplete_without_reason = partial or interrupted or (completed_known and not completed)
        max_iterations_reached = (
            explicit_termination_reason == "iteration_budget_exhausted"
            or (max_iterations > 0 and api_calls >= max_iterations and not completed)
        )
        termination_reason = explicit_termination_reason or (
            "iteration_budget_exhausted" if (max_iterations_reached or incomplete_without_reason) else "normal_completion"
        )
        partial_stop = termination_reason in _PARTIAL_COMPLETION_TERMINATION_REASONS
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

    def _record_lifecycle(self, event_type: str, *, status: str = "", payload: JsonObject | None = None) -> None:
        self._lifecycle.record(
            event_type,
            {
                "status": _string(status),
                "payload": _bounded_jsonable(payload or {}),
            },
        )

    def _reasoning_callback(self, text: Any) -> None:
        self._record_lifecycle(
            "model.reasoning.delta",
            status="streaming",
            payload={"delta": _stream_text(text)},
        )

    def _stream_delta_callback(self, text: Any) -> None:
        if text is None:
            self._record_lifecycle("model.output.delta", status="completed", payload={"delta": ""})
            return
        self._record_lifecycle(
            "model.output.delta",
            status="streaming",
            payload={"delta": _stream_text(text)},
        )

    def _thinking_callback(self, text: Any) -> None:
        message = _string(text)
        self._record_lifecycle(
            "model.thinking",
            status="running" if message else "completed",
            payload={"text": message},
        )

    def _tool_generation_callback(self, tool_name: Any) -> None:
        self._record_lifecycle(
            "tool.generation.started",
            status="running",
            payload={"tool_name": _string(tool_name)},
        )

    def _tool_progress_callback(self, *args: Any, **kwargs: Any) -> None:
        progress_event = _string(args[0] if args else kwargs.get("event"))
        tool_name = _string(args[1] if len(args) > 1 else kwargs.get("tool_name"))
        preview = _string(args[2] if len(args) > 2 else kwargs.get("preview"))
        tool_args = args[3] if len(args) > 3 else kwargs.get("args")
        event_type = "tool.call.progress"
        status = "running"
        if progress_event == "_thinking":
            event_type = "model.thinking"
            status = "running" if preview else "completed"
        elif progress_event == "reasoning.available":
            event_type = "model.reasoning.delta"
            status = "completed"
        elif progress_event == "tool.completed":
            status = "failed" if bool(kwargs.get("is_error")) else "completed"
        self._record_lifecycle(
            event_type,
            status=status,
            payload={
                "progress_event": progress_event,
                "tool_name": tool_name,
                "preview": preview,
                "args": _bounded_jsonable(tool_args),
                "duration_seconds": kwargs.get("duration"),
                "is_error": bool(kwargs.get("is_error")),
                "raw_args": _bounded_jsonable(list(args)),
                "raw_kwargs": _bounded_jsonable(kwargs),
            },
        )

    def _tool_start_callback(self, tool_call_id: Any, tool_name: Any, args: Any) -> None:
        self._record_lifecycle(
            "tool.call.started",
            status="running",
            payload={
                "tool_call_id": _string(tool_call_id),
                "tool_name": _string(tool_name),
                "args": _bounded_jsonable(args),
            },
        )

    def _tool_complete_callback(self, tool_call_id: Any, tool_name: Any, args: Any, result: Any) -> None:
        result_text = _string(result)
        failed = result_text.strip().lower().startswith("error:")
        self._record_lifecycle(
            "tool.call.completed",
            status="failed" if failed else "completed",
            payload={
                "tool_call_id": _string(tool_call_id),
                "tool_name": _string(tool_name),
                "args": _bounded_jsonable(args),
                "result": _bounded_jsonable(result),
            },
        )

    def _status_callback(self, *args: Any, **kwargs: Any) -> None:
        status_kind = _string(args[0] if args else kwargs.get("status_kind"))
        message = _string(args[1] if len(args) > 1 else kwargs.get("message"))
        self._record_lifecycle(
            "model.status",
            status=status_kind or "status",
            payload={
                "status_kind": status_kind,
                "message": message,
                "raw_args": _bounded_jsonable(list(args)),
                "raw_kwargs": _bounded_jsonable(kwargs),
            },
        )
