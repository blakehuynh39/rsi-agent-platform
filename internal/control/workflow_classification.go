package control

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	workflowKindIncident       = "incident"
	workflowKindFeatureRequest = "feature-request"
	workflowKindArchitecture   = "architecture"
)

type workflowExecutionClassification struct {
	Workflow   storepkg.Workflow
	Rationale  string
	Confidence float64
	Source     string
}

type workflowClassificationOutput struct {
	WorkflowKind string  `json:"workflow_kind"`
	Rationale    string  `json:"rationale"`
	Confidence   float64 `json:"confidence"`
}

func classifyWorkflowExecution(
	cfg config.Config,
	store storepkg.Store,
	runnerClient *clients.RunnerClient,
	role string,
	trace events.Trace,
	workflow storepkg.Workflow,
	ingestion slackpkg.Ingestion,
) workflowExecutionClassification {
	fallback := workflowExecutionClassification{
		Workflow:   workflow,
		Rationale:  "Using persisted workflow classification.",
		Confidence: 0.84,
		Source:     "persisted_workflow",
	}
	if strings.TrimSpace(workflow.Kind) != workflowKindFeatureRequest || runnerClient == nil {
		return fallback
	}
	resp, err := runnerClient.Execute(buildWorkflowClassificationTask(cfg, store, role, trace, workflow, ingestion))
	if err != nil {
		log.Printf("control-plane workflow classification fallback trace=%s workflow=%s error=%v", trace.Summary.TraceID, workflow.ID, err)
		return fallback
	}
	if !resp.OK {
		log.Printf("control-plane workflow classification fallback trace=%s workflow=%s provider=%s message=%q", trace.Summary.TraceID, workflow.ID, resp.Provider, resp.Message)
		return fallback
	}
	output, err := parseWorkflowClassification(resp)
	if err != nil {
		log.Printf("control-plane workflow classification fallback trace=%s workflow=%s parse_error=%v", trace.Summary.TraceID, workflow.ID, err)
		return fallback
	}
	kind := normalizeWorkflowKind(output.WorkflowKind)
	if kind == "" {
		log.Printf("control-plane workflow classification fallback trace=%s workflow=%s invalid_kind=%q", trace.Summary.TraceID, workflow.ID, output.WorkflowKind)
		return fallback
	}
	effective := workflowWithKind(workflow, kind)
	return workflowExecutionClassification{
		Workflow:   effective,
		Rationale:  firstNonEmpty(strings.TrimSpace(output.Rationale), "Runner reclassified the workflow for execution."),
		Confidence: confidenceOrDefault(output.Confidence, 0.88),
		Source:     "ai_classifier",
	}
}

