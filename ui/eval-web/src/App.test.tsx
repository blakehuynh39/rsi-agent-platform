import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";
import type { JsonValue } from "./types";

const conversationListResponse = {
  conversations: [
    {
      conversation_id: "conv-001",
      source: "slack",
      external_key: "slack:dm:D123",
      title: "Need help understanding trace rendering",
      status: "active",
      active_case: {
        case_id: "case-001",
        conversation_id: "conv-001",
        kind: "question",
        intent: "question",
        title: "Need help understanding trace rendering",
        summary: "How does the agent think through trace rendering?",
        status: "active",
        assigned_bot: "arch",
        latest_trace_id: "trace-001",
        latest_trace_verdict: "pass",
        recurrence: 1,
        linked_proposal_ids: [],
        updated_at: "2026-04-11T12:00:00Z"
      },
      latest_message_at: "2026-04-11T12:00:00Z",
      latest_trace_verdict: "pass",
      open_trace_count: 1,
      proposal_count: 1
    },
    ...Array.from({ length: 8 }, (_, index) => {
      const sequence = index + 2;
      return {
        conversation_id: `conv-${String(sequence).padStart(3, "0")}`,
        source: "slack",
        external_key: `slack:thread:C123:${sequence}`,
        title: `Older conversation ${sequence}`,
        status: "active",
        latest_message_at: `2026-04-${String(11 - sequence).padStart(2, "0")}T12:00:00Z`,
        open_trace_count: 0,
        proposal_count: 0
      };
    })
  ]
};

const casesListResponse = {
  cases: [
    {
      case_id: "case-001",
      conversation_id: "conv-001",
      kind: "question",
      intent: "question",
      title: "Need help understanding trace rendering",
      summary: "How does the agent think through trace rendering?",
      status: "active",
      assigned_bot: "arch",
      latest_trace_id: "trace-001",
      latest_trace_verdict: "pass",
      recurrence: 1,
      linked_proposal_ids: [],
      updated_at: "2026-04-11T12:00:00Z"
    }
  ]
};

const workflowLineSummary = {
  case_id: "case-001",
  conversation_id: "conv-001",
  status: "completed",
  current_workflow_id: "wf-001",
  latest_workflow_id: "wf-001",
  attempt_count: 1,
  auto_retry_budget_remaining: 2,
  updated_at: "2026-04-11T12:05:00Z"
};

const workflowAttemptSummaries = [
  {
    workflow_id: "wf-001",
    trace_id: "trace-001",
    conversation_id: "conv-001",
    case_id: "case-001",
    workflow_kind: "question",
    status: "completed",
    trace_status: "completed",
    attempt_number: 1,
    retry_decision: "none",
    repair_attempted: false,
    repair_succeeded: false,
    created_at: "2026-04-11T12:01:00Z",
    updated_at: "2026-04-11T12:05:00Z",
    completed_at: "2026-04-11T12:05:00Z"
  }
];

const conversationDetailResponse = {
  conversation: {
    id: "conv-001",
    source: "slack",
    external_key: "slack:dm:D123",
    title: "Need help understanding trace rendering",
    status: "active",
    active_case_id: "case-001"
  },
  active_case: casesListResponse.cases[0],
  workflow_line: workflowLineSummary,
  workflow_attempts: workflowAttemptSummaries,
  cases: casesListResponse.cases,
  transcript: [
    {
      id: "entry-001",
      event_id: "evt-001",
      source: "slack",
      source_event_id: "171000001.000100",
      entry_type: "external_event",
      actor_id: "U123",
      actor_type: "user",
      body: "Sharing context with @U0ASDQKU3UL in #C0AKH5SNGKH - see <https://example.com/runbook|runbook> <!here>",
      metadata: {
        slack_user_names: {
          U0ASDQKU3UL: "blake"
        },
        slack_channel_names: {
          C0AKH5SNGKH: "depin-backend"
        }
      },
      created_at: "2026-04-11T12:00:00Z"
    }
  ],
  trace_attempts: [
    {
      trace_id: "trace-001",
      conversation_id: "conv-001",
      case_id: "case-001",
      workflow_kind: "question",
      status: "completed",
      thread_key: "slack:dm:D123",
      started_at: "2026-04-11T12:01:00Z",
      event_count: 6,
      reasoning_count: 4,
      tool_call_count: 2,
      slack_action_count: 1,
      latest_eval: {
        run_id: "eval-001",
        verdict: "pass",
        score: 0.92,
        created_at: "2026-04-11T12:10:00Z",
        suite_name: "conversation"
      }
    }
  ],
  action_intents: [],
  action_results: [],
  outcomes: [],
  knowledge_entries: [],
  linked_proposals: []
};

