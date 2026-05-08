from __future__ import annotations

from copy import deepcopy
from typing import Any, Iterable

from .json_types import JsonObject, JsonToolFunctionSchema


HERMES_ARTIFACT_TOOLSET = "rsi-artifacts"
HERMES_DB_READ_TOOLSET = "rsi-db-read"
HERMES_RSI_SLACK_TOOLSET = "rsi-slack"
HERMES_RSI_NOTION_TOOLSET = "rsi-notion"
HERMES_RSI_KNOWLEDGE_TOOLSET = "rsi-knowledge"
HERMES_RSI_OBSERVABILITY_TOOLSET = "rsi-observability"

_JSON_OBJECT_SCHEMA: JsonObject = {"type": "object"}
_JSON_ARRAY_SCHEMA: JsonObject = {"type": "array"}
_STRING_ARRAY_SCHEMA: JsonObject = {"type": "array", "items": {"type": "string"}}


def _schema(name: str, description: str, properties: JsonObject, required: list[str] | None = None) -> JsonToolFunctionSchema:
    return {
        "name": name,
        "description": description,
        "parameters": {
            "type": "object",
            "properties": properties,
            "required": required or [],
        },
    }


def _write_schema(
    name: str,
    description: str,
    properties: JsonObject,
    *,
    destructive: bool = False,
    required: list[str] | None = None,
) -> JsonToolFunctionSchema:
    merged = {
        **properties,
        "reason": {"type": "string", "description": "Why this external mutation is needed."},
        "idempotency_key": {"type": "string", "description": "Stable replay key for this exact mutation."},
    }
    required_fields = [*(required or []), "reason", "idempotency_key"]
    if destructive:
        merged["confirm_destroy"] = {"type": "boolean", "description": "Must be true for destructive operations."}
        required_fields.append("confirm_destroy")
    return _schema(name, description, merged, required_fields)

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

