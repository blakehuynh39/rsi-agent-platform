package store

import (
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
	"github.com/piplabs/rsi-agent-platform/internal/workflowplan"
)

type commandApplyResult struct {
	receipt             transition.CommandReceipt
	bundle              transitionPersistBundle
	proposalID          string
	knowledgeID         string
	feedbackID          string
	eventID             string
	ingestionID         string
	threadKey           string
	attemptID           string
	traceID             string
	candidateKey        string
	evalRunID           string
	evalJudgments       []evals.Judgment
	outcomeID           string
	harnessOverlayID    string
	harnessExperimentID string
	harnessBindingKey   string
	harnessExecutionID  string
	prAttemptID         string
	workspaceID         string
}

func (s *MemoryStore) SubmitCommand(command transition.CommandEnvelope) (transition.CommandReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.submitCommandChainLocked(command)
}

func (p *PostgresStore) SubmitCommand(command transition.CommandEnvelope) (receipt transition.CommandReceipt, err error) {
	err = p.withTx(func(tx *sql.Tx) error {
		command = normalizeCommandEnvelope(command)
		if err := advisoryLock(tx, "command:"+command.CommandID); err != nil {
			return err
		}
		if command.AggregateID != "" {
			if err := advisoryLock(tx, fmt.Sprintf("aggregate:%s:%s", command.MachineKind, command.AggregateID)); err != nil {
				return err
			}
		}
		existing, scanErr := scanCommandReceipt(tx.QueryRow(`select command_id, machine_kind, aggregate_id, command_kind, causation_id, actor, decision_kind, reason, aggregate_version, result_ref, created_at, updated_at from command_receipt where command_id = $1`, strings.TrimSpace(command.CommandID)))
		if scanErr == nil {
			receipt = existing
			return nil
		}
		if scanErr != sql.ErrNoRows {
			return scanErr
		}
		store, loadErr := loadStore(tx)
		if loadErr != nil {
			return loadErr
		}
		rootReceipt, results, receipts, events, effects, execErr := store.executeCommandChainLocked(command)
		if execErr != nil {
			return execErr
		}
		for _, result := range results {
			if persistErr := persistCommandMutation(tx, store, result); persistErr != nil {
				return persistErr
			}
		}
		if persistErr := persistDomainEvents(tx, events); persistErr != nil {
			return persistErr
		}
		if persistErr := persistEffectExecutions(tx, effects); persistErr != nil {
			return persistErr
		}
		if persistErr := persistCommandReceipts(tx, receipts); persistErr != nil {
			return persistErr
		}
		receipt = rootReceipt
		return nil
	})
	return
}

func (s *MemoryStore) submitCommandChainLocked(command transition.CommandEnvelope) (transition.CommandReceipt, error) {
	rootReceipt, _, _, _, _, err := s.executeCommandChainLocked(command)
	return rootReceipt, err
}

func (s *MemoryStore) executeCommandChainLocked(command transition.CommandEnvelope) (transition.CommandReceipt, []commandApplyResult, []transition.CommandReceipt, []transition.DomainEvent, []transition.EffectExecution, error) {
	command = normalizeCommandEnvelope(command)
	queue := []transition.CommandEnvelope{command}
	results := make([]commandApplyResult, 0, 1)
	receipts := make([]transition.CommandReceipt, 0, 1)
	events := make([]transition.DomainEvent, 0, 4)
	effects := make([]transition.EffectExecution, 0, 2)
	var rootReceipt transition.CommandReceipt
	for len(queue) > 0 {
		current := normalizeCommandEnvelope(queue[0])
		queue = queue[1:]
		if existing, ok := s.commandReceipts[strings.TrimSpace(current.CommandID)]; ok {
			if current.CommandID == command.CommandID {
				rootReceipt = existing
			}
			continue
		}
		result, err := s.applyCommandLocked(current)
		if err != nil {
			return transition.CommandReceipt{}, nil, nil, nil, nil, err
		}
		s.appendTransitionBundleLocked(result.bundle)
		receipt, _, err := s.recordCommandReceiptLocked(result.receipt)
		if err != nil {
			return transition.CommandReceipt{}, nil, nil, nil, nil, err
		}
		result.receipt = receipt
		results = append(results, result)
		receipts = append(receipts, receipt)
		events = append(events, result.bundle.Events...)
		effects = append(effects, result.bundle.Effects...)
		queue = append(queue, result.bundle.Commands...)
		if current.CommandID == command.CommandID {
			rootReceipt = receipt
		}
	}
	if rootReceipt.CommandID == "" {
		return transition.CommandReceipt{}, nil, nil, nil, nil, fmt.Errorf("root command %s did not produce a receipt", command.CommandID)
	}
	return rootReceipt, results, receipts, events, effects, nil
}

func (s *MemoryStore) applyCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	command = normalizeCommandEnvelope(command)
	switch command.MachineKind {
	case transition.MachineIngress:
		return s.applyIngressCommandLocked(command)
	case transition.MachineWorkflow:
		return s.applyWorkflowCommandLocked(command)
	case transition.MachineProblemLine:
		return s.applyProblemLineCommandLocked(command)
	case transition.MachineAttempt:
		return s.applyAttemptCommandLocked(command)
	case transition.MachineAction:
		return s.applyActionCommandLocked(command)
	case transition.MachineHarness:
		return s.applyHarnessCommandLocked(command)
	case transition.MachineThreadPolicy:
		return s.applyThreadPolicyCommandLocked(command)
	case transition.MachineSettings:
		return s.applySettingsCommandLocked(command)
	case transition.MachineKnowledge:
		return s.applyKnowledgeCommandLocked(command)
	case transition.MachineProposalLine:
		return s.applyProposalCommandLocked(command)
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported machine kind %s", command.MachineKind)
	}
}

func (s *MemoryStore) applyIngressCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	decision := transition.ReduceIngress(transition.IngressSnapshot{
		State: transition.IngressStatePending,
	}, command)
	result := commandApplyResult{
		bundle: buildCommandBundle(command, decision.TransitionDecision, 1),
	}
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		switch transition.IngressCommandKind(command.CommandKind) {
		case transition.CommandIngressRecordEvent:
			event, err := s.createEventLocked(eventEnvelopeFromCommand(command))
			if err != nil {
				return commandApplyResult{}, err
			}
			result.eventID = event.ID
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(result.bundle.Events[0].Payload, map[string]any{
					"event_id":        event.ID,
					"source":          string(event.Source),
					"source_event_id": event.SourceEventID,
					"dedupe_key":      event.DedupeKey,
				})
			}
			s.appendWorkflowStartCommandFromEventLocked(&result.bundle, command, event.ID)
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, event.CreatedAt, 1, event.ID)
			return result, nil
		case transition.CommandIngressRecordSlack:
			item, err := s.createIngestionLocked(slackEnvelopeFromCommand(command))
			if err != nil {
				return commandApplyResult{}, err
			}
			result.eventID = item.EventID
			result.ingestionID = item.ID
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(result.bundle.Events[0].Payload, map[string]any{
					"event_id":     item.EventID,
					"ingestion_id": item.ID,
					"thread_key":   item.ThreadKey,
				})
			}
			s.appendWorkflowStartCommandFromEventLocked(&result.bundle, command, item.EventID)
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, item.CreatedAt, 1, item.ID)
			return result, nil
		default:
			return commandApplyResult{}, fmt.Errorf("unsupported ingress command kind %s", command.CommandKind)
		}
	case transition.DecisionReject, transition.DecisionNoop:
		result.receipt = buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, 1, "")
		return result, nil
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported ingress decision kind %s", decision.DecisionKind)
	}
}

func (s *MemoryStore) applyWorkflowCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	workflow, ok := findWorkflowByID(s.workflows, command.AggregateID)
	if !ok {
		return commandApplyResult{}, errors.New("workflow not found")
	}
	decision := transition.ReduceWorkflow(transition.WorkflowSnapshot{
		State:   workflowStateFromStatus(workflow.Status),
		TraceID: workflow.TraceID,
	}, command)
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		next, err := s.setWorkflowMachineStateLocked(command.AggregateID, decision.NextState, workflowLastErrorForCommand(command), command.OccurredAt)
		if err != nil {
			return commandApplyResult{}, err
		}
		workflow = next
		if err := s.projectWorkflowTraceLocked(workflow, command); err != nil {
			return commandApplyResult{}, err
		}
	case transition.DecisionReject, transition.DecisionNoop:
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported workflow decision kind %s", decision.DecisionKind)
	}
	result := commandApplyResult{
		receipt: buildCommandReceipt(command, decision.TransitionDecision, workflow.UpdatedAt, workflow.Version, workflow.ID),
		bundle:  buildCommandBundle(command, decision.TransitionDecision, workflow.Version),
	}
	if decision.DecisionKind == transition.DecisionAdvance &&
		transition.WorkflowCommandKind(command.CommandKind) == transition.CommandWorkflowStarted &&
		workflowPlanningRequested(command) {
		if err := s.appendWorkflowPlanningCommandsLocked(&result.bundle, command, workflow); err != nil {
			return commandApplyResult{}, err
		}
	}
	return result, nil
}

func (s *MemoryStore) appendWorkflowStartCommandFromEventLocked(bundle *transitionPersistBundle, parent transition.CommandEnvelope, eventID string) {
	trace, workflow, ok := s.workflowForTriggerEventLocked(eventID)
	if !ok || strings.TrimSpace(workflow.Status) != "queued" {
		return
	}
	appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflow.ID,
		CommandKind: string(transition.CommandWorkflowStarted),
		CommandID:   storeWorkflowCommandID(workflow.ID, transition.CommandWorkflowStarted),
		Actor:       parent.Actor,
		OccurredAt:  trace.Summary.StartedAt,
		Payload: map[string]any{
			"default_repo":         stringFromCommand(parent, "default_repo"),
			"allowed_target_repos": stringSliceFromCommand(parent, "allowed_target_repos"),
			"knowledge_base_url":   stringFromCommand(parent, "knowledge_base_url"),
			"sandbox_namespace":    stringFromCommand(parent, "sandbox_namespace"),
			"resume_queue":         string(queue.WorkflowQueue),
		},
	}, "ingress materialized workflow")
}

func (s *MemoryStore) appendWorkflowPlanningCommandsLocked(bundle *transitionPersistBundle, parent transition.CommandEnvelope, workflow Workflow) error {
	trace, ok := s.traces[strings.TrimSpace(workflow.TraceID)]
	if !ok {
		return fmt.Errorf("trace %s not found for workflow planning", workflow.TraceID)
	}
	ingestion, ok := findIngestion(s.ingestions, workflow.IngestionID)
	if !ok {
		return fmt.Errorf("ingestion %s not found for workflow planning", workflow.IngestionID)
	}
	planning := workflowPlanningConfigFromCommand(parent)
	resumeQueue := firstNonEmpty(stringFromCommand(parent, "resume_queue"), string(queue.WorkflowQueue))
	repo := workflowplan.ResolveTargetRepo(planning, strings.TrimSpace(ingestion.Text))
	toolNames := workflowplan.ToolPlan(workflow.Intent, strings.TrimSpace(ingestion.Text), repo)
	if len(toolNames) == 0 {
		appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandContextSkipped),
			CommandID:   storeWorkflowCommandID(workflow.ID, transition.CommandContextSkipped),
			Actor:       parent.Actor,
			OccurredAt:  parent.OccurredAt,
			Payload: map[string]any{
				"tool_count":   0,
				"resume_queue": resumeQueue,
				"trace_events": []events.TraceEvent{{
					TraceID:     trace.Summary.TraceID,
					IngestionID: trace.Summary.IngestionID,
					WorkflowID:  trace.Summary.WorkflowID,
					Plane:       "execution",
					Service:     "runner",
					Actor:       workflow.AssignedBot,
					EventType:   "runner.started",
					Status:      events.StatusRunning,
					StartedAt:   parent.OccurredAt,
					Description: "Runner task dispatched with verbose reasoning enabled.",
				}},
			},
		}, "workflow planning determined no context actions were required")
		return nil
	}

	traceEvents := make([]events.TraceEvent, 0, len(toolNames))
	for _, toolName := range toolNames {
		traceEvents = append(traceEvents, events.TraceEvent{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     "tool-gateway",
			Actor:       workflow.AssignedBot,
			EventType:   "tool.requested",
			Status:      events.StatusQueued,
			StartedAt:   parent.OccurredAt,
			Description: fmt.Sprintf("Requested %s.", toolName),
		})
	}
	appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflow.ID,
		CommandKind: string(transition.CommandContextActionsQueued),
		CommandID:   storeWorkflowCommandID(workflow.ID, transition.CommandContextActionsQueued),
		Actor:       parent.Actor,
		OccurredAt:  parent.OccurredAt,
		Payload: map[string]any{
			"tool_count":   len(toolNames),
			"resume_queue": resumeQueue,
			"trace_events": traceEvents,
		},
	}, "workflow planning queued context actions")

	for _, toolName := range toolNames {
		idempotencyKey := fmt.Sprintf("%s:%s:%s", trace.Summary.TraceID, toolName, trace.Summary.TriggerEventID)
		actionID := storeActionIntentIDFromIdempotencyKey(idempotencyKey)
		appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
			MachineKind: transition.MachineAction,
			AggregateID: actionID,
			CommandKind: string(transition.CommandActionQueue),
			CommandID:   storeActionCommandID(actionID, transition.CommandActionQueue, ""),
			Actor:       firstNonEmpty(parent.Actor, "control-plane"),
			OccurredAt:  parent.OccurredAt,
			Payload: map[string]any{
				"owner_plane":     "control",
				"conversation_id": trace.Summary.ConversationID,
				"case_id":         trace.Summary.CaseID,
				"trace_id":        trace.Summary.TraceID,
				"kind":            string(action.KindToolRead),
				"phase_key":       "collect_context",
				"target_ref":      toolName,
				"request_payload": mergeCommandMetadataPayload(
					workflowplan.BuildToolRequestPayload(planning, workflowplan.RequestContext{
						Trace:          trace.Summary,
						WorkflowID:     workflow.ID,
						ConversationID: workflow.ConversationID,
						CaseID:         workflow.CaseID,
						WorkflowKind:   workflow.Kind,
						AssignedBot:    workflow.AssignedBot,
						Question:       strings.TrimSpace(ingestion.Text),
						ChannelID:      ingestion.ChannelID,
						ThreadTS:       ingestion.ThreadTS,
					}, parent.OccurredAt),
					map[string]any{"resume_queue": resumeQueue},
				),
				"idempotency_key": idempotencyKey,
				"approval_mode":   "not_required",
				"approval_state":  "not_required",
				"requested_by":    firstNonEmpty(parent.Actor, "control-plane"),
				"rationale":       fmt.Sprintf("Collect context via %s.", toolName),
				"evidence_refs": []events.EvidenceRef{
					{Kind: "trace", Ref: trace.Summary.TraceID, Summary: trace.Summary.WorkflowKind},
				},
			},
		}, "workflow planning queued context action")
	}
	return nil
}

func (s *MemoryStore) workflowForTriggerEventLocked(eventID string) (events.Trace, Workflow, bool) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return events.Trace{}, Workflow{}, false
	}
	for _, trace := range s.traces {
		if strings.TrimSpace(trace.Summary.TriggerEventID) != eventID {
			continue
		}
		workflow, ok := findWorkflowByID(s.workflows, trace.Summary.WorkflowID)
		if !ok {
			return events.Trace{}, Workflow{}, false
		}
		return trace, workflow, true
	}
	return events.Trace{}, Workflow{}, false
}

func appendFollowOnCommand(bundle *transitionPersistBundle, parent transition.CommandEnvelope, next transition.CommandEnvelope, reason string) {
	payload := cloneMetadata(commandPayload(parent))
	for key, value := range cloneMetadata(next.Payload) {
		if payload == nil {
			payload = map[string]any{}
		}
		payload[key] = value
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["reason"] = reason
	payload["parent_command_id"] = parent.CommandID
	bundle.Commands = append(bundle.Commands, normalizeCommandEnvelope(transition.CommandEnvelope{
		MachineKind: next.MachineKind,
		AggregateID: next.AggregateID,
		CommandKind: next.CommandKind,
		CommandID:   next.CommandID,
		CausationID: firstNonEmpty(next.CausationID, parent.CommandID),
		Actor:       firstNonEmpty(next.Actor, parent.Actor),
		OccurredAt:  firstNonZeroTime(optionalTime(next.OccurredAt), parent.OccurredAt),
		Payload:     payload,
	}))
}

func workflowPlanningConfigFromCommand(command transition.CommandEnvelope) workflowplan.RuntimeConfig {
	return workflowplan.RuntimeConfig{
		DefaultRepo:      stringFromCommand(command, "default_repo"),
		AllowedRepos:     stringSliceFromCommand(command, "allowed_target_repos"),
		KnowledgeBaseURL: stringFromCommand(command, "knowledge_base_url"),
		SandboxNamespace: stringFromCommand(command, "sandbox_namespace"),
	}
}

func workflowPlanningRequested(command transition.CommandEnvelope) bool {
	if strings.TrimSpace(stringFromCommand(command, "resume_queue")) != "" {
		return true
	}
	cfg := workflowPlanningConfigFromCommand(command)
	return strings.TrimSpace(cfg.DefaultRepo) != "" ||
		len(cfg.AllowedRepos) > 0 ||
		strings.TrimSpace(cfg.KnowledgeBaseURL) != "" ||
		strings.TrimSpace(cfg.SandboxNamespace) != ""
}

func storeWorkflowCommandID(workflowID string, kind transition.WorkflowCommandKind) string {
	return fmt.Sprintf("cmd-workflow:%s:%s", strings.TrimSpace(workflowID), string(kind))
}

func storeActionCommandID(actionID string, kind transition.ActionExecutionCommandKind, operationID string) string {
	base := fmt.Sprintf("cmd-action:%s:%s", strings.TrimSpace(actionID), string(kind))
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return base
	}
	return base + ":" + operationID
}

