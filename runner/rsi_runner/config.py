from __future__ import annotations

from dataclasses import dataclass
import os
from urllib.parse import urlparse


class RunnerConfigError(ValueError):
    pass


@dataclass
class RunnerConfig:
    role: str
    host: str
    port: int
    model: str
    reasoning_effort: str
    public_base_url: str

    @classmethod
    def from_env(cls) -> "RunnerConfig":
        role = required_env("RSI_RUNNER_ROLE")
        host = required_env("RSI_RUNNER_HOST")
        port = parse_port(required_env("RSI_RUNNER_PORT"))
        model = required_env("RSI_RUNNER_MODEL")
        reasoning_effort = required_env("RSI_RUNNER_REASONING_EFFORT")
        public_base_url = required_url_env("RSI_RUNNER_PUBLIC_BASE_URL")
        if model.startswith("openai/"):
            required_env("OPENAI_API_KEY")
        return cls(
            role=role,
            host=host,
            port=port,
            model=model,
            reasoning_effort=reasoning_effort,
            public_base_url=public_base_url,
        )


def required_env(name: str) -> str:
    value = os.getenv(name, "").strip()
    if not value:
        raise RunnerConfigError(f"{name} is required")
    if value.lower().startswith("vault:"):
        raise RunnerConfigError(f"{name} must be resolved at runtime and may not start with vault:")
    return value


def required_url_env(name: str) -> str:
    value = required_env(name)
    parsed = urlparse(value)
    if not parsed.scheme or not parsed.netloc:
        raise RunnerConfigError(f"{name} must be a valid absolute URL")
    return value


def parse_port(raw: str) -> int:
    try:
        port = int(raw)
    except ValueError as exc:
        raise RunnerConfigError("RSI_RUNNER_PORT must be a positive integer") from exc
    if port <= 0:
        raise RunnerConfigError("RSI_RUNNER_PORT must be a positive integer")
    return port
