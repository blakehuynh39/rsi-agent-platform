from __future__ import annotations

import base64
import io
import json
import os
from pathlib import Path
import sqlite3
import subprocess
import sys
import tempfile
import threading
import time
import types
import unittest
from unittest import mock
from urllib import error as urlerror

from rsi_runner.config import RunnerConfig, RunnerConfigError
from rsi_runner.execution_contract import HermesCompanyComputer
from rsi_runner.file_utils import _json_object_from_string
from rsi_runner.hermes_adapter import HermesAdapter, _build_plugin_module
from rsi_runner.hermes_agent_adapter import (
    HermesAgentAdapter,
    HermesContractStatus,
    validate_hermes_contract,
    _validate_session_db_integrity,
)
from rsi_runner.hermes_mcp_adapter import TaskScopedMCPCleanupResult, TaskScopedMCPRegistration, TaskScopedMCPServer
from rsi_runner.hermes_runtime import (
    HermesExecutionResult,
    HermesRuntime,
    RunnerTaskRequest,
    _NativeLifecycleTailer,
    canonical_gateway_session_key,
    _redact_json_value,
    _redact_subprocess_output,
)
from rsi_runner.observability import ObservationEmitter, execution_observation_id
from rsi_runner.rsi_tools import rsi_plugin_toolset_definitions, transport_tool_schema
from rsi_runner.session_manager import MemoryTracker, SessionManager


HERMES_TEST_PIN = "0aa2b52d52e6fdaf7992b9d6ac224573f3212f5d"


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


class FakeNativeHonchoAIAgent(FakeAIAgent):
    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self.tools = [
            {"type": "function", "function": {"name": "honcho_profile"}},
            {"type": "function", "function": {"name": "honcho_conclude"}},
        ]
        self.valid_tool_names = {"honcho_profile", "honcho_conclude"}


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
        self.session_db_ref = types.SimpleNamespace(db_path=self.session_db_path)
        self.honcho_available = True
        hermes_home = Path(_config.hermes_home)
        hermes_home.mkdir(parents=True, exist_ok=True)
        hermes_home.joinpath("config.yaml").write_text(
            "plugins:\n  enabled:\n    - rsi_context_engine\n    - company_knowledge\n",
            encoding="utf-8",
        )

    def prepare(self, task: RunnerTaskRequest, *, load_history: bool = True):
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
            session_db_tracking_enabled=load_history,
            conversation_history=[{"role": "user", "content": "Earlier thread message"}] if load_history else [],
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


def write_native_test_envelope(request: dict[str, object], final_response: str = "Final reply") -> None:
    envelope_path = Path(str(request["runtime_envelope_path"]))
    envelope_path.parent.mkdir(parents=True, exist_ok=True)
    envelope_path.write_text(
        json.dumps(
            {
                "contract_version": "execution-envelope/v1",
                "producer": "rsi_platform_runtime",
                "producer_version": "0.1.0",
                "created_at": "2026-05-02T00:00:00Z",
                "facts_source": ["test"],
                "execution_id": str(request.get("execution_id") or ""),
                "operation_id": str(request.get("operation_id") or ""),
                "trace_id": str(request.get("trace_id") or ""),
                "workflow_id": str(request.get("workflow_id") or ""),
                "phase_runs": [
                    {
                        "phase_id": "main",
                        "phase_type": "workflow",
                        "status": "completed",
                        "completion_verdict": "complete",
                        "termination_reason": "normal_completion",
                    }
                ],
                "ledger_events": [],
                "artifacts": [],
                "deliveries": [],
                "completion": {
                    "ok": True,
                    "completion_verdict": "complete",
                    "termination_reason": "normal_completion",
                },
                "final_response": final_response,
            },
            sort_keys=True,
        ),
        encoding="utf-8",
    )


