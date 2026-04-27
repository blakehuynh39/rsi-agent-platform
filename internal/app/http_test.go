package app

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
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

func TestDrainStatusPayloadUsesDrainStatusField(t *testing.T) {
	StopDrainForTest()
	defer StopDrainForTest()
	cfg := testConfig()
	cfg.Environment = "stage"
	router := NewBaseRouter(cfg)

	statusRec := httptest.NewRecorder()
	router.ServeHTTP(statusRec, httptest.NewRequest(http.MethodGet, "/internal/drain/status", nil))
	if statusRec.Code != http.StatusOK {
		t.Fatalf("status endpoint code = %d, body=%s", statusRec.Code, statusRec.Body.String())
	}
	var statusPayload map[string]any
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusPayload); err != nil {
		t.Fatalf("decode status payload: %v", err)
	}
	if _, ok := statusPayload["status"]; ok {
		t.Fatalf("drain status payload should not expose ambiguous status field: %+v", statusPayload)
	}
	if statusPayload["drain_status"] != "active" || statusPayload["draining"] != false {
		t.Fatalf("unexpected active drain payload: %+v", statusPayload)
	}

	startRec := httptest.NewRecorder()
	router.ServeHTTP(startRec, httptest.NewRequest(http.MethodPost, "/internal/drain/start", nil))
	if startRec.Code != http.StatusAccepted {
		t.Fatalf("start endpoint code = %d, body=%s", startRec.Code, startRec.Body.String())
	}
	var startPayload map[string]any
	if err := json.Unmarshal(startRec.Body.Bytes(), &startPayload); err != nil {
		t.Fatalf("decode start payload: %v", err)
	}
	if _, ok := startPayload["status"]; ok {
		t.Fatalf("drain start payload should not expose ambiguous status field: %+v", startPayload)
	}
	if startPayload["drain_status"] != "draining" || startPayload["draining"] != true || startPayload["accepted"] != true {
		t.Fatalf("unexpected draining payload: %+v", startPayload)
	}
}

func TestListenAndServeStopsWhenDrainStarts(t *testing.T) {
	StopDrainForTest()
	defer StopDrainForTest()

	done := make(chan error, 1)
	go func() {
		done <- ListenAndServe(config.Config{HTTPPort: 0}, http.NewServeMux())
	}()

	time.Sleep(10 * time.Millisecond)
	StartDrain()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ListenAndServe() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ListenAndServe did not stop after drain started")
	}
}

func TestListenAndServeReportsBindErrorEvenWhenDraining(t *testing.T) {
	StopDrainForTest()
	defer StopDrainForTest()

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen on test port: %v", err)
	}
	defer listener.Close()
	_, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split listener address: %v", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("parse listener port: %v", err)
	}

	StartDrain()
	if err := ListenAndServe(config.Config{HTTPPort: port}, http.NewServeMux()); err == nil {
		t.Fatal("ListenAndServe should report bind errors before observing drain")
	}
}

func TestStopDrainForTestStopsSignalDrainWithoutRestartingDrain(t *testing.T) {
	StopDrainForTest()
	InstallSignalDrain()
	StopDrainForTest()
	time.Sleep(10 * time.Millisecond)
	if IsDraining() {
		t.Fatal("StopDrainForTest should not wake signal goroutine into StartDrain")
	}
}

func testConfig() config.Config {
	return config.Config{ServiceName: "test-service", Environment: "development"}
}
