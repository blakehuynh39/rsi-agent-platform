from __future__ import annotations

import tempfile
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

    def test_conversations_list_uses_honcho_sessions_and_allowlist(self) -> None:
        with mock.patch.object(
            slack_mcp_server,
            "_honcho_api",
            return_value={
                "items": [
                    {
                        "id": "slack_T123_C0AKH5SNGKH_1777650186_068179",
                        "metadata": {
                            "source": "slack",
                            "source_session_key": "slack:T123:C0AKH5SNGKH:1777650186.068179",
                        },
                        "created_at": "2026-05-01T17:00:00Z",
                    },
                    {
                        "id": "slack_T123_CPRIVATE_channel",
                        "metadata": {
                            "source": "slack",
                            "source_session_key": "slack:T123:CPRIVATE:channel",
                        },
                    },
                ],
                "page": 1,
                "pages": 1,
                "total": 2,
            },
        ) as honcho_api, mock.patch.dict(
            "os.environ",
            {
                "RSI_HONCHO_BASE_URL": "http://honcho.test",
                "RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH",
            },
            clear=False,
        ):
            result = slack_mcp_server.conversations_list(channel_id="C0AKH5SNGKH", limit=25)

        self.assertEqual([item["channel_id"] for item in result["conversations"]], ["C0AKH5SNGKH"])
        self.assertEqual(result["conversations"][0]["thread_ts"], "1777650186.068179")
        path = honcho_api.call_args.args[1]
        body = honcho_api.call_args.args[2]
        self.assertIn("/sessions/list?size=25&page=1", path)
        self.assertEqual(body["filters"]["AND"][0]["metadata"]["source"], "slack")
        self.assertEqual(body["filters"]["AND"][1]["metadata"]["source_session_key"]["icontains"], ":C0AKH5SNGKH:")

    def test_conversations_search_uses_honcho_not_slack_search(self) -> None:
        with mock.patch.object(
            slack_mcp_server,
            "_honcho_api_raw",
            return_value=[
                {
                    "id": "msg_1",
                    "content": "CORS allowlist lives in depin-backend.",
                    "metadata": {"source": "slack", "channel_id": "C0AKH5SNGKH"},
                },
                {
                    "id": "msg_private",
                    "content": "not visible",
                    "metadata": {"source": "slack", "channel_id": "CPRIVATE"},
                },
            ],
        ) as honcho_api, mock.patch.dict(
            "os.environ",
            {
                "RSI_HONCHO_BASE_URL": "http://honcho.test",
                "RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH",
            },
            clear=False,
        ):
            result = slack_mcp_server.conversations_search("CORS", limit=5)

        self.assertEqual([item["id"] for item in result["results"]], ["msg_1"])
        path = honcho_api.call_args.args[1]
        body = honcho_api.call_args.args[2]
        self.assertEqual(path, "/workspaces/rsi_company_knowledge/search")
        self.assertEqual(body["filters"]["metadata"]["channel_id"]["in"], ["C0AKH5SNGKH"])

    def test_normalize_messages_preserves_slack_file_metadata(self) -> None:
        messages = slack_mcp_server._normalize_messages(
            [
                {
                    "text": "see screenshot",
                    "ts": "1777650186.068179",
                    "files": [
                        {
                            "id": "F123",
                            "name": "screenshot.png",
                            "mimetype": "image/png",
                            "filetype": "png",
                            "size": 1234,
                            "permalink": "https://storyprotocol.slack.com/files/F123",
                        }
                    ],
                }
            ]
        )

        self.assertEqual(messages[0]["files"][0]["id"], "F123")
        self.assertEqual(messages[0]["files"][0]["mimetype"], "image/png")

    def test_attachments_fetch_returns_metadata_only_from_honcho(self) -> None:
        with mock.patch.object(
            slack_mcp_server,
            "conversation_get",
            return_value={
                "messages": [
                    {
                        "id": "msg_1",
                        "content": "see screenshot",
                        "metadata": {
                            "channel_id": "C0AKH5SNGKH",
                            "thread_ts": "1777650186.068179",
                            "slack_ts": "1777650187.000000",
                            "permalink": "https://storyprotocol.slack.com/archives/C0AKH5SNGKH/p1777650187000000",
                            "files": [
                                {
                                    "id": "F123",
                                    "name": "screenshot.png",
                                    "mimetype": "image/png",
                                    "filetype": "png",
                                    "size": 1234,
                                    "permalink": "https://storyprotocol.slack.com/files/F123",
                                }
                            ],
                        },
                    }
                ]
            },
        ), mock.patch.dict(
            "os.environ",
            {"RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH"},
            clear=False,
        ):
            result = slack_mcp_server.attachments_fetch(
                channel_id="C0AKH5SNGKH",
                thread_ts="1777650186.068179",
                message_ts="1777650187.000000",
            )

        self.assertEqual(result["attachment_count"], 1)
        self.assertEqual(result["attachments"][0]["file"]["id"], "F123")
        self.assertEqual(result["attachments"][0]["content_status"], "metadata_only")

    def test_attachments_fetch_extracts_text_and_persists_idempotently(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir, mock.patch.object(
            slack_mcp_server,
            "conversation_get",
            return_value={
                "messages": [
                    {
                        "id": "msg_1",
                        "metadata": {
                            "workspace_id": "T123",
                            "channel_id": "C0AKH5SNGKH",
                            "thread_ts": "1777650186.068179",
                            "slack_ts": "1777650187.000000",
                            "permalink": "https://storyprotocol.slack.com/archives/C0AKH5SNGKH/p1777650187000000",
                            "files": [{"id": "F123", "name": "notes.txt", "mimetype": "text/plain", "filetype": "txt"}],
                        },
                    }
                ]
            },
        ), mock.patch.object(
            slack_mcp_server,
            "_download_slack_file",
            return_value=(b"hello from a cached attachment\n", {"id": "F123", "name": "notes.txt", "mimetype": "text/plain", "filetype": "txt"}),
        ), mock.patch.object(
            slack_mcp_server,
            "_source_mirror_write_message",
            return_value={"record": {"status": "complete"}, "honcho_message_id": "msg_analysis_1", "should_write": True},
        ) as persist, mock.patch.dict(
            "os.environ",
            {
                "RSI_ATTACHMENT_CACHE_ROOT": tmpdir,
                "RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH",
            },
            clear=False,
        ):
            result = slack_mcp_server.attachments_fetch(
                channel_id="C0AKH5SNGKH",
                thread_ts="1777650186.068179",
                message_ts="1777650187.000000",
                include_content=True,
            )

        attachment = result["attachments"][0]
        self.assertEqual(attachment["content_status"], "cached")
        self.assertEqual(attachment["extraction_status"], "extracted")
        self.assertIn("hello from a cached attachment", attachment["extracted_text"])
        self.assertNotIn("Call attachments_fetch with include_content=true", attachment["extraction_note"])
        self.assertTrue(attachment["cache_path"].startswith(tmpdir))
        record = persist.call_args.args[0]
        message = persist.call_args.args[1]
        self.assertEqual(record["source_type"], "slack_attachment_analysis")
        self.assertEqual(record["source_key"], "slack_attachment_analysis:T123:C0AKH5SNGKH:1777650187.000000:F123:text")
        self.assertEqual(message["peer_id"], "rsi_attachment_analyzer")
        self.assertEqual(message["metadata"]["extraction_status"], "extracted")

    def test_attachments_fetch_analyzes_image_with_configured_vision_model(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir, mock.patch.object(
            slack_mcp_server,
            "conversation_get",
            return_value={
                "messages": [
                    {
                        "id": "msg_1",
                        "metadata": {
                            "workspace_id": "T123",
                            "channel_id": "C0AKH5SNGKH",
                            "thread_ts": "1777650186.068179",
                            "slack_ts": "1777650187.000000",
                            "files": [{"id": "FIMG", "name": "screen.png", "mimetype": "image/png", "filetype": "png"}],
                        },
                    }
                ]
            },
        ), mock.patch.object(
            slack_mcp_server,
            "_download_slack_file",
            return_value=(b"\x89PNG\r\n", {"id": "FIMG", "name": "screen.png", "mimetype": "image/png", "filetype": "png"}),
        ), mock.patch.object(
            slack_mcp_server,
            "_vision_analyze_image",
            return_value={"model": "qwen/qwen3.6-flash", "text": "Screenshot says CORS error."},
        ) as vision, mock.patch.object(
            slack_mcp_server,
            "_source_mirror_write_message",
            return_value={"record": {"status": "complete"}, "honcho_message_id": "msg_analysis_2", "should_write": True},
        ) as persist, mock.patch.dict(
            "os.environ",
            {
                "RSI_ATTACHMENT_CACHE_ROOT": tmpdir,
                "RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH",
            },
            clear=False,
        ):
            result = slack_mcp_server.attachments_fetch(
                channel_id="C0AKH5SNGKH",
                thread_ts="1777650186.068179",
                message_ts="1777650187.000000",
                include_content=True,
                analyze_images=True,
                analysis_prompt="read the UI",
            )

        attachment = result["attachments"][0]
        self.assertEqual(attachment["extraction_status"], "vision_analyzed")
        self.assertEqual(attachment["vision_model"], "qwen/qwen3.6-flash")
        self.assertEqual(attachment["extracted_text"], "Screenshot says CORS error.")
        self.assertIn("auxiliary vision model", attachment["extraction_note"])
        vision.assert_called_once()
        record = persist.call_args.args[0]
        self.assertEqual(record["source_key"], "slack_attachment_analysis:T123:C0AKH5SNGKH:1777650187.000000:FIMG:vision")
        self.assertIn("model:qwen/qwen3.6-flash", record["source_revision"])

    def test_attachments_fetch_records_unsupported_binary_without_fabricating_content(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir, mock.patch.object(
            slack_mcp_server,
            "conversation_get",
            return_value={
                "messages": [
                    {
                        "id": "msg_1",
                        "metadata": {
                            "workspace_id": "T123",
                            "channel_id": "C0AKH5SNGKH",
                            "thread_ts": "1777650186.068179",
                            "slack_ts": "1777650187.000000",
                            "files": [{"id": "FPDF", "name": "doc.pdf", "mimetype": "application/pdf", "filetype": "pdf"}],
                        },
                    }
                ]
            },
        ), mock.patch.object(
            slack_mcp_server,
            "_download_slack_file",
            return_value=(b"%PDF", {"id": "FPDF", "name": "doc.pdf", "mimetype": "application/pdf", "filetype": "pdf"}),
        ), mock.patch.object(
            slack_mcp_server,
            "_source_mirror_write_message",
            return_value={"record": {"status": "complete"}, "honcho_message_id": "msg_analysis_3", "should_write": True},
        ) as persist, mock.patch.dict(
            "os.environ",
            {
                "RSI_ATTACHMENT_CACHE_ROOT": tmpdir,
                "RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST": "C0AKH5SNGKH",
            },
            clear=False,
        ):
            result = slack_mcp_server.attachments_fetch(
                channel_id="C0AKH5SNGKH",
                thread_ts="1777650186.068179",
                message_ts="1777650187.000000",
                include_content=True,
            )

        attachment = result["attachments"][0]
        self.assertEqual(attachment["extraction_status"], "unsupported_binary")
        self.assertNotIn("extracted_text", attachment)
        self.assertIn("Unsupported", attachment["extraction_error"])
        self.assertIn("no supported extractor", attachment["extraction_note"])
        message = persist.call_args.args[1]
        self.assertEqual(message["metadata"]["extraction_status"], "unsupported_binary")

    def test_honcho_api_base_uses_v3_router(self) -> None:
        with mock.patch.dict("os.environ", {"RSI_HONCHO_BASE_URL": "http://honcho.test"}, clear=False):
            self.assertEqual(slack_mcp_server._honcho_api_base_url(), "http://honcho.test/v3")

        with mock.patch.dict("os.environ", {"RSI_HONCHO_BASE_URL": "http://honcho.test/v3"}, clear=False):
            self.assertEqual(slack_mcp_server._honcho_api_base_url(), "http://honcho.test/v3")

    def test_control_plane_base_url_requires_explicit_config(self) -> None:
        with mock.patch.dict(
            "os.environ",
            {
                "RSI_CONTROL_PLANE_BASE_URL": "",
                "USE1_STAGE_RSI_AGENT_PLATFORM_CONTROL_PLANE_SERVICE_HOST": "172.20.190.168",
                "USE1_STAGE_RSI_AGENT_PLATFORM_CONTROL_PLANE_SERVICE_PORT": "8080",
            },
            clear=False,
        ):
            self.assertEqual(slack_mcp_server._control_plane_base_url(), "")

        with mock.patch.dict("os.environ", {"RSI_CONTROL_PLANE_BASE_URL": "http://control-plane:8080/"}, clear=False):
            self.assertEqual(slack_mcp_server._control_plane_base_url(), "http://control-plane:8080")

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
