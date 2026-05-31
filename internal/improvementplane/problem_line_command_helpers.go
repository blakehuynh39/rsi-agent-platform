package improvementplane

import (
	"errors"
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func loadFeedbackFromReceipt(store storepkg.Store, receipt transition.CommandReceipt) (review.FeedbackRecord, error) {
	feedbackID := strings.TrimSpace(receipt.ResultRef)
	traceID := strings.TrimSpace(receipt.AggregateID)
	if feedbackID == "" {
		return review.FeedbackRecord{}, errors.New("missing feedback result ref")
	}
	if traceID == "" {
		return review.FeedbackRecord{}, errors.New("missing feedback trace ref")
	}
	for _, item := range store.ListFeedback(traceID) {
		if item.ID == feedbackID {
			return item, nil
		}
	}
	return review.FeedbackRecord{}, fmt.Errorf("feedback %s not found", feedbackID)
}

func resolveFeedbackTargetTraceID(store storepkg.Store, record review.FeedbackRecord) (string, error) {
	if traceID := strings.TrimSpace(record.TraceID); traceID != "" {
		if _, ok := store.GetTrace(traceID); ok {
			return traceID, nil
		}
		return "", errors.New("trace not found")
	}
	switch record.TargetType {
	case review.FeedbackTargetConversation:
		if _, ok := store.GetConversation(record.TargetID); !ok {
			return "", errors.New("conversation not found")
		}
		return latestTraceID(store, func(summary traceSummaryLike) bool {
			return summary.conversationID == record.TargetID
		}, "conversation")
	case review.FeedbackTargetCase:
		if _, ok := store.GetCase(record.TargetID); !ok {
			return "", errors.New("case not found")
		}
		return latestTraceID(store, func(summary traceSummaryLike) bool {
			return summary.caseID == record.TargetID
		}, "case")
	case review.FeedbackTargetProposal:
		for _, proposal := range store.ListProposals() {
			if proposal.ID != record.TargetID {
				continue
			}
			traceID := firstProblemLineTraceID(proposal.OriginTraceID, proposal.TraceID)
			if traceID == "" {
				return "", errors.New("trace not found for proposal")
			}
			if _, ok := store.GetTrace(traceID); !ok {
				return "", errors.New("trace not found for proposal")
			}
			return traceID, nil
		}
		return "", errors.New("proposal not found")
	case review.FeedbackTargetActionIntent:
		for _, intent := range store.ListActionIntents() {
			if intent.ID != record.TargetID {
				continue
			}
			switch {
			case strings.TrimSpace(intent.TraceID) != "":
				if _, ok := store.GetTrace(intent.TraceID); ok {
					return intent.TraceID, nil
				}
			case strings.TrimSpace(intent.ProposalID) != "":
				return resolveFeedbackTargetTraceID(store, review.FeedbackRecord{
					TargetType: review.FeedbackTargetProposal,
					TargetID:   intent.ProposalID,
				})
			case strings.TrimSpace(intent.CaseID) != "":
				return resolveFeedbackTargetTraceID(store, review.FeedbackRecord{
					TargetType: review.FeedbackTargetCase,
					TargetID:   intent.CaseID,
				})
			case strings.TrimSpace(intent.ConversationID) != "":
				return resolveFeedbackTargetTraceID(store, review.FeedbackRecord{
					TargetType: review.FeedbackTargetConversation,
					TargetID:   intent.ConversationID,
				})
			}
			return "", errors.New("trace not found")
		}
		return "", errors.New("action intent not found")
	case review.FeedbackTargetTrace:
		if _, ok := store.GetTrace(record.TargetID); ok {
			return record.TargetID, nil
		}
		return "", errors.New("trace not found")
	case review.FeedbackTargetReasoning, review.FeedbackTargetToolCall, review.FeedbackTargetSlackAction:
		for _, summary := range store.ListTraces() {
			trace, ok := store.GetTrace(summary.TraceID)
			if !ok {
				continue
			}
			switch record.TargetType {
			case review.FeedbackTargetReasoning:
				for _, step := range trace.Reasoning {
					if step.ID == record.TargetID {
						return trace.Summary.TraceID, nil
					}
				}
			case review.FeedbackTargetToolCall:
				for _, call := range trace.ToolCalls {
					if call.ID == record.TargetID || call.ToolCallID == record.TargetID {
						return trace.Summary.TraceID, nil
					}
				}
			case review.FeedbackTargetSlackAction:
				for _, action := range trace.SlackActions {
					if action.ID == record.TargetID {
						return trace.Summary.TraceID, nil
					}
				}
			}
		}
		return "", errors.New("trace not found")
	default:
		return "", errors.New("unsupported feedback target")
	}
}

type traceSummaryLike struct {
	conversationID string
	caseID         string
	traceID        string
	startedAtUnix  int64
}

func latestTraceID(store storepkg.Store, match func(traceSummaryLike) bool, scope string) (string, error) {
	var current traceSummaryLike
	found := false
	for _, summary := range store.ListTraces() {
		item := traceSummaryLike{
			conversationID: summary.ConversationID,
			caseID:         summary.CaseID,
			traceID:        summary.TraceID,
			startedAtUnix:  summary.StartedAt.UnixNano(),
		}
		if !match(item) {
			continue
		}
		if !found || item.startedAtUnix > current.startedAtUnix || (item.startedAtUnix == current.startedAtUnix && item.traceID > current.traceID) {
			current = item
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("trace not found for %s", scope)
	}
	return current.traceID, nil
}

func firstProblemLineTraceID(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