func storeAttemptCommandID(attemptID string, kind transition.AttemptPhaseCommandKind) string {
	return fmt.Sprintf("cmd-attempt:%s:%s", strings.TrimSpace(attemptID), string(kind))
}

func storeActionIntentIDFromIdempotencyKey(key string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(key)))
	return fmt.Sprintf("action-%x", sum[:8])
}

func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func (s *MemoryStore) applyProblemLineCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	trace, traceExists := s.traces[command.AggregateID]
	slots := s.proposalSlotsLocked(command.OccurredAt)
	snapshot := transition.ProblemLineSnapshot{
		State:                   transition.ProblemLineStateObserving,
		TraceExists:             traceExists,
		SlotsAvailable:          slots.Available > 0,
		HasPromotableCandidates: hasPromotableCandidatesLocked(s.candidates),
		PromotionLeaseBlocked:   promotionLeaseBlockedLocked(s.cronLeases, command.Actor, command.OccurredAt),
	}
	decision := transition.ReduceProblemLine(snapshot, command)
	result := commandApplyResult{
		bundle: buildCommandBundle(command, decision.TransitionDecision, 1),
	}
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		switch transition.ProblemLineCommandKind(command.CommandKind) {
		case transition.CommandProblemLineEvaluateTrace:
			run, judgments, err := s.evaluateTraceLocked(command.AggregateID, firstNonEmpty(stringFromCommand(command, "trigger"), "manual"))
			if err != nil {
				return commandApplyResult{}, err
			}
			candidateKey := candidateKeyForTrace(trace, s.findEventByTraceLocked(trace))
			result.evalRunID = run.ID
			result.traceID = run.TraceID
			result.evalJudgments = append([]evals.Judgment(nil), judgments...)
			result.candidateKey = candidateKey
			for idx := range result.bundle.Effects {
				if result.bundle.Effects[idx].MachineKind != transition.MachineProblemLine || result.bundle.Effects[idx].EffectKind != transition.EffectInvokeRunner {
					continue
				}
				result.bundle.Effects[idx].IdempotencyKey = fmt.Sprintf("problem-line-eval:%s", run.ID)
				result.bundle.Effects[idx].Payload = mergeCommandMetadataPayload(result.bundle.Effects[idx].Payload, map[string]any{
					"trace_id":      run.TraceID,
					"eval_run_id":   run.ID,
					"candidate_key": candidateKey,
				})
			}
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(
					result.bundle.Events[0].Payload,
					map[string]any{
						"eval_run_id":     run.ID,
						"trace_id":        trace.Summary.TraceID,
						"candidate_key":   candidateKey,
						"judgment_count":  len(judgments),
						"overall_verdict": run.OverallVerdict,
					},
				)
			}
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, run.CompletedAt, 1, run.ID)
			return result, nil
		case transition.CommandProblemLinePromote:
			holder := firstNonEmpty(command.Actor, stringFromCommand(command, "requested_by"), "improvement-plane-cron")
			s.cronLeases["improvement-plane-cron"] = improvement.CronLease{
				Name:      "improvement-plane-cron",
				Holder:    holder,
				ExpiresAt: command.OccurredAt.Add(proposalPromoterLease),
			}
			limit := int(floatFromCommand(command, "limit"))
			if limit <= 0 {
				limit = normalizedSettings(s.settings).ActiveProposalCap
			}
			promotion, err := s.promoteCandidatesLocked(holder, limit)
			if err != nil {
				return commandApplyResult{}, err
			}
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(
					result.bundle.Events[0].Payload,
					map[string]any{
						"promoted":           promotion.Promoted,
						"promoted_ids":       promotion.PromotedIDs,
						"blocked_by_cap":     promotion.BlockedByCap,
						"stale_proposal_ids": promotion.StaleProposalIDs,
					},
				)
			}
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, 1, command.CommandID)
			return result, nil
		case transition.CommandProblemLineRecordOutcome:
			record, err := s.recordOutcomeLocked(problemLineOutcomeFromCommand(command))
			if err != nil {
				return commandApplyResult{}, err
			}
			result.outcomeID = record.ID
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(
					result.bundle.Events[0].Payload,
					map[string]any{
						"outcome_id":  record.ID,
						"case_id":     record.CaseID,
						"proposal_id": record.ProposalID,
						"trace_id":    record.TraceID,
						"verdict":     string(record.Verdict),
					},
				)
			}
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, record.RecordedAt, 1, record.ID)
			return result, nil
		case transition.CommandProblemLineProjectTrace:
			traceID := firstNonEmpty(stringFromCommand(command, "trace_id"), command.AggregateID)
			if err := s.projectTraceFromCommandLocked(traceID, command); err != nil {
				return commandApplyResult{}, err
			}
			result.traceID = traceID
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, 1, traceID)
			return result, nil
		case transition.CommandProblemLineRecordFeedback:
			record, err := s.addFeedbackLocked(feedbackRecordFromCommand(command), command.OccurredAt)
			if err != nil {
				return commandApplyResult{}, err
			}
			result.feedbackID = record.ID
			result.traceID = record.TraceID
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(result.bundle.Events[0].Payload, map[string]any{
					"feedback_id": record.ID,
					"trace_id":    record.TraceID,
					"target_type": string(record.TargetType),
					"target_id":   record.TargetID,
					"reviewer_id": record.ReviewerID,
				})
			}
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, record.CreatedAt, 1, record.ID)
			return result, nil
		case transition.CommandProblemLineRecordRating:
			rating, err := s.addRatingLocked(command.AggregateID, humanRatingFromCommand(command), timeFromCommand(command, "created_at", command.OccurredAt))
			if err != nil {
				return commandApplyResult{}, err
			}
			result.traceID = rating.TraceID
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(result.bundle.Events[0].Payload, map[string]any{
					"trace_id":    rating.TraceID,
					"score":       rating.Score,
					"verdict":     rating.Verdict,
					"reviewer_id": rating.ReviewerID,
				})
			}
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, rating.CreatedAt, 1, rating.CreatedAt.Format(time.RFC3339Nano))
			return result, nil
		case transition.CommandProblemLineRecordImprovementNote:
			note, err := s.addImprovementNoteLocked(command.AggregateID, improvementNoteFromCommand(command), timeFromCommand(command, "created_at", command.OccurredAt))
			if err != nil {
				return commandApplyResult{}, err
			}
			result.traceID = note.TraceID
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(result.bundle.Events[0].Payload, map[string]any{
					"trace_id":        note.TraceID,
					"category":        note.Category,
					"suggested_owner": note.SuggestedOwner,
					"created_by":      note.CreatedBy,
				})
			}
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, note.CreatedAt, 1, note.CreatedAt.Format(time.RFC3339Nano))
			return result, nil
		case transition.CommandProblemLineScheduleReplay:
			result.traceID = command.AggregateID
			if len(result.bundle.Events) > 0 {
				result.bundle.Events[0].Payload = mergeCommandMetadataPayload(result.bundle.Events[0].Payload, map[string]any{
					"trace_id":     command.AggregateID,
					"requested_by": firstNonEmpty(stringFromCommand(command, "requested_by"), command.Actor),
				})
			}
			result.receipt = buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, 1, command.AggregateID)
			return result, nil
		default:
			return commandApplyResult{}, fmt.Errorf("unsupported problem line command kind %s", command.CommandKind)
		}
	case transition.DecisionNoop:
		result.receipt = buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, 1, command.CommandID)
		return result, nil
	case transition.DecisionReject:
		result.receipt = buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, 1, "")
		return result, nil
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported problem line decision kind %s", decision.DecisionKind)
	}
}

func (s *MemoryStore) applyActionCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	intent, ok := s.actionIntents[command.AggregateID]
	snapshot := transition.ActionExecutionSnapshot{}
	if ok {
		snapshot.State = intent.Status
		snapshot.Kind = intent.Kind
	}
	decision := transition.ReduceActionExecution(snapshot, command)
	var resultRef string
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		switch transition.ActionExecutionCommandKind(command.CommandKind) {
		case transition.CommandActionQueue:
			if !ok {
				created, err := actionIntentFromCommand(command)
				if err != nil {
					return commandApplyResult{}, err
				}
				next, err := s.upsertActionIntentLocked(created)
				if err != nil {
					return commandApplyResult{}, err
				}
				intent = next
				ok = true
			}
			resultRef = intent.ID
		case transition.CommandActionStart, transition.CommandActionCancel:
			if !ok {
				return commandApplyResult{}, errors.New("action intent not found")
			}
			next, err := s.setActionExecutionStateLocked(command.AggregateID, decision.NextState, command)
			if err != nil {
				return commandApplyResult{}, err
			}
			intent = next
			if err := s.projectActionTraceLocked(intent, command, decision.NextState); err != nil {
				return commandApplyResult{}, err
			}
			resultRef = intent.ID
		case transition.CommandActionSucceed, transition.CommandActionBlock, transition.CommandActionFail:
			if !ok {
				return commandApplyResult{}, errors.New("action intent not found")
			}
			if !boolFromCommand(command, "record_result", true) {
				next, err := s.setActionExecutionStateLocked(command.AggregateID, decision.NextState, command)
				if err != nil {
					return commandApplyResult{}, err
				}
				intent = next
				if err := s.projectActionTraceLocked(intent, command, decision.NextState); err != nil {
					return commandApplyResult{}, err
				}
				resultRef = intent.ID
				break
			}
			previous := intent
			intent.OperationID = firstNonEmpty(stringFromCommand(command, "operation_id"), intent.OperationID)
			intent.ApprovalState = firstNonEmpty(stringFromCommand(command, "approval_state"), intent.ApprovalState)
			intent.PolicyVerdict = firstNonEmpty(stringFromCommand(command, "policy_verdict"), intent.PolicyVerdict)
			s.actionIntents[intent.ID] = intent
			recorded, err := s.recordActionResultLocked(actionResultFromCommand(intent, decision.NextState, command))
			if err != nil {
				s.actionIntents[previous.ID] = previous
				return commandApplyResult{}, err
			}
			intent = s.actionIntents[recorded.ActionIntentID]
			resultRef = recorded.ID
			if err := s.projectActionTraceLocked(intent, command, decision.NextState); err != nil {
				return commandApplyResult{}, err
			}
		default:
			return commandApplyResult{}, fmt.Errorf("unsupported action command kind %s", command.CommandKind)
		}
	case transition.DecisionReject, transition.DecisionNoop:
		if !ok {
			return commandApplyResult{}, errors.New("action intent not found")
		}
		resultRef = intent.ID
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported action decision kind %s", decision.DecisionKind)
	}
	return commandApplyResult{
		receipt: buildCommandReceipt(command, decision.TransitionDecision, intent.UpdatedAt, 1, firstNonEmpty(resultRef, intent.ID)),
		bundle:  buildCommandBundle(command, decision.TransitionDecision, 1),
	}, nil
}

func (s *MemoryStore) applyHarnessCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	overlay, overlayExists := s.harnessOverlays[command.AggregateID]
	binding, bindingExists := s.harnessSessionBindings[command.AggregateID]
	execution, executionExists := s.findHarnessExecutionLocked(command.AggregateID, stringFromCommand(command, "operation_id"))
	snapshot := transition.HarnessSnapshot{
		SessionBound:  bindingExists,
		ExecutionSeen: executionExists,
	}
	if overlayExists {
		snapshot.OverlayStatus = overlay.Status
	}
	decision := transition.ReduceHarness(snapshot, command)
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		bundle := buildCommandBundle(command, decision.TransitionDecision, 1)
		switch transition.HarnessCommandKind(command.CommandKind) {
		case transition.CommandHarnessActivateOverlay:
			nextOverlay, err := harnessOverlayFromCommand(command)
			if err != nil {
				return commandApplyResult{}, err
			}
			overlay, err = s.upsertHarnessOverlayLocked(nextOverlay)
			if err != nil {
				return commandApplyResult{}, err
			}
			experiment, err := s.recordHarnessExperimentLocked(harnessExperimentFromCommand(command))
			if err != nil {
				return commandApplyResult{}, err
			}
			for idx := range bundle.Events {
				switch bundle.Events[idx].EventKind {
				case "harness_overlay_activated":
					bundle.Events[idx].Payload = mergeCommandMetadataPayload(bundle.Events[idx].Payload, map[string]any{
						"overlay_id": overlay.ID,
						"role":       overlay.Role,
						"version":    overlay.Version,
					})
				case "harness_experiment_recorded":
					bundle.Events[idx].Payload = mergeCommandMetadataPayload(bundle.Events[idx].Payload, map[string]any{
						"experiment_id": experiment.ID,
						"overlay_id":    experiment.OverlayID,
						"attempt_id":    experiment.AttemptID,
					})
				}
			}
			return commandApplyResult{
				receipt:             buildCommandReceipt(command, decision.TransitionDecision, overlay.UpdatedAt, 1, overlay.ID),
				bundle:              bundle,
				harnessOverlayID:    overlay.ID,
				harnessExperimentID: experiment.ID,
			}, nil
		case transition.CommandHarnessBindSession:
			nextBinding, err := harnessSessionBindingFromCommand(command)
			if err != nil {
				return commandApplyResult{}, err
			}
			binding, err = s.upsertHarnessSessionBindingLocked(nextBinding)
			if err != nil {
				return commandApplyResult{}, err
			}
			for idx := range bundle.Events {
				if bundle.Events[idx].EventKind == "harness_session_bound" || bundle.Events[idx].EventKind == "harness_session_refreshed" {
					bundle.Events[idx].Payload = mergeCommandMetadataPayload(bundle.Events[idx].Payload, map[string]any{
						"binding_key":        harnessSessionBindingKey(binding.Role, binding.ScopeKind, binding.ScopeID),
						"role":               binding.Role,
						"scope_kind":         binding.ScopeKind,
						"scope_id":           binding.ScopeID,
						"hermes_session_id":  binding.HermesSessionID,
						"effective_overlay":  binding.EffectiveOverlayVersion,
						"harness_profile_id": binding.HarnessProfileID,
					})
				}
			}
			return commandApplyResult{
				receipt:           buildCommandReceipt(command, decision.TransitionDecision, binding.UpdatedAt, 1, harnessSessionBindingKey(binding.Role, binding.ScopeKind, binding.ScopeID)),
				bundle:            bundle,
				harnessBindingKey: harnessSessionBindingKey(binding.Role, binding.ScopeKind, binding.ScopeID),
			}, nil
		case transition.CommandHarnessRecordExecution:
			nextExecution, err := harnessExecutionFromCommand(command)
			if err != nil {
				return commandApplyResult{}, err
			}
			execution, err = s.recordHarnessExecutionLocked(nextExecution)
			if err != nil {
				return commandApplyResult{}, err
			}
			for idx := range bundle.Events {
				if bundle.Events[idx].EventKind == "harness_execution_recorded" {
					bundle.Events[idx].Payload = mergeCommandMetadataPayload(bundle.Events[idx].Payload, map[string]any{
						"execution_id":       execution.ID,
						"operation_id":       execution.OperationID,
						"trace_id":           execution.TraceID,
						"proposal_id":        execution.ProposalID,
						"role":               execution.Role,
						"session_scope_kind": execution.SessionScopeKind,
						"session_scope_id":   execution.SessionScopeID,
					})
				}
			}
			return commandApplyResult{
				receipt:            buildCommandReceipt(command, decision.TransitionDecision, execution.CreatedAt, 1, execution.ID),
				bundle:             bundle,
				harnessExecutionID: execution.ID,
			}, nil
		default:
			return commandApplyResult{}, fmt.Errorf("unsupported harness command kind %s", command.CommandKind)
		}
	case transition.DecisionNoop, transition.DecisionReject:
		return commandApplyResult{
			receipt: buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, 1, strings.TrimSpace(command.AggregateID)),
			bundle:  buildCommandBundle(command, decision.TransitionDecision, 1),
		}, nil
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported harness decision kind %s", decision.DecisionKind)
	}
}

func (s *MemoryStore) findHarnessExecutionLocked(aggregateID string, operationID string) (harness.Execution, bool) {
	aggregateID = strings.TrimSpace(aggregateID)
	operationID = strings.TrimSpace(operationID)
	for _, item := range s.harnessExecutions {
		if aggregateID != "" && strings.TrimSpace(item.ID) == aggregateID {
			return item, true
		}
		if operationID != "" && strings.TrimSpace(item.OperationID) == operationID {
			return item, true
		}
	}
	return harness.Execution{}, false
}

func actionIntentFromCommand(command transition.CommandEnvelope) (action.Intent, error) {
	actionID := strings.TrimSpace(command.AggregateID)
	if actionID == "" {
		return action.Intent{}, errors.New("action aggregate id is required")
	}
	kind := action.Kind(strings.TrimSpace(stringFromCommand(command, "kind")))
	if kind == "" {
		return action.Intent{}, errors.New("action kind is required")
	}
	return action.Intent{
		ID:             actionID,
		OwnerPlane:     firstNonEmpty(stringFromCommand(command, "owner_plane"), string(command.MachineKind)),
		ConversationID: stringFromCommand(command, "conversation_id"),
		CaseID:         stringFromCommand(command, "case_id"),
		TraceID:        stringFromCommand(command, "trace_id"),
		ProposalID:     stringFromCommand(command, "proposal_id"),
		AttemptID:      stringFromCommand(command, "attempt_id"),
		Kind:           kind,
		PhaseKey:       stringFromCommand(command, "phase_key"),
		TargetRef:      stringFromCommand(command, "target_ref"),
		RequestPayload: anyMapFromCommand(command, "request_payload"),
		IdempotencyKey: stringFromCommand(command, "idempotency_key"),
		ApprovalMode:   stringFromCommand(command, "approval_mode"),
		ApprovalState:  stringFromCommand(command, "approval_state"),
		PolicyVerdict:  stringFromCommand(command, "policy_verdict"),
		Status:         action.StatusQueued,
		RequestedBy:    firstNonEmpty(stringFromCommand(command, "requested_by"), command.Actor),
		Rationale:      stringFromCommand(command, "rationale"),
		EvidenceRefs:   evidenceRefsFromCommand(command, "evidence_refs"),
		CreatedAt:      command.OccurredAt,
		UpdatedAt:      command.OccurredAt,
	}, nil
}