_SLACK_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "rsi_slack.channels_list": _schema(
        "rsi_slack.channels_list",
        "List Slack channels through the RSI native Slack gateway.",
        {
            "types": _STRING_ARRAY_SCHEMA,
            "include_archived": {"type": "boolean"},
            "limit": {"type": "integer"},
            "cursor": {"type": "string"},
        },
    ),
    "rsi_slack.channel_info": _schema(
        "rsi_slack.channel_info",
        "Read Slack channel metadata through the RSI native Slack gateway.",
        {"channel_id": {"type": "string"}},
        ["channel_id"],
    ),
    "rsi_slack.conversation_read": _schema(
        "rsi_slack.conversation_read",
        "Read Slack conversation or thread messages through the RSI native Slack gateway.",
        {
            "channel_id": {"type": "string"},
            "thread_ts": {"type": "string"},
            "oldest": {"type": "string"},
            "latest": {"type": "string"},
            "limit": {"type": "integer"},
            "include_replies": {"type": "boolean"},
        },
        ["channel_id"],
    ),
    "rsi_slack.user_lookup": _schema(
        "rsi_slack.user_lookup",
        "Resolve a Slack user by id, email, or display name through the RSI native Slack gateway.",
        {
            "user_id": {"type": "string"},
            "email": {"type": "string"},
            "name": {"type": "string"},
        },
    ),
    "rsi_slack.message_post": _write_schema(
        "rsi_slack.message_post",
        "Post a Slack message through the RSI native Slack gateway.",
        {
            "channel_id": {"type": "string"},
            "text": {"type": "string"},
            "thread_ts": {"type": "string"},
            "blocks": _JSON_ARRAY_SCHEMA,
            "attachments": _JSON_ARRAY_SCHEMA,
            "unfurl_links": {"type": "boolean"},
            "unfurl_media": {"type": "boolean"},
        },
        required=["channel_id", "text"],
    ),
    "rsi_slack.message_update": _write_schema(
        "rsi_slack.message_update",
        "Update a Slack message through the RSI native Slack gateway.",
        {
            "channel_id": {"type": "string"},
            "ts": {"type": "string"},
            "text": {"type": "string"},
            "blocks": _JSON_ARRAY_SCHEMA,
            "attachments": _JSON_ARRAY_SCHEMA,
        },
        required=["channel_id", "ts"],
    ),
    "rsi_slack.message_delete": _write_schema(
        "rsi_slack.message_delete",
        "Delete a Slack message through the RSI native Slack gateway.",
        {"channel_id": {"type": "string"}, "ts": {"type": "string"}},
        destructive=True,
        required=["channel_id", "ts"],
    ),
    "rsi_slack.reaction_add": _write_schema(
        "rsi_slack.reaction_add",
        "Add a Slack reaction through the RSI native Slack gateway.",
        {"channel_id": {"type": "string"}, "timestamp": {"type": "string"}, "name": {"type": "string"}},
        required=["channel_id", "timestamp", "name"],
    ),
    "rsi_slack.reaction_remove": _write_schema(
        "rsi_slack.reaction_remove",
        "Remove a Slack reaction through the RSI native Slack gateway.",
        {"channel_id": {"type": "string"}, "timestamp": {"type": "string"}, "name": {"type": "string"}},
        required=["channel_id", "timestamp", "name"],
    ),
    "rsi_slack.file_upload": _write_schema(
        "rsi_slack.file_upload",
        "Upload a file to Slack through the RSI native Slack gateway.",
        {
            "channel_id": {"type": "string"},
            "path": {"type": "string"},
            "artifact_ref": {"type": "string"},
            "content": {"type": "string"},
            "content_base64": {"type": "string"},
            "filename": {"type": "string"},
            "title": {"type": "string"},
            "initial_comment": {"type": "string"},
            "thread_ts": {"type": "string"},
        },
        required=["channel_id"],
    ),
    "rsi_slack.channel_create": _write_schema(
        "rsi_slack.channel_create",
        "Create a Slack channel through the RSI native Slack gateway.",
        {"name": {"type": "string"}, "is_private": {"type": "boolean"}},
        required=["name"],
    ),
    "rsi_slack.channel_rename": _write_schema(
        "rsi_slack.channel_rename",
        "Rename a Slack channel through the RSI native Slack gateway.",
        {"channel_id": {"type": "string"}, "name": {"type": "string"}},
        required=["channel_id", "name"],
    ),
    "rsi_slack.channel_archive": _write_schema(
        "rsi_slack.channel_archive",
        "Archive a Slack channel through the RSI native Slack gateway.",
        {"channel_id": {"type": "string"}},
        destructive=True,
        required=["channel_id"],
    ),
    "rsi_slack.channel_invite": _write_schema(
        "rsi_slack.channel_invite",
        "Invite users to a Slack channel through the RSI native Slack gateway.",
        {"channel_id": {"type": "string"}, "user_ids": _STRING_ARRAY_SCHEMA},
        required=["channel_id", "user_ids"],
    ),
}

