package transition

import "github.com/piplabs/rsi-agent-platform/internal/knowledge"

type KnowledgeCommandKind string

const (
	CommandKnowledgeRecordDraft KnowledgeCommandKind = "knowledge_record_draft"
	CommandKnowledgeApprove     KnowledgeCommandKind = "knowledge_approve"
	CommandKnowledgeReject      KnowledgeCommandKind = "knowledge_reject"
	CommandKnowledgeMarkStale   KnowledgeCommandKind = "knowledge_mark_stale"
	CommandKnowledgeArchive     KnowledgeCommandKind = "knowledge_archive"
)

type KnowledgeSnapshot struct {
	Exists bool             `json:"exists"`
	Status knowledge.Status `json:"status"`
}

type KnowledgeDecision struct {
	TransitionDecision
	NextStatus knowledge.Status `json:"next_status,omitempty"`
}

func ReduceKnowledge(snapshot KnowledgeSnapshot, command CommandEnvelope) KnowledgeDecision {
	commandKind := KnowledgeCommandKind(command.CommandKind)
	if !snapshot.Exists {
		switch commandKind {
		case CommandKnowledgeRecordDraft:
			return advanceKnowledge(snapshot.Status, knowledge.StatusDraft, "knowledge draft recorded", "knowledge_draft_recorded")
		default:
			return KnowledgeDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionReject,
					Reason:       "knowledge entry not found for command",
				},
				NextStatus: snapshot.Status,
			}
		}
	}
	switch snapshot.Status {
	case knowledge.StatusDraft, knowledge.StatusReviewPending:
		switch commandKind {
		case CommandKnowledgeRecordDraft:
			return advanceKnowledge(snapshot.Status, knowledge.StatusDraft, "knowledge draft refreshed", "knowledge_draft_recorded")
		case CommandKnowledgeApprove:
			return advanceKnowledge(snapshot.Status, knowledge.StatusCanonical, "knowledge entry promoted to canonical", "knowledge_promoted")
		case CommandKnowledgeReject:
			return noopOrAdvanceKnowledge(snapshot.Status, knowledge.StatusDraft, "knowledge review kept the entry as draft", "knowledge_rejected")
		case CommandKnowledgeMarkStale:
			return advanceKnowledge(snapshot.Status, knowledge.StatusStale, "knowledge entry marked stale", "knowledge_marked_stale")
		case CommandKnowledgeArchive:
			return advanceKnowledge(snapshot.Status, knowledge.StatusArchived, "knowledge entry archived", "knowledge_archived")
		}
	case knowledge.StatusCanonical:
		switch commandKind {
		case CommandKnowledgeRecordDraft:
			return KnowledgeDecision{TransitionDecision: TransitionDecision{DecisionKind: DecisionReject, Reason: "cannot overwrite canonical knowledge via draft record command"}, NextStatus: snapshot.Status}
		case CommandKnowledgeApprove:
			return KnowledgeDecision{TransitionDecision: TransitionDecision{DecisionKind: DecisionNoop, Reason: "knowledge entry already canonical"}, NextStatus: snapshot.Status}
		case CommandKnowledgeReject:
			return advanceKnowledge(snapshot.Status, knowledge.StatusDraft, "knowledge entry returned to draft", "knowledge_rejected")
		case CommandKnowledgeMarkStale:
			return advanceKnowledge(snapshot.Status, knowledge.StatusStale, "knowledge entry marked stale", "knowledge_marked_stale")
		case CommandKnowledgeArchive:
			return advanceKnowledge(snapshot.Status, knowledge.StatusArchived, "knowledge entry archived", "knowledge_archived")
		}
	case knowledge.StatusStale:
		switch commandKind {
		case CommandKnowledgeRecordDraft:
			return KnowledgeDecision{TransitionDecision: TransitionDecision{DecisionKind: DecisionReject, Reason: "cannot overwrite stale knowledge via draft record command"}, NextStatus: snapshot.Status}
		case CommandKnowledgeApprove:
			return advanceKnowledge(snapshot.Status, knowledge.StatusCanonical, "stale knowledge entry restored to canonical", "knowledge_promoted")
		case CommandKnowledgeReject:
			return advanceKnowledge(snapshot.Status, knowledge.StatusDraft, "stale knowledge entry returned to draft", "knowledge_rejected")
		case CommandKnowledgeMarkStale:
			return KnowledgeDecision{TransitionDecision: TransitionDecision{DecisionKind: DecisionNoop, Reason: "knowledge entry already stale"}, NextStatus: snapshot.Status}
		case CommandKnowledgeArchive:
			return advanceKnowledge(snapshot.Status, knowledge.StatusArchived, "knowledge entry archived", "knowledge_archived")
		}
	case knowledge.StatusContradicted:
		switch commandKind {
		case CommandKnowledgeRecordDraft:
			return KnowledgeDecision{TransitionDecision: TransitionDecision{DecisionKind: DecisionReject, Reason: "cannot overwrite contradicted knowledge via draft record command"}, NextStatus: snapshot.Status}
		case CommandKnowledgeArchive:
			return advanceKnowledge(snapshot.Status, knowledge.StatusArchived, "contradicted knowledge entry archived", "knowledge_archived")
		}
	case knowledge.StatusArchived:
		switch commandKind {
		case CommandKnowledgeRecordDraft:
			return KnowledgeDecision{TransitionDecision: TransitionDecision{DecisionKind: DecisionReject, Reason: "cannot overwrite archived knowledge via draft record command"}, NextStatus: snapshot.Status}
		case CommandKnowledgeArchive:
			return KnowledgeDecision{TransitionDecision: TransitionDecision{DecisionKind: DecisionNoop, Reason: "knowledge entry already archived"}, NextStatus: snapshot.Status}
		}
	}
	return KnowledgeDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       "unsupported knowledge review command for current state",
		},
		NextStatus: snapshot.Status,
	}
}

func advanceKnowledge(current knowledge.Status, next knowledge.Status, reason string, eventKind string) KnowledgeDecision {
	return KnowledgeDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: eventKind,
			}},
		},
		NextStatus: next,
	}
}

func noopOrAdvanceKnowledge(current knowledge.Status, next knowledge.Status, reason string, eventKind string) KnowledgeDecision {
	if current == next {
		return KnowledgeDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionNoop,
				Reason:       reason,
			},
			NextStatus: next,
		}
	}
	return advanceKnowledge(current, next, reason, eventKind)
}
