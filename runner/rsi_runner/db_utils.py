"""Shared database utility functions."""
from __future__ import annotations

import os


def float_env(name: str, default: float) -> float:
    """Parse a float value from an environment variable.
    
    Args:
        name: Environment variable name
        default: Default value if the env var is not set or invalid
        
    Returns:
        The parsed float value, or default if parsing fails or value is non-positive
    """
    raw = os.getenv(name)
    if raw is None or not str(raw).strip():
        return default
    try:
        value = float(str(raw).strip())
        if not (value > 0.0 and value != float('inf')):
            return default
        return value
    except ValueError:
        return default


def sqlite_error_is_locked(exc: Exception) -> bool:
    """Check if a SQLite exception indicates a database lock.
    
    Args:
        exc: Exception to check
        
    Returns:
        True if the exception message indicates a database lock
    """
    text = str(exc).lower()
    return "database is locked" in text or "database is busy" in text or "database table is locked" in text
