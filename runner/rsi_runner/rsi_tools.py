from __future__ import annotations

from copy import deepcopy
from typing import Any, Iterable

from .json_types import JsonObject, JsonToolFunctionSchema


HERMES_ARTIFACT_TOOLSET = "rsi-artifacts"
HERMES_DB_READ_TOOLSET = "rsi-db-read"
HERMES_RSI_SLACK_TOOLSET = "rsi-slack"
HERMES_RSI_NOTION_TOOLSET = "rsi-notion"
HERMES_RSI_KNOWLEDGE_TOOLSET = "rsi-knowledge"
HERMES_RSI_SENTRY_TOOLSET = "rsi-sentry"
HERMES_RSI_KANBAN_TOOLSET = "rsi-kanban"
HERMES_RSI_TEMPORAL_TOOLSET = "rsi-temporal"
HERMES_RSI_OBSERVABILITY_TOOLSET = "rsi-observability"
HERMES_RSI_AWS_TOOLSET = "rsi-aws"

_JSON_OBJECT_SCHEMA: JsonObject = {"type": "object"}
_JSON_ARRAY_SCHEMA: JsonObject = {"type": "array"}
_STRING_ARRAY_SCHEMA: JsonObject = {"type": "array", "items": {"type": "string"}}
_TEMPORAL_TARGET_PROPERTIES: JsonObject = {
    "environment": {"type": "string", "enum": ["stage", "prod"]},
    "target": {"type": "string", "description": "Configured Temporal target name such as royalty-graph-v2 or indexer."},
}
_TEMPORAL_MUTATION_PROPERTIES: JsonObject = {
    **_TEMPORAL_TARGET_PROPERTIES,
    "confirm": {"type": "boolean", "description": "Must be true after explicit operator authorization for live mutations."},
    "dry_run": {"type": "boolean", "description": "Validate policy and report what would be executed without connecting to Temporal."},
}
_AWS_READ_PROPERTIES: JsonObject = {
    "account": {"type": "string", "enum": ["stage", "staging", "prod", "production"], "description": "AWS account/environment to read."},
    "region": {"type": "string", "description": "AWS region, defaults to us-east-1."},
    "service": {
        "type": "string",
        "description": "Supported AWS diagnostic service, e.g. rds, cloudwatch, logs metadata, cloudtrail, ec2, eks, elbv2, autoscaling, sts.",
    },
    "operation": {
        "type": "string",
        "description": "Read-only AWS operation using CLI-style naming, e.g. describe-events, describe-db-instances, describe-db-parameters, lookup-events, get-caller-identity.",
    },
    "params": {
        "type": "object",
        "description": "Operation input object. Use AWS SDK-style field names or snake_case/kebab-case equivalents. Do not include secrets.",
    },
}
_SLACK_REPORT_COLUMN_SCHEMA: JsonObject = {
    "type": "object",
    "properties": {
        "key": {"type": "string"},
        "label": {"type": "string"},
        "align": {"type": "string", "enum": ["left", "center", "right"]},
    },
    "required": ["key", "label"],
}
_SLACK_REPORT_TABLE_SCHEMA: JsonObject = {
    "type": "object",
    "properties": {
        "title": {"type": "string"},
        "caption": {"type": "string"},
        "columns": {"type": "array", "items": _SLACK_REPORT_COLUMN_SCHEMA},
        "rows": {"type": "array", "items": {"type": "object"}},
    },
    "required": ["columns", "rows"],
}
_SLACK_REPORT_FILE_SCHEMA: JsonObject = {
    "type": "object",
    "properties": {
        "artifact_ref": {"type": "string"},
        "path": {"type": "string"},
        "filename": {"type": "string"},
        "title": {"type": "string"},
        "mime_type": {"type": "string"},
        "content": {"type": "string"},
        "content_base64": {"type": "string"},
    },
}


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
        "Post a Slack message through the RSI native Slack gateway. In RSI workflow threads, the platform supplies the bound channel and thread.",
        {
            "channel_id": {"type": "string"},
            "text": {"type": "string"},
            "thread_ts": {"type": "string"},
            "blocks": _JSON_ARRAY_SCHEMA,
            "attachments": _JSON_ARRAY_SCHEMA,
            "unfurl_links": {"type": "boolean"},
            "unfurl_media": {"type": "boolean"},
        },
        required=["text"],
    ),
    "rsi_slack.report_post": _write_schema(
        "rsi_slack.report_post",
        "Post a structured Slack report through the RSI native Slack gateway. Use this for rich final answers, tables, and artifact-backed report output. In RSI workflow threads, the platform supplies the bound channel and thread.",
        {
            "channel_id": {"type": "string"},
            "thread_ts": {"type": "string"},
            "report_schema_version": {"type": "integer", "enum": [1]},
            "summary": {"type": "string"},
            "sections": {
                "type": "array",
                "items": {
                    "type": "object",
                    "properties": {
                        "title": {"type": "string"},
                        "text": {"type": "string"},
                    },
                },
            },
            "tables": {"type": "array", "items": _SLACK_REPORT_TABLE_SCHEMA},
            "files": {"type": "array", "items": _SLACK_REPORT_FILE_SCHEMA},
            "images": {"type": "array", "items": _SLACK_REPORT_FILE_SCHEMA},
        },
        required=["report_schema_version", "summary"],
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
        "Upload a file to Slack through the RSI native Slack gateway. In RSI workflow threads, the platform supplies the bound channel and thread.",
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
        required=[],
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
    "rsi_notion.blocks_children": _schema("rsi_notion.blocks_children", "List Notion block children through the RSI native Notion gateway. Returns block summaries with plain_text, markdown, and typed block payloads for reading page content directly.", {"block_id": {"type": "string"}, "page_size": {"type": "integer"}, "cursor": {"type": "string"}}, ["block_id"]),
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
    "rsi_knowledge.document_get": _schema("rsi_knowledge.document_get", "Retrieve a mirrored company knowledge document. Use document_id only with mirrored knowledge/Honcho document IDs returned by rsi_knowledge.search. For raw Notion page/block IDs, pass source_ref when a mirror source ref is available, or use rsi_notion.page_get / rsi_notion.blocks_children for direct Notion reads.", {"source_ref": {"type": "string"}, "document_id": {"type": "string"}}),
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

_SENTRY_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "rsi_sentry.projects_list": _schema(
        "rsi_sentry.projects_list",
        "List Sentry projects in the server-configured RSI organization, optionally narrowed by project_ref.",
        {
            "project_ref": {"type": "string", "description": "Project slug such as depin-backend. Do not include an organization; RSI adds the configured org server-side."},
            "platform": {"type": "string"},
            "limit": {"type": "integer"},
            "cursor": {"type": "string"},
            "fresh": {"type": "boolean"},
        },
    ),
    "rsi_sentry.issues_list": _schema(
        "rsi_sentry.issues_list",
        "List Sentry issues in the server-configured RSI organization, optionally narrowed by project_ref.",
        {
            "project_ref": {"type": "string", "description": "Project slug such as depin-backend. Omit for org-wide issues."},
            "query": {"type": "string"},
            "limit": {"type": "integer"},
            "sort": {"type": "string", "enum": ["date", "new", "freq", "user"]},
            "period": {"type": "string", "description": "Time range such as 7d, 24h, or 2026-04-01..2026-05-01."},
            "cursor": {"type": "string"},
            "fresh": {"type": "boolean"},
        },
    ),
    "rsi_sentry.issue_view": _schema(
        "rsi_sentry.issue_view",
        "View one Sentry issue.",
        {"issue": {"type": "string"}, "spans": {"type": "string"}, "fresh": {"type": "boolean"}},
        ["issue"],
    ),
    "rsi_sentry.issue_events": _schema(
        "rsi_sentry.issue_events",
        "List events for one Sentry issue.",
        {
            "issue": {"type": "string"},
            "limit": {"type": "integer"},
            "query": {"type": "string"},
            "period": {"type": "string"},
            "cursor": {"type": "string"},
            "full": {"type": "boolean"},
            "fresh": {"type": "boolean"},
        },
        ["issue"],
    ),
    "rsi_sentry.releases_list": _schema(
        "rsi_sentry.releases_list",
        "List Sentry releases in the server-configured RSI organization, optionally narrowed by project_ref.",
        {"project_ref": {"type": "string"}, "limit": {"type": "integer"}, "cursor": {"type": "string"}, "fresh": {"type": "boolean"}},
    ),
}