const traceDetailResponse = {
  trace: {
    summary: {
      trace_id: "trace-001",
      conversation_id: "conv-001",
      case_id: "case-001",
      workflow_kind: "question",
      status: "completed",
      thread_key: "slack:dm:D123",
      started_at: "2026-04-11T12:01:00Z",
      event_count: 6,
      artifact_count: 1,
      reasoning_step_count: 4,
      tool_call_count: 2,
      slack_action_count: 1
    },
    events: [],
    artifacts: [],
    reasoning: [
      {
        id: "reason-001",
        step_type: "goal_framing",
        summary: "Frame the user question before collecting evidence.",
        evidence_refs: [],
        alternatives: [],
        confidence: 0.92,
        decision: "inspect trace data",
        created_at: "2026-04-11T12:02:00Z"
      }
    ],
    tool_calls: [],
    slack_actions: []
  },
  conversation: conversationListResponse.conversations[0],
  case: casesListResponse.cases[0],
  workflow_line: workflowLineSummary,
  workflow_attempts: workflowAttemptSummaries,
  transcript_slice: conversationDetailResponse.transcript,
  linked_eval_runs: [],
  judgments_by_eval_run: {},
  action_intents: [
    {
      id: "action-001",
      owner_plane: "control",
      trace_id: "trace-001",
      kind: "slack_post",
      status: "succeeded",
      created_at: "2026-04-11T12:03:00Z",
      updated_at: "2026-04-11T12:03:00Z"
    }
  ],
  action_results: [
    {
      id: "action-result-001",
      action_intent_id: "action-001",
      attempt_number: 1,
      executor: "native-hermes",
      provider: "slack",
      status: "succeeded",
      started_at: "2026-04-11T12:03:00Z",
      completed_at: "2026-04-11T12:03:01Z"
    }
  ],
  outcomes: [
    {
      id: "outcome-001",
      source: "operator",
      trace_id: "trace-001",
      outcome_type: "answer_quality",
      verdict: "positive",
      score: 1,
      summary: "The answer resolved the thread.",
      recorded_at: "2026-04-11T12:15:00Z"
    }
  ],
  knowledge_entries: [
    {
      id: "knowledge-001",
      tier: "working",
      kind: "fact",
      scope_type: "case",
      scope_id: "case-001",
      title: "Trace rendering note",
      summary: "Keep evidence grouped by trace attempt.",
      status: "draft",
      confidence: 0.82,
      source_type: "agent",
      created_at: "2026-04-11T12:16:00Z",
      updated_at: "2026-04-11T12:16:00Z"
    }
  ],
  feedback_records: [],
  linked_proposals: [],
  harness_executions: [
    {
      id: "hexec-001",
      trace_id: "trace-001",
      role: "prod",
      session_scope_kind: "conversation",
      session_scope_id: "conv-001",
      hermes_session_id: "rsi-prod-conversation-123",
      memory_backend: "honcho",
      memory_reads: [
        {
          kind: "session_history",
          summary: "user: How does the agent think through trace rendering?"
        }
      ],
      memory_writes: [
        {
          kind: "memory_sync_assistant",
          summary: "Explained trace rendering and evidence grouping."
        }
      ],
      created_at: "2026-04-11T12:03:00Z"
    }
  ]
};

