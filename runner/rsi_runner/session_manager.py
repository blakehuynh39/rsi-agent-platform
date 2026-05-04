from __future__ import annotations

from dataclasses import dataclass, field
import hashlib
import json
import os
from pathlib import Path
import random
import time
from typing import Any, Protocol

from .db_utils import float_env, sqlite_error_is_locked
from .json_types import JsonObject

try:
    from hermes_state import SessionDB  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    SessionDB = None

try:
    from tools.skills_sync import sync_skills  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    sync_skills = None

from .config import RunnerConfig


def _render_provider_routing(provider_routing: dict[str, object]) -> str:
    lines = ["provider_routing:"]
    for key in ("only", "ignore", "order"):
        value = provider_routing.get(key)
        if isinstance(value, list) and value:
            lines.append(f"  {key}:")
            for item in value:
                lines.append(f"    - {json.dumps(str(item))}")
    for key in ("sort", "data_collection"):
        value = provider_routing.get(key)
        if isinstance(value, str) and value:
            lines.append(f"  {key}: {json.dumps(value)}")
    value = provider_routing.get("require_parameters")
    if isinstance(value, bool):
        lines.append(f"  require_parameters: {str(value).lower()}")
    return "\n".join(lines) + "\n"


@dataclass
class SessionContext:
    session_id: str
    parent_session_id: str
    scope_kind: str
    scope_id: str
    parent_scope_kind: str
    parent_scope_id: str
    memory_backend: str
    assistant_peer_id: str
    user_peer_id: str
    hermes_home: str
    session_db_path: str
    session_db_tracking_enabled: bool = True
    conversation_history: list[JsonObject] = field(default_factory=list)


@dataclass(frozen=True)
class SessionDBRef:
    db_path: Path
    defer_integrity_check: bool = True


@dataclass
class MemoryTracker:
    reads: list[JsonObject] = field(default_factory=list)
    writes: list[JsonObject] = field(default_factory=list)
    warnings: list[JsonObject] = field(default_factory=list)

    def record_read(self, kind: str, summary: str, source: str = "", ref: str = "") -> None:
        text = (summary or "").strip()
        if not text:
            return
        self.reads.append({"kind": kind, "summary": text, "source": source, "ref": ref})

    def record_write(self, kind: str, summary: str, source: str = "", ref: str = "") -> None:
        text = (summary or "").strip()
        if not text:
            return
        self.writes.append({"kind": kind, "summary": text, "source": source, "ref": ref})

    def record_warning(self, kind: str, summary: str, source: str = "", ref: str = "") -> None:
        text = (summary or "").strip()
        if not text:
            return
        self.warnings.append({"kind": kind, "summary": text, "source": source, "ref": ref})


class RunnerTaskLike(Protocol):
    session_scope_kind: str | None
    session_scope_id: str | None
    parent_session_scope_kind: str | None
    parent_session_scope_id: str | None
    user_peer_id: str | None
    recent_conversation_entries: list[JsonObject]
    assistant_peer_id: str | None


