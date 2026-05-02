from __future__ import annotations

import base64
import json
import os
import re
import hashlib
import tempfile
import urllib.parse
import urllib.error
import urllib.request
from pathlib import Path
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


def _honcho_api_raw(method: str, path: str, payload: dict[str, Any] | None = None) -> Any:
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
        return json.loads(resp.read().decode("utf-8"))


def _honcho_api(method: str, path: str, payload: dict[str, Any] | None = None) -> dict[str, Any]:
    decoded = _honcho_api_raw(method, path, payload)
    if not isinstance(decoded, dict):
        raise RuntimeError(f"Honcho API {path} returned a non-object response")
    return decoded


def _control_plane_base_url() -> str:
    return _env("RSI_CONTROL_PLANE_BASE_URL").rstrip("/")


def _control_plane_api(method: str, path: str, payload: dict[str, Any]) -> dict[str, Any]:
    base_url = _control_plane_base_url()
    if not base_url:
        raise RuntimeError("RSI_CONTROL_PLANE_BASE_URL is required for Slack attachment extraction persistence")
    req = urllib.request.Request(
        f"{base_url}{path}",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Content-Type": "application/json",
            "User-Agent": "rsi-hermes-company-knowledge-mcp/1.0",
        },
        method=method,
    )
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            decoded = json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"Control plane API {path} failed: HTTP {exc.code}: {detail}") from exc
    if not isinstance(decoded, dict):
        raise RuntimeError(f"Control plane API {path} returned a non-object response")
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
                "files": _normalize_files(item.get("files")),
            }
        )
    return out


def _normalize_files(files: Any) -> list[dict[str, Any]]:
    out: list[dict[str, Any]] = []
    for item in files if isinstance(files, list) else []:
        if not isinstance(item, dict):
            continue
        out.append(
            {
                "id": str(item.get("id") or ""),
                "name": str(item.get("name") or ""),
                "title": str(item.get("title") or ""),
                "mimetype": str(item.get("mimetype") or item.get("mime_type") or ""),
                "filetype": str(item.get("filetype") or item.get("file_type") or ""),
                "size": item.get("size") or 0,
                "permalink": str(item.get("permalink") or ""),
            }
        )
    return out


def _slack_file_info(file_id: str) -> dict[str, Any]:
    file_id = str(file_id or "").strip()
    if not file_id:
        raise RuntimeError("Slack file id is required")
    payload = _slack_api("files.info", {"file": file_id})
    file_info = payload.get("file")
    if not isinstance(file_info, dict):
        raise RuntimeError(f"Slack files.info returned no file object for {file_id}")
    return file_info


