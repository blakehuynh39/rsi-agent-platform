"""Shared file and JSON utility functions."""
from __future__ import annotations

import json
import os
import tempfile
from pathlib import Path
from typing import Any

from .json_types import JsonObject


def _fsync_parent(path: Path) -> None:
    """Fsync the parent directory of a path to ensure durability.
    
    Args:
        path: Path whose parent directory should be fsync'd
    """
    try:
        flags = getattr(os, "O_DIRECTORY", 0) | os.O_RDONLY
        fd = os.open(str(path.parent), flags)
    except OSError:
        return
    try:
        os.fsync(fd)
    except OSError:
        pass
    finally:
        os.close(fd)


def _atomic_write_json(path: Path, payload: JsonObject) -> None:
    """Atomically write JSON payload to a file with fsync guarantees.
    
    Creates parent directories if needed, writes to a temporary file,
    fsyncs, then atomically replaces the target file.
    
    Args:
        path: Target file path
        payload: JSON-serializable dictionary to write
    """
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temp_name = tempfile.mkstemp(prefix=path.name + ".", suffix=".tmp", dir=str(path.parent))
    temp_path = Path(temp_name)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            json.dump(payload, handle, ensure_ascii=True, indent=2, sort_keys=True)
            handle.write("\n")
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(temp_path, path)
        _fsync_parent(path)
    finally:
        if temp_path.exists():
            temp_path.unlink(missing_ok=True)


def _json_object_from_string(value: Any) -> JsonObject:
    """Parse a JSON object from a string or return the value if already a dict.
    
    Args:
        value: String to parse or existing dict
        
    Returns:
        Parsed dictionary or empty dict if parsing fails
    """
    if isinstance(value, dict):
        return value
    if not isinstance(value, str) or not value.strip():
        return {}
    try:
        parsed = json.loads(value)
    except (TypeError, ValueError, json.JSONDecodeError):
        return {}
    return parsed if isinstance(parsed, dict) else {}
