from __future__ import annotations

import concurrent.futures
from dataclasses import asdict, dataclass, replace
import json
import os
import time
from typing import Any, Dict, List

from .config import RunnerConfig
from .rsi_tools import (
    BLOCKED_HONCHO_TOOLS,
    CompositeToolProvider,
    READ_ONLY_HONCHO_TOOLS,
    READ_ONLY_RSI_TOOL_NAMES,
    ReadOnlyToolBinding,
    normalize_tool_names,
    tool_schema_wrappers,
)
from .session_manager import SessionManager

ROLE_TASK_TYPES = {
    "prod": {"general", "workflow", "prod"},
    "proactive": {"general", "proactive"},
    "eval": {"general", "eval"},
    "proposal": {"general", "proposal", "repo-change"},
}


try:
    from run_agent import AIAgent  # type: ignore
    from hermes_constants import parse_reasoning_effort  # type: ignore
except Exception:  # pragma: no cover - import depends on external Hermes install
    AIAgent = None

    def parse_reasoning_effort(effort: str) -> Dict[str, Any] | None:
        level = (effort or "").strip().lower()
        if not level:
            return None
        if level == "none":
            return {"enabled": False}
        if level in {"minimal", "low", "medium", "high", "xhigh"}:
            return {"enabled": True, "effort": level}
        return None


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
    intent: str | None
    trace_id: str | None
    workflow_id: str | None
    conversation_id: str | None
    case_id: str | None
    trigger_event_id: str | None
    recent_conversation_entries: List[Dict[str, Any]]
    case_summary: Dict[str, Any] | None
    prior_trace_refs: List[Dict[str, Any]]
    repo_allowlist: List[str]
    tool_allowlist: List[str]
    response_mode: str | None
    context_refs: List[Dict[str, Any]]
    approval_mode: str | None
    reasoning_verbosity: str | None
    session_scope_kind: str | None
    session_scope_id: str | None
    parent_session_scope_kind: str | None
    parent_session_scope_id: str | None
    harness_profile_id: str | None
    harness_overlay_version: str | None
    memory_backend: str | None
    assistant_peer_id: str | None
    user_peer_id: str | None

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
            intent=task.get("intent"),
            trace_id=task.get("trace_id"),
            workflow_id=task.get("workflow_id"),
            conversation_id=task.get("conversation_id"),
            case_id=task.get("case_id"),
            trigger_event_id=task.get("trigger_event_id"),
            recent_conversation_entries=list(task.get("recent_conversation_entries", [])),
            case_summary=task.get("case_summary"),
            prior_trace_refs=list(task.get("prior_trace_refs", [])),
            repo_allowlist=[str(item) for item in task.get("repo_allowlist", [])],
            tool_allowlist=[str(item) for item in task.get("tool_allowlist", [])],
            response_mode=task.get("response_mode"),
            context_refs=list(task.get("context_refs", [])),
            approval_mode=task.get("approval_mode"),
            reasoning_verbosity=task.get("reasoning_verbosity"),
            session_scope_kind=task.get("session_scope_kind"),
            session_scope_id=task.get("session_scope_id"),
            parent_session_scope_kind=task.get("parent_session_scope_kind"),
            parent_session_scope_id=task.get("parent_session_scope_id"),
            harness_profile_id=task.get("harness_profile_id"),
            harness_overlay_version=task.get("harness_overlay_version"),
            memory_backend=task.get("memory_backend"),
            assistant_peer_id=task.get("assistant_peer_id"),
            user_peer_id=task.get("user_peer_id"),
        )


@dataclass
class ToolPolicy:
    mode: str
    requested: List[str]
    effective: List[str]
    blocked: List[str]
    memory_tools: List[str]
    custom_tools: List[str]


