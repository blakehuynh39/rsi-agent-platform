import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ReactNode, useEffect, useMemo, useState } from "react";

type NullableList<T> = T[] | null;
type TabKey = "traces" | "proposals";
type ProposalSegment = "active" | "candidates" | "history";

type TraceEvalSummary = {
  run_id: string;
  verdict: string;
  score: number;
  created_at: string;
  suite_name: string;
};

type TraceListItem = {
  trace_id: string;
  workflow_id: string;
  ingestion_id: string;
  workflow_kind: string;
  status: string;
  thread_key: string;
  started_at: string;
  event_count: number;
  reasoning_count: number;
  tool_call_count: number;
  slack_action_count: number;
  latest_eval?: TraceEvalSummary;
};

type TraceEvent = {
  trace_id: string;
  workflow_id?: string;
  plane: string;
  service: string;
  actor: string;
  event_type: string;
  status: string;
  description?: string;
  started_at: string;
  ended_at?: string;
};

type EvidenceRef = {
  kind: string;
  ref: string;
  summary?: string;
};

type ReasoningStep = {
  id: string;
  step_type: string;
  summary: string;
  evidence_refs?: NullableList<EvidenceRef>;
  alternatives?: NullableList<string>;
  confidence?: number;
  decision?: string;
  created_at: string;
};

type ToolCallRecord = {
  id: string;
  tool_name: string;
  tool_call_id: string;
  summary?: string;
  approval_state?: string;
  interpretation_summary?: string;
  status?: string;
  created_at: string;
};

type SlackActionRecord = {
  id: string;
  channel_id?: string;
  thread_ts?: string;
  draft_body?: string;
  final_body?: string;
  policy_verdict?: string;
  send_status?: string;
  created_at: string;
};

type Artifact = {
  id: string;
  kind: string;
  url: string;
  source: string;
};

type Trace = {
  summary: {
    trace_id: string;
    ingestion_id: string;
    workflow_id: string;
    thread_key: string;
    workflow_kind: string;
    status: string;
    last_verdict?: string;
    started_at: string;
    event_count: number;
    artifact_count: number;
    reasoning_step_count: number;
    tool_call_count: number;
    slack_action_count: number;
  };
  events: NullableList<TraceEvent>;
  artifacts: NullableList<Artifact>;
  reasoning: NullableList<ReasoningStep>;
  tool_calls: NullableList<ToolCallRecord>;
  slack_actions: NullableList<SlackActionRecord>;
};

type EvalRun = {
  id: string;
  trace_id: string;
  suite_name: string;
  trigger: string;
  overall_score: number;
  overall_verdict: string;
  created_at: string;
};

type EvalJudgment = {
  id: string;
  layer: string;
  category: string;
  score: number;
  passed: boolean;
  rationale: string;
};

type HumanRating = {
  verdict: string;
  score: number;
  notes: string;
  reviewer_id: string;
  created_at: string;
};

type ImprovementNote = {
  category: string;
  note: string;
  suggested_owner: string;
  created_by: string;
  created_at: string;
};

type TraceDetailResponse = {
  trace: Trace;
  linked_eval_runs: NullableList<EvalRun>;
  judgments_by_eval_run: Record<string, NullableList<EvalJudgment>>;
  linked_proposals: NullableList<Proposal>;
  ratings: NullableList<HumanRating>;
  improvement_notes: NullableList<ImprovementNote>;
};

type Candidate = {
  id: string;
  candidate_key: string;
  subsystem: string;
  failure_mode: string;
  intervention_type: string;
  status: string;
  severity: string;
  recurrence_count: number;
  priority_score: number;
  confidence_score: number;
  latest_trace_id?: string;
  new_evidence_since_last_rejection: boolean;
  prior_similar_proposal_ids: NullableList<string>;
};

type Proposal = {
  id: string;
  trace_id: string;
  title: string;
  category: string;
  summary: string;
  status: string;
  reviewer?: string;
  candidate_key: string;
  source_eval_ids?: NullableList<string>;
  risk_tier?: string;
  proposed_scope?: string;
  evidence_artifact_ids?: NullableList<string>;
  active_slot_consuming: boolean;
  review_deadline?: string;
  prior_similar_proposal_ids: NullableList<string>;
  new_evidence_since_last_rejection: boolean;
  created_at: string;
  reviews?: NullableList<ProposalReview>;
};

type ProposalReview = {
  proposal_id: string;
  decision: string;
  rationale: string;
  reviewer_id: string;
  failure_class?: string;
  failure_classes?: NullableList<string>;
  created_at: string;
};

