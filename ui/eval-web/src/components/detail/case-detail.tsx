import type { CaseDetailResponse, TraceDetailResponse, TraceInspectorTab, NullableList, EvalJudgment } from "@/types";
import { formatTime, listOrEmpty } from "@/hooks/api";
import { TraceInspector } from "./trace-inspector";

export function CaseDetail(props: {
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
  const workflowAttempts = listOrEmpty(props.detail.workflow_attempts);
  const workflowLine = props.detail.workflow_line;
  return (
    <div className="detail-stack">
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
        {workflowLine ? (
          <dl className="overview-grid">
            <div><dt>Line status</dt><dd>{workflowLine.status}</dd></div>
            <div><dt>Current attempt</dt><dd>{workflowLine.current_workflow_id || "none"}</dd></div>
            <div><dt>Retry budget</dt><dd>{workflowLine.auto_retry_budget_remaining}</dd></div>
            <div><dt>Retry at</dt><dd>{workflowLine.retry_after ? formatTime(workflowLine.retry_after) : "none"}</dd></div>
          </dl>
        ) : null}
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
  );
}