_KANBAN_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "rsi_kanban.list_projects": _schema(
        "rsi_kanban.list_projects",
        "List RSI-native Kanban projects and optionally Slack project routes.",
        {
            "include_routes": {"type": "boolean"},
        },
    ),
    "rsi_kanban.create_project": _write_schema(
        "rsi_kanban.create_project",
        "Create an RSI-native Kanban project. Project creation also creates its default board.",
        {
            "slug": {"type": "string", "description": "Stable URL-safe project slug, for example numo or trace."},
            "name": {"type": "string", "description": "Human-readable project name."},
            "description": {"type": "string"},
            "summary": {"type": "string"},
            "metadata": _JSON_OBJECT_SCHEMA,
        },
        required=["name"],
    ),
    "rsi_kanban.list_project_routes": _schema(
        "rsi_kanban.list_project_routes",
        "List Slack route bindings for RSI-native Kanban projects.",
        {
            "project_id": {"type": "string"},
            "project_slug": {"type": "string"},
        },
    ),
    "rsi_kanban.set_project_slack_route": _write_schema(
        "rsi_kanban.set_project_slack_route",
        "Bind a Slack channel or thread to an RSI-native Kanban project for future project resolution.",
        {
            "project_id": {"type": "string"},
            "project_slug": {"type": "string"},
            "channel_id": {"type": "string"},
            "thread_ts": {"type": "string"},
            "team_id": {"type": "string"},
        },
        required=["channel_id"],
    ),
    "rsi_kanban.create_ticket": _write_schema(
        "rsi_kanban.create_ticket",
        "Create an internal RSI Kanban ticket. Include project_id or project_slug unless the Slack channel has an unambiguous project default.",
        {
            "project_id": {"type": "string"},
            "project_slug": {"type": "string"},
            "title": {"type": "string"},
            "description": {"type": "string"},
            "priority": {"type": "string"},
            "assignee": {"type": "string"},
            "channel_id": {"type": "string"},
            "thread_ts": {"type": "string"},
            "message_ts": {"type": "string"},
            "team_id": {"type": "string"},
            "permalink": {"type": "string"},
            "metadata": _JSON_OBJECT_SCHEMA,
        },
        required=["title"],
    ),
    "rsi_kanban.update_ticket": _write_schema(
        "rsi_kanban.update_ticket",
        "Update an internal RSI Kanban ticket status or fields.",
        {
            "ticket_id": {"type": "string"},
            "title": {"type": "string"},
            "description": {"type": "string"},
            "status": {"type": "string", "enum": ["triage", "todo", "in_progress", "blocked", "done", "archived"]},
            "priority": {"type": "string"},
            "assignee": {"type": "string"},
            "metadata": _JSON_OBJECT_SCHEMA,
        },
        required=["ticket_id"],
    ),
    "rsi_kanban.list_tickets": _schema(
        "rsi_kanban.list_tickets",
        "List internal RSI Kanban tickets for a project.",
        {
            "project_id": {"type": "string"},
            "project_slug": {"type": "string"},
            "channel_id": {"type": "string"},
            "thread_ts": {"type": "string"},
        },
    ),
    "rsi_kanban.comment_ticket": _write_schema(
        "rsi_kanban.comment_ticket",
        "Add a comment to an internal RSI Kanban ticket.",
        {"ticket_id": {"type": "string"}, "body": {"type": "string"}, "metadata": _JSON_OBJECT_SCHEMA},
        required=["ticket_id", "body"],
    ),
    "rsi_kanban.link_ticket": _write_schema(
        "rsi_kanban.link_ticket",
        "Create an informational link between two RSI Kanban tickets.",
        {
            "from_ticket_id": {"type": "string"},
            "ticket_id": {"type": "string"},
            "to_ticket_id": {"type": "string"},
            "link_type": {"type": "string"},
            "metadata": _JSON_OBJECT_SCHEMA,
        },
        required=["to_ticket_id"],
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

_TEMPORAL_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "rsi_temporal.list_schedules": _schema(
        "rsi_temporal.list_schedules",
        "List Temporal schedules for a configured target using Temporal visibility APIs.",
        {
            **_TEMPORAL_TARGET_PROPERTIES,
            "query": {"type": "string"},
            "limit": {"type": "integer"},
        },
        ["environment", "target"],
    ),
    "rsi_temporal.describe_schedule": _schema(
        "rsi_temporal.describe_schedule",
        "Describe one Temporal schedule for debugging or status checks.",
        {
            **_TEMPORAL_TARGET_PROPERTIES,
            "schedule_id": {"type": "string"},
        },
        ["environment", "target", "schedule_id"],
    ),
    "rsi_temporal.list_workflows": _schema(
        "rsi_temporal.list_workflows",
        "List Temporal workflow executions for a configured target using Temporal visibility query syntax.",
        {
            **_TEMPORAL_TARGET_PROPERTIES,
            "query": {"type": "string"},
            "limit": {"type": "integer"},
        },
        ["environment", "target"],
    ),
    "rsi_temporal.count_workflows": _schema(
        "rsi_temporal.count_workflows",
        "Count Temporal workflow executions for a configured target using Temporal visibility query syntax.",
        {
            **_TEMPORAL_TARGET_PROPERTIES,
            "query": {"type": "string"},
        },
        ["environment", "target"],
    ),
    "rsi_temporal.describe_workflow": _schema(
        "rsi_temporal.describe_workflow",
        "Describe one Temporal workflow execution for debugging or status checks.",
        {
            **_TEMPORAL_TARGET_PROPERTIES,
            "workflow_id": {"type": "string"},
            "run_id": {"type": "string"},
        },
        ["environment", "target", "workflow_id"],
    ),
    "rsi_temporal.pause_schedule": _write_schema(
        "rsi_temporal.pause_schedule",
        "Pause one allowlisted Temporal schedule. Requires confirm=true unless dry_run=true.",
        {
            **_TEMPORAL_MUTATION_PROPERTIES,
            "schedule_id": {"type": "string"},
        },
        required=["environment", "target", "schedule_id"],
    ),
    "rsi_temporal.unpause_schedule": _write_schema(
        "rsi_temporal.unpause_schedule",
        "Unpause one allowlisted Temporal schedule. Requires confirm=true unless dry_run=true.",
        {
            **_TEMPORAL_MUTATION_PROPERTIES,
            "schedule_id": {"type": "string"},
        },
        required=["environment", "target", "schedule_id"],
    ),
    "rsi_temporal.trigger_schedule": _write_schema(
        "rsi_temporal.trigger_schedule",
        "Start one action from an allowlisted Temporal schedule immediately. Requires confirm=true unless dry_run=true.",
        {
            **_TEMPORAL_MUTATION_PROPERTIES,
            "schedule_id": {"type": "string"},
        },
        required=["environment", "target", "schedule_id"],
    ),
    "rsi_temporal.start_workflow": _write_schema(
        "rsi_temporal.start_workflow",
        "Start one Temporal workflow on a configured Temporal target. Use this to restart a failed same-ID manager workflow; running IDs fail closed. Requires confirm=true unless dry_run=true.",
        {
            **_TEMPORAL_MUTATION_PROPERTIES,
            "workflow_id": {"type": "string"},
            "new_workflow_id": {"type": "string"},
            "workflow_type": {"type": "string"},
            "task_queue": {"type": "string"},
            "args": {"type": "array"},
        },
        required=["environment", "target", "workflow_type", "task_queue"],
    ),
    "rsi_temporal.stop_workflow": _write_schema(
        "rsi_temporal.stop_workflow",
        "Request graceful cancellation for one Temporal workflow on a configured Temporal target. This does not terminate or delete history.",
        {
            **_TEMPORAL_MUTATION_PROPERTIES,
            "workflow_id": {"type": "string"},
            "run_id": {"type": "string"},
        },
        required=["environment", "target", "workflow_id"],
    ),
    "rsi_temporal.restart_workflow": _write_schema(
        "rsi_temporal.restart_workflow",
        "Request graceful cancellation for one Temporal workflow and start a replacement workflow on a configured Temporal target.",
        {
            **_TEMPORAL_MUTATION_PROPERTIES,
            "workflow_id": {"type": "string"},
            "run_id": {"type": "string"},
            "new_workflow_id": {"type": "string"},
            "workflow_type": {"type": "string"},
            "task_queue": {"type": "string"},
            "args": {"type": "array"},
        },
        required=["environment", "target", "workflow_id", "new_workflow_id", "workflow_type", "task_queue"],
    ),
}