_NOTION_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "rsi_notion.search": _schema(
        "rsi_notion.search",
        "Search Notion through the RSI native Notion gateway.",
        {"query": {"type": "string"}, "filter": _JSON_OBJECT_SCHEMA, "sort": _JSON_OBJECT_SCHEMA, "page_size": {"type": "integer"}, "cursor": {"type": "string"}},
    ),
    "rsi_notion.page_get": _schema("rsi_notion.page_get", "Retrieve a Notion page through the RSI native Notion gateway.", {"page_id": {"type": "string"}}, ["page_id"]),
    "rsi_notion.blocks_children": _schema("rsi_notion.blocks_children", "List Notion block children through the RSI native Notion gateway.", {"block_id": {"type": "string"}, "page_size": {"type": "integer"}, "cursor": {"type": "string"}}, ["block_id"]),
    "rsi_notion.database_get": _schema("rsi_notion.database_get", "Retrieve a Notion database through the RSI native Notion gateway.", {"database_id": {"type": "string"}}, ["database_id"]),
    "rsi_notion.data_source_get": _schema("rsi_notion.data_source_get", "Retrieve a Notion data source through the RSI native Notion gateway.", {"data_source_id": {"type": "string"}}, ["data_source_id"]),
    "rsi_notion.data_source_query": _schema(
        "rsi_notion.data_source_query",
        "Query a Notion data source through the RSI native Notion gateway.",
        {"data_source_id": {"type": "string"}, "filter": _JSON_OBJECT_SCHEMA, "sorts": _JSON_ARRAY_SCHEMA, "page_size": {"type": "integer"}, "cursor": {"type": "string"}},
        ["data_source_id"],
    ),
    "rsi_notion.page_create": _write_schema(
        "rsi_notion.page_create",
        "Create a Notion page through the RSI native Notion gateway.",
        {"parent": _JSON_OBJECT_SCHEMA, "properties": _JSON_OBJECT_SCHEMA, "children": _JSON_ARRAY_SCHEMA, "icon": _JSON_OBJECT_SCHEMA, "cover": _JSON_OBJECT_SCHEMA, "mirror_root_id": {"type": "string"}},
        required=["parent", "properties"],
    ),
    "rsi_notion.page_update": _write_schema(
        "rsi_notion.page_update",
        "Update a Notion page through the RSI native Notion gateway.",
        {"page_id": {"type": "string"}, "properties": _JSON_OBJECT_SCHEMA, "icon": _JSON_OBJECT_SCHEMA, "cover": _JSON_OBJECT_SCHEMA, "mirror_root_id": {"type": "string"}},
        required=["page_id"],
    ),
    "rsi_notion.page_archive": _write_schema(
        "rsi_notion.page_archive",
        "Archive a Notion page through the RSI native Notion gateway.",
        {"page_id": {"type": "string"}, "mirror_root_id": {"type": "string"}},
        destructive=True,
        required=["page_id"],
    ),
    "rsi_notion.blocks_append": _write_schema(
        "rsi_notion.blocks_append",
        "Append children to a Notion block through the RSI native Notion gateway.",
        {"block_id": {"type": "string"}, "children": _JSON_ARRAY_SCHEMA, "mirror_root_id": {"type": "string"}},
        required=["block_id", "children"],
    ),
    "rsi_notion.block_update": _write_schema(
        "rsi_notion.block_update",
        "Update a Notion block through the RSI native Notion gateway.",
        {"block_id": {"type": "string"}, "block": _JSON_OBJECT_SCHEMA, "mirror_root_id": {"type": "string"}},
        required=["block_id", "block"],
    ),
    "rsi_notion.block_delete": _write_schema(
        "rsi_notion.block_delete",
        "Delete a Notion block through the RSI native Notion gateway.",
        {"block_id": {"type": "string"}, "mirror_root_id": {"type": "string"}},
        destructive=True,
        required=["block_id"],
    ),
    "rsi_notion.comment_create": _write_schema(
        "rsi_notion.comment_create",
        "Create a Notion comment through the RSI native Notion gateway.",
        {"parent": _JSON_OBJECT_SCHEMA, "rich_text": _JSON_ARRAY_SCHEMA, "discussion_id": {"type": "string"}, "mirror_root_id": {"type": "string"}},
        required=["rich_text"],
    ),
}

