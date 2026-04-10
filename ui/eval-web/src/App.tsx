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

type Proposal = {
  id: string;
  trace_id: string;
  title: string;
  category: string;
  summary: string;
  status: string;
  reviewer?: string;
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
  const [selectedTraceId, setSelectedTraceId] = useState<string>("trace-oncall-001");
  const [ratingNotes, setRatingNotes] = useState("");
  const [improvementNote, setImprovementNote] = useState("");

  const tracesQuery = useQuery({
    queryKey: ["traces"],
    queryFn: () => getJSON<{ traces: TraceSummary[] }>("/api/traces")
  });

  const traceIds = tracesQuery.data?.traces ?? [];
  const activeTraceId = useMemo(() => {
    if (traceIds.some((trace) => trace.trace_id === selectedTraceId)) {
      return selectedTraceId;
    }
    return traceIds[0]?.trace_id ?? "trace-oncall-001";
  }, [selectedTraceId, traceIds]);

  const traceQuery = useQuery({
    queryKey: ["trace", activeTraceId],
    queryFn: () => getJSON<Trace>(`/api/traces/${activeTraceId}`),
    enabled: Boolean(activeTraceId)
  });

  const proposalsQuery = useQuery({
    queryKey: ["proposals"],
    queryFn: () => getJSON<{ proposals: Proposal[] }>("/api/proposals")
  });

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
      await queryClient.invalidateQueries({ queryKey: ["traces"] });
      await queryClient.invalidateQueries({ queryKey: ["trace", activeTraceId] });
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
      await queryClient.invalidateQueries({ queryKey: ["trace", activeTraceId] });
    }
  });

  const replayMutation = useMutation({
    mutationFn: () =>
      postJSON(`/api/traces/${activeTraceId}/replay`, { requested_by: "ui-operator" }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["traces"] });
      await queryClient.invalidateQueries({ queryKey: ["trace", activeTraceId] });
    }
  });

  const proposalDecisionMutation = useMutation({
    mutationFn: (proposal: Proposal) =>
      postJSON(`/api/proposals/${proposal.id}/decision`, {
        decision: "approved",
        rationale: "Looks safe to promote after replay.",
        reviewer_id: "ui-operator"
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["proposals"] });
    }
  });

  return (
    <div className="shell">
      <aside className="sidebar">
        <div>
          <p className="eyebrow">Improvement Plane</p>
          <h1>Trace Review</h1>
          <p className="muted">
            Inspect ingestion timelines, review proposals, and rate outcome quality.
          </p>
        </div>

        <div className="panel">
          <h2>Traces</h2>
          {tracesQuery.isLoading ? <p>Loading traces...</p> : null}
          <ul className="trace-list">
            {traceIds.map((trace) => (
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
          <h2>Proposals</h2>
          <ul className="proposal-list">
            {(proposalsQuery.data?.proposals ?? []).map((proposal) => (
              <li key={proposal.id}>
                <div>
                  <strong>{proposal.title}</strong>
                  <p>{proposal.summary}</p>
                  <small>{proposal.status}</small>
                </div>
                <button onClick={() => proposalDecisionMutation.mutate(proposal)}>Approve</button>
              </li>
            ))}
          </ul>
        </div>
      </aside>

      <main className="content">
        <section className="hero">
          <div>
            <p className="eyebrow">End-to-End Trace</p>
            <h2>{traceQuery.data?.summary.trace_id ?? activeTraceId}</h2>
          </div>
          <button className="ghost" onClick={() => replayMutation.mutate()}>
            Queue replay
          </button>
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
                <select name="verdict" defaultValue="correct">
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
                <input name="score" type="number" min={1} max={5} defaultValue={4} />
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

        <section className="panel">
          <h3>Improvement Note</h3>
          <div className="stack">
            <textarea
              value={improvementNote}
              onChange={(event) => setImprovementNote(event.target.value)}
              placeholder="Capture prompt, workflow, policy, tooling, or platform-bug observations."
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

