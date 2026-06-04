package control

import (
	"context"
	"net/http"
	"testing"
	"time"

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

func TestNativeTemporalBlocksDisallowedWorkflowType(t *testing.T) {
	cfg := nativeTemporalTestConfig()
	previous := newTemporalClient
	newTemporalClient = func(context.Context, config.TemporalTarget) (temporalOpsClient, error) {
		t.Fatal("disallowed workflow type should not connect to Temporal")
		return &fakeTemporalOpsClient{}, nil
	}
	t.Cleanup(func() { newTemporalClient = previous })

	resp, status, err := handleNativeToolAction(context.Background(), cfg, storepkg.NewMemoryStore(), nativeToolsValidClaims(time.Now().UTC(), "temporal"), nativeToolActionRequest{
		Surface:        "temporal",
		Operation:      "start_workflow",
		IdempotencyKey: "temporal-bad-type",
		Reason:         "test disallowed type",
		Arguments: map[string]any{
			"environment":     "stage",
			"target":          "royalty-graph-v2",
			"new_workflow_id": "royalty-graph-bad-type",
			"workflow_type":   "BadWorkflow",
			"task_queue":      "ROYALTY_GRAPH_TASK_QUEUE",
			"dry_run":         true,
		},
	})

	if err == nil {
		t.Fatal("expected disallowed workflow type error")
	}
	if status != http.StatusForbidden || resp.OK {
		t.Fatalf("status/ok = %d/%v, want 403/false: %#v", status, resp.OK, resp)
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

func nativeTemporalTestConfig() config.Config {
	cfg := nativeToolsTestConfig()
	cfg.TemporalControlEnabled = true
	cfg.TemporalTargets = []config.TemporalTarget{{
		Environment:               "stage",
		Name:                      "royalty-graph-v2",
		HostPort:                  "royalty-graph-v2-staging.koyiy.tmprl.cloud:7233",
		Namespace:                 "royalty-graph-v2-staging.koyiy",
		CertPEMEnv:                "TEMPORAL_CERT",
		KeyPEMEnv:                 "TEMPORAL_KEY",
		AllowedScheduleIDs:        []string{"royalty-graph-schedule-v2"},
		AllowedWorkflowIDPrefixes: []string{"royalty-graph-"},
		AllowedWorkflowTypes:      []string{"RoyaltyGraphManagerWorkflow"},
		AllowedTaskQueues:         []string{"ROYALTY_GRAPH_TASK_QUEUE"},
	}}
	return cfg
}

type fakeTemporalOpsClient struct {
	cancelWorkflowID string
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

func (f *fakeTemporalOpsClient) StartWorkflow(context.Context, temporalclient.StartWorkflowOptions, string, []any) (temporalWorkflowRun, error) {
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
