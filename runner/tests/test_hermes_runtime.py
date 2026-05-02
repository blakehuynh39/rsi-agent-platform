from __future__ import annotations

import base64
import io
import json
import os
from pathlib import Path
import sqlite3
import subprocess
import tempfile
import threading
import time
import types
import unittest
from unittest import mock

from rsi_runner.config import RunnerConfig, RunnerConfigError
from rsi_runner.execution_contract import HermesCompanyComputer
from rsi_runner.hermes_adapter import HermesAdapter, _build_plugin_module
from rsi_runner.hermes_agent_adapter import HermesAgentAdapter, HermesContractStatus, validate_hermes_contract
from rsi_runner.hermes_mcp_adapter import TaskScopedMCPCleanupResult, TaskScopedMCPRegistration, TaskScopedMCPServer
from rsi_runner.hermes_runtime import (
    HermesExecutionResult,
    HermesRuntime,
    RunnerTaskRequest,
    _NativeLifecycleTailer,
    _redact_json_value,
    _redact_subprocess_output,
)
from rsi_runner.observability import ObservationEmitter, execution_observation_id
from rsi_runner.rsi_tools import transport_tool_schema
from rsi_runner.session_manager import SessionManager


HERMES_TEST_PIN = "4712c9d34610e5bb8729b6989019205f7c1cfd26"


def runner_env(role: str = "prod") -> dict[str, str]:
    return {
        "RSI_RUNNER_ROLE": role,
        "RSI_RUNNER_HOST": "0.0.0.0",
        "RSI_RUNNER_PORT": "8090",
        "RSI_RUNNER_MODEL": "openrouter/deepseek/deepseek-v4-pro",
        "RSI_RUNNER_REASONING_EFFORT": "xhigh",
        "RSI_OPENROUTER_PROVIDER_ONLY": "deepseek",
        "RSI_OPENROUTER_PROVIDER_ORDER": "deepseek",
        "RSI_OPENROUTER_REQUIRE_PARAMETERS": "true",
        "RSI_HERMES_PIN": HERMES_TEST_PIN,
        "RSI_RUNNER_PUBLIC_BASE_URL": "https://staging-rsi-platform.storyprotocol.net",
        "HERMES_HOME": "/tmp/hermes",
        "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": "/tmp/hermes/workspace",
        "RSI_EXECUTION_ENVELOPE_V1_ENABLED": "true",
        "RSI_RUNNER_PLANNER_MODE": "runner_first",
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
        "RSI_RUNNER_PROD_TASK_TIMEOUT": "1800s",
        "RSI_RUNNER_EVAL_INACTIVITY_TIMEOUT": "240s",
        "RSI_RUNNER_PROPOSAL_INACTIVITY_TIMEOUT": "360s",
        "RSI_RUNNER_PROD_TIMEOUT": "1830s",
        "RSI_RUNNER_PROACTIVE_TIMEOUT": "60s",
        "RSI_RUNNER_EVAL_TIMEOUT": "330s",
        "RSI_RUNNER_PROPOSAL_TIMEOUT": "450s",
        "RSI_RUNNER_NATIVE_MAX_OUTPUT_TOKENS": "15000",
        "HONCHO_API_KEY": "honcho-test-key",
        "OPENROUTER_API_KEY": "openrouter-test-key",
        "SLACK_BOT_TOKEN": "xoxb-test",
    }


def openrouter_runner_env(role: str = "prod") -> dict[str, str]:
    return runner_env(role)


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
        hermes_home = Path(_config.hermes_home)
        hermes_home.mkdir(parents=True, exist_ok=True)
        hermes_home.joinpath("config.yaml").write_text(
            "plugins:\n  enabled:\n    - rsi_context_engine\n",
            encoding="utf-8",
        )

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


class FakeObservationSinkResponse:
    status = 200

    def read(self) -> bytes:
        return b'{"status":"ok"}'


