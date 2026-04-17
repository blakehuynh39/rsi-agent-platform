from __future__ import annotations

import json
import os
import tempfile
import time
import types
import unittest
from unittest import mock

from rsi_runner.config import RunnerConfig, RunnerConfigError
from rsi_runner.hermes_runtime import HermesRuntime, RunnerTaskRequest
from rsi_runner.rsi_tools import ReadOnlyToolBinding, tool_schema_wrappers
from rsi_runner.session_manager import SessionManager


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
        "RSI_RUNNER_PROD_MAX_ITERATIONS": "20",
        "RSI_RUNNER_PROACTIVE_MAX_ITERATIONS": "20",
        "RSI_RUNNER_EVAL_TASK_TIMEOUT": "300s",
        "RSI_RUNNER_PROPOSAL_TASK_TIMEOUT": "420s",
        "RSI_RUNNER_PROD_TASK_TIMEOUT": "300s",
        "RSI_RUNNER_EVAL_INACTIVITY_TIMEOUT": "240s",
        "RSI_RUNNER_PROPOSAL_INACTIVITY_TIMEOUT": "360s",
        "RSI_RUNNER_PROD_TIMEOUT": "330s",
        "RSI_RUNNER_PROACTIVE_TIMEOUT": "60s",
        "RSI_RUNNER_EVAL_TIMEOUT": "330s",
        "RSI_RUNNER_PROPOSAL_TIMEOUT": "450s",
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

    def test_config_rejects_timeout_contract_drift(self) -> None:
        with mock.patch.dict(
            os.environ,
            {**runner_env("eval"), "RSI_RUNNER_EVAL_TASK_TIMEOUT": "328s", "RSI_RUNNER_EVAL_TIMEOUT": "330s"},
            clear=True,
        ):
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
        ), mock.patch.dict(os.environ, {**runner_env("eval"), "HERMES_HOME": tempdir}, clear=True):
            config = RunnerConfig.from_env()
            SessionManager(config)

            with open(os.path.join(tempdir, "honcho.json"), "r", encoding="utf-8") as fh:
                payload = json.load(fh)

        self.assertEqual(payload["environment"], "production")

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
        self.assertEqual(result.raw["hermes_pin"], "0e336b0e717027cbb81fcb5816246b7aec2d4a47")
        self.assertIn("structured_output", result.raw)
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")

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

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict:
                type(self).calls += 1
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
                    "prompt": "Summarize the workflow.",
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
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertTrue(result.raw["repair_attempted"])
        self.assertTrue(result.raw["repair_succeeded"])
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")
        self.assertEqual(result.raw["repair_original_response"], "plain text response")

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
        self.assertEqual(result.raw["task_timeout_seconds"], 300)
        self.assertEqual(result.raw["transport_timeout_seconds"], 330)
        self.assertIn("github.create_pr", result.raw["blocked_tool_names"])
        self.assertIn("repo_context", result.raw["tool_transport_allowlist_effective"])
        self.assertIn("knowledge_context", result.raw["tool_transport_allowlist_effective"])

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
        class BudgetRecoveryAIAgent:
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
                if self.instance_index == 1:
                    time.sleep(0.6)
                    return {"final_response": ""}
                return {
                    "final_response": json.dumps(
                        {
                            "visible_reasoning": [
                                {
                                    "step_type": "analysis",
                                    "summary": "Recovered a partial answer from the persisted session context.",
                                    "alternatives": [],
                                    "confidence": 0.72,
                                    "decision": "post_partial_reply",
                                }
                            ],
                            "reply_draft": "Partial answer: grounded summary so far.",
                            "final_answer": "Partial answer: grounded summary so far.",
                            "confidence": 0.72,
                            "context_summary": "Recovered from the persisted session without new tools.",
                            "self_critique": "Additional repository and Slack reads would improve coverage.",
                            "proposed_actions": [
                                {
                                    "kind": "slack_post",
                                    "target_ref": "C123",
                                    "request_payload": {
                                        "body": "Partial answer: grounded summary so far.",
                                    },
                                    "approval_mode": "not_required",
                                    "idempotency_key": "partial-reply-1",
                                    "rationale": "Post the grounded partial answer.",
                                    "evidence_refs": [],
                                }
                            ],
                            "knowledge_drafts": [],
                            "outcome_hypotheses": [],
                        }
                    )
                }

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

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", BudgetRecoveryAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
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
        self.assertEqual(len(BudgetRecoveryAIAgent.init_history), 2)
        self.assertEqual(BudgetRecoveryAIAgent.init_history[0]["session_id"], BudgetRecoveryAIAgent.init_history[1]["session_id"])
        self.assertEqual(BudgetRecoveryAIAgent.init_history[0]["max_iterations"], 20)
        self.assertEqual(BudgetRecoveryAIAgent.init_history[1]["max_iterations"], 1)
        self.assertIn("repo_context", BudgetRecoveryAIAgent.run_history[0]["valid_tool_names"])
        self.assertEqual(BudgetRecoveryAIAgent.run_history[1]["valid_tool_names"], [])
        self.assertEqual(BudgetRecoveryAIAgent.run_history[1]["tool_names"], [])
        self.assertIn("Recovery instruction", BudgetRecoveryAIAgent.run_history[1]["prompt"])
        self.assertEqual(BudgetRecoveryAIAgent.run_history[1]["history"], [{"role": "user", "content": "Earlier thread message"}])
        self.assertTrue(any("iteration_budget_exhausted" in message for message in BudgetRecoveryAIAgent.interrupt_messages))

    def test_workflow_iteration_budget_exhaustion_fails_when_recovery_output_is_invalid(self) -> None:
        class InvalidRecoveryAIAgent:
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
                if self.instance_index == 1:
                    time.sleep(0.6)
                    return {"final_response": ""}
                return {"final_response": "not valid json"}

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

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", InvalidRecoveryAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_iteration_budget_exhausted")
        self.assertEqual(result.raw["termination_reason"], "iteration_budget_exhausted")
        self.assertTrue(result.raw["runner_diagnostics"]["max_iterations_reached"])
        self.assertEqual(result.raw["runner_diagnostics"]["budget_used"], 20)
        self.assertEqual(result.raw["runner_diagnostics"]["budget_max"], 20)
        self.assertIn("structured output", result.message.lower())

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
        self.assertEqual(runtime.metadata["honcho_environment_effective"], "production")
        self.assertIn("repo.context", runtime.metadata["tool_allowlist_effective"])

    def test_prod_runtime_metadata_reports_live_contract(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime.metadata["max_iterations"], 20)
        self.assertEqual(runtime.metadata["task_timeout_seconds"], 300)
        self.assertEqual(runtime.metadata["inactivity_timeout_seconds"], 300)
        self.assertEqual(runtime.metadata["transport_timeout_seconds"], 330)

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
