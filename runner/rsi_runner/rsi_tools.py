from __future__ import annotations

from copy import deepcopy
from typing import Any, Iterable

from .json_types import JsonObject, JsonToolFunctionSchema


HERMES_ARTIFACT_TOOLSET = "rsi-artifacts"

_ARTIFACT_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "artifact.list_files": {
        "name": "artifact.list_files",
        "description": "List files inside the staged Hermes artifact output directory.",
        "parameters": {
            "type": "object",
            "properties": {
                "path": {"type": "string"},
            },
        },
    },
    "artifact.write_file": {
        "name": "artifact.write_file",
        "description": "Write file content inside the staged Hermes artifact output directory.",
        "parameters": {
            "type": "object",
            "properties": {
                "path": {"type": "string"},
                "content": {"type": "string"},
            },
            "required": ["path", "content"],
        },
    },
}

_TOOL_SCHEMAS = dict(_ARTIFACT_TOOL_SCHEMAS)
_TRANSPORT_SAFE_TOOL_CHARS = frozenset("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-")


def _is_transport_safe_tool_name(name: str) -> bool:
    return bool(name) and all(char in _TRANSPORT_SAFE_TOOL_CHARS for char in name)


def _canonical_to_transport_tool_name(name: str) -> str:
    canonical = str(name or "").strip()
    if not canonical:
        raise ValueError("tool name is empty")
    transport = canonical.replace(".", "_")
    if not _is_transport_safe_tool_name(transport):
        raise ValueError(f"tool name {canonical!r} cannot be mapped to a provider-safe transport name")
    return transport


def _build_tool_transport_maps() -> tuple[dict[str, str], dict[str, str]]:
    canonical_to_transport: dict[str, str] = {}
    transport_to_canonical: dict[str, str] = {}
    for canonical in sorted(_TOOL_SCHEMAS):
        transport = _canonical_to_transport_tool_name(canonical)
        existing = transport_to_canonical.get(transport)
        if existing is not None and existing != canonical:
            raise ValueError(f"transport tool name collision for {canonical!r} and {existing!r}: {transport!r}")
        canonical_to_transport[canonical] = transport
        transport_to_canonical[transport] = canonical
    return canonical_to_transport, transport_to_canonical


_CANONICAL_TO_TRANSPORT_TOOL_NAMES, _TRANSPORT_TO_CANONICAL_TOOL_NAMES = _build_tool_transport_maps()


def tool_transport_name(name: str) -> str:
    canonical = str(name or "").strip()
    if not canonical:
        raise ValueError("tool name is empty")
    if canonical in _CANONICAL_TO_TRANSPORT_TOOL_NAMES:
        return _CANONICAL_TO_TRANSPORT_TOOL_NAMES[canonical]
    if "." not in canonical and _is_transport_safe_tool_name(canonical):
        return canonical
    raise ValueError(f"tool name {canonical!r} is not transport-safe")


def canonical_tool_name(name: str) -> str:
    tool = str(name or "").strip()
    if not tool:
        raise ValueError("tool name is empty")
    if tool in _TOOL_SCHEMAS:
        return tool
    if tool in _TRANSPORT_TO_CANONICAL_TOOL_NAMES:
        return _TRANSPORT_TO_CANONICAL_TOOL_NAMES[tool]
    if "." not in tool and _is_transport_safe_tool_name(tool):
        return tool
    raise ValueError(f"tool name {tool!r} is not recognized")


def _nullable_json_schema(value: Any) -> Any:
    if not isinstance(value, dict):
        return value
    out = deepcopy(value)
    schema_type = out.get("type")
    if isinstance(schema_type, str):
        if schema_type != "null":
            out["type"] = [schema_type, "null"]
        return out
    if isinstance(schema_type, list):
        normalized = [item for item in schema_type if isinstance(item, str)]
        if "null" not in normalized:
            out["type"] = [*schema_type, "null"]
        return out
    enum_values = out.get("enum")
    if isinstance(enum_values, list) and None not in enum_values:
        out["enum"] = [*enum_values, None]
    return out


def _strict_json_schema(value: Any) -> Any:
    if isinstance(value, dict):
        out = {key: _strict_json_schema(item) for key, item in value.items()}
        if out.get("type") == "object":
            properties = out.get("properties")
            if isinstance(properties, dict):
                existing_required = {
                    str(item).strip()
                    for item in out.get("required", [])
                    if isinstance(item, str) and str(item).strip()
                }
                ordered_keys = list(properties.keys())
                for key in ordered_keys:
                    if key not in existing_required:
                        properties[key] = _nullable_json_schema(properties[key])
                out["required"] = ordered_keys
            else:
                out["required"] = []
            out["additionalProperties"] = False
        return out
    if isinstance(value, list):
        return [_strict_json_schema(item) for item in value]
    return value


def transport_tool_schema(name: str) -> JsonToolFunctionSchema:
    schema = _TOOL_SCHEMAS.get(name)
    if schema is None:
        raise KeyError(name)
    wrapped = deepcopy(schema)
    wrapped["name"] = tool_transport_name(name)
    if "parameters" in wrapped:
        wrapped["parameters"] = _strict_json_schema(wrapped.get("parameters"))
    return wrapped


def rsi_plugin_toolset_definitions() -> list[JsonObject]:
    definitions: list[JsonObject] = []
    for canonical_name in sorted(_ARTIFACT_TOOL_SCHEMAS):
        schema = transport_tool_schema(canonical_name)
        definitions.append(
            {
                "canonical_name": canonical_name,
                "transport_name": schema["name"],
                "toolset": HERMES_ARTIFACT_TOOLSET,
                "schema": schema,
            }
        )
    return definitions


def normalize_tool_names(values: Iterable[str]) -> list[str]:
    seen: set[str] = set()
    out: list[str] = []
    for value in values:
        name = str(value or "").strip()
        if not name or name in seen:
            continue
        seen.add(name)
        out.append(name)
    return out
