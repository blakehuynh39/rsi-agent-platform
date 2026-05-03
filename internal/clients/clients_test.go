package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestHonchoClientCorpusAPIsUseV3AndAuthorization(t *testing.T) {
	errCh := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/workspaces" {
			errCh <- fmt.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer honcho-key" {
			errCh <- fmt.Errorf("authorization header = %q", got)
			http.Error(w, "bad auth", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "rsi_company_knowledge"})
	}))
	defer server.Close()

	client := NewHonchoClientWithAPIKey(server.URL, "honcho-key")
	if _, err := client.EnsureWorkspace("rsi_company_knowledge", nil); err != nil {
		t.Fatalf("EnsureWorkspace() error = %v", err)
	}
	select {
	case handlerErr := <-errCh:
		t.Fatal(handlerErr)
	default:
	}
}

func TestHonchoClientCreateConclusionsEnsuresPeers(t *testing.T) {
	var paths []string
	errCh := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if got := r.Header.Get("Authorization"); got != "Bearer honcho-key" {
			errCh <- fmt.Errorf("authorization header = %q", got)
			http.Error(w, "bad auth", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/v3/workspaces/rsi_company_knowledge/peers":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				errCh <- fmt.Errorf("decode peer request: %w", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			name, _ := body["name"].(string)
			if name != "notion_mirror" && name != "story_company" {
				errCh <- fmt.Errorf("unexpected peer name %q", name)
				http.Error(w, "bad peer", http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": name, "name": name})
		case "/v3/workspaces/rsi_company_knowledge/conclusions":
			var body map[string][]HonchoConclusionCreate
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				errCh <- fmt.Errorf("decode conclusion request: %w", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if len(body["conclusions"]) != 2 {
				errCh <- fmt.Errorf("conclusions length = %d", len(body["conclusions"]))
				http.Error(w, "bad conclusions", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode([]HonchoConclusion{{ID: "doc_1"}, {ID: "doc_2"}})
		default:
			errCh <- fmt.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := NewHonchoClientWithAPIKey(server.URL, "honcho-key")
	out, err := client.CreateConclusions("rsi_company_knowledge", []HonchoConclusionCreate{
		{Content: "one", ObserverID: "notion_mirror", ObservedID: "story_company"},
		{Content: "two", ObserverID: "notion_mirror", ObservedID: "story_company"},
	})
	if err != nil {
		t.Fatalf("CreateConclusions() error = %v", err)
	}
	select {
	case handlerErr := <-errCh:
		t.Fatal(handlerErr)
	default:
	}
	if len(out) != 2 {
		t.Fatalf("conclusions returned = %d, want 2", len(out))
	}
	got := strings.Join(paths, ",")
	want := "/v3/workspaces/rsi_company_knowledge/peers,/v3/workspaces/rsi_company_knowledge/peers,/v3/workspaces/rsi_company_knowledge/conclusions"
	if got != want {
		t.Fatalf("request paths = %s, want %s", got, want)
	}
}
