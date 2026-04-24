from __future__ import annotations

import base64
import io
import json
import os
from pathlib import Path
import subprocess
import tempfile
import threading
import time
import types
import unittest
from unittest import mock

from rsi_runner.config import RunnerConfig, RunnerConfigError
from rsi_runner.hermes_adapter import _build_plugin_module
from rsi_runner.hermes_executor_worker import _LocalArtifactToolBinding, _initialize_cli_agent
from rsi_runner.hermes_mcp_adapter import TaskScopedMCPCleanupResult, TaskScopedMCPRegistration, TaskScopedMCPServer
from rsi_runner.hermes_runtime import HermesExecutionResult, HermesRuntime, RunnerTaskRequest
from rsi_runner.observability import execution_observation_id
from rsi_runner.rsi_tools import ReadOnlyToolBinding, governed_toolset_definitions, tool_schema_wrappers, transport_tool_schema
from rsi_runner.session_manager import SessionManager


def runner_env(role: str = "prod") -> dict[str, str]:
    return {
        "RSI_RUNNER_ROLE": role,
        "RSI_RUNNER_HOST": "0.0.0.0",
        "RSI_RUNNER_PORT": "8090",
        "RSI_RUNNER_MODEL": "openai/gpt-5.4",
        "RSI_RUNNER_REASONING_EFFORT": "xhigh",
        "RSI_HERMES_PIN": "6fdbf2f2d76cf37393e657bf37ceda3d84589200",
        "RSI_RUNNER_PUBLIC_BASE_URL": "https://staging-rsi-platform.storyprotocol.net",
        "RSI_TOOL_GATEWAY_BASE_URL": "http://tool-gateway.internal:8080",
        "HERMES_HOME": "/tmp/hermes",
        "RSI_RUNNER_MEMORY_BACKEND": "honcho",
        "RSI_HONCHO_WORKSPACE": "rsi-stage",
        "RSI_HONCHO_RECALL_MODE": "hybrid",
        "RSI_HONCHO_WRITE_FREQUENCY": "async",
        "RSI_HONCHO_SESSION_STRATEGY": "hybrid",
        "RSI_HONCHO_AI_PEER": f"rsi:stage:{role}",
        "RSI_HONCHO_ENVIRONMENT": "stage",
        "RSI_RUNNER_EVAL_MAX_ITERATIONS": "5",
        "RSI_RUNNER_PROPOSAL_MAX_ITERATIONS": "5",
        "RSI_RUNNER_PROD_MAX_ITERATIONS": "20",
        "RSI_RUNNER_PROACTIVE_MAX_ITERATIONS": "20",
        "RSI_RUNNER_EVAL_TASK_TIMEOUT": "300s",
        "RSI_RUNNER_PROPOSAL_TASK_TIMEOUT": "420s",
        "RSI_RUNNER_PROD_TASK_TIMEOUT": "900s",
        "RSI_RUNNER_EVAL_INACTIVITY_TIMEOUT": "240s",
        "RSI_RUNNER_PROPOSAL_INACTIVITY_TIMEOUT": "360s",
        "RSI_RUNNER_PROD_TIMEOUT": "930s",
        "RSI_RUNNER_PROACTIVE_TIMEOUT": "60s",
        "RSI_RUNNER_EVAL_TIMEOUT": "330s",
        "RSI_RUNNER_PROPOSAL_TIMEOUT": "450s",
        "RSI_RUNNER_NATIVE_MAX_OUTPUT_TOKENS": "15000",
        "HONCHO_API_KEY": "honcho-test-key",
        "OPENAI_API_KEY": "openai-test-key",
    }


class FakeAIAgent:
    last_kwargs: dict[str, object] | None = None
    last_prompt: str | None = None
    last_system_message: str | None = None
    last_history: list[dict[str, object]] | None = None
    last_valid_tool_names: list[str] | None = None
    last_tool_names: list[str] | None = None
    last_interrupt_message: str | None = None
    sleep_seconds: float = 0.0
    budget_used: int = 1

    def __init__(self, **kwargs) -> None:
        type(self).last_kwargs = kwargs
        self._interrupted = False

    def run_conversation(
        self,
        prompt: str,
        system_message: str | None = None,
        conversation_history: list[dict] | None = None,
        task_id: str | None = None,
    ) -> dict[str, object]:
        if type(self).sleep_seconds > 0:
            time.sleep(type(self).sleep_seconds)
        type(self).last_prompt = prompt
        type(self).last_system_message = system_message
        type(self).last_history = conversation_history or []
        type(self).last_valid_tool_names = sorted(getattr(self, "valid_tool_names", []))
        type(self).last_tool_names = sorted(
            tool["function"]["name"]
            for tool in list(getattr(self, "tools", []) or [])
            if isinstance(tool, dict) and isinstance(tool.get("function"), dict) and tool["function"].get("name")
        )
        return {
            "final_response": json.dumps(
                {
                    "visible_reasoning": [
                        {
                            "step_type": "analysis",
                            "summary": "Collected context and prepared a reply.",
                            "alternatives": [],
                            "confidence": 0.91,
                            "decision": "reply_in_thread",
                        }
                    ],
                    "reply_draft": "Draft reply",
                    "final_answer": "Final reply",
                    "confidence": 0.91,
                    "context_summary": "Repo and KB context collected.",
                    "self_critique": "Follow up if channel policy changes.",
                    "proposed_actions": [],
                    "knowledge_drafts": [],
                    "outcome_hypotheses": [],
                }
            )
        }

    def interrupt(self, _message: str | None = None) -> None:
        self._interrupted = True
        type(self).last_interrupt_message = _message

    def get_activity_summary(self) -> dict[str, object]:
        return {
            "last_activity_desc": "waiting_on_tool_or_model",
            "current_tool": "rsi.trace_context",
            "api_call_count": 2,
            "budget_used": type(self).budget_used,
            "budget_max": type(self).last_kwargs.get("max_iterations", 1) if type(self).last_kwargs else 1,
        }


class FakeTracker:
    def __init__(self) -> None:
        self.reads = [{"kind": "session_history", "summary": "user: prior message"}]
        self.writes = [{"kind": "memory_sync_assistant", "summary": "assistant: reply"}]


class FakeSessionManager:
    def __init__(self, _config) -> None:
        self.ready_issues: list[str] = []
        self.available = True
        self.hermes_home = "/var/lib/hermes"
        self.session_db_path = "/var/lib/hermes/state.db"
        self.skills_dir = "/var/lib/hermes/skills"
        self.bundled_skills_available = True
        self.bundled_skills_sync_status = "synced"
        self.bundled_skills_sync_error = ""
        self.hermes_config_parity_status = "configured"
        self.hermes_config_parity_error = ""
        self.skills_healthy = True
        self.session_db = object()
        self.honcho_available = True

    def prepare(self, task: RunnerTaskRequest):
        return types.SimpleNamespace(
            session_id="rsi-prod-conversation-123",
            parent_session_id="",
            scope_kind=task.session_scope_kind or "conversation",
            scope_id=task.session_scope_id or "conv-001",
            parent_scope_kind=task.parent_session_scope_kind or "",
            parent_scope_id=task.parent_session_scope_id or "",
            memory_backend=task.memory_backend or "honcho",
            assistant_peer_id=task.assistant_peer_id or "rsi:stage:prod",
            user_peer_id=task.user_peer_id or "slack:U123",
            hermes_home=self.hermes_home,
            session_db_path=self.session_db_path,
            conversation_history=[{"role": "user", "content": "Earlier thread message"}],
        )

    def attach_tracking(self, _agent: object, _task: RunnerTaskRequest, _context: object) -> FakeTracker:
        return FakeTracker()

    def finalize(self, context: types.SimpleNamespace, tracker: FakeTracker) -> dict[str, object]:
        return {
            "hermes_session_id": context.session_id,
            "parent_session_id": context.parent_session_id,
            "session_scope_kind": context.scope_kind,
            "session_scope_id": context.scope_id,
            "parent_session_scope_kind": context.parent_scope_kind,
            "parent_session_scope_id": context.parent_scope_id,
            "memory_backend": context.memory_backend,
            "assistant_peer_id": context.assistant_peer_id,
            "user_peer_id": context.user_peer_id,
            "hermes_home": context.hermes_home,
            "session_db_path": context.session_db_path,
            "memory_reads": tracker.reads,
            "memory_writes": tracker.writes,
            "session_messages_delta": [],
        }


class RecordingObserver:
    def __init__(self) -> None:
        self.execution_id = "hexec-test"
        self.events: list[dict[str, object]] = []

    def emit(self, *, phase: str, event_type: str, status: str = "", payload: dict[str, object] | None = None) -> None:
        self.events.append(
            {
                "phase": phase,
                "event_type": event_type,
                "status": status,
                "payload": payload or {},
            }
        )

    def diagnostics(self) -> dict[str, object]:
        return {
            "observation_execution_id": self.execution_id,
            "observation_sink_status": "ok",
            "observation_seq": len(self.events),
        }


class FakeHTTPResponse:
    def __init__(self, payload: dict[str, object]) -> None:
        self._body = json.dumps(payload).encode("utf-8")

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb) -> bool:
        return False

    def read(self) -> bytes:
        return self._body


def partial_structured_output(
    *,
    reply_text: str,
    proposed_actions: list[dict[str, object]] | None = None,
    context_summary: str = "Partial answer grounded in bounded-stop evidence.",
) -> dict[str, object]:
    return {
        "visible_reasoning": [
            {
                "step_type": "analysis",
                "summary": "Condensed the captured evidence into a grounded partial answer.",
                "alternatives": [],
                "confidence": 0.71,
                "decision": "post_partial_reply",
            }
        ],
        "reply_draft": reply_text,
        "final_answer": reply_text,
        "confidence": 0.71,
        "context_summary": context_summary,
        "self_critique": "More time or reads could improve coverage.",
        "proposed_actions": proposed_actions or [],
        "knowledge_drafts": [],
        "outcome_hypotheses": [],
        "change_plan": "",
        "repo_patch": "",
        "validation_plan": "",
        "retry_assessment": {},
        "hypothesis_delta": "",
    }


