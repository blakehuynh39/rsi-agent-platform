from __future__ import annotations

import json
import os
import types
import unittest
from unittest import mock

from rsi_runner.config import RunnerConfig, RunnerConfigError
from rsi_runner.hermes_runtime import HermesRuntime, RunnerTaskRequest


def runner_env(role: str = "prod") -> dict[str, str]:
    return {
        "RSI_RUNNER_ROLE": role,
        "RSI_RUNNER_HOST": "0.0.0.0",
        "RSI_RUNNER_PORT": "8090",
        "RSI_RUNNER_MODEL": "openai/gpt-5.4",
        "RSI_RUNNER_REASONING_EFFORT": "xhigh",
        "RSI_RUNNER_PUBLIC_BASE_URL": "https://staging-rsi-platform.storyprotocol.net",
        "HERMES_HOME": "/tmp/hermes",
        "RSI_RUNNER_MEMORY_BACKEND": "honcho",
        "RSI_HONCHO_WORKSPACE": "rsi-stage",
        "RSI_HONCHO_RECALL_MODE": "hybrid",
        "RSI_HONCHO_WRITE_FREQUENCY": "async",
        "RSI_HONCHO_SESSION_STRATEGY": "hybrid",
        "RSI_HONCHO_AI_PEER": f"rsi:stage:{role}",
        "RSI_HONCHO_ENVIRONMENT": "stage",
        "HONCHO_API_KEY": "honcho-test-key",
        "OPENAI_API_KEY": "openai-test-key",
    }


class FakeAIAgent:
    last_kwargs: dict | None = None
    last_prompt: str | None = None
    last_system_message: str | None = None
    last_history: list[dict] | None = None

    def __init__(self, **kwargs) -> None:
        type(self).last_kwargs = kwargs

    def run_conversation(
        self,
        prompt: str,
        system_message: str | None = None,
        conversation_history: list[dict] | None = None,
    ) -> dict:
        type(self).last_prompt = prompt
        type(self).last_system_message = system_message
        type(self).last_history = conversation_history or []
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
        self.session_db = object()
        self.honcho_available = True

    def prepare(self, task):
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

    def attach_tracking(self, _agent, _task, _context):
        return FakeTracker()

    def finalize(self, context, tracker):
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
        }


class HermesRuntimeTests(unittest.TestCase):
    def setUp(self) -> None:
        FakeAIAgent.last_kwargs = None
        FakeAIAgent.last_prompt = None
        FakeAIAgent.last_system_message = None
        FakeAIAgent.last_history = None

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
        self.assertEqual(result.raw["honcho_recall_mode"], "hybrid")
        self.assertEqual(result.raw["honcho_write_frequency"], "async")
        self.assertEqual(result.raw["honcho_session_strategy"], "hybrid")
        self.assertEqual(result.raw["honcho_ai_peer"], "rsi:stage:eval")
        self.assertIn("structured_output", result.raw)
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")

    def test_system_message_is_forwarded_to_hermes_run_conversation(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute("User prompt", system_message="System directive")

        self.assertTrue(result.ok)
        self.assertEqual(FakeAIAgent.last_prompt, "User prompt")
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
