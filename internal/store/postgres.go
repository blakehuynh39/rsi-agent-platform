package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/registry"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
)

type PostgresStore struct {
	db *sql.DB
}

type sqlReader interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func OpenStore(cfg config.Config) (Store, error) {
	if cfg.StoreBackend == "postgres" {
		return NewPostgresStore(cfg)
	}
	return NewMemoryStore(), nil
}

func MustOpenStore(cfg config.Config) Store {
	store, err := OpenStore(cfg)
	if err != nil {
		panic(err)
	}
	return store
}

func NewPostgresStore(cfg config.Config) (*PostgresStore, error) {
	db, err := sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec(platformdb.SchemaSQL); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	store := &PostgresStore{db: db}
	if err := store.ensureSeed(); err != nil {
		return nil, err
	}
	return store, nil
}

func (p *PostgresStore) ensureSeed() error {
	var count int
	if err := p.db.QueryRow(`select count(*) from event_envelope`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := persistStore(tx, NewMemoryStore()); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *PostgresStore) readStore() (*MemoryStore, error) {
	return loadStore(p.db)
}

func (p *PostgresStore) mutate(fn func(*MemoryStore) error) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	store, err := loadStore(tx)
	if err != nil {
		return err
	}
	if err := fn(store); err != nil {
		return err
	}
	if err := persistStore(tx, store); err != nil {
		return err
	}
	return tx.Commit()
}

func loadStore(r sqlReader) (*MemoryStore, error) {
	store := newEmptyMemoryStore()

	if err := loadThreadPolicies(r, store); err != nil {
		return nil, err
	}
	if err := loadChannelPolicies(r, store); err != nil {
		return nil, err
	}
	if err := loadOwnership(r, store); err != nil {
		return nil, err
	}
	if err := loadCapabilities(r, store); err != nil {
		return nil, err
	}
	if err := loadTemplates(r, store); err != nil {
		return nil, err
	}
	if err := loadExperiments(r, store); err != nil {
		return nil, err
	}
	if err := loadEvents(r, store); err != nil {
		return nil, err
	}
	if err := loadIngestions(r, store); err != nil {
		return nil, err
	}
	if err := loadWorkflows(r, store); err != nil {
		return nil, err
	}
	if err := loadAssignments(r, store); err != nil {
		return nil, err
	}
	if err := loadTraces(r, store); err != nil {
		return nil, err
	}
	if err := loadRatings(r, store); err != nil {
		return nil, err
	}
	if err := loadNotes(r, store); err != nil {
		return nil, err
	}
	if err := loadEvalSuites(r, store); err != nil {
		return nil, err
	}
	if err := loadEvalRuns(r, store); err != nil {
		return nil, err
	}
	if err := loadEvalJudgments(r, store); err != nil {
		return nil, err
	}
	if err := loadSettings(r, store); err != nil {
		return nil, err
	}
	if err := loadWorkItems(r, store); err != nil {
		return nil, err
	}
	if err := loadCandidates(r, store); err != nil {
		return nil, err
	}
	if err := loadProposals(r, store); err != nil {
		return nil, err
	}
	if err := loadProposalReviews(r, store); err != nil {
		return nil, err
	}
	if err := loadProposalMemory(r, store); err != nil {
		return nil, err
	}
	if err := loadRepoChangeJobs(r, store); err != nil {
		return nil, err
	}
	if err := loadPRAttempts(r, store); err != nil {
		return nil, err
	}
	if err := loadPostMergeReplays(r, store); err != nil {
		return nil, err
	}
	if err := loadCronLeases(r, store); err != nil {
		return nil, err
	}

	if len(store.events) == 0 && len(store.workflows) == 0 {
		return NewMemoryStore(), nil
	}
	return store, nil
}

func persistStore(tx *sql.Tx, store *MemoryStore) error {
	for _, table := range []string{
		"proposal_review",
		"proposal_memory",
		"pr_attempt",
		"repo_change_job",
		"post_merge_replay",
		"cron_lease",
		"slack_action_record",
		"tool_call_record",
		"reasoning_step",
		"work_item",
		"improvement_settings",
		"proposal",
		"improvement_candidate",
		"eval_judgment",
		"eval_run",
		"eval_suite",
		"improvement_note",
		"human_rating",
		"artifact",
		"trace_event",
		"trace_summary",
		"assignment",
		"workflow",
		"ingestion",
		"event_envelope",
		"experiment_registry",
		"workflow_templates",
		"capability_registry",
		"ownership_registry",
		"channel_policy",
		"thread_policy",
	} {
		if _, err := tx.Exec(`delete from ` + table); err != nil {
			return err
		}
	}

	if err := persistThreadPolicies(tx, store); err != nil {
		return err
	}
	if err := persistChannelPolicies(tx, store); err != nil {
		return err
	}
	if err := persistOwnership(tx, store); err != nil {
		return err
	}
	if err := persistCapabilities(tx, store); err != nil {
		return err
	}
	if err := persistTemplates(tx, store); err != nil {
		return err
	}
	if err := persistExperiments(tx, store); err != nil {
		return err
	}
	if err := persistEvents(tx, store); err != nil {
		return err
	}
	if err := persistIngestions(tx, store); err != nil {
		return err
	}
	if err := persistWorkflows(tx, store); err != nil {
		return err
	}
	if err := persistAssignments(tx, store); err != nil {
		return err
	}
	if err := persistTraces(tx, store); err != nil {
		return err
	}
	if err := persistRatings(tx, store); err != nil {
		return err
	}
	if err := persistNotes(tx, store); err != nil {
		return err
	}
	if err := persistEvalSuites(tx, store); err != nil {
		return err
	}
	if err := persistEvalRuns(tx, store); err != nil {
		return err
	}
	if err := persistEvalJudgments(tx, store); err != nil {
		return err
	}
	if err := persistSettings(tx, store); err != nil {
		return err
	}
	if err := persistWorkItems(tx, store); err != nil {
		return err
	}
	if err := persistCandidates(tx, store); err != nil {
		return err
	}
	if err := persistProposals(tx, store); err != nil {
		return err
	}
	if err := persistProposalReviews(tx, store); err != nil {
		return err
	}
	if err := persistProposalMemory(tx, store); err != nil {
		return err
	}
	if err := persistRepoChangeJobs(tx, store); err != nil {
		return err
	}
	if err := persistPRAttempts(tx, store); err != nil {
		return err
	}
	if err := persistPostMergeReplays(tx, store); err != nil {
		return err
	}
	return persistCronLeases(tx, store)
}

func loadThreadPolicies(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select thread_key, state, owner_bot, muted, close_reason, last_policy_version, updated_at from thread_policy order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item policy.ThreadPolicy
		var state string
		var closeReason sql.NullString
		if err := rows.Scan(&item.ThreadKey, &state, &item.OwnerBot, &item.Muted, &closeReason, &item.LastPolicyVersion, &item.UpdatedAt); err != nil {
			return err
		}
		item.State = policy.ThreadState(state)
		item.CloseReason = closeReason.String
		store.threadPolicies[item.ThreadKey] = item
	}
	return rows.Err()
}

