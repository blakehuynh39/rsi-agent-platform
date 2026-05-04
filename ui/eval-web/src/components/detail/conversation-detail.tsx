import { useEffect, useState } from "react";

import type { ConversationDetailResponse, TraceDetailResponse, TraceInspectorTab, NullableList, EvalJudgment, SelfReviewCadence } from "@/types";
import { formatTime, listOrEmpty, pageCount, clampPage } from "@/hooks/api";
import { TraceInspector } from "./trace-inspector";
import { FormattedMessage } from "@/components/formatted-message";

const TRANSCRIPT_PAGE_SIZE = 5;
const WORKFLOW_ATTEMPT_PAGE_SIZE = 3;

function PageControls(props: {
  label: string;
  page: number;
  total: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}) {
  const pages = pageCount(props.total, props.pageSize);
  if (props.total <= props.pageSize) {
    return <span className="list-range">{props.total} total</span>;
  }
  const start = (props.page - 1) * props.pageSize + 1;
  const end = Math.min(props.total, props.page * props.pageSize);
  return (
    <div className="pagination-row compact" aria-label={`${props.label} pages`}>
      <span>{start}-{end} / {props.total}</span>
      <button
        className="pager-button"
        aria-label={`Previous ${props.label} page`}
        onClick={() => props.onPageChange(clampPage(props.page - 1, props.total, props.pageSize))}
        disabled={props.page <= 1}
      >
        Prev
      </button>
      <button
        className="pager-button"
        aria-label={`Next ${props.label} page`}
        onClick={() => props.onPageChange(clampPage(props.page + 1, props.total, props.pageSize))}
        disabled={props.page >= pages}
      >
        Next
      </button>
    </div>
  );
}

function cadenceValue(current?: number, threshold?: number) {
  if (typeof current !== "number") {
    return "n/a";
  }
  if (typeof threshold === "number" && threshold > 0) {
    return `${current} / ${threshold}`;
  }
  return String(current);
}

function cadencePercent(current?: number, threshold?: number) {
  if (typeof current !== "number" || typeof threshold !== "number" || threshold <= 0) {
    return 0;
  }
  return Math.max(0, Math.min(100, (current / threshold) * 100));
}

function SelfReviewCadencePanel(props: { cadence?: SelfReviewCadence }) {
  const cadence = props.cadence;
  if (!cadence) {
    return null;
  }
  const summary = cadence.latest_summary || cadence.latest_error || "No self-review actions recorded yet.";
  const status = cadence.review_status || cadence.candidate_status || "observed";
  return (
    <section className="stream-section self-review-cadence-panel">
      <div className="card-section-header">
        <div>
          <h3>Self-review cadence</h3>
          <p className="muted">{cadence.cadence_scope_key || cadence.gateway_session_key || cadence.execution_id || "conversation scope"}</p>
        </div>
        <span className={cadence.latest_error ? "status-chip warn" : "status-chip"}>{status}</span>
      </div>
      <div className="cadence-grid">
        <div className="cadence-meter">
          <div className="cadence-meter-header">
            <span>Memory</span>
            <strong>{cadenceValue(cadence.memory_iterations, cadence.memory_threshold)}</strong>
          </div>
          <div className="cadence-track" aria-hidden="true">
            <span style={{ width: `${cadencePercent(cadence.memory_iterations, cadence.memory_threshold)}%` }} />
          </div>
        </div>
        <div className="cadence-meter">
          <div className="cadence-meter-header">
            <span>Skills</span>
            <strong>{cadenceValue(cadence.skill_iterations, cadence.skill_threshold)}</strong>
          </div>
          <div className="cadence-track" aria-hidden="true">
            <span style={{ width: `${cadencePercent(cadence.skill_iterations, cadence.skill_threshold)}%` }} />
          </div>
        </div>
      </div>
      <dl className="mini-metrics cadence-details">
        <div><dt>Candidate</dt><dd>{cadence.candidate_id || "none"}</dd></div>
        <div><dt>Trigger</dt><dd>{cadence.trigger_kind || cadence.review_kind || listOrEmpty(cadence.work_created).join(", ") || "none"}</dd></div>
        <div><dt>Execution</dt><dd>{cadence.execution_id || "unknown"}</dd></div>
        <div><dt>Updated</dt><dd>{cadence.updated_at ? formatTime(cadence.updated_at) : "unknown"}</dd></div>
      </dl>
      <p className="detail-copy cadence-summary">{summary}</p>
    </section>
  );
}

