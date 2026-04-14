package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	platformdb "github.com/piplabs/rsi-agent-platform/internal/db"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/registry"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
)

type PostgresStore struct {
	db           *sql.DB
	schemaStatus platformdb.SchemaStatus
}

type sqlReader interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func OpenStore(cfg config.Config) (Store, error) {
	switch strings.TrimSpace(cfg.StoreBackend) {
	case "postgres":
		return NewPostgresStore(cfg)
	case "memory":
		return NewMemoryStore(), nil
	default:
		return nil, fmt.Errorf("unsupported RSI_STORE_BACKEND %q", cfg.StoreBackend)
	}
}

func MustOpenStore(cfg config.Config) Store {
	store, err := OpenStore(cfg)
	if err != nil {
		panic(err)
	}
	return store
}

func NewPostgresStore(cfg config.Config) (*PostgresStore, error) {
	db, err := platformdb.OpenPostgres(cfg.PostgresURL)
	if err != nil {
		return nil, err
	}
	status, err := platformdb.VerifyCompatible(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	store := &PostgresStore{db: db, schemaStatus: status}
	if err := store.ensureSeed(cfg); err != nil {
		return nil, err
	}
	return store, nil
}

func (p *PostgresStore) SchemaStatus() platformdb.SchemaStatus {
	return p.schemaStatus
}

func (p *PostgresStore) ensureSeed(cfg config.Config) error {
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

	bootstrap := newEmptyMemoryStore()
	if cfg.DefaultProposalCap > 0 {
		bootstrap.settings.ActiveProposalCap = cfg.DefaultProposalCap
	}
	if err := persistStore(tx, bootstrap); err != nil {
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
	if err := loadConversations(r, store); err != nil {
		return nil, err
	}
	if err := loadConversationEntries(r, store); err != nil {
		return nil, err
	}
	if err := loadCases(r, store); err != nil {
		return nil, err
	}
	if err := loadActionIntents(r, store); err != nil {
		return nil, err
	}
	if err := loadActionResults(r, store); err != nil {
		return nil, err
	}
	if err := loadOutcomes(r, store); err != nil {
		return nil, err
	}
	if err := loadKnowledgeEntries(r, store); err != nil {
		return nil, err
	}
	if err := loadKnowledgeEvidence(r, store); err != nil {
		return nil, err
	}
	if err := loadKnowledgeReviews(r, store); err != nil {
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
	if err := loadFeedback(r, store); err != nil {
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

	if len(store.events) == 0 && len(store.workflows) == 0 && len(store.conversations) == 0 {
		return store, nil
	}
	backfillConversationCaseV2(store)
	backfillActionOutcomeKnowledgeV3(store)
	return store, nil
}

func persistStore(tx *sql.Tx, store *MemoryStore) error {
	for _, table := range []string{
		"proposal_review",
		"proposal_memory",
		"pr_attempt",
		"repo_change_job",
		"post_merge_replay",
		"knowledge_review",
		"knowledge_evidence_link",
		"knowledge_entry",
		"outcome_record",
		"action_result",
		"action_intent",
		"cron_lease",
		"slack_action_record",
		"tool_call_record",
		"reasoning_step",
		"work_item",
		"feedback_record",
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
		"case_record",
		"conversation_entry",
		"conversation",
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
	if err := persistConversations(tx, store); err != nil {
		return err
	}
	if err := persistConversationEntries(tx, store); err != nil {
		return err
	}
	if err := persistCases(tx, store); err != nil {
		return err
	}
	if err := persistActionIntents(tx, store); err != nil {
		return err
	}
	if err := persistActionResults(tx, store); err != nil {
		return err
	}
	if err := persistOutcomes(tx, store); err != nil {
		return err
	}
	if err := persistKnowledgeEntries(tx, store); err != nil {
		return err
	}
	if err := persistKnowledgeEvidence(tx, store); err != nil {
		return err
	}
	if err := persistKnowledgeReviews(tx, store); err != nil {
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
	if err := persistFeedback(tx, store); err != nil {
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

func loadConversations(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, source, external_key, external_conversation, title, status, participant_ids, active_case_id, latest_event_id, created_at, updated_at from conversation order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item conversation.Conversation
		var source, status string
		var participantIDs []byte
		var activeCaseID, latestEventID sql.NullString
		if err := rows.Scan(&item.ID, &source, &item.ExternalKey, &item.ExternalConversation, &item.Title, &status, &participantIDs, &activeCaseID, &latestEventID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		item.Source = ingestion.Source(source)
		item.Status = conversation.Status(status)
		item.ParticipantIDs = decodeJSON(participantIDs, []string{})
		item.ActiveCaseID = activeCaseID.String
		item.LatestEventID = latestEventID.String
		store.conversations[item.ID] = item
	}
	return rows.Err()
}

func loadConversationEntries(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, conversation_id, event_id, trace_id, source, source_event_id, entry_type, actor_id, actor_type, body, metadata, created_at from conversation_entry order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item conversation.Entry
		var eventID, traceID, actorID, actorType sql.NullString
		var source string
		var metadata []byte
		if err := rows.Scan(&item.ID, &item.ConversationID, &eventID, &traceID, &source, &item.SourceEventID, &item.EntryType, &actorID, &actorType, &item.Body, &metadata, &item.CreatedAt); err != nil {
			return err
		}
		item.EventID = eventID.String
		item.TraceID = traceID.String
		item.Source = ingestion.Source(source)
		item.ActorID = actorID.String
		item.ActorType = actorType.String
		item.Metadata = decodeJSON(metadata, map[string]interface{}{})
		store.conversationEntries = append(store.conversationEntries, item)
	}
	return rows.Err()
}

func loadCases(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, conversation_id, kind, intent, title, summary, status, approval_mode, response_mode, assigned_bot, opened_by_event_id, closed_by_event_id, latest_trace_id, resolution_state, resolved_at, latest_outcome_id, outcome_score, superseded_by_case_id, created_at, updated_at, closed_at from case_record order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item conversation.Case
		var status string
		var approvalMode, responseMode, openedByEventID, closedByEventID, latestTraceID, resolutionState, latestOutcomeID, supersededByCaseID sql.NullString
		var resolvedAt, closedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.Kind, &item.Intent, &item.Title, &item.Summary, &status, &approvalMode, &responseMode, &item.AssignedBot, &openedByEventID, &closedByEventID, &latestTraceID, &resolutionState, &resolvedAt, &latestOutcomeID, &item.OutcomeScore, &supersededByCaseID, &item.CreatedAt, &item.UpdatedAt, &closedAt); err != nil {
			return err
		}
		item.Status = conversation.CaseStatus(status)
		item.ApprovalMode = approvalMode.String
		item.ResponseMode = responseMode.String
		item.OpenedByEventID = openedByEventID.String
		item.ClosedByEventID = closedByEventID.String
		item.LatestTraceID = latestTraceID.String
		item.ResolutionState = conversation.ResolutionState(resolutionState.String)
		item.LatestOutcomeID = latestOutcomeID.String
		item.SupersededByCaseID = supersededByCaseID.String
		if resolvedAt.Valid {
			t := resolvedAt.Time
			item.ResolvedAt = &t
		}
		if closedAt.Valid {
			t := closedAt.Time
			item.ClosedAt = &t
		}
		store.cases[item.ID] = item
	}
	return rows.Err()
}

func loadActionIntents(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, owner_plane, conversation_id, case_id, trace_id, proposal_id, kind, phase_key, target_ref, request_payload, idempotency_key, approval_mode, approval_state, policy_verdict, status, superseded_by_action_id, requested_by, rationale, evidence_refs, created_at, updated_at from action_intent order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item action.Intent
		var conversationID, caseID, traceID, proposalID, phaseKey, targetRef, idempotencyKey, approvalMode, approvalState, policyVerdict, supersededBy, requestedBy, rationale sql.NullString
		var requestPayload, evidenceRefs []byte
		var kind, status string
		if err := rows.Scan(&item.ID, &item.OwnerPlane, &conversationID, &caseID, &traceID, &proposalID, &kind, &phaseKey, &targetRef, &requestPayload, &idempotencyKey, &approvalMode, &approvalState, &policyVerdict, &status, &supersededBy, &requestedBy, &rationale, &evidenceRefs, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.TraceID = traceID.String
		item.ProposalID = proposalID.String
		item.Kind = action.Kind(kind)
		item.PhaseKey = phaseKey.String
		item.TargetRef = targetRef.String
		item.RequestPayload = decodeJSON(requestPayload, map[string]any{})
		item.IdempotencyKey = idempotencyKey.String
		item.ApprovalMode = approvalMode.String
		item.ApprovalState = approvalState.String
		item.PolicyVerdict = policyVerdict.String
		item.Status = action.Status(status)
		item.SupersededByActionID = supersededBy.String
		item.RequestedBy = requestedBy.String
		item.Rationale = rationale.String
		item.EvidenceRefs = decodeJSON(evidenceRefs, []events.EvidenceRef{})
		store.actionIntents[item.ID] = item
	}
	return rows.Err()
}

func loadActionResults(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, action_intent_id, attempt_number, executor, provider, provider_ref, request_artifact_id, response_artifact_id, status, error_code, error_message, started_at, completed_at from action_result order by action_intent_id asc, attempt_number asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item action.Result
		var provider, providerRef, requestArtifactID, responseArtifactID, errorCode, errorMessage sql.NullString
		var status string
		if err := rows.Scan(&item.ID, &item.ActionIntentID, &item.AttemptNumber, &item.Executor, &provider, &providerRef, &requestArtifactID, &responseArtifactID, &status, &errorCode, &errorMessage, &item.StartedAt, &item.CompletedAt); err != nil {
			return err
		}
		item.Provider = provider.String
		item.ProviderRef = providerRef.String
		item.RequestArtifactID = requestArtifactID.String
		item.ResponseArtifactID = responseArtifactID.String
		item.Status = action.Status(status)
		item.ErrorCode = errorCode.String
		item.ErrorMessage = errorMessage.String
		store.actionResults[item.ActionIntentID] = append(store.actionResults[item.ActionIntentID], item)
	}
	return rows.Err()
}

func loadOutcomes(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, source, source_event_id, conversation_id, case_id, trace_id, proposal_id, outcome_type, verdict, score, summary, details, external_ref, recorded_by, recorded_at from outcome_record order by recorded_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item outcome.Record
		var sourceEventID, conversationID, caseID, traceID, proposalID, summary, details, externalRef, recordedBy sql.NullString
		var outcomeType, verdict string
		if err := rows.Scan(&item.ID, &item.Source, &sourceEventID, &conversationID, &caseID, &traceID, &proposalID, &outcomeType, &verdict, &item.Score, &summary, &details, &externalRef, &recordedBy, &item.RecordedAt); err != nil {
			return err
		}
		item.SourceEventID = sourceEventID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.TraceID = traceID.String
		item.ProposalID = proposalID.String
		item.OutcomeType = outcome.Type(outcomeType)
		item.Verdict = outcome.Verdict(verdict)
		item.Summary = summary.String
		item.Details = details.String
		item.ExternalRef = externalRef.String
		item.RecordedBy = recordedBy.String
		store.outcomes[item.ID] = item
	}
	return rows.Err()
}

func loadKnowledgeEntries(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, tier, kind, scope_type, scope_id, title, summary, body, structured_facts, status, confidence, fresh_until, source_type, supersedes_entry_id, contradicted_by_entry_id, created_at, updated_at from knowledge_entry order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item knowledge.Entry
		var scopeID, summary, body, supersedesEntryID, contradictedByEntryID sql.NullString
		var structuredFacts []byte
		var freshUntil sql.NullTime
		var tier, kind, scopeType, status, sourceType string
		if err := rows.Scan(&item.ID, &tier, &kind, &scopeType, &scopeID, &item.Title, &summary, &body, &structuredFacts, &status, &item.Confidence, &freshUntil, &sourceType, &supersedesEntryID, &contradictedByEntryID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		item.Tier = knowledge.Tier(tier)
		item.Kind = knowledge.Kind(kind)
		item.ScopeType = knowledge.ScopeType(scopeType)
		item.ScopeID = scopeID.String
		item.Summary = summary.String
		item.Body = body.String
		item.StructuredFacts = decodeJSON(structuredFacts, map[string]any{})
		item.Status = knowledge.Status(status)
		if freshUntil.Valid {
			t := freshUntil.Time
			item.FreshUntil = &t
		}
		item.SourceType = knowledge.SourceType(sourceType)
		item.SupersedesEntryID = supersedesEntryID.String
		item.ContradictedByEntryID = contradictedByEntryID.String
		store.knowledgeEntries[item.ID] = item
	}
	return rows.Err()
}

func loadKnowledgeEvidence(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select knowledge_entry_id, evidence_type, evidence_id, relevance_summary, evidence_ref from knowledge_evidence_link order by knowledge_entry_id asc, id asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item knowledge.EvidenceLink
		var relevance sql.NullString
		var evidenceRef []byte
		if err := rows.Scan(&item.KnowledgeEntryID, &item.EvidenceType, &item.EvidenceID, &relevance, &evidenceRef); err != nil {
			return err
		}
		item.RelevanceSummary = relevance.String
		item.EvidenceRef = decodeJSON(evidenceRef, events.EvidenceRef{})
		store.knowledgeEvidence[item.KnowledgeEntryID] = append(store.knowledgeEvidence[item.KnowledgeEntryID], item)
	}
	return rows.Err()
}

func loadKnowledgeReviews(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, knowledge_entry_id, decision, reviewer_id, rationale, created_at from knowledge_review order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item knowledge.Review
		var rationale sql.NullString
		if err := rows.Scan(&item.ID, &item.KnowledgeEntryID, &item.Decision, &item.ReviewerID, &rationale, &item.CreatedAt); err != nil {
			return err
		}
		item.Rationale = rationale.String
		store.knowledgeReviews[item.KnowledgeEntryID] = append(store.knowledgeReviews[item.KnowledgeEntryID], item)
	}
	return rows.Err()
}

func loadIngestions(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, event_id, conversation_id, case_id, thread_key, thread_ts, workflow_hint, intent, bot_role, source, channel_id, user_id, text, created_at from ingestion order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item slack.Ingestion
		var eventID, conversationID, caseID, threadTS, intent, botRole sql.NullString
		if err := rows.Scan(&item.ID, &eventID, &conversationID, &caseID, &item.ThreadKey, &threadTS, &item.WorkflowHint, &intent, &botRole, &item.Source, &item.ChannelID, &item.UserID, &item.Text, &item.CreatedAt); err != nil {
			return err
		}
		item.EventID = eventID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.ThreadTS = threadTS.String
		item.Intent = intent.String
		item.BotRole = slack.BotRole(botRole.String)
		store.ingestions = append(store.ingestions, item)
	}
	return rows.Err()
}

