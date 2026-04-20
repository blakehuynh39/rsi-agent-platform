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
from rsi_runner.rsi_tools import ReadOnlyToolBinding, governed_toolset_definitions, tool_schema_wrappers, transport_tool_schema
from rsi_runner.session_manager import SessionManager


def runner_env(role: str = "prod") -> dict[str, str]:
    return {
        "RSI_RUNNER_ROLE": role,
        "RSI_RUNNER_HOST": "0.0.0.0",
        "RSI_RUNNER_PORT": "8090",
        "RSI_RUNNER_MODEL": "openai/gpt-5.4",
        "RSI_RUNNER_REASONING_EFFORT": "xhigh",
        "RSI_HERMES_PIN": "4a0358d2e741eb049a6ffb9b8e610db946a4fec5",
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
        self.assertEqual(result.raw["hermes_pin"], "4a0358d2e741eb049a6ffb9b8e610db946a4fec5")
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
        self.assertEqual(payload["tool_call_id"], "repo.context:1")
        self.assertEqual(len(binding.diagnostics()["tool_calls"]), 1)
        self.assertEqual(binding.diagnostics()["tool_calls"][0]["tool_name"], "repo.context")
        self.assertEqual(binding.diagnostics()["tool_calls"][0]["tool_call_id"], "repo.context:1")
        self.assertEqual(binding.diagnostics()["tool_calls"][0]["request"], {"trace_id": "trace-123", "repo": "rsi-agent-platform", "question": "What broke?"})
        self.assertEqual(binding.diagnostics()["tool_calls"][0]["summary"], "Context gathered.")

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

    def test_question_gather_treats_max_output_tokens_as_partial_bounded_stop(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "question_gather",
                    "repo": "depin-backend",
                    "prompt": json.dumps(
                        {
                            "investigation_spec": {
                                "user_request": "Did the linked Slack thread confirm the upload fix?",
                                "repo": "depin-backend",
                            },
                            "evidence_ledger": {
                                "reply_target": {"channel_id": "C123", "thread_ts": "171000001.000100"},
                                "evidence_items": [
                                    {
                                        "kind": "slack_message",
                                        "summary": "[alice] The linked thread discussed the upload fix.",
                                        "source_ref": "https://slack.example/messages/1",
                                        "tool_name": "slack.mcp.get_thread",
                                    }
                                ],
                                "open_questions": ["Need to confirm whether the linked thread contained a final resolution."],
                            },
                        }
                    ),
                    "system_message": "Use read-only tools and return JSON only.",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_read"}],
                }
            }
        )
        captured_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            body = json.loads(req.data.decode("utf-8"))
            captured_requests.append({"timeout": timeout, "body": body})
            return FakeHTTPResponse(
                {
                    "id": "resp-question-gather-1",
                    "status": "incomplete",
                    "incomplete_details": {"reason": "max_output_tokens"},
                    "usage": {
                        "output_tokens": 15000,
                        "output_tokens_details": {"reasoning_tokens": 14950},
                    },
                    "output": [{"type": "reasoning"}],
                }
            )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_resolved_task_mcp_servers", return_value=([], set())), mock.patch.object(
                runtime, "_direct_function_tools", return_value=[]
            ), mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen):
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(len(captured_requests), 1)
        self.assertEqual(captured_requests[0]["body"]["max_output_tokens"], 15000)
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["termination_reason"], "output_token_budget_exhausted")
        self.assertEqual(result.raw["provider_response_id"], "resp-question-gather-1")
        self.assertEqual(result.raw["runner_diagnostics"]["provider_status"], "incomplete")
        self.assertEqual(result.raw["runner_diagnostics"]["provider_incomplete_reason"], "max_output_tokens")
        self.assertEqual(result.raw["evidence_ledger"]["termination_reason"], "output_token_budget_exhausted")
        self.assertEqual(result.raw["evidence_ledger"]["evidence_items"][0]["tool_name"], "slack.mcp.get_thread")

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

    def test_native_mcp_preflight_rejects_invalid_function_schema_before_dispatch(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply in Slack.",
                    "allowed_tools": ["repo.context"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_read"}],
                }
            }
        )

        env = {
            **runner_env("prod"),
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_USER_TOKEN": "slack-mcp-test-token",
        }
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, env, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_resolved_task_mcp_servers", return_value=([], set())), mock.patch.object(
                runtime,
                "_direct_function_tools",
                return_value=[
                    {
                        "type": "function",
                        "name": "broken_tool",
                        "description": "Broken schema.",
                        "parameters": {
                            "type": "object",
                            "properties": {"repo": {"type": "string"}},
                        },
                        "strict": True,
                    }
                ],
            ), mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen") as urlopen:
                result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_invalid_tool_schema")
        self.assertEqual(result.raw["runner_diagnostics"]["failure_kind"], "invalid_tool_schema")
        self.assertEqual(result.raw["invalid_tool_schemas"][0]["tool_name"], "broken_tool")
        violation_kinds = {item["kind"] for item in result.raw["invalid_tool_schemas"][0]["violations"]}
        self.assertEqual(violation_kinds, {"additional_properties_not_false", "missing_required"})
        self.assertEqual(urlopen.call_count, 0)

    def test_native_mcp_workflow_uses_responses_with_function_tools_and_reply_delivery(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest Slack thread and reply in-thread.",
                    "system_message": "Return only valid JSON.",
                    "allowed_tools": ["repo.context", "slack.history", "slack.reply"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "context_summary": "Slack-bound workflow",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                    "mcp_servers": [
                        {
                            "server_label": "slack",
                            "profile": "slack_mcp_reply",
                            "headers": {"X-RSI-Channel-ID": "C123", "X-RSI-Thread-TS": "171000001.000100"},
                        }
                    ],
                }
            }
        )
        responses_requests: list[dict[str, object]] = []
        slack_methods: list[str] = []
        tool_gateway_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            body = json.loads(req.data.decode("utf-8"))
            if req.full_url == "https://mcp.slack.com/mcp":
                method = body["method"]
                slack_methods.append(method)
                if method == "initialize":
                    return FakeHTTPResponse({"result": {}})
                if method == "notifications/initialized":
                    return FakeHTTPResponse({})
                if method == "tools/list":
                    return FakeHTTPResponse(
                        {
                            "result": {
                                "tools": [
                                    {
                                        "name": "get_thread",
                                        "description": "Read a Slack thread.",
                                        "annotations": {"readOnlyHint": True},
                                    },
                                    {
                                        "name": "search_messages",
                                        "description": "Search Slack messages.",
                                        "annotations": {"readOnlyHint": True},
                                    },
                                    {
                                        "name": "send_message",
                                        "description": "Send a message to Slack.",
                                    },
                                ]
                            }
                        }
                    )
                raise AssertionError(f"unexpected Slack MCP method {method}")
            if req.full_url == "https://api.openai.com/v1/responses":
                responses_requests.append({"timeout": timeout, "body": body})
                if len(responses_requests) == 1:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-1",
                            "output": [
                                {
                                    "type": "mcp_call",
                                    "id": "mcp-read-1",
                                    "name": "get_thread",
                                    "arguments": json.dumps({"channel_id": "C123", "thread_ts": "171000001.000100"}),
                                    "status": "completed",
                                    "output": json.dumps({"messages": [{"text": "Thread summary"}]}),
                                },
                                {
                                    "type": "function_call",
                                    "id": "fc-1",
                                    "call_id": "repo-context-1",
                                    "name": "repo_context",
                                    "arguments": json.dumps({"repo": "rsi-agent-platform", "question": "What changed this week?"}),
                                },
                            ],
                        }
                    )
                return FakeHTTPResponse(
                    {
                        "id": "resp-native-2",
                        "output": [
                            {
                                "type": "mcp_call",
                                "id": "mcp-send-1",
                                "name": "send_message",
                                "arguments": json.dumps(
                                    {
                                        "channel_id": "C123",
                                        "thread_ts": "171000001.000100",
                                        "text": "Final reply from Slack MCP.",
                                    }
                                ),
                                "status": "completed",
                                "output": json.dumps({"ts": "171000001.000100", "text": "Final reply from Slack MCP."}),
                            }
                        ],
                        "output_text": json.dumps(
                            {
                                "visible_reasoning": [],
                                "reply_draft": "Final reply from Slack MCP.",
                                "final_answer": "Final reply from Slack MCP.",
                                "confidence": 0.87,
                                "context_summary": "Grounded in Slack and repo evidence.",
                                "self_critique": "None.",
                                "proposed_actions": [],
                                "knowledge_drafts": [],
                                "outcome_hypotheses": [],
                            }
                        ),
                    }
                )
            if req.full_url == "http://tool-gateway.internal:8080/api/tools/repo.context/execute":
                tool_gateway_requests.append(body)
                return FakeHTTPResponse(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Repo context loaded.",
                        "provider": "repo",
                        "provider_ref": "repo-context-1",
                        "tool_call_id": "repo.context:1",
                        "output": {"files": [{"path": "README.md", "summary": "Updated project overview."}]},
                    }
                )
            raise AssertionError(f"unexpected urlopen target {req.full_url}")

        env = {
            **runner_env("prod"),
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_USER_TOKEN": "slack-mcp-test-token",
        }
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen), mock.patch.dict(
            os.environ, {**env, "HERMES_HOME": tempdir}, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)
            log_path = result.raw["native_execution_log_path"]
            self.assertTrue(os.path.exists(log_path))
            with open(log_path, encoding="utf-8") as handle:
                events = [json.loads(line) for line in handle if line.strip()]

        self.assertTrue(result.ok)
        self.assertEqual(slack_methods, ["initialize", "notifications/initialized", "tools/list"])
        self.assertEqual(len(responses_requests), 2)
        self.assertEqual(responses_requests[0]["timeout"], 720)
        first_tools = responses_requests[0]["body"]["tools"]
        self.assertTrue(any(tool["type"] == "function" and tool["name"] == "repo_context" for tool in first_tools))
        mcp_tools = [tool for tool in first_tools if tool["type"] == "mcp"]
        self.assertEqual(len(mcp_tools), 1)
        self.assertEqual(mcp_tools[0]["server_label"], "slack")
        self.assertEqual(mcp_tools[0]["authorization"], "slack-mcp-test-token")
        self.assertEqual(
            mcp_tools[0]["allowed_tools"]["tool_names"],
            ["get_thread", "search_messages", "send_message"],
        )
        self.assertEqual(responses_requests[1]["body"]["previous_response_id"], "resp-native-1")
        self.assertEqual(responses_requests[1]["body"]["input"][0]["type"], "function_call_output")
        self.assertEqual(tool_gateway_requests[0]["repo"], "rsi-agent-platform")
        self.assertTrue(result.raw["native_mcp_enabled"])
        self.assertEqual(result.raw["runner_diagnostics"]["native_execution_mode"], "openai_responses_mcp")
        self.assertEqual(result.raw["reply_delivery"]["channel_id"], "C123")
        self.assertEqual(result.raw["reply_delivery"]["thread_ts"], "171000001.000100")
        self.assertEqual(result.raw["reply_delivery"]["body"], "Final reply from Slack MCP.")
        self.assertEqual(result.raw["reply_delivery"]["tool_name"], "slack.mcp.send_message")
        tool_names = [item["tool_name"] for item in result.raw["tool_calls"]]
        self.assertIn("repo.context", tool_names)
        self.assertIn("slack.mcp.get_thread", tool_names)
        self.assertIn("slack.mcp.send_message", tool_names)
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply from Slack MCP.")
        event_names = [item["event"] for item in events]
        self.assertIn("responses_request", event_names)
        self.assertIn("responses_response", event_names)
        self.assertIn("mcp_calls_observed", event_names)
        self.assertIn("function_call_outputs", event_names)
        self.assertEqual(events[-1]["event"], "execution_completed")

    def test_native_mcp_workflow_iteration_budget_exhaustion_recovers_partial_completion(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest Slack thread and reply in-thread.",
                    "system_message": "Return only valid JSON.",
                    "allowed_tools": ["repo.context"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "mcp_servers": [{"server_label": "notion", "server_url": "https://mcp.notion.com/mcp"}],
                }
            }
        )
        responses_requests: list[dict[str, object]] = []
        tool_gateway_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            body = json.loads(req.data.decode("utf-8"))
            if req.full_url == "https://api.openai.com/v1/responses":
                responses_requests.append({"timeout": timeout, "body": body})
                if len(responses_requests) == 1:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-1",
                            "output": [
                                {
                                    "type": "mcp_call",
                                    "id": "mcp-read-1",
                                    "name": "search",
                                    "arguments": json.dumps({"query": "Numo rollout status"}),
                                    "status": "completed",
                                    "output": json.dumps({"results": [{"title": "Numo TODO"}]}),
                                },
                                {
                                    "type": "function_call",
                                    "id": "fc-1",
                                    "call_id": "repo-context-1",
                                    "name": "repo_context",
                                    "arguments": json.dumps(
                                        {"repo": "rsi-agent-platform", "question": "What changed in the workflow code?"}
                                    ),
                                },
                            ],
                        }
                    )
                if len(responses_requests) == 2:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-2",
                            "output": [
                                {
                                    "type": "function_call",
                                    "id": "fc-2",
                                    "call_id": "repo-context-2",
                                    "name": "repo_context",
                                    "arguments": json.dumps(
                                        {"repo": "rsi-agent-platform", "question": "What is still missing?"}
                                    ),
                                }
                            ],
                        }
                    )
                return FakeHTTPResponse(
                    {
                        "id": "resp-partial-1",
                        "output_text": json.dumps(
                            partial_structured_output(
                                reply_text="Partial answer after hitting the native iteration cap.",
                                proposed_actions=[],
                            )
                        ),
                    }
                )
            if req.full_url == "http://tool-gateway.internal:8080/api/tools/repo.context/execute":
                tool_gateway_requests.append({"timeout": timeout, "body": body})
                return FakeHTTPResponse(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Repo context loaded.",
                        "provider": "repo",
                        "provider_ref": "repo-context-1",
                        "tool_call_id": "repo.context:1",
                        "output": {"files": [{"path": "runner/rsi_runner/hermes_runtime.py", "summary": "workflow logic"}]},
                    }
                )
            raise AssertionError(f"unexpected urlopen target {req.full_url}")

        env = {**runner_env("prod"), "RSI_RUNNER_PROD_MAX_ITERATIONS": "1"}
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen), mock.patch.dict(
            os.environ, {**env, "HERMES_HOME": tempdir}, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_resolved_task_mcp_servers", return_value=([], set())):
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["termination_reason"], "iteration_budget_exhausted")
        self.assertTrue(result.raw["max_iterations_reached"])
        self.assertEqual(result.raw["runner_diagnostics"]["completion_verdict"], "partial")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "direct_reducer")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_attempts"], 1)
        self.assertEqual(result.raw["runner_diagnostics"]["budget_used"], 1)
        self.assertEqual(result.raw["runner_diagnostics"]["budget_max"], 1)
        self.assertEqual(len(responses_requests), 3)
        self.assertEqual(responses_requests[0]["timeout"], 720)
        self.assertEqual(responses_requests[1]["body"]["previous_response_id"], "resp-native-1")
        self.assertEqual(responses_requests[2]["timeout"], 180)
        self.assertNotIn("tools", responses_requests[2]["body"])
        self.assertEqual(tool_gateway_requests[0]["body"]["repo"], "rsi-agent-platform")
        self.assertEqual(result.raw["mcp_calls"][0]["tool_name"], "slack.mcp.search")
        tool_names = [item["tool_name"] for item in result.raw["tool_calls"]]
        self.assertIn("repo.context", tool_names)
        self.assertIn("slack.mcp.search", tool_names)
        self.assertEqual(result.raw["structured_output"]["proposed_actions"][0]["kind"], "slack_post")
        self.assertTrue(result.raw["action_contract_repair_attempted"])
        self.assertTrue(result.raw["action_contract_repair_succeeded"])

    def test_native_mcp_workflow_max_output_tokens_recovers_partial_completion(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest Slack thread and reply in-thread.",
                    "allowed_tools": ["repo.context"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "mcp_servers": [{"server_label": "notion", "server_url": "https://mcp.notion.com/mcp"}],
                }
            }
        )
        responses_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            body = json.loads(req.data.decode("utf-8"))
            if req.full_url == "https://api.openai.com/v1/responses":
                responses_requests.append({"timeout": timeout, "body": body})
                if len(responses_requests) == 1:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-1",
                            "output": [
                                {
                                    "type": "function_call",
                                    "id": "fc-1",
                                    "call_id": "repo-context-1",
                                    "name": "repo_context",
                                    "arguments": json.dumps(
                                        {"repo": "rsi-agent-platform", "question": "What did the workflow collect?"}
                                    ),
                                }
                            ],
                        }
                    )
                if len(responses_requests) == 2:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-2",
                            "status": "incomplete",
                            "incomplete_details": {"reason": "max_output_tokens"},
                            "output": [
                                {
                                    "type": "mcp_call",
                                    "id": "mcp-read-1",
                                    "name": "fetch",
                                    "arguments": json.dumps({"id": "page-123"}),
                                    "status": "completed",
                                    "output": json.dumps({"title": "Numo TODO"}),
                                }
                            ],
                        }
                    )
                return FakeHTTPResponse(
                    {
                        "id": "resp-partial-2",
                        "output_text": json.dumps(
                            partial_structured_output(
                                reply_text="Partial answer after hitting the output token cap.",
                                proposed_actions=[],
                            )
                        ),
                    }
                )
            if req.full_url == "http://tool-gateway.internal:8080/api/tools/repo.context/execute":
                return FakeHTTPResponse(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Repo context loaded.",
                        "provider": "repo",
                        "provider_ref": "repo-context-1",
                        "tool_call_id": "repo.context:1",
                        "output": {"files": [{"path": "README.md", "summary": "Updated overview."}]},
                    }
                )
            raise AssertionError(f"unexpected urlopen target {req.full_url}")

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch(
            "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_resolved_task_mcp_servers", return_value=([], set())):
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["termination_reason"], "output_token_budget_exhausted")
        self.assertEqual(result.raw["provider_incomplete_reason"], "max_output_tokens")
        self.assertEqual(result.raw["runner_diagnostics"]["completion_verdict"], "partial")
        self.assertEqual(result.raw["runner_diagnostics"]["termination_reason"], "output_token_budget_exhausted")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "direct_reducer")
        self.assertEqual(len(responses_requests), 3)
        self.assertEqual(result.raw["mcp_calls"][0]["tool_name"], "slack.mcp.fetch")
        tool_names = [item["tool_name"] for item in result.raw["tool_calls"]]
        self.assertIn("repo.context", tool_names)
        self.assertIn("slack.mcp.fetch", tool_names)
        self.assertEqual(result.raw["structured_output"]["proposed_actions"][0]["kind"], "slack_post")

    def test_native_mcp_workflow_iteration_budget_exhaustion_after_reply_delivery_fails_closed(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest Slack thread and reply in-thread.",
                    "allowed_tools": ["repo.context"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_reply"}],
                }
            }
        )
        responses_seen = 0

        def fake_urlopen(req, timeout: int = 0):
            nonlocal responses_seen
            body = json.loads(req.data.decode("utf-8"))
            if req.full_url == "https://mcp.slack.com/mcp":
                if body["method"] == "initialize":
                    return FakeHTTPResponse({"result": {}})
                if body["method"] == "notifications/initialized":
                    return FakeHTTPResponse({})
                if body["method"] == "tools/list":
                    return FakeHTTPResponse(
                        {
                            "result": {
                                "tools": [
                                    {
                                        "name": "get_thread",
                                        "description": "Read a Slack thread.",
                                        "annotations": {"readOnlyHint": True},
                                    },
                                    {"name": "send_message", "description": "Send a message to Slack."},
                                ]
                            }
                        }
                    )
                raise AssertionError(f"unexpected Slack MCP method {body['method']}")
            if req.full_url == "https://api.openai.com/v1/responses":
                responses_seen += 1
                if responses_seen == 1:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-budget-1",
                            "output": [
                                {
                                    "type": "mcp_call",
                                    "id": "mcp-send-1",
                                    "name": "send_message",
                                    "arguments": json.dumps(
                                        {
                                            "channel_id": "C123",
                                            "thread_ts": "171000001.000100",
                                            "text": "Final reply from Slack MCP.",
                                        }
                                    ),
                                    "status": "completed",
                                    "output": json.dumps({"ts": "171000001.000100", "text": "Final reply from Slack MCP."}),
                                },
                                {
                                    "type": "function_call",
                                    "id": "fc-budget-1",
                                    "call_id": "repo-context-budget-1",
                                    "name": "repo_context",
                                    "arguments": json.dumps({"repo": "rsi-agent-platform", "question": "What changed?"}),
                                },
                            ],
                        }
                    )
                return FakeHTTPResponse(
                    {
                        "id": "resp-native-budget-2",
                        "output": [
                            {
                                "type": "function_call",
                                "id": "fc-budget-2",
                                "call_id": "repo-context-budget-2",
                                "name": "repo_context",
                                "arguments": json.dumps({"repo": "rsi-agent-platform", "question": "What else changed?"}),
                            }
                        ],
                    }
                )
            if req.full_url == "http://tool-gateway.internal:8080/api/tools/repo.context/execute":
                return FakeHTTPResponse(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Repo context loaded.",
                        "provider": "repo",
                        "provider_ref": "repo-context-budget-1",
                        "tool_call_id": "repo.context:1",
                        "output": {"files": []},
                    }
                )
            raise AssertionError(f"unexpected runtime target {req.full_url}")

        env = {
            **runner_env("prod"),
            "RSI_RUNNER_PROD_MAX_ITERATIONS": "1",
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_USER_TOKEN": "slack-mcp-test-token",
        }
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch(
            "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_reply_delivery_uncertain")
        self.assertTrue(result.raw["reply_delivery_attempted"])
        self.assertEqual(result.raw["reply_delivery"]["channel_id"], "C123")
        self.assertEqual(responses_seen, 2)

    def test_native_mcp_workflow_max_output_tokens_after_reply_delivery_fails_closed(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest Slack thread and reply in-thread.",
                    "allowed_tools": ["repo.context"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_reply"}],
                }
            }
        )
        responses_seen = 0

        def fake_urlopen(req, timeout: int = 0):
            nonlocal responses_seen
            body = json.loads(req.data.decode("utf-8"))
            if req.full_url == "https://mcp.slack.com/mcp":
                if body["method"] == "initialize":
                    return FakeHTTPResponse({"result": {}})
                if body["method"] == "notifications/initialized":
                    return FakeHTTPResponse({})
                if body["method"] == "tools/list":
                    return FakeHTTPResponse(
                        {
                            "result": {
                                "tools": [
                                    {
                                        "name": "get_thread",
                                        "description": "Read a Slack thread.",
                                        "annotations": {"readOnlyHint": True},
                                    },
                                    {"name": "send_message", "description": "Send a message to Slack."},
                                ]
                            }
                        }
                    )
                raise AssertionError(f"unexpected Slack MCP method {body['method']}")
            if req.full_url == "https://api.openai.com/v1/responses":
                responses_seen += 1
                if responses_seen == 1:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-output-1",
                            "output": [
                                {
                                    "type": "mcp_call",
                                    "id": "mcp-send-1",
                                    "name": "send_message",
                                    "arguments": json.dumps(
                                        {
                                            "channel_id": "C123",
                                            "thread_ts": "171000001.000100",
                                            "text": "Final reply from Slack MCP.",
                                        }
                                    ),
                                    "status": "completed",
                                    "output": json.dumps({"ts": "171000001.000100", "text": "Final reply from Slack MCP."}),
                                }
                            ],
                        }
                    )
                return FakeHTTPResponse(
                    {
                        "id": "resp-native-output-2",
                        "status": "incomplete",
                        "incomplete_details": {"reason": "max_output_tokens"},
                        "output": [
                            {
                                "type": "function_call",
                                "id": "fc-output-1",
                                "call_id": "repo-context-output-1",
                                "name": "repo_context",
                                "arguments": json.dumps({"repo": "rsi-agent-platform", "question": "What changed?"}),
                            }
                        ],
                    }
                )
            raise AssertionError(f"unexpected runtime target {req.full_url}")

        env = {
            **runner_env("prod"),
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_USER_TOKEN": "slack-mcp-test-token",
        }
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch(
            "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_reply_delivery_uncertain")
        self.assertTrue(result.raw["reply_delivery_attempted"])
        self.assertEqual(result.raw["reply_delivery"]["channel_id"], "C123")
        self.assertEqual(result.raw["provider_incomplete_reason"], "max_output_tokens")
        self.assertEqual(responses_seen, 2)
    def test_native_mcp_workflow_request_timeout_recovers_partial_completion_with_reserved_headroom(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest Slack thread and reply in-thread.",
                    "allowed_tools": ["repo.context"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "mcp_servers": [{"server_label": "notion", "server_url": "https://mcp.notion.com/mcp"}],
                }
            }
        )
        responses_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            body = json.loads(req.data.decode("utf-8"))
            if req.full_url == "https://api.openai.com/v1/responses":
                responses_requests.append({"timeout": timeout, "body": body})
                if len(responses_requests) == 1:
                    raise TimeoutError("native request timed out")
                return FakeHTTPResponse(
                    {
                        "id": "resp-partial-timeout",
                        "output_text": json.dumps(
                            partial_structured_output(
                                reply_text="Partial answer after native request timeout.",
                                proposed_actions=[],
                            )
                        ),
                    }
                )
            raise AssertionError(f"unexpected urlopen target {req.full_url}")

        env = {**runner_env("prod"), "RSI_RUNNER_PROD_TASK_TIMEOUT": "20s"}
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch(
            "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_resolved_task_mcp_servers", return_value=([], set())):
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["termination_reason"], "task_timeout")
        self.assertEqual(result.raw["timeout_kind"], "task_timeout")
        self.assertEqual(result.raw["task_timeout_seconds"], 20)
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "direct_reducer")
        self.assertEqual(len(responses_requests), 2)
        self.assertEqual(responses_requests[0]["timeout"], 10)
        self.assertEqual(responses_requests[1]["timeout"], 10)
        self.assertNotIn("tools", responses_requests[1]["body"])
        self.assertEqual(result.raw["structured_output"]["proposed_actions"][0]["kind"], "slack_post")

    def test_native_mcp_workflow_iteration_budget_exhaustion_still_fails_when_partial_recovery_is_ineligible(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the latest workflow state.",
                    "allowed_tools": ["repo.context"],
                    "mcp_servers": [{"server_label": "notion", "server_url": "https://mcp.notion.com/mcp"}],
                }
            }
        )
        responses_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            body = json.loads(req.data.decode("utf-8"))
            if req.full_url == "https://api.openai.com/v1/responses":
                responses_requests.append({"timeout": timeout, "body": body})
                if len(responses_requests) == 1:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-1",
                            "output": [
                                {
                                    "type": "function_call",
                                    "id": "fc-1",
                                    "call_id": "repo-context-1",
                                    "name": "repo_context",
                                    "arguments": json.dumps(
                                        {"repo": "rsi-agent-platform", "question": "What changed in the workflow code?"}
                                    ),
                                }
                            ],
                        }
                    )
                return FakeHTTPResponse(
                    {
                        "id": "resp-native-2",
                        "output": [
                            {
                                "type": "function_call",
                                "id": "fc-2",
                                "call_id": "repo-context-2",
                                "name": "repo_context",
                                "arguments": json.dumps(
                                    {"repo": "rsi-agent-platform", "question": "What is still missing?"}
                                ),
                            }
                        ],
                    }
                )
            if req.full_url == "http://tool-gateway.internal:8080/api/tools/repo.context/execute":
                return FakeHTTPResponse(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Repo context loaded.",
                        "provider": "repo",
                        "provider_ref": "repo-context-1",
                        "tool_call_id": "repo.context:1",
                        "output": {"files": [{"path": "runner/rsi_runner/hermes_runtime.py", "summary": "workflow logic"}]},
                    }
                )
            raise AssertionError(f"unexpected urlopen target {req.full_url}")

        env = {**runner_env("prod"), "RSI_RUNNER_PROD_MAX_ITERATIONS": "1"}
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch(
            "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_resolved_task_mcp_servers", return_value=([], set())):
                result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_iteration_budget_exhausted")
        self.assertTrue(result.raw["native_mcp_enabled"])
        self.assertIn("function-call budget", result.message)
        self.assertEqual(len(responses_requests), 2)

    def test_native_mcp_reply_profile_fails_closed_when_send_tool_discovery_is_ambiguous(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply in Slack.",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_reply"}],
                }
            }
        )
        runtime_requests: list[str] = []

        def fake_runtime_urlopen(req, timeout: int = 0):
            body = json.loads(req.data.decode("utf-8"))
            runtime_requests.append(body["method"])
            if body["method"] == "initialize":
                return FakeHTTPResponse({"result": {}})
            if body["method"] == "notifications/initialized":
                return FakeHTTPResponse({})
            if body["method"] == "tools/list":
                return FakeHTTPResponse(
                    {
                        "result": {
                            "tools": [
                                {"name": "send_message", "description": "Send a message to Slack."},
                                {"name": "post_message", "description": "Post a message to Slack."},
                            ]
                        }
                    }
                )
            raise AssertionError(f"unexpected method {body['method']}")

        env = {
            **runner_env("prod"),
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_USER_TOKEN": "slack-mcp-test-token",
        }
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch(
            "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_runtime_urlopen
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(runtime_requests, ["initialize", "notifications/initialized", "tools/list"])
        self.assertEqual(result.raw["failure_class"], "runner_non_ok")
        self.assertTrue(result.raw["native_mcp_enabled"])
        self.assertIn("expected exactly one candidate", result.message)

    def test_resolved_task_mcp_servers_supports_custom_authorization_env_var(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Search Notion.",
                    "mcp_servers": [
                        {
                            "server_label": "notion",
                            "server_url": "https://mcp.notion.com/mcp",
                            "authorization_env_var": "RSI_NOTION_MCP_AUTHORIZATION",
                            "allowed_tools": {"tool_names": ["search", "fetch"]},
                        }
                    ],
                }
            }
        )

        env = {
            **runner_env("prod"),
            "RSI_NOTION_MCP_AUTHORIZATION": "notion-oauth-token",
        }
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, env, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            resolved, send_tool_names = runtime._resolved_task_mcp_servers(task)

        self.assertEqual(send_tool_names, set())
        self.assertEqual(len(resolved), 1)
        self.assertEqual(resolved[0]["server_label"], "notion")
        self.assertEqual(resolved[0]["server_url"], "https://mcp.notion.com/mcp")
        self.assertEqual(resolved[0]["authorization"], "notion-oauth-token")
        self.assertEqual(resolved[0]["allowed_tools"]["tool_names"], ["search", "fetch"])

    def test_resolved_task_mcp_servers_supports_header_env_vars(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Search Notion.",
                    "mcp_servers": [
                        {
                            "server_label": "notion",
                            "server_url": "https://mcp.notion.com/mcp",
                            "header_env_vars": {
                                "CF-Access-Client-Id": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID",
                                "CF-Access-Client-Secret": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_SECRET",
                            },
                            "headers": {"X-Test": "static"},
                        }
                    ],
                }
            }
        )

        env = {
            **runner_env("prod"),
            "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID": "client-id",
            "RSI_NOTION_MCP_CF_ACCESS_CLIENT_SECRET": "client-secret",
        }
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, env, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            resolved, send_tool_names = runtime._resolved_task_mcp_servers(task)

        self.assertEqual(send_tool_names, set())
        self.assertEqual(
            resolved[0]["headers"],
            {
                "X-Test": "static",
                "CF-Access-Client-Id": "client-id",
                "CF-Access-Client-Secret": "client-secret",
            },
        )

    def test_resolved_task_mcp_servers_allows_custom_server_without_authorization(self) -> None:
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
            resolved, _ = runtime._resolved_task_mcp_servers(task)

        self.assertEqual(len(resolved), 1)
        self.assertNotIn("authorization", resolved[0])
        self.assertEqual(resolved[0]["server_label"], "docs")
        self.assertEqual(resolved[0]["allowed_tools"], {"read_only": True})

    def test_resolved_task_mcp_servers_fails_when_custom_authorization_env_var_is_missing(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Search Notion.",
                    "mcp_servers": [
                        {
                            "server_label": "notion",
                            "server_url": "https://mcp.notion.com/mcp",
                            "authorization_env_var": "RSI_NOTION_MCP_AUTHORIZATION",
                        }
                    ],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with self.assertRaisesRegex(RuntimeError, "RSI_NOTION_MCP_AUTHORIZATION"):
                runtime._resolved_task_mcp_servers(task)

    def test_resolved_task_mcp_servers_fails_when_header_env_var_is_missing(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Search Notion.",
                    "mcp_servers": [
                        {
                            "server_label": "notion",
                            "server_url": "https://mcp.notion.com/mcp",
                            "header_env_vars": {
                                "CF-Access-Client-Id": "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID",
                            },
                        }
                    ],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with self.assertRaisesRegex(RuntimeError, "RSI_NOTION_MCP_CF_ACCESS_CLIENT_ID"):
                runtime._resolved_task_mcp_servers(task)

    def test_native_mcp_write_timeout_returns_reply_delivery_uncertain(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the thread and reply in-thread.",
                    "allowed_tools": ["repo.context"],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_reply"}],
                }
            }
        )
        responses_seen = 0

        def fake_urlopen(req, timeout: int = 0):
            nonlocal responses_seen
            body = json.loads(req.data.decode("utf-8"))
            if req.full_url == "https://mcp.slack.com/mcp":
                if body["method"] == "initialize":
                    return FakeHTTPResponse({"result": {}})
                if body["method"] == "notifications/initialized":
                    return FakeHTTPResponse({})
                if body["method"] == "tools/list":
                    return FakeHTTPResponse(
                        {
                            "result": {
                                "tools": [
                                    {
                                        "name": "get_thread",
                                        "description": "Read a Slack thread.",
                                        "annotations": {"readOnlyHint": True},
                                    },
                                    {"name": "send_message", "description": "Send a message to Slack."},
                                ]
                            }
                        }
                    )
                raise AssertionError(f"unexpected Slack MCP method {body['method']}")
            if req.full_url == "https://api.openai.com/v1/responses":
                responses_seen += 1
                if responses_seen == 1:
                    return FakeHTTPResponse(
                        {
                            "id": "resp-native-timeout-1",
                            "output": [
                                {
                                    "type": "mcp_call",
                                    "id": "mcp-send-1",
                                    "name": "send_message",
                                    "arguments": json.dumps(
                                        {
                                            "channel_id": "C123",
                                            "thread_ts": "171000001.000100",
                                            "text": "Final reply from Slack MCP.",
                                        }
                                    ),
                                    "status": "completed",
                                    "output": json.dumps({"ts": "171000001.000100", "text": "Final reply from Slack MCP."}),
                                },
                                {
                                    "type": "function_call",
                                    "id": "fc-timeout-1",
                                    "call_id": "repo-context-timeout-1",
                                    "name": "repo_context",
                                    "arguments": json.dumps({"repo": "rsi-agent-platform", "question": "What changed?"}),
                                },
                            ],
                        }
                    )
                raise TimeoutError("simulated timeout after Slack write")
            if req.full_url == "http://tool-gateway.internal:8080/api/tools/repo.context/execute":
                return FakeHTTPResponse(
                    {
                        "status": "completed",
                        "available": True,
                        "summary": "Repo context loaded.",
                        "provider": "repo",
                        "provider_ref": "repo-context-timeout-1",
                        "tool_call_id": "repo.context:1",
                        "output": {"files": []},
                    }
                )
            raise AssertionError(f"unexpected runtime target {req.full_url}")

        env = {
            **runner_env("prod"),
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_USER_TOKEN": "slack-mcp-test-token",
        }
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch(
            "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen
        ), mock.patch.dict(
            os.environ, env, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_reply_delivery_uncertain")
        self.assertTrue(result.raw["native_mcp_enabled"])
        self.assertTrue(result.raw["reply_delivery_attempted"])
        self.assertEqual(result.raw["reply_delivery"]["channel_id"], "C123")
        self.assertEqual(result.raw["reply_delivery"]["body"], "Final reply from Slack MCP.")
        self.assertEqual(result.raw["reply_delivery"]["tool_name"], "slack.mcp.send_message")

    def test_native_question_expand_guards_json_input_prompt(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "question_expand",
                    "repo": "depin-backend",
                    "prompt": json.dumps(
                        {
                            "investigation_spec": {
                                "user_request": "How did depin-backend api progress this week for numo?",
                                "repo": "depin-backend",
                                "project_key": "numo",
                            },
                            "evidence_ledger": {
                                "evidence_items": [],
                                "open_questions": ["Need grounded Slack evidence from linked channels."],
                            },
                        }
                    ),
                    "system_message": "Use only governed read-only tools. Return only JSON with tool_calls, evidence_items, open_questions, insufficiency_markers, and confidence.",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_read"}],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                }
            }
        )
        captured_requests: list[dict[str, object]] = []

        def fake_urlopen(req, timeout: int = 0):
            self.assertEqual(req.full_url, "https://api.openai.com/v1/responses")
            body = json.loads(req.data.decode("utf-8"))
            captured_requests.append({"timeout": timeout, "body": body})
            return FakeHTTPResponse(
                {
                    "id": "resp-question-expand-1",
                    "output_text": json.dumps(
                        {
                            "tool_calls": [],
                            "evidence_items": [],
                            "open_questions": ["Need grounded Slack evidence from linked channels."],
                            "insufficiency_markers": ["slack evidence missing"],
                            "confidence": 0.41,
                        }
                    ),
                }
            )

        env = {
            **runner_env("prod"),
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_USER_TOKEN": "slack-mcp-test-token",
        }
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ, env, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_resolved_task_mcp_servers", return_value=([], set())), mock.patch(
                "rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen
            ):
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(len(captured_requests), 1)
        request_body = captured_requests[0]["body"]
        self.assertEqual(request_body["text"]["format"]["type"], "json_object")
        self.assertIn("json", request_body["input"].lower())
        self.assertTrue(request_body["input"].startswith("Return a JSON object only."))
        self.assertIn('"investigation_spec"', request_body["input"])
        self.assertEqual(
            result.raw["structured_output"]["insufficiency_markers"],
            ["slack evidence missing"],
        )

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
        self.assertEqual(runtime.metadata["hermes_pin"], "4a0358d2e741eb049a6ffb9b8e610db946a4fec5")
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