export function ConversationDetail(props: {
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
  const workflowAttempts = listOrEmpty(props.detail.workflow_attempts);
  const workflowLine = props.detail.workflow_line;
  const [transcriptPage, setTranscriptPage] = useState(1);
  const [workflowAttemptPage, setWorkflowAttemptPage] = useState(1);
  const transcriptPageCount = pageCount(transcript.length, TRANSCRIPT_PAGE_SIZE);
  const workflowAttemptPageCount = pageCount(workflowAttempts.length, WORKFLOW_ATTEMPT_PAGE_SIZE);
  const clampedTranscriptPage = Math.min(transcriptPage, transcriptPageCount);
  const clampedWorkflowAttemptPage = Math.min(workflowAttemptPage, workflowAttemptPageCount);
  const visibleTranscript = transcript.slice((clampedTranscriptPage - 1) * TRANSCRIPT_PAGE_SIZE, clampedTranscriptPage * TRANSCRIPT_PAGE_SIZE);
  const visibleWorkflowAttempts = workflowAttempts.slice(
    (clampedWorkflowAttemptPage - 1) * WORKFLOW_ATTEMPT_PAGE_SIZE,
    clampedWorkflowAttemptPage * WORKFLOW_ATTEMPT_PAGE_SIZE
  );

  useEffect(() => {
    setTranscriptPage(1);
    setWorkflowAttemptPage(1);
  }, [props.detail.conversation.id]);

  useEffect(() => {
    setTranscriptPage((current) => Math.min(current, transcriptPageCount));
  }, [transcriptPageCount]);

  useEffect(() => {
    setWorkflowAttemptPage((current) => Math.min(current, workflowAttemptPageCount));
  }, [workflowAttemptPageCount]);

  return (
    <div className="detail-stack conversation-workspace">
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

      <main className="conversation-stream">
        <div className="stream-hero">
          <div>
            <p className="eyebrow">Conversation stream</p>
            <h2>{props.detail.conversation.title || props.detail.conversation.external_key}</h2>
            <p className="muted">{props.detail.conversation.external_key}</p>
          </div>
          <div className="stream-status-strip">
            <span className="status-chip">{props.detail.conversation.status}</span>
            <span className="status-chip">{props.detail.conversation.source}</span>
            {workflowLine ? <span className={workflowLine.last_failure_class ? "status-chip warn" : "status-chip"}>{workflowLine.last_failure_class || workflowLine.status}</span> : null}
          </div>
        </div>

        <div className="stream-summary-grid">
          <div><span>Active case</span><strong>{props.detail.active_case?.title || "none"}</strong></div>
          <div><span>Line</span><strong>{workflowLine?.status || "none"}</strong></div>
          <div><span>Retry budget</span><strong>{workflowLine?.auto_retry_budget_remaining ?? "n/a"}</strong></div>
          <div><span>Proposals</span><strong>{listOrEmpty(props.detail.linked_proposals).length}</strong></div>
        </div>

        <SelfReviewCadencePanel cadence={props.detail.self_review_cadence} />

        <section className="stream-section">
          <div className="card-section-header">
            <div>
              <h3>Transcript</h3>
              <p className="muted">Most recent {transcript.length} entries</p>
            </div>
            <PageControls
              label="transcript"
              page={clampedTranscriptPage}
              total={transcript.length}
              pageSize={TRANSCRIPT_PAGE_SIZE}
              onPageChange={setTranscriptPage}
            />
          </div>
          <div className="nested-list">
            {visibleTranscript.map((entry) => (
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
            {visibleTranscript.length === 0 ? <div className="empty-list compact">No transcript entries.</div> : null}
          </div>
        </section>

        <section className="stream-section">
          <div className="card-section-header">
            <div>
              <h3>Workflow attempts</h3>
              <p className="muted">{workflowAttempts.length} recorded attempts</p>
            </div>
            <PageControls
              label="workflow attempts"
              page={clampedWorkflowAttemptPage}
              total={workflowAttempts.length}
              pageSize={WORKFLOW_ATTEMPT_PAGE_SIZE}
              onPageChange={setWorkflowAttemptPage}
            />
          </div>
          <div className="nested-list">
            {visibleWorkflowAttempts.map((attempt) => (
              <button key={attempt.workflow_id} className={attempt.trace_id === props.selectedTraceId ? "list-card selected" : "list-card"} onClick={() => attempt.trace_id ? props.onSelectTrace(attempt.trace_id) : undefined}>
                <div className="list-card-header">
                  <div className="card-title-block">
                    <strong>{attempt.workflow_id}</strong>
                    <p>attempt {attempt.attempt_number} · {attempt.status}</p>
                  </div>
                  {attempt.failure_class ? <span className="status-chip eval">{attempt.failure_class}</span> : null}
                </div>
                <dl className="mini-metrics">
                  <div><dt>Started</dt><dd>{formatTime(attempt.created_at)}</dd></div>
                  <div><dt>Trace</dt><dd>{attempt.trace_id || "none"}</dd></div>
                  <div><dt>Retry</dt><dd>{attempt.retry_decision || "none"}</dd></div>
                  <div><dt>Repair</dt><dd>{attempt.repair_attempted ? (attempt.repair_succeeded ? "succeeded" : "failed") : "not needed"}</dd></div>
                </dl>
              </button>
            ))}
            {visibleWorkflowAttempts.length === 0 ? <div className="empty-list compact">No workflow attempts.</div> : null}
          </div>
        </section>
      </main>
    </div>
  );
}
