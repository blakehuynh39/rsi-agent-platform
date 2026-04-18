package store

import (
	"errors"
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/questionrun"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func (s *MemoryStore) applyQuestionRunCommandLocked(command transition.CommandEnvelope) (commandApplyResult, error) {
	questionRun, exists := s.questionRuns[command.AggregateID]
	snapshot := transition.QuestionRunSnapshot{}
	if exists {
		snapshot.State = transition.QuestionRunState(questionRun.Status)
	}
	decision := transition.ReduceQuestionRun(snapshot, command)
	version := int64(1)
	if exists {
		version = questionRun.Version
	}
	if decision.DecisionKind == transition.DecisionAdvance {
		now := command.OccurredAt
		if !exists {
			questionRun = QuestionRun{
				ID:        strings.TrimSpace(command.AggregateID),
				CreatedAt: now,
			}
			version = 0
		}
		version++
		questionRun.Version = version
		questionRun.Status = string(decision.NextState)
		questionRun.WorkflowID = firstNonEmpty(stringFromCommand(command, "workflow_id"), questionRun.WorkflowID)
		questionRun.TraceID = firstNonEmpty(stringFromCommand(command, "trace_id"), questionRun.TraceID)
		questionRun.ConversationID = firstNonEmpty(stringFromCommand(command, "conversation_id"), questionRun.ConversationID)
		questionRun.CaseID = firstNonEmpty(stringFromCommand(command, "case_id"), questionRun.CaseID)
		questionRun.IngestionID = firstNonEmpty(stringFromCommand(command, "ingestion_id"), questionRun.IngestionID)
		questionRun.Role = firstNonEmpty(stringFromCommand(command, "role"), questionRun.Role)
		questionRun.Strategy = firstNonEmpty(stringFromCommand(command, "strategy"), questionRun.Strategy)
		if spec, ok := questionRunInvestigationSpecFromCommand(command, "investigation_spec"); ok {
			questionRun.InvestigationSpec = spec
			questionRun.EvidenceLedger = firstNonZeroEvidenceLedger(questionRun.EvidenceLedger, evidenceLedgerFromInvestigationSpec(spec))
		}
		if ledger, ok := questionRunEvidenceLedgerFromCommand(command, "evidence_ledger"); ok {
			questionRun.EvidenceLedger = ledger
		}
		if ledger, ok := questionRunAlignmentLedgerFromCommand(command, "alignment_ledger"); ok {
			evidenceLedger := questionRun.EvidenceLedger
			evidenceLedger.AlignmentLedger = ledger
			evidenceLedger.AlignmentRequired = evidenceLedger.AlignmentRequired || questionRun.InvestigationSpec.AlignmentRequired
			evidenceLedger.AlignmentDegraded = ledger.Degraded
			questionRun.EvidenceLedger = evidenceLedger
		}
		if result, ok := questionRunResultFromCommand(command, "result"); ok {
			questionRun.Result = result
		}
		if failureClass := strings.TrimSpace(stringFromCommand(command, "failure_class")); failureClass != "" {
			questionRun.FailureClass = failureClass
		}
		if failureSummary := strings.TrimSpace(stringFromCommand(command, "failure_summary")); failureSummary != "" {
			questionRun.FailureSummary = failureSummary
		}
		if lastError := strings.TrimSpace(stringFromCommand(command, "last_error")); lastError != "" {
			questionRun.LastError = lastError
		}
		if _, ok := command.Payload["runner_diagnostics"]; ok {
			questionRun.RunnerDiagnostics = anyMapFromCommand(command, "runner_diagnostics")
		}
		questionRun.UpdatedAt = now
		if isTerminalQuestionRunState(decision.NextState) {
			questionRun.CompletedAt = &now
		} else {
			questionRun.CompletedAt = nil
		}
		s.questionRuns[questionRun.ID] = questionRun
	}
	result := commandApplyResult{
		receipt: buildCommandReceipt(command, decision.TransitionDecision, firstNonZeroTime(optionalTime(questionRun.UpdatedAt), command.OccurredAt), version, questionRun.ID),
		bundle:  buildCommandBundle(command, decision.TransitionDecision, version),
	}
	if decision.DecisionKind == transition.DecisionAdvance {
		result.questionRunID = questionRun.ID
		s.appendWorkflowFollowOnCommandFromQuestionRunLocked(&result.bundle, command, questionRun)
	}
	if decision.DecisionKind == transition.DecisionReject && !exists {
		result.receipt = buildCommandReceipt(command, decision.TransitionDecision, command.OccurredAt, 1, "")
		return result, nil
	}
	if questionRun.ID == "" && exists {
		return commandApplyResult{}, errors.New("question run not found after command application")
	}
	return result, nil
}

