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


if __name__ == "__main__":
    unittest.main()