func loadChannelPolicies(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select channel_id, proactive_enabled, auto_post_allowed, allowed_workflow_kinds, updated_at from channel_policy order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item policy.ChannelPolicy
		var raw []byte
		if err := rows.Scan(&item.ChannelID, &item.ProactiveEnabled, &item.AutoPostAllowed, &raw, &item.UpdatedAt); err != nil {
			return err
		}
		item.AllowedWorkflowKinds = decodeJSON(raw, []string{})
		store.channelPolicy = append(store.channelPolicy, item)
	}
	return rows.Err()
}

func loadOwnership(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select domain, owner_team, escalation_slack from ownership_registry order by domain`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item registry.OwnershipRecord
		if err := rows.Scan(&item.Domain, &item.OwnerTeam, &item.EscalationSlack); err != nil {
			return err
		}
		store.ownership = append(store.ownership, item)
	}
	return rows.Err()
}

func loadCapabilities(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select name, kind, allowed_bots, approval_needed from capability_registry order by name`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item registry.CapabilityRecord
		var raw []byte
		if err := rows.Scan(&item.Name, &item.Kind, &raw, &item.ApprovalNeeded); err != nil {
			return err
		}
		item.AllowedBots = decodeJSON(raw, []string{})
		store.capabilities = append(store.capabilities, item)
	}
	return rows.Err()
}

func loadTemplates(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select name, kind, description, steps from workflow_templates order by name`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item registry.WorkflowTemplate
		var raw []byte
		if err := rows.Scan(&item.Name, &item.Kind, &item.Description, &raw); err != nil {
			return err
		}
		item.Steps = decodeJSON(raw, []string{})
		store.templates = append(store.templates, item)
	}
	return rows.Err()
}

func loadExperiments(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select name, candidate, baseline, state, reviewed_by from experiment_registry order by name`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item registry.ExperimentRecord
		var reviewedBy sql.NullString
		if err := rows.Scan(&item.Name, &item.Candidate, &item.Baseline, &item.State, &reviewedBy); err != nil {
			return err
		}
		item.ReviewedBy = reviewedBy.String
		store.experiments = append(store.experiments, item)
	}
	return rows.Err()
}

func loadEvents(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, source, source_event_id, thread_key, incident_key, dedupe_key, severity, normalized_problem_statement, ownership_hint, raw_payload_ref, workflow_hint, metadata, created_at from event_envelope order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item ingestion.EventEnvelope
		var threadKey, incidentKey, ownershipHint, rawPayloadRef, workflowHint sql.NullString
		var raw []byte
		var source, severity string
		if err := rows.Scan(&item.ID, &source, &item.SourceEventID, &threadKey, &incidentKey, &item.DedupeKey, &severity, &item.NormalizedProblemStatement, &ownershipHint, &rawPayloadRef, &workflowHint, &raw, &item.CreatedAt); err != nil {
			return err
		}
		item.Source = ingestion.Source(source)
		item.ThreadKey = threadKey.String
		item.IncidentKey = incidentKey.String
		item.Severity = ingestion.Severity(severity)
		item.OwnershipHint = ownershipHint.String
		item.RawPayloadRef = rawPayloadRef.String
		item.WorkflowHint = workflowHint.String
		item.Metadata = decodeJSON(raw, map[string]interface{}{})
		store.events = append(store.events, item)
	}
	return rows.Err()
}

func loadIngestions(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, event_id, thread_key, thread_ts, workflow_hint, intent, bot_role, source, channel_id, user_id, text, created_at from ingestion order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item slack.Ingestion
		var eventID, threadTS, intent, botRole sql.NullString
		if err := rows.Scan(&item.ID, &eventID, &item.ThreadKey, &threadTS, &item.WorkflowHint, &intent, &botRole, &item.Source, &item.ChannelID, &item.UserID, &item.Text, &item.CreatedAt); err != nil {
			return err
		}
		item.EventID = eventID.String
		item.ThreadTS = threadTS.String
		item.Intent = intent.String
		item.BotRole = slack.BotRole(botRole.String)
		store.ingestions = append(store.ingestions, item)
	}
	return rows.Err()
}

func loadWorkflows(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, ingestion_id, trace_id, thread_key, kind, intent, assigned_bot, approval_mode, response_mode, status, last_error, created_at, updated_at, completed_at from workflow order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item Workflow
		var ingestionID, traceID, intent, approvalMode, responseMode, lastError sql.NullString
		var completedAt sql.NullTime
		if err := rows.Scan(&item.ID, &ingestionID, &traceID, &item.ThreadKey, &item.Kind, &intent, &item.AssignedBot, &approvalMode, &responseMode, &item.Status, &lastError, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
			return err
		}
		item.IngestionID = ingestionID.String
		item.TraceID = traceID.String
		item.Intent = intent.String
		item.ApprovalMode = approvalMode.String
		item.ResponseMode = responseMode.String
		item.LastError = lastError.String
		if completedAt.Valid {
			t := completedAt.Time
			item.CompletedAt = &t
		}
		store.workflows = append(store.workflows, item)
	}
	return rows.Err()
}

func loadAssignments(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, thread_key, assigned_bot, confidence, rationale, created_at from assignment order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item Assignment
		if err := rows.Scan(&item.ID, &item.ThreadKey, &item.AssignedBot, &item.Confidence, &item.Rationale, &item.CreatedAt); err != nil {
			return err
		}
		store.assignments = append(store.assignments, item)
	}
	return rows.Err()
}

