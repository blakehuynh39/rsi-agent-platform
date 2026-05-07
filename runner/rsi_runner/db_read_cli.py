from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="rsi-db", description="RSI Slack-approved Postgres read gateway client")
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("sources", help="List configured DB read targets")
    schema = sub.add_parser("schema", help="Show allowlisted schema metadata for a target")
    schema.add_argument("--target", required=True)
    validate = sub.add_parser("validate", help="Validate SQL without requesting approval")
    validate.add_argument("--target", required=True)
    validate.add_argument("--sql", required=True)
    validate.add_argument("--purpose", default="query")
    query = sub.add_parser("query", help="Create a Slack approval request for a DB read")
    query.add_argument("--target", required=True)
    query.add_argument("--sql", required=True)
    query.add_argument("--purpose", default="query")
    query.add_argument("--requester", default=os.getenv("RSI_TASK_REQUESTER", "hermes"))
    query.add_argument("--conversation-id", default=os.getenv("RSI_CONVERSATION_ID", ""))
    query.add_argument("--workflow-id", default=os.getenv("RSI_WORKFLOW_ID", ""))
    query.add_argument("--trace-id", default=os.getenv("RSI_TRACE_ID", ""))
    query.add_argument("--channel-id", default=os.getenv("RSI_SLACK_CHANNEL_ID", ""))
    query.add_argument("--thread-ts", default=os.getenv("RSI_SLACK_THREAD_TS", ""))
    status = sub.add_parser("status", help="Show request status")
    status.add_argument("request_id")
    args = parser.parse_args(argv)

    try:
        if args.command == "sources":
            payload = request_json("GET", "/internal/db-read/sources")
        elif args.command == "schema":
            payload = request_json("GET", "/internal/db-read/schema?" + urllib.parse.urlencode({"target": args.target}))
        elif args.command == "validate":
            payload = request_json("POST", "/internal/db-read/validate", {"target": args.target, "sql": args.sql, "purpose": args.purpose})
        elif args.command == "query":
            payload = request_json(
                "POST",
                "/internal/db-read/query",
                {
                    "target": args.target,
                    "sql": args.sql,
                    "purpose": args.purpose,
                    "requester": args.requester,
                    "conversation_id": args.conversation_id,
                    "workflow_id": args.workflow_id,
                    "trace_id": args.trace_id,
                    "channel_id": args.channel_id,
                    "thread_ts": args.thread_ts,
                },
            )
        elif args.command == "status":
            payload = request_json("GET", f"/internal/db-read/requests/{urllib.parse.quote(args.request_id)}")
        else:
            parser.error("unknown command")
            return 2
    except RuntimeError as exc:
        print(str(exc), file=sys.stderr)
        return 1
    print(json.dumps(payload, indent=2, sort_keys=True))
    return 0


def request_json(method: str, path: str, body: dict[str, object] | None = None) -> object:
    base_url = os.getenv("RSI_CONTROL_PLANE_BASE_URL", "").strip().rstrip("/")
    token = os.getenv("RSI_DB_READ_CLIENT_TOKEN", "").strip()
    if not base_url:
        raise RuntimeError("RSI_CONTROL_PLANE_BASE_URL is required")
    if not token:
        raise RuntimeError("RSI_DB_READ_CLIENT_TOKEN is required")
    data = None
    headers = {"Authorization": f"Bearer {token}", "Accept": "application/json"}
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(base_url + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            raw = resp.read()
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"rsi-db request failed: HTTP {exc.code}: {detail}") from exc
    except urllib.error.URLError as exc:
        raise RuntimeError(f"rsi-db request failed: {exc}") from exc
    if not raw:
        return {}
    return json.loads(raw.decode("utf-8"))


if __name__ == "__main__":
    raise SystemExit(main())

