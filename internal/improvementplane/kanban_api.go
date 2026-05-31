package improvementplane

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type kanbanProjectRequest struct {
	Slug        string         `json:"slug"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	State       *string        `json:"state,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Actor       *kanbanActorIn `json:"actor,omitempty"`
}

type kanbanTicketRequest struct {
	ProjectID   string                                `json:"project_id"`
	ProjectSlug string                                `json:"project_slug"`
	BoardID     string                                `json:"board_id"`
	Title       string                                `json:"title"`
	Description string                                `json:"description"`
	Priority    string                                `json:"priority"`
	Assignee    string                                `json:"assignee"`
	Metadata    map[string]any                        `json:"metadata,omitempty"`
	Actor       *kanbanActorIn                        `json:"actor,omitempty"`
	SourceRefs  []storepkg.KanbanTicketSourceRefInput `json:"source_refs,omitempty"`
}

type kanbanTicketPatchRequest struct {
	Title       *string                      `json:"title,omitempty"`
	Description *string                      `json:"description,omitempty"`
	Status      *storepkg.KanbanTicketStatus `json:"status,omitempty"`
	Priority    *string                      `json:"priority,omitempty"`
	Assignee    *string                      `json:"assignee,omitempty"`
	Metadata    map[string]any               `json:"metadata,omitempty"`
	Actor       *kanbanActorIn               `json:"actor,omitempty"`
}

type kanbanCommentRequest struct {
	Body     string         `json:"body"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Actor    *kanbanActorIn `json:"actor,omitempty"`
}

type kanbanLinkRequest struct {
	FromTicketID string         `json:"from_ticket_id"`
	ToTicketID   string         `json:"to_ticket_id"`
	LinkType     string         `json:"link_type"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Actor        *kanbanActorIn `json:"actor,omitempty"`
}

type kanbanSlackRouteRequest struct {
	TeamID    string         `json:"team_id"`
	ChannelID string         `json:"channel_id"`
	ThreadTS  string         `json:"thread_ts"`
	Actor     *kanbanActorIn `json:"actor,omitempty"`
}

type kanbanActorIn struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Display string `json:"display"`
	Surface string `json:"surface"`
}

func decodeKanbanTicketRequest(r *http.Request) (kanbanTicketRequest, error) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return kanbanTicketRequest{}, err
	}
	if _, ok := raw["status"]; ok {
		return kanbanTicketRequest{}, errors.New("new kanban tickets always start in triage; update status after creation")
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		return kanbanTicketRequest{}, err
	}
	var body kanbanTicketRequest
	if err := json.Unmarshal(payload, &body); err != nil {
		return kanbanTicketRequest{}, err
	}
	return body, nil
}

