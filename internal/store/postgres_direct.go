package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func (p *PostgresStore) withTx(fn func(tx *sql.Tx) error) error {
	return p.withTxContext(context.Background(), fn)
}

func (p *PostgresStore) withTxContext(ctx context.Context, fn func(tx *sql.Tx) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *PostgresStore) withProposalLockedStoreTx(proposalID string, fn func(tx *sql.Tx, store *MemoryStore) error) error {
	return p.withTx(func(tx *sql.Tx) error {
		if err := lockProposal(tx, proposalID); err != nil {
			return err
		}
		store, err := loadStore(tx)
		if err != nil {
			return err
		}
		return fn(tx, store)
	})
}

func advisoryLock(tx *sql.Tx, key string) error {
	return advisoryLockContext(context.Background(), tx, key)
}

func advisoryLockContext(ctx context.Context, tx *sql.Tx, key string) error {
	_, err := tx.ExecContext(ctx, `select pg_advisory_xact_lock(hashtext($1)::bigint)`, key)
	return err
}

func lockProposal(tx *sql.Tx, proposalID string) error {
	if err := advisoryLock(tx, fmt.Sprintf("proposal:%s", proposalID)); err != nil {
		return err
	}
	var id string
	return tx.QueryRow(`select id from proposal where id = $1 for update`, proposalID).Scan(&id)
}

func newSubsetStore() *MemoryStore {
	return newEmptyMemoryStore()
}

func replaceTraceScope(tx *sql.Tx, store *MemoryStore, traceID string) error {
	trace, ok := store.traces[traceID]
	if !ok {
		return fmt.Errorf("trace %s not found", traceID)
	}
	for _, stmt := range []string{
		`delete from trace_event where trace_id = $1`,
		`delete from artifact where trace_id = $1`,
		`delete from reasoning_step where trace_id = $1`,
		`delete from tool_call_record where trace_id = $1`,
		`delete from slack_action_record where trace_id = $1`,
		`delete from conversation_entry where trace_id = $1 and entry_type = 'slack_action'`,
	} {
		if _, err := tx.Exec(stmt, traceID); err != nil {
			return err
		}
	}
	temp := newSubsetStore()
	temp.traces[traceID] = trace
	if err := persistTraces(tx, temp); err != nil {
		return err
	}
	for _, item := range store.conversationEntries {
		if item.TraceID == traceID && item.EntryType == "slack_action" {
			temp.conversationEntries = append(temp.conversationEntries, item)
		}
	}
	if len(temp.conversationEntries) > 0 {
		if err := persistConversationEntries(tx, temp); err != nil {
			return err
		}
	}
	return nil
}

func replaceKnowledgeEvidenceScope(tx *sql.Tx, store *MemoryStore, knowledgeID string) error {
	if _, err := tx.Exec(`delete from knowledge_evidence_link where knowledge_entry_id = $1`, knowledgeID); err != nil {
		return err
	}
	temp := newSubsetStore()
	if links := store.knowledgeEvidence[knowledgeID]; len(links) > 0 {
		temp.knowledgeEvidence[knowledgeID] = append([]knowledge.EvidenceLink(nil), links...)
	}
	if len(temp.knowledgeEvidence) == 0 {
		return nil
	}
	return persistKnowledgeEvidence(tx, temp)
}

func replaceKnowledgeReviewScope(tx *sql.Tx, store *MemoryStore, knowledgeID string) error {
	if _, err := tx.Exec(`delete from knowledge_review where knowledge_entry_id = $1`, knowledgeID); err != nil {
		return err
	}
	temp := newSubsetStore()
	if reviews := store.knowledgeReviews[knowledgeID]; len(reviews) > 0 {
		temp.knowledgeReviews[knowledgeID] = append([]knowledge.Review(nil), reviews...)
	}
	if len(temp.knowledgeReviews) == 0 {
		return nil
	}
	return persistKnowledgeReviews(tx, temp)
}