def _download_slack_file(file_id: str, max_bytes: int) -> tuple[bytes, dict[str, Any]]:
    file_info = _slack_file_info(file_id)
    url = str(file_info.get("url_private_download") or file_info.get("url_private") or "").strip()
    if not url:
        raise RuntimeError(f"Slack file {file_id} has no private download URL")
    req = urllib.request.Request(
        url,
        headers={
            "Authorization": f"Bearer {_bot_token()}",
            "User-Agent": "rsi-hermes-slack-mcp/1.0",
        },
        method="GET",
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        content_length = resp.headers.get("Content-Length")
        if content_length and int(content_length) > max_bytes:
            raise RuntimeError(f"Slack file {file_id} is too large to fetch: {content_length} bytes > {max_bytes}")
        data = resp.read(max_bytes + 1)
    if len(data) > max_bytes:
        raise RuntimeError(f"Slack file {file_id} is too large to fetch: read more than {max_bytes} bytes")
    return data, file_info


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
    channel_id = str(channel_id or "").strip()
    if not channel_id or channel_id in _denied_channels():
        return False
    if _slack_mirror_channel_discovery() == "joined":
        return True
    configured = _allowlisted_channels()
    return bool(configured) and channel_id in configured


def _slack_mirror_channel_discovery() -> str:
    return str(_env("RSI_SLACK_MIRROR_CHANNEL_DISCOVERY") or "joined").strip().lower() or "joined"


def _allowlisted_channels() -> list[str]:
    return [item.strip() for item in _env("RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST").split(",") if item.strip()]


def _denied_channels() -> set[str]:
    return {item.strip() for item in _env("RSI_SLACK_MIRROR_CHANNEL_DENYLIST").split(",") if item.strip()}


def _message_slack_ts(message: dict[str, Any]) -> str:
    metadata = message.get("metadata") if isinstance(message.get("metadata"), dict) else {}
    return str(metadata.get("slack_ts") or "")


def _message_metadata(message: dict[str, Any]) -> dict[str, Any]:
    metadata = message.get("metadata")
    return metadata if isinstance(metadata, dict) else {}


def _message_files(message: dict[str, Any]) -> list[dict[str, Any]]:
    metadata = _message_metadata(message)
    files = metadata.get("files")
    if not files:
        files = message.get("files")
    return _normalize_files(files)


def _parse_source_session_key(source_session_key: str) -> dict[str, str]:
    parts = str(source_session_key or "").split(":")
    if len(parts) < 4 or parts[0] != "slack":
        return {"workspace_id": "", "channel_id": "", "thread_ts": "", "conversation_type": ""}
    tail = ":".join(parts[3:])
    return {
        "workspace_id": parts[1],
        "channel_id": parts[2],
        "thread_ts": "" if tail == "channel" else tail,
        "conversation_type": "channel" if tail == "channel" else "thread",
    }


def _session_metadata(session: dict[str, Any]) -> dict[str, Any]:
    metadata = session.get("metadata")
    return metadata if isinstance(metadata, dict) else {}


def _normalize_session(session: dict[str, Any]) -> dict[str, Any] | None:
    metadata = _session_metadata(session)
    source_session_key = str(metadata.get("source_session_key") or "")
    parsed = _parse_source_session_key(source_session_key)
    channel_id = str(parsed.get("channel_id") or "")
    if not channel_id or not _allowlisted_channel(channel_id):
        return None
    return {
        "honcho_session_id": str(session.get("id") or ""),
        "source_session_key": source_session_key,
        "workspace_id": parsed.get("workspace_id", ""),
        "channel_id": channel_id,
        "thread_ts": parsed.get("thread_ts", ""),
        "conversation_type": parsed.get("conversation_type", ""),
        "created_at": session.get("created_at", ""),
        "metadata": metadata,
    }


def _ts_in_window(ts: str, oldest_ts: str, latest_ts: str) -> bool:
    if oldest_ts and ts and ts <= oldest_ts:
        return False
    if latest_ts and ts and ts > latest_ts:
        return False
    return True


def _slack_message_filters(channel_id: str = "") -> dict[str, Any]:
    metadata: dict[str, Any] = {"source": "slack"}
    if channel_id:
        if not _allowlisted_channel(channel_id):
            raise RuntimeError(f"Slack channel {channel_id} is not available in the mirrored Slack corpus")
        metadata["channel_id"] = channel_id
    else:
        channels = [] if _slack_mirror_channel_discovery() == "joined" else _allowlisted_channels()
        if _slack_mirror_channel_discovery() == "explicit" and not channels:
            raise RuntimeError("RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST is empty")
        if channels:
            filtered_channels = [channel for channel in channels if _allowlisted_channel(channel)]
            if not filtered_channels:
                raise RuntimeError("All allowlisted channels are denied by RSI_SLACK_MIRROR_CHANNEL_DENYLIST")
            metadata["channel_id"] = {"in": filtered_channels}
    return {"metadata": metadata}


def _slack_session_filters(channel_id: str = "") -> dict[str, Any]:
    channel_id = str(channel_id or "").strip()
    if channel_id:
        if not _allowlisted_channel(channel_id):
            raise RuntimeError(f"Slack channel {channel_id} is not available in the mirrored Slack corpus")
        return {
            "AND": [
                {"metadata": {"source": "slack"}},
                {"metadata": {"source_session_key": {"icontains": f":{channel_id}:"}}},
            ]
        }
    if _slack_mirror_channel_discovery() == "joined":
        return {"metadata": {"source": "slack"}}
    channels = _allowlisted_channels()
    if not channels:
        raise RuntimeError("RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST is empty")
    filtered_channels = [channel for channel in channels if _allowlisted_channel(channel)]
    if not filtered_channels:
        raise RuntimeError("All allowlisted channels are denied by RSI_SLACK_MIRROR_CHANNEL_DENYLIST")
    return {
        "AND": [
            {"metadata": {"source": "slack"}},
            {
                "OR": [
                    {"metadata": {"source_session_key": {"icontains": f":{allowed_channel}:"}}}
                    for allowed_channel in filtered_channels
                ]
            },
        ]
    }


def _document_filters(source: str = "") -> dict[str, Any]:
    if source == "notion":
        return {
            "observer_id": "notion_mirror",
            "observed_id": "story_company",
        }
    raise RuntimeError(f"Unsupported compiled document source {source!r}; supported sources: notion")


def _attachment_cache_root() -> Path:
    return Path(_env("RSI_ATTACHMENT_CACHE_ROOT", "/var/lib/hermes/attachments"))


def _safe_path_part(value: str) -> str:
    cleaned = "".join(ch if ch.isalnum() or ch in {"_", "-", "."} else "_" for ch in str(value or "").strip())
    cleaned = cleaned.strip("._-")
    return cleaned[:120] or "unknown"


def _cache_attachment_bytes(workspace_id: str, channel_id: str, file_id: str, file_name: str, data: bytes) -> dict[str, Any]:
    digest = hashlib.sha256(data).hexdigest()
    directory = _attachment_cache_root() / "slack" / _safe_path_part(workspace_id) / _safe_path_part(channel_id) / _safe_path_part(file_id) / digest[:24]
    directory.mkdir(parents=True, exist_ok=True)
    target = directory / _safe_path_part(file_name or file_id)
    with tempfile.NamedTemporaryFile(dir=directory, delete=False) as tmp:
        tmp.write(data)
        tmp_path = Path(tmp.name)
    os.replace(tmp_path, target)
    return {
        "cache_path": str(target),
        "content_sha256": digest,
        "content_size": len(data),
    }


def _is_text_attachment(file: dict[str, Any]) -> bool:
    mimetype = str(file.get("mimetype") or "").lower()
    filetype = str(file.get("filetype") or "").lower()
    name = str(file.get("name") or file.get("title") or "").lower()
    if mimetype.startswith("text/"):
        return True
    if mimetype in {
        "application/json",
        "application/xml",
        "application/yaml",
        "application/x-yaml",
        "application/javascript",
        "application/x-javascript",
        "application/typescript",
        "application/csv",
    }:
        return True
    return filetype in {"txt", "text", "md", "markdown", "json", "csv", "yaml", "yml", "xml", "log"} or name.endswith(
        (".txt", ".md", ".json", ".csv", ".yaml", ".yml", ".xml", ".log")
    )


def _is_image_attachment(file: dict[str, Any]) -> bool:
    mimetype = str(file.get("mimetype") or "").lower()
    filetype = str(file.get("filetype") or "").lower()
    name = str(file.get("name") or file.get("title") or "").lower()
    return mimetype.startswith("image/") or filetype in {"png", "jpg", "jpeg", "gif", "webp"} or name.endswith(
        (".png", ".jpg", ".jpeg", ".gif", ".webp")
    )


def _openrouter_api_key() -> str:
    return _env("RSI_OPENROUTER_API_KEY") or _env("OPENROUTER_API_KEY")


def _vision_model() -> str:
    return _env("RSI_ATTACHMENT_VISION_MODEL", "qwen/qwen3.6-flash")


def _openrouter_chat_completion(payload: dict[str, Any]) -> dict[str, Any]:
    token = _openrouter_api_key()
    if not token:
        raise RuntimeError("OPENROUTER_API_KEY is required for Slack attachment vision analysis")
    req = urllib.request.Request(
        _env("RSI_OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1").rstrip("/") + "/chat/completions",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
            "User-Agent": "rsi-hermes-company-knowledge-mcp/1.0",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            decoded = json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"OpenRouter vision analysis failed: HTTP {exc.code}: {detail}") from exc
    if not isinstance(decoded, dict):
        raise RuntimeError("OpenRouter vision analysis returned a non-object response")
    return decoded


def _vision_analyze_image(data: bytes, mimetype: str, prompt: str) -> dict[str, Any]:
    mimetype = str(mimetype or "image/png").strip() or "image/png"
    data_url = f"data:{mimetype};base64,{base64.b64encode(data).decode('ascii')}"
    model = _vision_model()
    payload = {
        "model": model,
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        "text": prompt
                        or "Extract all visible text and summarize the screenshot or image. Preserve concrete UI labels, errors, links, numbers, and code-like text.",
                    },
                    {"type": "image_url", "image_url": {"url": data_url}},
                ],
            }
        ],
        "max_tokens": 1200,
    }
    decoded = _openrouter_chat_completion(payload)
    choices = decoded.get("choices")
    if not isinstance(choices, list) or not choices:
        raise RuntimeError("OpenRouter vision analysis returned no choices")
    message = choices[0].get("message") if isinstance(choices[0], dict) else {}
    content = message.get("content") if isinstance(message, dict) else ""
    if isinstance(content, list):
        content = "\n".join(str(part.get("text") or "") for part in content if isinstance(part, dict))
    text = str(content or "").strip()
    if not text:
        raise RuntimeError("OpenRouter vision analysis returned empty text")
    return {"model": model, "text": text}