func normalizeCommandEnvelope(command transition.CommandEnvelope) transition.CommandEnvelope {
	command.AggregateID = strings.TrimSpace(command.AggregateID)
	command.CommandKind = strings.TrimSpace(command.CommandKind)
	command.CommandID = strings.TrimSpace(command.CommandID)
	command.CausationID = strings.TrimSpace(command.CausationID)
	command.Actor = strings.TrimSpace(command.Actor)
	if command.OccurredAt.IsZero() {
		command.OccurredAt = time.Now().UTC()
	}
	if command.Payload == nil {
		command.Payload = map[string]any{}
	}
	return command
}

func (s *MemoryStore) applyThreadPolicyCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	item, ok := s.threadPolicies[command.AggregateID]
	if !ok {
		return commandApplyResult{}, errors.New("thread policy not found")
	}
	decision := transition.ReduceThreadPolicy(transition.ThreadPolicySnapshot{State: item.State}, command)
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		next, err := s.setThreadStateLocked(command.AggregateID, decision.NextState, command.Actor)
		if err != nil {
			return commandApplyResult{}, err
		}
		item = next
	case transition.DecisionReject, transition.DecisionNoop:
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported thread policy decision kind %s", decision.DecisionKind)
	}
	return commandApplyResult{
		receipt:   buildCommandReceipt(command, decision.TransitionDecision, item.UpdatedAt, 1, item.ThreadKey),
		bundle:    buildCommandBundle(command, decision.TransitionDecision, 1),
		threadKey: item.ThreadKey,
	}, nil
}

func (s *MemoryStore) applySettingsCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	current := normalizedSettings(s.settings)
	decision := transition.ReduceSettings(transition.SettingsSnapshot{ActiveProposalCap: current.ActiveProposalCap}, command)
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		next, err := s.updateSettingsLocked(improvement.Settings{ActiveProposalCap: decision.NextCap})
		if err != nil {
			return commandApplyResult{}, err
		}
		current = next
	case transition.DecisionReject, transition.DecisionNoop:
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported settings decision kind %s", decision.DecisionKind)
	}
	return commandApplyResult{
		receipt: buildCommandReceipt(command, decision.TransitionDecision, current.UpdatedAt, 1, "settings"),
		bundle:  buildCommandBundle(command, decision.TransitionDecision, 1),
	}, nil
}

func (s *MemoryStore) applyKnowledgeCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	entry, ok := s.knowledgeEntries[command.AggregateID]
	decision := transition.ReduceKnowledge(transition.KnowledgeSnapshot{Exists: ok, Status: entry.Status}, command)
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		switch transition.KnowledgeCommandKind(command.CommandKind) {
		case transition.CommandKnowledgeRecordDraft:
			nextEntry, links, err := knowledgeEntryFromCommand(command)
			if err != nil {
				return commandApplyResult{}, err
			}
			next, err := s.upsertKnowledgeEntryLocked(nextEntry, links)
			if err != nil {
				return commandApplyResult{}, err
			}
			entry = next
		default:
			reviewItem := knowledge.Review{
				ID:         strings.TrimSpace(stringFromCommand(command, "review_id")),
				Decision:   knowledgeDecisionFromCommand(command.CommandKind),
				ReviewerID: firstNonEmpty(command.Actor, stringFromCommand(command, "reviewer_id")),
				Rationale:  stringFromCommand(command, "rationale"),
				CreatedAt:  command.OccurredAt,
			}
			next, err := s.reviewKnowledgeEntryLocked(command.AggregateID, reviewItem)
			if err != nil {
				return commandApplyResult{}, err
			}
			entry = next
		}
	case transition.DecisionReject, transition.DecisionNoop:
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported knowledge decision kind %s", decision.DecisionKind)
	}
	return commandApplyResult{
		receipt:     buildCommandReceipt(command, decision.TransitionDecision, entry.UpdatedAt, 1, entry.ID),
		bundle:      buildCommandBundle(command, decision.TransitionDecision, 1),
		knowledgeID: entry.ID,
	}, nil
}

func (s *MemoryStore) applyProposalCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	proposal, ok := s.proposals[command.AggregateID]
	if !ok {
		return commandApplyResult{}, errors.New("proposal not found")
	}
	decision := transition.ReduceProposalLine(transition.ProposalLineSnapshot{
		State:            proposal.Status,
		InterventionKind: proposal.RecommendedInterventionKind,
	}, command)
	attemptID := strings.TrimSpace(proposal.CurrentAttemptID)
	traceID := ""
	var nextAttemptCommand *transition.CommandEnvelope
	suppressResumeFollowOn := false
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		switch transition.ProposalLineCommandKind(command.CommandKind) {
		case transition.CommandProposalApproveIntervention, transition.CommandProposalRejectLine, transition.CommandProposalDismissLine:
			item, err := s.reviewProposalLocked(command.AggregateID, review.ProposalReview{
				IdempotencyKey: stringFromCommand(command, "idempotency_key"),
				Decision:       proposalDecisionFromCommand(command.CommandKind),
				Scope:          reviewScopeFromCommand(command),
				Rationale:      stringFromCommand(command, "rationale"),
				ReviewerID:     firstNonEmpty(command.Actor, stringFromCommand(command, "reviewer_id")),
				FailureClass:   stringFromCommand(command, "failure_class"),
				FailureClasses: stringSliceFromCommand(command, "failure_classes"),
				CreatedAt:      command.OccurredAt,
			})
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalMarkRepoChangeQueued:
			item, err := s.updateProposalStatusLocked(command.AggregateID, review.ProposalRepoChangeQueued)
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalMarkRepoChangeRunning:
			item, err := s.updateProposalStatusLocked(command.AggregateID, review.ProposalRepoChangeRunning)
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalMarkValidationPending:
			item, err := s.updateProposalStatusLocked(command.AggregateID, review.ProposalValidationPending)
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalMarkFailedValidation:
			item, err := s.updateProposalStatusLocked(command.AggregateID, review.ProposalFailedValidation)
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalMarkPROpen:
			item, err := s.updateProposalStatusLocked(command.AggregateID, review.ProposalPROpen)
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalMarkMerged:
			item, err := s.reviewProposalLocked(command.AggregateID, review.ProposalReview{
				IdempotencyKey: stringFromCommand(command, "idempotency_key"),
				Decision:       string(review.ProposalMerged),
				Scope:          reviewScopeFromCommand(command),
				Rationale:      stringFromCommand(command, "rationale"),
				ReviewerID:     firstNonEmpty(command.Actor, stringFromCommand(command, "reviewer_id")),
				CreatedAt:      command.OccurredAt,
			})
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalRetryableFailure:
			item, err := s.updateProposalStatusLocked(command.AggregateID, review.ProposalApproved)
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalNeedsReview:
			item, err := s.updateProposalStatusLocked(command.AggregateID, review.ProposalPendingReview)
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalStopLine:
			item, err := s.stopProposalLineLocked(command.AggregateID, command.Actor, stringFromCommand(command, "rationale"))
			if err != nil {
				return commandApplyResult{}, err
			}
			proposal = item
		case transition.CommandProposalRetryAttempt:
			if proposal.Status == review.ProposalFailedValidation {
				item, err := s.updateProposalStatusLocked(command.AggregateID, review.ProposalApproved)
				if err != nil {
					return commandApplyResult{}, err
				}
				proposal = item
			}
			if nextProposal, nextAttempt, nextTraceID, nextPhaseCommand, handled, err := s.materializeApprovedProposalAttemptLocked(command, proposal); err != nil {
				return commandApplyResult{}, err
			} else if handled {
				proposal = nextProposal
				attemptID = nextAttempt.ID
				traceID = nextTraceID
				nextAttemptCommand = attemptFollowOnCommand(command, nextProposal, nextAttempt, nextTraceID, nextPhaseCommand)
				suppressResumeFollowOn = true
				break
			}
			proposal = s.proposals[command.AggregateID]
			attemptID = strings.TrimSpace(proposal.CurrentAttemptID)
			suppressResumeFollowOn = true
		case transition.CommandProposalResumeExecution:
			if nextProposal, nextAttempt, nextTraceID, nextPhaseCommand, handled, err := s.materializeApprovedProposalAttemptLocked(command, proposal); err != nil {
				return commandApplyResult{}, err
			} else if handled {
				proposal = nextProposal
				attemptID = nextAttempt.ID
				traceID = nextTraceID
				nextAttemptCommand = attemptFollowOnCommand(command, nextProposal, nextAttempt, nextTraceID, nextPhaseCommand)
				break
			}
			proposal = s.proposals[command.AggregateID]
			attemptID = strings.TrimSpace(proposal.CurrentAttemptID)
		default:
			return commandApplyResult{}, fmt.Errorf("unsupported proposal command kind %s", command.CommandKind)
		}
	case transition.DecisionReject, transition.DecisionNoop:
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported proposal decision kind %s", decision.DecisionKind)
	}
	result := commandApplyResult{
		receipt:    buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, proposal.Version, proposal.ID),
		bundle:     buildCommandBundle(command, decision.TransitionDecision, proposal.Version),
		proposalID: proposal.ID,
		attemptID:  firstNonEmpty(attemptID, strings.TrimSpace(proposal.CurrentAttemptID)),
		traceID:    traceID,
	}
	if nextAttemptCommand != nil {
		appendFollowOnCommand(&result.bundle, command, *nextAttemptCommand, "proposal materialized attempt and queued the first executable attempt command")
	}
	if suppressResumeFollowOn {
		result.bundle.Commands = filterCommandEnvelopes(result.bundle.Commands, func(candidate transition.CommandEnvelope) bool {
			return candidate.MachineKind == transition.MachineProposalLine &&
				candidate.AggregateID == proposal.ID &&
				candidate.CommandKind == string(transition.CommandProposalResumeExecution)
		})
	}
	return result, nil
}

func (s *MemoryStore) materializeApprovedProposalAttemptLocked(command transition.CommandEnvelope, proposal review.Proposal) (review.Proposal, improvement.ChangeAttempt, string, transition.AttemptPhaseCommandKind, bool, error) {
	if proposal.Status != review.ProposalApproved {
		return review.Proposal{}, improvement.ChangeAttempt{}, "", "", false, nil
	}
	if !review.ProposalExecutableIntervention(proposal.RecommendedInterventionKind) {
		return review.Proposal{}, improvement.ChangeAttempt{}, "", "", false, nil
	}
	now := command.OccurredAt
	parentAttemptID := strings.TrimSpace(proposal.CurrentAttemptID)
	if parentAttemptID != "" {
		if current, ok := s.changeAttempts[parentAttemptID]; ok && !isTerminalAttemptState(current.State) {
			return review.Proposal{}, improvement.ChangeAttempt{}, "", "", false, nil
		}
	}
	s.supersedeNonCurrentActiveAttemptsLocked(proposal.ID, parentAttemptID, now)

	nextAttemptNumber := maxInt(1, proposal.AttemptCount+1)
	trigger := proposalAttemptTriggerFromCommand(command)
	attempt := normalizeChangeAttempt(improvement.ChangeAttempt{
		ID:              fmt.Sprintf("attempt-%s-%02d", strings.ReplaceAll(strings.TrimSpace(proposal.ID), "/", "-"), nextAttemptNumber),
		ProposalID:      proposal.ID,
		CandidateKey:    proposal.CandidateKey,
		AttemptNumber:   nextAttemptNumber,
		TargetLayer:     proposal.TargetLayer,
		TargetKind:      proposal.TargetKind,
		TargetRef:       proposal.TargetRef,
		Trigger:         trigger,
		State:           improvement.AttemptStatePatchPlan,
		ParentAttemptID: parentAttemptID,
		BranchName:      fmt.Sprintf("codex/%s/attempt-%02d", proposal.ID, nextAttemptNumber),
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		attempt.State = improvement.AttemptStateOverlayPlan
	}

	traceReq := DerivedTraceRequest{
		SourceTraceID:  firstNonEmpty(proposal.TraceID, proposal.OriginTraceID),
		ProposalID:     proposal.ID,
		AttemptID:      attempt.ID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		ThreadKey:      fmt.Sprintf("proposal:%s", proposal.ID),
		WorkflowKind:   "proposal_attempt",
		RequestedBy:    firstNonEmpty(command.Actor, "formal-transition"),
		Description:    fmt.Sprintf("Queued remediation attempt %d for proposal %s triggered by %s.", nextAttemptNumber, proposal.ID, trigger),
		TriggerEventID: proposal.OriginTraceID,
		CreatedAt:      now,
	}
	createdTrace, _, err := s.createDerivedTraceLocked(traceReq)
	if err != nil {
		return review.Proposal{}, improvement.ChangeAttempt{}, "", "", false, err
	}
	attempt.AttemptTraceID = createdTrace.Summary.TraceID
	recordedAttempt, err := s.upsertChangeAttemptLocked(attempt)
	if err != nil {
		return review.Proposal{}, improvement.ChangeAttempt{}, "", "", false, err
	}
	proposal = s.proposals[proposal.ID]

	nextCommand := transition.CommandAttemptPlannedWorkspace
	if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
		nextCommand = transition.CommandAttemptPlannedImplement
	}
	return proposal, recordedAttempt, createdTrace.Summary.TraceID, nextCommand, true, nil
}

func (s *MemoryStore) supersedeNonCurrentActiveAttemptsLocked(proposalID string, keepAttemptID string, now time.Time) {
	proposalID = strings.TrimSpace(proposalID)
	keepAttemptID = strings.TrimSpace(keepAttemptID)
	if proposalID == "" {
		return
	}

	staleAttemptIDs := make(map[string]struct{})
	for attemptID, item := range s.changeAttempts {
		if strings.TrimSpace(item.ProposalID) != proposalID || attemptID == keepAttemptID || isTerminalAttemptState(item.State) {
			continue
		}
		item.State = improvement.AttemptStateSuperseded
		item.FailureClass = firstNonEmpty(item.FailureClass, "superseded_by_new_attempt")
		item.FailureSummary = firstNonEmpty(item.FailureSummary, fmt.Sprintf("Attempt superseded while resuming proposal %s.", proposalID))
		item.RetryDecision = firstNonEmpty(item.RetryDecision, "superseded")
		item.RetryAfter = nil
		item.UpdatedAt = now
		if item.Version == 0 {
			item.Version = 1
		} else {
			item.Version++
		}
		s.changeAttempts[attemptID] = normalizeChangeAttempt(item)
		staleAttemptIDs[attemptID] = struct{}{}
	}

	if len(staleAttemptIDs) == 0 {
		return
	}
}

func proposalAttemptTriggerFromCommand(command transition.CommandEnvelope) improvement.ChangeAttemptTrigger {
	if raw := strings.TrimSpace(stringFromCommand(command, "trigger")); raw != "" {
		return improvement.ChangeAttemptTrigger(raw)
	}
	switch strings.TrimSpace(stringFromCommand(command, "failure_class")) {
	case "ci_regression":
		return improvement.AttemptTriggerCIFailed
	case "closed_unmerged":
		return improvement.AttemptTriggerPRClosed
	case "sandbox_failure":
		return improvement.AttemptTriggerSandboxFailed
	default:
		return improvement.AttemptTriggerProposalApproved
	}
}