func loadTraces(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select trace_id, ingestion_id, workflow_id, thread_key, workflow_kind, status, last_verdict, started_at, ended_at, event_count, artifact_count, reasoning_step_count, tool_call_count, slack_action_count from trace_summary order by started_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var summary events.TraceSummary
		var status string
		var lastVerdict sql.NullString
		if err := rows.Scan(&summary.TraceID, &summary.IngestionID, &summary.WorkflowID, &summary.ThreadKey, &summary.WorkflowKind, &status, &lastVerdict, &summary.StartedAt, &summary.EndedAt, &summary.EventCount, &summary.ArtifactCount, &summary.ReasoningStepCount, &summary.ToolCallCount, &summary.SlackActionCount); err != nil {
			return err
		}
		summary.Status = events.Status(status)
		summary.LastVerdict = lastVerdict.String
		store.traces[summary.TraceID] = events.Trace{Summary: summary}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = r.Query(`select trace_id, ingestion_id, workflow_id, parent_event_id, plane, service, actor, event_type, status, started_at, ended_at, payload_ref, artifact_ref, cost_tokens, latency_ms, description from trace_event order by started_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.TraceEvent
		var status string
		var parentEvent, payloadRef, artifactRef, description sql.NullString
		var endedAt sql.NullTime
		if err := rows.Scan(&item.TraceID, &item.IngestionID, &item.WorkflowID, &parentEvent, &item.Plane, &item.Service, &item.Actor, &item.EventType, &status, &item.StartedAt, &endedAt, &payloadRef, &artifactRef, &item.CostTokens, &item.LatencyMs, &description); err != nil {
			return err
		}
		item.Status = events.Status(status)
		item.ParentEvent = parentEvent.String
		item.PayloadRef = payloadRef.String
		item.ArtifactRef = artifactRef.String
		item.Description = description.String
		if endedAt.Valid {
			t := endedAt.Time
			item.EndedAt = &t
		}
		trace := store.traces[item.TraceID]
		trace.Events = append(trace.Events, item)
		store.traces[item.TraceID] = trace
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = r.Query(`select id, trace_id, kind, content_type, url, size_bytes, source from artifact order by trace_id, id`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.Artifact
		if err := rows.Scan(&item.ID, &item.TraceID, &item.Kind, &item.ContentType, &item.URL, &item.SizeBytes, &item.Source); err != nil {
			return err
		}
		trace := store.traces[item.TraceID]
		trace.Artifacts = append(trace.Artifacts, item)
		store.traces[item.TraceID] = trace
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = r.Query(`select id, trace_id, workflow_id, step_type, summary, evidence_refs, alternatives, confidence, decision, created_at from reasoning_step order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.ReasoningStep
		var workflowID, decision sql.NullString
		var evidenceRefs, alternatives []byte
		if err := rows.Scan(&item.ID, &item.TraceID, &workflowID, &item.StepType, &item.Summary, &evidenceRefs, &alternatives, &item.Confidence, &decision, &item.CreatedAt); err != nil {
			return err
		}
		item.WorkflowID = workflowID.String
		item.Decision = decision.String
		item.EvidenceRefs = decodeJSON(evidenceRefs, []events.EvidenceRef{})
		item.Alternatives = decodeJSON(alternatives, []string{})
		trace := store.traces[item.TraceID]
		trace.Reasoning = append(trace.Reasoning, item)
		store.traces[item.TraceID] = trace
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = r.Query(`select id, trace_id, workflow_id, tool_name, tool_call_id, request, summary, raw_artifact_refs, approval_state, interpretation_summary, status, created_at from tool_call_record order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.ToolCallRecord
		var workflowID, summary, approvalState, interpretationSummary, status sql.NullString
		var request, rawArtifactRefs []byte
		if err := rows.Scan(&item.ID, &item.TraceID, &workflowID, &item.ToolName, &item.ToolCallID, &request, &summary, &rawArtifactRefs, &approvalState, &interpretationSummary, &status, &item.CreatedAt); err != nil {
			return err
		}
		item.WorkflowID = workflowID.String
		item.Request = decodeJSON(request, map[string]interface{}{})
		item.Summary = summary.String
		item.RawArtifactRefs = decodeJSON(rawArtifactRefs, []string{})
		item.ApprovalState = approvalState.String
		item.InterpretationSummary = interpretationSummary.String
		item.Status = status.String
		trace := store.traces[item.TraceID]
		trace.ToolCalls = append(trace.ToolCalls, item)
		store.traces[item.TraceID] = trace
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = r.Query(`select id, trace_id, workflow_id, channel_id, thread_ts, idempotency_key, draft_body, final_body, policy_verdict, send_status, artifact_refs, created_at from slack_action_record order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.SlackActionRecord
		var workflowID, channelID, threadTS, draftBody, finalBody, policyVerdict, sendStatus sql.NullString
		var artifactRefs []byte
		if err := rows.Scan(&item.ID, &item.TraceID, &workflowID, &channelID, &threadTS, &item.IdempotencyKey, &draftBody, &finalBody, &policyVerdict, &sendStatus, &artifactRefs, &item.CreatedAt); err != nil {
			return err
		}
		item.WorkflowID = workflowID.String
		item.ChannelID = channelID.String
		item.ThreadTS = threadTS.String
		item.DraftBody = draftBody.String
		item.FinalBody = finalBody.String
		item.PolicyVerdict = policyVerdict.String
		item.SendStatus = sendStatus.String
		item.ArtifactRefs = decodeJSON(artifactRefs, []string{})
		trace := store.traces[item.TraceID]
		trace.SlackActions = append(trace.SlackActions, item)
		store.traces[item.TraceID] = trace
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for traceID, trace := range store.traces {
		recomputeTraceSummary(&trace)
		store.traces[traceID] = trace
	}
	return nil
}

func loadRatings(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select trace_id, score, verdict, labels, notes, reviewer_id, created_at from human_rating order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.HumanRating
		var raw []byte
		var notes sql.NullString
		if err := rows.Scan(&item.TraceID, &item.Score, &item.Verdict, &raw, &notes, &item.ReviewerID, &item.CreatedAt); err != nil {
			return err
		}
		item.Labels = decodeJSON(raw, []string{})
		item.Notes = notes.String
		store.ratings[item.TraceID] = append(store.ratings[item.TraceID], item)
	}
	return rows.Err()
}

func loadNotes(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select trace_id, category, note, suggested_owner, created_by, created_at from improvement_note order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.ImprovementNote
		var suggestedOwner sql.NullString
		if err := rows.Scan(&item.TraceID, &item.Category, &item.Note, &suggestedOwner, &item.CreatedBy, &item.CreatedAt); err != nil {
			return err
		}
		item.SuggestedOwner = suggestedOwner.String
		store.notes[item.TraceID] = append(store.notes[item.TraceID], item)
	}
	return rows.Err()
}

func loadEvalSuites(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select name, description, event_kinds, layers from eval_suite order by name`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item evals.Suite
		var eventKinds, layers []byte
		if err := rows.Scan(&item.Name, &item.Description, &eventKinds, &layers); err != nil {
			return err
		}
		item.EventKinds = decodeJSON(eventKinds, []string{})
		item.Layers = decodeJSON(layers, []evals.Layer{})
		store.evalSuites = append(store.evalSuites, item)
	}
	return rows.Err()
}

func loadEvalRuns(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, trace_id, event_id, suite_name, status, trigger, overall_score, overall_verdict, created_at, completed_at from eval_run order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item evals.Run
		var eventID sql.NullString
		var status string
		var completedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.TraceID, &eventID, &item.SuiteName, &status, &item.Trigger, &item.OverallScore, &item.OverallVerdict, &item.CreatedAt, &completedAt); err != nil {
			return err
		}
		item.EventID = eventID.String
		item.Status = evals.Status(status)
		if completedAt.Valid {
			item.CompletedAt = completedAt.Time
		}
		store.evalRuns[item.ID] = item
	}
	return rows.Err()
}

func loadEvalJudgments(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, eval_run_id, layer, category, score, passed, rationale, created_at from eval_judgment order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item evals.Judgment
		var layer string
		if err := rows.Scan(&item.ID, &item.EvalRunID, &layer, &item.Category, &item.Score, &item.Passed, &item.Rationale, &item.CreatedAt); err != nil {
			return err
		}
		item.Layer = evals.Layer(layer)
		store.evalJudgments[item.EvalRunID] = append(store.evalJudgments[item.EvalRunID], item)
	}
	return rows.Err()
}

func loadSettings(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select key, active_proposal_cap, updated_at from improvement_settings`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var item improvement.Settings
		if err := rows.Scan(&key, &item.ActiveProposalCap, &item.UpdatedAt); err != nil {
			return err
		}
		if key == "default" {
			store.settings = item
		}
	}
	store.settings = normalizedSettings(store.settings)
	return rows.Err()
}

