import type { HarnessResponse, HarnessExecution, HarnessOverlay, RuntimeRole } from "@/types";
import { formatTime, listOrEmpty } from "@/hooks/api";
import { EmptyDetail } from "./empty-detail";

function latestExecution(executions: HarnessExecution[], role: string): HarnessExecution | undefined {
  return executions.find((item) => item.role === role);
}

function activeOverlay(overlays: HarnessOverlay[], role: string): HarnessOverlay | undefined {
  return overlays.find((item) => item.role === role && item.status === "active");
}

function runtimeForRole(roles: RuntimeRole[], role: string): RuntimeRole | undefined {
  return roles.find((item) => item.role === role);
}

export function HarnessDetail(props: {
  detail?: HarnessResponse;
  selectedRole?: string;
}) {
  if (!props.selectedRole) {
    return <EmptyDetail title="Select a role agent" body="Choose a runner role to inspect its Hermes session continuity, active overlay, and memory activity." />;
  }
  if (!props.detail) {
    return <EmptyDetail title="Loading harness state" body="Fetching runner role health, active overlays, recent sessions, and experiment history." />;
  }

  const role = props.selectedRole;
  const runtime = runtimeForRole(listOrEmpty(props.detail.roles), role);
  const overlays = listOrEmpty(props.detail.overlays).filter((item) => item.role === role);
  const experiments = listOrEmpty(props.detail.experiments).filter((item) => item.role === role);
  const bindings = listOrEmpty(props.detail.session_bindings).filter((item) => item.role === role);
  const executions = listOrEmpty(props.detail.executions).filter((item) => item.role === role);
  const profile =
    listOrEmpty(props.detail.profiles).find((item) => item.role === role && item.id === runtime?.harness_profile_id) ??
    listOrEmpty(props.detail.profiles).find((item) => item.role === role);
  const latest = latestExecution(executions, role);
  const overlay = activeOverlay(overlays, role);

  return (
    <div className="detail-stack">
      <div className="detail-card">
        <div className="detail-header">
          <div>
            <p className="eyebrow">Harness role</p>
            <h2>{role}</h2>
          </div>
          <div className="detail-meta">
            <span className="status-chip">{runtime?.healthy ? "healthy" : "degraded"}</span>
            <span className="status-chip">{runtime?.model || "unreachable"}</span>
          </div>
        </div>
        <dl className="overview-grid">
          <div><dt>Reasoning</dt><dd>{runtime?.reasoning_effort || "n/a"}</dd></div>
          <div><dt>Persistence</dt><dd>{runtime?.persistence_enabled ? "enabled" : "disabled"}</dd></div>
          <div><dt>Honcho</dt><dd>{runtime?.honcho_available ? "available" : "unavailable"}</dd></div>
          <div><dt>Overlay</dt><dd>{runtime?.active_overlay_version || overlay?.version || "baseline"}</dd></div>
          <div><dt>Bindings</dt><dd>{bindings.length}</dd></div>
          <div><dt>Executions</dt><dd>{executions.length}</dd></div>
        </dl>
        {runtime?.error ? <p className="muted">Runtime error: {runtime.error}</p> : null}
      </div>

      <div className="review-grid">
        <div className="detail-card">
          <h3>Effective profile</h3>
          {profile ? (
            <div className="nested-list">
              <div className="nested-card">
                <div className="detail-row-header">
                  <strong>{profile.name}</strong>
                  <small>{profile.id}</small>
                </div>
                <p className="detail-copy">{profile.description || "No description."}</p>
                <p className="muted">Retrieval bias: {profile.retrieval_bias || "default"} · Memory read/write: {profile.memory_read_enabled ? "on" : "off"}/{profile.memory_write_enabled ? "on" : "off"}</p>
                {listOrEmpty(profile.prompt_fragments).length ? <p className="muted">Prompt fragments: {listOrEmpty(profile.prompt_fragments).join(" • ")}</p> : null}
                {listOrEmpty(profile.tool_preference_order).length ? <p className="muted">Tool order: {listOrEmpty(profile.tool_preference_order).join(" -> ")}</p> : null}
              </div>
              {overlay ? (
                <div className="nested-card">
                  <div className="detail-row-header">
                    <strong>Active overlay</strong>
                    <small>{overlay.version}</small>
                  </div>
                  <p className="detail-copy">{listOrEmpty(overlay.prompt_fragments).join(" ") || "No prompt fragments recorded."}</p>
                  <p className="muted">Created by {overlay.created_by || "unknown"} · {formatTime(overlay.updated_at)}</p>
                </div>
              ) : null}
            </div>
          ) : (
            <p className="detail-copy">No harness profile recorded for this role.</p>
          )}
        </div>

        <div className="detail-card">
          <h3>Latest execution</h3>
          {latest ? (
            <div className="nested-list">
              <div className="nested-card">
                <div className="detail-row-header">
                  <strong>{latest.hermes_session_id}</strong>
                  <small>{formatTime(latest.created_at)}</small>
                </div>
                <p className="muted">Scope: {latest.session_scope_kind}:{latest.session_scope_id}</p>
                {latest.parent_session_id ? <p className="muted">Parent session: {latest.parent_session_id}</p> : null}
                <p className="muted">Memory backend: {latest.memory_backend || "n/a"}</p>
              </div>
              {listOrEmpty(latest.memory_reads).length ? (
                <div className="nested-card">
                  <div className="detail-row-header">
                    <strong>Memory reads</strong>
                    <small>{listOrEmpty(latest.memory_reads).length}</small>
                  </div>
                  {listOrEmpty(latest.memory_reads).map((item, index) => (
                    <p key={`${item.kind}-${index}`} className="detail-copy">{item.kind}: {item.summary}</p>
                  ))}
                </div>
              ) : null}
              {listOrEmpty(latest.memory_writes).length ? (
                <div className="nested-card">
                  <div className="detail-row-header">
                    <strong>Memory writes</strong>
                    <small>{listOrEmpty(latest.memory_writes).length}</small>
                  </div>
                  {listOrEmpty(latest.memory_writes).map((item, index) => (
                    <p key={`${item.kind}-${index}`} className="detail-copy">{item.kind}: {item.summary}</p>
                  ))}
                </div>
              ) : null}
            </div>
          ) : (
            <p className="detail-copy">No persisted executions recorded for this role yet.</p>
          )}
        </div>
      </div>

      <div className="review-grid">
        <div className="detail-card">
          <h3>Session bindings</h3>
          <div className="nested-list">
            {bindings.map((item) => (
              <div key={`${item.scope_kind}:${item.scope_id}`} className="nested-card">
                <div className="detail-row-header">
                  <strong>{item.scope_kind}:{item.scope_id}</strong>
                  <small>{formatTime(item.last_used_at)}</small>
                </div>
                <p className="muted">Session: {item.hermes_session_id}</p>
                {item.parent_session_id ? <p className="muted">Parent: {item.parent_session_id}</p> : null}
              </div>
            ))}
            {!bindings.length ? <div className="nested-card"><p className="detail-copy">No active session bindings recorded.</p></div> : null}
          </div>
        </div>

        <div className="detail-card">
          <h3>Overlay experiments</h3>
          <div className="nested-list">
            {experiments.map((item) => (
              <div key={item.id} className="nested-card">
                <div className="detail-row-header">
                  <strong>{item.status}</strong>
                  <small>{formatTime(item.updated_at)}</small>
                </div>
                <p className="detail-copy">{item.summary || "No summary recorded."}</p>
              </div>
            ))}
            {!experiments.length ? <div className="nested-card"><p className="detail-copy">No harness experiments recorded yet.</p></div> : null}
          </div>
        </div>
      </div>
    </div>
  );
}
