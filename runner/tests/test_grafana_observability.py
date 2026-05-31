from __future__ import annotations

from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import threading
import unittest
from unittest import mock
from urllib import parse

from rsi_runner import grafana_observability


class GrafanaObservabilityHandler(BaseHTTPRequestHandler):
    requests: list[dict[str, str]] = []

    def do_GET(self) -> None:  # noqa: N802
        GrafanaObservabilityHandler.requests.append(
            {
                "path": self.path,
                "authorization": self.headers.get("Authorization", ""),
                "user_agent": self.headers.get("User-Agent", ""),
                "cf_access_client_id": self.headers.get("CF-Access-Client-Id", ""),
                "cf_access_client_secret": self.headers.get("CF-Access-Client-Secret", ""),
            }
        )
        if self.path.startswith("/api/datasources?") or self.path == "/api/datasources":
            self._json(
                [
                    {"uid": "loki", "name": "Loki", "type": "loki"},
                    {"uid": "thanos", "name": "Thanos", "type": "prometheus"},
                ]
            )
            return
        if self.path.startswith("/api/search?"):
            self._json([{"uid": "dash1", "title": "Depin Overview", "type": "dash-db"}])
            return
        if self.path == "/api/dashboards/uid/dash1":
            self._json({"dashboard": {"uid": "dash1", "title": "Depin Overview"}})
            return
        if self.path == "/api/v1/provisioning/alert-rules":
            self._json(
                [
                    {"uid": "rule1", "title": "Pod restarts", "folderUID": "infra", "ruleGroup": "kubernetes"},
                    {"uid": "rule2", "title": "Indexer lag", "folderUID": "story", "ruleGroup": "chain"},
                ]
            )
            return
        if self.path == "/api/v1/provisioning/alert-rules/rule1":
            self._json({"uid": "rule1", "title": "Pod restarts", "folderUID": "infra"})
            return
        if self.path.startswith("/api/alertmanager/grafana/api/v2/alerts?"):
            self._json([{"labels": {"alertname": "PodRestarts", "namespace": "rsi-platform"}}])
            return
        if "/loki/api/v1/query_range" in self.path:
            self._json(
                {
                    "status": "success",
                    "data": {
                        "result": [
                            {
                                "stream": {"namespace": "rsi-platform"},
                                "values": [["1700000000000000000", "first log line"]],
                            }
                        ]
                    },
                }
            )
            return
        if "/api/v1/query" in self.path:
            self._json({"status": "success", "data": {"resultType": "vector", "result": []}})
            return
        self.send_response(404)
        self.end_headers()

    def _json(self, payload: object) -> None:
        raw = json.dumps(payload).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def log_message(self, fmt: str, *args: object) -> None:
        return


