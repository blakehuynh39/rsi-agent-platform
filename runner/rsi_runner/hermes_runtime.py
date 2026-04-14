from __future__ import annotations

from dataclasses import asdict, dataclass, replace
import json
import os
from typing import Any, Dict, List

from .config import RunnerConfig
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
        return self._execute_task_request(task)

    def _create_agent(self, context: Any) -> Any:
        agent_kwargs: Dict[str, Any] = {
            "model": self._provider_model,
            "quiet_mode": True,
            "reasoning_config": self._reasoning_config,
            "enabled_toolsets": [],
            "skip_context_files": True,
            "skip_memory": False,
            "persist_session": True,
            "max_iterations": 1,
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

    def _execute_task_request(self, task: RunnerTaskRequest) -> HermesExecutionResult:
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
        try:
            agent = self._create_agent(context)
            tracker = self._session_manager.attach_tracking(agent, task, context)
            response = agent.run_conversation(
                task.prompt,
                system_message=task.system_message,
                conversation_history=context.conversation_history,
            )["final_response"]
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

    def execute_task(self, task: RunnerTaskRequest) -> HermesExecutionResult:
        if task.task_type not in ROLE_TASK_TYPES.get(self._role, {self._role}):
            return HermesExecutionResult(
                ok=False,
                message=f"Runner role {self._role} cannot execute task type {task.task_type}.",
                provider="policy",
                raw={"role": self._role, "task_type": task.task_type},
            )
        prompt = self._render_task_prompt(task)
        rendered_task = replace(task, prompt=prompt)
        result = self._execute_task_request(rendered_task)
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

    def _render_task_prompt(self, task: RunnerTaskRequest) -> str:
        parts = [
            f"Runner role: {self._role}",
            f"Task type: {task.task_type}",
            f"Repository: {task.repo}",
            f"Configured model: {self._configured_model}",
            f"Reasoning effort: {self._reasoning_effort}",
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
            parts.append(f"Allowed tools: {', '.join(task.allowed_tools)}")
        if task.tool_allowlist:
            parts.append(f"Tool allowlist: {', '.join(task.tool_allowlist)}")
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
        parts.append(
            "Return a JSON object with keys: visible_reasoning, reply_draft, final_answer, confidence, context_summary, self_critique, proposed_actions, knowledge_drafts, outcome_hypotheses."
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
        }


def first_non_empty(*values: str | None) -> str:
    for value in values:
        if value and value.strip():
            return value.strip()
    return ""
