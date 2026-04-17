import type { ConversationDetailResponse, TraceDetailResponse, TraceInspectorTab, NullableList, EvalJudgment } from "@/types";
import { formatTime, listOrEmpty } from "@/hooks/api";
import { TraceInspector } from "./trace-inspector";

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
          <div><dt>Workflow attempts</dt><dd>{workflowAttempts.length}</dd></div>
          <div><dt>Linked proposals</dt><dd>{listOrEmpty(props.detail.linked_proposals).length}</dd></div>
        </dl>
        {workflowLine ? (
          <dl className="overview-grid">
            <div><dt>Line status</dt><dd>{workflowLine.status}</dd></div>
            <div><dt>Current attempt</dt><dd>{workflowLine.current_workflow_id || "none"}</dd></div>
            <div><dt>Retry budget</dt><dd>{workflowLine.auto_retry_budget_remaining}</dd></div>
            <div><dt>Last failure</dt><dd>{workflowLine.last_failure_class || "none"}</dd></div>
          </dl>
        ) : null}
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
          <h3>Workflow attempts</h3>
          <div className="nested-list">
            {workflowAttempts.map((attempt) => (
              <button key={attempt.workflow_id} className={attempt.trace_id === props.selectedTraceId ? "list-card selected" : "list-card"} onClick={() => attempt.trace_id ? props.onSelectTrace(attempt.trace_id) : undefined}>
                <div className="list-card-header">
                  <div>
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