func filterCommandEnvelopes(items []transition.CommandEnvelope, drop func(transition.CommandEnvelope) bool) []transition.CommandEnvelope {
	if len(items) == 0 {
		return nil
	}
	filtered := make([]transition.CommandEnvelope, 0, len(items))
	for _, item := range items {
		if drop(item) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func (s *MemoryStore) applyAttemptCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	attempt, ok := s.changeAttempts[command.AggregateID]
	if !ok {
		return commandApplyResult{}, errors.New("change attempt not found")
	}
	proposal, ok := s.proposals[attempt.ProposalID]
	if !ok {
		return commandApplyResult{}, errors.New("proposal not found for attempt")
	}
	snapshot := transition.AttemptSnapshot{
		ProposalStatus:       transitionProposalStatusOr(&proposal),
		AttemptState:         transitionAttemptStateOr(&attempt),
		CurrentOperationKind: "",
	}
	decision := transition.ReduceAttempt(snapshot, command)
	if decision.DecisionKind == transition.DecisionReject {
		decision.Reason = fmt.Sprintf("%s (proposal=%s attempt=%s current_operation=%s)", decision.Reason, snapshot.ProposalStatus, snapshot.AttemptState, firstNonEmpty(snapshot.CurrentOperationKind, "<none>"))
	}
	switch decision.DecisionKind {
	case transition.DecisionAdvance:
		next := attempt
		next.UpdatedAt = command.OccurredAt
		switch transition.AttemptPhaseCommandKind(command.CommandKind) {
		case transition.CommandAttemptPlannedWorkspace, transition.CommandAttemptPlannedImplement:
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandWorkspaceOpenDeferred:
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandWorkspaceReady:
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandWorkspaceMetadataSynced,
			transition.CommandWorkspaceToolValidationStarted,
			transition.CommandWorkspaceToolValidationCompleted,
			transition.CommandWorkspaceToolValidationFailed,
			transition.CommandAttemptRunnerStarted,
			transition.CommandAttemptRunnerCompleted:
		case transition.CommandImplementationDeferred:
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandImplementationCompleted:
			if next.State == improvement.AttemptStateOverlayPlan {
				next.State = improvement.AttemptStateOverlayGenerated
			} else {
				next.State = improvement.AttemptStatePatchGenerated
			}
			next.ChangePlan = firstNonEmpty(stringFromCommand(command, "change_plan"), next.ChangePlan)
			next.RepoPatch = firstNonEmpty(stringFromCommand(command, "repo_patch"), next.RepoPatch)
			next.ValidationPlan = firstNonEmpty(stringFromCommand(command, "validation_plan"), next.ValidationPlan)
			next.HypothesisDelta = firstNonEmpty(stringFromCommand(command, "hypothesis_delta"), next.HypothesisDelta)
			next.DiffSummary = firstNonEmpty(stringFromCommand(command, "diff_summary"), next.DiffSummary)
			next.ValidationSummary = firstNonEmpty(stringFromCommand(command, "validation_summary"), next.ValidationSummary)
			next.ChangedFiles = firstNonEmptyStringSlice(stringSliceFromCommand(command, "changed_files"), next.ChangedFiles)
			next.HeadSHA = firstNonEmpty(stringFromCommand(command, "head_sha"), next.HeadSHA)
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandValidationStarted:
			next.State = improvement.AttemptStateValidationRunning
			next.ValidationSummary = firstNonEmpty(stringFromCommand(command, "validation_summary"), next.ValidationSummary)
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandValidationCompleted:
			next.State = improvement.AttemptStateValidationRunning
			next.ValidationSummary = firstNonEmpty(stringFromCommand(command, "validation_summary"), next.ValidationSummary)
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandOverlayActivated:
			next.State = improvement.AttemptStateOverlayActive
			next.ChangePlan = firstNonEmpty(stringFromCommand(command, "change_plan"), next.ChangePlan)
			next.ValidationPlan = firstNonEmpty(stringFromCommand(command, "validation_plan"), next.ValidationPlan)
			if payload := anyMapFromCommand(command, "overlay_payload"); len(payload) > 0 {
				next.OverlayPayload = payload
			}
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandWorkspaceFailedRetryable, transition.CommandImplementationFailedRetryable, transition.CommandValidationFailedRetryable, transition.CommandPROpenFailedRetryable:
			next.State = attemptFailureStateFromCommand(command.CommandKind, stringFromCommand(command, "failure_class"), true)
			next.FailureClass = firstNonEmpty(stringFromCommand(command, "failure_class"), next.FailureClass)
			next.FailureSummary = firstNonEmpty(stringFromCommand(command, "failure_summary"), next.FailureSummary)
			next.ValidationSummary = firstNonEmpty(stringFromCommand(command, "validation_summary"), next.ValidationSummary)
			next.RetryDecision = firstNonEmpty(stringFromCommand(command, "retry_decision"), next.RetryDecision)
			next.RetryAfter = optionalTimeFromCommand(command, "retry_after")
			next.MaterialHypothesisChange = boolFromCommand(command, "material_hypothesis_change", false)
		case transition.CommandWorkspaceFailedReview, transition.CommandImplementationFailedReview, transition.CommandValidationFailedReview, transition.CommandPROpenFailedReview:
			next.State = attemptFailureStateFromCommand(command.CommandKind, stringFromCommand(command, "failure_class"), false)
			next.FailureClass = firstNonEmpty(stringFromCommand(command, "failure_class"), next.FailureClass)
			next.FailureSummary = firstNonEmpty(stringFromCommand(command, "failure_summary"), next.FailureSummary)
			next.ValidationSummary = firstNonEmpty(stringFromCommand(command, "validation_summary"), next.ValidationSummary)
			next.RetryDecision = firstNonEmpty(stringFromCommand(command, "retry_decision"), next.RetryDecision)
			next.RetryAfter = nil
			next.MaterialHypothesisChange = boolFromCommand(command, "material_hypothesis_change", false)
		case transition.CommandAttemptPROpened:
			next.State = improvement.AttemptStateCIObserving
			next.PRURL = firstNonEmpty(stringFromCommand(command, "pr_url"), next.PRURL)
			next.HeadSHA = firstNonEmpty(stringFromCommand(command, "head_sha"), next.HeadSHA)
		case transition.CommandAttemptMerged:
			next.State = improvement.AttemptStateMerged
			next.PRURL = firstNonEmpty(stringFromCommand(command, "pr_url"), next.PRURL)
			next.HeadSHA = firstNonEmpty(stringFromCommand(command, "head_sha"), next.HeadSHA)
			next.FailureClass = ""
			next.FailureSummary = ""
			next.RetryDecision = ""
			next.RetryAfter = nil
		case transition.CommandAttemptClosedUnmerged:
			next.State = improvement.AttemptStateClosedUnmerged
			next.PRURL = firstNonEmpty(stringFromCommand(command, "pr_url"), next.PRURL)
			next.HeadSHA = firstNonEmpty(stringFromCommand(command, "head_sha"), next.HeadSHA)
			next.FailureClass = firstNonEmpty(stringFromCommand(command, "failure_class"), "closed_unmerged")
			next.FailureSummary = firstNonEmpty(stringFromCommand(command, "failure_summary"), next.FailureSummary)
			next.RetryDecision = firstNonEmpty(stringFromCommand(command, "retry_decision"), next.RetryDecision)
			next.RetryAfter = optionalTimeFromCommand(command, "retry_after")
			next.MaterialHypothesisChange = boolFromCommand(command, "material_hypothesis_change", false)
		case transition.CommandAttemptCIFailed:
			next.State = improvement.AttemptStateCIFailed
			next.PRURL = firstNonEmpty(stringFromCommand(command, "pr_url"), next.PRURL)
			next.HeadSHA = firstNonEmpty(stringFromCommand(command, "head_sha"), next.HeadSHA)
			next.FailureClass = firstNonEmpty(stringFromCommand(command, "failure_class"), "ci_regression")
			next.FailureSummary = firstNonEmpty(stringFromCommand(command, "failure_summary"), next.FailureSummary)
			next.RetryDecision = firstNonEmpty(stringFromCommand(command, "retry_decision"), next.RetryDecision)
			next.RetryAfter = optionalTimeFromCommand(command, "retry_after")
			next.MaterialHypothesisChange = boolFromCommand(command, "material_hypothesis_change", false)
		default:
			return commandApplyResult{}, fmt.Errorf("unsupported attempt command kind %s", command.CommandKind)
		}
		updated, err := s.upsertChangeAttemptLocked(next)
		if err != nil {
			return commandApplyResult{}, err
		}
		attempt = updated
		if err := s.applyRepoChangeJobForAttemptCommandLocked(proposal, attempt, command); err != nil {
			return commandApplyResult{}, err
		}
		workspaceID, err := s.applyAttemptWorkspaceForCommandLocked(attempt, command)
		if err != nil {
			return commandApplyResult{}, err
		}
		prAttemptID := ""
		if transition.AttemptPhaseCommandKind(command.CommandKind) == transition.CommandAttemptPROpened {
			recorded, err := s.recordPRAttemptLocked(prAttemptFromAttemptCommand(proposal, attempt, command))
			if err != nil {
				return commandApplyResult{}, err
			}
			prAttemptID = recorded.ID
		}
		if err := s.projectAttemptTraceLocked(attempt, command); err != nil {
			return commandApplyResult{}, err
		}
		return commandApplyResult{
			receipt:     buildCommandReceipt(command, decision.TransitionDecision, attempt.UpdatedAt, attempt.Version, attempt.ID),
			bundle:      buildCommandBundle(command, decision.TransitionDecision, attempt.Version),
			attemptID:   attempt.ID,
			proposalID:  attempt.ProposalID,
			traceID:     attemptTraceProjectionID(command, attempt),
			prAttemptID: prAttemptID,
			workspaceID: workspaceID,
		}, nil
	case transition.DecisionReject, transition.DecisionNoop:
	default:
		return commandApplyResult{}, fmt.Errorf("unsupported attempt decision kind %s", decision.DecisionKind)
	}
	return commandApplyResult{
		receipt:    buildCommandReceipt(command, decision.TransitionDecision, attempt.UpdatedAt, attempt.Version, attempt.ID),
		bundle:     buildCommandBundle(command, decision.TransitionDecision, attempt.Version),
		attemptID:  attempt.ID,
		proposalID: attempt.ProposalID,
		traceID:    attemptTraceProjectionID(command, attempt),
	}, nil
}

func (s *MemoryStore) applyRepoChangeJobForAttemptCommandLocked(proposal review.Proposal, attempt improvement.ChangeAttempt, command transition.CommandEnvelope) error {
	job, ok := repoChangeJobForAttemptCommandLocked(s.repoChangeJobs, attempt.ProposalID, attempt.ID, command)
	updated := job
	if !ok {
		switch transition.AttemptPhaseCommandKind(command.CommandKind) {
		case transition.CommandWorkspaceOpenDeferred, transition.CommandWorkspaceReady:
			created, createOK := repoChangeJobFromAttemptCommand(proposal, attempt, command)
			if !createOK {
				return nil
			}
			updated = created
		default:
			return nil
		}
	}
	updated.UpdatedAt = command.OccurredAt
	switch transition.AttemptPhaseCommandKind(command.CommandKind) {
	case transition.CommandWorkspaceOpenDeferred:
		updated.Status = string(review.ProposalRepoChangeQueued)
		updated.Repo = firstNonEmpty(stringFromCommand(command, "repo"), updated.Repo)
		updated.BaseRef = firstNonEmpty(stringFromCommand(command, "base_ref"), updated.BaseRef)
		updated.BranchName = firstNonEmpty(stringFromCommand(command, "branch_name"), updated.BranchName)
		if globs := stringSliceFromCommand(command, "allowed_path_globs"); len(globs) > 0 {
			updated.AllowedPathGlobs = globs
		}
		updated.SandboxNamespace = firstNonEmpty(stringFromCommand(command, "sandbox_namespace"), updated.SandboxNamespace)
		updated.SandboxJobName = firstNonEmpty(stringFromCommand(command, "sandbox_job_name"), updated.SandboxJobName)
		updated.SandboxPodName = firstNonEmpty(stringFromCommand(command, "sandbox_pod_name"), updated.SandboxPodName)
		updated.ValidationRef = firstNonEmpty(stringFromCommand(command, "validation_ref"), updated.ValidationRef)
		updated.ValidationError = ""
	case transition.CommandWorkspaceReady:
		updated.Status = string(review.ProposalRepoChangeRunning)
		updated.Repo = firstNonEmpty(stringFromCommand(command, "repo"), updated.Repo)
		updated.BaseRef = firstNonEmpty(stringFromCommand(command, "base_ref"), updated.BaseRef)
		updated.BranchName = firstNonEmpty(stringFromCommand(command, "branch_name"), updated.BranchName)
		if globs := stringSliceFromCommand(command, "allowed_path_globs"); len(globs) > 0 {
			updated.AllowedPathGlobs = globs
		}
		updated.SandboxNamespace = firstNonEmpty(stringFromCommand(command, "sandbox_namespace"), updated.SandboxNamespace)
		updated.SandboxJobName = firstNonEmpty(stringFromCommand(command, "sandbox_job_name"), updated.SandboxJobName)
		updated.SandboxPodName = firstNonEmpty(stringFromCommand(command, "sandbox_pod_name"), updated.SandboxPodName)
		updated.ValidationRef = firstNonEmpty(stringFromCommand(command, "validation_ref"), updated.ValidationRef)
		updated.ValidationError = ""
	case transition.CommandValidationStarted:
		updated.Status = string(review.ProposalRepoChangeRunning)
		updated.SandboxNamespace = firstNonEmpty(stringFromCommand(command, "sandbox_namespace"), updated.SandboxNamespace)
		updated.SandboxJobName = firstNonEmpty(stringFromCommand(command, "sandbox_job_name"), updated.SandboxJobName)
		updated.SandboxPodName = firstNonEmpty(stringFromCommand(command, "sandbox_pod_name"), updated.SandboxPodName)
		updated.ValidationRef = firstNonEmpty(stringFromCommand(command, "validation_ref"), updated.ValidationRef)
		updated.ValidationError = ""
	case transition.CommandValidationCompleted:
		updated.Status = string(review.ProposalValidationPending)
		updated.SandboxNamespace = firstNonEmpty(stringFromCommand(command, "sandbox_namespace"), updated.SandboxNamespace)
		updated.SandboxJobName = firstNonEmpty(stringFromCommand(command, "sandbox_job_name"), updated.SandboxJobName)
		updated.SandboxPodName = firstNonEmpty(stringFromCommand(command, "sandbox_pod_name"), updated.SandboxPodName)
		updated.ValidationRef = firstNonEmpty(stringFromCommand(command, "validation_ref"), updated.ValidationRef)
		updated.LogArtifactID = firstNonEmpty(stringFromCommand(command, "log_artifact_id"), updated.LogArtifactID)
		updated.ValidationError = ""
	case transition.CommandValidationFailedRetryable, transition.CommandValidationFailedReview:
		updated.Status = string(review.ProposalFailedValidation)
		updated.SandboxNamespace = firstNonEmpty(stringFromCommand(command, "sandbox_namespace"), updated.SandboxNamespace)
		updated.SandboxJobName = firstNonEmpty(stringFromCommand(command, "sandbox_job_name"), updated.SandboxJobName)
		updated.SandboxPodName = firstNonEmpty(stringFromCommand(command, "sandbox_pod_name"), updated.SandboxPodName)
		updated.ValidationRef = firstNonEmpty(stringFromCommand(command, "validation_ref"), updated.ValidationRef)
		updated.LogArtifactID = firstNonEmpty(stringFromCommand(command, "log_artifact_id"), updated.LogArtifactID)
		updated.ValidationError = firstNonEmpty(stringFromCommand(command, "validation_error"), stringFromCommand(command, "failure_summary"), updated.ValidationError)
	case transition.CommandAttemptPROpened:
		updated.Status = string(review.ProposalPROpen)
	default:
		return nil
	}
	_, err := s.upsertRepoChangeJobLocked(updated)
	return err
}

func (s *MemoryStore) applyAttemptWorkspaceForCommandLocked(attempt improvement.ChangeAttempt, command transition.CommandEnvelope) (string, error) {
	workspace, ok := attemptWorkspaceForCommandLocked(s.attemptWorkspaces, attempt.ID, stringFromCommand(command, "workspace_id"))
	createdWorkspace := false
	if !ok {
		created, createOK := attemptWorkspaceFromCommand(attempt, command)
		if !createOK {
			return "", nil
		}
		workspace = created
		createdWorkspace = true
	}
	updated := workspace
	touched := createdWorkspace
	if namespace := firstNonEmpty(stringFromCommand(command, "workspace_namespace"), stringFromCommand(command, "sandbox_namespace")); namespace != "" && namespace != updated.Namespace {
		updated.Namespace = namespace
		touched = true
	}
	if jobName := firstNonEmpty(stringFromCommand(command, "workspace_job_name"), stringFromCommand(command, "sandbox_job_name")); jobName != "" && jobName != updated.JobName {
		updated.JobName = jobName
		touched = true
	}
	if podName := firstNonEmpty(stringFromCommand(command, "workspace_pod_name"), stringFromCommand(command, "sandbox_pod_name")); podName != "" && podName != updated.PodName {
		updated.PodName = podName
		touched = true
	}
	if headSHA := stringFromCommand(command, "head_sha"); headSHA != "" && headSHA != updated.HeadSHA {
		updated.HeadSHA = headSHA
		touched = true
	}
	if diffSummary := stringFromCommand(command, "diff_summary"); diffSummary != "" && diffSummary != updated.DiffSummary {
		updated.DiffSummary = diffSummary
		touched = true
	}
	if repo := stringFromCommand(command, "repo"); repo != "" && repo != updated.Repo {
		updated.Repo = repo
		touched = true
	}
	if baseRef := stringFromCommand(command, "base_ref"); baseRef != "" && baseRef != updated.BaseRef {
		updated.BaseRef = baseRef
		touched = true
	}
	if branchName := stringFromCommand(command, "branch_name"); branchName != "" && branchName != updated.BranchName {
		updated.BranchName = branchName
		touched = true
	}
	if globs := stringSliceFromCommand(command, "allowed_path_globs"); len(globs) > 0 {
		updated.AllowedPathGlobs = globs
		touched = true
	}
	switch transition.AttemptPhaseCommandKind(command.CommandKind) {
	case transition.CommandWorkspaceOpenDeferred:
		if updated.Status != improvement.WorkspaceQueued {
			updated.Status = improvement.WorkspaceQueued
			touched = true
		}
	case transition.CommandWorkspaceReady:
		if updated.Status != improvement.WorkspaceReady {
			updated.Status = improvement.WorkspaceReady
			touched = true
		}
	case transition.CommandAttemptRunnerStarted:
		if updated.Status != improvement.WorkspaceExecuting {
			updated.Status = improvement.WorkspaceExecuting
			touched = true
		}
	case transition.CommandWorkspaceToolValidationStarted:
		if updated.Status != improvement.WorkspaceValidating {
			updated.Status = improvement.WorkspaceValidating
			touched = true
		}
	case transition.CommandWorkspaceToolValidationCompleted:
		if updated.Status != improvement.WorkspaceCompleted {
			updated.Status = improvement.WorkspaceCompleted
			touched = true
		}
	case transition.CommandWorkspaceToolValidationFailed:
		if updated.Status != improvement.WorkspaceFailed {
			updated.Status = improvement.WorkspaceFailed
			touched = true
		}
	case transition.CommandValidationStarted:
		if updated.Status != improvement.WorkspaceValidating {
			updated.Status = improvement.WorkspaceValidating
			touched = true
		}
	case transition.CommandValidationCompleted:
		if updated.Status != improvement.WorkspaceCompleted {
			updated.Status = improvement.WorkspaceCompleted
			touched = true
		}
	case transition.CommandWorkspaceFailedRetryable,
		transition.CommandWorkspaceFailedReview,
		transition.CommandImplementationFailedRetryable,
		transition.CommandImplementationFailedReview,
		transition.CommandValidationFailedRetryable,
		transition.CommandValidationFailedReview,
		transition.CommandPROpenFailedRetryable,
		transition.CommandPROpenFailedReview:
		if updated.Status != improvement.WorkspaceFailed {
			updated.Status = improvement.WorkspaceFailed
			touched = true
		}
	}
	if !touched {
		return "", nil
	}
	updated.UpdatedAt = command.OccurredAt
	recorded, err := s.upsertAttemptWorkspaceLocked(updated)
	if err != nil {
		return "", err
	}
	return recorded.ID, nil
}

func attemptWorkspaceFromCommand(attempt improvement.ChangeAttempt, command transition.CommandEnvelope) (improvement.AttemptWorkspace, bool) {
	workspaceID := strings.TrimSpace(stringFromCommand(command, "workspace_id"))
	repo := strings.TrimSpace(stringFromCommand(command, "repo"))
	baseRef := firstNonEmpty(strings.TrimSpace(stringFromCommand(command, "base_ref")), "main")
	branchName := firstNonEmpty(strings.TrimSpace(stringFromCommand(command, "branch_name")), attempt.BranchName)
	namespace := firstNonEmpty(strings.TrimSpace(stringFromCommand(command, "workspace_namespace")), strings.TrimSpace(stringFromCommand(command, "sandbox_namespace")))
	jobName := firstNonEmpty(strings.TrimSpace(stringFromCommand(command, "workspace_job_name")), strings.TrimSpace(stringFromCommand(command, "sandbox_job_name")))
	podName := firstNonEmpty(strings.TrimSpace(stringFromCommand(command, "workspace_pod_name")), strings.TrimSpace(stringFromCommand(command, "sandbox_pod_name")))
	if workspaceID == "" && repo == "" && branchName == "" && namespace == "" && jobName == "" && podName == "" {
		return improvement.AttemptWorkspace{}, false
	}
	now := command.OccurredAt
	if workspaceID == "" {
		workspaceID = fmt.Sprintf("workspace-%s", attempt.ID)
	}
	return improvement.AttemptWorkspace{
		ID:               workspaceID,
		AttemptID:        attempt.ID,
		ProposalID:       attempt.ProposalID,
		Repo:             repo,
		BaseRef:          baseRef,
		BranchName:       branchName,
		Namespace:        namespace,
		JobName:          jobName,
		PodName:          podName,
		AllowedPathGlobs: stringSliceFromCommand(command, "allowed_path_globs"),
		CreatedAt:        now,
		UpdatedAt:        now,
	}, true
}

func firstNonEmptyStringSlice(primary []string, fallback []string) []string {
	if len(primary) > 0 {
		return primary
	}
	return fallback
}

func repoChangeJobFromAttemptCommand(proposal review.Proposal, attempt improvement.ChangeAttempt, command transition.CommandEnvelope) (improvement.RepoChangeJob, bool) {
	jobID := strings.TrimSpace(stringFromCommand(command, "job_id"))
	repo := strings.TrimSpace(stringFromCommand(command, "repo"))
	baseRef := firstNonEmpty(strings.TrimSpace(stringFromCommand(command, "base_ref")), "main")
	branchName := strings.TrimSpace(stringFromCommand(command, "branch_name"))
	if jobID == "" && repo == "" && branchName == "" {
		return improvement.RepoChangeJob{}, false
	}
	now := command.OccurredAt
	if jobID == "" {
		jobID = fmt.Sprintf("job-%s", attempt.ID)
	}
	if branchName == "" {
		branchName = attempt.BranchName
	}
	return improvement.RepoChangeJob{
		ID:             jobID,
		ProposalID:     attempt.ProposalID,
		AttemptID:      attempt.ID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		OriginTraceID:  firstNonEmpty(attempt.AttemptTraceID, proposal.OriginTraceID, proposal.TraceID),
		CandidateKey:   proposal.CandidateKey,
		Status:         string(review.ProposalRepoChangeQueued),
		Repo:           repo,
		BaseRef:        baseRef,
		BranchName:     branchName,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, true
}

func repoChangeJobForAttemptCommandLocked(items map[string]improvement.RepoChangeJob, proposalID string, attemptID string, command transition.CommandEnvelope) (improvement.RepoChangeJob, bool) {
	jobID := strings.TrimSpace(stringFromCommand(command, "job_id"))
	if jobID != "" {
		item, ok := items[jobID]
		return item, ok
	}
	for _, item := range items {
		if item.ProposalID == proposalID && item.AttemptID == attemptID {
			return item, true
		}
	}
	return improvement.RepoChangeJob{}, false
}

func persistCommandMutation(tx *sql.Tx, store *MemoryStore, result commandApplyResult) error {
	switch result.receipt.MachineKind {
	case transition.MachineIngress:
		for _, item := range store.events {
			if item.ID != result.eventID {
				continue
			}
			return replaceEventMaterializationScope(tx, store, item)
		}
		if result.eventID == "" {
			return nil
		}
		return errors.New("event not found for ingress persistence")
	case transition.MachineWorkflow:
		workflow, ok := findWorkflowByID(store.workflows, result.receipt.AggregateID)
		if !ok {
			return errors.New("workflow not found for persistence")
		}
		return replaceWorkflowScope(tx, workflow)
	case transition.MachineAction:
		intent, ok := store.actionIntents[result.receipt.AggregateID]
		if !ok {
			return errors.New("action intent not found for persistence")
		}
		if err := replaceActionIntentScope(tx, intent); err != nil {
			return err
		}
		return replaceActionResultScope(tx, store, intent.ID)
	case transition.MachineProblemLine:
		switch transition.ProblemLineCommandKind(result.receipt.CommandKind) {
		case transition.CommandProblemLineEvaluateTrace:
			if strings.TrimSpace(result.evalRunID) == "" {
				if strings.TrimSpace(result.traceID) != "" {
					trace, ok := store.traces[result.traceID]
					if !ok {
						return errors.New("trace not found for persistence")
					}
					return replaceTraceAndWorkflowScope(tx, store, trace)
				}
				return nil
			}
			run, ok := findEvalRunByID(store.evalRuns, result.evalRunID)
			if !ok {
				return errors.New("eval run not found for persistence")
			}
			if err := replaceEvalRunScope(tx, run, result.evalJudgments); err != nil {
				return err
			}
			if strings.TrimSpace(result.candidateKey) != "" {
				return replaceCandidateScope(tx, store, result.candidateKey)
			}
			return nil
		case transition.CommandProblemLinePromote:
			return replaceProposalPromoterScope(tx, store)
		case transition.CommandProblemLineRecordOutcome:
			if strings.TrimSpace(result.outcomeID) == "" {
				return nil
			}
			record, ok := store.outcomes[result.outcomeID]
			if !ok {
				return errors.New("outcome not found for persistence")
			}
			if err := replaceOutcomeScope(tx, record); err != nil {
				return err
			}
			if strings.TrimSpace(record.CaseID) != "" {
				if err := replaceCaseScope(tx, record, store); err != nil {
					return err
				}
			}
			if strings.TrimSpace(record.ProposalID) != "" {
				if err := replaceProposalScope(tx, store, record.ProposalID); err != nil {
					return err
				}
			}
			return nil
		case transition.CommandProblemLineProjectTrace:
			trace, ok := store.traces[result.traceID]
			if !ok {
				return errors.New("trace not found for persistence")
			}
			return replaceTraceAndWorkflowScope(tx, store, trace)
		case transition.CommandProblemLineRecordFeedback:
			if strings.TrimSpace(result.traceID) == "" {
				return errors.New("feedback trace not found for persistence")
			}
			return replaceFeedbackScope(tx, store, result.traceID)
		case transition.CommandProblemLineScheduleReplay:
			return nil
		default:
			return fmt.Errorf("unsupported problem line command kind %s for persistence", result.receipt.CommandKind)
		}
	case transition.MachineThreadPolicy:
		item, ok := store.threadPolicies[result.threadKey]
		if !ok {
			return errors.New("thread policy not found for persistence")
		}
		return replaceThreadPolicyScope(tx, item)
	case transition.MachineSettings:
		return replaceSettingsScope(tx, store.settings)
	case transition.MachineKnowledge:
		if strings.TrimSpace(result.knowledgeID) == "" {
			return nil
		}
		entry, ok := store.knowledgeEntries[result.knowledgeID]
		if !ok {
			return errors.New("knowledge entry not found for persistence")
		}
		if err := replaceKnowledgeEntryScope(tx, entry); err != nil {
			return err
		}
		if err := replaceKnowledgeEvidenceScope(tx, store, result.knowledgeID); err != nil {
			return err
		}
		return replaceKnowledgeReviewScope(tx, store, result.knowledgeID)
	case transition.MachineProposalLine:
		proposal, ok := store.proposals[result.proposalID]
		if !ok {
			return errors.New("proposal not found for persistence")
		}
		if err := replaceProposalScope(tx, store, proposal.ID); err != nil {
			return err
		}
		if err := replaceProposalReviewScope(tx, store, proposal.ID); err != nil {
			return err
		}
		if err := replaceProposalMemoryScope(tx, store, proposal.ID); err != nil {
			return err
		}
		if err := replaceCandidateScope(tx, store, proposal.CandidateKey); err != nil {
			return err
		}
		if strings.TrimSpace(result.traceID) != "" {
			trace, ok := store.traces[result.traceID]
			if !ok {
				return errors.New("proposal-derived attempt trace not found for persistence")
			}
			if err := replaceTraceAndWorkflowScope(tx, store, trace); err != nil {
				return err
			}
		}
		if err := replaceRepoChangeJobScope(tx, store, proposal.ID); err != nil {
			return err
		}
		if attemptID := firstNonEmpty(result.attemptID, proposal.CurrentAttemptID); attemptID != "" {
			if attempt, ok := store.changeAttempts[attemptID]; ok {
				if err := replaceChangeAttemptScope(tx, attempt); err != nil {
					return err
				}
			}
		}
		return nil
	case transition.MachineAttempt:
		attempt, ok := store.changeAttempts[result.attemptID]
		if !ok {
			return errors.New("change attempt not found for persistence")
		}
		if err := replaceChangeAttemptScope(tx, attempt); err != nil {
			return err
		}
		if proposal, ok := store.proposals[attempt.ProposalID]; ok {
			if err := replaceProposalScope(tx, store, proposal.ID); err != nil {
				return err
			}
			if err := replaceProposalMemoryScope(tx, store, proposal.ID); err != nil {
				return err
			}
		}
		if attempt.CandidateKey != "" {
			if err := replaceCandidateScope(tx, store, attempt.CandidateKey); err != nil {
				return err
			}
		}
		if strings.TrimSpace(result.workspaceID) != "" {
			if workspace, ok := store.attemptWorkspaces[result.workspaceID]; ok {
				if err := replaceAttemptWorkspaceScope(tx, workspace); err != nil {
					return err
				}
			}
		}
		if err := replaceRepoChangeJobScope(tx, store, attempt.ProposalID); err != nil {
			return err
		}
		if result.traceID != "" {
			trace, ok := store.traces[result.traceID]
			if !ok {
				return errors.New("attempt trace not found for persistence")
			}
			if err := replaceTraceAndWorkflowScope(tx, store, trace); err != nil {
				return err
			}
		}
		if strings.TrimSpace(result.prAttemptID) != "" {
			if prAttempt, ok := store.prAttempts[result.prAttemptID]; ok {
				if err := replacePRAttemptScope(tx, prAttempt); err != nil {
					return err
				}
			}
		}
		return nil
	case transition.MachineHarness:
		if strings.TrimSpace(result.harnessOverlayID) != "" {
			if err := replaceHarnessOverlayScope(tx, store, result.harnessOverlayID); err != nil {
				return err
			}
		}
		if strings.TrimSpace(result.harnessExperimentID) != "" {
			if err := replaceHarnessExperimentScope(tx, store, result.harnessExperimentID); err != nil {
				return err
			}
		}
		if strings.TrimSpace(result.harnessBindingKey) != "" {
			if err := replaceHarnessSessionBindingScope(tx, store, result.harnessBindingKey); err != nil {
				return err
			}
		}
		if strings.TrimSpace(result.harnessExecutionID) != "" {
			if err := replaceHarnessExecutionScope(tx, store, result.harnessExecutionID); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported machine kind %s for persistence", result.receipt.MachineKind)
	}
}

func buildCommandReceipt(command transition.CommandEnvelope, decision transition.TransitionDecision, updatedAt time.Time, version int64, resultRef string) transition.CommandReceipt {
	return transition.CommandReceipt{
		CommandID:        command.CommandID,
		MachineKind:      command.MachineKind,
		AggregateID:      command.AggregateID,
		CommandKind:      command.CommandKind,
		CausationID:      command.CausationID,
		Actor:            command.Actor,
		DecisionKind:     decision.DecisionKind,
		Reason:           decision.Reason,
		AggregateVersion: version,
		ResultRef:        strings.TrimSpace(resultRef),
		CreatedAt:        command.OccurredAt,
		UpdatedAt:        updatedAt,
	}
}

func buildCommandBundle(command transition.CommandEnvelope, decision transition.TransitionDecision, version int64) transitionPersistBundle {
	if decision.DecisionKind != transition.DecisionAdvance {
		return transitionPersistBundle{}
	}
	now := command.OccurredAt
	bundle := transitionPersistBundle{}
	for idx, event := range decision.Events {
		payload := cloneMetadata(commandPayload(command))
		for key, value := range cloneMetadata(event.Payload) {
			if payload == nil {
				payload = map[string]interface{}{}
			}
			payload[key] = value
		}
		if payload == nil {
			payload = map[string]interface{}{}
		}
		payload["command_kind"] = command.CommandKind
		payload["reason"] = decision.Reason
		bundle.Events = append(bundle.Events, transition.DomainEvent{
			ID:               fmt.Sprintf("evt:%s:%d", command.CommandID, idx),
			MachineKind:      command.MachineKind,
			AggregateID:      command.AggregateID,
			AggregateVersion: version,
			EventKind:        event.Kind,
			CommandID:        command.CommandID,
			CausationID:      command.CausationID,
			Payload:          payload,
			CreatedAt:        now,
		})
	}
	for idx, effect := range decision.Effects {
		payload := cloneMetadata(commandPayload(command))
		for key, value := range cloneMetadata(effect.Payload) {
			if payload == nil {
				payload = map[string]interface{}{}
			}
			payload[key] = value
		}
		if payload == nil {
			payload = map[string]interface{}{}
		}
		payload["command_kind"] = command.CommandKind
		payload["reason"] = decision.Reason
		bundle.Effects = append(bundle.Effects, transition.EffectExecution{
			ID:             nextUUID("eff"),
			MachineKind:    command.MachineKind,
			AggregateID:    command.AggregateID,
			EffectKind:     effect.Kind,
			Status:         effect.Status,
			IdempotencyKey: fmt.Sprintf("%s:%s:%d", command.AggregateID, effect.IdempotencyKey, idx),
			Payload:        payload,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	for idx, nextCommand := range decision.Commands {
		payload := cloneMetadata(commandPayload(command))
		for key, value := range cloneMetadata(nextCommand.Payload) {
			if payload == nil {
				payload = map[string]interface{}{}
			}
			payload[key] = value
		}
		if payload == nil {
			payload = map[string]interface{}{}
		}
		payload["reason"] = decision.Reason
		payload["parent_command_id"] = command.CommandID
		nextID := strings.TrimSpace(nextCommand.CommandID)
		if nextID == "" {
			nextID = fmt.Sprintf("%s:follow:%d:%s", command.CommandID, idx, strings.TrimSpace(nextCommand.CommandKind))
		}
		bundle.Commands = append(bundle.Commands, normalizeCommandEnvelope(transition.CommandEnvelope{
			MachineKind: nextCommand.MachineKind,
			AggregateID: nextCommand.AggregateID,
			CommandKind: nextCommand.CommandKind,
			CommandID:   nextID,
			CausationID: command.CommandID,
			Actor:       firstNonEmpty(nextCommand.Actor, command.Actor),
			OccurredAt:  now,
			Payload:     payload,
		}))
	}
	return bundle
}

func commandPayload(command transition.CommandEnvelope) map[string]interface{} {
	if command.Payload == nil {
		return nil
	}
	payload := map[string]interface{}{}
	for key, value := range command.Payload {
		payload[key] = value
	}
	return payload
}

func mergeCommandMetadataPayload(base map[string]any, extra map[string]any) map[string]any {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	for key, value := range extra {
		out[key] = value
	}
	return out
}

func optionalTimeFromCommand(command transition.CommandEnvelope, key string) *time.Time {
	raw, ok := command.Payload[key]
	if !ok {
		return nil
	}
	switch value := raw.(type) {
	case time.Time:
		if value.IsZero() {
			return nil
		}
		parsed := value
		return &parsed
	case string:
		if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value)); err == nil && !parsed.IsZero() {
			return &parsed
		}
	}
	return nil
}

func proposalDecisionFromCommand(commandKind string) string {
	switch transition.ProposalLineCommandKind(commandKind) {
	case transition.CommandProposalApproveIntervention:
		return string(review.ProposalApproved)
	case transition.CommandProposalRejectLine:
		return string(review.ProposalRejected)
	case transition.CommandProposalDismissLine:
		return string(review.ProposalDismissed)
	default:
		return ""
	}
}

func attemptFollowOnCommand(parent transition.CommandEnvelope, proposal review.Proposal, attempt improvement.ChangeAttempt, traceID string, kind transition.AttemptPhaseCommandKind) *transition.CommandEnvelope {
	if strings.TrimSpace(attempt.ID) == "" || kind == "" {
		return nil
	}
	payload := map[string]any{
		"proposal_id": proposal.ID,
		"attempt_id":  attempt.ID,
		"trace_id":    firstNonEmpty(strings.TrimSpace(traceID), strings.TrimSpace(attempt.AttemptTraceID), strings.TrimSpace(proposal.TraceID)),
		"branch_name": attempt.BranchName,
	}
	return &transition.CommandEnvelope{
		MachineKind: transition.MachineAttempt,
		AggregateID: attempt.ID,
		CommandKind: string(kind),
		CommandID:   storeAttemptCommandID(attempt.ID, kind),
		Actor:       firstNonEmpty(parent.Actor, "formal-transition"),
		OccurredAt:  parent.OccurredAt,
		Payload:     payload,
	}
}

func proposalCommandKindForDecision(decision string) (transition.ProposalLineCommandKind, error) {
	switch review.ProposalStatus(strings.TrimSpace(decision)) {
	case review.ProposalApproved:
		return transition.CommandProposalApproveIntervention, nil
	case review.ProposalRejected:
		return transition.CommandProposalRejectLine, nil
	case review.ProposalDismissed:
		return transition.CommandProposalDismissLine, nil
	case review.ProposalMerged:
		return transition.CommandProposalMarkMerged, nil
	default:
		return "", fmt.Errorf("unsupported proposal review decision %q", decision)
	}
}

func knowledgeDecisionFromCommand(commandKind string) string {
	switch transition.KnowledgeCommandKind(commandKind) {
	case transition.CommandKnowledgeApprove:
		return "approve"
	case transition.CommandKnowledgeReject:
		return "reject"
	case transition.CommandKnowledgeMarkStale:
		return "mark_stale"
	case transition.CommandKnowledgeArchive:
		return "archive"
	default:
		return ""
	}
}

func reviewScopeFromCommand(command transition.CommandEnvelope) review.ProposalFeedbackScope {
	scope := review.ProposalFeedbackScope(stringFromCommand(command, "scope"))
	if scope == "" {
		return review.FeedbackScopeLine
	}
	return scope
}

func stringFromCommand(command transition.CommandEnvelope, key string) string {
	raw, ok := command.Payload[key]
	if !ok {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}

func stringSliceFromCommand(command transition.CommandEnvelope, key string) []string {
	raw, ok := command.Payload[key]
	if !ok {
		return nil
	}
	switch value := raw.(type) {
	case []string:
		return append([]string(nil), value...)
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			text := strings.TrimSpace(fmt.Sprintf("%v", item))
			if text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		text := strings.TrimSpace(fmt.Sprintf("%v", value))
		if text == "" {
			return nil
		}
		return []string{text}
	}
}

func anyMapFromCommand(command transition.CommandEnvelope, key string) map[string]any {
	raw, ok := command.Payload[key]
	if !ok {
		return nil
	}
	switch value := raw.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for k, v := range value {
			out[k] = v
		}
		return out
	default:
		return nil
	}
}

func floatFromCommand(command transition.CommandEnvelope, key string) float64 {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return 0
	}
	switch value := raw.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func optionalBoolPointerFromCommand(command transition.CommandEnvelope, key string) *bool {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case bool:
		copied := value
		return &copied
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true":
			copied := true
			return &copied
		case "false":
			copied := false
			return &copied
		}
	}
	return nil
}

func harnessOverlayFromCommand(command transition.CommandEnvelope) (harness.Overlay, error) {
	overlayID := strings.TrimSpace(command.AggregateID)
	if overlayID == "" {
		return harness.Overlay{}, errors.New("harness overlay aggregate id is required")
	}
	status := harness.OverlayStatus(firstNonEmpty(stringFromCommand(command, "status"), string(harness.OverlayStatusActive)))
	return harness.Overlay{
		ID:                  overlayID,
		ProfileID:           stringFromCommand(command, "profile_id"),
		Role:                stringFromCommand(command, "role"),
		Version:             stringFromCommand(command, "version"),
		Status:              status,
		TargetKind:          stringFromCommand(command, "target_kind"),
		TargetRef:           stringFromCommand(command, "target_ref"),
		ProposalID:          stringFromCommand(command, "proposal_id"),
		PromptFragments:     stringSliceFromCommand(command, "prompt_fragments"),
		FewShotSnippets:     stringSliceFromCommand(command, "few_shot_snippets"),
		ToolPreferenceOrder: stringSliceFromCommand(command, "tool_preference_order"),
		RetrievalBias:       stringFromCommand(command, "retrieval_bias"),
		ReasoningVerbosity:  stringFromCommand(command, "reasoning_verbosity"),
		MemoryReadEnabled:   optionalBoolPointerFromCommand(command, "memory_read_enabled"),
		MemoryWriteEnabled:  optionalBoolPointerFromCommand(command, "memory_write_enabled"),
		CreatedBy:           firstNonEmpty(stringFromCommand(command, "created_by"), command.Actor),
		ApprovedBy:          firstNonEmpty(stringFromCommand(command, "approved_by"), command.Actor),
		CreatedAt:           command.OccurredAt,
		UpdatedAt:           command.OccurredAt,
		ActivatedAt:         ptrTimeValue(command.OccurredAt),
	}, nil
}

func harnessExperimentFromCommand(command transition.CommandEnvelope) harness.Experiment {
	return harness.Experiment{
		ID:         stringFromCommand(command, "experiment_id"),
		ProfileID:  stringFromCommand(command, "profile_id"),
		OverlayID:  strings.TrimSpace(command.AggregateID),
		ProposalID: stringFromCommand(command, "proposal_id"),
		AttemptID:  stringFromCommand(command, "attempt_id"),
		Role:       stringFromCommand(command, "role"),
		Status:     harness.ExperimentStatus(firstNonEmpty(stringFromCommand(command, "experiment_status"), string(harness.ExperimentStatusSucceeded))),
		Summary:    stringFromCommand(command, "experiment_summary"),
		Metrics:    anyMapFromCommand(command, "experiment_metrics"),
		CreatedAt:  command.OccurredAt,
		UpdatedAt:  command.OccurredAt,
	}
}

func prAttemptFromAttemptCommand(proposal review.Proposal, attempt improvement.ChangeAttempt, command transition.CommandEnvelope) improvement.PRAttempt {
	item := improvement.PRAttempt{
		ProposalID:       attempt.ProposalID,
		AttemptID:        attempt.ID,
		ConversationID:   proposal.ConversationID,
		CaseID:           proposal.CaseID,
		OriginTraceID:    firstNonEmpty(attempt.AttemptTraceID, proposal.OriginTraceID, proposal.TraceID),
		Repo:             stringFromCommand(command, "repo"),
		BranchName:       firstNonEmpty(stringFromCommand(command, "branch_name"), attempt.BranchName),
		PRURL:            stringFromCommand(command, "pr_url"),
		HeadSHA:          firstNonEmpty(stringFromCommand(command, "head_sha"), attempt.HeadSHA),
		Status:           string(review.ProposalPROpen),
		ValidationStatus: firstNonEmpty(stringFromCommand(command, "validation_status"), "pending"),
		CreatedAt:        timeFromCommand(command, "created_at", command.OccurredAt),
	}
	if prID := strings.TrimSpace(stringFromCommand(command, "pr_attempt_id")); prID != "" {
		item.ID = prID
	}
	return item
}

func harnessSessionBindingFromCommand(command transition.CommandEnvelope) (harness.SessionBinding, error) {
	role := strings.TrimSpace(stringFromCommand(command, "role"))
	scopeKind := firstNonEmpty(stringFromCommand(command, "scope_kind"), stringFromCommand(command, "session_scope_kind"))
	scopeID := firstNonEmpty(stringFromCommand(command, "scope_id"), stringFromCommand(command, "session_scope_id"))
	hermesSessionID := strings.TrimSpace(stringFromCommand(command, "hermes_session_id"))
	if role == "" {
		return harness.SessionBinding{}, errors.New("harness session binding role is required")
	}
	if scopeKind == "" || scopeID == "" {
		return harness.SessionBinding{}, errors.New("harness session binding scope is required")
	}
	if hermesSessionID == "" {
		return harness.SessionBinding{}, errors.New("harness session binding hermes_session_id is required")
	}
	return harness.SessionBinding{
		Role:                    role,
		ScopeKind:               scopeKind,
		ScopeID:                 scopeID,
		ParentScopeKind:         firstNonEmpty(stringFromCommand(command, "parent_scope_kind"), stringFromCommand(command, "requested_parent_scope_kind")),
		ParentScopeID:           firstNonEmpty(stringFromCommand(command, "parent_scope_id"), stringFromCommand(command, "requested_parent_scope_id")),
		HermesSessionID:         hermesSessionID,
		ParentSessionID:         stringFromCommand(command, "parent_session_id"),
		MemoryBackend:           stringFromCommand(command, "memory_backend"),
		AssistantPeerID:         stringFromCommand(command, "assistant_peer_id"),
		UserPeerID:              stringFromCommand(command, "user_peer_id"),
		HarnessProfileID:        stringFromCommand(command, "harness_profile_id"),
		EffectiveOverlayID:      stringFromCommand(command, "effective_overlay_id"),
		EffectiveOverlayVersion: stringFromCommand(command, "effective_overlay_version"),
		LastUsedAt:              timeFromCommand(command, "last_used_at", command.OccurredAt),
		CreatedAt:               timeFromCommand(command, "created_at", command.OccurredAt),
		UpdatedAt:               timeFromCommand(command, "updated_at", command.OccurredAt),
	}, nil
}

func harnessExecutionFromCommand(command transition.CommandEnvelope) (harness.Execution, error) {
	executionID := strings.TrimSpace(command.AggregateID)
	if executionID == "" {
		return harness.Execution{}, errors.New("harness execution aggregate id is required")
	}
	role := strings.TrimSpace(stringFromCommand(command, "role"))
	scopeKind := firstNonEmpty(stringFromCommand(command, "session_scope_kind"), stringFromCommand(command, "scope_kind"))
	scopeID := firstNonEmpty(stringFromCommand(command, "session_scope_id"), stringFromCommand(command, "scope_id"))
	if role == "" {
		return harness.Execution{}, errors.New("harness execution role is required")
	}
	if scopeKind == "" || scopeID == "" {
		return harness.Execution{}, errors.New("harness execution scope is required")
	}
	return harness.Execution{
		ID:                      executionID,
		OperationID:             stringFromCommand(command, "operation_id"),
		TraceID:                 stringFromCommand(command, "trace_id"),
		ProposalID:              stringFromCommand(command, "proposal_id"),
		Role:                    role,
		SessionScopeKind:        scopeKind,
		SessionScopeID:          scopeID,
		HermesSessionID:         stringFromCommand(command, "hermes_session_id"),
		ParentSessionID:         stringFromCommand(command, "parent_session_id"),
		HarnessProfileID:        stringFromCommand(command, "harness_profile_id"),
		EffectiveOverlayID:      stringFromCommand(command, "effective_overlay_id"),
		EffectiveOverlayVersion: stringFromCommand(command, "effective_overlay_version"),
		MemoryBackend:           stringFromCommand(command, "memory_backend"),
		MemoryReads:             memoryArtifactsFromCommand(command, "memory_reads"),
		MemoryWrites:            memoryArtifactsFromCommand(command, "memory_writes"),
		CreatedAt:               timeFromCommand(command, "created_at", command.OccurredAt),
	}, nil
}

func knowledgeEntryFromCommand(command transition.CommandEnvelope) (knowledge.Entry, []knowledge.EvidenceLink, error) {
	knowledgeID := strings.TrimSpace(command.AggregateID)
	if knowledgeID == "" {
		return knowledge.Entry{}, nil, errors.New("knowledge aggregate id is required")
	}
	entry := knowledge.Entry{
		ID:                    knowledgeID,
		Tier:                  knowledge.Tier(firstNonEmpty(stringFromCommand(command, "tier"), string(knowledge.TierWorking))),
		Kind:                  knowledge.Kind(firstNonEmpty(stringFromCommand(command, "kind"), string(knowledge.KindFact))),
		ScopeType:             knowledge.ScopeType(firstNonEmpty(stringFromCommand(command, "scope_type"), string(knowledge.ScopeCase))),
		ScopeID:               stringFromCommand(command, "scope_id"),
		Title:                 stringFromCommand(command, "title"),
		Summary:               stringFromCommand(command, "summary"),
		Body:                  stringFromCommand(command, "body"),
		StructuredFacts:       anyMapFromCommand(command, "structured_facts"),
		Status:                knowledge.Status(firstNonEmpty(stringFromCommand(command, "status"), string(knowledge.StatusDraft))),
		Confidence:            floatFromCommand(command, "confidence"),
		FreshUntil:            optionalTimeFromCommand(command, "fresh_until"),
		SourceType:            knowledge.SourceType(firstNonEmpty(stringFromCommand(command, "source_type"), string(knowledge.SourceAgent))),
		SupersedesEntryID:     stringFromCommand(command, "supersedes_entry_id"),
		ContradictedByEntryID: stringFromCommand(command, "contradicted_by_entry_id"),
		CreatedAt:             timeFromCommand(command, "created_at", command.OccurredAt),
		UpdatedAt:             timeFromCommand(command, "updated_at", command.OccurredAt),
	}
	links := knowledgeEvidenceLinksFromCommand(command, "evidence_links")
	if len(links) == 0 {
		for _, ref := range evidenceRefsFromCommand(command, "evidence_refs") {
			links = append(links, knowledge.EvidenceLink{
				KnowledgeEntryID: knowledgeID,
				EvidenceType:     ref.Kind,
				EvidenceID:       ref.Ref,
				RelevanceSummary: ref.Summary,
				EvidenceRef:      ref,
			})
		}
	}
	for idx := range links {
		links[idx].KnowledgeEntryID = knowledgeID
	}
	return entry, links, nil
}

func problemLineOutcomeFromCommand(command transition.CommandEnvelope) outcome.Record {
	return outcome.Record{
		ID:             stringFromCommand(command, "outcome_id"),
		ConversationID: stringFromCommand(command, "conversation_id"),
		CaseID:         stringFromCommand(command, "case_id"),
		TraceID:        stringFromCommand(command, "trace_id"),
		ProposalID:     stringFromCommand(command, "proposal_id"),
		AttemptID:      stringFromCommand(command, "attempt_id"),
		OperationID:    stringFromCommand(command, "operation_id"),
		OutcomeType:    outcome.Type(firstNonEmpty(stringFromCommand(command, "outcome_type"), string(outcome.TypeAnswerQuality))),
		Verdict:        outcome.Verdict(stringFromCommand(command, "verdict")),
		Score:          floatFromCommand(command, "score"),
		Summary:        stringFromCommand(command, "summary"),
		Details:        stringFromCommand(command, "details"),
		Source:         firstNonEmpty(stringFromCommand(command, "source"), "operator"),
		RecordedBy:     firstNonEmpty(command.Actor, stringFromCommand(command, "recorded_by")),
		RecordedAt:     command.OccurredAt,
	}
}

func hasPromotableCandidatesLocked(items map[string]improvement.Candidate) bool {
	for _, item := range items {
		if item.Status == improvement.CandidateQueued {
			return true
		}
	}
	return false
}

func promotionLeaseBlockedLocked(leases map[string]improvement.CronLease, holder string, now time.Time) bool {
	lease, ok := leases["improvement-plane-cron"]
	if !ok {
		return false
	}
	if !lease.ExpiresAt.After(now) {
		return false
	}
	return strings.TrimSpace(lease.Holder) != "" && strings.TrimSpace(lease.Holder) != strings.TrimSpace(holder)
}

func stringFromPayload(payload map[string]interface{}, key string) string {
	if payload == nil {
		return ""
	}
	raw, ok := payload[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", raw))
}

func findEvalRunByID(items map[string]evals.Run, runID string) (evals.Run, bool) {
	item, ok := items[strings.TrimSpace(runID)]
	return item, ok
}

func workflowStateFromStatus(status string) transition.WorkflowStateKind {
	switch strings.TrimSpace(status) {
	case "", string(transition.WorkflowStateQueued):
		return transition.WorkflowStateQueued
	case "running", string(transition.WorkflowStateCollectingContext):
		return transition.WorkflowStateCollectingContext
	case string(transition.WorkflowStateWaitingOnActions):
		return transition.WorkflowStateWaitingOnActions
	case string(transition.WorkflowStateReasoning):
		return transition.WorkflowStateReasoning
	case string(transition.WorkflowStateReplyPending):
		return transition.WorkflowStateReplyPending
	case "needs-human", string(transition.WorkflowStateNeedsHuman):
		return transition.WorkflowStateNeedsHuman
	case string(transition.WorkflowStateCompleted):
		return transition.WorkflowStateCompleted
	case string(transition.WorkflowStateFailed):
		return transition.WorkflowStateFailed
	case string(transition.WorkflowStateSuperseded):
		return transition.WorkflowStateSuperseded
	default:
		return transition.WorkflowStateKind(strings.TrimSpace(status))
	}
}

func workflowStatusFromState(state transition.WorkflowStateKind) string {
	return strings.TrimSpace(string(state))
}

func workflowLastErrorForCommand(command transition.CommandEnvelope) string {
	switch transition.WorkflowCommandKind(command.CommandKind) {
	case transition.CommandWorkflowBlocked, transition.CommandWorkflowFailed:
		return stringFromCommand(command, "last_error")
	default:
		return ""
	}
}

func attemptFailureStateFromCommand(commandKind string, failureClass string, retryable bool) improvement.ChangeAttemptState {
	switch transition.AttemptPhaseCommandKind(commandKind) {
	case transition.CommandWorkspaceFailedRetryable, transition.CommandWorkspaceFailedReview,
		transition.CommandImplementationFailedRetryable, transition.CommandImplementationFailedReview,
		transition.CommandValidationFailedRetryable, transition.CommandValidationFailedReview:
		switch strings.TrimSpace(failureClass) {
		case "sandbox_failure":
			return improvement.AttemptStateSandboxFailed
		case "ci_regression":
			return improvement.AttemptStateCIFailed
		case "closed_unmerged":
			return improvement.AttemptStateClosedUnmerged
		}
	case transition.CommandPROpenFailedRetryable, transition.CommandPROpenFailedReview:
		return improvement.AttemptStateNeedsReview
	}
	if retryable {
		return improvement.AttemptStateSandboxFailed
	}
	return improvement.AttemptStateNeedsReview
}

func boolFromCommand(command transition.CommandEnvelope, key string, fallback bool) bool {
	raw, ok := command.Payload[key]
	if !ok {
		return fallback
	}
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}
	return fallback
}

func feedbackRecordFromCommand(command transition.CommandEnvelope) review.FeedbackRecord {
	return review.FeedbackRecord{
		ID:         stringFromCommand(command, "feedback_id"),
		TraceID:    firstNonEmpty(stringFromCommand(command, "trace_id"), command.AggregateID),
		TargetType: review.FeedbackTargetType(stringFromCommand(command, "target_type")),
		TargetID:   stringFromCommand(command, "target_id"),
		Score:      int(floatFromCommand(command, "score")),
		Verdict:    stringFromCommand(command, "verdict"),
		Labels:     stringSliceFromCommand(command, "labels"),
		Notes:      stringFromCommand(command, "notes"),
		ReviewerID: firstNonEmpty(command.Actor, stringFromCommand(command, "reviewer_id")),
	}
}

func humanRatingFromCommand(command transition.CommandEnvelope) review.HumanRating {
	return review.HumanRating{
		TraceID:    firstNonEmpty(stringFromCommand(command, "trace_id"), command.AggregateID),
		Score:      int(floatFromCommand(command, "score")),
		Verdict:    stringFromCommand(command, "verdict"),
		Labels:     stringSliceFromCommand(command, "labels"),
		Notes:      stringFromCommand(command, "notes"),
		ReviewerID: firstNonEmpty(command.Actor, stringFromCommand(command, "reviewer_id")),
		CreatedAt:  timeFromCommand(command, "created_at", command.OccurredAt),
	}
}

func improvementNoteFromCommand(command transition.CommandEnvelope) review.ImprovementNote {
	return review.ImprovementNote{
		TraceID:        firstNonEmpty(stringFromCommand(command, "trace_id"), command.AggregateID),
		Category:       stringFromCommand(command, "category"),
		Note:           stringFromCommand(command, "note"),
		SuggestedOwner: stringFromCommand(command, "suggested_owner"),
		CreatedBy:      firstNonEmpty(command.Actor, stringFromCommand(command, "created_by")),
		CreatedAt:      timeFromCommand(command, "created_at", command.OccurredAt),
	}
}

func eventEnvelopeFromCommand(command transition.CommandEnvelope) ingestion.EventEnvelope {
	return ingestion.EventEnvelope{
		ID:                         stringFromCommand(command, "event_id"),
		Source:                     ingestion.Source(firstNonEmpty(stringFromCommand(command, "source"), string(ingestion.SourceSystem))),
		SourceEventID:              firstNonEmpty(stringFromCommand(command, "source_event_id"), strings.TrimSpace(command.AggregateID)),
		ThreadKey:                  stringFromCommand(command, "thread_key"),
		IncidentKey:                stringFromCommand(command, "incident_key"),
		DedupeKey:                  firstNonEmpty(stringFromCommand(command, "dedupe_key"), strings.TrimSpace(command.AggregateID)),
		Severity:                   ingestion.Severity(firstNonEmpty(stringFromCommand(command, "severity"), string(ingestion.SeverityWarning))),
		NormalizedProblemStatement: stringFromCommand(command, "normalized_problem_statement"),
		OwnershipHint:              stringFromCommand(command, "ownership_hint"),
		RawPayloadRef:              stringFromCommand(command, "raw_payload_ref"),
		WorkflowHint:               stringFromCommand(command, "workflow_hint"),
		Metadata:                   anyMapFromCommand(command, "metadata"),
		CreatedAt:                  firstNonZeroTime(optionalTimeFromCommand(command, "created_at"), command.OccurredAt),
	}
}

func slackEnvelopeFromCommand(command transition.CommandEnvelope) slack.SlackEnvelope {
	return slack.SlackEnvelope{
		BotRole:   slack.BotRole(stringFromCommand(command, "bot_role")),
		TeamID:    stringFromCommand(command, "team_id"),
		ChannelID: stringFromCommand(command, "channel_id"),
		ThreadTS:  stringFromCommand(command, "thread_ts"),
		UserID:    stringFromCommand(command, "user_id"),
		Text:      stringFromCommand(command, "text"),
		TS:        stringFromCommand(command, "ts"),
		Files:     stringSliceFromCommand(command, "files"),
		CreatedAt: firstNonZeroTime(optionalTimeFromCommand(command, "created_at"), command.OccurredAt),
	}
}

func firstNonZeroTime(value *time.Time, fallback time.Time) time.Time {
	if value != nil && !value.IsZero() {
		return *value
	}
	return fallback
}

func memoryArtifactsFromCommand(command transition.CommandEnvelope, key string) []harness.MemoryArtifact {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case []harness.MemoryArtifact:
		out := make([]harness.MemoryArtifact, len(value))
		copy(out, value)
		return out
	}
	var out []harness.MemoryArtifact
	if decodeCommandPayload(raw, &out) {
		return out
	}
	var single harness.MemoryArtifact
	if decodeCommandPayload(raw, &single) {
		return []harness.MemoryArtifact{single}
	}
	return nil
}

func actionResultFromCommand(intent action.Intent, status action.Status, command transition.CommandEnvelope) action.Result {
	startedAt := timeFromCommand(command, "started_at", command.OccurredAt)
	completedAt := timeFromCommand(command, "completed_at", command.OccurredAt)
	return action.Result{
		OperationID:        firstNonEmpty(stringFromCommand(command, "operation_id"), intent.OperationID),
		ActionIntentID:     intent.ID,
		AttemptID:          firstNonEmpty(stringFromCommand(command, "attempt_id"), intent.AttemptID),
		Executor:           stringFromCommand(command, "executor"),
		Provider:           stringFromCommand(command, "provider"),
		ProviderRef:        stringFromCommand(command, "provider_ref"),
		RequestArtifactID:  stringFromCommand(command, "request_artifact_id"),
		ResponseArtifactID: stringFromCommand(command, "response_artifact_id"),
		Status:             status,
		ErrorCode:          stringFromCommand(command, "error_code"),
		ErrorMessage:       stringFromCommand(command, "error_message"),
		StartedAt:          startedAt,
		CompletedAt:        completedAt,
	}
}

func firstNonEmptyMap(primary map[string]any, fallback map[string]any) map[string]any {
	if len(primary) > 0 {
		return primary
	}
	return fallback
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func ptrStatus(status events.Status) *events.Status {
	return &status
}

func (s *MemoryStore) projectActionTraceLocked(intent action.Intent, command transition.CommandEnvelope, status action.Status) error {
	traceID := strings.TrimSpace(intent.TraceID)
	if traceID == "" {
		return nil
	}
	trace, ok := s.traces[traceID]
	if !ok {
		return nil
	}
	workflow, _ := findWorkflowByID(s.workflows, trace.Summary.WorkflowID)
	startedAt := timeFromCommand(command, "started_at", command.OccurredAt)
	completedAt := timeFromCommand(command, "completed_at", command.OccurredAt)
	summary := firstNonEmpty(
		stringFromCommand(command, "summary"),
		stringFromCommand(command, "error_message"),
		intent.PolicyVerdict,
		string(status),
	)

	update := TraceUpdate{
		Status:    traceStatusFromCommand(command, "trace_status"),
		Events:    traceEventsFromCommand(command, "trace_events"),
		Artifacts: traceArtifactsFromCommand(command, "trace_artifacts"),
		Reasoning: reasoningStepsFromCommand(command, "reasoning_steps"),
	}
	switch intent.Kind {
	case action.KindToolRead:
		if status != action.StatusExecuting && status != action.StatusQueued && status != action.StatusCanceled {
			eventStatus := events.StatusCompleted
			eventType := "tool.completed"
			switch status {
			case action.StatusBlocked:
				eventStatus = events.StatusNeedsHuman
				eventType = "tool.blocked"
			case action.StatusFailed:
				eventStatus = events.StatusNeedsHuman
				eventType = "tool.failed"
			}
			update.Events = append(update.Events, events.TraceEvent{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     firstNonEmpty(stringFromCommand(command, "executor"), "tool-gateway"),
				Actor:       workflow.AssignedBot,
				EventType:   eventType,
				Status:      eventStatus,
				StartedAt:   startedAt,
				EndedAt:     ptrTime(completedAt),
				Description: summary,
			})
			update.ToolCalls = []events.ToolCallRecord{{
				ID:                    fmt.Sprintf("tool-record-%s", strings.TrimSpace(command.CommandID)),
				TraceID:               trace.Summary.TraceID,
				WorkflowID:            trace.Summary.WorkflowID,
				ConversationID:        intent.ConversationID,
				CaseID:                intent.CaseID,
				ToolName:              intent.TargetRef,
				ToolCallID:            firstNonEmpty(stringFromCommand(command, "tool_call_id"), stringFromCommand(command, "provider_ref"), intent.ID),
				Request:               firstNonEmptyMap(anyMapFromCommand(command, "request_payload"), intent.RequestPayload),
				Summary:               summary,
				RawArtifactRefs:       stringSliceFromCommand(command, "raw_artifact_refs"),
				ApprovalState:         intent.ApprovalState,
				InterpretationSummary: summary,
				Status:                string(status),
				CreatedAt:             completedAt,
			}}
		}
	case action.KindSlackPost:
		if status != action.StatusExecuting && status != action.StatusQueued && status != action.StatusCanceled {
			eventStatus := events.StatusCompleted
			eventType := "slack.reply.posted"
			switch status {
			case action.StatusBlocked:
				eventStatus = events.StatusNeedsHuman
				eventType = "slack.reply.blocked"
			case action.StatusFailed:
				eventStatus = events.StatusNeedsHuman
				eventType = "slack.reply.failed"
			}
			sendStatus := stringFromCommand(command, "send_status")
			if status == action.StatusSucceeded && strings.TrimSpace(sendStatus) == "" {
				sendStatus = "posted"
			}
			update.Events = append(update.Events, events.TraceEvent{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "edge",
				Service:     firstNonEmpty(stringFromCommand(command, "executor"), "tool-gateway"),
				Actor:       workflow.AssignedBot,
				EventType:   eventType,
				Status:      eventStatus,
				StartedAt:   startedAt,
				EndedAt:     ptrTime(completedAt),
				Description: summary,
			})
			update.SlackActions = []events.SlackActionRecord{{
				ID:             fmt.Sprintf("slack-action-%s", strings.TrimSpace(command.CommandID)),
				TraceID:        trace.Summary.TraceID,
				WorkflowID:     trace.Summary.WorkflowID,
				ConversationID: intent.ConversationID,
				CaseID:         intent.CaseID,
				ChannelID:      firstNonEmpty(stringFromCommand(command, "channel_id"), stringFromPayload(intent.RequestPayload, "channel_id")),
				ThreadTS:       firstNonEmpty(stringFromCommand(command, "thread_ts"), stringFromPayload(intent.RequestPayload, "thread_ts")),
				IdempotencyKey: intent.IdempotencyKey,
				DraftBody:      firstNonEmpty(stringFromCommand(command, "draft_body"), stringFromPayload(intent.RequestPayload, "draft_body")),
				FinalBody:      firstNonEmpty(stringFromCommand(command, "final_body"), stringFromPayload(intent.RequestPayload, "final_body"), stringFromPayload(intent.RequestPayload, "body")),
				PolicyVerdict:  firstNonEmpty(stringFromCommand(command, "policy_verdict"), intent.PolicyVerdict),
				SendStatus:     sendStatus,
				ArtifactRefs:   stringSliceFromCommand(command, "artifact_refs"),
				CreatedAt:      completedAt,
			}}
		}
	default:
	}
	if update.Status == nil && len(update.Events) == 0 && len(update.Artifacts) == 0 && len(update.Reasoning) == 0 && len(update.ToolCalls) == 0 && len(update.SlackActions) == 0 {
		return nil
	}
	_, err := s.applyTraceUpdateLocked(traceID, update)
	return err
}

func (s *MemoryStore) projectWorkflowTraceLocked(workflow Workflow, command transition.CommandEnvelope) error {
	traceID := strings.TrimSpace(workflow.TraceID)
	if traceID == "" {
		return nil
	}
	trace, ok := s.traces[traceID]
	if !ok {
		return nil
	}
	commandKind := transition.WorkflowCommandKind(command.CommandKind)
	update := TraceUpdate{}
	switch commandKind {
	case transition.CommandWorkflowStarted:
		if traceHasEventType(trace, "workflow.started") {
			return nil
		}
		update.Status = ptrStatus(events.StatusRunning)
		update.Events = []events.TraceEvent{{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     firstNonEmpty(command.Actor, "control-plane"),
			Actor:       "worker",
			EventType:   "workflow.started",
			Status:      events.StatusRunning,
			StartedAt:   command.OccurredAt,
			Description: firstNonEmpty(workflow.Intent, "workflow") + " workflow started.",
		}}
		update.Reasoning = append(update.Reasoning, reasoningStepsFromCommand(command, "reasoning_steps")...)
		if len(update.Reasoning) == 0 {
			update.Reasoning = append(update.Reasoning, events.ReasoningStep{
				ID:         fmt.Sprintf("reason-start-%d", command.OccurredAt.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "pre_action_summary",
				Summary:    fmt.Sprintf("Preparing %s response for conversation %s.", workflow.Intent, trace.Summary.ConversationID),
				Confidence: 0.9,
				Decision:   fmt.Sprintf("response_mode:%s", workflow.ResponseMode),
				CreatedAt:  command.OccurredAt,
			})
		}
	case transition.CommandContextActionsQueued, transition.CommandContextSkipped, transition.CommandContextCompleted:
		update.Events = append(update.Events, traceEventsFromCommand(command, "trace_events")...)
		update.Reasoning = append(update.Reasoning, reasoningStepsFromCommand(command, "reasoning_steps")...)
	case transition.CommandRunnerCompleted:
		update.Events = append(update.Events, traceEventsFromCommand(command, "trace_events")...)
		update.Reasoning = append(update.Reasoning, reasoningStepsFromCommand(command, "reasoning_steps")...)
	case transition.CommandWorkflowBlocked:
		if traceHasEventType(trace, "workflow.blocked") {
			return nil
		}
		update.Status = ptrStatus(events.StatusNeedsHuman)
		update.Events = []events.TraceEvent{{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     firstNonEmpty(command.Actor, "control-plane"),
			Actor:       "worker",
			EventType:   "workflow.blocked",
			Status:      events.StatusNeedsHuman,
			StartedAt:   command.OccurredAt,
			Description: workflowLastErrorForCommand(command),
		}}
		update.Events = append(update.Events, traceEventsFromCommand(command, "trace_events")...)
	case transition.CommandWorkflowFailed:
		if traceHasEventType(trace, "workflow.failed") {
			return nil
		}
		update.Status = ptrStatus(events.StatusFailed)
		update.Events = []events.TraceEvent{{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     firstNonEmpty(command.Actor, "control-plane"),
			Actor:       "worker",
			EventType:   "workflow.failed",
			Status:      events.StatusFailed,
			StartedAt:   command.OccurredAt,
			Description: workflowLastErrorForCommand(command),
		}}
	case transition.CommandReplyPosted, transition.CommandRunnerCompletedNoReply:
		if traceHasEventType(trace, "workflow.completed") {
			return nil
		}
		update.Status = ptrStatus(events.StatusCompleted)
		update.Events = []events.TraceEvent{{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     firstNonEmpty(command.Actor, "control-plane"),
			Actor:       "worker",
			EventType:   "workflow.completed",
			Status:      events.StatusCompleted,
			StartedAt:   command.OccurredAt,
			Description: "Workflow completed.",
		}}
		update.Events = append(traceEventsFromCommand(command, "trace_events"), update.Events...)
		update.Reasoning = append(update.Reasoning, reasoningStepsFromCommand(command, "reasoning_steps")...)
	default:
		return nil
	}
	if update.Status == nil && len(update.Events) == 0 && len(update.Reasoning) == 0 {
		return nil
	}
	_, err := s.applyTraceUpdateLocked(traceID, update)
	return err
}

func traceHasEventType(trace events.Trace, eventType string) bool {
	for _, item := range trace.Events {
		if item.EventType == eventType {
			return true
		}
	}
	return false
}

func (s *MemoryStore) projectAttemptTraceLocked(attempt improvement.ChangeAttempt, command transition.CommandEnvelope) error {
	traceID := strings.TrimSpace(attempt.AttemptTraceID)
	if traceID == "" {
		return nil
	}
	return s.projectTraceFromCommandLocked(traceID, command)
}

func (s *MemoryStore) projectTraceFromCommandLocked(traceID string, command transition.CommandEnvelope) error {
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return nil
	}
	update := TraceUpdate{
		Status:         traceStatusFromCommand(command, "trace_status"),
		LastVerdict:    optionalStringFromCommand(command, "last_verdict"),
		WorkflowStatus: stringFromCommand(command, "workflow_status"),
		WorkflowError:  stringFromCommand(command, "workflow_error"),
		Events:         traceEventsFromCommand(command, "trace_events"),
		Artifacts:      traceArtifactsFromCommand(command, "trace_artifacts"),
		Reasoning:      reasoningStepsFromCommand(command, "reasoning_steps"),
		ToolCalls:      toolCallRecordsFromCommand(command, "tool_calls"),
		SlackActions:   slackActionsFromCommand(command, "slack_actions"),
	}
	if update.Status == nil &&
		update.LastVerdict == nil &&
		update.WorkflowStatus == "" &&
		update.WorkflowError == "" &&
		len(update.Events) == 0 &&
		len(update.Artifacts) == 0 &&
		len(update.Reasoning) == 0 &&
		len(update.ToolCalls) == 0 &&
		len(update.SlackActions) == 0 {
		return nil
	}
	_, err := s.applyTraceUpdateLocked(traceID, update)
	return err
}

func traceStatusFromCommand(command transition.CommandEnvelope, key string) *events.Status {
	status := strings.TrimSpace(stringFromCommand(command, key))
	if status == "" {
		return nil
	}
	value := events.Status(status)
	return &value
}

func traceEventsFromCommand(command transition.CommandEnvelope, key string) []events.TraceEvent {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case []events.TraceEvent:
		out := make([]events.TraceEvent, len(value))
		copy(out, value)
		return out
	case events.TraceEvent:
		return []events.TraceEvent{value}
	}
	var out []events.TraceEvent
	if decodeCommandPayload(raw, &out) {
		return out
	}
	var single events.TraceEvent
	if decodeCommandPayload(raw, &single) {
		return []events.TraceEvent{single}
	}
	return nil
}

func traceArtifactsFromCommand(command transition.CommandEnvelope, key string) []events.Artifact {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case []events.Artifact:
		out := make([]events.Artifact, len(value))
		copy(out, value)
		return out
	case events.Artifact:
		return []events.Artifact{value}
	}
	var out []events.Artifact
	if decodeCommandPayload(raw, &out) {
		return out
	}
	var single events.Artifact
	if decodeCommandPayload(raw, &single) {
		return []events.Artifact{single}
	}
	return nil
}

