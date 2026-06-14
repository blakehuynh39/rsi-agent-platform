package control

import (
	"context"
	"net/http"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestNativeTemporalDryRunDoesNotConnect(t *testing.T) {
	cfg := nativeTemporalTestConfig()
	called := false
	previous := newTemporalClient
	newTemporalClient = func(context.Context, config.TemporalTarget) (temporalOpsClient, error) {
		called = true
		return &fakeTemporalOpsClient{}, nil
	}
	t.Cleanup(func() { newTemporalClient = previous })

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "temporal"), nativeToolActionRequest{
		Surface:        "temporal",
		Operation:      "pause_schedule",
		IdempotencyKey: "temporal-dry-run",
		Reason:         "validate Temporal pause",
		Arguments: map[string]any{
			"environment": "stage",
			"target":      "royalty-graph-v2",
			"schedule_id": "royalty-graph-schedule-v2",
			"dry_run":     true,
		},
	})

	if err != nil {
		t.Fatalf("handleNativeToolAction() error = %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status/ok = %d/%v, want 200/true: %#v", status, resp.OK, resp)
	}
	if called {
		t.Fatal("dry-run Temporal mutation connected to Temporal")
	}
}

func TestNativeTemporalStartsAnyWorkflowOnConfiguredTarget(t *testing.T) {
	cfg := nativeTemporalTestConfig()
	fake := &fakeTemporalOpsClient{}
	previous := newTemporalClient
	newTemporalClient = func(_ context.Context, target config.TemporalTarget) (temporalOpsClient, error) {
		if target.Name != "royalty-graph-v2" {
			t.Fatalf("target = %q, want royalty-graph-v2", target.Name)
		}
		return fake, nil
	}
	t.Cleanup(func() { newTemporalClient = previous })

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "temporal"), nativeToolActionRequest{
		Surface:        "temporal",
		Operation:      "start_workflow",
		IdempotencyKey: "temporal-future-workflow",
		Reason:         "start future workflow on configured target",
		Arguments: map[string]any{
			"environment":     "stage",
			"target":          "royalty-graph-v2",
			"new_workflow_id": "future-manager-workflow",
			"workflow_type":   "FutureManagerWorkflow",
			"task_queue":      "FUTURE_MANAGER_TASK_QUEUE",
			"confirm":         true,
		},
	})

	if err != nil {
		t.Fatalf("handleNativeToolAction() error = %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status/ok = %d/%v, want 200/true: %#v", status, resp.OK, resp)
	}
	if fake.startOptions.ID != "future-manager-workflow" {
		t.Fatalf("started workflow id = %q, want future-manager-workflow", fake.startOptions.ID)
	}
	if fake.startOptions.TaskQueue != "FUTURE_MANAGER_TASK_QUEUE" {
		t.Fatalf("started task queue = %q, want FUTURE_MANAGER_TASK_QUEUE", fake.startOptions.TaskQueue)
	}
}

func TestNativeTemporalStartsIndexerManagerWithFailedOnlyReuse(t *testing.T) {
	cfg := nativeTemporalTestConfig()
	cfg.TemporalTargets = append(cfg.TemporalTargets, nativeTemporalIndexerTarget("stage"))
	fake := &fakeTemporalOpsClient{}
	previous := newTemporalClient
	newTemporalClient = func(_ context.Context, target config.TemporalTarget) (temporalOpsClient, error) {
		if target.Name != "indexer" {
			t.Fatalf("target = %q, want indexer", target.Name)
		}
		return fake, nil
	}
	t.Cleanup(func() { newTemporalClient = previous })

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "temporal"), nativeToolActionRequest{
		Surface:        "temporal",
		Operation:      "start_workflow",
		IdempotencyKey: "temporal-root-ip-start",
		Reason:         "restart failed root IP manager",
		Arguments: map[string]any{
			"environment":   "stage",
			"target":        "indexer",
			"workflow_id":   "RootIPManagerWorkflow",
			"workflow_type": "RootIPManagerWorkflow",
			"task_queue":    "ROOT_IP_TASK_QUEUE",
			"args":          []any{float64(5000)},
			"confirm":       true,
		},
	})

	if err != nil {
		t.Fatalf("handleNativeToolAction() error = %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status/ok = %d/%v, want 200/true: %#v", status, resp.OK, resp)
	}
	if fake.startOptions.ID != "RootIPManagerWorkflow" {
		t.Fatalf("started workflow id = %q, want RootIPManagerWorkflow", fake.startOptions.ID)
	}
	if fake.startOptions.TaskQueue != "ROOT_IP_TASK_QUEUE" {
		t.Fatalf("started task queue = %q, want ROOT_IP_TASK_QUEUE", fake.startOptions.TaskQueue)
	}
	if fake.startOptions.WorkflowIDConflictPolicy != enumspb.WORKFLOW_ID_CONFLICT_POLICY_FAIL {
		t.Fatalf("WorkflowIDConflictPolicy = %v, want FAIL", fake.startOptions.WorkflowIDConflictPolicy)
	}
	if fake.startOptions.WorkflowIDReusePolicy != enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY {
		t.Fatalf("WorkflowIDReusePolicy = %v, want ALLOW_DUPLICATE_FAILED_ONLY", fake.startOptions.WorkflowIDReusePolicy)
	}
	if len(fake.startArgs) != 1 || fake.startArgs[0] != float64(5000) {
		t.Fatalf("start args = %#v, want [5000]", fake.startArgs)
	}
}

