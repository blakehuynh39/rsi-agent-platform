package clients

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNotionClientRetriesRetryableStatus(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			http.Error(w, `{"error":"rate limited"}`, http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"page","id":"page_abc","url":"https://notion.so/page_abc","last_edited_time":"2026-05-02T10:00:00.000Z","created_time":"2026-05-01T10:00:00.000Z","properties":{}}`))
	}))
	defer server.Close()

	client := NewNotionClientWithConfig(NotionClientOptions{
		BaseURL:        server.URL,
		Token:          "ntn-token",
		MaxRetries:     1,
		RetryBaseDelay: time.Millisecond,
	})
	page, err := client.RetrievePage(context.Background(), "page_abc")
	if err != nil {
		t.Fatalf("RetrievePage() error = %v", err)
	}
	if page.ID != "page_abc" {
		t.Fatalf("page id = %q", page.ID)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
}

func TestNotionClientClosesSuccessResponseBody(t *testing.T) {
	var closed int32
	client := NewNotionClientWithConfig(NotionClientOptions{
		BaseURL:    "https://notion.test",
		Token:      "ntn-token",
		MaxRetries: 0,
	})
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body: &closeTrackingBody{
					data:   []byte(`{"object":"page","id":"page_abc","url":"https://notion.so/page_abc","last_edited_time":"2026-05-02T10:00:00.000Z","created_time":"2026-05-01T10:00:00.000Z","properties":{}}`),
					closed: &closed,
				},
			}, nil
		}),
	}

	page, err := client.RetrievePage(context.Background(), "page_abc")
	if err != nil {
		t.Fatalf("RetrievePage() error = %v", err)
	}
	if page.ID != "page_abc" {
		t.Fatalf("page id = %q", page.ID)
	}
	if got := atomic.LoadInt32(&closed); got != 1 {
		t.Fatalf("closed = %d, want 1", got)
	}
}

func TestNotionClientQueryDataSourceUsesDeltaFilterSortAndFilterProperties(t *testing.T) {
	var gotPath string
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.String()
		if r.Header.Get("Notion-Version") != "2026-03-11" {
			t.Fatalf("Notion-Version = %q", r.Header.Get("Notion-Version"))
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false}`))
	}))
	defer server.Close()

	client := NewNotionClientWithConfig(NotionClientOptions{
		BaseURL: server.URL,
		Token:   "ntn-token",
	})
	_, err := client.QueryDataSource(context.Background(), "ds_abc", NotionDataSourceQueryOptions{
		Cursor:                  "cursor_1",
		PageSize:                50,
		LastEditedTimeOnOrAfter: "2026-05-02T09:50:00Z",
		SortTimestamp:           "last_edited_time",
		SortDirection:           "ascending",
		FilterProperties:        []string{"title", "Status"},
	})
	if err != nil {
		t.Fatalf("QueryDataSource() error = %v", err)
	}
	if !strings.HasPrefix(gotPath, "/v1/data_sources/ds_abc/query?") {
		t.Fatalf("path = %q", gotPath)
	}
	query := strings.TrimPrefix(gotPath, "/v1/data_sources/ds_abc/query?")
	if !strings.Contains(query, "filter_properties%5B%5D=title") || !strings.Contains(query, "filter_properties%5B%5D=Status") {
		t.Fatalf("filter_properties query = %q", query)
	}
	if gotPayload["page_size"] != float64(50) || gotPayload["start_cursor"] != "cursor_1" {
		t.Fatalf("unexpected pagination payload: %#v", gotPayload)
	}
	filter, ok := gotPayload["filter"].(map[string]any)
	if !ok || filter["timestamp"] != "last_edited_time" {
		t.Fatalf("unexpected filter payload: %#v", gotPayload["filter"])
	}
	lastEdited, ok := filter["last_edited_time"].(map[string]any)
	if !ok || lastEdited["on_or_after"] != "2026-05-02T09:50:00Z" {
		t.Fatalf("unexpected last_edited_time filter: %#v", filter)
	}
	sorts, ok := gotPayload["sorts"].([]any)
	if !ok || len(sorts) != 1 {
		t.Fatalf("unexpected sorts payload: %#v", gotPayload["sorts"])
	}
	sort0, ok := sorts[0].(map[string]any)
	if !ok || sort0["timestamp"] != "last_edited_time" || sort0["direction"] != "ascending" {
		t.Fatalf("unexpected sort: %#v", sorts[0])
	}
}

func TestNotionClientMutationMethodsUseExpectedEndpoints(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"ok","id":"new_id"}`))
	}))
	defer server.Close()

	client := NewNotionClientWithConfig(NotionClientOptions{BaseURL: server.URL, Token: "ntn-token"})
	if _, err := client.Search(context.Background(), NotionSearchOptions{Query: "tiger"}); err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if _, err := client.CreatePage(context.Background(), map[string]any{"parent": map[string]any{"page_id": "root"}}); err != nil {
		t.Fatalf("CreatePage() error = %v", err)
	}
	if _, err := client.UpdatePage(context.Background(), "page_1", map[string]any{"properties": map[string]any{}}); err != nil {
		t.Fatalf("UpdatePage() error = %v", err)
	}
	if _, err := client.AppendBlockChildren(context.Background(), "block_1", []any{map[string]any{"type": "paragraph"}}); err != nil {
		t.Fatalf("AppendBlockChildren() error = %v", err)
	}
	if _, err := client.UpdateBlock(context.Background(), "block_1", map[string]any{"paragraph": map[string]any{}}); err != nil {
		t.Fatalf("UpdateBlock() error = %v", err)
	}
	if _, err := client.DeleteBlock(context.Background(), "block_1"); err != nil {
		t.Fatalf("DeleteBlock() error = %v", err)
	}
	if _, err := client.CreateComment(context.Background(), map[string]any{"rich_text": []any{}}); err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}

	want := []string{
		"POST /v1/search",
		"POST /v1/pages",
		"PATCH /v1/pages/page_1",
		"PATCH /v1/blocks/block_1/children",
		"PATCH /v1/blocks/block_1",
		"DELETE /v1/blocks/block_1",
		"POST /v1/comments",
	}
	if strings.Join(requests, "\n") != strings.Join(want, "\n") {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type closeTrackingBody struct {
	data   []byte
	offset int
	closed *int32
}

func (b *closeTrackingBody) Read(p []byte) (int, error) {
	if b.offset >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.offset:])
	b.offset += n
	return n, nil
}

func (b *closeTrackingBody) Close() error {
	atomic.AddInt32(b.closed, 1)
	return nil
}
