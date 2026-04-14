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
    hermes_home: str
    memory_backend: str
    honcho_workspace: str
    honcho_recall_mode: str
    honcho_write_frequency: str
    honcho_session_strategy: str
    honcho_ai_peer: str
    honcho_base_url: str | None
    honcho_environment: str
    honcho_api_key_configured: bool

    @classmethod
    def from_env(cls) -> "RunnerConfig":
        role = required_env("RSI_RUNNER_ROLE")
        host = required_env("RSI_RUNNER_HOST")
        port = parse_port(required_env("RSI_RUNNER_PORT"))
        model = required_env("RSI_RUNNER_MODEL")
        reasoning_effort = required_env("RSI_RUNNER_REASONING_EFFORT")
        public_base_url = required_url_env("RSI_RUNNER_PUBLIC_BASE_URL")
        hermes_home = required_env("HERMES_HOME")
        memory_backend = required_env("RSI_RUNNER_MEMORY_BACKEND")
        honcho_workspace = required_env("RSI_HONCHO_WORKSPACE")
        honcho_recall_mode = required_env("RSI_HONCHO_RECALL_MODE")
        honcho_write_frequency = required_env("RSI_HONCHO_WRITE_FREQUENCY")
        honcho_session_strategy = required_env("RSI_HONCHO_SESSION_STRATEGY")
        honcho_ai_peer = required_env("RSI_HONCHO_AI_PEER")
        honcho_base_url = optional_url_env("RSI_HONCHO_BASE_URL")
        honcho_environment = optional_env("RSI_HONCHO_ENVIRONMENT") or "production"
        honcho_api_key = optional_env("HONCHO_API_KEY")
        if model.startswith("openai/"):
            required_env("OPENAI_API_KEY")
        if memory_backend != "honcho":
            raise RunnerConfigError("RSI_RUNNER_MEMORY_BACKEND must be set to honcho")
        if not honcho_api_key and not honcho_base_url:
            raise RunnerConfigError("HONCHO_API_KEY or RSI_HONCHO_BASE_URL is required when RSI_RUNNER_MEMORY_BACKEND=honcho")
        return cls(
            role=role,
            host=host,
            port=port,
            model=model,
            reasoning_effort=reasoning_effort,
            public_base_url=public_base_url,
            hermes_home=hermes_home,
            memory_backend=memory_backend,
            honcho_workspace=honcho_workspace,
            honcho_recall_mode=honcho_recall_mode,
            honcho_write_frequency=honcho_write_frequency,
            honcho_session_strategy=honcho_session_strategy,
            honcho_ai_peer=honcho_ai_peer,
            honcho_base_url=honcho_base_url or None,
            honcho_environment=honcho_environment,
            honcho_api_key_configured=bool(honcho_api_key),
        )


def required_env(name: str) -> str:
    value = os.getenv(name, "").strip()
    if not value:
        raise RunnerConfigError(f"{name} is required")
    if value.lower().startswith("vault:"):
        raise RunnerConfigError(f"{name} must be resolved at runtime and may not start with vault:")
    return value


def optional_env(name: str) -> str:
    value = os.getenv(name, "").strip()
    if value.lower().startswith("vault:"):
        raise RunnerConfigError(f"{name} must be resolved at runtime and may not start with vault:")
    return value


def required_url_env(name: str) -> str:
    value = required_env(name)
    parsed = urlparse(value)
    if not parsed.scheme or not parsed.netloc:
        raise RunnerConfigError(f"{name} must be a valid absolute URL")
    return value


def optional_url_env(name: str) -> str:
    value = optional_env(name)
    if not value:
        return ""
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