class HermesRuntime:
    def __init__(self, config: RunnerConfig) -> None:
        self._config = config
        self._configured_model = config.model
        self._reasoning_effort = config.reasoning_effort
        self._role = config.role
        self._backend = "hermes-aiagent"
        self._provider = "hermes"
        self._api_mode = ""
        self._base_url = ""
        self._api_key = ""
        self._provider_model = config.model
        self._provider_hint = ""
        self._reasoning_config = parse_reasoning_effort(config.reasoning_effort) or {"enabled": True, "effort": "medium"}
        self._openai_configured = False
        self._session_manager = SessionManager(config)
        self._max_iterations = config.max_iterations
        self._default_task_timeout_seconds = config.task_timeout_seconds
        self._transport_timeout_seconds = config.transport_timeout_seconds
        self._tool_policy_mode = config.tool_policy_mode
        self._configure_runtime()
        self._available = AIAgent is not None and self._runtime_has_credentials() and self._session_manager.available

    def _configure_runtime(self) -> None:
        if self._configured_model.startswith("openai/"):
            self._provider = "openai"
            self._provider_hint = "custom"
            self._provider_model = self._configured_model.split("/", 1)[1]
            self._api_mode = "codex_responses"
            self._base_url = first_non_empty(
                os.getenv("RSI_OPENAI_BASE_URL"),
                os.getenv("OPENAI_BASE_URL"),
                "https://api.openai.com/v1",
            )
            self._api_key = first_non_empty(os.getenv("RSI_OPENAI_API_KEY"), os.getenv("OPENAI_API_KEY"))
            self._openai_configured = bool(self._api_key)
            return

        self._provider = first_non_empty(os.getenv("RSI_HERMES_PROVIDER"), "hermes")
        self._provider_hint = first_non_empty(os.getenv("RSI_HERMES_PROVIDER_HINT"), "")
        self._base_url = first_non_empty(os.getenv("RSI_HERMES_BASE_URL"), "")
        self._api_key = first_non_empty(os.getenv("RSI_HERMES_API_KEY"), "")
        self._api_mode = first_non_empty(os.getenv("RSI_HERMES_API_MODE"), "")

    def _runtime_has_credentials(self) -> bool:
        if self._configured_model.startswith("openai/"):
            return bool(self._api_key)
        return True

    @property
    def available(self) -> bool:
        return self._available

    @property
    def metadata(self) -> Dict[str, Any]:
        return {
            "status": "ok" if self.available else "degraded",
            "role": self._role,
            "backend": self._backend,
            "provider": self._provider,
            "model": self._configured_model,
            "provider_model": self._provider_model,
            "reasoning_effort": self._reasoning_effort,
            "api_mode": self._api_mode,
            "available": self.available,
            "hermes_available": AIAgent is not None,
            "openai_configured": self._openai_configured,
            "persistence_enabled": self._session_manager.available,
            "hermes_home": self._session_manager.hermes_home,
            "session_db_path": self._session_manager.session_db_path,
            "memory_backend": self._config.memory_backend,
            "max_iterations": self._max_iterations,
            "task_timeout_seconds": self._default_task_timeout_seconds,
            "transport_timeout_seconds": self._transport_timeout_seconds,
            "tool_policy_mode": self._tool_policy_mode,
            "tool_allowlist_effective": self._default_policy_allowlist(),
            "blocked_tool_names": [],
            "honcho_configured": self._config.honcho_api_key_configured or bool(self._config.honcho_base_url),
            "honcho_available": self._session_manager.honcho_available,
            "honcho_base_url": self._config.honcho_base_url or "",
            "honcho_workspace": self._config.honcho_workspace,
            "honcho_environment": self._config.honcho_environment,
            "honcho_recall_mode": self._config.honcho_recall_mode,
            "honcho_write_frequency": self._config.honcho_write_frequency,
            "honcho_session_strategy": self._config.honcho_session_strategy,
            "honcho_ai_peer": self._config.honcho_ai_peer,
            "issues": self._session_manager.ready_issues,
        }

    def execute(self, prompt: str, system_message: str | None = None) -> HermesExecutionResult:
        task = RunnerTaskRequest.from_payload(
            {
                "task": {
                    "task_type": "general",
                    "repo": "rsi-agent-platform",
                    "prompt": prompt,
                    "system_message": system_message,
                    "session_scope_kind": "adhoc",
                    "session_scope_id": self._role,
                    "memory_backend": self._config.memory_backend,
                    "assistant_peer_id": self._config.honcho_ai_peer,
                }
            }
        )
        return self._execute_task_request(task, self._resolve_tool_policy(task))

    def _create_agent(self, context: Any) -> Any:
        agent_kwargs: Dict[str, Any] = {
            "model": self._provider_model,
            "quiet_mode": True,
            "reasoning_config": self._reasoning_config,
            "enabled_toolsets": [],
            "skip_context_files": True,
            "skip_memory": False,
            "persist_session": True,
            "max_iterations": self._max_iterations,
            "session_id": context.session_id,
            "parent_session_id": context.parent_session_id or None,
            "session_db": self._session_manager.session_db,
        }
        if self._provider_hint:
            agent_kwargs["provider"] = self._provider_hint
        if self._api_mode:
            agent_kwargs["api_mode"] = self._api_mode
        if self._base_url:
            agent_kwargs["base_url"] = self._base_url
        if self._api_key:
            agent_kwargs["api_key"] = self._api_key
        return AIAgent(**agent_kwargs)

    def _execute_task_request(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> HermesExecutionResult:
        if AIAgent is None:
            return HermesExecutionResult(
                ok=False,
                message="Hermes runtime is not installed in this environment.",
                provider=self._backend,
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )
        if not self._runtime_has_credentials():
            return HermesExecutionResult(
                ok=False,
                message="Hermes OpenAI runtime selected but RSI_OPENAI_API_KEY / OPENAI_API_KEY is not configured.",
                provider=self._backend,
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )
        if not self._session_manager.available:
            return HermesExecutionResult(
                ok=False,
                message="Hermes persistent session runtime is unavailable.",
                provider=self._backend,
                raw=self._base_raw(prompt=task.prompt, system_message=task.system_message),
            )

        context = self._session_manager.prepare(task)
        effective_task_timeout = self._effective_task_timeout(task)
        try:
            agent = self._create_agent(context)
            self._attach_tool_policy(agent, task, tool_policy)
            tracker = self._session_manager.attach_tracking(agent, task, context)
            timed_out, run_result, timeout_meta = self._run_with_task_timeout(agent, task, context, effective_task_timeout)
            if timed_out:
                finalized = self._session_manager.finalize(context, tracker)
                return HermesExecutionResult(
                    ok=False,
                    message=f"Hermes execution timed out after {effective_task_timeout}s.",
                    provider=self._backend,
                    raw={
                        **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                        **finalized,
                        **timeout_meta,
                        "task_timeout_seconds": effective_task_timeout,
                        "transport_timeout_seconds": self._transport_timeout_seconds,
                        "max_iterations": self._max_iterations,
                        "tool_policy_mode": tool_policy.mode,
                        "tool_allowlist_effective": tool_policy.effective,
                        "blocked_tool_names": tool_policy.blocked,
                        "termination_reason": "task_timeout",
                    },
                )
            response = str((run_result or {}).get("final_response", "") or "")
        except Exception as exc:
            return HermesExecutionResult(
                ok=False,
                message=f"Hermes execution failed: {exc}",
                provider=self._backend,
                raw={**self._base_raw(prompt=task.prompt, system_message=task.system_message), "error": str(exc)},
            )

        finalized = self._session_manager.finalize(context, tracker)
        return HermesExecutionResult(
            ok=True,
            message=response,
            provider=self._backend,
            raw={
                **self._base_raw(prompt=task.prompt, system_message=task.system_message),
                **finalized,
                "task_timeout_seconds": effective_task_timeout,
                "transport_timeout_seconds": self._transport_timeout_seconds,
                "max_iterations": self._max_iterations,
                "tool_policy_mode": tool_policy.mode,
                "tool_allowlist_effective": tool_policy.effective,
                "blocked_tool_names": tool_policy.blocked,
                "termination_reason": "normal_completion",
                "max_iterations_reached": False,
                "harness_profile_id": task.harness_profile_id,
                "effective_overlay_version": task.harness_overlay_version,
            },
        )

    def _base_raw(self, prompt: str = "", system_message: str | None = None) -> Dict[str, Any]:
        return {
            "role": self._role,
            "backend": self._backend,
            "provider": self._provider,
            "provider_hint": self._provider_hint,
            "api_mode": self._api_mode,
            "model": self._configured_model,
            "provider_model": self._provider_model,
            "reasoning_effort": self._reasoning_effort,
            "reasoning_config": self._reasoning_config,
            "base_url": self._base_url,
            "honcho_base_url": self._config.honcho_base_url or "",
            "honcho_workspace": self._config.honcho_workspace,
            "honcho_environment": self._config.honcho_environment,
            "honcho_recall_mode": self._config.honcho_recall_mode,
            "honcho_write_frequency": self._config.honcho_write_frequency,
            "honcho_session_strategy": self._config.honcho_session_strategy,
            "honcho_ai_peer": self._config.honcho_ai_peer,
            "prompt": prompt,
            "system_message": system_message,
        }

    def _default_policy_allowlist(self) -> List[str]:
        if self._role not in {"eval", "proposal"}:
            return sorted(READ_ONLY_HONCHO_TOOLS)
        return sorted(list(READ_ONLY_HONCHO_TOOLS) + list(READ_ONLY_RSI_TOOL_NAMES))

    def _resolve_tool_policy(self, task: RunnerTaskRequest) -> ToolPolicy:
        requested = normalize_tool_names([*task.allowed_tools, *task.tool_allowlist])
        if self._role not in {"eval", "proposal"}:
            effective = requested
            return ToolPolicy(
                mode=self._tool_policy_mode,
                requested=requested,
                effective=effective,
                blocked=[],
                memory_tools=sorted(READ_ONLY_HONCHO_TOOLS),
                custom_tools=[],
            )
        permitted = set(READ_ONLY_HONCHO_TOOLS) | set(READ_ONLY_RSI_TOOL_NAMES)
        effective = normalize_tool_names(requested or sorted(permitted))
        effective = [name for name in effective if name in permitted]
        blocked = [name for name in requested if name not in permitted]
        return ToolPolicy(
            mode=self._tool_policy_mode,
            requested=requested,
            effective=effective,
            blocked=blocked,
            memory_tools=sorted([name for name in effective if name in READ_ONLY_HONCHO_TOOLS]),
            custom_tools=sorted([name for name in effective if name in READ_ONLY_RSI_TOOL_NAMES]),
        )

    def _effective_task_timeout(self, task: RunnerTaskRequest) -> int:
        requested = int(task.timeout_seconds or 0)
        candidates = [self._default_task_timeout_seconds]
        if requested > 0:
            candidates.append(requested)
        if self._transport_timeout_seconds > 5:
            candidates.append(self._transport_timeout_seconds - 5)
        return max(1, min(value for value in candidates if value > 0))

    def _attach_tool_policy(self, agent: Any, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> None:
        current_tools = list(getattr(agent, "tools", []) or [])
        current_valid = set(getattr(agent, "valid_tool_names", set()) or set())
        filtered_tools = []
        if self._role in {"eval", "proposal"}:
            allowed_names = set(tool_policy.effective)
            filtered_tools = [tool for tool in current_tools if tool_name(tool) in allowed_names]
        else:
            filtered_tools = current_tools
        custom_tool_names = [name for name in tool_policy.custom_tools if self._config.tool_gateway_base_url]
        if custom_tool_names:
            filtered_tools.extend(tool_schema_wrappers(custom_tool_names))
            readonly_tools = ReadOnlyToolBinding(
                base_url=self._config.tool_gateway_base_url or "",
                task_repo=task.repo,
                task_prompt=task.prompt,
                task_context_summary=task.context_summary or "",
                trace_id=task.trace_id or "",
                session_scope_kind=task.session_scope_kind or "",
                session_scope_id=task.session_scope_id or "",
                context_refs=task.context_refs,
            )
            agent._memory_manager = CompositeToolProvider(getattr(agent, "_memory_manager", None), readonly_tools)
        elif self._role in {"eval", "proposal"} and getattr(agent, "_memory_manager", None) is not None:
            agent._memory_manager = CompositeToolProvider(getattr(agent, "_memory_manager", None), ReadOnlyToolBinding(
                base_url=self._config.tool_gateway_base_url or "",
                task_repo=task.repo,
                task_prompt=task.prompt,
                task_context_summary=task.context_summary or "",
                trace_id=task.trace_id or "",
                session_scope_kind=task.session_scope_kind or "",
                session_scope_id=task.session_scope_id or "",
                context_refs=task.context_refs,
            ))
        if self._role in {"eval", "proposal"}:
            effective_names = set(tool_policy.effective)
            current_valid = {name for name in current_valid if name in effective_names and name not in BLOCKED_HONCHO_TOOLS}
            current_valid.update(custom_tool_names)
        agent.tools = filtered_tools
        agent.valid_tool_names = current_valid

    def _run_with_task_timeout(self, agent: Any, task: RunnerTaskRequest, context: Any, timeout_seconds: int) -> tuple[bool, Dict[str, Any] | None, Dict[str, Any]]:
        executor = concurrent.futures.ThreadPoolExecutor(max_workers=1)
        future = executor.submit(
            agent.run_conversation,
            task.prompt,
            task.system_message,
            context.conversation_history,
        )
        try:
            result = future.result(timeout=timeout_seconds)
            activity = safe_activity_summary(agent)
            return False, result, {
                "last_activity": activity,
                "last_tool_invoked": string_from_map(activity, "current_tool"),
                "max_iterations_reached": bool(activity.get("budget_used", 0) >= activity.get("budget_max", 0) and activity.get("budget_max", 0) > 0),
            }
        except concurrent.futures.TimeoutError:
            agent.interrupt(f"runner task timeout after {timeout_seconds}s")
            try:
                future.result(timeout=min(5, max(1, timeout_seconds//10)))
            except Exception:
                pass
            activity = safe_activity_summary(agent)
            return True, None, {
                "timeout_kind": "task_timeout",
                "last_activity": activity,
                "last_activity_desc": string_from_map(activity, "last_activity_desc"),
                "last_tool_invoked": string_from_map(activity, "current_tool"),
                "max_iterations_reached": bool(activity.get("budget_used", 0) >= activity.get("budget_max", 0) and activity.get("budget_max", 0) > 0),
                "timed_out_after_seconds": timeout_seconds,
            }
        finally:
            executor.shutdown(wait=False, cancel_futures=True)

    def execute_task(self, task: RunnerTaskRequest) -> HermesExecutionResult:
        if task.task_type not in ROLE_TASK_TYPES.get(self._role, {self._role}):
            return HermesExecutionResult(
                ok=False,
                message=f"Runner role {self._role} cannot execute task type {task.task_type}.",
                provider="policy",
                raw={"role": self._role, "task_type": task.task_type},
            )
        tool_policy = self._resolve_tool_policy(task)
        prompt = self._render_task_prompt(task, tool_policy)
        rendered_task = replace(task, prompt=prompt)
        result = self._execute_task_request(rendered_task, tool_policy)
        structured_output = self._extract_structured_output(result.message)
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
            "intent": task.intent,
            "trace_id": task.trace_id,
            "workflow_id": task.workflow_id,
            "repo_allowlist": task.repo_allowlist,
            "tool_allowlist": task.tool_allowlist,
            "tool_policy_mode": tool_policy.mode,
            "tool_allowlist_effective": tool_policy.effective,
            "blocked_tool_names": tool_policy.blocked,
            "response_mode": task.response_mode,
            "context_refs": task.context_refs,
            "approval_mode": task.approval_mode,
            "reasoning_verbosity": task.reasoning_verbosity,
            "session_scope_kind": task.session_scope_kind,
            "session_scope_id": task.session_scope_id,
            "parent_session_scope_kind": task.parent_session_scope_kind,
            "parent_session_scope_id": task.parent_session_scope_id,
            "harness_profile_id": task.harness_profile_id,
            "harness_overlay_version": task.harness_overlay_version,
            "memory_backend": task.memory_backend,
            "assistant_peer_id": task.assistant_peer_id,
            "user_peer_id": task.user_peer_id,
            "structured_output": structured_output,
        }
        return result

    def _render_task_prompt(self, task: RunnerTaskRequest, tool_policy: ToolPolicy) -> str:
        parts = [
            f"Runner role: {self._role}",
            f"Task type: {task.task_type}",
            f"Repository: {task.repo}",
            f"Configured model: {self._configured_model}",
            f"Reasoning effort: {self._reasoning_effort}",
            f"Max iterations: {self._max_iterations}",
            f"Task timeout seconds: {self._effective_task_timeout(task)}",
            f"Transport timeout seconds: {self._transport_timeout_seconds}",
            f"Tool policy mode: {tool_policy.mode}",
        ]
        if task.repo_ref:
            parts.append(f"Repository ref: {task.repo_ref}")
        if task.intent:
            parts.append(f"Intent: {task.intent}")
        if task.trace_id:
            parts.append(f"Trace ID: {task.trace_id}")
        if task.workflow_id:
            parts.append(f"Workflow ID: {task.workflow_id}")
        if task.context_summary:
            parts.append(f"Context: {task.context_summary}")
        if task.context_refs:
            parts.append(f"Context refs: {json.dumps(task.context_refs)}")
        if task.allowed_tools:
            parts.append(f"Requested allowed tools: {', '.join(task.allowed_tools)}")
        if task.tool_allowlist:
            parts.append(f"Requested tool allowlist: {', '.join(task.tool_allowlist)}")
        if tool_policy.effective:
            parts.append(f"Effective tool allowlist: {', '.join(tool_policy.effective)}")
        if tool_policy.blocked:
            parts.append(f"Blocked tools by policy: {', '.join(tool_policy.blocked)}")
        if task.allowed_commands:
            parts.append(f"Allowed commands: {', '.join(task.allowed_commands)}")
        if task.repo_allowlist:
            parts.append(f"Repo allowlist: {', '.join(task.repo_allowlist)}")
        if task.expected_outputs:
            parts.append(f"Expected outputs: {', '.join(task.expected_outputs)}")
        if task.artifact_destination:
            parts.append(f"Artifact destination: {task.artifact_destination}")
        if task.rejected_proposal_context:
            parts.append(f"Prior rejected/dismissed context: {json.dumps(task.rejected_proposal_context)}")
        if task.response_mode:
            parts.append(f"Response mode: {task.response_mode}")
        if task.approval_mode:
            parts.append(f"Approval mode: {task.approval_mode}")
        if task.reasoning_verbosity:
            parts.append(f"Reasoning verbosity: {task.reasoning_verbosity}")
        if task.session_scope_kind:
            parts.append(f"Session scope: {task.session_scope_kind}:{task.session_scope_id}")
        if task.parent_session_scope_kind:
            parts.append(f"Parent session scope: {task.parent_session_scope_kind}:{task.parent_session_scope_id}")
        if task.harness_profile_id:
            parts.append(f"Harness profile: {task.harness_profile_id}")
        if task.harness_overlay_version:
            parts.append(f"Effective harness overlay: {task.harness_overlay_version}")
        if task.memory_backend:
            parts.append(f"Memory backend: {task.memory_backend}")
        parts.append(f"Timeout seconds: {task.timeout_seconds}")
        parts.append("Use only the effective tool allowlist above. Proposal and eval roles are read-only; they must not mutate repos, launch jobs, open PRs, or post to Slack.")
        parts.append(
            "Return a JSON object with keys: visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, outcome_hypotheses, change_plan, repo_patch, validation_plan, retry_assessment, hypothesis_delta."
        )
        parts.append(
            "Each proposed action must include: kind, target_ref, request_payload, approval_mode, idempotency_key, rationale, evidence_refs."
        )
        parts.append(
            "Each knowledge draft must include: kind, scope_type, scope_id, title, summary, body, confidence, fresh_until, evidence_refs."
        )
        parts.append(
            "Each outcome hypothesis must include: outcome_type, success_condition, measurement_ref, expected_time_horizon."
        )
        parts.append(
            "For proposal or repo-change tasks, change_plan must explain the concrete remediation, repo_patch must contain a unified diff when target_layer is repo_change, validation_plan must name the checks to run, retry_assessment must include failure_class, failure_summary, retry_decision, material_hypothesis_change, and changed_files, and hypothesis_delta must explain what changed from the prior failed attempt."
        )
        parts.append(f"Task prompt:\n{task.prompt}")
        return "\n".join(parts)

    def _extract_structured_output(self, message: str) -> Dict[str, Any]:
        text = (message or "").strip()
        if text:
            try:
                parsed = json.loads(text)
                if isinstance(parsed, dict):
                    return parsed
            except json.JSONDecodeError:
                pass
        return {
            "visible_reasoning": [
                {
                    "step_type": "fallback",
                    "summary": "Runtime returned unstructured text; preserving it as the visible answer.",
                    "alternatives": [],
                    "confidence": 0.5,
                    "decision": text,
                }
            ],
            "reply_draft": text,
            "final_answer": text,
            "confidence": 0.5,
            "context_summary": "",
            "self_critique": "",
            "proposed_actions": [],
            "knowledge_drafts": [],
            "outcome_hypotheses": [],
            "change_plan": "",
            "repo_patch": "",
            "validation_plan": "",
            "retry_assessment": {
                "failure_class": "",
                "failure_summary": "",
                "retry_decision": "",
                "material_hypothesis_change": False,
                "changed_files": [],
            },
            "hypothesis_delta": "",
        }


def first_non_empty(*values: str | None) -> str:
    for value in values:
        if value and value.strip():
            return value.strip()
    return ""


def tool_name(schema: Dict[str, Any]) -> str:
    if not isinstance(schema, dict):
        return ""
    function = schema.get("function", {})
    if not isinstance(function, dict):
        return ""
    value = function.get("name", "")
    return str(value).strip()


def safe_activity_summary(agent: Any) -> Dict[str, Any]:
    try:
        summary = agent.get_activity_summary()
        if isinstance(summary, dict):
            return summary
    except Exception:
        pass
    return {}


def string_from_map(values: Dict[str, Any], key: str) -> str:
    value = values.get(key, "")
    return str(value or "").strip()