func replaceProposalReviewScope(tx *sql.Tx, store *MemoryStore, proposalID string) error {
	if _, err := tx.Exec(`delete from proposal_review where proposal_id = $1`, proposalID); err != nil {
		return err
	}
	temp := newSubsetStore()
	if proposal, ok := store.proposals[proposalID]; ok {
		temp.proposals[proposalID] = proposal
	}
	return persistProposalReviews(tx, temp)
}

func replaceThreadPolicyScope(tx *sql.Tx, item policy.ThreadPolicy) error {
	temp := newSubsetStore()
	temp.threadPolicies[item.ThreadKey] = item
	return persistThreadPolicies(tx, temp)
}

func replaceActionIntentScope(tx *sql.Tx, item action.Intent) error {
	temp := newSubsetStore()
	temp.actionIntents[item.ID] = item
	return persistActionIntents(tx, temp)
}

func replaceActionResultScope(tx *sql.Tx, store *MemoryStore, actionIntentID string) error {
	temp := newSubsetStore()
	temp.actionResults[actionIntentID] = append([]action.Result(nil), store.actionResults[actionIntentID]...)
	return persistActionResults(tx, temp)
}

func replaceWorkflowScope(tx *sql.Tx, item Workflow) error {
	temp := newSubsetStore()
	temp.workflows = append(temp.workflows, item)
	return persistWorkflows(tx, temp)
}

func replaceCaseScope(tx *sql.Tx, item outcome.Record, store *MemoryStore) error {
	if item.CaseID == "" {
		return nil
	}
	caseRecord, ok := store.cases[item.CaseID]
	if !ok {
		return nil
	}
	temp := newSubsetStore()
	temp.cases[caseRecord.ID] = caseRecord
	return persistCases(tx, temp)
}

func replaceProposalScope(tx *sql.Tx, store *MemoryStore, proposalID string) error {
	proposal, ok := store.proposals[proposalID]
	if !ok {
		return nil
	}
	temp := newSubsetStore()
	temp.proposals[proposal.ID] = proposal
	return persistProposals(tx, temp)
}

func replaceCandidateScope(tx *sql.Tx, store *MemoryStore, candidateKey string) error {
	candidate, ok := store.candidates[candidateKey]
	if !ok {
		return nil
	}
	temp := newSubsetStore()
	temp.candidates[candidateKey] = candidate
	return persistCandidates(tx, temp)
}

func replaceRuntimeDiagnosisScope(tx *sql.Tx, store *MemoryStore, diagnosisID string) error {
	diagnosis, ok := store.runtimeDiagnoses[diagnosisID]
	if !ok {
		return nil
	}
	temp := newSubsetStore()
	temp.runtimeDiagnoses[diagnosisID] = diagnosis
	return persistRuntimeDiagnoses(tx, temp)
}

func replaceQuestionRunScope(tx *sql.Tx, store *MemoryStore, questionRunID string) error {
	item, ok := store.questionRuns[questionRunID]
	if !ok {
		return nil
	}
	temp := newSubsetStore()
	temp.questionRuns[questionRunID] = item
	return persistQuestionRuns(tx, temp)
}

func replaceChangeAttemptScope(tx *sql.Tx, item improvement.ChangeAttempt) error {
	temp := newSubsetStore()
	temp.changeAttempts[item.ID] = item
	return persistChangeAttempts(tx, temp)
}

func replaceAllCandidates(tx *sql.Tx, store *MemoryStore) error {
	temp := newSubsetStore()
	temp.candidates = store.candidates
	return persistCandidates(tx, temp)
}

func replaceAllProposals(tx *sql.Tx, store *MemoryStore) error {
	temp := newSubsetStore()
	temp.proposals = store.proposals
	return persistProposals(tx, temp)
}

func replaceProposalMemoryScope(tx *sql.Tx, store *MemoryStore, proposalID string) error {
	if _, err := tx.Exec(`delete from proposal_memory where proposal_id = $1`, proposalID); err != nil {
		return err
	}
	temp := newSubsetStore()
	for _, item := range store.proposalMemory {
		if item.ProposalID == proposalID {
			temp.proposalMemory = append(temp.proposalMemory, item)
		}
	}
	if len(temp.proposalMemory) == 0 {
		return nil
	}
	return persistProposalMemory(tx, temp)
}