func registerKanbanRoutes(r chi.Router, repo storepkg.Repository) {
	kanban := storepkg.KanbanStore(repo)

	r.Get("/api/kanban/projects", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{"projects": kanban.ListKanbanProjects()})
	})
	r.Post("/api/kanban/projects", func(w http.ResponseWriter, r *http.Request) {
		var body kanbanProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		actor := kanbanActorFromRequest(r, body.Actor, "dashboard")
		item, err := kanban.CreateKanbanProject(storepkg.KanbanProjectCreateInput{
			Slug:        body.Slug,
			Name:        body.Name,
			Description: body.Description,
			Metadata:    body.Metadata,
			Actor:       actor,
		}, time.Now().UTC())
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, map[string]any{"project": item})
	})
	r.Patch("/api/kanban/projects/{projectID}", func(w http.ResponseWriter, r *http.Request) {
		var body kanbanProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		actor := kanbanActorFromRequest(r, body.Actor, "dashboard")
		item, err := kanban.UpdateKanbanProject(chi.URLParam(r, "projectID"), storepkg.KanbanProjectUpdateInput{
			Name:        optionalStringFromValue(body.Name),
			Description: optionalStringFromValue(body.Description),
			State:       body.State,
			Metadata:    body.Metadata,
			Actor:       actor,
		}, time.Now().UTC())
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusOK, map[string]any{"project": item})
	})
	r.Get("/api/kanban/projects/{projectID}/board", func(w http.ResponseWriter, r *http.Request) {
		snapshot, ok := kanban.GetKanbanBoardSnapshot(chi.URLParam(r, "projectID"))
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("kanban project not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, snapshot)
	})
	r.Get("/api/kanban/projects/{projectID}/slack-routes", func(w http.ResponseWriter, r *http.Request) {
		projectID := strings.TrimSpace(chi.URLParam(r, "projectID"))
		app.WriteJSON(w, http.StatusOK, map[string]any{"routes": kanban.ListKanbanSlackProjectRoutes(projectID)})
	})
	r.Post("/api/kanban/projects/{projectID}/slack-routes", func(w http.ResponseWriter, r *http.Request) {
		var body kanbanSlackRouteRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		actor := kanbanActorFromRequest(r, body.Actor, "dashboard")
		route, err := kanban.SetKanbanSlackProjectRoute(storepkg.KanbanProjectSlackRouteInput{
			ProjectID: chi.URLParam(r, "projectID"),
			TeamID:    body.TeamID,
			ChannelID: body.ChannelID,
			ThreadTS:  body.ThreadTS,
			Actor:     actor,
		}, time.Now().UTC())
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, map[string]any{"route": route})
	})
	r.Post("/api/kanban/tickets", func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeKanbanTicketRequest(r)
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		actor := kanbanActorFromRequest(r, body.Actor, "dashboard")
		item, err := kanban.CreateKanbanTicket(storepkg.KanbanTicketCreateInput{
			ProjectID:   body.ProjectID,
			ProjectSlug: body.ProjectSlug,
			BoardID:     body.BoardID,
			Title:       body.Title,
			Description: body.Description,
			Priority:    body.Priority,
			Assignee:    body.Assignee,
			CreatedBy:   actor.ID,
			Metadata:    body.Metadata,
			Actor:       actor,
			SourceRefs:  body.SourceRefs,
		}, time.Now().UTC())
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, map[string]any{"ticket": item})
	})
	r.Get("/api/kanban/tickets/{ticketID}", func(w http.ResponseWriter, r *http.Request) {
		item, ok := kanban.GetKanbanTicket(chi.URLParam(r, "ticketID"))
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("kanban ticket not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, map[string]any{"ticket": item})
	})
	r.Patch("/api/kanban/tickets/{ticketID}", func(w http.ResponseWriter, r *http.Request) {
		var body kanbanTicketPatchRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		actor := kanbanActorFromRequest(r, body.Actor, "dashboard")
		item, err := kanban.UpdateKanbanTicket(chi.URLParam(r, "ticketID"), storepkg.KanbanTicketUpdateInput{
			Title:       body.Title,
			Description: body.Description,
			Status:      body.Status,
			Priority:    body.Priority,
			Assignee:    body.Assignee,
			Metadata:    body.Metadata,
			Actor:       actor,
		}, time.Now().UTC())
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusOK, map[string]any{"ticket": item})
	})
	r.Post("/api/kanban/tickets/{ticketID}/comments", func(w http.ResponseWriter, r *http.Request) {
		var body kanbanCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		actor := kanbanActorFromRequest(r, body.Actor, "dashboard")
		item, err := kanban.AddKanbanTicketComment(chi.URLParam(r, "ticketID"), storepkg.KanbanTicketCommentInput{
			Body:     body.Body,
			Metadata: body.Metadata,
			Actor:    actor,
		}, time.Now().UTC())
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, map[string]any{"comment": item})
	})
	r.Post("/api/kanban/tickets/{ticketID}/links", func(w http.ResponseWriter, r *http.Request) {
		var body kanbanLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		fromTicketID := strings.TrimSpace(body.FromTicketID)
		if fromTicketID == "" {
			fromTicketID = chi.URLParam(r, "ticketID")
		}
		actor := kanbanActorFromRequest(r, body.Actor, "dashboard")
		item, err := kanban.AddKanbanTicketLink(storepkg.KanbanTicketLinkInput{
			FromTicketID: fromTicketID,
			ToTicketID:   body.ToTicketID,
			LinkType:     body.LinkType,
			CreatedBy:    actor.ID,
			Metadata:     body.Metadata,
			Actor:        actor,
		}, time.Now().UTC())
		if err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		app.WriteJSON(w, http.StatusCreated, map[string]any{"link": item})
	})
	r.Get("/api/kanban/events", func(w http.ResponseWriter, r *http.Request) {
		projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
		after := strings.TrimSpace(r.URL.Query().Get("after"))
		limit := intQuery(r, "limit", 100)
		app.WriteJSON(w, http.StatusOK, map[string]any{"events": kanban.ListKanbanEvents(projectID, after, limit)})
	})
	r.Get("/api/kanban/stream", func(w http.ResponseWriter, r *http.Request) {
		streamKanbanEvents(w, r, kanban)
	})
}

func streamKanbanEvents(w http.ResponseWriter, r *http.Request, kanban storepkg.KanbanStore) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		app.WriteError(w, http.StatusInternalServerError, errors.New("streaming not supported"))
		return
	}
	projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
	after := strings.TrimSpace(r.URL.Query().Get("after"))
	explicitAfter := after != ""
	if header := strings.TrimSpace(r.Header.Get("Last-Event-ID")); header != "" && after == "" {
		after = header
		explicitAfter = true
	}
	after = kanbanStreamStartCursor(kanban, projectID, after, explicitAfter)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	send := func() {
		events := kanban.ListKanbanEvents(projectID, after, 100)
		for _, event := range events {
			writeKanbanSSEEvent(w, event)
			after = event.ID
		}
		flusher.Flush()
	}
	send()
	heartbeat := time.NewTicker(10 * time.Second)
	poll := time.NewTicker(1 * time.Second)
	defer heartbeat.Stop()
	defer poll.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-poll.C:
			send()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

