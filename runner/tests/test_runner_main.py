from __future__ import annotations

from http.server import ThreadingHTTPServer
import json
import threading
import types
import unittest
from urllib import request

from rsi_runner import main as runner_main


class _Runtime:
    metadata = {"executor_instance_id": "test-executor"}
    available = True

    def __init__(self) -> None:
        self.drain_requested = 0

    def active_execution_snapshot(self) -> dict[str, object]:
        return {
            "executor_instance_id": "test-executor",
            "active_execution_count": 0,
            "active_execution_ids": [],
        }

    def request_drain(self) -> None:
        self.drain_requested += 1

    def wait_for_active_executions(self, _timeout_seconds: int) -> dict[str, object]:
        return {"active_execution_count": 0, "active_execution_ids": []}

    def terminate_self_review_processes(self, *, timeout_seconds: float) -> None:
        raise AssertionError(f"unexpected termination path: {timeout_seconds}")


class RunnerDrainHTTPTest(unittest.TestCase):
    def setUp(self) -> None:
        runner_main._DRAINING.clear()
        runner_main._DRAIN_STARTED_AT_UNIX = 0.0
        runner_main._DRAIN_DEADLINE_UNIX = 0.0

    def tearDown(self) -> None:
        runner_main._DRAINING.clear()
        runner_main._DRAIN_STARTED_AT_UNIX = 0.0
        runner_main._DRAIN_DEADLINE_UNIX = 0.0

    def _serve(self, runtime: _Runtime):
        runner_main.RunnerHandler.runtime = runtime
        runner_main.RunnerHandler.config = types.SimpleNamespace(
            drain_timeout_seconds=5,
            hermes_executor_service_only=False,
        )
        server = ThreadingHTTPServer(("127.0.0.1", 0), runner_main.RunnerHandler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        return server, f"http://127.0.0.1:{server.server_port}"

    def test_kubernetes_lifecycle_get_can_start_drain(self) -> None:
        runtime = _Runtime()
        server, base_url = self._serve(runtime)
        try:
            with request.urlopen(base_url + "/internal/drain/start", timeout=2) as response:
                payload = json.loads(response.read().decode("utf-8"))
        finally:
            server.shutdown()
            server.server_close()

        self.assertEqual(202, response.status)
        self.assertEqual("draining", payload["drain_status"])
        self.assertEqual(1, runtime.drain_requested)

    def test_kubernetes_lifecycle_get_can_wait_for_prestop_drain(self) -> None:
        runtime = _Runtime()
        server, base_url = self._serve(runtime)
        try:
            with request.urlopen(base_url + "/internal/drain/prestop", timeout=2) as response:
                payload = json.loads(response.read().decode("utf-8"))
        finally:
            server.shutdown()
            server.server_close()

        self.assertEqual(200, response.status)
        self.assertEqual("drained", payload["drain_status"])
        self.assertEqual(1, runtime.drain_requested)


if __name__ == "__main__":
    unittest.main()
