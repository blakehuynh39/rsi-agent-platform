import type { ProposalDetailResponse } from "@/types";
import { formatTime, latestActionResult, listOrEmpty, scoreBadge } from "@/hooks/api";

export function ProposalDetail(props: {
  detail: ProposalDetailResponse;
  proposalRationale: string;
  setProposalRationale: (value: string) => void;
  onDecision: (decision: string) => void;
  onRetry: () => void;
  canRetry: boolean;
}) {
  const prAttempt = listOrEmpty(props.detail.pr_attempts)[0];
  const actionIntents = listOrEmpty(props.detail.action_intents);
  const actionResults = listOrEmpty(props.detail.action_results);
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
          <div><dt>Target layer</dt><dd>{props.detail.proposal.target_layer || "repo_change"}</dd></div>
          <div><dt>Target</dt><dd>{props.detail.proposal.target_kind || "n/a"} {props.detail.proposal.target_ref ? `· ${props.detail.proposal.target_ref}` : ""}</dd></div>
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
            {!listOrEmpty(props.detail.related_proposal_memory).length ? (
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
            {props.canRetry ? <button className="secondary" onClick={props.onRetry}>Retry repo change</button> : null}
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
              {job.validation_error ? <p className="muted">Validation: {job.validation_error}</p> : null}
              {job.validation_ref ? <p className="muted">Sandbox ref: {job.validation_ref}</p> : null}
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

      <div className="review-grid">
        <div className="detail-card">
          <h3>Action chain</h3>
          <div className="nested-list">
            {actionIntents.map((intent) => {
              const result = latestActionResult(intent.id, actionResults);
              return (
                <div key={intent.id} className="nested-card">
                  <div className="detail-row-header">
                    <strong>{intent.kind}</strong>
                    <small>{result?.status || intent.status}</small>
                  </div>
                  <p className="detail-copy">{intent.rationale || intent.target_ref || "No rationale recorded."}</p>
                  <p className="muted">{intent.target_ref || intent.policy_verdict || "No target."}</p>
                  {result?.error_message ? <p className="muted">Error: {result.error_message}</p> : null}
                </div>
              );
            })}
          </div>
        </div>

        <div className="detail-card">
          <h3>Hermes harness executions</h3>
          <div className="nested-list">
            {listOrEmpty(props.detail.harness_executions).map((item) => (
              <div key={item.id} className="nested-card">
                <div className="detail-row-header">
                  <strong>{item.role}</strong>
                  <small>{formatTime(item.created_at)}</small>
                </div>
                <p className="detail-copy">{item.hermes_session_id}</p>
                <p className="muted">Scope: {item.session_scope_kind}:{item.session_scope_id}</p>
                {item.effective_overlay_version ? <p className="muted">Overlay: {item.effective_overlay_version}</p> : null}
              </div>
            ))}
            {!listOrEmpty(props.detail.harness_executions).length ? (
              <div className="nested-card"><p className="detail-copy">No harness execution metadata recorded for this proposal yet.</p></div>
            ) : null}
          </div>
        </div>

        <div className="detail-card">
          <h3>Knowledge and outcomes</h3>
          <div className="nested-list">
            {listOrEmpty(props.detail.outcomes).map((item) => (
              <div key={item.id} className="nested-card">
                <div className="detail-row-header">
                  <strong>{item.outcome_type}</strong>
                  <small>{item.verdict} · {scoreBadge(item.score)}</small>
                </div>
                <p className="detail-copy">{item.summary || item.details || "No summary."}</p>
              </div>
            ))}
            {listOrEmpty(props.detail.knowledge_entries).map((item) => (
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
      </div>
    </div>
  );
}
