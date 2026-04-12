import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

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
    }
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
      body: "How does the agent think through trace rendering?",
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
  transcript_slice: conversationDetailResponse.transcript,
  linked_eval_runs: [],
  judgments_by_eval_run: {},
  feedback_records: [],
  linked_proposals: []
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
      status: "pr_open",
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
  proposal_memory: [],
  repo_change_jobs: [
    {
      id: "job-001",
      proposal_id: "proposal-001",
      status: "pr_open",
      repo: "rsi-agent-platform",
      branch_name: "codex/proposal-001",
      context_summary: "Detail payload refactor."
    }
  ],
  pr_attempts: [
    {
      id: "pr-001",
      proposal_id: "proposal-001",
      pr_url: "https://github.com/piplabs/rsi-agent-platform/pull/42",
      status: "pr_open",
      validation_status: "pending",
      created_at: "2026-04-11T12:30:00Z"
    }
  ],
  post_merge_replays: [],
  settings: {
    active_proposal_cap: 2,
    updated_at: "2026-04-11T12:00:00Z"
  }
};

const proposalDetailResponse = {
  proposal: proposalListResponse.proposals[0],
  reviews: [],
  related_proposal_memory: [],
  repo_change_jobs: proposalListResponse.repo_change_jobs,
  pr_attempts: proposalListResponse.pr_attempts,
  post_merge_replays: [],
  linked_trace_summaries: conversationDetailResponse.trace_attempts,
  linked_eval_runs: []
};

const runtimeResponse = {
  roles: [
    {
      role: "eval",
      reported_role: "eval",
      base_url: "http://runner-eval",
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
      hermes_available: true
    }
  ]
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

describe("App", () => {
  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    vi.spyOn(global, "fetch").mockImplementation((input) => {
      const url = String(input);
      const payload =
        url.endsWith("/api/conversations") ? conversationListResponse :
        url.endsWith("/api/cases") ? casesListResponse :
        url.endsWith("/api/proposals") ? proposalListResponse :
        url.endsWith("/api/runtime") ? runtimeResponse :
        url.endsWith("/api/conversations/conv-001") ? conversationDetailResponse :
        url.endsWith("/api/traces/trace-001") ? traceDetailResponse :
        url.endsWith("/api/proposals/proposal-001") ? proposalDetailResponse :
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
  });

  it("defaults to conversations with no detail selection", async () => {
    renderApp();

    expect(
      await screen.findByRole("button", { name: /Need help understanding trace rendering/i })
    ).toBeInTheDocument();
    expect(screen.getByText("Select a conversation")).toBeInTheDocument();
    expect(window.location.search).toBe("");
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
    fireEvent.click(screen.getByRole("button", { name: "reasoning" }));
    expect(screen.getByText("goal_framing")).toBeInTheDocument();
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
});
