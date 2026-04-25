package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestRunnerClientExecute(t *testing.T) {
	errCh := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/execute" {
			errCh <- fmt.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		var body map[string]RunnerTask
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			errCh <- fmt.Errorf("decode request: %w", err)
			http.Error(w, "decode request", http.StatusBadRequest)
			return
		}
		task := body["task"]
		if task.TaskType != "workflow" || task.Repo != "repo" {
			errCh <- fmt.Errorf("unexpected task payload: %+v", task)
			http.Error(w, "unexpected task payload", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(RunnerResponse{
			OK:       true,
			Provider: "fake",
			Message:  "ok",
			Raw:      map[string]any{"structured_output": map[string]any{}},
		})
	}))
	defer server.Close()

	client := NewRunnerClientWithTimeout(server.URL, time.Second)
	resp, err := client.Execute(RunnerTask{TaskType: "workflow", Repo: "repo"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	select {
	case handlerErr := <-errCh:
		t.Fatal(handlerErr)
	default:
	}
	if !resp.OK || resp.Provider != "fake" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRunnerClientExtendsHTTPTimeoutForExplicitTaskBudget(t *testing.T) {
	client := NewRunnerClientWithTimeout("http://runner.test", time.Second)

	httpClient := client.httpClientForTask(RunnerTask{TaskType: "workflow", TimeoutSeconds: 1800})

	if httpClient.Timeout != 1830*time.Second {
		t.Fatalf("extended timeout = %s, want 1830s", httpClient.Timeout)
	}
	if client.httpClientForTask(RunnerTask{TaskType: "workflow"}).Timeout != time.Second {
		t.Fatalf("expected default timeout for ordinary task")
	}
}

func TestToolGatewayClientExecute(t *testing.T) {
	errCh := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/tools/slack.reply/execute" {
			errCh <- fmt.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			errCh <- fmt.Errorf("decode request: %w", err)
			http.Error(w, "decode request", http.StatusBadRequest)
			return
		}
		if body["channel_id"] != "C123" {
			errCh <- fmt.Errorf("unexpected body: %#v", body)
			http.Error(w, "unexpected body", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(storepkg.ToolResult{
			Name:        "slack.reply",
			ToolCallID:  "call-1",
			Provider:    "slack",
			ProviderRef: "171000001.000100",
			Available:   true,
			Status:      "completed",
			Summary:     "posted",
		})
	}))
	defer server.Close()

	client := NewToolGatewayClient(server.URL)
	resp, err := client.Execute("slack.reply", map[string]any{"channel_id": "C123"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	select {
	case handlerErr := <-errCh:
		t.Fatal(handlerErr)
	default:
	}
	if resp.Provider != "slack" || resp.ProviderRef != "171000001.000100" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHonchoClientRuntimeInitializesEmptyDialecticLevels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/runtimez" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewHonchoClient(server.URL)
	resp, err := client.Runtime()
	if err != nil {
		t.Fatalf("Runtime() error = %v", err)
	}
	if resp.DialecticLevels == nil {
		t.Fatal("expected DialecticLevels to be initialized")
	}
}
