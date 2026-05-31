package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

func NewBaseRouter(cfg config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(timeoutExceptEventStreams(15 * time.Second))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": cfg.ServiceName})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		payload := map[string]any{
			"status":                  "ready",
			"service":                 cfg.ServiceName,
			"service_kind":            cfg.ServiceKind,
			"mode":                    cfg.RuntimeMode,
			"config_validated":        cfg.ConfigValidated,
			"schema_current_version":  cfg.SchemaVersionCurrent,
			"schema_expected_version": cfg.SchemaVersionExpected,
			"schema_state":            cfg.SchemaCompatibility,
			"drain_status":            drainStatus(),
		}
		draining := IsDraining()
		configValid := cfg.ConfigValidated
		if draining && !configValid {
			status = http.StatusServiceUnavailable
			payload["status"] = "draining_not_ready"
		} else if draining {
			status = http.StatusServiceUnavailable
			payload["status"] = "draining"
		} else if !configValid {
			status = http.StatusServiceUnavailable
			payload["status"] = "not_ready"
		}
		WriteJSON(w, status, payload)
	})
	startDrain := func(w http.ResponseWriter, r *http.Request) {
		StartDrain()
		payload := drainStatusPayload()
		payload["accepted"] = true
		WriteJSON(w, http.StatusAccepted, payload)
	}
	// Drain hooks are intentionally unauthenticated. They are internal-only
	// operational endpoints protected by cluster/network boundaries, and need to
	// remain usable during rollout/drain incidents without extra auth wiring.
	r.Post("/internal/drain/start", startDrain)
	r.Get("/internal/drain/start", startDrain)
	r.Get("/internal/drain/status", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, drainStatusPayload())
	})
	r.Get("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"service":          cfg.ServiceName,
			"service_kind":     cfg.ServiceKind,
			"mode":             cfg.RuntimeMode,
			"env":              cfg.Environment,
			"config_validated": cfg.ConfigValidated,
			"store_backend":    cfg.StoreBackend,
			"default_repo":     cfg.DefaultRepo,
			"execution_ledger_first_projection_enabled": cfg.ExecutionLedgerFirstProjection,
			"effect_scheduler_mode":                     EffectSchedulerModeName(cfg.EffectFairClaimEnabled),
			"async_hermes_execution_enabled":            cfg.AsyncHermesExecutionEnabled,
			"deployment_active_execution_policy":        cfg.DeploymentActiveExecutionPolicy,
			"dependencies":                              cfg.DependencyTargets(),
			"schema_current_version":                    cfg.SchemaVersionCurrent,
			"schema_expected_version":                   cfg.SchemaVersionExpected,
			"schema_state":                              cfg.SchemaCompatibility,
			"drain_status":                              drainStatus(),
		})
	})
	return r
}

func timeoutExceptEventStreams(timeout time.Duration) func(http.Handler) http.Handler {
	timeoutMiddleware := middleware.Timeout(timeout)
	return func(next http.Handler) http.Handler {
		timed := timeoutMiddleware(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requestIsEventStream(r) {
				next.ServeHTTP(w, r)
				return
			}
			timed.ServeHTTP(w, r)
		})
	}
}

func requestIsEventStream(r *http.Request) bool {
	if r == nil || r.URL == nil {
		return false
	}
	return strings.HasSuffix(strings.TrimRight(r.URL.Path, "/"), "/stream")
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		panic(fmt.Errorf("encode json response: %w", err))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func ListenAndServe(cfg config.Config, handler http.Handler) error {
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server := &http.Server{Addr: addr, Handler: handler}
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()
	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-signalCh:
		StartDrain()
		if err := shutdownHTTPServer(server); err != nil {
			return err
		}
		select {
		case err := <-errCh:
			if err != nil && err != http.ErrServerClosed {
				return err
			}
		default:
		}
		return nil
	case <-DrainStarted():
		if err := shutdownHTTPServer(server); err != nil {
			return err
		}
		select {
		case err := <-errCh:
			if err != nil && err != http.ErrServerClosed {
				return err
			}
		default:
		}
		return nil
	}
}

func shutdownHTTPServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func drainStatus() string {
	if IsDraining() {
		return "draining"
	}
	return "active"
}

func drainStatusPayload() map[string]any {
	return map[string]any{
		"drain_status": drainStatus(),
		"draining":     IsDraining(),
	}
}

func EffectSchedulerModeName(enabled bool) string {
	if enabled {
		return "fair_claim"
	}
	return "legacy_list_scan"
}

func SanitizedTracePath(traceID string) string {
	return path.Clean("/" + traceID)
}
