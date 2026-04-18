import type { TraceDetailResponse, TraceInspectorTab, NullableList, EvalJudgment } from "@/types";
import { formatTime, latestActionResult, listOrEmpty, scoreBadge } from "@/hooks/api";
import { EmptyDetail } from "./empty-detail";
import { FormattedMessage } from "@/components/formatted-message";

export function TraceInspector(props: {
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

  const traceDetail = props.traceDetail;
  const trace = traceDetail.trace;
  const inspectorTabs: TraceInspectorTab[] = ["overview", "timeline", "reasoning", "tools", "actions", "slack", "outcomes", "evals", "feedback", "proposals"];

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
            <div><dt>Actions</dt><dd>{listOrEmpty(props.traceDetail.action_intents).length}</dd></div>
            <div><dt>Outcomes</dt><dd>{listOrEmpty(props.traceDetail.outcomes).length}</dd></div>
            <div><dt>Knowledge</dt><dd>{listOrEmpty(props.traceDetail.knowledge_entries).length}</dd></div>
            <div><dt>Harness runs</dt><dd>{listOrEmpty(props.traceDetail.harness_executions).length}</dd></div>
          </dl>
          {props.traceDetail.workflow_line ? (
            <div className="detail-card">
              <h3>Workflow line</h3>
              <dl className="overview-grid">
                <div><dt>Status</dt><dd>{props.traceDetail.workflow_line.status}</dd></div>
                <div><dt>Current attempt</dt><dd>{props.traceDetail.workflow_line.current_workflow_id || "none"}</dd></div>
                <div><dt>Retry budget</dt><dd>{props.traceDetail.workflow_line.auto_retry_budget_remaining}</dd></div>
                <div><dt>Last failure</dt><dd>{props.traceDetail.workflow_line.last_failure_class || "none"}</dd></div>
                <div><dt>Retry at</dt><dd>{props.traceDetail.workflow_line.retry_after ? formatTime(props.traceDetail.workflow_line.retry_after) : "none"}</dd></div>
                <div><dt>Next action</dt><dd>{props.traceDetail.workflow_line.next_retry_action || "none"}</dd></div>
              </dl>
            </div>
          ) : null}
          {listOrEmpty(props.traceDetail.workflow_attempts).length ? (
            <div className="detail-card">
              <h3>Attempt lineage</h3>
              <div className="nested-list">
                {listOrEmpty(props.traceDetail.workflow_attempts).map((attempt) => (
                  <div key={attempt.workflow_id} className="nested-card">
                    <div className="detail-row-header">
                      <strong>{attempt.workflow_id}</strong>
                      <small>attempt {attempt.attempt_number} · {attempt.status}</small>
                    </div>
                    <p className="detail-copy">{attempt.trace_id || "No linked trace."}</p>
                    <p className="muted">
                      Parent: {attempt.parent_workflow_id || "none"} · Supersedes trace: {attempt.supersedes_trace_id || "none"}
                    </p>
                    <p className="muted">
                      Failure: {attempt.failure_class || "none"} · Retry: {attempt.retry_decision || "none"} · Repair: {attempt.repair_attempted ? (attempt.repair_succeeded ? "succeeded" : "failed") : "not needed"}
                    </p>
                  </div>
                ))}
              </div>
            </div>
          ) : null}
          {listOrEmpty(props.traceDetail.harness_executions).length ? (
            <div className="detail-card">
              <h3>Hermes session continuity</h3>
              <div className="nested-list">
                {listOrEmpty(props.traceDetail.harness_executions).map((item) => (
                  <div key={item.id} className="nested-card">
                    <div className="detail-row-header">
                      <strong>{item.role}</strong>
                      <small>{formatTime(item.created_at)}</small>
                    </div>
                    <p className="detail-copy">{item.hermes_session_id}</p>
                    <p className="muted">Scope: {item.session_scope_kind}:{item.session_scope_id}</p>
                    {item.parent_session_id ? <p className="muted">Parent session: {item.parent_session_id}</p> : null}
                    {listOrEmpty(item.memory_reads).length ? <p className="muted">Reads: {listOrEmpty(item.memory_reads).map((memory) => memory.summary).join(" • ")}</p> : null}
                    {listOrEmpty(item.memory_writes).length ? <p className="muted">Writes: {listOrEmpty(item.memory_writes).map((memory) => memory.summary).join(" • ")}</p> : null}
                  </div>
                ))}
              </div>
            </div>
          ) : null}
          <div className="detail-card">
            <h3>Transcript slice used</h3>
            <div className="nested-list">
              {listOrEmpty(props.traceDetail.transcript_slice).map((entry) => (
                <div key={entry.id} className="nested-card">
                  <div className="detail-row-header">
                    <strong>{entry.actor_type || "actor"}</strong>
                    <small>{formatTime(entry.created_at)}</small>
                  </div>
                  <p className="detail-copy">
                    <FormattedMessage source={entry.source} text={entry.body} metadata={entry.metadata} />
                  </p>
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

      {props.tab === "actions" ? (
        <div className="detail-section-body">
          {listOrEmpty(traceDetail.action_intents).map((intent) => {
            const result = latestActionResult(intent.id, traceDetail.action_results);
            return (
              <div key={intent.id} className="detail-row">
                <div className="detail-row-header">
                  <strong>{intent.kind}</strong>
                  <small>{result?.status || intent.status}</small>
                </div>
                <p className="detail-copy">{intent.rationale || intent.target_ref || "No rationale recorded."}</p>
                <p className="muted">{intent.target_ref || intent.policy_verdict || "No target reference."}</p>
                {result?.provider ? <p className="muted">Provider: {result.provider}{result.provider_ref ? ` · ${result.provider_ref}` : ""}</p> : null}
                {result?.error_message ? <p className="muted">Error: {result.error_message}</p> : null}
              </div>
            );
          })}
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
              <p className="detail-copy">
                <FormattedMessage source="slack" text={action.final_body || action.draft_body} />
              </p>
            </div>
          ))}
        </div>
      ) : null}

      {props.tab === "outcomes" ? (
        <div className="detail-section-body">
          {listOrEmpty(props.traceDetail.outcomes).map((item) => (
            <div key={item.id} className="detail-row">
              <div className="detail-row-header">
                <strong>{item.outcome_type}</strong>
                <small>{item.verdict} · {scoreBadge(item.score)}</small>
              </div>
              <p className="detail-copy">{item.summary || item.details || "No summary."}</p>
              <p className="muted">{item.source} · {formatTime(item.recorded_at)}</p>
            </div>
          ))}
          {listOrEmpty(props.traceDetail.knowledge_entries).length ? (
            <div className="detail-card">
              <h3>Related knowledge</h3>
              <div className="nested-list">
                {listOrEmpty(props.traceDetail.knowledge_entries).map((item) => (
                  <div key={item.id} className="nested-card">
                    <div className="detail-row-header">
                      <strong>{item.title}</strong>
                      <small>{item.status} · {item.tier}</small>
                    </div>
                    <p className="detail-copy">{item.summary || item.body || "No summary."}</p>
                  </div>
                ))}
              </div>
            </div>
          ) : null}
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
