from __future__ import annotations

from dataclasses import dataclass
import os


@dataclass
class RunnerConfig:
    role: str = "prod"
    host: str = "0.0.0.0"
    port: int = 8090
    model: str = "openai/gpt-5.4"
    reasoning_effort: str = "xhigh"
    public_base_url: str = "http://localhost:8090"

    @classmethod
    def from_env(cls) -> "RunnerConfig":
        return cls(
            role=os.getenv("RSI_RUNNER_ROLE", "prod"),
            host=os.getenv("RSI_RUNNER_HOST", "0.0.0.0"),
            port=int(os.getenv("RSI_RUNNER_PORT", "8090")),
            model=os.getenv("RSI_RUNNER_MODEL", "openai/gpt-5.4"),
            reasoning_effort=os.getenv("RSI_RUNNER_REASONING_EFFORT", "xhigh"),
            public_base_url=os.getenv("RSI_RUNNER_PUBLIC_BASE_URL", "http://localhost:8090"),
        )