func kanbanStreamStartCursor(kanban storepkg.KanbanStore, projectID string, after string, explicitAfter bool) string {
	after = strings.TrimSpace(after)
	if after == "" {
		return kanban.LatestKanbanEventID(projectID)
	}
	if kanban.KanbanEventExists(projectID, after) {
		return after
	}
	if explicitAfter {
		return ""
	}
	return kanban.LatestKanbanEventID(projectID)
}

func writeKanbanSSEEvent(w http.ResponseWriter, event storepkg.KanbanTicketEvent) {
	payload, _ := json.Marshal(event)
	_, _ = fmt.Fprintf(w, "id: %s\n", event.ID)
	_, _ = fmt.Fprint(w, "event: kanban.event\n")
	_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
}

func kanbanActor(input *kanbanActorIn, surface string) storepkg.KanbanActor {
	if input == nil {
		return storepkg.KanbanActor{Type: "user", ID: "dashboard", Display: "Dashboard", Surface: surface}
	}
	return storepkg.KanbanActor{Type: input.Type, ID: input.ID, Display: input.Display, Surface: firstNonEmptyString(input.Surface, surface)}
}

func kanbanActorFromRequest(r *http.Request, input *kanbanActorIn, surface string) storepkg.KanbanActor {
	if actor, ok := kanbanDashboardActorFromHeaders(r, surface); ok {
		return actor
	}
	return kanbanActor(input, surface)
}

func kanbanDashboardActorFromHeaders(r *http.Request, surface string) (storepkg.KanbanActor, bool) {
	if r == nil {
		return storepkg.KanbanActor{}, false
	}
	cloudflareClaims := kanbanJWTClaimsFromHeader(r.Header.Get("Cf-Access-Jwt-Assertion"))
	oidcClaims := kanbanJWTClaimsFromHeader(r.Header.Get("X-Amzn-Oidc-Data"))
	email := firstNonEmptyString(
		kanbanIdentityHeader(r, "Cf-Access-Authenticated-User-Email"),
		kanbanClaimString(cloudflareClaims, "email"),
		kanbanIdentityHeader(r, "X-Auth-Request-Email"),
		kanbanIdentityHeader(r, "X-Forwarded-Email"),
		kanbanIdentityHeader(r, "X-Pomerium-Claim-Email"),
		kanbanIdentityHeader(r, "X-Authenticated-User-Email"),
		kanbanClaimString(oidcClaims, "email"),
	)
	email = strings.ToLower(email)
	if !kanbanIsCompanyEmail(email) {
		return storepkg.KanbanActor{}, false
	}
	display := firstNonEmptyString(
		kanbanClaimString(cloudflareClaims, "name"),
		kanbanClaimString(cloudflareClaims, "preferred_username"),
		kanbanIdentityHeader(r, "X-Auth-Request-Name"),
		kanbanIdentityHeader(r, "X-Auth-Request-Preferred-Username"),
		kanbanIdentityHeader(r, "X-Auth-Request-User"),
		kanbanIdentityHeader(r, "X-Forwarded-User"),
		kanbanIdentityHeader(r, "X-Pomerium-Claim-Name"),
		kanbanIdentityHeader(r, "X-Pomerium-Claim-User"),
		kanbanIdentityHeader(r, "X-Authenticated-User"),
		kanbanClaimString(oidcClaims, "name"),
		kanbanClaimString(oidcClaims, "preferred_username"),
	)
	id := email
	if id == "" {
		return storepkg.KanbanActor{}, false
	}
	if display == "" || strings.EqualFold(display, email) || strings.Contains(display, "@") {
		display = firstNonEmptyString(kanbanDisplayNameFromEmail(email), kanbanDisplayNameFromEmail(display))
	}
	display = firstNonEmptyString(display, id)
	return storepkg.KanbanActor{Type: "user", ID: id, Display: display, Surface: surface}, true
}

func kanbanIsCompanyEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	return strings.HasSuffix(email, "@piplabs.xyz")
}

func kanbanIdentityHeader(r *http.Request, key string) string {
	if r == nil {
		return ""
	}
	return kanbanCleanIdentityValue(r.Header.Get(key))
}

func kanbanCleanIdentityValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.Contains(value, ",") {
		value = strings.TrimSpace(strings.Split(value, ",")[0])
	}
	value = strings.Trim(value, `"'`)
	value = strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(value)
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 200 {
		value = value[:200]
	}
	return value
}

func kanbanJWTClaimsFromHeader(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payload, err = base64.URLEncoding.DecodeString(parts[1])
	}
	if err != nil {
		return nil
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil
	}
	return claims
}

func kanbanClaimString(claims map[string]any, key string) string {
	if claims == nil {
		return ""
	}
	value, ok := claims[key]
	if !ok || value == nil {
		return ""
	}
	return kanbanCleanIdentityValue(fmt.Sprint(value))
}

func kanbanDisplayNameFromEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return ""
	}
	local := strings.Split(email, "@")[0]
	local = strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(local)
	parts := strings.Fields(local)
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, " ")
}

func optionalStringFromValue(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func intQuery(r *http.Request, key string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