func loadWorkItems(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, queue, kind, status, trace_id, workflow_id, ingestion_id, proposal_id, thread_key, intent, repo_scope, requested_by, approval_mode, response_mode, payload, attempts, lease_owner, lease_expires_at, last_error, created_at, updated_at, completed_at from work_item order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item queue.WorkItem
		var queueName, status string
		var traceID, workflowID, ingestionID, proposalID, threadKey, intent, repoScope, requestedBy, approvalMode, responseMode, leaseOwner, lastError sql.NullString
		var payload []byte
		var leaseExpiresAt, completedAt sql.NullTime
		if err := rows.Scan(&item.ID, &queueName, &item.Kind, &status, &traceID, &workflowID, &ingestionID, &proposalID, &threadKey, &intent, &repoScope, &requestedBy, &approvalMode, &responseMode, &payload, &item.Attempts, &leaseOwner, &leaseExpiresAt, &lastError, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
			return err
		}
		item.Queue = queue.QueueName(queueName)
		item.Status = queue.WorkItemStatus(status)
		item.TraceID = traceID.String
		item.WorkflowID = workflowID.String
		item.IngestionID = ingestionID.String
		item.ProposalID = proposalID.String
		item.ThreadKey = threadKey.String
		item.Intent = intent.String
		item.RepoScope = repoScope.String
		item.RequestedBy = requestedBy.String
		item.ApprovalMode = approvalMode.String
		item.ResponseMode = responseMode.String
		item.Payload = decodeJSON(payload, map[string]interface{}{})
		item.LeaseOwner = leaseOwner.String
		item.LastError = lastError.String
		if leaseExpiresAt.Valid {
			t := leaseExpiresAt.Time
			item.LeaseExpiresAt = &t
		}
		if completedAt.Valid {
			t := completedAt.Time
			item.CompletedAt = &t
		}
		store.workItems[item.ID] = item
	}
	return rows.Err()
}

func loadCandidates(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, candidate_key, subsystem, failure_mode, intervention_type, status, severity, recurrence_count, expected_impact, novelty_score, confidence_score, freshness_score, priority_score, risk_tier, hypothesis, proposed_scope, latest_trace_id, source_eval_ids, evidence_artifact_ids, prior_similar_proposal_ids, new_evidence_since_last_rejection, last_evaluated_at, created_at, updated_at from improvement_candidate order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.Candidate
		var status, riskTier string
		var latestTraceID sql.NullString
		var sourceEvalIDs, evidenceArtifactIDs, priorSimilarProposalIDs []byte
		var lastEvaluatedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.CandidateKey, &item.Subsystem, &item.FailureMode, &item.InterventionType, &status, &item.Severity, &item.RecurrenceCount, &item.ExpectedImpact, &item.NoveltyScore, &item.ConfidenceScore, &item.FreshnessScore, &item.PriorityScore, &riskTier, &item.Hypothesis, &item.ProposedScope, &latestTraceID, &sourceEvalIDs, &evidenceArtifactIDs, &priorSimilarProposalIDs, &item.NewEvidenceSinceLastRejection, &lastEvaluatedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		item.Status = improvement.CandidateStatus(status)
		item.RiskTier = improvement.RiskTier(riskTier)
		item.LatestTraceID = latestTraceID.String
		item.SourceEvalIDs = decodeJSON(sourceEvalIDs, []string{})
		item.EvidenceArtifactIDs = decodeJSON(evidenceArtifactIDs, []string{})
		item.PriorSimilarProposalIDs = decodeJSON(priorSimilarProposalIDs, []string{})
		if lastEvaluatedAt.Valid {
			item.LastEvaluatedAt = lastEvaluatedAt.Time
		}
		store.candidates[item.CandidateKey] = item
	}
	return rows.Err()
}

func loadProposals(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, trace_id, title, category, summary, status, reviewer, candidate_key, source_eval_ids, risk_tier, proposed_scope, evidence_artifact_ids, active_slot_consuming, review_deadline, prior_similar_proposal_ids, new_evidence_since_last_rejection, created_at from proposal order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.Proposal
		var status string
		var reviewer sql.NullString
		var sourceEvalIDs, evidenceArtifactIDs, priorSimilarProposalIDs []byte
		var reviewDeadline sql.NullTime
		if err := rows.Scan(&item.ID, &item.TraceID, &item.Title, &item.Category, &item.Summary, &status, &reviewer, &item.CandidateKey, &sourceEvalIDs, &item.RiskTier, &item.ProposedScope, &evidenceArtifactIDs, &item.ActiveSlotConsuming, &reviewDeadline, &priorSimilarProposalIDs, &item.NewEvidenceSinceLastRejection, &item.CreatedAt); err != nil {
			return err
		}
		item.Status = review.ProposalStatus(status)
		item.Reviewer = reviewer.String
		item.SourceEvalIDs = decodeJSON(sourceEvalIDs, []string{})
		item.EvidenceArtifactIDs = decodeJSON(evidenceArtifactIDs, []string{})
		item.PriorSimilarProposalIDs = decodeJSON(priorSimilarProposalIDs, []string{})
		if reviewDeadline.Valid {
			item.ReviewDeadline = reviewDeadline.Time
		}
		store.proposals[item.ID] = item
	}
	return rows.Err()
}

func loadProposalReviews(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select proposal_id, decision, rationale, reviewer_id, failure_class, failure_classes, created_at from proposal_review order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.ProposalReview
		var failureClass sql.NullString
		var failureClasses []byte
		if err := rows.Scan(&item.ProposalID, &item.Decision, &item.Rationale, &item.ReviewerID, &failureClass, &failureClasses, &item.CreatedAt); err != nil {
			return err
		}
		item.FailureClass = failureClass.String
		item.FailureClasses = decodeJSON(failureClasses, []string{})
		proposal := store.proposals[item.ProposalID]
		proposal.Reviews = append(proposal.Reviews, item)
		store.proposals[item.ProposalID] = proposal
	}
	return rows.Err()
}

