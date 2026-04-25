package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONPanicsOnEncodeError(t *testing.T) {
	recorder := httptest.NewRecorder()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected WriteJSON to panic on encode failure")
		}
	}()

	WriteJSON(recorder, http.StatusOK, map[string]any{
		"bad": make(chan int),
	})
}

func TestRequestIsEventStream(t *testing.T) {
	streamReq := httptest.NewRequest(http.MethodGet, "/api/traces/trace-1/stream?scope=all", nil)
	if !requestIsEventStream(streamReq) {
		t.Fatal("expected trace stream path to bypass the short request timeout")
	}

	acceptReq := httptest.NewRequest(http.MethodGet, "/api/anything", nil)
	acceptReq.Header.Set("Accept", "text/event-stream")
	if requestIsEventStream(acceptReq) {
		t.Fatal("expected event-stream Accept header alone to keep the short request timeout")
	}

	apiReq := httptest.NewRequest(http.MethodGet, "/api/traces/trace-1", nil)
	if requestIsEventStream(apiReq) {
		t.Fatal("expected ordinary API request to keep the short request timeout")
	}
}
