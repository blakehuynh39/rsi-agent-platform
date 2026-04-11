import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { FormEvent, useMemo, useState } from "react";

type TraceSummary = {
  trace_id: string;
  ingestion_id: string;
  workflow_id: string;
  thread_key: string;
  workflow_kind: string;
  status: string;
  last_verdict?: string;
  started_at: string;
  ended_at: string;
  event_count: number;
  artifact_count: number;
};

type TraceEvent = {
  plane: string;
  service: string;
  actor: string;
  event_type: string;
  status: string;
  description?: string;
  started_at: string;
  ended_at?: string;
};

type Artifact = {
  id: string;
  kind: string;
  url: string;
  source: string;
};

type Trace = {
  summary: TraceSummary;
  events: TraceEvent[];
  artifacts: Artifact[];
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
  new_evidence_since_last_rejection: boolean;
  prior_similar_proposal_ids: string[];
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
  risk_tier?: string;
  review_deadline?: string;
  active_slot_consuming: boolean;
  prior_similar_proposal_ids: string[];
  new_evidence_since_last_rejection: boolean;
};

type ProposalMemory = {
  id: string;
  proposal_id: string;
  candidate_key: string;
  disposition: string;
  review_rationale: string;
  failure_class?: string;
  created_at: string;
};

type ProposalSlots = {
  cap: number;
  active: number;
  available: number;
  active_proposal_ids: string[];
  stale_proposal_ids: string[];
};

type RepoChangeJob = {
  id: string;
  proposal_id: string;
  status: string;
  branch_name: string;
  context_summary: string;
};

type PRAttempt = {
  id: string;
  proposal_id: string;
  pr_url?: string;
  status: string;
  validation_status: string;
};

type PostMergeReplay = {
  id: string;
  proposal_id: string;
  trace_id: string;
  baseline_score: number;
  candidate_score: number;
  improved: boolean;
};

type ProposalResponse = {
  proposals: Proposal[];
  proposal_slots: ProposalSlots;
  candidates: Candidate[];
  proposal_memory: ProposalMemory[];
  repo_change_jobs: RepoChangeJob[];
  pr_attempts: PRAttempt[];
  post_merge_replays: PostMergeReplay[];
};

type EvalResponse = {
  eval_runs: EvalRun[];
  judgments: Record<string, EvalJudgment[]>;
};

async function getJSON<T>(url: string): Promise<T> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status}`);
  }
  return response.json();
}

async function postJSON<T>(url: string, body: Record<string, unknown>): Promise<T> {
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  });
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status}`);
  }
  return response.json();
}

