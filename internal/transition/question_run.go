package transition

type QuestionRunState string

const (
	QuestionRunStateQueued                    QuestionRunState = "queued"
	QuestionRunStateCompilingSpec             QuestionRunState = "compiling_spec"
	QuestionRunStateGatheringEvidence         QuestionRunState = "gathering_evidence"
	QuestionRunStateReducing                  QuestionRunState = "reducing"
	QuestionRunStateCompleted                 QuestionRunState = "completed"
	QuestionRunStateNeedsHuman                QuestionRunState = "needs_human"
	QuestionRunStateFailed                    QuestionRunState = "failed"
	QuestionRunStateSuperseded                QuestionRunState = "superseded"
	QuestionRunStateRefreshingAlignmentLedger QuestionRunState = "refreshing_alignment_ledger"
	QuestionRunStateCollectingSeedEvidence    QuestionRunState = "collecting_seed_evidence"
	QuestionRunStateExpandingEvidence         QuestionRunState = "expanding_evidence"
)

type QuestionRunCommandKind string

const (
	CommandQuestionRunStarted      QuestionRunCommandKind = "question_run_started"
	CommandInvestigationSpecBuilt  QuestionRunCommandKind = "investigation_spec_built"
	CommandEvidenceGathered        QuestionRunCommandKind = "evidence_gathered"
	CommandEvidenceGatheredPartial QuestionRunCommandKind = "evidence_gathered_partial"
	CommandReplyReduced            QuestionRunCommandKind = "reply_reduced"
	CommandReplyReducedPartial     QuestionRunCommandKind = "reply_reduced_partial"
	CommandReplyBlocked            QuestionRunCommandKind = "reply_blocked"
	CommandQuestionRunFailed       QuestionRunCommandKind = "question_run_failed"
	CommandQuestionRunSuperseded   QuestionRunCommandKind = "question_run_superseded"
	CommandAlignmentLedgerReady    QuestionRunCommandKind = "alignment_ledger_ready"
	CommandAlignmentLedgerDegraded QuestionRunCommandKind = "alignment_ledger_degraded"
	CommandSeedEvidenceCollected   QuestionRunCommandKind = "seed_evidence_collected"
	CommandEvidenceExpanded        QuestionRunCommandKind = "evidence_expanded"
)

type QuestionRunSnapshot struct {
	State QuestionRunState `json:"state,omitempty"`
}

type QuestionRunDecision struct {
	TransitionDecision
	NextState QuestionRunState `json:"next_state,omitempty"`
}

func ReduceQuestionRun(snapshot QuestionRunSnapshot, command CommandEnvelope) QuestionRunDecision {
	current := snapshot.State
	if current == "" {
		current = QuestionRunStateQueued
	}
	switch current {
	case QuestionRunStateQueued:
		switch QuestionRunCommandKind(command.CommandKind) {
		case CommandQuestionRunStarted:
			return questionRunAdvance(current, QuestionRunStateCompilingSpec, "question run started and investigation spec compilation queued", "question_run_started", EffectCompileInvestigationSpec, "compile_investigation_spec")
		case CommandQuestionRunFailed:
			return questionRunFailure(current)
		case CommandQuestionRunSuperseded:
			return questionRunSuperseded(current)
		}
	case QuestionRunStateCompilingSpec:
		switch QuestionRunCommandKind(command.CommandKind) {
		case CommandInvestigationSpecBuilt:
			return questionRunAdvance(current, QuestionRunStateGatheringEvidence, "investigation spec built and evidence gathering queued", "question_run_investigation_spec_built", EffectGatherEvidence, "gather_evidence")
		case CommandQuestionRunFailed:
			return questionRunFailure(current)
		case CommandQuestionRunSuperseded:
			return questionRunSuperseded(current)
		}
	case QuestionRunStateGatheringEvidence:
		switch QuestionRunCommandKind(command.CommandKind) {
		case CommandEvidenceGathered, CommandEvidenceGatheredPartial:
			eventKind := "question_run_evidence_gathered"
			reason := "evidence gathered and reducer queued"
			if QuestionRunCommandKind(command.CommandKind) == CommandEvidenceGatheredPartial {
				eventKind = "question_run_evidence_gathered_partial"
				reason = "evidence gathered partially and reducer queued"
			}
			return questionRunAdvance(current, QuestionRunStateReducing, reason, eventKind, EffectReduceReply, "reduce_reply")
		case CommandQuestionRunFailed:
			return questionRunFailure(current)
		case CommandQuestionRunSuperseded:
			return questionRunSuperseded(current)
		}
	case QuestionRunStateReducing:
		switch QuestionRunCommandKind(command.CommandKind) {
		case CommandReplyReduced, CommandReplyReducedPartial:
			eventKind := "question_run_reply_reduced"
			if QuestionRunCommandKind(command.CommandKind) == CommandReplyReducedPartial {
				eventKind = "question_run_reply_reduced_partial"
			}
			return QuestionRunDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "question run reduced a final reply",
					Events: []DomainEventDescriptor{{
						Kind: eventKind,
					}},
				},
				NextState: QuestionRunStateCompleted,
			}
		case CommandReplyBlocked:
			return QuestionRunDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "question run produced a reply but workflow policy requires human handling",
					Events: []DomainEventDescriptor{{
						Kind: "question_run_reply_blocked",
					}},
				},
				NextState: QuestionRunStateNeedsHuman,
			}
		case CommandQuestionRunFailed:
			return questionRunFailure(current)
		case CommandQuestionRunSuperseded:
			return questionRunSuperseded(current)
		}
	case QuestionRunStateCompleted, QuestionRunStateNeedsHuman, QuestionRunStateFailed:
		if QuestionRunCommandKind(command.CommandKind) == CommandQuestionRunSuperseded {
			return questionRunSuperseded(current)
		}
	case QuestionRunStateSuperseded:
	}
	return QuestionRunDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       "unsupported question run command for current state",
		},
		NextState: current,
	}
}

func questionRunAdvance(current QuestionRunState, next QuestionRunState, reason string, eventKind string, effectKind EffectKind, idempotencyKey string) QuestionRunDecision {
	return QuestionRunDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: eventKind,
			}},
			Effects: []EffectRequest{{
				Kind:           effectKind,
				Status:         EffectQueued,
				IdempotencyKey: idempotencyKey,
			}},
		},
		NextState: next,
	}
}

func questionRunFailure(current QuestionRunState) QuestionRunDecision {
	return QuestionRunDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "question run failed",
			Events: []DomainEventDescriptor{{
				Kind: "question_run_failed",
			}},
		},
		NextState: QuestionRunStateFailed,
	}
}

func questionRunSuperseded(current QuestionRunState) QuestionRunDecision {
	return QuestionRunDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "question run superseded",
			Events: []DomainEventDescriptor{{
				Kind: "question_run_superseded",
			}},
		},
		NextState: QuestionRunStateSuperseded,
	}
}

func commandPayloadBool(command CommandEnvelope, key string) bool {
	if command.Payload == nil {
		return false
	}
	value, ok := command.Payload[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch typed {
		case "true", "1", "yes":
			return true
		}
	}
	return false
}