def _source_mirror_write_message(record: dict[str, Any], message: dict[str, Any]) -> dict[str, Any]:
    return _control_plane_api(
        "POST",
        "/internal/source-mirror/messages",
        {
            "record": record,
            "message": message,
            "lease_seconds": 300,
        },
    )


def _persist_attachment_analysis(
    *,
    workspace_id: str,
    channel_id: str,
    thread_ts: str,
    message_ts: str,
    message_permalink: str,
    message_id: str,
    file: dict[str, Any],
    content_sha256: str,
    extraction_kind: str,
    extraction_status: str,
    extracted_text: str,
    cache_path: str,
    error: str = "",
    model: str = "",
) -> dict[str, Any]:
    source_session_key = _source_session_key(workspace_id, channel_id, thread_ts)
    source_key = ":".join(
        [
            "slack_attachment_analysis",
            workspace_id,
            channel_id,
            message_ts,
            str(file.get("id") or ""),
            extraction_kind,
        ]
    )
    source_revision = f"{extraction_kind}:sha256:{content_sha256}:status:{extraction_status}:model:{model or 'none'}"
    metadata = {
        "source": "slack_attachment_analysis",
        "source_key": source_key,
        "source_dedupe_key": source_key,
        "source_revision": source_revision,
        "source_session_key": source_session_key,
        "workspace_id": workspace_id,
        "channel_id": channel_id,
        "thread_ts": thread_ts,
        "slack_ts": message_ts,
        "permalink": message_permalink,
        "source_message_id": message_id,
        "file": file,
        "file_id": str(file.get("id") or ""),
        "content_sha256": content_sha256,
        "cache_path": cache_path,
        "extraction_kind": extraction_kind,
        "extraction_status": extraction_status,
        "vision_model": model,
        "error": error,
    }
    content = extracted_text
    if not content:
        content = f"Slack attachment {file.get('id') or ''} extraction status: {extraction_status}. {error}".strip()
    return _source_mirror_write_message(
        {
            "source_type": "slack_attachment_analysis",
            "source_key": source_key,
            "workspace": workspace_id,
            "environment": _env("RSI_HONCHO_ENVIRONMENT", "stage"),
            "source_session_key": source_session_key,
            "honcho_workspace": _honcho_workspace_id(),
            "honcho_session_id": _honcho_session_id_for_source(source_session_key),
            "source_revision": source_revision,
            "metadata": metadata,
        },
        {
            "content": content,
            "peer_id": "rsi_attachment_analyzer",
            "metadata": metadata,
        },
    )


