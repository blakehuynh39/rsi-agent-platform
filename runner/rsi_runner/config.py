from __future__ import annotations

from dataclasses import dataclass
import os
import re
from urllib.parse import urlparse


class RunnerConfigError(ValueError):
    pass


TRANSPORT_TIMEOUT_SAFETY_MARGIN_SECONDS = 5
DEFAULT_PROD_TASK_TIMEOUT_SECONDS = 1800


@dataclass
class ModelProfile:
    name: str
    provider: str
    configured_model: str
    provider_model: str
    base_url: str
    api_key_configured: bool
    reasoning_effort: str
    thinking_mode: str
    openrouter_provider_routing: dict[str, object]


@dataclass
class RunnerConfig:
    role: str
    executor_instance_id: str
    host: str
    port: int
    model: str
    reasoning_effort: str
    model_provider: str
    provider_model: str
    model_base_url: str
    model_api_key_configured: bool
    thinking_mode: str
    summary_model: str
    summary_reasoning_effort: str
    summary_model_provider: str
    summary_provider_model: str
    summary_model_base_url: str
    summary_model_api_key_configured: bool
    summary_thinking_mode: str
    openrouter_api_key_configured: bool
    openrouter_provider_routing: dict[str, object]
    hermes_pin: str
    public_base_url: str
    runtime_observation_sink_url: str | None
    hermes_home: str
    memory_backend: str
    honcho_workspace: str
    honcho_recall_mode: str
    honcho_write_frequency: str
    honcho_session_strategy: str
    honcho_ai_peer: str
    honcho_base_url: str | None
    honcho_environment: str
    honcho_environment_effective: str
    honcho_api_key_configured: bool
    max_iterations: int
    task_timeout_seconds: int
    inactivity_timeout_seconds: int
    transport_timeout_seconds: int
    native_max_output_tokens: int
    workflow_runner_repair_attempts: int
    hermes_executor_enabled: bool
    hermes_executor_service_only: bool
    hermes_executor_workspace_root: str
    hermes_computer_root: str
    hermes_run_root: str
    hermes_artifact_root: str
    company_wiki_root: str
    hermes_native_terminal_enabled: bool
    hermes_native_toolsets: list[str]
    hermes_terminal_env: str
    hermes_terminal_cwd: str
    hermes_terminal_timeout_seconds: int
    hermes_terminal_lifetime_seconds: int
    hermes_terminal_local_persistent: bool
    hermes_company_bin_dir: str
    hermes_kubernetes_context_enabled: bool
    hermes_prod_kubernetes_context_enabled: bool
    hermes_prod_kubernetes_context_name: str
    hermes_prod_kubernetes_cluster_name: str
    hermes_prod_kubernetes_cluster_server: str
    hermes_prod_kubernetes_cluster_ca_data: str
    hermes_prod_kubernetes_role_arn: str
    hermes_prod_kubernetes_region: str
    hermes_prod_kubernetes_namespace: str
    hermes_kubeconfig_path: str
    hermes_kubernetes_service_account_token_path: str
    hermes_kubernetes_service_account_ca_path: str
    hermes_kubernetes_service_account_namespace_path: str
    grafana_observability_configured: bool
    grafana_metrics_datasource_uid: str
    grafana_logs_datasource_uid: str
    db_read_gateway_configured: bool
    native_tools_enabled: bool
    native_tools_client_token_configured: bool
    native_tools_surfaces: list[str]
    execution_envelope_v1_enabled: bool
    execution_ledger_first_projection_enabled: bool
    runner_planner_mode: str
    verbose_trace_logging: bool
    verbose_trace_log_limit: int
    drain_timeout_seconds: int

    @classmethod
    def from_env(cls) -> "RunnerConfig":
        role = required_env("RSI_RUNNER_ROLE")
        executor_instance_id = optional_env("RSI_HERMES_EXECUTOR_INSTANCE_ID") or optional_env("HOSTNAME") or role
        host = required_env("RSI_RUNNER_HOST")
        port = parse_port(required_env("RSI_RUNNER_PORT"))
        model = optional_env("RSI_RUNNER_MODEL") or "deepseek-v4-pro"
        reasoning_effort = optional_env("RSI_RUNNER_REASONING_EFFORT") or "xhigh"
        thinking_mode = normalize_thinking_mode(optional_env("RSI_RUNNER_THINKING") or "enabled", "RSI_RUNNER_THINKING")
        openrouter_api_key = optional_env("RSI_OPENROUTER_API_KEY") or optional_env("OPENROUTER_API_KEY")
        openrouter_provider_routing = parse_openrouter_provider_routing()
        main_profile = resolve_model_profile(
            name="main",
            provider_raw=optional_env("RSI_RUNNER_PROVIDER"),
            model_raw=model,
            reasoning_effort=reasoning_effort,
            thinking_mode=thinking_mode,
            openrouter_provider_routing=openrouter_provider_routing,
        )
        summary_model = optional_env("RSI_SUMMARY_MODEL") or "deepseek-v4-flash"
        summary_reasoning_effort = optional_env("RSI_SUMMARY_REASONING_EFFORT") or "none"
        summary_thinking_mode = normalize_thinking_mode(
            optional_env("RSI_SUMMARY_THINKING") or "disabled",
            "RSI_SUMMARY_THINKING",
        )
        summary_profile = resolve_model_profile(
            name="summary",
            provider_raw=optional_env("RSI_SUMMARY_PROVIDER"),
            model_raw=summary_model,
            reasoning_effort=summary_reasoning_effort,
            thinking_mode=summary_thinking_mode,
            openrouter_provider_routing=openrouter_provider_routing,
        )
        hermes_pin = optional_env("RSI_HERMES_PIN")
        public_base_url = required_url_env("RSI_RUNNER_PUBLIC_BASE_URL")
        runtime_observation_sink_url = optional_url_env("RSI_RUNTIME_OBSERVATION_SINK_URL")
        hermes_home = required_env("HERMES_HOME")
        memory_backend = required_env("RSI_RUNNER_MEMORY_BACKEND")
        honcho_workspace = required_env("RSI_HONCHO_WORKSPACE")
        honcho_recall_mode = required_env("RSI_HONCHO_RECALL_MODE")
        honcho_write_frequency = required_env("RSI_HONCHO_WRITE_FREQUENCY")
        honcho_session_strategy = required_env("RSI_HONCHO_SESSION_STRATEGY")
        honcho_ai_peer = required_env("RSI_HONCHO_AI_PEER")
        honcho_base_url = optional_url_env("RSI_HONCHO_BASE_URL")
        honcho_environment = optional_env("RSI_HONCHO_ENVIRONMENT") or "production"
        honcho_environment_effective = normalize_honcho_environment(honcho_environment)
        honcho_api_key = optional_env("HONCHO_API_KEY")
        max_iterations = role_max_iterations(role)
        task_timeout_seconds = role_task_timeout_seconds(role)
        inactivity_timeout_seconds = role_inactivity_timeout_seconds(role, task_timeout_seconds)
        transport_timeout_seconds = role_transport_timeout_seconds(role)
        native_max_output_tokens = parse_positive_int(
            required_env("RSI_RUNNER_NATIVE_MAX_OUTPUT_TOKENS"),
            "RSI_RUNNER_NATIVE_MAX_OUTPUT_TOKENS",
        )
        validate_timeout_contract(
            role,
            task_timeout_seconds,
            inactivity_timeout_seconds,
            transport_timeout_seconds,
        )
        workflow_runner_repair_attempts = parse_non_negative_int(optional_env("RSI_WORKFLOW_RUNNER_REPAIR_ATTEMPTS") or "2", "RSI_WORKFLOW_RUNNER_REPAIR_ATTEMPTS")
        hermes_executor_enabled = parse_bool(optional_env("RSI_HERMES_EXECUTOR_ENABLED") or "false", "RSI_HERMES_EXECUTOR_ENABLED")
        hermes_executor_service_only = parse_bool(optional_env("RSI_HERMES_EXECUTOR_SERVICE_ONLY") or "false", "RSI_HERMES_EXECUTOR_SERVICE_ONLY")
        hermes_executor_workspace_root = optional_env("RSI_HERMES_EXECUTOR_WORKSPACE_ROOT") or "/workspace"
        hermes_computer_root = optional_env("RSI_HERMES_COMPUTER_ROOT") or path_join(hermes_executor_workspace_root, "company")
        hermes_run_root = optional_env("RSI_HERMES_RUN_ROOT") or path_join(hermes_computer_root, ".rsi", "runs")
        hermes_artifact_root = optional_env("RSI_HERMES_ARTIFACT_ROOT") or path_join(hermes_computer_root, "artifacts")
        company_wiki_root = optional_env("RSI_COMPANY_WIKI_ROOT") or path_join(hermes_computer_root, "wiki")
        hermes_native_terminal_enabled = parse_bool(optional_env("RSI_HERMES_NATIVE_TERMINAL_ENABLED") or "false", "RSI_HERMES_NATIVE_TERMINAL_ENABLED")
        hermes_native_toolsets = filter_hermes_native_toolsets(
            parse_csv_list(optional_env("RSI_HERMES_NATIVE_TOOLSETS") or "terminal,file,company_knowledge")
        )
        hermes_terminal_env = optional_env("TERMINAL_ENV") or "local"
        hermes_terminal_cwd = optional_env("TERMINAL_CWD") or hermes_computer_root
        hermes_terminal_timeout_seconds = parse_positive_int(optional_env("TERMINAL_TIMEOUT") or "180", "TERMINAL_TIMEOUT")
        hermes_terminal_lifetime_seconds = parse_positive_int(optional_env("TERMINAL_LIFETIME_SECONDS") or "900", "TERMINAL_LIFETIME_SECONDS")
        hermes_terminal_local_persistent = parse_bool(optional_env("TERMINAL_LOCAL_PERSISTENT") or "true", "TERMINAL_LOCAL_PERSISTENT")
        hermes_company_bin_dir = optional_env("RSI_HERMES_COMPANY_BIN_DIR") or path_join(hermes_computer_root, ".rsi", "bin")
        hermes_kubernetes_context_enabled = parse_bool(optional_env("RSI_HERMES_KUBERNETES_CONTEXT_ENABLED") or "false", "RSI_HERMES_KUBERNETES_CONTEXT_ENABLED")
        hermes_prod_kubernetes_context_enabled = parse_bool(optional_env("RSI_HERMES_PROD_KUBERNETES_CONTEXT_ENABLED") or "false", "RSI_HERMES_PROD_KUBERNETES_CONTEXT_ENABLED")
        hermes_prod_kubernetes_context_name = optional_env("RSI_HERMES_PROD_KUBERNETES_CONTEXT_NAME") or "use1-prod"
        hermes_prod_kubernetes_cluster_name = optional_env("RSI_HERMES_PROD_KUBERNETES_CLUSTER_NAME") or "use1-prod"
        hermes_prod_kubernetes_cluster_server = optional_url_env("RSI_HERMES_PROD_KUBERNETES_CLUSTER_SERVER")
        hermes_prod_kubernetes_cluster_ca_data = optional_env("RSI_HERMES_PROD_KUBERNETES_CLUSTER_CA_DATA")
        hermes_prod_kubernetes_role_arn = optional_env("RSI_HERMES_PROD_KUBERNETES_ROLE_ARN")
        hermes_prod_kubernetes_region = optional_env("RSI_HERMES_PROD_KUBERNETES_REGION") or "us-east-1"
        hermes_prod_kubernetes_namespace = optional_env("RSI_HERMES_PROD_KUBERNETES_NAMESPACE") or "story"
        hermes_kubeconfig_path = optional_env("KUBECONFIG") or path_join(hermes_computer_root, ".rsi", "kube", "config")
        hermes_kubernetes_service_account_token_path = optional_env("RSI_HERMES_KUBERNETES_SERVICE_ACCOUNT_TOKEN_PATH") or "/var/run/secrets/kubernetes.io/serviceaccount/token"
        hermes_kubernetes_service_account_ca_path = optional_env("RSI_HERMES_KUBERNETES_SERVICE_ACCOUNT_CA_PATH") or "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
        hermes_kubernetes_service_account_namespace_path = optional_env("RSI_HERMES_KUBERNETES_SERVICE_ACCOUNT_NAMESPACE_PATH") or "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
        grafana_observability_configured = bool(
            optional_env("RSI_GRAFANA_BASE_URL") and optional_env("RSI_GRAFANA_SERVICE_ACCOUNT_TOKEN")
        )
        grafana_metrics_datasource_uid = optional_env("RSI_GRAFANA_METRICS_DATASOURCE_UID") or "thanos"
        grafana_logs_datasource_uid = optional_env("RSI_GRAFANA_LOGS_DATASOURCE_UID") or "loki"
        db_read_gateway_configured = bool(
            parse_bool(optional_env("RSI_DB_READ_ENABLED") or "false", "RSI_DB_READ_ENABLED")
            and optional_env("RSI_CONTROL_PLANE_BASE_URL")
            and optional_env("RSI_DB_READ_CLIENT_TOKEN")
        )
        native_tools_enabled_env = optional_env("RSI_NATIVE_TOOLS_ENABLED")
        native_tools_enabled = True
        if native_tools_enabled_env and not parse_bool(native_tools_enabled_env, "RSI_NATIVE_TOOLS_ENABLED"):
            raise RunnerConfigError("RSI_NATIVE_TOOLS_ENABLED cannot be false; RSI native tools are required")
        native_tools_client_token = optional_env("RSI_NATIVE_TOOLS_CLIENT_TOKEN")
        if not native_tools_client_token:
            raise RunnerConfigError("RSI_NATIVE_TOOLS_CLIENT_TOKEN is required; RSI native tools are the canonical tool path")
        native_tools_surfaces = parse_csv_list(optional_env("RSI_NATIVE_TOOLS_SURFACES") or "slack,notion,knowledge,sentry,kanban,temporal")
        execution_envelope_v1_enabled = parse_bool(optional_env("RSI_EXECUTION_ENVELOPE_V1_ENABLED") or "true", "RSI_EXECUTION_ENVELOPE_V1_ENABLED")
        execution_ledger_first_projection_enabled = parse_bool(optional_env("RSI_EXECUTION_LEDGER_FIRST_PROJECTION_ENABLED") or "false", "RSI_EXECUTION_LEDGER_FIRST_PROJECTION_ENABLED")
        runner_planner_mode = optional_env("RSI_RUNNER_PLANNER_MODE") or "runner_first"
        verbose_trace_logging = parse_bool(optional_env("RSI_VERBOSE_TRACE_LOGGING") or "false", "RSI_VERBOSE_TRACE_LOGGING")
        verbose_trace_log_limit = parse_non_negative_int(optional_env("RSI_VERBOSE_TRACE_LOG_LIMIT") or "100000", "RSI_VERBOSE_TRACE_LOG_LIMIT")
        drain_timeout_seconds = parse_positive_int(optional_env("RSI_RUNNER_DRAIN_TIMEOUT_SECONDS") or "900", "RSI_RUNNER_DRAIN_TIMEOUT_SECONDS")
        if not main_profile.api_key_configured:
            raise RunnerConfigError(
                f"{provider_api_key_label(main_profile.provider)} is required when RSI_RUNNER_PROVIDER={main_profile.provider}"
            )
        if memory_backend != "honcho":
            raise RunnerConfigError("RSI_RUNNER_MEMORY_BACKEND must be set to honcho")
        if not honcho_api_key and not honcho_base_url:
            raise RunnerConfigError("HONCHO_API_KEY or RSI_HONCHO_BASE_URL is required when RSI_RUNNER_MEMORY_BACKEND=honcho")
        if hermes_executor_service_only and not runtime_observation_sink_url:
            raise RunnerConfigError("RSI_RUNTIME_OBSERVATION_SINK_URL is required when RSI_HERMES_EXECUTOR_SERVICE_ONLY=true")
        return cls(
            role=role,
            executor_instance_id=executor_instance_id,
            host=host,
            port=port,
            model=model,
            reasoning_effort=reasoning_effort,
            model_provider=main_profile.provider,
            provider_model=main_profile.provider_model,
            model_base_url=main_profile.base_url,
            model_api_key_configured=main_profile.api_key_configured,
            thinking_mode=main_profile.thinking_mode,
            summary_model=summary_model,
            summary_reasoning_effort=summary_reasoning_effort,
            summary_model_provider=summary_profile.provider,
            summary_provider_model=summary_profile.provider_model,
            summary_model_base_url=summary_profile.base_url,
            summary_model_api_key_configured=summary_profile.api_key_configured,
            summary_thinking_mode=summary_profile.thinking_mode,
            openrouter_api_key_configured=bool(openrouter_api_key),
            openrouter_provider_routing=openrouter_provider_routing,
            hermes_pin=hermes_pin,
            public_base_url=public_base_url,
            runtime_observation_sink_url=runtime_observation_sink_url or None,
            hermes_home=hermes_home,
            memory_backend=memory_backend,
            honcho_workspace=honcho_workspace,
            honcho_recall_mode=honcho_recall_mode,
            honcho_write_frequency=honcho_write_frequency,
            honcho_session_strategy=honcho_session_strategy,
            honcho_ai_peer=honcho_ai_peer,
            honcho_base_url=honcho_base_url or None,
            honcho_environment=honcho_environment,
            honcho_environment_effective=honcho_environment_effective,
            honcho_api_key_configured=bool(honcho_api_key),
            max_iterations=max_iterations,
            task_timeout_seconds=task_timeout_seconds,
            inactivity_timeout_seconds=inactivity_timeout_seconds,
            transport_timeout_seconds=transport_timeout_seconds,
            native_max_output_tokens=native_max_output_tokens,
            workflow_runner_repair_attempts=workflow_runner_repair_attempts,
            hermes_executor_enabled=hermes_executor_enabled,
            hermes_executor_service_only=hermes_executor_service_only,
            hermes_executor_workspace_root=hermes_executor_workspace_root,
            hermes_computer_root=hermes_computer_root,
            hermes_run_root=hermes_run_root,
            hermes_artifact_root=hermes_artifact_root,
            company_wiki_root=company_wiki_root,
            hermes_native_terminal_enabled=hermes_native_terminal_enabled,
            hermes_native_toolsets=hermes_native_toolsets,
            hermes_terminal_env=hermes_terminal_env,
            hermes_terminal_cwd=hermes_terminal_cwd,
            hermes_terminal_timeout_seconds=hermes_terminal_timeout_seconds,
            hermes_terminal_lifetime_seconds=hermes_terminal_lifetime_seconds,
            hermes_terminal_local_persistent=hermes_terminal_local_persistent,
            hermes_company_bin_dir=hermes_company_bin_dir,
            hermes_kubernetes_context_enabled=hermes_kubernetes_context_enabled,
            hermes_prod_kubernetes_context_enabled=hermes_prod_kubernetes_context_enabled,
            hermes_prod_kubernetes_context_name=hermes_prod_kubernetes_context_name,
            hermes_prod_kubernetes_cluster_name=hermes_prod_kubernetes_cluster_name,
            hermes_prod_kubernetes_cluster_server=hermes_prod_kubernetes_cluster_server,
            hermes_prod_kubernetes_cluster_ca_data=hermes_prod_kubernetes_cluster_ca_data,
            hermes_prod_kubernetes_role_arn=hermes_prod_kubernetes_role_arn,
            hermes_prod_kubernetes_region=hermes_prod_kubernetes_region,
            hermes_prod_kubernetes_namespace=hermes_prod_kubernetes_namespace,
            hermes_kubeconfig_path=hermes_kubeconfig_path,
            hermes_kubernetes_service_account_token_path=hermes_kubernetes_service_account_token_path,
            hermes_kubernetes_service_account_ca_path=hermes_kubernetes_service_account_ca_path,
            hermes_kubernetes_service_account_namespace_path=hermes_kubernetes_service_account_namespace_path,
            grafana_observability_configured=grafana_observability_configured,
            grafana_metrics_datasource_uid=grafana_metrics_datasource_uid,
            grafana_logs_datasource_uid=grafana_logs_datasource_uid,
            db_read_gateway_configured=db_read_gateway_configured,
            native_tools_enabled=native_tools_enabled,
            native_tools_client_token_configured=bool(native_tools_client_token),
            native_tools_surfaces=native_tools_surfaces,
            execution_envelope_v1_enabled=execution_envelope_v1_enabled,
            execution_ledger_first_projection_enabled=execution_ledger_first_projection_enabled,
            runner_planner_mode=runner_planner_mode,
            verbose_trace_logging=verbose_trace_logging,
            verbose_trace_log_limit=max(1024, verbose_trace_log_limit),
            drain_timeout_seconds=drain_timeout_seconds,
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


def resolve_model_profile(
    *,
    name: str,
    provider_raw: str,
    model_raw: str,
    reasoning_effort: str,
    thinking_mode: str,
    openrouter_provider_routing: dict[str, object],
) -> ModelProfile:
    provider, provider_model, configured_model = normalize_provider_model(provider_raw, model_raw)
    base_url = provider_base_url(provider)
    api_key_configured = bool(provider_api_key(provider))
    return ModelProfile(
        name=name,
        provider=provider,
        configured_model=configured_model,
        provider_model=provider_model,
        base_url=base_url,
        api_key_configured=api_key_configured,
        reasoning_effort=str(reasoning_effort or "").strip().lower(),
        thinking_mode=normalize_thinking_mode(thinking_mode, f"RSI_{name.upper()}_THINKING"),
        openrouter_provider_routing=dict(openrouter_provider_routing or {}) if provider == "openrouter" else {},
    )


def normalize_provider_model(provider_raw: str, model_raw: str) -> tuple[str, str, str]:
    model = str(model_raw or "").strip()
    if not model:
        raise RunnerConfigError("model must be configured")
    explicit_provider = str(provider_raw or "").strip().lower()
    if explicit_provider and explicit_provider not in {"deepseek", "openrouter"}:
        raise RunnerConfigError("model provider must be one of: deepseek, openrouter")

    if model.startswith("openrouter/"):
        provider = explicit_provider or "openrouter"
        if provider != "openrouter":
            raise RunnerConfigError("openrouter/<model> strings require provider=openrouter")
        provider_model = model.split("/", 1)[1].strip()
    elif model.startswith("deepseek/"):
        provider = explicit_provider or "deepseek"
        if provider == "openrouter":
            provider_model = model
            model = "openrouter/" + model
        else:
            provider_model = model.split("/", 1)[1].strip()
    elif model.startswith("deepseek-"):
        provider = explicit_provider or "deepseek"
        provider_model = model
        if provider == "openrouter":
            provider_model = "deepseek/" + model
            model = "openrouter/" + provider_model
    else:
        provider = explicit_provider or "openrouter"
        provider_model = model
        if provider == "openrouter" and "/" not in model:
            raise RunnerConfigError("OpenRouter model strings must include a provider namespace or use openrouter/<provider>/<model>")

    if not provider_model:
        raise RunnerConfigError("model must include a provider model name")
    return provider, provider_model, model


def normalize_thinking_mode(raw: str, name: str) -> str:
    value = str(raw or "").strip().lower()
    if value in {"enabled", "enable", "on", "true", "1", "yes"}:
        return "enabled"
    if value in {"disabled", "disable", "off", "false", "0", "no", "none"}:
        return "disabled"
    raise RunnerConfigError(f"{name} must be enabled or disabled")


def provider_base_url(provider: str) -> str:
    provider = str(provider or "").strip().lower()
    if provider == "deepseek":
        return normalize_base_url(
            first_non_empty_env(
                "RSI_DEEPSEEK_BASE_URL",
                "DEEPSEEK_BASE_URL",
            )
            or "https://api.deepseek.com",
            "DEEPSEEK_BASE_URL",
        )
    if provider == "openrouter":
        return normalize_base_url(
            first_non_empty_env("RSI_OPENROUTER_BASE_URL", "OPENROUTER_BASE_URL") or "",
            "RSI_OPENROUTER_BASE_URL",
        )
    raise RunnerConfigError("model provider must be one of: deepseek, openrouter")


def provider_api_key(provider: str) -> str:
    provider = str(provider or "").strip().lower()
    if provider == "deepseek":
        return first_non_empty_env("RSI_DEEPSEEK_API_KEY", "DEEPSEEK_API_KEY")
    if provider == "openrouter":
        return first_non_empty_env("RSI_OPENROUTER_API_KEY", "OPENROUTER_API_KEY")
    return ""


def provider_api_key_label(provider: str) -> str:
    provider = str(provider or "").strip().lower()
    if provider == "deepseek":
        return "RSI_DEEPSEEK_API_KEY or DEEPSEEK_API_KEY"
    if provider == "openrouter":
        return "RSI_OPENROUTER_API_KEY or OPENROUTER_API_KEY"
    return "provider API key"


def first_non_empty_env(*names: str) -> str:
    for name in names:
        value = optional_env(name)
        if value:
            return value
    return ""


def normalize_base_url(raw: str, name: str) -> str:
    value = str(raw or "").strip().rstrip("/")
    if not value:
        return ""
    parsed = urlparse(value)
    if not parsed.scheme or not parsed.netloc:
        raise RunnerConfigError(f"{name} must be a valid absolute URL")
    return value


def path_join(*parts: str) -> str:
    stripped = [str(part or "").strip().strip("/") for part in parts if str(part or "").strip()]
    if not stripped:
        return ""
    prefix = "/" if str(parts[0] or "").strip().startswith("/") else ""
    return prefix + "/".join(stripped)


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
    transport_headroom = max(1, role_transport_timeout_seconds(role) - TRANSPORT_TIMEOUT_SAFETY_MARGIN_SECONDS)
    if role in {"prod", "proactive"}:
        return min(DEFAULT_PROD_TASK_TIMEOUT_SECONDS, transport_headroom)
    return transport_headroom


def role_transport_timeout_seconds(role: str) -> int:
    return parse_duration_seconds(required_env(role_env_name(role, "TIMEOUT")), role_env_name(role, "TIMEOUT"))


def role_inactivity_timeout_seconds(role: str, task_timeout_seconds: int) -> int:
    raw = optional_env(role_env_name(role, "INACTIVITY_TIMEOUT"))
    if not raw:
        return max(1, task_timeout_seconds)
    value = parse_duration_seconds(raw, role_env_name(role, "INACTIVITY_TIMEOUT"))
    return max(1, min(value, task_timeout_seconds))


def validate_timeout_contract(role: str, task_timeout_seconds: int, inactivity_timeout_seconds: int, transport_timeout_seconds: int) -> None:
    role = role.strip().upper()
    if inactivity_timeout_seconds > task_timeout_seconds:
        raise RunnerConfigError(
            f"RSI_RUNNER_{role}_INACTIVITY_TIMEOUT must be less than or equal to RSI_RUNNER_{role}_TASK_TIMEOUT"
        )
    if task_timeout_seconds >= transport_timeout_seconds:
        raise RunnerConfigError(
            f"RSI_RUNNER_{role}_TASK_TIMEOUT must be less than RSI_RUNNER_{role}_TIMEOUT"
        )
    if (transport_timeout_seconds - task_timeout_seconds) < TRANSPORT_TIMEOUT_SAFETY_MARGIN_SECONDS:
        raise RunnerConfigError(
            f"RSI_RUNNER_{role}_TIMEOUT must exceed RSI_RUNNER_{role}_TASK_TIMEOUT by at least {TRANSPORT_TIMEOUT_SAFETY_MARGIN_SECONDS}s"
        )


def normalize_honcho_environment(raw: str) -> str:
    value = str(raw or "").strip().lower()
    if value in {"prod", "production", "stage", "staging"}:
        return "production"
    if value in {"local", "dev", "development"}:
        return "local"
    raise RunnerConfigError(
        "RSI_HONCHO_ENVIRONMENT must be one of: stage, prod, production, local, dev, development"
    )


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


def parse_csv_list(raw: str) -> list[str]:
    seen: set[str] = set()
    out: list[str] = []
    for item in str(raw or "").split(","):
        value = item.strip()
        if not value or value in seen:
            continue
        seen.add(value)
        out.append(value)
    return out


def filter_hermes_native_toolsets(toolsets: list[str]) -> list[str]:
    return [item for item in toolsets if item.strip() != "kanban"]


def parse_openrouter_provider_routing() -> dict[str, object]:
    routing: dict[str, object] = {}
    only = parse_csv_list(optional_env("RSI_OPENROUTER_PROVIDER_ONLY") or "")
    ignore = parse_csv_list(optional_env("RSI_OPENROUTER_PROVIDER_IGNORE") or "")
    order = parse_csv_list(optional_env("RSI_OPENROUTER_PROVIDER_ORDER") or "")
    sort = optional_env("RSI_OPENROUTER_PROVIDER_SORT")
    require_parameters = optional_env("RSI_OPENROUTER_REQUIRE_PARAMETERS")
    data_collection = optional_env("RSI_OPENROUTER_DATA_COLLECTION")

    if only:
        routing["only"] = only
    if ignore:
        routing["ignore"] = ignore
    if order:
        routing["order"] = order
    if sort:
        if sort not in {"price", "throughput", "latency"}:
            raise RunnerConfigError("RSI_OPENROUTER_PROVIDER_SORT must be one of: price, throughput, latency")
        routing["sort"] = sort
    if require_parameters:
        routing["require_parameters"] = parse_bool(require_parameters, "RSI_OPENROUTER_REQUIRE_PARAMETERS")
    if data_collection:
        if data_collection not in {"allow", "deny"}:
            raise RunnerConfigError("RSI_OPENROUTER_DATA_COLLECTION must be one of: allow, deny")
        routing["data_collection"] = data_collection
    return routing


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
