package clients

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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