def _fetch_and_extract_attachment(
    *,
    workspace_id: str,
    channel_id: str,
    thread_ts: str,
    message_ts: str,
    message_permalink: str,
    message_id: str,
    file: dict[str, Any],
    analyze_images: bool,
    max_bytes: int,
    analysis_prompt: str,
) -> dict[str, Any]:
    file_id = str(file.get("id") or "").strip()
    if not file_id:
        raise RuntimeError("Slack attachment is missing file.id; cannot fetch content")
    data, file_info = _download_slack_file(file_id, max_bytes)
    merged_file = dict(file)
    for key in ("id", "name", "title", "mimetype", "filetype", "size", "permalink"):
        if not merged_file.get(key) and file_info.get(key) is not None:
            merged_file[key] = file_info.get(key)
    cache = _cache_attachment_bytes(
        workspace_id,
        channel_id,
        file_id,
        str(merged_file.get("name") or merged_file.get("title") or file_id),
        data,
    )
    content_sha256 = str(cache["content_sha256"])
    mimetype = str(merged_file.get("mimetype") or file_info.get("mimetype") or "")
    extraction_kind = "unsupported"
    extraction_status = "unsupported_binary"
    extracted_text = ""
    model = ""
    error = ""
    if _is_text_attachment(merged_file):
        extraction_kind = "text"
        extraction_status = "extracted"
        extracted_text = data.decode("utf-8", errors="replace")
    elif _is_image_attachment(merged_file):
        extraction_kind = "vision"
        if analyze_images:
            analysis = _vision_analyze_image(data, mimetype, analysis_prompt)
            extraction_status = "vision_analyzed"
            extracted_text = analysis["text"]
            model = analysis["model"]
        else:
            extraction_status = "requires_vision"
            error = "Image attachment cached but not analyzed; call with analyze_images=true when image content is required."
    else:
        error = "Unsupported Slack attachment type; no text extractor is configured for this MIME/filetype."

    persisted = _persist_attachment_analysis(
        workspace_id=workspace_id,
        channel_id=channel_id,
        thread_ts=thread_ts,
        message_ts=message_ts,
        message_permalink=message_permalink,
        message_id=message_id,
        file=_normalize_files([merged_file])[0],
        content_sha256=content_sha256,
        extraction_kind=extraction_kind,
        extraction_status=extraction_status,
        extracted_text=extracted_text,
        cache_path=str(cache["cache_path"]),
        error=error,
        model=model,
    )
    out = {
        "content_status": "cached",
        "cache_path": cache["cache_path"],
        "content_sha256": content_sha256,
        "content_size": cache["content_size"],
        "extraction_kind": extraction_kind,
        "extraction_status": extraction_status,
        "extraction_error": error,
        "extraction_note": _attachment_extraction_note(extraction_status),
        "honcho_persistence": persisted,
    }
    if extracted_text:
        out["extracted_text"] = extracted_text
    if model:
        out["vision_model"] = model
    return out