_KNOWLEDGE_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "rsi_knowledge.search": _schema("rsi_knowledge.search", "Search mirrored company knowledge through the RSI native knowledge gateway.", {"query": {"type": "string"}, "limit": {"type": "integer"}, "source_types": _STRING_ARRAY_SCHEMA}, ["query"]),
    "rsi_knowledge.document_get": _schema("rsi_knowledge.document_get", "Retrieve a mirrored company knowledge document.", {"source_ref": {"type": "string"}, "document_id": {"type": "string"}}),
    "rsi_knowledge.conversation_get": _schema("rsi_knowledge.conversation_get", "Retrieve a mirrored conversation by source reference.", {"conversation_ref": {"type": "string"}, "channel_id": {"type": "string"}, "thread_ts": {"type": "string"}}),
    "rsi_knowledge.messages_read": _schema("rsi_knowledge.messages_read", "Read mirrored Slack messages from the company knowledge corpus.", {"channel_id": {"type": "string"}, "thread_ts": {"type": "string"}, "oldest": {"type": "string"}, "latest": {"type": "string"}, "limit": {"type": "integer"}}, ["channel_id"]),
    "rsi_knowledge.wiki_search": _schema("rsi_knowledge.wiki_search", "Search the synthesized company wiki.", {"query": {"type": "string"}, "limit": {"type": "integer"}}, ["query"]),
    "rsi_knowledge.wiki_page_get": _schema("rsi_knowledge.wiki_page_get", "Read a synthesized company wiki page.", {"page_ref": {"type": "string"}, "slug": {"type": "string"}}, ["page_ref"]),
    "rsi_knowledge.wiki_index_get": _schema("rsi_knowledge.wiki_index_get", "Read the synthesized company wiki index.", {}),
    "rsi_knowledge.wiki_log_get": _schema("rsi_knowledge.wiki_log_get", "Read the synthesized company wiki log.", {"limit": {"type": "integer"}}),
    "rsi_knowledge.source_status": _schema("rsi_knowledge.source_status", "Read source mirror freshness and status metadata.", {"source_types": _STRING_ARRAY_SCHEMA, "source_type": {"type": "string"}, "limit": {"type": "integer"}, "max_age_seconds": {"type": "integer"}}),
    "rsi_knowledge.wiki_edit_propose": _write_schema(
        "rsi_knowledge.wiki_edit_propose",
        "Record an audited proposal to edit the synthesized company wiki.",
        {"slug": {"type": "string"}, "page_ref": {"type": "string"}, "title": {"type": "string"}, "body": {"type": "string"}, "content": {"type": "string"}, "metadata": _JSON_OBJECT_SCHEMA},
        required=["title"],
    ),
    "rsi_knowledge.wiki_edit_apply": _write_schema(
        "rsi_knowledge.wiki_edit_apply",
        "Apply an audited edit to the synthesized company wiki.",
        {"slug": {"type": "string"}, "page_ref": {"type": "string"}, "title": {"type": "string"}, "body": {"type": "string"}, "content": {"type": "string"}, "metadata": _JSON_OBJECT_SCHEMA},
        required=["title", "body"],
    ),
}

_OBSERVABILITY_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "rsi_observability.datasources": _schema(
        "rsi_observability.datasources",
        "List Grafana datasources visible to the Hermes read-only observability token.",
        {
            "type": {
                "type": "string",
                "enum": ["loki", "prometheus", "tempo", "pyroscope"],
            },
        },
    ),
    "rsi_observability.metrics_query": _schema(
        "rsi_observability.metrics_query",
        "Run a read-only PromQL query through the Grafana datasource proxy.",
        {
            "expr": {"type": "string"},
            "datasource": {"type": "string"},
            "range": {"type": "boolean"},
            "since": {"type": "string", "description": "Duration such as 30m, 6h, or 7d when start is omitted."},
            "start": {"type": "string", "description": "Unix seconds, RFC3339, now, or now-<duration>."},
            "end": {"type": "string", "description": "Unix seconds, RFC3339, now, or now-<duration>."},
            "step": {"type": "string", "description": "Prometheus range step such as 60s or 5m."},
        },
        ["expr"],
    ),
    "rsi_observability.logs_query": _schema(
        "rsi_observability.logs_query",
        "Run a read-only LogQL range query through the Grafana Loki datasource proxy.",
        {
            "expr": {"type": "string"},
            "datasource": {"type": "string"},
            "since": {"type": "string", "description": "Duration such as 30m, 6h, or 7d when start is omitted."},
            "start": {"type": "string", "description": "Unix seconds, Unix nanoseconds, RFC3339, now, or now-<duration>."},
            "end": {"type": "string", "description": "Unix seconds, Unix nanoseconds, RFC3339, now, or now-<duration>."},
            "limit": {"type": "integer"},
            "direction": {
                "type": "string",
                "enum": ["forward", "backward"],
            },
            "step": {"type": "string", "description": "Optional Loki query step."},
        },
        ["expr"],
    ),
    "rsi_observability.dashboards_search": _schema(
        "rsi_observability.dashboards_search",
        "Search Grafana dashboards visible to the Hermes read-only observability token.",
        {
            "query": {"type": "string"},
            "tags": _STRING_ARRAY_SCHEMA,
            "limit": {"type": "integer"},
        },
    ),
    "rsi_observability.dashboard_get": _schema(
        "rsi_observability.dashboard_get",
        "Read one Grafana dashboard by UID through the read-only observability token.",
        {
            "uid": {"type": "string"},
        },
        ["uid"],
    ),
    "rsi_observability.alert_rules_search": _schema(
        "rsi_observability.alert_rules_search",
        "Search Grafana-managed alert rules visible to the Hermes read-only observability token.",
        {
            "query": {"type": "string"},
            "folder_uid": {"type": "string"},
            "limit": {"type": "integer"},
        },
    ),
    "rsi_observability.alert_rule_get": _schema(
        "rsi_observability.alert_rule_get",
        "Read one Grafana-managed alert rule by UID through the read-only observability token.",
        {
            "uid": {"type": "string"},
        },
        ["uid"],
    ),
    "rsi_observability.active_alerts": _schema(
        "rsi_observability.active_alerts",
        "List active Grafana Alertmanager alerts visible to the Hermes read-only observability token.",
        {
            "filters": _STRING_ARRAY_SCHEMA,
            "active": {"type": "boolean"},
            "silenced": {"type": "boolean"},
            "inhibited": {"type": "boolean"},
            "limit": {"type": "integer"},
        },
    ),
}

