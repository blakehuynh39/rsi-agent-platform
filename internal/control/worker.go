package control

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type runnerReasoningStep struct {
	StepType     string   `json:"step_type"`
	Summary      string   `json:"summary"`
	Alternatives []string `json:"alternatives"`
	Confidence   float64  `json:"confidence"`
	Decision     string   `json:"decision"`
}

type runnerStructuredOutput struct {
	ContextSummary   string                `json:"context_summary"`
	ReplyDraft       string                `json:"reply_draft"`
	FinalAnswer      string                `json:"final_answer"`
	Confidence       float64               `json:"confidence"`
	SelfCritique     string                `json:"self_critique"`
	VisibleReasoning []runnerReasoningStep `json:"visible_reasoning"`
}

func RunWorker(cfg config.Config, store storepkg.Store) error {
	runnerClient := clients.NewRunnerClient(cfg.RunnerBaseURL)
	toolClient := clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL)
	workerID := fmt.Sprintf("%s-worker", cfg.ServiceName)
	for {
		item, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.WorkflowQueue, queue.ProactiveQueue}, workerID, cfg.WorkItemLeaseDuration)
		if err != nil {
			return err
		}
		if !ok {
			time.Sleep(cfg.WorkerPollInterval)
			continue
		}
		if err := processWorkflowItem(cfg, store, runnerClient, toolClient, item); err != nil {
			log.Printf("control-plane worker item=%s error=%v", item.ID, err)
			_, _ = store.FailWorkItem(item.ID, err.Error())
			_, _ = store.ApplyTraceUpdate(item.TraceID, storepkg.TraceUpdate{
				Status: ptrStatus(events.StatusFailed),
				Events: []events.TraceEvent{
					{
						TraceID:     item.TraceID,
						IngestionID: item.IngestionID,
						WorkflowID:  item.WorkflowID,
						Plane:       "control",
						Service:     cfg.ServiceName,
						Actor:       "worker",
						EventType:   "workflow.failed",
						Status:      events.StatusFailed,
						StartedAt:   time.Now().UTC(),
						Description: err.Error(),
					},
				},
				WorkflowStatus: "failed",
				WorkflowError:  err.Error(),
			})
			continue
		}
		_, _ = store.CompleteWorkItem(item.ID)
	}
}

