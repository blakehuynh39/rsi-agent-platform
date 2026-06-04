package control

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

type temporalClientFactory func(context.Context, config.TemporalTarget) (temporalOpsClient, error)

type temporalOpsClient interface {
	Close()
	ListSchedules(context.Context, string, int) ([]*temporalclient.ScheduleListEntry, error)
	DescribeSchedule(context.Context, string) (*temporalclient.ScheduleDescription, error)
	PauseSchedule(context.Context, string, string) error
	UnpauseSchedule(context.Context, string, string) error
	TriggerSchedule(context.Context, string) error
	ListWorkflows(context.Context, string, int) (*workflowservice.ListWorkflowExecutionsResponse, error)
	CountWorkflows(context.Context, string) (*workflowservice.CountWorkflowExecutionsResponse, error)
	DescribeWorkflow(context.Context, string, string) (*workflowservice.DescribeWorkflowExecutionResponse, error)
	CancelWorkflow(context.Context, string, string) error
	StartWorkflow(context.Context, temporalclient.StartWorkflowOptions, string, []any) (temporalWorkflowRun, error)
}

type temporalWorkflowRun interface {
	GetID() string
	GetRunID() string
}

var newTemporalClient temporalClientFactory = defaultTemporalClientFactory

type sdkTemporalClient struct {
	client temporalclient.Client
	target config.TemporalTarget
}

func defaultTemporalClientFactory(ctx context.Context, target config.TemporalTarget) (temporalOpsClient, error) {
	cert, err := temporalCertificate(target)
	if err != nil {
		return nil, err
	}
	client, err := temporalclient.DialContext(ctx, temporalclient.Options{
		HostPort:  target.HostPort,
		Namespace: target.Namespace,
		Identity:  "rsi-native-temporal-control",
		ConnectionOptions: temporalclient.ConnectionOptions{
			TLS: &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cert},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return sdkTemporalClient{client: client, target: target}, nil
}

func (c sdkTemporalClient) Close() {
	c.client.Close()
}

func (c sdkTemporalClient) ListSchedules(ctx context.Context, query string, pageSize int) ([]*temporalclient.ScheduleListEntry, error) {
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}
	iter, err := c.client.ScheduleClient().List(ctx, temporalclient.ScheduleListOptions{
		PageSize: pageSize,
		Query:    strings.TrimSpace(query),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*temporalclient.ScheduleListEntry, 0, pageSize)
	for iter.HasNext() && len(out) < pageSize {
		item, err := iter.Next()
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (c sdkTemporalClient) DescribeSchedule(ctx context.Context, scheduleID string) (*temporalclient.ScheduleDescription, error) {
	return c.client.ScheduleClient().GetHandle(ctx, scheduleID).Describe(ctx)
}

func (c sdkTemporalClient) PauseSchedule(ctx context.Context, scheduleID string, note string) error {
	return c.client.ScheduleClient().GetHandle(ctx, scheduleID).Pause(ctx, temporalclient.SchedulePauseOptions{Note: note})
}

func (c sdkTemporalClient) UnpauseSchedule(ctx context.Context, scheduleID string, note string) error {
	return c.client.ScheduleClient().GetHandle(ctx, scheduleID).Unpause(ctx, temporalclient.ScheduleUnpauseOptions{Note: note})
}

func (c sdkTemporalClient) TriggerSchedule(ctx context.Context, scheduleID string) error {
	return c.client.ScheduleClient().GetHandle(ctx, scheduleID).Trigger(ctx, temporalclient.ScheduleTriggerOptions{})
}

func (c sdkTemporalClient) ListWorkflows(ctx context.Context, query string, pageSize int) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}
	return c.client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Namespace: c.target.Namespace,
		PageSize:  int32(pageSize),
		Query:     strings.TrimSpace(query),
	})
}

func (c sdkTemporalClient) CountWorkflows(ctx context.Context, query string) (*workflowservice.CountWorkflowExecutionsResponse, error) {
	return c.client.CountWorkflow(ctx, &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: c.target.Namespace,
		Query:     strings.TrimSpace(query),
	})
}