func evidenceRefsFromCommand(command transition.CommandEnvelope, key string) []events.EvidenceRef {
	raw, ok := command.Payload[key]
	if !ok {
		return nil
	}
	items, ok := raw.([]events.EvidenceRef)
	if !ok {
		return nil
	}
	out := make([]events.EvidenceRef, len(items))
	copy(out, items)
	return out
}

func knowledgeEvidenceLinksFromCommand(command transition.CommandEnvelope, key string) []knowledge.EvidenceLink {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case []knowledge.EvidenceLink:
		out := make([]knowledge.EvidenceLink, len(value))
		copy(out, value)
		return out
	case []events.EvidenceRef:
		out := make([]knowledge.EvidenceLink, 0, len(value))
		for _, ref := range value {
			out = append(out, knowledge.EvidenceLink{
				EvidenceType:     ref.Kind,
				EvidenceID:       ref.Ref,
				RelevanceSummary: ref.Summary,
				EvidenceRef:      ref,
			})
		}
		return out
	}
	var out []knowledge.EvidenceLink
	if decodeCommandPayload(raw, &out) {
		return out
	}
	var single knowledge.EvidenceLink
	if decodeCommandPayload(raw, &single) {
		return []knowledge.EvidenceLink{single}
	}
	return nil
}

func reasoningStepsFromCommand(command transition.CommandEnvelope, key string) []events.ReasoningStep {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case []events.ReasoningStep:
		out := make([]events.ReasoningStep, len(value))
		copy(out, value)
		return out
	case events.ReasoningStep:
		return []events.ReasoningStep{value}
	}
	var out []events.ReasoningStep
	if decodeCommandPayload(raw, &out) {
		return out
	}
	var single events.ReasoningStep
	if decodeCommandPayload(raw, &single) {
		return []events.ReasoningStep{single}
	}
	return nil
}

