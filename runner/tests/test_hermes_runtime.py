from __future__ import annotations

import json
import os
import time
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
        "RSI_HERMES_PIN": "0e336b0e717027cbb81fcb5816246b7aec2d4a47",
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
        "RSI_RUNNER_EVAL_TASK_TIMEOUT": "300s",
        "RSI_RUNNER_PROPOSAL_TASK_TIMEOUT": "420s",
        "RSI_RUNNER_EVAL_INACTIVITY_TIMEOUT": "240s",
        "RSI_RUNNER_PROPOSAL_INACTIVITY_TIMEOUT": "360s",
        "RSI_RUNNER_PROD_TIMEOUT": "60s",
        "RSI_RUNNER_PROACTIVE_TIMEOUT": "60s",
        "RSI_RUNNER_EVAL_TIMEOUT": "330s",
        "RSI_RUNNER_PROPOSAL_TIMEOUT": "450s",
        "HONCHO_API_KEY": "honcho-test-key",
        "OPENAI_API_KEY": "openai-test-key",
    }


class FakeAIAgent:
    last_kwargs: dict | None = None
    last_prompt: str | None = None
    last_system_message: str | None = None
    last_history: list[dict] | None = None
    last_valid_tool_names: list[str] | None = None
    last_tool_names: list[str] | None = None
    sleep_seconds: float = 0.0

    def __init__(self, **kwargs) -> None:
        type(self).last_kwargs = kwargs
        self._interrupted = False

    def run_conversation(
        self,
        prompt: str,
        system_message: str | None = None,
        conversation_history: list[dict] | None = None,
    ) -> dict:
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

    def get_activity_summary(self) -> dict:
        return {
            "last_activity_desc": "waiting_on_tool_or_model",
            "current_tool": "rsi.trace_context",
            "budget_used": 1,
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
        FakeAIAgent.last_valid_tool_names = None
        FakeAIAgent.last_tool_names = None
        FakeAIAgent.sleep_seconds = 0.0

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
        self.assertEqual(result.raw["hermes_pin"], "0e336b0e717027cbb81fcb5816246b7aec2d4a47")
        self.assertIn("structured_output", result.raw)
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")

    def test_non_json_runner_output_fails_closed(self) -> None:
        class UnstructuredAIAgent(FakeAIAgent):
            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
            ) -> dict:
                type(self).last_prompt = prompt
                type(self).last_system_message = system_message
                type(self).last_history = conversation_history or []
                return {"final_response": "plain text response"}

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "eval",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the eval.",
                    "session_scope_kind": "eval_line",
                    "session_scope_id": "shared-store:pk-collision",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:eval",
                    "user_peer_id": "operator:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", UnstructuredAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("eval"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertIn("structured output", result.message)
        self.assertEqual(result.raw["raw_response"], "plain text response")
        self.assertIn("structured_output_error", result.raw)

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
        self.assertIn("repo.context", FakeAIAgent.last_valid_tool_names)
        self.assertIn("rsi.candidate_context", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("github.create_pr", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("honcho_conclude", FakeAIAgent.last_valid_tool_names)
        self.assertEqual(result.raw["tool_policy_mode"], "enforced_read_only")
        self.assertIn("github.create_pr", result.raw["blocked_tool_names"])
        self.assertIn("honcho_conclude", result.raw["blocked_tool_names"])
        self.assertIn("repo.context", result.raw["tool_allowlist_effective"])
        self.assertIn("rsi.candidate_context", result.raw["tool_allowlist_effective"])

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
        self.assertIn("repo.context", FakeAIAgent.last_valid_tool_names)
        self.assertIn("knowledge.context", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("github.create_pr", FakeAIAgent.last_valid_tool_names)
        self.assertEqual(result.raw["tool_policy_mode"], "enforced_read_only")
        self.assertIn("github.create_pr", result.raw["blocked_tool_names"])

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
        self.assertEqual(runtime.metadata["hermes_pin"], "0e336b0e717027cbb81fcb5816246b7aec2d4a47")
        self.assertEqual(runtime.metadata["session_continuity_status"], "ok")
        self.assertIn("repo.context", runtime.metadata["tool_allowlist_effective"])

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
