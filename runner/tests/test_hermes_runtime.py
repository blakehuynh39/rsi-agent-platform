from __future__ import annotations

import json
import os
import unittest
from unittest import mock

from rsi_runner.config import RunnerConfig
from rsi_runner.hermes_runtime import HermesRuntime, RunnerTaskRequest


class FakeAIAgent:
    last_kwargs: dict | None = None
    last_prompt: str | None = None
    last_system_message: str | None = None

    def __init__(self, **kwargs) -> None:
        type(self).last_kwargs = kwargs

    def run_conversation(self, prompt: str, system_message: str | None = None) -> dict:
        type(self).last_prompt = prompt
        type(self).last_system_message = system_message
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
                }
            )
        }


class HermesRuntimeTests(unittest.TestCase):
    def setUp(self) -> None:
        FakeAIAgent.last_kwargs = None
        FakeAIAgent.last_prompt = None
        FakeAIAgent.last_system_message = None

    def test_defaults_use_gpt54_xhigh(self) -> None:
        config = RunnerConfig.from_env()

        self.assertEqual(config.model, "openai/gpt-5.4")
        self.assertEqual(config.reasoning_effort, "xhigh")

    def test_openai_models_use_hermes_codex_responses_with_xhigh(self) -> None:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "eval",
                    "repo": "rsi-agent-platform",
                    "prompt": "Summarize the eval.",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "reasoning_verbosity": "verbose",
                }
            }
        )
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch.dict(
            os.environ,
            {"OPENAI_API_KEY": "test-key"},
            clear=False,
        ):
            runtime = HermesRuntime(model="openai/gpt-5.4", reasoning_effort="xhigh", role="eval")

            result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.provider, "hermes-aiagent")
        self.assertEqual(FakeAIAgent.last_kwargs["model"], "gpt-5.4")
        self.assertEqual(FakeAIAgent.last_kwargs["api_mode"], "codex_responses")
        self.assertEqual(FakeAIAgent.last_kwargs["provider"], "custom")
        self.assertEqual(FakeAIAgent.last_kwargs["reasoning_config"], {"enabled": True, "effort": "xhigh"})
        self.assertEqual(FakeAIAgent.last_kwargs["enabled_toolsets"], [])
        self.assertIsNone(FakeAIAgent.last_system_message)
        self.assertEqual(result.raw["model"], "openai/gpt-5.4")
        self.assertEqual(result.raw["provider_model"], "gpt-5.4")
        self.assertEqual(result.raw["api_mode"], "codex_responses")
        self.assertEqual(result.raw["reasoning_effort"], "xhigh")
        self.assertIn("structured_output", result.raw)
        self.assertEqual(result.raw["structured_output"]["final_answer"], "Final reply")

    def test_system_message_is_forwarded_to_hermes_run_conversation(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch.dict(
            os.environ,
            {"OPENAI_API_KEY": "test-key"},
            clear=False,
        ):
            runtime = HermesRuntime(model="openai/gpt-5.4", reasoning_effort="xhigh", role="prod")
            result = runtime.execute("User prompt", system_message="System directive")

        self.assertTrue(result.ok)
        self.assertEqual(FakeAIAgent.last_prompt, "User prompt")
        self.assertEqual(FakeAIAgent.last_system_message, "System directive")

    def test_openai_runtime_requires_api_key(self) -> None:
        with mock.patch("rsi_runner.hermes_runtime.AIAgent", FakeAIAgent), mock.patch.dict(os.environ, {}, clear=True):
            runtime = HermesRuntime(model="openai/gpt-5.4", reasoning_effort="xhigh", role="prod")
            result = runtime.execute("Hello")

        self.assertFalse(result.ok)
        self.assertIn("OPENAI_API_KEY", result.message)

    def test_eval_role_rejects_repo_change_task(self) -> None:
        runtime = HermesRuntime(model="openai/gpt-5.4", reasoning_effort="xhigh", role="eval")
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "repo-change",
                    "repo": "rsi-agent-platform",
                    "prompt": "This should be blocked.",
                }
            }
        )

        result = runtime.execute_task(task)

        self.assertFalse(result.ok)
        self.assertEqual(result.provider, "policy")
        self.assertIn("cannot execute", result.message)


if __name__ == "__main__":
    unittest.main()
