package store

import (
	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

type Repository = Store

func newEmptyMemoryStore() *MemoryStore {
	return &MemoryStore{
		threadPolicies:    map[string]policy.ThreadPolicy{},
		conversations:     map[string]conversation.Conversation{},
		cases:             map[string]conversation.Case{},
		feedbackRecords:   map[string][]review.FeedbackRecord{},
		actionIntents:     map[string]action.Intent{},
		actionResults:     map[string][]action.Result{},
		outcomes:          map[string]outcome.Record{},
		knowledgeEntries:  map[string]knowledge.Entry{},
		knowledgeEvidence: map[string][]knowledge.EvidenceLink{},
		knowledgeReviews:  map[string][]knowledge.Review{},
		traces:            map[string]events.Trace{},
		ratings:           map[string][]review.HumanRating{},
		notes:             map[string][]review.ImprovementNote{},
		evalRuns:          map[string]evals.Run{},
		evalJudgments:     map[string][]evals.Judgment{},
		candidates:        map[string]improvement.Candidate{},
		proposals:         map[string]review.Proposal{},
		workItems:         map[string]queue.WorkItem{},
		repoChangeJobs:    map[string]improvement.RepoChangeJob{},
		prAttempts:        map[string]improvement.PRAttempt{},
		postMergeReplay:   map[string]improvement.PostMergeReplay{},
		cronLeases:        map[string]improvement.CronLease{},
		settings: improvement.Settings{
			ActiveProposalCap: defaultProposalSlotCap,
		},
	}
}