class HermesRuntimeTests(unittest.TestCase):
    def setUp(self) -> None:
        for fake_agent in (FakeAIAgent, FakeNativeHonchoAIAgent):
            fake_agent.last_kwargs = None
            fake_agent.last_prompt = None
            fake_agent.last_system_message = None
            fake_agent.last_history = None
            fake_agent.last_valid_tool_names = None
            fake_agent.last_tool_names = None
            fake_agent.last_interrupt_message = None
            fake_agent.sleep_seconds = 0.0
            fake_agent.budget_used = 1
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
        self.assertEqual(config.hermes_native_toolsets, ["terminal", "file", "company_knowledge"])
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
        self.assertIn("db_read_query", source)
        self.assertIn("external_tool.pending", source)

    def test_github_repo_activity_default_payload_ignores_context_windows(self) -> None:
        namespace: dict[str, object] = {}
        exec(_build_plugin_module(), namespace)

        payload = namespace["_default_payload"](
            "github.repo_activity",
            {
                "task_repo": "rsi-agent-platform",
                "context_refs": [
                    {
                        "kind": "repo_activity_window",
                        "since": "2026-04-30T17:15:39Z",
                        "until": "2026-05-07T17:15:39Z",
                    }
                ],
            },
        )

        self.assertEqual(payload, {"repo": "rsi-agent-platform"})

    def test_task_prompt_advertises_grafana_when_configured(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "story-infra-aws",
                    "prompt": "Read Grafana metrics.",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "RSI_GRAFANA_BASE_URL": "https://grafana.ops.storyprotocol.net",
                "RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN": "grafana-secret-token",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        prompt = runtime._render_task_prompt(task)

        self.assertIn("Grafana read-only observability is available through native `rsi_observability.*` tools", prompt)
        self.assertIn("Grafana Viewer/RBAC is the read boundary", prompt)
        self.assertIn("`rsi_observability.metrics_query`", prompt)
        self.assertIn("`rsi_observability.logs_query`", prompt)
        self.assertIn("`rsi_observability.dashboards_search`/`dashboard_get`", prompt)
        self.assertIn("`rsi_observability.alert_rules_search`/`alert_rule_get`/`active_alerts`", prompt)
        self.assertIn("`rsi_observability.datasources`", prompt)
        self.assertNotIn("rsi-grafana", prompt)
        self.assertNotIn("gcx", prompt)
        self.assertIn("token is mounted in the executor environment and is shell-visible", prompt)
        self.assertIn("Dashboard edits and imports must be PRs to storyprotocol/story-infra-aws", prompt)
        self.assertIn("rsi-observability", runtime._hermes_native_toolsets())



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
                    "execution_id": "hexec-gh-app-force",
                    "trace_id": "trace-gh-app-force",
                    "workflow_id": "wf-gh-app-force",
                    "operation_id": "op-gh-app-force",
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
                write_native_test_envelope(request, final_response="Final reply")
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
            legacy_bin_dir = Path(computer_root, ".rsi", "bin")
            legacy_bin_dir.mkdir(parents=True)
            legacy_bin_dir.joinpath("rsi-db").write_text("#!/bin/sh\nexit 1\n", encoding="utf-8")
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
                    "RSI_GRAFANA_BASE_URL": "https://grafana.ops.storyprotocol.net/",
                    "RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN": "grafana-secret-token",
                    "RSI_GRAFANA_METRICS_DATASOURCE_UID": "thanos",
                    "RSI_GRAFANA_LOGS_DATASOURCE_UID": "loki",
                },
                clear=True,
            ):
                runtime = HermesRuntime(RunnerConfig.from_env())
                captured_env = {
                    "TERMINAL_CWD": os.environ["TERMINAL_CWD"],
                    "PATH": os.environ["PATH"],
                    "KUBECONFIG": os.environ["KUBECONFIG"],
                    "GRAFANA_SERVER": os.environ["GRAFANA_SERVER"],
                    "GRAFANA_TOKEN": os.environ["GRAFANA_TOKEN"],
                    "RSI_GRAFANA_METRICS_DATASOURCE_UID": os.environ["RSI_GRAFANA_METRICS_DATASOURCE_UID"],
                    "RSI_GRAFANA_LOGS_DATASOURCE_UID": os.environ["RSI_GRAFANA_LOGS_DATASOURCE_UID"],
                }

            kubeconfig = kubeconfig_path.read_text(encoding="utf-8")
            manifest = json.loads(Path(computer_root, ".rsi", "computer.json").read_text(encoding="utf-8"))
            bin_dir_entries = sorted(path.name for path in Path(computer_root).resolve().joinpath(".rsi", "bin").iterdir())
            grafana_status = runtime.metadata["company_computer_bootstrap_status"]["grafana_observability"]

        self.assertTrue(runtime.metadata["company_computer_bootstrap_status"]["ok"])
        self.assertEqual(captured_env["TERMINAL_CWD"], str(Path(computer_root).resolve()))
        self.assertTrue(captured_env["PATH"].split(os.pathsep)[0].endswith("/.rsi/bin"))
        self.assertEqual(captured_env["KUBECONFIG"], str(kubeconfig_path.resolve()))
        self.assertEqual(captured_env["GRAFANA_SERVER"], "https://grafana.ops.storyprotocol.net")
        self.assertEqual(captured_env["GRAFANA_TOKEN"], "grafana-secret-token")
        self.assertEqual(captured_env["RSI_GRAFANA_METRICS_DATASOURCE_UID"], "thanos")
        self.assertEqual(captured_env["RSI_GRAFANA_LOGS_DATASOURCE_UID"], "loki")
        self.assertEqual(bin_dir_entries, [])
        self.assertEqual(runtime.metadata["company_computer_bootstrap_status"]["removed_legacy_tools"], ["rsi-db"])
        self.assertEqual(grafana_status["tool"], "rsi_observability")
        self.assertEqual(grafana_status["toolset"], "rsi-observability")
        self.assertTrue(grafana_status["configured"])
        self.assertEqual(grafana_status["transport"], "grafana_datasource_proxy")
        self.assertEqual(grafana_status["policy_boundary"], "grafana_rbac")
        self.assertFalse(grafana_status["query_guardrails_enforced"])
        self.assertEqual(grafana_status["metrics_datasource_uid"], "thanos")
        self.assertEqual(grafana_status["logs_datasource_uid"], "loki")
        self.assertEqual(manifest["grafana_observability"]["tool"], "rsi_observability")
        self.assertEqual(manifest["grafana_observability"]["toolset"], "rsi-observability")
        self.assertTrue(manifest["grafana_observability"]["configured"])
        self.assertEqual(manifest["grafana_observability"]["policy_boundary"], "grafana_rbac")
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
        self.assertEqual(manifest["native_toolsets"], ["terminal", "file", "company_knowledge", "rsi-observability"])
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
        self.assertEqual(result.raw["failure_class"], "native_workflow_preflight_failed")

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
        self.assertEqual(result.raw["failure_class"], "native_workflow_preflight_failed")

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
        self.assertIn("Hermes config is missing required plugin(s): rsi_context_engine, company_knowledge.", status.errors)

    def test_hermes_contract_requires_platform_runtime_capability_for_executor(self) -> None:
        plugin_manager = types.SimpleNamespace(
            list_plugins=lambda: [
                {"enabled": True, "name": "rsi_context_engine"},
                {"enabled": True, "name": "company_knowledge"},
                {"enabled": True, "name": "rsi_platform_runtime", "capabilities": {}},
            ]
        )
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
            Path(tempdir, "config.yaml").write_text(
                "plugins:\n  enabled:\n    - rsi_context_engine\n    - company_knowledge\n    - rsi_platform_runtime\n",
                encoding="utf-8",
            )
            status = validate_hermes_contract(
                expected_pin=HERMES_TEST_PIN,
                hermes_home=tempdir,
                session_db=object(),
                required_toolsets=[],
                require_platform_runtime=True,
            )

        self.assertFalse(status.ok)
        self.assertEqual(status.plugin_status, "capability_missing")
        self.assertIn("Hermes plugin capability check failed for: rsi_platform_runtime.", status.errors)

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

    def test_hermes_contract_defers_parent_runtime_session_db_integrity_check(self) -> None:
        plugin_manager = types.SimpleNamespace(
            list_plugins=lambda: [
                {"enabled": True, "name": "rsi_context_engine"},
                {"enabled": True, "name": "company_knowledge"},
                {
                    "enabled": True,
                    "name": "rsi_platform_runtime",
                    "capabilities": {"execution_scoped_context_supported": True},
                },
            ]
        )
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
            Path(tempdir, "config.yaml").write_text(
                "plugins:\n  enabled:\n    - rsi_context_engine\n    - company_knowledge\n    - rsi_platform_runtime\n",
                encoding="utf-8",
            )
            Path(tempdir, "state.db").write_bytes(b"not sqlite")
            with mock.patch("rsi_runner.hermes_agent_adapter.sqlite3.connect") as connect_mock:
                status = validate_hermes_contract(
                    expected_pin=HERMES_TEST_PIN,
                    hermes_home=tempdir,
                    session_db=types.SimpleNamespace(db_path=Path(tempdir, "state.db"), defer_integrity_check=True),
                    required_toolsets=[],
                    require_platform_runtime=True,
                    require_session_db_ready=False,
                )

        self.assertTrue(status.ok, status.errors)
        self.assertEqual(status.session_db_status, "deferred")
        connect_mock.assert_not_called()

    def test_hermes_contract_retries_locked_session_db_integrity_check(self) -> None:
        class FakeCursorResult:
            def __init__(self, rows: list[tuple[str]]) -> None:
                self._rows = rows

            def fetchone(self) -> tuple[str] | None:
                return self._rows[0] if self._rows else None

            def fetchall(self) -> list[tuple[str]]:
                return list(self._rows)

        class FakeConnection:
            def __enter__(self):
                return self

            def __exit__(self, _exc_type, _exc, _tb) -> bool:
                return False

            def execute(self, query: str):
                if "quick_check" in query:
                    return FakeCursorResult([("ok",)])
                return FakeCursorResult([("sessions",), ("messages",)])

        with tempfile.TemporaryDirectory() as tempdir:
            db_path = Path(tempdir, "state.db")
            db_path.touch()
            with mock.patch(
                "rsi_runner.hermes_agent_adapter.sqlite3.connect",
                side_effect=[sqlite3.OperationalError("database is locked"), FakeConnection()],
            ), mock.patch("rsi_runner.hermes_agent_adapter.time.sleep") as sleep_mock, mock.patch(
                "rsi_runner.hermes_agent_adapter.random.uniform", return_value=0
            ), mock.patch.dict(
                os.environ,
                {"RSI_HERMES_SESSION_DB_INTEGRITY_RETRY_SECONDS": "1"},
                clear=True,
            ):
                status, error = _validate_session_db_integrity(types.SimpleNamespace(db_path=db_path))

        self.assertEqual(status, "ok")
        self.assertEqual(error, "")
        sleep_mock.assert_called_once()

    def test_hermes_contract_reports_locked_session_db_integrity_without_corruption(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            db_path = Path(tempdir, "state.db")
            db_path.touch()
            with mock.patch(
                "rsi_runner.hermes_agent_adapter.sqlite3.connect",
                side_effect=sqlite3.OperationalError("database is locked"),
            ), mock.patch(
                "rsi_runner.hermes_agent_adapter.time.monotonic",
                side_effect=[0.0, 2.0],
            ), mock.patch.dict(
                os.environ,
                {"RSI_HERMES_SESSION_DB_INTEGRITY_RETRY_SECONDS": "1"},
                clear=True,
            ):
                status, error = _validate_session_db_integrity(types.SimpleNamespace(db_path=db_path))

        self.assertEqual(status, "locked")
        self.assertEqual(error, "")

    def test_hermes_contract_tolerates_uninitialized_session_db_for_parent_runtime(self) -> None:
        plugin_manager = types.SimpleNamespace(
            list_plugins=lambda: [
                {"enabled": True, "name": "rsi_context_engine"},
                {"enabled": True, "name": "company_knowledge"},
                {
                    "enabled": True,
                    "name": "rsi_platform_runtime",
                    "capabilities": {"execution_scoped_context_supported": True},
                },
            ]
        )
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
            Path(tempdir, "config.yaml").write_text(
                "plugins:\n  enabled:\n    - rsi_context_engine\n    - company_knowledge\n    - rsi_platform_runtime\n",
                encoding="utf-8",
            )
            status = validate_hermes_contract(
                expected_pin=HERMES_TEST_PIN,
                hermes_home=tempdir,
                session_db=types.SimpleNamespace(db_path=Path(tempdir, "state.db")),
                required_toolsets=[],
                require_platform_runtime=True,
                require_session_db_ready=False,
            )

        self.assertTrue(status.ok, status.errors)
        self.assertEqual(status.session_db_status, "uninitialized")

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

    def test_hermes_agent_adapter_preserves_native_timeout_meta(self) -> None:
        adapter = HermesAgentAdapter({"session_id": "session-1", "max_iterations": 20})

        meta = adapter._completion_meta(
            {
                "completed": False,
                "partial": True,
                "api_calls": 3,
                "final_response": "partial reply",
                "termination_reason": "task_timeout",
                "completion_verdict": "partial",
                "timeout_kind": "task_timeout",
            }
        )

        self.assertEqual(meta["termination_reason"], "task_timeout")
        self.assertEqual(meta["completion_verdict"], "partial")
        self.assertEqual(meta["timeout_kind"], "task_timeout")
        self.assertFalse(meta["max_iterations_reached"])
        self.assertTrue(meta["native_result_partial"])

    def test_json_object_from_string_accepts_json_code_fences(self) -> None:
        payload = {"final_answer": "Clean answer", "reply_delivery": {"body": "Slack body"}}

        self.assertEqual(_json_object_from_string("```json\n" + json.dumps(payload) + "\n```"), payload)
        self.assertEqual(_json_object_from_string("draft:\n```JSON\n" + json.dumps(payload) + "\n```"), payload)

    def test_hermes_agent_adapter_final_delivery_body_extracts_fenced_contract(self) -> None:
        adapter = HermesAgentAdapter({"session_id": "session-1"})
        response = "```json\n" + json.dumps({"final_answer": "Clean answer"}) + "\n```"

        self.assertEqual(adapter._final_delivery_body(response, {"final_response": response}), "Clean answer")

    def test_hermes_agent_adapter_final_delivery_body_uses_nested_reply_delivery_body(self) -> None:
        adapter = HermesAgentAdapter({"session_id": "session-1"})
        response = "```json\n" + json.dumps({"reply_delivery": {"body": "Slack body"}}) + "\n```"

        self.assertEqual(adapter._final_delivery_body(response, {"final_response": response}), "Slack body")

    def test_hermes_agent_adapter_final_delivery_body_drops_malformed_json_contract_fallback(self) -> None:
        adapter = HermesAgentAdapter({"session_id": "session-1"})
        malformed_response = '```json\n{"visible_reasoning": "done", "final_answer": "Raw contract"\n```'

        self.assertEqual(
            adapter._final_delivery_body(
                malformed_response,
                {"final_response": malformed_response, "structured_output": {"reply_draft": "Recovered draft"}},
            ),
            "Recovered draft",
        )
        self.assertEqual(adapter._final_delivery_body(malformed_response, {"final_response": malformed_response}), "")

    def test_hermes_agent_adapter_writes_self_review_candidate_with_runner_execution_id(self) -> None:
        captured_payload: dict[str, object] = {}

        def fake_run(cmd, check, capture_output, text, timeout):
            captured_payload.update(json.loads(Path(cmd[-1]).read_text(encoding="utf-8")))
            observation = captured_payload["observation"]
            assert isinstance(observation, dict)
            return subprocess.CompletedProcess(
                cmd,
                0,
                stdout=json.dumps(
                    {
                        "candidate_id": 7,
                        "candidate_status": "candidate",
                        "status": "candidate",
                        "execution_id": observation.get("execution_id"),
                        "agent_identity": "rsi:stage:prod",
                        "snapshot_ref": "snapshots/7.json",
                        "snapshot_hash": "abc123",
                        "snapshot_size": 123,
                    }
                )
                + "\n",
                stderr="",
            )

        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True), mock.patch(
            "rsi_runner.hermes_agent_adapter.subprocess.run",
            side_effect=fake_run,
        ):
            adapter = HermesAgentAdapter(
                {
                    "session_id": "rsi-prod-conversation-123",
                    "execution_id": "hexec-runner-123",
                }
            )
            candidate = adapter._write_self_review_candidate(
                {
                    "self_review_observation": {
                        "execution_id": "rsi-prod-conversation-123",
                        "agent_identity": "rsi:stage:prod",
                    }
                }
            )

        observation = captured_payload["observation"]
        self.assertIsInstance(observation, dict)
        self.assertEqual(observation["execution_id"], "hexec-runner-123")
        self.assertEqual(candidate["execution_id"], "hexec-runner-123")

    def test_hermes_agent_adapter_commits_direct_delivery_from_final_response(self) -> None:
        calls: list[tuple[dict[str, object], dict[str, object]]] = []
        send_module = types.ModuleType("tools.send_message_tool")

        def fake_send_message(args, **kwargs):
            calls.append((dict(args), dict(kwargs)))
            return json.dumps(
                {
                    "success": True,
                    "platform": "slack",
                    "chat_id": "C123",
                    "thread_id": "171000001.000100",
                    "message_id": "171000001.000200",
                    "message_link": "https://slack.example/archives/C123/p171000001000200",
                    "idempotency_key": args["idempotency_key"],
                }
            )

        send_module.send_message_tool = fake_send_message
        tools_module = types.ModuleType("tools")
        tools_module.__path__ = []

        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(
            sys.modules,
            {
                "tools": tools_module,
                "tools.send_message_tool": send_module,
            },
        ), mock.patch.dict(os.environ, {"HERMES_HOME": tempdir}, clear=True):
            envelope_path = Path(tempdir, "envelope.json")
            write_native_test_envelope(
                {
                    "runtime_envelope_path": str(envelope_path),
                    "execution_id": "hexec-direct",
                    "operation_id": "op-direct",
                    "trace_id": "trace-direct",
                    "workflow_id": "wf-direct",
                },
                final_response="😺",
            )
            adapter = HermesAgentAdapter(
                {
                    "session_id": "session-direct",
                    "execution_id": "hexec-direct",
                    "trace_id": "trace-direct",
                    "runtime_envelope_path": str(envelope_path),
                    "delivery_policy": {
                        "bound_channel_id": "C123",
                        "bound_thread_ts": "171000001.000100",
                        "direct_send_allowed": True,
                        "idempotency_key_base": "C123:171000001.000100:trace-direct",
                    },
                }
            )

            delivery = adapter._commit_final_delivery_if_required("😺", {"final_response": "😺"})
            envelope = json.loads(envelope_path.read_text(encoding="utf-8"))

        self.assertEqual(len(calls), 1)
        args, kwargs = calls[0]
        self.assertEqual(args["target"], "slack:C123:171000001.000100")
        self.assertEqual(args["message"], "😺")
        self.assertEqual(args["idempotency_key"], "C123:171000001.000100:trace-direct")
        self.assertEqual(kwargs["idempotency_key"], "C123:171000001.000100:trace-direct")
        self.assertEqual(delivery["send_status"], "posted")
        self.assertEqual(delivery["body"], "😺")
        self.assertEqual(envelope["deliveries"][0]["body"], "😺")
        self.assertEqual(envelope["deliveries"][0]["send_status"], "posted")
        self.assertIn("rsi_runner.hermes_agent_adapter.delivery_commit", envelope["facts_source"])
        self.assertIn("slack.message.sent", [item["kind"] for item in envelope["ledger_events"]])
        self.assertEqual(envelope["phase_runs"][-1]["phase_id"], "deliver")
        self.assertEqual(envelope["phase_runs"][-1]["status"], "completed")

    def test_gateway_session_key_slack_thread_root_and_fallback(self) -> None:
        threaded = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv:one",
                }
            }
        )
        same_thread = RunnerTaskRequest.from_payload({"task": {"channel_id": "C123", "thread_ts": "171000001.000100"}})
        other_thread = RunnerTaskRequest.from_payload({"task": {"channel_id": "C123", "thread_ts": "171000002.000200"}})
        root = RunnerTaskRequest.from_payload({"task": {"channel_id": "C123", "message_ts": "171000003.000300"}})
        derived_root = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "channel_id": "C123",
                    "recent_conversation_entries": [
                        {"channel_id": "C123", "message_ts": "171000004.000400"},
                    ],
                }
            }
        )
        fallback = RunnerTaskRequest.from_payload(
            {"task": {"session_scope_kind": "Conversation", "session_scope_id": "conv:with/slash%20encoded"}}
        )

        self.assertEqual(canonical_gateway_session_key(threaded, "Prod"), canonical_gateway_session_key(same_thread, "prod"))
        self.assertNotEqual(canonical_gateway_session_key(threaded, "prod"), canonical_gateway_session_key(other_thread, "prod"))
        self.assertEqual(canonical_gateway_session_key(root, "prod"), "rsi:prod:slack:C123:171000003.000300")
        self.assertEqual(canonical_gateway_session_key(derived_root, "prod"), "rsi:prod:slack:C123:171000004.000400")
        self.assertEqual(
            canonical_gateway_session_key(fallback, "Prod"),
            "rsi:prod:scope:conversation:conv%3Awith%2Fslash%2520encoded",
        )

    def test_native_runtime_env_exposes_db_read_request_scope(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": tempdir},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "execution_id": "hexec-1",
                        "operation_id": "op-1",
                        "trace_id": "trace-1",
                        "workflow_id": "wf-1",
                        "conversation_id": "conv-1",
                        "channel_id": "C123",
                        "thread_ts": "171000001.000100",
                        "user_peer_id": "user:U123",
                    }
                }
            )
            env = runtime._native_runtime_env(
                task,
                context_path=Path(tempdir, "context.json"),
                envelope_path=Path(tempdir, "envelope.json"),
            )

        self.assertEqual(env["RSI_WORKFLOW_ID"], "wf-1")
        self.assertEqual(env["RSI_CONVERSATION_ID"], "conv-1")
        self.assertEqual(env["RSI_SLACK_CHANNEL_ID"], "C123")
        self.assertEqual(env["RSI_SLACK_THREAD_TS"], "171000001.000100")
        self.assertEqual(env["RSI_TASK_REQUESTER"], "user:U123")
        self.assertNotIn("RSI_DB_READ_EXECUTION_TOKEN", env)
        self.assertNotIn("RSI_DB_READ_CLIENT_TOKEN", env)

    def test_native_runtime_env_injects_execution_scoped_db_read_token(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": tempdir,
                "RSI_DB_READ_ENABLED": "true",
                "RSI_CONTROL_PLANE_BASE_URL": "http://control-plane.rsi-platform.svc.cluster.local:8080",
                "RSI_DB_READ_CLIENT_TOKEN": "static-secret",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "execution_id": "hexec-1",
                        "operation_id": "op-1",
                        "trace_id": "trace-1",
                        "workflow_id": "wf-1",
                        "conversation_id": "conv-1",
                        "channel_id": "C123",
                        "thread_ts": "171000001.000100",
                        "user_peer_id": "user:U123",
                    }
                }
            )
            env = runtime._native_runtime_env(
                task,
                context_path=Path(tempdir, "context.json"),
                envelope_path=Path(tempdir, "envelope.json"),
            )

        token = env.get("RSI_DB_READ_EXECUTION_TOKEN", "")
        self.assertTrue(token.startswith("v1."))
        self.assertEqual(env["RSI_DB_READ_AUTH_MODE"], "execution_scoped")
        self.assertNotIn("RSI_DB_READ_CLIENT_TOKEN", env)
        parts = token.split(".")
        claims = json.loads(base64.urlsafe_b64decode(parts[1] + "=" * (-len(parts[1]) % 4)).decode("utf-8"))
        self.assertEqual(claims["execution_id"], "hexec-1")
        self.assertEqual(claims["operation_id"], "op-1")
        self.assertEqual(claims["conversation_id"], "conv-1")
        self.assertEqual(claims["workflow_id"], "wf-1")
        self.assertEqual(claims["trace_id"], "trace-1")
        self.assertEqual(claims["channel_id"], "C123")
        self.assertEqual(claims["thread_ts"], "171000001.000100")
        self.assertEqual(claims["requester"], "user:U123")
        self.assertTrue(claims["db_read_query_allowed"])

    def test_hermes_agent_adapter_passes_rich_context_to_aiagent(self) -> None:
        captured: dict[str, object] = {}

        class CapturingAIAgent:
            def __init__(self, **kwargs):
                captured.update(kwargs)

        with mock.patch("rsi_runner.hermes_agent_adapter.AIAgent", CapturingAIAgent):
            adapter = HermesAgentAdapter(
                {
                    "session_id": "session-1",
                    "parent_session_id": "parent-1",
                    "model": "deepseek/deepseek-v4-pro",
                    "max_iterations": 5,
                    "toolsets": ["memory", "skills"],
                    "runtime": {"provider": "openrouter", "base_url": "https://openrouter.ai/api/v1"},
                    "user_peer_id": "user:U123",
                    "chat_id": "C123",
                    "thread_id": "171000001.000100",
                    "gateway_session_key": "rsi:prod:slack:C123:171000001.000100",
                }
            )
            adapter._create_agent("session-1", object(), "system")

        self.assertEqual(captured["user_id"], "user:U123")
        self.assertEqual(captured["chat_id"], "C123")
        self.assertEqual(captured["thread_id"], "171000001.000100")
        self.assertEqual(captured["gateway_session_key"], "rsi:prod:slack:C123:171000001.000100")

    def test_hermes_agent_adapter_ensures_missing_parent_session_before_child(self) -> None:
        try:
            from hermes_state import SessionDB as HermesSessionDB  # type: ignore
        except (ImportError, ModuleNotFoundError):
            self.skipTest("hermes_state is unavailable")

        with tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(
            os.environ,
            {"HERMES_HOME": tempdir},
            clear=False,
        ):
            db = HermesSessionDB(db_path=Path(tempdir, "state.db"))
            adapter = HermesAgentAdapter(
                {
                    "session_id": "child-session",
                    "parent_session_id": "parent-session",
                    "model": "deepseek/deepseek-v4-pro",
                }
            )

            adapter._prepare_session_db("child-session", db, "system prompt")

            parent = db.get_session("parent-session")
            child = db.get_session("child-session")

        self.assertIsNotNone(parent)
        self.assertIsNotNone(child)
        self.assertEqual(parent["source"], "rsi_executor_parent")
        self.assertEqual(child["parent_session_id"], "parent-session")
        self.assertEqual(child["system_prompt"], "system prompt")

    def test_hermes_agent_adapter_treats_concurrent_parent_insert_as_success(self) -> None:
        class RacingParentDB:
            def __init__(self) -> None:
                self.parent_exists = False
                self.created_children: list[tuple[str, str | None]] = []

            def get_session(self, session_id: str) -> dict[str, str] | None:
                if session_id == "parent-session" and self.parent_exists:
                    return {"id": session_id}
                return None

            def create_session(self, session_id: str, source: str, **kwargs) -> str:
                if session_id == "parent-session":
                    self.parent_exists = True
                    raise sqlite3.IntegrityError("UNIQUE constraint failed: sessions.id")
                self.created_children.append((session_id, kwargs.get("parent_session_id")))
                return session_id

            def reopen_session(self, session_id: str) -> None:
                return None

            def update_system_prompt(self, session_id: str, system_prompt: str) -> None:
                return None

        db = RacingParentDB()
        adapter = HermesAgentAdapter(
            {
                "session_id": "child-session",
                "parent_session_id": "parent-session",
                "model": "deepseek/deepseek-v4-pro",
            }
        )

        adapter._prepare_session_db("child-session", db, "system prompt")

        self.assertEqual(db.created_children, [("child-session", "parent-session")])

    def test_hermes_agent_adapter_retries_locked_session_db_open(self) -> None:
        attempts = 0

        def fake_session_db(db_path: Path):
            nonlocal attempts
            attempts += 1
            if attempts == 1:
                raise sqlite3.OperationalError("database is locked")
            return types.SimpleNamespace(db_path=db_path)

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_agent_adapter.SessionDB",
            side_effect=fake_session_db,
        ), mock.patch("rsi_runner.hermes_agent_adapter.time.sleep") as sleep_mock, mock.patch(
            "rsi_runner.hermes_agent_adapter.random.uniform", return_value=0
        ), mock.patch.dict(
            os.environ,
            {"HERMES_HOME": tempdir, "RSI_HERMES_SESSION_DB_OPEN_RETRY_SECONDS": "1"},
            clear=True,
        ):
            adapter = HermesAgentAdapter({"session_id": "session-1"})
            db = adapter._open_session_db()

        self.assertEqual(attempts, 2)
        self.assertEqual(Path(db.db_path), Path(tempdir, "state.db"))
        sleep_mock.assert_called_once()

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

    def test_session_manager_does_not_open_session_db_during_startup(self) -> None:
        class ExplodingSessionDB:
            def __init__(self, db_path: str) -> None:
                raise AssertionError(f"SessionDB should not open during runner startup: {db_path}")

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB", ExplodingSessionDB
        ), mock.patch("rsi_runner.session_manager.sync_skills") as sync_mock, mock.patch.dict(
            os.environ, {**runner_env("eval"), "HERMES_HOME": tempdir}, clear=True
        ):
            sync_mock.return_value = {"copied": []}
            manager = SessionManager(RunnerConfig.from_env())

        self.assertTrue(manager.available)
        self.assertEqual(manager.session_db_path, os.path.join(tempdir, "state.db"))

    def test_session_manager_retries_locked_session_db_when_loading_history(self) -> None:
        attempts = 0

        class FlakySessionDB:
            def __init__(self, db_path: str) -> None:
                nonlocal attempts
                attempts += 1
                self.db_path = db_path
                if attempts == 1:
                    raise sqlite3.OperationalError("database is locked")

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        task = types.SimpleNamespace(
            session_scope_kind="conversation",
            session_scope_id="conv-1",
            parent_session_scope_kind="",
            parent_session_scope_id="",
            user_peer_id="user:U123",
            recent_conversation_entries=[],
            assistant_peer_id="rsi:test:prod",
        )
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB",
            FlakySessionDB,
        ), mock.patch("rsi_runner.session_manager.time.sleep") as sleep_mock, mock.patch(
            "rsi_runner.session_manager.random.uniform", return_value=0
        ), mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": tempdir, "RSI_HERMES_SESSION_DB_OPEN_RETRY_SECONDS": "1"},
            clear=True,
        ):
            manager = SessionManager(RunnerConfig.from_env())
            context = manager.prepare(task, load_history=True)

        self.assertEqual(attempts, 2)
        self.assertTrue(context.session_id.startswith("rsi-prod-conversation-"))
        sleep_mock.assert_called_once()

    def test_session_manager_retries_locked_session_history_read(self) -> None:
        read_attempts = 0

        class FlakyReadSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                nonlocal read_attempts
                read_attempts += 1
                if read_attempts == 1:
                    raise sqlite3.OperationalError("database is locked")
                return [{"role": "user", "content": "prior"}]

        task = types.SimpleNamespace(
            session_scope_kind="conversation",
            session_scope_id="conv-1",
            parent_session_scope_kind="",
            parent_session_scope_id="",
            user_peer_id="user:U123",
            recent_conversation_entries=[],
            assistant_peer_id="rsi:test:prod",
        )
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB",
            FlakyReadSessionDB,
        ), mock.patch("rsi_runner.session_manager.time.sleep") as sleep_mock, mock.patch(
            "rsi_runner.session_manager.random.uniform", return_value=0
        ), mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": tempdir, "RSI_HERMES_SESSION_DB_READ_RETRY_SECONDS": "1"},
            clear=True,
        ):
            manager = SessionManager(RunnerConfig.from_env())
            context = manager.prepare(task, load_history=True)

        self.assertEqual(read_attempts, 2)
        self.assertEqual(context.conversation_history, [{"role": "user", "content": "prior"}])
        sleep_mock.assert_called_once()

    def test_session_manager_finalize_records_warning_when_history_read_stays_locked(self) -> None:
        class LockedReadSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                raise sqlite3.OperationalError("database is locked")

        context = types.SimpleNamespace(
            session_id="session-1",
            parent_session_id="",
            scope_kind="conversation",
            scope_id="conv-1",
            parent_scope_kind="",
            parent_scope_id="",
            memory_backend="honcho",
            assistant_peer_id="rsi:test:prod",
            user_peer_id="user:U123",
            hermes_home="/tmp/hermes",
            session_db_path="/tmp/hermes/state.db",
            session_db_tracking_enabled=True,
            conversation_history=[{"role": "user", "content": "prior"}],
        )
        tracker = MemoryTracker()
        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB",
            LockedReadSessionDB,
        ), mock.patch(
            "rsi_runner.session_manager.time.monotonic",
            side_effect=[0.0, 0.0, 2.0],
        ), mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": tempdir, "RSI_HERMES_SESSION_DB_READ_RETRY_SECONDS": "1"},
            clear=True,
        ):
            manager = SessionManager(RunnerConfig.from_env())
            payload = manager.finalize(context, tracker)

        self.assertEqual(payload["session_messages_delta"], [])
        self.assertEqual(tracker.warnings[0]["kind"], "session_history_read_failed")

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
        self.assertIn("    - company_knowledge", config_text)
        self.assertNotIn("    - rsi_platform_runtime", config_text)
        self.assertEqual(manager.hermes_config_parity_status, "configured")
        self.assertEqual(manager.hermes_config_parity_error, "")

    def test_session_manager_enables_platform_runtime_for_hermes_executor(self) -> None:
        class FakeSessionDB:
            def __init__(self, db_path: str) -> None:
                self.db_path = db_path

            def get_messages_as_conversation(self, _session_id: str) -> list[dict[str, object]]:
                return []

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.session_manager.SessionDB", FakeSessionDB
        ), mock.patch("rsi_runner.session_manager.sync_skills") as sync_mock, mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": tempdir, "RSI_HERMES_EXECUTOR_ENABLED": "true"},
            clear=True,
        ):
            sync_mock.return_value = {"copied": []}
            manager = SessionManager(RunnerConfig.from_env())
            config_text = Path(tempdir, "config.yaml").read_text(encoding="utf-8")

        self.assertIn("    - rsi_context_engine", config_text)
        self.assertIn("    - company_knowledge", config_text)
        self.assertIn("    - rsi_platform_runtime", config_text)
        self.assertEqual(manager.hermes_config_parity_status, "configured")

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
                    "allowed_commands": "python -m pytest",
                    "expected_outputs": "final answer",
                    "rejected_proposal_context": "nope",
                    "recent_conversation_entries": "nope",
                    "prior_trace_refs": "nope",
                    "repo_allowlist": "nope",
                    "context_refs": "nope",
                    "allowed_path_globs": "nope",
                    "case_summary": "nope",
                }
            }
        )

        self.assertEqual(task.allowed_commands, [])
        self.assertEqual(task.expected_outputs, [])
        self.assertEqual(task.rejected_proposal_context, [])
        self.assertEqual(task.recent_conversation_entries, [])
        self.assertEqual(task.prior_trace_refs, [])
        self.assertEqual(task.repo_allowlist, [])
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

    def test_runner_task_request_ignores_legacy_requested_artifacts(self) -> None:
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

        self.assertFalse(hasattr(task, "requested_artifacts"))
        self.assertFalse(hasattr(task, "artifact_optional"))

    def test_runner_task_request_ignores_legacy_requested_skills(self) -> None:
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

        self.assertFalse(hasattr(task, "requested_skills"))

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

    def test_active_execution_snapshot_and_drain_wait_track_background_threads(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        done = threading.Event()

        def run() -> None:
            time.sleep(0.05)
            done.set()

        thread = threading.Thread(target=run)
        with runtime._executor_process_lock:
            runtime._executor_threads["hexec-active"] = thread
        thread.start()

        snapshot = runtime.active_execution_snapshot()
        self.assertEqual(snapshot["active_execution_count"], 1)
        self.assertEqual(snapshot["active_execution_ids"], ["hexec-active"])

        drained = runtime.wait_for_active_executions(2)
        self.assertTrue(done.is_set())
        self.assertEqual(drained["drain_status"], "drained")
        self.assertEqual(drained["active_execution_count"], 0)

    def test_active_execution_snapshot_tracks_native_process_registry_checkpoint(self) -> None:
        with tempfile.TemporaryDirectory() as hermes_home, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

            process = subprocess.Popen(
                [sys.executable, "-c", "import time; time.sleep(0.5)"],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
            )
            try:
                session_key = "rsi:prod:slack:C123:1700000000.000000"
                with runtime._executor_process_lock:
                    runtime._executor_native_session_ids["hexec-native-bg"] = "session-native-bg"
                    runtime._executor_native_session_keys["hexec-native-bg"] = session_key
                    runtime._executor_native_started_at_unix["hexec-native-bg"] = time.time()
                Path(hermes_home, "processes.json").write_text(
                    json.dumps(
                        [
                            {
                                "session_id": "proc-test",
                                "command": "python long-native-eval.py",
                                "pid": process.pid,
                                "pid_scope": "host",
                                "task_id": "default",
                                "session_key": session_key,
                                "started_at": time.time(),
                            }
                        ]
                    ),
                    encoding="utf-8",
                )

                snapshot = runtime.active_execution_snapshot(include_self_review_queue=False)
                self.assertEqual(snapshot["active_execution_count"], 1)
                self.assertEqual(snapshot["active_execution_ids"], ["hexec-native-bg"])
                self.assertEqual(snapshot["active_native_subprocess_execution_ids"], ["hexec-native-bg"])
                self.assertEqual(snapshot["active_native_subprocess_count"], 1)
                self.assertEqual(snapshot["active_native_subprocesses"]["hexec-native-bg"][0]["pid"], process.pid)
            finally:
                process.terminate()
                try:
                    process.wait(timeout=2)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait(timeout=2)

            drained = runtime.wait_for_active_executions(2)
            self.assertEqual(drained["drain_status"], "drained")
            self.assertEqual(drained["active_execution_count"], 0)

    def test_active_execution_snapshot_keeps_native_descendants_after_worker_exits(self) -> None:
        with tempfile.TemporaryDirectory() as hermes_home, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_EXECUTOR_ENABLED": "true",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        class FakeProcess:
            pid = 12345

            def poll(self) -> None:
                return None

        with runtime._executor_process_lock:
            runtime._executor_processes["hexec-native-descendant"] = FakeProcess()  # type: ignore[assignment]

        with mock.patch.object(runtime, "_descendant_process_ids", return_value={23456}), mock.patch.object(
            runtime,
            "_pid_is_active",
            side_effect=lambda pid: pid == 23456,
        ):
            snapshot = runtime.active_execution_snapshot(include_self_review_queue=False)
            self.assertEqual(snapshot["active_execution_count"], 1)
            self.assertEqual(snapshot["active_native_subprocess_count"], 1)
            self.assertEqual(snapshot["active_native_subprocess_execution_ids"], ["hexec-native-descendant"])

            with runtime._executor_process_lock:
                runtime._executor_processes.pop("hexec-native-descendant", None)

            snapshot = runtime.active_execution_snapshot(include_self_review_queue=False)
            self.assertEqual(snapshot["active_execution_count"], 1)
            self.assertEqual(snapshot["active_execution_ids"], ["hexec-native-descendant"])

    def test_executor_status_orphaned_diagnostics_include_pod_and_status_file(self) -> None:
        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as workspace_root, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HOSTNAME": "executor-pod-1",
                "HERMES_HOME": hermes_home,
                "RSI_HERMES_COMPUTER_ROOT": str(Path(workspace_root) / "company"),
                "RSI_HERMES_RUN_ROOT": str(Path(workspace_root) / "company" / ".rsi" / "runs"),
                "RSI_HERMES_ARTIFACT_ROOT": str(Path(workspace_root) / "company" / "artifacts"),
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            runtime._store_executor_result("hexec-orphan", {"execution_id": "hexec-orphan", "status": "running"})
            status = runtime.executor_status("hexec-orphan")

        self.assertEqual(status["status"], "orphaned")
        self.assertEqual(status["executor_instance_id"], "executor-pod-1")
        self.assertEqual(status["current_executor_instance_id"], "executor-pod-1")
        self.assertEqual(status["last_observed_status"], "running")
        self.assertTrue(str(status["status_file_path"]).endswith("hexec-orphan.json"))

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

    def test_workflow_provider_runtime_error_is_not_reported_as_structured_output_parse_failure(self) -> None:
        class ProviderRuntimeErrorAIAgent(FakeAIAgent):
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
                    "final_response": "API call failed after 3 retries: HTTP 403: Key limit exceeded (monthly limit)."
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
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", ProviderRuntimeErrorAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_provider_error")
        self.assertEqual(result.raw["runner_diagnostics"]["failure_kind"], "provider_runtime_error")
        self.assertEqual(result.raw["runner_diagnostics"]["provider_status_code"], 403)
        self.assertIn("Key limit exceeded", result.message)
        self.assertNotIn("structured output", result.message)
        self.assertFalse(result.raw["repair_attempted"])
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
                write_native_test_envelope(request, final_response="Final reply")
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
            self.assertNotIn("requested_skills", process._request)
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
        ), mock.patch.object(
            HermesCompanyComputer,
            "attach_envelope",
            side_effect=AssertionError("native strict success must not synthesize an envelope"),
        ) as attach_mock, mock.patch.dict(
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
                            "execution_id": "hexec-123",
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
        self.assertEqual(result.message, "Final reply")
        self.assertEqual(result.raw["execution_envelope"]["final_response"], "Final reply")
        self.assertNotIn("structured_output", result.raw)
        self.assertEqual(result.raw["system_message"], "System directive")
        self.assertEqual(result.raw["runner_diagnostics"]["agentic_mcp_cleanup_status"], "cleaned")
        attach_mock.assert_not_called()
        self.assertEqual(len(request_paths), 1)
        self.assertTrue(run_dir_exists_before_temp_cleanup)

    def test_execute_task_skips_attach_envelope_for_native_strict_preflight_failure(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Run without required workflow identifiers.",
                    "trace_id": "trace-preflight-missing-id",
                    "workflow_id": "wf-preflight-missing-id",
                    "operation_id": "op-preflight-missing-id",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-preflight-missing-id",
                }
            }
        )
        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen",
            side_effect=AssertionError("Hermes subprocess must not start after native strict preflight failure"),
        ) as popen_mock, mock.patch.object(
            HermesCompanyComputer,
            "attach_envelope",
            side_effect=AssertionError("native strict failure must not synthesize an envelope"),
        ) as attach_mock, mock.patch.dict(
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
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.provider, "hermes-native-executor")
        self.assertEqual(result.raw["failure_class"], "native_workflow_preflight_failed")
        self.assertTrue(result.raw["native_strict"])
        self.assertNotIn("execution_envelope", result.raw)
        attach_mock.assert_not_called()
        popen_mock.assert_not_called()

    def test_native_executor_fails_before_subprocess_when_github_app_credentials_missing(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "Read the private repo.",
                    "execution_id": "hexec-gh-required",
                    "trace_id": "trace-gh-required",
                    "workflow_id": "wf-gh-required",
                    "operation_id": "op-gh-required",
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
        self.assertEqual(result.raw["failure_class"], "native_workflow_preflight_failed")
        self.assertEqual(result.raw["termination_reason"], "native_workflow_preflight_failed")
        self.assertEqual(result.raw["runner_diagnostics"]["preflight_failure_class"], "github_app_credentials_unavailable")
        self.assertEqual(result.raw["runner_diagnostics"]["github_credentials"]["reason"], "missing_github_app_credentials")
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
                write_native_test_envelope(request, final_response="Grounded but partial answer.")
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
                    "execution_id": "hexec-partial-native",
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
        self.assertEqual(result.message, "Grounded but partial answer.")
        self.assertEqual(result.raw["execution_envelope"]["final_response"], "Grounded but partial answer.")
        self.assertNotIn("structured_output", result.raw)
        self.assertEqual(result.raw["runner_diagnostics"]["completion_verdict"], "partial")
        self.assertEqual(captured_requests[0]["phase_contract"]["history_policy"], "session")
        self.assertNotIn("rsi-governed-readonly", captured_requests[0]["phase_contract"]["required_toolsets"])

    def test_native_executor_extracts_trailing_json_fence_without_repair_mcp_registration(self) -> None:
        structured = {
            "visible_reasoning": [],
            "reply_draft": "I found the architecture details.",
            "final_answer": "Architecture details are ready.",
            "confidence": 0.91,
            "context_summary": "Grounded context.",
            "self_critique": "",
            "proposed_actions": [],
            "knowledge_drafts": [],
            "outcome_hypotheses": [],
            "artifact_render_briefs": [
                {
                    "kind": "diagram",
                    "title": "RSI Platform",
                    "render_prompt": "Render the grounded architecture.",
                }
            ],
            "produced_artifacts": [],
            "artifact_failure_reason": "",
        }
        response = "\n\n".join(
            [
                "I investigated the repo and deployment first.",
                "```json\n{\"admin\": false, \"push\": false}\n```",
                "Now the structured output.",
                f"```json\n{json.dumps(structured, sort_keys=True)}\n```",
            ]
        )

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
                Path(request["result_path"]).write_text(
                    json.dumps(
                        {
                            "ok": True,
                            "mcp_cleanup_errors": [],
                            "mcp_cleanup_status": "cleaned",
                            "response": response,
                            "result": {"final_response": response, "completed": True, "api_calls": 7},
                            "session_id": "rsi-prod-conversation-123",
                        },
                        sort_keys=True,
                    ),
                    encoding="utf-8",
                )
                write_native_test_envelope(request, final_response="Architecture details are ready.")
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

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Investigate and prepare a diagram brief.",
                    "execution_id": "hexec-native-fenced-json",
                    "trace_id": "trace-native-fenced-json",
                    "workflow_id": "wf-native-fenced-json",
                    "operation_id": "op-native-fenced-json",
                    "reply_delivery_mode": "none",
                    "mcp_servers": [{"server_label": "slack", "profile": "slack_mcp_read"}],
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-native-fenced-json",
                }
            }
        )

        with tempfile.TemporaryDirectory() as tempdir, mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch(
            "rsi_runner.hermes_runtime.subprocess.Popen", FakePopen
        ), mock.patch.dict(
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
            with mock.patch.object(runtime._mcp_adapter, "plan_task_servers", return_value=TaskScopedMCPRegistration()), mock.patch.object(
                runtime._mcp_adapter,
                "register_task_servers",
                side_effect=RuntimeError("repair should not register MCP servers"),
            ) as register_mock:
                result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.message, "Architecture details are ready.")
        self.assertEqual(result.raw["execution_envelope"]["final_response"], "Architecture details are ready.")
        self.assertNotIn("structured_output", result.raw)
        register_mock.assert_not_called()

    def test_native_strict_mediated_workflow_projects_slack_report_action(self) -> None:
        structured = {
            "visible_reasoning": [],
            "reply_draft": "Campaign report is ready.",
            "final_answer": "Campaign report is ready.",
            "confidence": 0.92,
            "context_summary": "DB read result summarized.",
            "self_critique": "",
            "knowledge_drafts": [],
            "outcome_hypotheses": [],
            "produced_artifacts": [],
            "artifact_failure_reason": "",
            "proposed_actions": [
                {
                    "kind": "slack_report",
                    "target_ref": "slack:thread",
                    "request_payload": {
                        "report_schema_version": 1,
                        "summary": "Campaign report is ready.",
                        "tables": [
                            {
                                "columns": [
                                    {"key": "campaign", "label": "Campaign"},
                                    {"key": "submissions", "label": "Submissions"},
                                ],
                                "rows": [{"campaign": "Vietnamese", "submissions": 12815}],
                            }
                        ],
                    },
                    "approval_mode": "deterministic",
                    "idempotency_key": "report-1",
                    "rationale": "Deliver the requested report in Slack.",
                    "evidence_refs": [],
                }
            ],
        }
        native_response = "\n".join(["Done.", f"```json\n{json.dumps(structured, sort_keys=True)}\n```"])
        native_result = HermesExecutionResult(
            ok=True,
            message=native_response,
            provider="hermes-native-executor",
            raw={
                "native_strict": True,
                "completion_verdict": "complete",
                "termination_reason": "normal_completion",
                "execution_envelope": {
                    "completion": {"completion_verdict": "complete", "termination_reason": "normal_completion"},
                    "deliveries": [],
                    "final_response": native_response,
                },
            },
        )
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize campaign submissions.",
                    "reply_delivery_mode": "mediated",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-native-report",
                    "workflow_id": "wf-native-report",
                    "operation_id": "op-native-report",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-native-report",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "RSI_HERMES_EXECUTOR_ENABLED": "true"},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_execute_native_workflow_task_request", return_value=native_result):
                result = runtime._execute_task_internal(task)

        self.assertTrue(result.ok)
        self.assertTrue(result.raw["native_strict"])
        self.assertEqual(result.raw["structured_output"]["proposed_actions"][0]["kind"], "slack_report")
        self.assertFalse(result.raw["action_contract_repair_attempted"])

    def test_native_strict_external_tool_pending_is_not_projected(self) -> None:
        native_result = HermesExecutionResult(
            ok=True,
            message="",
            provider="hermes-native-executor",
            raw={
                "native_strict": True,
                "completion_verdict": "paused",
                "termination_reason": "external_tool_pending",
                "external_tool_pause_id": "pause-123",
                "suppress_delivery": True,
            },
        )
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Run an approved DB read.",
                    "reply_delivery_mode": "mediated",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-native-pending",
                    "workflow_id": "wf-native-pending",
                    "operation_id": "op-native-pending",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-native-pending",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "RSI_HERMES_EXECUTOR_ENABLED": "true"},
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            with mock.patch.object(runtime, "_execute_native_workflow_task_request", return_value=native_result):
                result = runtime._execute_task_internal(task)

        self.assertTrue(result.ok)
        self.assertTrue(result.raw["native_strict"])
        self.assertEqual(result.raw["termination_reason"], "external_tool_pending")
        self.assertNotIn("structured_output", result.raw)

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
            write_native_test_envelope(request, final_response="Rendered artifact")
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
                        "execution_id": "hexec-render",
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

    def test_native_executor_request_payload_includes_external_tool_resume(self) -> None:
        resume_payload = {
            "kind": "external_tool_result",
            "session_id": "rsi-prod-conversation-123",
            "tool_call_id": "call-db-read",
            "tool_name": "db_read_query",
            "status": "ok",
            "content": {
                "kind": "db_read_result",
                "request_id": "dbread_1",
                "target": "depin-prod",
                "sample": [{"unique_users": "2", "script_count": "506"}],
            },
            "transcript_snapshot": [
                {"role": "user", "content": "original request"},
                {
                    "role": "assistant",
                    "content": "",
                    "tool_calls": [
                        {
                            "id": "call-db-read",
                            "type": "function",
                            "function": {
                                "name": "db_read_query",
                                "arguments": "{\"target\":\"depin-prod\"}",
                            },
                        }
                    ],
                },
            ],
        }
        with tempfile.TemporaryDirectory() as hermes_home, tempfile.TemporaryDirectory() as tempdir, mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
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
                        "prompt": "Resume the DB read.",
                        "execution_id": "hexec-db-resume",
                        "trace_id": "trace-db-resume",
                        "workflow_id": "wf-db-resume",
                        "operation_id": "op-db-resume",
                        "session_scope_kind": "conversation",
                        "session_scope_id": "conv-db-resume",
                        "external_tool_resume": resume_payload,
                    }
                }
            )
            context = FakeSessionManager(runtime._config).prepare(task)
            request = runtime._native_executor_request_payload(
                task,
                context,
                toolsets=["rsi-db-read"],
                task_scoped_mcp_registration=TaskScopedMCPRegistration(),
                max_iterations=5,
                workdir=Path(tempdir),
                result_path=Path(tempdir) / "result.json",
                phase_contract={},
                github_cli_credentials={},
            )

        self.assertEqual(request["external_tool_resume"], resume_payload)

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

    def test_plugin_db_read_validate_posts_to_control_plane(self) -> None:
        class Response:
            def __enter__(self):
                return self

            def __exit__(self, *_args):
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "ok",
                        "validation": {"ok": True},
                    }
                ).encode("utf-8")

        seen: dict[str, object] = {}

        def fake_urlopen(req, timeout):  # type: ignore[no-untyped-def]
            seen["url"] = req.full_url
            seen["headers"] = dict(req.header_items())
            seen["body"] = json.loads(req.data.decode("utf-8"))
            seen["timeout"] = timeout
            return Response()

        with mock.patch.dict(
            os.environ,
            {
                "RSI_CONTROL_PLANE_BASE_URL": "https://control.example.test",
                "RSI_DB_READ_EXECUTION_TOKEN": "scoped-token",
            },
            clear=True,
        ), mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
            namespace: dict[str, object] = {}
            exec(_build_plugin_module(), namespace)
            handler = namespace["_tool_handler"]("db_read_validate")
            payload = json.loads(handler({"target": "depin-prod", "sql": "SELECT 1", "purpose": "query"}, task_id="sess-db"))

        self.assertEqual(payload["status"], "ok")
        self.assertEqual(payload["output"]["validation"]["ok"], True)
        self.assertEqual(seen["url"], "https://control.example.test/internal/db-read/validate")
        self.assertEqual(seen["headers"].get("Authorization"), "Bearer scoped-token")
        self.assertEqual(seen["body"]["target"], "depin-prod")
        self.assertEqual(seen["body"]["sql"], "SELECT 1")
        self.assertEqual(seen["body"]["purpose"], "query")

    def test_plugin_observability_logs_query_uses_native_handler(self) -> None:
        seen: dict[str, object] = {}

        def fake_logs_query(expr: str, **kwargs: object) -> dict[str, object]:
            seen["expr"] = expr
            seen["kwargs"] = kwargs
            return {
                "status": "success",
                "data": {
                    "result": [
                        {
                            "stream": {"namespace": "rsi-platform"},
                            "values": [["1700000000000000000", "hello from loki"]],
                        }
                    ]
                },
            }

        with tempfile.TemporaryDirectory() as hermes_home, mock.patch.dict(
            os.environ,
            {
                "HERMES_HOME": hermes_home,
                "GRAFANA_SERVER": "https://grafana.ops.storyprotocol.net",
                "GRAFANA_TOKEN": "grafana-token",
            },
            clear=True,
        ):
            namespace: dict[str, object] = {}
            exec(_build_plugin_module(), namespace)
            namespace["grafana_logs_query"] = fake_logs_query
            handler = namespace["_tool_handler"]("rsi_observability_logs_query")
            payload = json.loads(
                handler(
                    {
                        "expr": '{namespace="rsi-platform"}',
                        "since": "30m",
                        "limit": 5,
                        "direction": "backward",
                    },
                    task_id="sess-obs",
                )
            )
            lifecycle_path = Path(hermes_home, "rsi_runtime", "lifecycle", "sess-obs.jsonl")
            lifecycle_events = [json.loads(line) for line in lifecycle_path.read_text(encoding="utf-8").splitlines()]

        self.assertEqual(payload["status"], "ok")
        self.assertEqual(payload["summary"], "Returned 1 Loki log line(s).")
        self.assertEqual(payload["log_lines"], ["hello from loki"])
        self.assertEqual(seen["expr"], '{namespace="rsi-platform"}')
        self.assertEqual(seen["kwargs"]["since"], "30m")
        self.assertEqual(seen["kwargs"]["limit"], 5)
        self.assertIn("observability.query.completed", [event["event"] for event in lifecycle_events])

    def test_plugin_observability_dashboard_and_alert_tools_use_native_handlers(self) -> None:
        seen: dict[str, object] = {}

        def fake_dashboards_search(query: str = "", **kwargs: object) -> dict[str, object]:
            seen["dashboards_query"] = query
            seen["dashboards_kwargs"] = kwargs
            return {"dashboards": [{"uid": "dash1", "title": "Depin Overview"}]}

        def fake_alert_rules_search(query: str = "", **kwargs: object) -> dict[str, object]:
            seen["alerts_query"] = query
            seen["alerts_kwargs"] = kwargs
            return {"alert_rules": [{"uid": "rule1", "title": "Pod restarts"}]}

        with tempfile.TemporaryDirectory() as hermes_home, mock.patch.dict(
            os.environ,
            {
                "HERMES_HOME": hermes_home,
                "GRAFANA_SERVER": "https://grafana.ops.storyprotocol.net",
                "GRAFANA_TOKEN": "grafana-token",
            },
            clear=True,
        ):
            namespace: dict[str, object] = {}
            exec(_build_plugin_module(), namespace)
            namespace["grafana_dashboards_search"] = fake_dashboards_search
            namespace["grafana_alert_rules_search"] = fake_alert_rules_search
            dashboard_handler = namespace["_tool_handler"]("rsi_observability_dashboards_search")
            alert_handler = namespace["_tool_handler"]("rsi_observability_alert_rules_search")
            dashboard_payload = json.loads(dashboard_handler({"query": "depin", "tags": ["prod"], "limit": 3}, task_id="sess-obs"))
            alert_payload = json.loads(alert_handler({"query": "pod", "folder_uid": "infra", "limit": 4}, task_id="sess-obs"))

        self.assertEqual(dashboard_payload["summary"], "Returned 1 Grafana dashboard(s).")
        self.assertEqual(alert_payload["summary"], "Returned 1 Grafana alert rule(s).")
        self.assertEqual(seen["dashboards_query"], "depin")
        self.assertEqual(seen["dashboards_kwargs"], {"tags": ["prod"], "limit": 3})
        self.assertEqual(seen["alerts_query"], "pod")
        self.assertEqual(seen["alerts_kwargs"], {"folder_uid": "infra", "limit": 4})

    def test_plugin_db_read_query_raises_external_pause_when_enabled(self) -> None:
        class FakeExternalToolPending(Exception):
            pass

        tools_module = types.ModuleType("tools")
        pause_module = types.ModuleType("tools.external_tool_pause")
        pause_module.ExternalToolPending = FakeExternalToolPending
        original_tools = sys.modules.get("tools")
        original_pause = sys.modules.get("tools.external_tool_pause")
        sys.modules["tools"] = tools_module
        sys.modules["tools.external_tool_pause"] = pause_module

        class Response:
            def __enter__(self):
                return self

            def __exit__(self, *_args):
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "delivery_mode": "external_tool_resume",
                        "status": "validating",
                        "request": {
                            "id": "dbread_1",
                            "target": "depin-prod",
                            "state": "validating",
                            "sql_sha256": "sha256:abc",
                        },
                        "validation": {"ok": True},
                        "external_tool_pause": {
                            "id": "etpause_1",
                            "transport_tool_name": "db_read_query",
                            "tool_call_id": "call_1",
                            "hermes_session_id": "sess-db",
                        },
                    }
                ).encode("utf-8")

        seen: dict[str, object] = {}

        def fake_urlopen(req, timeout):  # type: ignore[no-untyped-def]
            seen["body"] = json.loads(req.data.decode("utf-8"))
            return Response()

        try:
            with tempfile.TemporaryDirectory() as hermes_home, mock.patch.dict(
                os.environ,
                {
                    "HERMES_HOME": hermes_home,
                    "RSI_CONTROL_PLANE_BASE_URL": "https://control.example.test",
                    "RSI_DB_READ_EXECUTION_TOKEN": "scoped-token",
                    "RSI_WORKFLOW_ID": "wf-1",
                    "RSI_TRACE_ID": "trace-1",
                    "RSI_CONVERSATION_ID": "conv-1",
                },
                clear=True,
            ), mock.patch("urllib.request.urlopen", side_effect=fake_urlopen):
                namespace: dict[str, object] = {}
                exec(_build_plugin_module(), namespace)
                handler = namespace["_tool_handler"]("db_read_query")
                with self.assertRaises(FakeExternalToolPending) as raised:
                    handler(
                        {"target": "depin-prod", "sql": "SELECT 1", "purpose": "query"},
                        task_id="sess-db",
                        session_id="sess-db",
                        tool_call_id="call_1",
                    )
                lifecycle_path = Path(hermes_home, "rsi_runtime", "lifecycle", "sess-db.jsonl")
                lifecycle_events = [json.loads(line) for line in lifecycle_path.read_text(encoding="utf-8").splitlines()]
        finally:
            if original_tools is None:
                sys.modules.pop("tools", None)
            else:
                sys.modules["tools"] = original_tools
            if original_pause is None:
                sys.modules.pop("tools.external_tool_pause", None)
            else:
                sys.modules["tools.external_tool_pause"] = original_pause

        pending = raised.exception.args[0]
        self.assertEqual(pending["kind"], "external_tool_pending")
        self.assertEqual(pending["external_tool_pause_id"], "etpause_1")
        self.assertEqual(pending["request_id"], "dbread_1")
        self.assertEqual(pending["target"], "depin-prod")
        self.assertEqual(pending["sql_sha256"], "sha256:abc")
        self.assertEqual(pending["tool_call_id"], "call_1")
        self.assertEqual(seen["body"]["workflow_id"], "wf-1")
        self.assertEqual(seen["body"]["hermes_tool_call_id"], "call_1")
        self.assertIn("external_tool.pending", [event["event"] for event in lifecycle_events])

    def test_plugin_db_read_query_returns_validation_feedback_without_pause(self) -> None:
        class Response:
            def __enter__(self):
                return self

            def __exit__(self, *_args):
                return False

            def read(self) -> bytes:
                return json.dumps(
                    {
                        "status": "validation_failed",
                        "request": {
                            "id": "dbread_bad",
                            "target": "depin-prod",
                            "state": "validation_failed",
                            "sql_sha256": "sha256:bad",
                        },
                        "validation": {
                            "ok": False,
                            "message": "syntax error at or near FROM",
                            "error_code": "offline_parse_failed",
                        },
                    }
                ).encode("utf-8")

        with mock.patch.dict(
            os.environ,
            {
                "RSI_CONTROL_PLANE_BASE_URL": "https://control.example.test",
                "RSI_DB_READ_EXECUTION_TOKEN": "scoped-token",
            },
            clear=True,
        ), mock.patch("urllib.request.urlopen", return_value=Response()):
            namespace: dict[str, object] = {}
            exec(_build_plugin_module(), namespace)
            handler = namespace["_tool_handler"]("db_read_query")
            payload = json.loads(handler({"target": "depin-prod", "sql": "SELECT FROM", "purpose": "query"}, task_id="sess-db"))

        self.assertEqual(payload["status"], "ok")
        self.assertEqual(payload["output"]["status"], "validation_failed")
        self.assertIn("repair the SQL", payload["output"]["message"])

    def test_plugin_db_read_tools_require_execution_scoped_token(self) -> None:
        with mock.patch.dict(
            os.environ,
            {
                "RSI_CONTROL_PLANE_BASE_URL": "https://control.example.test",
                "RSI_DB_READ_CLIENT_TOKEN": "static-token",
            },
            clear=True,
        ):
            namespace: dict[str, object] = {}
            exec(_build_plugin_module(), namespace)

        self.assertFalse(namespace["_db_read_tools_available"]())

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
                write_native_test_envelope(request, final_response="Final reply")
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

    def test_self_review_finalizer_stops_when_candidate_delivery_is_missing(self) -> None:
        calls: list[tuple[str, object]] = []
        fake_queue = types.ModuleType("self_review_queue")

        class FakeSelfReviewConfig:
            @classmethod
            def from_env(cls, **kwargs):
                return {"kwargs": kwargs}

        def mark_candidate_delivered(config, execution_id, result_hash, result_ref=None):
            calls.append(("delivered", execution_id))
            return {"status": "missing", "execution_id": execution_id}

        def promote_review_candidate(config, candidate_id):
            calls.append(("promote", candidate_id))
            return {"status": "enqueued", "candidate_id": candidate_id}

        fake_queue.SelfReviewConfig = FakeSelfReviewConfig
        fake_queue.mark_candidate_delivered = mark_candidate_delivered
        fake_queue.promote_review_candidate = promote_review_candidate

        with tempfile.TemporaryDirectory() as hermes_home, mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": hermes_home},
            clear=True,
        ), mock.patch(
            "rsi_runner.hermes_runtime.AIAgent",
            FakeAIAgent,
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager",
            FakeSessionManager,
        ), mock.patch.dict(
            "sys.modules",
            {"self_review_queue": fake_queue},
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            runtime._start_self_review_worker = mock.Mock()
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "execution_id": "hexec-runner-123",
                    }
                }
            )
            result = HermesExecutionResult(
                ok=True,
                message="ok",
                provider="hermes-native-executor",
                raw={
                    "native_strict": True,
                    "self_review_candidate": {"candidate_id": 7},
                    "runner_diagnostics": {"native_envelope_validated": True},
                },
            )
            final_status = {
                "status": "completed",
                "result": {"ok": True, "message": "ok", "provider": "hermes-native-executor", "raw": {}},
            }
            runtime._finalize_self_review_candidate(task, result, final_status, execution_id="hexec-runner-123")

        self.assertEqual(calls, [("delivered", "hexec-runner-123")])
        runtime._start_self_review_worker.assert_not_called()

    def test_self_review_finalizer_persists_cadence_status(self) -> None:
        calls: list[tuple[str, object]] = []
        fake_queue = types.ModuleType("self_review_queue")

        class FakeSelfReviewConfig:
            @classmethod
            def from_env(cls, **kwargs):
                return {"kwargs": kwargs}

        def mark_candidate_delivered(config, execution_id, result_hash, result_ref=None):
            calls.append(("delivered", execution_id))
            return {"status": "validated", "execution_id": execution_id}

        def promote_review_candidate(config, candidate_id):
            calls.append(("promote", candidate_id))
            return {
                "status": "enqueued",
                "candidate_id": candidate_id,
                "cadence_scope_key": "rsi:prod:slack:D123:171000001.000100",
                "memory_turns_after": 2,
                "skill_iterations_after": 0,
                "review_memory": False,
                "review_skills": True,
                "work_created": ["skill"],
            }

        def candidate_status(config, execution_id):
            return {
                "candidate_id": 7,
                "self_review_candidate_status": "enqueued",
                "self_review_status": "skill:pending",
                "self_review_work": [{"kind": "skill", "status": "pending", "review_kind": "skill", "trigger_kind": "skill"}],
            }

        fake_queue.SelfReviewConfig = FakeSelfReviewConfig
        fake_queue.mark_candidate_delivered = mark_candidate_delivered
        fake_queue.promote_review_candidate = promote_review_candidate
        fake_queue.candidate_status = candidate_status

        with tempfile.TemporaryDirectory() as hermes_home, mock.patch.dict(
            os.environ,
            {**runner_env("prod"), "HERMES_HOME": hermes_home},
            clear=True,
        ), mock.patch(
            "rsi_runner.hermes_runtime.AIAgent",
            FakeAIAgent,
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager",
            FakeSessionManager,
        ), mock.patch.dict(
            "sys.modules",
            {"self_review_queue": fake_queue},
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            runtime._start_self_review_worker = mock.Mock()
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "execution_id": "hexec-runner-456",
                    }
                }
            )
            result = HermesExecutionResult(
                ok=True,
                message="ok",
                provider="hermes-native-executor",
                raw={
                    "native_strict": True,
                    "self_review_candidate": {
                        "candidate_id": 7,
                        "cadence_scope_key": "rsi:prod:slack:D123:171000001.000100",
                        "memory_nudge_interval": 10,
                        "skill_nudge_interval": 10,
                    },
                    "runner_diagnostics": {"native_envelope_validated": True},
                },
            )
            final_status = {
                "execution_id": "hexec-runner-456",
                "status": "completed",
                "result": {"ok": True, "message": "ok", "provider": "hermes-native-executor", "raw": dict(result.raw)},
            }
            runtime._finalize_self_review_candidate(task, result, final_status, execution_id="hexec-runner-456")
            status = runtime.executor_status("hexec-runner-456")

        self.assertEqual(calls, [("delivered", "hexec-runner-456"), ("promote", 7)])
        runtime._start_self_review_worker.assert_called_once_with(7)
        self.assertEqual(status["self_review"]["memory_turns_after"], 2)
        self.assertEqual(status["self_review"]["skill_iterations_after"], 0)
        self.assertEqual(status["self_review"]["skill_nudge_interval"], 10)
        self.assertEqual(status["result"]["raw"]["self_review"]["cadence_scope_key"], "rsi:prod:slack:D123:171000001.000100")

    def test_self_review_worker_env_maps_runner_config_and_secret_allowlist(self) -> None:
        with tempfile.TemporaryDirectory() as hermes_home, mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "HERMES_HOME": hermes_home,
                "OPENROUTER_API_KEY": "openrouter-secret",
                "UNRELATED_SECRET": "must-not-copy",
                "RSI_HERMES_SELF_REVIEW_MAX_BATCH_ROWS": "1",
                "HERMES_STATE_BACKEND": "postgres",
                "HERMES_STATE_POSTGRES_URL": "postgres://user:pass@postgres/hermes",
                "HERMES_STATE_POSTGRES_SCHEMA": "hermes_state",
                "HERMES_STATE_SEARCH_MODE": "hybrid",
                "HERMES_STATE_EMBEDDINGS_ENABLED": "true",
                "HERMES_STATE_EMBEDDING_MODEL": "text-embedding-3-small",
                "HERMES_STATE_EMBEDDING_DIMENSIONS": "1536",
                "HERMES_STATE_EMBEDDING_BATCH_SIZE": "32",
            },
            clear=True,
        ), mock.patch(
            "rsi_runner.hermes_runtime.AIAgent",
            FakeAIAgent,
        ), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager",
            FakeSessionManager,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            env = runtime._self_review_worker_env()

        self.assertEqual(env["RSI_MODEL"], "deepseek/deepseek-v4-pro")
        self.assertEqual(env["RSI_PROVIDER"], "openrouter")
        self.assertEqual(env["RSI_MEMORY_BACKEND"], "honcho")
        self.assertEqual(env["RSI_HONCHO_WORKSPACE"], "rsi-stage")
        self.assertEqual(env["OPENROUTER_API_KEY"], "openrouter-secret")
        self.assertEqual(env["RSI_HERMES_SELF_REVIEW_MAX_BATCH_ROWS"], "1")
        self.assertEqual(env["HERMES_STATE_BACKEND"], "postgres")
        self.assertEqual(env["HERMES_STATE_POSTGRES_URL"], "postgres://user:pass@postgres/hermes")
        self.assertEqual(env["HERMES_STATE_POSTGRES_SCHEMA"], "hermes_state")
        self.assertEqual(env["HERMES_STATE_SEARCH_MODE"], "hybrid")
        self.assertEqual(env["HERMES_STATE_EMBEDDINGS_ENABLED"], "true")
        self.assertEqual(env["HERMES_STATE_EMBEDDING_MODEL"], "text-embedding-3-small")
        self.assertEqual(env["HERMES_STATE_EMBEDDING_DIMENSIONS"], "1536")
        self.assertEqual(env["HERMES_STATE_EMBEDDING_BATCH_SIZE"], "32")
        self.assertNotIn("UNRELATED_SECRET", env)

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
                write_native_test_envelope(request, final_response="Final reply")
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
                    "grafana=grafana-service-account-secret\n"
                    "cloudflare=cf-access-client-secret\n"
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
                "RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN": "grafana-service-account-secret",
                "RSI_UNUSED_CLIENT_SECRET": "unused-client-secret",
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
                        "execution_id": "hexec-streams",
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
        self.assertNotIn("grafana-service-account-secret", stderr_text)
        self.assertNotIn("unused-client-secret", stderr_text)
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

    def test_native_executor_subprocess_observes_without_parent_timeout_kill(self) -> None:
        source = Path(__file__).parents[1].joinpath("rsi_runner", "hermes_runtime.py").read_text(encoding="utf-8")
        method = source.split("def _execute_native_workflow_task_request", 1)[1].split("\n    def ", 1)[0]

        self.assertNotIn("effective_task_timeout + 5", method)
        self.assertNotIn("executor.inactivity_timeout", method)
        self.assertNotIn("effective_inactivity_timeout", method)
        self.assertIn('"timeout_source": "hermes_subprocess"', method)
        self.assertIn("process.poll()", method)

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
                        "execution_id": "hexec-missing-envelope",
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
        self.assertEqual(result.raw["failure_class"], "plugin_execution_envelope_missing")
        self.assertNotIn("execution_envelope", result.raw)

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
                write_native_test_envelope(request, final_response="Final reply")
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
                        "execution_id": "hexec-stat-failure",
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

    def test_native_executor_cancel_path_keeps_waits_bounded_after_failed_kill(self) -> None:
        wait_calls: list[object] = []
        runtime_holder: dict[str, HermesRuntime] = {}

        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                self.returncode = 0
                self.stdout = io.StringIO("")
                self.stderr = io.StringIO("")

            def poll(self):
                runtime = runtime_holder.get("runtime")
                if runtime is not None:
                    with runtime._executor_process_lock:
                        runtime._executor_cancel_requests.add("hexec-cancel")
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
            "rsi_runner.hermes_runtime.time.sleep", side_effect=lambda _seconds: None
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
            runtime_holder["runtime"] = runtime
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": "workflow",
                        "repo": "depin-backend",
                        "prompt": "User prompt",
                        "execution_id": "hexec-cancel",
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
        self.assertEqual(result.raw["failure_class"], "runner_cancelled")
        self.assertEqual(wait_calls, [5, 5])

    def test_native_executor_low_task_timeout_does_not_reclassify_success_result(self) -> None:
        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
                payload = {
                    "ok": True,
                    "response": "{}",
                    "result": {"final_response": "{}"},
                    "session_id": "rsi-prod-conversation-timeout-result",
                }
                Path(request["result_path"]).write_text(json.dumps(payload, sort_keys=True), encoding="utf-8")
                write_native_test_envelope(request, final_response="Final reply")
                self.returncode = 0
                self.stdout = io.StringIO("")
                self.stderr = io.StringIO("")
                self.poll_calls = 0

            def poll(self):
                self.poll_calls += 1
                return None if self.poll_calls == 1 else 0

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
        ), mock.patch(
            "rsi_runner.hermes_runtime.time.sleep", side_effect=lambda _seconds: None
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
                        "execution_id": "hexec-timeout-success-result",
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
            result = runtime._execute_native_workflow_task_request(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw.get("native_strict"), True)
        self.assertEqual(result.raw.get("native_timeout_source"), "hermes_subprocess")
        self.assertEqual(result.raw.get("termination_reason"), "normal_completion")
        self.assertNotIn("timeout_kind", result.raw)

    def test_native_partial_recovery_failure_skips_attach_envelope(self) -> None:
        class FakePopen:
            def __init__(self, cmd, cwd, env, stdout, stderr, text):
                request = json.loads(Path(cmd[-1]).read_text(encoding="utf-8"))
                payload = {
                    "ok": False,
                    "error": "Hermes subprocess reported task timeout.",
                    "response": "",
                    "result": {
                        "final_response": "",
                        "completed": False,
                        "partial": True,
                        "api_calls": 3,
                        "termination_reason": "task_timeout",
                        "completion_verdict": "partial",
                        "timeout_kind": "task_timeout",
                        "stopped_after_seconds": 1,
                    },
                    "session_id": "rsi-prod-conversation-123",
                    "termination_reason": "task_timeout",
                    "completion_verdict": "partial",
                    "timeout_kind": "task_timeout",
                    "stopped_after_seconds": 1,
                }
                Path(request["result_path"]).write_text(json.dumps(payload, sort_keys=True), encoding="utf-8")
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

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "depin-backend",
                    "prompt": "User prompt",
                    "execution_id": "hexec-native-partial-unrecoverable",
                    "trace_id": "trace-native-partial-unrecoverable",
                    "workflow_id": "wf-native-partial-unrecoverable",
                    "operation_id": "op-native-partial-unrecoverable",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-native-partial-unrecoverable",
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
            "rsi_runner.hermes_runtime.subprocess.Popen", side_effect=FakePopen
        ), mock.patch.object(
            HermesCompanyComputer,
            "attach_envelope",
            side_effect=AssertionError("native strict partial recovery failure must not synthesize an envelope"),
        ) as attach_mock, mock.patch.dict(
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
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.provider, "hermes-native-executor")
        self.assertEqual(result.raw["failure_class"], "runner_partial_completion_unrecoverable")
        self.assertEqual(result.raw["termination_reason"], "task_timeout")
        self.assertEqual(result.raw["timeout_kind"], "task_timeout")
        self.assertTrue(result.raw["native_strict"])
        self.assertEqual(result.raw["native_timeout_source"], "hermes_subprocess")
        self.assertNotIn("execution_envelope", result.raw)
        attach_mock.assert_not_called()

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
                write_native_test_envelope(request, final_response="Final reply")
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
                        "execution_id": "hexec-finalize-raises",
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
        self.assertFalse(runtime.metadata["direct_delivery_phase_enabled"])

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
        self.assertEqual(result.raw["prompt_skills"], ["architecture-diagram"])
        self.assertEqual(result.raw["resolved_skills"], ["architecture-diagram"])
        self.assertEqual(result.raw["missing_skills"], [])
        self.assertEqual(result.raw["skill_injection_mode"], "slash_command")

    def test_workflow_task_preloads_prompt_skill_once(self) -> None:
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
        self.assertEqual(result.raw["prompt_skills"], ["architecture-diagram"])
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
        self.assertEqual(result.raw["action_contract_repair_attempts"], 1)
        self.assertEqual(ActionRepairAIAgent.prompts[1].count("Runner role:"), 1)
        self.assertEqual(ActionRepairAIAgent.prompts[1].count("[PRELOADED architecture-diagram]"), 1)

    def test_action_contract_repair_default_budget_is_two(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(runtime._config.workflow_runner_repair_attempts, 2)
        self.assertEqual(runtime._action_contract_repair_attempt_limit(), 2)

    def test_action_contract_repair_preserves_zero_attempt_count_over_stale_diagnostics(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertEqual(
            runtime._action_contract_repair_attempt_count(
                {"action_contract_repair_attempts": 0},
                {"action_contract_repair_attempts": 2},
            ),
            0,
        )
        self.assertEqual(
            runtime._action_contract_repair_attempt_count(
                {},
                {"action_contract_repair_attempts": 2},
            ),
            2,
        )

    def test_action_contract_repair_retries_once_after_invalid_repair(self) -> None:
        class RetryingActionRepairAIAgent(FakeAIAgent):
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
                    "proposed_actions": [],
                }
                if type(self).calls == 3:
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
                    "prompt": "Summarize the workflow.",
                    "reply_delivery_mode": "mediated",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", RetryingActionRepairAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertTrue(result.raw["action_contract_repair_attempted"])
        self.assertTrue(result.raw["action_contract_repair_succeeded"])
        self.assertEqual(result.raw["action_contract_repair_attempts"], 2)
        self.assertEqual(RetryingActionRepairAIAgent.calls, 3)
        self.assertIn("Repair attempt: 1 of 2", RetryingActionRepairAIAgent.prompts[1])
        self.assertIn("Repair attempt: 2 of 2", RetryingActionRepairAIAgent.prompts[2])
        self.assertEqual(result.raw["structured_output"]["proposed_actions"][0]["kind"], "slack_post")

    def test_action_contract_repair_fails_closed_after_two_attempts(self) -> None:
        class FailingActionRepairAIAgent(FakeAIAgent):
            calls = 0

            def run_conversation(
                self,
                prompt: str,
                system_message: str | None = None,
                conversation_history: list[dict] | None = None,
                task_id: str | None = None,
            ) -> dict[str, object]:
                type(self).calls += 1
                return {
                    "final_response": json.dumps(
                        {
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
                            "proposed_actions": [],
                        }
                    )
                }

        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the workflow.",
                    "reply_delivery_mode": "mediated",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FailingActionRepairAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertTrue(result.raw["action_contract_repair_attempted"])
        self.assertFalse(result.raw["action_contract_repair_succeeded"])
        self.assertEqual(result.raw["action_contract_repair_attempts"], 2)
        self.assertEqual(FailingActionRepairAIAgent.calls, 3)
        self.assertIn("final_action_contract_invalid", result.raw["action_contract_repair_error"])
        self.assertEqual(result.raw["structured_output"]["proposed_actions"], [])

    def test_missing_prompt_skill_is_recorded_without_failing_workflow(self) -> None:
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
        self.assertEqual(result.raw["prompt_skills"], ["missing-skill"])
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
                    "allowed_tools": ["repo.context", "github.create_pr", "rsi.candidate_context"],
                    "tool_allowlist": ["repo.context", "github.create_pr", "rsi.candidate_context"],
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
        self.assertNotIn("tool_policy_mode", result.raw)
        self.assertNotIn("blocked_tool_names", result.raw)
        self.assertNotIn("tool_allowlist_effective", result.raw)
        self.assertNotIn("Blocked tools", FakeAIAgent.last_prompt or "")
        self.assertNotIn("Tool allowlist", FakeAIAgent.last_prompt or "")
        self.assertNotIn("github.create_pr", FakeAIAgent.last_prompt or "")

    def test_proposal_role_preserves_native_honcho_conclude_tool(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "proposal",
                    "repo": "rsi-agent-platform",
                    "prompt": "Produce a fix plan.",
                    "session_scope_kind": "proposal_candidate",
                    "session_scope_id": "shared-store:pk-collision",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:proposal",
                    "user_peer_id": "operator:alice",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeNativeHonchoAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("proposal"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertIn("honcho_profile", FakeNativeHonchoAIAgent.last_valid_tool_names)
        self.assertIn("honcho_profile", FakeNativeHonchoAIAgent.last_tool_names)
        self.assertIn("honcho_conclude", FakeNativeHonchoAIAgent.last_valid_tool_names)
        self.assertIn("honcho_conclude", FakeNativeHonchoAIAgent.last_tool_names)

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

    def test_workflow_native_toolsets_expose_artifact_tools_by_default(self) -> None:
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

        self.assertIn("rsi-artifacts", runtime._native_toolsets_for_task(lease_only_task))
        self.assertIn("rsi-artifacts", runtime._native_toolsets_for_task(requested_artifact_task))

    def test_workflow_native_toolsets_expose_db_read_when_configured(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Read from prod DB.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-123",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "true",
                "RSI_DB_READ_ENABLED": "true",
                "RSI_CONTROL_PLANE_BASE_URL": "http://control-plane.rsi-platform.svc.cluster.local:8080",
                "RSI_DB_READ_CLIENT_TOKEN": "static-secret",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertIn("rsi-db-read", runtime._native_toolsets_for_task(task))

    def test_workflow_native_toolsets_expose_observability_when_configured(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Read Loki logs.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-obs",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "false",
                "RSI_GRAFANA_BASE_URL": "https://grafana.ops.storyprotocol.net",
                "RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN": "grafana-secret-token",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        self.assertIn("rsi-observability", runtime._native_toolsets_for_task(task))
        self.assertNotIn("terminal", runtime._native_toolsets_for_task(task))

    def test_rsi_native_toolsets_are_decoupled_from_terminal_native(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Search company knowledge.",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-native-tools",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_NATIVE_TOOLS_ENABLED": "true",
                "RSI_NATIVE_TOOLS_CLIENT_TOKEN": "native-secret",
                "RSI_CONTROL_PLANE_BASE_URL": "http://control-plane.internal:8080",
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "false",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())

        toolsets = runtime._native_toolsets_for_task(task)

        self.assertIn("rsi-slack", toolsets)
        self.assertIn("rsi-notion", toolsets)
        self.assertIn("rsi-knowledge", toolsets)
        self.assertNotIn("terminal", toolsets)

    def test_native_runtime_env_mints_execution_token_and_strips_source_credentials(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Search company knowledge.",
                    "execution_id": "exec-native",
                    "operation_id": "op-native",
                    "trace_id": "trace-native",
                    "workflow_id": "wf-native",
                    "conversation_id": "conv-native",
                    "session_scope_kind": "conversation",
                    "session_scope_id": "conv-native",
                    "memory_backend": "honcho",
                    "assistant_peer_id": "rsi:stage:prod",
                    "user_peer_id": "user:alice",
                }
            }
        )

        with mock.patch("rsi_runner.hermes_runtime.SessionManager", FakeSessionManager), mock.patch.dict(
            os.environ,
            {
                **runner_env("prod"),
                "RSI_NATIVE_TOOLS_ENABLED": "true",
                "RSI_NATIVE_TOOLS_CLIENT_TOKEN": "native-secret",
                "RSI_CONTROL_PLANE_BASE_URL": "http://control-plane.internal:8080",
                "RSI_HERMES_NATIVE_TERMINAL_ENABLED": "false",
                "SLACK_BOT_TOKEN": "xoxb-test",
                "NOTION_TOKEN": "ntn_test",
                "NOTION_API_KEY": "secret_test",
                "RSI_NOTION_MCP_AUTHORIZATION": "Bearer notion",
                "RSI_DB_READ_CLIENT_TOKEN": "db-read-static",
            },
            clear=True,
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            env = os.environ.copy()
            runtime._strip_native_worker_source_credentials(env)
            native_env = runtime._native_runtime_env(
                task,
                context_path=Path("/tmp/context.json"),
                envelope_path=Path("/tmp/envelope.json"),
            )

        self.assertNotIn("SLACK_BOT_TOKEN", env)
        self.assertNotIn("NOTION_TOKEN", env)
        self.assertNotIn("NOTION_API_KEY", env)
        self.assertNotIn("RSI_NOTION_MCP_AUTHORIZATION", env)
        self.assertNotIn("RSI_NATIVE_TOOLS_CLIENT_TOKEN", env)
        self.assertNotIn("RSI_DB_READ_CLIENT_TOKEN", env)
        self.assertIn("RSI_NATIVE_TOOLS_EXECUTION_TOKEN", native_env)
        self.assertNotEqual(native_env["RSI_NATIVE_TOOLS_EXECUTION_TOKEN"], "native-secret")

    def test_generated_plugin_registers_rsi_native_toolsets(self) -> None:
        definitions = rsi_plugin_toolset_definitions()
        by_toolset: dict[str, set[str]] = {}
        for item in definitions:
            by_toolset.setdefault(str(item["toolset"]), set()).add(str(item["canonical_name"]))

        self.assertIn("rsi_slack.message_post", by_toolset["rsi-slack"])
        self.assertIn("rsi_notion.page_create", by_toolset["rsi-notion"])
        self.assertIn("rsi_knowledge.wiki_search", by_toolset["rsi-knowledge"])
        self.assertIn("rsi_observability.logs_query", by_toolset["rsi-observability"])
        self.assertIn("rsi_observability.metrics_query", by_toolset["rsi-observability"])
        self.assertIn("rsi_observability.dashboards_search", by_toolset["rsi-observability"])
        self.assertIn("rsi_observability.dashboard_get", by_toolset["rsi-observability"])
        self.assertIn("rsi_observability.alert_rules_search", by_toolset["rsi-observability"])
        self.assertIn("rsi_observability.alert_rule_get", by_toolset["rsi-observability"])
        self.assertIn("rsi_observability.active_alerts", by_toolset["rsi-observability"])
        self.assertIn("db_read.query", by_toolset["rsi-db-read"])

        message_post = transport_tool_schema("rsi_slack.message_post")
        required = set(message_post["parameters"]["required"])
        self.assertIn("reason", required)
        self.assertIn("idempotency_key", required)
        message_delete = transport_tool_schema("rsi_slack.message_delete")
        self.assertIn("confirm_destroy", set(message_delete["parameters"]["required"]))

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

    def test_prod_role_preserves_native_honcho_conclude_tool(self) -> None:
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
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeNativeHonchoAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertIn("honcho_profile", FakeNativeHonchoAIAgent.last_valid_tool_names)
        self.assertIn("honcho_profile", FakeNativeHonchoAIAgent.last_tool_names)
        self.assertIn("honcho_conclude", FakeNativeHonchoAIAgent.last_valid_tool_names)
        self.assertIn("honcho_conclude", FakeNativeHonchoAIAgent.last_tool_names)

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

    def test_removed_question_task_types_are_rejected(self) -> None:
        for task_type in ("question_gather", "question_expand", "question_reduce"):
            task = RunnerTaskRequest.from_payload(
                {
                    "task": {
                        "task_type": task_type,
                        "repo": "depin-backend",
                        "prompt": "Legacy question task.",
                    }
                }
            )
            with self.subTest(task_type=task_type), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
                runtime = HermesRuntime(RunnerConfig.from_env())
                result = runtime.execute_task(task)

            self.assertFalse(result.ok)
            self.assertEqual(result.provider, "policy")
            self.assertEqual(result.raw["task_type"], task_type)

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

        db_read_query = transport_tool_schema("db_read.query")
        self.assertEqual(db_read_query["name"], "db_read_query")
        self.assertEqual(db_read_query["parameters"]["required"], ["target", "sql", "purpose"])
        self.assertEqual(db_read_query["parameters"]["properties"]["target"]["type"], "string")
        self.assertEqual(db_read_query["parameters"]["properties"]["sql"]["type"], "string")

        slack_report_post = transport_tool_schema("rsi_slack.report_post")
        self.assertEqual(slack_report_post["name"], "rsi_slack_report_post")
        self.assertIn("summary", slack_report_post["parameters"]["properties"])
        self.assertIn("tables", slack_report_post["parameters"]["properties"])
        self.assertIn("reason", slack_report_post["parameters"]["required"])
        self.assertIn("idempotency_key", slack_report_post["parameters"]["required"])

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
                "rsi-artifacts",
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

    def test_execution_envelope_v1_legacy_capability_fields_are_ignored(self) -> None:
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

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["execution_envelope"]["contract_version"], "execution-envelope/v2")
        self.assertNotIn("capability_leases", result.raw["execution_envelope"])

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

    def test_workflow_iteration_budget_exhaustion_reports_reducer_provider_failure(self) -> None:
        class ProviderFailureReducerAIAgent:
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
                if self.instance_index == 1:
                    return {"final_response": ""}
                return {
                    "api_calls": 1,
                    "completed": False,
                    "error": "HTTP 403: Key limit exceeded (monthly limit). Manage it using https://openrouter.ai/settings/keys",
                    "failed": True,
                    "final_response": "API call failed after 3 retries: HTTP 403: Key limit exceeded (monthly limit). Manage it using https://openrouter.ai/settings/keys",
                }

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
                    "last_activity_desc": "provider quota exhausted",
                    "current_tool": "",
                    "api_call_count": 1,
                    "budget_used": 1,
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
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", ProviderFailureReducerAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(
            os.environ, runner_env("prod"), clear=True
        ):
            runtime = HermesRuntime(RunnerConfig.from_env())
            result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.raw["failure_class"], "runner_partial_completion_unrecoverable")
        self.assertIn("Hermes reducer provider call failed", result.message)
        self.assertIn("Key limit exceeded", result.message)
        self.assertNotIn("invalid JSON", result.message)
        diagnostics = result.raw["runner_diagnostics"]
        self.assertEqual(diagnostics["partial_finalization_mode"], "hermes_reducer")
        self.assertEqual(diagnostics["partial_finalization_attempts"], 1)
        self.assertIn("Hermes reducer provider call failed", diagnostics["provider_error_message"])
        self.assertIn("HTTP 403", diagnostics["provider_error_message"])
        self.assertEqual(diagnostics["reducer_attempt_errors"], [diagnostics["provider_error_message"]])
        self.assertEqual(ProviderFailureReducerAIAgent.created_instances, 2)
        self.assertEqual(ProviderFailureReducerAIAgent.init_history[1]["enabled_toolsets"], [])

    def test_workflow_evidence_ledger_projects_compact_tool_calls_and_evidence_items(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the Slack thread.",
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
        self.assertNotIn("requested_artifacts", ledger)
        self.assertNotIn("artifact_optional", ledger)
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

    def test_native_lifecycle_reply_delivery_records_rsi_slack_delivery(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply with a report.",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-rsi-slack-report",
                    "workflow_id": "wf-rsi-slack-report",
                }
            }
        )
        lifecycle_events = [
            {
                "event": "reply_delivery",
                "recorded_at_unix": 1710000003.0,
                "tool_name": "rsi_slack_report_post",
                "transport_tool_name": "rsi_slack_report_post",
                "tool_call_id": "call-report-1",
                "channel_id": "C123",
                "thread_ts": "171000001.000100",
                "body": "Structured report posted.",
                "send_status": "posted",
                "provider_ref": "slack:C123:171000002.000200",
                "artifact_refs": ["external_tool_action:extact-1"],
            },
        ]
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        observed = runtime._observability_metadata(None, task, lifecycle_events=lifecycle_events)

        self.assertEqual(observed["reply_delivery"]["tool_name"], "rsi_slack.report_post")
        self.assertEqual(observed["reply_delivery"]["send_status"], "posted")
        self.assertEqual(observed["reply_delivery"]["artifact_refs"], ["external_tool_action:extact-1"])

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
        self.assertEqual(runtime.metadata["execution_contract_version"], "execution-envelope/v2")
        self.assertEqual(runtime.metadata["runner_planner_mode"], "runner_first")
        self.assertEqual(runtime.metadata["company_computer_root"], "/tmp/hermes/workspace/company")
        self.assertNotIn("required_capabilities", runtime.metadata)
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
                self.headers = {"Content-Type": "application/json"}

            def __enter__(self) -> "FakeResponse":
                return self

            def __exit__(self, *_args: object) -> None:
                return None

            def read(self) -> bytes:
                return json.dumps(self._payload).encode("utf-8")

        observed_methods: list[str] = []

        def fake_urlopen(req, timeout: int = 0):
            self.assertEqual(req.get_header("Accept"), "application/json, text/event-stream")
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
        self.assertEqual(metadata["slack_mcp_discovery_error"], "")
        self.assertEqual(observed_methods, ["initialize", "notifications/initialized", "tools/list"])

    def test_runtime_metadata_retries_slack_mcp_after_discovery_failure(self) -> None:
        class FakeResponse:
            def __init__(self, payload: dict[str, object]) -> None:
                self._payload = payload
                self.headers = {"Content-Type": "application/json"}

            def __enter__(self) -> "FakeResponse":
                return self

            def __exit__(self, *_args: object) -> None:
                return None

            def read(self) -> bytes:
                return json.dumps(self._payload).encode("utf-8")

        calls = 0
        observed_methods: list[str] = []

        def fake_urlopen(req, timeout: int = 0):
            nonlocal calls
            calls += 1
            if calls == 1:
                raise urlerror.HTTPError(
                    req.full_url,
                    406,
                    "Not Acceptable",
                    {},
                    io.BytesIO(b'{"error":{"message":"Not Acceptable: Client must accept application/json"}}'),
                )
            payload = json.loads(req.data.decode("utf-8"))
            method = str(payload["method"])
            observed_methods.append(method)
            if method == "tools/list":
                return FakeResponse(
                    {
                        "jsonrpc": "2.0",
                        "id": payload.get("id"),
                        "result": {"tools": [{"name": "slack_read_thread", "description": "Read Slack", "annotations": {"readOnlyHint": True}}]},
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
            first_metadata = runtime.metadata
            second_metadata = runtime.metadata

        self.assertFalse(first_metadata["slack_mcp_available"])
        self.assertEqual(first_metadata["slack_mcp_tool_count"], 0)
        self.assertIn("Slack MCP initialize returned 406", first_metadata["slack_mcp_discovery_error"])
        self.assertTrue(second_metadata["slack_mcp_available"])
        self.assertEqual(second_metadata["slack_mcp_tool_count"], 1)
        self.assertEqual(second_metadata["slack_mcp_discovery_error"], "")
        self.assertEqual(observed_methods, ["initialize", "notifications/initialized", "tools/list"])

    def test_slack_mcp_response_parser_accepts_sse_json_rpc_message(self) -> None:
        env = {
            **runner_env("prod"),
            "RSI_SLACK_MCP_ENABLED": "true",
            "RSI_SLACK_MCP_SERVER_URL": "https://slack-mcp.test/mcp",
        }
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, env, clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())

        parsed = runtime._parse_slack_mcp_response(
            'event: message\n'
            'data: {"jsonrpc":"2.0","id":"tools_list","result":{"tools":[{"name":"slack_read_thread"}]}}\n\n',
            content_type="text/event-stream",
        )

        self.assertEqual(parsed["result"], {"tools": [{"name": "slack_read_thread"}]})

    def test_workflow_final_action_contract_rejects_pipe_table_slack_post(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply with a table.",
                    "reply_delivery_mode": "mediated",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                }
            }
        )
        structured_output = {
            "reply_draft": "Here is a table.",
            "final_answer": "Here is a table.",
            "proposed_actions": [
                {
                    "kind": "slack_post",
                    "request_payload": {
                        "body": "| Campaign | Count |\n|---|---|\n| Hindi | 10 |",
                    },
                }
            ],
        }

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            errors = runtime._workflow_reply_action_contract_errors(task, structured_output)

        self.assertEqual(errors[0]["code"], "unsupported_markdown_table")
        self.assertIn("slack_report.tables", errors[0]["message"])

    def test_partial_slack_post_synthesis_skips_pipe_table_with_diagnostic(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply with a table.",
                    "reply_delivery_mode": "mediated",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                    "trace_id": "trace-table",
                }
            }
        )
        structured_output = {
            "reply_draft": "| Campaign | Count |\n|---|---|\n| Hindi | 10 |",
            "final_answer": "| Campaign | Count |\n|---|---|\n| Hindi | 10 |",
            "proposed_actions": [],
        }

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            normalized, synthesized = runtime._synthesize_partial_slack_post_action(
                task,
                structured_output,
                "max_iterations",
            )

        self.assertFalse(synthesized)
        self.assertEqual(normalized.get("proposed_actions"), [])
        diagnostics = normalized["runner_diagnostics"]
        self.assertEqual(
            diagnostics["action_contract_synthesis_skipped"],
            "markdown_pipe_table_requires_slack_report",
        )
        self.assertEqual(
            normalized["action_contract_synthesis_error"]["code"],
            "markdown_pipe_table_requires_slack_report",
        )

    def test_workflow_final_action_contract_matches_platform_body_precedence(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply with a table.",
                    "reply_delivery_mode": "mediated",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                }
            }
        )
        structured_output = {
            "reply_draft": "Clean draft.",
            "final_answer": "Clean final answer.",
            "proposed_actions": [
                {
                    "kind": "slack_post",
                    "request_payload": {
                        "body": "Clean prose body.",
                        "final_body": "| Campaign | Count |\n|---|---|\n| Hindi | 10 |",
                    },
                }
            ],
        }

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            errors = runtime._workflow_reply_action_contract_errors(task, structured_output)

        self.assertEqual(errors[0]["code"], "unsupported_markdown_table")

    def test_workflow_final_action_contract_validates_slack_report_schema(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Reply with a report.",
                    "reply_delivery_mode": "mediated",
                    "channel_id": "C123",
                    "thread_ts": "171000001.000100",
                }
            }
        )
        structured_output = {
            "reply_draft": "Report ready.",
            "final_answer": "Report ready.",
            "proposed_actions": [
                {
                    "kind": "slack_report",
                    "request_payload": {
                        "report_schema_version": 1,
                        "summary": "",
                        "tables": [
                            {
                                "columns": [{"key": "campaign", "label": "Campaign"}],
                                "rows": [{"campaign": {"nested": "invalid"}}],
                            }
                        ],
                    },
                }
            ],
        }

        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch(
            "rsi_runner.hermes_runtime.SessionManager", FakeSessionManager
        ), mock.patch.dict(os.environ, runner_env("prod"), clear=True):
            runtime = HermesRuntime(RunnerConfig.from_env())
            errors = runtime._workflow_reply_action_contract_errors(task, structured_output)

        codes = {error["code"] for error in errors}
        self.assertIn("required", codes)
        self.assertIn("invalid_cell_type", codes)

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
