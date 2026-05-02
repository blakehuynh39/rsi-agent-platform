from __future__ import annotations

import json
import os
import re
import urllib.parse
import urllib.request
from typing import Any

from mcp.server.fastmcp import FastMCP
from mcp.types import ToolAnnotations


def _env(name: str, default: str = "") -> str:
    return os.getenv(name, default).strip()


def _bot_token() -> str:
    token = _env("SLACK_BOT_TOKEN")
    if not token:
        raise RuntimeError("SLACK_BOT_TOKEN is not configured")
    return token


def _slack_api(method: str, params: dict[str, Any]) -> dict[str, Any]:
    data = urllib.parse.urlencode({key: str(value) for key, value in params.items() if value is not None and str(value) != ""}).encode("utf-8")
    req = urllib.request.Request(
        f"https://slack.com/api/{method}",
        data=data,
        headers={
            "Authorization": f"Bearer {_bot_token()}",
            "Content-Type": "application/x-www-form-urlencoded",
            "User-Agent": "rsi-hermes-slack-mcp/1.0",
        },
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=20) as resp:
        payload = json.loads(resp.read().decode("utf-8"))
    if not isinstance(payload, dict):
        raise RuntimeError(f"Slack API {method} returned a non-object response")
    if not payload.get("ok"):
        error = str(payload.get("error") or "unknown_error")
        needed = str(payload.get("needed") or "")
        provided = str(payload.get("provided") or "")
        detail = f"Slack API {method} failed: {error}"
        if needed:
            detail += f" needed={needed}"
        if provided:
            detail += f" provided={provided}"
        raise RuntimeError(detail)
    return payload


def _ts_from_permalink(permalink: str) -> tuple[str, str]:
    text = str(permalink or "").strip()
    match = re.search(r"/archives/([A-Z0-9]+)/p([0-9]{10})([0-9]{6})", text)
    if not match:
        raise RuntimeError("Slack permalink must look like https://.../archives/<channel>/p<seconds><micros>")
    return match.group(1), f"{match.group(2)}.{match.group(3)}"


def _normalize_messages(messages: Any) -> list[dict[str, Any]]:
    out: list[dict[str, Any]] = []
    for item in messages if isinstance(messages, list) else []:
        if not isinstance(item, dict):
            continue
        out.append(
            {
                "type": item.get("type", ""),
                "subtype": item.get("subtype", ""),
                "user": item.get("user", ""),
                "bot_id": item.get("bot_id", ""),
                "username": item.get("username", ""),
                "text": item.get("text", ""),
                "ts": item.get("ts", ""),
                "thread_ts": item.get("thread_ts", ""),
                "reply_count": item.get("reply_count", 0),
                "permalink": item.get("permalink", ""),
            }
        )
    return out


def _message_permalink(channel_id: str, ts: str) -> str:
    try:
        payload = _slack_api("chat.getPermalink", {"channel": channel_id, "message_ts": ts})
        return str(payload.get("permalink") or "")
    except Exception:
        return ""


read_only = ToolAnnotations(readOnlyHint=True, destructiveHint=False, idempotentHint=True, openWorldHint=True)
mcp = FastMCP(
    "rsi-hermes-slack",
    instructions="Read Slack context visible to the RSI Slack bot token.",
    host=_env("RSI_SLACK_MCP_HOST", "127.0.0.1"),
    port=int(_env("RSI_SLACK_MCP_PORT", "8092")),
    streamable_http_path=_env("RSI_SLACK_MCP_PATH", "/mcp"),
    stateless_http=True,
    json_response=True,
)


@mcp.tool(
    name="slack_read_thread",
    description="Read a Slack thread by channel_id and thread_ts using the configured Slack bot token.",
    annotations=read_only,
)
def slack_read_thread(channel_id: str, thread_ts: str, limit: int = 100, cursor: str = "") -> dict[str, Any]:
    safe_limit = max(1, min(int(limit or 100), 200))
    payload = _slack_api(
        "conversations.replies",
        {
            "channel": channel_id,
            "ts": thread_ts,
            "limit": safe_limit,
            "cursor": cursor,
            "inclusive": "true",
        },
    )
    messages = _normalize_messages(payload.get("messages"))
    for message in messages:
        if message.get("ts") and not message.get("permalink"):
            message["permalink"] = _message_permalink(channel_id, str(message["ts"]))
    metadata = payload.get("response_metadata") if isinstance(payload.get("response_metadata"), dict) else {}
    return {
        "channel_id": channel_id,
        "thread_ts": thread_ts,
        "messages": messages,
        "next_cursor": str(metadata.get("next_cursor") or ""),
    }


@mcp.tool(
    name="slack_read_permalink",
    description="Read the Slack thread containing a permalink. The bot must be a member of the target channel.",
    annotations=read_only,
)
def slack_read_permalink(permalink: str, limit: int = 100) -> dict[str, Any]:
    channel_id, message_ts = _ts_from_permalink(permalink)
    thread = slack_read_thread(channel_id=channel_id, thread_ts=message_ts, limit=limit)
    thread["permalink"] = permalink
    return thread


def main() -> None:
    _bot_token()
    mcp.run(transport="streamable-http")


if __name__ == "__main__":
    main()