func toolCallRecordsFromCommand(command transition.CommandEnvelope, key string) []events.ToolCallRecord {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case []events.ToolCallRecord:
		out := make([]events.ToolCallRecord, len(value))
		copy(out, value)
		return out
	case events.ToolCallRecord:
		return []events.ToolCallRecord{value}
	}
	var out []events.ToolCallRecord
	if decodeCommandPayload(raw, &out) {
		return out
	}
	var single events.ToolCallRecord
	if decodeCommandPayload(raw, &single) {
		return []events.ToolCallRecord{single}
	}
	return nil
}

func slackActionsFromCommand(command transition.CommandEnvelope, key string) []events.SlackActionRecord {
	raw, ok := command.Payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case []events.SlackActionRecord:
		out := make([]events.SlackActionRecord, len(value))
		copy(out, value)
		return out
	case events.SlackActionRecord:
		return []events.SlackActionRecord{value}
	}
	var out []events.SlackActionRecord
	if decodeCommandPayload(raw, &out) {
		return out
	}
	var single events.SlackActionRecord
	if decodeCommandPayload(raw, &single) {
		return []events.SlackActionRecord{single}
	}
	return nil
}

func optionalStringFromCommand(command transition.CommandEnvelope, key string) *string {
	value := strings.TrimSpace(stringFromCommand(command, key))
	if value == "" {
		return nil
	}
	return &value
}