func processWorkflowItem(cfg config.Config, store storepkg.Store, runnerClient *clients.RunnerClient, toolClient *clients.ToolGatewayClient, item queue.WorkItem) error {
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), item.WorkflowID)
	if !ok {
		return fmt.Errorf("workflow %s not found", item.WorkflowID)
	}
	ingestion, ok := findIngestion(store.ListIngestions(), trace.Summary.IngestionID)
	if !ok {
		return fmt.Errorf("ingestion %s not found", trace.Summary.IngestionID)
	}
	now := time.Now().UTC()
	_, _ = store.UpdateWorkflowStatus(workflow.ID, "running", "")
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Status: ptrStatus(events.StatusRunning),
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "workflow.started",
				Status:      events.StatusRunning,
				StartedAt:   now,
				Description: fmt.Sprintf("Started %s workflow worker loop.", workflow.Intent),
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:         fmt.Sprintf("reason-start-%d", now.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "pre_action_summary",
				Summary:    fmt.Sprintf("Preparing %s response for thread %s.", workflow.Intent, trace.Summary.ThreadKey),
				Confidence: 0.9,
				Decision:   fmt.Sprintf("response_mode:%s", workflow.ResponseMode),
				CreatedAt:  now,
			},
		},
		WorkflowStatus: "running",
	})

	toolNames := toolPlanForIntent(workflow.Intent)
	toolCalls, toolEvents, contextRefs, contextSummary, err := collectContext(cfg, toolClient, trace, workflow, ingestion, toolNames)
	if err != nil {
		return err
	}
	if len(toolCalls) > 0 || len(toolEvents) > 0 {
		_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
			Events:    toolEvents,
			ToolCalls: toolCalls,
			Reasoning: []events.ReasoningStep{
				{
					ID:           fmt.Sprintf("reason-context-%d", time.Now().UTC().UnixNano()),
					TraceID:      trace.Summary.TraceID,
					WorkflowID:   trace.Summary.WorkflowID,
					StepType:     "context_collected",
					Summary:      contextSummary,
					EvidenceRefs: evidenceRefsFromContext(contextRefs),
					Confidence:   0.82,
					Decision:     strings.Join(toolNames, ","),
					CreatedAt:    time.Now().UTC(),
				},
			},
		})
	}

	runnerStarted := time.Now().UTC()
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       workflow.AssignedBot,
				EventType:   "runner.started",
				Status:      events.StatusRunning,
				StartedAt:   runnerStarted,
				Description: "Runner task dispatched with verbose reasoning enabled.",
			},
		},
	})

	runnerTask := buildRunnerTask(cfg, trace, workflow, ingestion, contextSummary, contextRefs, toolNames)
	runnerResp, err := runnerClient.Execute(runnerTask)
	if err != nil {
		return err
	}
	runnerOutput := parseRunnerOutput(runnerResp)
	runnerCompleted := time.Now().UTC()

	allowed, policyVerdict := replyPolicy(store, workflow.Kind, trace.Summary.ThreadKey, ingestion.ChannelID)
	replyBody := firstNonEmpty(strings.TrimSpace(runnerOutput.FinalAnswer), strings.TrimSpace(runnerOutput.ReplyDraft), strings.TrimSpace(runnerResp.Message))
	slackAction := events.SlackActionRecord{
		ID:             fmt.Sprintf("slack-action-%d", runnerCompleted.UnixNano()),
		TraceID:        trace.Summary.TraceID,
		WorkflowID:     trace.Summary.WorkflowID,
		ChannelID:      ingestion.ChannelID,
		ThreadTS:       ingestion.ThreadTS,
		IdempotencyKey: fmt.Sprintf("%s:%s:%s", ingestion.ChannelID, ingestion.ThreadTS, trace.Summary.TraceID),
		DraftBody:      firstNonEmpty(runnerOutput.ReplyDraft, replyBody),
		FinalBody:      replyBody,
		PolicyVerdict:  policyVerdict,
		SendStatus:     "blocked_by_policy",
		CreatedAt:      runnerCompleted,
	}
	slackEvents := []events.TraceEvent{
		{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     cfg.ServiceName,
			Actor:       workflow.AssignedBot,
			EventType:   "slack.reply.drafted",
			Status:      events.StatusCompleted,
			StartedAt:   runnerCompleted,
			EndedAt:     ptrTime(runnerCompleted),
			Description: "Drafted in-thread Slack response.",
		},
	}
	if allowed && replyBody != "" {
		result, toolErr := toolClient.Execute("slack.reply", map[string]any{
			"channel_id": ingestion.ChannelID,
			"thread_ts":  ingestion.ThreadTS,
			"body":       replyBody,
		})
		if toolErr != nil {
			return toolErr
		}
		slackAction.SendStatus = statusString(result.Output["posted"], "draft_only")
		if slackAction.SendStatus == "" {
			slackAction.SendStatus = "posted"
		}
		slackEvents = append(slackEvents, events.TraceEvent{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "edge",
			Service:     "tool-gateway",
			Actor:       workflow.AssignedBot,
			EventType:   "slack.reply.posted",
			Status:      events.StatusCompleted,
			StartedAt:   time.Now().UTC(),
			EndedAt:     ptrTime(time.Now().UTC()),
			Description: result.Summary,
		})
	}

	finalReasoning := reasoningStepsFromRunner(trace.Summary.TraceID, trace.Summary.WorkflowID, runnerOutput, runnerCompleted)
	if runnerOutput.SelfCritique != "" {
		finalReasoning = append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-self-critique-%d", runnerCompleted.UnixNano()),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "self_critique",
			Summary:    runnerOutput.SelfCritique,
			Confidence: runnerOutput.Confidence,
			CreatedAt:  runnerCompleted,
		})
	}
	completedStatus := events.StatusCompleted
	workflowStatus := "completed"
	if !allowed {
		completedStatus = events.StatusNeedsHuman
		workflowStatus = "needs-human"
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Status: ptrStatus(completedStatus),
		Events: append([]events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       workflow.AssignedBot,
				EventType:   "runner.completed",
				Status:      completedStatus,
				StartedAt:   runnerStarted,
				EndedAt:     ptrTime(runnerCompleted),
				Description: "Runner returned visible reasoning and reply body.",
			},
		}, slackEvents...),
		Reasoning: append(finalReasoning, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-final-%d", runnerCompleted.UnixNano()),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "final_answer_rationale",
			Summary:    firstNonEmpty(runnerOutput.ContextSummary, "Prepared final response from collected context."),
			Confidence: runnerOutput.Confidence,
			Decision:   replyBody,
			CreatedAt:  runnerCompleted,
		}),
		SlackActions:   []events.SlackActionRecord{slackAction},
		WorkflowStatus: workflowStatus,
	})

	_, err = store.EnqueueWorkItem(queue.WorkItem{
		Queue:        queue.EvalQueue,
		Kind:         "evaluate_trace",
		Status:       queue.WorkQueued,
		TraceID:      trace.Summary.TraceID,
		WorkflowID:   trace.Summary.WorkflowID,
		IngestionID:  trace.Summary.IngestionID,
		ThreadKey:    trace.Summary.ThreadKey,
		RequestedBy:  cfg.ServiceName,
		ApprovalMode: workflow.ApprovalMode,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     "improvement-plane",
				Actor:       "eval-scheduler",
				EventType:   "eval_trace.queued",
				Status:      events.StatusQueued,
				StartedAt:   time.Now().UTC(),
				Description: "Queued trace for post-completion evaluation.",
			},
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "control",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "workflow.completed",
				Status:      completedStatus,
				StartedAt:   time.Now().UTC(),
				Description: fmt.Sprintf("Workflow ended with %s.", workflowStatus),
			},
		},
	})
	return nil
}