_AWS_TOOL_SCHEMAS: dict[str, JsonToolFunctionSchema] = {
    "rsi_aws.read": _schema(
        "rsi_aws.read",
        "Read AWS operational metadata through the RSI native AWS gateway. This accepts read-only operations for supported diagnostic services and blocks Secrets Manager, SSM parameter reads, KMS decrypt, S3 object reads, raw log reads, ECR authorization tokens, IAM, and mutations.",
        _AWS_READ_PROPERTIES,
        ["account", "service", "operation"],
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
        "Submit one read-only SQL query. This pauses Hermes at the tool call; after read-only validation (auto-approved, audited in Slack) and execution, RSI resumes this same session with a sanitized tool result.",
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
    **_SENTRY_TOOL_SCHEMAS,
    **_KANBAN_TOOL_SCHEMAS,
    **_TEMPORAL_TOOL_SCHEMAS,
    **_AWS_TOOL_SCHEMAS,
    **_OBSERVABILITY_TOOL_SCHEMAS,
}
_TOOLSET_SCHEMAS = {
    HERMES_ARTIFACT_TOOLSET: _ARTIFACT_TOOL_SCHEMAS,
    HERMES_DB_READ_TOOLSET: _DB_READ_TOOL_SCHEMAS,
    HERMES_RSI_SLACK_TOOLSET: _SLACK_TOOL_SCHEMAS,
    HERMES_RSI_NOTION_TOOLSET: _NOTION_TOOL_SCHEMAS,
    HERMES_RSI_KNOWLEDGE_TOOLSET: _KNOWLEDGE_TOOL_SCHEMAS,
    HERMES_RSI_SENTRY_TOOLSET: _SENTRY_TOOL_SCHEMAS,
    HERMES_RSI_KANBAN_TOOLSET: _KANBAN_TOOL_SCHEMAS,
    HERMES_RSI_TEMPORAL_TOOLSET: _TEMPORAL_TOOL_SCHEMAS,
    HERMES_RSI_AWS_TOOLSET: _AWS_TOOL_SCHEMAS,
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