func findWorkflowByID(items []Workflow, workflowID string) (Workflow, bool) {
	for _, item := range items {
		if item.ID == workflowID {
			return item, true
		}
	}
	return Workflow{}, false
}

func findIngestion(items []slack.Ingestion, ingestionID string) (slack.Ingestion, bool) {
	for _, item := range items {
		if item.ID == ingestionID {
			return item, true
		}
	}
	return slack.Ingestion{}, false
}

func (s *MemoryStore) setThreadStateLocked(threadKey string, state policy.ThreadState, owner string) (policy.ThreadPolicy, error) {
	item, ok := s.threadPolicies[threadKey]
	if !ok {
		return policy.ThreadPolicy{}, errors.New("thread policy not found")
	}
	item.State = state
	item.Muted = state == policy.ThreadStateMuted || state == policy.ThreadStateMuteUntilMention
	if owner != "" {
		item.OwnerBot = owner
	}
	item.UpdatedAt = time.Now().UTC()
	s.threadPolicies[threadKey] = item
	return item, nil
}

func (s *MemoryStore) setActionExecutionStateLocked(actionID string, state action.Status, command transition.CommandEnvelope) (action.Intent, error) {
	intent, ok := s.actionIntents[actionID]
	if !ok {
		return action.Intent{}, errors.New("action intent not found")
	}
	intent.Status = state
	intent.OperationID = firstNonEmpty(stringFromCommand(command, "operation_id"), intent.OperationID)
	intent.ApprovalState = firstNonEmpty(stringFromCommand(command, "approval_state"), intent.ApprovalState)
	intent.PolicyVerdict = firstNonEmpty(stringFromCommand(command, "policy_verdict"), intent.PolicyVerdict)
	intent.UpdatedAt = command.OccurredAt
	s.actionIntents[actionID] = intent
	return intent, nil
}

