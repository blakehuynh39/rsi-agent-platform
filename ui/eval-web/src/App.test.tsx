import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

const traceListResponse = {
  traces: [
    {
      trace_id: "trace-001",
      workflow_id: "wf-001",
      ingestion_id: "ing-001",
      workflow_kind: "question",
      status: "completed",
      thread_key: "slack:CENG:171000001.000100",
      started_at: "2026-04-11T12:00:00Z",
      event_count: 6,
      reasoning_count: 4,
      tool_call_count: 2,
      slack_action_count: 1,
      latest_eval: {
        run_id: "eval-001",
        verdict: "needs_improvement",
        score: 0.72,
        created_at: "2026-04-11T12:10:00Z",
        suite_name: "conversation"
      }
    }
  ]
};

const proposalListResponse = {
  proposals: [
    {
      id: "proposal-001",
      trace_id: "trace-001",
      title: "Improve trace rendering",
      category: "architecture",
      summary: "Split improvement-plane detail payloads into list and detail contracts.",
      status: "pr_open",
      candidate_key: "improvement-plane:detail-contracts",
      risk_tier: "medium",
      proposed_scope: "ui/eval-web and internal/improvementplane",
      active_slot_consuming: true,
      prior_similar_proposal_ids: null,
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
  candidates: null,
  proposal_memory: null,
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
  post_merge_replays: null,
  settings: {
    active_proposal_cap: 2,
    updated_at: "2026-04-11T12:00:00Z"
  }
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

const traceDetailResponse = {
  trace: {
    summary: {
      trace_id: "trace-001",
      ingestion_id: "ing-001",
      workflow_id: "wf-001",
      thread_key: "slack:CENG:171000001.000100",
      workflow_kind: "question",
      status: "completed",
      started_at: "2026-04-11T12:00:00Z",
      event_count: 6,
      artifact_count: 1,
      reasoning_step_count: 4,
      tool_call_count: 2,
      slack_action_count: 1
    },
    events: [],
    artifacts: [],
    reasoning: [],
    tool_calls: [],
    slack_actions: []
  },
  linked_eval_runs: [],
  judgments_by_eval_run: {},
  linked_proposals: [],
  ratings: [],
  improvement_notes: []
};

const proposalDetailResponse = {
  proposal: proposalListResponse.proposals[0],
  reviews: [],
  related_proposal_memory: [],
  repo_change_jobs: proposalListResponse.repo_change_jobs,
  pr_attempts: proposalListResponse.pr_attempts,
  post_merge_replays: [],
  linked_trace_summaries: [
    {
      trace_id: "trace-001",
      workflow_id: "wf-001",
      ingestion_id: "ing-001",
      workflow_kind: "question",
      status: "completed",
      thread_key: "slack:CENG:171000001.000100",
      started_at: "2026-04-11T12:00:00Z"
    }
  ],
  linked_eval_runs: []
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
        url.endsWith("/api/traces") ? traceListResponse :
        url.endsWith("/api/proposals") ? proposalListResponse :
        url.endsWith("/api/runtime") ? runtimeResponse :
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

  it("defaults to traces with no detail selection", async () => {
    renderApp();

    expect(await screen.findByText("trace-001")).toBeInTheDocument();
    expect(screen.getByText("Select a trace")).toBeInTheDocument();
    expect(window.location.search).toBe("");
  });

  it("opens trace detail and persists the trace id in the URL", async () => {
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /trace-001/i }));

    await waitFor(() => {
      expect(window.location.search).toContain("tab=traces");
      expect(window.location.search).toContain("trace=trace-001");
    });
    expect(await screen.findByText("Overview")).toBeInTheDocument();
  });

  it("shows proposal detail with a clickable PR link", async () => {
    renderApp();

    fireEvent.click(await screen.findByRole("button", { name: /proposals/i }));
    fireEvent.click(await screen.findByRole("button", { name: /improve trace rendering/i }));

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
