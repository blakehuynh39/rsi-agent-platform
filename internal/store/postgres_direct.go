package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func (p *PostgresStore) withTx(fn func(tx *sql.Tx) error) error {
	tx, err := p.db.Begin()
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

func (p *PostgresStore) withLoadedStoreTx(fn func(tx *sql.Tx, store *MemoryStore) error) error {
	return p.withTx(func(tx *sql.Tx) error {
		store, err := loadStore(tx)
		if err != nil {
			return err
		}
		return fn(tx, store)
	})
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
	_, err := tx.Exec(`select pg_advisory_xact_lock(hashtext($1)::bigint)`, key)
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

func replaceWorkItemScope(tx *sql.Tx, item queue.WorkItem) error {
	temp := newSubsetStore()
	temp.workItems[item.ID] = item
	return persistWorkItems(tx, temp)
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
	for _, item := range store.workItems {
		if _, ok := caseIDs[item.CaseID]; ok {
			if err := replaceWorkItemScope(tx, item); err != nil {
				return err
			}
			continue
		}
		if _, ok := traceIDs[item.TraceID]; ok {
			if err := replaceWorkItemScope(tx, item); err != nil {
				return err
			}
			continue
		}
		if _, ok := workflowIDs[item.WorkflowID]; ok {
			if err := replaceWorkItemScope(tx, item); err != nil {
				return err
			}
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
	_, err := tx.Exec(`insert into action_result (id, action_intent_id, attempt_id, attempt_number, executor, provider, provider_ref, request_artifact_id, response_artifact_id, status, error_code, error_message, started_at, completed_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		item.ID,
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

func scanWorkItem(scanner rowScanner) (queue.WorkItem, error) {
	var item queue.WorkItem
	var queueName, status string
	var traceID, workflowID, ingestionID, conversationID, caseID, triggerEventID, proposalID, threadKey, intent, repoScope, requestedBy, approvalMode, responseMode, leaseOwner, lastError sql.NullString
	var payload []byte
	var leaseExpiresAt, completedAt sql.NullTime
	if err := scanner.Scan(&item.ID, &queueName, &item.Kind, &status, &traceID, &workflowID, &ingestionID, &conversationID, &caseID, &triggerEventID, &proposalID, &threadKey, &intent, &repoScope, &requestedBy, &approvalMode, &responseMode, &payload, &item.Attempts, &leaseOwner, &leaseExpiresAt, &lastError, &item.CreatedAt, &item.UpdatedAt, &completedAt); err != nil {
		return queue.WorkItem{}, err
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
	return item, nil
}

func workItemSelectColumns() string {
	return `id, queue, kind, status, trace_id, workflow_id, ingestion_id, conversation_id, case_id, trigger_event_id, proposal_id, thread_key, intent, repo_scope, requested_by, approval_mode, response_mode, payload, attempts, lease_owner, lease_expires_at, last_error, created_at, updated_at, completed_at`
}

func workItemSelectColumnsWithAlias(alias string) string {
	columns := []string{
		"id",
		"queue",
		"kind",
		"status",
		"trace_id",
		"workflow_id",
		"ingestion_id",
		"conversation_id",
		"case_id",
		"trigger_event_id",
		"proposal_id",
		"thread_key",
		"intent",
		"repo_scope",
		"requested_by",
		"approval_mode",
		"response_mode",
		"payload",
		"attempts",
		"lease_owner",
		"lease_expires_at",
		"last_error",
		"created_at",
		"updated_at",
		"completed_at",
	}
	qualified := make([]string, 0, len(columns))
	for _, column := range columns {
		qualified = append(qualified, alias+"."+column)
	}
	return strings.Join(qualified, ", ")
}

func findExistingWorkItemByDedupe(tx *sql.Tx, item queue.WorkItem) (queue.WorkItem, bool, error) {
	dedupeKey := workItemDedupeKey(item)
	if dedupeKey == "" {
		return queue.WorkItem{}, false, nil
	}
	if err := advisoryLock(tx, fmt.Sprintf("work-item:%s:%s:%s", item.Queue, item.Kind, dedupeKey)); err != nil {
		return queue.WorkItem{}, false, err
	}
	row := tx.QueryRow(
		`select `+workItemSelectColumns()+` from work_item where queue = $1 and kind = $2 and payload->>'dedupe_key' = $3 and status not in ($4, $5) order by created_at desc, id desc limit 1`,
		string(item.Queue),
		item.Kind,
		dedupeKey,
		string(queue.WorkFailed),
		string(queue.WorkCanceled),
	)
	existing, err := scanWorkItem(row)
	if err == sql.ErrNoRows {
		return queue.WorkItem{}, false, nil
	}
	if err != nil {
		return queue.WorkItem{}, false, err
	}
	return existing, true, nil
}

func queuePredicate(queues []queue.QueueName, startIndex int) (string, []any) {
	if len(queues) == 0 {
		return "1=1", nil
	}
	parts := make([]string, 0, len(queues))
	args := make([]any, 0, len(queues))
	for idx, name := range queues {
		parts = append(parts, fmt.Sprintf("$%d", startIndex+idx))
		args = append(args, string(name))
	}
	return fmt.Sprintf("queue in (%s)", strings.Join(parts, ",")), args
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
