package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
)

func TestHermesExecutorPoolSelectsReadyNonDrainingEndpoint(t *testing.T) {
	draining := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":                 "ok",
				"available":              true,
				"drain_status":           "draining",
				"executor_instance_id":   "executor-draining",
				"active_execution_count": 0,
			})
			return
		}
		t.Fatalf("draining endpoint received unexpected request %s", r.URL.Path)
	}))
	defer draining.Close()

	active := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/readyz":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":                 "ok",
				"available":              true,
				"drain_status":           "active",
				"executor_instance_id":   "executor-active",
				"active_execution_count": 1,
			})
		case "/internal/hermes-executions":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"execution_id": "hexec-1",
				"status":       "accepted",
			})
		default:
			t.Fatalf("active endpoint received unexpected request %s", r.URL.Path)
		}
	}))
	defer active.Close()

	pool := newHermesExecutorPool(
		config.Config{
			HermesExecutorPoolURLs: []string{draining.URL, active.URL},
		},
		"prod",
		clients.NewRunnerClient(draining.URL),
	)
	status, endpoint, err := pool.startExecution(clients.RunnerTask{ExecutionID: "hexec-1"})
	if err != nil {
		t.Fatalf("startExecution() error = %v", err)
	}
	if status.Status != "accepted" {
		t.Fatalf("status = %q, want accepted", status.Status)
	}
	if endpoint.instanceID != "executor-active" {
		t.Fatalf("endpoint.instanceID = %q, want executor-active", endpoint.instanceID)
	}
}