func loadProposalMemory(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, proposal_id, candidate_key, hypothesis, diff_summary, review_rationale, disposition, disposition_reason, failure_class, failure_classes, source_eval_ids, linked_artifact_ids, linked_proposal_ids, created_at from proposal_memory order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.ProposalMemory
		var disposition string
		var dispositionReason, failureClass sql.NullString
		var failureClasses, sourceEvalIDs, linkedArtifactIDs, linkedProposalIDs []byte
		if err := rows.Scan(&item.ID, &item.ProposalID, &item.CandidateKey, &item.Hypothesis, &item.DiffSummary, &item.ReviewRationale, &disposition, &dispositionReason, &failureClass, &failureClasses, &sourceEvalIDs, &linkedArtifactIDs, &linkedProposalIDs, &item.CreatedAt); err != nil {
			return err
		}
		item.Disposition = review.ProposalStatus(disposition)
		item.DispositionReason = dispositionReason.String
		item.FailureClass = failureClass.String
		item.FailureClasses = decodeJSON(failureClasses, []string{})
		item.SourceEvalIDs = decodeJSON(sourceEvalIDs, []string{})
		item.LinkedArtifactIDs = decodeJSON(linkedArtifactIDs, []string{})
		item.LinkedProposalIDs = decodeJSON(linkedProposalIDs, []string{})
		store.proposalMemory = append(store.proposalMemory, item)
	}
	return rows.Err()
}

func loadRepoChangeJobs(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, proposal_id, candidate_key, status, repo, base_ref, branch_name, allowed_path_globs, context_summary, created_at from repo_change_job order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.RepoChangeJob
		var allowed []byte
		if err := rows.Scan(&item.ID, &item.ProposalID, &item.CandidateKey, &item.Status, &item.Repo, &item.BaseRef, &item.BranchName, &allowed, &item.ContextSummary, &item.CreatedAt); err != nil {
			return err
		}
		item.AllowedPathGlobs = decodeJSON(allowed, []string{})
		store.repoChangeJobs[item.ID] = item
	}
	return rows.Err()
}

func loadPRAttempts(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, proposal_id, repo, branch_name, pr_url, status, validation_status, created_at from pr_attempt order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.PRAttempt
		var prURL sql.NullString
		if err := rows.Scan(&item.ID, &item.ProposalID, &item.Repo, &item.BranchName, &prURL, &item.Status, &item.ValidationStatus, &item.CreatedAt); err != nil {
			return err
		}
		item.PRURL = prURL.String
		store.prAttempts[item.ID] = item
	}
	return rows.Err()
}

func loadPostMergeReplays(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, proposal_id, trace_id, baseline_score, candidate_score, improved, created_at from post_merge_replay order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.PostMergeReplay
		if err := rows.Scan(&item.ID, &item.ProposalID, &item.TraceID, &item.BaselineScore, &item.CandidateScore, &item.Improved, &item.CreatedAt); err != nil {
			return err
		}
		store.postMergeReplay[item.ID] = item
	}
	return rows.Err()
}

