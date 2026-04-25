from __future__ import annotations

import base64
from pathlib import Path
import re
import shutil
import subprocess

from .json_types import JsonObject


LOCAL_UPLOAD_MAX_BYTES = 10 * 1024 * 1024
SLACK_RENDERABLE_SOURCE_SUFFIXES = {".html", ".htm", ".svg"}
SLACK_RENDERED_IMAGE_SUFFIXES = {".png", ".jpg", ".jpeg", ".gif", ".webp"}


def prepare_local_slack_upload_payload(payload: JsonObject, resolved_path: Path) -> JsonObject:
    upload_path = slack_upload_share_path(resolved_path)
    stat = upload_path.stat()
    if stat.st_size <= 0:
        raise ValueError(f"slack.upload_file local file is empty: {upload_path}")
    if stat.st_size > LOCAL_UPLOAD_MAX_BYTES:
        raise ValueError(f"slack.upload_file local file exceeds {LOCAL_UPLOAD_MAX_BYTES} bytes: {upload_path}")
    updated = dict(payload)
    updated["content_base64"] = base64.b64encode(upload_path.read_bytes()).decode("ascii")
    if upload_path != resolved_path:
        updated["filename"] = upload_path.name
        updated["title"] = _first_non_empty(_string_value(updated.get("title")), resolved_path.stem)
        updated["alt_txt"] = _first_non_empty(
            _string_value(updated.get("alt_txt")),
            f"Rendered preview of {resolved_path.name}",
        )
        updated["source_path"] = str(resolved_path)
        updated["rendered_from_path"] = str(resolved_path)
    else:
        updated["filename"] = _first_non_empty(_string_value(updated.get("filename")), upload_path.name)
    updated["path"] = str(upload_path)
    return updated


def slack_upload_share_path(path: Path) -> Path:
    suffix = path.suffix.lower()
    if suffix in SLACK_RENDERED_IMAGE_SUFFIXES:
        return path
    if suffix not in SLACK_RENDERABLE_SOURCE_SUFFIXES:
        return path
    converter = shutil.which("rsvg-convert")
    if not converter:
        raise ValueError(
            "slack.upload_file cannot upload HTML/SVG diagram source; "
            "rsvg-convert is required to render a PNG preview"
        )
    share_dir = path.parent / ".slack-share"
    share_dir.mkdir(parents=True, exist_ok=True)
    png_path = share_dir / f"{path.stem}.png"
    svg_path = path
    if suffix in {".html", ".htm"}:
        svg_path = share_dir / f"{path.stem}.svg"
        svg_path.write_text(extract_inline_svg(path), encoding="utf-8")
    try:
        subprocess.run(
            [converter, "-f", "png", "-o", str(png_path), str(svg_path)],
            check=True,
            capture_output=True,
            timeout=30,
        )
    except (OSError, subprocess.SubprocessError) as exc:
        raise ValueError(f"slack.upload_file failed to render {path.name} as PNG: {exc}") from exc
    if not png_path.exists() or png_path.stat().st_size <= 0:
        raise ValueError(f"slack.upload_file rendered PNG is empty for {path.name}")
    return png_path


def extract_inline_svg(path: Path) -> str:
    text = path.read_text(encoding="utf-8", errors="replace")
    match = re.search(r"<svg\b[\s\S]*</svg>", text, flags=re.IGNORECASE)
    if not match:
        raise ValueError(f"slack.upload_file cannot render HTML artifact without an inline SVG: {path}")
    return match.group(0)


def _first_non_empty(*values: str) -> str:
    for value in values:
        text = str(value or "").strip()
        if text:
            return text
    return ""


def _string_value(value: object) -> str:
    return str(value or "").strip()