func TestNativeTemporalStartsFutureIndexerWorkflowOnConfiguredTarget(t *testing.T) {
	cfg := nativeTemporalTestConfig()
	cfg.TemporalTargets = append(cfg.TemporalTargets, nativeTemporalIndexerTarget("stage"))
	fake := &fakeTemporalOpsClient{}
	previous := newTemporalClient
	newTemporalClient = func(_ context.Context, target config.TemporalTarget) (temporalOpsClient, error) {
		if target.Name != "indexer" {
			t.Fatalf("target = %q, want indexer", target.Name)
		}
		return fake, nil
	}
	t.Cleanup(func() { newTemporalClient = previous })

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "temporal"), nativeToolActionRequest{
		Surface:        "temporal",
		Operation:      "start_workflow",
		IdempotencyKey: "temporal-future-indexer-start",
		Reason:         "start future indexer manager",
		Arguments: map[string]any{
			"environment":   "stage",
			"target":        "indexer",
			"workflow_id":   "FutureIndexerManagerWorkflow",
			"workflow_type": "FutureIndexerManagerWorkflow",
			"task_queue":    "FUTURE_INDEXER_TASK_QUEUE",
			"confirm":       true,
		},
	})

	if err != nil {
		t.Fatalf("handleNativeToolAction() error = %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status/ok = %d/%v, want 200/true: %#v", status, resp.OK, resp)
	}
	if fake.startOptions.ID != "FutureIndexerManagerWorkflow" {
		t.Fatalf("started workflow id = %q, want FutureIndexerManagerWorkflow", fake.startOptions.ID)
	}
	if fake.startOptions.TaskQueue != "FUTURE_INDEXER_TASK_QUEUE" {
		t.Fatalf("started task queue = %q, want FUTURE_INDEXER_TASK_QUEUE", fake.startOptions.TaskQueue)
	}
}

func TestNativeTemporalStopUsesGracefulCancel(t *testing.T) {
	cfg := nativeTemporalTestConfig()
	fake := &fakeTemporalOpsClient{}
	previous := newTemporalClient
	newTemporalClient = func(context.Context, config.TemporalTarget) (temporalOpsClient, error) {
		return fake, nil
	}
	t.Cleanup(func() { newTemporalClient = previous })

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "temporal"), nativeToolActionRequest{
		Surface:        "temporal",
		Operation:      "stop_workflow",
		IdempotencyKey: "temporal-stop",
		Reason:         "test graceful stop",
		Arguments: map[string]any{
			"environment": "stage",
			"target":      "royalty-graph-v2",
			"workflow_id": "royalty-graph-manager",
			"confirm":     true,
		},
	})

	if err != nil {
		t.Fatalf("handleNativeToolAction() error = %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status/ok = %d/%v, want 200/true: %#v", status, resp.OK, resp)
	}
	if fake.cancelWorkflowID != "royalty-graph-manager" {
		t.Fatalf("CancelWorkflow workflow id = %q, want royalty-graph-manager", fake.cancelWorkflowID)
	}
}

func TestNativeTemporalRestartRequiresReplacementWorkflowID(t *testing.T) {
	cfg := nativeTemporalTestConfig()
	previous := newTemporalClient
	newTemporalClient = func(context.Context, config.TemporalTarget) (temporalOpsClient, error) {
		t.Fatal("same-ID restart should not connect to Temporal")
		return &fakeTemporalOpsClient{}, nil
	}
	t.Cleanup(func() { newTemporalClient = previous })

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "temporal"), nativeToolActionRequest{
		Surface:        "temporal",
		Operation:      "restart_workflow",
		IdempotencyKey: "temporal-restart-same-id",
		Reason:         "test same-id restart guard",
		Arguments: map[string]any{
			"environment":     "stage",
			"target":          "royalty-graph-v2",
			"workflow_id":     "royalty-graph-manager",
			"new_workflow_id": "royalty-graph-manager",
			"workflow_type":   "RoyaltyGraphManagerWorkflow",
			"task_queue":      "ROYALTY_GRAPH_TASK_QUEUE",
			"confirm":         true,
		},
	})

	if err == nil {
		t.Fatal("expected same-ID restart error")
	}
	if status != http.StatusBadRequest || resp.OK {
		t.Fatalf("status/ok = %d/%v, want 400/false: %#v", status, resp.OK, resp)
	}
}