func loadCronLeases(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select name, holder, expires_at from cron_lease`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.CronLease
		if err := rows.Scan(&item.Name, &item.Holder, &item.ExpiresAt); err != nil {
			return err
		}
		store.cronLeases[item.Name] = item
	}
	return rows.Err()
}

func persistThreadPolicies(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.threadPolicies)
	for _, key := range keys {
		item := store.threadPolicies[key]
		if _, err := tx.Exec(`insert into thread_policy (thread_key, state, owner_bot, muted, close_reason, last_policy_version, updated_at) values ($1,$2,$3,$4,$5,$6,$7)`,
			item.ThreadKey, string(item.State), item.OwnerBot, item.Muted, nullString(item.CloseReason), item.LastPolicyVersion, item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistChannelPolicies(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.channelPolicy {
		if _, err := tx.Exec(`insert into channel_policy (channel_id, proactive_enabled, auto_post_allowed, allowed_workflow_kinds, updated_at) values ($1,$2,$3,$4::jsonb,$5)`,
			item.ChannelID, item.ProactiveEnabled, item.AutoPostAllowed, jsonString(item.AllowedWorkflowKinds), item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistOwnership(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.ownership {
		if _, err := tx.Exec(`insert into ownership_registry (domain, owner_team, escalation_slack) values ($1,$2,$3)`,
			item.Domain, item.OwnerTeam, item.EscalationSlack,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistCapabilities(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.capabilities {
		if _, err := tx.Exec(`insert into capability_registry (name, kind, allowed_bots, approval_needed) values ($1,$2,$3::jsonb,$4)`,
			item.Name, item.Kind, jsonString(item.AllowedBots), item.ApprovalNeeded,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistTemplates(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.templates {
		if _, err := tx.Exec(`insert into workflow_templates (name, kind, description, steps) values ($1,$2,$3,$4::jsonb)`,
			item.Name, item.Kind, item.Description, jsonString(item.Steps),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistExperiments(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.experiments {
		if _, err := tx.Exec(`insert into experiment_registry (name, candidate, baseline, state, reviewed_by) values ($1,$2,$3,$4,$5)`,
			item.Name, item.Candidate, item.Baseline, item.State, nullString(item.ReviewedBy),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistEvents(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.events {
		if _, err := tx.Exec(`insert into event_envelope (id, source, source_event_id, thread_key, incident_key, dedupe_key, severity, normalized_problem_statement, ownership_hint, raw_payload_ref, workflow_hint, metadata, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,$13)`,
			item.ID, string(item.Source), item.SourceEventID, nullString(item.ThreadKey), nullString(item.IncidentKey), item.DedupeKey, string(item.Severity), item.NormalizedProblemStatement, nullString(item.OwnershipHint), nullString(item.RawPayloadRef), nullString(item.WorkflowHint), jsonString(item.Metadata), item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistIngestions(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.ingestions {
		if _, err := tx.Exec(`insert into ingestion (id, event_id, thread_key, thread_ts, workflow_hint, intent, bot_role, source, channel_id, user_id, text, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			item.ID, nullString(item.EventID), item.ThreadKey, nullString(item.ThreadTS), item.WorkflowHint, nullString(item.Intent), nullString(string(item.BotRole)), item.Source, item.ChannelID, item.UserID, item.Text, item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistWorkflows(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.workflows {
		if _, err := tx.Exec(`insert into workflow (id, ingestion_id, trace_id, thread_key, kind, intent, assigned_bot, approval_mode, response_mode, status, last_error, created_at, updated_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
			item.ID, nullString(item.IngestionID), nullString(item.TraceID), item.ThreadKey, item.Kind, nullString(item.Intent), item.AssignedBot, nullString(item.ApprovalMode), nullString(item.ResponseMode), item.Status, nullString(item.LastError), item.CreatedAt, item.UpdatedAt, nullTime(item.CompletedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistAssignments(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.assignments {
		if _, err := tx.Exec(`insert into assignment (id, thread_key, assigned_bot, confidence, rationale, created_at) values ($1,$2,$3,$4,$5,$6)`,
			item.ID, item.ThreadKey, item.AssignedBot, item.Confidence, item.Rationale, item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistTraces(tx *sql.Tx, store *MemoryStore) error {
	traceIDs := sortedMapKeys(store.traces)
	for _, traceID := range traceIDs {
		trace := store.traces[traceID]
		if _, err := tx.Exec(`insert into trace_summary (trace_id, ingestion_id, workflow_id, thread_key, workflow_kind, status, last_verdict, started_at, ended_at, event_count, artifact_count, reasoning_step_count, tool_call_count, slack_action_count) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
			trace.Summary.TraceID, trace.Summary.IngestionID, trace.Summary.WorkflowID, trace.Summary.ThreadKey, trace.Summary.WorkflowKind, string(trace.Summary.Status), nullString(trace.Summary.LastVerdict), trace.Summary.StartedAt, trace.Summary.EndedAt, trace.Summary.EventCount, trace.Summary.ArtifactCount, trace.Summary.ReasoningStepCount, trace.Summary.ToolCallCount, trace.Summary.SlackActionCount,
		); err != nil {
			return err
		}
		for _, event := range trace.Events {
			if _, err := tx.Exec(`insert into trace_event (trace_id, ingestion_id, workflow_id, parent_event_id, plane, service, actor, event_type, status, started_at, ended_at, payload_ref, artifact_ref, cost_tokens, latency_ms, description) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
				event.TraceID, event.IngestionID, event.WorkflowID, nullString(event.ParentEvent), event.Plane, event.Service, event.Actor, event.EventType, string(event.Status), event.StartedAt, nullTime(event.EndedAt), nullString(event.PayloadRef), nullString(event.ArtifactRef), event.CostTokens, event.LatencyMs, nullString(event.Description),
			); err != nil {
				return err
			}
		}
		for _, artifact := range trace.Artifacts {
			if _, err := tx.Exec(`insert into artifact (id, trace_id, kind, content_type, url, size_bytes, source) values ($1,$2,$3,$4,$5,$6,$7)`,
				artifact.ID, artifact.TraceID, artifact.Kind, artifact.ContentType, artifact.URL, artifact.SizeBytes, artifact.Source,
			); err != nil {
				return err
			}
		}
		for _, item := range trace.Reasoning {
			if _, err := tx.Exec(`insert into reasoning_step (id, trace_id, workflow_id, step_type, summary, evidence_refs, alternatives, confidence, decision, created_at) values ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb,$8,$9,$10)`,
				item.ID, item.TraceID, nullString(item.WorkflowID), item.StepType, item.Summary, jsonString(item.EvidenceRefs), jsonString(item.Alternatives), item.Confidence, nullString(item.Decision), item.CreatedAt,
			); err != nil {
				return err
			}
		}
		for _, item := range trace.ToolCalls {
			if _, err := tx.Exec(`insert into tool_call_record (id, trace_id, workflow_id, tool_name, tool_call_id, request, summary, raw_artifact_refs, approval_state, interpretation_summary, status, created_at) values ($1,$2,$3,$4,$5,$6::jsonb,$7,$8::jsonb,$9,$10,$11,$12)`,
				item.ID, item.TraceID, nullString(item.WorkflowID), item.ToolName, item.ToolCallID, jsonString(item.Request), nullString(item.Summary), jsonString(item.RawArtifactRefs), nullString(item.ApprovalState), nullString(item.InterpretationSummary), nullString(item.Status), item.CreatedAt,
			); err != nil {
				return err
			}
		}
		for _, item := range trace.SlackActions {
			if _, err := tx.Exec(`insert into slack_action_record (id, trace_id, workflow_id, channel_id, thread_ts, idempotency_key, draft_body, final_body, policy_verdict, send_status, artifact_refs, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12)`,
				item.ID, item.TraceID, nullString(item.WorkflowID), nullString(item.ChannelID), nullString(item.ThreadTS), item.IdempotencyKey, nullString(item.DraftBody), nullString(item.FinalBody), nullString(item.PolicyVerdict), nullString(item.SendStatus), jsonString(item.ArtifactRefs), item.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistRatings(tx *sql.Tx, store *MemoryStore) error {
	traceIDs := sortedMapKeys(store.ratings)
	for _, traceID := range traceIDs {
		for _, item := range store.ratings[traceID] {
			if _, err := tx.Exec(`insert into human_rating (trace_id, score, verdict, labels, notes, reviewer_id, created_at) values ($1,$2,$3,$4::jsonb,$5,$6,$7)`,
				item.TraceID, item.Score, item.Verdict, jsonString(item.Labels), nullString(item.Notes), item.ReviewerID, item.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistNotes(tx *sql.Tx, store *MemoryStore) error {
	traceIDs := sortedMapKeys(store.notes)
	for _, traceID := range traceIDs {
		for _, item := range store.notes[traceID] {
			if _, err := tx.Exec(`insert into improvement_note (trace_id, category, note, suggested_owner, created_by, created_at) values ($1,$2,$3,$4,$5,$6)`,
				item.TraceID, item.Category, item.Note, nullString(item.SuggestedOwner), item.CreatedBy, item.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistEvalSuites(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.evalSuites {
		if _, err := tx.Exec(`insert into eval_suite (name, description, event_kinds, layers) values ($1,$2,$3::jsonb,$4::jsonb)`,
			item.Name, item.Description, jsonString(item.EventKinds), jsonString(item.Layers),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistEvalRuns(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.evalRuns)
	for _, key := range keys {
		item := store.evalRuns[key]
		if _, err := tx.Exec(`insert into eval_run (id, trace_id, event_id, suite_name, status, trigger, overall_score, overall_verdict, created_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			item.ID, item.TraceID, nullString(item.EventID), item.SuiteName, string(item.Status), item.Trigger, item.OverallScore, item.OverallVerdict, item.CreatedAt, nullTimeValue(item.CompletedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistEvalJudgments(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.evalJudgments)
	for _, key := range keys {
		for _, item := range store.evalJudgments[key] {
			if _, err := tx.Exec(`insert into eval_judgment (id, eval_run_id, layer, category, score, passed, rationale, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
				item.ID, item.EvalRunID, string(item.Layer), item.Category, item.Score, item.Passed, item.Rationale, item.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistCandidates(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.candidates)
	for _, key := range keys {
		item := store.candidates[key]
		if _, err := tx.Exec(`insert into improvement_candidate (id, candidate_key, subsystem, failure_mode, intervention_type, status, severity, recurrence_count, expected_impact, novelty_score, confidence_score, freshness_score, priority_score, risk_tier, hypothesis, proposed_scope, latest_trace_id, source_eval_ids, evidence_artifact_ids, prior_similar_proposal_ids, new_evidence_since_last_rejection, last_evaluated_at, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18::jsonb,$19::jsonb,$20::jsonb,$21,$22,$23,$24)`,
			item.ID, item.CandidateKey, item.Subsystem, item.FailureMode, item.InterventionType, string(item.Status), item.Severity, item.RecurrenceCount, item.ExpectedImpact, item.NoveltyScore, item.ConfidenceScore, item.FreshnessScore, item.PriorityScore, string(item.RiskTier), item.Hypothesis, item.ProposedScope, nullString(item.LatestTraceID), jsonString(item.SourceEvalIDs), jsonString(item.EvidenceArtifactIDs), jsonString(item.PriorSimilarProposalIDs), item.NewEvidenceSinceLastRejection, nullTimeValue(item.LastEvaluatedAt), item.CreatedAt, item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistProposals(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.proposals)
	for _, key := range keys {
		item := store.proposals[key]
		if _, err := tx.Exec(`insert into proposal (id, trace_id, title, category, summary, status, reviewer, candidate_key, source_eval_ids, risk_tier, proposed_scope, evidence_artifact_ids, active_slot_consuming, review_deadline, prior_similar_proposal_ids, new_evidence_since_last_rejection, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11,$12::jsonb,$13,$14,$15::jsonb,$16,$17)`,
			item.ID, item.TraceID, item.Title, item.Category, item.Summary, string(item.Status), nullString(item.Reviewer), item.CandidateKey, jsonString(item.SourceEvalIDs), item.RiskTier, item.ProposedScope, jsonString(item.EvidenceArtifactIDs), item.ActiveSlotConsuming, nullTimeValue(item.ReviewDeadline), jsonString(item.PriorSimilarProposalIDs), item.NewEvidenceSinceLastRejection, item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistProposalReviews(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.proposals)
	for _, key := range keys {
		for _, item := range store.proposals[key].Reviews {
			if _, err := tx.Exec(`insert into proposal_review (proposal_id, decision, rationale, reviewer_id, failure_class, failure_classes, created_at) values ($1,$2,$3,$4,$5,$6::jsonb,$7)`,
				item.ProposalID, item.Decision, item.Rationale, item.ReviewerID, nullString(item.FailureClass), jsonString(item.FailureClasses), item.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistProposalMemory(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.proposalMemory {
		if _, err := tx.Exec(`insert into proposal_memory (id, proposal_id, candidate_key, hypothesis, diff_summary, review_rationale, disposition, disposition_reason, failure_class, failure_classes, source_eval_ids, linked_artifact_ids, linked_proposal_ids, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11::jsonb,$12::jsonb,$13::jsonb,$14)`,
			item.ID, item.ProposalID, item.CandidateKey, item.Hypothesis, item.DiffSummary, item.ReviewRationale, string(item.Disposition), nullString(item.DispositionReason), nullString(item.FailureClass), jsonString(item.FailureClasses), jsonString(item.SourceEvalIDs), jsonString(item.LinkedArtifactIDs), jsonString(item.LinkedProposalIDs), item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistSettings(tx *sql.Tx, store *MemoryStore) error {
	item := normalizedSettings(store.settings)
	if _, err := tx.Exec(`insert into improvement_settings (key, active_proposal_cap, updated_at) values ($1,$2,$3)`,
		"default", item.ActiveProposalCap, item.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}

func persistWorkItems(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.workItems)
	for _, key := range keys {
		item := store.workItems[key]
		if _, err := tx.Exec(`insert into work_item (id, queue, kind, status, trace_id, workflow_id, ingestion_id, proposal_id, thread_key, intent, repo_scope, requested_by, approval_mode, response_mode, payload, attempts, lease_owner, lease_expires_at, last_error, created_at, updated_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::jsonb,$16,$17,$18,$19,$20,$21,$22)`,
			item.ID, string(item.Queue), item.Kind, string(item.Status), nullString(item.TraceID), nullString(item.WorkflowID), nullString(item.IngestionID), nullString(item.ProposalID), nullString(item.ThreadKey), nullString(item.Intent), nullString(item.RepoScope), nullString(item.RequestedBy), nullString(item.ApprovalMode), nullString(item.ResponseMode), jsonString(item.Payload), item.Attempts, nullString(item.LeaseOwner), nullTime(item.LeaseExpiresAt), nullString(item.LastError), item.CreatedAt, item.UpdatedAt, nullTime(item.CompletedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistRepoChangeJobs(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.repoChangeJobs)
	for _, key := range keys {
		item := store.repoChangeJobs[key]
		if _, err := tx.Exec(`insert into repo_change_job (id, proposal_id, candidate_key, status, repo, base_ref, branch_name, allowed_path_globs, context_summary, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9,$10)`,
			item.ID, item.ProposalID, item.CandidateKey, item.Status, item.Repo, item.BaseRef, item.BranchName, jsonString(item.AllowedPathGlobs), item.ContextSummary, item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistPRAttempts(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.prAttempts)
	for _, key := range keys {
		item := store.prAttempts[key]
		if _, err := tx.Exec(`insert into pr_attempt (id, proposal_id, repo, branch_name, pr_url, status, validation_status, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
			item.ID, item.ProposalID, item.Repo, item.BranchName, nullString(item.PRURL), item.Status, item.ValidationStatus, item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistPostMergeReplays(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.postMergeReplay)
	for _, key := range keys {
		item := store.postMergeReplay[key]
		if _, err := tx.Exec(`insert into post_merge_replay (id, proposal_id, trace_id, baseline_score, candidate_score, improved, created_at) values ($1,$2,$3,$4,$5,$6,$7)`,
			item.ID, item.ProposalID, item.TraceID, item.BaselineScore, item.CandidateScore, item.Improved, item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistCronLeases(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.cronLeases)
	for _, key := range keys {
		item := store.cronLeases[key]
		if _, err := tx.Exec(`insert into cron_lease (name, holder, expires_at) values ($1,$2,$3)`,
			item.Name, item.Holder, item.ExpiresAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func decodeJSON[T any](raw []byte, fallback T) T {
	if len(raw) == 0 {
		return fallback
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return fallback
	}
	return out
}

func jsonString(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullTime(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return *value
}

func nullTimeValue(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func sortedMapKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func (p *PostgresStore) ListEvents() []ingestion.EventEnvelope {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListEvents()
}

func (p *PostgresStore) CreateEvent(event ingestion.EventEnvelope) (created ingestion.EventEnvelope, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var createErr error
		created, createErr = store.CreateEvent(event)
		return createErr
	})
	return
}

func (p *PostgresStore) ListIngestions() []slack.Ingestion {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListIngestions()
}

func (p *PostgresStore) CreateIngestion(envelope slack.SlackEnvelope) slack.Ingestion {
	var created slack.Ingestion
	_ = p.mutate(func(store *MemoryStore) error {
		created = store.CreateIngestion(envelope)
		return nil
	})
	return created
}

func (p *PostgresStore) ListWorkflows() []Workflow {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListWorkflows()
}

func (p *PostgresStore) ListAssignments() []Assignment {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListAssignments()
}

func (p *PostgresStore) ListThreadPolicies() []policy.ThreadPolicy {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListThreadPolicies()
}

func (p *PostgresStore) ListChannelPolicies() []policy.ChannelPolicy {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListChannelPolicies()
}

func (p *PostgresStore) SetThreadState(threadKey string, state policy.ThreadState, owner string) (item policy.ThreadPolicy, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.SetThreadState(threadKey, state, owner)
		return inner
	})
	return
}

func (p *PostgresStore) ListOwnershipRecords() []registry.OwnershipRecord {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListOwnershipRecords()
}

func (p *PostgresStore) ListCapabilities() []registry.CapabilityRecord {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListCapabilities()
}

func (p *PostgresStore) ListTemplates() []registry.WorkflowTemplate {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListTemplates()
}

func (p *PostgresStore) ListExperiments() []registry.ExperimentRecord {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListExperiments()
}

func (p *PostgresStore) ListTraces() []events.TraceSummary {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListTraces()
}

func (p *PostgresStore) GetTrace(traceID string) (events.Trace, bool) {
	store, err := p.readStore()
	if err != nil {
		return events.Trace{}, false
	}
	return store.GetTrace(traceID)
}

func (p *PostgresStore) ListRatings(traceID string) []review.HumanRating {
	store, err := p.readStore()
	if err != nil {
		return []review.HumanRating{}
	}
	return store.ListRatings(traceID)
}

func (p *PostgresStore) ListImprovementNotes(traceID string) []review.ImprovementNote {
	store, err := p.readStore()
	if err != nil {
		return []review.ImprovementNote{}
	}
	return store.ListImprovementNotes(traceID)
}

func (p *PostgresStore) AddRating(traceID string, rating review.HumanRating) (item review.HumanRating, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.AddRating(traceID, rating)
		return inner
	})
	return
}

func (p *PostgresStore) AddImprovementNote(traceID string, note review.ImprovementNote) (item review.ImprovementNote, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.AddImprovementNote(traceID, note)
		return inner
	})
	return
}

func (p *PostgresStore) ScheduleReplay(traceID string, requestedBy string) (item queue.WorkItem, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.ScheduleReplay(traceID, requestedBy)
		return inner
	})
	return
}

func (p *PostgresStore) ListEvalSuites() []evals.Suite {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListEvalSuites()
}

func (p *PostgresStore) ListEvalRuns() []evals.Run {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListEvalRuns()
}

func (p *PostgresStore) ListEvalJudgments(evalRunID string) []evals.Judgment {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListEvalJudgments(evalRunID)
}

func (p *PostgresStore) EvaluateTrace(traceID string, trigger string) (run evals.Run, judgments []evals.Judgment, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		run, judgments, inner = store.EvaluateTrace(traceID, trigger)
		return inner
	})
	return
}

func (p *PostgresStore) GetSettings() improvement.Settings {
	store, err := p.readStore()
	if err != nil {
		return improvement.Settings{ActiveProposalCap: defaultProposalSlotCap}
	}
	return store.GetSettings()
}

func (p *PostgresStore) UpdateSettings(settings improvement.Settings) (item improvement.Settings, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.UpdateSettings(settings)
		return inner
	})
	return
}

func (p *PostgresStore) ListWorkItems() []queue.WorkItem {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListWorkItems()
}

func (p *PostgresStore) EnqueueWorkItem(item queue.WorkItem) (created queue.WorkItem, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		created, inner = store.EnqueueWorkItem(item)
		return inner
	})
	return
}

func (p *PostgresStore) ClaimNextWorkItem(queues []queue.QueueName, holder string, lease time.Duration) (item queue.WorkItem, ok bool, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, ok, inner = store.ClaimNextWorkItem(queues, holder, lease)
		return inner
	})
	return
}

