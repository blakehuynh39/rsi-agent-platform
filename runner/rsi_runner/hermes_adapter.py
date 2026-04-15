from __future__ import annotations

from dataclasses import dataclass
import importlib.metadata
import json
from pathlib import Path

from .json_types import JsonObject

from .config import RunnerConfig

try:
    from hermes_cli.plugins import discover_plugins  # type: ignore
except (ImportError, ModuleNotFoundError):  # pragma: no cover - depends on Hermes install
    discover_plugins = None


_PLUGIN_MANIFEST = """name: rsi_context_engine
version: "1.0.0"
description: "RSI governed context injection and lifecycle capture"
"""


_PLUGIN_MODULE = """from __future__ import annotations

import json
import os
import time
from pathlib import Path


def _runtime_root() -> Path:
    hermes_home = Path(os.getenv("HERMES_HOME", "~/.hermes")).expanduser()
    return hermes_home / "rsi_runtime"


def _context_path(session_id: str) -> Path:
    return _runtime_root() / "context" / f"{session_id}.json"


def _lifecycle_path(session_id: str) -> Path:
    return _runtime_root() / "lifecycle" / f"{session_id}.jsonl"


def _append_event(session_id: str, event: str, payload: dict) -> None:
    path = _lifecycle_path(session_id)
    path.parent.mkdir(parents=True, exist_ok=True)
    item = {
        "event": event,
        "recorded_at_unix": time.time(),
        **(payload or {}),
    }
    with path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(item, sort_keys=True) + "\\n")

def _load_context(session_id: str) -> dict:
    path = _context_path(session_id)
    if not path.exists():
        return {}
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"Invalid RSI context payload for session {session_id}.") from exc


def _json_block(label: str, value) -> str:
    if not value:
        return ""
    return f"{label}:\\n" + json.dumps(value, ensure_ascii=True, sort_keys=True)


def _render_context(payload: dict) -> str:
    parts: list[str] = []
    summary = str(payload.get("context_summary", "") or "").strip()
    if summary:
        parts.append("Context summary:\\n" + summary)
    refs = payload.get("context_refs") or []
    if refs:
        parts.append(_json_block("Context refs", refs))
    tool_allowlist = payload.get("tool_allowlist_effective") or []
    if tool_allowlist:
        parts.append("Governed tool allowlist: " + ", ".join(str(item) for item in tool_allowlist))
    blocked = payload.get("blocked_tool_names") or []
    if blocked:
        parts.append("Blocked tools by RSI policy: " + ", ".join(str(item) for item in blocked))
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
    return {"context": "RSI governed context engine\\n\\n" + rendered}


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
    ctx.register_hook("pre_llm_call", pre_llm_call)
    ctx.register_hook("on_session_start", on_session_start)
    ctx.register_hook("on_session_end", on_session_end)
"""


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

    def stage_context(self, session_id: str, payload: JsonObject) -> None:
        self._context_dir.mkdir(parents=True, exist_ok=True)
        path = self._context_dir / f"{session_id}.json"
        tmp_path = path.with_suffix(".json.tmp")
        tmp_path.write_text(json.dumps(payload, indent=2, sort_keys=True), encoding="utf-8")
        tmp_path.replace(path)

    def lifecycle_events(self, session_id: str) -> list[JsonObject]:
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
        return out[-8:]

    def _install(self) -> None:
        try:
            self._plugin_dir.mkdir(parents=True, exist_ok=True)
            self._context_dir.mkdir(parents=True, exist_ok=True)
            self._lifecycle_dir.mkdir(parents=True, exist_ok=True)
            (self._plugin_dir / "plugin.yaml").write_text(_PLUGIN_MANIFEST, encoding="utf-8")
            (self._plugin_dir / "__init__.py").write_text(_PLUGIN_MODULE, encoding="utf-8")
            if callable(discover_plugins):
                discover_plugins()
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

    def stage_task_context(self, session_id: str, payload: JsonObject) -> None:
        self._context_engine.stage_context(session_id, payload)

    def lifecycle_events(self, session_id: str) -> list[JsonObject]:
        return self._context_engine.lifecycle_events(session_id)

    def _detect_version(self) -> str:
        try:
            return importlib.metadata.version("hermes-agent")
        except importlib.metadata.PackageNotFoundError:
            return ""
        except Exception:
            return ""
