from __future__ import annotations

import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import logging
import os
import signal
import threading
import time

from .json_types import JsonObject

from .config import RunnerConfig, RunnerConfigError
from .hermes_runtime import HermesRuntime, RunnerTaskRequest

logger = logging.getLogger(__name__)
_DRAINING = threading.Event()
_DRAIN_LOCK = threading.Lock()
_DRAIN_STARTED_AT_UNIX = 0.0
_DRAIN_DEADLINE_UNIX = 0.0

_SENSITIVE_KEY_FRAGMENTS = (
    "authorization",
    "api_key",
    "apikey",
    "token",
    "secret",
    "private_key",
    "password",
)


def _is_sensitive_key(key: str) -> bool:
    lower = str(key or "").strip().lower()
    if not lower:
        return False
    return any(fragment in lower for fragment in _SENSITIVE_KEY_FRAGMENTS)


def _sanitize_for_log(value, key: str = ""):
    if _is_sensitive_key(key):
        return "[redacted]"
    if isinstance(value, dict):
        return {child_key: _sanitize_for_log(child_value, child_key) for child_key, child_value in value.items()}
    if isinstance(value, list):
        return [_sanitize_for_log(item) for item in value]
    return value


def _truncate_for_log(text: str, limit: int) -> str:
    value = str(text or "").strip()
    limit = max(1024, int(limit or 0))
    if len(value) <= limit:
        return value
    return value[: limit - 20] + "...[truncated]"


def _json_for_log(value, limit: int) -> str:
    return _truncate_for_log(json.dumps(_sanitize_for_log(value), ensure_ascii=True, sort_keys=True), limit)


def _configure_logging(config: RunnerConfig) -> None:
    level_name = os.getenv("RSI_RUNNER_LOG_LEVEL", "INFO").strip().upper()
    level = getattr(logging, level_name, logging.INFO)
    if config.verbose_trace_logging and level > logging.INFO:
        level = logging.INFO
    logging.basicConfig(level=level, format="%(asctime)s %(levelname)s %(name)s %(message)s")


def _mark_draining(config: RunnerConfig) -> JsonObject:
    global _DRAIN_STARTED_AT_UNIX, _DRAIN_DEADLINE_UNIX
    with _DRAIN_LOCK:
        now = time.time()
        if not _DRAINING.is_set():
            _DRAINING.set()
            _DRAIN_STARTED_AT_UNIX = now
            _DRAIN_DEADLINE_UNIX = now + float(max(1, config.drain_timeout_seconds))
        elif _DRAIN_STARTED_AT_UNIX <= 0:
            _DRAIN_STARTED_AT_UNIX = now
            _DRAIN_DEADLINE_UNIX = now + float(max(1, config.drain_timeout_seconds))
    return {
        "status": "draining",
        "drain_status": "draining",
        "started_at_unix": _DRAIN_STARTED_AT_UNIX,
        "deadline_unix": _DRAIN_DEADLINE_UNIX,
    }


def _drain_status_payload(
    runtime: HermesRuntime,
    config: RunnerConfig,
    *,
    include_self_review_queue: bool = True,
) -> JsonObject:
    status = "draining" if _DRAINING.is_set() else "active"
    payload = runtime.active_execution_snapshot(include_self_review_queue=include_self_review_queue)
    payload["status"] = status
    payload["drain_status"] = status
    if _DRAINING.is_set():
        payload["started_at_unix"] = _DRAIN_STARTED_AT_UNIX
        payload["deadline_unix"] = _DRAIN_DEADLINE_UNIX or (time.time() + float(max(1, config.drain_timeout_seconds)))
    return payload


def _wait_for_drain(runtime: HermesRuntime, config: RunnerConfig) -> JsonObject:
    start_payload = _mark_draining(config)
    runtime.request_drain()
    timeout_seconds = max(1, int(_DRAIN_DEADLINE_UNIX - time.time()))
    snapshot = runtime.active_execution_snapshot()
    logger.info(
        "rsi-runner draining executor_instance=%s timeout_seconds=%s active=%s",
        snapshot.get("executor_instance_id"),
        timeout_seconds,
        snapshot.get("active_execution_ids"),
    )
    result = runtime.wait_for_active_executions(timeout_seconds)
    if int(result.get("active_execution_count") or 0) != 0:
        runtime.terminate_self_review_processes(timeout_seconds=5.0)
    result.update(start_payload)
    if int(result.get("active_execution_count") or 0) == 0:
        result["status"] = "drained"
        result["drain_status"] = "drained"
    else:
        result["status"] = "timeout"
        result["drain_status"] = "timeout"
    return result