def _attachment_extraction_note(extraction_status: str) -> str:
    if extraction_status == "extracted":
        return "Text attachment content was extracted and persisted into Honcho with provenance."
    if extraction_status == "vision_analyzed":
        return "Image attachment content was analyzed by the configured auxiliary vision model and persisted into Honcho with provenance."
    if extraction_status == "requires_vision":
        return "Image attachment bytes were cached; call attachments_fetch with analyze_images=true when visual content is required."
    if extraction_status == "unsupported_binary":
        return "Attachment bytes were cached, but no supported extractor exists for this binary type."
    return f"Attachment extraction completed with status {extraction_status}."


read_only = ToolAnnotations(readOnlyHint=True, destructiveHint=False, idempotentHint=True, openWorldHint=True)
idempotent_write = ToolAnnotations(readOnlyHint=False, destructiveHint=False, idempotentHint=True, openWorldHint=True)
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
        raise RuntimeError(f"Slack channel {channel_id} is not available in the mirrored Slack corpus")
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
    name="conversations_list",
    description="List mirrored Slack conversations available in the Honcho company corpus. Results follow RSI Slack mirror channel policy.",
    annotations=read_only,
)
def conversations_list(channel_id: str = "", limit: int = 50, page: int = 1) -> dict[str, Any]:
    channel_id = str(channel_id or "").strip()
    safe_limit = max(1, min(int(limit or 50), 200))
    safe_page = max(1, int(page or 1))
    payload = _honcho_api(
        "POST",
        f"/workspaces/{_honcho_workspace_id()}/sessions/list?size={safe_limit}&page={safe_page}",
        {"filters": _slack_session_filters(channel_id)},
    )
    items = payload.get("items") if isinstance(payload.get("items"), list) else []
    conversations: list[dict[str, Any]] = []
    for item in items:
        if not isinstance(item, dict):
            continue
        normalized = _normalize_session(item)
        if normalized is not None:
            conversations.append(normalized)
    return {
        "source": "honcho_slack_corpus",
        "channel_id": channel_id,
        "conversations": conversations,
        "page": payload.get("page", safe_page),
        "pages": payload.get("pages", 1),
        "total": payload.get("total", len(conversations)),
    }


