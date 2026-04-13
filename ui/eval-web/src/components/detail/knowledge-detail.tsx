import type { KnowledgeDetailResponse } from "@/types";
import { formatTime, listOrEmpty, scoreBadge } from "@/hooks/api";

export function KnowledgeDetail(props: {
  detail: KnowledgeDetailResponse;
  reviewRationale: string;
  setReviewRationale: (value: string) => void;
  onDecision: (decision: string) => void;
}) {
  return (
    <div className="detail-stack">
      <div className="detail-card">
        <div className="detail-header">
          <div>
            <p className="eyebrow">Knowledge</p>
            <h2>{props.detail.knowledge_entry.title}</h2>
          </div>
          <div className="detail-meta">
            <span className="status-chip">{props.detail.knowledge_entry.status}</span>
            <span className="status-chip">{props.detail.knowledge_entry.tier}</span>
          </div>
        </div>
        <p className="detail-copy">{props.detail.knowledge_entry.summary || props.detail.knowledge_entry.body || "No summary."}</p>
        <dl className="overview-grid">
          <div><dt>Kind</dt><dd>{props.detail.knowledge_entry.kind}</dd></div>
          <div><dt>Scope</dt><dd>{props.detail.knowledge_entry.scope_type}</dd></div>
          <div><dt>Scope ID</dt><dd>{props.detail.knowledge_entry.scope_id || "global"}</dd></div>
          <div><dt>Confidence</dt><dd>{scoreBadge(props.detail.knowledge_entry.confidence)}</dd></div>
        </dl>
      </div>

      <div className="review-grid">
        <div className="detail-card">
          <h3>Evidence links</h3>
          <div className="nested-list">
            {listOrEmpty(props.detail.evidence_links).map((link) => (
              <div key={`${link.evidence_type}:${link.evidence_id}`} className="nested-card">
                <div className="detail-row-header">
                  <strong>{link.evidence_type}</strong>
                  <small>{link.evidence_id}</small>
                </div>
                <p className="detail-copy">{link.relevance_summary || link.evidence_ref.summary || link.evidence_ref.ref}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="detail-card">
          <h3>Review actions</h3>
          <label className="field">
            Review rationale
            <textarea value={props.reviewRationale} onChange={(event) => props.setReviewRationale(event.target.value)} placeholder="Why this entry should be approved, rejected, or marked stale." />
          </label>
          <div className="button-row">
            <button onClick={() => props.onDecision("approve")}>Approve</button>
            <button className="secondary" onClick={() => props.onDecision("reject")}>Reject</button>
            <button className="secondary" onClick={() => props.onDecision("mark_stale")}>Mark stale</button>
            <button className="secondary" onClick={() => props.onDecision("archive")}>Archive</button>
          </div>
        </div>
      </div>

      <div className="detail-card">
        <h3>Review history</h3>
        <div className="nested-list">
          {listOrEmpty(props.detail.reviews).map((review) => (
            <div key={review.id} className="nested-card">
              <div className="detail-row-header">
                <strong>{review.decision}</strong>
                <small>{formatTime(review.created_at)}</small>
              </div>
              <p className="detail-copy">{review.rationale || "No rationale."}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
