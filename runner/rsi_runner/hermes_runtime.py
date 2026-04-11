from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Dict, List

ROLE_TASK_TYPES = {
    "prod": {"general", "workflow", "prod"},
    "proactive": {"general", "proactive"},
    "eval": {"general", "eval"},
    "proposal": {"general", "proposal", "repo-change"},
}


try:
    from run_agent import AIAgent  # type: ignore
except Exception:  # pragma: no cover - best-effort runtime import
    AIAgent = None


@dataclass
class HermesExecutionResult:
    ok: bool
    message: str
    provider: str
    raw: Dict[str, Any]


@dataclass
class RunnerTaskRequest:
    task_type: str
    repo: str
    repo_ref: str | None
    prompt: str
    system_message: str | None
    allowed_tools: List[str]
    allowed_commands: List[str]
    timeout_seconds: int
    expected_outputs: List[str]
    artifact_destination: str | None
    context_summary: str | None
    rejected_proposal_context: List[Dict[str, Any]]

    @classmethod
    def from_payload(cls, payload: Dict[str, Any]) -> "RunnerTaskRequest":
        task = payload.get("task", payload)
        return cls(
            task_type=str(task.get("task_type", "general")),
            repo=str(task.get("repo", "rsi-agent-platform")),
            repo_ref=task.get("repo_ref"),
            prompt=str(task.get("prompt", payload.get("prompt", ""))),
            system_message=task.get("system_message", payload.get("system_message")),
            allowed_tools=[str(item) for item in task.get("allowed_tools", [])],
            allowed_commands=[str(item) for item in task.get("allowed_commands", [])],
            timeout_seconds=int(task.get("timeout_seconds", 900)),
            expected_outputs=[str(item) for item in task.get("expected_outputs", [])],
            artifact_destination=task.get("artifact_destination"),
            context_summary=task.get("context_summary"),
            rejected_proposal_context=list(task.get("rejected_proposal_context", [])),
        )


class HermesRuntime:
    def __init__(self, model: str, reasoning_effort: str, role: str = "prod") -> None:
        self._model = model
        self._reasoning_effort = reasoning_effort
        self._role = role
        self._available = AIAgent is not None

    @property
    def available(self) -> bool:
        return self._available

    def execute(self, prompt: str, system_message: str | None = None) -> HermesExecutionResult:
        if not self._available:
            return HermesExecutionResult(
                ok=False,
                message="Hermes runtime is not installed in this environment.",
                provider="stub",
                raw={"prompt": prompt, "system_message": system_message},
            )

        agent = AIAgent(model=self._model, quiet_mode=True)
        response = agent.chat(prompt if system_message is None else f"{system_message}\n\n{prompt}")
        return HermesExecutionResult(
            ok=True,
            message=response,
            provider="hermes",
            raw={"model": self._model, "reasoning_effort": self._reasoning_effort},
        )

    def execute_task(self, task: RunnerTaskRequest) -> HermesExecutionResult:
        if task.task_type not in ROLE_TASK_TYPES.get(self._role, {self._role}):
            return HermesExecutionResult(
                ok=False,
                message=f"Runner role {self._role} cannot execute task type {task.task_type}.",
                provider="policy",
                raw={"role": self._role, "task_type": task.task_type},
            )
        prompt = self._render_task_prompt(task)
        result = self.execute(prompt, system_message=task.system_message)
        result.raw = {
            **result.raw,
            "role": self._role,
            "task_type": task.task_type,
            "repo": task.repo,
            "repo_ref": task.repo_ref,
            "allowed_tools": task.allowed_tools,
            "allowed_commands": task.allowed_commands,
            "timeout_seconds": task.timeout_seconds,
            "expected_outputs": task.expected_outputs,
            "artifact_destination": task.artifact_destination,
            "context_summary": task.context_summary,
            "rejected_proposal_context": task.rejected_proposal_context,
        }
        return result

    def _render_task_prompt(self, task: RunnerTaskRequest) -> str:
        parts = [
            f"Runner role: {self._role}",
            f"Task type: {task.task_type}",
            f"Repository: {task.repo}",
        ]
        if task.repo_ref:
            parts.append(f"Repository ref: {task.repo_ref}")
        if task.context_summary:
            parts.append(f"Context: {task.context_summary}")
        if task.allowed_tools:
            parts.append(f"Allowed tools: {', '.join(task.allowed_tools)}")
        if task.allowed_commands:
            parts.append(f"Allowed commands: {', '.join(task.allowed_commands)}")
        if task.expected_outputs:
            parts.append(f"Expected outputs: {', '.join(task.expected_outputs)}")
        if task.artifact_destination:
            parts.append(f"Artifact destination: {task.artifact_destination}")
        if task.rejected_proposal_context:
            parts.append(f"Prior rejected/dismissed context: {task.rejected_proposal_context}")
        parts.append(f"Timeout seconds: {task.timeout_seconds}")
        parts.append(f"Task prompt:\n{task.prompt}")
        return "\n".join(parts)
