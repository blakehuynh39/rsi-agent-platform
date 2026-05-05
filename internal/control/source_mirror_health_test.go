package control

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

func TestCheckNotionMirrorRootsAcceptsDatabaseRootTypeMismatch(t *testing.T) {
	rootID := normalizeNotionID("34f05129-9a54-8064-9b98-c7337f8c8084")
	pageCalls := 0
	databaseCalls := 0
	notion := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/pages/"+rootID:
			pageCalls++
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"object":"error","status":400,"code":"validation_error","message":"Provided ID 34f05129-9a54-8064-9b98-c7337f8c8084 is a database, not a page. Use the retrieve database API instead."}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/databases/"+rootID:
			databaseCalls++
			_, _ = w.Write([]byte(`{"object":"database","id":"` + rootID + `","url":"https://notion.so/` + rootID + `","title":[],"properties":{},"archived":false,"in_trash":false}`))
		default:
			t.Errorf("unexpected notion request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer notion.Close()

	err := checkNotionMirrorRoots(context.Background(), config.Config{
		NotionAPIBaseURL:      notion.URL,
		NotionToken:           "secret",
		NotionAPIVersion:      "2026-03-11",
		NotionMirrorAllowlist: []string{rootID},
	})

	if err != nil {
		t.Fatalf("checkNotionMirrorRoots() error = %v", err)
	}
	if pageCalls != 1 || databaseCalls != 1 {
		t.Fatalf("notion calls page=%d database=%d, want 1 each", pageCalls, databaseCalls)
	}
}

func TestNotionTypeMismatchClassifierRejectsUnrelatedValidationError(t *testing.T) {
	err := notionEndpointTypeMismatchError("database", "page")
	if !isNotionPageEndpointTypeMismatch(err) {
		t.Fatal("expected page endpoint database type mismatch")
	}
	unrelated := err
	unrelated.Body = strings.ReplaceAll(unrelated.Body, "is a database, not a page", "is invalid")
	if isNotionPageEndpointTypeMismatch(unrelated) {
		t.Fatal("unrelated validation error should not be classified as a Notion type mismatch")
	}
}
