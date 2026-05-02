from __future__ import annotations

from unittest import mock
import unittest

from rsi_runner import slack_mcp_server


class SlackMCPServerTests(unittest.TestCase):
    def test_permalink_parser_converts_slack_url_to_channel_and_thread_ts(self) -> None:
        channel_id, thread_ts = slack_mcp_server._ts_from_permalink(
            "https://storyprotocol.slack.com/archives/C0AKH5SNGKH/p1777650186068179"
        )

        self.assertEqual(channel_id, "C0AKH5SNGKH")
        self.assertEqual(thread_ts, "1777650186.068179")

    def test_read_permalink_uses_thread_reader(self) -> None:
        with mock.patch.dict(
            "os.environ",
            {"RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH"},
            clear=False,
        ), mock.patch.object(
            slack_mcp_server,
            "slack_read_thread",
            return_value={"channel_id": "C0AKH5SNGKH", "thread_ts": "1777650186.068179", "messages": []},
        ) as read_thread:
            result = slack_mcp_server.slack_read_permalink(
                "https://storyprotocol.slack.com/archives/C0AKH5SNGKH/p1777650186068179",
                limit=50,
            )

        read_thread.assert_called_once_with(channel_id="C0AKH5SNGKH", thread_ts="1777650186.068179", limit=50)
        self.assertEqual(result["permalink"], "https://storyprotocol.slack.com/archives/C0AKH5SNGKH/p1777650186068179")

    def test_messages_read_refuses_unbounded_channel_reads(self) -> None:
        with mock.patch.dict("os.environ", {"RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH"}, clear=False):
            with self.assertRaisesRegex(RuntimeError, "unbounded channel history reads are refused"):
                slack_mcp_server.messages_read(channel_id="C0AKH5SNGKH")

    def test_conversation_get_reads_honcho_corpus_session(self) -> None:
        with mock.patch.object(slack_mcp_server, "_slack_workspace_id", return_value="T123"), mock.patch.object(
            slack_mcp_server,
            "_honcho_api",
            return_value={
                "items": [
                    {
                        "id": "msg_1",
                        "content": "CORS allowlist is configured in depin-backend.",
                        "metadata": {"slack_ts": "1777650186.068179"},
                    }
                ],
                "page": 1,
                "pages": 1,
                "total": 1,
            },
        ) as honcho_api, mock.patch.dict(
            "os.environ",
            {
                "RSI_HONCHO_BASE_URL": "http://honcho.test",
                "RSI_HONCHO_WORKSPACE_ID": "rsi_company_knowledge",
                "RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH",
            },
            clear=False,
        ):
            result = slack_mcp_server.conversation_get(
                channel_id="C0AKH5SNGKH",
                thread_ts="1777650186.068179",
                limit=25,
            )

        self.assertEqual(result["source"], "honcho_slack_corpus")
        self.assertEqual(result["source_session_key"], "slack:T123:C0AKH5SNGKH:1777650186.068179")
        self.assertEqual(result["messages"][0]["id"], "msg_1")
        honcho_api.assert_called_once()
        path = honcho_api.call_args.args[1]
        self.assertIn("/workspaces/rsi_company_knowledge/sessions/slack_T123_C0AKH5SNGKH_1777650186_068179/messages/list", path)

    def test_honcho_api_base_uses_v3_router(self) -> None:
        with mock.patch.dict("os.environ", {"RSI_HONCHO_BASE_URL": "http://honcho.test"}, clear=False):
            self.assertEqual(slack_mcp_server._honcho_api_base_url(), "http://honcho.test/v3")

        with mock.patch.dict("os.environ", {"RSI_HONCHO_BASE_URL": "http://honcho.test/v3"}, clear=False):
            self.assertEqual(slack_mcp_server._honcho_api_base_url(), "http://honcho.test/v3")

    def test_messages_read_filters_channel_window(self) -> None:
        with mock.patch.dict("os.environ", {"RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH"}, clear=False), mock.patch.object(slack_mcp_server, "conversation_get") as conversation_get:
            conversation_get.return_value = {
                "messages": [
                    {"id": "old", "metadata": {"slack_ts": "1777650100.000000"}},
                    {"id": "hit", "metadata": {"slack_ts": "1777650186.068179"}},
                    {"id": "new", "metadata": {"slack_ts": "1777650300.000000"}},
                ]
            }
            result = slack_mcp_server.messages_read(
                channel_id="C0AKH5SNGKH",
                oldest_ts="1777650180.000000",
                latest_ts="1777650200.000000",
            )

        self.assertEqual([item["id"] for item in result["messages"]], ["hit"])


if __name__ == "__main__":
    unittest.main()