type ProposalMemory = {
  id: string;
  proposal_id: string;
  candidate_key: string;
  hypothesis: string;
  diff_summary: string;
  review_rationale: string;
  disposition: string;
  disposition_reason?: string;
  failure_class?: string;
  failure_classes?: NullableList<string>;
  created_at: string;
};

type ProposalSlots = {
  cap: number;
  active: number;
  available: number;
  active_proposal_ids: NullableList<string>;
  stale_proposal_ids: NullableList<string>;
};

type RepoChangeJob = {
  id: string;
  proposal_id: string;
  status: string;
  repo: string;
  branch_name: string;
  context_summary: string;
};

type PRAttempt = {
  id: string;
  proposal_id: string;
  pr_url?: string;
  status: string;
  validation_status: string;
  created_at: string;
};

type PostMergeReplay = {
  id: string;
  proposal_id: string;
  trace_id: string;
  baseline_score: number;
  candidate_score: number;
  improved: boolean;
  created_at: string;
};

type ImprovementSettings = {
  active_proposal_cap: number;
  updated_at: string;
};

type ProposalResponse = {
  proposals: NullableList<Proposal>;
  proposal_slots: ProposalSlots;
  candidates: NullableList<Candidate>;
  proposal_memory: NullableList<ProposalMemory>;
  repo_change_jobs: NullableList<RepoChangeJob>;
  pr_attempts: NullableList<PRAttempt>;
  post_merge_replays: NullableList<PostMergeReplay>;
  settings: ImprovementSettings;
};

type ProposalDetailResponse = {
  proposal: Proposal;
  reviews: NullableList<ProposalReview>;
  related_proposal_memory: NullableList<ProposalMemory>;
  repo_change_jobs: NullableList<RepoChangeJob>;
  pr_attempts: NullableList<PRAttempt>;
  post_merge_replays: NullableList<PostMergeReplay>;
  linked_trace_summaries: NullableList<TraceListItem>;
  linked_eval_runs: NullableList<EvalRun>;
};

type RuntimeRole = {
  role: string;
  reported_role?: string;
  base_url: string;
  status: string;
  backend: string;
  provider: string;
  model: string;
  provider_model?: string;
  api_mode?: string;
  reasoning_effort: string;
  available: boolean;
  healthy: boolean;
  openai_configured: boolean;
  hermes_available: boolean;
  error?: string;
};

type RuntimeResponse = {
  roles: NullableList<RuntimeRole>;
};

type ViewState = {
  tab: TabKey;
  trace?: string;
  proposal?: string;
};

function listOrEmpty<T>(items: NullableList<T> | undefined): T[] {
  return items ?? [];
}

function recordOrEmpty<T>(items: Record<string, T> | undefined): Record<string, T> {
  return items ?? {};
}

function getJSON<T>(url: string): Promise<T> {
  return fetch(url).then(async (response) => {
    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }
    return response.json();
  });
}

function postJSON<T>(url: string, body: Record<string, unknown>): Promise<T> {
  return fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  }).then(async (response) => {
    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }
    return response.json();
  });
}

function readViewState(): ViewState {
  const params = new URLSearchParams(window.location.search);
  const tab = params.get("tab") === "proposals" ? "proposals" : "traces";
  const trace = params.get("trace") || undefined;
  const proposal = params.get("proposal") || undefined;
  return {
    tab,
    trace: tab === "traces" ? trace : undefined,
    proposal: tab === "proposals" ? proposal : undefined
  };
}

function writeViewState(next: ViewState) {
  const params = new URLSearchParams();
  params.set("tab", next.tab);
  if (next.tab === "traces" && next.trace) {
    params.set("trace", next.trace);
  }
  if (next.tab === "proposals" && next.proposal) {
    params.set("proposal", next.proposal);
  }
  const query = params.toString();
  const target = `${window.location.pathname}${query ? `?${query}` : ""}`;
  window.history.replaceState({}, "", target);
}

function formatTime(value?: string) {
  if (!value) return "Unknown";
  const date = new Date(value);
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit"
  }).format(date);
}

function scoreBadge(score?: number) {
  if (typeof score !== "number") return "n/a";
  return score.toFixed(2);
}

const ACTIVE_PROPOSAL_STATES = new Set([
  "pending_review",
  "approved",
  "repo_change_queued",
  "repo_change_running",
  "validation_pending",
  "pr_open"
]);

