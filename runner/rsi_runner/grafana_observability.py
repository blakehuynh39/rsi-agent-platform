from __future__ import annotations

from datetime import datetime
import json
import os
import re
import time
from urllib import error, parse, request

from .json_types import JsonObject, JsonValue

DEFAULT_METRICS_DATASOURCE_UID = "thanos"
DEFAULT_LOGS_DATASOURCE_UID = "loki"
DEFAULT_USER_AGENT = "rsi-agent-platform-observability/1.0"


class GrafanaObservabilityError(RuntimeError):
    pass


def _env(name: str, default: str = "") -> str:
    return os.getenv(name, default).strip()


def _grafana_server() -> str:
    server = (_env("GRAFANA_SERVER") or _env("RSI_GRAFANA_BASE_URL")).rstrip("/")
    if not server:
        raise GrafanaObservabilityError("GRAFANA_SERVER or RSI_GRAFANA_BASE_URL is required")
    return server


def _grafana_token() -> str:
    token = _env("GRAFANA_TOKEN") or _env("RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN")
    if not token:
        raise GrafanaObservabilityError("GRAFANA_TOKEN or RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN is required")
    return token


def _request_json(path: str, params: JsonObject | None = None) -> JsonValue:
    server = _grafana_server()
    query = ""
    if params:
        query = "?" + parse.urlencode({key: value for key, value in params.items() if value not in (None, "", [])}, doseq=True)
    req = request.Request(server + path + query)
    req.add_header("Authorization", f"Bearer {_grafana_token()}")
    req.add_header("Accept", "application/json")
    req.add_header("User-Agent", _env("RSI_GRAFANA_USER_AGENT", DEFAULT_USER_AGENT))
    cf_client_id = _env("RSI_GRAFANA_CF_ACCESS_CLIENT_ID")
    cf_client_secret = _env("RSI_GRAFANA_CF_ACCESS_CLIENT_SECRET")
    if cf_client_id and cf_client_secret:
        req.add_header("CF-Access-Client-Id", cf_client_id)
        req.add_header("CF-Access-Client-Secret", cf_client_secret)
    try:
        with request.urlopen(req, timeout=30) as response:
            raw = response.read().decode("utf-8")
    except error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        raise GrafanaObservabilityError(f"Grafana API request failed: HTTP {exc.code}: {detail}") from exc
    return json.loads(raw) if raw else {}


def _request_object(path: str, params: JsonObject | None = None) -> JsonObject:
    payload = _request_json(path, params)
    if not isinstance(payload, dict):
        raise GrafanaObservabilityError("Grafana API response was not a JSON object")
    return payload


def _duration_seconds(value: str) -> int:
    match = re.fullmatch(r"\s*(\d+)\s*([smhdw])\s*", value or "")
    if not match:
        raise GrafanaObservabilityError("--since must be a duration like 30m, 6h, or 7d")
    amount = int(match.group(1))
    unit = match.group(2)
    multiplier = {"s": 1, "m": 60, "h": 3600, "d": 86400, "w": 604800}[unit]
    return amount * multiplier


def _unix_seconds(value: str) -> str:
    text = str(value or "").strip()
    if not text:
        return ""
    if text == "now":
        return str(_now_seconds())
    if text.startswith("now-"):
        return str(_now_seconds() - _duration_seconds(text.removeprefix("now-")))
    if re.fullmatch(r"\d+(?:\.\d+)?", text):
        return text
    parsed = datetime.fromisoformat(text.replace("Z", "+00:00"))
    return str(parsed.timestamp())


def _unix_nanoseconds(value: str) -> str:
    text = str(value or "").strip()
    if not text:
        return ""
    if text == "now":
        return str(int(_now_seconds() * 1_000_000_000))
    if text.startswith("now-"):
        return str(int((_now_seconds() - _duration_seconds(text.removeprefix("now-"))) * 1_000_000_000))
    if re.fullmatch(r"\d{16,}", text):
        return text
    if re.fullmatch(r"\d+(?:\.\d+)?", text):
        return str(int(float(text) * 1_000_000_000))
    parsed = datetime.fromisoformat(text.replace("Z", "+00:00"))
    return str(int(parsed.timestamp() * 1_000_000_000))


def _now_seconds() -> float:
    return time.time()


def _positive_limit(value: int, default: int = 50) -> int:
    try:
        parsed = int(value)
    except (TypeError, ValueError):
        return default
    return parsed if parsed > 0 else default


def datasources_query(datasource_type: str = "") -> JsonObject:
    payload = _request_json("/api/datasources")
    if not isinstance(payload, list):
        raise GrafanaObservabilityError("Grafana datasource list response was not a JSON array")
    items = payload
    if datasource_type:
        items = [item for item in items if isinstance(item, dict) and item.get("type") == datasource_type]
    return {"datasources": items}