func replaceRepoChangeJobScope(tx *sql.Tx, store *MemoryStore, proposalID string) error {
	temp := newSubsetStore()
	for key, item := range store.repoChangeJobs {
		if item.ProposalID == proposalID {
			temp.repoChangeJobs[key] = item
		}
	}
	if len(temp.repoChangeJobs) == 0 {
		return nil
	}
	return persistRepoChangeJobs(tx, temp)
}

func replacePRAttemptScope(tx *sql.Tx, attempt improvement.PRAttempt) error {
	temp := newSubsetStore()
	temp.prAttempts[attempt.ID] = attempt
	return persistPRAttempts(tx, temp)
}

func replaceAttemptWorkspaceScope(tx *sql.Tx, item improvement.AttemptWorkspace) error {
	temp := newSubsetStore()
	temp.attemptWorkspaces[item.ID] = item
	return persistAttemptWorkspaces(tx, temp)
}

func replacePostMergeReplayScope(tx *sql.Tx, store *MemoryStore, proposalID string) error {
	temp := newSubsetStore()
	for key, item := range store.postMergeReplay {
		if item.ProposalID == proposalID {
			temp.postMergeReplay[key] = item
		}
	}
	if len(temp.postMergeReplay) == 0 {
		return nil
	}
	return persistPostMergeReplays(tx, temp)
}

func replaceEvalRunScope(tx *sql.Tx, run evals.Run, judgments []evals.Judgment) error {
	temp := newSubsetStore()
	temp.evalRuns[run.ID] = run
	temp.evalJudgments[run.ID] = append([]evals.Judgment(nil), judgments...)
	if err := persistEvalRuns(tx, temp); err != nil {
		return err
	}
	return persistEvalJudgments(tx, temp)
}

func replaceOutcomeScope(tx *sql.Tx, record outcome.Record) error {
	temp := newSubsetStore()
	temp.outcomes[record.ID] = record
	return persistOutcomes(tx, temp)
}

func replaceHarnessOverlayScope(tx *sql.Tx, store *MemoryStore, overlayID string) error {
	item, ok := store.harnessOverlays[strings.TrimSpace(overlayID)]
	if !ok {
		return fmt.Errorf("harness overlay %s not found for persistence", overlayID)
	}
	item = normalizeHarnessOverlay(item)
	if item.Status == harness.OverlayStatusActive {
		if _, err := tx.Exec(`
			update harness_overlay
			set status = $1, updated_at = $2
			where role = $3 and status = $4 and id <> $5
		`, string(harness.OverlayStatusSuperseded), item.UpdatedAt, item.Role, string(harness.OverlayStatusActive), item.ID); err != nil {
			return err
		}
	}
	_, err := tx.Exec(`
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
	)
	return err
}

func replaceHarnessExperimentScope(tx *sql.Tx, store *MemoryStore, experimentID string) error {
	item, ok := store.harnessExperiments[strings.TrimSpace(experimentID)]
	if !ok {
		return fmt.Errorf("harness experiment %s not found for persistence", experimentID)
	}
	item = normalizeHarnessExperiment(item)
	_, err := tx.Exec(`
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
	)
	return err
}

func replaceHarnessSessionBindingScope(tx *sql.Tx, store *MemoryStore, bindingKey string) error {
	item, ok := store.harnessSessionBindings[strings.TrimSpace(bindingKey)]
	if !ok {
		return fmt.Errorf("harness session binding %s not found for persistence", bindingKey)
	}
	item = normalizeHarnessSessionBinding(item)
	_, err := tx.Exec(`
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
	)
	return err
}

func replaceHarnessExecutionScope(tx *sql.Tx, store *MemoryStore, executionID string) error {
	var (
		item harness.Execution
		ok   bool
	)
	executionID = strings.TrimSpace(executionID)
	for _, candidate := range store.harnessExecutions {
		if strings.TrimSpace(candidate.ID) == executionID {
			item = candidate
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("harness execution %s not found for persistence", executionID)
	}
	item = normalizeHarnessExecution(item)
	_, err := tx.Exec(`
		insert into harness_execution (
			id, operation_id, trace_id, proposal_id, role, session_scope_kind, session_scope_id, hermes_session_id, parent_session_id, harness_profile_id, effective_overlay_id, effective_overlay_version, memory_backend, memory_reads, memory_writes, created_at
		) values (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::jsonb,$15::jsonb,$16
		)
		on conflict (id) do update set
			operation_id = excluded.operation_id,
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
		firstNonEmpty(item.OperationID),
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
	)
	return err
}