class FakeObservationSinkConnection:
    requests: list[dict[str, object]] = []

    def __init__(self, netloc: str, timeout: int | float | None = None) -> None:
        self.netloc = netloc
        self.timeout = timeout

    def request(self, method: str, path: str, body: bytes | None = None, headers: dict[str, str] | None = None) -> None:
        type(self).requests.append(
            {
                "netloc": self.netloc,
                "timeout": self.timeout,
                "method": method,
                "path": path,
                "body": body or b"",
                "headers": headers or {},
            }
        )

    def getresponse(self) -> FakeObservationSinkResponse:
        return FakeObservationSinkResponse()

    def close(self) -> None:
        return None


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
        self._runtime_contract_patch = mock.patch(
            "rsi_runner.hermes_runtime.validate_hermes_contract",
            return_value=HermesContractStatus(
                ok=True,
                expected_pin=HERMES_TEST_PIN,
                installed_commit=HERMES_TEST_PIN,
                hermes_version="test",
                api_signature_status="ok",
                pin_status="ok",
                plugin_status="ok",
                required_toolsets=["hermes-api-server"],
                toolset_status={"hermes-api-server": "ok"},
                session_db_status="ok",
                errors=[],
                checked_at_unix=1.0,
            ),
        )
        self._runtime_contract_patch.start()
        self.addCleanup(self._runtime_contract_patch.stop)

    def test_config_requires_explicit_env(self) -> None:
        with mock.patch.dict(os.environ, {}, clear=True):
            with self.assertRaises(RunnerConfigError):
                RunnerConfig.from_env()

    def test_config_reads_explicit_openrouter_xhigh_and_honcho(self) -> None:
        with mock.patch.dict(os.environ, runner_env("eval"), clear=True):
            config = RunnerConfig.from_env()

        self.assertEqual(config.model, "openrouter/deepseek/deepseek-v4-pro")
        self.assertEqual(config.reasoning_effort, "xhigh")
        self.assertEqual(config.memory_backend, "honcho")
        self.assertEqual(config.honcho_workspace, "rsi-stage")
        self.assertEqual(config.honcho_environment_effective, "production")
        self.assertEqual(config.native_max_output_tokens, 15000)

    def test_config_reads_openrouter_routing_and_requires_key(self) -> None:
        env = openrouter_runner_env("eval")
        with mock.patch.dict(os.environ, env, clear=True):
            config = RunnerConfig.from_env()

        self.assertEqual(config.model, "openrouter/deepseek/deepseek-v4-pro")
        self.assertTrue(config.openrouter_api_key_configured)
        self.assertEqual(
            config.openrouter_provider_routing,
            {
                "only": ["deepseek"],
                "order": ["deepseek"],
                "require_parameters": True,
            },
        )

        env.pop("OPENROUTER_API_KEY")
        with mock.patch.dict(os.environ, env, clear=True):
            with self.assertRaisesRegex(RunnerConfigError, "OPENROUTER_API_KEY"):
                RunnerConfig.from_env()

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
                "RSI_RUNTIME_OBSERVATION_SINK_URL": "http://control-plane.internal:8080/internal/runtime/observations",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": "/var/lib/hermes-executor",
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "TERMINAL_ENV": "local",
                "TERMINAL_CWD": "/var/lib/hermes-executor/company",
                "TERMINAL_TIMEOUT": "240",
                "TERMINAL_LIFETIME_SECONDS": "1200",
                "TERMINAL_LOCAL_PERSISTENT": "false",
                "RSI_HERMES_COMPANY_BIN_DIR": "/var/lib/hermes-executor/company/.rsi/bin",
                "RSI_HERMES_KUBERNETES_CONTEXT_ENABLED": "true",
                "KUBECONFIG": "/var/lib/hermes-executor/company/.rsi/kube/config",
            },
            clear=True,
        ):
            config = RunnerConfig.from_env()

        self.assertTrue(config.hermes_executor_enabled)
        self.assertTrue(config.hermes_executor_service_only)
        self.assertEqual(config.runtime_observation_sink_url, "http://control-plane.internal:8080/internal/runtime/observations")
        self.assertEqual(config.hermes_executor_workspace_root, "/var/lib/hermes-executor")
        self.assertEqual(config.hermes_computer_root, "/var/lib/hermes-executor/company")
        self.assertEqual(config.hermes_run_root, "/var/lib/hermes-executor/company/.rsi/runs")
        self.assertEqual(config.hermes_artifact_root, "/var/lib/hermes-executor/company/artifacts")
        self.assertTrue(config.hermes_native_terminal_enabled)
        self.assertEqual(config.hermes_native_toolsets, ["terminal", "file"])
        self.assertEqual(config.hermes_terminal_env, "local")
        self.assertEqual(config.hermes_terminal_cwd, "/var/lib/hermes-executor/company")
        self.assertEqual(config.hermes_terminal_timeout_seconds, 240)
        self.assertEqual(config.hermes_terminal_lifetime_seconds, 1200)
        self.assertFalse(config.hermes_terminal_local_persistent)
        self.assertEqual(config.hermes_company_bin_dir, "/var/lib/hermes-executor/company/.rsi/bin")
        self.assertTrue(config.hermes_kubernetes_context_enabled)
        self.assertEqual(config.hermes_kubeconfig_path, "/var/lib/hermes-executor/company/.rsi/kube/config")

    def test_config_requires_observation_sink_for_service_only_executor(self) -> None:
        with mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_SERVICE_ONLY": "true",
            },
            clear=True,
        ):
            with self.assertRaisesRegex(RunnerConfigError, "RSI_RUNTIME_OBSERVATION_SINK_URL"):
                RunnerConfig.from_env()

    def test_observation_emitter_posts_to_direct_sink(self) -> None:
        FakeObservationSinkConnection.requests = []
        with mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "RSI_RUNTIME_OBSERVATION_SINK_URL": "http://control-plane.internal:8080/internal/runtime/observations"},
            clear=True,
        ), mock.patch("rsi_runner.observability.http.client.HTTPConnection", FakeObservationSinkConnection):
            config = RunnerConfig.from_env()
            emitter = ObservationEmitter.create(
                config,
                trace_id="trace-obs",
                workflow_id="wf-obs",
                operation_id="op-obs",
                role="prod",
                hermes_session_id="session-obs",
            )
            emitter.emit(phase="main", event_type="model.reasoning.delta", status="streaming", payload={"delta": "hello"})

        self.assertEqual(emitter.sink_status, "ok")
        self.assertEqual(len(FakeObservationSinkConnection.requests), 1)
        request = FakeObservationSinkConnection.requests[0]
        self.assertEqual(request["netloc"], "control-plane.internal:8080")
        self.assertEqual(request["path"], "/internal/runtime/observations")
        body = json.loads(request["body"])
        self.assertEqual(body["trace_id"], "trace-obs")
        self.assertEqual(body["event_type"], "model.reasoning.delta")
        self.assertEqual(body["payload"]["delta"], "hello")

    def test_context_engine_plugin_module_compiles_as_python(self) -> None:
        source = _build_plugin_module()

        compile(source, "rsi_context_engine/__init__.py", "exec")
        self.assertNotIn(": false", source)
        self.assertIn(": False", source)
        self.assertIn("artifact.write.started", source)
        self.assertIn("artifact.write.completed", source)
        self.assertIn("artifact_write_file", source)

    def test_runner_task_reads_kubernetes_namespace_scope(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Inspect runtime.",
                    "kubernetes_read_namespaces": ["story", "rsi-platform"],
                }
            }
        )

        self.assertEqual(task.kubernetes_read_namespaces, ["story", "rsi-platform"])

    def test_task_prompt_advertises_kubernetes_read_scope(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Inspect runtime.",
                    "kubernetes_read_namespaces": ["story", "rsi-platform"],
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            runner_env("prod"),
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        prompt = runtime._render_task_prompt(task)

        self.assertIn("Kubernetes read namespace scope: story, rsi-platform", prompt)
        self.assertIn("do not probe unlisted or archived namespaces", prompt)



    def test_native_workflow_enables_configured_company_computer_toolsets(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Inspect the company computer.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-terminal",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "RSI_HERMES_NATIVE_TOOLSETS": "hermes-api-server",
                "TERMINAL_CWD": "/tmp/hermes/workspace/company",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        toolsets = runtime._native_toolsets_for_task(task)
        phase_contract = runtime._native_executor_phase_contract(task, toolsets)

        self.assertIn("hermes-api-server", toolsets)
        self.assertIn("hermes-api-server", phase_contract["required_toolsets"])

    def test_native_terminal_refuses_pod_github_token_without_app_credentials(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "repo_allowlist": ["rsi-agent-platform", "depin-backend"],
                    "prompt": "Create the requested GitHub issue.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-gh",
                }
            }
        )
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": str(Path(tempdir, "hermes-home")),
                "RSI_HERMES_COMPUTER_ROOT": str(Path(tempdir, "company")),
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "GH_TOKEN": "github-installation-token",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            env, status = runtime._github_cli_environment(task)

        self.assertEqual(env, {})
        self.assertFalse(status["configured"])
        self.assertEqual(status["status"], "failed")
        self.assertEqual(status["reason"], "missing_github_app_credentials")
        self.assertIn("RSI_GITHUB_APP_ID", status["missing"])
        self.assertNotIn("token", status)

    def test_native_terminal_mints_github_cli_credentials_from_app_environment(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "repo_allowlist": ["rsi-agent-platform", "story-deployments"],
                    "prompt": "Read the private repo.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-gh-app",
                }
            }
        )
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.object(
            HermesRuntime,
            "_github_app_installation_token",
            return_value=(
                "github-app-token",
                {
                    "status": "configured",
                    "reason": "github_app_installation",
                    "expires_at": "2026-05-01T08:00:00Z",
                    "installation_id": "123754864",
                    "owner": "piplabs",
                },
            ),
        ) as token_mock, mock.patch.object(
            HermesRuntime,
            "_github_app_installation_token_for_owner",
            return_value=(
                "storyprotocol-token",
                {
                    "status": "configured",
                    "reason": "github_app_installation",
                    "expires_at": "2026-05-01T08:05:00Z",
                    "installation_id": "123754958",
                    "owner": "storyprotocol",
                },
            ),
        ) as owner_token_mock, mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": str(Path(tempdir, "hermes-home")),
                "RSI_HERMES_COMPUTER_ROOT": str(Path(tempdir, "company")),
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "RSI_GITHUB_OWNER": "piplabs",
                "RSI_GITHUB_APP_ID": "3370196",
                "RSI_GITHUB_APP_INSTALLATION_ID": "123754864",
                "RSI_GITHUB_APP_PRIVATE_KEY": "private-key",
                "RSI_GITHUB_REPO_OWNERS": "story-deployments=storyprotocol,story-infra-aws=storyprotocol",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            env, status = runtime._github_cli_environment(task)

        self.assertEqual(env["GH_TOKEN"], "github-app-token")
        self.assertEqual(env["GITHUB_TOKEN"], "github-app-token")
        self.assertEqual(env["_HERMES_FORCE_GH_TOKEN"], "github-app-token")
        self.assertEqual(env["_HERMES_FORCE_GITHUB_TOKEN"], "github-app-token")
        self.assertEqual(env["GH_PROMPT_DISABLED"], "1")
        self.assertEqual(
            json.loads(env["_HERMES_GITHUB_TOKEN_MAP_JSON"]),
            {"piplabs": "github-app-token", "storyprotocol": "storyprotocol-token"},
        )
        self.assertEqual(env["_HERMES_GITHUB_DEFAULT_OWNER"], "piplabs")
        self.assertEqual(status["provider_ref"], "github_app_installation")
        self.assertEqual(status["installation_id"], "123754864")
        self.assertEqual(status["owner"], "piplabs")
        self.assertEqual(status["owners"], ["piplabs", "storyprotocol"])
        token_mock.assert_called_once()
        owner_token_mock.assert_called_once_with("storyprotocol", repo="depin-backend")

    def test_native_terminal_forces_github_app_token_through_hermes_env_scrubber(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Read the private repo.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-gh-app-force",
                }
            }
        )
        captured_env: dict[str, str] = {}

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                captured_env.update(env)
                captured_env["_test_git_askpass_exists"] = str(Path(env["GIT_ASKPASS"]).exists())
                request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
                payload = {
                    "ok": True,
                    "mcp_cleanup_errors": [],
                    "mcp_cleanup_status": "cleaned",
                    "response": "{}",
                    "result": {"final_response": "{}"},
                    "session_id": "rsi-prod-conversation-gh-force",
                }
                Path(request["result_path"]).write_text(json.dumps(payload, sort_keys=True), encoding="utf-8")
                self.returncode = 0
                self.stdout = io.StringIO("RSI_EXECUTOR_RESULT::" + json.dumps(payload, sort_keys=True) + "\n")
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
        ), mock.patch.object(
            HermesRuntime,
            "_github_app_installation_token",
            return_value=(
                "github-app-token",
                {
                    "status": "configured",
                    "reason": "github_app_installation",
                    "expires_at": "2026-05-01T08:00:00Z",
                    "installation_id": "123754864",
                    "owner": "piplabs",
                },
            ),
        ), mock.patch("rsi_runner.hermes_runtime.subprocess.Popen", side_effect=FakePopen), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
                "RSI_GITHUB_OWNER": "piplabs",
                "RSI_GITHUB_APP_ID": "3370196",
                "RSI_GITHUB_APP_INSTALLATION_ID": "123754864",
                "RSI_GITHUB_APP_PRIVATE_KEY": "private-key",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(captured_env["GH_TOKEN"], "github-app-token")
        self.assertEqual(captured_env["GITHUB_TOKEN"], "github-app-token")
        self.assertEqual(captured_env["_HERMES_FORCE_GH_TOKEN"], "github-app-token")
        self.assertEqual(captured_env["_HERMES_FORCE_GITHUB_TOKEN"], "github-app-token")
        self.assertEqual(json.loads(captured_env["_HERMES_GITHUB_TOKEN_MAP_JSON"]), {"piplabs": "github-app-token"})
        self.assertEqual(captured_env["GIT_TERMINAL_PROMPT"], "0")
        self.assertEqual(captured_env["_test_git_askpass_exists"], "True")
        self.assertGreaterEqual(int(captured_env["GIT_CONFIG_COUNT"]), 2)
        git_config = {
            captured_env[f"GIT_CONFIG_KEY_{index}"]: captured_env[f"GIT_CONFIG_VALUE_{index}"]
            for index in range(int(captured_env["GIT_CONFIG_COUNT"]))
        }
        self.assertEqual(git_config["credential.useHttpPath"], "true")
        self.assertIn("git-credential-rsi-github.py", git_config["credential.helper"])

    def test_git_askpass_uses_github_cli_token_for_private_clone(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            askpass = runtime._write_git_askpass(Path(tempdir))

            username = subprocess.check_output([str(askpass), "Username for https://github.com"], text=True).strip()
            password = subprocess.check_output(
                [str(askpass), "Password for https://github.com"],
                env={**os.environ, "GH_TOKEN": "github-token"},
                text=True,
            ).strip()

        self.assertEqual(username, "x-access-token")
        self.assertEqual(password, "github-token")

    def test_git_credential_helper_uses_owner_specific_installation_token(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            helper = runtime._write_git_credential_helper(Path(tempdir))

            env = {
                **os.environ,
                "_HERMES_GITHUB_DEFAULT_OWNER": "piplabs",
                "_HERMES_GITHUB_TOKEN_MAP_JSON": json.dumps(
                    {"piplabs": "piplabs-token", "storyprotocol": "storyprotocol-token"},
                    sort_keys=True,
                ),
            }
            storyprotocol_output = subprocess.check_output(
                [str(helper), "get"],
                input="protocol=https\nhost=github.com\npath=storyprotocol/story-deployments.git\n\n",
                env=env,
                text=True,
            )
            piplabs_output = subprocess.check_output(
                [str(helper), "get"],
                input="protocol=https\nhost=github.com\npath=piplabs/cloudflare.git\n\n",
                env=env,
                text=True,
            )

        self.assertIn("username=x-access-token", storyprotocol_output)
        self.assertIn("password=storyprotocol-token", storyprotocol_output)
        self.assertIn("username=x-access-token", piplabs_output)
        self.assertIn("password=piplabs-token", piplabs_output)

    def test_github_app_repo_scope_accepts_owner_qualified_repositories(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_GITHUB_OWNER": "piplabs",
                "RSI_GITHUB_APP_INSTALLATION_ID": "123754864",
                "RSI_GITHUB_APP_INSTALLATION_IDS": "storyprotocol=123754958",
                "RSI_GITHUB_REPO_OWNERS": "story-deployments=storyprotocol,story-infra-aws=storyprotocol",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "repo_allowlist": ["story-deployments"],
                        "prompt": "Read deployment config.",
                    }
                }
            )

            self.assertEqual(runtime._github_owner_for_repo("piplabs/numo-monorepo"), "piplabs")
            self.assertEqual(runtime._github_owner_for_repo("story-deployments"), "storyprotocol")
            self.assertEqual(
                runtime._github_repositories_for_owner(
                    "storyprotocol",
                    ["piplabs/numo-monorepo", "storyprotocol/story-deployments", "story-infra-aws"],
                ),
                ["story-deployments", "story-infra-aws"],
            )
            guidance = runtime._github_repository_guidance(task)
            self.assertIn("story-deployments -> storyprotocol/story-deployments", guidance)
            self.assertIn("configured GitHub App owner(s): piplabs, storyprotocol", guidance)

    def test_company_computer_bootstrap_writes_service_account_kubeconfig_for_terminal(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, tempfile.TemporaryDirectory() as computer_root:
            service_account = Path(tempdir, "serviceaccount")
            service_account.mkdir(parents=True)
            service_account_data = Path(tempdir, "serviceaccount-data")
            service_account_data.mkdir(parents=True)
            token_path = service_account / "token"
            ca_path = service_account / "ca.crt"
            namespace_path = service_account / "namespace"
            token_target = service_account_data / "token"
            ca_target = service_account_data / "ca.crt"
            token_target.write_text("token-value", encoding="utf-8")
            ca_target.write_text("ca-value", encoding="utf-8")
            token_path.symlink_to(token_target)
            ca_path.symlink_to(ca_target)
            namespace_path.write_text("rsi-platform", encoding="utf-8")
            kubeconfig_path = Path(computer_root, ".rsi", "kube", "config")
            captured_env: dict[str, str] = {}
            with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
                os.environ,
                {
                    **runner_env("prod"),
                    "HERMES_HOME": str(Path(tempdir, "hermes-home")),
                    "RSI_HERMES_EXECUTOR_ENABLED": "true",
                    "RSI_HERMES_COMPUTER_ROOT": computer_root,
                    "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                    "RSI_HERMES_KUBERNETES_CONTEXT_ENABLED": "true",
                    "RSI_HERMES_PROD_KUBERNETES_CONTEXT_ENABLED": "true",
                    "RSI_HERMES_PROD_KUBERNETES_CONTEXT_NAME": "use1-prod",
                    "RSI_HERMES_PROD_KUBERNETES_CLUSTER_NAME": "use1-prod",
                    "RSI_HERMES_PROD_KUBERNETES_CLUSTER_SERVER": "https://prod.eks.example",
                    "RSI_HERMES_PROD_KUBERNETES_CLUSTER_CA_DATA": "prod-ca-data",
                    "RSI_HERMES_PROD_KUBERNETES_ROLE_ARN": "arn:aws:iam::625966732747:role/use1-prod-rsi-stage-hermes-k8s-readonly",
                    "RSI_HERMES_PROD_KUBERNETES_REGION": "us-east-1",
                    "RSI_HERMES_PROD_KUBERNETES_NAMESPACE": "story",
                    "KUBECONFIG": str(kubeconfig_path),
                    "KUBERNETES_SERVICE_HOST": "10.0.0.1",
                    "KUBERNETES_SERVICE_PORT": "443",
                    "RSI_HERMES_KUBERNETES_SERVICE_ACCOUNT_TOKEN_PATH": str(token_path),
                    "RSI_HERMES_KUBERNETES_SERVICE_ACCOUNT_CA_PATH": str(ca_path),
                    "RSI_HERMES_KUBERNETES_SERVICE_ACCOUNT_NAMESPACE_PATH": str(namespace_path),
                },
                clear=True,
            ):
                runtime = HermesRuntime(RunnerConfig.from_env())
                captured_env = {
                    "TERMINAL_CWD": os.environ["TERMINAL_CWD"],
                    "PATH": os.environ["PATH"],
                    "KUBECONFIG": os.environ["KUBECONFIG"],
                }

            kubeconfig = kubeconfig_path.read_text(encoding="utf-8")
            manifest = json.loads(Path(computer_root, ".rsi", "computer.json").read_text(encoding="utf-8"))

        self.assertTrue(runtime.metadata["company_computer_bootstrap_status"]["ok"])
        self.assertEqual(captured_env["TERMINAL_CWD"], str(Path(computer_root).resolve()))
        self.assertTrue(captured_env["PATH"].split(os.pathsep)[0].endswith("/.rsi/bin"))
        self.assertEqual(captured_env["KUBECONFIG"], str(kubeconfig_path.resolve()))
        self.assertIn("server: https://10.0.0.1:443", kubeconfig)
        self.assertIn(f"tokenFile: {token_path}", kubeconfig)
        self.assertIn(f"certificate-authority: {ca_path}", kubeconfig)
        self.assertNotIn(str(token_target), kubeconfig)
        self.assertNotIn(str(ca_target), kubeconfig)
        self.assertIn("namespace: rsi-platform", kubeconfig)
        self.assertIn("- name: use1-prod", kubeconfig)
        self.assertIn("server: https://prod.eks.example", kubeconfig)
        self.assertIn("certificate-authority-data: prod-ca-data", kubeconfig)
        self.assertIn("command: aws", kubeconfig)
        self.assertIn("- get-token", kubeconfig)
        self.assertIn("- --cluster-name", kubeconfig)
        self.assertIn("- --role-arn", kubeconfig)
        self.assertIn("- arn:aws:iam::625966732747:role/use1-prod-rsi-stage-hermes-k8s-readonly", kubeconfig)
        self.assertIn("namespace: story", kubeconfig)
        self.assertIn("current-context: hermes-company-computer", kubeconfig)
        self.assertEqual(manifest["terminal"]["bin_dir"], str(Path(computer_root, ".rsi", "bin")))
        self.assertEqual(manifest["native_toolsets"], ["terminal", "file"])
        self.assertEqual(manifest["prod_kubernetes_context"]["name"], "use1-prod")
        self.assertEqual(manifest["prod_kubernetes_context"]["auth"], "aws_eks_exec_assume_role")

    def test_company_computer_bootstrap_failure_fails_closed(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Inspect Kubernetes.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-kube-fail",
                }
            }
        )
        with tempfile.TemporaryDirectory() as tempdir, tempfile.TemporaryDirectory() as computer_root, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": str(Path(tempdir, "hermes-home")),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_COMPUTER_ROOT": computer_root,
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "RSI_HERMES_KUBERNETES_CONTEXT_ENABLED": "true",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime._execute_native_workflow_task_request(task)

        self.assertFalse(runtime.available)
        self.assertFalse(runtime.metadata["company_computer_bootstrap_status"]["ok"])
        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "hermes_company_computer_bootstrap_failed")

    def test_company_computer_bootstrap_rejects_incomplete_prod_kubernetes_context(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Inspect production Kubernetes.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-prod-kube-fail",
                }
            }
        )
        with tempfile.TemporaryDirectory() as tempdir, tempfile.TemporaryDirectory() as computer_root, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": str(Path(tempdir, "hermes-home")),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_COMPUTER_ROOT": computer_root,
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "RSI_HERMES_PROD_KUBERNETES_CONTEXT_ENABLED": "true",
                "RSI_HERMES_PROD_KUBERNETES_CLUSTER_SERVER": "https://prod.eks.example",
                "RSI_HERMES_PROD_KUBERNETES_CLUSTER_CA_DATA": "prod-ca-data",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime._execute_native_workflow_task_request(task)

        self.assertFalse(runtime.available)
        self.assertFalse(runtime.metadata["company_computer_bootstrap_status"]["ok"])
        self.assertIn(
            "RSI_HERMES_PROD_KUBERNETES_ROLE_ARN",
            "\n".join(runtime.metadata["company_computer_bootstrap_status"]["errors"]),
        )
        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "hermes_company_computer_bootstrap_failed")

    def test_hermes_contract_rejects_missing_pin(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True):
            Path(tempdir, "config.yaml").write_text("plugins:\n  enabled:\n    - rsi_context_engine\n", encoding="utf-8")
            status = validate_hermes_contract(
                expected_pin="",
                hermes_home=tempdir,
                session_db=object(),
                required_toolsets=[],
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.pin_status, "missing_expected_pin")
        self.assertIn("RSI_HERMES_PIN is required.", status.errors)

    def test_hermes_contract_rejects_pin_mismatch(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True), mock.patch(
            "rsi_runner.hermes_agent_adapter._read_direct_url_commit",
            return_value=("0.0.0", "cafebabe"),
        ):
            Path(tempdir, "config.yaml").write_text("plugins:\n  enabled:\n    - rsi_context_engine\n", encoding="utf-8")
            status = validate_hermes_contract(
                expected_pin="deadbeef",
                hermes_home=tempdir,
                session_db=object(),
                required_toolsets=[],
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.pin_status, "mismatch")

    def test_hermes_contract_rejects_missing_plugin_enablement(self) -> None:
        plugin_manager = types.SimpleNamespace(list_plugins=lambda: [{"enabled": True, "name": "rsi_context_engine"}])
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True), mock.patch(
            "rsi_runner.hermes_agent_adapter.discover_plugins",
            side_effect=lambda force=False: None,
        ), mock.patch(
            "rsi_runner.hermes_agent_adapter.get_plugin_manager",
            return_value=plugin_manager,
        ):
            status = validate_hermes_contract(
                expected_pin=HERMES_TEST_PIN,
                hermes_home=tempdir,
                session_db=object(),
                required_toolsets=[],
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.plugin_status, "config_missing")
        self.assertIn("rsi_context_engine is not enabled in Hermes config.", status.errors)

    def test_hermes_contract_rejects_missing_session_db_and_toolset(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True), mock.patch(
            "rsi_runner.hermes_agent_adapter.validate_toolset",
            side_effect=lambda _name: False,
        ):
            Path(tempdir, "config.yaml").write_text("plugins:\n  enabled:\n    - rsi_context_engine\n", encoding="utf-8")
            status = validate_hermes_contract(
                expected_pin=HERMES_TEST_PIN,
                hermes_home=tempdir,
                session_db=None,
                required_toolsets=["rsi-missing-toolset"],
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.session_db_status, "missing")
        self.assertEqual(status.toolset_status["rsi-missing-toolset"], "failed")

    def test_hermes_contract_rejects_corrupt_session_db(self) -> None:
        plugin_manager = types.SimpleNamespace(list_plugins=lambda: [{"enabled": True, "name": "rsi_context_engine"}])
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True), mock.patch(
            "rsi_runner.hermes_agent_adapter._read_direct_url_commit",
            return_value=("0.12.0", HERMES_TEST_PIN),
        ), mock.patch(
            "rsi_runner.hermes_agent_adapter.discover_plugins",
            side_effect=lambda force=False: None,
        ), mock.patch(
            "rsi_runner.hermes_agent_adapter.get_plugin_manager",
            return_value=plugin_manager,
        ):
            Path(tempdir, "config.yaml").write_text("plugins:\n  enabled:\n    - rsi_context_engine\n", encoding="utf-8")
            db_path = Path(tempdir, "state.db")
            sqlite3.connect(db_path).close()
            db_path.write_bytes(b"not a sqlite database")
            status = validate_hermes_contract(
                expected_pin=HERMES_TEST_PIN,
                hermes_home=tempdir,
                session_db=types.SimpleNamespace(db_path=db_path),
                required_toolsets=[],
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.session_db_status, "corrupt")
        self.assertTrue(any("Hermes SessionDB integrity check failed" in error for error in status.errors))

    def test_hermes_contract_rejects_session_db_missing_schema(self) -> None:
        plugin_manager = types.SimpleNamespace(list_plugins=lambda: [{"enabled": True, "name": "rsi_context_engine"}])
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True), mock.patch(
            "rsi_runner.hermes_agent_adapter._read_direct_url_commit",
            return_value=("0.12.0", HERMES_TEST_PIN),
        ), mock.patch(
            "rsi_runner.hermes_agent_adapter.discover_plugins",
            side_effect=lambda force=False: None,
        ), mock.patch(
            "rsi_runner.hermes_agent_adapter.get_plugin_manager",
            return_value=plugin_manager,
        ):
            Path(tempdir, "config.yaml").write_text("plugins:\n  enabled:\n    - rsi_context_engine\n", encoding="utf-8")
            db_path = Path(tempdir, "state.db")
            sqlite3.connect(db_path).close()
            status = validate_hermes_contract(
                expected_pin=HERMES_TEST_PIN,
                hermes_home=tempdir,
                session_db=types.SimpleNamespace(db_path=db_path),
                required_toolsets=[],
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.session_db_status, "missing_schema")
        self.assertTrue(any("Hermes SessionDB schema is missing required table(s)" in error for error in status.errors))

    def test_hermes_contract_rejects_missing_toolset_validation_api_once(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True), mock.patch(
            "rsi_runner.hermes_agent_adapter.validate_toolset",
            None,
        ):
            Path(tempdir, "config.yaml").write_text("plugins:\n  enabled:\n    - rsi_context_engine\n", encoding="utf-8")
            status = validate_hermes_contract(
                expected_pin=HERMES_TEST_PIN,
                hermes_home=tempdir,
                session_db=object(),
                required_toolsets=["rsi-missing-toolset-a", "rsi-missing-toolset-b"],
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.toolset_status["rsi-missing-toolset-a"], "validate_api_missing")
        self.assertEqual(status.toolset_status["rsi-missing-toolset-b"], "validate_api_missing")
        self.assertEqual(status.errors.count("Hermes toolset validation API is unavailable."), 1)

    def test_hermes_agent_adapter_marks_incomplete_native_result_partial(self) -> None:
        adapter = HermesAgentAdapter({"session_id": "session-1", "max_iterations": 20})

        meta = adapter._completion_meta({"completed": False, "api_calls": 20, "final_response": "{}"})

        self.assertEqual(meta["termination_reason"], "iteration_budget_exhausted")
        self.assertEqual(meta["completion_verdict"], "partial")
        self.assertTrue(meta["max_iterations_reached"])
        self.assertFalse(meta["native_result_completed"])

    def test_hermes_contract_rejects_missing_aiagent_kwargs(self) -> None:
        class BadAIAgent:
            def __init__(self, model=None):
                pass

            def run_conversation(self, user_message=None):
                return {}

        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True), mock.patch(
            "rsi_runner.hermes_agent_adapter.AIAgent", BadAIAgent
        ):
            Path(tempdir, "config.yaml").write_text("plugins:\n  enabled:\n    - rsi_context_engine\n", encoding="utf-8")
            status = validate_hermes_contract(
                expected_pin=HERMES_TEST_PIN,
                hermes_home=tempdir,
                session_db=object(),
                required_toolsets=[],
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.api_signature_status, "failed")
        self.assertTrue(any("AIAgent.__init__ missing" in error for error in status.errors))

    def test_hermes_agent_adapter_rejects_custom_runtime_provider(self) -> None:
        class CaptureAIAgent:
            last_kwargs: dict[str, object] = {}

            def __init__(self, **kwargs) -> None:
                type(self).last_kwargs = kwargs

        with mock.patch("rsi_runner.hermes_agent_adapter.AIAgent", CaptureAIAgent), mock.patch.dict(
            os.environ,
            {
                "OPENROUTER_API_KEY": "openrouter-key",
            },
            clear=True,
        ):
            adapter = HermesAgentAdapter(
                {
                    "model": "custom-model",
                    "runtime": {
                        "provider": "custom",
                        "api_mode": "",
                        "base_url": "https://models.internal.example",
                    },
                    "toolsets": [],
                }
            )
            with self.assertRaisesRegex(RuntimeError, "OpenRouter"):
                adapter._create_agent("session-1", object(), "")

    def test_hermes_agent_adapter_uses_openrouter_key_and_provider_routing(self) -> None:
        class CaptureAIAgent:
            last_kwargs: dict[str, object] = {}

            def __init__(self, **kwargs) -> None:
                type(self).last_kwargs = kwargs

        with mock.patch("rsi_runner.hermes_agent_adapter.AIAgent", CaptureAIAgent), mock.patch.dict(
            os.environ,
            {
                "OPENROUTER_API_KEY": "openrouter-key",
            },
            clear=True,
        ):
            adapter = HermesAgentAdapter(
                {
                    "model": "deepseek/deepseek-v4-pro",
                    "runtime": {
                        "provider": "openrouter",
                        "api_mode": "",
                        "base_url": "",
                        "provider_routing": {
                            "only": ["deepseek"],
                            "order": ["deepseek"],
                            "require_parameters": True,
                        },
                    },
                    "toolsets": [],
                }
            )
            adapter._create_agent("session-1", object(), "")

        self.assertEqual(CaptureAIAgent.last_kwargs["api_key"], "openrouter-key")
        self.assertEqual(CaptureAIAgent.last_kwargs["provider"], "openrouter")
        self.assertEqual(CaptureAIAgent.last_kwargs["providers_allowed"], ["deepseek"])
        self.assertEqual(CaptureAIAgent.last_kwargs["providers_order"], ["deepseek"])
        self.assertTrue(CaptureAIAgent.last_kwargs["provider_require_parameters"])

    def test_hermes_agent_adapter_normalizes_stream_callbacks_to_lifecycle_events(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True):
            adapter = HermesAgentAdapter({"session_id": "session-callbacks"})
            adapter._reasoning_callback(" private surfaced reasoning")
            adapter._stream_delta_callback(" visible output")
            adapter._thinking_callback("thinking")
            adapter._tool_generation_callback("terminal")
            adapter._tool_progress_callback("tool.completed", "terminal", None, None, duration=1.2, is_error=False)
            adapter._tool_start_callback("call-1", "terminal", {"cmd": "pwd"})
            adapter._tool_complete_callback("call-1", "terminal", {"cmd": "pwd"}, "ok")
            adapter._status_callback("lifecycle", "context loaded")

            events = [
                json.loads(line)
                for line in adapter.lifecycle_path.read_text(encoding="utf-8").splitlines()
                if line.strip()
            ]

        self.assertEqual(
            [item["event_type"] for item in events],
            [
                "model.reasoning.delta",
                "model.output.delta",
                "model.thinking",
                "tool.generation.started",
                "tool.call.progress",
                "tool.call.started",
                "tool.call.completed",
                "model.status",
            ],
        )
        self.assertEqual(events[0]["payload"]["delta"], " private surfaced reasoning")
        self.assertEqual(events[1]["payload"]["delta"], " visible output")
        self.assertEqual(events[5]["payload"]["tool_name"], "terminal")

    def test_native_lifecycle_tailer_emits_redacted_live_observations(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            path = Path(tempdir) / "session.jsonl"
            observer = RecordingObserver()
            emitted: list[dict[str, object]] = []

            def emit_event(target, phase, item, *, secret_values):
                payload = item.get("payload", {})
                if isinstance(payload, dict):
                    payload = {
                        key: str(value).replace("secret-token", "[redacted]")
                        for key, value in payload.items()
                    }
                target.emit(phase=phase, event_type=str(item.get("event_type")), status=str(item.get("status")), payload=payload)
                emitted.append(item)

            tailer = _NativeLifecycleTailer(
                path=path,
                phase="investigate",
                observer=observer,
                secret_values=["secret-token"],
                emit_event=emit_event,
            )
            tailer.start()
            path.write_text(
                json.dumps(
                    {
                        "event_type": "model.output.delta",
                        "status": "streaming",
                        "payload": {"delta": "hello secret-token"},
                    }
                )
                + "\n",
                encoding="utf-8",
            )
            deadline = time.time() + 2
            while not observer.events and time.time() < deadline:
                time.sleep(0.05)
            tailer.stop()

        self.assertEqual(tailer.emitted, 1)
        self.assertEqual(observer.events[0]["event_type"], "model.output.delta")
        self.assertEqual(observer.events[0]["payload"]["delta"], "hello [redacted]")

    def test_native_lifecycle_emission_preserves_stream_delta_whitespace(self) -> None:
        with mock.patch.dict(os.environ, runner_env(), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
        observer = RecordingObserver()

        runtime._emit_native_lifecycle_event(
            observer,
            "investigate",
            {
                "event_type": "model.reasoning.delta",
                "status": "streaming",
                "payload": {"delta": " hello secret-token"},
            },
            secret_values=["secret-token"],
        )

        self.assertEqual(observer.events[0]["event_type"], "model.reasoning.delta")
        self.assertEqual(observer.events[0]["payload"]["delta"], " hello [redacted]")

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
            {**runner_env("prod"), "HERMES_HOME": tempdir, "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true"},
            clear=True,
        ):
            sync_mock.return_value = {"copied": ["architecture-diagram"]}
            config = RunnerConfig.from_env()
            manager = SessionManager(config)
            config_text = Path(tempdir, "config.yaml").read_text(encoding="utf-8")

        self.assertIn("model:", config_text)
        self.assertIn('default: "deepseek/deepseek-v4-pro"', config_text)
        self.assertIn("provider: openrouter", config_text)
        self.assertNotIn("base_url:", config_text)
        self.assertIn('api_key: ""', config_text)
        self.assertIn("provider_routing:", config_text)
        self.assertIn('  only:\n    - "deepseek"', config_text)
        self.assertIn('  order:\n    - "deepseek"', config_text)
        self.assertIn("  require_parameters: true", config_text)
        self.assertIn("terminal:", config_text)
        self.assertIn('backend: "local"', config_text)
        self.assertIn('cwd: "/tmp/hermes/workspace/company"', config_text)
        self.assertIn("timeout: 180", config_text)
        self.assertIn("lifetime_seconds: 900", config_text)
        self.assertIn("plugins:", config_text)
        self.assertIn("  enabled:", config_text)
        self.assertIn("    - rsi_context_engine", config_text)
        self.assertEqual(manager.hermes_config_parity_status, "configured")
        self.assertEqual(manager.hermes_config_parity_error, "")

    def test_session_manager_writes_native_openrouter_model_config(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch("rsi_runner.session_manager.sync_skills") as sync_mock, mock.patch.dict(
            os.environ,
            {**openrouter_runner_env("prod"), "HERMES_HOME": tempdir},
            clear=True,
        ):
            sync_mock.return_value = {"copied": ["architecture-diagram"]}
            config = RunnerConfig.from_env()
            SessionManager(config)
            config_text = Path(tempdir, "config.yaml").read_text(encoding="utf-8")

        self.assertIn("model:", config_text)
        self.assertIn('default: "deepseek/deepseek-v4-pro"', config_text)
        self.assertIn("provider: openrouter", config_text)
        self.assertNotIn("base_url:", config_text)
        self.assertIn("provider_routing:", config_text)
        self.assertIn('  only:\n    - "deepseek"', config_text)
        self.assertIn('  order:\n    - "deepseek"', config_text)
        self.assertIn("  require_parameters: true", config_text)
        self.assertIn("plugins:", config_text)

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
                "RSI_RUNNER_MODEL": "openrouter/deepseek/deepseek-v4-pro:beta # unsafe",
            },
            clear=True,
        ):
            sync_mock.return_value = {"copied": ["architecture-diagram"]}
            config = RunnerConfig.from_env()
            SessionManager(config)
            config_text = Path(tempdir, "config.yaml").read_text(encoding="utf-8")

        self.assertIn('default: "deepseek/deepseek-v4-pro:beta # unsafe"', config_text)
        self.assertNotIn("base_url:", config_text)

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

    def test_openrouter_models_use_persisted_hermes_sessions_with_xhigh(self) -> None:
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
        self.assertEqual(FakeAIAgent.last_kwargs["model"], "deepseek/deepseek-v4-pro")
        self.assertNotIn("api_mode", FakeAIAgent.last_kwargs)
        self.assertNotIn("base_url", FakeAIAgent.last_kwargs)
        self.assertEqual(FakeAIAgent.last_kwargs["provider"], "openrouter")
        self.assertEqual(FakeAIAgent.last_kwargs["api_key"], "openrouter-test-key")
        self.assertEqual(FakeAIAgent.last_kwargs["providers_allowed"], ["deepseek"])
        self.assertEqual(FakeAIAgent.last_kwargs["providers_order"], ["deepseek"])
        self.assertTrue(FakeAIAgent.last_kwargs["provider_require_parameters"])
        self.assertEqual(FakeAIAgent.last_kwargs["reasoning_config"], {"enabled": True, "effort": "xhigh"})
        self.assertEqual(FakeAIAgent.last_kwargs["enabled_toolsets"], [])
        self.assertEqual(FakeAIAgent.last_kwargs["session_id"], "rsi-prod-conversation-123")
        self.assertNotIn("persist_session", FakeAIAgent.last_kwargs)
        self.assertFalse(FakeAIAgent.last_kwargs["skip_memory"])
        self.assertEqual(FakeAIAgent.last_history, [{"role": "user", "content": "Earlier thread message"}])
        self.assertEqual(result.raw["model"], "openrouter/deepseek/deepseek-v4-pro")
        self.assertEqual(result.raw["provider"], "openrouter")
        self.assertEqual(result.raw["provider_model"], "deepseek/deepseek-v4-pro")
        self.assertEqual(result.raw["api_mode"], "")
        self.assertEqual(result.raw["provider_routing"]["only"], ["deepseek"])
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
        self.assertEqual(result.raw["hermes_pin"], HERMES_TEST_PIN)
        self.assertIn("structured_output", result.raw)
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")

    def test_openrouter_models_use_native_provider_and_routing(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "eval",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the eval.",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "session_scope_kind": "eval_line",
                    "session_scope_id": "shared-store:openrouter",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:eval",
                    "user_peer_id": "operator:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, openrouter_runner_env("eval"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(FakeAIAgent.last_kwargs["model"], "deepseek/deepseek-v4-pro")
        self.assertEqual(FakeAIAgent.last_kwargs["provider"], "openrouter")
        self.assertNotIn("api_mode", FakeAIAgent.last_kwargs)
        self.assertNotIn("base_url", FakeAIAgent.last_kwargs)
        self.assertEqual(FakeAIAgent.last_kwargs["api_key"], "openrouter-test-key")
        self.assertEqual(FakeAIAgent.last_kwargs["providers_allowed"], ["deepseek"])
        self.assertEqual(FakeAIAgent.last_kwargs["providers_order"], ["deepseek"])
        self.assertTrue(FakeAIAgent.last_kwargs["provider_require_parameters"])
        self.assertEqual(result.raw["model"], "openrouter/deepseek/deepseek-v4-pro")
        self.assertEqual(result.raw["provider"], "openrouter")
        self.assertEqual(result.raw["provider_model"], "deepseek/deepseek-v4-pro")
        self.assertEqual(result.raw["provider_routing"]["only"], ["deepseek"])

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
        status = runtime.executor_status("hexec-test-routing")
        self.assertEqual(status["status"], "completed")
        self.assertEqual(status["executor_instance_id"], "prod")
        self.assertTrue(status["result"]["ok"])


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

    def test_executor_status_persists_completed_result_for_recovery(self) -> None:
        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as workspace_root, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_COMPUTER_ROOT": str(Path(workspace_root) / "company"),
                "RSI_HERMES_RUN_ROOT": str(Path(workspace_root) / "company" / ".rsi" / "runs"),
                "RSI_HERMES_ARTIFACT_ROOT": str(Path(workspace_root) / "company" / "artifacts"),
            },
            clear=True,
        ):
            first = HermesRuntime(RunnerConfig.from_env())
            first._store_executor_result(
                "hexec-persist",
                {
                    "execution_id": "hexec-persist",
                    "status": "completed",
                    "result": {
                        "ok": True,
                        "message": "Recovered",
                        "provider": "fake",
                        "raw": {"structured_output": {"final_answer": "Recovered"}},
                    },
                },
            )
            second = HermesRuntime(RunnerConfig.from_env())
            status = second.executor_status("hexec-persist")

        self.assertEqual(status["status"], "completed")
        self.assertTrue(status["result"]["ok"])
        self.assertEqual(status["result"]["raw"]["structured_output"]["final_answer"], "Recovered")

    def test_executor_status_marks_persisted_active_without_process_as_orphaned(self) -> None:
        for persisted_status in ("running", "starting", "cancel_requested"):
            with self.subTest(persisted_status=persisted_status):
                with (
                    tempfile.TemporaryDirectory() as hermes_home,
                    tempfile.TemporaryDirectory() as workspace_root,
                    mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager),
                    mock.patch.dict(
                        os.environ,
                        {
                            **runner_env("prod"),
                            "HERMES_HOME": hermes_home,
                            "RSI_HERMES_COMPUTER_ROOT": str(Path(workspace_root) / "company"),
                            "RSI_HERMES_RUN_ROOT": str(Path(workspace_root) / "company" / ".rsi" / "runs"),
                            "RSI_HERMES_ARTIFACT_ROOT": str(Path(workspace_root) / "company" / "artifacts"),
                        },
                        clear=True,
                    ),
                ):
                    execution_id = f"hexec-orphan-{persisted_status}"
                    first = HermesRuntime(RunnerConfig.from_env())
                    first._store_executor_result(
                        execution_id,
                        {"execution_id": execution_id, "status": persisted_status},
                    )
                    second = HermesRuntime(RunnerConfig.from_env())
                    status = second.executor_status(execution_id)

                self.assertEqual(status["status"], "orphaned")
                self.assertIn("no local execution process", status["message"])









    def test_legacy_tool_allowlist_does_not_construct_provider_tools(self) -> None:
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

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertNotIn("runner_invalid_request", json.dumps(result.raw))
        self.assertIsNotNone(FakeAIAgent.last_kwargs)

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
                    "prompt": "User request: Please use /architecture-diagram skill and summarize the workflow.\n\nInvestigate using Hermes-native tools and configured MCP servers.",
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

    def test_workflow_accepts_json_code_fenced_structured_output(self) -> None:
        class FencedStructuredOutputAIAgent(FakeAIAgent):
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
                    "final_response": "```json\n"
                    + json.dumps(
                        {
                            "visible_reasoning": "Validated the workflow and sent the reply.",
                            "final_answer": "Created the requested follow-up.",
                            "confidence": 0.91,
                            "context_summary": "The response is complete.",
                            "proposed_actions": [],
                            "knowledge_drafts": [],
                            "outcome_hypotheses": [],
                            "produced_artifacts": [],
                            "artifact_failure_reason": "",
                        }
                    )
                    + "\n```"
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
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FencedStructuredOutputAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertFalse(result.raw["repair_attempted"])
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Created the requested follow-up.")

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
            self.assertNotIn("rsi-governed-readonly", process._request["toolsets"])
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

    def test_native_executor_fails_before_subprocess_when_github_app_credentials_missing(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Read the private repo.",
                    "trace_id": "trace-gh-required",
                    "workflow_id": "wf-gh-required",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-gh-required",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                }
            }
        )
        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", FakeAIAgent
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen",
            side_effect=AssertionError("Hermes subprocess must not start without GitHub App credentials"),
        ) as popen_mock, mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
                "GH_TOKEN": "ignored-pod-token",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.provider, "hermes-native-executor")
        self.assertEqual(result.raw["failure_class"], "github_app_credentials_unavailable")
        self.assertEqual(result.raw["termination_reason"], "github_app_credentials_unavailable")
        self.assertEqual(result.raw["github_credentials"]["reason"], "missing_github_app_credentials")
        popen_mock.assert_not_called()

    def test_native_executor_incomplete_result_preserves_partial_contract(self) -> None:
        structured = json.dumps(
            partial_structured_output(
                reply_text="Grounded but partial answer.",
                proposed_actions=[],
            )
        )
        captured_requests: list[dict[str, object]] = []

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
                captured_requests.append(request)
                self.returncode = 0
                payload = {
                    "ok": True,
                    "mcp_cleanup_errors": [],
                    "mcp_cleanup_status": "cleaned",
                    "response": structured,
                    "result": {
                        "final_response": structured,
                        "completed": False,
                        "api_calls": 20,
                        "partial": False,
                        "interrupted": False,
                    },
                    "session_id": "rsi-prod-conversation-123",
                }
                Path(request["result_path"]).write_text(json.dumps(payload, sort_keys=True), encoding="utf-8")
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

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Investigate the latest trace.",
                    "trace_id": "trace-partial-native",
                    "workflow_id": "wf-partial-native",
                    "operation_id": "op-partial-native",
                    "reply_delivery_mode": "none",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-partial-native",
                }
            }
        )

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.subprocess.Popen", FakePopen), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
                "RSI_HERMES_EXECUTOR_SERVICE_ONLY": "true",
                "RSI_RUNTIME_OBSERVATION_SINK_URL": "http://control-plane.internal:8080/internal/runtime/observations",
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": tempdir,
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["termination_reason"], "iteration_budget_exhausted")
        self.assertTrue(result.raw["max_iterations_reached"])
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Grounded but partial answer.")
        self.assertEqual(result.raw["runner_diagnostics"]["completion_verdict"], "partial")
        self.assertEqual(captured_requests[0]["phase_contract"]["history_policy"], "session")
        self.assertNotIn("rsi-governed-readonly", captured_requests[0]["phase_contract"]["required_toolsets"])

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
        mcp_registration = TaskScopedMCPRegistration(
            servers=[
                TaskScopedMCPServer(
                    source_label="render-context",
                    profile="custom_read_only",
                    server_name="rsi-task-render-context",
                    toolset_alias="mcp-render-context",
                    included_tool_names=["fetch_context"],
                    hermes_config={},
                )
            ]
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
            self.assertIn("rsi-artifacts", request["toolsets"])
            self.assertIn("mcp-render-context", request["toolsets"])
            self.assertEqual(len(request["toolsets"]), len(set(request["toolsets"])))
            self.assertEqual(request["conversation_history"], [])
            self.assertEqual(request["phase_contract"]["history_policy"], "empty")
            self.assertEqual(request["phase_contract"]["required_toolsets"], ["rsi-artifacts"])
            self.assertEqual(request["phase_contract"]["missing_required_toolsets"], [])

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
                        "mcp_servers": [{"server_label": "render-context", "profile": "custom_read_only"}],
                    }
                }
            )
            with mock.patch.object(runtime._mcp_adapter, "plan_task_servers", return_value=mcp_registration):
                result = runtime._execute_native_workflow_task_request(
                    task,
                    observer=RecordingObserver(),
                )

        self.assertTrue(result.ok)

    def test_plugin_artifact_write_creates_workspace_metadata_and_lifecycle_events(self) -> None:
        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as artifact_dir, mock.patch.dict(
            os.environ, {"HERMES_HOME": hermes_home}, clear=True
        ):
            context_dir = Path(hermes_home, "rsi_runtime", "context")
            context_dir.mkdir(parents=True, exist_ok=True)
            session_id = "sess-artifact"
            context_dir.joinpath(f"{session_id}.json").write_text(
                json.dumps(
                    {
                        "artifact_output_dir": artifact_dir,
                        "execution_id": "exec-artifact",
                    }
                ),
                encoding="utf-8",
            )
            namespace: dict[str, object] = {}
            exec(_build_plugin_module(), namespace)
            handler = namespace["_tool_handler"]("artifact_write_file")
            list_handler = namespace["_tool_handler"]("artifact_list_files")

            payload = json.loads(handler({"path": "diagram.html", "content": "<html></html>"}, task_id=session_id))
            escaped = json.loads(handler({"path": "../escape.html", "content": "nope"}, task_id=session_id))
            listed = json.loads(list_handler({"path": ""}, task_id=session_id))
            list_escaped = json.loads(list_handler({"path": "../escape.html"}, task_id=session_id))
            non_dict_args = json.loads(handler("not a dict", task_id=session_id))
            events_path = Path(hermes_home, "rsi_runtime", "lifecycle", f"{session_id}.jsonl")
            events = [json.loads(line) for line in events_path.read_text(encoding="utf-8").splitlines()]
            self.assertEqual(payload["status"], "ok")
            self.assertTrue(Path(artifact_dir, "diagram.html").exists())
            self.assertEqual(payload["output"]["workspace_path"], str(Path(artifact_dir, "diagram.html").resolve()))
            self.assertEqual(payload["output"]["created_by_execution_id"], "exec-artifact")
            self.assertEqual(payload["output"]["share_status"], "local")
            self.assertEqual(escaped["status"], "error")
            self.assertIn("artifact_path_outside_root", escaped["error"])
            self.assertEqual(listed["status"], "ok")
            self.assertEqual(list_escaped["status"], "error")
            self.assertIn("artifact_path_outside_root", list_escaped["error"])
            self.assertEqual(non_dict_args["status"], "error")
            self.assertNotIn("object has no attribute", non_dict_args["error"])
            self.assertIn("artifact.write.started", [event["event"] for event in events])
            self.assertIn("artifact.write.completed", [event["event"] for event in events])
            self.assertIn("artifact.write.failed", [event["event"] for event in events])
            self.assertIn("artifact.list.completed", [event["event"] for event in events])
            self.assertIn("artifact.list.failed", [event["event"] for event in events])

    def test_plugin_slack_upload_tool_is_not_registered(self) -> None:
        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as artifact_dir, mock.patch.dict(
            os.environ, {"HERMES_HOME": hermes_home}, clear=True
        ):
            artifact_path = Path(artifact_dir, "diagram.html")
            artifact_path.write_text("<html><body><svg width='10' height='10'><rect width='10' height='10'/></svg></body></html>", encoding="utf-8")
            context_dir = Path(hermes_home, "rsi_runtime", "context")
            context_dir.mkdir(parents=True, exist_ok=True)
            session_id = "sess-upload"
            context_dir.joinpath(f"{session_id}.json").write_text(
                json.dumps(
                    {
                        "artifact_output_dir": artifact_dir,
                        "hermes_computer_root": artifact_dir,
                        "task_channel_id": "C123",
                        "task_thread_ts": "171000001.000100",
                    }
                ),
                encoding="utf-8",
            )
            namespace: dict[str, object] = {}
            exec(_build_plugin_module(), namespace)
            self.assertNotIn("slack_upload_file", namespace["_TRANSPORT_TO_CANONICAL"])
            with self.assertRaises(KeyError):
                namespace["_tool_handler"]("slack_upload_file")

    def test_native_worker_uses_aiagent_adapter_not_hermes_cli(self) -> None:
        source = Path(__file__).parents[1].joinpath("rsi_runner", "hermes_executor_worker.py").read_text(encoding="utf-8")

        self.assertIn("HermesAgentAdapter", source)
        self.assertNotIn("HermesCLI", source)
        self.assertNotIn("_init_agent", source)

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
                observer=None,
            )

        self.assertTrue(result.ok)
        self.assertEqual(runtime.executor_status("hexec-explicit")["status"], "completed")
        self.assertEqual(runtime.executor_status("hexec-explicit")["execution_id"], "hexec-explicit")
        self.assertEqual(runtime.executor_status("hexec-explicit")["executor_instance_id"], "prod")
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
                    "openrouter=sk-secret-openrouter-key\n"
                    "aws=aws-session-secret\n"
                    "env=openrouter-test-key\n"
                    "JOB_HELPER_TOKEN_SECRET=fake-secret-value-for-redaction\n"
                    '{\n  "name": "INTERNAL_ADMIN_READ_API_KEY",\n  "value": "admin-read-key-from-cluster"\n}\n'
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
                observer=observer,
            )

        self.assertTrue(result.ok)
        output_events = [event for event in observer.events if event["event_type"] == "terminal.output"]
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
        self.assertIn("[redacted-api-key]", stderr_text)
        self.assertIn("[redacted]", stderr_text)
        self.assertNotIn("secret-bearer-token", stderr_text)
        self.assertNotIn("xoxb-123456789-secret", stderr_text)
        self.assertNotIn("sk-secret-openrouter-key", stderr_text)
        self.assertNotIn("aws-session-secret", stderr_text)
        self.assertNotIn("openrouter-test-key", stderr_text)
        self.assertNotIn("fake-secret-value-for-redaction", stderr_text)
        self.assertNotIn("admin-read-key-from-cluster", stderr_text)
        self.assertNotIn("secret-bearer-token", result.raw["native_executor_stderr"])
        self.assertNotIn("xoxb-123456789-secret", result.raw["native_executor_stderr"])
        self.assertTrue(any(event["event_type"] == "executor.result.persisted" for event in observer.events))
        self.assertTrue(any(event["event_type"] == "executor.result.loaded" for event in observer.events))

    def test_native_executor_stream_activity_counts_even_when_output_is_suppressed(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        activities: list[str] = []
        chunks: list[str] = []
        observer = RecordingObserver()

        runtime._read_native_executor_stream(
            io.StringIO("Warning: Unknown toolsets: \nmcp-test-slack\n\n"),
            stream_name="stdout",
            phase="main",
            observer=observer,
            chunk_store=chunks,
            secret_values=[],
            result_detected=threading.Event(),
            activity_callback=activities.append,
        )

        self.assertEqual(activities, ["stdout output"])
        self.assertEqual(len(chunks), 1)
        self.assertFalse([event for event in observer.events if event["event_type"] == "terminal.output"])

    def test_native_lifecycle_tailer_reports_activity_for_lifecycle_events(self) -> None:
        activities: list[str] = []
        observer = RecordingObserver()

        def emit_event(
            target_observer: RecordingObserver,
            phase: str,
            item: dict[str, object],
            *,
            secret_values: list[str],
        ) -> None:
            target_observer.emit(
                phase=phase,
                event_type=str(item.get("event_type") or ""),
                status=str(item.get("status") or ""),
                payload={"secret_count": len(secret_values)},
            )

        with tempfile.TemporaryDirectory() as tempdir:
            path = Path(tempdir) / "lifecycle.jsonl"
            path.write_text(
                json.dumps({"event_type": "tool.generation.started", "status": "running"}) + "\n",
                encoding="utf-8",
            )
            tailer = _NativeLifecycleTailer(
                path=path,
                phase="main",
                observer=observer,
                secret_values=[],
                emit_event=emit_event,
                activity_callback=activities.append,
            )
            tailer.drain()

        self.assertEqual(activities, ["lifecycle:tool.generation.started:running"])
        self.assertEqual(observer.events[0]["event_type"], "tool.generation.started")

    def test_native_executor_subprocess_uses_inactivity_guard_not_wall_clock_task_kill(self) -> None:
        source = Path(__file__).parents[1].joinpath("rsi_runner", "hermes_runtime.py").read_text(encoding="utf-8")
        method = source.split("def _execute_native_workflow_task_request", 1)[1].split("\n    def ", 1)[0]

        self.assertNotIn("effective_task_timeout + 5", method)
        self.assertIn("executor.inactivity_timeout", method)
        self.assertIn("effective_inactivity_timeout", method)

    def test_redaction_masks_unknown_secret_values_by_key_shape(self) -> None:
        text = "\n".join(
            [
                "JOB_HELPER_TOKEN_SECRET=fake-secret-value-for-redaction",
                '{"name":"INTERNAL_ADMIN_READ_API_KEY","value":"admin-read-key-from-cluster"}',
                '{"value":"reverse-order-secret","name":"DEPIN_ADMIN_READ_API_KEY"}',
                '{"DEPIN_ADMIN_READ_API_KEY":"object-key-secret"}',
                '{"message":"missing authorization header"}',
            ]
        )

        redacted = _redact_subprocess_output(text, secret_values=[])

        self.assertIn("JOB_HELPER_TOKEN_SECRET=[redacted]", redacted)
        self.assertIn('"name":"INTERNAL_ADMIN_READ_API_KEY","value":"[redacted]"', redacted)
        self.assertIn('"value":"[redacted]","name":"DEPIN_ADMIN_READ_API_KEY"', redacted)
        self.assertIn('"DEPIN_ADMIN_READ_API_KEY":"[redacted]"', redacted)
        self.assertIn("missing authorization header", redacted)
        self.assertNotIn("fake-secret-value-for-redaction", redacted)
        self.assertNotIn("admin-read-key-from-cluster", redacted)
        self.assertNotIn("reverse-order-secret", redacted)
        self.assertNotIn("object-key-secret", redacted)

    def test_redaction_masks_structured_secret_name_value_payloads(self) -> None:
        payload = {
            "name": "JOB_HELPER_TOKEN_SECRET",
            "value": "cluster-only-secret",
            "nested": {"DEPIN_ADMIN_READ_API_KEY": "nested-secret"},
        }

        redacted = _redact_json_value(payload, secret_values=[], limit=1000)

        self.assertEqual(redacted["name"], "JOB_HELPER_TOKEN_SECRET")
        self.assertEqual(redacted["value"], "[redacted]")
        self.assertEqual(redacted["nested"]["DEPIN_ADMIN_READ_API_KEY"], "[redacted]")

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
        self.assertFalse(runtime.metadata["observation_sink_configured"])
        self.assertEqual(runtime.metadata["observation_sink_status"], "not_configured")
        self.assertTrue(runtime.metadata["direct_delivery_phase_enabled"])

    def test_runtime_metadata_exposes_direct_observation_sink(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "RSI_RUNTIME_OBSERVATION_SINK_URL": "http://control-plane.internal:8080/internal/runtime/observations"},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertTrue(runtime.metadata["observation_sink_configured"])
        self.assertEqual(runtime.metadata["observation_sink_status"], "configured")
        self.assertEqual(runtime.metadata["live_stream_status"], "configured")

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
                    "prompt": "User request: Please use /architecture-diagram skill for this system diagram.\n\nInvestigate using Hermes-native tools and configured MCP servers.",
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
                    "prompt": "User request: Please use /architecture-diagram skill and summarize the workflow.\n\nInvestigate using Hermes-native tools and configured MCP servers.",
                    "requested_skills": ["architecture-diagram"],
                    "reply_delivery_mode": "mediated",
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
                    "prompt": "User request: Please use /missing-skill.\n\nInvestigate using Hermes-native tools and configured MCP servers.",
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
        self.assertNotIn("repo_context", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("rsi_candidate_context", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("github.create_pr", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("honcho_conclude", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("tool_policy_mode", result.raw)
        self.assertNotIn("blocked_tool_names", result.raw)
        self.assertNotIn("tool_allowlist_effective", result.raw)
        self.assertNotIn("Blocked tools", FakeAIAgent.last_prompt or "")
        self.assertNotIn("Tool allowlist", FakeAIAgent.last_prompt or "")
        self.assertNotIn("github.create_pr", FakeAIAgent.last_prompt or "")

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
            runner_env("proposal"),
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        toolsets = runtime._native_toolsets_for_task(task)

        self.assertIn("todo", toolsets)
        self.assertIn("session_search", toolsets)

    def test_workflow_toolsets_never_include_governed_entries(self) -> None:
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
            runner_env("proposal"),
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        toolsets = runtime._native_toolsets_for_task(task)

        self.assertEqual(toolsets.count("rsi-governed-readonly"), 0)
        self.assertEqual(toolsets.count("rsi-governed-workspace"), 0)

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
            runner_env("prod"),
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        toolsets = runtime._native_toolsets_for_task(task)

        self.assertEqual(toolsets.count("rsi-governed-readonly"), 0)
        self.assertEqual(toolsets.count("rsi-governed-workspace"), 0)

    def test_workflow_artifact_toolset_requires_artifact_task_scope(self) -> None:
        lease_only_task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Investigate the active workflow.",
                    "execution_mode": "diagnose",
                    "capability_leases": [
                        {"capability": "artifact_write", "granted": True},
                    ],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
        requested_artifact_task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Draw a diagram.",
                    "execution_mode": "diagnose",
                    "capability_leases": [
                        {"capability": "artifact_write", "granted": True},
                    ],
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
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
            runner_env("prod"),
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertNotIn("rsi-artifacts", runtime._native_toolsets_for_task(lease_only_task))
        self.assertIn("rsi-artifacts", runtime._native_toolsets_for_task(requested_artifact_task))

    def test_prod_role_hides_legacy_tool_policy_from_native_prompt(self) -> None:
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
        self.assertNotIn("repo_context", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("knowledge_context", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("github.create_pr", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("tool_policy_mode", result.raw)
        self.assertEqual(result.raw["task_timeout_seconds"], 1800)
        self.assertEqual(result.raw["transport_timeout_seconds"], 1830)
        self.assertNotIn("blocked_tool_names", result.raw)
        self.assertNotIn("tool_allowlist_effective", result.raw)
        self.assertNotIn("Blocked tools", FakeAIAgent.last_prompt or "")
        self.assertNotIn("Tool allowlist", FakeAIAgent.last_prompt or "")
        self.assertNotIn("github.create_pr", FakeAIAgent.last_prompt or "")

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
        self.assertNotIn("workspace_git_history", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("workspace_read_file", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("workspace_write_file", FakeAIAgent.last_valid_tool_names)
        self.assertNotIn("tool_allowlist_effective", result.raw)
        self.assertNotIn("blocked_tool_names", result.raw)

    def test_question_reduce_uses_openrouter_hermes_reducer_without_tools(self) -> None:
        class QuestionReducerAIAgent:
            last_kwargs: dict[str, object] = {}
            run_history: list[dict[str, object]] = []

            def __init__(self, **kwargs) -> None:
                type(self).last_kwargs = kwargs

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict[str, object]:
                type(self).run_history.append(
                    {
                        "prompt": prompt,
                        "system_message": system_message,
                        "history": list(conversation_history or []),
                        "task_id": task_id,
                    }
                )
                return {
                    "final_response": json.dumps(
                        {
                            "reply_markdown": "Partial rundown: pagination cleanup landed, but the weekly picture is incomplete.",
                            "confidence": 0.68,
                            "alignment_degraded": True,
                            "alignment_notice": "NUMO alignment is degraded because no fresh canonical project ledger was available.",
                        }
                    )
                }

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

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.AIAgent", QuestionReducerAIAgent
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": tempdir,
                "RSI_HERMES_EXECUTOR_WORKSPACE_ROOT": str(Path(tempdir) / "workspace"),
            },
            clear=True,
        ), mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)
            log_path = result.raw["native_execution_log_path"]
            self.assertTrue(os.path.exists(log_path))
            with open(log_path, encoding="utf-8") as handle:
                events = [json.loads(line) for line in handle if line.strip()]

        self.assertTrue(result.ok)
        self.assertEqual(QuestionReducerAIAgent.last_kwargs["provider"], "openrouter")
        self.assertEqual(QuestionReducerAIAgent.last_kwargs["enabled_toolsets"], [])
        self.assertEqual(QuestionReducerAIAgent.run_history[0]["history"], [])
        self.assertEqual(result.raw["question_reduce_mode"], "hermes_reducer")
        self.assertEqual(result.raw["completion_verdict"], "partial")
        self.assertEqual(result.raw["termination_reason"], "task_timeout")
        self.assertEqual(result.raw["structured_output"]["reply_markdown"], "Partial rundown: pagination cleanup landed, but the weekly picture is incomplete.")
        self.assertEqual(events[0]["event"], "execution_started")
        self.assertEqual(events[1]["event"], "hermes_reducer_request")
        self.assertEqual(events[2]["event"], "hermes_reducer_response")
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
            ) as cleanup_mock:
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(register_mock.call_count, 1)
        self.assertEqual(cleanup_mock.call_count, 1)
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

    def test_transport_tool_schemas_are_provider_strict(self) -> None:
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

        artifact_write = transport_tool_schema("artifact.write_file")
        self.assertEqual(artifact_write["name"], "artifact_write_file")
        self.assertEqual(artifact_write["parameters"]["required"], ["path", "content"])
        self.assertEqual(artifact_write["parameters"]["properties"]["path"]["type"], "string")
        self.assertEqual(artifact_write["parameters"]["properties"]["content"]["type"], "string")
        self.assertFalse(artifact_write["parameters"]["additionalProperties"])

        artifact_list = transport_tool_schema("artifact.list_files")
        self.assertEqual(artifact_list["name"], "artifact_list_files")
        self.assertEqual(artifact_list["parameters"]["required"], ["path"])
        self.assertEqual(artifact_list["parameters"]["properties"]["path"]["type"], ["string", "null"])

        invalid: dict[str, list[str]] = {}
        from rsi_runner.rsi_tools import rsi_plugin_toolset_definitions
        for item in rsi_plugin_toolset_definitions():
            schema = item["schema"]
            paths = strict_schema_violations(schema.get("parameters"), "parameters")
            if paths:
                invalid[schema["name"]] = paths
        self.assertEqual(invalid, {})


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
            ) as cleanup_mock:
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertGreaterEqual(register_mock.call_count, 1)
        self.assertGreaterEqual(cleanup_mock.call_count, 1)
        self.assertEqual(
            FakeAIAgent.last_kwargs["enabled_toolsets"],
            [
                "todo",
                "session_search",
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
                    included_tool_names=[
                        "API-post-search",
                        "API-retrieve-a-page",
                        "API-get-block-children",
                    ],
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
                            "server_url": "https://docs.example.com/mcp",
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
        self.assertNotIn("tool_policy_mode", result.raw)

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
        self.assertNotIn("tool_policy_mode", result.raw)
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
                if self.instance_index > 1:
                    return {
                        "final_response": json.dumps(
                            partial_structured_output(
                                reply_text="Partial answer: grounded summary so far.",
                                proposed_actions=[],
                            )
                        )
                    }
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
                    "reply_delivery_mode": "mediated",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )
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

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", BudgetReducerAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.object(
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
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "hermes_reducer")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_attempts"], 1)
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_attempted"])
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_succeeded"])
        self.assertEqual(BudgetReducerAIAgent.created_instances, 2)
        self.assertEqual(BudgetReducerAIAgent.init_history[0]["max_iterations"], 20)
        self.assertEqual(BudgetReducerAIAgent.init_history[1]["max_iterations"], 1)
        self.assertEqual(BudgetReducerAIAgent.init_history[1]["enabled_toolsets"], [])
        self.assertEqual(BudgetReducerAIAgent.init_history[1]["provider"], "openrouter")
        self.assertNotIn("repo_context", BudgetReducerAIAgent.run_history[0]["valid_tool_names"])
        self.assertEqual(BudgetReducerAIAgent.run_history[1]["valid_tool_names"], [])
        self.assertIn('"termination_reason": "iteration_budget_exhausted"', BudgetReducerAIAgent.run_history[1]["prompt"])
        self.assertNotIn("Earlier thread message", BudgetReducerAIAgent.run_history[1]["prompt"])
        self.assertTrue(any("iteration_budget_exhausted" in message for message in BudgetReducerAIAgent.interrupt_messages))
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
                                    "name": "send_message",
                                    "arguments": json.dumps(
                                        {
                                            "target": "slack:C123:171000001.000100",
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
                                        "success": True,
                                        "platform": "slack",
                                        "chat_id": "C123",
                                        "thread_id": "171000001.000100",
                                        "message_id": "171000001.000100",
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
                                        "name": "send_message",
                                        "arguments": json.dumps(
                                            {
                                                "target": "slack:C123:171000001.000100",
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
                                            "success": True,
                                            "platform": "slack",
                                            "chat_id": "C123",
                                            "thread_id": "171000001.000100",
                                            "message_id": "171000001.000100",
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
        self.assertNotIn("slack_upload_file", ArtifactWorkflowAIAgent.run_history[2]["valid_tool_names"])
        produced = result.raw["structured_output"]["produced_artifacts"]
        self.assertEqual(len(produced), 1)
        self.assertTrue(artifact_ref.startswith("file://"))
        self.assertTrue(artifact_exists)
        self.assertEqual(result.raw["reply_delivery"]["status"], "posted")
        self.assertEqual(result.raw["runner_diagnostics"]["artifact_phase_enabled"], True)
        self.assertEqual(result.raw["runner_diagnostics"]["observation_seq"] > 0, True)
        envelope = result.raw["execution_envelope"]
        self.assertEqual(envelope["contract_version"], "execution-envelope/v1")
        self.assertEqual(envelope["execution_plan"]["mode"], "runner_first")
        self.assertEqual([item["phase_type"] for item in envelope["phase_runs"]], ["plan", "investigate", "render", "deliver", "reflect"])
        planned_events = [item for item in envelope["ledger_events"] if item["kind"] == "phase.planned"]
        self.assertEqual(len(planned_events), len(envelope["phase_runs"]))
        self.assertEqual(
            [item["phase_id"] for item in planned_events],
            [item["phase_id"] for item in envelope["phase_runs"]],
        )
        self.assertEqual({item["payload"]["status"] for item in planned_events}, {"planned"})
        self.assertEqual([item["payload"].get("output_refs") for item in planned_events], [[] for _ in planned_events])
        self.assertEqual(envelope["artifacts"][0]["file_ref"], artifact_ref)
        self.assertIn("workspace_path", envelope["artifacts"][0])
        self.assertIn("sha256", envelope["artifacts"][0])
        self.assertEqual(envelope["deliveries"][0]["send_status"], "posted")
        self.assertIn("artifact.created", {item["kind"] for item in envelope["ledger_events"]})
        self.assertIn("slack.message.sent", {item["kind"] for item in envelope["ledger_events"]})
        slack_events = [item for item in envelope["ledger_events"] if item["kind"] == "slack.message.sent"]
        self.assertEqual(slack_events[0]["status"], "posted")

    def test_execution_envelope_contract_fails_closed_for_unknown_capability(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Answer safely.",
                    "trace_id": "trace-contract",
                    "workflow_id": "wf-contract",
                    "contract_version": "execution-envelope/v1",
                    "capability_leases": [{"capability": "aws_admin"}],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-contract",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_contract_failed")
        self.assertEqual(result.raw["execution_envelope"]["completion"]["termination_reason"], "runner_contract_failed")
        self.assertIn("Unknown capability lease aws_admin.", result.message)

    def test_observation_ledger_event_preserves_empty_payload(self) -> None:
        computer = HermesCompanyComputer(
            computer_root="/workspace/company",
            run_root="/workspace/company/.rsi/runs",
            artifact_root="/workspace/company/artifacts",
            hermes_pin=HERMES_TEST_PIN,
        )
        event = computer._ledger_event_from_observation(
            {
                "execution_id": "hexec-test",
                "hermes_session_id": "session-test",
                "event_type": "phase.started",
                "phase": "operate",
                "payload": {},
                "role": "prod",
                "seq": 1,
                "status": "running",
            },
            sequence=1,
            execution_id="hexec-test",
        )

        self.assertEqual(event["payload"], {})

    def test_execution_envelope_promotes_generic_write_file_for_artifact_intent(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            computer_root = Path(tempdir) / "company"
            artifact_root = computer_root / "artifacts"
            computer_root.mkdir(parents=True)
            source_path = computer_root / "depin-backend-architecture.html"
            source_path.write_text("<html><body><svg></svg></body></html>", encoding="utf-8")
            task = types.SimpleNamespace(
                task_type="workflow",
                repo="depin-backend",
                trace_id="trace-artifact",
                workflow_id="wf-artifact",
                operation_id="op-artifact",
                requested_artifacts=[],
                capability_leases=[{"capability": "artifact_write", "granted": True}],
                execution_intent={"kind": "architecture", "user_request": "draw an architecture diagram"},
                workspace_policy={
                    "computer_root": str(computer_root),
                    "run_root": str(computer_root / ".rsi" / "runs"),
                    "artifact_root": str(artifact_root),
                    "allowed_path_roots": [str(computer_root), str(artifact_root)],
                },
            )
            observer = types.SimpleNamespace(
                execution_id="hexec-artifact",
                events=lambda: [
                    {
                        "execution_id": "hexec-artifact",
                        "trace_id": "trace-artifact",
                        "workflow_id": "wf-artifact",
                        "event_type": "tool.call.completed",
                        "phase": "main",
                        "status": "completed",
                        "seq": 7,
                        "payload": {
                            "tool_name": "write_file",
                            "tool_call_id": "call-write",
                            "args": {"path": "depin-backend-architecture.html"},
                            "result": json.dumps({"bytes_written": source_path.stat().st_size}),
                        },
                        "recorded_at": "2026-04-25T23:20:00Z",
                    }
                ],
            )
            result = types.SimpleNamespace(
                ok=True,
                message="done",
                raw={"structured_output": {"final_answer": "done"}},
            )
            computer = HermesCompanyComputer(
                computer_root=str(computer_root),
                run_root=str(computer_root / ".rsi" / "runs"),
                artifact_root=str(artifact_root),
                hermes_pin=HERMES_TEST_PIN,
            )

            result = computer.attach_envelope(task, result, observer=observer)

            artifact = result.raw["execution_envelope"]["artifacts"][0]
            artifact_path = Path(artifact["workspace_path"])
            self.assertEqual(artifact["kind"], "architecture")
            self.assertTrue(str(artifact_path).startswith(str(artifact_root)))
            self.assertTrue(artifact_path.exists())
            self.assertNotEqual(str(artifact_path), str(source_path))
            self.assertEqual(result.raw["structured_output"]["produced_artifacts"][0]["file_ref"], artifact["file_ref"])
            ledger_kinds = {item["kind"] for item in result.raw["execution_envelope"]["ledger_events"]}
            self.assertIn("artifact.created", ledger_kinds)
            self.assertIn("file.written", ledger_kinds)

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

    def test_artifact_phase_budgets_scale_when_total_is_small(self) -> None:
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
        self.assertGreater(budgets["investigate"], budgets["deliver"])
        self.assertEqual(budgets["investigate"], budgets["render"])
        self.assertLessEqual(
            budgets["investigate"] + budgets["render"] + budgets["deliver"] + budgets["reducer_reserve"],
            budgets["total"],
        )

    def test_explicit_artifact_timeout_can_exceed_default_task_timeout(self) -> None:
        env = {
            **runner_env("prod"),
            "RSI_RUNNER_PROD_TIMEOUT": "4110s",
            "RSI_RUNNER_PROD_TASK_TIMEOUT": "1800s",
            "RSI_RUNNER_PROD_ARTIFACT_MAX_ITERATIONS": "40",
        }
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Render an architecture diagram artifact.",
                    "timeout_seconds": 4080,
                    "requested_skills": ["architecture-diagram"],
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime._effective_task_timeout(task), 4080)
        self.assertEqual(runtime._phase_max_iterations_override(task), 40)
        budgets = runtime._artifact_phase_budgets(task)
        self.assertEqual(budgets["total"], 4080)
        self.assertEqual(budgets["investigate"], 1800)
        self.assertEqual(budgets["render"], 1800)
        self.assertEqual(budgets["deliver"], 300)
        self.assertEqual(budgets["reducer_reserve"], 180)

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
            native_toolsets = runtime._native_toolsets_for_task(task)

        self.assertNotIn("rsi-governed-workspace", native_toolsets)

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
            result = runtime._execute_artifact_workflow_task(task)

        self.assertTrue(result.ok)
        self.assertEqual([item.execution_phase for item in executed_tasks], ["investigate", "render", "deliver"])
        expected_artifact_root = str((Path(tempdir) / "company" / "artifacts" / "depin-backend" / time.strftime("%Y-%m-%d", time.gmtime()) / "op-artifacts").resolve())
        self.assertTrue(all(item.artifact_destination == expected_artifact_root for item in executed_tasks))
        self.assertIn(f"Output path: {expected_artifact_root}", executed_tasks[1].prompt)
        self.assertNotIn("/var/lib/hermes/rsi_runtime/artifacts", executed_tasks[1].prompt)

    def test_artifact_workflow_does_not_render_timeout_fallback_brief(self) -> None:
        executed_tasks: list[RunnerTaskRequest] = []

        def fake_execute_task_internal(task: RunnerTaskRequest, observer=None):
            executed_tasks.append(task)
            if task.execution_phase == "investigate":
                return HermesExecutionResult(
                    ok=True,
                    message="investigate timeout",
                    provider="test",
                    raw={
                        "completion_verdict": "partial",
                        "termination_reason": "task_timeout",
                        "structured_output": {
                            "final_answer": "I could not gather enough grounded evidence before timing out.",
                            "reply_draft": "I could not gather enough grounded evidence before timing out.",
                            "context_summary": "Terminated before evidence collection completed.",
                            "artifact_render_briefs": [],
                            "produced_artifacts": [],
                            "artifact_failure_reason": "",
                        },
                    },
                )
            if task.execution_phase == "render":
                self.fail("timed-out investigation without explicit briefs must not enter render")
            self.assertEqual(task.execution_phase, "deliver")
            self.assertIn("No file artifacts were produced.", task.prompt)
            self.assertNotIn("slack.upload_file", task.allowed_tools)
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
                        "trace_id": "trace-artifacts-timeout",
                        "workflow_id": "wf-artifacts-timeout",
                        "operation_id": "op-artifacts-timeout",
                        "channel_id": "C123",
                        "thread_ts": "171000001.000100",
                        "reply_delivery_mode": "direct",
                        "requested_skills": ["architecture-diagram"],
                        "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                    }
                }
            )
            result = runtime._execute_artifact_workflow_task(task)

        self.assertTrue(result.ok)
        self.assertEqual([item.execution_phase for item in executed_tasks], ["investigate", "deliver"])
        self.assertEqual(result.raw["structured_output"]["produced_artifacts"], [])
        self.assertIn(
            "Artifact render skipped because investigation ended with task_timeout before producing renderable briefs.",
            result.raw["structured_output"]["artifact_failure_reason"],
        )
        self.assertEqual(result.raw["runner_diagnostics"]["artifact_delivery_branch"], "text_only")

    def test_artifact_workflow_renders_grounded_timeout_evidence_brief(self) -> None:
        executed_tasks: list[RunnerTaskRequest] = []

        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir:

            def fake_execute_task_internal(task: RunnerTaskRequest, observer=None):
                executed_tasks.append(task)
                if task.execution_phase == "investigate":
                    return HermesExecutionResult(
                        ok=True,
                        message="investigate timeout",
                        provider="test",
                        raw={
                            "completion_verdict": "partial",
                            "termination_reason": "task_timeout",
                            "structured_output": {
                                "final_answer": "Partial answer from grounded evidence.",
                                "reply_draft": "Partial answer from grounded evidence.",
                                "context_summary": "story-deployments and Kubernetes evidence were loaded.",
                                "artifact_render_briefs": [],
                                "produced_artifacts": [],
                                "artifact_failure_reason": "",
                            },
                            "evidence_ledger": {
                                "termination_reason": "task_timeout",
                                "tool_calls": [
                                    {
                                        "tool_name": "repo.read_file",
                                        "request": {
                                            "repo": "story-deployments",
                                            "path": "story/depin-backend/use1-prod.yaml",
                                        },
                                        "status": "ok",
                                        "summary": "Repository file story/depin-backend/use1-prod.yaml loaded.",
                                    },
                                    {
                                        "tool_name": "kubernetes.inspect",
                                        "request": {"namespace": "story", "target": "depin-backend"},
                                        "status": "ok",
                                        "summary": "Kubernetes inspection found two running pods.",
                                    },
                                ],
                                "evidence_items": [
                                    {
                                        "kind": "repo_file",
                                        "summary": "Production depin-backend Helm values.",
                                        "source_ref": "story/depin-backend/use1-prod.yaml",
                                        "path": "story/depin-backend/use1-prod.yaml",
                                        "repo": "story-deployments",
                                        "tool_name": "repo.read_file",
                                    },
                                    {
                                        "kind": "kubernetes_inspect",
                                        "summary": "Live namespace story has two running pods.",
                                        "source_ref": "kubernetes://story/depin-backend",
                                        "tool_name": "kubernetes.inspect",
                                    },
                                ],
                                "open_questions": [],
                            },
                        },
                    )
                if task.execution_phase == "render":
                    self.assertIn("bounded-stop investigation", task.prompt)
                    self.assertIn("story/depin-backend/use1-prod.yaml", task.prompt)
                    self.assertIn("kubernetes://story/depin-backend", task.prompt)
                    artifact_path = Path(task.artifact_destination) / "architecture.html"
                    artifact_path.parent.mkdir(parents=True, exist_ok=True)
                    artifact_path.write_text("<html>diagram</html>", encoding="utf-8")
                    return HermesExecutionResult(
                        ok=True,
                        message="render",
                        provider="test",
                        raw={
                            "structured_output": {
                                "produced_artifacts": [
                                    {
                                        "kind": "diagram",
                                        "title": "Architecture diagram",
                                        "workspace_path": str(artifact_path),
                                        "file_ref": f"file://{artifact_path}",
                                        "artifact_refs": [f"file://{artifact_path}"],
                                    }
                                ]
                            }
                        },
                )
                self.assertEqual(task.execution_phase, "deliver")
                self.assertEqual(task.allowed_tools, [])
                self.assertIn("Attach produced local artifacts through Hermes native send_message", task.prompt)
                self.assertIn("Hermes native send_message", task.prompt)
                self.assertNotIn("Produced artifacts:", task.prompt)
                return HermesExecutionResult(
                    ok=True,
                    message="deliver",
                    provider="test",
                    raw={"structured_output": {"reply_delivery": {"status": "posted", "provider_ref": "slack-message-1"}}},
                )

            with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
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
                            "prompt": "Render the architecture diagram from repo and story-deployments.",
                            "trace_id": "trace-artifacts-grounded-timeout",
                            "workflow_id": "wf-artifacts-grounded-timeout",
                            "operation_id": "op-artifacts-grounded-timeout",
                            "channel_id": "C123",
                            "thread_ts": "171000001.000100",
                            "reply_delivery_mode": "direct",
                            "requested_skills": ["architecture-diagram"],
                            "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                        }
                    }
                )
                result = runtime._execute_artifact_workflow_task(task)

        self.assertTrue(result.ok)
        self.assertEqual([item.execution_phase for item in executed_tasks], ["investigate", "render", "deliver"])
        self.assertEqual(result.raw["runner_diagnostics"]["artifact_delivery_branch"], "native_slack_media_with_artifacts")
        self.assertEqual(result.raw["structured_output"]["reply_delivery"]["status"], "posted")
        self.assertTrue(result.raw["structured_output"]["produced_artifacts"])
        self.assertEqual(result.raw["structured_output"]["produced_artifacts"][0]["share_status"], "shared")
        self.assertIn("artifact_render_briefs", result.raw["structured_output"])

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
            self.assertEqual(task.allowed_tools, [])
            self.assertIn("Attach produced local artifacts through Hermes native send_message", task.prompt)
            self.assertIn("Hermes native send_message", task.prompt)
            self.assertNotIn("Produced artifacts:", task.prompt)
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
            result = runtime._execute_artifact_workflow_task(task)

        self.assertTrue(result.ok)
        produced = result.raw["structured_output"]["produced_artifacts"]
        self.assertEqual(len(produced), 1)
        self.assertTrue(produced[0]["artifact_refs"][0].startswith("file://"))

    def test_artifact_workflow_treats_failed_reply_delivery_as_failed(self) -> None:
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
                raw={"structured_output": {"reply_delivery": {"status": "failed", "tool_name": "slack.upload_file"}}},
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
                        "trace_id": "trace-delivery-failed",
                        "workflow_id": "wf-delivery-failed",
                        "operation_id": "op-delivery-failed",
                        "channel_id": "C123",
                        "thread_ts": "171000001.000100",
                        "reply_delivery_mode": "direct",
                        "requested_skills": ["architecture-diagram"],
                        "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram"}],
                    }
                }
            )
            result = runtime._execute_artifact_workflow_task(task)

        self.assertTrue(result.ok)
        self.assertEqual([item.execution_phase for item in executed_tasks], ["investigate", "render", "deliver"])
        self.assertEqual(result.raw["structured_output"]["reply_delivery"]["status"], "failed")
        self.assertEqual(
            result.raw["runner_diagnostics"]["direct_delivery_phase_failed"],
            "direct delivery phase reported unsuccessful status failed",
        )

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

        self.assertEqual(deliver_task.allowed_tools, [])
        self.assertEqual(deliver_task.tool_allowlist, [])
        self.assertIn("Hermes native send_message", deliver_task.prompt)
        self.assertIn("Render failure reason: none", deliver_task.prompt)
        self.assertIn("Do not upload files", deliver_task.prompt)

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
                if self.instance_index > 1:
                    return {
                        "final_response": json.dumps(
                            partial_structured_output(reply_text="Partial answer after timeout.")
                        )
                    }
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
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", TimeoutReducerAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
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
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "hermes_reducer")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_attempts"], 1)
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_attempted"])
        self.assertEqual(TimeoutReducerAIAgent.created_instances, 2)
        self.assertEqual(TimeoutReducerAIAgent.init_history[1]["enabled_toolsets"], [])
        self.assertEqual(TimeoutReducerAIAgent.run_history[1]["valid_tool_names"], [])
        self.assertIn('"termination_reason": "task_timeout"', TimeoutReducerAIAgent.run_history[1]["prompt"])
        self.assertNotIn("Earlier thread message", TimeoutReducerAIAgent.run_history[1]["prompt"])
        self.assertTrue(any(message == "runner task_timeout after 10s" for message in TimeoutReducerAIAgent.interrupt_messages))

    def test_workflow_iteration_budget_exhaustion_fails_when_hermes_reducer_cannot_return_valid_output(self) -> None:
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
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", InvalidReducerAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
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
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_mode"], "hermes_reducer")
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_attempts"], 1)
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_attempted"])
        self.assertFalse(result.raw["runner_diagnostics"]["partial_finalization_retry_succeeded"])
        self.assertEqual(result.raw["runner_diagnostics"]["partial_finalization_timeout_seconds"], 180)
        self.assertEqual(InvalidReducerAIAgent.created_instances, 2)
        self.assertEqual(InvalidReducerAIAgent.init_history[1]["enabled_toolsets"], [])
        self.assertEqual(InvalidReducerAIAgent.init_history[1]["provider"], "openrouter")
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

        rendered = runtime._render_task_prompt(task)
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

    def test_workflow_evidence_ledger_folds_native_lifecycle_tool_events(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Draw the depin-backend architecture from repo and story-deployments.",
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram."}],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-native-lifecycle",
                    "workflow_id": "wf-native-lifecycle",
                }
            }
        )
        lifecycle_events = [
            {
                "event": "tool_call_started",
                "recorded_at_unix": 1710000001.0,
                "tool_name": "repo.search",
                "request_payload": {"repo": "story-deployments", "pattern": "depin-backend"},
            },
            {
                "event": "tool_call_completed",
                "recorded_at_unix": 1710000002.0,
                "tool_name": "repo.search",
                "status": "completed",
                "summary": "Found depin backend deployment values.",
                "provider_ref": "story-deployments",
                "output": {
                    "repo": "story-deployments",
                    "pattern": "depin-backend",
                    "matches": [
                        {
                            "path": "story/depin-backend/use1-stage.yaml",
                            "snippet": "replicaCount: 2\nimage:\n  repository: depin-backend-api",
                        }
                    ],
                },
            },
            {
                "event": "tool_call_completed",
                "recorded_at_unix": 1710000003.0,
                "tool_name": "repo.read_file",
                "status": "completed",
                "summary": "Read deployed worker values.",
                "output": {
                    "repo": "story-deployments",
                    "path": "story/depin-ip-registration/use1-stage.yaml",
                    "content": "poller:\n  enabled: true\nsubmitter:\n  enabled: true\nconfirmer:\n  enabled: true",
                },
            },
        ]
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        observed = runtime._observability_metadata(None, task, lifecycle_events=lifecycle_events)
        ledger = runtime._build_evidence_ledger(task, observed, "task_timeout")

        self.assertEqual([item["tool_name"] for item in ledger["tool_calls"]], ["repo.search", "repo.read_file"])
        self.assertEqual(ledger["tool_calls"][0]["request"], {"repo": "story-deployments", "pattern": "depin-backend"})
        self.assertEqual(ledger["evidence_items"][0]["path"], "story/depin-backend/use1-stage.yaml")
        self.assertIn("replicaCount: 2", ledger["evidence_items"][0]["snippet"])
        self.assertEqual(ledger["evidence_items"][1]["path"], "story/depin-ip-registration/use1-stage.yaml")
        self.assertNotIn("No grounded evidence", " ".join(ledger["open_questions"]))

    def test_workflow_evidence_ledger_preserves_late_sot_evidence_after_long_run(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Draw all depin-backend architecture using the repo plus story-deployments as source of truth.",
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram."}],
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-long-sot-run",
                    "workflow_id": "wf-long-sot-run",
                }
            }
        )
        tool_calls: list[dict[str, object]] = []
        evidence_items: list[dict[str, object]] = []
        for index in range(72):
            tool_calls.append(
                {
                    "tool_name": "repo.search",
                    "tool_call_id": f"repo.search:{index}",
                    "request": {"repo": "depin-backend", "pattern": f"old-api-{index}"},
                    "summary": f"Found early depin backend source evidence {index}.",
                    "status": "completed",
                }
            )
            evidence_items.append(
                {
                    "kind": "repo_search_match",
                    "summary": f"Early depin backend source evidence {index}.",
                    "source_ref": f"docs/old-{index}.md",
                    "tool_name": "repo.search",
                    "path": f"docs/old-{index}.md",
                    "repo": "depin-backend",
                }
            )
        tool_calls.extend(
            [
                {
                    "tool_name": "repo.read_file",
                    "tool_call_id": "repo.read_file:stage-api",
                    "request": {"repo": "story-deployments", "path": "story/depin-backend/use1-prod.yaml"},
                    "summary": "Read deployed depin-backend API values from story-deployments.",
                    "status": "completed",
                },
                {
                    "tool_name": "repo.read_file",
                    "tool_call_id": "repo.read_file:stage-worker",
                    "request": {"repo": "story-deployments", "path": "story/depin-ip-registration/use1-prod.yaml"},
                    "summary": "Read deployed ip-registration worker values from story-deployments.",
                    "status": "completed",
                },
                {
                    "tool_name": "repo.read_file",
                    "tool_call_id": "repo.read_file:applicationset",
                    "request": {"repo": "story-deployments", "path": "applicationset/use1-prod.yaml"},
                    "summary": "Read Argo applicationset values for story deployments.",
                    "status": "completed",
                },
                {
                    "tool_name": "kubernetes.inspect",
                    "tool_call_id": "kubernetes.inspect:depin",
                    "request": {"namespace": "story", "target": "depin-backend"},
                    "summary": "Live namespace story has two running depin-backend pods.",
                    "status": "completed",
                },
            ]
        )
        evidence_items.extend(
            [
                {
                    "kind": "repo_file",
                    "summary": "Production depin-backend API Helm values.",
                    "source_ref": "story/depin-backend/use1-prod.yaml",
                    "tool_name": "repo.read_file",
                    "path": "story/depin-backend/use1-prod.yaml",
                    "repo": "story-deployments",
                },
                {
                    "kind": "repo_file",
                    "summary": "Production depin-ip-registration Helm values.",
                    "source_ref": "story/depin-ip-registration/use1-prod.yaml",
                    "tool_name": "repo.read_file",
                    "path": "story/depin-ip-registration/use1-prod.yaml",
                    "repo": "story-deployments",
                },
                {
                    "kind": "repo_file",
                    "summary": "Production applicationset values.",
                    "source_ref": "applicationset/use1-prod.yaml",
                    "tool_name": "repo.read_file",
                    "path": "applicationset/use1-prod.yaml",
                    "repo": "story-deployments",
                },
                {
                    "kind": "kubernetes_inspect",
                    "summary": "Live namespace story has two running depin-backend pods.",
                    "source_ref": "kubernetes://story/depin-backend",
                    "tool_name": "kubernetes.inspect",
                },
            ]
        )
        for index in range(36):
            tool_calls.append(
                {
                    "tool_name": "terminal.run",
                    "tool_call_id": f"terminal.run:{index}",
                    "request": {"command": f"printf trailing-{index}"},
                    "summary": f"Trailing terminal observation {index}.",
                    "status": "completed",
                }
            )
            evidence_items.append(
                {
                    "kind": "terminal_output",
                    "summary": f"Trailing non-SoT output {index}.",
                    "source_ref": f"terminal://trailing-{index}",
                    "tool_name": "terminal.run",
                }
            )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        ledger = runtime._build_evidence_ledger(task, {"tool_calls": tool_calls, "evidence_items": evidence_items}, "task_timeout")
        tool_blob = json.dumps(ledger["tool_calls"], sort_keys=True)
        evidence_blob = json.dumps(ledger["evidence_items"], sort_keys=True)

        self.assertLessEqual(len(ledger["tool_calls"]), 30)
        self.assertLessEqual(len(ledger["evidence_items"]), 40)
        self.assertIn("story/depin-backend/use1-prod.yaml", tool_blob)
        self.assertIn("story/depin-ip-registration/use1-prod.yaml", tool_blob)
        self.assertIn("applicationset/use1-prod.yaml", tool_blob)
        self.assertIn("kubernetes.inspect", tool_blob)
        self.assertIn("story/depin-backend/use1-prod.yaml", evidence_blob)
        self.assertIn("story/depin-ip-registration/use1-prod.yaml", evidence_blob)
        self.assertIn("applicationset/use1-prod.yaml", evidence_blob)
        self.assertIn("kubernetes://story/depin-backend", evidence_blob)
        self.assertNotIn("No grounded evidence", " ".join(ledger["open_questions"]))

    def test_hermes_adapter_keeps_lifecycle_history_beyond_last_eight_events(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": tempdir},
            clear=True,
        ):
            adapter = HermesAdapter(RunnerConfig.from_env())
            lifecycle_dir = Path(tempdir) / "rsi_runtime" / "lifecycle"
            lifecycle_dir.mkdir(parents=True, exist_ok=True)
            lifecycle_path = lifecycle_dir / "session-123.jsonl"
            lifecycle_path.write_text(
                "\n".join(json.dumps({"event": "tool_call_completed", "tool_name": "repo.search", "seq": index}) for index in range(12)),
                encoding="utf-8",
            )

            events = adapter.lifecycle_events("session-123")

        self.assertEqual(len(events), 12)
        self.assertEqual(events[0]["seq"], 0)
        self.assertEqual(events[-1]["seq"], 11)

    def test_artifact_phase_merge_preserves_investigation_lifecycle_and_partial_verdict(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Draw the depin-backend architecture.",
                    "requested_artifacts": [{"kind": "diagram", "description": "Architecture diagram."}],
                    "reply_delivery_mode": "direct",
                    "trace_id": "trace-artifact-merge",
                    "workflow_id": "wf-artifact-merge",
                    "operation_id": "op-artifact-merge",
                }
            }
        )
        investigate_result = HermesExecutionResult(
            ok=True,
            message="partial",
            provider="hermes-native-executor",
            raw={
                "completion_verdict": "partial",
                "termination_reason": "task_timeout",
                "structured_output": {
                    "final_answer": "Partial grounded answer.",
                    "reply_draft": "Partial grounded answer.",
                    "artifact_render_briefs": [],
                    "produced_artifacts": [],
                },
                "lifecycle_events": [
                    {
                        "event": "tool_call_completed",
                        "tool_name": "repo.search",
                        "summary": "Found deployed values.",
                    }
                ],
                "tool_calls": [{"tool_name": "repo.search", "tool_call_id": "repo.search:1"}],
                "evidence_items": [{"kind": "repo_search_match", "summary": "Found deployed values."}],
                "evidence_ledger": {"tool_calls": [{"tool_name": "repo.search"}], "evidence_items": [{"summary": "Found deployed values."}]},
                "runner_diagnostics": {"completion_verdict": "partial"},
            },
        )
        delivery_result = HermesExecutionResult(
            ok=True,
            message="delivered",
            provider="hermes-native-executor",
            raw={
                "completion_verdict": "complete",
                "termination_reason": "normal_completion",
                "structured_output": {
                    "reply_delivery": {
                        "status": "sent",
                        "channel_id": "C123",
                        "thread_ts": "171000001.000100",
                    }
                },
                "lifecycle_events": [
                    {
                        "event": "tool_call_completed",
                        "tool_name": "slack.reply",
                        "summary": "Sent Slack reply.",
                    }
                ],
                "runner_diagnostics": {"completion_verdict": "complete"},
            },
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        merged = runtime._merge_artifact_phase_result(
            task,
            investigate_result,
            investigate_result.raw["structured_output"],
            [],
            "render skipped after timeout",
            delivery_result=delivery_result,
            delivery_output=delivery_result.raw["structured_output"],
        )

        self.assertEqual(merged.raw["completion_verdict"], "partial")
        self.assertEqual(merged.raw["termination_reason"], "task_timeout")
        self.assertEqual([item["tool_name"] for item in merged.raw["lifecycle_events"]], ["repo.search", "slack.reply"])
        self.assertEqual(merged.raw["evidence_ledger"]["tool_calls"][0]["tool_name"], "repo.search")
        self.assertEqual(merged.raw["tool_calls"][0]["tool_name"], "repo.search")

    def test_runtime_metadata_reports_role_contract(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("proposal"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime.metadata["max_iterations"], 5)
        self.assertEqual(runtime.metadata["task_timeout_seconds"], 420)
        self.assertEqual(runtime.metadata["inactivity_timeout_seconds"], 360)
        self.assertEqual(runtime.metadata["transport_timeout_seconds"], 450)
        self.assertNotIn("tool_policy_mode", runtime.metadata)
        self.assertEqual(runtime.metadata["hermes_pin"], HERMES_TEST_PIN)
        self.assertEqual(runtime.metadata["execution_contract_version"], "execution-envelope/v1")
        self.assertEqual(runtime.metadata["runner_planner_mode"], "runner_first")
        self.assertEqual(runtime.metadata["company_computer_root"], "/tmp/hermes/workspace/company")
        self.assertIn("artifact_write", runtime.metadata["required_capabilities"])
        self.assertEqual(runtime.metadata["honcho_runtime_status"]["workspace"], "rsi-stage")
        self.assertEqual(runtime.metadata["session_continuity_status"], "ok")
        self.assertEqual(runtime.metadata["honcho_environment_effective"], "production")
        self.assertNotIn("tool_allowlist_effective", runtime.metadata)
        self.assertNotIn("blocked_tool_names", runtime.metadata)

    def test_prod_runtime_metadata_reports_live_contract(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime.metadata["max_iterations"], 20)
        self.assertEqual(runtime.metadata["task_timeout_seconds"], 1800)
        self.assertEqual(runtime.metadata["inactivity_timeout_seconds"], 1800)
        self.assertEqual(runtime.metadata["transport_timeout_seconds"], 1830)
        self.assertEqual(runtime.metadata["native_max_output_tokens"], 15000)

    def test_runtime_metadata_probes_slack_mcp(self) -> None:
        class FakeResponse:
            def __init__(self, payload: dict[str, object]) -> None:
                self._payload = payload

            def __enter__(self) -> "FakeResponse":
                return self

            def __exit__(self, *_args: object) -> None:
                return None

            def read(self) -> bytes:
                return json.dumps(self._payload).encode("utf-8")

        observed_methods: list[str] = []

        def fake_urlopen(req, timeout: int = 0):
            payload = json.loads(req.data.decode("utf-8"))
            method = str(payload["method"])
            observed_methods.append(method)
            if method == "tools/list":
                return FakeResponse(
                    {
                        "jsonrpc": "2.0",
                        "id": payload.get("id"),
                        "result": {"tools": [{"name": "search_messages", "description": "Search Slack"}]},
                    }
                )
            return FakeResponse({"jsonrpc": "2.0", "id": payload.get("id"), "result": {}})

        env = {
            **runner_env("prod"),
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_MCP_SERVER_URL": "https://slack-mcp.test/mcp",
        }
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch("rsi_runner.hermes_runtime.urlrequest.urlopen", side_effect=fake_urlopen), mock.patch.dict(
            os.environ, env, clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            metadata = runtime.metadata

        self.assertTrue(metadata["slack_mcp_available"])
        self.assertEqual(metadata["slack_mcp_tool_count"], 1)
        self.assertEqual(observed_methods, ["initialize", "notifications/initialized", "tools/list"])

    def test_direct_slack_delivery_env_uses_canonical_bot_token_name(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply in Slack.",
                    "reply_delivery_mode": "direct",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                }
            }
        )
        env = runner_env("prod")

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            direct_env, error = runtime._direct_slack_delivery_env(task)

        self.assertEqual(error, "")
        self.assertEqual(direct_env["SLACK_BOT_TOKEN"], "xoxb-test")
        self.assertEqual(
            set(direct_env),
            {"HERMES_SESSION_PLATFORM", "HERMES_SESSION_CHAT_ID", "HERMES_SESSION_THREAD_ID", "SLACK_BOT_TOKEN"},
        )

    def test_prod_task_timeout_defaults_to_1800_when_transport_is_extended(self) -> None:
        env = {**runner_env("prod"), "RSI_RUNNER_PROD_TIMEOUT": "1830s"}
        env.pop("RSI_RUNNER_PROD_TASK_TIMEOUT", None)
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Investigate normally.",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(task.timeout_seconds, 0)
        self.assertEqual(runtime.metadata["task_timeout_seconds"], 1800)
        self.assertEqual(runtime.metadata["transport_timeout_seconds"], 1830)
        self.assertEqual(runtime._effective_task_timeout(task), 1800)

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