_DB_READ_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "db_read.sources": _schema(
        "db_read.sources",
        "List RSI Slack-approved Postgres read targets available to this execution.",
        {},
    ),
    "db_read.schema": _schema(
        "db_read.schema",
        "Show allowlisted schema metadata for one RSI DB-read target.",
        {"target": {"type": "string"}},
        ["target"],
    ),
    "db_read.validate": _schema(
        "db_read.validate",
        "Validate a read-only SQL query for one RSI DB-read target without requesting Slack approval.",
        {"target": {"type": "string"}, "sql": {"type": "string"}, "purpose": {"type": "string"}},
        ["target", "sql"],
    ),
    "db_read.query": _schema(
        "db_read.query",
        "Submit one Slack-approved read-only SQL query. After this tool succeeds, stop; the Slack approval/result card owns the response.",
        {"target": {"type": "string"}, "sql": {"type": "string"}, "purpose": {"type": "string"}},
        ["target", "sql"],
    ),
    "db_read.status": _schema(
        "db_read.status",
        "Show status for one RSI DB-read request.",
        {"request_id": {"type": "string"}},
        ["request_id"],
    ),
}

_TOOL_SCHEMAS = {
    **_ARTIFACT_TOOL_SCHEMAS,
    **_DB_READ_TOOL_SCHEMAS,
    **_SLACK_TOOL_SCHEMAS,
    **_NOTION_TOOL_SCHEMAS,
    **_KNOWLEDGE_TOOL_SCHEMAS,
    **_OBSERVABILITY_TOOL_SCHEMAS,
}
_TOOLSET_SCHEMAS = {
    HERMES_ARTIFACT_TOOLSET: _ARTIFACT_TOOL_SCHEMAS,
    HERMES_DB_READ_TOOLSET: _DB_READ_TOOL_SCHEMAS,
    HERMES_RSI_SLACK_TOOLSET: _SLACK_TOOL_SCHEMAS,
    HERMES_RSI_NOTION_TOOLSET: _NOTION_TOOL_SCHEMAS,
    HERMES_RSI_KNOWLEDGE_TOOLSET: _KNOWLEDGE_TOOL_SCHEMAS,
    HERMES_RSI_OBSERVABILITY_TOOLSET: _OBSERVABILITY_TOOL_SCHEMAS,
}
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
    for toolset, schemas in _TOOLSET_SCHEMAS.items():
        for canonical_name in sorted(schemas):
            schema = transport_tool_schema(canonical_name)
            definitions.append(
                {
                    "canonical_name": canonical_name,
                    "transport_name": schema["name"],
                    "toolset": toolset,
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