func replaceKnowledgeEntryScope(tx *sql.Tx, entry knowledge.Entry) error {
	temp := newSubsetStore()
	temp.knowledgeEntries[entry.ID] = entry
	return persistKnowledgeEntries(tx, temp)
}

func replaceSettingsScope(tx *sql.Tx, settings improvement.Settings) error {
	temp := newSubsetStore()
	temp.settings = settings
	return persistSettings(tx, temp)
}

func replaceCronLeasesScope(tx *sql.Tx, store *MemoryStore) error {
	temp := newSubsetStore()
	temp.cronLeases = store.cronLeases
	return persistCronLeases(tx, temp)
}

func replaceFeedbackScope(tx *sql.Tx, store *MemoryStore, traceID string) error {
	temp := newSubsetStore()
	if items := store.feedbackRecords[traceID]; len(items) > 0 {
		temp.feedbackRecords[traceID] = append([]review.FeedbackRecord(nil), items...)
	}
	if len(temp.feedbackRecords) == 0 {
		return nil
	}
	return persistFeedback(tx, temp)
}

func replaceRatingScope(tx *sql.Tx, store *MemoryStore, traceID string) error {
	temp := newSubsetStore()
	if items := store.ratings[traceID]; len(items) > 0 {
		temp.ratings[traceID] = append([]review.HumanRating(nil), items...)
	}
	if len(temp.ratings) == 0 {
		return nil
	}
	return persistRatings(tx, temp)
}

func replaceImprovementNotesScope(tx *sql.Tx, store *MemoryStore, traceID string) error {
	temp := newSubsetStore()
	if items := store.notes[traceID]; len(items) > 0 {
		temp.notes[traceID] = append([]review.ImprovementNote(nil), items...)
	}
	if len(temp.notes) == 0 {
		return nil
	}
	return persistNotes(tx, temp)
}

func replaceProposalPromoterScope(tx *sql.Tx, store *MemoryStore) error {
	if err := replaceAllCandidates(tx, store); err != nil {
		return err
	}
	if err := replaceAllProposals(tx, store); err != nil {
		return err
	}
	return replaceCronLeasesScope(tx, store)
}

func replaceTraceAndWorkflowScope(tx *sql.Tx, store *MemoryStore, trace events.Trace) error {
	if err := replaceTraceScope(tx, store, trace.Summary.TraceID); err != nil {
		return err
	}
	if workflowID := trace.Summary.WorkflowID; workflowID != "" {
		for _, item := range store.workflows {
			if item.ID == workflowID {
				return replaceWorkflowScope(tx, item)
			}
		}
	}
	return nil
}