class HermesRuntimeTests(unittest.TestCase):
    def setUp(self) -> None:
        FakeAIAgent.last_kwargs = None
        FakeAIAgent.last_prompt = None
        FakeAIAgent.last_system_message = None
        FakeAIAgent.last_history = None
        FakeAIAgent.last_valid_tool_names = None
        FakeAIAgent.last_tool_names = None
        FakeAIAgent.last_interrupt_message = None
        FakeAIAgent.sleep_seconds = 0.0
        FakeAIAgent.budget_used = 1

    def test_config_requires_explicit_env(self) -> None:
        with mock.patch.dict(os.environ, {}, clear=True):
            with self.assertRaises(RunnerConfigError):
                RunnerConfig.from_env()

    def test_config_reads_explicit_gpt54_xhigh_and_honcho(self) -> None:
        with mock.patch.dict(os.environ, runner_env("eval"), clear=True):
            config = RunnerConfig.from_env()

        self.assertEqual(config.model, "openai/gpt-5.4")
        self.assertEqual(config.reasoning_effort, "xhigh")
        self.assertEqual(config.memory_backend, "honcho")
        self.assertEqual(config.honcho_workspace, "rsi-stage")
        self.assertEqual(config.honcho_environment_effective, "production")
        self.assertEqual(config.native_max_output_tokens, 15000)

    def test_config_reads_verbose_trace_logging(self) -> None:
        with mock.patch.dict(
            os.environ,
            {**runner_env("eval"), "RSI_VERBOSE_TRACE_LOGGING": "true", "RSI_VERBOSE_TRACE_LOG_LIMIT": "2048"},
            clear=True,
        ):
            config = RunnerConfig.from_env()

        self.assertTrue(config.verbose_trace_logging)
        self.assertEqual(config.verbose_trace_log_limit, 2048)

    def test_config_reads_native_executor_settings(self) -> None:
        with mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_SERVICE_ONLY": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": "/var/lib/hermes-executor",
            },
            clear=True,
        ):
            config = RunnerConfig.from_env()

        self.assertTrue(config.hermes_executor_enabled)
        self.assertTrue(config.hermes_executor_service_only)
        self.assertEqual(config.hermes_executor_workspace_root, "/var/lib/hermes-executor")
        self.assertEqual(config.hermes_computer_root, "/var/lib/hermes-executor/company")
        self.assertEqual(config.hermes_run_root, "/var/lib/hermes-executor/company/.rsi/runs")
        self.assertEqual(config.hermes_artifact_root, "/var/lib/hermes-executor/company/artifacts")

    def test_context_engine_plugin_module_compiles_as_python(self) -> None:
        source = _build_plugin_module()

        compile(source, "rsi_context_engine/__init__.py", "exec")
        self.assertNotIn(": false", source)
        self.assertIn(": False", source)
        self.assertIn("tool_call_started", source)
        self.assertIn("tool_call_completed", source)

    def test_config_rejects_timeout_contract_drift(self) -> None:
        with mock.patch.dict(
            os.environ,
            {**runner_env("eval"), "RSI_RUNNER_EVAL_TASK_TIMEOUT": "328s", "RSI_RUNNER_EVAL_TIMEOUT": "330s"},
            clear=True,
        ):
            with self.assertRaises(RunnerConfigError):
                RunnerConfig.from_env()

    def test_config_requires_explicit_native_output_token_budget(self) -> None:
        env = runner_env("eval")
        env.pop("RSI_RUNNER_NATIVE_MAX_OUTPUT_TOKENS", None)
        with mock.patch.dict(os.environ, env, clear=True):
            with self.assertRaises(RunnerConfigError):
                RunnerConfig.from_env()

    def test_session_manager_writes_honcho_safe_environment(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch(
            "rsi_runner.session_manager.sync_skills"
        ) as sync_mock, mock.patch.dict(os.environ, {**runner_env("eval"), "HERMES_HOME": tempdir}, clear=True):
            sync_mock.return_value = {"copied": ["architecture-diagram"]}
            config = RunnerConfig.from_env()
            SessionManager(config)

            with open(os.path.join(tempdir, "honcho.json"), "r", encoding="utf-8") as fh:
                payload = json.load(fh)

        self.assertEqual(payload["environment"], "production")

    def test_session_manager_syncs_bundled_skills_when_configured(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, tempfile.TemporaryDirectory() as bundled, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch.dict(os.environ, {**runner_env("eval"), "HERMES_HOME": tempdir}, clear=True):
            with mock.patch("rsi_runner.session_manager.sync_skills") as sync_mock, mock.patch.dict(
                os.environ,
                {**runner_env("eval"), "HERMES_HOME": tempdir, "HERMES_BUNDLED_SKILLS": bundled},
                clear=True,
            ):
                sync_mock.return_value = {"copied": ["architecture-diagram"]}
                config = RunnerConfig.from_env()
                manager = SessionManager(config)

        self.assertEqual(manager.skills_dir, os.path.join(tempdir, "skills"))
        self.assertTrue(manager.bundled_skills_available)
        self.assertEqual(manager.bundled_skills_sync_status, "synced")
        self.assertEqual(manager.bundled_skills_sync_error, "")

    def test_execution_observation_id_distinguishes_invocations(self) -> None:
        first = execution_observation_id("", "trace-1", "wf-1", "sess-1", "invoke-a")
        second = execution_observation_id("", "trace-1", "wf-1", "sess-1", "invoke-b")

        self.assertNotEqual(first, second)

    def test_execution_observation_id_preserves_operation_id_hash_when_present(self) -> None:
        first = execution_observation_id("op-1", "trace-1", "wf-1", "sess-1", "invoke-a")
        second = execution_observation_id("op-1", "trace-1", "wf-1", "sess-1", "invoke-b")

        self.assertEqual(first, second)

    def test_read_native_executor_stream_waits_for_read_completion_before_closing(self) -> None:
        finished = threading.Event()
        result_detected = threading.Event()

        class SlowStream:
            def __init__(self) -> None:
                self.closed_before_finish = False
                self.calls = 0

            def read(self, _size: int) -> str:
                if self.calls == 0:
                    self.calls += 1
                    time.sleep(0.05)
                    return "chunk"
                finished.set()
                return ""

            def close(self) -> None:
                if not finished.is_set():
                    self.closed_before_finish = True

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        stream = SlowStream()
        chunks: list[str] = []
        runtime._read_native_executor_stream(
            stream,
            stream_name="stderr",
            phase="investigate",
            observer=None,
            chunk_store=chunks,
            secret_values=[],
            result_detected=result_detected,
        )

        self.assertEqual("".join(chunks), "chunk")
        self.assertTrue(finished.is_set())
        self.assertFalse(stream.closed_before_finish)

    def test_session_manager_writes_hermes_cli_parity_model_config(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch("rsi_runner.session_manager.sync_skills") as sync_mock, mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": tempdir, "OPENAI_BASE_URL": "https://api.openai.com/v1"},
            clear=True,
        ):
            sync_mock.return_value = {"copied": ["architecture-diagram"]}
            config = RunnerConfig.from_env()
            manager = SessionManager(config)
            config_text = Path(tempdir, "config.yaml").read_text(encoding="utf-8")

        self.assertIn("model:", config_text)
        self.assertIn('default: "gpt-5.4"', config_text)
        self.assertIn("provider: custom", config_text)
        self.assertIn('base_url: "https://api.openai.com/v1"', config_text)
        self.assertIn('api_key: ""', config_text)
        self.assertEqual(manager.hermes_config_parity_status, "configured")
        self.assertEqual(manager.hermes_config_parity_error, "")

    def test_session_manager_quotes_yaml_sensitive_model_values(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch("rsi_runner.session_manager.sync_skills") as sync_mock, mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": tempdir,
                "RSI_RUNNER_MODEL": "openai/gpt-5.4:beta # unsafe",
                "OPENAI_BASE_URL": "https://api.openai.com/v1?x=#frag",
            },
            clear=True,
        ):
            sync_mock.return_value = {"copied": ["architecture-diagram"]}
            config = RunnerConfig.from_env()
            SessionManager(config)
            config_text = Path(tempdir, "config.yaml").read_text(encoding="utf-8")

        self.assertIn('default: "gpt-5.4:beta # unsafe"', config_text)
        self.assertIn('base_url: "https://api.openai.com/v1?x=#frag"', config_text)

    def test_session_manager_records_skill_sync_failure_without_disabling_runner(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, tempfile.TemporaryDirectory() as bundled, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch("rsi_runner.session_manager.sync_skills", side_effect=RuntimeError("sync failed")), mock.patch.dict(
            os.environ,
            {**runner_env("eval"), "HERMES_HOME": tempdir, "HERMES_BUNDLED_SKILLS": bundled},
            clear=True,
        ):
            config = RunnerConfig.from_env()
            manager = SessionManager(config)

        self.assertTrue(manager.available)
        self.assertEqual(manager.bundled_skills_sync_status, "failed")
        self.assertEqual(manager.bundled_skills_sync_error, "sync failed")
        self.assertFalse(manager.skills_healthy)

    def test_session_manager_preserves_hermes_config_parity_when_honcho_write_fails(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch("rsi_runner.session_manager.sync_skills") as sync_mock, mock.patch(
            "rsi_runner.session_manager.SessionManager._write_honcho_config", side_effect=RuntimeError("honcho write failed")
        ), mock.patch.dict(
            os.environ,
            {**runner_env("eval"), "HERMES_HOME": tempdir},
            clear=True,
        ):
            sync_mock.return_value = {"copied": ["architecture-diagram"]}
            config = RunnerConfig.from_env()
            manager = SessionManager(config)

        self.assertEqual(manager.hermes_config_parity_status, "configured")
        self.assertEqual(manager.hermes_config_parity_error, "")
        self.assertIn("configure Honcho persistence failed: honcho write failed", manager.ready_issues)

    def test_session_manager_reports_not_configured_before_sync_unavailable(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch("rsi_runner.session_manager.sync_skills", None), mock.patch.dict(
            os.environ,
            {**runner_env("eval"), "HERMES_HOME": tempdir},
            clear=True,
        ):
            config = RunnerConfig.from_env()
            manager = SessionManager(config)

        self.assertFalse(manager.bundled_skills_available)
        self.assertEqual(manager.bundled_skills_sync_status, "not_configured")
        self.assertEqual(manager.bundled_skills_sync_error, "")

    def test_runner_task_request_normalizes_non_list_payload_fields(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "eval",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the eval.",
                    "allowed_tools": "repo.context",
                    "allowed_commands": "python -m pytest",
                    "expected_outputs": "final answer",
                    "rejected_proposal_context": "nope",
                    "recent_conversation_entries": "nope",
                    "prior_trace_refs": "nope",
                    "repo_allowlist": "nope",
                    "tool_allowlist": "nope",
                    "context_refs": "nope",
                    "allowed_path_globs": "nope",
                    "case_summary": "nope",
                }
            }
        )

        self.assertEqual(task.allowed_tools, [])
        self.assertEqual(task.allowed_commands, [])
        self.assertEqual(task.expected_outputs, [])
        self.assertEqual(task.rejected_proposal_context, [])
        self.assertEqual(task.recent_conversation_entries, [])
        self.assertEqual(task.prior_trace_refs, [])
        self.assertEqual(task.repo_allowlist, [])
        self.assertEqual(task.tool_allowlist, [])
        self.assertEqual(task.context_refs, [])
        self.assertEqual(task.allowed_path_globs, [])
        self.assertIsNone(task.case_summary)

    def test_observation_execution_id_matches_harness_execution_prefix_for_operation_id(self) -> None:
        observation_id = execution_observation_id("op-123", "trace-123", "wf-123", "session-123")
        self.assertTrue(observation_id.startswith("hexec-"))

    def test_readonly_tool_binding_uses_execution_phase_for_observations(self) -> None:
        class RecordingObserver:
            def __init__(self) -> None:
                self.events: list[dict[str, object]] = []

            def emit(self, *, phase: str, event_type: str, status: str, payload: dict[str, object] | None = None) -> None:
                self.events.append(
                    {
                        "phase": phase,
                        "event_type": event_type,
                        "status": status,
                        "payload": dict(payload or {}),
                    }
                )

        observer = RecordingObserver()
        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal:8080",
            allowed_tool_names=[],
            task_repo="depin-backend",
            task_repo_ref="",
            task_prompt="Deliver the final reply.",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
            execution_phase="deliver",
            observer=observer,
        )
        binding._record_tool_call(
            canonical_name="slack.upload_file",
            tool_call_id="tool-1",
            request_payload={"filename": "diagram.html"},
            started_at=0.0,
            completed_at=1.0,
            status="completed",
            summary="uploaded",
            provider_ref="file-123",
        )

        self.assertEqual(observer.events[0]["phase"], "deliver")
        self.assertEqual(observer.events[0]["event_type"], "tool.call.completed")

    def test_runner_task_request_falls_back_to_top_level_payload_when_task_is_not_object(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": "not-an-object",
                "task_type": "general",
                "repo": "rsi-agent-platform",
                "prompt": "Use the top-level prompt.",
            }
        )

        self.assertEqual(task.task_type, "general")
        self.assertEqual(task.repo, "rsi-agent-platform")
        self.assertEqual(task.prompt, "Use the top-level prompt.")

    def test_runner_task_request_normalizes_requested_artifacts(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Render the architecture diagram.",
                    "requested_artifacts": [
                        {"kind": "diagram", "description": "Render the requested system diagram."},
                        {"kind": "", "description": "skip empty"},
                        "skip non-object",
                    ],
                    "artifact_optional": True,
                }
            }
        )

        self.assertEqual(
            task.requested_artifacts,
            [{"kind": "diagram", "description": "Render the requested system diagram."}],
        )
        self.assertTrue(task.artifact_optional)

    def test_runner_task_request_normalizes_requested_skills(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Use the architecture skill.",
                    "requested_skills": ["architecture_diagram", "architecture-diagram", "", 1],
                }
            }
        )

        self.assertEqual(task.requested_skills, ["architecture-diagram"])

    def test_openai_models_use_persisted_hermes_sessions_with_xhigh(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "eval",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the eval.",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "reasoning_verbosity": "verbose",
                    "session_scope_kind": "eval_line",
                    "session_scope_id": "shared-store:pk-collision",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:eval",
                    "user_peer_id": "operator:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("eval"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.provider, "hermes-aiagent")
        self.assertEqual(FakeAIAgent.last_kwargs["model"], "gpt-5.4")
        self.assertEqual(FakeAIAgent.last_kwargs["api_mode"], "codex_responses")
        self.assertEqual(FakeAIAgent.last_kwargs["provider"], "custom")
        self.assertEqual(FakeAIAgent.last_kwargs["reasoning_config"], {"enabled": True, "effort": "xhigh"})
        self.assertEqual(FakeAIAgent.last_kwargs["enabled_toolsets"], [])
        self.assertEqual(FakeAIAgent.last_kwargs["session_id"], "rsi-prod-conversation-123")
        self.assertTrue(FakeAIAgent.last_kwargs["persist_session"])
        self.assertFalse(FakeAIAgent.last_kwargs["skip_memory"])
        self.assertEqual(FakeAIAgent.last_history, [{"role": "user", "content": "Earlier thread message"}])
        self.assertEqual(result.raw["model"], "openai/gpt-5.4")
        self.assertEqual(result.raw["provider_model"], "gpt-5.4")
        self.assertEqual(result.raw["api_mode"], "codex_responses")
        self.assertEqual(result.raw["reasoning_effort"], "xhigh")
        self.assertEqual(result.raw["hermes_session_id"], "rsi-prod-conversation-123")
        self.assertEqual(result.raw["memory_backend"], "honcho")
        self.assertEqual(result.raw["honcho_workspace"], "rsi-stage")
        self.assertEqual(result.raw["honcho_environment"], "stage")
        self.assertEqual(result.raw["honcho_environment_effective"], "production")
        self.assertEqual(result.raw["honcho_recall_mode"], "hybrid")
        self.assertEqual(result.raw["honcho_write_frequency"], "async")
        self.assertEqual(result.raw["honcho_session_strategy"], "hybrid")
        self.assertEqual(result.raw["honcho_ai_peer"], "rsi:stage:eval")
        self.assertEqual(result.raw["hermes_pin"], "6fdbf2f2d76cf37393e657bf37ceda3d84589200")
        self.assertIn("structured_output", result.raw)
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")

    def test_execute_task_routes_workflow_to_native_executor_when_enabled(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Render the current architecture diagram.",
                    "execution_id": "hexec-test-routing",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "operation_id": "op-123",
                }
            }
        )
        expected = HermesExecutionResult(
            ok=True,
            message=json.dumps(
                partial_structured_output(
                    reply_text="Native executor reply",
                    proposed_actions=[
                        {
                            "kind": "slack_post",
                            "target_ref": "C123",
                            "request_payload": {"body": "Native executor reply"},
                            "idempotency_key": "reply-1",
                            "rationale": "Post the final answer.",
                        }
                    ],
                )
            ),
            provider="hermes-native-executor",
            raw={},
        )
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_execute_native_workflow_task_request", return_value=expected) as native_mock, mock.patch.object(
                runtime, "_execute_task_request", autospec=True
            ) as standard_mock:
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        native_mock.assert_called_once()
        standard_mock.assert_not_called()
        self.assertEqual(runtime.executor_status("hexec-test-routing")["status"], "running")

    def test_rsi_tool_wrappers_use_transport_names_and_reverse_map_to_canonical_gateway_ids(self) -> None:
        wrappers = tool_schema_wrappers(["repo.context", "rsi.workflow_context"])
        self.assertEqual(
            [item["function"]["name"] for item in wrappers],
            ["repo_context", "rsi_workflow_context"],
        )

        captured: dict[str, object] = {}

        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Context gathered.",
                        "provider": "tool-gateway",
                        "provider_ref": "repo.context-call",
                        "output": {"ok": True},
                    }
                ).encode("utf-8")

        def fake_urlopen(req, timeout: int = 0):
            captured["url"] = req.full_url
            captured["timeout"] = timeout
            captured["body"] = json.loads(req.data.decode("utf-8"))
            return FakeResponse()

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["repo.context", "rsi.workflow_context"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="What broke?",
            task_channel_id="",
            task_thread_ts="",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_urlopen):
            payload = json.loads(binding.handle_tool_call("repo_context", {"question": "What broke?"}))

        self.assertEqual(captured["url"], "http://tool-gateway.internal/api/tools/repo.context/execute")
        self.assertEqual(captured["body"], {"trace_id": "trace-123", "repo": "rsi-agent-platform", "question": "What broke?"})
        self.assertEqual(payload["tool_name"], "repo.context")
        self.assertEqual(payload["transport_tool_name"], "repo_context")
        self.assertEqual(payload["tool_call_id"], "repo.context:1")
        self.assertEqual(len(binding.diagnostics()["tool_calls"]), 1)
        self.assertEqual(binding.diagnostics()["tool_calls"][0]["tool_name"], "repo.context")
        self.assertEqual(binding.diagnostics()["tool_calls"][0]["tool_call_id"], "repo.context:1")
        self.assertEqual(binding.diagnostics()["tool_calls"][0]["request"], {"trace_id": "trace-123", "repo": "rsi-agent-platform", "question": "What broke?"})
        self.assertEqual(binding.diagnostics()["tool_calls"][0]["summary"], "Context gathered.")

    def test_cancel_execution_marks_active_execution_as_cancelling(self) -> None:
        class FakeProcess:
            def __init__(self) -> None:
                self.terminated = False

            def poll(self):
                return None

            def terminate(self):
                self.terminated = True

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        process = FakeProcess()
        runtime._store_executor_result("hexec-cancel", {"execution_id": "hexec-cancel", "status": "running"})
        with runtime._executor_process_lock:
            runtime._executor_processes["hexec-cancel"] = process

        status = runtime.cancel_execution("hexec-cancel")

        self.assertTrue(process.terminated)
        self.assertEqual(status["status"], "cancelling")
        self.assertEqual(runtime.executor_status("hexec-cancel")["status"], "cancelling")

    def test_slack_history_binding_defaults_to_bound_channel_context(self) -> None:
        captured: dict[str, object] = {}

        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Slack thread history loaded.",
                        "provider": "slack",
                        "provider_ref": "171000001.000100",
                        "output": {"messages": []},
                    }
                ).encode("utf-8")

        def fake_urlopen(req, timeout: int = 0):
            captured["url"] = req.full_url
            captured["timeout"] = timeout
            captured["body"] = json.loads(req.data.decode("utf-8"))
            return FakeResponse()

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["slack.history"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="What did we say in the latest convo?",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_urlopen):
            payload = json.loads(binding.handle_tool_call("slack_history", {}))

        self.assertEqual(captured["url"], "http://tool-gateway.internal/api/tools/slack.history/execute")
        self.assertEqual(
            captured["body"],
            {
                "trace_id": "trace-123",
                "channel_id": "C123",
                "thread_ts": "171000001.000100",
                "question": "What did we say in the latest convo?",
            },
        )
        self.assertEqual(payload["tool_name"], "slack.history")
        self.assertEqual(payload["transport_tool_name"], "slack_history")

    def test_slack_history_binding_records_grounded_message_evidence_items(self) -> None:
        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Slack thread history loaded.",
                        "provider": "slack",
                        "provider_ref": "171000001.000100",
                        "output": {
                            "channel_id": "C123",
                            "thread_ts": "171000001.000100",
                            "messages": [
                                {
                                    "author_name": "blake",
                                    "user": "U123",
                                    "ts": "171000001.000100",
                                    "thread_ts": "171000001.000100",
                                    "text": "Pinned the timeout increase to five minutes for this rollout.",
                                    "permalink": "https://slack.example/C123/p1710000010000100",
                                }
                            ],
                        },
                    }
                ).encode("utf-8")

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["slack.history"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="Summarize the thread.",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", return_value=FakeResponse()):
            payload = json.loads(binding.handle_tool_call("slack_history", {}))

        self.assertEqual(payload["status"], "completed")
        evidence_items = binding.diagnostics()["evidence_items"]
        self.assertEqual(len(evidence_items), 1)
        self.assertEqual(evidence_items[0]["kind"], "slack_message")
        self.assertEqual(evidence_items[0]["summary"], "[blake] Pinned the timeout increase to five minutes for this rollout.")
        self.assertEqual(evidence_items[0]["snippet"], "Pinned the timeout increase to five minutes for this rollout.")
        self.assertEqual(evidence_items[0]["author"], "blake")
        self.assertEqual(evidence_items[0]["message_ts"], "171000001.000100")
        self.assertEqual(evidence_items[0]["thread_ts"], "171000001.000100")
        self.assertEqual(evidence_items[0]["source_ref"], "https://slack.example/C123/p1710000010000100")

    def test_slack_search_binding_defaults_to_bound_channel_context(self) -> None:
        captured: dict[str, object] = {}

        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Slack search loaded.",
                        "provider": "slack",
                        "provider_ref": "control plane timeout",
                        "output": {"messages": []},
                    }
                ).encode("utf-8")

        def fake_urlopen(req, timeout: int = 0):
            captured["url"] = req.full_url
            captured["timeout"] = timeout
            captured["body"] = json.loads(req.data.decode("utf-8"))
            return FakeResponse()

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["slack.search"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="Where did we decide to bump the control plane to 5 minutes?",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_urlopen):
            payload = json.loads(binding.handle_tool_call("slack_search", {}))

        self.assertEqual(captured["url"], "http://tool-gateway.internal/api/tools/slack.search/execute")
        self.assertEqual(
            captured["body"],
            {
                "trace_id": "trace-123",
                "channel_ids": ["C123"],
                "query": "Where did we decide to bump the control plane to 5 minutes?",
            },
        )
        self.assertEqual(payload["tool_name"], "slack.search")
        self.assertEqual(payload["transport_tool_name"], "slack_search")

    def test_readonly_tool_binding_null_args_do_not_clobber_defaults(self) -> None:
        captured: dict[str, object] = {}

        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Repo context loaded.",
                        "provider": "github",
                        "provider_ref": "https://github.com/piplabs/rsi-agent-platform",
                        "output": {},
                    }
                ).encode("utf-8")

        def fake_urlopen(req, timeout: int = 0):
            captured["url"] = req.full_url
            captured["timeout"] = timeout
            captured["body"] = json.loads(req.data.decode("utf-8"))
            return FakeResponse()

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["repo.context"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="Summarize the latest workflow fix.",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_urlopen):
            payload = json.loads(binding.handle_tool_call("repo_context", {"repo": None, "question": None}))

        self.assertEqual(captured["url"], "http://tool-gateway.internal/api/tools/repo.context/execute")
        self.assertEqual(
            captured["body"],
            {
                "trace_id": "trace-123",
                "repo": "rsi-agent-platform",
                "question": "Summarize the latest workflow fix.",
            },
        )
        self.assertEqual(payload["tool_name"], "repo.context")
        self.assertEqual(payload["transport_tool_name"], "repo_context")

    def test_slack_upload_file_binding_defaults_to_bound_thread(self) -> None:
        captured: dict[str, object] = {}

        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Slack file uploaded.",
                        "provider": "slack",
                        "provider_ref": "F123",
                        "output": {"uploaded": True, "file_id": "F123"},
                    }
                ).encode("utf-8")

        def fake_urlopen(req, timeout: int = 0):
            captured["url"] = req.full_url
            captured["timeout"] = timeout
            captured["body"] = json.loads(req.data.decode("utf-8"))
            return FakeResponse()

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["slack.upload_file"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="Reply with the generated file.",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_urlopen):
            payload = json.loads(binding.handle_tool_call("slack_upload_file", {"filename": "diagram.svg", "content": "<svg />"}))

        self.assertEqual(captured["url"], "http://tool-gateway.internal/api/tools/slack.upload_file/execute")
        self.assertEqual(
            captured["body"],
            {
                "trace_id": "trace-123",
                "channel_id": "C123",
                "thread_ts": "171000001.000100",
                "filename": "diagram.svg",
                "content": "<svg />",
            },
        )
        self.assertEqual(payload["tool_name"], "slack.upload_file")
        self.assertEqual(payload["transport_tool_name"], "slack_upload_file")

    def test_slack_upload_file_binding_reads_local_path_and_forwards_base64(self) -> None:
        captured: dict[str, object] = {}

        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Slack file uploaded.",
                        "provider": "slack",
                        "provider_ref": "F456",
                        "output": {"uploaded": True, "file_id": "F456"},
                    }
                ).encode("utf-8")

        def fake_urlopen(req, timeout: int = 0):
            captured["body"] = json.loads(req.data.decode("utf-8"))
            return FakeResponse()

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["slack.upload_file"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="Reply with the generated file.",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "diagram.svg"
            file_path.write_text("<svg>generated</svg>", encoding="utf-8")
            with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_urlopen):
                payload = json.loads(binding.handle_tool_call("slack_upload_file", {"path": str(file_path)}))

        body = captured["body"]
        self.assertEqual(body["channel_id"], "C123")
        self.assertEqual(body["thread_ts"], "171000001.000100")
        self.assertEqual(body["filename"], "diagram.svg")
        self.assertEqual(Path(body["path"]).name, "diagram.svg")
        self.assertEqual(base64.b64decode(body["content_base64"]).decode("utf-8"), "<svg>generated</svg>")
        self.assertEqual(payload["tool_name"], "slack.upload_file")
        self.assertEqual(payload["transport_tool_name"], "slack_upload_file")

    def test_slack_upload_file_binding_reads_file_artifact_ref(self) -> None:
        captured: dict[str, object] = {}

        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Slack file uploaded.",
                        "provider": "slack",
                        "provider_ref": "F789",
                        "output": {"uploaded": True, "file_id": "F789"},
                    }
                ).encode("utf-8")

        def fake_urlopen(req, timeout: int = 0):
            captured["body"] = json.loads(req.data.decode("utf-8"))
            return FakeResponse()

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["slack.upload_file"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="Reply with the generated file.",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "diagram.png"
            file_path.write_bytes(b"\x89PNG\r\n")
            artifact_ref = file_path.as_uri()
            with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_urlopen):
                _ = json.loads(binding.handle_tool_call("slack_upload_file", {"artifact_ref": artifact_ref}))

        body = captured["body"]
        self.assertEqual(body["filename"], "diagram.png")
        self.assertEqual(Path(body["path"]).name, "diagram.png")
        self.assertEqual(base64.b64decode(body["content_base64"]), b"\x89PNG\r\n")

    def test_repo_context_binding_records_grounded_match_snippets(self) -> None:
        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "GitHub-backed repo context loaded for piplabs/rsi-agent-platform with 1 relevant code match(es).",
                        "provider": "github",
                        "provider_ref": "https://github.com/piplabs/rsi-agent-platform",
                        "output": {
                            "repo": "rsi-agent-platform",
                            "default_branch": "main",
                            "description": "Runner support for bounded-stop workflow replies.",
                            "html_url": "https://github.com/piplabs/rsi-agent-platform",
                            "matches": [
                                {
                                    "path": "runner/rsi_runner/hermes_runtime.py",
                                    "html_url": "https://github.com/piplabs/rsi-agent-platform/blob/main/runner/rsi_runner/hermes_runtime.py",
                                    "snippet": "completion_verdict = partial when the reducer had enough grounded evidence",
                                }
                            ],
                        },
                    }
                ).encode("utf-8")

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["repo.context"],
            task_repo="rsi-agent-platform",
            task_repo_ref="main",
            task_prompt="Where is partial completion handled?",
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="workflow summary",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", return_value=FakeResponse()):
            payload = json.loads(binding.handle_tool_call("repo_context", {}))

        self.assertEqual(payload["status"], "completed")
        evidence_items = binding.diagnostics()["evidence_items"]
        self.assertEqual(len(evidence_items), 2)
        self.assertEqual(evidence_items[0]["kind"], "repo_context")
        self.assertEqual(evidence_items[0]["snippet"], "Runner support for bounded-stop workflow replies.")
        self.assertEqual(evidence_items[0]["default_branch"], "main")
        self.assertEqual(evidence_items[1]["kind"], "repo_context_match")
        self.assertEqual(
            evidence_items[1]["snippet"],
            "completion_verdict = partial when the reducer had enough grounded evidence",
        )
        self.assertEqual(evidence_items[1]["path"], "runner/rsi_runner/hermes_runtime.py")
        self.assertEqual(
            evidence_items[1]["source_ref"],
            "https://github.com/piplabs/rsi-agent-platform/blob/main/runner/rsi_runner/hermes_runtime.py",
        )

    def test_invalid_tool_name_preflight_fails_before_provider_execution(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the workflow.",
                    "allowed_tools": ["repo.context"],
                    "tool_allowlist": ["repo.context"],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )

        def fake_transport_name(name: str) -> str:
            if name == "repo.context":
                raise ValueError("unsafe tool name")
            return name

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.tool_transport_name", side_effect=fake_transport_name), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_invalid_request")
        self.assertNotIn("structured_output_error", result.raw)
        self.assertEqual(result.raw["runner_diagnostics"]["failure_kind"], "invalid_request")
        self.assertEqual(result.raw["runner_diagnostics"]["provider_error_param"], "tools[0].name")
        self.assertEqual(result.raw["runner_diagnostics"]["invalid_tool_names"], ["repo.context"])
        self.assertIsNone(FakeAIAgent.last_kwargs)

    def test_workflow_non_json_output_repairs_successfully(self) -> None:
        class RepairingAIAgent(FakeAIAgent):
            calls = 0
            prompts: list[str] = []

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict:
                type(self).calls += 1
                type(self).prompts.append(prompt)
                type(self).last_prompt = prompt
                type(self).last_system_message = system_message
                type(self).last_history = conversation_history or []
                if type(self).calls == 1:
                    return {"final_response": "plain text response"}
                return {
                    "final_response": json.dumps(
                        {
                            "visible_reasoning": [],
                            "reply_draft": "Draft reply",
                            "final_answer": "Final reply",
                            "confidence": 0.91,
                            "context_summary": "Repo and KB context collected.",
                            "self_critique": "",
                            "proposed_actions": [],
                            "knowledge_drafts": [],
                            "outcome_hypotheses": [],
                        }
                    )
                }

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "User request: Please use /architecture-diagram skill and summarize the workflow.\n\nInvestigate within the governed tool boundary.",
                    "requested_skills": ["architecture-diagram"],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", RepairingAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_skill_invocation_message",
            return_value=None,
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_preloaded_skills_prompt",
            return_value=("[PRELOADED architecture-diagram]", ["architecture-diagram"], []),
        ), mock.patch(
            "rsi_runner.hermes_runtime.resolve_skill_command_key",
            return_value="/architecture-diagram",
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertTrue(result.raw["repair_attempted"])
        self.assertTrue(result.raw["repair_succeeded"])
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")
        self.assertEqual(result.raw["repair_original_response"], "plain text response")
        self.assertGreaterEqual(len(RepairingAIAgent.prompts), 2)
        self.assertEqual(RepairingAIAgent.prompts[1].count("Runner role:"), 1)
        self.assertEqual(RepairingAIAgent.prompts[1].count("[PRELOADED architecture-diagram]"), 1)

    def test_workflow_non_json_output_fails_closed_after_single_repair(self) -> None:
        class UnstructuredAIAgent(FakeAIAgent):
            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict:
                type(self).last_prompt = prompt
                type(self).last_system_message = system_message
                type(self).last_history = conversation_history or []
                return {"final_response": "plain text response"}

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the workflow.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", UnstructuredAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertIn("structured output", result.message)
        self.assertEqual(result.raw["raw_response"], "plain text response")
        self.assertIn("structured_output_error", result.raw)
        self.assertTrue(result.raw["repair_attempted"])
        self.assertFalse(result.raw["repair_succeeded"])

    def test_workflow_structured_output_is_normalized_to_contract_shape(self) -> None:
        class MessyStructuredOutputAIAgent(FakeAIAgent):
            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict:
                type(self).last_prompt = prompt
                type(self).last_system_message = system_message
                type(self).last_history = conversation_history or []
                return {
                    "final_response": json.dumps(
                        {
                            "visible_reasoning": [
                                "Loaded workflow context for the active trace.",
                            ],
                            "reply_draft": "Draft reply",
                            "final_answer": "Final reply",
                            "confidence": 0.91,
                            "context_summary": {
                                "time_window": "2026-04-10T05:01:43Z to 2026-04-17T05:01:43Z",
                                "workflow_context": "Only the inbound Slack request was surfaced.",
                            },
                            "self_critique": "",
                            "proposed_actions": [],
                            "knowledge_drafts": [
                                {
                                    "kind": "investigation_gap",
                                    "scope_type": "repo",
                                    "scope_id": "depin-backend",
                                    "title": "Gap in accessible evidence",
                                    "summary": "The run lacked channel excerpts.",
                                    "body": "Only the inbound Slack request was available.",
                                    "confidence": 0.78,
                                    "fresh_until": "2026-04-17T06:01:43Z",
                                    "evidence_refs": [
                                        "rsi.workflow_context.output.recent_conversation_entries",
                                    ],
                                }
                            ],
                            "outcome_hypotheses": [
                                {
                                    "outcome_type": "answer_limitation",
                                    "success_condition": "Avoid unsupported claims.",
                                    "measurement_ref": "final_answer",
                                    "expected_time_horizon": "immediate",
                                }
                            ],
                            "validation_plan": [
                                "Check workflow context for actual discussion evidence.",
                                "Check repo context for Numo-related activity.",
                            ],
                            "retry_assessment": {
                                "failure_class": "runner_structured_output_parse_failure",
                                "failure_summary": "non-canonical payload shape",
                                "retry_decision": "retry",
                                "material_hypothesis_change": True,
                                "changed_files": ["runner/rsi_runner/hermes_runtime.py"],
                            },
                            "hypothesis_delta": "Need canonical string/object normalization.",
                        }
                    )
                }

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the workflow.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", MessyStructuredOutputAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        normalized = result.raw["structured_output"]
        self.assertEqual(
            normalized["context_summary"],
            json.dumps(
                {
                    "time_window": "2026-04-10T05:01:43Z to 2026-04-17T05:01:43Z",
                    "workflow_context": "Only the inbound Slack request was surfaced.",
                },
                ensure_ascii=True,
                sort_keys=True,
            ),
        )
        self.assertEqual(normalized["validation_plan"], "Check workflow context for actual discussion evidence.\nCheck repo context for Numo-related activity.")
        self.assertEqual(normalized["visible_reasoning"][0]["step_type"], "analysis")
        self.assertEqual(normalized["visible_reasoning"][0]["summary"], "Loaded workflow context for the active trace.")
        self.assertEqual(
            normalized["knowledge_drafts"][0]["evidence_refs"],
            [{"kind": "reference", "ref": "rsi.workflow_context.output.recent_conversation_entries"}],
        )

    def test_system_message_is_forwarded_to_hermes_run_conversation(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute("User prompt", system_message="System directive")

        self.assertTrue(result.ok)
        self.assertEqual(FakeAIAgent.last_prompt, "User prompt")
        self.assertEqual(FakeAIAgent.last_system_message, "System directive")

    def test_workflow_uses_native_executor_subprocess_when_enabled(self) -> None:
        structured = json.dumps(
            {
                "visible_reasoning": [],
                "reply_draft": "Draft reply",
                "final_answer": "Final reply",
                "confidence": 0.9,
                "context_summary": "Grounded context",
                "self_critique": "",
                "proposed_actions": [],
                "knowledge_drafts": [],
                "outcome_hypotheses": [],
                "produced_artifacts": [],
                "artifact_failure_reason": "",
            }
        )
        request_paths: list[Path] = []
        run_dir_exists_before_temp_cleanup = False
        mcp_registration = TaskScopedMCPRegistration(
            servers=[
                TaskScopedMCPServer(
                    source_label="slack",
                    profile="slack_mcp_reply",
                    server_name="rsi-task-trace-123-0-slack-abc",
                    toolset_alias="mcp-rsi-task-trace-123-0-slack-abc",
                    included_tool_names=["slack_send_message"],
                    hermes_config={"url": "https://mcp.slack.com/mcp"},
                )
            ]
        )

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request_path = Path(cmd[-1])
                request_paths.append(request_path)
                request = json.loads(request_path.read_text(encoding="utf-8"))
                self.returncode = 0
                self._payload = {
                    "ok": True,
                    "mcp_cleanup_errors": [],
                    "mcp_cleanup_status": "cleaned",
                    "response": structured,
                    "result": {"final_response": structured},
                    "session_id": "rsi-prod-conversation-123",
                }
                Path(request["result_path"]).write_text(json.dumps(self._payload, sort_keys=True), encoding="utf-8")
                self.stdout = io.StringIO("noise only\n")
                self.stderr = io.StringIO("")
                self._request = request
                self._cwd = cwd
                self._text = text
                self._env = env

            def poll(self):
                return 0

            def wait(self, timeout=None):
                return self.returncode

            def communicate(self, timeout=None):
                raise AssertionError("native executor path should not rely on communicate() once pipes are drained asynchronously")

            def terminate(self):
                self.returncode = -15

            def kill(self):
                self.returncode = -9

        def fake_popen(cmd, cwd, env, stdout, stderr, text):
            process = FakePopen(cmd, cwd, env, stdout, stderr, text)
            self.assertEqual(process._request["prompt"], "User prompt")
            self.assertEqual(process._request["system_message"], "System directive")
            self.assertEqual(process._request["requested_skills"], ["architecture-diagram"])
            self.assertIn("todo", process._request["toolsets"])
            self.assertIn("rsi-governed-readonly", process._request["toolsets"])
            self.assertIn("mcp-rsi-task-trace-123-0-slack-abc", process._request["toolsets"])
            self.assertEqual(
                process._request["task_scoped_mcp_servers"],
                [
                    {
                        "source_label": "slack",
                        "profile": "slack_mcp_reply",
                        "server_name": "rsi-task-trace-123-0-slack-abc",
                        "toolset_alias": "mcp-rsi-task-trace-123-0-slack-abc",
                        "included_tool_names": ["slack_send_message"],
                        "hermes_config": {"url": "https://mcp.slack.com/mcp"},
                    }
                ],
            )
            self.assertTrue(process._request["result_path"])
            self.assertNotIn("api_key", process._request["runtime"])
            self.assertEqual(cwd, str((Path(tempdir) / "company").resolve()))
            self.assertTrue(process._text)
            return process

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=fake_popen
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime._mcp_adapter, "plan_task_servers", return_value=mcp_registration):
                task = RunnerTaskRequest.from_payload(
                    {
                        "task": {
                            "task_type": "workflow",
                            "repo": "depin-backend",
                            "prompt": "User prompt",
                            "system_message": "System directive",
                            "trace_id": "trace-123",
                            "workflow_id": "wf-123",
                            "operation_id": "op-123",
                            "session_scope_kind": "conversation",
                            "session_scope_id": "conv-123",
                            "memory_backend": "honcho",
                            "assistant_peer_id": "rsi:stage:prod",
                            "requested_skills": ["architecture-diagram"],
                            "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_reply"}],
                        }
                    }
                )
                result = runtime.execute_task(task)
            run_dir_exists_before_temp_cleanup = bool(
                request_paths and request_paths[0].parent.exists() and request_paths[0].exists()
            )

        self.assertTrue(result.ok)
        self.assertEqual(result.provider, "hermes-native-executor")
        self.assertEqual(result.raw["native_executor_mode"], "subprocess")
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")
        self.assertEqual(result.raw["system_message"], "System directive")
        self.assertEqual(result.raw["runner_diagnostics"]["agentic_mcp_cleanup_status"], "cleaned")
        self.assertEqual(len(request_paths), 1)
        self.assertTrue(run_dir_exists_before_temp_cleanup)

    def test_native_render_request_payload_uses_local_artifact_dir_and_strips_governed_toolsets(self) -> None:
        structured = json.dumps(
            {
                "visible_reasoning": [],
                "reply_draft": "",
                "final_answer": "",
                "confidence": 0.9,
                "context_summary": "Rendered artifact",
                "self_critique": "",
                "proposed_actions": [],
                "knowledge_drafts": [],
                "outcome_hypotheses": [],
                "produced_artifacts": [],
                "artifact_failure_reason": "",
            }
        )

        def fake_popen(cmd, cwd, env, stdout, stderr, text):
            request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
            Path(request["result_path"]).write_text(
                json.dumps(
                    {
                        "ok": True,
                        "response": structured,
                        "result": {"final_response": structured},
                        "session_id": "rsi-prod-conversation-123",
                        "artifact_tool_events": [],
                    },
                    sort_keys=True,
                ),
                encoding="utf-8",
            )
            self.assertEqual(request["execution_phase"], "render")
            self.assertEqual(request["artifact_output_dir"], str((Path(tempdir) / "op-render" / "artifacts").resolve()))
            self.assertNotIn("rsi-governed-readonly", request["toolsets"])
            self.assertNotIn("rsi-governed-workspace", request["toolsets"])

            class FakePopen:
                def __init__(self) -> None:
                    self.returncode = 0
                    self.stdout = io.StringIO("")
                    self.stderr = io.StringIO("")

                def poll(self):
                    return 0

                def wait(self, timeout=None):
                    return self.returncode

                def terminate(self):
                    self.returncode = -15

                def kill(self):
                    self.returncode = -9

            return FakePopen()

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=fake_popen
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "Render the diagram.",
                        "trace_id": "trace-render",
                        "workflow_id": "wf-render",
                        "operation_id": "op-render",
                        "execution_phase": "render",
                        "execution_mode": "artifact_render",
                        "artifact_destination": os.path.join(tempdir, "op-render", "artifacts"),
                        "requested_skills": ["architecture-diagram"],
                        "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-render",
                        "memory_backend": "honcho",
                        "assistant_peer_id": "rsi:stage:prod",
                    }
                }
            )
            result = runtime._execute_native_workflow_task_request(
                task,
                runtime._resolve_tool_policy(task),
                observer=RecordingObserver(),
            )

        self.assertTrue(result.ok)

    def test_local_artifact_tool_binding_rejects_writes_outside_root(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            binding = _LocalArtifactToolBinding(Path(tempdir))
            payload = json.loads(
                binding.handle_tool_call(
                    "artifact_write_file",
                    {"path": "diagram.html", "content": "<html></html>"},
                )
            )
            self.assertEqual(payload["status"], "ok")
            self.assertTrue((Path(tempdir) / "diagram.html").exists())
            with self.assertRaises(RuntimeError):
                binding.handle_tool_call(
                    "artifact_write_file",
                    {"path": "../escape.html", "content": "nope"},
                )

    def test_native_worker_initializes_current_hermes_cli_signature(self) -> None:
        class CurrentHermesCLI:
            init_kwargs: dict[str, object] | None = None

            def _init_agent(self, *, model_override=None, runtime_override=None):
                type(self).init_kwargs = {
                    "model_override": model_override,
                    "runtime_override": runtime_override,
                }
                return True

        payload = {
            "model": "openai/gpt-5.4",
            "runtime": {
                "api_key": "test-key",
                "base_url": "https://api.openai.com/v1",
                "provider": "openai",
                "api_mode": "codex_responses",
            },
        }

        cli = CurrentHermesCLI()
        _initialize_cli_agent(cli, payload)

        self.assertEqual(
            CurrentHermesCLI.init_kwargs,
            {
                "model_override": "openai/gpt-5.4",
                "runtime_override": {
                    "api_key": "test-key",
                    "base_url": "https://api.openai.com/v1",
                    "provider": "openai",
                    "api_mode": "codex_responses",
                    "command": None,
                    "args": [],
                    "credential_pool": None,
                },
            },
        )

    def test_native_executor_stores_final_status_under_explicit_execution_id_without_observer(self) -> None:
        structured = json.dumps(
            {
                "visible_reasoning": [],
                "reply_draft": "Draft reply",
                "final_answer": "Final reply",
                "confidence": 0.9,
                "context_summary": "Grounded context",
                "self_critique": "",
                "proposed_actions": [],
                "knowledge_drafts": [],
                "outcome_hypotheses": [],
                "produced_artifacts": [],
                "artifact_failure_reason": "",
            }
        )

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
                self.returncode = 0
                self._payload = {
                    "ok": True,
                    "response": structured,
                    "result": {"final_response": structured},
                    "session_id": "rsi-prod-conversation-123",
                }
                Path(request["result_path"]).write_text(json.dumps(self._payload, sort_keys=True), encoding="utf-8")
                self.stdout = io.StringIO("")
                self.stderr = io.StringIO("")

            def poll(self):
                return 0

            def wait(self, timeout=None):
                return self.returncode

            def communicate(self, timeout=None):
                raise AssertionError("native executor path should not rely on communicate() once pipes are drained asynchronously")

            def terminate(self):
                self.returncode = -15

            def kill(self):
                self.returncode = -9

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=FakePopen
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "trace_id": "trace-123",
                        "workflow_id": "wf-123",
                        "operation_id": "op-123",
                        "execution_id": "hexec-explicit",
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-123",
                        "memory_backend": "honcho",
                        "assistant_peer_id": "rsi:stage:prod",
                    }
                }
            )
            result = runtime._execute_native_workflow_task_request(
                task,
                runtime._resolve_tool_policy(task),
                observer=None,
            )

        self.assertTrue(result.ok)
        self.assertEqual(runtime.executor_status("hexec-explicit")["status"], "completed")
        self.assertEqual(runtime.executor_status("hexec-explicit")["execution_id"], "hexec-explicit")
        self.assertEqual(runtime.executor_status("rsi-prod-conversation-123"), {})

    def test_native_executor_streams_output_detects_result_and_redacts_secrets(self) -> None:
        structured = json.dumps(
            {
                "visible_reasoning": [],
                "reply_draft": "Draft reply",
                "final_answer": "Final reply",
                "confidence": 0.9,
                "context_summary": "Grounded context",
                "self_critique": "",
                "proposed_actions": [],
                "knowledge_drafts": [],
                "outcome_hypotheses": [],
                "produced_artifacts": [],
                "artifact_failure_reason": "",
            }
        )

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
                payload = {
                    "ok": True,
                    "mcp_cleanup_errors": [],
                    "mcp_cleanup_status": "cleaned",
                    "response": structured,
                    "result": {"final_response": structured},
                    "session_id": "rsi-prod-conversation-123",
                }
                self.returncode = 0
                Path(request["result_path"]).write_text(json.dumps(payload, sort_keys=True), encoding="utf-8")
                self.stdout = io.StringIO(
                    "Warning: Unknown toolsets: \n"
                    "mcp-test-slack,\n"
                    "mcp-test-notion\n\n"
                    "prelude\n"
                    "RSI_EXECUTOR_RESULT::"
                    + json.dumps(payload, sort_keys=True)
                    + "\n"
                )
                self.stderr = io.StringIO(
                    "Authorization: Bearer secret-bearer-token\n"
                    "slack=xoxb-123456789-secret\n"
                    "openai=sk-secret-openai-key\n"
                    "aws=aws-session-secret\n"
                    "env=openai-test-key\n"
                )

            def poll(self):
                return 0

            def wait(self, timeout=None):
                return self.returncode

            def terminate(self):
                self.returncode = -15

            def kill(self):
                self.returncode = -9

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=FakePopen
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
                "AWS_SESSION_TOKEN": "aws-session-secret",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            observer = RecordingObserver()
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "trace_id": "trace-123",
                        "workflow_id": "wf-123",
                        "operation_id": "op-123",
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-123",
                        "memory_backend": "honcho",
                        "assistant_peer_id": "rsi:stage:prod",
                    }
                }
            )
            result = runtime._execute_native_workflow_task_request(
                task,
                runtime._resolve_tool_policy(task),
                observer=observer,
            )

        self.assertTrue(result.ok)
        output_events = [event for event in observer.events if event["event_type"] == "executor.subprocess.output"]
        self.assertGreaterEqual(len(output_events), 2)
        stdout_events = [event for event in output_events if event["payload"].get("stream") == "stdout"]
        stderr_events = [event for event in output_events if event["payload"].get("stream") == "stderr"]
        self.assertTrue(any(event["payload"].get("contains_result_marker") for event in stdout_events))
        self.assertTrue(any(event["event_type"] == "executor.result.detected" for event in observer.events))
        self.assertTrue(any(event["event_type"] == "executor.subprocess.completed" for event in observer.events))
        stdout_text = "\n".join(str(event["payload"].get("chunk_text", "")) for event in stdout_events)
        stderr_text = "\n".join(str(event["payload"].get("chunk_text", "")) for event in stderr_events)
        self.assertNotIn("Warning: Unknown toolsets:", stdout_text)
        self.assertNotIn("mcp-test-slack", stdout_text)
        self.assertIn("prelude", stdout_text)
        self.assertIn("Bearer [redacted]", stderr_text)
        self.assertIn("[redacted-slack-token]", stderr_text)
        self.assertIn("[redacted-openai-key]", stderr_text)
        self.assertIn("[redacted]", stderr_text)
        self.assertNotIn("secret-bearer-token", stderr_text)
        self.assertNotIn("xoxb-123456789-secret", stderr_text)
        self.assertNotIn("sk-secret-openai-key", stderr_text)
        self.assertNotIn("aws-session-secret", stderr_text)
        self.assertNotIn("openai-test-key", stderr_text)
        self.assertNotIn("secret-bearer-token", result.raw["native_executor_stderr"])
        self.assertNotIn("xoxb-123456789-secret", result.raw["native_executor_stderr"])
        self.assertTrue(any(event["event_type"] == "executor.result.persisted" for event in observer.events))
        self.assertTrue(any(event["event_type"] == "executor.result.loaded" for event in observer.events))

    def test_native_executor_requires_result_file_for_success(self) -> None:
        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                self.returncode = 0
                self.stdout = io.StringIO("")
                self.stderr = io.StringIO("")

            def poll(self):
                return 0

            def wait(self, timeout=None):
                return self.returncode

            def terminate(self):
                self.returncode = -15

            def kill(self):
                self.returncode = -9

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=FakePopen
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "trace_id": "trace-123",
                        "workflow_id": "wf-123",
                        "operation_id": "op-123",
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-123",
                        "memory_backend": "honcho",
                        "assistant_peer_id": "rsi:stage:prod",
                    }
                }
            )
            result = runtime._execute_native_workflow_task_request(
                task,
                runtime._resolve_tool_policy(task),
                observer=RecordingObserver(),
            )

        self.assertFalse(result.ok)
        self.assertIn("result file", result.message)

    def test_native_executor_result_stat_failure_does_not_break_loaded_result(self) -> None:
        structured = json.dumps(
            {
                "visible_reasoning": [],
                "reply_draft": "Draft reply",
                "final_answer": "Final reply",
                "confidence": 0.9,
                "context_summary": "Grounded context",
                "self_critique": "",
                "proposed_actions": [],
                "knowledge_drafts": [],
                "outcome_hypotheses": [],
                "produced_artifacts": [],
                "artifact_failure_reason": "",
            }
        )
        result_paths: list[Path] = []

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
                payload = {
                    "ok": True,
                    "response": structured,
                    "result": {"final_response": structured},
                    "session_id": "rsi-prod-conversation-123",
                }
                result_path = Path(request["result_path"])
                result_paths.append(result_path)
                result_path.write_text(json.dumps(payload, sort_keys=True), encoding="utf-8")
                self.returncode = 0
                self.stdout = io.StringIO("")
                self.stderr = io.StringIO("")

            def poll(self):
                return 0

            def wait(self, timeout=None):
                return self.returncode

            def terminate(self):
                self.returncode = -15

            def kill(self):
                self.returncode = -9

        original_stat = Path.stat
        stat_calls: dict[str, int] = {}

        def flaky_stat(path: Path, *args, **kwargs):
            path_key = str(Path(path))
            stat_calls[path_key] = stat_calls.get(path_key, 0) + 1
            if result_paths and Path(path) == result_paths[0] and stat_calls[path_key] >= 2:
                raise OSError("stat failed")
            return original_stat(path, *args, **kwargs)

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=FakePopen
        ), mock.patch(
            "pathlib.Path.stat", new=flaky_stat
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            observer = RecordingObserver()
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "trace_id": "trace-123",
                        "workflow_id": "wf-123",
                        "operation_id": "op-123",
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-123",
                        "memory_backend": "honcho",
                        "assistant_peer_id": "rsi:stage:prod",
                    }
                }
            )
            result = runtime._execute_native_workflow_task_request(
                task,
                runtime._resolve_tool_policy(task),
                observer=observer,
            )

        self.assertTrue(result.ok)
        persisted_events = [event for event in observer.events if event["event_type"] == "executor.result.persisted"]
        self.assertEqual(len(persisted_events), 1)
        self.assertIsNone(persisted_events[0]["payload"].get("bytes"))

    def test_native_executor_timeout_path_keeps_waits_bounded_after_failed_kill(self) -> None:
        wait_calls: list[object] = []

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                self.returncode = 0
                self.stdout = io.StringIO("")
                self.stderr = io.StringIO("")

            def poll(self):
                return None

            def wait(self, timeout=None):
                wait_calls.append(timeout)
                if timeout == 5:
                    raise subprocess.TimeoutExpired(cmd="worker", timeout=5)
                return self.returncode

            def terminate(self):
                self.returncode = -15

            def kill(self):
                raise OSError("kill failed")

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=FakePopen
        ), mock.patch(
            "rsi_runner.hermes_runtime.time.monotonic", side_effect=[0.0, 7.0]
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
                "RSI_RUNNER_PROD_TASK_TIMEOUT": "1s",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "trace_id": "trace-123",
                        "workflow_id": "wf-123",
                        "operation_id": "op-123",
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-123",
                        "memory_backend": "honcho",
                        "assistant_peer_id": "rsi:stage:prod",
                    }
                }
            )
            result = runtime._execute_native_workflow_task_request(
                task,
                runtime._resolve_tool_policy(task),
                observer=None,
            )

        self.assertFalse(result.ok)
        self.assertEqual(wait_calls, [5, 5])

    def test_native_executor_persists_run_dir_when_postprocess_raises(self) -> None:
        request_paths: list[Path] = []
        run_dir_exists_before_temp_cleanup = False

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request_path = Path(cmd[-1])
                request = json.loads(request_path.read_text(encoding="utf-8"))
                request_paths.append(request_path)
                Path(request["result_path"]).write_text(
                    json.dumps(
                        {
                            "ok": True,
                            "response": "{}",
                            "result": {"final_response": "{}"},
                            "session_id": "rsi-prod-conversation-123",
                        },
                        sort_keys=True,
                    ),
                    encoding="utf-8",
                )
                self.returncode = 0
                self.stdout = io.StringIO("")
                self.stderr = io.StringIO("")

            def poll(self):
                return 0

            def wait(self, timeout=None):
                return self.returncode

            def terminate(self):
                self.returncode = -15

            def kill(self):
                self.returncode = -9

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=FakePopen
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "trace_id": "trace-123",
                        "workflow_id": "wf-123",
                        "operation_id": "op-123",
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-123",
                        "memory_backend": "honcho",
                        "assistant_peer_id": "rsi:stage:prod",
                    }
                }
            )
            with mock.patch.object(
                runtime._session_manager, "finalize", side_effect=RuntimeError("finalize exploded")
            ):
                with self.assertRaisesRegex(RuntimeError, "finalize exploded"):
                    runtime._execute_native_workflow_task_request(
                        task,
                        runtime._resolve_tool_policy(task),
                        observer=RecordingObserver(),
                    )
            run_dir_exists_before_temp_cleanup = bool(request_paths and request_paths[0].parent.exists() and request_paths[0].exists())

        self.assertEqual(len(request_paths), 1)
        self.assertTrue(run_dir_exists_before_temp_cleanup)

    def test_execute_does_not_expand_skills_for_adhoc_prompts(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_skill_invocation_message",
            return_value="[INVOKED architecture-diagram]",
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_preloaded_skills_prompt",
            return_value=("[PRELOADED architecture-diagram]", ["architecture-diagram"], []),
        ), mock.patch(
            "rsi_runner.hermes_runtime.resolve_skill_command_key",
            return_value="/architecture-diagram",
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute("/architecture-diagram map the active depin services", system_message="System directive")

        self.assertTrue(result.ok)
        self.assertEqual(FakeAIAgent.last_prompt, "/architecture-diagram map the active depin services")
        self.assertNotIn("[INVOKED architecture-diagram]", FakeAIAgent.last_prompt)
        self.assertNotIn("[PRELOADED architecture-diagram]", FakeAIAgent.last_prompt)
        self.assertEqual(FakeAIAgent.last_system_message, "System directive")

    def test_runtime_reports_degraded_when_session_manager_unavailable(self) -> None:
        class BrokenSessionManager(FakeSessionManager):
            def __init__(self, config) -> None:
                super().__init__(config)
                self.ready_issues = ["session db unavailable"]
                self.available = False

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", BrokenSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertFalse(runtime.available)
        self.assertEqual(runtime.metadata["status"], "degraded")
        self.assertFalse(runtime.metadata["persistence_enabled"])

    def test_runtime_metadata_exposes_honcho_configuration(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {**runner_env("proposal"), "RSI_HONCHO_BASE_URL": "http://honcho.internal:8000"},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime.metadata["honcho_base_url"], "http://honcho.internal:8000")
        self.assertEqual(runtime.metadata["honcho_workspace"], "rsi-stage")
        self.assertEqual(runtime.metadata["honcho_environment"], "stage")
        self.assertEqual(runtime.metadata["honcho_recall_mode"], "hybrid")
        self.assertEqual(runtime.metadata["honcho_write_frequency"], "async")
        self.assertEqual(runtime.metadata["honcho_session_strategy"], "hybrid")
        self.assertEqual(runtime.metadata["honcho_ai_peer"], "rsi:stage:proposal")

    def test_runtime_metadata_exposes_skill_runtime_state(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime.metadata["skills_dir"], "/var/lib/hermes/skills")
        self.assertTrue(runtime.metadata["bundled_skills_available"])
        self.assertEqual(runtime.metadata["bundled_skills_sync_status"], "synced")
        self.assertEqual(runtime.metadata["hermes_config_parity_status"], "configured")
        self.assertTrue(runtime.metadata["observation_sink_configured"])
        self.assertEqual(runtime.metadata["observation_sink_status"], "configured")
        self.assertTrue(runtime.metadata["direct_delivery_phase_enabled"])

    def test_skill_mentions_detect_leading_slash_command(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(
            runtime._skill_mentions_from_text("/architecture-diagram map the active depin services"),
            ["architecture-diagram"],
        )

    def test_workflow_task_expands_explicit_slash_skill_before_render(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "/architecture-diagram map the active depin services",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-001",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.resolve_skill_command_key", return_value="/architecture-diagram"), mock.patch(
            "rsi_runner.hermes_runtime.build_skill_invocation_message",
            return_value="[INVOKED architecture-diagram]",
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_preloaded_skills_prompt",
            return_value=("", [], []),
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertIn("[INVOKED architecture-diagram]", FakeAIAgent.last_prompt)
        self.assertEqual(result.raw["requested_skills"], ["architecture-diagram"])
        self.assertEqual(result.raw["resolved_skills"], ["architecture-diagram"])
        self.assertEqual(result.raw["missing_skills"], [])
        self.assertEqual(result.raw["skill_injection_mode"], "slash_command")

    def test_workflow_task_preloads_inline_and_requested_skills_once(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "User request: Please use /architecture-diagram skill for this system diagram.\n\nInvestigate within the governed tool boundary.",
                    "requested_skills": ["architecture-diagram"],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-001",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.resolve_skill_command_key", return_value="/architecture-diagram"), mock.patch(
            "rsi_runner.hermes_runtime.build_skill_invocation_message",
            return_value=None,
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_preloaded_skills_prompt",
            return_value=("[PRELOADED architecture-diagram]", ["architecture-diagram"], []),
        ) as preload_mock, mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(preload_mock.call_args.args[0], ["architecture-diagram"])
        self.assertIn("[PRELOADED architecture-diagram]", FakeAIAgent.last_prompt)
        self.assertEqual(result.raw["requested_skills"], ["architecture-diagram"])
        self.assertEqual(result.raw["resolved_skills"], ["architecture-diagram"])
        self.assertEqual(result.raw["missing_skills"], [])
        self.assertEqual(result.raw["skill_injection_mode"], "preloaded")

    def test_action_contract_repair_does_not_reexpand_skills_or_rerender_prompt(self) -> None:
        class ActionRepairAIAgent(FakeAIAgent):
            calls = 0
            prompts: list[str] = []

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict[str, object]:
                type(self).calls += 1
                type(self).prompts.append(prompt)
                type(self).last_prompt = prompt
                type(self).last_system_message = system_message
                type(self).last_history = conversation_history or []
                payload = {
                    "visible_reasoning": [],
                    "reply_draft": "Draft reply",
                    "final_answer": "Final reply",
                    "confidence": 0.91,
                    "context_summary": "Repo and KB context collected.",
                    "self_critique": "",
                    "knowledge_drafts": [],
                    "outcome_hypotheses": [],
                    "produced_artifacts": [],
                    "artifact_failure_reason": "",
                }
                if type(self).calls == 1:
                    payload["proposed_actions"] = []
                else:
                    payload["proposed_actions"] = [
                        {
                            "kind": "slack_post",
                            "target_ref": "slack:thread",
                            "request_payload": {"body": "Final reply"},
                            "approval_mode": "deterministic",
                            "idempotency_key": "reply-1",
                            "rationale": "Reply in thread",
                            "evidence_refs": [],
                        }
                    ]
                return {"final_response": json.dumps(payload)}

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "User request: Please use /architecture-diagram skill and summarize the workflow.\n\nInvestigate within the governed tool boundary.",
                    "requested_skills": ["architecture-diagram"],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", ActionRepairAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_skill_invocation_message",
            return_value=None,
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_preloaded_skills_prompt",
            return_value=("[PRELOADED architecture-diagram]", ["architecture-diagram"], []),
        ), mock.patch(
            "rsi_runner.hermes_runtime.resolve_skill_command_key",
            return_value="/architecture-diagram",
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertTrue(result.raw["action_contract_repair_attempted"])
        self.assertTrue(result.raw["action_contract_repair_succeeded"])
        self.assertEqual(ActionRepairAIAgent.prompts[1].count("Runner role:"), 1)
        self.assertEqual(ActionRepairAIAgent.prompts[1].count("[PRELOADED architecture-diagram]"), 1)

    def test_missing_requested_skill_is_recorded_without_failing_workflow(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "User request: Please use /missing-skill.\n\nInvestigate within the governed tool boundary.",
                    "requested_skills": ["missing-skill"],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-001",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.build_skill_invocation_message",
            return_value=None,
        ), mock.patch("rsi_runner.hermes_runtime.resolve_skill_command_key", return_value=None), mock.patch(
            "rsi_runner.hermes_runtime.build_preloaded_skills_prompt",
            return_value=("", [], ["missing-skill"]),
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["requested_skills"], ["missing-skill"])
        self.assertEqual(result.raw["resolved_skills"], [])
        self.assertEqual(result.raw["missing_skills"], ["missing-skill"])
        self.assertEqual(result.raw["skill_injection_mode"], "preloaded")

    def test_proposal_role_enforces_read_only_tool_policy(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "proposal",
                    "repo": "rsi-agent-platform",
                    "prompt": "Produce a fix plan.",
                    "allowed_tools": ["repo.context", "github.create_pr", "rsi.candidate_context", "honcho_conclude"],
                    "tool_allowlist": ["repo.context", "github.create_pr", "rsi.candidate_context", "honcho_conclude"],
                    "session_scope_kind": "proposal_candidate",
                    "session_scope_id": "shared-store:pk-collision",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:proposal",
                    "user_peer_id": "operator:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("proposal"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(FakeAIAgent.last_kwargs["max_iterations"], 5)
        self.assertIn("repo_context", FakeAIAgent.last_valid_tool_names)
        self.assertIn("rsi_candidate_context", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("github.create_pr", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("honcho_conclude", FakeAIAgent.last_valid_tool_names)
        self.assertEqual(result.raw["tool_policy_mode"], "enforced_read_only")
        self.assertIn("github.create_pr", result.raw["blocked_tool_names"])
        self.assertIn("honcho_conclude", result.raw["blocked_tool_names"])
        self.assertIn("repo.context", result.raw["tool_allowlist_effective"])
        self.assertIn("rsi.candidate_context", result.raw["tool_allowlist_effective"])
        self.assertIn("repo_context", result.raw["tool_transport_allowlist_effective"])
        self.assertIn("rsi_candidate_context", result.raw["tool_transport_allowlist_effective"])

    def test_proposal_role_keeps_helper_toolsets_for_governed_tasks(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "proposal",
                    "repo": "rsi-agent-platform",
                    "prompt": "Produce a fix plan.",
                    "execution_mode": "diagnose",
                    "session_scope_kind": "proposal_candidate",
                    "session_scope_id": "shared-store:pk-collision",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:proposal",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {**runner_env("proposal"), "RSI_HERMES_NATIVE_GOVERNED_TOOLS_ENABLED": "true"},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        toolsets = runtime._native_toolsets_for_task(task)

        self.assertIn("todo", toolsets)
        self.assertIn("session_search", toolsets)

    def test_workflow_toolsets_do_not_duplicate_governed_entries(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Investigate the active workflow.",
                    "execution_mode": "implement",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {**runner_env("proposal"), "RSI_TOOL_GATEWAY_BASE_URL": "http://tool-gateway.test", "RSI_HERMES_NATIVE_GOVERNED_TOOLS_ENABLED": "true"},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        toolsets = runtime._native_toolsets_for_task(task)

        self.assertEqual(toolsets.count("rsi-governed-readonly"), 1)
        self.assertEqual(toolsets.count("rsi-governed-workspace"), 1)

    def test_read_only_workflow_toolsets_do_not_include_workspace_writes(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Investigate the active workflow.",
                    "execution_mode": "diagnose",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "RSI_TOOL_GATEWAY_BASE_URL": "http://tool-gateway.test"},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        toolsets = runtime._native_toolsets_for_task(task)

        self.assertEqual(toolsets.count("rsi-governed-readonly"), 1)
        self.assertEqual(toolsets.count("rsi-governed-workspace"), 0)

    def test_prod_role_uses_governed_read_only_tool_policy(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the active workflow.",
                    "allowed_tools": ["repo.context", "knowledge.context", "github.create_pr"],
                    "tool_allowlist": ["repo.context", "knowledge.context", "github.create_pr"],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(FakeAIAgent.last_kwargs["max_iterations"], 20)
        self.assertIn("repo_context", FakeAIAgent.last_valid_tool_names)
        self.assertIn("knowledge_context", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("github.create_pr", FakeAIAgent.last_valid_tool_names)
        self.assertEqual(result.raw["tool_policy_mode"], "enforced_read_only")
        self.assertEqual(result.raw["task_timeout_seconds"], 900)
        self.assertEqual(result.raw["transport_timeout_seconds"], 930)
        self.assertIn("github.create_pr", result.raw["blocked_tool_names"])
        self.assertIn("repo_context", result.raw["tool_transport_allowlist_effective"])
        self.assertIn("knowledge_context", result.raw["tool_transport_allowlist_effective"])

    def test_prod_role_with_bound_workspace_admits_read_only_workspace_tools(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Trace the workspace history for this file.",
                    "allowed_tools": ["workspace.git_history", "workspace.read_file", "workspace.write_file"],
                    "tool_allowlist": ["workspace.git_history", "workspace.read_file", "workspace.write_file"],
                    "workspace_id": "workspace-123",
                    "attempt_id": "attempt-123",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertIn("workspace_git_history", FakeAIAgent.last_valid_tool_names)
        self.assertIn("workspace_read_file", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("workspace_write_file", FakeAIAgent.last_valid_tool_names)
        self.assertIn("workspace.git_history", result.raw["tool_allowlist_effective"])
        self.assertIn("workspace.read_file", result.raw["tool_allowlist_effective"])
        self.assertIn("workspace.write_file", result.raw["blocked_tool_names"])

    def test_question_reduce_uses_direct_responses_reducer_without_hermes_loop(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "question_reduce",
                    "repo": "depin-backend",
                    "prompt": json.dumps(
                        {
                            "investigation_spec": {
                                "user_request": "How did depin-backend API progress this week for the numo project?",
                                "repo": "depin-backend",
                                "project_key": "numo",
                            },
                            "evidence_ledger": {
                                "user_request": "How did depin-backend API progress this week for the numo project?",
                                "reply_target": {"channel_id": "C123", "thread_ts": "171000001.000100"},
                                "evidence_items": [
                                    {
                                        "kind": "slack_message",
                                        "summary": "[alice] Merged the pagination cleanup PR.",
                                        "source_ref": "https://slack.example/messages/1",
                                        "tool_name": "slack.history",
                                    }
                                ],
                            },
                            "runner_diagnostics": {
                                "termination_reason": "task_timeout",
                            },
                        }
                    ),
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        captured_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            captured_requests.append(
                {
                    "url": req.full_url,
                    "timeout": timeout,
                    "body": json.loads(req.data.decode("utf-8")),
                }
            )
            return FakeHTTPResponse(
                {
                    "id": "resp_question_reduce_1",
                    "output_text": json.dumps(
                        {
                            "reply_markdown": "Partial rundown: pagination cleanup landed, but the weekly picture is incomplete.",
                            "confidence": 0.68,
                            "alignment_degraded": True,
                            "alignment_notice": "NUMO alignment is degraded because no fresh canonical project ledger was available.",
                        }
                    ),
                }
            )

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch(
            "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen
        ), mock.patch.dict(
            os.environ, {**runner_env("prod"), "HERMES_HOME": tempdir}, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)
            log_path = result.raw["native_execution_log_path"]
            self.assertTrue(os.path.exists(log_path))
            with open(log_path, encoding="utf-8") as handle:
                events = [json.loads(line) for line in handle if line.strip()]

        self.assertTrue(result.ok)
        self.assertIsNone(FakeAIAgent.last_kwargs)
        self.assertEqual(len(captured_requests), 1)
        self.assertEqual(captured_requests[0]["url"], "https://api.openai.com/v1/responses")
        self.assertNotIn("tools", captured_requests[0]["body"])
        self.assertEqual(result.raw["question_reduce_mode"], "direct_responses_api")
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["termination_reason"], "task_timeout")
        self.assertEqual(result.raw["structured_output"]["reply_markdown"], "Partial rundown: pagination cleanup landed, but the weekly picture is incomplete.")
        self.assertEqual(events[0]["event"], "execution_started")
        self.assertEqual(events[1]["event"], "direct_response_request")
        self.assertEqual(events[2]["event"], "direct_response_response")
        self.assertEqual(events[-1]["event"], "execution_completed")
        self.assertEqual(events[-1]["termination_reason"], "task_timeout")

    def test_question_gather_with_mcp_uses_hermes_loop(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "question_gather",
                    "repo": "depin-backend",
                    "prompt": "Did the linked Slack thread confirm the upload fix?",
                    "system_message": "Use read-only tools and return JSON only.",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_read"}],
                }
            }
        )
        registration = TaskScopedMCPRegistration(
            servers=[
                TaskScopedMCPServer(
                    source_label="slack",
                    profile="slack_mcp_read",
                    server_name="rsi-task-trace-123-0-slack-abc",
                    toolset_alias="mcp-rsi-task-trace-123-0-slack-abc",
                    included_tool_names=["get_thread", "search_messages"],
                    hermes_config={},
                )
            ]
        )
        cleanup = TaskScopedMCPCleanupResult(
            status="cleaned",
            cleaned_server_names=registration.server_names,
            failed_server_names=[],
            errors=[],
        )

        def fake_cleanup(_registration: TaskScopedMCPRegistration) -> TaskScopedMCPCleanupResult:
            _registration.cleanup_result = cleanup
            return cleanup

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime._mcp_adapter, "register_task_servers", return_value=registration) as register_mock, mock.patch.object(
                runtime._mcp_adapter, "cleanup_registration", side_effect=fake_cleanup
            ) as cleanup_mock, mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen") as urlopen:
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(register_mock.call_count, 1)
        self.assertEqual(cleanup_mock.call_count, 1)
        self.assertEqual(urlopen.call_count, 0)
        self.assertEqual(
            FakeAIAgent.last_kwargs["enabled_toolsets"],
            ["todo", "session_search", "mcp-rsi-task-trace-123-0-slack-abc"],
        )
        self.assertTrue(result.raw["agentic_mcp_enabled"])
        self.assertEqual(result.raw["agentic_mcp_server_names"], ["rsi-task-trace-123-0-slack-abc"])
        self.assertEqual(result.raw["runner_diagnostics"]["agentic_mcp_cleanup_status"], "cleaned")

    def test_question_reduce_defaults_partial_for_output_token_budget_exhaustion(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "question_reduce",
                    "repo": "depin-backend",
                    "prompt": json.dumps(
                        {
                            "investigation_spec": {
                                "user_request": "Did the linked Slack thread confirm the upload fix?",
                                "repo": "depin-backend",
                            },
                            "evidence_ledger": {
                                "termination_reason": "output_token_budget_exhausted",
                            },
                            "runner_diagnostics": {
                                "termination_reason": "output_token_budget_exhausted",
                            },
                        }
                    ),
                }
            }
        )

        with mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        verdict, termination_reason = runtime._question_reduce_defaults(task)

        self.assertEqual(verdict, "partial")
        self.assertEqual(termination_reason, "output_token_budget_exhausted")

    def test_transport_tool_schemas_are_openai_strict(self) -> None:
        def strict_schema_violations(node: object, path: str) -> list[str]:
            violations: list[str] = []
            if isinstance(node, dict):
                if node.get("type") == "object":
                    properties = node.get("properties") or {}
                    if node.get("additionalProperties") is not False:
                        violations.append(f"{path}:additionalProperties")
                    required = node.get("required")
                    if not isinstance(required, list):
                        violations.append(f"{path}:required")
                    else:
                        missing = [key for key in properties if key not in required]
                        if missing:
                            violations.append(f"{path}:missing_required={','.join(missing)}")
                for key, value in node.items():
                    violations.extend(strict_schema_violations(value, f"{path}.{key}"))
                return violations
            if isinstance(node, list):
                for idx, value in enumerate(node):
                    violations.extend(strict_schema_violations(value, f"{path}[{idx}]"))
            return violations

        cloudflare = transport_tool_schema("cloudflare.inspect")
        self.assertEqual(cloudflare["parameters"]["required"], ["resource"])
        self.assertEqual(cloudflare["parameters"]["properties"]["resource"]["type"], ["string", "null"])
        self.assertFalse(cloudflare["parameters"]["additionalProperties"])

        repo_read_file = transport_tool_schema("repo.read_file")
        self.assertEqual(repo_read_file["parameters"]["required"], ["repo", "path", "ref"])
        self.assertEqual(repo_read_file["parameters"]["properties"]["repo"]["type"], ["string", "null"])
        self.assertEqual(repo_read_file["parameters"]["properties"]["path"]["type"], "string")
        self.assertEqual(repo_read_file["parameters"]["properties"]["ref"]["type"], ["string", "null"])

        runtime_config = transport_tool_schema("rsi.runtime_config")
        self.assertEqual(runtime_config["parameters"]["required"], [])

        invalid: dict[str, list[str]] = {}
        for item in governed_toolset_definitions():
            schema = item["schema"]
            paths = strict_schema_violations(schema.get("parameters"), "parameters")
            if paths:
                invalid[schema["name"]] = paths
        self.assertEqual(invalid, {})

    def test_default_policy_allowlist_includes_slack_upload_file_when_tool_gateway_enabled(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        allowlist = runtime._default_policy_allowlist(execution_mode="")

        self.assertIn("slack.upload_file", allowlist)

    def test_workflow_with_mcp_routes_through_hermes_loop_and_records_agentic_diagnostics(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest Slack thread and reply in-thread.",
                    "allowed_tools": ["repo.context", "slack.history"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-workflow-123",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_reply"}],
                }
            }
        )
        registration = TaskScopedMCPRegistration(
            servers=[
                TaskScopedMCPServer(
                    source_label="slack",
                    profile="slack_mcp_reply",
                    server_name="rsi-task-trace-workflow-123-0-slack-abc",
                    toolset_alias="mcp-rsi-task-trace-workflow-123-0-slack-abc",
                    included_tool_names=["get_thread", "search_messages", "send_message"],
                    hermes_config={},
                )
            ]
        )
        cleanup = TaskScopedMCPCleanupResult(
            status="cleaned",
            cleaned_server_names=registration.server_names,
            failed_server_names=[],
            errors=[],
        )

        def fake_cleanup(_registration: TaskScopedMCPRegistration) -> TaskScopedMCPCleanupResult:
            _registration.cleanup_result = cleanup
            return cleanup

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime._mcp_adapter, "register_task_servers", return_value=registration) as register_mock, mock.patch.object(
                runtime._mcp_adapter, "cleanup_registration", side_effect=fake_cleanup
            ) as cleanup_mock, mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen") as urlopen:
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertGreaterEqual(register_mock.call_count, 1)
        self.assertGreaterEqual(cleanup_mock.call_count, 1)
        self.assertEqual(urlopen.call_count, 0)
        self.assertEqual(
            FakeAIAgent.last_kwargs["enabled_toolsets"],
            [
                "todo",
                "session_search",
                "rsi-governed-readonly",
                "mcp-rsi-task-trace-workflow-123-0-slack-abc",
            ],
        )
        self.assertTrue(result.raw["agentic_mcp_enabled"])
        self.assertEqual(result.raw["agentic_mcp_server_names"], ["rsi-task-trace-workflow-123-0-slack-abc"])
        self.assertEqual(
            result.raw["runner_diagnostics"]["agentic_mcp_toolsets"],
            ["mcp-rsi-task-trace-workflow-123-0-slack-abc"],
        )
        self.assertEqual(result.raw["runner_diagnostics"]["agentic_mcp_cleanup_status"], "cleaned")
        self.assertNotIn("native_execution_mode", result.raw["runner_diagnostics"])

    def test_workflow_with_mcp_registration_failure_returns_specific_runner_error(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply in Slack.",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_reply"}],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(
                runtime._mcp_adapter,
                "register_task_servers",
                side_effect=RuntimeError("expected exactly one candidate"),
            ), mock.patch.object(runtime, "_create_agent", side_effect=AssertionError("agent should not be created")):
                result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_non_ok")
        self.assertEqual(result.raw["runner_diagnostics"]["failure_kind"], "agentic_mcp_registration_failed")
        self.assertEqual(result.raw["runner_diagnostics"]["agentic_mcp_cleanup_status"], "not_needed")
        self.assertIn("expected exactly one candidate", result.message)

    def test_workflow_with_mcp_cleanup_runs_on_agent_error(self) -> None:
        class FailingAIAgent(FakeAIAgent):
            def run_conversation(self, *args, **kwargs) -> dict[str, object]:
                raise RuntimeError("boom")

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest Slack thread and reply in-thread.",
                    "mcp_servers": [{"server_label": "notion", "profile": "notion_mcp_read"}],
                }
            }
        )
        registration = TaskScopedMCPRegistration(
            servers=[
                TaskScopedMCPServer(
                    source_label="notion",
                    profile="notion_mcp_read",
                    server_name="rsi-task-trace-workflow-456-0-notion-abc",
                    toolset_alias="mcp-rsi-task-trace-workflow-456-0-notion-abc",
                    included_tool_names=["search", "fetch"],
                    hermes_config={},
                )
            ]
        )
        cleanup = TaskScopedMCPCleanupResult(
            status="cleaned",
            cleaned_server_names=registration.server_names,
            failed_server_names=[],
            errors=[],
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FailingAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime._mcp_adapter, "register_task_servers", return_value=registration), mock.patch.object(
                runtime._mcp_adapter, "cleanup_registration", return_value=cleanup
            ) as cleanup_mock:
                result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(cleanup_mock.call_count, 1)
        self.assertEqual(result.raw["runner_diagnostics"]["agentic_mcp_cleanup_status"], "cleaned")
        self.assertTrue(result.raw["agentic_mcp_enabled"])
        self.assertIn("boom", result.message)

    def test_workflow_attach_tool_policy_preserves_task_scoped_mcp_and_helper_tools(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Investigate and reply.",
                    "allowed_tools": ["repo.context", "slack.upload_file"],
                    "reply_delivery_mode": "direct",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            tool_policy = runtime._resolve_tool_policy(task)

        agent = types.SimpleNamespace(
            tools=[
                {"type": "function", "function": {"name": "search_messages"}},
                {"type": "function", "function": {"name": "send_message"}},
                {"type": "function", "function": {"name": "todo_write"}},
            ],
            valid_tool_names={"search_messages", "send_message", "todo_write"},
            _memory_manager=None,
        )
        runtime._attach_tool_policy(agent, task, tool_policy)

        tool_names = sorted(tool["function"]["name"] for tool in agent.tools)
        self.assertIn("search_messages", tool_names)
        self.assertIn("send_message", tool_names)
        self.assertIn("slack_upload_file", tool_names)
        self.assertIn("todo_write", tool_names)
        self.assertIn("repo_context", agent.valid_tool_names)
        self.assertIn("search_messages", agent.valid_tool_names)
        self.assertIn("send_message", agent.valid_tool_names)
        self.assertIn("slack_upload_file", agent.valid_tool_names)
        self.assertIn("todo_write", agent.valid_tool_names)

    def test_task_scoped_mcp_adapter_fails_closed_for_custom_read_only_server_without_tool_names(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Read docs.",
                    "mcp_servers": [
                        {
                            "server_label": "docs",
                            "server_url": "https://developers.openai.com/mcp",
                            "allowed_tools": {"read_only": True},
                        }
                    ],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with self.assertRaisesRegex(RuntimeError, "refusing to expose the full server"):
                runtime._mcp_adapter._translate_task_servers(task)

    def test_question_expand_with_mcp_uses_hermes_loop(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "question_expand",
                    "repo": "depin-backend",
                    "prompt": "Investigate the linked thread.",
                    "system_message": "Return only JSON.",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-expand-123",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_reply"}],
                }
            }
        )
        registration = TaskScopedMCPRegistration(
            servers=[
                TaskScopedMCPServer(
                    source_label="slack",
                    profile="slack_mcp_reply",
                    server_name="rsi-task-trace-expand-123-0-slack-abc",
                    toolset_alias="mcp-rsi-task-trace-expand-123-0-slack-abc",
                    included_tool_names=["get_thread", "send_message"],
                    hermes_config={},
                )
            ]
        )
        cleanup = TaskScopedMCPCleanupResult(
            status="cleaned",
            cleaned_server_names=registration.server_names,
            failed_server_names=[],
            errors=[],
        )

        def fake_cleanup(_registration: TaskScopedMCPRegistration) -> TaskScopedMCPCleanupResult:
            _registration.cleanup_result = cleanup
            return cleanup

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime._mcp_adapter, "register_task_servers", return_value=registration), mock.patch.object(
                runtime._mcp_adapter, "cleanup_registration", side_effect=fake_cleanup
            ):
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(
            FakeAIAgent.last_kwargs["enabled_toolsets"],
            ["todo", "session_search", "mcp-rsi-task-trace-expand-123-0-slack-abc"],
        )
        self.assertTrue(result.raw["agentic_mcp_enabled"])
        self.assertEqual(result.raw["runner_diagnostics"]["agentic_mcp_cleanup_status"], "cleaned")

    def test_question_expand_query_hints_prevent_prompt_blob_leakage_into_tool_defaults(self) -> None:
        captured: dict[str, object] = {}

        class FakeResponse:
            def __enter__(self):
                return self

            def __exit__(self, exc_type, exc, tb) -> bool:
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Slack thread history loaded.",
                        "provider": "slack",
                        "provider_ref": "171000001.000100",
                        "output": {"messages": []},
                    }
                ).encode("utf-8")

        def fake_tool_urlopen(req, timeout: int = 0):
            captured["body"] = json.loads(req.data.decode("utf-8"))
            return FakeResponse()

        binding = ReadOnlyToolBinding(
            base_url="http://tool-gateway.internal",
            allowed_tool_names=["slack.history", "slack.search", "knowledge.context", "repo.context"],
            task_repo="depin-backend",
            task_repo_ref="main",
            task_prompt=json.dumps(
                {
                    "investigation_spec": {
                        "user_request": "How did depin-backend API progress this week for the numo project?",
                        "repo": "depin-backend",
                        "project_key": "numo",
                    },
                    "evidence_ledger": {"open_questions": ["Need better Slack evidence."]},
                }
            ),
            task_channel_id="C123",
            task_thread_ts="171000001.000100",
            task_context_summary="read-heavy qna",
            trace_id="trace-123",
            session_scope_kind="conversation",
            session_scope_id="conv-123",
            context_refs=[],
            default_question="How did depin-backend API progress this week for the numo project?",
            repo_question="How did depin-backend API progress this week for the numo project?",
            knowledge_topic="numo",
            knowledge_question="What are the current goals, constraints, and expected outcomes for numo?",
            slack_history_focus="Extract the most relevant messages for answering: How did depin-backend API progress this week for the numo project?",
            slack_search_query="depin-backend numo",
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_tool_urlopen):
            _ = json.loads(binding.handle_tool_call("slack_history", {}))
        self.assertEqual(
            captured["body"]["question"],
            "Extract the most relevant messages for answering: How did depin-backend API progress this week for the numo project?",
        )

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_tool_urlopen):
            _ = json.loads(binding.handle_tool_call("slack_search", {}))
        self.assertEqual(captured["body"]["query"], "depin-backend numo")

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_tool_urlopen):
            _ = json.loads(binding.handle_tool_call("knowledge_context", {}))
        self.assertEqual(captured["body"]["topic"], "numo")
        self.assertEqual(captured["body"]["question"], "What are the current goals, constraints, and expected outcomes for numo?")

        with mock.patch("rsi_runner.rsi_tools.urlrequest.urlopen", side_effect=fake_tool_urlopen):
            _ = json.loads(binding.handle_tool_call("repo_context", {}))
        self.assertEqual(captured["body"]["question"], "How did depin-backend API progress this week for the numo project?")

    def test_proactive_role_accepts_workflow_task_type(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the proactive workflow.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:proactive",
                    "user_peer_id": "user:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("proactive"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["tool_policy_mode"], "enforced_read_only")

    def test_eval_task_timeout_returns_structured_timeout(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "eval",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the eval.",
                    "timeout_seconds": 1,
                    "session_scope_kind": "eval_line",
                    "session_scope_id": "shared-store:pk-collision",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:eval",
                    "user_peer_id": "operator:alice",
                }
            }
        )
        FakeAIAgent.sleep_seconds = 1.2
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, {**runner_env("eval"), "RSI_RUNNER_EVAL_TASK_TIMEOUT": "1s"}, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertIn("timed out", result.message)
        self.assertEqual(result.raw["timeout_kind"], "task_timeout")
        self.assertEqual(result.raw["task_timeout_seconds"], 1)
        self.assertEqual(result.raw["transport_timeout_seconds"], 330)
        self.assertEqual(result.raw["tool_policy_mode"], "enforced_read_only")
        self.assertEqual(result.raw["failure_class"], "runner_transport_timeout")
        self.assertEqual(result.raw["runner_diagnostics"]["failure_kind"], "transport_timeout")
        self.assertEqual(result.raw["runner_diagnostics"]["timeout_kind"], "task_timeout")
        self.assertEqual(result.raw["runner_diagnostics"]["termination_reason"], "task_timeout")
        self.assertEqual(result.raw["runner_diagnostics"]["last_activity_desc"], "waiting_on_tool_or_model")
        self.assertEqual(result.raw["runner_diagnostics"]["current_tool"], "rsi.trace_context")
        self.assertEqual(result.raw["runner_diagnostics"]["api_call_count"], 2)
        self.assertEqual(result.raw["runner_diagnostics"]["budget_used"], 1)

    def test_eval_inactivity_timeout_returns_structured_timeout(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "eval",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the eval.",
                    "timeout_seconds": 10,
                    "session_scope_kind": "eval_line",
                    "session_scope_id": "shared-store:pk-collision",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:eval",
                    "user_peer_id": "operator:alice",
                }
            }
        )
        FakeAIAgent.sleep_seconds = 1.2
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {**runner_env("eval"), "RSI_RUNNER_EVAL_TASK_TIMEOUT": "10s", "RSI_RUNNER_EVAL_INACTIVITY_TIMEOUT": "1s"},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertIn("inactivity timeout", result.message)
        self.assertEqual(result.raw["timeout_kind"], "inactivity_timeout")
        self.assertEqual(result.raw["inactivity_timeout_seconds"], 1)

    def test_prod_iteration_budget_exhaustion_returns_structured_failure_when_partial_recovery_is_ineligible(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the active workflow.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        FakeAIAgent.sleep_seconds = 1.2
        FakeAIAgent.budget_used = 20
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertIn("exhausted", result.message)
        self.assertEqual(result.raw["failure_class"], "runner_iteration_budget_exhausted")
        self.assertEqual(result.raw["termination_reason"], "iteration_budget_exhausted")
        self.assertTrue(result.raw["max_iterations_reached"])
        self.assertEqual(result.raw["runner_diagnostics"]["failure_kind"], "iteration_budget_exhausted")
        self.assertEqual(result.raw["runner_diagnostics"]["termination_reason"], "iteration_budget_exhausted")
        self.assertEqual(result.raw["runner_diagnostics"]["budget_used"], 20)
        self.assertEqual(result.raw["runner_diagnostics"]["budget_max"], 20)
        self.assertEqual(FakeAIAgent.last_interrupt_message, "runner iteration_budget_exhausted")

    def test_workflow_iteration_budget_exhaustion_recovers_partial_completion_without_tools(self) -> None:
        class BudgetReducerAIAgent:
            init_history: list[dict[str, object]] = []
            run_history: list[dict[str, object]] = []
            interrupt_messages: list[str] = []
            created_instances = 0

            def __init__(self, **kwargs) -> None:
                type(self).created_instances += 1
                self.instance_index = type(self).created_instances
                type(self).init_history.append(dict(kwargs))

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict[str, object]:
                type(self).run_history.append(
                    {
                        "instance_index": self.instance_index,
                        "prompt": prompt,
                        "system_message": system_message,
                        "history": list(conversation_history or []),
                        "valid_tool_names": sorted(getattr(self, "valid_tool_names", [])),
                        "tool_names": sorted(
                            tool["function"]["name"]
                            for tool in list(getattr(self, "tools", []) or [])
                            if isinstance(tool, dict) and isinstance(tool.get("function"), dict) and tool["function"].get("name")
                        ),
                        "task_id": task_id,
                    }
                )
                time.sleep(0.6)
                return {"final_response": ""}

            def interrupt(self, message: str | None = None) -> None:
                type(self).interrupt_messages.append(message or "")

            def get_activity_summary(self) -> dict[str, object]:
                budget_max = int(type(self).init_history[self.instance_index - 1].get("max_iterations", 1))
                if self.instance_index == 1:
                    return {
                        "last_activity_desc": "iteration budget exhausted",
                        "current_tool": "slack_history",
                        "api_call_count": budget_max,
                        "budget_used": budget_max,
                        "budget_max": budget_max,
                    }
                return {
                    "last_activity_desc": "composing partial reply",
                    "current_tool": "",
                    "api_call_count": 0,
                    "budget_used": 0,
                    "budget_max": budget_max,
                }

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the Slack thread.",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        captured_requests: list[dict[str, object]] = []
        observed = {
            "candidate_read_surfaces": [{"channel_id": "C123", "thread_ts": "171000001.000100", "ref": "", "source": "task_binding"}],
            "selected_context_surfaces": [{"channel_id": "C123", "thread_ts": "171000001.000100", "scope": "bound_thread"}],
            "memory_warnings": [],
            "tool_calls": [
                {
                    "tool_name": "slack.history",
                    "tool_call_id": "slack.history:1",
                    "request": {"channel_id": "C123", "thread_ts": "171000001.000100"},
                    "summary": "Loaded the bound Slack thread.",
                    "status": "completed",
                    "provider_ref": "slack-history-1",
                },
                {
                    "tool_name": "slack.search",
                    "tool_call_id": "slack.search:2",
                    "request": {"query": "depin backend api numo"},
                    "summary": "slack.search is unavailable in this workflow because the bound Slack event has no action_token.",
                    "status": "error",
                },
            ],
            "evidence_items": [
                {
                    "kind": "slack_message",
                    "summary": "The thread discussed depin-backend API progress for the numo project.",
                    "source_ref": "https://slack.example/messages/1",
                    "tool_name": "slack.history",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "permalink": "https://slack.example/messages/1",
                }
            ],
        }

        def fake_urlopen(req, timeout: int = 0):
            captured_requests.append(
                {
                    "url": req.full_url,
                    "timeout": timeout,
                    "body": json.loads(req.data.decode("utf-8")),
                }
            )
            return FakeHTTPResponse(
                {
                    "id": "resp_partial_1",
                    "output_text": json.dumps(
                        partial_structured_output(reply_text="Partial answer: grounded summary so far.", proposed_actions=[])
                    ),
                }
            )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", BudgetReducerAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen), mock.patch.object(
            HermesRuntime, "_observability_metadata", return_value=observed
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["termination_reason"], "iteration_budget_exhausted")
        self.assertTrue(result.raw["max_iterations_reached"])
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["runner_diagnostics"]["completion_verdict"], "partial")
        self.assertEqual(result.raw["runner_diagnostics"]["budget_used"], 20)
        self.assertEqual(result.raw["runner_diagnostics"]["budget_max"], 20)
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "direct_reducer")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_attempts"], 1)
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_attempted"])
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_succeeded"])
        self.assertEqual(BudgetReducerAIAgent.created_instances, 1)
        self.assertEqual(BudgetReducerAIAgent.init_history[0]["max_iterations"], 20)
        self.assertIn("repo_context", BudgetReducerAIAgent.run_history[0]["valid_tool_names"])
        self.assertTrue(any("iteration_budget_exhausted" in message for message in BudgetReducerAIAgent.interrupt_messages))
        self.assertEqual(len(captured_requests), 1)
        self.assertEqual(captured_requests[0]["url"], "https://api.openai.com/v1/responses")
        self.assertEqual(captured_requests[0]["timeout"], 180)
        self.assertNotIn("tools", captured_requests[0]["body"])
        self.assertIn('"termination_reason": "iteration_budget_exhausted"', str(captured_requests[0]["body"]["input"]))
        self.assertNotIn("Earlier thread message", str(captured_requests[0]["body"]["input"]))
        self.assertEqual(result.raw["structured_output"]["proposed_actions"][0]["kind"], "slack_post")
        self.assertTrue(result.raw["action_contract_repair_attempted"])
        self.assertTrue(result.raw["action_contract_repair_succeeded"])
        self.assertEqual(result.raw["evidence_ledger"]["tool_calls"][0]["tool_name"], "slack.history")
        self.assertEqual(len(result.raw["evidence_ledger"]["evidence_items"]), 1)
        self.assertTrue(result.raw["evidence_ledger"]["open_questions"])

    def test_partial_completion_default_reserve_uses_single_three_minute_reducer_window(self) -> None:
        with mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime._partial_completion_finalization_reserve_seconds(300), 180)
        self.assertEqual(runtime._partial_completion_attempt_budgets(180), [180])

    def test_workflow_direct_reply_delivery_uses_session_delta_without_action_repair(self) -> None:
        class DirectReplySessionManager(FakeSessionManager):
            def finalize(self, context: types.SimpleNamespace, tracker: FakeTracker) -> dict[str, object]:
                payload = super().finalize(context, tracker)
                payload["session_messages_delta"] = [
                    {
                        "role": "assistant",
                        "tool_calls": [
                            {
                                "id": "call_send_1",
                                "call_id": "call_send_1",
                                "function": {
                                    "name": "mcp_rsi_task_trace_123_0_slack_deadbeef_slack_send_message",
                                    "arguments": json.dumps(
                                        {
                                            "channel_id": "C123",
                                            "thread_ts": "171000001.000100",
                                            "message": "Final reply",
                                        }
                                    ),
                                },
                            }
                        ],
                    },
                    {
                        "role": "tool",
                        "tool_call_id": "call_send_1",
                        "content": json.dumps(
                            {
                                "result": json.dumps(
                                    {
                                        "message_link": "https://storyprotocol.slack.com/archives/C123/p171000001000100",
                                        "message_context": {
                                            "message_ts": "171000001.000100",
                                            "channel_id": "C123",
                                        },
                                    }
                                )
                            }
                        ),
                    },
                ]
                return payload

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Answer the thread and post directly.",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "reply_delivery_mode": "direct",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", DirectReplySessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertFalse(result.raw["action_contract_repair_attempted"])
        self.assertFalse(result.raw["action_contract_repair_succeeded"])
        self.assertEqual(result.raw["reply_delivery"]["status"], "posted")
        self.assertEqual(result.raw["reply_delivery"]["channel_id"], "C123")
        self.assertEqual(result.raw["reply_delivery"]["thread_ts"], "171000001.000100")
        self.assertEqual(result.raw["reply_delivery"]["tool_call_id"], "call_send_1")
        self.assertEqual(result.raw["reply_delivery"]["provider_ref"], "171000001.000100")
        self.assertEqual(result.raw["structured_output"]["reply_delivery"]["status"], "posted")

    def test_artifact_workflow_executes_investigate_render_and_deliver_phases(self) -> None:
        class ArtifactPhaseSessionManager(FakeSessionManager):
            def prepare(self, task: RunnerTaskRequest):
                context = super().prepare(task)
                context.execution_phase = task.execution_phase or "main"
                return context

            def finalize(self, context: types.SimpleNamespace, tracker: FakeTracker) -> dict[str, object]:
                payload = super().finalize(context, tracker)
                if getattr(context, "execution_phase", "") == "deliver":
                    payload["session_messages_delta"] = [
                        {
                            "role": "assistant",
                            "tool_calls": [
                                {
                                    "id": "call_send_deliver",
                                    "call_id": "call_send_deliver",
                                    "function": {
                                        "name": "mcp_rsi_task_trace_artifact_0_slack_deadbeef_slack_send_message",
                                        "arguments": json.dumps(
                                            {
                                                "channel_id": "C123",
                                                "thread_ts": "171000001.000100",
                                                "message": "Grounded answer with diagram attached.",
                                            }
                                        ),
                                    },
                                }
                            ],
                        },
                        {
                            "role": "tool",
                            "tool_call_id": "call_send_deliver",
                            "content": json.dumps(
                                {
                                    "result": json.dumps(
                                        {
                                            "message_link": "https://storyprotocol.slack.com/archives/C123/p171000001000100",
                                            "message_context": {
                                                "message_ts": "171000001.000100",
                                                "channel_id": "C123",
                                            },
                                        }
                                    )
                                }
                            ),
                        },
                    ]
                return payload

        class ArtifactWorkflowAIAgent:
            run_history: list[dict[str, object]] = []

            def __init__(self, **kwargs) -> None:
                self._kwargs = dict(kwargs)

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict[str, object]:
                valid_tool_names = sorted(getattr(self, "valid_tool_names", []))
                if "Investigation phase only" in (system_message or ""):
                    phase = "investigate"
                    payload = {
                        "visible_reasoning": [],
                        "reply_draft": "Grounded answer with diagram attached.",
                        "final_answer": "Grounded answer with diagram attached.",
                        "confidence": 0.9,
                        "context_summary": "Compact grounded context for render.",
                        "self_critique": "",
                        "proposed_actions": [],
                        "knowledge_drafts": [],
                        "outcome_hypotheses": [],
                        "produced_artifacts": [],
                        "artifact_failure_reason": "",
                        "artifact_render_briefs": [
                            {
                                "kind": "diagram",
                                "skill": "/architecture-diagram",
                                "title": "DePIN Backend",
                                "render_prompt": "Render the grounded depin backend architecture.",
                                "inputs": {"services": ["depin-api", "ip-registration-worker"]},
                                "output_path_hint": "",
                            }
                        ],
                    }
                elif "Render phase only" in (system_message or ""):
                    phase = "render"
                    output_path = ""
                    for line in prompt.splitlines():
                        if line.startswith("Output path: "):
                            output_path = line.split("Output path: ", 1)[1].strip()
                            break
                    Path(output_path).write_text("<html><body>diagram</body></html>", encoding="utf-8")
                    payload = {
                        "visible_reasoning": [],
                        "reply_draft": "",
                        "final_answer": "",
                        "confidence": 0.82,
                        "context_summary": "Rendered from compact brief.",
                        "self_critique": "",
                        "proposed_actions": [],
                        "knowledge_drafts": [],
                        "outcome_hypotheses": [],
                        "produced_artifacts": [
                            {
                                "kind": "diagram",
                                "title": "DePIN Backend",
                                "artifact_refs": [f"file://{output_path}"],
                                "delivery_status": "generated",
                                "failure_reason": "",
                            }
                        ],
                        "artifact_failure_reason": "",
                    }
                else:
                    phase = "deliver"
                    payload = {
                        "visible_reasoning": [],
                        "reply_draft": "Grounded answer with diagram attached.",
                        "final_answer": "Grounded answer with diagram attached.",
                        "confidence": 0.88,
                        "context_summary": "Delivery only.",
                        "self_critique": "",
                        "proposed_actions": [],
                        "knowledge_drafts": [],
                        "outcome_hypotheses": [],
                        "produced_artifacts": [],
                        "artifact_failure_reason": "",
                    }
                type(self).run_history.append(
                    {
                        "phase": phase,
                        "prompt": prompt,
                        "system_message": system_message,
                        "valid_tool_names": valid_tool_names,
                        "tool_names": sorted(
                            tool["function"]["name"]
                            for tool in list(getattr(self, "tools", []) or [])
                            if isinstance(tool, dict) and isinstance(tool.get("function"), dict) and tool["function"].get("name")
                        ),
                        "task_id": task_id,
                    }
                )
                return {"final_response": json.dumps(payload)}

            def interrupt(self, _message: str | None = None) -> None:
                return None

            def get_activity_summary(self) -> dict[str, object]:
                return {
                    "last_activity_desc": "completed",
                    "current_tool": "",
                    "api_call_count": 1,
                    "budget_used": 1,
                    "budget_max": 4,
                }

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Draw the grounded deployment architecture and attach the diagram.",
                    "trace_id": "trace-artifact",
                    "workflow_id": "wf-artifact",
                    "operation_id": "op-artifact",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "reply_delivery_mode": "direct",
                    "requested_skills": ["architecture-diagram"],
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-artifact",
                }
            }
        )

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", ArtifactWorkflowAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", ArtifactPhaseSessionManager
        ), mock.patch.dict(
            os.environ, {**runner_env("prod"), "HERMES_HOME": tempdir}, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)
            artifact_ref = result.raw["structured_output"]["produced_artifacts"][0]["artifact_refs"][0]
            artifact_exists = Path(artifact_ref.removeprefix("file://")).exists()

        self.assertTrue(result.ok)
        self.assertEqual([item["phase"] for item in ArtifactWorkflowAIAgent.run_history], ["investigate", "render", "deliver"])
        self.assertNotIn("slack_upload_file", ArtifactWorkflowAIAgent.run_history[0]["valid_tool_names"])
        self.assertIn("Output path: ", ArtifactWorkflowAIAgent.run_history[1]["prompt"])
        self.assertIn("Grounded context summary: Compact grounded context for render.", ArtifactWorkflowAIAgent.run_history[1]["prompt"])
        self.assertNotIn("repo_context", ArtifactWorkflowAIAgent.run_history[1]["valid_tool_names"])
        self.assertNotIn("repo_context", ArtifactWorkflowAIAgent.run_history[2]["valid_tool_names"])
        self.assertIn("slack_upload_file", ArtifactWorkflowAIAgent.run_history[2]["valid_tool_names"])
        produced = result.raw["structured_output"]["produced_artifacts"]
        self.assertEqual(len(produced), 1)
        self.assertTrue(artifact_ref.startswith("file://"))
        self.assertTrue(artifact_exists)
        self.assertEqual(result.raw["reply_delivery"]["status"], "posted")
        self.assertEqual(result.raw["runner_diagnostics"]["artifact_phase_enabled"], True)
        self.assertEqual(result.raw["runner_diagnostics"]["observation_seq"] > 0, True)

    def test_artifact_phase_budgets_never_exceed_total_timeout(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Render a diagram artifact.",
                    "timeout_seconds": 100,
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            budgets = runtime._artifact_phase_budgets(task)

        self.assertEqual(budgets["total"], 100)
        self.assertGreaterEqual(budgets["investigate"], 1)
        self.assertGreaterEqual(budgets["render"], 0)
        self.assertGreaterEqual(budgets["deliver"], 0)
        self.assertGreaterEqual(budgets["reducer_reserve"], 0)
        self.assertLessEqual(
            budgets["investigate"] + budgets["render"] + budgets["deliver"] + budgets["reducer_reserve"],
            budgets["total"],
        )

    def test_architecture_diagram_render_budget_gets_five_minutes(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Render an architecture diagram artifact.",
                    "timeout_seconds": 900,
                    "requested_skills": ["architecture-diagram"],
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            budgets = runtime._artifact_phase_budgets(task)

        self.assertEqual(budgets["total"], 900)
        self.assertEqual(budgets["render"], 300)
        self.assertLessEqual(
            budgets["investigate"] + budgets["render"] + budgets["deliver"] + budgets["reducer_reserve"],
            budgets["total"],
        )

    def test_artifact_investigate_task_clears_requested_artifacts(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Draw the architecture diagram.",
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                    "allowed_tools": ["slack.upload_file"],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            investigate_task = runtime._build_artifact_investigate_task(
                task,
                {"investigate": 60},
            )

        self.assertEqual(investigate_task.requested_artifacts, [])
        self.assertEqual(investigate_task.execution_phase, "investigate")

    def test_render_phase_allows_only_explicit_workspace_file_helpers(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Render the artifact using the workspace.",
                    "execution_phase": "render",
                    "execution_mode": "artifact_render",
                    "allowed_tools": [
                        "workspace.write_file",
                        "workspace.read_file",
                        "workspace.git_history",
                        "repo.context",
                        "slack.upload_file",
                    ],
                    "tool_allowlist": [
                        "workspace.write_file",
                        "workspace.read_file",
                        "workspace.git_history",
                        "repo.context",
                        "slack.upload_file",
                    ],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            tool_policy = runtime._resolve_tool_policy(task)
            native_toolsets = runtime._native_toolsets_for_task(task)

        self.assertIn("workspace.write_file", tool_policy.effective)
        self.assertIn("workspace.read_file", tool_policy.effective)
        self.assertNotIn("workspace.git_history", tool_policy.effective)
        self.assertNotIn("repo.context", tool_policy.effective)
        self.assertNotIn("slack.upload_file", tool_policy.effective)
        self.assertIn("rsi-governed-workspace", native_toolsets)

    def test_artifact_workflow_preserves_native_artifact_destination_across_phases(self) -> None:
        executed_tasks: list[RunnerTaskRequest] = []

        def fake_execute_task_internal(task: RunnerTaskRequest, observer=None):
            executed_tasks.append(task)
            if task.execution_phase == "investigate":
                return HermesExecutionResult(
                    ok=True,
                    message="investigate",
                    provider="test",
                    raw={
                        "structured_output": {
                            "final_answer": "Grounded answer",
                            "context_summary": "Grounded context",
                            "artifact_render_briefs": [
                                {
                                    "kind": "diagram",
                                    "skill": "architecture-diagram",
                                    "title": "DePIN Backend",
                                    "render_prompt": "Render the grounded backend diagram.",
                                }
                            ],
                        }
                    },
                )
            if task.execution_phase == "render":
                return HermesExecutionResult(
                    ok=True,
                    message="render",
                    provider="test",
                    raw={
                        "structured_output": {
                            "produced_artifacts": [
                                {
                                    "kind": "diagram",
                                    "title": "DePIN Backend",
                                    "artifact_refs": [f"file://{Path(task.artifact_destination) / 'depin-backend.html'}"],
                                    "delivery_status": "generated",
                                    "failure_reason": "",
                                }
                            ],
                            "artifact_failure_reason": "",
                        }
                    },
                )
            return HermesExecutionResult(
                ok=True,
                message="deliver",
                provider="test",
                raw={"structured_output": {"reply_delivery": {"status": "posted"}}},
            )

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            runtime._execute_task_internal = fake_execute_task_internal  # type: ignore[method-assign]
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "Render the architecture diagram.",
                        "trace_id": "trace-artifacts",
                        "workflow_id": "wf-artifacts",
                        "operation_id": "op-artifacts",
                        "channel_id": "C123",
                        "thread_ts": "171000001.000100",
                        "reply_delivery_mode": "direct",
                        "requested_skills": ["architecture-diagram"],
                        "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                    }
                }
            )
            result = runtime._execute_artifact_workflow_task(task, runtime._resolve_tool_policy(task))

        self.assertTrue(result.ok)
        self.assertEqual([item.execution_phase for item in executed_tasks], ["investigate", "render", "deliver"])
        expected_artifact_root = str((Path(tempdir) / "company" / "artifacts" / "depin-backend" / time.strftime("%Y-%m-%d", time.gmtime()) / "op-artifacts").resolve())
        self.assertTrue(all(item.artifact_destination == expected_artifact_root for item in executed_tasks))
        self.assertIn(f"Output path: {expected_artifact_root}", executed_tasks[1].prompt)
        self.assertNotIn("/var/lib/hermes/rsi_runtime/artifacts", executed_tasks[1].prompt)

    def test_artifact_workflow_synthesizes_native_artifacts_for_delivery(self) -> None:
        executed_tasks: list[RunnerTaskRequest] = []

        def fake_execute_task_internal(task: RunnerTaskRequest, observer=None):
            executed_tasks.append(task)
            if task.execution_phase == "investigate":
                return HermesExecutionResult(
                    ok=True,
                    message="investigate",
                    provider="test",
                    raw={
                        "structured_output": {
                            "final_answer": "Grounded answer",
                            "context_summary": "Grounded context",
                            "artifact_render_briefs": [
                                {
                                    "kind": "diagram",
                                    "skill": "architecture-diagram",
                                    "title": "DePIN Backend",
                                    "render_prompt": "Render the grounded backend diagram.",
                                }
                            ],
                        }
                    },
                )
            if task.execution_phase == "render":
                artifact_path = str((Path(task.artifact_destination) / "depin-backend.html").resolve())
                return HermesExecutionResult(
                    ok=True,
                    message="render",
                    provider="test",
                    raw={
                        "structured_output": {
                            "produced_artifacts": [],
                            "artifact_failure_reason": "",
                        },
                        "native_artifact_paths": [artifact_path],
                    },
                )
            self.assertEqual(task.allowed_tools, ["slack.upload_file"])
            self.assertIn("Produced artifacts:", task.prompt)
            return HermesExecutionResult(
                ok=True,
                message="deliver",
                provider="test",
                raw={"structured_output": {"reply_delivery": {"status": "posted"}}},
            )

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            runtime._execute_task_internal = fake_execute_task_internal  # type: ignore[method-assign]
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "Render the architecture diagram.",
                        "trace_id": "trace-artifacts",
                        "workflow_id": "wf-artifacts",
                        "operation_id": "op-artifacts",
                        "channel_id": "C123",
                        "thread_ts": "171000001.000100",
                        "reply_delivery_mode": "direct",
                        "requested_skills": ["architecture-diagram"],
                        "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                    }
                }
            )
            result = runtime._execute_artifact_workflow_task(task, runtime._resolve_tool_policy(task))

        self.assertTrue(result.ok)
        produced = result.raw["structured_output"]["produced_artifacts"]
        self.assertEqual(len(produced), 1)
        self.assertTrue(produced[0]["artifact_refs"][0].startswith("file://"))

    def test_artifact_render_briefs_fallback_uses_user_request_text(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "User request: /architecture-diagram map the services and data flows",
                    "requested_skills": ["architecture-diagram"],
                    "requested_artifacts": [{"kind": "diagram", "description": ""}],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            briefs = runtime._artifact_render_briefs(task, {})

        self.assertEqual(len(briefs), 1)
        self.assertEqual(briefs[0]["kind"], "diagram")
        self.assertEqual(briefs[0]["skill"], "architecture-diagram")
        self.assertEqual(briefs[0]["title"], "diagram-1")
        self.assertEqual(briefs[0]["render_prompt"], "/architecture-diagram map the services and data flows")

    def test_artifact_render_briefs_hydrate_missing_fields(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Render the backend diagram",
                    "requested_skills": ["architecture-diagram"],
                    "requested_artifacts": [{"kind": "diagram", "description": "Backend architecture"}],
                    "trace_id": "trace-brief",
                }
            }
        )

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, {**runner_env("prod"), "HERMES_HOME": tempdir}, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            briefs = runtime._artifact_render_briefs(
                task,
                {
                    "final_answer": "Grounded answer",
                    "context_summary": "Grounded summary",
                    "artifact_render_briefs": [
                        {
                            "kind": "diagram",
                            "skill": "",
                            "title": "",
                            "render_prompt": "",
                            "inputs": {},
                            "output_path_hint": "",
                        }
                    ],
                },
            )

        self.assertEqual(len(briefs), 1)
        self.assertEqual(briefs[0]["skill"], "architecture-diagram")
        self.assertEqual(briefs[0]["title"], "Backend architecture")
        self.assertIn("Grounded answer", briefs[0]["render_prompt"])
        self.assertTrue(briefs[0]["output_path_hint"].endswith(".html"))

    def test_artifact_render_briefs_normalize_slash_prefixed_skill(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Render the backend diagram",
                    "requested_skills": ["architecture-diagram"],
                    "requested_artifacts": [{"kind": "diagram", "description": "Backend architecture"}],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            briefs = runtime._artifact_render_briefs(
                task,
                {
                    "final_answer": "Grounded answer",
                    "context_summary": "Grounded summary",
                    "artifact_render_briefs": [
                        {
                            "kind": "diagram",
                            "skill": "/architecture-diagram",
                            "title": "Backend architecture",
                            "render_prompt": "Render it",
                            "inputs": {},
                            "output_path_hint": "",
                        }
                    ],
                },
            )
            render_task = runtime._build_artifact_render_task(task, briefs[0], {"context_summary": "", "final_answer": ""}, {"render": 60}, 0)

        self.assertEqual(briefs[0]["skill"], "architecture-diagram")
        self.assertEqual(render_task.requested_skills, ["architecture-diagram"])
        self.assertIn("Selected skill: architecture-diagram", render_task.prompt)

    def test_artifact_render_briefs_tolerate_non_string_skill_fields(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Render the backend diagram",
                    "requested_skills": ["architecture-diagram"],
                    "requested_artifacts": [{"kind": "diagram", "description": "Backend architecture"}],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            briefs = runtime._artifact_render_briefs(
                task,
                {
                    "artifact_render_briefs": [
                        {
                            "kind": "diagram",
                            "requested_skill": {"unexpected": "shape"},
                            "title": "Backend architecture",
                            "render_prompt": "Render it",
                            "inputs": {},
                            "output_path_hint": "",
                        },
                        {
                            "kind": "diagram",
                            "skill": 123,
                            "title": "Backend architecture 2",
                            "render_prompt": "Render it again",
                            "inputs": {},
                            "output_path_hint": "",
                        },
                    ]
                },
            )

        self.assertEqual(len(briefs), 2)
        self.assertEqual(briefs[0]["requested_skill"], '{"unexpected": "shape"}')
        self.assertEqual(briefs[0]["skill"], "")
        self.assertEqual(briefs[1]["requested_skill"], "123")
        self.assertEqual(briefs[1]["skill"], "123")

    def test_artifact_render_invalid_skill_stays_structured(self) -> None:
        class InvalidSkillAIAgent:
            run_history: list[dict[str, object]] = []

            def __init__(self, **kwargs) -> None:
                self._kwargs = kwargs

            def run_conversation(self, prompt: str, system_message: str | None = None, conversation_history=None, task_id: str | None = None):
                if "Investigation phase only" in (system_message or ""):
                    payload = {
                        "visible_reasoning": [],
                        "reply_draft": "Grounded answer without diagram.",
                        "final_answer": "Grounded answer without diagram.",
                        "confidence": 0.8,
                        "context_summary": "Compact grounded context.",
                        "self_critique": "",
                        "proposed_actions": [],
                        "knowledge_drafts": [],
                        "outcome_hypotheses": [],
                        "produced_artifacts": [],
                        "artifact_failure_reason": "",
                        "artifact_render_briefs": [
                            {
                                "kind": "diagram",
                                "skill": "/bad skill!",
                                "title": "Broken diagram",
                                "render_prompt": "Render a diagram",
                                "inputs": {},
                                "output_path_hint": "",
                            }
                        ],
                    }
                else:
                    payload = {
                        "visible_reasoning": [],
                        "reply_draft": "Grounded answer without diagram.",
                        "final_answer": "Grounded answer without diagram.",
                        "confidence": 0.8,
                        "context_summary": "Delivery only.",
                        "self_critique": "",
                        "proposed_actions": [],
                        "knowledge_drafts": [],
                        "outcome_hypotheses": [],
                        "produced_artifacts": [],
                        "artifact_failure_reason": "",
                    }
                type(self).run_history.append({"system_message": system_message or "", "prompt": prompt})
                return {"final_response": json.dumps(payload)}

            def interrupt(self, _message: str | None = None) -> None:
                return None

            def get_activity_summary(self) -> dict[str, object]:
                return {"last_activity_desc": "completed", "current_tool": "", "api_call_count": 1, "budget_used": 1, "budget_max": 4}

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", InvalidSkillAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, {**runner_env("prod"), "HERMES_HOME": tempdir}, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "Draw the grounded deployment architecture and attach the diagram.",
                        "trace_id": "trace-invalid-skill",
                        "workflow_id": "wf-invalid-skill",
                        "reply_delivery_mode": "none",
                        "requested_skills": ["architecture-diagram"],
                        "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-invalid-skill",
                    }
                }
            )
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual([entry["system_message"] for entry in InvalidSkillAIAgent.run_history].count(""), 0)
        self.assertIn("Artifact render skill identifier is invalid after normalization.", result.raw["structured_output"]["artifact_failure_reason"])
        self.assertEqual(result.raw["structured_output"]["produced_artifacts"], [])

    def test_artifact_delivery_without_artifacts_disables_upload_tool(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Deliver the result",
                    "reply_delivery_mode": "direct",
                    "allowed_tools": ["slack.upload_file"],
                    "tool_allowlist": ["slack.upload_file"],
                    "mcp_servers": [{"name": "slack", "profile": "slack_mcp_reply"}],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            deliver_task = runtime._build_artifact_delivery_task(
                task,
                {"final_answer": "Reply body", "context_summary": "Summary"},
                [],
                {"deliver": 60},
            )
            tool_policy = runtime._resolve_tool_policy(deliver_task)

        self.assertEqual(deliver_task.allowed_tools, [])
        self.assertEqual(deliver_task.tool_allowlist, [])
        self.assertEqual(tool_policy.effective, [])
        self.assertIn("Slack MCP", deliver_task.prompt)
        self.assertIn("Render failure reason: none", deliver_task.prompt)
        self.assertIn("Do not call slack.upload_file", deliver_task.prompt)

    def test_workflow_task_timeout_enters_partial_finalization_before_hard_deadline(self) -> None:
        class TimeoutReducerAIAgent:
            init_history: list[dict[str, object]] = []
            run_history: list[dict[str, object]] = []
            interrupt_messages: list[str] = []
            created_instances = 0

            def __init__(self, **kwargs) -> None:
                type(self).created_instances += 1
                self.instance_index = type(self).created_instances
                type(self).init_history.append(dict(kwargs))

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict[str, object]:
                type(self).run_history.append(
                    {
                        "instance_index": self.instance_index,
                        "prompt": prompt,
                        "system_message": system_message,
                        "history": list(conversation_history or []),
                        "valid_tool_names": sorted(getattr(self, "valid_tool_names", [])),
                        "tool_names": sorted(
                            tool["function"]["name"]
                            for tool in list(getattr(self, "tools", []) or [])
                            if isinstance(tool, dict) and isinstance(tool.get("function"), dict) and tool["function"].get("name")
                        ),
                        "task_id": task_id,
                    }
                )
                time.sleep(1.2)
                return {"final_response": ""}

            def interrupt(self, message: str | None = None) -> None:
                type(self).interrupt_messages.append(message or "")

            def get_activity_summary(self) -> dict[str, object]:
                budget_max = int(type(self).init_history[self.instance_index - 1].get("max_iterations", 1))
                if self.instance_index == 1:
                    return {
                        "last_activity_desc": "starting API call #9",
                        "current_tool": "",
                        "api_call_count": 9,
                        "budget_used": 9,
                        "budget_max": budget_max,
                    }
                return {
                    "last_activity_desc": "composing partial reply",
                    "current_tool": "",
                    "api_call_count": 0,
                    "budget_used": 0,
                    "budget_max": budget_max,
                }

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the Slack thread.",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        captured_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            captured_requests.append(
                {
                    "url": req.full_url,
                    "timeout": timeout,
                    "body": json.loads(req.data.decode("utf-8")),
                }
            )
            return FakeHTTPResponse(
                {
                    "id": "resp_timeout_partial",
                    "output_text": json.dumps(partial_structured_output(reply_text="Partial answer after timeout.")),
                }
            )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", TimeoutReducerAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen), mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "RSI_RUNNER_PROD_TASK_TIMEOUT": "20s"},
            clear=True,
        ), mock.patch(
            "rsi_runner.hermes_runtime.time.monotonic",
            side_effect=[100.0, 110.0],
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["termination_reason"], "task_timeout")
        self.assertFalse(result.raw["max_iterations_reached"])
        self.assertEqual(result.raw["timeout_kind"], "task_timeout")
        self.assertEqual(result.raw["task_timeout_seconds"], 20)
        self.assertEqual(result.raw["stopped_after_seconds"], 10)
        self.assertEqual(result.raw["runner_diagnostics"]["completion_verdict"], "partial")
        self.assertEqual(result.raw["runner_diagnostics"]["termination_reason"], "task_timeout")
        self.assertEqual(result.raw["runner_diagnostics"]["timeout_kind"], "task_timeout")
        self.assertEqual(result.raw["runner_diagnostics"]["budget_used"], 9)
        self.assertEqual(result.raw["runner_diagnostics"]["budget_max"], 20)
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "direct_reducer")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_attempts"], 1)
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_attempted"])
        self.assertEqual(TimeoutReducerAIAgent.created_instances, 1)
        self.assertEqual(len(captured_requests), 1)
        self.assertIn('"termination_reason": "task_timeout"', str(captured_requests[0]["body"]["input"]))
        self.assertNotIn("Earlier thread message", str(captured_requests[0]["body"]["input"]))
        self.assertTrue(any(message == "runner task_timeout after 10s" for message in TimeoutReducerAIAgent.interrupt_messages))

    def test_workflow_iteration_budget_exhaustion_fails_when_direct_reducer_cannot_return_valid_output(self) -> None:
        class InvalidReducerAIAgent:
            init_history: list[dict[str, object]] = []
            created_instances = 0

            def __init__(self, **kwargs) -> None:
                type(self).created_instances += 1
                self.instance_index = type(self).created_instances
                type(self).init_history.append(dict(kwargs))

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict[str, object]:
                time.sleep(0.6)
                return {"final_response": ""}

            def interrupt(self, message: str | None = None) -> None:
                return None

            def get_activity_summary(self) -> dict[str, object]:
                budget_max = int(type(self).init_history[self.instance_index - 1].get("max_iterations", 1))
                if self.instance_index == 1:
                    return {
                        "last_activity_desc": "iteration budget exhausted",
                        "current_tool": "slack_history",
                        "api_call_count": budget_max,
                        "budget_used": budget_max,
                        "budget_max": budget_max,
                    }
                return {
                    "last_activity_desc": "invalid recovery output",
                    "current_tool": "",
                    "api_call_count": 0,
                    "budget_used": 0,
                    "budget_max": budget_max,
                }

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the Slack thread.",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        captured_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            captured_requests.append(
                {
                    "url": req.full_url,
                    "timeout": timeout,
                    "body": json.loads(req.data.decode("utf-8")),
                }
            )
            if len(captured_requests) == 1:
                return FakeHTTPResponse({"id": "resp_fail_1", "output_text": "not valid json"})
            return FakeHTTPResponse({"id": "resp_fail_2", "output_text": "still not valid json"})

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", InvalidReducerAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_partial_completion_unrecoverable")
        self.assertEqual(result.raw["termination_reason"], "iteration_budget_exhausted")
        self.assertTrue(result.raw["runner_diagnostics"]["max_iterations_reached"])
        self.assertEqual(result.raw["runner_diagnostics"]["budget_used"], 20)
        self.assertEqual(result.raw["runner_diagnostics"]["budget_max"], 20)
        self.assertTrue(result.raw["partial_recovery_attempted"])
        self.assertFalse(result.raw["partial_recovery_succeeded"])
        self.assertTrue(result.raw["runner_diagnostics"]["partial_completion_attempted"])
        self.assertFalse(result.raw["runner_diagnostics"]["partial_completion_succeeded"])
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "direct_reducer")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_attempts"], 1)
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_attempted"])
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_succeeded"])
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_timeout_seconds"], 180)
        self.assertEqual(InvalidReducerAIAgent.created_instances, 1)
        self.assertEqual(len(captured_requests), 1)
        self.assertEqual(captured_requests[0]["timeout"], 180)
        self.assertIn("structured output", result.message.lower())

    def test_workflow_evidence_ledger_projects_compact_tool_calls_and_evidence_items(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the Slack thread.",
                    "requested_artifacts": [{"kind": "diagram", "description": "Render the architecture diagram."}],
                    "artifact_optional": True,
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "context_summary": "Slack and repo context were being gathered.",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        rendered = runtime._render_task_prompt(task, runtime._resolve_tool_policy(task))
        task_with_rendered_prompt = RunnerTaskRequest.from_payload({"task": {**task.__dict__, "prompt": rendered}})
        ledger = runtime._build_evidence_ledger(
            task_with_rendered_prompt,
            {
                "tool_calls": [
                    {
                        "tool_name": "repo.search",
                        "tool_call_id": "repo.search:1",
                        "request": {"path": "internal/control", "pattern": "completion_verdict"},
                        "summary": "Found reducer handling in worker.go.",
                        "status": "completed",
                        "provider_ref": "search-1",
                    },
                    {
                        "tool_name": "slack.history",
                        "tool_call_id": "slack.history:2",
                        "request": {"channel_id": "C123", "thread_ts": "171000001.000100"},
                        "summary": "Loaded the bound Slack thread.",
                        "status": "completed",
                        "provider_ref": "slack-history-1",
                    },
                    {
                        "tool_name": "slack.search",
                        "tool_call_id": "slack.search:3",
                        "request": {"query": "depin backend api numo"},
                        "summary": "missing_action_token",
                        "status": "error",
                    },
                ],
                "evidence_items": [
                    {
                        "kind": "repo_search_match",
                        "summary": "worker.go treats completion_verdict=partial as a normal success path.",
                        "snippet": "if completion_verdict == partial { return nil }",
                        "source_ref": "internal/control/worker.go",
                        "tool_name": "repo.search",
                        "path": "internal/control/worker.go",
                        "repo": "rsi-agent-platform",
                    },
                    {
                        "kind": "slack_message",
                        "summary": "[blake] The reducer fallback is safe to ship once the ledger keeps snippets.",
                        "snippet": "The reducer fallback is safe to ship once the ledger keeps snippets.",
                        "source_ref": "https://slack.example/messages/1",
                        "tool_name": "slack.history",
                        "channel_id": "C123",
                        "thread_ts": "171000001.000100",
                        "message_ts": "171000001.000100",
                        "author": "blake",
                        "permalink": "https://slack.example/messages/1",
                    }
                ],
            },
            "task_timeout",
        )

        self.assertEqual(ledger["user_request"], "Summarize the Slack thread.")
        self.assertEqual(ledger["reply_target"], {"channel_id": "C123", "thread_ts": "171000001.000100"})
        self.assertEqual(ledger["termination_reason"], "task_timeout")
        self.assertEqual(ledger["requested_artifacts"], [{"kind": "diagram", "description": "Render the architecture diagram."}])
        self.assertTrue(ledger["artifact_optional"])
        self.assertEqual(len(ledger["tool_calls"]), 3)
        self.assertEqual(ledger["tool_calls"][0]["tool_name"], "repo.search")
        self.assertEqual(ledger["evidence_items"][0]["source_ref"], "internal/control/worker.go")
        self.assertEqual(ledger["evidence_items"][0]["snippet"], "if completion_verdict == partial { return nil }")
        self.assertEqual(ledger["evidence_items"][1]["tool_name"], "slack.history")
        self.assertEqual(ledger["evidence_items"][1]["author"], "blake")
        self.assertEqual(ledger["evidence_items"][1]["message_ts"], "171000001.000100")
        self.assertTrue(ledger["open_questions"])

    def test_runtime_metadata_reports_role_contract(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("proposal"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime.metadata["max_iterations"], 5)
        self.assertEqual(runtime.metadata["task_timeout_seconds"], 420)
        self.assertEqual(runtime.metadata["inactivity_timeout_seconds"], 360)
        self.assertEqual(runtime.metadata["transport_timeout_seconds"], 450)
        self.assertEqual(runtime.metadata["tool_policy_mode"], "enforced_read_only")
        self.assertEqual(runtime.metadata["hermes_pin"], "6fdbf2f2d76cf37393e657bf37ceda3d84589200")
        self.assertEqual(runtime.metadata["session_continuity_status"], "ok")
        self.assertEqual(runtime.metadata["honcho_environment_effective"], "production")
        self.assertIn("repo.context", runtime.metadata["tool_allowlist_effective"])

    def test_prod_runtime_metadata_reports_live_contract(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime.metadata["max_iterations"], 20)
        self.assertEqual(runtime.metadata["task_timeout_seconds"], 900)
        self.assertEqual(runtime.metadata["inactivity_timeout_seconds"], 900)
        self.assertEqual(runtime.metadata["transport_timeout_seconds"], 930)
        self.assertEqual(runtime.metadata["native_max_output_tokens"], 15000)

    def test_eval_role_rejects_repo_change_task(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "repo-change",
                    "repo": "rsi-agent-platform",
                    "prompt": "This should be blocked.",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, runner_env("eval"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.provider, "policy")
        self.assertIn("cannot execute", result.message)


if __name__ == "__main__":
    unittest.main()
