package store

import (
	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type Repository = Store

func newEmptyMemoryStore() *MemoryStore {
	return &MemoryStore{
		threadPolicies:                    map[string]policy.ThreadPolicy{},
		sourceMirrorRecords:               map[string]SourceMirrorRecord{},
		conversations:                     map[string]conversation.Conversation{},
		cases:                             map[string]conversation.Case{},
		workflowLines:                     map[string]WorkflowLine{},
		feedbackRecords:                   map[string][]review.FeedbackRecord{},
		actionIntents:                     map[string]action.Intent{},
		actionResults:                     map[string][]action.Result{},
		externalToolActions:               map[string]ExternalToolAction{},
		externalToolActionByIdempotency:   map[string]string{},
		kanbanProjects:                    map[string]KanbanProject{},
		kanbanProjectBySlug:               map[string]string{},
		kanbanBoards:                      map[string]KanbanBoard{},
		kanbanDefaultBoardByProject:       map[string]string{},
		kanbanTickets:                     map[string]KanbanTicket{},
		kanbanComments:                    map[string]KanbanTicketComment{},
		kanbanLinks:                       map[string]KanbanTicketLink{},
		kanbanSourceRefs:                  map[string]KanbanTicketSourceRef{},
		kanbanSlackSourceRefByKey:         map[string]string{},
		kanbanEvents:                      []KanbanTicketEvent{},
		kanbanSlackRoutes:                 map[string]KanbanProjectSlackRoute{},
		kanbanSlackProjectRoutes:          map[string]string{},
		dbReadRequests:                    map[string]DBReadRequest{},
		dbReadRequestByIdempotencyKey:     map[string]string{},
		dbReadValidationAttempts:          map[string][]DBReadValidationAttempt{},
		dbReadExecutionResults:            map[string][]DBReadExecutionResult{},
		externalToolPauses:                map[string]ExternalToolPause{},
		externalToolPauseByIdempotencyKey: map[string]string{},
		domainEvents:                      []transition.DomainEvent{},
		effectExecutions:                  map[string]transition.EffectExecution{},
		runnerExecutions:                  map[string]RunnerExecution{},
		commandReceipts:                   map[string]transition.CommandReceipt{},
		outcomes:                          map[string]outcome.Record{},
		knowledgeEntries:                  map[string]knowledge.Entry{},
		knowledgeEvidence:                 map[string][]knowledge.EvidenceLink{},
		knowledgeReviews:                  map[string][]knowledge.Review{},
		harnessProfiles:                   map[string]harness.Profile{},
		harnessOverlays:                   map[string]harness.Overlay{},
		harnessExperiments:                map[string]harness.Experiment{},
		harnessSessionBindings:            map[string]harness.SessionBinding{},
		harnessExecutions:                 []harness.Execution{},
		traces:                            map[string]events.Trace{},
		ratings:                           map[string][]review.HumanRating{},
		notes:                             map[string][]review.ImprovementNote{},
		evalRuns:                          map[string]evals.Run{},
		evalJudgments:                     map[string][]evals.Judgment{},
		candidates:                        map[string]improvement.Candidate{},
		runtimeDiagnoses:                  map[string]improvement.RuntimeDiagnosis{},
		proposals:                         map[string]review.Proposal{},
		changeAttempts:                    map[string]improvement.ChangeAttempt{},
		attemptWorkspaces:                 map[string]improvement.AttemptWorkspace{},
		validationRuns:                    map[string]improvement.ValidationRun{},
		repoChangeJobs:                    map[string]improvement.RepoChangeJob{},
		prAttempts:                        map[string]improvement.PRAttempt{},
		postMergeReplay:                   map[string]improvement.PostMergeReplay{},
		cronLeases:                        map[string]improvement.CronLease{},
		settings: improvement.Settings{
			ActiveProposalCap: defaultProposalSlotCap,
		},
	}
}
