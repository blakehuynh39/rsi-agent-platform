from __future__ import annotations

import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import logging
import os

from .json_types import JsonObject

from .config import RunnerConfig, RunnerConfigError
from .hermes_runtime import HermesRuntime, RunnerTaskRequest

logger = logging.getLogger(__name__)

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


class RunnerHandler(BaseHTTPRequestHandler):
    runtime: HermesRuntime
    config: RunnerConfig

    def _json(self, status: int, payload: JsonObject) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self) -> None:  # noqa: N802
        if self.path == "/healthz":
            self._json(200, self.runtime.metadata)
            return
        if self.path == "/readyz":
            status = 200 if self.runtime.available else 503
            self._json(status, self.runtime.metadata)
            return
        if self.path == "/runtimez":
            self._json(200, self.runtime.metadata)
            return
        self._json(404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        if self.path != "/execute":
            self._json(404, {"error": "not found"})
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
    logger.info("rsi-runner listening on %s:%s role=%s", config.host, config.port, config.role)
    server.serve_forever()


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
