package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
)

func (p *PostgresStore) ListHarnessProfiles() []harness.Profile {
	rows, err := p.db.Query(`
		select id, role, name, description, model, reasoning_effort, prompt_fragments, few_shot_snippets, tool_preference_order, retrieval_bias, reasoning_verbosity, memory_read_enabled, memory_write_enabled, repo_ref, created_at, updated_at
		from harness_profile
		order by role asc, id asc
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []harness.Profile{}
	for rows.Next() {
		item, err := scanHarnessProfile(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) GetHarnessProfile(profileID string) (harness.Profile, bool) {
	row := p.db.QueryRow(`
		select id, role, name, description, model, reasoning_effort, prompt_fragments, few_shot_snippets, tool_preference_order, retrieval_bias, reasoning_verbosity, memory_read_enabled, memory_write_enabled, repo_ref, created_at, updated_at
		from harness_profile
		where id = $1
	`, strings.TrimSpace(profileID))
	item, err := scanHarnessProfileRow(row)
	if err != nil {
		return harness.Profile{}, false
	}
	return item, true
}

func (p *PostgresStore) ListHarnessOverlays() []harness.Overlay {
	rows, err := p.db.Query(`
		select id, profile_id, role, version, status, target_kind, target_ref, proposal_id, prompt_fragments, few_shot_snippets, tool_preference_order, retrieval_bias, reasoning_verbosity, memory_read_enabled, memory_write_enabled, created_by, approved_by, created_at, updated_at, activated_at
		from harness_overlay
		order by updated_at desc, id asc
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []harness.Overlay{}
	for rows.Next() {
		item, err := scanHarnessOverlay(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) GetActiveHarnessOverlay(role string) (harness.Overlay, bool) {
	row := p.db.QueryRow(`
		select id, profile_id, role, version, status, target_kind, target_ref, proposal_id, prompt_fragments, few_shot_snippets, tool_preference_order, retrieval_bias, reasoning_verbosity, memory_read_enabled, memory_write_enabled, created_by, approved_by, created_at, updated_at, activated_at
		from harness_overlay
		where role = $1 and status = $2
		order by updated_at desc, id asc
		limit 1
	`, strings.TrimSpace(role), string(harness.OverlayStatusActive))
	item, err := scanHarnessOverlayRow(row)
	if err != nil {
		return harness.Overlay{}, false
	}
	return item, true
}

func (p *PostgresStore) UpsertHarnessOverlay(item harness.Overlay) (harness.Overlay, error) {
	item = normalizeHarnessOverlay(item)
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextUUID("overlay")
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = now
	}
	if item.Status == harness.OverlayStatusActive && item.ActivatedAt == nil {
		item.ActivatedAt = &now
	}
	tx, err := p.db.Begin()
	if err != nil {
		return harness.Overlay{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if item.Status == harness.OverlayStatusActive {
		if _, err := tx.Exec(`
			update harness_overlay
			set status = $1, updated_at = $2
			where role = $3 and status = $4 and id <> $5
		`, string(harness.OverlayStatusSuperseded), now, item.Role, string(harness.OverlayStatusActive), item.ID); err != nil {
			return harness.Overlay{}, err
		}
	}
	if _, err := tx.Exec(`
		insert into harness_overlay (
			id, profile_id, role, version, status, target_kind, target_ref, proposal_id, prompt_fragments, few_shot_snippets, tool_preference_order, retrieval_bias, reasoning_verbosity, memory_read_enabled, memory_write_enabled, created_by, approved_by, created_at, updated_at, activated_at
		) values (
			$1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10::jsonb,$11::jsonb,$12,$13,$14,$15,$16,$17,$18,$19,$20
		)
		on conflict (id) do update set
			profile_id = excluded.profile_id,
			role = excluded.role,
			version = excluded.version,
			status = excluded.status,
			target_kind = excluded.target_kind,
			target_ref = excluded.target_ref,
			proposal_id = excluded.proposal_id,
			prompt_fragments = excluded.prompt_fragments,
			few_shot_snippets = excluded.few_shot_snippets,
			tool_preference_order = excluded.tool_preference_order,
			retrieval_bias = excluded.retrieval_bias,
			reasoning_verbosity = excluded.reasoning_verbosity,
			memory_read_enabled = excluded.memory_read_enabled,
			memory_write_enabled = excluded.memory_write_enabled,
			created_by = excluded.created_by,
			approved_by = excluded.approved_by,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			activated_at = excluded.activated_at
	`,
		item.ID,
		item.ProfileID,
		item.Role,
		item.Version,
		string(item.Status),
		item.TargetKind,
		item.TargetRef,
		nullString(item.ProposalID),
		jsonString(item.PromptFragments),
		jsonString(item.FewShotSnippets),
		jsonString(item.ToolPreferenceOrder),
		item.RetrievalBias,
		item.ReasoningVerbosity,
		item.MemoryReadEnabled,
		item.MemoryWriteEnabled,
		item.CreatedBy,
		item.ApprovedBy,
		item.CreatedAt,
		item.UpdatedAt,
		nullTime(item.ActivatedAt),
	); err != nil {
		return harness.Overlay{}, err
	}
	if err := tx.Commit(); err != nil {
		return harness.Overlay{}, err
	}
	return item, nil
}

func (p *PostgresStore) ListHarnessExperiments() []harness.Experiment {
	rows, err := p.db.Query(`
		select id, profile_id, overlay_id, proposal_id, attempt_id, role, status, summary, metrics, created_at, updated_at
		from harness_experiment
		order by updated_at desc, id asc
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []harness.Experiment{}
	for rows.Next() {
		item, err := scanHarnessExperiment(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) RecordHarnessExperiment(item harness.Experiment) (harness.Experiment, error) {
	item = normalizeHarnessExperiment(item)
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextUUID("hexp")
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = now
	}
	if _, err := p.db.Exec(`
		insert into harness_experiment (id, profile_id, overlay_id, proposal_id, attempt_id, role, status, summary, metrics, created_at, updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11)
		on conflict (id) do update set
			profile_id = excluded.profile_id,
			overlay_id = excluded.overlay_id,
			proposal_id = excluded.proposal_id,
			attempt_id = excluded.attempt_id,
			role = excluded.role,
			status = excluded.status,
			summary = excluded.summary,
			metrics = excluded.metrics,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at
	`,
		item.ID,
		item.ProfileID,
		nullString(item.OverlayID),
		nullString(item.ProposalID),
		firstNonEmpty(item.AttemptID),
		item.Role,
		string(item.Status),
		item.Summary,
		jsonString(item.Metrics),
		item.CreatedAt,
		item.UpdatedAt,
	); err != nil {
		return harness.Experiment{}, err
	}
	return item, nil
}

func (p *PostgresStore) ListHarnessSessionBindings() []harness.SessionBinding {
	rows, err := p.db.Query(`
		select role, scope_kind, scope_id, parent_scope_kind, parent_scope_id, hermes_session_id, parent_session_id, memory_backend, assistant_peer_id, user_peer_id, harness_profile_id, effective_overlay_id, effective_overlay_version, last_used_at, created_at, updated_at
		from harness_session_binding
		order by last_used_at desc, role asc, scope_kind asc, scope_id asc
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []harness.SessionBinding{}
	for rows.Next() {
		item, err := scanHarnessSessionBinding(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) UpsertHarnessSessionBinding(item harness.SessionBinding) (harness.SessionBinding, error) {
	item = normalizeHarnessSessionBinding(item)
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.LastUsedAt.IsZero() {
		item.LastUsedAt = now
	}
	if item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		item.UpdatedAt = now
	}
	if _, err := p.db.Exec(`
		insert into harness_session_binding (
			role, scope_kind, scope_id, parent_scope_kind, parent_scope_id, hermes_session_id, parent_session_id, memory_backend, assistant_peer_id, user_peer_id, harness_profile_id, effective_overlay_id, effective_overlay_version, last_used_at, created_at, updated_at
		) values (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
		)
		on conflict (role, scope_kind, scope_id) do update set
			parent_scope_kind = excluded.parent_scope_kind,
			parent_scope_id = excluded.parent_scope_id,
			hermes_session_id = excluded.hermes_session_id,
			parent_session_id = excluded.parent_session_id,
			memory_backend = excluded.memory_backend,
			assistant_peer_id = excluded.assistant_peer_id,
			user_peer_id = excluded.user_peer_id,
			harness_profile_id = excluded.harness_profile_id,
			effective_overlay_id = excluded.effective_overlay_id,
			effective_overlay_version = excluded.effective_overlay_version,
			last_used_at = excluded.last_used_at,
			updated_at = excluded.updated_at
	`,
		item.Role,
		item.ScopeKind,
		item.ScopeID,
		item.ParentScopeKind,
		item.ParentScopeID,
		item.HermesSessionID,
		item.ParentSessionID,
		item.MemoryBackend,
		item.AssistantPeerID,
		item.UserPeerID,
		item.HarnessProfileID,
		item.EffectiveOverlayID,
		item.EffectiveOverlayVersion,
		item.LastUsedAt,
		item.CreatedAt,
		item.UpdatedAt,
	); err != nil {
		return harness.SessionBinding{}, err
	}
	return item, nil
}

func (p *PostgresStore) ListHarnessExecutions() []harness.Execution {
	rows, err := p.db.Query(`
		select id, trace_id, proposal_id, role, session_scope_kind, session_scope_id, hermes_session_id, parent_session_id, harness_profile_id, effective_overlay_id, effective_overlay_version, memory_backend, memory_reads, memory_writes, created_at
		from harness_execution
		order by created_at desc, id asc
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []harness.Execution{}
	for rows.Next() {
		item, err := scanHarnessExecution(rows)
		if err != nil {
			return nil
		}
		out = append(out, item)
	}
	return out
}

func (p *PostgresStore) RecordHarnessExecution(item harness.Execution) (harness.Execution, error) {
	item = normalizeHarnessExecution(item)
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextUUID("hexec")
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if _, err := p.db.Exec(`
		insert into harness_execution (
			id, trace_id, proposal_id, role, session_scope_kind, session_scope_id, hermes_session_id, parent_session_id, harness_profile_id, effective_overlay_id, effective_overlay_version, memory_backend, memory_reads, memory_writes, created_at
		) values (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13::jsonb,$14::jsonb,$15
		)
		on conflict (id) do update set
			trace_id = excluded.trace_id,
			proposal_id = excluded.proposal_id,
			role = excluded.role,
			session_scope_kind = excluded.session_scope_kind,
			session_scope_id = excluded.session_scope_id,
			hermes_session_id = excluded.hermes_session_id,
			parent_session_id = excluded.parent_session_id,
			harness_profile_id = excluded.harness_profile_id,
			effective_overlay_id = excluded.effective_overlay_id,
			effective_overlay_version = excluded.effective_overlay_version,
			memory_backend = excluded.memory_backend,
			memory_reads = excluded.memory_reads,
			memory_writes = excluded.memory_writes,
			created_at = excluded.created_at
	`,
		item.ID,
		nullString(item.TraceID),
		nullString(item.ProposalID),
		item.Role,
		item.SessionScopeKind,
		item.SessionScopeID,
		item.HermesSessionID,
		item.ParentSessionID,
		item.HarnessProfileID,
		item.EffectiveOverlayID,
		item.EffectiveOverlayVersion,
		item.MemoryBackend,
		jsonString(item.MemoryReads),
		jsonString(item.MemoryWrites),
		item.CreatedAt,
	); err != nil {
		return harness.Execution{}, err
	}
	return item, nil
}

type harnessScanner interface {
	Scan(dest ...any) error
}

func scanHarnessProfile(scanner harnessScanner) (harness.Profile, error) {
	var (
		item                harness.Profile
		promptFragments     []byte
		fewShotSnippets     []byte
		toolPreferenceOrder []byte
	)
	err := scanner.Scan(
		&item.ID,
		&item.Role,
		&item.Name,
		&item.Description,
		&item.Model,
		&item.ReasoningEffort,
		&promptFragments,
		&fewShotSnippets,
		&toolPreferenceOrder,
		&item.RetrievalBias,
		&item.ReasoningVerbosity,
		&item.MemoryReadEnabled,
		&item.MemoryWriteEnabled,
		&item.RepoRef,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return harness.Profile{}, err
	}
	item.PromptFragments = decodeJSON(promptFragments, []string{})
	item.FewShotSnippets = decodeJSON(fewShotSnippets, []string{})
	item.ToolPreferenceOrder = decodeJSON(toolPreferenceOrder, []string{})
	return normalizeHarnessProfile(item), nil
}

func scanHarnessProfileRow(row *sql.Row) (harness.Profile, error) {
	return scanHarnessProfile(row)
}

func scanHarnessOverlay(scanner harnessScanner) (harness.Overlay, error) {
	var (
		item                harness.Overlay
		proposalID          sql.NullString
		promptFragments     []byte
		fewShotSnippets     []byte
		toolPreferenceOrder []byte
		memoryReadEnabled   sql.NullBool
		memoryWriteEnabled  sql.NullBool
		activatedAt         sql.NullTime
	)
	err := scanner.Scan(
		&item.ID,
		&item.ProfileID,
		&item.Role,
		&item.Version,
		&item.Status,
		&item.TargetKind,
		&item.TargetRef,
		&proposalID,
		&promptFragments,
		&fewShotSnippets,
		&toolPreferenceOrder,
		&item.RetrievalBias,
		&item.ReasoningVerbosity,
		&memoryReadEnabled,
		&memoryWriteEnabled,
		&item.CreatedBy,
		&item.ApprovedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
		&activatedAt,
	)
	if err != nil {
		return harness.Overlay{}, err
	}
	item.ProposalID = proposalID.String
	item.PromptFragments = decodeJSON(promptFragments, []string{})
	item.FewShotSnippets = decodeJSON(fewShotSnippets, []string{})
	item.ToolPreferenceOrder = decodeJSON(toolPreferenceOrder, []string{})
	if memoryReadEnabled.Valid {
		item.MemoryReadEnabled = &memoryReadEnabled.Bool
	}
	if memoryWriteEnabled.Valid {
		item.MemoryWriteEnabled = &memoryWriteEnabled.Bool
	}
	if activatedAt.Valid {
		item.ActivatedAt = &activatedAt.Time
	}
	return normalizeHarnessOverlay(item), nil
}

func scanHarnessOverlayRow(row *sql.Row) (harness.Overlay, error) {
	return scanHarnessOverlay(row)
}

func scanHarnessExperiment(scanner harnessScanner) (harness.Experiment, error) {
	var (
		item       harness.Experiment
		overlayID  sql.NullString
		proposalID sql.NullString
		attemptID  sql.NullString
		metrics    []byte
	)
	err := scanner.Scan(
		&item.ID,
		&item.ProfileID,
		&overlayID,
		&proposalID,
		&attemptID,
		&item.Role,
		&item.Status,
		&item.Summary,
		&metrics,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return harness.Experiment{}, err
	}
	item.OverlayID = overlayID.String
	item.ProposalID = proposalID.String
	item.AttemptID = attemptID.String
	item.Metrics = decodeJSON(metrics, map[string]any{})
	return normalizeHarnessExperiment(item), nil
}

func scanHarnessSessionBinding(scanner harnessScanner) (harness.SessionBinding, error) {
	var item harness.SessionBinding
	err := scanner.Scan(
		&item.Role,
		&item.ScopeKind,
		&item.ScopeID,
		&item.ParentScopeKind,
		&item.ParentScopeID,
		&item.HermesSessionID,
		&item.ParentSessionID,
		&item.MemoryBackend,
		&item.AssistantPeerID,
		&item.UserPeerID,
		&item.HarnessProfileID,
		&item.EffectiveOverlayID,
		&item.EffectiveOverlayVersion,
		&item.LastUsedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return harness.SessionBinding{}, err
	}
	return normalizeHarnessSessionBinding(item), nil
}

func scanHarnessExecution(scanner harnessScanner) (harness.Execution, error) {
	var (
		item         harness.Execution
		traceID      sql.NullString
		proposalID   sql.NullString
		memoryReads  []byte
		memoryWrites []byte
	)
	err := scanner.Scan(
		&item.ID,
		&traceID,
		&proposalID,
		&item.Role,
		&item.SessionScopeKind,
		&item.SessionScopeID,
		&item.HermesSessionID,
		&item.ParentSessionID,
		&item.HarnessProfileID,
		&item.EffectiveOverlayID,
		&item.EffectiveOverlayVersion,
		&item.MemoryBackend,
		&memoryReads,
		&memoryWrites,
		&item.CreatedAt,
	)
	if err != nil {
		return harness.Execution{}, err
	}
	item.TraceID = traceID.String
	item.ProposalID = proposalID.String
	item.MemoryReads = decodeJSON(memoryReads, []harness.MemoryArtifact{})
	item.MemoryWrites = decodeJSON(memoryWrites, []harness.MemoryArtifact{})
	return normalizeHarnessExecution(item), nil
}

func nextUUID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, strings.ReplaceAll(uuid.NewString(), "-", ""))
}