func (s *MemoryStore) setWorkflowMachineStateLocked(workflowID string, state transition.WorkflowStateKind, lastError string, updatedAt time.Time) (Workflow, error) {
	for i := range s.workflows {
		if s.workflows[i].ID != workflowID {
			continue
		}
		s.workflows[i].Status = workflowStatusFromState(state)
		s.workflows[i].LastError = strings.TrimSpace(lastError)
		s.workflows[i].UpdatedAt = updatedAt
		s.workflows[i].Version++
		switch state {
		case transition.WorkflowStateCompleted, transition.WorkflowStateFailed, transition.WorkflowStateNeedsHuman, transition.WorkflowStateSuperseded:
			completedAt := updatedAt
			s.workflows[i].CompletedAt = &completedAt
		default:
			s.workflows[i].CompletedAt = nil
		}
		return s.workflows[i], nil
	}
	return Workflow{}, errors.New("workflow not found")
}

func timeFromCommand(command transition.CommandEnvelope, key string, fallback time.Time) time.Time {
	raw, ok := command.Payload[key]
	if !ok {
		return fallback
	}
	switch value := raw.(type) {
	case time.Time:
		if !value.IsZero() {
			return value
		}
	case string:
		if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value)); err == nil && !parsed.IsZero() {
			return parsed
		}
	}
	return fallback
}

func attemptTraceProjectionID(command transition.CommandEnvelope, attempt improvement.ChangeAttempt) string {
	if attempt.AttemptTraceID == "" {
		return ""
	}
	if len(traceEventsFromCommand(command, "trace_events")) == 0 &&
		len(reasoningStepsFromCommand(command, "reasoning_steps")) == 0 &&
		len(traceArtifactsFromCommand(command, "trace_artifacts")) == 0 {
		return ""
	}
	return attempt.AttemptTraceID
}

func attemptWorkspaceForCommandLocked(items map[string]improvement.AttemptWorkspace, attemptID string, workspaceID string) (improvement.AttemptWorkspace, bool) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID != "" {
		item, ok := items[workspaceID]
		return item, ok
	}
	attemptID = strings.TrimSpace(attemptID)
	for _, item := range items {
		if strings.TrimSpace(item.AttemptID) == attemptID {
			return item, true
		}
	}
	return improvement.AttemptWorkspace{}, false
}

func decodeCommandPayload(raw any, target any) bool {
	data, err := json.Marshal(raw)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, target) == nil
}

func (s *MemoryStore) updateSettingsLocked(settings improvement.Settings) (improvement.Settings, error) {
	settings = normalizedSettings(settings)
	settings.UpdatedAt = time.Now().UTC()
	s.settings = settings
	return settings, nil
}

func (s *MemoryStore) reviewKnowledgeEntryLocked(knowledgeID string, item knowledge.Review) (knowledge.Entry, error) {
	entry, ok := s.knowledgeEntries[knowledgeID]
	if !ok {
		return knowledge.Entry{}, errors.New("knowledge entry not found")
	}
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextID("knowledge-review", len(s.knowledgeReviews[knowledgeID])+1)
	}
	item.KnowledgeEntryID = knowledgeID
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	s.knowledgeReviews[knowledgeID] = append(s.knowledgeReviews[knowledgeID], item)
	entry.UpdatedAt = item.CreatedAt
	switch strings.ToLower(strings.TrimSpace(item.Decision)) {
	case "approve":
		entry.Tier = knowledge.TierCanonical
		entry.Status = knowledge.StatusCanonical
		entry.SourceType = knowledge.SourcePromoted
	case "reject":
		entry.Status = knowledge.StatusDraft
	case "mark_stale":
		entry.Status = knowledge.StatusStale
	case "archive":
		entry.Status = knowledge.StatusArchived
	default:
		return knowledge.Entry{}, errors.New("unsupported knowledge review decision")
	}
	s.knowledgeEntries[knowledgeID] = entry
	return entry, nil
}

func (s *MemoryStore) reviewProposalLocked(proposalID string, decision review.ProposalReview) (review.Proposal, error) {
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return review.Proposal{}, errors.New("proposal not found")
	}
	decision.ProposalID = proposalID
	decision.IdempotencyKey = firstNonEmpty(strings.TrimSpace(decision.IdempotencyKey), proposalDecisionIdempotencyKey(proposalID, decision.Decision, decision.Scope))
	if decision.Scope == "" {
		decision.Scope = review.FeedbackScopeLine
	}
	for _, existing := range proposal.Reviews {
		if existing.IdempotencyKey == decision.IdempotencyKey {
			return proposal, nil
		}
	}
	decision.ID = int64(len(proposal.Reviews) + 1)
	if decision.CreatedAt.IsZero() {
		decision.CreatedAt = time.Now().UTC()
	}
	if len(decision.FailureClasses) == 0 && decision.FailureClass != "" {
		decision.FailureClasses = []string{decision.FailureClass}
	}
	proposal.Reviews = append(proposal.Reviews, decision)
	proposal.Reviewer = decision.ReviewerID
	proposal.Status = review.ProposalStatus(decision.Decision)
	proposal.ActiveSlotConsuming = review.ConsumesActiveProposalSlot(proposal.Status)
	if proposal.Status == review.ProposalApproved && proposal.AutoRetryBudgetRemaining == 0 {
		proposal.AutoRetryBudgetRemaining = defaultProposalRetryBudget
	}
	if proposal.Status == review.ProposalRejected || proposal.Status == review.ProposalDismissed {
		proposal.NextRetryAction = ""
	}
	if proposal.Version == 0 {
		proposal.Version = 1
	} else {
		proposal.Version++
	}
	s.proposals[proposalID] = proposal

	s.proposalMemory = append(s.proposalMemory, review.ProposalMemory{
		ID:                nextID("memory", len(s.proposalMemory)+1),
		ReviewID:          decision.ID,
		ProposalID:        proposalID,
		CandidateKey:      proposal.CandidateKey,
		ConversationID:    proposal.ConversationID,
		CaseID:            proposal.CaseID,
		OriginTraceID:     proposal.OriginTraceID,
		EvidenceTraceIDs:  append([]string(nil), proposal.EvidenceTraceIDs...),
		Hypothesis:        proposal.Summary,
		DiffSummary:       proposal.ProposedScope,
		ReviewRationale:   decision.Rationale,
		Disposition:       proposal.Status,
		DispositionReason: decision.Rationale,
		FailureClass:      decision.FailureClass,
		FailureClasses:    append([]string(nil), decision.FailureClasses...),
		SourceEvalIDs:     append([]string(nil), proposal.SourceEvalIDs...),
		LinkedArtifactIDs: append([]string(nil), proposal.EvidenceArtifactIDs...),
		LinkedProposalIDs: append([]string(nil), proposal.PriorSimilarProposalIDs...),
		CreatedAt:         decision.CreatedAt,
	})

	candidate := s.candidates[proposal.CandidateKey]
	switch proposal.Status {
	case review.ProposalApproved:
		candidate.Status = improvement.CandidatePromoted
		candidate.LineStatus = improvement.LineActive
		candidate.AutoRetryBudgetRemaining = proposal.AutoRetryBudgetRemaining
		candidate.CurrentTargetLayer = proposal.TargetLayer
	case review.ProposalRejected, review.ProposalDismissed:
		candidate.Status = improvement.CandidateNeedsEvidence
		candidate.LineStatus = improvement.LineClosed
		candidate.AutoRetryBudgetRemaining = 0
		candidate.NewEvidenceSinceLastRejection = false
	case review.ProposalMerged:
		candidate.Status = improvement.CandidateDormant
		candidate.LineStatus = improvement.LineClosed
		candidate.AutoRetryBudgetRemaining = 0
		replayID := nextID("pmr", len(s.postMergeReplay)+1)
		s.postMergeReplay[replayID] = improvement.PostMergeReplay{
			ID:             replayID,
			ProposalID:     proposal.ID,
			TraceID:        proposal.TraceID,
			ConversationID: proposal.ConversationID,
			CaseID:         proposal.CaseID,
			BaselineScore:  latestEvalScoreForTrace(s.evalRuns, proposal.TraceID),
			CandidateScore: minFloat(1.0, latestEvalScoreForTrace(s.evalRuns, proposal.TraceID)+0.15),
			Improved:       true,
			CreatedAt:      decision.CreatedAt,
		}
	default:
		candidate.Status = improvement.CandidateDormant
		if candidate.LineStatus == "" {
			candidate.LineStatus = improvement.LineDormant
		}
	}
	candidate.UpdatedAt = decision.CreatedAt
	s.candidates[proposal.CandidateKey] = candidate
	return proposal, nil
}

func (s *MemoryStore) stopProposalLineLocked(proposalID string, requestedBy string, rationale string) (review.Proposal, error) {
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return review.Proposal{}, errors.New("proposal not found")
	}
	now := time.Now().UTC()
	proposal.Status = review.ProposalCanceled
	proposal.ActiveSlotConsuming = false
	proposal.NextRetryAction = ""
	proposal.LineStoppedBy = requestedBy
	proposal.LineStopReason = rationale
	proposal.LineStoppedAt = &now
	if proposal.Version == 0 {
		proposal.Version = 1
	} else {
		proposal.Version++
	}
	s.proposals[proposal.ID] = proposal
	if candidate, ok := s.candidates[proposal.CandidateKey]; ok {
		candidate.LineStatus = improvement.LineClosed
		candidate.AutoRetryBudgetRemaining = 0
		candidate.UpdatedAt = now
		s.candidates[candidate.CandidateKey] = candidate
	}
	if proposal.CurrentAttemptID != "" {
		if attempt, ok := s.changeAttempts[proposal.CurrentAttemptID]; ok && !isTerminalAttemptState(attempt.State) {
			attempt.State = improvement.AttemptStateAbandoned
			attempt.FailureClass = firstNonEmpty(attempt.FailureClass, "stopped_by_operator")
			attempt.FailureSummary = firstNonEmpty(attempt.FailureSummary, rationale)
			attempt.RetryDecision = "stop_line"
			attempt.UpdatedAt = now
			if attempt.Version == 0 {
				attempt.Version = 1
			} else {
				attempt.Version++
			}
			s.changeAttempts[attempt.ID] = normalizeChangeAttempt(attempt)
		}
	}
	s.proposalMemory = append(s.proposalMemory, review.ProposalMemory{
		ID:                nextID("memory", len(s.proposalMemory)+1),
		ProposalID:        proposal.ID,
		CandidateKey:      proposal.CandidateKey,
		ConversationID:    proposal.ConversationID,
		CaseID:            proposal.CaseID,
		OriginTraceID:     proposal.OriginTraceID,
		EvidenceTraceIDs:  append([]string(nil), proposal.EvidenceTraceIDs...),
		Hypothesis:        proposal.Summary,
		DiffSummary:       proposal.ProposedScope,
		ReviewRationale:   firstNonEmpty(rationale, "Line stopped by operator."),
		Disposition:       review.ProposalCanceled,
		DispositionReason: firstNonEmpty(rationale, "Line stopped by operator."),
		FailureClass:      "stopped_by_operator",
		LinkedProposalIDs: append([]string(nil), proposal.PriorSimilarProposalIDs...),
		CreatedAt:         now,
	})
	return proposal, nil
}