func buildRunnerTask(cfg config.Config, trace events.Trace, workflow storepkg.Workflow, ingestion slackpkg.Ingestion, contextSummary string, contextRefs []map[string]any, toolNames []string) clients.RunnerTask {
	prompt := fmt.Sprintf("User request: %s\n\nRespond in-thread for intent=%s. Use the gathered context and cite concrete evidence when possible. If unsure, say so explicitly.", ingestion.Text, workflow.Intent)
	return clients.RunnerTask{
		TaskType:                "workflow",
		Repo:                    cfg.DefaultRepo,
		RepoRef:                 "main",
		Prompt:                  prompt,
		SystemMessage:           "Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, and self_critique.",
		AllowedTools:            toolNames,
		AllowedCommands:         []string{},
		TimeoutSeconds:          120,
		ExpectedOutputs:         []string{"visible_reasoning", "final_answer"},
		ArtifactDestination:     fmt.Sprintf("trace:%s", trace.Summary.TraceID),
		ContextSummary:          contextSummary,
		Intent:                  workflow.Intent,
		TraceID:                 trace.Summary.TraceID,
		WorkflowID:              trace.Summary.WorkflowID,
		RepoAllowlist:           cfg.AllowedTargetRepos,
		ToolAllowlist:           toolNames,
		ResponseMode:            workflow.ResponseMode,
		ContextRefs:             contextRefs,
		ApprovalMode:            workflow.ApprovalMode,
		ReasoningVerbosity:      cfg.DefaultReasoningVerbosity,
		RejectedProposalContext: []map[string]any{},
	}
}

func collectContext(cfg config.Config, toolClient *clients.ToolGatewayClient, trace events.Trace, workflow storepkg.Workflow, ingestion slackpkg.Ingestion, toolNames []string) ([]events.ToolCallRecord, []events.TraceEvent, []map[string]any, string, error) {
	toolCalls := make([]events.ToolCallRecord, 0, len(toolNames))
	toolEvents := make([]events.TraceEvent, 0, len(toolNames)*2)
	contextRefs := make([]map[string]any, 0, len(toolNames))
	summaries := make([]string, 0, len(toolNames))
	for _, toolName := range toolNames {
		started := time.Now().UTC()
		input := map[string]any{
			"repo":               cfg.DefaultRepo,
			"question":           ingestion.Text,
			"service":            workflow.AssignedBot,
			"alert":              ingestion.Text,
			"namespace":          cfg.SandboxNamespace,
			"target":             workflow.Kind,
			"knowledge_base_url": cfg.DefaultKnowledgeBaseURL,
			"channel_id":         ingestion.ChannelID,
			"thread_ts":          ingestion.ThreadTS,
		}
		toolEvents = append(toolEvents, events.TraceEvent{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     "tool-gateway",
			Actor:       workflow.AssignedBot,
			EventType:   "tool.requested",
			Status:      events.StatusQueued,
			StartedAt:   started,
			Description: fmt.Sprintf("Requested %s.", toolName),
		})
		result, err := toolClient.Execute(toolName, input)
		if err != nil {
			return nil, nil, nil, "", err
		}
		completed := time.Now().UTC()
		toolCalls = append(toolCalls, events.ToolCallRecord{
			ID:                    fmt.Sprintf("tool-record-%d", completed.UnixNano()),
			TraceID:               trace.Summary.TraceID,
			WorkflowID:            trace.Summary.WorkflowID,
			ToolName:              toolName,
			ToolCallID:            result.ToolCallID,
			Request:               result.Input,
			Summary:               result.Summary,
			RawArtifactRefs:       result.RawArtifactRefs,
			ApprovalState:         result.ApprovalState,
			InterpretationSummary: result.Summary,
			Status:                "completed",
			CreatedAt:             completed,
		})
		toolEvents = append(toolEvents, events.TraceEvent{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     "tool-gateway",
			Actor:       workflow.AssignedBot,
			EventType:   "tool.completed",
			Status:      events.StatusCompleted,
			StartedAt:   started,
			EndedAt:     ptrTime(completed),
			Description: result.Summary,
		})
		contextRefs = append(contextRefs, map[string]any{
			"kind":         "tool_call",
			"ref":          result.ToolCallID,
			"tool_call_id": result.ToolCallID,
			"summary":      result.Summary,
			"tool_name":    toolName,
		})
		summaries = append(summaries, result.Summary)
	}
	return toolCalls, toolEvents, contextRefs, strings.Join(summaries, " "), nil
}

