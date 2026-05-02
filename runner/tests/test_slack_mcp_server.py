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
        with mock.patch.object(
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


if __name__ == "__main__":
    unittest.main()