func replaceEventMaterializationScope(tx *sql.Tx, store *MemoryStore, event ingestion.EventEnvelope) error {
	temp := newSubsetStore()
	temp.events = append(temp.events, event)
	if err := persistEvents(tx, temp); err != nil {
		return err
	}

	conversationIDs := map[string]struct{}{}
	caseIDs := map[string]struct{}{}
	traceIDs := map[string]struct{}{}
	workflowIDs := map[string]struct{}{}
	proposalIDs := map[string]struct{}{}
	actionIntentIDs := map[string]struct{}{}

	addID := func(target map[string]struct{}, value string) {
		if value != "" {
			target[value] = struct{}{}
		}
	}

	for _, item := range store.conversationEntries {
		if item.EventID == event.ID {
			temp.conversationEntries = append(temp.conversationEntries, item)
			addID(conversationIDs, item.ConversationID)
			addID(traceIDs, item.TraceID)
		}
	}
	for _, item := range store.ingestions {
		if item.EventID == event.ID {
			temp.ingestions = append(temp.ingestions, item)
			addID(conversationIDs, item.ConversationID)
			addID(caseIDs, item.CaseID)
		}
	}
	for _, item := range store.outcomes {
		if item.SourceEventID == event.SourceEventID {
			temp.outcomes[item.ID] = item
			addID(conversationIDs, item.ConversationID)
			addID(caseIDs, item.CaseID)
			addID(traceIDs, item.TraceID)
			addID(proposalIDs, item.ProposalID)
		}
	}
	for id, item := range store.conversations {
		if item.LatestEventID == event.ID {
			conversationIDs[id] = struct{}{}
		}
	}
	for id, item := range store.cases {
		if item.OpenedByEventID == event.ID || item.ClosedByEventID == event.ID {
			caseIDs[id] = struct{}{}
			addID(conversationIDs, item.ConversationID)
		}
	}
	for traceID, trace := range store.traces {
		if trace.Summary.TriggerEventID == event.ID {
			traceIDs[traceID] = struct{}{}
		}
		for _, step := range trace.Events {
			if step.TriggerEventID == event.ID {
				traceIDs[traceID] = struct{}{}
				break
			}
		}
	}
	if proposalID, ok := event.Metadata["proposal_id"].(string); ok {
		addID(proposalIDs, proposalID)
	}

	for caseID := range caseIDs {
		if item, ok := store.cases[caseID]; ok {
			temp.cases[caseID] = item
			addID(conversationIDs, item.ConversationID)
			addID(traceIDs, item.LatestTraceID)
		}
		if item, ok := store.workflowLines[caseID]; ok {
			temp.workflowLines[caseID] = item
		}
	}
	for convID := range conversationIDs {
		if item, ok := store.conversations[convID]; ok {
			temp.conversations[convID] = item
			if policyItem, ok := store.threadPolicies[item.ExternalKey]; ok {
				temp.threadPolicies[policyItem.ThreadKey] = policyItem
			}
		}
	}
	for _, item := range store.assignments {
		if _, ok := caseIDs[item.CaseID]; ok {
			temp.assignments = append(temp.assignments, item)
			continue
		}
		if _, ok := conversationIDs[item.ConversationID]; ok {
			temp.assignments = append(temp.assignments, item)
		}
	}
	for _, item := range store.workflows {
		if _, ok := caseIDs[item.CaseID]; ok {
			temp.workflows = append(temp.workflows, item)
			addID(workflowIDs, item.ID)
			continue
		}
		if _, ok := conversationIDs[item.ConversationID]; ok {
			temp.workflows = append(temp.workflows, item)
			addID(workflowIDs, item.ID)
		}
	}
	for traceID, trace := range store.traces {
		if _, ok := caseIDs[trace.Summary.CaseID]; ok {
			traceIDs[traceID] = struct{}{}
			addID(workflowIDs, trace.Summary.WorkflowID)
			continue
		}
		if _, ok := conversationIDs[trace.Summary.ConversationID]; ok && trace.Summary.TriggerEventID == event.ID {
			traceIDs[traceID] = struct{}{}
			addID(workflowIDs, trace.Summary.WorkflowID)
		}
	}
	for proposalID, item := range store.proposals {
		if _, ok := proposalIDs[proposalID]; ok {
			temp.proposals[proposalID] = item
			addID(caseIDs, item.CaseID)
			addID(conversationIDs, item.ConversationID)
			continue
		}
		if _, ok := caseIDs[item.CaseID]; ok && item.OriginTraceID != "" {
			temp.proposals[proposalID] = item
		}
	}

	if len(temp.conversations) > 0 {
		if err := persistConversations(tx, temp); err != nil {
			return err
		}
	}
	if len(temp.threadPolicies) > 0 {
		if err := persistThreadPolicies(tx, temp); err != nil {
			return err
		}
	}
	if len(temp.conversationEntries) > 0 {
		if err := persistConversationEntries(tx, temp); err != nil {
			return err
		}
	}
	if len(temp.cases) > 0 {
		if err := persistCases(tx, temp); err != nil {
			return err
		}
	}
	if len(temp.workflowLines) > 0 {
		if err := persistWorkflowLines(tx, temp); err != nil {
			return err
		}
	}
	if len(temp.ingestions) > 0 {
		if err := persistIngestions(tx, temp); err != nil {
			return err
		}
	}
	if len(temp.assignments) > 0 {
		if err := persistAssignments(tx, temp); err != nil {
			return err
		}
	}
	if len(temp.workflows) > 0 {
		if err := persistWorkflows(tx, temp); err != nil {
			return err
		}
	}
	for traceID := range traceIDs {
		if _, ok := store.traces[traceID]; !ok {
			continue
		}
		if err := replaceTraceScope(tx, store, traceID); err != nil {
			return err
		}
	}
	for id, item := range store.actionIntents {
		if _, ok := caseIDs[item.CaseID]; ok {
			actionIntentIDs[id] = struct{}{}
		} else if _, ok := traceIDs[item.TraceID]; ok {
			actionIntentIDs[id] = struct{}{}
		}
		if _, ok := actionIntentIDs[id]; !ok {
			continue
		}
		if err := replaceActionIntentScope(tx, item); err != nil {
			return err
		}
	}
	for actionIntentID := range actionIntentIDs {
		if err := replaceActionResultScope(tx, store, actionIntentID); err != nil {
			return err
		}
	}
	if len(temp.outcomes) > 0 {
		if err := persistOutcomes(tx, temp); err != nil {
			return err
		}
	}
	if len(temp.proposals) > 0 {
		if err := persistProposals(tx, temp); err != nil {
			return err
		}
	}
	return nil
}

