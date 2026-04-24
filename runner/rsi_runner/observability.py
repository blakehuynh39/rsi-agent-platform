from __future__ import annotations

from dataclasses import dataclass, field
import hashlib
import http.client
import json
import logging
import threading
import time
from typing import Any
from urllib.parse import urlparse

from .config import RunnerConfig
from .json_types import JsonObject

logger = logging.getLogger(__name__)


def _string(value: Any) -> str:
    if value is None:
        return ""
    return str(value).strip()


def execution_observation_id(operation_id: str, trace_id: str, workflow_id: str, session_id: str, invocation_id: str = "") -> str:
    primary = _string(operation_id)
    seed = primary or "|".join(
        item
        for item in [
            _string(trace_id),
            _string(workflow_id),
            _string(session_id),
            _string(invocation_id),
        ]
        if item
    ) or f"runner:{time.time_ns()}"
    digest = hashlib.sha1(seed.encode("utf-8")).hexdigest()[:16]
    return f"hexec-{digest}"


@dataclass
class ObservationEmitter:
    config: RunnerConfig
    trace_id: str
    workflow_id: str
    operation_id: str
    role: str
    hermes_session_id: str
    execution_id: str
    seq: int = 0
    sink_status: str = "not_configured"
    sink_errors: list[str] = field(default_factory=list)
    invocation_id: str = ""
    _events: list[JsonObject] = field(default_factory=list)
    _lock: threading.Lock = field(default_factory=threading.Lock)

    @classmethod
    def create(
        cls,
        config: RunnerConfig,
        *,
        trace_id: str,
        workflow_id: str,
        operation_id: str,
        role: str,
        hermes_session_id: str,
        execution_id: str = "",
    ) -> "ObservationEmitter":
        invocation_id = f"invoke-{time.time_ns()}"
        resolved_execution_id = _string(execution_id) or execution_observation_id(operation_id, trace_id, workflow_id, hermes_session_id, invocation_id)
        sink_status = "configured" if config.tool_gateway_base_url else "not_configured"
        return cls(
            config=config,
            trace_id=_string(trace_id),
            workflow_id=_string(workflow_id),
            operation_id=_string(operation_id),
            role=_string(role),
            hermes_session_id=_string(hermes_session_id),
            execution_id=resolved_execution_id,
            sink_status=sink_status,
            invocation_id=invocation_id,
        )

    def emit(self, *, phase: str, event_type: str, status: str = "", payload: JsonObject | None = None) -> None:
        with self._lock:
            self.seq += 1
            item: JsonObject = {
                "execution_id": self.execution_id,
                "operation_id": self.operation_id,
                "trace_id": self.trace_id,
                "workflow_id": self.workflow_id,
                "hermes_session_id": self.hermes_session_id,
                "role": self.role,
                "phase": _string(phase),
                "event_type": _string(event_type),
                "status": _string(status),
                "seq": self.seq,
                "payload": payload or {},
                "recorded_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                "invocation_id": self.invocation_id,
            }
            self._events.append(dict(item))
        logger.info("runner observation %s", json.dumps(item, ensure_ascii=True, sort_keys=True))
        if not self.config.tool_gateway_base_url:
            return
        body = json.dumps(item, ensure_ascii=True, sort_keys=True).encode("utf-8")
        try:
            self._post_observation(body)
            with self._lock:
                self.sink_status = "ok"
        except (TimeoutError, OSError, ValueError) as exc:
            with self._lock:
                self.sink_status = "degraded"
                error_text = _string(exc)
                if error_text:
                    self.sink_errors.append(error_text)

    def _post_observation(self, body: bytes) -> None:
        parsed = urlparse(f"{self.config.tool_gateway_base_url.rstrip('/')}/api/runtime/observations")
        if parsed.scheme not in {"http", "https"} or not parsed.netloc:
            raise ValueError("tool gateway base URL must be an absolute http(s) URL")
        connection_cls = http.client.HTTPSConnection if parsed.scheme == "https" else http.client.HTTPConnection
        path = parsed.path or "/"
        if parsed.query:
            path = f"{path}?{parsed.query}"
        connection = connection_cls(parsed.netloc, timeout=5)
        try:
            connection.request("POST", path, body=body, headers={"Content-Type": "application/json"})
            response = connection.getresponse()
            response.read()
            if response.status >= 400:
                raise OSError(f"observation sink returned HTTP {response.status}")
        finally:
            connection.close()

    def diagnostics(self) -> JsonObject:
        out: JsonObject = {
            "observation_execution_id": self.execution_id,
            "observation_sink_status": self.sink_status,
            "observation_seq": self.seq,
        }
        if self.sink_errors:
            out["observation_sink_errors"] = list(self.sink_errors[-5:])
        return out

    def events(self) -> list[JsonObject]:
        with self._lock:
            return [dict(item) for item in self._events]
