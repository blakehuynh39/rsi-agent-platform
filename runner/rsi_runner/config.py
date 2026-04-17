from __future__ import annotations

from dataclasses import dataclass
import os
import re
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
    hermes_pin: str
    public_base_url: str
    tool_gateway_base_url: str | None
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
    max_iterations: int
    task_timeout_seconds: int
    inactivity_timeout_seconds: int
    transport_timeout_seconds: int
    tool_policy_mode: str
    workflow_runner_repair_attempts: int
    hermes_native_governed_tools_enabled: bool

    @classmethod
    def from_env(cls) -> "RunnerConfig":
        role = required_env("RSI_RUNNER_ROLE")
        host = required_env("RSI_RUNNER_HOST")
        port = parse_port(required_env("RSI_RUNNER_PORT"))
        model = required_env("RSI_RUNNER_MODEL")
        reasoning_effort = required_env("RSI_RUNNER_REASONING_EFFORT")
        hermes_pin = optional_env("RSI_HERMES_PIN")
        public_base_url = required_url_env("RSI_RUNNER_PUBLIC_BASE_URL")
        tool_gateway_base_url = optional_url_env("RSI_TOOL_GATEWAY_BASE_URL")
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
        max_iterations = role_max_iterations(role)
        task_timeout_seconds = role_task_timeout_seconds(role)
        inactivity_timeout_seconds = role_inactivity_timeout_seconds(role, task_timeout_seconds)
        transport_timeout_seconds = role_transport_timeout_seconds(role)
        tool_policy_mode = role_tool_policy_mode(role)
        workflow_runner_repair_attempts = parse_non_negative_int(optional_env("RSI_WORKFLOW_RUNNER_REPAIR_ATTEMPTS") or "1", "RSI_WORKFLOW_RUNNER_REPAIR_ATTEMPTS")
        hermes_native_governed_tools_enabled = parse_bool(optional_env("RSI_HERMES_NATIVE_GOVERNED_TOOLS_ENABLED") or "false", "RSI_HERMES_NATIVE_GOVERNED_TOOLS_ENABLED")
        if model.startswith("openai/"):
            required_env("OPENAI_API_KEY")
        if memory_backend != "honcho":
            raise RunnerConfigError("RSI_RUNNER_MEMORY_BACKEND must be set to honcho")
        if not honcho_api_key and not honcho_base_url:
            raise RunnerConfigError("HONCHO_API_KEY or RSI_HONCHO_BASE_URL is required when RSI_RUNNER_MEMORY_BACKEND=honcho")
        if role in {"eval", "proposal"} and not tool_gateway_base_url:
            raise RunnerConfigError("RSI_TOOL_GATEWAY_BASE_URL is required for eval and proposal runner roles")
        return cls(
            role=role,
            host=host,
            port=port,
            model=model,
            reasoning_effort=reasoning_effort,
            hermes_pin=hermes_pin,
            public_base_url=public_base_url,
            tool_gateway_base_url=tool_gateway_base_url or None,
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
            max_iterations=max_iterations,
            task_timeout_seconds=task_timeout_seconds,
            inactivity_timeout_seconds=inactivity_timeout_seconds,
            transport_timeout_seconds=transport_timeout_seconds,
            tool_policy_mode=tool_policy_mode,
            workflow_runner_repair_attempts=workflow_runner_repair_attempts,
            hermes_native_governed_tools_enabled=hermes_native_governed_tools_enabled,
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


def role_env_name(role: str, suffix: str) -> str:
    return f"RSI_RUNNER_{role.strip().upper()}_{suffix.strip().upper()}"


def role_max_iterations(role: str) -> int:
    env_name = role_env_name(role, "MAX_ITERATIONS")
    raw = required_env(env_name) if role in {"eval", "proposal"} else optional_env(env_name)
    if not raw:
        return 1
    value = parse_positive_int(raw, env_name)
    if value <= 0:
        raise RunnerConfigError(f"{env_name} must be greater than 0")
    return value


def role_task_timeout_seconds(role: str) -> int:
    env_name = role_env_name(role, "TASK_TIMEOUT")
    raw = required_env(env_name) if role in {"eval", "proposal"} else optional_env(env_name)
    if raw:
        return parse_duration_seconds(raw, env_name)
    return max(1, role_transport_timeout_seconds(role)-5)


def role_transport_timeout_seconds(role: str) -> int:
    return parse_duration_seconds(required_env(role_env_name(role, "TIMEOUT")), role_env_name(role, "TIMEOUT"))


def role_inactivity_timeout_seconds(role: str, task_timeout_seconds: int) -> int:
    raw = optional_env(role_env_name(role, "INACTIVITY_TIMEOUT"))
    if not raw:
        return max(1, task_timeout_seconds)
    value = parse_duration_seconds(raw, role_env_name(role, "INACTIVITY_TIMEOUT"))
    return max(1, min(value, task_timeout_seconds))


def role_tool_policy_mode(role: str) -> str:
    return "enforced_read_only"


def parse_positive_int(raw: str, name: str) -> int:
    try:
        value = int(raw)
    except ValueError as exc:
        raise RunnerConfigError(f"{name} must be a positive integer") from exc
    if value <= 0:
        raise RunnerConfigError(f"{name} must be a positive integer")
    return value


def parse_non_negative_int(raw: str, name: str) -> int:
    try:
        value = int(raw)
    except ValueError as exc:
        raise RunnerConfigError(f"{name} must be a non-negative integer") from exc
    if value < 0:
        raise RunnerConfigError(f"{name} must be a non-negative integer")
    return value


def parse_bool(raw: str, name: str) -> bool:
    text = str(raw or "").strip().lower()
    if text in {"1", "true", "t", "yes", "y", "on"}:
        return True
    if text in {"0", "false", "f", "no", "n", "off"}:
        return False
    raise RunnerConfigError(f"{name} must be a boolean")


_DURATION_RE = re.compile(r"^(?P<value>\d+(?:\.\d+)?)(?P<unit>ms|s|m)?$")


def parse_duration_seconds(raw: str, name: str) -> int:
    match = _DURATION_RE.match((raw or "").strip())
    if not match:
        raise RunnerConfigError(f"{name} must be a duration like 300s, 5m, or 500ms")
    value = float(match.group("value"))
    unit = match.group("unit") or "s"
    multiplier = 1.0
    if unit == "ms":
        multiplier = 0.001
    elif unit == "m":
        multiplier = 60.0
    seconds = int(max(1, round(value * multiplier)))
    if seconds <= 0:
        raise RunnerConfigError(f"{name} must resolve to at least 1 second")
    return seconds
