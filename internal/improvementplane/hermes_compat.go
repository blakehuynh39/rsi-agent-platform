package improvementplane

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	hermesassets "github.com/piplabs/rsi-agent-platform/hermes"
	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func registerHermesCompatibilityRoutes(r chi.Router, cfg config.Config, store storepkg.Repository) {
	r.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, buildHermesStatus(cfg, store))
	})
	r.Get("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		limit := boundedQueryInt(r, "limit", 20, 1, 200)
		offset := boundedQueryInt(r, "offset", 0, 0, 1_000_000)
		sessions := buildHermesSessions(cfg, store)
		total := len(sessions)
		if offset > total {
			offset = total
		}
		end := offset + limit
		if end > total {
			end = total
		}
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"sessions": sessions[offset:end],
			"total":    total,
			"limit":    limit,
			"offset":   offset,
		})
	})
	r.Get("/api/sessions/search", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"results": searchHermesSessions(cfg, store, r.URL.Query().Get("q")),
		})
	})
	r.Get("/api/sessions/{sessionID}", func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionID")
		if session, ok := buildHermesSession(cfg, store, sessionID); ok {
			app.WriteJSON(w, http.StatusOK, session)
			return
		}
		app.WriteError(w, http.StatusNotFound, errors.New("session not found"))
	})
	r.Get("/api/sessions/{sessionID}/messages", func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionID")
		messages, ok := buildHermesSessionMessages(store, sessionID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("session not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"session_id": sessionID,
			"messages":   messages,
		})
	})
	r.Get("/api/sessions/{sessionID}/stream", func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionID")
		if _, ok := store.GetTrace(sessionID); !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("session stream is only available for trace-backed sessions"))
			return
		}
		streamHermesSessionEvents(w, r, store, sessionID)
	})
	r.Delete("/api/sessions/{sessionID}", hermesUnsupported("RSI conversations and traces are immutable from the Hermes dashboard"))

	r.Get("/api/skills", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, buildHermesSkills(store))
	})
	r.Put("/api/skills/toggle", hermesUnsupported("RSI skills are managed by runner configuration"))
	r.Get("/api/tools/toolsets", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, buildHermesToolsets(cfg, store))
	})

	r.Get("/api/cron/jobs", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, buildHermesCronJobs(cfg, store))
	})
	r.Post("/api/cron/jobs", hermesUnsupported("RSI cron jobs are managed by platform scheduler configuration"))
	r.Post("/api/cron/jobs/{jobID}/pause", hermesUnsupported("RSI cron jobs are managed by platform scheduler configuration"))
	r.Post("/api/cron/jobs/{jobID}/resume", hermesUnsupported("RSI cron jobs are managed by platform scheduler configuration"))
	r.Post("/api/cron/jobs/{jobID}/trigger", hermesUnsupported("RSI cron jobs are managed by platform scheduler configuration"))
	r.Delete("/api/cron/jobs/{jobID}", hermesUnsupported("RSI cron jobs are managed by platform scheduler configuration"))

	r.Get("/api/dashboard/themes", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"active": "default",
			"themes": []map[string]string{
				{"name": "default", "label": "Default", "description": "Hermes dashboard default theme"},
				{"name": "midnight", "label": "Midnight", "description": "Dark blue dashboard theme"},
				{"name": "ember", "label": "Ember", "description": "Warm dashboard theme"},
				{"name": "mono", "label": "Mono", "description": "Monochrome dashboard theme"},
				{"name": "cyberpunk", "label": "Cyberpunk", "description": "High contrast neon dashboard theme"},
				{"name": "rose", "label": "Rose", "description": "Soft editorial dashboard theme"},
			},
		})
	})
	r.Put("/api/dashboard/theme", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		app.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "theme": firstNonEmptyString(body.Name, "default")})
	})
	r.Get("/api/dashboard/plugins", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, []map[string]any{})
	})
	r.Get("/api/dashboard/plugins/hub", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"plugins":                  []map[string]any{},
			"orphan_dashboard_plugins": []map[string]any{},
			"providers": map[string]any{
				"memory_provider": "rsi-platform",
				"memory_options":  []map[string]string{},
				"context_engine":  "rsi-platform",
				"context_options": []map[string]string{},
			},
		})
	})
	r.Get("/api/dashboard/plugins/rescan", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "count": 0})
	})
	r.Post("/api/dashboard/agent-plugins/install", hermesUnsupported("Dashboard plugin installation is disabled in RSI staging"))
	r.Post("/api/dashboard/agent-plugins/{name}/enable", hermesUnsupported("Dashboard plugin mutation is disabled in RSI staging"))
	r.Post("/api/dashboard/agent-plugins/{name}/disable", hermesUnsupported("Dashboard plugin mutation is disabled in RSI staging"))
	r.Post("/api/dashboard/agent-plugins/{name}/update", hermesUnsupported("Dashboard plugin mutation is disabled in RSI staging"))
	r.Delete("/api/dashboard/agent-plugins/{name}", hermesUnsupported("Dashboard plugin mutation is disabled in RSI staging"))
	r.Put("/api/dashboard/plugin-providers", hermesUnsupported("Dashboard plugin provider mutation is disabled in RSI staging"))
	r.Post("/api/dashboard/plugins/{name}/visibility", hermesUnsupported("Dashboard plugin visibility mutation is disabled in RSI staging"))

	r.Get("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"file":  "rsi-platform",
			"lines": []string{"Logs are not exposed through the Hermes compatibility layer yet."},
		})
	})
	r.Get("/api/analytics/usage", func(w http.ResponseWriter, r *http.Request) {
		days := boundedQueryInt(r, "days", 7, 1, 365)
		app.WriteJSON(w, http.StatusOK, emptyHermesUsageAnalytics(days))
	})
	r.Get("/api/analytics/models", func(w http.ResponseWriter, r *http.Request) {
		days := boundedQueryInt(r, "days", 7, 1, 365)
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"models":      []map[string]any{},
			"period_days": days,
			"totals": map[string]any{
				"distinct_models":      0,
				"total_input":          0,
				"total_output":         0,
				"total_cache_read":     0,
				"total_reasoning":      0,
				"total_estimated_cost": 0,
				"total_actual_cost":    0,
				"total_sessions":       0,
				"total_api_calls":      0,
			},
		})
	})

	r.Get("/api/config", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, hermesConfigSnapshot(cfg))
	})
	r.Put("/api/config", hermesUnsupported("RSI configuration is server-side and cannot be edited from the dashboard"))
	r.Get("/api/config/defaults", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, hermesConfigSnapshot(cfg))
	})
	r.Get("/api/config/schema", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, hermesConfigSchema())
	})
	r.Get("/api/config/raw", func(w http.ResponseWriter, r *http.Request) {
		payload, _ := json.MarshalIndent(hermesConfigSnapshot(cfg), "", "  ")
		app.WriteJSON(w, http.StatusOK, map[string]string{"yaml": string(payload)})
	})
	r.Put("/api/config/raw", hermesUnsupported("RSI configuration is server-side and cannot be edited from the dashboard"))
	r.Get("/api/env", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, hermesEnvSnapshot(cfg))
	})
	r.Put("/api/env", hermesUnsupported("RSI environment variables are managed outside the dashboard"))
	r.Delete("/api/env", hermesUnsupported("RSI environment variables are managed outside the dashboard"))
	r.Post("/api/env/reveal", hermesUnsupported("RSI secret reveal is disabled from the dashboard"))

	r.Get("/api/model/info", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, hermesModelInfo(cfg, store))
	})
	r.Get("/api/model/options", func(w http.ResponseWriter, r *http.Request) {
		info := hermesModelInfo(cfg, store)
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"model":    info["model"],
			"provider": info["provider"],
			"providers": []map[string]any{
				{
					"name":       info["provider"],
					"slug":       info["provider"],
					"models":     []string{fmt.Sprint(info["model"])},
					"is_current": true,
					"source":     "rsi-runner",
				},
			},
		})
	})
	r.Get("/api/model/auxiliary", func(w http.ResponseWriter, r *http.Request) {
		info := hermesModelInfo(cfg, store)
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"tasks": []map[string]any{},
			"main":  map[string]any{"provider": info["provider"], "model": info["model"]},
		})
	})
	r.Post("/api/model/set", hermesUnsupported("RSI model assignment is managed by runner deployment configuration"))

	r.Get("/api/profiles", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"profiles": []map[string]any{
				{
					"name":        "rsi-platform",
					"path":        "server-side",
					"is_default":  true,
					"model":       hermesModelInfo(cfg, store)["model"],
					"provider":    hermesModelInfo(cfg, store)["provider"],
					"has_env":     true,
					"skill_count": len(buildHermesSkills(store)),
				},
			},
		})
	})
	r.Post("/api/profiles", hermesUnsupported("RSI profiles are not editable from the dashboard"))
	r.Patch("/api/profiles/{name}", hermesUnsupported("RSI profiles are not editable from the dashboard"))
	r.Delete("/api/profiles/{name}", hermesUnsupported("RSI profiles are not editable from the dashboard"))
	r.Get("/api/profiles/{name}/setup-command", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]string{"command": "RSI profiles are managed by deployment configuration."})
	})
	r.Get("/api/profiles/{name}/soul", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{"content": "", "exists": false})
	})
	r.Put("/api/profiles/{name}/soul", hermesUnsupported("RSI profile content is not editable from the dashboard"))

	r.Get("/api/providers/oauth", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{"providers": []map[string]any{}})
	})
	r.Delete("/api/providers/oauth/{providerID}", hermesUnsupported("OAuth provider mutation is disabled in RSI staging"))
	r.Post("/api/providers/oauth/{providerID}/start", hermesUnsupported("OAuth provider mutation is disabled in RSI staging"))
	r.Post("/api/providers/oauth/{providerID}/submit", hermesUnsupported("OAuth provider mutation is disabled in RSI staging"))
	r.Get("/api/providers/oauth/{providerID}/poll/{sessionID}", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"session_id":    chi.URLParam(r, "sessionID"),
			"status":        "error",
			"error_message": "OAuth provider mutation is disabled in RSI staging",
		})
	})
	r.Delete("/api/providers/oauth/sessions/{sessionID}", hermesUnsupported("OAuth provider mutation is disabled in RSI staging"))

	r.Post("/api/gateway/restart", hermesUnsupported("Local Hermes gateway restart is disabled in RSI staging"))
	r.Post("/api/hermes/update", hermesUnsupported("Local Hermes update is disabled in RSI staging"))
	r.Get("/api/actions/{actionName}/status", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"name":      chi.URLParam(r, "actionName"),
			"pid":       nil,
			"running":   false,
			"exit_code": 1,
			"lines":     []string{"Local Hermes process actions are disabled in RSI staging."},
		})
	})
}

