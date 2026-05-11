from __future__ import annotations

from dataclasses import dataclass
import importlib.metadata
import json
import os
from pathlib import Path
import tempfile

from .json_types import JsonObject

from .config import RunnerConfig
from .file_utils import _atomic_write_json, _fsync_parent
from .rsi_tools import rsi_plugin_toolset_definitions

try:
    from hermes_cli.plugins import discover_plugins  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    discover_plugins = None


_PLUGIN_MANIFEST = """name: rsi_context_engine
version: "1.2.0"
description: "RSI context injection and lifecycle capture"
provides_hooks:
  - pre_llm_call
  - on_session_start
  - on_session_end
capabilities:
  execution_scoped_context_supported: true
"""


_PLUGIN_MODULE_TEMPLATE = """from __future__ import annotations

import json
import hashlib
import os
import time
from pathlib import Path
from urllib import error as urlerror
from urllib import parse as urlparse
from urllib import request as urlrequest

from rsi_runner.grafana_observability import (
    active_alerts as grafana_active_alerts,
    alert_rule_get as grafana_alert_rule_get,
    alert_rules_search as grafana_alert_rules_search,
    dashboard_get as grafana_dashboard_get,
    dashboards_search as grafana_dashboards_search,
    datasources_query as grafana_datasources_query,
    loki_log_lines as grafana_loki_log_lines,
    logs_query as grafana_logs_query,
    metrics_query as grafana_metrics_query,
)
from rsi_runner.slack_uploads import prepare_local_slack_upload_payload

try:
    from tools.external_tool_pause import ExternalToolPending
except (ImportError, ModuleNotFoundError):
    ExternalToolPending = None


_PLUGIN_TOOLS = __PLUGIN_TOOLS__
_TRANSPORT_TO_CANONICAL = {item["transport_name"]: item["canonical_name"] for item in _PLUGIN_TOOLS}
_ARTIFACT_CANONICAL_NAMES = {"artifact.list_files", "artifact.write_file"}
_DB_READ_CANONICAL_NAMES = {"db_read.sources", "db_read.schema", "db_read.validate", "db_read.query", "db_read.status"}
_OBSERVABILITY_CANONICAL_NAMES = {
    "rsi_observability.datasources",
    "rsi_observability.metrics_query",
    "rsi_observability.logs_query",
    "rsi_observability.dashboards_search",
    "rsi_observability.dashboard_get",
    "rsi_observability.alert_rules_search",
    "rsi_observability.alert_rule_get",
    "rsi_observability.active_alerts",
}
_NATIVE_CANONICAL_PREFIXES = ("rsi_slack.", "rsi_notion.", "rsi_knowledge.", "rsi_sentry.")


def _runtime_root() -> Path:
    hermes_home = Path(os.getenv("HERMES_HOME", "~/.hermes")).expanduser()
    return hermes_home / "rsi_runtime"


def _context_path(session_id: str) -> Path:
    explicit_path = os.getenv("RSI_RUNTIME_CONTEXT_PATH", "").strip()
    if explicit_path:
        return Path(explicit_path).expanduser()
    if os.getenv("RSI_EXECUTION_ID", "").strip():
        raise RuntimeError("RSI_RUNTIME_CONTEXT_PATH is required for RSI workflow execution context.")
    return _runtime_root() / "context" / f"{session_id}.json"


def _lifecycle_path(session_id: str) -> Path:
    return _runtime_root() / "lifecycle" / f"{session_id}.jsonl"


def _append_event(session_id: str, event: str, payload: JsonObject) -> None:
    path = _lifecycle_path(session_id)
    path.parent.mkdir(parents=True, exist_ok=True)
    item = {
        "event": event,
        "recorded_at_unix": time.time(),
        **(payload or {}),
    }
    with path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(item, sort_keys=True) + "\\n")


def _load_context(session_id: str) -> JsonObject:
    path = _context_path(session_id)
    if not path.exists():
        if os.getenv("RSI_EXECUTION_ID", "").strip() or os.getenv("RSI_RUNTIME_CONTEXT_PATH", "").strip():
            raise RuntimeError(f"Missing RSI workflow context payload at {path}.")
        return {}
    try:
        parsed = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"Invalid RSI context payload for session {session_id}.") from exc
    if not isinstance(parsed, dict):
        raise RuntimeError(f"Invalid RSI context payload for session {session_id}.")
    _verify_execution_context(parsed, path)
    return parsed


def _verify_execution_context(payload: JsonObject, path: Path) -> None:
    for key, env_name in (
        ("execution_id", "RSI_EXECUTION_ID"),
        ("operation_id", "RSI_OPERATION_ID"),
        ("trace_id", "RSI_TRACE_ID"),
        ("workflow_id", "RSI_WORKFLOW_ID"),
    ):
        expected = os.getenv(env_name, "").strip()
        if not expected:
            continue
        actual = str(payload.get(key, "") or "").strip()
        if actual != expected:
            raise RuntimeError(f"RSI context {path} has {key}={actual!r}, expected {expected!r}.")


def _active_context(task_id: str = "", **kwargs) -> JsonObject:
    candidates = [
        str(task_id or "").strip(),
        str(kwargs.get("task_id", "") or "").strip(),
        str(kwargs.get("session_id", "") or "").strip(),
    ]
    seen: set[str] = set()
    for candidate in candidates:
        if not candidate or candidate in seen:
            continue
        seen.add(candidate)
        payload = _load_context(candidate)
        if payload:
            return payload
    return {}


def _artifact_tools_available() -> bool:
    return True


def _db_read_auth_token() -> str:
    return os.getenv("RSI_DB_READ_EXECUTION_TOKEN", "").strip()


def _db_read_tools_available() -> bool:
    return bool(os.getenv("RSI_CONTROL_PLANE_BASE_URL", "").strip() and _db_read_auth_token())


def _native_tools_available() -> bool:
    return bool(_control_plane_base_url() and os.getenv("RSI_NATIVE_TOOLS_EXECUTION_TOKEN", "").strip())


def _observability_tools_available() -> bool:
    base_url = os.getenv("GRAFANA_SERVER", "").strip() or os.getenv("RSI_GRAFANA_BASE_URL", "").strip()
    token = os.getenv("GRAFANA_TOKEN", "").strip() or os.getenv("RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN", "").strip()
    return bool(base_url and token)


def _control_plane_base_url() -> str:
    return os.getenv("RSI_CONTROL_PLANE_BASE_URL", "").strip().rstrip("/")


def _native_tools_execution_token() -> str:
    return os.getenv("RSI_NATIVE_TOOLS_EXECUTION_TOKEN", "").strip()


def _native_surface_and_operation(canonical_name: str) -> tuple[str, str]:
    if canonical_name.startswith("rsi_slack."):
        return "slack", canonical_name.split(".", 1)[1]
    if canonical_name.startswith("rsi_notion."):
        return "notion", canonical_name.split(".", 1)[1]
    if canonical_name.startswith("rsi_knowledge."):
        return "knowledge", canonical_name.split(".", 1)[1]
    if canonical_name.startswith("rsi_sentry."):
        return "sentry", canonical_name.split(".", 1)[1]
    raise ValueError(f"unknown native RSI tool {canonical_name!r}")


def _rsi_tool_available(canonical_name: str) -> bool:
    if canonical_name in _ARTIFACT_CANONICAL_NAMES:
        return _artifact_tools_available()
    if canonical_name in _DB_READ_CANONICAL_NAMES:
        return _db_read_tools_available()
    if canonical_name in _OBSERVABILITY_CANONICAL_NAMES:
        return _observability_tools_available()
    if canonical_name.startswith(_NATIVE_CANONICAL_PREFIXES):
        return _native_tools_available()
    return False


def _first_non_empty(*values) -> str:
    for value in values:
        text = str(value or "").strip()
        if text:
            return text
    return ""


def _string_value(value) -> str:
    return str(value or "").strip()


def _bool_value(value) -> bool:
    if isinstance(value, bool):
        return value
    return _string_value(value).lower() in {"1", "true", "yes", "on"}


def _bool_value_or_default(value, default: bool) -> bool:
    return default if value is None else _bool_value(value)


def _int_value(value, default: int) -> int:
    try:
        parsed = int(value)
    except (TypeError, ValueError):
        return default
    return parsed if parsed > 0 else default


def _string_list(value) -> list[str]:
    if not isinstance(value, list):
        return []
    out: list[str] = []
    seen: set[str] = set()
    for item in value:
        text = _string_value(item)
        if not text or text in seen:
            continue
        seen.add(text)
        out.append(text)
    return out


def _artifact_output_dir(payload: JsonObject) -> Path:
    raw = str(payload.get("artifact_output_dir", "") or "").strip()
    if not raw:
        raise RuntimeError("artifact_output_dir unavailable for this session")
    root = Path(raw).expanduser().resolve()
    root.mkdir(parents=True, exist_ok=True)
    return root


def _artifact_output_dir_text(payload: JsonObject) -> str:
    try:
        return str(_artifact_output_dir(payload))
    except Exception:
        return str(payload.get("artifact_output_dir", "") or "").strip()


def _artifact_path(payload: JsonObject, requested_path) -> tuple[str, Path]:
    root = _artifact_output_dir(payload)
    raw = str(requested_path or "").strip()
    candidate = Path(raw).expanduser() if raw.startswith("/") else (root / raw)
    resolved = candidate.resolve()
    try:
        resolved.relative_to(root)
    except ValueError as exc:
        raise ValueError("artifact_path_outside_root") from exc
    return raw, resolved


def _artifact_entry(path: Path) -> JsonObject:
    stat = path.stat()
    return {
        "name": path.name,
        "path": str(path),
        "is_dir": path.is_dir(),
        "size_bytes": 0 if path.is_dir() else stat.st_size,
    }


def _artifact_handler(canonical_name: str, transport_name: str, args: JsonObject, task_id: str = "", **kwargs):
    payload = _active_context(task_id=task_id, **kwargs)
    session_id = str(task_id or kwargs.get("task_id", "") or kwargs.get("session_id", "")).strip()
    safe_args = args if isinstance(args, dict) else {}
    if canonical_name == "artifact.list_files":
        requested_path = str(safe_args.get("path", "") or "").strip()
        try:
            requested_path, target = _artifact_path(payload, requested_path)
            entries = []
            if target.exists():
                if target.is_dir():
                    entries = [_artifact_entry(item) for item in sorted(target.iterdir())]
                else:
                    entries = [_artifact_entry(target)]
            output = {
                "artifact_output_dir": str(_artifact_output_dir(payload)),
                "requested_path": requested_path,
                "path": str(target),
                "entries": entries,
            }
            if session_id:
                _append_event(
                    session_id,
                    "artifact.list.completed",
                    {
                        "event_type": "artifact.list.completed",
                        "status": "completed",
                        "payload": {
                            "tool_name": canonical_name,
                            "transport_tool_name": transport_name,
                            **output,
                        },
                    },
                )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "ok",
                    "summary": f"Listed {len(entries)} item(s) in the native artifact directory.",
                    "output": output,
                },
                sort_keys=True,
            )
        except Exception as exc:
            failed_payload = {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "requested_path": requested_path,
                "artifact_output_dir": _artifact_output_dir_text(payload),
                "error": str(exc),
            }
            if session_id:
                _append_event(
                    session_id,
                    "artifact.list.failed",
                    {
                        "event_type": "artifact.list.failed",
                        "status": "failed",
                        "payload": failed_payload,
                    },
                )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "error",
                    "error": str(exc),
                    "output": failed_payload,
                },
                sort_keys=True,
            )
    if canonical_name == "artifact.write_file":
        requested_path = str(safe_args.get("path", "") or "").strip()
        try:
            artifact_dir = str(_artifact_output_dir(payload))
            started_payload = {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "requested_path": requested_path,
                "artifact_output_dir": artifact_dir,
            }
            if session_id:
                _append_event(
                    session_id,
                    "artifact.write.started",
                    {
                        "event_type": "artifact.write.started",
                        "status": "running",
                        "payload": started_payload,
                    },
                )
            _, target = _artifact_path(payload, requested_path)
            target.parent.mkdir(parents=True, exist_ok=True)
            content = str(safe_args.get("content", "") or "")
            target.write_text(content, encoding="utf-8")
            content_bytes = content.encode("utf-8")
            file_ref = f"file://{target}"
            completed_payload = {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "path": str(target),
                "workspace_path": str(target),
                "file_ref": file_ref,
                "artifact_output_dir": str(_artifact_output_dir(payload)),
                "bytes_written": len(content_bytes),
                "size_bytes": len(content_bytes),
                "sha256": hashlib.sha256(content_bytes).hexdigest(),
                "created_by_execution_id": _first_non_empty(
                    payload.get("execution_id"),
                    payload.get("operation_id"),
                    payload.get("trace_id"),
                ),
                "share_status": "local",
            }
            if session_id:
                _append_event(
                    session_id,
                    "artifact.write.completed",
                    {
                        "event_type": "artifact.write.completed",
                        "status": "completed",
                        "payload": completed_payload,
                    },
                )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "ok",
                    "summary": f"Wrote artifact file to {target}.",
                    "output": completed_payload,
                    "raw_artifact_refs": [file_ref],
                },
                sort_keys=True,
            )
        except Exception as exc:
            try:
                artifact_dir_for_error = str(_artifact_output_dir(payload))
            except Exception:
                artifact_dir_for_error = ""
            failed_payload = {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "requested_path": requested_path,
                "artifact_output_dir": artifact_dir_for_error or _artifact_output_dir_text(payload),
                "error": str(exc),
            }
            if session_id:
                _append_event(
                    session_id,
                    "artifact.write.failed",
                    {
                        "event_type": "artifact.write.failed",
                        "status": "failed",
                        "payload": failed_payload,
                    },
                )
            return json.dumps(
                {
                    "tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "status": "error",
                    "error": str(exc),
                    "output": failed_payload,
                },
                sort_keys=True,
            )
    return json.dumps({
        "tool_name": canonical_name,
        "transport_tool_name": transport_name,
        "status": "error",
        "error": "unknown RSI artifact tool",
    }, sort_keys=True)


def _db_read_request_json(method: str, path: str, body: JsonObject | None = None) -> JsonObject:
    base_url = os.getenv("RSI_CONTROL_PLANE_BASE_URL", "").strip().rstrip("/")
    token = _db_read_auth_token()
    if not base_url:
        raise RuntimeError("RSI_CONTROL_PLANE_BASE_URL is required")
    if not token:
        raise RuntimeError("RSI_DB_READ_EXECUTION_TOKEN is required")
    data = None
    headers = {"Authorization": f"Bearer {token}", "Accept": "application/json"}
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urlrequest.Request(base_url + path, data=data, headers=headers, method=method)
    try:
        with urlrequest.urlopen(req, timeout=30) as resp:
            raw = resp.read()
    except urlerror.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        if exc.code == 409:
            try:
                parsed = json.loads(detail)
                if isinstance(parsed, dict):
                    return parsed
            except json.JSONDecodeError:
                pass
        raise RuntimeError(f"db-read request failed: HTTP {exc.code}: {detail}") from exc
    except urlerror.URLError as exc:
        raise RuntimeError(f"db-read request failed: {exc}") from exc
    if not raw:
        return {}
    parsed = json.loads(raw.decode("utf-8"))
    if not isinstance(parsed, dict):
        raise RuntimeError("db-read response was not a JSON object")
    return parsed


def _db_read_external_pause_event(payload: JsonObject) -> JsonObject:
    request = payload.get("request")
    if not isinstance(request, dict):
        return {}
    validation = payload.get("validation")
    if isinstance(validation, dict) and validation.get("ok") is False:
        return {}
    state = str(request.get("state") or payload.get("status") or "").strip()
    if state == "validation_failed":
        return {}
    request_id = str(request.get("id") or "").strip()
    if not request_id:
        return {}
    pause = payload.get("external_tool_pause")
    if not isinstance(pause, dict):
        return {}
    pause_id = str(pause.get("id") or "").strip()
    if not pause_id:
        return {}
    return {
        "kind": "external_tool_pending",
        "external_tool_pause_id": pause_id,
        "external_action_id": pause_id,
        "request_ref": request_id,
        "request_id": request_id,
        "target": str(request.get("target") or "").strip(),
        "state": state,
        "sql_sha256": str(request.get("sql_sha256") or "").strip(),
        "status": str(payload.get("status") or state).strip(),
        "tool_name": str(pause.get("transport_tool_name") or "db_read_query").strip(),
        "tool_call_id": str(pause.get("tool_call_id") or "").strip(),
        "session_id": str(pause.get("hermes_session_id") or "").strip(),
        "summary": "DB read request is pending Slack admin approval and execution.",
    }


def _record_external_tool_pending(session_id: str, event: JsonObject) -> None:
    if not event:
        return
    if session_id:
        _append_event(
            session_id,
            "external_tool.pending",
            {
                "event_type": "external_tool.pending",
                "status": "pending",
                "payload": dict(event),
            },
        )


def _db_read_args_hash(target: str, sql: str, purpose: str) -> str:
    raw = json.dumps(
        {"target": target.strip(), "sql": sql.strip(), "purpose": (purpose or "query").strip() or "query"},
        ensure_ascii=True,
        sort_keys=True,
        separators=(",", ":"),
    ).encode("utf-8")
    return "sha256:" + hashlib.sha256(raw).hexdigest()


def _db_read_handler(canonical_name: str, transport_name: str, args: JsonObject, task_id: str = "", **kwargs):
    session_id = str(task_id or kwargs.get("task_id", "") or kwargs.get("session_id", "")).strip()
    tool_call_id = str(kwargs.get("tool_call_id", "") or "").strip()
    safe_args = args if isinstance(args, dict) else {}
    try:
        if canonical_name == "db_read.sources":
            payload = _db_read_request_json("GET", "/internal/db-read/sources")
        elif canonical_name == "db_read.schema":
            target = _string_value(safe_args.get("target"))
            payload = _db_read_request_json("GET", "/internal/db-read/schema?" + urlparse.urlencode({"target": target}))
        elif canonical_name == "db_read.validate":
            payload = _db_read_request_json(
                "POST",
                "/internal/db-read/validate",
                {
                    "target": _string_value(safe_args.get("target")),
                    "sql": str(safe_args.get("sql") or ""),
                    "purpose": _string_value(safe_args.get("purpose")) or "query",
                },
            )
        elif canonical_name == "db_read.query":
            target = _string_value(safe_args.get("target"))
            sql = str(safe_args.get("sql") or "")
            purpose = _string_value(safe_args.get("purpose")) or "query"
            payload = _db_read_request_json(
                "POST",
                "/internal/db-read/query",
                {
                    "target": target,
                    "sql": sql,
                    "purpose": purpose,
                    "hermes_session_id": session_id,
                    "hermes_tool_call_id": tool_call_id,
                    "canonical_tool_name": canonical_name,
                    "transport_tool_name": transport_name,
                    "args_hash": _db_read_args_hash(target, sql, purpose),
                    "operation_id": os.getenv("RSI_OPERATION_ID", "").strip(),
                    "execution_id": os.getenv("RSI_EXECUTION_ID", "").strip(),
                    "conversation_id": os.getenv("RSI_CONVERSATION_ID", "").strip(),
                    "workflow_id": os.getenv("RSI_WORKFLOW_ID", "").strip(),
                    "trace_id": os.getenv("RSI_TRACE_ID", "").strip(),
                    "channel_id": os.getenv("RSI_SLACK_CHANNEL_ID", "").strip(),
                    "thread_ts": os.getenv("RSI_SLACK_THREAD_TS", "").strip(),
                },
            )
            pending = _db_read_external_pause_event(payload)
            if pending:
                if ExternalToolPending is None:
                    raise RuntimeError("Hermes external tool pause support is unavailable for db_read.query")
                _record_external_tool_pending(session_id, pending)
                raise ExternalToolPending(pending)
            validation = payload.get("validation")
            request = payload.get("request")
            if isinstance(validation, dict) and validation.get("ok") is False:
                payload = {
                    "status": "validation_failed",
                    "request": request,
                    "validation": validation,
                    "message": "DB read SQL validation failed; repair the SQL and call db_read.query again.",
                }
            else:
                raise RuntimeError("db-read query did not create an external tool pause")
        elif canonical_name == "db_read.status":
            request_id = urlparse.quote(_string_value(safe_args.get("request_id")))
            payload = _db_read_request_json("GET", f"/internal/db-read/requests/{request_id}")
        else:
            raise RuntimeError("unknown RSI DB-read tool")
        output = {
            "tool_name": canonical_name,
            "transport_tool_name": transport_name,
            "status": "ok",
            "output": payload,
        }
        return json.dumps(output, sort_keys=True)
    except Exception as exc:
        if ExternalToolPending is not None and isinstance(exc, ExternalToolPending):
            raise
        failed_payload = {
            "tool_name": canonical_name,
            "transport_tool_name": transport_name,
            "error": str(exc),
        }
        if session_id:
            _append_event(
                session_id,
                "db_read.tool_failed",
                {
                    "event_type": "db_read.tool_failed",
                    "status": "failed",
                    "payload": failed_payload,
                },
            )
        return json.dumps(
            {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "status": "error",
                "error": str(exc),
                "output": failed_payload,
            },
            sort_keys=True,
        )


def _default_payload(canonical_name: str, payload: JsonObject) -> JsonObject:
    task_repo = str(payload.get("task_repo", "") or payload.get("repo", "")).strip()
    task_repo_ref = str(payload.get("task_repo_ref", "") or payload.get("repo_ref", "")).strip()
    task_prompt = str(payload.get("task_prompt", "") or payload.get("prompt", "")).strip()
    task_channel_id = str(payload.get("task_channel_id", "") or payload.get("channel_id", "")).strip()
    task_thread_ts = str(payload.get("task_thread_ts", "") or payload.get("thread_ts", "")).strip()
    context_summary = str(payload.get("context_summary", "") or "").strip()
    default_question = str(payload.get("task_default_question", "") or task_prompt).strip()
    repo_question = str(payload.get("task_repo_question", "") or default_question or context_summary).strip()
    knowledge_topic = str(payload.get("task_knowledge_topic", "") or context_summary or task_repo).strip()
    knowledge_question = str(payload.get("task_knowledge_question", "") or default_question or context_summary).strip()
    slack_history_focus = str(payload.get("task_slack_history_focus", "") or default_question).strip()
    slack_search_query = str(payload.get("task_slack_search_query", "") or default_question).strip()
    session_scope_kind = str(payload.get("session_scope_kind", "") or "").strip()
    session_scope_id = str(payload.get("session_scope_id", "") or "").strip()
    trace_id = str(payload.get("trace_id", "") or "").strip()
    attempt_id = str(payload.get("attempt_id", "") or "").strip()
    workspace_id = str(payload.get("workspace_id", "") or "").strip()
    out: JsonObject = {}
    if trace_id:
        out["trace_id"] = trace_id
    if canonical_name in {"repo.context", "repo.read_file", "repo.search", "github.repo_activity", "github.repo_context"} and task_repo:
        out["repo"] = task_repo
    if canonical_name in {"repo.read_file", "repo.search"} and task_repo_ref:
        out["ref"] = task_repo_ref
    if canonical_name == "repo.context" and repo_question:
        out["question"] = repo_question
    if canonical_name == "knowledge.context":
        if knowledge_question:
            out["question"] = knowledge_question
        if knowledge_topic:
            out["topic"] = knowledge_topic
        out["scope_id"] = task_repo
    if canonical_name == "slack.history":
        surface = _default_slack_surface(payload)
        if surface.get("channel_id"):
            out["channel_id"] = str(surface.get("channel_id", "")).strip()
        if surface.get("thread_ts"):
            out["thread_ts"] = str(surface.get("thread_ts", "")).strip()
        if slack_history_focus:
            out["question"] = slack_history_focus
    if canonical_name == "slack.search":
        channel_ids = _default_slack_channel_ids(payload)
        if channel_ids:
            out["channel_ids"] = channel_ids
        if slack_search_query:
            out["query"] = slack_search_query
    if canonical_name == "slack.upload_file":
        surface = _default_slack_surface(payload)
        channel_id = str(surface.get("channel_id", "") or task_channel_id).strip()
        thread_ts = str(surface.get("thread_ts", "") or task_thread_ts).strip()
        if channel_id:
            out["channel_id"] = channel_id
        if thread_ts:
            out["thread_ts"] = thread_ts
    if canonical_name == "rsi.workflow_context":
        out["trace_id"] = trace_id
    if canonical_name == "rsi.action_chain":
        out["trace_id"] = trace_id
    if canonical_name == "rsi.runner_execution":
        out["trace_id"] = trace_id
    if canonical_name == "sentry.lookup":
        out["alert"] = context_summary or task_prompt
    if canonical_name in {"rsi.proposal_memory", "rsi.candidate_context"} and session_scope_kind == "proposal_candidate":
        out["candidate_key"] = session_scope_id
    if canonical_name == "rsi.attempt_context" and attempt_id:
        out["attempt_id"] = attempt_id
    if canonical_name.startswith("workspace."):
        if workspace_id:
            out["workspace_id"] = workspace_id
        if attempt_id:
            out["attempt_id"] = attempt_id
    return out


def _candidate_read_surfaces(payload: JsonObject) -> list[JsonObject]:
    refs = payload.get("context_refs") or []
    seen: set[str] = set()
    out: list[JsonObject] = []
    for item in refs:
        if not isinstance(item, dict):
            continue
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
        out.append(candidate)
    task_channel_id = str(payload.get("task_channel_id", "") or payload.get("channel_id", "")).strip()
    task_thread_ts = str(payload.get("task_thread_ts", "") or payload.get("thread_ts", "")).strip()
    if task_channel_id:
        fallback = {
            "channel_id": task_channel_id,
            "thread_ts": task_thread_ts,
            "ref": "",
            "source": "task_binding",
        }
        encoded = json.dumps(fallback, sort_keys=True)
        if encoded not in seen:
            out.insert(0, fallback)
    return out


def _default_slack_surface(payload: JsonObject) -> JsonObject:
    candidates = _candidate_read_surfaces(payload)
    if candidates:
        return candidates[0]
    return {}


def _default_slack_channel_ids(payload: JsonObject) -> list[str]:
    out: list[str] = []
    seen: set[str] = set()
    for item in _candidate_read_surfaces(payload):
        channel_id = str(item.get("channel_id", "")).strip()
        if not channel_id or channel_id in seen:
            continue
        seen.add(channel_id)
        out.append(channel_id)
    return out


def _allowed_upload_roots(payload: JsonObject) -> list[Path]:
    roots: list[Path] = []
    for value in [
        payload.get("artifact_output_dir"),
        payload.get("hermes_computer_root"),
        payload.get("hermes_artifact_root"),
    ]:
        text = _string_value(value)
        if text:
            roots.append(Path(text).expanduser().resolve())
    workspace_policy = payload.get("workspace_policy")
    if isinstance(workspace_policy, dict):
        for item in workspace_policy.get("allowed_path_roots") or []:
            text = _string_value(item)
            if text:
                roots.append(Path(text).expanduser().resolve())
    deduped: list[Path] = []
    seen: set[str] = set()
    for root in roots:
        key = str(root)
        if key not in seen:
            seen.add(key)
            deduped.append(root)
    return deduped


def _path_is_under(path: Path, root: Path) -> bool:
    try:
        path.relative_to(root)
        return True
    except ValueError:
        return False


def _resolve_slack_upload_path(payload: JsonObject) -> Path | None:
    raw_path = _first_non_empty(_string_value(payload.get("path")), _string_value(payload.get("artifact_ref")))
    if not raw_path:
        return None
    candidate = raw_path.strip()
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
    path = path.resolve() if path.is_absolute() else (Path.cwd() / path).resolve()
    roots = _allowed_upload_roots(payload)
    if not roots:
        raise ValueError("slack.upload_file no allowed upload roots configured")
    if not any(_path_is_under(path, root) for root in roots):
        raise ValueError("slack.upload_file local file is outside allowed workspace roots")
    if not path.exists():
        raise ValueError(f"slack.upload_file local file does not exist: {path}")
    if not path.is_file():
        raise ValueError(f"slack.upload_file local path is not a file: {path}")
    return path


def _resolve_slack_upload_payload(payload: JsonObject) -> JsonObject:
    if _string_value(payload.get("content")) or _string_value(payload.get("content_base64")):
        return payload
    resolved_path = _resolve_slack_upload_path(payload)
    if resolved_path is None:
        return payload
    return prepare_local_slack_upload_payload(payload, resolved_path)


def _payload_with_upload_context(args: JsonObject, context: JsonObject) -> JsonObject:
    payload = dict(args or {})
    for key in ("artifact_output_dir", "hermes_computer_root", "hermes_artifact_root", "workspace_policy"):
        if key not in payload and key in context:
            payload[key] = context[key]
    return payload


def _strip_upload_context(payload: JsonObject) -> JsonObject:
    cleaned = dict(payload or {})
    for key in ("artifact_output_dir", "hermes_computer_root", "hermes_artifact_root", "workspace_policy"):
        cleaned.pop(key, None)
    return cleaned


def _resolve_slack_report_attachment_payload(item: object, context: JsonObject) -> object:
    if not isinstance(item, dict):
        return item
    payload = _payload_with_upload_context(dict(item), context)
    resolved = _resolve_slack_upload_payload(payload)
    return _strip_upload_context(resolved)


def _resolve_native_action_args(canonical_name: str, args: JsonObject, context: JsonObject) -> JsonObject:
    if canonical_name == "rsi_slack.file_upload":
        return _strip_upload_context(_resolve_slack_upload_payload(_payload_with_upload_context(args, context)))
    if canonical_name == "rsi_slack.report_post":
        resolved = dict(args or {})
        for key in ("files", "images"):
            value = resolved.get(key)
            if isinstance(value, list):
                resolved[key] = [_resolve_slack_report_attachment_payload(item, context) for item in value]
        report = resolved.get("report")
        if isinstance(report, dict):
            report_payload = dict(report)
            for key in ("files", "images"):
                value = report_payload.get(key)
                if isinstance(value, list):
                    report_payload[key] = [_resolve_slack_report_attachment_payload(item, context) for item in value]
            resolved["report"] = report_payload
        return resolved
    return args


def _native_target_ref(args: JsonObject) -> str:
    for key in (
        "target_ref",
        "channel_id",
        "page_id",
        "database_id",
        "data_source_id",
        "block_id",
        "source_ref",
        "page_ref",
        "slug",
        "issue",
        "issue_ref",
        "short_id",
        "project_ref",
        "project",
        "org",
        "release",
    ):
        value = _string_value(args.get(key))
        if value:
            return value
    return ""


def _native_action_payload(canonical_name: str, args: JsonObject) -> JsonObject:
    surface, operation = _native_surface_and_operation(canonical_name)
    safe_args = dict(args or {})
    payload: JsonObject = {
        "surface": surface,
        "operation": operation,
        "target_ref": _native_target_ref(safe_args),
        "arguments": safe_args,
    }
    for key in ("reason", "idempotency_key", "confirm_destroy"):
        if key in safe_args:
            payload[key] = safe_args[key]
    destructive_ops = {
        "rsi_slack.message_delete",
        "rsi_slack.channel_archive",
        "rsi_notion.page_archive",
        "rsi_notion.block_delete",
    }
    if canonical_name in destructive_ops:
        payload["destructive"] = True
    return payload


def _native_action_request(payload: JsonObject) -> tuple[int, JsonObject]:
    base_url = _control_plane_base_url()
    token = _native_tools_execution_token()
    if not base_url:
        raise RuntimeError("RSI_CONTROL_PLANE_BASE_URL is required for RSI native tools")
    if not token:
        raise RuntimeError("RSI_NATIVE_TOOLS_EXECUTION_TOKEN is required for RSI native tools")
    body = json.dumps(payload, ensure_ascii=True, sort_keys=True).encode("utf-8")
    req = urlrequest.Request(
        base_url + "/internal/native-tools/actions",
        data=body,
        headers={
            "Authorization": "Bearer " + token,
            "Content-Type": "application/json",
        },
        method="POST",
    )
    try:
        with urlrequest.urlopen(req, timeout=60) as response:
            raw = response.read().decode("utf-8")
            parsed = json.loads(raw) if raw else {}
            return int(response.status), parsed if isinstance(parsed, dict) else {"raw": parsed}
    except urlerror.HTTPError as exc:
        raw = exc.read().decode("utf-8", errors="replace")
        try:
            parsed = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            parsed = {"error": raw}
        if not isinstance(parsed, dict):
            parsed = {"raw": parsed}
        return int(exc.code), parsed


def _native_action_response_map(response: JsonObject, key: str) -> JsonObject:
    value = response.get(key)
    return value if isinstance(value, dict) else {}


def _native_slack_reply_delivery(
    canonical_name: str,
    transport_name: str,
    args: JsonObject,
    response: JsonObject,
    *,
    ok: bool,
    status_code: int,
    tool_call_id: str,
) -> JsonObject:
    if canonical_name not in {"rsi_slack.message_post", "rsi_slack.report_post"}:
        return {}
    output = _native_action_response_map(response, "output")
    action = _native_action_response_map(response, "action")
    manifest = output.get("render_manifest") if isinstance(output.get("render_manifest"), dict) else {}
    main_message = manifest.get("main_message") if isinstance(manifest.get("main_message"), dict) else {}
    action_id = _string_value(action.get("id"))
    body = _first_non_empty(
        args.get("text"),
        args.get("summary"),
        action.get("response_summary"),
        response.get("error"),
    )
    source_ref = _first_non_empty(
        action.get("source_ref"),
        main_message.get("source_ref"),
        output.get("source_ref"),
    )
    channel_id = _first_non_empty(
        output.get("channel_id"),
        main_message.get("channel_id"),
        args.get("channel_id"),
    )
    thread_ts = _first_non_empty(
        args.get("thread_ts"),
        output.get("thread_ts"),
        main_message.get("thread_ts"),
        output.get("ts"),
        main_message.get("ts"),
    )
    artifact_refs: list[str] = []
    if action_id:
        artifact_refs.append("external_tool_action:" + action_id)
    uploaded_files = output.get("uploaded_files")
    if isinstance(uploaded_files, list):
        for item in uploaded_files:
            if isinstance(item, dict):
                source = _string_value(item.get("source_ref"))
                if source:
                    artifact_refs.append(source)
    delivery_id = _first_non_empty(tool_call_id, action_id, args.get("idempotency_key"))
    delivery: JsonObject = {
        "tool_name": canonical_name,
        "transport_tool_name": transport_name,
        "tool_call_id": delivery_id,
        "channel_id": channel_id,
        "thread_ts": thread_ts,
        "body": body,
        "body_excerpt": body[:600],
        "body_sha1": hashlib.sha1(body.encode("utf-8")).hexdigest() if body else "",
        "send_status": "posted" if ok else "failed",
        "status_code": status_code,
        "provider_ref": source_ref,
        "artifact_refs": artifact_refs,
    }
    if action_id:
        delivery["native_action_id"] = action_id
    if output.get("renderer_version") is not None:
        delivery["renderer_version"] = output.get("renderer_version")
    if response.get("replayed"):
        delivery["replayed"] = True
    return delivery


def _native_action_handler(canonical_name: str, transport_name: str, args: JsonObject, task_id: str = "", **kwargs):
    session_id = str(task_id or kwargs.get("task_id", "") or kwargs.get("session_id", "")).strip()
    tool_call_id = str(kwargs.get("tool_call_id", "") or kwargs.get("call_id", "") or "").strip()
    context = _active_context(task_id=task_id, **kwargs)
    safe_args = args if isinstance(args, dict) else {}
    try:
        safe_args = _resolve_native_action_args(canonical_name, safe_args, context)
        request_payload = _native_action_payload(canonical_name, safe_args)
        status_code, response = _native_action_request(request_payload)
        ok = 200 <= status_code < 300 and bool(response.get("ok", False))
        event_payload = {
            "tool_name": canonical_name,
            "transport_tool_name": transport_name,
            "status_code": status_code,
            "ok": ok,
            "action": response.get("action"),
        }
        reply_delivery: JsonObject = {}
        if session_id:
            _append_event(
                session_id,
                "native_tool.completed" if ok else "native_tool.failed",
                {
                    "event_type": "native_tool.completed" if ok else "native_tool.failed",
                    "status": "completed" if ok else "failed",
                    "payload": event_payload,
                },
            )
            reply_delivery = _native_slack_reply_delivery(
                canonical_name,
                transport_name,
                safe_args,
                response,
                ok=ok,
                status_code=status_code,
                tool_call_id=tool_call_id,
            )
            if reply_delivery:
                _append_event(
                    session_id,
                    "reply_delivery",
                    {
                        "event_type": "reply_delivery",
                        "status": reply_delivery.get("send_status", "posted" if ok else "failed"),
                        **reply_delivery,
                    },
                )
        return json.dumps(
            {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "status": "ok" if ok else "error",
                "status_code": status_code,
                "summary": response.get("action", {}).get("response_summary", ""),
                "error": response.get("error", ""),
                "reply_delivery": reply_delivery,
                "output": response,
            },
            sort_keys=True,
        )
    except Exception as exc:
        if session_id:
            _append_event(
                session_id,
                "native_tool.failed",
                {
                    "event_type": "native_tool.failed",
                    "status": "failed",
                    "payload": {
                        "tool_name": canonical_name,
                        "transport_tool_name": transport_name,
                        "error": str(exc),
                    },
                },
            )
        return json.dumps(
            {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "status": "error",
                "error": str(exc),
            },
            sort_keys=True,
        )


def _observability_summary(canonical_name: str, output: JsonObject, lines: list[str]) -> str:
    if canonical_name == "rsi_observability.datasources":
        datasources = output.get("datasources")
        count = len(datasources) if isinstance(datasources, list) else 0
        return f"Returned {count} Grafana datasource(s)."
    if canonical_name == "rsi_observability.logs_query":
        return f"Returned {len(lines)} Loki log line(s)."
    if canonical_name == "rsi_observability.dashboards_search":
        dashboards = output.get("dashboards")
        count = len(dashboards) if isinstance(dashboards, list) else 0
        return f"Returned {count} Grafana dashboard(s)."
    if canonical_name == "rsi_observability.dashboard_get":
        dashboard = output.get("dashboard")
        title = _string_value(dashboard.get("title")) if isinstance(dashboard, dict) else ""
        return f"Returned Grafana dashboard{(': ' + title) if title else ''}."
    if canonical_name == "rsi_observability.alert_rules_search":
        rules = output.get("alert_rules")
        count = len(rules) if isinstance(rules, list) else 0
        return f"Returned {count} Grafana alert rule(s)."
    if canonical_name == "rsi_observability.alert_rule_get":
        title = _string_value(output.get("title"))
        return f"Returned Grafana alert rule{(': ' + title) if title else ''}."
    if canonical_name == "rsi_observability.active_alerts":
        alerts = output.get("alerts")
        count = len(alerts) if isinstance(alerts, list) else 0
        return f"Returned {count} active Grafana alert(s)."
    result = output.get("data", {}).get("result") if isinstance(output.get("data"), dict) else None
    count = len(result) if isinstance(result, list) else 0
    return f"Returned {count} Prometheus result item(s)."


def _observability_handler(canonical_name: str, transport_name: str, args: JsonObject, task_id: str = "", **kwargs):
    session_id = str(task_id or kwargs.get("task_id", "") or kwargs.get("session_id", "")).strip()
    safe_args = args if isinstance(args, dict) else {}
    lines: list[str] = []
    try:
        if canonical_name == "rsi_observability.datasources":
            output = grafana_datasources_query(_string_value(safe_args.get("type")))
        elif canonical_name == "rsi_observability.metrics_query":
            output = grafana_metrics_query(
                str(safe_args.get("expr") or ""),
                datasource=_string_value(safe_args.get("datasource")),
                range_query=_bool_value(safe_args.get("range")),
                since=_string_value(safe_args.get("since")) or "1h",
                start=_string_value(safe_args.get("start")),
                end=_string_value(safe_args.get("end")),
                step=_string_value(safe_args.get("step")),
            )
        elif canonical_name == "rsi_observability.logs_query":
            output = grafana_logs_query(
                str(safe_args.get("expr") or ""),
                datasource=_string_value(safe_args.get("datasource")),
                since=_string_value(safe_args.get("since")) or "1h",
                start=_string_value(safe_args.get("start")),
                end=_string_value(safe_args.get("end")),
                limit=_int_value(safe_args.get("limit"), 50),
                direction=_string_value(safe_args.get("direction")) or "backward",
                step=_string_value(safe_args.get("step")),
            )
            lines = grafana_loki_log_lines(output)
        elif canonical_name == "rsi_observability.dashboards_search":
            output = grafana_dashboards_search(
                _string_value(safe_args.get("query")),
                tags=_string_list(safe_args.get("tags")),
                limit=_int_value(safe_args.get("limit"), 50),
            )
        elif canonical_name == "rsi_observability.dashboard_get":
            output = grafana_dashboard_get(_string_value(safe_args.get("uid")))
        elif canonical_name == "rsi_observability.alert_rules_search":
            output = grafana_alert_rules_search(
                _string_value(safe_args.get("query")),
                folder_uid=_string_value(safe_args.get("folder_uid")),
                limit=_int_value(safe_args.get("limit"), 100),
            )
        elif canonical_name == "rsi_observability.alert_rule_get":
            output = grafana_alert_rule_get(_string_value(safe_args.get("uid")))
        elif canonical_name == "rsi_observability.active_alerts":
            output = grafana_active_alerts(
                _string_list(safe_args.get("filters")),
                active=_bool_value_or_default(safe_args.get("active"), True),
                silenced=_bool_value_or_default(safe_args.get("silenced"), False),
                inhibited=_bool_value_or_default(safe_args.get("inhibited"), False),
                limit=_int_value(safe_args.get("limit"), 100),
            )
        else:
            raise RuntimeError("unknown RSI observability tool")
        summary = _observability_summary(canonical_name, output, lines)
        if session_id:
            _append_event(
                session_id,
                "observability.query.completed",
                {
                    "event_type": "observability.query.completed",
                    "status": "completed",
                    "payload": {
                        "tool_name": canonical_name,
                        "transport_tool_name": transport_name,
                        "summary": summary,
                    },
                },
            )
        result: JsonObject = {
            "tool_name": canonical_name,
            "transport_tool_name": transport_name,
            "status": "ok",
            "summary": summary,
            "output": output,
        }
        if lines:
            result["log_lines"] = lines
        return json.dumps(result, sort_keys=True)
    except Exception as exc:
        if session_id:
            _append_event(
                session_id,
                "observability.query.failed",
                {
                    "event_type": "observability.query.failed",
                    "status": "failed",
                    "payload": {
                        "tool_name": canonical_name,
                        "transport_tool_name": transport_name,
                        "error": str(exc),
                    },
                },
            )
        return json.dumps(
            {
                "tool_name": canonical_name,
                "transport_tool_name": transport_name,
                "status": "error",
                "error": str(exc),
            },
            sort_keys=True,
        )


def _tool_handler(transport_name: str):
    canonical_name = _TRANSPORT_TO_CANONICAL[transport_name]

    def handler(args: JsonObject, task_id: str = "", **kwargs):
        if canonical_name in _DB_READ_CANONICAL_NAMES:
            return _db_read_handler(canonical_name, transport_name, args, task_id=task_id, **kwargs)
        if canonical_name in _OBSERVABILITY_CANONICAL_NAMES:
            return _observability_handler(canonical_name, transport_name, args, task_id=task_id, **kwargs)
        if canonical_name.startswith(_NATIVE_CANONICAL_PREFIXES):
            return _native_action_handler(canonical_name, transport_name, args, task_id=task_id, **kwargs)
        return _artifact_handler(canonical_name, transport_name, args, task_id=task_id, **kwargs)

    return handler


def _json_block(label: str, value: object) -> str:
    if not value:
        return ""
    return f"{label}:\\n" + json.dumps(value, ensure_ascii=True, sort_keys=True)


def _render_context(payload: JsonObject) -> str:
    parts: list[str] = []
    summary = str(payload.get("context_summary", "") or "").strip()
    if summary:
        parts.append("Context summary:\\n" + summary)
    refs = payload.get("context_refs") or []
    if refs:
        parts.append(_json_block("Context refs", refs))
    execution_mode = str(payload.get("execution_mode", "") or "").strip()
    if execution_mode:
        parts.append("Execution mode: " + execution_mode)
    for label, key in (
        ("Trace ID", "trace_id"),
        ("Workflow ID", "workflow_id"),
        ("Proposal ID", "proposal_id"),
        ("Attempt ID", "attempt_id"),
        ("Workspace ID", "workspace_id"),
    ):
        value = str(payload.get(key, "") or "").strip()
        if value:
            parts.append(f"{label}: {value}")
    return "\\n\\n".join(part for part in parts if part)


def pre_llm_call(session_id: str, **kwargs):
    payload = _load_context(session_id)
    if not payload:
        return None
    rendered = _render_context(payload)
    _append_event(
        session_id,
        "pre_llm_call",
        {
            "is_first_turn": bool(kwargs.get("is_first_turn", False)),
            "context_available": bool(rendered),
        },
    )
    if not rendered:
        return None
    return {"context": "RSI context engine\\n\\n" + rendered}


def on_session_start(session_id: str, **kwargs):
    _append_event(
        session_id,
        "on_session_start",
        {
            "model": kwargs.get("model", ""),
            "platform": kwargs.get("platform", ""),
        },
    )


def on_session_end(session_id: str, **kwargs):
    _append_event(
        session_id,
        "on_session_end",
        {
            "completed": bool(kwargs.get("completed", False)),
            "interrupted": bool(kwargs.get("interrupted", False)),
            "model": kwargs.get("model", ""),
            "platform": kwargs.get("platform", ""),
        },
    )


def register(ctx):
    for item in _PLUGIN_TOOLS:
        canonical_name = str(item["canonical_name"])
        ctx.register_tool(
            name=str(item["transport_name"]),
            toolset=str(item["toolset"]),
            schema=dict(item["schema"]),
            handler=_tool_handler(str(item["transport_name"])),
            check_fn=lambda canonical_name=canonical_name: _rsi_tool_available(canonical_name),
            description=str(item["schema"].get("description", "")),
            external_pause=canonical_name == "db_read.query",
        )
    ctx.register_hook("pre_llm_call", pre_llm_call)
    ctx.register_hook("on_session_start", on_session_start)
    ctx.register_hook("on_session_end", on_session_end)
"""