class SessionManager:
    def __init__(self, config: RunnerConfig) -> None:
        self._config = config
        self._hermes_home = Path(config.hermes_home)
        self._session_db_path = self._hermes_home / "state.db"
        self._skills_dir = self._hermes_home / "skills"
        bundled_override = str(os.getenv("HERMES_BUNDLED_SKILLS", "") or "").strip()
        self._bundled_skills_path = Path(bundled_override).expanduser() if bundled_override else None
        self._bundled_skills_available = bool(self._bundled_skills_path and self._bundled_skills_path.exists())
        self._bundled_skills_sync_status = "not_configured"
        self._bundled_skills_sync_error = ""
        self._hermes_config_parity_status = "pending"
        self._hermes_config_parity_error = ""
        self._db = None
        self._ready_issues = self._configure_home()

    @property
    def ready_issues(self) -> list[str]:
        return list(self._ready_issues)

    @property
    def available(self) -> bool:
        return not self._ready_issues

    @property
    def session_db_path(self) -> str:
        return str(self._session_db_path)

    @property
    def hermes_home(self) -> str:
        return str(self._hermes_home)

    @property
    def skills_dir(self) -> str:
        return str(self._skills_dir)

    @property
    def bundled_skills_available(self) -> bool:
        return self._bundled_skills_available

    @property
    def bundled_skills_sync_status(self) -> str:
        return self._bundled_skills_sync_status

    @property
    def bundled_skills_sync_error(self) -> str:
        return self._bundled_skills_sync_error

    @property
    def hermes_config_parity_status(self) -> str:
        return self._hermes_config_parity_status

    @property
    def hermes_config_parity_error(self) -> str:
        return self._hermes_config_parity_error

    @property
    def skills_healthy(self) -> bool:
        return self._bundled_skills_sync_status not in {"failed"}

    @property
    def session_db(self) -> Any:
        return self._get_db()

    @property
    def session_db_ref(self) -> Any:
        if SessionDB is None:
            return None
        return SessionDBRef(db_path=self._session_db_path)

    @property
    def honcho_available(self) -> bool:
        return self._config.memory_backend == "honcho"

    def prepare(self, task: RunnerTaskLike, *, load_history: bool = True) -> SessionContext:
        scope_kind = first_non_empty(getattr(task, "session_scope_kind", None), "role")
        scope_id = first_non_empty(getattr(task, "session_scope_id", None), self._config.role)
        parent_scope_kind = first_non_empty(getattr(task, "parent_session_scope_kind", None))
        parent_scope_id = first_non_empty(getattr(task, "parent_session_scope_id", None))
        session_id = stable_session_id(self._config.role, scope_kind, scope_id)
        parent_session_id = stable_session_id(self._config.role, parent_scope_kind, parent_scope_id) if parent_scope_kind and parent_scope_id else ""
        history: list[JsonObject] = []
        db = self._get_db() if load_history else None
        if db is not None:
            history = list(db.get_messages_as_conversation(session_id) or [])
        tracker_user_peer = first_non_empty(
            getattr(task, "user_peer_id", None),
            infer_user_peer_id(getattr(task, "recent_conversation_entries", []) or [], scope_kind, scope_id, self._config.role),
        )
        return SessionContext(
            session_id=session_id,
            parent_session_id=parent_session_id,
            scope_kind=scope_kind,
            scope_id=scope_id,
            parent_scope_kind=parent_scope_kind,
            parent_scope_id=parent_scope_id,
            memory_backend=self._config.memory_backend,
            assistant_peer_id=first_non_empty(getattr(task, "assistant_peer_id", None), self._config.honcho_ai_peer),
            user_peer_id=tracker_user_peer,
            hermes_home=str(self._hermes_home),
            session_db_path=str(self._session_db_path),
            session_db_tracking_enabled=load_history,
            conversation_history=history,
        )

    def attach_tracking(self, agent: Any, task: RunnerTaskLike, context: SessionContext) -> MemoryTracker:
        tracker = MemoryTracker()
        for entry in summarize_history(context.conversation_history):
            tracker.record_read("session_history", entry, source="session_db", ref=context.session_id)
        memory_manager = getattr(agent, "_memory_manager", None)
        if memory_manager is None:
            return tracker

        original_prefetch = getattr(memory_manager, "prefetch_all", None)
        if callable(original_prefetch):
            def tracked_prefetch(query: str, *, session_id: str = "") -> str:
                try:
                    result = original_prefetch(query, session_id=session_id)
                except Exception as exc:
                    tracker.record_warning(
                        "memory_prefetch_failed",
                        truncate_text(str(exc), 320),
                        source=self._config.memory_backend,
                        ref=session_id or context.session_id,
                    )
                    return ""
                if result and str(result).strip():
                    tracker.record_read("memory_prefetch", truncate_text(str(result), 800), source=self._config.memory_backend, ref=session_id or context.session_id)
                return result
            memory_manager.prefetch_all = tracked_prefetch

        original_sync = getattr(memory_manager, "sync_all", None)
        if callable(original_sync):
            def tracked_sync(user_content: str, assistant_content: str, *, session_id: str = "") -> None:
                tracker.record_write("memory_sync_user", truncate_text(user_content, 320), source=self._config.memory_backend, ref=session_id or context.session_id)
                tracker.record_write("memory_sync_assistant", truncate_text(assistant_content, 320), source=self._config.memory_backend, ref=session_id or context.session_id)
                try:
                    return original_sync(user_content, assistant_content, session_id=session_id)
                except Exception as exc:
                    tracker.record_warning(
                        "memory_sync_failed",
                        truncate_text(str(exc), 320),
                        source=self._config.memory_backend,
                        ref=session_id or context.session_id,
                    )
                    return None
            memory_manager.sync_all = tracked_sync
        return tracker

    def finalize(self, context: SessionContext, tracker: MemoryTracker) -> JsonObject:
        history: list[JsonObject] = list(context.conversation_history)
        delta: list[JsonObject] = []
        db = self._get_db() if getattr(context, "session_db_tracking_enabled", True) else None
        if db is not None:
            history = list(db.get_messages_as_conversation(context.session_id) or [])
            if len(history) > len(context.conversation_history):
                tracker.record_write(
                    "session_append",
                    f"Persisted {len(history) - len(context.conversation_history)} new session messages.",
                    source="session_db",
                    ref=context.session_id,
                )
            if len(history) >= len(context.conversation_history):
                delta = history[len(context.conversation_history):]
            else:
                delta = history
        return {
            "hermes_session_id": context.session_id,
            "parent_session_id": context.parent_session_id,
            "session_scope_kind": context.scope_kind,
            "session_scope_id": context.scope_id,
            "parent_session_scope_kind": context.parent_scope_kind,
            "parent_session_scope_id": context.parent_scope_id,
            "memory_backend": context.memory_backend,
            "assistant_peer_id": context.assistant_peer_id,
            "user_peer_id": context.user_peer_id,
            "hermes_home": context.hermes_home,
            "session_db_path": context.session_db_path,
            "memory_reads": tracker.reads,
            "memory_writes": tracker.writes,
            "memory_warnings": tracker.warnings,
            "session_messages_delta": delta,
        }

    def _get_db(self) -> Any:
        if not self.available or SessionDB is None:
            return None
        if self._db is None:
            self._db = self._open_session_db_with_retry()
        return self._db

    def _open_session_db_with_retry(self) -> Any:
        deadline = time.monotonic() + float_env("RSI_HERMES_SESSION_DB_OPEN_RETRY_SECONDS", 20.0)
        while True:
            try:
                return SessionDB(db_path=self._session_db_path)
            except Exception as exc:
                if not sqlite_error_is_locked(exc) or time.monotonic() >= deadline:
                    raise
                time.sleep(random.uniform(0.05, 0.25))

    def _configure_home(self) -> list[str]:
        issues: list[str] = []
        try:
            self._hermes_home.mkdir(parents=True, exist_ok=True)
        except Exception as exc:  # pragma: no cover - OS-specific failure
            return [f"create HERMES_HOME failed: {exc}"]
        try:
            self._write_hermes_config()
        except Exception as exc:  # pragma: no cover - OS-specific failure
            self._hermes_config_parity_status = "failed"
            self._hermes_config_parity_error = str(exc)
            issues.append(f"configure Hermes persistence failed: {exc}")
        try:
            self._write_honcho_config()
        except Exception as exc:  # pragma: no cover - OS-specific failure
            issues.append(f"configure Honcho persistence failed: {exc}")
        self._sync_bundled_skills()
        if SessionDB is None:
            issues.append("Hermes SessionDB is unavailable in this environment.")
        return issues

    def _write_hermes_config(self) -> None:
        config_path = self._hermes_home / "config.yaml"
        configured_model = str(self._config.model or "").strip()
        provider_model = configured_model.split("/", 1)[1]
        provider_routing = dict(self._config.openrouter_provider_routing or {})

        content = "model:\n"
        content += f"  default: {json.dumps(provider_model)}\n"
        content += "  provider: openrouter\n"
        content += "  api_key: \"\"\n"
        if provider_routing:
            content += _render_provider_routing(provider_routing)
        content += (
            "memory:\n"
            "  provider: honcho\n"
        )
        if self._config.hermes_native_terminal_enabled:
            content += (
                "terminal:\n"
                f"  backend: {json.dumps(self._config.hermes_terminal_env)}\n"
                f"  cwd: {json.dumps(self._config.hermes_terminal_cwd)}\n"
                f"  timeout: {self._config.hermes_terminal_timeout_seconds}\n"
                f"  lifetime_seconds: {self._config.hermes_terminal_lifetime_seconds}\n"
            )
        enabled_plugins = ["rsi_context_engine", "company_knowledge"]
        if self._config.hermes_executor_enabled:
            enabled_plugins.append("rsi_platform_runtime")
        content += "plugins:\n  enabled:\n"
        for plugin_name in enabled_plugins:
            content += f"    - {plugin_name}\n"
        config_path.write_text(content, encoding="utf-8")
        self._hermes_config_parity_status = "configured"
        self._hermes_config_parity_error = ""

    def _write_honcho_config(self) -> None:
        honcho_path = self._hermes_home / "honcho.json"
        payload = {
            "workspace": self._config.honcho_workspace,
            "environment": self._config.honcho_environment_effective,
            "hosts": {
                "hermes": {
                    "enabled": True,
                    "workspace": self._config.honcho_workspace,
                    "aiPeer": self._config.honcho_ai_peer,
                    "peerName": f"rsi:{self._config.role}:user",
                    "recallMode": self._config.honcho_recall_mode,
                    "writeFrequency": self._config.honcho_write_frequency,
                    "sessionStrategy": self._config.honcho_session_strategy,
                }
            },
        }
        if self._config.honcho_base_url:
            payload["baseUrl"] = self._config.honcho_base_url
        honcho_path.write_text(json.dumps(payload, indent=2, sort_keys=True), encoding="utf-8")

    def _sync_bundled_skills(self) -> None:
        if not self._bundled_skills_available:
            self._bundled_skills_sync_status = "not_configured"
            return
        if sync_skills is None:
            self._bundled_skills_sync_status = "sync_unavailable"
            self._bundled_skills_sync_error = "Hermes skills_sync module is unavailable in this environment."
            return
        try:
            sync_skills(quiet=True)
        except Exception as exc:
            self._bundled_skills_sync_status = "failed"
            self._bundled_skills_sync_error = str(exc)
            return
        self._bundled_skills_sync_status = "synced"
        self._bundled_skills_sync_error = ""


