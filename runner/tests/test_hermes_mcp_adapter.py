from __future__ import annotations

import os
import types
import unittest
from unittest import mock

from rsi_runner.hermes_mcp_adapter import HermesTaskScopedMCPAdapter, TaskScopedMCPRegistration


class FakeServerTask:
    def __init__(self, name: str, shutdown_log: list[str], discovered_tool_names: list[str]) -> None:
        self.name = name
        self._shutdown_log = shutdown_log
        self._tools = [types.SimpleNamespace(name=tool_name) for tool_name in discovered_tool_names]

    def shutdown(self) -> dict[str, object]:
        self._shutdown_log.append(self.name)
        return {"closed": True}


class FakeMCPToolModule:
    def __init__(self, *, mcp_available: bool = True, discovered_tool_names: list[str] | None = None) -> None:
        self._MCP_AVAILABLE = mcp_available
        self._servers: dict[str, FakeServerTask] = {}
        self.shutdown_log: list[str] = []
        self.discovered_tool_names = discovered_tool_names

    def register_mcp_servers(self, configs: dict[str, dict[str, object]]) -> list[str]:
        for name, config in configs.items():
            tools_config = config.get("tools") if isinstance(config, dict) else {}
            include = []
            if isinstance(tools_config, dict):
                include = [str(item) for item in tools_config.get("include") or []]
            discovered_tool_names = list(
                self.discovered_tool_names if self.discovered_tool_names is not None else include
            )
            self._servers[name] = FakeServerTask(name, self.shutdown_log, discovered_tool_names)
        return list(configs.keys())

    def _run_on_mcp_loop(self, value):
        return value