@mcp.tool(
    name="conversations_search",
    description="Search mirrored Slack messages in the Honcho company corpus. This does not call Slack search.messages and follows RSI Slack mirror channel policy.",
    annotations=read_only,
)
def conversations_search(query: str, channel_id: str = "", limit: int = 10) -> dict[str, Any]:
    query = str(query or "").strip()
    if not query:
        raise RuntimeError("conversations_search requires a non-empty query")
    safe_limit = max(1, min(int(limit or 10), 50))
    channel_id = str(channel_id or "").strip()
    payload = _honcho_api_raw(
        "POST",
        f"/workspaces/{_honcho_workspace_id()}/search",
        {
            "query": query,
            "filters": _slack_message_filters(channel_id),
            "limit": safe_limit,
        },
    )
    if not isinstance(payload, list):
        raise RuntimeError("Honcho workspace search returned a non-list response")
    results: list[dict[str, Any]] = []
    for item in payload:
        if not isinstance(item, dict):
            continue
        metadata = _message_metadata(item)
        result_channel = str(metadata.get("channel_id") or "")
        if result_channel and _allowlisted_channel(result_channel):
            results.append(item)
    return {
        "source": "honcho_slack_corpus",
        "query": query,
        "channel_id": channel_id,
        "results": results,
        "limit": safe_limit,
    }


@mcp.tool(
    name="documents_list",
    description="List mirrored company documents from Honcho. Currently exposes Notion mirrored documents with source provenance embedded in content.",
    annotations=read_only,
)
def documents_list(source: str = "notion", limit: int = 50, page: int = 1) -> dict[str, Any]:
    safe_limit = max(1, min(int(limit or 50), 200))
    safe_page = max(1, int(page or 1))
    source = str(source or "notion").strip().lower()
    payload = _honcho_api(
        "POST",
        f"/workspaces/{_honcho_workspace_id()}/conclusions/list?size={safe_limit}&page={safe_page}",
        {"filters": _document_filters(source)},
    )
    return {
        "source": f"honcho_{source}_documents",
        "documents": payload.get("items") if isinstance(payload.get("items"), list) else [],
        "page": payload.get("page", safe_page),
        "pages": payload.get("pages", 1),
        "total": payload.get("total", 0),
    }


@mcp.tool(
    name="documents_search",
    description="Search mirrored company documents in Honcho. Use this for Notion/wiki-like company context before live source fetches.",
    annotations=read_only,
)
def documents_search(query: str, source: str = "notion", limit: int = 10) -> dict[str, Any]:
    query = str(query or "").strip()
    if not query:
        raise RuntimeError("documents_search requires a non-empty query")
    safe_limit = max(1, min(int(limit or 10), 50))
    source = str(source or "notion").strip().lower()
    payload = _honcho_api_raw(
        "POST",
        f"/workspaces/{_honcho_workspace_id()}/conclusions/query",
        {
            "query": query,
            "top_k": safe_limit,
            "filters": _document_filters(source),
        },
    )
    if not isinstance(payload, list):
        raise RuntimeError("Honcho conclusions query returned a non-list response")
    return {
        "source": f"honcho_{source}_documents",
        "query": query,
        "results": payload,
        "limit": safe_limit,
    }


@mcp.tool(
    name="document_get",
    description="Read a mirrored company document by Honcho document/conclusion id.",
    annotations=read_only,
)
def document_get(document_id: str, source: str = "notion") -> dict[str, Any]:
    document_id = str(document_id or "").strip()
    if not document_id:
        raise RuntimeError("document_get requires document_id")
    source = str(source or "notion").strip().lower()
    filters = {
        "AND": [
            {"id": document_id},
            _document_filters(source),
        ]
    }
    payload = _honcho_api(
        "POST",
        f"/workspaces/{_honcho_workspace_id()}/conclusions/list?size=1&page=1",
        {"filters": filters},
    )
    items = payload.get("items") if isinstance(payload.get("items"), list) else []
    if not items:
        raise RuntimeError(f"Document {document_id} was not found in mirrored {source} documents")
    return {
        "source": f"honcho_{source}_documents",
        "document": items[0],
    }