def _build_plugin_module() -> str:
    return _PLUGIN_MODULE_TEMPLATE.replace(
        "__PLUGIN_TOOLS__",
        repr(rsi_plugin_toolset_definitions()),
    )


@dataclass
class HermesAdapterMetadata:
    version: str
    pin: str
    context_engine_mode: str
    context_engine_status: str
    lifecycle_hook_status: str


class HermesContextEngine:
    def __init__(self, hermes_home: str) -> None:
        self._hermes_home = Path(hermes_home)
        self._runtime_root = self._hermes_home / "rsi_runtime"
        self._plugin_dir = self._hermes_home / "plugins" / "rsi_context_engine"
        self._context_dir = self._runtime_root / "context"
        self._lifecycle_dir = self._runtime_root / "lifecycle"
        self._status = "unknown"
        self._install_error = ""
        self._install()

    @property
    def status(self) -> str:
        return self._status

    @property
    def error(self) -> str:
        return self._install_error

    def stage_context(self, session_id: str, payload: JsonObject, *, context_path: str | Path | None = None) -> Path:
        if context_path is not None:
            path = Path(context_path).expanduser().resolve()
        else:
            path = (self._context_dir / f"{session_id}.json").resolve()
        staged_payload = dict(payload)
        staged_payload["rsi_runtime_context_path"] = str(path)
        _atomic_write_json(path, staged_payload)
        return path

    def lifecycle_events(self, session_id: str, *, limit: int = 256) -> list[JsonObject]:
        path = self._lifecycle_dir / f"{session_id}.jsonl"
        if not path.exists():
            return []
        out: list[JsonObject] = []
        for line in path.read_text(encoding="utf-8").splitlines():
            if not line.strip():
                continue
            parsed = json.loads(line)
            if not isinstance(parsed, dict):
                raise RuntimeError(f"Invalid RSI lifecycle event payload for session {session_id}.")
            out.append(parsed)
        if limit <= 0:
            return out
        return out[-limit:]

    def _install(self) -> None:
        try:
            self._plugin_dir.mkdir(parents=True, exist_ok=True)
            self._context_dir.mkdir(parents=True, exist_ok=True)
            self._lifecycle_dir.mkdir(parents=True, exist_ok=True)
            (self._plugin_dir / "plugin.yaml").write_text(_PLUGIN_MANIFEST, encoding="utf-8")
            (self._plugin_dir / "__init__.py").write_text(_build_plugin_module(), encoding="utf-8")
            if callable(discover_plugins):
                discover_plugins(force=True)
            self._status = "ready"
        except Exception as exc:  # pragma: no cover - filesystem/env dependent
            self._install_error = str(exc)
            self._status = "degraded"


class HermesAdapter:
    def __init__(self, config: RunnerConfig) -> None:
        self._config = config
        self._context_engine = HermesContextEngine(config.hermes_home)

    @property
    def metadata(self) -> HermesAdapterMetadata:
        return HermesAdapterMetadata(
            version=self._detect_version(),
            pin=(self._config.hermes_pin or "").strip(),
            context_engine_mode="pre_llm_call_plugin",
            context_engine_status=self._context_engine.status,
            lifecycle_hook_status=self._context_engine.status,
        )

    def stage_task_context(self, session_id: str, payload: JsonObject, *, context_path: str | Path | None = None) -> Path:
        return self._context_engine.stage_context(session_id, payload, context_path=context_path)

    def lifecycle_events(self, session_id: str, *, limit: int = 256) -> list[JsonObject]:
        return self._context_engine.lifecycle_events(session_id, limit=limit)

    def _detect_version(self) -> str:
        try:
            return importlib.metadata.version("hermes-agent")
        except importlib.metadata.PackageNotFoundError:
            return ""
        except Exception:
            return ""
