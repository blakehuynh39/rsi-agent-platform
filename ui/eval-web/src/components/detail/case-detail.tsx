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