type hermesSessionInfo struct {
	ID              string  `json:"id"`
	Type            string  `json:"type,omitempty"`
	Source          string  `json:"source"`
	Model           string  `json:"model"`
	Title           string  `json:"title"`
	StartedAt       int64   `json:"started_at"`
	EndedAt         *int64  `json:"ended_at"`
	LastActive      int64   `json:"last_active"`
	IsActive        bool    `json:"is_active"`
	MessageCount    int     `json:"message_count"`
	ToolCallCount   int     `json:"tool_call_count"`
	InputTokens     int     `json:"input_tokens"`
	OutputTokens    int     `json:"output_tokens"`
	Preview         *string `json:"preview"`
	ConversationID  string  `json:"conversation_id,omitempty"`
	TraceID         string  `json:"trace_id,omitempty"`
	ParentSessionID string  `json:"parent_session_id,omitempty"`
	CaseID          string  `json:"case_id,omitempty"`
	TriggerEventID  string  `json:"trigger_event_id,omitempty"`
	ThreadKey       string  `json:"thread_key,omitempty"`
	WorkflowKind    string  `json:"workflow_kind,omitempty"`
	Status          string  `json:"status,omitempty"`
	TraceCount      int     `json:"trace_count,omitempty"`
	OpenTraceCount  int     `json:"open_trace_count,omitempty"`
	ProposalCount   int     `json:"proposal_count,omitempty"`
}