def stable_session_id(role: str, scope_kind: str, scope_id: str) -> str:
    role = first_non_empty(role, "prod")
    scope_kind = first_non_empty(scope_kind, "role")
    scope_id = first_non_empty(scope_id, role)
    digest = hashlib.sha256(f"{role}:{scope_kind}:{scope_id}".encode("utf-8")).hexdigest()[:20]
    return f"rsi-{role}-{scope_kind}-{digest}"


def summarize_history(history: list[JsonObject]) -> list[str]:
    if not history:
        return []
    out: list[str] = []
    for item in history[-4:]:
        role = first_non_empty(str(item.get("role", "")).strip(), "message")
        content = truncate_text(str(item.get("content", "") or ""), 160)
        if not content:
            continue
        out.append(f"{role}: {content}")
    return out


def truncate_text(value: str, limit: int) -> str:
    text = (value or "").strip()
    if len(text) <= limit:
        return text
    return text[: max(0, limit - 3)].rstrip() + "..."


def infer_user_peer_id(entries: list[JsonObject], scope_kind: str, scope_id: str, role: str) -> str:
    for entry in reversed(entries):
        actor_id = str(entry.get("actor_id", "") or "").strip()
        actor_type = str(entry.get("actor_type", "") or "").strip()
        if actor_id and actor_type in {"user", "operator"}:
            return f"{actor_type}:{actor_id}"
    return f"session:{role}:{scope_kind}:{scope_id}"


def first_non_empty(*values: str | None) -> str:
    for value in values:
        if value and value.strip():
            return value.strip()
    return ""