def metrics_query(
    expr: str,
    *,
    datasource: str = "",
    range_query: bool = False,
    since: str = "1h",
    start: str = "",
    end: str = "",
    step: str = "",
) -> JsonObject:
    datasource_uid = datasource or _env("RSI_GRAFANA_METRICS_DATASOURCE_UID", DEFAULT_METRICS_DATASOURCE_UID)
    if range_query or start or end or step:
        now = _now_seconds()
        end_value = _unix_seconds(end) if end else str(now)
        start_value = _unix_seconds(start) if start else str(float(end_value) - _duration_seconds(since or "1h"))
        params: JsonObject = {"query": expr, "start": start_value, "end": end_value, "step": step or "60s"}
        path = f"/api/datasources/proxy/uid/{parse.quote(datasource_uid, safe='')}/api/v1/query_range"
    else:
        params = {"query": expr}
        path = f"/api/datasources/proxy/uid/{parse.quote(datasource_uid, safe='')}/api/v1/query"
    return _request_object(path, params)


def logs_query(
    expr: str,
    *,
    datasource: str = "",
    since: str = "1h",
    start: str = "",
    end: str = "",
    limit: int = 50,
    direction: str = "backward",
    step: str = "",
) -> JsonObject:
    if direction not in {"forward", "backward"}:
        raise GrafanaObservabilityError("direction must be forward or backward")
    datasource_uid = datasource or _env("RSI_GRAFANA_LOGS_DATASOURCE_UID", DEFAULT_LOGS_DATASOURCE_UID)
    now = _now_seconds()
    end_value = _unix_nanoseconds(end) if end else str(int(now * 1_000_000_000))
    start_value = _unix_nanoseconds(start) if start else str(int(end_value) - _duration_seconds(since or "1h") * 1_000_000_000)
    params: JsonObject = {
        "query": expr,
        "start": start_value,
        "end": end_value,
        "limit": limit,
        "direction": direction,
    }
    if step:
        params["step"] = step
    path = f"/api/datasources/proxy/uid/{parse.quote(datasource_uid, safe='')}/loki/api/v1/query_range"
    return _request_object(path, params)


def loki_log_lines(payload: JsonObject) -> list[str]:
    lines: list[str] = []
    for stream in payload.get("data", {}).get("result", []):
        if not isinstance(stream, dict):
            continue
        for value in stream.get("values", []):
            if isinstance(value, list) and len(value) >= 2:
                lines.append(str(value[1]))
    return lines


def dashboards_search(query: str = "", *, tags: list[str] | None = None, limit: int = 50) -> JsonObject:
    params: JsonObject = {
        "type": "dash-db",
        "query": query,
        "limit": _positive_limit(limit, 50),
    }
    if tags:
        params["tag"] = [str(item) for item in tags if str(item or "").strip()]
    payload = _request_json("/api/search", params)
    if not isinstance(payload, list):
        raise GrafanaObservabilityError("Grafana dashboard search response was not a JSON array")
    return {"dashboards": payload}


def dashboard_get(uid: str) -> JsonObject:
    uid_text = str(uid or "").strip()
    if not uid_text:
        raise GrafanaObservabilityError("dashboard uid is required")
    return _request_object(f"/api/dashboards/uid/{parse.quote(uid_text, safe='')}")


def alert_rules_search(query: str = "", *, folder_uid: str = "", limit: int = 100) -> JsonObject:
    payload = _request_json("/api/v1/provisioning/alert-rules")
    if not isinstance(payload, list):
        raise GrafanaObservabilityError("Grafana alert-rules response was not a JSON array")
    query_text = str(query or "").strip().lower()
    folder_text = str(folder_uid or "").strip()
    rules: list[object] = []
    for item in payload:
        if not isinstance(item, dict):
            continue
        if folder_text and str(item.get("folderUID", "") or "").strip() != folder_text:
            continue
        if query_text:
            searchable = " ".join(
                str(item.get(key, "") or "")
                for key in ("uid", "title", "ruleGroup", "folderUID", "condition", "dashboardUid", "panelId")
            ).lower()
            if query_text not in searchable:
                continue
        rules.append(item)
        if len(rules) >= _positive_limit(limit, 100):
            break
    return {"alert_rules": rules}


def alert_rule_get(uid: str) -> JsonObject:
    uid_text = str(uid or "").strip()
    if not uid_text:
        raise GrafanaObservabilityError("alert rule uid is required")
    return _request_object(f"/api/v1/provisioning/alert-rules/{parse.quote(uid_text, safe='')}")


def active_alerts(filters: list[str] | None = None, *, active: bool = True, silenced: bool = False, inhibited: bool = False, limit: int = 100) -> JsonObject:
    payload = _request_json(
        "/api/alertmanager/grafana/api/v2/alerts",
        {
            "active": str(bool(active)).lower(),
            "silenced": str(bool(silenced)).lower(),
            "inhibited": str(bool(inhibited)).lower(),
            "filter": [str(item) for item in (filters or []) if str(item or "").strip()],
        },
    )
    if not isinstance(payload, list):
        raise GrafanaObservabilityError("Grafana active-alerts response was not a JSON array")
    return {"alerts": payload[: _positive_limit(limit, 100)]}