func (c sdkTemporalClient) DescribeWorkflow(ctx context.Context, workflowID string, runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	return c.client.DescribeWorkflowExecution(ctx, workflowID, runID)
}

func (c sdkTemporalClient) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	return c.client.CancelWorkflow(ctx, workflowID, runID)
}

func (c sdkTemporalClient) StartWorkflow(ctx context.Context, options temporalclient.StartWorkflowOptions, workflowType string, args []any) (temporalWorkflowRun, error) {
	return c.client.ExecuteWorkflow(ctx, options, workflowType, args...)
}

func temporalCertificate(target config.TemporalTarget) (tls.Certificate, error) {
	if target.CertPEMEnv != "" || target.KeyPEMEnv != "" {
		certPEM := normalizePEM(os.Getenv(target.CertPEMEnv))
		keyPEM := normalizePEM(os.Getenv(target.KeyPEMEnv))
		if strings.TrimSpace(certPEM) == "" || strings.TrimSpace(keyPEM) == "" {
			return tls.Certificate{}, fmt.Errorf("target %s/%s requires populated cert/key env vars %s and %s", target.Environment, target.Name, target.CertPEMEnv, target.KeyPEMEnv)
		}
		return tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	}
	if target.CertPath == "" || target.KeyPath == "" {
		return tls.Certificate{}, errors.New("target cert/key are not configured")
	}
	return tls.LoadX509KeyPair(target.CertPath, target.KeyPath)
}

func normalizePEM(value string) string {
	return strings.ReplaceAll(value, `\n`, "\n")
}