const proposalListResponse = {
  proposals: [
    {
      id: "proposal-001",
      trace_id: "trace-001",
      conversation_id: "conv-001",
      case_id: "case-001",
      origin_trace_id: "trace-001",
      evidence_trace_ids: ["trace-001"],
      title: "Improve trace rendering",
      category: "architecture",
      summary: "Split detail payloads into conversation, case, and trace evidence objects.",
      status: "in_progress",
      repo_change_status: "validation_pending",
      pr_status: "open",
      pr_url: "https://github.com/piplabs/rsi-agent-platform/pull/42",
      candidate_key: "improvement-plane:detail-contracts",
      risk_tier: "medium",
      proposed_scope: "ui/eval-web and internal/improvementplane",
      active_slot_consuming: true,
      prior_similar_proposal_ids: [],
      new_evidence_since_last_rejection: true,
      created_at: "2026-04-11T12:20:00Z"
    }
  ],
  proposal_slots: {
    cap: 2,
    active: 1,
    available: 1,
    active_proposal_ids: ["proposal-001"],
    stale_proposal_ids: []
  },
  candidates: [],
  settings: {
    active_proposal_cap: 2,
    updated_at: "2026-04-11T12:00:00Z"
  }
};

const proposalDetailResponse = {
  proposal: proposalListResponse.proposals[0],
  current_phase: {
    attempt_id: "attempt-001",
    attempt_state: "observing_ci",
    reconcile_status: "healthy",
    reconciliation_needed: false
  },
  attempts: [
    {
      id: "attempt-001",
      proposal_id: "proposal-001",
      candidate_key: "shared-store:action_result_pkey",
      attempt_number: 1,
      target_layer: "repo_change",
      target_kind: "repo",
      target_ref: "rsi-agent-platform",
      trigger: "proposal_approved",
      state: "observing_ci",
      branch_name: "codex/proposal-001",
      diff_summary: "Update persistence contract.",
      changed_files: ["internal/store/postgres.go"],
      validation_summary: "make test passed in workspace",
      change_plan: "Patch direct write path and keep action result identity append-only.",
      validation_plan: "make test",
      created_at: "2026-04-11T12:24:00Z",
      updated_at: "2026-04-11T12:30:00Z"
    }
  ],
  workspace_sessions: [
    {
      id: "workspace-001",
      attempt_id: "attempt-001",
      proposal_id: "proposal-001",
      operation_id: "attempt-001:workspace:001",
      generation: 1,
      repo: "rsi-agent-platform",
      base_ref: "main",
      branch_name: "codex/proposal-001",
      namespace: "rsi-platform",
      job_name: "rsi-workspace-attempt-001",
      pod_name: "rsi-workspace-attempt-001-pod",
      status: "ready",
      allowed_path_globs: ["internal/**", "cmd/**"],
      diff_summary: "1 file changed",
      created_at: "2026-04-11T12:24:00Z",
      updated_at: "2026-04-11T12:30:00Z"
    }
  ],
  effects: [
    {
      id: "eff-001",
      machine_kind: "proposal_line",
      aggregate_id: "proposal-001",
      effect_kind: "open_draft_pr",
      status: "completed",
      idempotency_key: "proposal-001:pr_open",
      retry_count: 0,
      attempt_id: "attempt-001",
      started_at: "2026-04-11T12:28:00Z",
      completed_at: "2026-04-11T12:30:00Z"
    }
  ],
  reviews: [],
  related_proposal_memory: [],
  validation_runs: [
    {
      id: "validation-001",
      proposal_id: "proposal-001",
      attempt_id: "attempt-001",
      workspace_id: "workspace-001",
      repo: "rsi-agent-platform",
      branch_name: "codex/proposal-001",
      command: "make test",
      status: "passed",
      validation_ref: "rsi-platform/rsi-sandbox-trace-001",
      sandbox_namespace: "rsi-platform",
      sandbox_job_name: "rsi-sandbox-trace-001",
      sandbox_pod_name: "rsi-sandbox-trace-001",
      log_artifact_id: "",
      created_at: "2026-04-11T12:25:00Z",
      updated_at: "2026-04-11T12:30:00Z"
    }
  ],
  pr_attempts: [
    {
      id: "pr-001",
      proposal_id: "proposal-001",
      pr_url: "https://github.com/piplabs/rsi-agent-platform/pull/42",
      status: "open",
      validation_status: "passed",
      created_at: "2026-04-11T12:30:00Z"
    }
  ],
  post_merge_replays: [],
  linked_trace_summaries: conversationDetailResponse.trace_attempts,
  linked_eval_runs: [],
  action_intents: traceDetailResponse.action_intents,
  action_results: traceDetailResponse.action_results,
  outcomes: traceDetailResponse.outcomes,
  knowledge_entries: traceDetailResponse.knowledge_entries,
  harness_executions: traceDetailResponse.harness_executions
};