type hermesSessionMessage struct {
	Role       string           `json:"role"`
	Content    *string          `json:"content"`
	ToolCalls  []hermesToolCall `json:"tool_calls,omitempty"`
	ToolName   string           `json:"tool_name,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Timestamp  int64            `json:"timestamp,omitempty"`
}

type hermesToolCall struct {
	ID       string             `json:"id"`
	Function hermesToolFunction `json:"function"`
}

type hermesToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func buildHermesStatus(cfg config.Config, store storepkg.Repository) map[string]any {
	roles := buildHermesRuntimeStatus(cfg, store)
	platforms := map[string]map[string]string{}
	now := time.Now().UTC().Format(time.RFC3339)
	gatewayRunning := false
	gatewayState := "disconnected"
	for _, role := range roles {
		state := "disconnected"
		if strings.EqualFold(role.Status, "disabled") {
			state = "disabled"
		} else if role.Healthy {
			state = "connected"
			gatewayRunning = true
			gatewayState = "connected"
		} else if strings.EqualFold(role.Status, "fatal") {
			state = "fatal"
		}
		item := map[string]string{"state": state, "updated_at": now}
		if role.Error != "" {
			item["error_message"] = role.Error
		}
		platforms[role.Role] = item
	}
	if len(roles) == 0 {
		gatewayState = "unknown"
	}
	return map[string]any{
		"active_sessions":       countActiveHermesSessions(cfg, store),
		"config_path":           "server-side",
		"config_version":        cfg.SchemaVersionCurrent,
		"env_path":              "server-side",
		"gateway_exit_reason":   nil,
		"gateway_health_url":    cfg.PublicBaseURL,
		"gateway_pid":           nil,
		"gateway_platforms":     platforms,
		"gateway_running":       gatewayRunning,
		"gateway_state":         gatewayState,
		"gateway_updated_at":    now,
		"hermes_home":           "rsi-platform",
		"latest_config_version": cfg.SchemaVersionExpected,
		"release_date":          "",
		"version":               "rsi-platform",
	}
}

func buildHermesSessions(cfg config.Config, store storepkg.Repository) []hermesSessionInfo {
	items := []hermesSessionInfo{}
	live := buildHermesLiveSessionIndex(cfg, store)
	traceCountByConversation := map[string]int{}
	for _, trace := range store.ListTraces() {
		if trace.ConversationID != "" {
			traceCountByConversation[trace.ConversationID]++
		}
	}
	for _, conv := range buildConversationList(store) {
		entries := store.ListConversationEntries(conv.ConversationID)
		preview := lastConversationPreview(entries, conv.ActiveCase)
		isLive := live.conversations[conv.ConversationID]
		var endedAt *int64
		if !isLive {
			endedAt = unixPtr(conv.LatestMessageAt)
		}
		items = append(items, hermesSessionInfo{
			ID:             conv.ConversationID,
			Type:           "conversation",
			Source:         firstNonEmptyString(conv.Source, "conversation"),
			Model:          "rsi-platform",
			Title:          firstNonEmptyString(conv.Title, conv.ConversationID),
			StartedAt:      unixSeconds(conv.CreatedAt),
			EndedAt:        endedAt,
			LastActive:     unixSeconds(conv.LatestMessageAt),
			IsActive:       isLive,
			MessageCount:   len(entries),
			ToolCallCount:  0,
			InputTokens:    0,
			OutputTokens:   0,
			Preview:        preview,
			ConversationID: conv.ConversationID,
			ThreadKey:      conv.ExternalKey,
			Status:         conv.Status,
			TraceCount:     traceCountByConversation[conv.ConversationID],
			OpenTraceCount: conv.OpenTraceCount,
			ProposalCount:  conv.ProposalCount,
		})
	}
	for _, trace := range store.ListTraces() {
		last := trace.EndedAt
		if last.IsZero() {
			last = trace.StartedAt
		}
		title := fmt.Sprintf("Trace %s", trace.TraceID)
		if trace.WorkflowKind != "" {
			title = fmt.Sprintf("%s trace %s", trace.WorkflowKind, trace.TraceID)
		}
		preview := stringPtr(firstNonEmptyString(trace.LastVerdict, string(trace.Status)))
		items = append(items, hermesSessionInfo{
			ID:              trace.TraceID,
			Type:            "trace",
			Source:          "trace",
			Model:           firstNonEmptyString(trace.WorkflowKind, "rsi-trace"),
			Title:           title,
			StartedAt:       unixSeconds(trace.StartedAt),
			EndedAt:         unixPtrIfSet(trace.EndedAt),
			LastActive:      unixSeconds(last),
			IsActive:        live.traces[trace.TraceID],
			MessageCount:    trace.EventCount + trace.ReasoningStepCount + trace.ToolCallCount,
			ToolCallCount:   trace.ToolCallCount,
			InputTokens:     0,
			OutputTokens:    0,
			Preview:         preview,
			ConversationID:  trace.ConversationID,
			TraceID:         trace.TraceID,
			ParentSessionID: trace.ConversationID,
			CaseID:          trace.CaseID,
			TriggerEventID:  trace.TriggerEventID,
			ThreadKey:       trace.ThreadKey,
			WorkflowKind:    trace.WorkflowKind,
			Status:          string(trace.Status),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].LastActive == items[j].LastActive {
			return items[i].ID < items[j].ID
		}
		return items[i].LastActive > items[j].LastActive
	})
	return items
}

func buildHermesSession(cfg config.Config, store storepkg.Repository, sessionID string) (hermesSessionInfo, bool) {
	live := buildHermesLiveSessionIndex(cfg, store)
	if conv, ok := store.GetConversation(sessionID); ok {
		entries := store.ListConversationEntries(conv.ID)
		var activeCase *caseSummary
		if conv.ActiveCaseID != "" {
			if c, ok := store.GetCase(conv.ActiveCaseID); ok {
				activeCase = &caseSummary{Summary: c.Summary}
			}
		}
		preview := lastConversationPreview(entries, activeCase)
		isLive := live.conversations[conv.ID]
		var endedAt *int64
		latestMessageAt := conv.UpdatedAt
		if !isLive {
			endedAt = unixPtr(latestMessageAt)
		}
		traces := storeTraceSummariesForConversation(store, conv.ID)
		openTraceCount := 0
		for _, t := range traces {
			if t.StartedAt.After(latestMessageAt) {
				latestMessageAt = t.StartedAt
			}
			if isOpenTraceStatus(t.Status) {
				openTraceCount++
			}
		}
		proposals := storeProposalsByConversation(store, conv.ID)
		return hermesSessionInfo{
			ID:             conv.ID,
			Type:           "conversation",
			Source:         firstNonEmptyString(string(conv.Source), "conversation"),
			Model:          "rsi-platform",
			Title:          firstNonEmptyString(conv.Title, conv.ID),
			StartedAt:      unixSeconds(conv.CreatedAt),
			EndedAt:        endedAt,
			LastActive:     unixSeconds(latestMessageAt),
			IsActive:       isLive,
			MessageCount:   len(entries),
			ToolCallCount:  0,
			InputTokens:    0,
			OutputTokens:   0,
			Preview:        preview,
			ConversationID: conv.ID,
			ThreadKey:      conv.ExternalKey,
			Status:         string(conv.Status),
			TraceCount:     len(traces),
			OpenTraceCount: openTraceCount,
			ProposalCount:  len(proposals),
		}, true
	}
	if trace, ok := store.GetTrace(sessionID); ok {
		summary := trace.Summary
		last := summary.EndedAt
		if last.IsZero() {
			last = summary.StartedAt
		}
		title := fmt.Sprintf("Trace %s", summary.TraceID)
		if summary.WorkflowKind != "" {
			title = fmt.Sprintf("%s trace %s", summary.WorkflowKind, summary.TraceID)
		}
		preview := stringPtr(firstNonEmptyString(summary.LastVerdict, string(summary.Status)))
		return hermesSessionInfo{
			ID:              summary.TraceID,
			Type:            "trace",
			Source:          "trace",
			Model:           firstNonEmptyString(summary.WorkflowKind, "rsi-trace"),
			Title:           title,
			StartedAt:       unixSeconds(summary.StartedAt),
			EndedAt:         unixPtrIfSet(summary.EndedAt),
			LastActive:      unixSeconds(last),
			IsActive:        live.traces[summary.TraceID],
			MessageCount:    summary.EventCount + summary.ReasoningStepCount + summary.ToolCallCount,
			ToolCallCount:   summary.ToolCallCount,
			InputTokens:     0,
			OutputTokens:    0,
			Preview:         preview,
			ConversationID:  summary.ConversationID,
			TraceID:         summary.TraceID,
			ParentSessionID: summary.ConversationID,
			CaseID:          summary.CaseID,
			TriggerEventID:  summary.TriggerEventID,
			ThreadKey:       summary.ThreadKey,
			WorkflowKind:    summary.WorkflowKind,
			Status:          string(summary.Status),
		}, true
	}
	return hermesSessionInfo{}, false
}

type hermesLiveSessionIndex struct {
	conversations map[string]bool
	traces        map[string]bool
}

func buildHermesLiveSessionIndex(cfg config.Config, store storepkg.Repository) hermesLiveSessionIndex {
	out := hermesLiveSessionIndex{
		conversations: map[string]bool{},
		traces:        map[string]bool{},
	}
	now := time.Now().UTC()
	for _, execution := range store.ListActiveRunnerExecutions() {
		if !hermesRunnerExecutionFresh(cfg, execution, now) {
			continue
		}
		traceID := strings.TrimSpace(execution.TraceID)
		if traceID != "" {
			out.traces[traceID] = true
		}
		conversationID := strings.TrimSpace(execution.ConversationID)
		if conversationID == "" && strings.TrimSpace(execution.CaseID) != "" {
			if c, ok := store.GetCase(strings.TrimSpace(execution.CaseID)); ok {
				conversationID = strings.TrimSpace(c.ConversationID)
			}
		}
		if conversationID == "" && traceID != "" {
			if trace, ok := store.GetTrace(traceID); ok {
				conversationID = strings.TrimSpace(trace.Summary.ConversationID)
			}
		}
		if conversationID != "" {
			out.conversations[conversationID] = true
		}
	}
	return out
}

func hermesRunnerExecutionFresh(cfg config.Config, execution storepkg.RunnerExecution, now time.Time) bool {
	timeout := cfg.HermesExecutionHeartbeatTimeout
	if timeout <= 0 {
		return true
	}
	reference := execution.UpdatedAt.UTC()
	if execution.HeartbeatAt != nil {
		reference = execution.HeartbeatAt.UTC()
	} else if reference.IsZero() {
		reference = execution.CreatedAt.UTC()
	}
	if reference.IsZero() {
		return true
	}
	return now.Sub(reference) <= timeout
}

func countActiveHermesSessions(cfg config.Config, store storepkg.Repository) int {
	live := buildHermesLiveSessionIndex(cfg, store)
	return len(live.conversations) + len(live.traces)
}

func searchHermesSessions(cfg config.Config, store storepkg.Repository, query string) []map[string]any {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return []map[string]any{}
	}
	results := []map[string]any{}
	for _, session := range buildHermesSessions(cfg, store) {
		haystack := strings.ToLower(strings.Join([]string{
			session.ID,
			session.Source,
			session.Model,
			session.Title,
			stringValueFromPtr(session.Preview),
		}, " "))
		if !strings.Contains(haystack, query) {
			continue
		}
		results = append(results, map[string]any{
			"session_id":        session.ID,
			"snippet":           highlightSnippet(firstNonEmptyString(stringValueFromPtr(session.Preview), session.Title), query),
			"role":              nil,
			"source":            session.Source,
			"model":             session.Model,
			"session_started":   session.StartedAt,
			"type":              session.Type,
			"conversation_id":   session.ConversationID,
			"trace_id":          session.TraceID,
			"parent_session_id": session.ParentSessionID,
		})
	}
	return results
}

func buildHermesSessionMessages(store storepkg.Repository, sessionID string) ([]hermesSessionMessage, bool) {
	if detail, ok := buildConversationDetailWithOptions(store, sessionID, conversationDetailOptions{
		Includes: map[string]bool{
			"traces":     true,
			"transcript": true,
		},
	}); ok {
		out := make([]hermesSessionMessage, 0, len(detail.Transcript)+len(detail.TraceAttempts))
		for _, entry := range detail.Transcript {
			out = append(out, hermesConversationMessage(entry))
		}
		for _, trace := range detail.TraceAttempts {
			content := fmt.Sprintf("Trace `%s` is `%s` for `%s`.", trace.TraceID, trace.Status, trace.WorkflowKind)
			out = append(out, hermesTextMessage("system", content, trace.StartedAt))
		}
		sortHermesMessages(out)
		return out, true
	}
	trace, ok := store.GetTrace(sessionID)
	if !ok {
		return nil, false
	}
	out := []hermesSessionMessage{}
	for _, ev := range trace.Events {
		content := strings.TrimSpace(strings.Join([]string{ev.EventType, string(ev.Status), ev.Description}, " "))
		out = append(out, hermesTextMessage("system", content, ev.StartedAt))
	}
	for _, step := range trace.Reasoning {
		parts := []string{}
		if step.StepType != "" {
			parts = append(parts, "### "+step.StepType)
		}
		if step.Summary != "" {
			parts = append(parts, step.Summary)
		}
		if step.Decision != "" {
			parts = append(parts, "Decision: "+step.Decision)
		}
		out = append(out, hermesTextMessage("assistant", strings.Join(parts, "\n\n"), step.CreatedAt))
	}
	for _, call := range trace.ToolCalls {
		args, _ := json.Marshal(call.Request)
		content := firstNonEmptyString(call.InterpretationSummary, call.Summary, call.Status)
		msg := hermesTextMessage("tool", content, call.CreatedAt)
		msg.ToolName = call.ToolName
		msg.ToolCallID = call.ToolCallID
		msg.ToolCalls = []hermesToolCall{{
			ID: firstNonEmptyString(call.ToolCallID, call.ID),
			Function: hermesToolFunction{
				Name:      call.ToolName,
				Arguments: string(args),
			},
		}}
		out = append(out, msg)
	}
	sortHermesMessages(out)
	return out, true
}

func hermesConversationMessage(entry conversation.Entry) hermesSessionMessage {
	role := "user"
	entryType := strings.ToLower(entry.EntryType)
	actorType := strings.ToLower(entry.ActorType)
	if strings.Contains(entryType, "system") {
		role = "system"
	} else if strings.Contains(actorType, "bot") || strings.Contains(actorType, "assistant") || strings.Contains(actorType, "service") || strings.Contains(entryType, "response") || strings.Contains(entryType, "reply") {
		role = "assistant"
	}
	content := entry.Body
	if content == "" {
		content = firstNonEmptyString(entry.EntryType, entry.ID)
	}
	return hermesTextMessage(role, content, entry.CreatedAt)
}

func buildHermesSkills(store storepkg.Repository) []map[string]any {
	out := []map[string]any{}
	seen := map[string]bool{}
	for _, skill := range hermesassets.ExportedSkills() {
		if seen[skill.Name] {
			continue
		}
		seen[skill.Name] = true
		out = append(out, map[string]any{
			"name":        skill.Name,
			"description": firstNonEmptyString(skill.Description, fmt.Sprintf("Hermes skill exported from %s.", skill.Path)),
			"category":    firstNonEmptyString(skill.Category, "hermes"),
			"enabled":     true,
		})
	}
	for _, capability := range store.ListCapabilities() {
		if !strings.EqualFold(capability.Kind, "skill") {
			continue
		}
		if seen[capability.Name] {
			continue
		}
		seen[capability.Name] = true
		out = append(out, map[string]any{
			"name":        capability.Name,
			"description": fmt.Sprintf("RSI capability available to %s", strings.Join(capability.AllowedBots, ", ")),
			"category":    "rsi/capability",
			"enabled":     true,
		})
	}
	sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]["name"]) < fmt.Sprint(out[j]["name"]) })
	return out
}

func buildHermesToolsets(cfg config.Config, store storepkg.Repository) []map[string]any {
	out := []map[string]any{}
	tools := []string{}
	for _, capability := range store.ListCapabilities() {
		if strings.EqualFold(capability.Kind, "tool") {
			tools = append(tools, capability.Name)
		}
	}
	sort.Strings(tools)
	if len(tools) > 0 {
		out = append(out, map[string]any{
			"name":        "rsi-tool-gateway",
			"label":       "RSI Tool Gateway",
			"description": "Capabilities registered with the RSI platform tool gateway.",
			"enabled":     true,
			"configured":  true,
			"tools":       tools,
		})
	}
	return out
}

func buildHermesCronJobs(cfg config.Config, store storepkg.Repository) []map[string]any {
	slots := normalizeProposalSlots(store.GetProposalSlots())
	settings := store.GetSettings()
	state := "enabled"
	if cfg.ProposalPromoterInterval <= 0 {
		state = "paused"
	}
	scheduleDisplay := "disabled"
	if cfg.ProposalPromoterInterval > 0 {
		scheduleDisplay = "every " + cfg.ProposalPromoterInterval.String()
	}
	return []map[string]any{
		{
			"id":               "proposal-promoter",
			"name":             "Proposal promoter",
			"prompt":           "Promote eligible problem lines into active RSI proposal slots.",
			"schedule":         map[string]string{"kind": "interval", "expr": cfg.ProposalPromoterInterval.String(), "display": scheduleDisplay},
			"schedule_display": scheduleDisplay,
			"enabled":          cfg.ProposalPromoterInterval > 0,
			"state":            state,
			"deliver":          "rsi-platform",
			"last_run_at":      nil,
			"next_run_at":      nil,
			"last_error":       nil,
		},
		{
			"id":               "proposal-slots",
			"name":             "Proposal slots",
			"prompt":           fmt.Sprintf("Active cap %d; active proposals %d; stale proposals %d.", settings.ActiveProposalCap, len(slots.ActiveProposalIDs), len(slots.StaleProposalIDs)),
			"schedule":         map[string]string{"kind": "state", "expr": "platform", "display": "platform state"},
			"schedule_display": "platform state",
			"enabled":          true,
			"state":            "enabled",
			"deliver":          "rsi-platform",
			"last_run_at":      nil,
			"next_run_at":      nil,
			"last_error":       nil,
		},
	}
}

func emptyHermesUsageAnalytics(days int) map[string]any {
	return map[string]any{
		"daily":    []map[string]any{},
		"by_model": []map[string]any{},
		"totals": map[string]any{
			"total_input":          0,
			"total_output":         0,
			"total_cache_read":     0,
			"total_reasoning":      0,
			"total_estimated_cost": 0,
			"total_actual_cost":    0,
			"total_sessions":       0,
			"total_api_calls":      0,
			"period_days":          days,
		},
		"skills": map[string]any{
			"summary": map[string]any{
				"total_skill_loads":    0,
				"total_skill_edits":    0,
				"total_skill_actions":  0,
				"distinct_skills_used": 0,
			},
			"top_skills": []map[string]any{},
		},
	}
}

func hermesConfigSnapshot(cfg config.Config) map[string]any {
	redactURL := func(url string) string {
		if strings.TrimSpace(url) != "" {
			return "********"
		}
		return ""
	}
	return map[string]any{
		"service": map[string]any{
			"name":            cfg.ServiceName,
			"kind":            cfg.ServiceKind,
			"environment":     cfg.Environment,
			"public_base_url": cfg.PublicBaseURL,
		},
		"runtime": map[string]any{
			"runner_base_url":          redactURL(cfg.RunnerBaseURL),
			"hermes_executor_base_url": redactURL(cfg.HermesExecutorBaseURL),
			"honcho_runtime_base_url":  redactURL(cfg.HonchoRuntimeBaseURL),
		},
		"repos": map[string]any{
			"default_repo":         cfg.DefaultRepo,
			"allowed_target_repos": cfg.AllowedTargetRepos,
		},
	}
}

func hermesConfigSchema() map[string]any {
	field := func(category, description string) map[string]any {
		return map[string]any{
			"category":    category,
			"description": description,
			"type":        "string",
			"read_only":   true,
		}
	}
	return map[string]any{
		"category_order": []string{"service", "runtime", "repos"},
		"fields": map[string]any{
			"service.name":                     field("service", "RSI service name."),
			"service.kind":                     field("service", "RSI service kind."),
			"service.environment":              field("service", "Deployment environment."),
			"service.public_base_url":          field("service", "Public dashboard base URL."),
			"runtime.runner_base_url":          field("runtime", "Default runner base URL."),
			"runtime.hermes_executor_base_url": field("runtime", "Hermes executor base URL."),
			"runtime.honcho_runtime_base_url":  field("runtime", "Honcho runtime base URL."),
			"repos.default_repo":               field("repos", "Default target repository."),
			"repos.allowed_target_repos":       map[string]any{"category": "repos", "description": "Allowed target repositories.", "type": "array", "read_only": true},
		},
	}
}

func hermesEnvSnapshot(cfg config.Config) map[string]any {
	return map[string]any{
		"RSI_PUBLIC_BASE_URL":          hermesEnvInfo(cfg.PublicBaseURL, "Dashboard public base URL.", "platform", false),
		"RSI_RUNNER_BASE_URL":          hermesEnvInfo(cfg.RunnerBaseURL, "Default runner base URL.", "runtime", true),
		"RSI_HERMES_EXECUTOR_BASE_URL": hermesEnvInfo(cfg.HermesExecutorBaseURL, "Hermes executor base URL.", "runtime", true),
		"RSI_HONCHO_RUNTIME_BASE_URL":  hermesEnvInfo(cfg.HonchoRuntimeBaseURL, "Honcho runtime base URL.", "runtime", true),
		"RSI_DEFAULT_REPO":             hermesEnvInfo(cfg.DefaultRepo, "Default target repository.", "platform", false),
		"SLACK_BOT_TOKEN":              hermesEnvInfo(cfg.SlackBotToken, "Slack bot token.", "secrets", true),
		"GITHUB_APP_PRIVATE_KEY":       hermesEnvInfo(cfg.GitHubAppPrivateKey, "GitHub App private key.", "secrets", true),
	}
}

func hermesEnvInfo(value string, description string, category string, password bool) map[string]any {
	isSet := strings.TrimSpace(value) != ""
	redacted := ""
	if isSet {
		if password {
			redacted = "********"
		} else {
			redacted = value
		}
	}
	return map[string]any{
		"is_set":         isSet,
		"redacted_value": nullableString(redacted),
		"description":    description,
		"url":            nil,
		"category":       category,
		"is_password":    password,
		"tools":          []string{},
		"advanced":       password,
	}
}

func hermesModelInfo(cfg config.Config, store storepkg.Repository) map[string]any {
	model := "openai/gpt-5.4"
	provider := "openai"
	for _, role := range buildHermesRuntimeStatus(cfg, store) {
		if role.Model != "" {
			model = role.Model
			provider = firstNonEmptyString(role.Provider, provider)
			break
		}
	}
	return map[string]any{
		"model":                    model,
		"provider":                 provider,
		"auto_context_length":      0,
		"config_context_length":    0,
		"effective_context_length": 0,
		"capabilities": map[string]any{
			"supports_tools":     true,
			"supports_reasoning": true,
		},
	}
}

func buildHermesRuntimeStatus(cfg config.Config, store storepkg.Repository) []runtimeRoleStatus {
	if hermesHasRunnerURL(cfg) {
		return buildRuntimeStatus(cfg, store)
	}
	roles := []string{"prod", "proactive", "eval", "proposal"}
	out := make([]runtimeRoleStatus, 0, len(roles))
	for _, role := range roles {
		out = append(out, runtimeRoleStatus{
			Role:            role,
			Status:          "disabled",
			Model:           "openai/gpt-5.4",
			Provider:        "openai",
			ReasoningEffort: "xhigh",
		})
	}
	return out
}

func hermesHasRunnerURL(cfg config.Config) bool {
	for _, value := range cfg.RunnerURLs() {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func streamHermesSessionEvents(w http.ResponseWriter, r *http.Request, store storepkg.Repository, traceID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		app.WriteError(w, http.StatusInternalServerError, errors.New("streaming not supported"))
		return
	}
	lastEventID := strings.TrimSpace(r.Header.Get("Last-Event-ID"))
	if queryAfter := strings.TrimSpace(r.URL.Query().Get("after")); queryAfter != "" {
		lastEventID = queryAfter
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	sent := map[string]bool{}
	sendBatch := func(backfill bool) {
		items := hermesLedgerStreamEvents(store, traceID)
		if backfill && lastEventID != "" {
			resumeIndex := -1
			for index, item := range items {
				if item.ID == lastEventID {
					resumeIndex = index
					break
				}
			}
			if resumeIndex >= 0 {
				for _, item := range items[:resumeIndex+1] {
					if item.ID != "" {
						sent[item.ID] = true
					}
				}
				items = items[resumeIndex+1:]
			}
		}
		for _, item := range items {
			if item.ID == "" || sent[item.ID] {
				continue
			}
			writeHermesSSEEvent(w, item)
			sent[item.ID] = true
		}
		flusher.Flush()
	}

	sendBatch(true)
	heartbeat := time.NewTicker(10 * time.Second)
	poll := time.NewTicker(1 * time.Second)
	defer heartbeat.Stop()
	defer poll.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-poll.C:
			sendBatch(false)
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

func hermesLedgerStreamEvents(store storepkg.Repository, traceID string) []events.ExecutionLedgerEvent {
	items := store.ListExecutionLedgerEventsByTrace(traceID)
	sort.Slice(items, func(i, j int) bool {
		if items[i].RecordedAt.Equal(items[j].RecordedAt) {
			return items[i].Seq < items[j].Seq
		}
		return items[i].RecordedAt.Before(items[j].RecordedAt)
	})
	return items
}

func writeHermesSSEEvent(w http.ResponseWriter, item events.ExecutionLedgerEvent) {
	eventName := hermesEventName(item)
	payload := map[string]any{
		"type":       eventName,
		"session_id": item.TraceID,
		"payload":    hermesEventPayload(item),
		"timestamp":  item.RecordedAt.Format(time.RFC3339Nano),
		"id":         item.ID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	id := strings.NewReplacer("\n", "", "\r", "").Replace(item.ID)
	_, _ = fmt.Fprintf(w, "id: %s\n", id)
	_, _ = fmt.Fprintf(w, "event: %s\n", eventName)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
}

func hermesEventName(item events.ExecutionLedgerEvent) string {
	kind := strings.ToLower(item.Kind)
	status := strings.ToLower(item.Status)
	if strings.Contains(kind, "tool") || strings.TrimSpace(stringValue(item.Payload["tool_name"])) != "" {
		switch {
		case strings.Contains(status, "start"), strings.Contains(status, "running"), strings.Contains(kind, "start"):
			return "tool.start"
		case strings.Contains(status, "complete"), strings.Contains(status, "success"), strings.Contains(status, "failed"), strings.Contains(kind, "complete"):
			return "tool.complete"
		default:
			return "tool.progress"
		}
	}
	if strings.Contains(kind, "thinking") {
		return "thinking.delta"
	}
	if strings.Contains(kind, "reason") || strings.Contains(kind, "model") {
		return "reasoning.delta"
	}
	if strings.Contains(kind, "message") || strings.Contains(kind, "slack") {
		return "message.delta"
	}
	return "status.update"
}

func hermesEventPayload(item events.ExecutionLedgerEvent) map[string]any {
	payload := map[string]any{
		"ledger_id":    item.ID,
		"execution_id": item.ExecutionID,
		"operation_id": item.OperationID,
		"phase_id":     item.PhaseID,
		"kind":         item.Kind,
		"status":       item.Status,
		"seq":          item.Seq,
	}
	for k, v := range item.Payload {
		if k == "ledger_id" || k == "execution_id" || k == "operation_id" || k == "phase_id" || k == "kind" || k == "status" || k == "seq" {
			continue
		}
		payload[k] = v
	}
	if _, ok := payload["text"]; !ok {
		if text := firstNonEmptyString(stringValue(item.Payload["summary"]), stringValue(item.Payload["message"]), stringValue(item.Payload["status_message"]), item.Status, item.Kind); text != "" {
			payload["text"] = text
		}
	}
	return payload
}

func boundedQueryInt(r *http.Request, key string, fallback int, min int, max int) int {
	value, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get(key)))
	if err != nil {
		return fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func hermesUnsupported(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.WriteError(w, http.StatusMethodNotAllowed, errors.New(message))
	}
}

func lastConversationPreview(entries []conversation.Entry, activeCase *caseSummary) *string {
	for i := len(entries) - 1; i >= 0; i-- {
		if text := strings.TrimSpace(entries[i].Body); text != "" {
			return stringPtr(text)
		}
	}
	if activeCase != nil && activeCase.Summary != "" {
		return stringPtr(activeCase.Summary)
	}
	return nil
}

func hermesTextMessage(role string, content string, timestamp time.Time) hermesSessionMessage {
	return hermesSessionMessage{
		Role:      role,
		Content:   stringPtr(content),
		Timestamp: unixSeconds(timestamp),
	}
}

func sortHermesMessages(items []hermesSessionMessage) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp < items[j].Timestamp
	})
}

func highlightSnippet(text string, query string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	lowerQuery := strings.ToLower(query)

	// Find case-insensitive match by converting rune-by-rune
	textRunes := []rune(text)
	queryRunes := []rune(lowerQuery)
	matchIndex := -1

	for i := 0; i <= len(textRunes)-len(queryRunes); i++ {
		match := true
		for j := 0; j < len(queryRunes); j++ {
			if strings.ToLower(string(textRunes[i+j])) != string(queryRunes[j]) {
				match = false
				break
			}
		}
		if match {
			matchIndex = i
			break
		}
	}

	if matchIndex < 0 {
		if len(textRunes) > 180 {
			return string(textRunes[:180]) + "..."
		}
		return text
	}

	start := matchIndex - 60
	if start < 0 {
		start = 0
	}
	end := matchIndex + len(queryRunes) + 80
	if end > len(textRunes) {
		end = len(textRunes)
	}

	return string(textRunes[start:matchIndex]) + ">>>" + string(textRunes[matchIndex:matchIndex+len(queryRunes)]) + "<<<" + string(textRunes[matchIndex+len(queryRunes):end])
}

func unixSeconds(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func unixPtr(t time.Time) *int64 {
	v := unixSeconds(t)
	return &v
}

func unixPtrIfSet(t time.Time) *int64 {
	if t.IsZero() {
		return nil
	}
	return unixPtr(t)
}

func stringPtr(value string) *string {
	return &value
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func stringValueFromPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