func insertActionResult(tx *sql.Tx, item action.Result) error {
	_, err := tx.Exec(`insert into action_result (id, operation_id, action_intent_id, attempt_id, attempt_number, executor, provider, provider_ref, request_artifact_id, response_artifact_id, status, error_code, error_message, started_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		on conflict (id) do update set
			operation_id = excluded.operation_id,
			action_intent_id = excluded.action_intent_id,
			attempt_id = excluded.attempt_id,
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
		item.ID,
		firstNonEmpty(item.OperationID),
		item.ActionIntentID,
		firstNonEmpty(item.AttemptID),
		item.AttemptNumber,
		item.Executor,
		nullString(item.Provider),
		nullString(item.ProviderRef),
		nullString(item.RequestArtifactID),
		nullString(item.ResponseArtifactID),
		string(item.Status),
		nullString(item.ErrorCode),
		nullString(item.ErrorMessage),
		item.StartedAt,
		item.CompletedAt,
	)
	return err
}

func sameProposalDecision(current review.ProposalReview, incoming review.ProposalReview) bool {
	currentClasses := append([]string(nil), current.FailureClasses...)
	if len(currentClasses) == 0 && strings.TrimSpace(current.FailureClass) != "" {
		currentClasses = []string{current.FailureClass}
	}
	incomingClasses := append([]string(nil), incoming.FailureClasses...)
	if len(incomingClasses) == 0 && strings.TrimSpace(incoming.FailureClass) != "" {
		incomingClasses = []string{incoming.FailureClass}
	}
	if current.Decision != incoming.Decision || strings.TrimSpace(current.ReviewerID) != strings.TrimSpace(incoming.ReviewerID) || strings.TrimSpace(current.Rationale) != strings.TrimSpace(incoming.Rationale) || len(currentClasses) != len(incomingClasses) {
		return false
	}
	for idx := range currentClasses {
		if currentClasses[idx] != incomingClasses[idx] {
			return false
		}
	}
	return true
}

func proposalDecisionAlreadyApplied(store *MemoryStore, proposalID string, incoming review.ProposalReview) (review.Proposal, bool) {
	proposal, ok := store.proposals[proposalID]
	if !ok || len(proposal.Reviews) == 0 {
		return review.Proposal{}, false
	}
	if proposal.Status != review.ProposalStatus(incoming.Decision) {
		return review.Proposal{}, false
	}
	latest := proposal.Reviews[len(proposal.Reviews)-1]
	if !sameProposalDecision(latest, incoming) {
		return review.Proposal{}, false
	}
	return proposal, true
}

func nowOr(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
