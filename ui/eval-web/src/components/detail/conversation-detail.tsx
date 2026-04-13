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