class HermesTaskScopedMCPAdapterTests(unittest.TestCase):
    def test_slack_reply_profile_translation_builds_hermes_headers_and_tool_include(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-123",
            workflow_id="wf-123",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "slack",
                    "profile": "slack_mcp_reply",
                    "headers": {"X-Test": "static"},
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter(
            default_slack_server_url="https://mcp.slack.com/mcp",
            slack_read_tool_names_resolver=lambda: ["get_thread", "search_messages"],
            slack_send_tool_name_resolver=lambda: "send_message",
        )

        with mock.patch.dict(os.environ, {"RSI_SLACK_USER_TOKEN": "slack-token"}, clear=True):
            translated = adapter._translate_task_servers(task)

        self.assertEqual(len(translated), 1)
        server = translated[0]
        self.assertEqual(server.hermes_config["url"], "https://mcp.slack.com/mcp")
        self.assertEqual(
            server.hermes_config["headers"],
            {
                "X-Test": "static",
                "Authorization": "Bearer slack-token",
            },
        )
        self.assertEqual(
            server.hermes_config["tools"]["include"],
            ["get_thread", "search_messages", "send_message"],
        )
        self.assertTrue(server.toolset_alias.startswith("mcp-rsi-task-trace-123-0-slack-"))

    def test_notion_read_profile_defaults_to_notion_api_read_tools(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-456",
            workflow_id="wf-456",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "notion",
                    "profile": "notion_mcp_read",
                    "server_url": "https://mcp.notion.com/mcp",
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()

        translated = adapter._translate_task_servers(task)

        self.assertEqual(
            translated[0].hermes_config["tools"],
            {
                "include": ["API-post-search", "API-retrieve-a-page", "API-get-block-children"],
                "resources": False,
                "prompts": False,
            },
        )

    def test_custom_server_authorization_env_var_populates_bearer_header(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-auth",
            workflow_id="wf-auth",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "notion",
                    "server_url": "https://mcp.notion.com/mcp",
                    "authorization_env_var": "RSI_NOTION_MCP_AUTHORIZATION",
                    "allowed_tools": {"tool_names": ["search", "fetch"]},
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()

        with mock.patch.dict(os.environ, {"RSI_NOTION_MCP_AUTHORIZATION": "notion-oauth-token"}, clear=True):
            translated = adapter._translate_task_servers(task)

        self.assertEqual(
            translated[0].hermes_config["headers"],
            {"Authorization": "Bearer notion-oauth-token"},
        )
        self.assertEqual(translated[0].hermes_config["tools"]["include"], ["search", "fetch"])

    def test_custom_server_header_env_vars_merge_with_static_headers(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-header",
            workflow_id="wf-header",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "notion",
                    "server_url": "https://mcp.notion.com/mcp",
                    "header_env_vars": {
                        "CF-Access-Client-Id": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID",
                        "CF-Access-Client-Secret": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_SECRET",
                    },
                    "headers": {"X-Test": "static"},
                    "allowed_tools": {"tool_names": ["search"]},
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()

        with mock.patch.dict(
            os.environ,
            {
                "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID": "client-id",
                "RSI_NOTION_MCP_CF_ACCESS_CLIENT_SECRET": "client-secret",
            },
            clear=True,
        ):
            translated = adapter._translate_task_servers(task)

        self.assertEqual(
            translated[0].hermes_config["headers"],
            {
                "X-Test": "static",
                "CF-Access-Client-Id": "client-id",
                "CF-Access-Client-Secret": "client-secret",
            },
        )

    def test_custom_server_missing_authorization_env_var_fails(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-auth-missing",
            workflow_id="wf-auth-missing",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "notion",
                    "server_url": "https://mcp.notion.com/mcp",
                    "authorization_env_var": "RSI_NOTION_MCP_AUTHORIZATION",
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()

        with mock.patch.dict(os.environ, {}, clear=True):
            with self.assertRaisesRegex(RuntimeError, "RSI_NOTION_MCP_AUTHORIZATION"):
                adapter._translate_task_servers(task)

    def test_custom_server_missing_header_env_var_fails(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-header-missing",
            workflow_id="wf-header-missing",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "notion",
                    "server_url": "https://mcp.notion.com/mcp",
                    "header_env_vars": {
                        "CF-Access-Client-Id": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID",
                    },
                    "allowed_tools": {"tool_names": ["search"]},
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()

        with mock.patch.dict(os.environ, {}, clear=True):
            with self.assertRaisesRegex(RuntimeError, "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID"):
                adapter._translate_task_servers(task)

    def test_custom_read_only_server_without_explicit_tools_fails_closed(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-789",
            workflow_id="wf-789",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "docs",
                    "server_url": "https://docs.example.com/mcp",
                    "allowed_tools": {"read_only": True},
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()

        with self.assertRaisesRegex(RuntimeError, "refusing to expose the full server"):
            adapter._translate_task_servers(task)

    def test_register_and_cleanup_manage_task_scoped_server_lifecycle(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-999",
            workflow_id="wf-999",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "notion",
                    "profile": "notion_mcp_read",
                    "server_url": "https://mcp.notion.com/mcp",
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()
        fake_mcp_tool = FakeMCPToolModule()

        with mock.patch.object(adapter, "_load_hermes_mcp_tool", return_value=fake_mcp_tool):
            registration = adapter.register_task_servers(task)
            cleanup = adapter.cleanup_registration(registration)

        self.assertTrue(registration.enabled)
        self.assertEqual(registration.enabled_toolsets, [f"mcp-{registration.server_names[0]}"])
        self.assertEqual(cleanup.status, "cleaned")
        self.assertEqual(cleanup.cleaned_server_names, registration.server_names)
        self.assertEqual(fake_mcp_tool.shutdown_log, registration.server_names)
        self.assertEqual(fake_mcp_tool._servers, {})

    def test_register_task_servers_fails_clearly_when_mcp_sdk_extra_is_unavailable(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-no-sdk",
            workflow_id="wf-no-sdk",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "notion",
                    "profile": "notion_mcp_read",
                    "server_url": "https://mcp.notion.com/mcp",
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()
        fake_mcp_tool = FakeMCPToolModule(mcp_available=False)

        with mock.patch.object(adapter, "_load_hermes_mcp_tool", return_value=fake_mcp_tool):
            with self.assertRaisesRegex(RuntimeError, "MCP SDK extra is unavailable"):
                adapter.register_task_servers(task)

    def test_register_task_servers_fails_when_included_tools_are_not_exposed(self) -> None:
        task = types.SimpleNamespace(
            trace_id="trace-missing-tool",
            workflow_id="wf-missing-tool",
            conversation_id="",
            session_scope_id="",
            task_type="workflow",
            mcp_servers=[
                {
                    "server_label": "notion",
                    "profile": "notion_mcp_read",
                    "server_url": "https://mcp.notion.com/mcp",
                }
            ],
        )
        adapter = HermesTaskScopedMCPAdapter()
        fake_mcp_tool = FakeMCPToolModule(discovered_tool_names=["API-post-search"])

        with mock.patch.object(adapter, "_load_hermes_mcp_tool", return_value=fake_mcp_tool):
            with self.assertRaisesRegex(RuntimeError, "did not expose included tool"):
                adapter.register_task_servers(task)

    def test_cleanup_marks_missing_server_as_failed(self) -> None:
        registration = TaskScopedMCPRegistration(
            servers=[
                types.SimpleNamespace(
                    source_label="notion",
                    profile="notion_mcp_read",
                    server_name="rsi-task-missing",
                    toolset_alias="mcp-rsi-task-missing",
                    included_tool_names=["search"],
                    hermes_config={},
                )
            ]
        )
        adapter = HermesTaskScopedMCPAdapter()
        fake_mcp_tool = FakeMCPToolModule()

        with mock.patch.object(adapter, "_load_hermes_mcp_tool", return_value=fake_mcp_tool):
            cleanup = adapter.cleanup_registration(registration)

        self.assertEqual(cleanup.status, "cleanup_failed")
        self.assertEqual(cleanup.cleaned_server_names, [])
        self.assertEqual(cleanup.failed_server_names, ["rsi-task-missing"])
        self.assertIn("server not registered", cleanup.errors[0])


if __name__ == "__main__":
    unittest.main()
