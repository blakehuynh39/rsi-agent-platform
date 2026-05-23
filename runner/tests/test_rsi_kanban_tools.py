from __future__ import annotations

import unittest

from rsi_runner.config import filter_hermes_native_toolsets
from rsi_runner.rsi_tools import rsi_plugin_toolset_definitions, transport_tool_schema


class RSIKanbanToolsTest(unittest.TestCase):
    def test_rsi_kanban_toolset_is_exposed_without_local_hermes_kanban_tools(self) -> None:
        definitions = rsi_plugin_toolset_definitions()
        by_canonical = {item["canonical_name"]: item for item in definitions}

        self.assertIn("rsi_kanban.create_ticket", by_canonical)
        self.assertEqual(by_canonical["rsi_kanban.create_ticket"]["toolset"], "rsi-kanban")
        self.assertEqual(transport_tool_schema("rsi_kanban.create_ticket")["name"], "rsi_kanban_create_ticket")
        self.assertNotIn(
            "status",
            transport_tool_schema("rsi_kanban.create_ticket")["parameters"]["properties"],
        )

        local_hermes_kanban = [item for item in definitions if str(item["canonical_name"]).startswith("kanban_")]
        self.assertEqual(local_hermes_kanban, [])

    def test_local_hermes_kanban_toolset_is_filtered_from_config(self) -> None:
        self.assertEqual(
            filter_hermes_native_toolsets(["terminal", "kanban", "file"]),
            ["terminal", "file"],
        )


if __name__ == "__main__":
    unittest.main()