func loadWorkflows(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, ingestion_id, trace_id, conversation_id, case_id, thread_key, kind, intent, assigned_bot, approval_mode, response_mode, status, last_error, created_at, updated_at, completed_at from workflow order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item Workflow
		var ingestionID, traceID, conversationID, caseID, intent, approvalMode, responseMode, lastError sql.NullString
		var completedAt sql.NullTime
		if err := rows.Scan(&item.ID, &ingestionID, &traceID, &conversationID, &caseID, &item.ThreadKey, &item.Kind, &intent, &item.AssignedBot, &approvalMode, &responseMode, &item.Status, &lastError, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
			return err
		}
		item.IngestionID = ingestionID.String
		item.TraceID = traceID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
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
	rows, err := r.Query(`select id, conversation_id, case_id, thread_key, assigned_bot, confidence, rationale, created_at from assignment order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item Assignment
		var conversationID, caseID sql.NullString
		if err := rows.Scan(&item.ID, &conversationID, &caseID, &item.ThreadKey, &item.AssignedBot, &item.Confidence, &item.Rationale, &item.CreatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		store.assignments = append(store.assignments, item)
	}
	return rows.Err()
}

func loadTraces(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select trace_id, ingestion_id, workflow_id, conversation_id, case_id, trigger_event_id, supersedes_trace_id, thread_key, workflow_kind, status, last_verdict, started_at, ended_at, event_count, artifact_count, reasoning_step_count, tool_call_count, slack_action_count from trace_summary order by started_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var summary events.TraceSummary
		var status string
		var conversationID, caseID, triggerEventID, supersedesTraceID, lastVerdict sql.NullString
		if err := rows.Scan(&summary.TraceID, &summary.IngestionID, &summary.WorkflowID, &conversationID, &caseID, &triggerEventID, &supersedesTraceID, &summary.ThreadKey, &summary.WorkflowKind, &status, &lastVerdict, &summary.StartedAt, &summary.EndedAt, &summary.EventCount, &summary.ArtifactCount, &summary.ReasoningStepCount, &summary.ToolCallCount, &summary.SlackActionCount); err != nil {
			return err
		}
		summary.ConversationID = conversationID.String
		summary.CaseID = caseID.String
		summary.TriggerEventID = triggerEventID.String
		summary.SupersedesTraceID = supersedesTraceID.String
		summary.Status = events.Status(status)
		summary.LastVerdict = lastVerdict.String
		store.traces[summary.TraceID] = events.Trace{Summary: summary}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	rows, err = r.Query(`select trace_id, ingestion_id, workflow_id, conversation_id, case_id, trigger_event_id, parent_event_id, plane, service, actor, event_type, status, started_at, ended_at, payload_ref, artifact_ref, cost_tokens, latency_ms, description from trace_event order by started_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.TraceEvent
		var status string
		var conversationID, caseID, triggerEventID, parentEvent, payloadRef, artifactRef, description sql.NullString
		var endedAt sql.NullTime
		if err := rows.Scan(&item.TraceID, &item.IngestionID, &item.WorkflowID, &conversationID, &caseID, &triggerEventID, &parentEvent, &item.Plane, &item.Service, &item.Actor, &item.EventType, &status, &item.StartedAt, &endedAt, &payloadRef, &artifactRef, &item.CostTokens, &item.LatencyMs, &description); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.TriggerEventID = triggerEventID.String
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

	rows, err = r.Query(`select id, trace_id, workflow_id, conversation_id, case_id, step_type, summary, evidence_refs, alternatives, confidence, decision, created_at from reasoning_step order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.ReasoningStep
		var workflowID, conversationID, caseID, decision sql.NullString
		var evidenceRefs, alternatives []byte
		if err := rows.Scan(&item.ID, &item.TraceID, &workflowID, &conversationID, &caseID, &item.StepType, &item.Summary, &evidenceRefs, &alternatives, &item.Confidence, &decision, &item.CreatedAt); err != nil {
			return err
		}
		item.WorkflowID = workflowID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
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

	rows, err = r.Query(`select id, trace_id, workflow_id, conversation_id, case_id, tool_name, tool_call_id, request, summary, raw_artifact_refs, approval_state, interpretation_summary, status, created_at from tool_call_record order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.ToolCallRecord
		var workflowID, conversationID, caseID, summary, approvalState, interpretationSummary, status sql.NullString
		var request, rawArtifactRefs []byte
		if err := rows.Scan(&item.ID, &item.TraceID, &workflowID, &conversationID, &caseID, &item.ToolName, &item.ToolCallID, &request, &summary, &rawArtifactRefs, &approvalState, &interpretationSummary, &status, &item.CreatedAt); err != nil {
			return err
		}
		item.WorkflowID = workflowID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
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

	rows, err = r.Query(`select id, trace_id, workflow_id, conversation_id, case_id, channel_id, thread_ts, idempotency_key, draft_body, final_body, policy_verdict, send_status, artifact_refs, created_at from slack_action_record order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item events.SlackActionRecord
		var workflowID, conversationID, caseID, channelID, threadTS, draftBody, finalBody, policyVerdict, sendStatus sql.NullString
		var artifactRefs []byte
		if err := rows.Scan(&item.ID, &item.TraceID, &workflowID, &conversationID, &caseID, &channelID, &threadTS, &item.IdempotencyKey, &draftBody, &finalBody, &policyVerdict, &sendStatus, &artifactRefs, &item.CreatedAt); err != nil {
			return err
		}
		item.WorkflowID = workflowID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
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

func loadFeedback(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, conversation_id, case_id, trace_id, target_type, target_id, score, verdict, labels, notes, reviewer_id, created_at from feedback_record order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.FeedbackRecord
		var conversationID, caseID, traceID, verdict, notes sql.NullString
		var labels []byte
		var targetType string
		if err := rows.Scan(&item.ID, &conversationID, &caseID, &traceID, &targetType, &item.TargetID, &item.Score, &verdict, &labels, &notes, &item.ReviewerID, &item.CreatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.TraceID = traceID.String
		item.TargetType = review.FeedbackTargetType(targetType)
		item.Verdict = verdict.String
		item.Labels = decodeJSON(labels, []string{})
		item.Notes = notes.String
		store.feedbackRecords[item.TraceID] = append(store.feedbackRecords[item.TraceID], item)
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
	rows, err := r.Query(`select id, queue, kind, status, trace_id, workflow_id, ingestion_id, conversation_id, case_id, trigger_event_id, proposal_id, thread_key, intent, repo_scope, requested_by, approval_mode, response_mode, payload, attempts, lease_owner, lease_expires_at, last_error, created_at, updated_at, completed_at from work_item order by created_at asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item queue.WorkItem
		var queueName, status string
		var traceID, workflowID, ingestionID, conversationID, caseID, triggerEventID, proposalID, threadKey, intent, repoScope, requestedBy, approvalMode, responseMode, leaseOwner, lastError sql.NullString
		var payload []byte
		var leaseExpiresAt, completedAt sql.NullTime
		if err := rows.Scan(&item.ID, &queueName, &item.Kind, &status, &traceID, &workflowID, &ingestionID, &conversationID, &caseID, &triggerEventID, &proposalID, &threadKey, &intent, &repoScope, &requestedBy, &approvalMode, &responseMode, &payload, &item.Attempts, &leaseOwner, &leaseExpiresAt, &lastError, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
			return err
		}
		item.Queue = queue.QueueName(queueName)
		item.Status = queue.WorkItemStatus(status)
		item.TraceID = traceID.String
		item.WorkflowID = workflowID.String
		item.IngestionID = ingestionID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.TriggerEventID = triggerEventID.String
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
	rows, err := r.Query(`select id, candidate_key, conversation_id, case_id, origin_trace_id, evidence_trace_ids, subsystem, failure_mode, intervention_type, target_layer, target_kind, target_ref, status, severity, recurrence_count, expected_impact, novelty_score, confidence_score, freshness_score, priority_score, risk_tier, hypothesis, proposed_scope, latest_trace_id, source_eval_ids, evidence_artifact_ids, prior_similar_proposal_ids, new_evidence_since_last_rejection, last_evaluated_at, created_at, updated_at from improvement_candidate order by updated_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.Candidate
		var status, riskTier, targetLayer string
		var conversationID, caseID, originTraceID, latestTraceID, targetKind, targetRef sql.NullString
		var evidenceTraceIDs, sourceEvalIDs, evidenceArtifactIDs, priorSimilarProposalIDs []byte
		var lastEvaluatedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.CandidateKey, &conversationID, &caseID, &originTraceID, &evidenceTraceIDs, &item.Subsystem, &item.FailureMode, &item.InterventionType, &targetLayer, &targetKind, &targetRef, &status, &item.Severity, &item.RecurrenceCount, &item.ExpectedImpact, &item.NoveltyScore, &item.ConfidenceScore, &item.FreshnessScore, &item.PriorityScore, &riskTier, &item.Hypothesis, &item.ProposedScope, &latestTraceID, &sourceEvalIDs, &evidenceArtifactIDs, &priorSimilarProposalIDs, &item.NewEvidenceSinceLastRejection, &lastEvaluatedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.OriginTraceID = originTraceID.String
		item.EvidenceTraceIDs = decodeJSON(evidenceTraceIDs, []string{})
		item.TargetLayer = harness.TargetLayer(targetLayer)
		item.TargetKind = targetKind.String
		item.TargetRef = targetRef.String
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
	rows, err := r.Query(`select id, trace_id, conversation_id, case_id, origin_trace_id, evidence_trace_ids, title, category, summary, status, reviewer, candidate_key, target_layer, target_kind, target_ref, source_eval_ids, risk_tier, proposed_scope, evidence_artifact_ids, active_slot_consuming, review_deadline, prior_similar_proposal_ids, new_evidence_since_last_rejection, created_at from proposal order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.Proposal
		var status, targetLayer string
		var conversationID, caseID, originTraceID, reviewer, targetKind, targetRef sql.NullString
		var evidenceTraceIDs, sourceEvalIDs, evidenceArtifactIDs, priorSimilarProposalIDs []byte
		var reviewDeadline sql.NullTime
		if err := rows.Scan(&item.ID, &item.TraceID, &conversationID, &caseID, &originTraceID, &evidenceTraceIDs, &item.Title, &item.Category, &item.Summary, &status, &reviewer, &item.CandidateKey, &targetLayer, &targetKind, &targetRef, &sourceEvalIDs, &item.RiskTier, &item.ProposedScope, &evidenceArtifactIDs, &item.ActiveSlotConsuming, &reviewDeadline, &priorSimilarProposalIDs, &item.NewEvidenceSinceLastRejection, &item.CreatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.OriginTraceID = originTraceID.String
		item.EvidenceTraceIDs = decodeJSON(evidenceTraceIDs, []string{})
		item.Status = review.ProposalStatus(status)
		item.Reviewer = reviewer.String
		item.TargetLayer = harness.TargetLayer(targetLayer)
		item.TargetKind = targetKind.String
		item.TargetRef = targetRef.String
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
	rows, err := r.Query(`select id, proposal_id, idempotency_key, decision, rationale, reviewer_id, failure_class, failure_classes, created_at from proposal_review order by created_at asc, id asc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.ProposalReview
		var idempotencyKey sql.NullString
		var failureClass sql.NullString
		var failureClasses []byte
		if err := rows.Scan(&item.ID, &item.ProposalID, &idempotencyKey, &item.Decision, &item.Rationale, &item.ReviewerID, &failureClass, &failureClasses, &item.CreatedAt); err != nil {
			return err
		}
		item.IdempotencyKey = idempotencyKey.String
		item.FailureClass = failureClass.String
		item.FailureClasses = decodeJSON(failureClasses, []string{})
		proposal := store.proposals[item.ProposalID]
		proposal.Reviews = append(proposal.Reviews, item)
		store.proposals[item.ProposalID] = proposal
	}
	return rows.Err()
}

func loadProposalMemory(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, review_id, proposal_id, candidate_key, conversation_id, case_id, origin_trace_id, evidence_trace_ids, hypothesis, diff_summary, review_rationale, disposition, disposition_reason, failure_class, failure_classes, source_eval_ids, linked_artifact_ids, linked_proposal_ids, created_at from proposal_memory order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item review.ProposalMemory
		var disposition string
		var reviewID sql.NullInt64
		var conversationID, caseID, originTraceID, dispositionReason, failureClass sql.NullString
		var evidenceTraceIDs, failureClasses, sourceEvalIDs, linkedArtifactIDs, linkedProposalIDs []byte
		if err := rows.Scan(&item.ID, &reviewID, &item.ProposalID, &item.CandidateKey, &conversationID, &caseID, &originTraceID, &evidenceTraceIDs, &item.Hypothesis, &item.DiffSummary, &item.ReviewRationale, &disposition, &dispositionReason, &failureClass, &failureClasses, &sourceEvalIDs, &linkedArtifactIDs, &linkedProposalIDs, &item.CreatedAt); err != nil {
			return err
		}
		if reviewID.Valid {
			item.ReviewID = reviewID.Int64
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.OriginTraceID = originTraceID.String
		item.EvidenceTraceIDs = decodeJSON(evidenceTraceIDs, []string{})
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
	rows, err := r.Query(`select id, proposal_id, conversation_id, case_id, origin_trace_id, candidate_key, status, repo, base_ref, branch_name, allowed_path_globs, context_summary, sandbox_namespace, sandbox_job_name, sandbox_pod_name, validation_error, validation_ref, log_artifact_id, created_at, updated_at from repo_change_job order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.RepoChangeJob
		var conversationID, caseID, originTraceID, sandboxNamespace, sandboxJobName, sandboxPodName, validationError, validationRef, logArtifactID sql.NullString
		var allowed []byte
		if err := rows.Scan(&item.ID, &item.ProposalID, &conversationID, &caseID, &originTraceID, &item.CandidateKey, &item.Status, &item.Repo, &item.BaseRef, &item.BranchName, &allowed, &item.ContextSummary, &sandboxNamespace, &sandboxJobName, &sandboxPodName, &validationError, &validationRef, &logArtifactID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.OriginTraceID = originTraceID.String
		item.AllowedPathGlobs = decodeJSON(allowed, []string{})
		item.SandboxNamespace = sandboxNamespace.String
		item.SandboxJobName = sandboxJobName.String
		item.SandboxPodName = sandboxPodName.String
		item.ValidationError = validationError.String
		item.ValidationRef = validationRef.String
		item.LogArtifactID = logArtifactID.String
		store.repoChangeJobs[item.ID] = item
	}
	return rows.Err()
}

func loadPRAttempts(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, proposal_id, conversation_id, case_id, origin_trace_id, repo, branch_name, pr_url, status, validation_status, created_at from pr_attempt order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.PRAttempt
		var conversationID, caseID, originTraceID, prURL sql.NullString
		if err := rows.Scan(&item.ID, &item.ProposalID, &conversationID, &caseID, &originTraceID, &item.Repo, &item.BranchName, &prURL, &item.Status, &item.ValidationStatus, &item.CreatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.OriginTraceID = originTraceID.String
		item.PRURL = prURL.String
		store.prAttempts[item.ID] = item
	}
	return rows.Err()
}

func loadPostMergeReplays(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, proposal_id, trace_id, conversation_id, case_id, baseline_score, candidate_score, improved, created_at from post_merge_replay order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.PostMergeReplay
		var conversationID, caseID sql.NullString
		if err := rows.Scan(&item.ID, &item.ProposalID, &item.TraceID, &conversationID, &caseID, &item.BaselineScore, &item.CandidateScore, &item.Improved, &item.CreatedAt); err != nil {
			return err
		}
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
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
		if _, err := tx.Exec(`insert into thread_policy (thread_key, state, owner_bot, muted, close_reason, last_policy_version, updated_at) values ($1,$2,$3,$4,$5,$6,$7)
			on conflict (thread_key) do update set
				state = excluded.state,
				owner_bot = excluded.owner_bot,
				muted = excluded.muted,
				close_reason = excluded.close_reason,
				last_policy_version = excluded.last_policy_version,
				updated_at = excluded.updated_at`,
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
		if _, err := tx.Exec(`insert into event_envelope (id, source, source_event_id, thread_key, incident_key, dedupe_key, severity, normalized_problem_statement, ownership_hint, raw_payload_ref, workflow_hint, metadata, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,$13)
			on conflict (id) do update set
				source = excluded.source,
				source_event_id = excluded.source_event_id,
				thread_key = excluded.thread_key,
				incident_key = excluded.incident_key,
				dedupe_key = excluded.dedupe_key,
				severity = excluded.severity,
				normalized_problem_statement = excluded.normalized_problem_statement,
				ownership_hint = excluded.ownership_hint,
				raw_payload_ref = excluded.raw_payload_ref,
				workflow_hint = excluded.workflow_hint,
				metadata = excluded.metadata,
				created_at = excluded.created_at`,
			item.ID, string(item.Source), item.SourceEventID, nullString(item.ThreadKey), nullString(item.IncidentKey), item.DedupeKey, string(item.Severity), item.NormalizedProblemStatement, nullString(item.OwnershipHint), nullString(item.RawPayloadRef), nullString(item.WorkflowHint), jsonString(item.Metadata), item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistConversations(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.conversations)
	for _, key := range keys {
		item := store.conversations[key]
		if _, err := tx.Exec(`insert into conversation (id, source, external_key, external_conversation, title, status, participant_ids, active_case_id, latest_event_id, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9,$10,$11)
			on conflict (id) do update set
				source = excluded.source,
				external_key = excluded.external_key,
				external_conversation = excluded.external_conversation,
				title = excluded.title,
				status = excluded.status,
				participant_ids = excluded.participant_ids,
				active_case_id = excluded.active_case_id,
				latest_event_id = excluded.latest_event_id,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at`,
			item.ID, string(item.Source), item.ExternalKey, item.ExternalConversation, item.Title, string(item.Status), jsonString(item.ParticipantIDs), nullString(item.ActiveCaseID), nullString(item.LatestEventID), item.CreatedAt, item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistConversationEntries(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.conversationEntries {
		if _, err := tx.Exec(`insert into conversation_entry (id, conversation_id, event_id, trace_id, source, source_event_id, entry_type, actor_id, actor_type, body, metadata, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12)
			on conflict (id) do update set
				conversation_id = excluded.conversation_id,
				event_id = excluded.event_id,
				trace_id = excluded.trace_id,
				source = excluded.source,
				source_event_id = excluded.source_event_id,
				entry_type = excluded.entry_type,
				actor_id = excluded.actor_id,
				actor_type = excluded.actor_type,
				body = excluded.body,
				metadata = excluded.metadata,
				created_at = excluded.created_at`,
			item.ID, item.ConversationID, nullString(item.EventID), nullString(item.TraceID), string(item.Source), item.SourceEventID, item.EntryType, nullString(item.ActorID), nullString(item.ActorType), item.Body, jsonString(item.Metadata), item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistCases(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.cases)
	for _, key := range keys {
		item := store.cases[key]
		if _, err := tx.Exec(`insert into case_record (id, conversation_id, kind, intent, title, summary, status, approval_mode, response_mode, assigned_bot, opened_by_event_id, closed_by_event_id, latest_trace_id, resolution_state, resolved_at, latest_outcome_id, outcome_score, superseded_by_case_id, created_at, updated_at, closed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
			on conflict (id) do update set
				conversation_id = excluded.conversation_id,
				kind = excluded.kind,
				intent = excluded.intent,
				title = excluded.title,
				summary = excluded.summary,
				status = excluded.status,
				approval_mode = excluded.approval_mode,
				response_mode = excluded.response_mode,
				assigned_bot = excluded.assigned_bot,
				opened_by_event_id = excluded.opened_by_event_id,
				closed_by_event_id = excluded.closed_by_event_id,
				latest_trace_id = excluded.latest_trace_id,
				resolution_state = excluded.resolution_state,
				resolved_at = excluded.resolved_at,
				latest_outcome_id = excluded.latest_outcome_id,
				outcome_score = excluded.outcome_score,
				superseded_by_case_id = excluded.superseded_by_case_id,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at,
				closed_at = excluded.closed_at`,
			item.ID, item.ConversationID, item.Kind, item.Intent, item.Title, item.Summary, string(item.Status), nullString(item.ApprovalMode), nullString(item.ResponseMode), item.AssignedBot, nullString(item.OpenedByEventID), nullString(item.ClosedByEventID), nullString(item.LatestTraceID), string(item.ResolutionState), nullTime(item.ResolvedAt), nullString(item.LatestOutcomeID), item.OutcomeScore, nullString(item.SupersededByCaseID), item.CreatedAt, item.UpdatedAt, nullTime(item.ClosedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistActionIntents(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.actionIntents)
	for _, key := range keys {
		item := store.actionIntents[key]
		if _, err := tx.Exec(`insert into action_intent (id, owner_plane, conversation_id, case_id, trace_id, proposal_id, kind, phase_key, target_ref, request_payload, idempotency_key, approval_mode, approval_state, policy_verdict, status, superseded_by_action_id, requested_by, rationale, evidence_refs, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11,$12,$13,$14,$15,$16,$17,$18,$19::jsonb,$20,$21)
			on conflict (id) do update set
				owner_plane = excluded.owner_plane,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				trace_id = excluded.trace_id,
				proposal_id = excluded.proposal_id,
				kind = excluded.kind,
				phase_key = excluded.phase_key,
				target_ref = excluded.target_ref,
				request_payload = excluded.request_payload,
				idempotency_key = excluded.idempotency_key,
				approval_mode = excluded.approval_mode,
				approval_state = excluded.approval_state,
				policy_verdict = excluded.policy_verdict,
				status = excluded.status,
				superseded_by_action_id = excluded.superseded_by_action_id,
				requested_by = excluded.requested_by,
				rationale = excluded.rationale,
				evidence_refs = excluded.evidence_refs,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at`,
			item.ID, item.OwnerPlane, nullString(item.ConversationID), nullString(item.CaseID), nullString(item.TraceID), nullString(item.ProposalID), string(item.Kind), nullString(item.PhaseKey), nullString(item.TargetRef), jsonString(item.RequestPayload), nullString(item.IdempotencyKey), nullString(item.ApprovalMode), nullString(item.ApprovalState), nullString(item.PolicyVerdict), string(item.Status), nullString(item.SupersededByActionID), nullString(item.RequestedBy), nullString(item.Rationale), jsonString(item.EvidenceRefs), item.CreatedAt, item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistActionResults(tx *sql.Tx, store *MemoryStore) error {
	intentKeys := sortedMapKeys(store.actionResults)
	for _, key := range intentKeys {
		for _, item := range store.actionResults[key] {
			if _, err := tx.Exec(`insert into action_result (id, action_intent_id, attempt_number, executor, provider, provider_ref, request_artifact_id, response_artifact_id, status, error_code, error_message, started_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
				on conflict (id) do update set
					action_intent_id = excluded.action_intent_id,
					attempt_number = excluded.attempt_number,
					executor = excluded.executor,
					provider = excluded.provider,
					provider_ref = excluded.provider_ref,
					request_artifact_id = excluded.request_artifact_id,
					response_artifact_id = excluded.response_artifact_id,
					status = excluded.status,
					error_code = excluded.error_code,
					error_message = excluded.error_message,
					started_at = excluded.started_at,
					completed_at = excluded.completed_at`,
				item.ID, item.ActionIntentID, item.AttemptNumber, item.Executor, nullString(item.Provider), nullString(item.ProviderRef), nullString(item.RequestArtifactID), nullString(item.ResponseArtifactID), string(item.Status), nullString(item.ErrorCode), nullString(item.ErrorMessage), item.StartedAt, item.CompletedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistOutcomes(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.outcomes)
	for _, key := range keys {
		item := store.outcomes[key]
		if _, err := tx.Exec(`insert into outcome_record (id, source, source_event_id, conversation_id, case_id, trace_id, proposal_id, outcome_type, verdict, score, summary, details, external_ref, recorded_by, recorded_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			on conflict (id) do update set
				source = excluded.source,
				source_event_id = excluded.source_event_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				trace_id = excluded.trace_id,
				proposal_id = excluded.proposal_id,
				outcome_type = excluded.outcome_type,
				verdict = excluded.verdict,
				score = excluded.score,
				summary = excluded.summary,
				details = excluded.details,
				external_ref = excluded.external_ref,
				recorded_by = excluded.recorded_by,
				recorded_at = excluded.recorded_at`,
			item.ID, item.Source, nullString(item.SourceEventID), nullString(item.ConversationID), nullString(item.CaseID), nullString(item.TraceID), nullString(item.ProposalID), string(item.OutcomeType), string(item.Verdict), item.Score, nullString(item.Summary), nullString(item.Details), nullString(item.ExternalRef), nullString(item.RecordedBy), item.RecordedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistKnowledgeEntries(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.knowledgeEntries)
	for _, key := range keys {
		item := store.knowledgeEntries[key]
		if _, err := tx.Exec(`insert into knowledge_entry (id, tier, kind, scope_type, scope_id, title, summary, body, structured_facts, status, confidence, fresh_until, source_type, supersedes_entry_id, contradicted_by_entry_id, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11,$12,$13,$14,$15,$16,$17)
			on conflict (id) do update set
				tier = excluded.tier,
				kind = excluded.kind,
				scope_type = excluded.scope_type,
				scope_id = excluded.scope_id,
				title = excluded.title,
				summary = excluded.summary,
				body = excluded.body,
				structured_facts = excluded.structured_facts,
				status = excluded.status,
				confidence = excluded.confidence,
				fresh_until = excluded.fresh_until,
				source_type = excluded.source_type,
				supersedes_entry_id = excluded.supersedes_entry_id,
				contradicted_by_entry_id = excluded.contradicted_by_entry_id,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at`,
			item.ID, string(item.Tier), string(item.Kind), string(item.ScopeType), nullString(item.ScopeID), item.Title, nullString(item.Summary), nullString(item.Body), jsonString(item.StructuredFacts), string(item.Status), item.Confidence, nullTime(item.FreshUntil), string(item.SourceType), nullString(item.SupersedesEntryID), nullString(item.ContradictedByEntryID), item.CreatedAt, item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistKnowledgeEvidence(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.knowledgeEvidence)
	for _, key := range keys {
		for _, item := range store.knowledgeEvidence[key] {
			if _, err := tx.Exec(`insert into knowledge_evidence_link (knowledge_entry_id, evidence_type, evidence_id, relevance_summary, evidence_ref) values ($1,$2,$3,$4,$5::jsonb)`,
				item.KnowledgeEntryID, item.EvidenceType, item.EvidenceID, nullString(item.RelevanceSummary), jsonString(item.EvidenceRef),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistKnowledgeReviews(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.knowledgeReviews)
	for _, key := range keys {
		for _, item := range store.knowledgeReviews[key] {
			if _, err := tx.Exec(`insert into knowledge_review (id, knowledge_entry_id, decision, reviewer_id, rationale, created_at) values ($1,$2,$3,$4,$5,$6)`,
				item.ID, item.KnowledgeEntryID, item.Decision, item.ReviewerID, nullString(item.Rationale), item.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistIngestions(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.ingestions {
		if _, err := tx.Exec(`insert into ingestion (id, event_id, conversation_id, case_id, thread_key, thread_ts, workflow_hint, intent, bot_role, source, channel_id, user_id, text, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			on conflict (id) do update set
				event_id = excluded.event_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				thread_key = excluded.thread_key,
				thread_ts = excluded.thread_ts,
				workflow_hint = excluded.workflow_hint,
				intent = excluded.intent,
				bot_role = excluded.bot_role,
				source = excluded.source,
				channel_id = excluded.channel_id,
				user_id = excluded.user_id,
				text = excluded.text,
				created_at = excluded.created_at`,
			item.ID, nullString(item.EventID), nullString(item.ConversationID), nullString(item.CaseID), item.ThreadKey, nullString(item.ThreadTS), item.WorkflowHint, nullString(item.Intent), nullString(string(item.BotRole)), item.Source, item.ChannelID, item.UserID, item.Text, item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistWorkflows(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.workflows {
		if _, err := tx.Exec(`insert into workflow (id, ingestion_id, trace_id, conversation_id, case_id, thread_key, kind, intent, assigned_bot, approval_mode, response_mode, status, last_error, created_at, updated_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			on conflict (id) do update set
				ingestion_id = excluded.ingestion_id,
				trace_id = excluded.trace_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				thread_key = excluded.thread_key,
				kind = excluded.kind,
				intent = excluded.intent,
				assigned_bot = excluded.assigned_bot,
				approval_mode = excluded.approval_mode,
				response_mode = excluded.response_mode,
				status = excluded.status,
				last_error = excluded.last_error,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at,
				completed_at = excluded.completed_at`,
			item.ID, nullString(item.IngestionID), nullString(item.TraceID), nullString(item.ConversationID), nullString(item.CaseID), item.ThreadKey, item.Kind, nullString(item.Intent), item.AssignedBot, nullString(item.ApprovalMode), nullString(item.ResponseMode), item.Status, nullString(item.LastError), item.CreatedAt, item.UpdatedAt, nullTime(item.CompletedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistAssignments(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.assignments {
		if _, err := tx.Exec(`insert into assignment (id, conversation_id, case_id, thread_key, assigned_bot, confidence, rationale, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8)
			on conflict (id) do update set
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				thread_key = excluded.thread_key,
				assigned_bot = excluded.assigned_bot,
				confidence = excluded.confidence,
				rationale = excluded.rationale,
				created_at = excluded.created_at`,
			item.ID, nullString(item.ConversationID), nullString(item.CaseID), item.ThreadKey, item.AssignedBot, item.Confidence, item.Rationale, item.CreatedAt,
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
		if _, err := tx.Exec(`insert into trace_summary (trace_id, ingestion_id, workflow_id, conversation_id, case_id, trigger_event_id, supersedes_trace_id, thread_key, workflow_kind, status, last_verdict, started_at, ended_at, event_count, artifact_count, reasoning_step_count, tool_call_count, slack_action_count) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
			on conflict (trace_id) do update set
				ingestion_id = excluded.ingestion_id,
				workflow_id = excluded.workflow_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				trigger_event_id = excluded.trigger_event_id,
				supersedes_trace_id = excluded.supersedes_trace_id,
				thread_key = excluded.thread_key,
				workflow_kind = excluded.workflow_kind,
				status = excluded.status,
				last_verdict = excluded.last_verdict,
				started_at = excluded.started_at,
				ended_at = excluded.ended_at,
				event_count = excluded.event_count,
				artifact_count = excluded.artifact_count,
				reasoning_step_count = excluded.reasoning_step_count,
				tool_call_count = excluded.tool_call_count,
				slack_action_count = excluded.slack_action_count`,
			trace.Summary.TraceID, trace.Summary.IngestionID, trace.Summary.WorkflowID, nullString(trace.Summary.ConversationID), nullString(trace.Summary.CaseID), nullString(trace.Summary.TriggerEventID), nullString(trace.Summary.SupersedesTraceID), trace.Summary.ThreadKey, trace.Summary.WorkflowKind, string(trace.Summary.Status), nullString(trace.Summary.LastVerdict), trace.Summary.StartedAt, trace.Summary.EndedAt, trace.Summary.EventCount, trace.Summary.ArtifactCount, trace.Summary.ReasoningStepCount, trace.Summary.ToolCallCount, trace.Summary.SlackActionCount,
		); err != nil {
			return err
		}
		for _, event := range trace.Events {
			if _, err := tx.Exec(`insert into trace_event (trace_id, ingestion_id, workflow_id, conversation_id, case_id, trigger_event_id, parent_event_id, plane, service, actor, event_type, status, started_at, ended_at, payload_ref, artifact_ref, cost_tokens, latency_ms, description) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
				event.TraceID, event.IngestionID, event.WorkflowID, nullString(event.ConversationID), nullString(event.CaseID), nullString(event.TriggerEventID), nullString(event.ParentEvent), event.Plane, event.Service, event.Actor, event.EventType, string(event.Status), event.StartedAt, nullTime(event.EndedAt), nullString(event.PayloadRef), nullString(event.ArtifactRef), event.CostTokens, event.LatencyMs, nullString(event.Description),
			); err != nil {
				return err
			}
		}
		for _, artifact := range trace.Artifacts {
			if _, err := tx.Exec(`insert into artifact (id, trace_id, kind, content_type, url, size_bytes, source) values ($1,$2,$3,$4,$5,$6,$7)
				on conflict (id) do update set
					trace_id = excluded.trace_id,
					kind = excluded.kind,
					content_type = excluded.content_type,
					url = excluded.url,
					size_bytes = excluded.size_bytes,
					source = excluded.source`,
				artifact.ID, artifact.TraceID, artifact.Kind, artifact.ContentType, artifact.URL, artifact.SizeBytes, artifact.Source,
			); err != nil {
				return err
			}
		}
		for _, item := range trace.Reasoning {
			if _, err := tx.Exec(`insert into reasoning_step (id, trace_id, workflow_id, conversation_id, case_id, step_type, summary, evidence_refs, alternatives, confidence, decision, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9::jsonb,$10,$11,$12)
				on conflict (id) do update set
					trace_id = excluded.trace_id,
					workflow_id = excluded.workflow_id,
					conversation_id = excluded.conversation_id,
					case_id = excluded.case_id,
					step_type = excluded.step_type,
					summary = excluded.summary,
					evidence_refs = excluded.evidence_refs,
					alternatives = excluded.alternatives,
					confidence = excluded.confidence,
					decision = excluded.decision,
					created_at = excluded.created_at`,
				item.ID, item.TraceID, nullString(item.WorkflowID), nullString(item.ConversationID), nullString(item.CaseID), item.StepType, item.Summary, jsonString(item.EvidenceRefs), jsonString(item.Alternatives), item.Confidence, nullString(item.Decision), item.CreatedAt,
			); err != nil {
				return err
			}
		}
		for _, item := range trace.ToolCalls {
			if _, err := tx.Exec(`insert into tool_call_record (id, trace_id, workflow_id, conversation_id, case_id, tool_name, tool_call_id, request, summary, raw_artifact_refs, approval_state, interpretation_summary, status, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9,$10::jsonb,$11,$12,$13,$14)
				on conflict (id) do update set
					trace_id = excluded.trace_id,
					workflow_id = excluded.workflow_id,
					conversation_id = excluded.conversation_id,
					case_id = excluded.case_id,
					tool_name = excluded.tool_name,
					tool_call_id = excluded.tool_call_id,
					request = excluded.request,
					summary = excluded.summary,
					raw_artifact_refs = excluded.raw_artifact_refs,
					approval_state = excluded.approval_state,
					interpretation_summary = excluded.interpretation_summary,
					status = excluded.status,
					created_at = excluded.created_at`,
				item.ID, item.TraceID, nullString(item.WorkflowID), nullString(item.ConversationID), nullString(item.CaseID), item.ToolName, item.ToolCallID, jsonString(item.Request), nullString(item.Summary), jsonString(item.RawArtifactRefs), nullString(item.ApprovalState), nullString(item.InterpretationSummary), nullString(item.Status), item.CreatedAt,
			); err != nil {
				return err
			}
		}
		for _, item := range trace.SlackActions {
			if _, err := tx.Exec(`insert into slack_action_record (id, trace_id, workflow_id, conversation_id, case_id, channel_id, thread_ts, idempotency_key, draft_body, final_body, policy_verdict, send_status, artifact_refs, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13::jsonb,$14)
				on conflict (id) do update set
					trace_id = excluded.trace_id,
					workflow_id = excluded.workflow_id,
					conversation_id = excluded.conversation_id,
					case_id = excluded.case_id,
					channel_id = excluded.channel_id,
					thread_ts = excluded.thread_ts,
					idempotency_key = excluded.idempotency_key,
					draft_body = excluded.draft_body,
					final_body = excluded.final_body,
					policy_verdict = excluded.policy_verdict,
					send_status = excluded.send_status,
					artifact_refs = excluded.artifact_refs,
					created_at = excluded.created_at`,
				item.ID, item.TraceID, nullString(item.WorkflowID), nullString(item.ConversationID), nullString(item.CaseID), nullString(item.ChannelID), nullString(item.ThreadTS), item.IdempotencyKey, nullString(item.DraftBody), nullString(item.FinalBody), nullString(item.PolicyVerdict), nullString(item.SendStatus), jsonString(item.ArtifactRefs), item.CreatedAt,
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

func persistFeedback(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.feedbackRecords)
	for _, key := range keys {
		for _, item := range store.feedbackRecords[key] {
			if _, err := tx.Exec(`insert into feedback_record (id, conversation_id, case_id, trace_id, target_type, target_id, score, verdict, labels, notes, reviewer_id, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11,$12)`,
				item.ID, nullString(item.ConversationID), nullString(item.CaseID), nullString(item.TraceID), string(item.TargetType), item.TargetID, item.Score, nullString(item.Verdict), jsonString(item.Labels), nullString(item.Notes), item.ReviewerID, item.CreatedAt,
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
		if _, err := tx.Exec(`insert into eval_run (id, trace_id, event_id, suite_name, status, trigger, overall_score, overall_verdict, created_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			on conflict (id) do update set
				trace_id = excluded.trace_id,
				event_id = excluded.event_id,
				suite_name = excluded.suite_name,
				status = excluded.status,
				trigger = excluded.trigger,
				overall_score = excluded.overall_score,
				overall_verdict = excluded.overall_verdict,
				created_at = excluded.created_at,
				completed_at = excluded.completed_at`,
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
			if _, err := tx.Exec(`insert into eval_judgment (id, eval_run_id, layer, category, score, passed, rationale, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8)
				on conflict (id) do update set
					eval_run_id = excluded.eval_run_id,
					layer = excluded.layer,
					category = excluded.category,
					score = excluded.score,
					passed = excluded.passed,
					rationale = excluded.rationale,
					created_at = excluded.created_at`,
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
		item := normalizeCandidateTargetFields(store.candidates[key])
		if _, err := tx.Exec(`insert into improvement_candidate (id, candidate_key, conversation_id, case_id, origin_trace_id, evidence_trace_ids, subsystem, failure_mode, intervention_type, target_layer, target_kind, target_ref, status, severity, recurrence_count, expected_impact, novelty_score, confidence_score, freshness_score, priority_score, risk_tier, hypothesis, proposed_scope, latest_trace_id, source_eval_ids, evidence_artifact_ids, prior_similar_proposal_ids, new_evidence_since_last_rejection, last_evaluated_at, created_at, updated_at) values ($1,$2,$3,$4,$5,$6::jsonb,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25::jsonb,$26::jsonb,$27::jsonb,$28,$29,$30,$31)
			on conflict (id) do update set
				candidate_key = excluded.candidate_key,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				origin_trace_id = excluded.origin_trace_id,
				evidence_trace_ids = excluded.evidence_trace_ids,
				subsystem = excluded.subsystem,
				failure_mode = excluded.failure_mode,
				intervention_type = excluded.intervention_type,
				target_layer = excluded.target_layer,
				target_kind = excluded.target_kind,
				target_ref = excluded.target_ref,
				status = excluded.status,
				severity = excluded.severity,
				recurrence_count = excluded.recurrence_count,
				expected_impact = excluded.expected_impact,
				novelty_score = excluded.novelty_score,
				confidence_score = excluded.confidence_score,
				freshness_score = excluded.freshness_score,
				priority_score = excluded.priority_score,
				risk_tier = excluded.risk_tier,
				hypothesis = excluded.hypothesis,
				proposed_scope = excluded.proposed_scope,
				latest_trace_id = excluded.latest_trace_id,
				source_eval_ids = excluded.source_eval_ids,
				evidence_artifact_ids = excluded.evidence_artifact_ids,
				prior_similar_proposal_ids = excluded.prior_similar_proposal_ids,
				new_evidence_since_last_rejection = excluded.new_evidence_since_last_rejection,
				last_evaluated_at = excluded.last_evaluated_at,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at`,
			item.ID, item.CandidateKey, nullString(item.ConversationID), nullString(item.CaseID), nullString(item.OriginTraceID), jsonString(item.EvidenceTraceIDs), item.Subsystem, item.FailureMode, item.InterventionType, string(item.TargetLayer), nullString(item.TargetKind), nullString(item.TargetRef), string(item.Status), item.Severity, item.RecurrenceCount, item.ExpectedImpact, item.NoveltyScore, item.ConfidenceScore, item.FreshnessScore, item.PriorityScore, string(item.RiskTier), item.Hypothesis, item.ProposedScope, nullString(item.LatestTraceID), jsonString(item.SourceEvalIDs), jsonString(item.EvidenceArtifactIDs), jsonString(item.PriorSimilarProposalIDs), item.NewEvidenceSinceLastRejection, nullTimeValue(item.LastEvaluatedAt), item.CreatedAt, item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistProposals(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.proposals)
	for _, key := range keys {
		item := normalizeProposalTargetFields(store.proposals[key])
		if _, err := tx.Exec(`insert into proposal (id, trace_id, conversation_id, case_id, origin_trace_id, evidence_trace_ids, title, category, summary, status, reviewer, candidate_key, target_layer, target_kind, target_ref, source_eval_ids, risk_tier, proposed_scope, evidence_artifact_ids, active_slot_consuming, review_deadline, prior_similar_proposal_ids, new_evidence_since_last_rejection, created_at) values ($1,$2,$3,$4,$5,$6::jsonb,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16::jsonb,$17,$18,$19::jsonb,$20,$21,$22::jsonb,$23,$24)
			on conflict (id) do update set
				trace_id = excluded.trace_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				origin_trace_id = excluded.origin_trace_id,
				evidence_trace_ids = excluded.evidence_trace_ids,
				title = excluded.title,
				category = excluded.category,
				summary = excluded.summary,
				status = excluded.status,
				reviewer = excluded.reviewer,
				candidate_key = excluded.candidate_key,
				target_layer = excluded.target_layer,
				target_kind = excluded.target_kind,
				target_ref = excluded.target_ref,
				source_eval_ids = excluded.source_eval_ids,
				risk_tier = excluded.risk_tier,
				proposed_scope = excluded.proposed_scope,
				evidence_artifact_ids = excluded.evidence_artifact_ids,
				active_slot_consuming = excluded.active_slot_consuming,
				review_deadline = excluded.review_deadline,
				prior_similar_proposal_ids = excluded.prior_similar_proposal_ids,
				new_evidence_since_last_rejection = excluded.new_evidence_since_last_rejection,
				created_at = excluded.created_at`,
			item.ID, item.TraceID, nullString(item.ConversationID), nullString(item.CaseID), nullString(item.OriginTraceID), jsonString(item.EvidenceTraceIDs), item.Title, item.Category, item.Summary, string(item.Status), nullString(item.Reviewer), item.CandidateKey, string(item.TargetLayer), nullString(item.TargetKind), nullString(item.TargetRef), jsonString(item.SourceEvalIDs), item.RiskTier, item.ProposedScope, jsonString(item.EvidenceArtifactIDs), item.ActiveSlotConsuming, nullTimeValue(item.ReviewDeadline), jsonString(item.PriorSimilarProposalIDs), item.NewEvidenceSinceLastRejection, item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func normalizeCandidateTargetFields(item improvement.Candidate) improvement.Candidate {
	if strings.TrimSpace(string(item.TargetLayer)) == "" {
		item.TargetLayer = deriveCandidateTargetLayer(item)
	}
	if strings.TrimSpace(item.TargetKind) == "" {
		if item.TargetLayer == harness.TargetLayerHarnessOverlay {
			item.TargetKind = "runner_role"
		} else {
			item.TargetKind = "repo"
		}
	}
	if strings.TrimSpace(item.TargetRef) == "" {
		if item.TargetLayer == harness.TargetLayerHarnessOverlay {
			item.TargetRef = "prod"
		} else {
			item.TargetRef = "rsi-agent-platform"
		}
	}
	return item
}

func normalizeProposalTargetFields(item review.Proposal) review.Proposal {
	if strings.TrimSpace(string(item.TargetLayer)) == "" {
		item.TargetLayer = deriveProposalTargetLayer(item)
	}
	if strings.TrimSpace(item.TargetKind) == "" {
		if item.TargetLayer == harness.TargetLayerHarnessOverlay {
			item.TargetKind = "runner_role"
		} else {
			item.TargetKind = "repo"
		}
	}
	if strings.TrimSpace(item.TargetRef) == "" {
		if item.TargetLayer == harness.TargetLayerHarnessOverlay {
			item.TargetRef = "prod"
		} else {
			item.TargetRef = "rsi-agent-platform"
		}
	}
	return item
}

func deriveCandidateTargetLayer(item improvement.Candidate) harness.TargetLayer {
	lowerFailure := strings.ToLower(strings.TrimSpace(item.FailureMode))
	lowerIntervention := strings.ToLower(strings.TrimSpace(item.InterventionType))
	switch {
	case strings.Contains(lowerFailure, "memory"),
		strings.Contains(lowerFailure, "prompt"),
		strings.Contains(lowerFailure, "tool_selection"),
		strings.Contains(lowerFailure, "behavioral"):
		return harness.TargetLayerHarnessOverlay
	case strings.Contains(lowerIntervention, "overlay"),
		strings.Contains(lowerIntervention, "prompt"),
		strings.Contains(lowerIntervention, "behavior"):
		return harness.TargetLayerHarnessOverlay
	default:
		return harness.TargetLayerRepoChange
	}
}

func deriveProposalTargetLayer(item review.Proposal) harness.TargetLayer {
	if strings.TrimSpace(string(item.TargetLayer)) != "" {
		return item.TargetLayer
	}
	switch {
	case strings.Contains(strings.ToLower(strings.TrimSpace(item.CandidateKey)), "memory"),
		strings.Contains(strings.ToLower(strings.TrimSpace(item.CandidateKey)), "prompt"),
		strings.Contains(strings.ToLower(strings.TrimSpace(item.CandidateKey)), "behavioral"),
		strings.Contains(strings.ToLower(strings.TrimSpace(item.Category)), "overlay"):
		return harness.TargetLayerHarnessOverlay
	default:
		return harness.TargetLayerRepoChange
	}
}

func persistProposalReviews(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.proposals)
	for _, key := range keys {
		for _, item := range store.proposals[key].Reviews {
			if _, err := tx.Exec(`insert into proposal_review (id, proposal_id, idempotency_key, decision, rationale, reviewer_id, failure_class, failure_classes, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9)
				on conflict (id) do update set
					proposal_id = excluded.proposal_id,
					idempotency_key = excluded.idempotency_key,
					decision = excluded.decision,
					rationale = excluded.rationale,
					reviewer_id = excluded.reviewer_id,
					failure_class = excluded.failure_class,
					failure_classes = excluded.failure_classes,
					created_at = excluded.created_at`,
				item.ID, item.ProposalID, nullString(item.IdempotencyKey), item.Decision, item.Rationale, item.ReviewerID, nullString(item.FailureClass), jsonString(item.FailureClasses), item.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistProposalMemory(tx *sql.Tx, store *MemoryStore) error {
	for _, item := range store.proposalMemory {
		if _, err := tx.Exec(`insert into proposal_memory (id, review_id, proposal_id, candidate_key, conversation_id, case_id, origin_trace_id, evidence_trace_ids, hypothesis, diff_summary, review_rationale, disposition, disposition_reason, failure_class, failure_classes, source_eval_ids, linked_artifact_ids, linked_proposal_ids, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9,$10,$11,$12,$13,$14,$15::jsonb,$16::jsonb,$17::jsonb,$18::jsonb,$19)
			on conflict (id) do update set
				review_id = excluded.review_id,
				proposal_id = excluded.proposal_id,
				candidate_key = excluded.candidate_key,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				origin_trace_id = excluded.origin_trace_id,
				evidence_trace_ids = excluded.evidence_trace_ids,
				hypothesis = excluded.hypothesis,
				diff_summary = excluded.diff_summary,
				review_rationale = excluded.review_rationale,
				disposition = excluded.disposition,
				disposition_reason = excluded.disposition_reason,
				failure_class = excluded.failure_class,
				failure_classes = excluded.failure_classes,
				source_eval_ids = excluded.source_eval_ids,
				linked_artifact_ids = excluded.linked_artifact_ids,
				linked_proposal_ids = excluded.linked_proposal_ids,
				created_at = excluded.created_at`,
			item.ID, nullInt64(item.ReviewID), item.ProposalID, item.CandidateKey, nullString(item.ConversationID), nullString(item.CaseID), nullString(item.OriginTraceID), jsonString(item.EvidenceTraceIDs), item.Hypothesis, item.DiffSummary, item.ReviewRationale, string(item.Disposition), nullString(item.DispositionReason), nullString(item.FailureClass), jsonString(item.FailureClasses), jsonString(item.SourceEvalIDs), jsonString(item.LinkedArtifactIDs), jsonString(item.LinkedProposalIDs), item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistSettings(tx *sql.Tx, store *MemoryStore) error {
	item := normalizedSettings(store.settings)
	if _, err := tx.Exec(`insert into improvement_settings (key, active_proposal_cap, updated_at) values ($1,$2,$3)
		on conflict (key) do update set
			active_proposal_cap = excluded.active_proposal_cap,
			updated_at = excluded.updated_at`,
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
		if _, err := tx.Exec(`insert into work_item (id, queue, kind, status, trace_id, workflow_id, ingestion_id, conversation_id, case_id, trigger_event_id, proposal_id, thread_key, intent, repo_scope, requested_by, approval_mode, response_mode, payload, attempts, lease_owner, lease_expires_at, last_error, created_at, updated_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18::jsonb,$19,$20,$21,$22,$23,$24,$25)
			on conflict (id) do update set
				queue = excluded.queue,
				kind = excluded.kind,
				status = excluded.status,
				trace_id = excluded.trace_id,
				workflow_id = excluded.workflow_id,
				ingestion_id = excluded.ingestion_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				trigger_event_id = excluded.trigger_event_id,
				proposal_id = excluded.proposal_id,
				thread_key = excluded.thread_key,
				intent = excluded.intent,
				repo_scope = excluded.repo_scope,
				requested_by = excluded.requested_by,
				approval_mode = excluded.approval_mode,
				response_mode = excluded.response_mode,
				payload = excluded.payload,
				attempts = excluded.attempts,
				lease_owner = excluded.lease_owner,
				lease_expires_at = excluded.lease_expires_at,
				last_error = excluded.last_error,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at,
				completed_at = excluded.completed_at`,
			item.ID, string(item.Queue), item.Kind, string(item.Status), nullString(item.TraceID), nullString(item.WorkflowID), nullString(item.IngestionID), nullString(item.ConversationID), nullString(item.CaseID), nullString(item.TriggerEventID), nullString(item.ProposalID), nullString(item.ThreadKey), nullString(item.Intent), nullString(item.RepoScope), nullString(item.RequestedBy), nullString(item.ApprovalMode), nullString(item.ResponseMode), jsonString(item.Payload), item.Attempts, nullString(item.LeaseOwner), nullTime(item.LeaseExpiresAt), nullString(item.LastError), item.CreatedAt, item.UpdatedAt, nullTime(item.CompletedAt),
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
		if _, err := tx.Exec(`insert into repo_change_job (id, proposal_id, conversation_id, case_id, origin_trace_id, candidate_key, status, repo, base_ref, branch_name, allowed_path_globs, context_summary, sandbox_namespace, sandbox_job_name, sandbox_pod_name, validation_error, validation_ref, log_artifact_id, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13,$14,$15,$16,$17,$18,$19,$20)
			on conflict (id) do update set
				proposal_id = excluded.proposal_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				origin_trace_id = excluded.origin_trace_id,
				candidate_key = excluded.candidate_key,
				status = excluded.status,
				repo = excluded.repo,
				base_ref = excluded.base_ref,
				branch_name = excluded.branch_name,
				allowed_path_globs = excluded.allowed_path_globs,
				context_summary = excluded.context_summary,
				sandbox_namespace = excluded.sandbox_namespace,
				sandbox_job_name = excluded.sandbox_job_name,
				sandbox_pod_name = excluded.sandbox_pod_name,
				validation_error = excluded.validation_error,
				validation_ref = excluded.validation_ref,
				log_artifact_id = excluded.log_artifact_id,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at`,
			item.ID, item.ProposalID, nullString(item.ConversationID), nullString(item.CaseID), nullString(item.OriginTraceID), item.CandidateKey, item.Status, item.Repo, item.BaseRef, item.BranchName, jsonString(item.AllowedPathGlobs), item.ContextSummary, nullString(item.SandboxNamespace), nullString(item.SandboxJobName), nullString(item.SandboxPodName), nullString(item.ValidationError), nullString(item.ValidationRef), nullString(item.LogArtifactID), item.CreatedAt, nullTimeValue(item.UpdatedAt),
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
		if _, err := tx.Exec(`insert into pr_attempt (id, proposal_id, conversation_id, case_id, origin_trace_id, repo, branch_name, pr_url, status, validation_status, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			on conflict (id) do update set
				proposal_id = excluded.proposal_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				origin_trace_id = excluded.origin_trace_id,
				repo = excluded.repo,
				branch_name = excluded.branch_name,
				pr_url = excluded.pr_url,
				status = excluded.status,
				validation_status = excluded.validation_status,
				created_at = excluded.created_at`,
			item.ID, item.ProposalID, nullString(item.ConversationID), nullString(item.CaseID), nullString(item.OriginTraceID), item.Repo, item.BranchName, nullString(item.PRURL), item.Status, item.ValidationStatus, item.CreatedAt,
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
		if _, err := tx.Exec(`insert into post_merge_replay (id, proposal_id, trace_id, conversation_id, case_id, baseline_score, candidate_score, improved, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			on conflict (id) do update set
				proposal_id = excluded.proposal_id,
				trace_id = excluded.trace_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				baseline_score = excluded.baseline_score,
				candidate_score = excluded.candidate_score,
				improved = excluded.improved,
				created_at = excluded.created_at`,
			item.ID, item.ProposalID, item.TraceID, nullString(item.ConversationID), nullString(item.CaseID), item.BaselineScore, item.CandidateScore, item.Improved, item.CreatedAt,
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
		if _, err := tx.Exec(`insert into cron_lease (name, holder, expires_at) values ($1,$2,$3)
			on conflict (name) do update set
				holder = excluded.holder,
				expires_at = excluded.expires_at`,
			item.Name, item.Holder, item.ExpiresAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func backfillConversationCaseV2(store *MemoryStore) {
	if len(store.conversations) == 0 {
		for i := range store.ingestions {
			ingestionItem := &store.ingestions[i]
			externalKey := firstNonEmpty(strings.TrimSpace(ingestionItem.ThreadKey), legacyConversationKeyFromIngestion(*ingestionItem))
			conversationID := legacyConversationID(externalKey)
			item, ok := store.conversations[conversationID]
			if !ok {
				item = conversation.Conversation{
					ID:                   conversationID,
					Source:               ingestion.Source(ingestionItem.Source),
					ExternalKey:          externalKey,
					ExternalConversation: externalKey,
					Title:                conversation.NormalizeTitle(ingestionItem.WorkflowHint, ingestionItem.Text),
					Status:               conversation.StatusActive,
					ParticipantIDs:       compactStrings([]string{ingestionItem.UserID}),
					LatestEventID:        ingestionItem.EventID,
					CreatedAt:            ingestionItem.CreatedAt,
					UpdatedAt:            ingestionItem.CreatedAt,
				}
			}
			item.ParticipantIDs = appendUnique(item.ParticipantIDs, ingestionItem.UserID)
			item.LatestEventID = firstNonEmpty(ingestionItem.EventID, item.LatestEventID)
			if ingestionItem.CreatedAt.After(item.UpdatedAt) {
				item.UpdatedAt = ingestionItem.CreatedAt
			}
			store.conversations[conversationID] = item
			ingestionItem.ConversationID = conversationID
		}
	}

	if len(store.cases) == 0 {
		for i := range store.workflows {
			workflowItem := &store.workflows[i]
			conversationID := workflowItem.ConversationID
			if conversationID == "" {
				conversationID = legacyConversationID(workflowItem.ThreadKey)
				workflowItem.ConversationID = conversationID
			}
			caseID := workflowItem.CaseID
			if caseID == "" {
				caseID = legacyCaseID(workflowItem.ID)
				workflowItem.CaseID = caseID
			}
			item := conversation.Case{
				ID:             caseID,
				ConversationID: conversationID,
				Kind:           workflowItem.Kind,
				Intent:         workflowItem.Intent,
				Title:          conversation.NormalizeTitle(workflowItem.Kind, workflowItem.Kind),
				Summary:        workflowItem.Kind,
				Status:         legacyCaseStatus(workflowItem.Status),
				ApprovalMode:   workflowItem.ApprovalMode,
				ResponseMode:   workflowItem.ResponseMode,
				AssignedBot:    workflowItem.AssignedBot,
				LatestTraceID:  workflowItem.TraceID,
				CreatedAt:      workflowItem.CreatedAt,
				UpdatedAt:      workflowItem.UpdatedAt,
				ClosedAt:       workflowItem.CompletedAt,
			}
			if workflowItem.IngestionID != "" {
				if ingestionItem, ok := findLoadedIngestion(store.ingestions, workflowItem.IngestionID); ok {
					item.Summary = ingestionItem.Text
					item.Title = conversation.NormalizeTitle(workflowItem.Kind, ingestionItem.Text)
					item.OpenedByEventID = ingestionItem.EventID
				}
			}
			store.cases[item.ID] = item
			if conversationItem, ok := store.conversations[conversationID]; ok {
				if conversationItem.ActiveCaseID == "" && item.Status == conversation.CaseActive {
					conversationItem.ActiveCaseID = item.ID
				}
				store.conversations[conversationID] = conversationItem
			}
		}
	}

	if len(store.conversationEntries) == 0 {
		for _, ingestionItem := range store.ingestions {
			store.conversationEntries = append(store.conversationEntries, conversation.Entry{
				ID:             legacyEntryID(ingestionItem.ID),
				ConversationID: ingestionItem.ConversationID,
				EventID:        ingestionItem.EventID,
				Source:         ingestion.Source(ingestionItem.Source),
				SourceEventID:  firstNonEmpty(ingestionItem.EventID, ingestionItem.ID),
				EntryType:      "external_event",
				ActorID:        ingestionItem.UserID,
				ActorType:      "user",
				Body:           ingestionItem.Text,
				Metadata: map[string]interface{}{
					"channel_id": ingestionItem.ChannelID,
					"thread_ts":  ingestionItem.ThreadTS,
				},
				CreatedAt: ingestionItem.CreatedAt,
			})
		}
	}

	for i := range store.ingestions {
		if store.ingestions[i].CaseID == "" {
			for _, workflowItem := range store.workflows {
				if workflowItem.IngestionID == store.ingestions[i].ID {
					store.ingestions[i].CaseID = workflowItem.CaseID
					store.ingestions[i].ConversationID = workflowItem.ConversationID
					break
				}
			}
		}
	}

	for traceID, trace := range store.traces {
		if trace.Summary.ConversationID == "" || trace.Summary.CaseID == "" {
			for _, workflowItem := range store.workflows {
				if workflowItem.ID != trace.Summary.WorkflowID {
					continue
				}
				trace.Summary.ConversationID = firstNonEmpty(trace.Summary.ConversationID, workflowItem.ConversationID)
				trace.Summary.CaseID = firstNonEmpty(trace.Summary.CaseID, workflowItem.CaseID)
				break
			}
		}
		if trace.Summary.TriggerEventID == "" {
			for _, ingestionItem := range store.ingestions {
				if ingestionItem.ID == trace.Summary.IngestionID {
					trace.Summary.TriggerEventID = ingestionItem.EventID
					break
				}
			}
		}
		for i := range trace.Events {
			trace.Events[i].ConversationID = firstNonEmpty(trace.Events[i].ConversationID, trace.Summary.ConversationID)
			trace.Events[i].CaseID = firstNonEmpty(trace.Events[i].CaseID, trace.Summary.CaseID)
			trace.Events[i].TriggerEventID = firstNonEmpty(trace.Events[i].TriggerEventID, trace.Summary.TriggerEventID)
		}
		for i := range trace.Reasoning {
			trace.Reasoning[i].ConversationID = firstNonEmpty(trace.Reasoning[i].ConversationID, trace.Summary.ConversationID)
			trace.Reasoning[i].CaseID = firstNonEmpty(trace.Reasoning[i].CaseID, trace.Summary.CaseID)
		}
		for i := range trace.ToolCalls {
			trace.ToolCalls[i].ConversationID = firstNonEmpty(trace.ToolCalls[i].ConversationID, trace.Summary.ConversationID)
			trace.ToolCalls[i].CaseID = firstNonEmpty(trace.ToolCalls[i].CaseID, trace.Summary.CaseID)
		}
		for i := range trace.SlackActions {
			trace.SlackActions[i].ConversationID = firstNonEmpty(trace.SlackActions[i].ConversationID, trace.Summary.ConversationID)
			trace.SlackActions[i].CaseID = firstNonEmpty(trace.SlackActions[i].CaseID, trace.Summary.CaseID)
		}
		store.traces[traceID] = trace
	}

	for workID, item := range store.workItems {
		if item.ConversationID == "" || item.CaseID == "" || item.TriggerEventID == "" {
			if trace, ok := store.traces[item.TraceID]; ok {
				item.ConversationID = firstNonEmpty(item.ConversationID, trace.Summary.ConversationID)
				item.CaseID = firstNonEmpty(item.CaseID, trace.Summary.CaseID)
				item.TriggerEventID = firstNonEmpty(item.TriggerEventID, trace.Summary.TriggerEventID)
			}
			store.workItems[workID] = item
		}
	}

	for key, candidate := range store.candidates {
		if trace, ok := store.traces[candidate.LatestTraceID]; ok {
			candidate.ConversationID = firstNonEmpty(candidate.ConversationID, trace.Summary.ConversationID)
			candidate.CaseID = firstNonEmpty(candidate.CaseID, trace.Summary.CaseID)
			candidate.OriginTraceID = firstNonEmpty(candidate.OriginTraceID, trace.Summary.TraceID)
			if len(candidate.EvidenceTraceIDs) == 0 {
				candidate.EvidenceTraceIDs = []string{trace.Summary.TraceID}
			}
			store.candidates[key] = candidate
		}
	}
	for id, proposal := range store.proposals {
		traceID := firstNonEmpty(proposal.OriginTraceID, proposal.TraceID)
		if trace, ok := store.traces[traceID]; ok {
			proposal.ConversationID = firstNonEmpty(proposal.ConversationID, trace.Summary.ConversationID)
			proposal.CaseID = firstNonEmpty(proposal.CaseID, trace.Summary.CaseID)
			proposal.OriginTraceID = firstNonEmpty(proposal.OriginTraceID, trace.Summary.TraceID)
			if len(proposal.EvidenceTraceIDs) == 0 {
				proposal.EvidenceTraceIDs = []string{trace.Summary.TraceID}
			}
			store.proposals[id] = proposal
		}
	}
	if len(store.feedbackRecords) == 0 {
		for traceID, ratings := range store.ratings {
			trace := store.traces[traceID]
			for idx, rating := range ratings {
				store.feedbackRecords[traceID] = append(store.feedbackRecords[traceID], review.FeedbackRecord{
					ID:             fmt.Sprintf("feedback-rating-%s-%d", traceID, idx+1),
					ConversationID: trace.Summary.ConversationID,
					CaseID:         trace.Summary.CaseID,
					TraceID:        traceID,
					TargetType:     review.FeedbackTargetTrace,
					TargetID:       traceID,
					Score:          rating.Score,
					Verdict:        rating.Verdict,
					Labels:         append([]string(nil), rating.Labels...),
					Notes:          rating.Notes,
					ReviewerID:     rating.ReviewerID,
					CreatedAt:      rating.CreatedAt,
				})
			}
		}
		for traceID, notes := range store.notes {
			trace := store.traces[traceID]
			for idx, note := range notes {
				store.feedbackRecords[traceID] = append(store.feedbackRecords[traceID], review.FeedbackRecord{
					ID:             fmt.Sprintf("feedback-note-%s-%d", traceID, idx+1),
					ConversationID: trace.Summary.ConversationID,
					CaseID:         trace.Summary.CaseID,
					TraceID:        traceID,
					TargetType:     review.FeedbackTargetTrace,
					TargetID:       traceID,
					Verdict:        note.Category,
					Labels:         []string{"improvement_note"},
					Notes:          note.Note,
					ReviewerID:     note.CreatedBy,
					CreatedAt:      note.CreatedAt,
				})
			}
		}
	}
}

func backfillActionOutcomeKnowledgeV3(store *MemoryStore) {
	for key, caseRecord := range store.cases {
		if caseRecord.ResolutionState == "" {
			caseRecord.ResolutionState = conversation.ResolutionUnresolved
		}
		store.cases[key] = caseRecord
	}

	if len(store.actionIntents) == 0 {
		for _, trace := range store.traces {
			for _, item := range trace.SlackActions {
				intentID := fmt.Sprintf("backfill-action-slack-%s", item.ID)
				store.actionIntents[intentID] = action.Intent{
					ID:             intentID,
					OwnerPlane:     "control",
					ConversationID: firstNonEmpty(item.ConversationID, trace.Summary.ConversationID),
					CaseID:         firstNonEmpty(item.CaseID, trace.Summary.CaseID),
					TraceID:        trace.Summary.TraceID,
					Kind:           action.KindSlackPost,
					TargetRef:      firstNonEmpty(item.ChannelID, trace.Summary.ThreadKey),
					RequestPayload: map[string]any{
						"channel_id": item.ChannelID,
						"thread_ts":  item.ThreadTS,
						"body":       firstNonEmpty(item.FinalBody, item.DraftBody),
					},
					IdempotencyKey: item.IdempotencyKey,
					ApprovalState:  item.PolicyVerdict,
					PolicyVerdict:  item.PolicyVerdict,
					Status:         backfilledActionStatus(item.SendStatus),
					RequestedBy:    "backfill",
					Rationale:      "Backfilled from existing slack_action_record.",
					EvidenceRefs: []events.EvidenceRef{
						{Kind: "slack_action", Ref: item.ID, Summary: firstNonEmpty(item.FinalBody, item.DraftBody)},
					},
					CreatedAt: item.CreatedAt,
					UpdatedAt: item.CreatedAt,
				}
				store.actionResults[intentID] = []action.Result{
					{
						ID:             fmt.Sprintf("backfill-action-result-%s", item.ID),
						ActionIntentID: intentID,
						AttemptNumber:  1,
						Executor:       "backfill",
						Provider:       "slack",
						ProviderRef:    firstNonEmpty(item.ThreadTS, item.ChannelID),
						Status:         backfilledActionStatus(item.SendStatus),
						StartedAt:      item.CreatedAt,
						CompletedAt:    item.CreatedAt,
					},
				}
			}
		}
		for _, proposal := range store.proposals {
			for _, job := range store.repoChangeJobs {
				if job.ProposalID != proposal.ID {
					continue
				}
				intentID := fmt.Sprintf("backfill-action-sandbox-%s", job.ID)
				if _, ok := store.actionIntents[intentID]; !ok {
					store.actionIntents[intentID] = action.Intent{
						ID:             intentID,
						OwnerPlane:     "improvement",
						ConversationID: job.ConversationID,
						CaseID:         job.CaseID,
						TraceID:        firstNonEmpty(job.OriginTraceID, proposal.TraceID),
						ProposalID:     proposal.ID,
						Kind:           action.KindSandboxLaunch,
						TargetRef:      job.Repo,
						RequestPayload: map[string]any{"branch_name": job.BranchName, "base_ref": job.BaseRef},
						Status:         statusForProposalJob(job.Status),
						RequestedBy:    "backfill",
						Rationale:      "Backfilled from existing repo_change_job.",
						CreatedAt:      job.CreatedAt,
						UpdatedAt:      job.CreatedAt,
					}
				}
			}
			for _, attempt := range store.prAttempts {
				if attempt.ProposalID != proposal.ID {
					continue
				}
				intentID := fmt.Sprintf("backfill-action-pr-%s", attempt.ID)
				store.actionIntents[intentID] = action.Intent{
					ID:             intentID,
					OwnerPlane:     "improvement",
					ConversationID: attempt.ConversationID,
					CaseID:         attempt.CaseID,
					TraceID:        firstNonEmpty(attempt.OriginTraceID, proposal.TraceID),
					ProposalID:     proposal.ID,
					Kind:           action.KindDraftPROpen,
					TargetRef:      attempt.Repo,
					RequestPayload: map[string]any{"branch_name": attempt.BranchName},
					Status:         action.StatusSucceeded,
					RequestedBy:    "backfill",
					Rationale:      "Backfilled from existing pr_attempt.",
					CreatedAt:      attempt.CreatedAt,
					UpdatedAt:      attempt.CreatedAt,
				}
				store.actionResults[intentID] = []action.Result{
					{
						ID:             fmt.Sprintf("backfill-action-result-pr-%s", attempt.ID),
						ActionIntentID: intentID,
						AttemptNumber:  1,
						Executor:       "backfill",
						Provider:       "github",
						ProviderRef:    attempt.PRURL,
						Status:         action.StatusSucceeded,
						StartedAt:      attempt.CreatedAt,
						CompletedAt:    attempt.CreatedAt,
					},
				}
			}
		}
	}
}

func backfilledActionStatus(sendStatus string) action.Status {
	switch strings.ToLower(strings.TrimSpace(sendStatus)) {
	case "posted", "ok", "succeeded":
		return action.StatusSucceeded
	case "blocked_by_policy", "blocked":
		return action.StatusBlocked
	case "canceled":
		return action.StatusCanceled
	default:
		return action.StatusSucceeded
	}
}

func statusForProposalJob(status string) action.Status {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case string(review.ProposalRepoChangeQueued):
		return action.StatusQueued
	case string(review.ProposalRepoChangeRunning), string(review.ProposalValidationPending):
		return action.StatusExecuting
	case string(review.ProposalPROpen), string(review.ProposalMerged):
		return action.StatusSucceeded
	case string(review.ProposalFailedValidation):
		return action.StatusFailed
	default:
		return action.StatusApproved
	}
}

func legacyConversationKeyFromIngestion(item slack.Ingestion) string {
	if item.ThreadKey != "" {
		return item.ThreadKey
	}
	if strings.HasPrefix(item.ChannelID, "D") {
		return fmt.Sprintf("slack:dm:%s", item.ChannelID)
	}
	return fmt.Sprintf("%s:%s", item.Source, item.ID)
}

func legacyConversationID(externalKey string) string {
	return fmt.Sprintf("conv-%s", sanitizeIDComponent(externalKey))
}

func legacyCaseID(workflowID string) string {
	return fmt.Sprintf("case-%s", sanitizeIDComponent(workflowID))
}

func legacyEntryID(ingestionID string) string {
	return fmt.Sprintf("entry-%s", sanitizeIDComponent(ingestionID))
}

func sanitizeIDComponent(value string) string {
	replacer := strings.NewReplacer(":", "-", "/", "-", ".", "-", " ", "-", "#", "-")
	value = replacer.Replace(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}
	return value
}

func legacyCaseStatus(status string) conversation.CaseStatus {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed":
		return conversation.CaseResolved
	case "failed":
		return conversation.CaseClosed
	default:
		return conversation.CaseActive
	}
}

func findLoadedIngestion(items []slack.Ingestion, id string) (slack.Ingestion, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return slack.Ingestion{}, false
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

func nullInt64(value int64) any {
	if value == 0 {
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

func (p *PostgresStore) ListConversations() []conversation.Conversation {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListConversations()
}

func (p *PostgresStore) GetConversation(conversationID string) (conversation.Conversation, bool) {
	store, err := p.readStore()
	if err != nil {
		return conversation.Conversation{}, false
	}
	return store.GetConversation(conversationID)
}

func (p *PostgresStore) ListConversationEntries(conversationID string) []conversation.Entry {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListConversationEntries(conversationID)
}

func (p *PostgresStore) ListCases() []conversation.Case {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListCases()
}

func (p *PostgresStore) GetCase(caseID string) (conversation.Case, bool) {
	store, err := p.readStore()
	if err != nil {
		return conversation.Case{}, false
	}
	return store.GetCase(caseID)
}

func (p *PostgresStore) ListActionIntents() []action.Intent {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListActionIntents()
}

func (p *PostgresStore) GetActionIntent(actionID string) (action.Intent, bool) {
	store, err := p.readStore()
	if err != nil {
		return action.Intent{}, false
	}
	return store.GetActionIntent(actionID)
}

func (p *PostgresStore) UpsertActionIntent(intent action.Intent) (item action.Intent, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.UpsertActionIntent(intent)
		if err != nil {
			return err
		}
		return replaceActionIntentScope(tx, item)
	})
	return
}

func (p *PostgresStore) ListActionResults(actionIntentID string) []action.Result {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListActionResults(actionIntentID)
}

func (p *PostgresStore) RecordActionResult(result action.Result) (item action.Result, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.RecordActionResult(result)
		if err != nil {
			return err
		}
		// Keep action_result persistence on plain insert semantics so RSI can
		// still reproduce and self-repair the original primary-key collision.
		if err := insertActionResult(tx, item); err != nil {
			return err
		}
		intent, ok := store.actionIntents[item.ActionIntentID]
		if !ok {
			return nil
		}
		return replaceActionIntentScope(tx, intent)
	})
	return
}

func (p *PostgresStore) ListOutcomes() []outcome.Record {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListOutcomes()
}

func (p *PostgresStore) RecordOutcome(record outcome.Record) (item outcome.Record, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.RecordOutcome(record)
		if err != nil {
			return err
		}
		if err := replaceOutcomeScope(tx, item); err != nil {
			return err
		}
		if err := replaceCaseScope(tx, item, store); err != nil {
			return err
		}
		if item.ProposalID != "" {
			return replaceProposalScope(tx, store, item.ProposalID)
		}
		return nil
	})
	return
}

func (p *PostgresStore) ListKnowledgeEntries() []knowledge.Entry {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListKnowledgeEntries()
}

func (p *PostgresStore) GetKnowledgeEntry(knowledgeID string) (knowledge.Entry, bool) {
	store, err := p.readStore()
	if err != nil {
		return knowledge.Entry{}, false
	}
	return store.GetKnowledgeEntry(knowledgeID)
}

func (p *PostgresStore) UpsertKnowledgeEntry(entry knowledge.Entry, links []knowledge.EvidenceLink) (item knowledge.Entry, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.UpsertKnowledgeEntry(entry, links)
		if err != nil {
			return err
		}
		if err := replaceKnowledgeEntryScope(tx, item); err != nil {
			return err
		}
		return replaceKnowledgeEvidenceScope(tx, store, item.ID)
	})
	return
}

func (p *PostgresStore) ListKnowledgeEvidenceLinks(knowledgeID string) []knowledge.EvidenceLink {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListKnowledgeEvidenceLinks(knowledgeID)
}

func (p *PostgresStore) ListKnowledgeReviews(knowledgeID string) []knowledge.Review {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListKnowledgeReviews(knowledgeID)
}

func (p *PostgresStore) ReviewKnowledgeEntry(knowledgeID string, item knowledge.Review) (entry knowledge.Entry, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		entry, err = store.ReviewKnowledgeEntry(knowledgeID, item)
		if err != nil {
			return err
		}
		if err := replaceKnowledgeEntryScope(tx, entry); err != nil {
			return err
		}
		if err := replaceKnowledgeEvidenceScope(tx, store, knowledgeID); err != nil {
			return err
		}
		return replaceKnowledgeReviewScope(tx, store, knowledgeID)
	})
	return
}

func (p *PostgresStore) CreateEvent(event ingestion.EventEnvelope) (created ingestion.EventEnvelope, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		var createErr error
		created, createErr = store.CreateEvent(event)
		if createErr != nil {
			return createErr
		}
		return replaceEventMaterializationScope(tx, store, created)
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

func (p *PostgresStore) CreateIngestion(envelope slack.SlackEnvelope) (slack.Ingestion, error) {
	return p.createIngestionDirect(envelope)
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
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		var inner error
		item, inner = store.SetThreadState(threadKey, state, owner)
		if inner != nil {
			return inner
		}
		return replaceThreadPolicyScope(tx, item)
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

func (p *PostgresStore) ListFeedback(traceID string) []review.FeedbackRecord {
	store, err := p.readStore()
	if err != nil {
		return []review.FeedbackRecord{}
	}
	return store.ListFeedback(traceID)
}

func (p *PostgresStore) AddFeedback(record review.FeedbackRecord) (item review.FeedbackRecord, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.AddFeedback(record)
		if err != nil {
			return err
		}
		return replaceFeedbackScope(tx, store, item.TraceID)
	})
	return
}

func (p *PostgresStore) AddRating(traceID string, rating review.HumanRating) (item review.HumanRating, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.AddRating(traceID, rating)
		if err != nil {
			return err
		}
		return replaceRatingScope(tx, store, traceID)
	})
	return
}

func (p *PostgresStore) AddImprovementNote(traceID string, note review.ImprovementNote) (item review.ImprovementNote, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.AddImprovementNote(traceID, note)
		if err != nil {
			return err
		}
		return replaceImprovementNotesScope(tx, store, traceID)
	})
	return
}

func (p *PostgresStore) ScheduleReplay(traceID string, requestedBy string) (item queue.WorkItem, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		var inner error
		item, inner = store.ScheduleReplay(traceID, requestedBy)
		if inner != nil {
			return inner
		}
		if err := replaceWorkItemScope(tx, item); err != nil {
			return err
		}
		return replaceTraceScope(tx, store, traceID)
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
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		run, judgments, err = store.EvaluateTrace(traceID, trigger)
		if err != nil {
			return err
		}
		if err := replaceEvalRunScope(tx, run, judgments); err != nil {
			return err
		}
		return replaceAllCandidates(tx, store)
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
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.UpdateSettings(settings)
		if err != nil {
			return err
		}
		return replaceSettingsScope(tx, item)
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
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		if item.ID == "" {
			item.ID = nextID("work", 0)
		}
		if item.Status == "" {
			item.Status = queue.WorkQueued
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = now
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = item.CreatedAt
		}
		if item.Payload == nil {
			item.Payload = map[string]interface{}{}
		}
		existing, ok, findErr := findExistingWorkItemByDedupe(tx, item)
		if findErr != nil {
			return findErr
		}
		if ok {
			created = existing
			return nil
		}
		if err := replaceWorkItemScope(tx, item); err != nil {
			return err
		}
		created = item
		return nil
	})
	return
}

func (p *PostgresStore) RescheduleWorkItem(id string, payload map[string]interface{}, lastError string, availableAt time.Time) (queue.WorkItem, error) {
	return p.rescheduleWorkItemDirect(id, payload, lastError, availableAt)
}

func (p *PostgresStore) ClaimNextWorkItem(queues []queue.QueueName, holder string, lease time.Duration) (item queue.WorkItem, ok bool, err error) {
	if holder == "" {
		return queue.WorkItem{}, false, fmt.Errorf("holder is required")
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		expires := now.Add(lease)
		queueClause, queueArgs := queuePredicate(queues, 1)
		statusQueuedArg := len(queueArgs) + 1
		statusLeasedArg := len(queueArgs) + 2
		nowArg := len(queueArgs) + 3
		nowUnixArg := len(queueArgs) + 4
		leasedStatusArg := len(queueArgs) + 5
		holderArg := len(queueArgs) + 6
		expiresArg := len(queueArgs) + 7
		updatedArg := len(queueArgs) + 8
		args := append(queueArgs,
			string(queue.WorkQueued),
			string(queue.WorkLeased),
			now,
			now.Unix(),
			string(queue.WorkLeased),
			holder,
			expires,
			now,
		)
		query := fmt.Sprintf(`
with next_item as (
	select id
	from work_item
	where %s
	  and coalesce(nullif(payload->>'retry_after_unix', ''), '0')::bigint <= $%d
	  and (status = $%d or (status = $%d and lease_expires_at is not null and lease_expires_at < $%d))
	order by created_at asc, id asc
	for update skip locked
	limit 1
)
update work_item wi
set status = $%d,
	attempts = wi.attempts + 1,
	lease_owner = $%d,
	lease_expires_at = $%d,
	updated_at = $%d,
	completed_at = null
from next_item
where wi.id = next_item.id
returning %s`, queueClause, nowUnixArg, statusQueuedArg, statusLeasedArg, nowArg, leasedStatusArg, holderArg, expiresArg, updatedArg, workItemSelectColumnsWithAlias("wi"))
		row := tx.QueryRow(query, args...)
		var scanErr error
		item, scanErr = scanWorkItem(row)
		if scanErr == sql.ErrNoRows {
			ok = false
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		ok = true
		return nil
	})
	return
}

func (p *PostgresStore) CompleteWorkItem(id string) (item queue.WorkItem, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		row := tx.QueryRow(
			`update work_item set status = $2, lease_owner = null, lease_expires_at = null, updated_at = $3, completed_at = $3 where id = $1 returning `+workItemSelectColumns(),
			id,
			string(queue.WorkCompleted),
			now,
		)
		item, err = scanWorkItem(row)
		return err
	})
	return
}

func (p *PostgresStore) FailWorkItem(id string, lastError string) (item queue.WorkItem, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		now := time.Now().UTC()
		row := tx.QueryRow(
			`update work_item set status = $2, lease_owner = null, lease_expires_at = null, last_error = $3, updated_at = $4, completed_at = $4 where id = $1 returning `+workItemSelectColumns(),
			id,
			string(queue.WorkFailed),
			lastError,
			now,
		)
		item, err = scanWorkItem(row)
		return err
	})
	return
}

func (p *PostgresStore) UpdateWorkflowStatus(workflowID string, status string, lastError string) (item Workflow, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		item, err = store.UpdateWorkflowStatus(workflowID, status, lastError)
		if err != nil {
			return err
		}
		return replaceWorkflowScope(tx, item)
	})
	return
}

func (p *PostgresStore) ApplyTraceUpdate(traceID string, update TraceUpdate) (trace events.Trace, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		trace, err = store.ApplyTraceUpdate(traceID, update)
		if err != nil {
			return err
		}
		if err := replaceTraceAndWorkflowScope(tx, store, trace); err != nil {
			return err
		}
		if trace.Summary.CaseID != "" {
			if caseItem, ok := store.cases[trace.Summary.CaseID]; ok {
				temp := newSubsetStore()
				temp.cases[caseItem.ID] = caseItem
				if err := persistCases(tx, temp); err != nil {
					return err
				}
			}
		}
		return nil
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
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		result, err = store.PromoteCandidates(requestedBy, limit)
		if err != nil {
			return err
		}
		if err := replaceAllCandidates(tx, store); err != nil {
			return err
		}
		return replaceAllProposals(tx, store)
	})
	return
}

func (p *PostgresStore) RunProposalPromoter(holder string) (result PromotionResult, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		result, err = store.RunProposalPromoter(holder)
		if err != nil {
			return err
		}
		return replaceProposalPromoterScope(tx, store)
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

func (p *PostgresStore) ReviewProposal(proposalID string, decision review.ProposalReview) (review.Proposal, error) {
	return p.reviewProposalDirect(proposalID, decision)
}

func (p *PostgresStore) UpdateProposalStatus(proposalID string, status review.ProposalStatus) (proposal review.Proposal, err error) {
	err = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		proposal, err = store.UpdateProposalStatus(proposalID, status)
		if err != nil {
			return err
		}
		return replaceProposalScope(tx, store, proposalID)
	})
	return
}

func (p *PostgresStore) MaterializeApprovedProposal(proposalID string, requestedBy string) (improvement.RepoChangeJob, error) {
	return p.materializeApprovedProposalDirect(proposalID, requestedBy)
}

func (p *PostgresStore) RetryProposalRepoChange(proposalID string, requestedBy string) (queue.WorkItem, error) {
	return p.retryProposalRepoChangeDirect(proposalID, requestedBy)
}

func (p *PostgresStore) ListRepoChangeJobs() []improvement.RepoChangeJob {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListRepoChangeJobs()
}

func (p *PostgresStore) UpdateRepoChangeJobStatus(jobID string, status string) (item improvement.RepoChangeJob, err error) {
	return p.updateRepoChangeJobStatusDirect(jobID, status)
}

func (p *PostgresStore) UpsertRepoChangeJob(job improvement.RepoChangeJob) (improvement.RepoChangeJob, error) {
	return p.upsertRepoChangeJobDirect(job)
}

func (p *PostgresStore) ListPRAttempts() []improvement.PRAttempt {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListPRAttempts()
}

func (p *PostgresStore) RecordPRAttempt(attempt improvement.PRAttempt) (item improvement.PRAttempt, err error) {
	return p.recordPRAttemptDirect(attempt)
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
	_ = p.withLoadedStoreTx(func(tx *sql.Tx, store *MemoryStore) error {
		beforeAttempts := len(store.prAttempts)
		result = store.ExecuteTool(name, input)
		if name == "github.create_pr" {
			if proposalID, _ := input["proposal_id"].(string); proposalID != "" {
				if err := replaceProposalScope(tx, store, proposalID); err != nil {
					return err
				}
			}
			if len(store.prAttempts) > beforeAttempts {
				temp := newSubsetStore()
				temp.prAttempts = store.prAttempts
				if err := persistPRAttempts(tx, temp); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return result
}
