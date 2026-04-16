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