export function App() {
  const queryClient = useQueryClient();
  const [viewState, setViewState] = useState<ViewState>(() => readViewState());
  const [proposalSegment, setProposalSegment] = useState<ProposalSegment>("active");
  const [proposalCapInput, setProposalCapInput] = useState("2");
  const [ratingScore, setRatingScore] = useState("3");
  const [ratingVerdict, setRatingVerdict] = useState("partial");
  const [ratingNotes, setRatingNotes] = useState("");
  const [improvementNote, setImprovementNote] = useState("");
  const [proposalRationale, setProposalRationale] = useState("");

  useEffect(() => {
    const handlePopState = () => setViewState(readViewState());
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  const navigate = (next: ViewState) => {
    writeViewState(next);
    setViewState(next);
  };

  const tracesQuery = useQuery({
    queryKey: ["traces"],
    queryFn: () => getJSON<{ traces: TraceListItem[] }>("/api/traces")
  });

  const proposalsQuery = useQuery({
    queryKey: ["proposals"],
    queryFn: () => getJSON<ProposalResponse>("/api/proposals")
  });

  const runtimeQuery = useQuery({
    queryKey: ["runtime"],
    queryFn: () => getJSON<RuntimeResponse>("/api/runtime")
  });

  const selectedTraceId = viewState.tab === "traces" ? viewState.trace : undefined;
  const selectedProposalId = viewState.tab === "proposals" ? viewState.proposal : undefined;

  const traceDetailQuery = useQuery({
    queryKey: ["trace", selectedTraceId],
    queryFn: () => getJSON<TraceDetailResponse>(`/api/traces/${selectedTraceId}`),
    enabled: Boolean(selectedTraceId)
  });

  const proposalDetailQuery = useQuery({
    queryKey: ["proposal", selectedProposalId],
    queryFn: () => getJSON<ProposalDetailResponse>(`/api/proposals/${selectedProposalId}`),
    enabled: Boolean(selectedProposalId)
  });

  const traces = listOrEmpty(tracesQuery.data?.traces);
  const proposals = listOrEmpty(proposalsQuery.data?.proposals);
  const candidates = listOrEmpty(proposalsQuery.data?.candidates);
  const prAttempts = listOrEmpty(proposalsQuery.data?.pr_attempts);
  const repoChangeJobs = listOrEmpty(proposalsQuery.data?.repo_change_jobs);
  const runtimeRoles = listOrEmpty(runtimeQuery.data?.roles);

  useEffect(() => {
    const settingValue = proposalsQuery.data?.settings?.active_proposal_cap;
    if (typeof settingValue === "number") {
      setProposalCapInput(String(settingValue));
    }
  }, [proposalsQuery.data?.settings?.active_proposal_cap]);

  useEffect(() => {
    if (selectedTraceId && !traces.some((trace) => trace.trace_id === selectedTraceId)) {
      navigate({ tab: "traces" });
    }
  }, [selectedTraceId, traces]);

  useEffect(() => {
    if (selectedProposalId && !proposals.some((proposal) => proposal.id === selectedProposalId)) {
      navigate({ tab: "proposals" });
    }
  }, [selectedProposalId, proposals]);

  useEffect(() => {
    if (!selectedProposalId) return;
    const selectedProposal = proposals.find((proposal) => proposal.id === selectedProposalId);
    if (!selectedProposal) return;
    setProposalSegment(ACTIVE_PROPOSAL_STATES.has(selectedProposal.status) ? "active" : "history");
  }, [selectedProposalId, proposals]);

  const activeProposals = useMemo(
    () => proposals.filter((proposal) => ACTIVE_PROPOSAL_STATES.has(proposal.status)),
    [proposals]
  );
  const historyProposals = useMemo(
    () => proposals.filter((proposal) => !ACTIVE_PROPOSAL_STATES.has(proposal.status)),
    [proposals]
  );

  const refreshEverything = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["traces"] }),
      queryClient.invalidateQueries({ queryKey: ["trace"] }),
      queryClient.invalidateQueries({ queryKey: ["proposals"] }),
      queryClient.invalidateQueries({ queryKey: ["proposal"] }),
      queryClient.invalidateQueries({ queryKey: ["runtime"] })
    ]);
  };

  const evaluateMutation = useMutation({
    mutationFn: () => postJSON(`/api/traces/${selectedTraceId}/evaluate`, {}),
    onSuccess: refreshEverything
  });

  const replayMutation = useMutation({
    mutationFn: () => postJSON(`/api/traces/${selectedTraceId}/replay`, { requested_by: "ui-operator" }),
    onSuccess: refreshEverything
  });

  const ratingMutation = useMutation({
    mutationFn: () =>
      postJSON(`/api/traces/${selectedTraceId}/ratings`, {
        score: Number(ratingScore),
        verdict: ratingVerdict,
        labels: ["ui-review"],
        notes: ratingNotes,
        reviewer_id: "ui-operator"
      }),
    onSuccess: async () => {
      setRatingNotes("");
      await refreshEverything();
    }
  });

  const noteMutation = useMutation({
    mutationFn: () =>
      postJSON(`/api/traces/${selectedTraceId}/notes`, {
        category: "platform-bug",
        note: improvementNote,
        suggested_owner: "platform",
        created_by: "ui-operator"
      }),
    onSuccess: async () => {
      setImprovementNote("");
      await refreshEverything();
    }
  });

  const promoteMutation = useMutation({
    mutationFn: () => postJSON(`/api/proposals/promote`, { requested_by: "ui-operator" }),
    onSuccess: refreshEverything
  });

  const settingsMutation = useMutation({
    mutationFn: () => postJSON(`/api/settings`, { active_proposal_cap: Number(proposalCapInput) }),
    onSuccess: refreshEverything
  });

  const proposalDecisionMutation = useMutation({
    mutationFn: (decision: string) =>
      postJSON(`/api/proposals/${selectedProposalId}/decision`, {
        decision,
        rationale: proposalRationale || `UI operator recorded ${decision}.`,
        reviewer_id: "ui-operator",
        failure_class: decision === "rejected" ? "insufficient_evidence" : ""
      }),
    onSuccess: async () => {
      setProposalRationale("");
      await refreshEverything();
    }
  });

  const traceJudgments = recordOrEmpty(traceDetailQuery.data?.judgments_by_eval_run);
  const proposalSlotState = proposalsQuery.data?.proposal_slots;

  const proposalPRMap = useMemo(() => {
    return new Map(prAttempts.map((attempt) => [attempt.proposal_id, attempt]));
  }, [prAttempts]);

  const proposalJobMap = useMemo(() => {
    return new Map(repoChangeJobs.map((job) => [job.proposal_id, job]));
  }, [repoChangeJobs]);

  const proposalRows = proposalSegment === "active" ? activeProposals : proposalSegment === "history" ? historyProposals : [];

  return (
    <div className="app-shell">
      <aside className="nav-rail">
        <div className="brand-block">
          <p className="eyebrow">Improvement Plane</p>
          <h1>Recursive operator workspace</h1>
          <p className="muted">
            Inspect trace runs, move through proposal review, and keep the active proposal path capped.
          </p>
        </div>

        <nav className="tab-nav" aria-label="Sections">
          <button
            className={viewState.tab === "traces" ? "tab-button active" : "tab-button"}
            onClick={() => navigate({ tab: "traces", trace: selectedTraceId })}
          >
            <span>Traces</span>
            <strong>{traces.length}</strong>
          </button>
          <button
            className={viewState.tab === "proposals" ? "tab-button active" : "tab-button"}
            onClick={() => navigate({ tab: "proposals", proposal: selectedProposalId })}
          >
            <span>Proposals</span>
            <strong>{proposals.length}</strong>
          </button>
        </nav>

        <section className="operations-card">
          <div className="section-header">
            <div>
              <p className="eyebrow">Operations</p>
              <h2>Proposal cap</h2>
            </div>
            <span className="status-chip">{proposalSlotState?.active ?? 0}/{proposalSlotState?.cap ?? 0}</span>
          </div>
          <dl className="slot-grid">
            <div><dt>Active</dt><dd>{proposalSlotState?.active ?? 0}</dd></div>
            <div><dt>Available</dt><dd>{proposalSlotState?.available ?? 0}</dd></div>
            <div><dt>Stale</dt><dd>{listOrEmpty(proposalSlotState?.stale_proposal_ids).length}</dd></div>
            <div><dt>Queued candidates</dt><dd>{candidates.length}</dd></div>
          </dl>
          <label className="field">
            Active proposal cap
            <input type="number" min={1} value={proposalCapInput} onChange={(event) => setProposalCapInput(event.target.value)} />
          </label>
          <div className="button-row">
            <button onClick={() => settingsMutation.mutate()} disabled={settingsMutation.isPending}>
              Save cap
            </button>
            <button className="secondary" onClick={() => promoteMutation.mutate()} disabled={promoteMutation.isPending || (proposalSlotState?.available ?? 0) === 0}>
              Run promoter
            </button>
          </div>
        </section>

        <section className="operations-card">
          <div className="section-header">
            <div>
              <p className="eyebrow">Runtime</p>
              <h2>Runner roles</h2>
            </div>
          </div>
          <div className="runtime-list">
            {runtimeRoles.map((role) => (
              <div key={role.role} className="runtime-row">
                <div>
                  <strong>{role.role}</strong>
                  <p>{role.backend || "unreachable"} · {role.api_mode || "n/a"}</p>
                </div>
                <div className="runtime-meta">
                  <span className={role.healthy ? "status-dot ok" : "status-dot"} />
                  <small>{role.model}</small>
                  <small>{role.reasoning_effort}</small>
                </div>
              </div>
            ))}
          </div>
        </section>
      </aside>

      <section className="list-pane">
        {viewState.tab === "traces" ? (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Trace runs</p>
                <h2>Inbound events and workflow traces</h2>
              </div>
            </header>
            <div className="list-stack">
              {traces.map((trace) => (
                <button
                  key={trace.trace_id}
                  className={trace.trace_id === selectedTraceId ? "list-card selected" : "list-card"}
                  onClick={() => navigate({ tab: "traces", trace: trace.trace_id })}
                >
                  <div className="list-card-header">
                    <div>
                      <strong>{trace.trace_id}</strong>
                      <p>{trace.workflow_kind} · {trace.status}</p>
                    </div>
                    {trace.latest_eval ? <span className="status-chip eval">{trace.latest_eval.verdict}</span> : null}
                  </div>
                  <p className="trace-thread">{trace.thread_key}</p>
                  <dl className="mini-metrics">
                    <div><dt>Started</dt><dd>{formatTime(trace.started_at)}</dd></div>
                    <div><dt>Events</dt><dd>{trace.event_count}</dd></div>
                    <div><dt>Reasoning</dt><dd>{trace.reasoning_count}</dd></div>
                    <div><dt>Tools</dt><dd>{trace.tool_call_count}</dd></div>
                    <div><dt>Slack</dt><dd>{trace.slack_action_count}</dd></div>
                    <div><dt>Latest eval</dt><dd>{trace.latest_eval ? scoreBadge(trace.latest_eval.score) : "none"}</dd></div>
                  </dl>
                </button>
              ))}
            </div>
          </>
        ) : (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Proposals</p>
                <h2>Review path and PR-linked change state</h2>
              </div>
              <div className="segment-row">
                {(["active", "candidates", "history"] as ProposalSegment[]).map((segment) => (
                  <button
                    key={segment}
                    className={proposalSegment === segment ? "segment-button active" : "segment-button"}
                    onClick={() => setProposalSegment(segment)}
                  >
                    {segment}
                  </button>
                ))}
              </div>
            </header>

            {proposalSegment === "candidates" ? (
              <div className="list-stack">
                {candidates.map((candidate) => (
                  <div key={candidate.id} className="list-card static">
                    <div className="list-card-header">
                      <div>
                        <strong>{candidate.subsystem}</strong>
                        <p>{candidate.failure_mode}</p>
                      </div>
                      <span className="status-chip">{scoreBadge(candidate.priority_score)}</span>
                    </div>
                    <dl className="mini-metrics">
                      <div><dt>Status</dt><dd>{candidate.status}</dd></div>
                      <div><dt>Severity</dt><dd>{candidate.severity}</dd></div>
                      <div><dt>Recurrence</dt><dd>{candidate.recurrence_count}</dd></div>
                      <div><dt>Latest trace</dt><dd>{candidate.latest_trace_id || "none"}</dd></div>
                    </dl>
                  </div>
                ))}
              </div>
            ) : (
              <div className="list-stack">
                {proposalRows.map((proposal) => {
                  const prAttempt = proposalPRMap.get(proposal.id);
                  const repoJob = proposalJobMap.get(proposal.id);
                  return (
                    <button
                      key={proposal.id}
                      className={proposal.id === selectedProposalId ? "list-card selected" : "list-card"}
                      onClick={() => navigate({ tab: "proposals", proposal: proposal.id })}
                    >
                      <div className="list-card-header">
                        <div>
                          <strong>{proposal.title}</strong>
                          <p>{proposal.status} · {proposal.candidate_key}</p>
                        </div>
                        <span className="status-chip">{prAttempt?.pr_url ? "PR open" : repoJob?.status || proposal.status}</span>
                      </div>
                      <p className="trace-thread">{proposal.summary}</p>
                      <dl className="mini-metrics">
                        <div><dt>Risk</dt><dd>{proposal.risk_tier || "n/a"}</dd></div>
                        <div><dt>Trace</dt><dd>{proposal.trace_id}</dd></div>
                        <div><dt>Slot</dt><dd>{proposal.active_slot_consuming ? "consuming" : "inactive"}</dd></div>
                        <div><dt>PR</dt><dd>{prAttempt?.status || "none"}</dd></div>
                      </dl>
                    </button>
                  );
                })}
              </div>
            )}
          </>
        )}
      </section>

      <section className="detail-pane">
        {viewState.tab === "traces" ? (
          selectedTraceId && traceDetailQuery.data ? (
            <div className="detail-stack">
              <header className="detail-header">
                <div>
                  <p className="eyebrow">Trace detail</p>
                  <h2>{traceDetailQuery.data.trace.summary.trace_id}</h2>
                  <p className="muted">{traceDetailQuery.data.trace.summary.workflow_kind} · {traceDetailQuery.data.trace.summary.thread_key}</p>
                </div>
                <div className="button-row">
                  <button className="secondary" onClick={() => evaluateMutation.mutate()} disabled={evaluateMutation.isPending}>Run eval</button>
                  <button className="secondary" onClick={() => replayMutation.mutate()} disabled={replayMutation.isPending}>Queue replay</button>
                </div>
              </header>

              <section className="detail-card">
                <h3>Overview</h3>
                <dl className="overview-grid">
                  <div><dt>Status</dt><dd>{traceDetailQuery.data.trace.summary.status}</dd></div>
                  <div><dt>Started</dt><dd>{formatTime(traceDetailQuery.data.trace.summary.started_at)}</dd></div>
                  <div><dt>Events</dt><dd>{traceDetailQuery.data.trace.summary.event_count}</dd></div>
                  <div><dt>Artifacts</dt><dd>{traceDetailQuery.data.trace.summary.artifact_count}</dd></div>
                  <div><dt>Reasoning</dt><dd>{traceDetailQuery.data.trace.summary.reasoning_step_count}</dd></div>
                  <div><dt>Tool calls</dt><dd>{traceDetailQuery.data.trace.summary.tool_call_count}</dd></div>
                  <div><dt>Slack actions</dt><dd>{traceDetailQuery.data.trace.summary.slack_action_count}</dd></div>
                  <div><dt>Linked proposals</dt><dd>{listOrEmpty(traceDetailQuery.data.linked_proposals).length}</dd></div>
                </dl>
              </section>

              <DetailListSection title="Event timeline">
                {listOrEmpty(traceDetailQuery.data.trace.events).map((event) => (
                  <div key={`${event.event_type}-${event.started_at}`} className="detail-row">
                    <div>
                      <strong>{event.event_type}</strong>
                      <p>{event.plane} · {event.service} · {event.actor}</p>
                    </div>
                    <div className="detail-meta">
                      <small>{event.status}</small>
                      <small>{formatTime(event.started_at)}</small>
                    </div>
                    {event.description ? <p className="detail-copy">{event.description}</p> : null}
                  </div>
                ))}
              </DetailListSection>

              <DetailListSection title="Visible reasoning">
                {listOrEmpty(traceDetailQuery.data.trace.reasoning).map((step) => (
                  <div key={step.id} className="detail-row">
                    <div>
                      <strong>{step.step_type}</strong>
                      <p>{formatTime(step.created_at)}</p>
                    </div>
                    <p className="detail-copy">{step.summary}</p>
                    {step.decision ? <small>Decision: {step.decision}</small> : null}
                  </div>
                ))}
              </DetailListSection>

              <DetailListSection title="Tool lineage">
                {listOrEmpty(traceDetailQuery.data.trace.tool_calls).map((call) => (
                  <div key={call.id} className="detail-row">
                    <div>
                      <strong>{call.tool_name}</strong>
                      <p>{call.tool_call_id}</p>
                    </div>
                    <div className="detail-meta">
                      <small>{call.status || "completed"}</small>
                      <small>{call.approval_state || "n/a"}</small>
                    </div>
                    <p className="detail-copy">{call.interpretation_summary || call.summary}</p>
                  </div>
                ))}
              </DetailListSection>

              <DetailListSection title="Outbound Slack actions">
                {listOrEmpty(traceDetailQuery.data.trace.slack_actions).map((action) => (
                  <div key={action.id} className="detail-row">
                    <div>
                      <strong>{action.channel_id || "direct message"}</strong>
                      <p>{action.thread_ts || "no thread"}</p>
                    </div>
                    <div className="detail-meta">
                      <small>{action.policy_verdict || "n/a"}</small>
                      <small>{action.send_status || "draft_only"}</small>
                    </div>
                    <p className="detail-copy">{action.final_body || action.draft_body || "No reply body recorded."}</p>
                  </div>
                ))}
              </DetailListSection>

              <DetailListSection title="Eval runs and judgments">
                {listOrEmpty(traceDetailQuery.data.linked_eval_runs).map((run) => (
                  <div key={run.id} className="detail-row">
                    <div className="detail-row-header">
                      <div>
                        <strong>{run.suite_name}</strong>
                        <p>{run.overall_verdict} · {scoreBadge(run.overall_score)}</p>
                      </div>
                      <small>{formatTime(run.created_at)}</small>
                    </div>
                    <div className="nested-list">
                      {listOrEmpty(traceJudgments[run.id]).map((judgment) => (
                        <div key={judgment.id} className="nested-card">
                          <strong>{judgment.layer}</strong>
                          <p>{judgment.category} · {scoreBadge(judgment.score)}</p>
                          <small>{judgment.rationale}</small>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </DetailListSection>

              <section className="detail-card">
                <h3>Human review actions</h3>
                <div className="review-grid">
                  <div className="stack">
                    <label className="field">
                      Verdict
                      <select value={ratingVerdict} onChange={(event) => setRatingVerdict(event.target.value)}>
                        <option value="correct">correct</option>
                        <option value="partial">partial</option>
                        <option value="wrong">wrong</option>
                        <option value="unsafe">unsafe</option>
                        <option value="needs-human">needs-human</option>
                      </select>
                    </label>
                    <label className="field">
                      Score
                      <input type="number" min={1} max={5} value={ratingScore} onChange={(event) => setRatingScore(event.target.value)} />
                    </label>
                    <label className="field">
                      Rating notes
                      <textarea value={ratingNotes} onChange={(event) => setRatingNotes(event.target.value)} />
                    </label>
                    <button onClick={() => ratingMutation.mutate()} disabled={ratingMutation.isPending}>Submit rating</button>
                  </div>

                  <div className="stack">
                    <label className="field">
                      Improvement note
                      <textarea value={improvementNote} onChange={(event) => setImprovementNote(event.target.value)} />
                    </label>
                    <button className="secondary" onClick={() => noteMutation.mutate()} disabled={noteMutation.isPending}>Add note</button>
                    <div className="nested-list">
                      {listOrEmpty(traceDetailQuery.data.ratings).map((rating, index) => (
                        <div key={`${rating.created_at}-${index}`} className="nested-card">
                          <strong>{rating.verdict}</strong>
                          <p>{rating.reviewer_id} · {rating.score}/5</p>
                          <small>{rating.notes}</small>
                        </div>
                      ))}
                      {listOrEmpty(traceDetailQuery.data.improvement_notes).map((note, index) => (
                        <div key={`${note.created_at}-${index}`} className="nested-card">
                          <strong>{note.category}</strong>
                          <p>{note.suggested_owner || "unassigned"}</p>
                          <small>{note.note}</small>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </section>
            </div>
          ) : (
            <EmptyDetail title="Select a trace" body="Choose a trace from the center pane to load its event timeline, visible reasoning, tool lineage, evals, and review actions." />
          )
        ) : selectedProposalId && proposalDetailQuery.data ? (
          <div className="detail-stack">
            <header className="detail-header">
              <div>
                <p className="eyebrow">Proposal detail</p>
                <h2>{proposalDetailQuery.data.proposal.title}</h2>
                <p className="muted">{proposalDetailQuery.data.proposal.status} · {proposalDetailQuery.data.proposal.candidate_key}</p>
              </div>
              <div className="button-row">
                <button onClick={() => proposalDecisionMutation.mutate("approved")} disabled={proposalDecisionMutation.isPending}>Approve</button>
                <button className="secondary" onClick={() => proposalDecisionMutation.mutate("dismissed")} disabled={proposalDecisionMutation.isPending}>Dismiss</button>
                <button className="secondary" onClick={() => proposalDecisionMutation.mutate("rejected")} disabled={proposalDecisionMutation.isPending}>Reject</button>
                <button className="secondary" onClick={() => proposalDecisionMutation.mutate("merged")} disabled={proposalDecisionMutation.isPending}>Mark merged</button>
              </div>
            </header>

            <section className="detail-card">
              <h3>Proposal summary and hypothesis</h3>
              <p className="detail-copy">{proposalDetailQuery.data.proposal.summary}</p>
              <dl className="overview-grid">
                <div><dt>Risk tier</dt><dd>{proposalDetailQuery.data.proposal.risk_tier || "n/a"}</dd></div>
                <div><dt>Trace</dt><dd>{proposalDetailQuery.data.proposal.trace_id}</dd></div>
                <div><dt>Scope</dt><dd>{proposalDetailQuery.data.proposal.proposed_scope || "n/a"}</dd></div>
                <div><dt>Slot</dt><dd>{proposalDetailQuery.data.proposal.active_slot_consuming ? "consuming" : "inactive"}</dd></div>
              </dl>
            </section>

            <section className="detail-card">
              <h3>Why this proposal exists</h3>
              <label className="field">
                Review rationale
                <textarea value={proposalRationale} onChange={(event) => setProposalRationale(event.target.value)} placeholder="Add review rationale or closure context." />
              </label>
              <div className="nested-list">
                {listOrEmpty(proposalDetailQuery.data.linked_eval_runs).map((run) => (
                  <div key={run.id} className="nested-card">
                    <strong>{run.suite_name}</strong>
                    <p>{run.overall_verdict} · {scoreBadge(run.overall_score)}</p>
                    <small>{formatTime(run.created_at)}</small>
                  </div>
                ))}
              </div>
            </section>

            <DetailListSection title="Linked trace and eval references">
              {listOrEmpty(proposalDetailQuery.data.linked_trace_summaries).map((trace) => (
                <div key={trace.trace_id} className="detail-row">
                  <div>
                    <strong>{trace.trace_id}</strong>
                    <p>{trace.workflow_kind} · {trace.status}</p>
                  </div>
                  <div className="detail-meta">
                    <small>{trace.thread_key}</small>
                    <small>{formatTime(trace.started_at)}</small>
                  </div>
                </div>
              ))}
            </DetailListSection>

            <DetailListSection title="Proposal memory and prior rejection context">
              {listOrEmpty(proposalDetailQuery.data.related_proposal_memory).map((memory) => (
                <div key={memory.id} className="detail-row">
                  <div>
                    <strong>{memory.disposition}</strong>
                    <p>{memory.candidate_key}</p>
                  </div>
                  <div className="detail-meta">
                    <small>{memory.failure_class || "no failure class"}</small>
                    <small>{formatTime(memory.created_at)}</small>
                  </div>
                  <p className="detail-copy">{memory.review_rationale}</p>
                </div>
              ))}
            </DetailListSection>

            <DetailListSection title="Review history">
              {listOrEmpty(proposalDetailQuery.data.reviews).map((review) => (
                <div key={`${review.decision}-${review.created_at}`} className="detail-row">
                  <div>
                    <strong>{review.decision}</strong>
                    <p>{review.reviewer_id}</p>
                  </div>
                  <p className="detail-copy">{review.rationale}</p>
                </div>
              ))}
            </DetailListSection>

            <DetailListSection title="Repo-change job state">
              {listOrEmpty(proposalDetailQuery.data.repo_change_jobs).map((job) => (
                <div key={job.id} className="detail-row">
                  <div>
                    <strong>{job.repo}</strong>
                    <p>{job.branch_name}</p>
                  </div>
                  <div className="detail-meta">
                    <small>{job.status}</small>
                  </div>
                  <p className="detail-copy">{job.context_summary}</p>
                </div>
              ))}
            </DetailListSection>

            <DetailListSection title="PR attempts">
              {listOrEmpty(proposalDetailQuery.data.pr_attempts).map((attempt) => (
                <div key={attempt.id} className="detail-row">
                  <div>
                    <strong>{attempt.status}</strong>
                    <p>{attempt.validation_status}</p>
                  </div>
                  {attempt.pr_url ? (
                    <a className="detail-link" href={attempt.pr_url} target="_blank" rel="noreferrer">
                      Open PR
                    </a>
                  ) : (
                    <small>No PR URL</small>
                  )}
                </div>
              ))}
            </DetailListSection>

            <DetailListSection title="Post-merge replay state">
              {listOrEmpty(proposalDetailQuery.data.post_merge_replays).map((replay) => (
                <div key={replay.id} className="detail-row">
                  <div>
                    <strong>{replay.improved ? "Improved" : "No improvement"}</strong>
                    <p>{replay.trace_id}</p>
                  </div>
                  <div className="detail-meta">
                    <small>{scoreBadge(replay.baseline_score)} → {scoreBadge(replay.candidate_score)}</small>
                  </div>
                </div>
              ))}
            </DetailListSection>
          </div>
        ) : (
          <EmptyDetail title="Select a proposal" body="Choose a proposal from the center pane to inspect review history, prior memory, repo-change state, and linked PR attempts." />
        )}
      </section>
    </div>
  );
}

function DetailListSection(props: { title: string; children: ReactNode }) {
  return (
    <section className="detail-card">
      <h3>{props.title}</h3>
      <div className="detail-section-body">{props.children}</div>
    </section>
  );
}

function EmptyDetail(props: { title: string; body: string }) {
  return (
    <div className="empty-detail">
      <p className="eyebrow">Detail pane</p>
      <h2>{props.title}</h2>
      <p className="muted">{props.body}</p>
    </div>
  );
}