func buildWorkflowClassificationTask(
	cfg config.Config,
	store storepkg.Store,
	role string,
	trace events.Trace,
	workflow storepkg.Workflow,
	ingestion slackpkg.Ingestion,
) clients.RunnerTask {
	effectiveHarness := harness.ResolveEffectiveConfig(store, role, cfg.DefaultReasoningVerbosity)
	sessionScopeKind, sessionScopeID, parentScopeKind, parentScopeID := workflowSessionScope(trace, workflow)
	payload := map[string]any{
		"user_request":          firstNonEmpty(strings.TrimSpace(ingestion.Prompt.RenderedText), strings.TrimSpace(ingestion.Text)),
		"current_workflow_kind": workflow.Kind,
		"current_intent":        workflow.Intent,
		"channel_id":            ingestion.ChannelID,
		"thread_ts":             ingestion.ThreadTS,
		"entity_refs":           ingestion.EntityRefs,
	}
	prompt := strings.Join([]string{
		"Classify this Slack request for RSI execution.",
		"Choose exactly one workflow_kind: incident, feature-request, or architecture.",
		"Prefer architecture for questions, investigations, summaries, progress checks, trace/debug requests, verification, or alignment checks that are primarily read-heavy.",
		"Prefer feature-request only when the user is primarily asking RSI to propose, plan, build, or change behavior, product, or implementation.",
		"Prefer incident only for active breakage, outages, alerts, failures, or operational debugging.",
		"Return JSON only.",
		mustJSONString(payload),
	}, "\n\n")
	return clients.RunnerTask{
		TaskType:                  "general",
		Repo:                      firstNonEmpty(cfg.DefaultRepo, "rsi-agent-platform"),
		RepoRef:                   "main",
		Prompt:                    prompt,
		SystemMessage:             harness.ComposeSystemMessage("Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with workflow_kind, rationale, and confidence.", effectiveHarness),
		AllowedTools:              []string{},
		AllowedCommands:           []string{},
		TimeoutSeconds:            20,
		ExpectedOutputs:           []string{"workflow_kind", "rationale", "confidence"},
		ArtifactDestination:       fmt.Sprintf("trace:%s:workflow_classification", trace.Summary.TraceID),
		Intent:                    workflow.Intent,
		TraceID:                   trace.Summary.TraceID,
		WorkflowID:                trace.Summary.WorkflowID,
		ConversationID:            trace.Summary.ConversationID,
		CaseID:                    trace.Summary.CaseID,
		ChannelID:                 ingestion.ChannelID,
		ThreadTS:                  ingestion.ThreadTS,
		RecentConversationEntries: recentConversationEntries(store.ListConversationEntries(trace.Summary.ConversationID)),
		ResponseMode:              workflow.ResponseMode,
		ReasoningVerbosity:        effectiveHarness.ReasoningVerbosity,
		SessionScopeKind:          sessionScopeKind,
		SessionScopeID:            sessionScopeID,
		ParentSessionScopeKind:    parentScopeKind,
		ParentSessionScopeID:      parentScopeID,
		HarnessProfileID:          effectiveHarness.Profile.ID,
		HarnessOverlayVersion:     effectiveHarness.EffectiveOverlayVersion,
		MemoryBackend:             harness.DefaultMemoryBackend,
		AssistantPeerID:           fmt.Sprintf("rsi:%s:%s", cfg.Environment, role),
		UserPeerID:                workflowUserPeerID(store.ListConversationEntries(trace.Summary.ConversationID), sessionScopeKind, sessionScopeID),
		ContractVersion:           clients.RunnerExecutionContractVersion,
		ExecutionIntent: map[string]any{
			"kind":                "workflow_classification",
			"intent":              workflow.Intent,
			"runner_planner_mode": firstNonEmpty(cfg.RunnerPlannerMode, "runner_first"),
		},
		CapabilityLeases: questionCapabilityLeases(false, nil),
		DeliveryPolicy:   runnerDeliveryPolicy(ingestion.ChannelID, ingestion.ThreadTS, "none", trace.Summary.TraceID),
		WorkspacePolicy:  runnerutil.WorkspacePolicyFromConfig(cfg),
		ApprovalPolicy:   runnerApprovalPolicy(false),
	}
}

func parseWorkflowClassification(resp clients.RunnerResponse) (workflowClassificationOutput, error) {
	raw := mapValue(resp.Raw["structured_output"])
	if len(raw) == 0 {
		return workflowClassificationOutput{}, fmt.Errorf("runner response missing structured_output")
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return workflowClassificationOutput{}, fmt.Errorf("marshal workflow classification: %w", err)
	}
	var out workflowClassificationOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return workflowClassificationOutput{}, fmt.Errorf("parse workflow classification: %w", err)
	}
	return out, nil
}

func normalizeWorkflowKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "incident":
		return workflowKindIncident
	case "feature-request", "feature_request", "feature request":
		return workflowKindFeatureRequest
	case "architecture", "question", "general", "read_heavy_slack_qna":
		return workflowKindArchitecture
	default:
		return ""
	}
}

func workflowWithKind(workflow storepkg.Workflow, kind string) storepkg.Workflow {
	workflow.Kind = kind
	switch kind {
	case workflowKindIncident:
		workflow.Intent = "incident"
		workflow.AssignedBot = "oncall"
		workflow.ApprovalMode = "policy_gated"
		workflow.ResponseMode = "thread_updates"
	case workflowKindFeatureRequest:
		workflow.Intent = "feature_request"
		workflow.AssignedBot = "fr"
		workflow.ApprovalMode = "human_required"
		workflow.ResponseMode = "reply_in_thread"
	default:
		workflow.Intent = "question"
		workflow.AssignedBot = "arch"
		workflow.ApprovalMode = "policy_gated"
		workflow.ResponseMode = "reply_in_thread"
	}
	return workflow
}

func confidenceOrDefault(value float64, fallback float64) float64 {
	if value <= 0 {
		return fallback
	}
	if value > 1 {
		return 1
	}
	return value
}

func mustJSONString(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
