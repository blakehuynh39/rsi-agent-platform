from __future__ import annotations

import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any, Dict

from .config import RunnerConfig, RunnerConfigError
from .hermes_runtime import HermesRuntime, RunnerTaskRequest


class RunnerHandler(BaseHTTPRequestHandler):
    runtime: HermesRuntime
    config: RunnerConfig

    def _json(self, status: int, payload: Dict[str, Any]) -> None:
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
        payload = json.loads(self.rfile.read(content_length) or "{}")
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


def run_server() -> None:
    config = RunnerConfig.from_env()
    runtime = HermesRuntime(config)
    RunnerHandler.config = config
    RunnerHandler.runtime = runtime

    server = ThreadingHTTPServer((config.host, config.port), RunnerHandler)
    print(f"rsi-runner listening on {config.host}:{config.port} role={config.role}")
    server.serve_forever()


def run_once() -> None:
    config = RunnerConfig.from_env()
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