export function App() {
  const queryClient = useQueryClient();
  const [selectedTraceId, setSelectedTraceId] = useState("");
  const [ratingNotes, setRatingNotes] = useState("");
  const [improvementNote, setImprovementNote] = useState("");

  const tracesQuery = useQuery({
    queryKey: ["traces"],
    queryFn: () => getJSON<{ traces: TraceSummary[] }>("/api/traces")
  });

  const activeTraceId = useMemo(() => {
    const traceIds = tracesQuery.data?.traces ?? [];
    if (selectedTraceId && traceIds.some((trace) => trace.trace_id === selectedTraceId)) {
      return selectedTraceId;
    }
    return traceIds[0]?.trace_id ?? "";
  }, [selectedTraceId, tracesQuery.data?.traces]);

  const traceQuery = useQuery({
    queryKey: ["trace", activeTraceId],
    queryFn: () => getJSON<Trace>(`/api/traces/${activeTraceId}`),
    enabled: Boolean(activeTraceId)
  });

  const proposalsQuery = useQuery({
    queryKey: ["proposals"],
    queryFn: () => getJSON<ProposalResponse>("/api/proposals")
  });

  const evalsQuery = useQuery({
    queryKey: ["evals"],
    queryFn: () => getJSON<EvalResponse>("/api/evals")
  });

  const traceEvals = useMemo(() => {
    const evalRuns = evalsQuery.data?.eval_runs ?? [];
    return evalRuns.filter((run) => run.trace_id === activeTraceId).slice(0, 3);
  }, [activeTraceId, evalsQuery.data?.eval_runs]);

  const slotState = proposalsQuery.data?.proposal_slots;

  const refreshEverything = async () => {
    await queryClient.invalidateQueries({ queryKey: ["traces"] });
    await queryClient.invalidateQueries({ queryKey: ["trace", activeTraceId] });
    await queryClient.invalidateQueries({ queryKey: ["proposals"] });
    await queryClient.invalidateQueries({ queryKey: ["evals"] });
  };

  const ratingMutation = useMutation({
    mutationFn: async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      return postJSON(`/api/traces/${activeTraceId}/ratings`, {
        score: Number(form.get("score") ?? 3),
        verdict: form.get("verdict"),
        labels: ["ui-review"],
        notes: ratingNotes,
        reviewer_id: "ui-operator"
      });
    },
    onSuccess: async () => {
      setRatingNotes("");
      await refreshEverything();
    }
  });

  const noteMutation = useMutation({
    mutationFn: () =>
      postJSON(`/api/traces/${activeTraceId}/notes`, {
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

  const replayMutation = useMutation({
    mutationFn: () => postJSON(`/api/traces/${activeTraceId}/replay`, { requested_by: "ui-operator" }),
    onSuccess: refreshEverything
  });

  const evaluateMutation = useMutation({
    mutationFn: () => postJSON(`/api/traces/${activeTraceId}/evaluate`, {}),
    onSuccess: refreshEverything
  });

  const promoteMutation = useMutation({
    mutationFn: () => postJSON(`/api/proposals/promote`, { requested_by: "ui-operator" }),
    onSuccess: refreshEverything
  });

  const proposalDecisionMutation = useMutation({
    mutationFn: ({ proposal, decision }: { proposal: Proposal; decision: string }) =>
      postJSON(`/api/proposals/${proposal.id}/decision`, {
        decision,
        rationale: `UI review recorded ${decision}.`,
        reviewer_id: "ui-operator",
        failure_class: decision === "rejected" ? "insufficient_evidence" : ""
      }),
    onSuccess: refreshEverything
  });

  return (
    <div className="shell">
      <aside className="sidebar">
        <div>
          <p className="eyebrow">Improvement Plane</p>
          <h1>Recursive Eval Control</h1>
          <p className="muted">
            Review traces, inspect eval history, and gate proposal slots before the agent can open more repo changes.
          </p>
        </div>

        <div className="panel">
          <h2>Proposal Slots</h2>
          <dl className="summary compact">
            <div><dt>Cap</dt><dd>{slotState?.cap ?? 0}</dd></div>
            <div><dt>Active</dt><dd>{slotState?.active ?? 0}</dd></div>
            <div><dt>Available</dt><dd>{slotState?.available ?? 0}</dd></div>
            <div><dt>Stale</dt><dd>{slotState?.stale_proposal_ids.length ?? 0}</dd></div>
          </dl>
          <button onClick={() => promoteMutation.mutate()} disabled={(slotState?.available ?? 0) === 0}>
            Run proposal promoter
          </button>
        </div>

        <div className="panel">
          <h2>Traces</h2>
          <ul className="trace-list">
            {(tracesQuery.data?.traces ?? []).map((trace) => (
              <li key={trace.trace_id}>
                <button
                  className={trace.trace_id === activeTraceId ? "trace-button active" : "trace-button"}
                  onClick={() => setSelectedTraceId(trace.trace_id)}
                >
                  <span>{trace.workflow_kind}</span>
                  <strong>{trace.trace_id}</strong>
                  <small>{trace.status}</small>
                </button>
              </li>
            ))}
          </ul>
        </div>

        <div className="panel">
          <h2>Queued Candidates</h2>
          <ul className="candidate-list">
            {(proposalsQuery.data?.candidates ?? []).slice(0, 6).map((candidate) => (
              <li key={candidate.id}>
                <div>
                  <strong>{candidate.subsystem}</strong>
                  <p>{candidate.failure_mode}</p>
                  <small>{candidate.status}</small>
                </div>
                <span className="pill score">{candidate.priority_score.toFixed(2)}</span>
              </li>
            ))}
          </ul>
        </div>
      </aside>

      <main className="content">
        <section className="hero">
          <div>
            <p className="eyebrow">Trace Detail</p>
            <h2>{traceQuery.data?.summary.trace_id ?? (activeTraceId || "No trace selected")}</h2>
          </div>
          <div className="actions">
            <button className="ghost" onClick={() => evaluateMutation.mutate()} disabled={!activeTraceId}>
              Run eval
            </button>
            <button className="ghost" onClick={() => replayMutation.mutate()} disabled={!activeTraceId}>
              Queue replay
            </button>
          </div>
        </section>

        <section className="grid">
          <article className="panel">
            <h3>Summary</h3>
            {traceQuery.data ? (
              <dl className="summary">
                <div><dt>Workflow</dt><dd>{traceQuery.data.summary.workflow_kind}</dd></div>
                <div><dt>Status</dt><dd>{traceQuery.data.summary.status}</dd></div>
                <div><dt>Ingestion</dt><dd>{traceQuery.data.summary.ingestion_id}</dd></div>
                <div><dt>Thread</dt><dd>{traceQuery.data.summary.thread_key}</dd></div>
                <div><dt>Events</dt><dd>{traceQuery.data.summary.event_count}</dd></div>
                <div><dt>Artifacts</dt><dd>{traceQuery.data.summary.artifact_count}</dd></div>
              </dl>
            ) : (
              <p>Loading trace...</p>
            )}
          </article>

          <article className="panel">
            <h3>Human Rating</h3>
            <form
              onSubmit={(event) => {
                void ratingMutation.mutateAsync(event);
              }}
              className="stack"
            >
              <label>
                Verdict
                <select name="verdict" defaultValue="partial">
                  <option value="correct">correct</option>
                  <option value="partial">partial</option>
                  <option value="wrong">wrong</option>
                  <option value="unsafe">unsafe</option>
                  <option value="spam">spam</option>
                  <option value="needs-human">needs-human</option>
                </select>
              </label>
              <label>
                Score
                <input name="score" type="number" min={1} max={5} defaultValue={3} />
              </label>
              <label>
                Notes
                <textarea value={ratingNotes} onChange={(event) => setRatingNotes(event.target.value)} />
              </label>
              <button type="submit">Submit rating</button>
            </form>
          </article>
        </section>

        <section className="grid">
          <article className="panel">
            <h3>Recent Evals</h3>
            <ul className="eval-list">
              {traceEvals.map((run) => (
                <li key={run.id}>
                  <div>
                    <strong>{run.suite_name}</strong>
                    <p>{run.trigger}</p>
                    <small>{run.overall_verdict}</small>
                  </div>
                  <span className="pill score">{run.overall_score.toFixed(2)}</span>
                </li>
              ))}
            </ul>
            {traceEvals[0] ? (
              <div className="stack detail-block">
                {(evalsQuery.data?.judgments?.[traceEvals[0].id] ?? []).map((judgment) => (
                  <div key={judgment.id} className="judgment">
                    <strong>{judgment.layer}</strong>
                    <p>{judgment.category}</p>
                    <small>{judgment.rationale}</small>
                  </div>
                ))}
              </div>
            ) : (
              <p className="muted">No eval runs yet.</p>
            )}
          </article>

          <article className="panel">
            <h3>Active Proposals</h3>
            <ul className="proposal-list">
              {(proposalsQuery.data?.proposals ?? []).map((proposal) => (
                <li key={proposal.id}>
                  <div>
                    <strong>{proposal.title}</strong>
                    <p>{proposal.summary}</p>
                    <small>{proposal.status}</small>
                  </div>
                  <div className="stack inline-actions">
                    <button onClick={() => proposalDecisionMutation.mutate({ proposal, decision: "approved" })}>
                      Approve
                    </button>
                    <button onClick={() => proposalDecisionMutation.mutate({ proposal, decision: "dismissed" })}>
                      Dismiss
                    </button>
                    <button onClick={() => proposalDecisionMutation.mutate({ proposal, decision: "rejected" })}>
                      Reject
                    </button>
                    <button onClick={() => proposalDecisionMutation.mutate({ proposal, decision: "merged" })}>
                      Mark merged
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </article>
        </section>

        <section className="grid">
          <article className="panel">
            <h3>Timeline</h3>
            <ul className="timeline">
              {(traceQuery.data?.events ?? []).map((event, index) => (
                <li key={`${event.event_type}-${index}`}>
                  <span className="pill">{event.plane}</span>
                  <div>
                    <strong>{event.event_type}</strong>
                    <p>{event.description ?? `${event.actor} via ${event.service}`}</p>
                  </div>
                  <small>{event.status}</small>
                </li>
              ))}
            </ul>
          </article>

          <article className="panel">
            <h3>Artifacts</h3>
            <ul className="artifact-list">
              {(traceQuery.data?.artifacts ?? []).map((artifact) => (
                <li key={artifact.id}>
                  <strong>{artifact.kind}</strong>
                  <a href={artifact.url}>{artifact.id}</a>
                  <small>{artifact.source}</small>
                </li>
              ))}
            </ul>
          </article>
        </section>

        <section className="grid">
          <article className="panel">
            <h3>Proposal Memory</h3>
            <ul className="memory-list">
              {(proposalsQuery.data?.proposal_memory ?? []).slice(0, 6).map((memory) => (
                <li key={memory.id}>
                  <div>
                    <strong>{memory.candidate_key}</strong>
                    <p>{memory.disposition}</p>
                    <small>{memory.review_rationale}</small>
                  </div>
                </li>
              ))}
            </ul>
          </article>

          <article className="panel">
            <h3>Repo Change Context</h3>
            <ul className="memory-list">
              {(proposalsQuery.data?.repo_change_jobs ?? []).slice(0, 4).map((job) => (
                <li key={job.id}>
                  <div>
                    <strong>{job.branch_name}</strong>
                    <p>{job.status}</p>
                    <small>{job.context_summary}</small>
                  </div>
                </li>
              ))}
              {(proposalsQuery.data?.pr_attempts ?? []).slice(0, 4).map((attempt) => (
                <li key={attempt.id}>
                  <div>
                    <strong>{attempt.proposal_id}</strong>
                    <p>{attempt.status}</p>
                    <small>{attempt.pr_url ?? attempt.validation_status}</small>
                  </div>
                </li>
              ))}
            </ul>
          </article>
        </section>

        <section className="panel">
          <h3>Improvement Note</h3>
          <div className="stack">
            <textarea
              value={improvementNote}
              onChange={(event) => setImprovementNote(event.target.value)}
              placeholder="Capture prompt, workflow, policy, or architecture observations."
            />
            <button onClick={() => noteMutation.mutate()} disabled={!improvementNote.trim()}>
              Save improvement note
            </button>
          </div>
        </section>
      </main>
    </div>
  );
}