@mcp.tool(
    name="conversation_get",
    description="Read one compiled Slack conversation from Honcho by channel_id and optional thread_ts. Uses mirrored company memory, not live Slack search.",
    annotations=read_only,
)
def conversation_get(channel_id: str, thread_ts: str = "", limit: int = 50, page: int = 1) -> dict[str, Any]:
    if not _allowlisted_channel(channel_id):
        raise RuntimeError(f"Slack channel {channel_id} is not available in the mirrored Slack corpus")
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
    name="attachments_fetch",
    description="Fetch Slack attachment metadata, optionally lazy-download/cache content, extract text, and persist extraction provenance into Honcho idempotently.",
    annotations=idempotent_write,
)
def attachments_fetch(
    channel_id: str = "",
    thread_ts: str = "",
    message_ts: str = "",
    permalink: str = "",
    limit: int = 50,
    page: int = 1,
    include_content: bool = False,
    analyze_images: bool = False,
    max_bytes: int = 2_000_000,
    analysis_prompt: str = "",
) -> dict[str, Any]:
    source = "honcho_slack_corpus"
    if str(permalink or "").strip():
        channel_id, parsed_ts = _ts_from_permalink(permalink)
        message_ts = str(message_ts or "").strip() or parsed_ts
        thread_ts = str(thread_ts or "").strip() or parsed_ts
        conversation = slack_read_thread(channel_id=channel_id, thread_ts=thread_ts, limit=limit)
        source = "slack_live_permalink"
    else:
        channel_id = str(channel_id or "").strip()
        if not channel_id:
            raise RuntimeError("attachments_fetch requires channel_id or permalink")
        conversation = conversation_get(channel_id=channel_id, thread_ts=thread_ts, limit=limit, page=page)

    if not _allowlisted_channel(channel_id):
        raise RuntimeError(f"Slack channel {channel_id} is not available in the mirrored Slack corpus")

    wanted_ts = str(message_ts or "").strip()
    attachments: list[dict[str, Any]] = []
    messages = conversation.get("messages") if isinstance(conversation.get("messages"), list) else []
    for message in messages:
        if not isinstance(message, dict):
            continue
        metadata = _message_metadata(message)
        slack_ts = str(metadata.get("slack_ts") or message.get("ts") or "")
        if wanted_ts and slack_ts != wanted_ts:
            continue
        effective_thread_ts = str(metadata.get("thread_ts") or message.get("thread_ts") or thread_ts or "")
        message_permalink = str(metadata.get("permalink") or message.get("permalink") or "")
        for file in _message_files(message):
            attachment = {
                "source": "slack",
                "source_channel_id": channel_id,
                "source_thread_ts": effective_thread_ts,
                "source_message_ts": slack_ts,
                "source_message_id": str(message.get("id") or ""),
                "source_message_permalink": message_permalink,
                "file": file,
                "content_status": "metadata_only",
                "extraction_status": "not_requested",
                "extraction_note": "Call attachments_fetch with include_content=true to lazily cache and extract supported attachment content.",
            }
            if include_content:
                attachment.update(
                    _fetch_and_extract_attachment(
                        workspace_id=str(metadata.get("workspace_id") or _slack_workspace_id()),
                        channel_id=channel_id,
                        thread_ts=effective_thread_ts,
                        message_ts=slack_ts,
                        message_permalink=message_permalink,
                        message_id=str(message.get("id") or ""),
                        file=file,
                        analyze_images=bool(analyze_images),
                        max_bytes=max(1, min(int(max_bytes or 2_000_000), 10_000_000)),
                        analysis_prompt=analysis_prompt,
                    )
                )
            attachments.append(
                attachment
            )
    return {
        "source": source,
        "channel_id": channel_id,
        "thread_ts": thread_ts,
        "message_ts": wanted_ts,
        "attachments": attachments,
        "attachment_count": len(attachments),
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
        raise RuntimeError(f"Slack channel {channel_id} is not available in the mirrored Slack corpus")
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
