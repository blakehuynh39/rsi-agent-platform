import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";

type NullableList<T> = T[] | null;
type TabKey = "conversations" | "cases" | "proposals";
type ProposalSegment = "active" | "candidates" | "history";
type TraceInspectorTab = "overview" | "timeline" | "reasoning" | "tools" | "slack" | "evals" | "feedback" | "proposals";

type TraceEvalSummary = {
  run_id: string;
  verdict: string;
  score: number;
  created_at: string;
  suite_name: string;
};

type TraceAttemptSummary = {
  trace_id: string;
  conversation_id: string;
  case_id: string;
  trigger_event_id?: string;
  supersedes_trace_id?: string;
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

type CaseSummary = {
  case_id: string;
  conversation_id: string;
  kind: string;
  intent: string;
  title: string;
  summary: string;
  status: string;
  assigned_bot: string;
  latest_trace_id?: string;
  latest_trace_verdict?: string;
  recurrence: number;
  linked_proposal_ids: NullableList<string>;
  updated_at: string;
};

type ConversationListItem = {
  conversation_id: string;
  source: string;
  external_key: string;
  title: string;
  status: string;
  active_case?: CaseSummary;
  latest_message_at: string;
  latest_trace_verdict?: string;
  open_trace_count: number;
  proposal_count: number;
};

type ConversationEntry = {
  id: string;
  event_id?: string;
  trace_id?: string;
  source: string;
  source_event_id: string;
  entry_type: string;
  actor_id?: string;
  actor_type?: string;
  body: string;
  created_at: string;
};

type ConversationDetailResponse = {
  conversation: {
    id: string;
    source: string;
    external_key: string;
    title: string;
    status: string;
    active_case_id?: string;
  };
  active_case?: CaseSummary;
  cases: NullableList<CaseSummary>;
  transcript: NullableList<ConversationEntry>;
  trace_attempts: NullableList<TraceAttemptSummary>;
  linked_proposals: NullableList<Proposal>;
};

type CaseDetailResponse = {
  case: CaseSummary;
  conversation: ConversationListItem;
  trace_attempts: NullableList<TraceAttemptSummary>;
  latest_eval_runs: NullableList<EvalRun>;
  linked_proposals: NullableList<Proposal>;
};

type TraceEvent = {
  trace_id: string;
  event_type: string;
  plane: string;
  service: string;
  actor: string;
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

type TraceDetailResponse = {
  trace: {
    summary: {
      trace_id: string;
      conversation_id: string;
      case_id: string;
      trigger_event_id?: string;
      workflow_kind: string;
      status: string;
      thread_key: string;
      started_at: string;
      event_count: number;
      artifact_count: number;
      reasoning_step_count: number;
      tool_call_count: number;
      slack_action_count: number;
      last_verdict?: string;
    };
    events: NullableList<TraceEvent>;
    artifacts: NullableList<Artifact>;
    reasoning: NullableList<ReasoningStep>;
    tool_calls: NullableList<ToolCallRecord>;
    slack_actions: NullableList<SlackActionRecord>;
  };
  conversation: ConversationListItem;
  case?: CaseSummary;
  transcript_slice: NullableList<ConversationEntry>;
  linked_eval_runs: NullableList<EvalRun>;
  judgments_by_eval_run: Record<string, NullableList<EvalJudgment>>;
  feedback_records: NullableList<FeedbackRecord>;
  linked_proposals: NullableList<Proposal>;
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

type FeedbackRecord = {
  id: string;
  conversation_id?: string;
  case_id?: string;
  trace_id?: string;
  target_type: string;
  target_id: string;
  score?: number;
  verdict?: string;
  labels?: NullableList<string>;
  notes?: string;
  reviewer_id: string;
  created_at: string;
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
  prior_similar_proposal_ids?: NullableList<string>;
};

type Proposal = {
  id: string;
  trace_id: string;
  conversation_id?: string;
  case_id?: string;
  origin_trace_id?: string;
  evidence_trace_ids?: NullableList<string>;
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
  prior_similar_proposal_ids?: NullableList<string>;
  new_evidence_since_last_rejection: boolean;
  created_at: string;
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
  conversation_id?: string;
  case_id?: string;
  origin_trace_id?: string;
  evidence_trace_ids?: NullableList<string>;
  hypothesis: string;
  diff_summary: string;
  review_rationale: string;
  disposition: string;
  disposition_reason?: string;
  failure_class?: string;
  failure_classes?: NullableList<string>;
  created_at: string;
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

type ProposalSlots = {
  cap: number;
  active: number;
  available: number;
  active_proposal_ids: NullableList<string>;
  stale_proposal_ids: NullableList<string>;
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
  linked_trace_summaries: NullableList<TraceAttemptSummary>;
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
  conversation?: string;
  case?: string;
  trace?: string;
  proposal?: string;
};

const ACTIVE_PROPOSAL_STATES = new Set([
  "pending_review",
  "approved",
  "repo_change_queued",
  "repo_change_running",
  "validation_pending",
  "pr_open"
]);

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
  const tabValue = params.get("tab");
  const tab: TabKey = tabValue === "cases" ? "cases" : tabValue === "proposals" ? "proposals" : "conversations";
  return {
    tab,
    conversation: params.get("conversation") || undefined,
    case: params.get("case") || undefined,
    trace: params.get("trace") || undefined,
    proposal: params.get("proposal") || undefined
  };
}

function writeViewState(next: ViewState) {
  const params = new URLSearchParams();
  params.set("tab", next.tab);
  if (next.conversation) params.set("conversation", next.conversation);
  if (next.case) params.set("case", next.case);
  if (next.trace) params.set("trace", next.trace);
  if (next.proposal) params.set("proposal", next.proposal);
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

function proposalPRState(proposalId: string, attempts: PRAttempt[]) {
  return attempts.find((attempt) => attempt.proposal_id === proposalId);
}

function proposalJobState(proposalId: string, jobs: RepoChangeJob[]) {
  return jobs.find((job) => job.proposal_id === proposalId);
}

export function App() {
  const queryClient = useQueryClient();
  const [viewState, setViewState] = useState<ViewState>(() => readViewState());
  const [proposalSegment, setProposalSegment] = useState<ProposalSegment>("active");
  const [traceInspectorTab, setTraceInspectorTab] = useState<TraceInspectorTab>("overview");
  const [proposalCapInput, setProposalCapInput] = useState("2");
  const [feedbackTargetType, setFeedbackTargetType] = useState("trace");
  const [feedbackTargetID, setFeedbackTargetID] = useState("");
  const [feedbackScore, setFeedbackScore] = useState("3");
  const [feedbackVerdict, setFeedbackVerdict] = useState("useful");
  const [feedbackNotes, setFeedbackNotes] = useState("");
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

  const conversationsQuery = useQuery({
    queryKey: ["conversations"],
    queryFn: () => getJSON<{ conversations: ConversationListItem[] }>("/api/conversations")
  });

  const casesQuery = useQuery({
    queryKey: ["cases"],
    queryFn: () => getJSON<{ cases: CaseSummary[] }>("/api/cases")
  });

  const proposalsQuery = useQuery({
    queryKey: ["proposals"],
    queryFn: () => getJSON<ProposalResponse>("/api/proposals")
  });

  const runtimeQuery = useQuery({
    queryKey: ["runtime"],
    queryFn: () => getJSON<RuntimeResponse>("/api/runtime")
  });

  const conversationDetailQuery = useQuery({
    queryKey: ["conversation", viewState.conversation],
    queryFn: () => getJSON<ConversationDetailResponse>(`/api/conversations/${viewState.conversation}`),
    enabled: Boolean(viewState.tab === "conversations" && viewState.conversation)
  });

  const caseDetailQuery = useQuery({
    queryKey: ["case", viewState.case],
    queryFn: () => getJSON<CaseDetailResponse>(`/api/cases/${viewState.case}`),
    enabled: Boolean(viewState.tab === "cases" && viewState.case)
  });

  const traceDetailQuery = useQuery({
    queryKey: ["trace", viewState.trace],
    queryFn: () => getJSON<TraceDetailResponse>(`/api/traces/${viewState.trace}`),
    enabled: Boolean(viewState.trace)
  });

  const proposalDetailQuery = useQuery({
    queryKey: ["proposal", viewState.proposal],
    queryFn: () => getJSON<ProposalDetailResponse>(`/api/proposals/${viewState.proposal}`),
    enabled: Boolean(viewState.tab === "proposals" && viewState.proposal)
  });

  const conversations = listOrEmpty(conversationsQuery.data?.conversations);
  const cases = listOrEmpty(casesQuery.data?.cases);
  const proposals = listOrEmpty(proposalsQuery.data?.proposals);
  const candidates = listOrEmpty(proposalsQuery.data?.candidates);
  const proposalMemories = listOrEmpty(proposalsQuery.data?.proposal_memory);
  const repoChangeJobs = listOrEmpty(proposalsQuery.data?.repo_change_jobs);
  const prAttempts = listOrEmpty(proposalsQuery.data?.pr_attempts);
  const runtimeRoles = listOrEmpty(runtimeQuery.data?.roles);
  const proposalSlotState = proposalsQuery.data?.proposal_slots;

  useEffect(() => {
    const settingValue = proposalsQuery.data?.settings?.active_proposal_cap;
    if (typeof settingValue === "number") {
      setProposalCapInput(String(settingValue));
    }
  }, [proposalsQuery.data?.settings?.active_proposal_cap]);

  useEffect(() => {
    if (viewState.conversation && !conversations.some((item) => item.conversation_id === viewState.conversation)) {
      navigate({ tab: "conversations" });
    }
  }, [viewState.conversation, conversations]);

  useEffect(() => {
    if (viewState.case && !cases.some((item) => item.case_id === viewState.case)) {
      navigate({ tab: "cases" });
    }
  }, [viewState.case, cases]);

  useEffect(() => {
    if (viewState.proposal && !proposals.some((item) => item.id === viewState.proposal)) {
      navigate({ tab: "proposals" });
    }
  }, [viewState.proposal, proposals]);

  const refreshEverything = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["conversations"] }),
      queryClient.invalidateQueries({ queryKey: ["conversation"] }),
      queryClient.invalidateQueries({ queryKey: ["cases"] }),
      queryClient.invalidateQueries({ queryKey: ["case"] }),
      queryClient.invalidateQueries({ queryKey: ["trace"] }),
      queryClient.invalidateQueries({ queryKey: ["proposals"] }),
      queryClient.invalidateQueries({ queryKey: ["proposal"] }),
      queryClient.invalidateQueries({ queryKey: ["runtime"] })
    ]);
  };

  const evaluateMutation = useMutation({
    mutationFn: () => postJSON(`/api/traces/${viewState.trace}/evaluate`, {}),
    onSuccess: refreshEverything
  });

  const replayMutation = useMutation({
    mutationFn: () => postJSON(`/api/traces/${viewState.trace}/replay`, { requested_by: "ui-operator" }),
    onSuccess: refreshEverything
  });

  const feedbackMutation = useMutation({
    mutationFn: () =>
      postJSON(`/api/feedback`, {
        target_type: feedbackTargetType,
        target_id: feedbackTargetID,
        score: Number(feedbackScore),
        verdict: feedbackVerdict,
        labels: ["operator-ui"],
        notes: feedbackNotes,
        reviewer_id: "ui-operator"
      }),
    onSuccess: async () => {
      setFeedbackNotes("");
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
      postJSON(`/api/proposals/${viewState.proposal}/decision`, {
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

  const activeProposals = useMemo(
    () => proposals.filter((proposal) => ACTIVE_PROPOSAL_STATES.has(proposal.status)),
    [proposals]
  );
  const historyProposals = useMemo(
    () => proposals.filter((proposal) => !ACTIVE_PROPOSAL_STATES.has(proposal.status)),
    [proposals]
  );
  const proposalRows = proposalSegment === "active" ? activeProposals : proposalSegment === "history" ? historyProposals : [];

  const traceDetail = traceDetailQuery.data;
  const feedbackTargets = useMemo(() => {
    if (!traceDetail) return [];
    const targets = [{ label: `Trace ${traceDetail.trace.summary.trace_id}`, value: traceDetail.trace.summary.trace_id, type: "trace" }];
    for (const step of listOrEmpty(traceDetail.trace.reasoning)) {
      targets.push({ label: `Reasoning: ${step.step_type}`, value: step.id, type: "reasoning_step" });
    }
    for (const call of listOrEmpty(traceDetail.trace.tool_calls)) {
      targets.push({ label: `Tool: ${call.tool_name}`, value: call.id, type: "tool_call" });
    }
    for (const action of listOrEmpty(traceDetail.trace.slack_actions)) {
      targets.push({ label: `Slack action ${formatTime(action.created_at)}`, value: action.id, type: "slack_action" });
    }
    return targets;
  }, [traceDetail]);

  useEffect(() => {
    if (feedbackTargets.length > 0) {
      setFeedbackTargetType(feedbackTargets[0].type);
      setFeedbackTargetID(feedbackTargets[0].value);
    }
  }, [traceDetail?.trace.summary.trace_id]);

  const traceJudgments = recordOrEmpty(traceDetail?.judgments_by_eval_run);

  return (
    <div className="app-shell">
      <aside className="nav-rail">
        <div className="brand-block">
          <p className="eyebrow">Improvement Plane</p>
          <h1>Evidence-first operator workspace</h1>
          <p className="muted">
            Start from conversations, move into cases, and inspect the exact trace evidence that produced a proposal.
          </p>
        </div>

        <nav className="tab-nav" aria-label="Sections">
          <button className={viewState.tab === "conversations" ? "tab-button active" : "tab-button"} onClick={() => navigate({ tab: "conversations" })}>
            <span>Conversations</span>
            <strong>{conversations.length}</strong>
          </button>
          <button className={viewState.tab === "cases" ? "tab-button active" : "tab-button"} onClick={() => navigate({ tab: "cases" })}>
            <span>Cases</span>
            <strong>{cases.length}</strong>
          </button>
          <button className={viewState.tab === "proposals" ? "tab-button active" : "tab-button"} onClick={() => navigate({ tab: "proposals" })}>
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
            <div><dt>Candidates</dt><dd>{candidates.length}</dd></div>
          </dl>
          <label className="field">
            Active proposal cap
            <input type="number" min={1} value={proposalCapInput} onChange={(event) => setProposalCapInput(event.target.value)} />
          </label>
          <div className="button-row">
            <button onClick={() => settingsMutation.mutate()} disabled={settingsMutation.isPending}>Save cap</button>
            <button className="secondary" onClick={() => promoteMutation.mutate()} disabled={promoteMutation.isPending || (proposalSlotState?.available ?? 0) === 0}>
              Run promoter
            </button>
          </div>
        </section>

        <section className="operations-card">
          <div className="section-header">
            <div>
              <p className="eyebrow">Runtime</p>
              <h2>Model runtime</h2>
            </div>
          </div>
          <div className="runtime-list">
            {runtimeRoles.map((role) => (
              <div key={role.role} className="runtime-row">
                <div>
                  <strong>{role.role}</strong>
                  <p>{role.backend || "unreachable"} · {role.provider || "n/a"}</p>
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
        {viewState.tab === "conversations" ? (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Conversations</p>
                <h2>Slack threads, DMs, and incident rooms</h2>
              </div>
            </header>
            <div className="list-stack">
              {conversations.map((item) => (
                <button
                  key={item.conversation_id}
                  className={item.conversation_id === viewState.conversation ? "list-card selected" : "list-card"}
                  onClick={() => navigate({ tab: "conversations", conversation: item.conversation_id, trace: item.active_case?.latest_trace_id })}
                >
                  <div className="list-card-header">
                    <div>
                      <strong>{item.title}</strong>
                      <p>{item.source} · {item.status}</p>
                    </div>
                    {item.latest_trace_verdict ? <span className="status-chip eval">{item.latest_trace_verdict}</span> : null}
                  </div>
                  <p className="trace-thread">{item.external_key}</p>
                  <dl className="mini-metrics">
                    <div><dt>Active case</dt><dd>{item.active_case?.title || "none"}</dd></div>
                    <div><dt>Latest</dt><dd>{formatTime(item.latest_message_at)}</dd></div>
                    <div><dt>Open traces</dt><dd>{item.open_trace_count}</dd></div>
                    <div><dt>Proposals</dt><dd>{item.proposal_count}</dd></div>
                  </dl>
                </button>
              ))}
            </div>
          </>
        ) : viewState.tab === "cases" ? (
          <>
            <header className="pane-header">
              <div>
                <p className="eyebrow">Cases</p>
                <h2>Cross-conversation objectives</h2>
              </div>
            </header>
            <div className="list-stack">
              {cases.map((item) => (
                <button
                  key={item.case_id}
                  className={item.case_id === viewState.case ? "list-card selected" : "list-card"}
                  onClick={() => navigate({ tab: "cases", case: item.case_id, trace: item.latest_trace_id })}
                >
                  <div className="list-card-header">
                    <div>
                      <strong>{item.title}</strong>
                      <p>{item.kind} · {item.status}</p>
                    </div>
                    {item.latest_trace_verdict ? <span className="status-chip eval">{item.latest_trace_verdict}</span> : null}
                  </div>
                  <p className="trace-thread">{item.summary}</p>
                  <dl className="mini-metrics">
                    <div><dt>Conversation</dt><dd>{item.conversation_id}</dd></div>
                    <div><dt>Bot</dt><dd>{item.assigned_bot}</dd></div>
                    <div><dt>Recurrence</dt><dd>{item.recurrence}</dd></div>
                    <div><dt>Proposals</dt><dd>{listOrEmpty(item.linked_proposal_ids).length}</dd></div>
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
                <h2>Review path and PR readiness</h2>
              </div>
              <div className="segment-row">
                {(["active", "candidates", "history"] as ProposalSegment[]).map((segment) => (
                  <button key={segment} className={proposalSegment === segment ? "segment-button active" : "segment-button"} onClick={() => setProposalSegment(segment)}>
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
                  const prAttempt = proposalPRState(proposal.id, prAttempts);
                  const repoJob = proposalJobState(proposal.id, repoChangeJobs);
                  return (
                    <button
                      key={proposal.id}
                      className={proposal.id === viewState.proposal ? "list-card selected" : "list-card"}
                      onClick={() => navigate({ tab: "proposals", proposal: proposal.id })}
                    >
                      <div className="list-card-header">
                        <div>
                          <strong>{proposal.title}</strong>
                          <p>{proposal.status} · {proposal.candidate_key}</p>
                        </div>
                        <span className="status-chip">{prAttempt?.status || repoJob?.status || proposal.status}</span>
                      </div>
                      <p className="trace-thread">{proposal.summary}</p>
                      <dl className="mini-metrics">
                        <div><dt>Risk</dt><dd>{proposal.risk_tier || "n/a"}</dd></div>
                        <div><dt>Case</dt><dd>{proposal.case_id || "none"}</dd></div>
                        <div><dt>Slot</dt><dd>{proposal.active_slot_consuming ? "occupied" : "free"}</dd></div>
                        <div><dt>PR</dt><dd>{prAttempt?.pr_url ? "linked" : "none"}</dd></div>
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
        {viewState.tab === "conversations" ? (
          !viewState.conversation ? (
            <EmptyDetail title="Select a conversation" body="Start from the conversation list. Once selected, you’ll see transcript context, case continuity, trace attempts, and the evidence behind the latest run." />
          ) : conversationDetailQuery.isLoading ? (
            <EmptyDetail title="Loading conversation" body="Fetching transcript, cases, and trace attempts." />
          ) : conversationDetailQuery.data ? (
            <ConversationDetail
              detail={conversationDetailQuery.data}
              selectedTraceId={viewState.trace}
              traceDetail={traceDetail}
              traceInspectorTab={traceInspectorTab}
              setTraceInspectorTab={setTraceInspectorTab}
              onSelectTrace={(traceId) => navigate({ tab: "conversations", conversation: viewState.conversation, trace: traceId })}
              onRunEval={() => evaluateMutation.mutate()}
              onReplay={() => replayMutation.mutate()}
              traceJudgments={traceJudgments}
              feedbackTargets={feedbackTargets}
              feedbackTargetType={feedbackTargetType}
              setFeedbackTargetType={setFeedbackTargetType}
              feedbackTargetID={feedbackTargetID}
              setFeedbackTargetID={setFeedbackTargetID}
              feedbackScore={feedbackScore}
              setFeedbackScore={setFeedbackScore}
              feedbackVerdict={feedbackVerdict}
              setFeedbackVerdict={setFeedbackVerdict}
              feedbackNotes={feedbackNotes}
              setFeedbackNotes={setFeedbackNotes}
              onSubmitFeedback={() => feedbackMutation.mutate()}
            />
          ) : (
            <EmptyDetail title="Conversation not found" body="The selected conversation no longer exists." />
          )
        ) : viewState.tab === "cases" ? (
          !viewState.case ? (
            <EmptyDetail title="Select a case" body="Cases are the active objectives inside conversations. Pick one to inspect its attempts and current evidence." />
          ) : caseDetailQuery.isLoading ? (
            <EmptyDetail title="Loading case" body="Fetching case summary and associated traces." />
          ) : caseDetailQuery.data ? (
            <CaseDetail
              detail={caseDetailQuery.data}
              selectedTraceId={viewState.trace}
              traceDetail={traceDetail}
              traceInspectorTab={traceInspectorTab}
              setTraceInspectorTab={setTraceInspectorTab}
              onSelectTrace={(traceId) => navigate({ tab: "cases", case: viewState.case, trace: traceId })}
              onRunEval={() => evaluateMutation.mutate()}
              onReplay={() => replayMutation.mutate()}
              traceJudgments={traceJudgments}
              feedbackTargets={feedbackTargets}
              feedbackTargetType={feedbackTargetType}
              setFeedbackTargetType={setFeedbackTargetType}
              feedbackTargetID={feedbackTargetID}
              setFeedbackTargetID={setFeedbackTargetID}
              feedbackScore={feedbackScore}
              setFeedbackScore={setFeedbackScore}
              feedbackVerdict={feedbackVerdict}
              setFeedbackVerdict={setFeedbackVerdict}
              feedbackNotes={feedbackNotes}
              setFeedbackNotes={setFeedbackNotes}
              onSubmitFeedback={() => feedbackMutation.mutate()}
            />
          ) : (
            <EmptyDetail title="Case not found" body="The selected case no longer exists." />
          )
        ) : !viewState.proposal ? (
          <EmptyDetail title="Select a proposal" body="Proposals remain the review and PR-path surface. Select one to inspect reasoning, memory, evidence traces, and linked PR state." />
        ) : proposalDetailQuery.isLoading ? (
          <EmptyDetail title="Loading proposal" body="Fetching proposal reviews, memory, repo-change state, and linked traces." />
        ) : proposalDetailQuery.data ? (
          <ProposalDetail
            detail={proposalDetailQuery.data}
            proposalRationale={proposalRationale}
            setProposalRationale={setProposalRationale}
            onDecision={(decision) => proposalDecisionMutation.mutate(decision)}
            proposalMemories={proposalMemories}
          />
        ) : (
          <EmptyDetail title="Proposal not found" body="The selected proposal no longer exists." />
        )}
      </section>
    </div>
  );
}

function ConversationDetail(props: {
  detail: ConversationDetailResponse;
  selectedTraceId?: string;
  traceDetail?: TraceDetailResponse;
  traceInspectorTab: TraceInspectorTab;
  setTraceInspectorTab: (tab: TraceInspectorTab) => void;
  onSelectTrace: (traceId: string) => void;
  onRunEval: () => void;
  onReplay: () => void;
  traceJudgments: Record<string, NullableList<EvalJudgment>>;
  feedbackTargets: { label: string; value: string; type: string }[];
  feedbackTargetType: string;
  setFeedbackTargetType: (value: string) => void;
  feedbackTargetID: string;
  setFeedbackTargetID: (value: string) => void;
  feedbackScore: string;
  setFeedbackScore: (value: string) => void;
  feedbackVerdict: string;
  setFeedbackVerdict: (value: string) => void;
  feedbackNotes: string;
  setFeedbackNotes: (value: string) => void;
  onSubmitFeedback: () => void;
}) {
  const transcript = listOrEmpty(props.detail.transcript);
  const traces = listOrEmpty(props.detail.trace_attempts);
  return (
    <div className="detail-stack">
      <div className="detail-card">
        <div className="detail-header">
          <div>
            <p className="eyebrow">Conversation</p>
            <h2>{props.detail.conversation.title || props.detail.conversation.external_key}</h2>
          </div>
          <div className="detail-meta">
            <span className="status-chip">{props.detail.conversation.status}</span>
            <span className="status-chip">{props.detail.conversation.source}</span>
          </div>
        </div>
        <dl className="overview-grid">
          <div><dt>External key</dt><dd>{props.detail.conversation.external_key}</dd></div>
          <div><dt>Active case</dt><dd>{props.detail.active_case?.title || "none"}</dd></div>
          <div><dt>Trace attempts</dt><dd>{traces.length}</dd></div>
          <div><dt>Linked proposals</dt><dd>{listOrEmpty(props.detail.linked_proposals).length}</dd></div>
        </dl>
      </div>

      <div className="review-grid">
        <div className="detail-card">
          <h3>Transcript</h3>
          <div className="nested-list">
            {transcript.map((entry) => (
              <div key={entry.id} className="nested-card">
                <div className="detail-row-header">
                  <strong>{entry.actor_type || "actor"}</strong>
                  <small>{formatTime(entry.created_at)}</small>
                </div>
                <p className="detail-copy">{entry.body}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="detail-card">
          <h3>Trace attempts</h3>
          <div className="nested-list">
            {traces.map((trace) => (
              <button key={trace.trace_id} className={trace.trace_id === props.selectedTraceId ? "list-card selected" : "list-card"} onClick={() => props.onSelectTrace(trace.trace_id)}>
                <div className="list-card-header">
                  <div>
                    <strong>{trace.trace_id}</strong>
                    <p>{trace.workflow_kind} · {trace.status}</p>
                  </div>
                  {trace.latest_eval ? <span className="status-chip eval">{trace.latest_eval.verdict}</span> : null}
                </div>
                <dl className="mini-metrics">
                  <div><dt>Started</dt><dd>{formatTime(trace.started_at)}</dd></div>
                  <div><dt>Events</dt><dd>{trace.event_count}</dd></div>
                  <div><dt>Reasoning</dt><dd>{trace.reasoning_count}</dd></div>
                  <div><dt>Tools</dt><dd>{trace.tool_call_count}</dd></div>
                </dl>
              </button>
            ))}
          </div>
        </div>
      </div>

      <TraceInspector
        selectedTraceId={props.selectedTraceId}
        traceDetail={props.traceDetail}
        tab={props.traceInspectorTab}
        setTab={props.setTraceInspectorTab}
        onRunEval={props.onRunEval}
        onReplay={props.onReplay}
        traceJudgments={props.traceJudgments}
        feedbackTargets={props.feedbackTargets}
        feedbackTargetType={props.feedbackTargetType}
        setFeedbackTargetType={props.setFeedbackTargetType}
        feedbackTargetID={props.feedbackTargetID}
        setFeedbackTargetID={props.setFeedbackTargetID}
        feedbackScore={props.feedbackScore}
        setFeedbackScore={props.setFeedbackScore}
        feedbackVerdict={props.feedbackVerdict}
        setFeedbackVerdict={props.setFeedbackVerdict}
        feedbackNotes={props.feedbackNotes}
        setFeedbackNotes={props.setFeedbackNotes}
        onSubmitFeedback={props.onSubmitFeedback}
      />
    </div>
  );
}

function CaseDetail(props: {
  detail: CaseDetailResponse;
  selectedTraceId?: string;
  traceDetail?: TraceDetailResponse;
  traceInspectorTab: TraceInspectorTab;
  setTraceInspectorTab: (tab: TraceInspectorTab) => void;
  onSelectTrace: (traceId: string) => void;
  onRunEval: () => void;
  onReplay: () => void;
  traceJudgments: Record<string, NullableList<EvalJudgment>>;
  feedbackTargets: { label: string; value: string; type: string }[];
  feedbackTargetType: string;
  setFeedbackTargetType: (value: string) => void;
  feedbackTargetID: string;
  setFeedbackTargetID: (value: string) => void;
  feedbackScore: string;
  setFeedbackScore: (value: string) => void;
  feedbackVerdict: string;
  setFeedbackVerdict: (value: string) => void;
  feedbackNotes: string;
  setFeedbackNotes: (value: string) => void;
  onSubmitFeedback: () => void;
}) {
  return (
    <div className="detail-stack">
      <div className="detail-card">
        <div className="detail-header">
          <div>
            <p className="eyebrow">Case</p>
            <h2>{props.detail.case.title}</h2>
          </div>
          <div className="detail-meta">
            <span className="status-chip">{props.detail.case.status}</span>
            <span className="status-chip">{props.detail.case.kind}</span>
          </div>
        </div>
        <p className="detail-copy">{props.detail.case.summary}</p>
        <dl className="overview-grid">
          <div><dt>Conversation</dt><dd>{props.detail.conversation.title}</dd></div>
          <div><dt>Assigned bot</dt><dd>{props.detail.case.assigned_bot}</dd></div>
          <div><dt>Recurrence</dt><dd>{props.detail.case.recurrence}</dd></div>
          <div><dt>Linked proposals</dt><dd>{listOrEmpty(props.detail.case.linked_proposal_ids).length}</dd></div>
        </dl>
      </div>

      <div className="detail-card">
        <h3>Trace attempts</h3>
        <div className="nested-list">
          {listOrEmpty(props.detail.trace_attempts).map((trace) => (
            <button key={trace.trace_id} className={trace.trace_id === props.selectedTraceId ? "list-card selected" : "list-card"} onClick={() => props.onSelectTrace(trace.trace_id)}>
              <div className="list-card-header">
                <div>
                  <strong>{trace.trace_id}</strong>
                  <p>{trace.workflow_kind} · {trace.status}</p>
                </div>
                {trace.latest_eval ? <span className="status-chip eval">{trace.latest_eval.verdict}</span> : null}
              </div>
              <dl className="mini-metrics">
                <div><dt>Started</dt><dd>{formatTime(trace.started_at)}</dd></div>
                <div><dt>Events</dt><dd>{trace.event_count}</dd></div>
                <div><dt>Tools</dt><dd>{trace.tool_call_count}</dd></div>
                <div><dt>Slack</dt><dd>{trace.slack_action_count}</dd></div>
              </dl>
            </button>
          ))}
        </div>
      </div>

      <TraceInspector
        selectedTraceId={props.selectedTraceId}
        traceDetail={props.traceDetail}
        tab={props.traceInspectorTab}
        setTab={props.setTraceInspectorTab}
        onRunEval={props.onRunEval}
        onReplay={props.onReplay}
        traceJudgments={props.traceJudgments}
        feedbackTargets={props.feedbackTargets}
        feedbackTargetType={props.feedbackTargetType}
        setFeedbackTargetType={props.setFeedbackTargetType}
        feedbackTargetID={props.feedbackTargetID}
        setFeedbackTargetID={props.setFeedbackTargetID}
        feedbackScore={props.feedbackScore}
        setFeedbackScore={props.setFeedbackScore}
        feedbackVerdict={props.feedbackVerdict}
        setFeedbackVerdict={props.setFeedbackVerdict}
        feedbackNotes={props.feedbackNotes}
        setFeedbackNotes={props.setFeedbackNotes}
        onSubmitFeedback={props.onSubmitFeedback}
      />
    </div>
  );
}

function ProposalDetail(props: {
  detail: ProposalDetailResponse;
  proposalMemories: ProposalMemory[];
  proposalRationale: string;
  setProposalRationale: (value: string) => void;
  onDecision: (decision: string) => void;
}) {
  const prAttempt = listOrEmpty(props.detail.pr_attempts)[0];
  return (
    <div className="detail-stack">
      <div className="detail-card">
        <div className="detail-header">
          <div>
            <p className="eyebrow">Proposal</p>
            <h2>{props.detail.proposal.title}</h2>
          </div>
          <div className="detail-meta">
            <span className="status-chip">{props.detail.proposal.status}</span>
            {prAttempt?.pr_url ? <a className="detail-link" href={prAttempt.pr_url} target="_blank" rel="noreferrer">Open PR</a> : null}
          </div>
        </div>
        <p className="detail-copy">{props.detail.proposal.summary}</p>
        <dl className="overview-grid">
          <div><dt>Candidate key</dt><dd>{props.detail.proposal.candidate_key}</dd></div>
          <div><dt>Risk</dt><dd>{props.detail.proposal.risk_tier || "n/a"}</dd></div>
          <div><dt>Origin trace</dt><dd>{props.detail.proposal.origin_trace_id || props.detail.proposal.trace_id}</dd></div>
          <div><dt>Evidence traces</dt><dd>{listOrEmpty(props.detail.proposal.evidence_trace_ids).length}</dd></div>
        </dl>
      </div>

      <div className="review-grid">
        <div className="detail-card">
          <h3>Proposal memory</h3>
          <div className="nested-list">
            {listOrEmpty(props.detail.related_proposal_memory).map((memory) => (
              <div key={memory.id} className="nested-card">
                <div className="detail-row-header">
                  <strong>{memory.disposition}</strong>
                  <small>{formatTime(memory.created_at)}</small>
                </div>
                <p className="detail-copy">{memory.review_rationale}</p>
              </div>
            ))}
            {!listOrEmpty(props.detail.related_proposal_memory).length && !props.proposalMemories.length ? (
              <div className="nested-card"><p className="detail-copy">No prior memory recorded.</p></div>
            ) : null}
          </div>
        </div>

        <div className="detail-card">
          <h3>Review actions</h3>
          <label className="field">
            Decision rationale
            <textarea value={props.proposalRationale} onChange={(event) => props.setProposalRationale(event.target.value)} placeholder="Why this should advance, be dismissed, or be rejected." />
          </label>
          <div className="button-row">
            <button onClick={() => props.onDecision("approved")}>Approve</button>
            <button className="secondary" onClick={() => props.onDecision("dismissed")}>Dismiss</button>
            <button className="secondary" onClick={() => props.onDecision("rejected")}>Reject</button>
            <button className="secondary" onClick={() => props.onDecision("merged")}>Mark merged</button>
          </div>
        </div>
      </div>

      <div className="detail-card">
        <h3>Linked traces and PR path</h3>
        <div className="nested-list">
          {listOrEmpty(props.detail.linked_trace_summaries).map((trace) => (
            <div key={trace.trace_id} className="nested-card">
              <div className="detail-row-header">
                <strong>{trace.trace_id}</strong>
                <small>{trace.status}</small>
              </div>
              <p className="detail-copy">{trace.workflow_kind} · {formatTime(trace.started_at)}</p>
            </div>
          ))}
          {listOrEmpty(props.detail.repo_change_jobs).map((job) => (
            <div key={job.id} className="nested-card">
              <div className="detail-row-header">
                <strong>{job.repo}</strong>
                <small>{job.status}</small>
              </div>
              <p className="detail-copy">{job.branch_name}</p>
            </div>
          ))}
          {listOrEmpty(props.detail.pr_attempts).map((attempt) => (
            <div key={attempt.id} className="nested-card">
              <div className="detail-row-header">
                <strong>{attempt.status}</strong>
                <small>{attempt.validation_status}</small>
              </div>
              {attempt.pr_url ? <a className="detail-link" href={attempt.pr_url} target="_blank" rel="noreferrer">{attempt.pr_url}</a> : null}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function TraceInspector(props: {
  selectedTraceId?: string;
  traceDetail?: TraceDetailResponse;
  tab: TraceInspectorTab;
  setTab: (tab: TraceInspectorTab) => void;
  onRunEval: () => void;
  onReplay: () => void;
  traceJudgments: Record<string, NullableList<EvalJudgment>>;
  feedbackTargets: { label: string; value: string; type: string }[];
  feedbackTargetType: string;
  setFeedbackTargetType: (value: string) => void;
  feedbackTargetID: string;
  setFeedbackTargetID: (value: string) => void;
  feedbackScore: string;
  setFeedbackScore: (value: string) => void;
  feedbackVerdict: string;
  setFeedbackVerdict: (value: string) => void;
  feedbackNotes: string;
  setFeedbackNotes: (value: string) => void;
  onSubmitFeedback: () => void;
}) {
  if (!props.selectedTraceId) {
    return <EmptyDetail title="Select a trace" body="Choose a trace attempt to inspect its timeline, visible reasoning, tools, Slack actions, evals, and operator feedback." />;
  }
  if (!props.traceDetail) {
    return <EmptyDetail title="Loading trace" body="Fetching the bounded evidence object for the selected trace." />;
  }

  const trace = props.traceDetail.trace;
  const inspectorTabs: TraceInspectorTab[] = ["overview", "timeline", "reasoning", "tools", "slack", "evals", "feedback", "proposals"];

  return (
    <div className="detail-card">
      <div className="detail-header">
        <div>
          <p className="eyebrow">Trace inspector</p>
          <h2>{trace.summary.trace_id}</h2>
        </div>
        <div className="button-row">
          <button onClick={props.onRunEval}>Run eval</button>
          <button className="secondary" onClick={props.onReplay}>Queue replay</button>
        </div>
      </div>
      <div className="segment-row">
        {inspectorTabs.map((tab) => (
          <button key={tab} className={props.tab === tab ? "segment-button active" : "segment-button"} onClick={() => props.setTab(tab)}>
            {tab}
          </button>
        ))}
      </div>

      {props.tab === "overview" ? (
        <div className="detail-section-body">
          <dl className="overview-grid">
            <div><dt>Conversation</dt><dd>{props.traceDetail.conversation.title || props.traceDetail.conversation.external_key}</dd></div>
            <div><dt>Case</dt><dd>{props.traceDetail.case?.title || trace.summary.case_id}</dd></div>
            <div><dt>Status</dt><dd>{trace.summary.status}</dd></div>
            <div><dt>Thread key</dt><dd>{trace.summary.thread_key}</dd></div>
            <div><dt>Events</dt><dd>{trace.summary.event_count}</dd></div>
            <div><dt>Reasoning</dt><dd>{trace.summary.reasoning_step_count}</dd></div>
            <div><dt>Tools</dt><dd>{trace.summary.tool_call_count}</dd></div>
            <div><dt>Slack</dt><dd>{trace.summary.slack_action_count}</dd></div>
          </dl>
          <div className="detail-card">
            <h3>Transcript slice used</h3>
            <div className="nested-list">
              {listOrEmpty(props.traceDetail.transcript_slice).map((entry) => (
                <div key={entry.id} className="nested-card">
                  <div className="detail-row-header">
                    <strong>{entry.actor_type || "actor"}</strong>
                    <small>{formatTime(entry.created_at)}</small>
                  </div>
                  <p className="detail-copy">{entry.body}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : null}

      {props.tab === "timeline" ? (
        <div className="detail-section-body">
          {listOrEmpty(trace.events).map((event) => (
            <div key={`${event.event_type}-${event.started_at}`} className="detail-row">
              <div className="detail-row-header">
                <strong>{event.event_type}</strong>
                <small>{event.status}</small>
              </div>
              <p className="detail-copy">{event.description || `${event.plane} · ${event.service} · ${event.actor}`}</p>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "reasoning" ? (
        <div className="detail-section-body">
          {listOrEmpty(trace.reasoning).map((step) => (
            <div key={step.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{step.step_type}</strong>
                <small>{step.confidence ? scoreBadge(step.confidence) : "n/a"}</small>
              </div>
              <p className="detail-copy">{step.summary}</p>
              {listOrEmpty(step.evidence_refs).length ? <p className="muted">Evidence: {listOrEmpty(step.evidence_refs).map((ref) => ref.summary || ref.ref).join(" • ")}</p> : null}
              {listOrEmpty(step.alternatives).length ? <p className="muted">Alternatives: {listOrEmpty(step.alternatives).join(" • ")}</p> : null}
              {step.decision ? <p className="muted">Decision: {step.decision}</p> : null}
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "tools" ? (
        <div className="detail-section-body">
          {listOrEmpty(trace.tool_calls).map((call) => (
            <div key={call.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{call.tool_name}</strong>
                <small>{call.approval_state || call.status}</small>
              </div>
              <p className="detail-copy">{call.summary || call.interpretation_summary}</p>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "slack" ? (
        <div className="detail-section-body">
          {listOrEmpty(trace.slack_actions).map((action) => (
            <div key={action.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{action.send_status || "draft"}</strong>
                <small>{formatTime(action.created_at)}</small>
              </div>
              <p className="detail-copy">{action.final_body || action.draft_body}</p>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "evals" ? (
        <div className="detail-section-body">
          {listOrEmpty(props.traceDetail.linked_eval_runs).map((run) => (
            <div key={run.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{run.suite_name}</strong>
                <small>{run.overall_verdict} · {scoreBadge(run.overall_score)}</small>
              </div>
              <p className="detail-copy">Triggered by {run.trigger} at {formatTime(run.created_at)}</p>
              <div className="nested-list">
                {listOrEmpty(props.traceJudgments[run.id]).map((judgment) => (
                  <div key={judgment.id} className="nested-card">
                    <div className="detail-row-header">
                      <strong>{judgment.layer}/{judgment.category}</strong>
                      <small>{judgment.passed ? "pass" : "needs work"} · {scoreBadge(judgment.score)}</small>
                    </div>
                    <p className="detail-copy">{judgment.rationale}</p>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "feedback" ? (
        <div className="review-grid">
          <div className="detail-card">
            <h3>Recorded feedback</h3>
            <div className="nested-list">
              {listOrEmpty(props.traceDetail.feedback_records).map((item) => (
                <div key={item.id} className="nested-card">
                  <div className="detail-row-header">
                    <strong>{item.target_type}</strong>
                    <small>{item.verdict || "no verdict"} · {item.score || 0}</small>
                  </div>
                  <p className="detail-copy">{item.notes || "No notes."}</p>
                </div>
              ))}
            </div>
          </div>
          <div className="detail-card">
            <h3>Add feedback</h3>
            <label className="field">
              Target
              <select
                value={`${props.feedbackTargetType}:${props.feedbackTargetID}`}
                onChange={(event) => {
                  const [type, value] = event.target.value.split(":");
                  props.setFeedbackTargetType(type);
                  props.setFeedbackTargetID(value);
                }}
              >
                {props.feedbackTargets.map((target) => (
                  <option key={`${target.type}:${target.value}`} value={`${target.type}:${target.value}`}>{target.label}</option>
                ))}
              </select>
            </label>
            <label className="field">
              Score
              <input value={props.feedbackScore} onChange={(event) => props.setFeedbackScore(event.target.value)} />
            </label>
            <label className="field">
              Verdict
              <input value={props.feedbackVerdict} onChange={(event) => props.setFeedbackVerdict(event.target.value)} />
            </label>
            <label className="field">
              Notes
              <textarea value={props.feedbackNotes} onChange={(event) => props.setFeedbackNotes(event.target.value)} />
            </label>
            <button onClick={props.onSubmitFeedback}>Submit feedback</button>
          </div>
        </div>
      ) : null}

      {props.tab === "proposals" ? (
        <div className="detail-section-body">
          {listOrEmpty(props.traceDetail.linked_proposals).map((proposal) => (
            <div key={proposal.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{proposal.title}</strong>
                <small>{proposal.status}</small>
              </div>
              <p className="detail-copy">{proposal.summary}</p>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}

function EmptyDetail(props: { title: string; body: string }) {
  return (
    <div className="empty-detail">
      <p className="eyebrow">Detail</p>
      <h2>{props.title}</h2>
      <p className="muted">{props.body}</p>
    </div>
  );
}
