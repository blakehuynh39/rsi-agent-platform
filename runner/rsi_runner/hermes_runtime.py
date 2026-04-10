from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Dict


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


class HermesRuntime:
    def __init__(self, model: str, reasoning_effort: str) -> None:
        self._model = model
        self._reasoning_effort = reasoning_effort
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

