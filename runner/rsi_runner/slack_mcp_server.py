from __future__ import annotations

import json
import os
import re
import hashlib
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


def _honcho_base_url() -> str:
    return _env("RSI_HONCHO_BASE_URL").rstrip("/")


def _honcho_api_base_url() -> str:
    base_url = _honcho_base_url()
    if not base_url:
        return ""
    if base_url.endswith("/v3"):
        return base_url
    return f"{base_url}/v3"


def _honcho_workspace_id() -> str:
    return _honcho_name(_env("RSI_HONCHO_WORKSPACE_ID", "rsi_company_knowledge"))


def _honcho_headers() -> dict[str, str]:
    headers = {
        "Content-Type": "application/json",
        "User-Agent": "rsi-hermes-company-knowledge-mcp/1.0",
    }
    token = _env("HONCHO_API_KEY")
    if token:
        headers["Authorization"] = token if token.lower().startswith("bearer ") else f"Bearer {token}"
    return headers


def _honcho_api(method: str, path: str, payload: dict[str, Any] | None = None) -> dict[str, Any]:
    base_url = _honcho_api_base_url()
    if not base_url:
        raise RuntimeError("RSI_HONCHO_BASE_URL is not configured")
    data = None
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        f"{base_url}{path}",
        data=data,
        headers=_honcho_headers(),
        method=method,
    )
    with urllib.request.urlopen(req, timeout=20) as resp:
        decoded = json.loads(resp.read().decode("utf-8"))
    if not isinstance(decoded, dict):
        raise RuntimeError(f"Honcho API {path} returned a non-object response")
    return decoded


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


def _slack_workspace_id() -> str:
    configured = _env("RSI_SLACK_WORKSPACE_ID")
    if configured:
        return configured
    payload = _slack_api("auth.test", {})
    team_id = str(payload.get("team_id") or "").strip()
    if not team_id:
        raise RuntimeError("Slack auth.test did not return team_id")
    return team_id


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


def _honcho_name(value: str) -> str:
    text = str(value or "").strip()
    cleaned = "".join(ch if ch.isalnum() or ch in {"_", "-"} else "_" for ch in text).strip("_-")
    if cleaned and len(cleaned) <= 100 and re.fullmatch(r"[A-Za-z0-9_-]+", cleaned):
        return cleaned
    digest = hashlib.sha256(text.encode("utf-8")).hexdigest()[:48]
    return f"rsi_{digest}"


def _source_session_key(workspace_id: str, channel_id: str, thread_ts: str = "") -> str:
    if str(thread_ts or "").strip():
        return f"slack:{workspace_id}:{channel_id}:{thread_ts}"
    return f"slack:{workspace_id}:{channel_id}:channel"


def _honcho_session_id_for_source(source_session_key: str) -> str:
    return _honcho_name(source_session_key)


def _allowlisted_channel(channel_id: str) -> bool:
    configured = [item.strip() for item in _env("RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST").split(",") if item.strip()]
    return bool(configured) and channel_id in configured


def _message_slack_ts(message: dict[str, Any]) -> str:
    metadata = message.get("metadata") if isinstance(message.get("metadata"), dict) else {}
    return str(metadata.get("slack_ts") or "")


def _ts_in_window(ts: str, oldest_ts: str, latest_ts: str) -> bool:
    if oldest_ts and ts and ts <= oldest_ts:
        return False
    if latest_ts and ts and ts > latest_ts:
        return False
    return True


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
    if not _allowlisted_channel(channel_id):
        raise RuntimeError(f"Slack channel {channel_id} is not in RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST")
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


@mcp.tool(
    name="conversation_get",
    description="Read one compiled Slack conversation from Honcho by channel_id and optional thread_ts. Uses mirrored company memory, not live Slack search.",
    annotations=read_only,
)
def conversation_get(channel_id: str, thread_ts: str = "", limit: int = 50, page: int = 1) -> dict[str, Any]:
    if not _allowlisted_channel(channel_id):
        raise RuntimeError(f"Slack channel {channel_id} is not in RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST")
    workspace_id = _slack_workspace_id()
    source_key = _source_session_key(workspace_id, channel_id, thread_ts)
    session_id = _honcho_session_id_for_source(source_key)
    safe_limit = max(1, min(int(limit or 50), 200))
    safe_page = max(1, int(page or 1))
    payload = _honcho_api(
        "POST",
        f"/workspaces/{_honcho_workspace_id()}/sessions/{session_id}/messages/list?size={safe_limit}&page={safe_page}",
        {},
    )
    items = payload.get("items") if isinstance(payload.get("items"), list) else []
    return {
        "source": "honcho_slack_corpus",
        "workspace_id": workspace_id,
        "channel_id": channel_id,
        "thread_ts": thread_ts,
        "source_session_key": source_key,
        "honcho_session_id": session_id,
        "messages": items,
        "page": payload.get("page", safe_page),
        "pages": payload.get("pages", 1),
        "total": payload.get("total", len(items)),
    }


@mcp.tool(
    name="messages_read",
    description="Read mirrored Slack messages from Honcho. Channel-wide reads must include oldest_ts or latest_ts and are paginated.",
    annotations=read_only,
)
def messages_read(
    channel_id: str,
    thread_ts: str = "",
    oldest_ts: str = "",
    latest_ts: str = "",
    limit: int = 50,
    page: int = 1,
) -> dict[str, Any]:
    if not _allowlisted_channel(channel_id):
        raise RuntimeError(f"Slack channel {channel_id} is not in RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST")
    if not str(thread_ts or "").strip() and not (str(oldest_ts or "").strip() or str(latest_ts or "").strip()):
        raise RuntimeError("Channel-wide messages_read requires oldest_ts or latest_ts; unbounded channel history reads are refused.")
    conversation = conversation_get(channel_id=channel_id, thread_ts=thread_ts, limit=limit, page=page)
    items = conversation.get("messages") if isinstance(conversation.get("messages"), list) else []
    if oldest_ts or latest_ts:
        items = [item for item in items if isinstance(item, dict) and _ts_in_window(_message_slack_ts(item), oldest_ts, latest_ts)]
    conversation["messages"] = items
    conversation["oldest_ts"] = oldest_ts
    conversation["latest_ts"] = latest_ts
    return conversation


def main() -> None:
    _bot_token()
    mcp.run(transport="streamable-http")


if __name__ == "__main__":
    main()