func (s *MemoryStore) appendWorkflowFollowOnCommandFromQuestionRunLocked(bundle *transitionPersistBundle, parent transition.CommandEnvelope, item QuestionRun) {
	if bundle == nil || strings.TrimSpace(item.WorkflowID) == "" {
		return
	}
	var next transition.WorkflowCommandKind
	switch transition.QuestionRunCommandKind(parent.CommandKind) {
	case transition.CommandReplyReduced:
		next = workflowExecutionCommandFromQuestionRun(parent, false)
	case transition.CommandReplyReducedPartial:
		next = workflowExecutionCommandFromQuestionRun(parent, true)
	case transition.CommandReplyBlocked:
		next = transition.CommandWorkflowExecutionNeedsHuman
	case transition.CommandQuestionRunFailed:
		next = transition.CommandWorkflowExecutionFailed
	default:
		return
	}
	appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: item.WorkflowID,
		CommandKind: string(next),
		CommandID:   fmt.Sprintf("%s:workflow-execution", parent.CommandID),
		Actor:       parent.Actor,
		OccurredAt:  parent.OccurredAt,
	}, "question_run emitted terminal workflow execution result")
}

func workflowExecutionCommandFromQuestionRun(command transition.CommandEnvelope, partial bool) transition.WorkflowCommandKind {
	hasReply := strings.TrimSpace(stringFromCommand(command, "reply_action_id")) != ""
	verdict := "complete"
	if partial {
		verdict = "partial"
	}
	return transition.WorkflowExecutionCompletionCommand(verdict, hasReply)
}

func (s *MemoryStore) appendQuestionRunFollowOnCommandLocked(bundle *transitionPersistBundle, parent transition.CommandEnvelope, workflow Workflow) {
	if bundle == nil || !workflowUsesQuestionRunStrategy(parent) {
		return
	}
	commandKind := transition.WorkflowCommandKind(parent.CommandKind)
	if commandKind != transition.CommandContextSkipped && commandKind != transition.CommandContextCompleted {
		return
	}
	questionRunID := questionRunIDForWorkflow(workflow.ID)
	appendFollowOnCommand(bundle, parent, transition.CommandEnvelope{
		MachineKind: transition.MachineQuestionRun,
		AggregateID: questionRunID,
		CommandKind: string(transition.CommandQuestionRunStarted),
		CommandID:   fmt.Sprintf("%s:question-run:start", parent.CommandID),
		Actor:       parent.Actor,
		OccurredAt:  parent.OccurredAt,
		Payload: map[string]any{
			"workflow_id":     workflow.ID,
			"trace_id":        workflow.TraceID,
			"conversation_id": workflow.ConversationID,
			"case_id":         workflow.CaseID,
			"ingestion_id":    workflow.IngestionID,
			"strategy":        "read_heavy_slack_qna",
			"role":            firstNonEmpty(stringFromCommand(parent, "execution_role"), stringFromCommand(parent, "role")),
		},
	}, "workflow execution strategy delegated to question_run")
}

func questionRunIDForWorkflow(workflowID string) string {
	return fmt.Sprintf("qrun:%s", strings.TrimSpace(workflowID))
}

func workflowUsesQuestionRunStrategy(command transition.CommandEnvelope) bool {
	return strings.TrimSpace(stringFromCommand(command, "execution_strategy")) == "read_heavy_slack_qna"
}

func questionRunInvestigationSpecFromCommand(command transition.CommandEnvelope, key string) (questionrun.InvestigationSpec, bool) {
	var out questionrun.InvestigationSpec
	if !decodeCommandPayload(command.Payload[key], &out) {
		return questionrun.InvestigationSpec{}, false
	}
	return out, true
}

func questionRunEvidenceLedgerFromCommand(command transition.CommandEnvelope, key string) (questionrun.EvidenceLedger, bool) {
	var out questionrun.EvidenceLedger
	if !decodeCommandPayload(command.Payload[key], &out) {
		return questionrun.EvidenceLedger{}, false
	}
	return out, true
}

func questionRunAlignmentLedgerFromCommand(command transition.CommandEnvelope, key string) (*questionrun.ProjectAlignmentLedger, bool) {
	var out questionrun.ProjectAlignmentLedger
	if !decodeCommandPayload(command.Payload[key], &out) {
		return nil, false
	}
	return &out, true
}

func questionRunResultFromCommand(command transition.CommandEnvelope, key string) (questionrun.Result, bool) {
	var out questionrun.Result
	if !decodeCommandPayload(command.Payload[key], &out) {
		return questionrun.Result{}, false
	}
	return out, true
}

func evidenceLedgerFromInvestigationSpec(spec questionrun.InvestigationSpec) questionrun.EvidenceLedger {
	return questionrun.EvidenceLedger{
		UserRequest:       spec.UserRequest,
		ReplyTarget:       spec.ReplyTarget,
		Repo:              spec.Repo,
		ProjectKey:        spec.ProjectKey,
		Since:             spec.Since,
		Until:             spec.Until,
		AlignmentRequired: spec.AlignmentRequired,
	}
}

func firstNonZeroEvidenceLedger(primary questionrun.EvidenceLedger, fallback questionrun.EvidenceLedger) questionrun.EvidenceLedger {
	if primary.UserRequest != "" || primary.ReplyTarget.ChannelID != "" || primary.ReplyTarget.ThreadTS != "" || primary.Repo != "" || len(primary.EvidenceItems) > 0 || len(primary.ToolCalls) > 0 {
		return primary
	}
	return fallback
}

func isTerminalQuestionRunState(state transition.QuestionRunState) bool {
	switch state {
	case transition.QuestionRunStateCompleted, transition.QuestionRunStateNeedsHuman, transition.QuestionRunStateFailed, transition.QuestionRunStateSuperseded:
		return true
	default:
		return false
	}
}
