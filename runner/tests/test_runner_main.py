from __future__ import annotations

from http.server import ThreadingHTTPServer
import json
import threading
import types
import unittest
from urllib import request

from rsi_runner import main as runner_main


class _Runtime:
    available = True

    def __init__(self) -> None:
        self.drain_requested = 0
        self.probe_calls: list[str] = []
        self.snapshot_include_review_queue: list[bool] = []

    @property
    def metadata(self) -> dict[str, object]:
        return {"executor_instance_id": "test-executor"}

    def probe_metadata(self) -> dict[str, object]:
        self.probe_calls.append("probe")
        return {
            "executor_instance_id": "test-executor",
            "status": "ok",
            "available": self.available,
        }

    def active_execution_snapshot(self, *, include_self_review_queue: bool = True) -> dict[str, object]:
        self.snapshot_include_review_queue.append(include_self_review_queue)
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


class _ProbeRuntime(_Runtime):
    @property
    def metadata(self) -> dict[str, object]:
        raise AssertionError("probe endpoints must not read rich runtime metadata")


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

    def test_healthz_uses_probe_metadata_without_rich_runtime_metadata(self) -> None:
        runtime = _ProbeRuntime()
        server, base_url = self._serve(runtime)
        try:
            with request.urlopen(base_url + "/healthz", timeout=2) as response:
                payload = json.loads(response.read().decode("utf-8"))
        finally:
            server.shutdown()
            server.server_close()

        self.assertEqual(200, response.status)
        self.assertEqual("test-executor", payload["executor_instance_id"])
        self.assertEqual(["probe"], runtime.probe_calls)
        self.assertEqual([], runtime.snapshot_include_review_queue)

    def test_readyz_uses_probe_metadata_and_fast_active_snapshot(self) -> None:
        runtime = _ProbeRuntime()
        server, base_url = self._serve(runtime)
        try:
            with request.urlopen(base_url + "/readyz", timeout=2) as response:
                payload = json.loads(response.read().decode("utf-8"))
        finally:
            server.shutdown()
            server.server_close()

        self.assertEqual(200, response.status)
        self.assertEqual("active", payload["drain_status"])
        self.assertEqual(0, payload["active_execution_count"])
        self.assertEqual(["probe"], runtime.probe_calls)
        self.assertEqual([False], runtime.snapshot_include_review_queue)


if __name__ == "__main__":
    unittest.main()