func TestNativeTemporalRestartStartsReplacementWithoutTerminateExisting(t *testing.T) {
	cfg := nativeTemporalTestConfig()
	fake := &fakeTemporalOpsClient{}
	previous := newTemporalClient
	newTemporalClient = func(context.Context, config.TemporalTarget) (temporalOpsClient, error) {
		return fake, nil
	}
	t.Cleanup(func() { newTemporalClient = previous })

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "temporal"), nativeToolActionRequest{
		Surface:        "temporal",
		Operation:      "restart_workflow",
		IdempotencyKey: "temporal-restart-replacement",
		Reason:         "test restart replacement",
		Arguments: map[string]any{
			"environment":     "stage",
			"target":          "royalty-graph-v2",
			"workflow_id":     "royalty-graph-manager",
			"new_workflow_id": "royalty-graph-manager-replacement",
			"workflow_type":   "RoyaltyGraphManagerWorkflow",
			"task_queue":      "ROYALTY_GRAPH_TASK_QUEUE",
			"confirm":         true,
		},
	})

	if err != nil {
		t.Fatalf("handleNativeToolAction() error = %v", err)
	}
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status/ok = %d/%v, want 200/true: %#v", status, resp.OK, resp)
	}
	if fake.cancelWorkflowID != "royalty-graph-manager" {
		t.Fatalf("CancelWorkflow workflow id = %q, want royalty-graph-manager", fake.cancelWorkflowID)
	}
	if fake.startOptions.WorkflowIDConflictPolicy != enumspb.WORKFLOW_ID_CONFLICT_POLICY_FAIL {
		t.Fatalf("WorkflowIDConflictPolicy = %v, want FAIL", fake.startOptions.WorkflowIDConflictPolicy)
	}
	if fake.startOptions.ID != "royalty-graph-manager-replacement" {
		t.Fatalf("started workflow id = %q, want royalty-graph-manager-replacement", fake.startOptions.ID)
	}
}

func nativeTemporalTestConfig() config.Config {
	cfg := nativeToolsTestConfig()
	cfg.TemporalControlEnabled = true
	cfg.TemporalTargets = []config.TemporalTarget{{
		Environment:        "stage",
		Name:               "royalty-graph-v2",
		HostPort:           "royalty-graph-v2-staging.koyiy.tmprl.cloud:7233",
		Namespace:          "royalty-graph-v2-staging.koyiy",
		CertPEMEnv:         "TEMPORAL_CERT",
		KeyPEMEnv:          "TEMPORAL_KEY",
		AllowedScheduleIDs: []string{"royalty-graph-schedule-v2"},
	}}
	return cfg
}

func nativeTemporalIndexerTarget(environment string) config.TemporalTarget {
	return config.TemporalTarget{
		Environment: environment,
		Name:        "indexer",
		HostPort:    "indexer-" + environment + ".koyiy.tmprl.cloud:7233",
		Namespace:   "indexer-" + environment + ".koyiy",
		CertPEMEnv:  "INDEXER_TEMPORAL_CERT",
		KeyPEMEnv:   "INDEXER_TEMPORAL_KEY",
	}
}

type fakeTemporalOpsClient struct {
	cancelWorkflowID string
	startOptions     temporalclient.StartWorkflowOptions
	startArgs        []any
}

func (f *fakeTemporalOpsClient) Close() {}

func (f *fakeTemporalOpsClient) ListSchedules(context.Context, string, int) ([]*temporalclient.ScheduleListEntry, error) {
	return nil, nil
}

func (f *fakeTemporalOpsClient) DescribeSchedule(context.Context, string) (*temporalclient.ScheduleDescription, error) {
	return &temporalclient.ScheduleDescription{}, nil
}

func (f *fakeTemporalOpsClient) PauseSchedule(context.Context, string, string) error {
	return nil
}

func (f *fakeTemporalOpsClient) UnpauseSchedule(context.Context, string, string) error {
	return nil
}

func (f *fakeTemporalOpsClient) TriggerSchedule(context.Context, string) error {
	return nil
}

func (f *fakeTemporalOpsClient) ListWorkflows(context.Context, string, int) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	return &workflowservice.ListWorkflowExecutionsResponse{}, nil
}

func (f *fakeTemporalOpsClient) CountWorkflows(context.Context, string) (*workflowservice.CountWorkflowExecutionsResponse, error) {
	return &workflowservice.CountWorkflowExecutionsResponse{}, nil
}

func (f *fakeTemporalOpsClient) DescribeWorkflow(context.Context, string, string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	return &workflowservice.DescribeWorkflowExecutionResponse{}, nil
}

func (f *fakeTemporalOpsClient) CancelWorkflow(_ context.Context, workflowID string, _ string) error {
	f.cancelWorkflowID = workflowID
	return nil
}

func (f *fakeTemporalOpsClient) StartWorkflow(_ context.Context, options temporalclient.StartWorkflowOptions, _ string, args []any) (temporalWorkflowRun, error) {
	f.startOptions = options
	f.startArgs = args
	return fakeTemporalWorkflowRun{id: "started", runID: "run-1"}, nil
}

type fakeTemporalWorkflowRun struct {
	id    string
	runID string
}

func (r fakeTemporalWorkflowRun) GetID() string {
	return r.id
}

func (r fakeTemporalWorkflowRun) GetRunID() string {
	return r.runID
}