func (p *PostgresStore) CompleteWorkItem(id string) (item queue.WorkItem, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.CompleteWorkItem(id)
		return inner
	})
	return
}

func (p *PostgresStore) FailWorkItem(id string, lastError string) (item queue.WorkItem, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.FailWorkItem(id, lastError)
		return inner
	})
	return
}

func (p *PostgresStore) UpdateWorkflowStatus(workflowID string, status string, lastError string) (item Workflow, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.UpdateWorkflowStatus(workflowID, status, lastError)
		return inner
	})
	return
}

func (p *PostgresStore) ApplyTraceUpdate(traceID string, update TraceUpdate) (trace events.Trace, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		trace, inner = store.ApplyTraceUpdate(traceID, update)
		return inner
	})
	return
}

func (p *PostgresStore) ListCandidates() []improvement.Candidate {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListCandidates()
}

func (p *PostgresStore) ListProposalMemories() []review.ProposalMemory {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListProposalMemories()
}

func (p *PostgresStore) GetProposalSlots() ProposalSlotState {
	store, err := p.readStore()
	if err != nil {
		return ProposalSlotState{Cap: defaultProposalSlotCap}
	}
	return store.GetProposalSlots()
}

func (p *PostgresStore) PromoteCandidates(requestedBy string, limit int) (result PromotionResult, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		result, inner = store.PromoteCandidates(requestedBy, limit)
		return inner
	})
	return
}