func findWorkflow(items []storepkg.Workflow, workflowID string) (storepkg.Workflow, bool) {
	for _, item := range items {
		if item.ID == workflowID {
			return item, true
		}
	}
	return storepkg.Workflow{}, false
}

func findIngestion(items []slackpkg.Ingestion, ingestionID string) (slackpkg.Ingestion, bool) {
	for _, item := range items {
		if item.ID == ingestionID {
			return item, true
		}
	}
	return slackpkg.Ingestion{}, false
}

func toolPlanForIntent(intent string) []string {
	switch intent {
	case "incident":
		return []string{"sentry.lookup", "kubernetes.inspect"}
	case "feature_request":
		return []string{"repo.context", "github.repo_context"}
	default:
		return []string{"repo.context", "knowledge.context"}
	}
}

func replyPolicy(store storepkg.Store, workflowKind string, threadKey string, channelID string) (bool, string) {
	for _, item := range store.ListThreadPolicies() {
		if item.ThreadKey == threadKey && item.Muted {
			return false, "thread_muted"
		}
	}
	for _, item := range store.ListChannelPolicies() {
		if item.ChannelID != channelID {
			continue
		}
		if !item.AutoPostAllowed {
			return false, "channel_autopost_disabled"
		}
		for _, allowed := range item.AllowedWorkflowKinds {
			if allowed == workflowKind {
				return true, "allowed"
			}
		}
		return false, "workflow_kind_not_allowed"
	}
	return false, "channel_policy_missing"
}

func parseRunnerOutput(resp clients.RunnerResponse) runnerStructuredOutput {
	if raw, ok := resp.Raw["structured_output"]; ok {
		data, _ := json.Marshal(raw)
		var out runnerStructuredOutput
		if err := json.Unmarshal(data, &out); err == nil {
			return out
		}
	}
	var out runnerStructuredOutput
	if err := json.Unmarshal([]byte(resp.Message), &out); err == nil {
		return out
	}
	return runnerStructuredOutput{
		FinalAnswer: resp.Message,
		Confidence:  0.5,
		VisibleReasoning: []runnerReasoningStep{
			{
				StepType:   "fallback",
				Summary:    "Runner returned unstructured output; stored raw response as the visible answer.",
				Confidence: 0.5,
				Decision:   resp.Message,
			},
		},
	}
}

func reasoningStepsFromRunner(traceID string, workflowID string, output runnerStructuredOutput, createdAt time.Time) []events.ReasoningStep {
	out := make([]events.ReasoningStep, 0, len(output.VisibleReasoning))
	for index, step := range output.VisibleReasoning {
		out = append(out, events.ReasoningStep{
			ID:           fmt.Sprintf("reason-runner-%d-%d", createdAt.UnixNano(), index),
			TraceID:      traceID,
			WorkflowID:   workflowID,
			StepType:     firstNonEmpty(step.StepType, "visible_reasoning"),
			Summary:      step.Summary,
			Alternatives: step.Alternatives,
			Confidence:   step.Confidence,
			Decision:     step.Decision,
			CreatedAt:    createdAt,
		})
	}
	return out
}

func evidenceRefsFromContext(contextRefs []map[string]any) []events.EvidenceRef {
	out := make([]events.EvidenceRef, 0, len(contextRefs))
	for _, ref := range contextRefs {
		out = append(out, events.EvidenceRef{
			Kind:    stringFromMap(ref, "kind"),
			Ref:     firstNonEmpty(stringFromMap(ref, "ref"), stringFromMap(ref, "tool_call_id")),
			Summary: stringFromMap(ref, "summary"),
		})
	}
	return out
}

func stringFromMap(item map[string]any, key string) string {
	if item == nil {
		return ""
	}
	value, _ := item[key].(string)
	return value
}

func ptrStatus(status events.Status) *events.Status {
	return &status
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func statusString(value any, fallback string) string {
	if posted, ok := value.(bool); ok {
		if posted {
			return "posted"
		}
		return fallback
	}
	return fallback
}
