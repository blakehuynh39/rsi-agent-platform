from __future__ import annotations

import unittest

from rsi_runner.hermes_runtime import HermesExecutionResult, HermesRuntime, RunnerTaskRequest


class FakeHermesRuntime(HermesRuntime):
    def execute(self, prompt: str, system_message: str | None = None) -> HermesExecutionResult:
        return HermesExecutionResult(
            ok=True,
            message=prompt,
            provider="fake",
            raw={"system_message": system_message},
        )


class HermesRuntimeRoleTests(unittest.TestCase):
    def test_proposal_role_accepts_repo_change_task(self) -> None:
        runtime = FakeHermesRuntime(model="test", reasoning_effort="low", role="proposal")
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "repo-change",
                    "repo": "rsi-agent-platform",
                    "prompt": "Open a proposal PR.",
                }
            }
        )

        result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["role"], "proposal")
        self.assertEqual(result.raw["task_type"], "repo-change")

    def test_eval_role_rejects_repo_change_task(self) -> None:
        runtime = FakeHermesRuntime(model="test", reasoning_effort="low", role="eval")
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

    def test_task_records_structured_output_and_verbose_fields(self) -> None:
        runtime = FakeHermesRuntime(model="test", reasoning_effort="low", role="prod")
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "workflow",
                    "repo": "rsi-agent-platform",
                    "prompt": "Answer the thread.",
                    "intent": "question",
                    "trace_id": "trace-123",
                    "workflow_id": "wf-123",
                    "repo_allowlist": ["rsi-agent-platform"],
                    "tool_allowlist": ["repo.context"],
                    "response_mode": "reply_in_thread",
                    "approval_mode": "policy_gated",
                    "reasoning_verbosity": "verbose",
                }
            }
        )

        result = runtime.execute_task(task)

        self.assertTrue(result.ok)
        self.assertEqual(result.raw["intent"], "question")
        self.assertEqual(result.raw["trace_id"], "trace-123")
        self.assertEqual(result.raw["reasoning_verbosity"], "verbose")
        self.assertIn("structured_output", result.raw)
        self.assertIn("final_answer", result.raw["structured_output"])


if __name__ == "__main__":
    unittest.main()