func (p *PostgresStore) RunProposalPromoter(holder string) (result PromotionResult, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		result, inner = store.RunProposalPromoter(holder)
		return inner
	})
	return
}

func (p *PostgresStore) ListProposals() []review.Proposal {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListProposals()
}

func (p *PostgresStore) ReviewProposal(proposalID string, decision review.ProposalReview) (proposal review.Proposal, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		proposal, inner = store.ReviewProposal(proposalID, decision)
		return inner
	})
	return
}

func (p *PostgresStore) UpdateProposalStatus(proposalID string, status review.ProposalStatus) (proposal review.Proposal, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		proposal, inner = store.UpdateProposalStatus(proposalID, status)
		return inner
	})
	return
}

func (p *PostgresStore) MaterializeApprovedProposal(proposalID string, requestedBy string) (job improvement.RepoChangeJob, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		job, inner = store.MaterializeApprovedProposal(proposalID, requestedBy)
		return inner
	})
	return
}

func (p *PostgresStore) ListRepoChangeJobs() []improvement.RepoChangeJob {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListRepoChangeJobs()
}

func (p *PostgresStore) UpdateRepoChangeJobStatus(jobID string, status string) (item improvement.RepoChangeJob, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.UpdateRepoChangeJobStatus(jobID, status)
		return inner
	})
	return
}

func (p *PostgresStore) ListPRAttempts() []improvement.PRAttempt {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListPRAttempts()
}

func (p *PostgresStore) RecordPRAttempt(attempt improvement.PRAttempt) (item improvement.PRAttempt, err error) {
	err = p.mutate(func(store *MemoryStore) error {
		var inner error
		item, inner = store.RecordPRAttempt(attempt)
		return inner
	})
	return
}

func (p *PostgresStore) ListPostMergeReplays() []improvement.PostMergeReplay {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListPostMergeReplays()
}

func (p *PostgresStore) ExecuteTool(name string, input map[string]interface{}) ToolResult {
	var result ToolResult
	_ = p.mutate(func(store *MemoryStore) error {
		result = store.ExecuteTool(name, input)
		return nil
	})
	return result
}