const knowledgeListResponse = {
  knowledge_entries: traceDetailResponse.knowledge_entries
};

const knowledgeDetailResponse = {
  knowledge_entry: traceDetailResponse.knowledge_entries[0],
  evidence_links: [
    {
      knowledge_entry_id: "knowledge-001",
      evidence_type: "trace",
      evidence_id: "trace-001",
      relevance_summary: "Derived from the successful question trace.",
      evidence_ref: {
        kind: "trace",
        ref: "trace-001",
        summary: "question"
      }
    }
  ],
  reviews: []
};

const runtimeResponse = {
  roles: [
    {
      role: "prod",
      reported_role: "prod",
      base_url: "http://runner-prod",
      timeout_seconds: 60,
      status: "ok",
      backend: "hermes-aiagent",
      provider: "openai",
      model: "openai/gpt-5.4",
      provider_model: "gpt-5.4",
      api_mode: "codex_responses",
      reasoning_effort: "xhigh",
      available: true,
      healthy: true,
      openai_configured: true,
      hermes_available: true,
      persistence_enabled: true,
      hermes_home: "/var/lib/hermes",
      session_db_path: "/var/lib/hermes/state.db",
      memory_backend: "honcho",
      honcho_configured: true,
      honcho_available: true,
      harness_profile_id: "harness-profile-prod",
      active_overlay_version: "baseline"
    },
    {
      role: "eval",
      reported_role: "eval",
      base_url: "http://runner-eval",
      timeout_seconds: 120,
      status: "ok",
      backend: "hermes-aiagent",
      provider: "openai",
      model: "openai/gpt-5.4",
      provider_model: "gpt-5.4",
      api_mode: "codex_responses",
      reasoning_effort: "xhigh",
      available: true,
      healthy: true,
      openai_configured: true,
      hermes_available: true,
      persistence_enabled: true,
      hermes_home: "/var/lib/hermes",
      session_db_path: "/var/lib/hermes/state.db",
      memory_backend: "honcho",
      honcho_configured: true,
      honcho_available: true,
      harness_profile_id: "harness-profile-eval",
      active_overlay_version: "baseline"
    }
  ]
};

const harnessResponse = {
  profiles: [
    {
      id: "harness-profile-prod",
      role: "prod",
      name: "Production conversational",
      description: "Baseline prod role profile.",
      model: "openai/gpt-5.4",
      reasoning_effort: "xhigh",
      prompt_fragments: ["Use explicit visible reasoning."],
      few_shot_snippets: [],
      tool_preference_order: ["repo.context", "knowledge.context"],
      retrieval_bias: "canonical_first",
      reasoning_verbosity: "verbose",
      memory_read_enabled: true,
      memory_write_enabled: true,
      created_at: "2026-04-11T12:00:00Z",
      updated_at: "2026-04-11T12:00:00Z"
    }
  ],
  overlays: [],
  experiments: [],
  session_bindings: [
    {
      role: "prod",
      scope_kind: "conversation",
      scope_id: "conv-001",
      hermes_session_id: "rsi-prod-conversation-123",
      memory_backend: "honcho",
      harness_profile_id: "harness-profile-prod",
      effective_overlay_version: "baseline",
      last_used_at: "2026-04-11T12:03:00Z",
      created_at: "2026-04-11T12:00:00Z",
      updated_at: "2026-04-11T12:03:00Z"
    }
  ],
  executions: traceDetailResponse.harness_executions,
  roles: runtimeResponse.roles
};