class GrafanaObservabilityTest(unittest.TestCase):
    def setUp(self) -> None:
        GrafanaObservabilityHandler.requests = []
        self.server = ThreadingHTTPServer(("127.0.0.1", 0), GrafanaObservabilityHandler)
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()
        self.base_url = f"http://127.0.0.1:{self.server.server_port}"

    def tearDown(self) -> None:
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=5)

    def grafana_env(self) -> mock._patch_dict:
        return mock.patch.dict(
            "os.environ",
            {
                "GRAFANA_SERVER": self.base_url,
                "GRAFANA_TOKEN": "secret-token",
                "RSI_GRAFANA_METRICS_DATASOURCE_UID": "thanos",
                "RSI_GRAFANA_LOGS_DATASOURCE_UID": "loki",
            },
            clear=True,
        )

    def test_datasources_filters_by_type_without_printing_token(self) -> None:
        with self.grafana_env():
            payload = grafana_observability.datasources_query("loki")

        self.assertEqual([{"name": "Loki", "type": "loki", "uid": "loki"}], payload["datasources"])
        self.assertEqual("Bearer secret-token", GrafanaObservabilityHandler.requests[0]["authorization"])
        self.assertEqual("rsi-agent-platform-observability/1.0", GrafanaObservabilityHandler.requests[0]["user_agent"])
        self.assertNotIn("secret-token", json.dumps(payload))

    def test_datasources_sends_cloudflare_access_headers_when_configured(self) -> None:
        with self.grafana_env(), mock.patch.dict(
            "os.environ",
            {
                "RSI_GRAFANA_CF_ACCESS_CLIENT_ID": "cf-client",
                "RSI_GRAFANA_CF_ACCESS_CLIENT_SECRET": "cf-secret",
                "RSI_GRAFANA_USER_AGENT": "custom-rsi-agent",
            },
        ):
            grafana_observability.datasources_query()

        self.assertEqual("custom-rsi-agent", GrafanaObservabilityHandler.requests[0]["user_agent"])
        self.assertEqual("cf-client", GrafanaObservabilityHandler.requests[0]["cf_access_client_id"])
        self.assertEqual("cf-secret", GrafanaObservabilityHandler.requests[0]["cf_access_client_secret"])

    def test_metrics_query_uses_prometheus_datasource_proxy(self) -> None:
        with self.grafana_env():
            grafana_observability.metrics_query("vector(1)")

        parsed = parse.urlparse(GrafanaObservabilityHandler.requests[0]["path"])
        self.assertEqual("/api/datasources/proxy/uid/thanos/api/v1/query", parsed.path)
        self.assertEqual({"query": ["vector(1)"]}, parse.parse_qs(parsed.query))

    def test_logs_query_uses_loki_datasource_proxy(self) -> None:
        with self.grafana_env(), mock.patch("rsi_runner.grafana_observability._now_seconds", return_value=1_700_000_000):
            grafana_observability.logs_query('{namespace="rsi-platform"}', limit=5, since="30m")

        parsed = parse.urlparse(GrafanaObservabilityHandler.requests[0]["path"])
        self.assertEqual("/api/datasources/proxy/uid/loki/loki/api/v1/query_range", parsed.path)
        params = parse.parse_qs(parsed.query)
        self.assertEqual(['{namespace="rsi-platform"}'], params["query"])
        self.assertEqual(["5"], params["limit"])
        self.assertEqual(["1699998200000000000"], params["start"])
        self.assertEqual(["1700000000000000000"], params["end"])

    def test_dashboard_and_alert_reads_use_grafana_read_apis(self) -> None:
        with self.grafana_env():
            dashboards = grafana_observability.dashboards_search("depin", tags=["prod"], limit=10)
            dashboard = grafana_observability.dashboard_get("dash1")
            rules = grafana_observability.alert_rules_search("pod", folder_uid="infra", limit=10)
            rule = grafana_observability.alert_rule_get("rule1")
            alerts = grafana_observability.active_alerts(["namespace=rsi-platform"], limit=5)

        self.assertEqual("dash1", dashboards["dashboards"][0]["uid"])
        self.assertEqual("Depin Overview", dashboard["dashboard"]["title"])
        self.assertEqual([{"folderUID": "infra", "ruleGroup": "kubernetes", "title": "Pod restarts", "uid": "rule1"}], rules["alert_rules"])
        self.assertEqual("Pod restarts", rule["title"])
        self.assertEqual("PodRestarts", alerts["alerts"][0]["labels"]["alertname"])
        self.assertEqual("/api/search", parse.urlparse(GrafanaObservabilityHandler.requests[0]["path"]).path)
        self.assertEqual("/api/dashboards/uid/dash1", GrafanaObservabilityHandler.requests[1]["path"])
        self.assertEqual("/api/v1/provisioning/alert-rules", GrafanaObservabilityHandler.requests[2]["path"])
        self.assertEqual("/api/v1/provisioning/alert-rules/rule1", GrafanaObservabilityHandler.requests[3]["path"])
        self.assertEqual("/api/alertmanager/grafana/api/v2/alerts", parse.urlparse(GrafanaObservabilityHandler.requests[4]["path"]).path)

    def test_logs_query_accepts_time_aliases_and_extracts_lines(self) -> None:
        with self.grafana_env(), mock.patch("rsi_runner.grafana_observability._now_seconds", return_value=1_700_000_000):
            payload = grafana_observability.logs_query(
                '{namespace="rsi-platform"}',
                limit=5,
                start="now-30m",
                end="now",
            )

        self.assertEqual(["first log line"], grafana_observability.loki_log_lines(payload))
        parsed = parse.urlparse(GrafanaObservabilityHandler.requests[0]["path"])
        params = parse.parse_qs(parsed.query)
        self.assertEqual(["1699998200000000000"], params["start"])
        self.assertEqual(["1700000000000000000"], params["end"])


if __name__ == "__main__":
    unittest.main()