func executeTemporalNativeToolAction(ctx context.Context, cfg config.Config, input nativeToolActionRequest) (any, string, string, string, map[string]any, int, error) {
	operation := normalizeTemporalOperation(input.Operation)
	environment := normalizeTemporalEnv(firstNonEmpty(stringArg(input.Arguments, "environment"), stringArg(input.Arguments, "env"), cfg.Environment))
	targetName := firstNonEmpty(stringArg(input.Arguments, "target"), stringArg(input.Arguments, "service"), stringArg(input.Arguments, "namespace"))
	target, ok := findTemporalTarget(cfg.TemporalTargets, environment, targetName)
	output := map[string]any{
		"operation":   operation,
		"environment": environment,
		"target":      strings.TrimSpace(targetName),
	}
	if ok {
		output["environment"] = target.Environment
		output["target"] = target.Name
		output["namespace"] = target.Namespace
		output["host_port"] = target.HostPort
	}
	mirrorEffect := map[string]any{"status": "not_applicable"}
	sourceRef := fmt.Sprintf("temporal:%s/%s", environment, strings.TrimSpace(targetName))
	if !cfg.TemporalControlEnabled {
		return output, "", sourceRef, "", mirrorEffect, http.StatusForbidden, errors.New("Temporal workflow control is not enabled")
	}
	if operation == "" {
		output["supported_operations"] = supportedTemporalOperations()
		return withTemporalError(output, "invalid_operation"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("missing or unsupported Temporal operation")
	}
	if !ok {
		output["targets"] = temporalTargetRefs(cfg.TemporalTargets)
		return withTemporalError(output, "target_not_configured"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("requested Temporal target is not configured")
	}
	sourceRef = fmt.Sprintf("temporal:%s/%s", target.Environment, target.Name)

	scheduleID := strings.TrimSpace(stringArg(input.Arguments, "schedule_id"))
	workflowID := strings.TrimSpace(stringArg(input.Arguments, "workflow_id"))
	newWorkflowID := strings.TrimSpace(stringArg(input.Arguments, "new_workflow_id"))
	runID := strings.TrimSpace(stringArg(input.Arguments, "run_id"))
	query := strings.TrimSpace(stringArg(input.Arguments, "query"))
	pageSize := intArg(input.Arguments, "limit", 0)
	reason := firstNonEmpty(input.Reason, stringArg(input.Arguments, "reason"), fmt.Sprintf("RSI Temporal %s", operation))
	dryRun := boolArg(input.Arguments, "dry_run", false)
	confirm := boolArg(input.Arguments, "confirm", false)
	workflowType := strings.TrimSpace(stringArg(input.Arguments, "workflow_type"))
	taskQueue := strings.TrimSpace(stringArg(input.Arguments, "task_queue"))
	workflowArgs, argsErr := temporalWorkflowArgsFromInput(input.Arguments)

	output["dry_run"] = dryRun
	if query != "" {
		output["query"] = query
	}
	if pageSize > 0 {
		output["limit"] = pageSize
	}
	if argsErr != nil {
		return withTemporalError(output, argsErr.Error()), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("workflow args must be a JSON array or array value")
	}
	if temporalScheduleOperation(operation) {
		if scheduleID == "" && len(target.AllowedScheduleIDs) == 1 {
			scheduleID = target.AllowedScheduleIDs[0]
		}
		output["schedule_id"] = scheduleID
		if operation != "list_schedules" {
			if scheduleID == "" {
				return withTemporalError(output, "missing_schedule_id"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("schedule_id is required for this Temporal schedule operation")
			}
			if !temporalAllowed(scheduleID, target.AllowedScheduleIDs, target.AllowedSchedulePrefixes) {
				return withTemporalError(output, "schedule_not_allowed"), "", sourceRef, "", mirrorEffect, http.StatusForbidden, errors.New("schedule_id is outside the configured Temporal allowlist")
			}
		}
	}
	if temporalWorkflowIDOperation(operation) {
		output["workflow_id"] = workflowID
		output["run_id"] = runID
		if workflowID == "" {
			return withTemporalError(output, "missing_workflow_id"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("workflow_id is required for this Temporal workflow operation")
		}
		if !temporalAllowed(workflowID, target.AllowedWorkflowIDs, target.AllowedWorkflowIDPrefixes) {
			return withTemporalError(output, "workflow_not_allowed"), "", sourceRef, "", mirrorEffect, http.StatusForbidden, errors.New("workflow_id is outside the configured Temporal allowlist")
		}
	}
	if temporalWorkflowStartOperation(operation) {
		if newWorkflowID == "" {
			newWorkflowID = workflowID
		}
		output["new_workflow_id"] = newWorkflowID
		output["workflow_type"] = workflowType
		output["task_queue"] = taskQueue
		if workflowType == "" {
			return withTemporalError(output, "missing_workflow_type"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("workflow_type is required for Temporal workflow start/restart")
		}
		if taskQueue == "" {
			return withTemporalError(output, "missing_task_queue"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("task_queue is required for Temporal workflow start/restart")
		}
		if newWorkflowID == "" {
			return withTemporalError(output, "missing_new_workflow_id"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("workflow_id or new_workflow_id is required for Temporal workflow start/restart")
		}
		if operation == "restart_workflow" && newWorkflowID == workflowID {
			return withTemporalError(output, "replacement_workflow_id_required"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("restart_workflow requires a distinct new_workflow_id so the existing workflow is not terminated or raced")
		}
		if !temporalAllowed(newWorkflowID, target.AllowedWorkflowIDs, target.AllowedWorkflowIDPrefixes) {
			return withTemporalError(output, "new_workflow_not_allowed"), "", sourceRef, "", mirrorEffect, http.StatusForbidden, errors.New("new_workflow_id is outside the configured Temporal allowlist")
		}
		if !containsTemporalString(target.AllowedWorkflowTypes, workflowType) {
			return withTemporalError(output, "workflow_type_not_allowed"), "", sourceRef, "", mirrorEffect, http.StatusForbidden, errors.New("workflow_type is outside the configured Temporal allowlist")
		}
		if !containsTemporalString(target.AllowedTaskQueues, taskQueue) {
			return withTemporalError(output, "task_queue_not_allowed"), "", sourceRef, "", mirrorEffect, http.StatusForbidden, errors.New("task_queue is outside the configured Temporal allowlist")
		}
		output["args_count"] = len(workflowArgs)
	}
	if temporalOperationMutates(operation) {
		output["reason"] = reason
		if dryRun {
			output["would_execute"] = true
			return output, fmt.Sprintf("Dry run accepted for Temporal %s on %s/%s.", operation, target.Environment, target.Name), sourceRef, "", mirrorEffect, http.StatusOK, nil
		}
		if !confirm {
			return withTemporalError(output, "confirmation_required"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("mutating Temporal operations require confirm=true after explicit operator authorization")
		}
	}

	callCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	client, err := newTemporalClient(callCtx, target)
	if err != nil {
		return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, http.StatusBadGateway, errors.New("failed to connect to Temporal target: " + err.Error())
	}
	defer client.Close()

	switch operation {
	case "list_schedules":
		items, err := client.ListSchedules(callCtx, query, pageSize)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		output["schedules"] = temporalScheduleListOutput(items)
		return output, fmt.Sprintf("listed %d Temporal schedule(s)", len(items)), sourceRef, "", mirrorEffect, http.StatusOK, nil
	case "describe_schedule":
		desc, err := client.DescribeSchedule(callCtx, scheduleID)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		output["schedule"] = temporalScheduleDescriptionOutput(desc)
		return output, "described Temporal schedule " + scheduleID, sourceRef + ":schedule:" + scheduleID, "", mirrorEffect, http.StatusOK, nil
	case "pause_schedule":
		before, err := client.DescribeSchedule(callCtx, scheduleID)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		if err := client.PauseSchedule(callCtx, scheduleID, reason); err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		after, _ := client.DescribeSchedule(callCtx, scheduleID)
		output["before"] = temporalScheduleDescriptionOutput(before)
		output["after"] = temporalScheduleDescriptionOutput(after)
		return output, "paused Temporal schedule " + scheduleID, sourceRef + ":schedule:" + scheduleID, "", mirrorEffect, http.StatusOK, nil
	case "unpause_schedule":
		before, err := client.DescribeSchedule(callCtx, scheduleID)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		if err := client.UnpauseSchedule(callCtx, scheduleID, reason); err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		after, _ := client.DescribeSchedule(callCtx, scheduleID)
		output["before"] = temporalScheduleDescriptionOutput(before)
		output["after"] = temporalScheduleDescriptionOutput(after)
		return output, "unpaused Temporal schedule " + scheduleID, sourceRef + ":schedule:" + scheduleID, "", mirrorEffect, http.StatusOK, nil
	case "trigger_schedule":
		before, err := client.DescribeSchedule(callCtx, scheduleID)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		if err := client.TriggerSchedule(callCtx, scheduleID); err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		after, _ := client.DescribeSchedule(callCtx, scheduleID)
		output["before"] = temporalScheduleDescriptionOutput(before)
		output["after"] = temporalScheduleDescriptionOutput(after)
		return output, "started one Temporal schedule action for " + scheduleID, sourceRef + ":schedule:" + scheduleID, "", mirrorEffect, http.StatusOK, nil
	case "list_workflows":
		resp, err := client.ListWorkflows(callCtx, query, pageSize)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		output["workflows"] = temporalWorkflowExecutionsOutput(resp.GetExecutions())
		output["has_next_page"] = len(resp.GetNextPageToken()) > 0
		return output, fmt.Sprintf("listed %d Temporal workflow(s)", len(resp.GetExecutions())), sourceRef, "", mirrorEffect, http.StatusOK, nil
	case "count_workflows":
		resp, err := client.CountWorkflows(callCtx, query)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		output["count"] = resp.GetCount()
		return output, "counted Temporal workflows", sourceRef, "", mirrorEffect, http.StatusOK, nil
	case "describe_workflow":
		desc, err := client.DescribeWorkflow(callCtx, workflowID, runID)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		output["workflow"] = temporalWorkflowDescriptionOutput(desc)
		return output, "described Temporal workflow " + workflowID, sourceRef + ":workflow:" + workflowID, "", mirrorEffect, http.StatusOK, nil
	case "stop_workflow":
		before, err := client.DescribeWorkflow(callCtx, workflowID, runID)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		if err := client.CancelWorkflow(callCtx, workflowID, runID); err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		after, _ := client.DescribeWorkflow(callCtx, workflowID, runID)
		output["before"] = temporalWorkflowDescriptionOutput(before)
		output["after"] = temporalWorkflowDescriptionOutput(after)
		return output, "requested graceful cancellation for Temporal workflow " + workflowID, sourceRef + ":workflow:" + workflowID, "", mirrorEffect, http.StatusOK, nil
	case "start_workflow":
		run, err := client.StartWorkflow(callCtx, temporalStartOptions(newWorkflowID, taskQueue), workflowType, workflowArgs)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		output["started_workflow_id"] = run.GetID()
		output["started_run_id"] = run.GetRunID()
		return output, "started Temporal workflow " + run.GetID(), sourceRef + ":workflow:" + run.GetID(), "", mirrorEffect, http.StatusOK, nil
	case "restart_workflow":
		before, err := client.DescribeWorkflow(callCtx, workflowID, runID)
		if err != nil {
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		run, err := client.StartWorkflow(callCtx, temporalStartOptions(newWorkflowID, taskQueue), workflowType, workflowArgs)
		if err != nil {
			output["before"] = temporalWorkflowDescriptionOutput(before)
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), err
		}
		if err := client.CancelWorkflow(callCtx, workflowID, runID); err != nil {
			output["before"] = temporalWorkflowDescriptionOutput(before)
			output["started_workflow_id"] = run.GetID()
			output["started_run_id"] = run.GetRunID()
			return withTemporalError(output, err.Error()), "", sourceRef, "", mirrorEffect, statusFromErr(err), errors.New("started replacement workflow " + run.GetID() + " but cancellation request failed: " + err.Error())
		}
		output["before"] = temporalWorkflowDescriptionOutput(before)
		output["started_workflow_id"] = run.GetID()
		output["started_run_id"] = run.GetRunID()
		return output, "started replacement Temporal workflow " + run.GetID() + " and requested graceful cancellation of " + workflowID, sourceRef + ":workflow:" + workflowID, "", mirrorEffect, http.StatusOK, nil
	default:
		return withTemporalError(output, "invalid_operation"), "", sourceRef, "", mirrorEffect, http.StatusBadRequest, errors.New("unsupported Temporal operation")
	}
}

func normalizeTemporalOperation(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "list_schedules", "schedules":
		return "list_schedules"
	case "describe_schedule", "schedule_status", "status_schedule":
		return "describe_schedule"
	case "pause", "pause_schedule":
		return "pause_schedule"
	case "resume", "unpause", "unpause_schedule", "resume_schedule":
		return "unpause_schedule"
	case "trigger", "trigger_schedule", "start_schedule", "start_now":
		return "trigger_schedule"
	case "list", "list_workflows", "visibility", "visibility_list":
		return "list_workflows"
	case "count", "count_workflows":
		return "count_workflows"
	case "describe", "describe_workflow", "workflow_status", "status_workflow":
		return "describe_workflow"
	case "stop", "cancel", "cancel_workflow", "stop_workflow":
		return "stop_workflow"
	case "start", "start_workflow":
		return "start_workflow"
	case "restart", "restart_workflow":
		return "restart_workflow"
	default:
		return ""
	}
}

func supportedTemporalOperations() []string {
	return []string{
		"list_schedules",
		"describe_schedule",
		"pause_schedule",
		"unpause_schedule",
		"trigger_schedule",
		"list_workflows",
		"count_workflows",
		"describe_workflow",
		"start_workflow",
		"stop_workflow",
		"restart_workflow",
	}
}

func temporalOperationMutates(operation string) bool {
	switch operation {
	case "pause_schedule", "unpause_schedule", "trigger_schedule", "start_workflow", "stop_workflow", "restart_workflow":
		return true
	default:
		return false
	}
}

func temporalScheduleOperation(operation string) bool {
	switch operation {
	case "list_schedules", "describe_schedule", "pause_schedule", "unpause_schedule", "trigger_schedule":
		return true
	default:
		return false
	}
}

func temporalWorkflowIDOperation(operation string) bool {
	switch operation {
	case "describe_workflow", "stop_workflow", "restart_workflow":
		return true
	default:
		return false
	}
}

func temporalWorkflowStartOperation(operation string) bool {
	return operation == "start_workflow" || operation == "restart_workflow"
}

func temporalStartOptions(workflowID string, taskQueue string) temporalclient.StartWorkflowOptions {
	return temporalclient.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                taskQueue,
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_FAIL,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
}

func normalizeTemporalEnv(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "stg", "stage", "staging":
		return "stage"
	case "prod", "production":
		return "prod"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func findTemporalTarget(targets []config.TemporalTarget, environment string, name string) (config.TemporalTarget, bool) {
	environment = normalizeTemporalEnv(environment)
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return config.TemporalTarget{}, false
	}
	for _, target := range targets {
		if normalizeTemporalEnv(target.Environment) != environment {
			continue
		}
		if name == strings.ToLower(target.Name) || name == strings.ToLower(target.Namespace) {
			return target, true
		}
	}
	return config.TemporalTarget{}, false
}

func temporalTargetRefs(targets []config.TemporalTarget) []map[string]any {
	out := make([]map[string]any, 0, len(targets))
	for _, target := range targets {
		out = append(out, map[string]any{
			"environment": target.Environment,
			"target":      target.Name,
			"namespace":   target.Namespace,
			"host_port":   target.HostPort,
		})
	}
	return out
}

func temporalAllowed(value string, exact []string, prefixes []string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if len(exact) == 0 && len(prefixes) == 0 {
		return false
	}
	for _, item := range exact {
		if strings.TrimSpace(item) == value {
			return true
		}
	}
	for _, prefix := range prefixes {
		if prefix = strings.TrimSpace(prefix); prefix != "" && strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func temporalWorkflowArgsFromInput(args map[string]any) ([]any, error) {
	value, ok := args["args"]
	if !ok || value == nil {
		return nil, nil
	}
	switch typed := value.(type) {
	case []any:
		return typed, nil
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return nil, nil
		}
		var out []any
		if err := json.Unmarshal([]byte(text), &out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported args type %T", value)
	}
}

func withTemporalError(output map[string]any, err string) map[string]any {
	out := make(map[string]any, len(output)+1)
	for key, value := range output {
		out[key] = value
	}
	out["error"] = err
	return out
}

func temporalScheduleListOutput(items []*temporalclient.ScheduleListEntry) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, map[string]any{
			"id":                item.ID,
			"paused":            item.Paused,
			"note":              item.Note,
			"workflow_type":     item.WorkflowType.Name,
			"next_action_times": temporalTimeList(item.NextActionTimes),
			"recent_actions":    temporalScheduleActionResultsOutput(item.RecentActions),
		})
	}
	return out
}

func temporalScheduleDescriptionOutput(desc *temporalclient.ScheduleDescription) map[string]any {
	if desc == nil {
		return map[string]any{}
	}
	state := desc.Schedule.State
	output := map[string]any{
		"num_actions":                       desc.Info.NumActions,
		"num_actions_missed_catchup_window": desc.Info.NumActionsMissedCatchupWindow,
		"num_actions_skipped_overlap":       desc.Info.NumActionsSkippedOverlap,
		"created_at":                        temporalTimeString(desc.Info.CreatedAt),
		"last_update_at":                    temporalTimeString(desc.Info.LastUpdateAt),
		"next_action_times":                 temporalTimeList(desc.Info.NextActionTimes),
		"running_workflows":                 temporalScheduleWorkflowExecutionsOutput(desc.Info.RunningWorkflows),
		"recent_actions":                    temporalScheduleActionResultsOutput(desc.Info.RecentActions),
	}
	if state != nil {
		output["paused"] = state.Paused
		output["note"] = state.Note
		output["limited_actions"] = state.LimitedActions
		output["remaining_actions"] = state.RemainingActions
	}
	if desc.Schedule.Policy != nil {
		output["overlap_policy"] = desc.Schedule.Policy.Overlap.String()
		output["catchup_window_seconds"] = int(desc.Schedule.Policy.CatchupWindow.Seconds())
		output["pause_on_failure"] = desc.Schedule.Policy.PauseOnFailure
	}
	if action, ok := desc.Schedule.Action.(*temporalclient.ScheduleWorkflowAction); ok {
		output["workflow_id_template"] = action.ID
		output["task_queue"] = action.TaskQueue
		output["workflow"] = fmt.Sprint(action.Workflow)
	}
	return output
}

func temporalScheduleWorkflowExecutionsOutput(values []temporalclient.ScheduleWorkflowExecution) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, item := range values {
		out = append(out, map[string]any{
			"workflow_id":            item.WorkflowID,
			"first_execution_run_id": item.FirstExecutionRunID,
		})
	}
	return out
}

func temporalScheduleActionResultsOutput(values []temporalclient.ScheduleActionResult) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, item := range values {
		entry := map[string]any{
			"schedule_time": temporalTimeString(item.ScheduleTime),
			"actual_time":   temporalTimeString(item.ActualTime),
		}
		if item.StartWorkflowResult != nil {
			entry["workflow_id"] = item.StartWorkflowResult.WorkflowID
			entry["first_execution_run_id"] = item.StartWorkflowResult.FirstExecutionRunID
		}
		out = append(out, entry)
	}
	return out
}

func temporalWorkflowExecutionsOutput(values []*workflowpb.WorkflowExecutionInfo) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, item := range values {
		if item == nil {
			continue
		}
		out = append(out, temporalWorkflowExecutionInfoOutput(item))
	}
	return out
}

func temporalWorkflowExecutionInfoOutput(info *workflowpb.WorkflowExecutionInfo) map[string]any {
	if info == nil {
		return map[string]any{}
	}
	return map[string]any{
		"workflow_id":            temporalWorkflowExecutionID(info.GetExecution()),
		"run_id":                 temporalWorkflowRunID(info.GetExecution()),
		"workflow_type":          temporalWorkflowTypeName(info.GetType()),
		"task_queue":             info.GetTaskQueue(),
		"status":                 info.GetStatus().String(),
		"start_time":             temporalPBTimeString(info.GetStartTime()),
		"execution_time":         temporalPBTimeString(info.GetExecutionTime()),
		"close_time":             temporalPBTimeString(info.GetCloseTime()),
		"history_length":         info.GetHistoryLength(),
		"history_size_bytes":     info.GetHistorySizeBytes(),
		"state_transition_count": info.GetStateTransitionCount(),
		"first_run_id":           info.GetFirstRunId(),
		"parent_workflow_id":     temporalWorkflowExecutionID(info.GetParentExecution()),
		"parent_run_id":          temporalWorkflowRunID(info.GetParentExecution()),
		"root_workflow_id":       temporalWorkflowExecutionID(info.GetRootExecution()),
		"root_run_id":            temporalWorkflowRunID(info.GetRootExecution()),
		"worker_deployment_name": info.GetWorkerDeploymentName(),
		"assigned_build_id":      info.GetAssignedBuildId(),
		"inherited_build_id":     info.GetInheritedBuildId(),
	}
}

func temporalWorkflowDescriptionOutput(desc *workflowservice.DescribeWorkflowExecutionResponse) map[string]any {
	if desc == nil {
		return map[string]any{}
	}
	info := desc.GetWorkflowExecutionInfo()
	output := map[string]any{
		"pending_activities": len(desc.GetPendingActivities()),
		"pending_children":   len(desc.GetPendingChildren()),
	}
	for key, value := range temporalWorkflowExecutionInfoOutput(info) {
		output[key] = value
	}
	return output
}

func temporalWorkflowExecutionID(execution *commonpb.WorkflowExecution) string {
	if execution == nil {
		return ""
	}
	return execution.GetWorkflowId()
}

func temporalWorkflowRunID(execution *commonpb.WorkflowExecution) string {
	if execution == nil {
		return ""
	}
	return execution.GetRunId()
}

func temporalWorkflowTypeName(workflowType *commonpb.WorkflowType) string {
	if workflowType == nil {
		return ""
	}
	return workflowType.GetName()
}

func temporalPBTimeString(value *timestamppb.Timestamp) string {
	if value == nil {
		return ""
	}
	return temporalTimeString(value.AsTime())
}

func temporalTimeString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func temporalTimeList(values []time.Time) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if text := temporalTimeString(value); text != "" {
			out = append(out, text)
		}
	}
	return out
}

func containsTemporalString(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}
