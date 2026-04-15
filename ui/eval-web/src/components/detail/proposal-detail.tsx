import type { ProposalDetailResponse } from "@/types";
import { formatTime, latestActionResult, listOrEmpty, scoreBadge } from "@/hooks/api";

export function ProposalDetail(props: {
  detail: ProposalDetailResponse;
  proposalRationale: string;
  setProposalRationale: (value: string) => void;
  onDecision: (decision: string) => void;
  onRetry: () => void;
  onStop: () => void;
  canRetry: boolean;
  canStop: boolean;
}) {
  const prAttempt = listOrEmpty(props.detail.pr_attempts)[0];
  const actionIntents = listOrEmpty(props.detail.action_intents);
  const actionResults = listOrEmpty(props.detail.action_results);
  const attempts = listOrEmpty(props.detail.attempts);
  const workspaces = listOrEmpty(props.detail.attempt_workspaces);
  const operations = listOrEmpty(props.detail.operations);
  const lineOperations = operations.filter((item) => item.scope_kind === "proposal");
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
          <div><dt>Intervention</dt><dd>{props.detail.proposal.recommended_intervention_kind || "repo_change"}</dd></div>
          <div><dt>Disposition</dt><dd>{props.detail.proposal.recommended_disposition || "approve_intervention"}</dd></div>
          <div><dt>Origin trace</dt><dd>{props.detail.proposal.origin_trace_id || props.detail.proposal.trace_id}</dd></div>
          <div><dt>Evidence traces</dt><dd>{listOrEmpty(props.detail.proposal.evidence_trace_ids).length}</dd></div>
          <div><dt>Target layer</dt><dd>{props.detail.proposal.target_layer || "repo_change"}</dd></div>
          <div><dt>Target</dt><dd>{props.detail.proposal.target_kind || "n/a"} {props.detail.proposal.target_ref ? `· ${props.detail.proposal.target_ref}` : ""}</dd></div>
          <div><dt>Target surface</dt><dd>{props.detail.proposal.target_surface || props.detail.proposal.proposed_scope || "n/a"}</dd></div>
          <div><dt>Current attempt</dt><dd>{props.detail.proposal.current_attempt_id || "n/a"}</dd></div>
          <div><dt>Attempt count</dt><dd>{props.detail.proposal.attempt_count ?? 0}</dd></div>
          <div><dt>Retry budget</dt><dd>{props.detail.proposal.auto_retry_budget_remaining ?? 0}</dd></div>
          <div><dt>Next retry action</dt><dd>{props.detail.proposal.next_retry_action || "n/a"}</dd></div>
          <div><dt>Last failure</dt><dd>{props.detail.proposal.last_failure_class || "n/a"}</dd></div>
        </dl>
        <div className="nested-list">
          <div className="nested-card">
            <div className="detail-row-header">
              <strong>Recommended intervention</strong>
              <small>{props.detail.proposal.recommended_intervention_kind || "repo_change"}</small>
            </div>
            <p className="detail-copy">{props.detail.proposal.recommended_intervention_rationale || props.detail.proposal.summary}</p>
            <p className="muted">{props.detail.proposal.material_risk_summary || "No material risk summary recorded."}</p>
            {props.detail.proposal.validation_plan ? <p className="muted">Validation: {props.detail.proposal.validation_plan}</p> : null}
            {listOrEmpty(props.detail.proposal.touched_files).length ? (
              <p className="muted">Expected files: {listOrEmpty(props.detail.proposal.touched_files).join(", ")}</p>
            ) : null}
          </div>
        </div>
        {props.detail.proposal.line_stop_reason ? (
          <p className="muted">Line stopped: {props.detail.proposal.line_stop_reason}</p>
        ) : null}
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
            <button onClick={() => props.onDecision("approved")}>Approve intervention</button>
            <button className="secondary" onClick={() => props.onDecision("dismissed")}>Dismiss line</button>
            <button className="secondary" onClick={() => props.onDecision("rejected")}>Reject line</button>
            <button className="secondary" onClick={() => props.onDecision("merged")}>Mark merged</button>
            {props.canRetry ? <button className="secondary" onClick={props.onRetry}>Resume attempt</button> : null}
            {props.canStop ? <button className="secondary" onClick={props.onStop}>Stop line</button> : null}
          </div>
        </div>
      </div>

      <div className="detail-card">
        <h3>Attempts</h3>
        <div className="nested-list">
          {attempts.map((attempt) => {
            const attemptJobs = listOrEmpty(props.detail.repo_change_jobs).filter((job) => job.attempt_id === attempt.id);
            const attemptPRs = listOrEmpty(props.detail.pr_attempts).filter((item) => item.attempt_id === attempt.id);
            const attemptWorkspace = workspaces.find((item) => item.attempt_id === attempt.id);
            const attemptOperations = operations.filter((item) => item.attempt_id === attempt.id || (item.scope_kind === "attempt" && item.scope_id === attempt.id));
            return (
              <div key={attempt.id} className="nested-card">
                <div className="detail-row-header">
                  <strong>Attempt {attempt.attempt_number}</strong>
                  <small>{attempt.state}</small>
                </div>
                <p className="detail-copy">{attempt.change_plan || attempt.failure_summary || attempt.validation_summary || "No attempt summary recorded."}</p>
                <p className="muted">Trigger: {attempt.trigger} · Branch: {attempt.branch_name || "n/a"}</p>
                <p className="muted">Failure: {attempt.failure_class || "n/a"} · Retry: {attempt.retry_decision || "n/a"}</p>
                {attemptWorkspace ? (
                  <p className="muted">
                    Workspace: {attemptWorkspace.status}
                    {attemptWorkspace.repo ? ` · ${attemptWorkspace.repo}` : ""}
                    {attemptWorkspace.branch_name ? ` · ${attemptWorkspace.branch_name}` : ""}
                    {attemptWorkspace.pod_name ? ` · pod ${attemptWorkspace.pod_name}` : ""}
                  </p>
                ) : null}
                {attemptWorkspace?.diff_summary ? <p className="muted">Diff: {attemptWorkspace.diff_summary}</p> : null}
                {listOrEmpty(attemptWorkspace?.allowed_path_globs).length ? (
                  <p className="muted">Allowed paths: {listOrEmpty(attemptWorkspace?.allowed_path_globs).join(", ")}</p>
                ) : null}
                {listOrEmpty(attempt.changed_files).length ? (
                  <p className="muted">Files: {listOrEmpty(attempt.changed_files).join(", ")}</p>
                ) : null}
                {attempt.validation_plan ? <p className="muted">Validation: {attempt.validation_plan}</p> : null}
                {attempt.hypothesis_delta ? <p className="muted">Delta: {attempt.hypothesis_delta}</p> : null}
                {attemptOperations.length ? (
                  <div className="nested-list">
                    {attemptOperations.map((item) => (
                      <div key={item.id} className="nested-card">
                        <div className="detail-row-header">
                          <strong>{item.operation_kind}</strong>
                          <small>{item.status}</small>
                        </div>
                        <p className="detail-copy">{item.operation_key}{item.queue ? ` · ${item.queue}` : ""}</p>
                        <p className="muted">
                          {item.started_at ? `Started ${formatTime(item.started_at)}` : "Not started"}
                          {item.completed_at ? ` · Completed ${formatTime(item.completed_at)}` : ""}
                          {typeof item.retry_count === "number" ? ` · Retries ${item.retry_count}` : ""}
                        </p>
                        {item.last_error ? <p className="muted">Error: {item.last_error}</p> : null}
                        {item.result_ref ? <p className="muted">Result: {item.result_ref}</p> : null}
                      </div>
                    ))}
                  </div>
                ) : null}
                {attemptJobs.map((job) => (
                  <p key={job.id} className="muted">Sandbox: {job.status}{job.sandbox_job_name ? ` · ${job.sandbox_job_name}` : ""}{job.validation_error ? ` · ${job.validation_error}` : ""}</p>
                ))}
                {attemptPRs.map((item) => (
                  <p key={item.id} className="muted">
                    PR: {item.status}
                    {item.pr_url ? <> · <a className="detail-link" href={item.pr_url} target="_blank" rel="noreferrer">{item.pr_url}</a></> : null}
                  </p>
                ))}
              </div>
            );
          })}
          {!attempts.length ? (
            <div className="nested-card"><p className="detail-copy">No change attempts recorded yet.</p></div>
          ) : null}
        </div>
      </div>

      <div className="detail-card">
        <h3>Operation ledger</h3>
        <div className="nested-list">
          {lineOperations.map((item) => (
            <div key={item.id} className="nested-card">
              <div className="detail-row-header">
                <strong>{item.operation_kind}</strong>
                <small>{item.status}</small>
              </div>
              <p className="detail-copy">{item.operation_key}{item.queue ? ` · ${item.queue}` : ""}</p>
              <p className="muted">
                Scope: {item.scope_kind}:{item.scope_id}
                {item.attempt_id ? ` · Attempt ${item.attempt_id}` : ""}
                {typeof item.retry_count === "number" ? ` · Retries ${item.retry_count}` : ""}
              </p>
              {item.last_error ? <p className="muted">Error: {item.last_error}</p> : null}
              {item.result_ref ? <p className="muted">Result: {item.result_ref}</p> : null}
            </div>
          ))}
          {!lineOperations.length ? (
            <div className="nested-card"><p className="detail-copy">No proposal-line operations recorded yet.</p></div>
          ) : null}
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