function renderApp() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false
      }
    }
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  );
}

class MockEventSource {
  static instances: MockEventSource[] = [];
  url: string;
  onopen: (() => void) | null = null;
  onerror: (() => void) | null = null;
  private listeners: Record<string, ((event: MessageEvent) => void)[]> = {};

  constructor(url: string) {
    this.url = url;
    MockEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: EventListener) {
    this.listeners[type] = [...(this.listeners[type] || []), listener as (event: MessageEvent) => void];
  }

  removeEventListener(type: string, listener: EventListener) {
    this.listeners[type] = (this.listeners[type] || []).filter((item) => item !== listener);
  }

  emit(type: string, data: JsonValue) {
    const event = { data: JSON.stringify(data) } as MessageEvent;
    for (const listener of this.listeners[type] || []) {
      listener(event);
    }
  }

  close() {}
}

describe("App", () => {
  beforeEach(() => {
    MockEventSource.instances = [];
    window.history.replaceState({}, "", "/");
    vi.spyOn(window, "confirm").mockReturnValue(true);
    vi.spyOn(globalThis, "fetch").mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      const path = new URL(url, window.location.origin).pathname;
      const method = init?.method ?? "GET";
      const payload =
        method === "POST" && path === "/api/app-data/reset" ? {
          backend: "postgres",
          reset_at: "2026-04-11T12:45:00Z",
          truncated_tables: ["public.event_envelope", "honcho.queue"],
          preserved_tables: ["public.rsi_schema_migrations", "honcho.alembic_version"]
        } :
        path === "/api/conversations" ? conversationListResponse :
        path === "/api/cases" ? casesListResponse :
        path === "/api/proposals" ? proposalListResponse :
        path === "/api/knowledge" ? knowledgeListResponse :
        path === "/api/runtime" ? runtimeResponse :
        path === "/api/harness" ? harnessResponse :
        path === "/api/conversations/conv-001" ? conversationDetailResponse :
        path === "/api/traces/trace-001" ? traceDetailResponse :
        path === "/api/traces/trace-001/ledger" ? {
          events: [
            {
              id: "xled-older-0",
              execution_id: "hexec-live",
              trace_id: "trace-001",
              workflow_id: "wf-001",
              phase_id: "investigate",
              kind: "model.reasoning.delta",
              status: "streaming",
              seq: 0,
              recorded_at: "2026-04-11T12:01:59Z",
              payload: { role: "prod", delta: "older" }
            }
          ],
          paging: { has_more: false, next_before: "" }
        } :
        path === "/api/proposals/proposal-001" ? proposalDetailResponse :
        path === "/api/knowledge/knowledge-001" ? knowledgeDetailResponse :
        undefined;

      if (!payload) {
        return Promise.resolve(new Response("not found", { status: 404 }));
      }
      return Promise.resolve(
        new Response(JSON.stringify(payload), {
          status: 200,
          headers: { "Content-Type": "application/json" }
        })
      );
    });
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("defaults to conversations with no detail selection", async () => {
    renderApp();

    expect(
      await screen.findByRole("button", { name: /Need help understanding trace rendering/i })
    ).toBeInTheDocument();
    expect(screen.getByText("Select a conversation")).toBeInTheDocument();
    expect(window.location.search).toBe("");
  });

  it("paginates conversation cards without losing the selected route state", async () => {
    renderApp();

    expect(await screen.findByRole("button", { name: /Need help understanding trace rendering/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Older conversation 7/i })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Next conversations page" }));

    expect(await screen.findByRole("button", { name: /Older conversation 7/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Need help understanding trace rendering/i })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Previous conversations page" }));

    expect(await screen.findByRole("button", { name: /Need help understanding trace rendering/i })).toBeInTheDocument();
  });

  it("opens conversation trace detail and persists conversation and trace in the URL", async () => {
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /Need help understanding trace rendering/i }));
    fireEvent.click(await screen.findByRole("button", { name: /trace-001/i }));

    await waitFor(() => {
      expect(window.location.search).toContain("tab=conversations");
      expect(window.location.search).toContain("conversation=conv-001");
      expect(window.location.search).toContain("trace=trace-001");
    });
    expect(await screen.findByText("Trace inspector")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Evidence" }));
    expect(screen.getByText("goal_framing")).toBeInTheDocument();
  });

  it("streams live trace ledger events into the Live tab", async () => {
    vi.stubGlobal("EventSource", MockEventSource);
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /Need help understanding trace rendering/i }));
    fireEvent.click(await screen.findByRole("button", { name: /trace-001/i }));
    expect(await screen.findByText("Trace inspector")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Raw" }));

    await waitFor(() => expect(MockEventSource.instances.length).toBe(1));
    expect(MockEventSource.instances[0].url).toBe("/api/traces/trace-001/stream?scope=main&limit=100");
    MockEventSource.instances[0].emit("ledger", {
      id: "xled-live-1",
      execution_id: "hexec-live",
      trace_id: "trace-001",
      workflow_id: "wf-001",
      phase_id: "investigate",
      kind: "model.reasoning.delta",
      status: "streaming",
      seq: 1,
      recorded_at: "2026-04-11T12:02:00Z",
      payload: { delta: "reading " }
    });
    MockEventSource.instances[0].emit("ledger", {
      id: "xled-live-2",
      execution_id: "hexec-live",
      trace_id: "trace-001",
      workflow_id: "wf-001",
      phase_id: "investigate",
      kind: "model.reasoning.delta",
      status: "streaming",
      seq: 2,
      recorded_at: "2026-04-11T12:02:01Z",
      payload: { delta: "repo files" }
    });
    MockEventSource.instances[0].emit("ledger", {
      id: "xled-live-3",
      execution_id: "hexec-live",
      trace_id: "trace-001",
      workflow_id: "wf-001",
      phase_id: "investigate",
      kind: "tool.call.started",
      status: "running",
      seq: 3,
      recorded_at: "2026-04-11T12:02:02Z",
      payload: {
        tool_call_id: "tool-1",
        tool_name: "repo_read_file"
      }
    });
    MockEventSource.instances[0].emit("ledger", {
      id: "xled-live-4",
      execution_id: "hexec-live",
      trace_id: "trace-001",
      workflow_id: "wf-001",
      phase_id: "investigate",
      kind: "tool.call.progress",
      status: "running",
      seq: 4,
      recorded_at: "2026-04-11T12:02:03Z",
      payload: {
        tool_call_id: "tool-1",
        tool_name: "repo_read_file"
      }
    });
    MockEventSource.instances[0].emit("ledger", {
      id: "xled-live-5",
      execution_id: "hexec-live",
      trace_id: "trace-001",
      workflow_id: "wf-001",
      phase_id: "investigate",
      kind: "tool.call.completed",
      status: "completed",
      seq: 5,
      recorded_at: "2026-04-11T12:02:04Z",
      payload: {
        tool_call_id: "tool-1",
        tool_name: "repo_read_file",
        result: JSON.stringify({
          status: "ok",
          summary: "Repository file README.md loaded."
        })
      }
    });

    expect(await screen.findByText("Reasoning stream")).toBeInTheDocument();
    expect(screen.getByText("model.reasoning.delta · 2 chunks")).toBeInTheDocument();
    expect(screen.getByText("reading repo files")).toBeInTheDocument();
    expect(screen.getByText("repo_read_file")).toBeInTheDocument();
    expect(screen.getByText("3 updates")).toBeInTheDocument();
    expect(screen.getByText("Repository file README.md loaded. · status: ok")).toBeInTheDocument();
  });

  it("probes ledger history before showing the Load older control", async () => {
    vi.stubGlobal("EventSource", MockEventSource);
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /Need help understanding trace rendering/i }));
    fireEvent.click(await screen.findByRole("button", { name: /trace-001/i }));
    expect(await screen.findByText("Trace inspector")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Raw" }));

    await waitFor(() => expect(MockEventSource.instances.length).toBe(1));
    for (let seq = 1; seq <= 100; seq += 1) {
      MockEventSource.instances[0].emit("ledger", {
        id: `xled-page-${seq}`,
        execution_id: "hexec-live",
        trace_id: "trace-001",
        workflow_id: "wf-001",
        phase_id: "investigate",
        kind: "model.reasoning.delta",
        status: "streaming",
        seq,
        recorded_at: `2026-04-11T12:02:${String(seq % 60).padStart(2, "0")}Z`,
        payload: { role: "prod", delta: `chunk ${seq} ` }
      });
    }

    await waitFor(() => {
      expect(
        vi.mocked(globalThis.fetch).mock.calls.some(([input]) =>
          String(input).includes("/api/traces/trace-001/ledger?scope=main&limit=1&before=xled-page-1")
        )
      ).toBe(true);
    });
    expect(await screen.findByRole("button", { name: "Load older" })).toBeInTheDocument();
  });

  it("renders Slack transcript entries with readable Slack names and channel labels", async () => {
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /Need help understanding trace rendering/i }));

    const transcriptEntries = await screen.findAllByText((_, element) =>
      element !== null &&
      element.classList.contains("detail-copy") &&
      element.textContent === "Sharing context with @blake in #depin-backend - see runbook @here"
    );
    expect(transcriptEntries.length).toBeGreaterThan(0);
    expect(screen.queryByText(/@U0ASDQKU3UL/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/#C0AKH5SNGKH/i)).not.toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: "runbook" })[0]).toHaveAttribute("href", "https://example.com/runbook");
  });

  it("shows proposal detail with a clickable PR link", async () => {
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /Proposals/i }));
    fireEvent.click(await screen.findByRole("button", { name: /Improve trace rendering/i }));

    await waitFor(() => {
      expect(window.location.search).toContain("tab=proposals");
      expect(window.location.search).toContain("proposal=proposal-001");
    });
    expect(await screen.findByRole("link", { name: "Open PR" })).toHaveAttribute(
      "href",
      "https://github.com/piplabs/rsi-agent-platform/pull/42"
    );
  });

  it("shows harness role detail with Hermes persistence metadata", async () => {
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /Harness/i }));
    fireEvent.click(await screen.findByRole("button", { name: /prod/i }));

    await waitFor(() => {
      expect(window.location.search).toContain("tab=harness");
      expect(window.location.search).toContain("role=prod");
    });
    expect(await screen.findByText("Persistent Hermes role agents")).toBeInTheDocument();
    expect(screen.getByText("rsi-prod-conversation-123")).toBeInTheDocument();
  });

  it("shows knowledge detail and trace action evidence", async () => {
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /Need help understanding trace rendering/i }));
    fireEvent.click(await screen.findByRole("button", { name: /trace-001/i }));
    fireEvent.click(await screen.findByRole("button", { name: "Evidence" }));

    expect(await screen.findByText("slack_post")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Knowledge/i }));
    fireEvent.click(await screen.findByRole("button", { name: /Trace rendering note/i }));

    expect(await screen.findByText("Evidence links")).toBeInTheDocument();
    expect(screen.getByText("Derived from the successful question trace.")).toBeInTheDocument();
  });

  it("posts reset app data after confirmation and returns to the conversations tab", async () => {
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /Proposals/i }));
    await waitFor(() => {
      expect(window.location.search).toContain("tab=proposals");
    });

    fireEvent.click(screen.getByRole("button", { name: "Reset app data" }));

    await waitFor(() => {
      expect(window.confirm).toHaveBeenCalled();
      expect(globalThis.fetch).toHaveBeenCalledWith(
        "/api/app-data/reset",
        expect.objectContaining({ method: "POST" })
      );
      expect(window.location.search).toContain("tab=conversations");
    });
    expect(await screen.findByText(/Reset finished/i)).toBeInTheDocument();
  });
});
