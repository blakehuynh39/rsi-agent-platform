from __future__ import annotations

from dataclasses import dataclass, field
import hashlib
import http.client
import json
import logging
import os
import queue
import threading
import time
from typing import Any, Callable
from urllib.parse import urlparse

from .config import RunnerConfig
from .json_types import JsonObject

logger = logging.getLogger(__name__)

_SINK_BATCH_MAX_ITEMS = 64
_SINK_FLUSH_INTERVAL_SECONDS = 0.25
_SINK_IDLE_EXIT_SECONDS = 30.0
_SINK_CONNECT_TIMEOUT_SECONDS = 10
_SINK_DIAGNOSTIC_FLUSH_SECONDS = 2.0


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
    _sink_queue: "queue.Queue[JsonObject]" = field(default_factory=queue.Queue)
    _sink_worker: threading.Thread | None = None
    _sink_worker_lock: threading.Lock = field(default_factory=threading.Lock)
    on_emit: Callable[[JsonObject], None] | None = None

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
        sink_status = "configured" if config.runtime_observation_sink_url else "local_only"
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
            observation_payload = dict(payload or {})
            if self.config.executor_instance_id:
                observation_payload.setdefault("executor_instance_id", self.config.executor_instance_id)
            pod_uid = _string(os.getenv("POD_UID"))
            if pod_uid:
                observation_payload.setdefault("pod_uid", pod_uid)
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
                "payload": observation_payload,
                "recorded_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                "invocation_id": self.invocation_id,
            }
            self._events.append(dict(item))
        logger.info("runner observation %s", json.dumps(item, ensure_ascii=True, sort_keys=True))
        if self.on_emit is not None:
            try:
                self.on_emit(dict(item))
            except Exception:
                logger.exception("runner observation callback failed execution_id=%s", self.execution_id)
        if not self.config.runtime_observation_sink_url:
            return
        self._ensure_sink_worker()
        self._sink_queue.put(dict(item))

    def _ensure_sink_worker(self) -> None:
        if not self.config.runtime_observation_sink_url:
            return
        with self._sink_worker_lock:
            if self._sink_worker is not None and self._sink_worker.is_alive():
                return
            self._sink_worker = threading.Thread(
                target=self._sink_loop,
                name=f"rsi-observation-sink-{self.execution_id}",
                daemon=True,
            )
            self._sink_worker.start()

    def _sink_loop(self) -> None:
        while True:
            try:
                first = self._sink_queue.get(timeout=_SINK_IDLE_EXIT_SECONDS)
            except queue.Empty:
                return
            batch = [first]
            flush_deadline = time.monotonic() + (
                0.0 if self._observation_requires_fast_flush(first) else _SINK_FLUSH_INTERVAL_SECONDS
            )
            while len(batch) < _SINK_BATCH_MAX_ITEMS:
                timeout = max(0.0, flush_deadline - time.monotonic())
                if timeout <= 0:
                    break
                try:
                    item = self._sink_queue.get(timeout=timeout)
                except queue.Empty:
                    break
                batch.append(item)
                if self._observation_requires_fast_flush(item):
                    flush_deadline = time.monotonic()
            try:
                self._post_observation_batch(batch)
                with self._lock:
                    self.sink_status = "ok"
            except (TimeoutError, OSError, ValueError, http.client.HTTPException) as exc:
                with self._lock:
                    self.sink_status = "degraded"
                    error_text = _string(exc)
                    if error_text:
                        self.sink_errors.append(error_text)
            finally:
                for _ in batch:
                    self._sink_queue.task_done()

    @staticmethod
    def _observation_requires_fast_flush(item: JsonObject) -> bool:
        event_type = _string(item.get("event_type"))
        return not event_type.startswith("model.reasoning.delta") and not event_type.startswith("model.output.delta")

    def _post_observation_batch(self, observations: list[JsonObject]) -> None:
        if not observations:
            return
        body = json.dumps({"observations": observations}, ensure_ascii=True, sort_keys=True).encode("utf-8")
        self._post_observation(body, batch=True)

    def _post_observation(self, body: bytes, *, batch: bool = False) -> None:
        parsed = urlparse(self.config.runtime_observation_sink_url or "")
        if parsed.scheme not in {"http", "https"} or not parsed.netloc:
            raise ValueError("runtime observation sink URL must be an absolute http(s) URL")
        connection_cls = http.client.HTTPSConnection if parsed.scheme == "https" else http.client.HTTPConnection
        path = parsed.path or "/"
        if batch and not path.rstrip("/").endswith("/batch"):
            path = f"{path.rstrip('/')}/batch"
        if parsed.query:
            path = f"{path}?{parsed.query}"
        connection = connection_cls(parsed.netloc, timeout=_SINK_CONNECT_TIMEOUT_SECONDS)
        try:
            connection.request("POST", path, body=body, headers={"Content-Type": "application/json"})
            response = connection.getresponse()
            response.read()
            if response.status >= 400:
                raise OSError(f"observation sink returned HTTP {response.status}")
        finally:
            connection.close()

    def flush(self, timeout_seconds: float = _SINK_DIAGNOSTIC_FLUSH_SECONDS) -> bool:
        if not self.config.runtime_observation_sink_url:
            return True
        self._ensure_sink_worker()
        deadline = time.monotonic() + max(0.0, timeout_seconds)
        while time.monotonic() < deadline:
            if self._sink_queue.unfinished_tasks == 0:
                return True
            time.sleep(0.01)
        with self._lock:
            self.sink_status = "degraded"
            self.sink_errors.append("observation sink flush timed out")
        return False

    def diagnostics(self) -> JsonObject:
        self.flush()
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
