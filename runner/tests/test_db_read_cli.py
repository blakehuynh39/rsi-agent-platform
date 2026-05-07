from __future__ import annotations

import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest import mock

from rsi_runner import db_read_cli


class DBReadCLITests(unittest.TestCase):
    def test_query_records_terminal_handoff_event(self) -> None:
        payload = {
            "status": "pending_approval",
            "message": "approval requested",
            "request": {
                "id": "dbread_1",
                "target": "depin-prod",
                "state": "pending_approval",
                "sql_sha256": "sha256:abc",
            },
            "validation": {"ok": True},
        }
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(
            os.environ,
            {
                "RSI_DB_READ_SUBMISSION_PATH": str(Path(tempdir, "dbread.json")),
                "RSI_TASK_REQUESTER": "user:U123",
                "RSI_CONVERSATION_ID": "conv-1",
                "RSI_WORKFLOW_ID": "wf-1",
                "RSI_TRACE_ID": "trace-1",
                "RSI_SLACK_CHANNEL_ID": "C123",
                "RSI_SLACK_THREAD_TS": "171000001.000100",
            },
            clear=True,
        ), mock.patch("rsi_runner.db_read_cli.request_json", return_value=payload) as request_json, mock.patch(
            "sys.stdout"
        ):
            code = db_read_cli.main(["query", "--target", "depin-prod", "--sql", "SELECT 1"])
            event = json.loads(Path(tempdir, "dbread.json").read_text(encoding="utf-8"))

        self.assertEqual(code, 0)
        request_json.assert_called_once()
        self.assertEqual(event["kind"], "db_read_request_submitted")
        self.assertEqual(event["request_id"], "dbread_1")
        self.assertEqual(event["target"], "depin-prod")
        self.assertEqual(event["state"], "pending_approval")
        self.assertEqual(event["sql_sha256"], "sha256:abc")

    def test_validation_failure_does_not_record_handoff_event(self) -> None:
        payload = {
            "status": "validation_failed",
            "request": {
                "id": "dbread_bad",
                "target": "depin-prod",
                "state": "validation_failed",
                "sql_sha256": "sha256:bad",
            },
            "validation": {"ok": False},
        }
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(
            os.environ,
            {"RSI_DB_READ_SUBMISSION_PATH": str(Path(tempdir, "dbread.json"))},
            clear=True,
        ):
            db_read_cli.record_db_read_submission(payload)
            self.assertFalse(Path(tempdir, "dbread.json").exists())


if __name__ == "__main__":
    unittest.main()