class RunnerHandler(BaseHTTPRequestHandler):
    runtime: HermesRuntime
    config: RunnerConfig

    def _json(self, status: int, payload: JsonObject) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        try:
            self.end_headers()
            self.wfile.write(body)
        except (BrokenPipeError, ConnectionResetError):
            logger.debug("client disconnected before runner response body was written path=%s status=%s", self.path, status)

    def _handle_drain_start(self) -> None:
        payload = _mark_draining(self.config)
        self.runtime.request_drain()
        payload.update(self.runtime.active_execution_snapshot())
        self._json(202, payload)

    def _handle_drain_prestop(self) -> None:
        payload = _wait_for_drain(self.runtime, self.config)
        status = 200 if int(payload.get("active_execution_count") or 0) == 0 else 503
        self._json(status, payload)

    def do_GET(self) -> None:  # noqa: N802
        if self.path == "/healthz":
            self._json(200, self.runtime.probe_metadata())
            return
        if self.path == "/readyz":
            payload = self.runtime.probe_metadata()
            payload.update(_drain_status_payload(self.runtime, self.config, include_self_review_queue=False))
            status = 200 if self.runtime.available and not _DRAINING.is_set() else 503
            self._json(status, payload)
            return
        if self.path == "/runtimez":
            self._json(200, self.runtime.metadata)
            return
        if self.path == "/internal/drain/start":
            self._handle_drain_start()
            return
        if self.path == "/internal/drain/prestop":
            self._handle_drain_prestop()
            return
        if self.path == "/internal/drain/status":
            self._json(200, _drain_status_payload(self.runtime, self.config))
            return
        if self.path.startswith("/internal/hermes-executions/"):
            execution_id = self.path.rsplit("/", 1)[-1]
            payload = self.runtime.executor_status(execution_id)
            if payload:
                self._json(200, payload)
            else:
                self._json(404, {"error": "not found"})
            return
        self._json(404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        if self.path.startswith("/internal/hermes-executions/") and self.path.endswith("/cancel"):
            execution_id = self.path.removeprefix("/internal/hermes-executions/").removesuffix("/cancel").strip("/")
            payload = self.runtime.cancel_execution(execution_id)
            status = 200 if execution_id else 400
            self._json(status, payload)
            return

        if self.path == "/internal/drain/start":
            self._handle_drain_start()
            return

        if self.path == "/internal/drain/prestop":
            self._handle_drain_prestop()
            return

        if self.config.hermes_executor_service_only and self.path == "/execute":
            self._json(404, {"error": "not found"})
            return

        if self.path not in {"/execute", "/internal/hermes-executions"}:
            self._json(404, {"error": "not found"})
            return
        if _DRAINING.is_set():
            self._json(503, {"error": "runner is draining", "drain_status": "draining"})
            return

        content_length = int(self.headers.get("Content-Length", "0"))
        parsed = json.loads(self.rfile.read(content_length) or "{}")
        if not isinstance(parsed, dict):
            self._json(400, {"error": "request body must be a JSON object"})
            return
        payload: JsonObject = parsed
        task_payload = payload.get("task", payload)
        if isinstance(task_payload, dict):
            logger.info(
                "runner execute request task_type=%s trace=%s workflow=%s channel=%s thread=%s",
                str(task_payload.get("task_type", "") or "").strip(),
                str(task_payload.get("trace_id", "") or "").strip(),
                str(task_payload.get("workflow_id", "") or "").strip(),
                str(task_payload.get("channel_id", "") or "").strip(),
                str(task_payload.get("thread_ts", "") or "").strip(),
            )
        if self.config.verbose_trace_logging:
            logger.info("runner execute request payload=%s", _json_for_log(payload, self.config.verbose_trace_log_limit))
        if "task" in payload or "task_type" in payload:
            task = RunnerTaskRequest.from_payload(payload)
            if self.path == "/internal/hermes-executions" and bool(payload.get("async")):
                accepted = self.runtime.start_executor_task(task)
                self._json(202, accepted)
                return
            result = self.runtime.execute_task(task)
        else:
            prompt = payload.get("prompt", "")
            system_message = payload.get("system_message")
            result = self.runtime.execute(prompt, system_message=system_message)
        self._json(
            200,
            {
                "ok": result.ok,
                "message": result.message,
                "provider": result.provider,
                "raw": result.raw,
            },
        )
        logger.info(
            "runner execute response ok=%s provider=%s termination_reason=%s completion_verdict=%s native_log=%s",
            result.ok,
            result.provider,
            str(result.raw.get("termination_reason", "") or "").strip(),
            str(result.raw.get("completion_verdict", "") or "").strip(),
            str(result.raw.get("native_execution_log_path", "") or "").strip(),
        )
        if self.config.verbose_trace_logging:
            logger.info(
                "runner execute response payload=%s",
                _json_for_log(
                    {
                        "ok": result.ok,
                        "message": result.message,
                        "provider": result.provider,
                        "raw": result.raw,
                    },
                    self.config.verbose_trace_log_limit,
                ),
            )


def run_server() -> None:
    config = RunnerConfig.from_env()
    _configure_logging(config)
    runtime = HermesRuntime(config)
    RunnerHandler.config = config
    RunnerHandler.runtime = runtime

    server = ThreadingHTTPServer((config.host, config.port), RunnerHandler)

    def _start_drain(_signum, _frame) -> None:
        def _shutdown_after_drain() -> None:
            payload = _wait_for_drain(runtime, config)
            logger.info("rsi-runner drain waiter finished status=%s active=%s", payload.get("drain_status"), payload.get("active_execution_ids"))
            server.shutdown()

        _mark_draining(config)
        runtime.request_drain()
        threading.Thread(target=_shutdown_after_drain, name="rsi-runner-shutdown", daemon=False).start()

    signal.signal(signal.SIGTERM, _start_drain)
    signal.signal(signal.SIGINT, _start_drain)
    logger.info("rsi-runner listening on %s:%s role=%s", config.host, config.port, config.role)
    try:
        server.serve_forever()
    finally:
        server.server_close()


def run_once() -> None:
    config = RunnerConfig.from_env()
    _configure_logging(config)
    runtime = HermesRuntime(config)
    result = runtime.execute("Summarize the current RSI runner bootstrap state in one sentence.")
    print(
        json.dumps(
            {
                "ok": result.ok,
                "message": result.message,
                "provider": result.provider,
                "runtime": runtime.metadata,
            }
        )
    )


def main() -> None:
    parser = argparse.ArgumentParser(description="RSI Hermes runner wrapper")
    parser.add_argument("--once", action="store_true", help="Run one health check execution and exit")
    args = parser.parse_args()
    try:
        if args.once:
            run_once()
            return
        run_server()
    except RunnerConfigError as exc:
        raise SystemExit(str(exc)) from exc


if __name__ == "__main__":
    main()
