package store

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type KanbanTicketStatus string

const (
	KanbanStatusTriage     KanbanTicketStatus = "triage"
	KanbanStatusTodo       KanbanTicketStatus = "todo"
	KanbanStatusInProgress KanbanTicketStatus = "in_progress"
	KanbanStatusBlocked    KanbanTicketStatus = "blocked"
	KanbanStatusDone       KanbanTicketStatus = "done"
	KanbanStatusArchived   KanbanTicketStatus = "archived"
)

var validKanbanStatuses = map[KanbanTicketStatus]bool{
	KanbanStatusTriage:     true,
	KanbanStatusTodo:       true,
	KanbanStatusInProgress: true,
	KanbanStatusBlocked:    true,
	KanbanStatusDone:       true,
	KanbanStatusArchived:   true,
}

var ErrKanbanProjectSlugExists = errors.New("kanban project slug already exists")

type KanbanActor struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Display string `json:"display,omitempty"`
	Surface string `json:"surface,omitempty"`
}

type KanbanProject struct {
	ID          string         `json:"id"`
	Slug        string         `json:"slug"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	State       string         `json:"state"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type KanbanBoard struct {
	ID        string         `json:"id"`
	ProjectID string         `json:"project_id"`
	Slug      string         `json:"slug"`
	Name      string         `json:"name"`
	IsDefault bool           `json:"is_default"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type KanbanTicket struct {
	ID          string             `json:"id"`
	ProjectID   string             `json:"project_id"`
	BoardID     string             `json:"board_id"`
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Status      KanbanTicketStatus `json:"status"`
	Priority    string             `json:"priority,omitempty"`
	Assignee    string             `json:"assignee,omitempty"`
	CreatedBy   string             `json:"created_by"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
	ArchivedAt  *time.Time         `json:"archived_at,omitempty"`
}

type KanbanTicketComment struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
	TicketID      string         `json:"ticket_id"`
	Body          string         `json:"body"`
	ActorType     string         `json:"actor_type"`
	ActorID       string         `json:"actor_id"`
	ActorDisplay  string         `json:"actor_display,omitempty"`
	SourceSurface string         `json:"source_surface,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

type KanbanTicketLink struct {
	ID           string         `json:"id"`
	ProjectID    string         `json:"project_id"`
	FromTicketID string         `json:"from_ticket_id"`
	ToTicketID   string         `json:"to_ticket_id"`
	LinkType     string         `json:"link_type"`
	CreatedBy    string         `json:"created_by"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type KanbanTicketSourceRef struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	TicketID       string         `json:"ticket_id"`
	SourceType     string         `json:"source_type"`
	ActionKind     string         `json:"action_kind"`
	TeamID         string         `json:"team_id,omitempty"`
	ChannelID      string         `json:"channel_id,omitempty"`
	ThreadTS       string         `json:"thread_ts,omitempty"`
	MessageTS      string         `json:"message_ts,omitempty"`
	Permalink      string         `json:"permalink,omitempty"`
	ConversationID string         `json:"conversation_id,omitempty"`
	TraceID        string         `json:"trace_id,omitempty"`
	WorkflowID     string         `json:"workflow_id,omitempty"`
	ProposalID     string         `json:"proposal_id,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type KanbanTicketEvent struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
	TicketID      string         `json:"ticket_id,omitempty"`
	EventType     string         `json:"event_type"`
	ActorType     string         `json:"actor_type"`
	ActorID       string         `json:"actor_id"`
	ActorDisplay  string         `json:"actor_display,omitempty"`
	SourceSurface string         `json:"source_surface,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

type KanbanProjectSlackRoute struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	TeamID    string    `json:"team_id,omitempty"`
	ChannelID string    `json:"channel_id"`
	ThreadTS  string    `json:"thread_ts,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type KanbanProjectCreateInput struct {
	Slug        string
	Name        string
	Description string
	Metadata    map[string]any
	Actor       KanbanActor
}

type KanbanProjectUpdateInput struct {
	Name        *string
	Description *string
	State       *string
	Metadata    map[string]any
	Actor       KanbanActor
}

type KanbanTicketCreateInput struct {
	ProjectID   string
	ProjectSlug string
	BoardID     string
	Title       string
	Description string
	Status      KanbanTicketStatus
	Priority    string
	Assignee    string
	CreatedBy   string
	Metadata    map[string]any
	Actor       KanbanActor
	SourceRefs  []KanbanTicketSourceRefInput
}

type KanbanTicketUpdateInput struct {
	Title       *string
	Description *string
	Status      *KanbanTicketStatus
	Priority    *string
	Assignee    *string
	Metadata    map[string]any
	Actor       KanbanActor
}

type KanbanTicketCommentInput struct {
	Body     string
	Actor    KanbanActor
	Metadata map[string]any
}

type KanbanTicketLinkInput struct {
	FromTicketID string
	ToTicketID   string
	LinkType     string
	CreatedBy    string
	Metadata     map[string]any
	Actor        KanbanActor
}

type KanbanTicketSourceRefInput struct {
	SourceType     string
	ActionKind     string
	TeamID         string
	ChannelID      string
	ThreadTS       string
	MessageTS      string
	Permalink      string
	ConversationID string
	TraceID        string
	WorkflowID     string
	ProposalID     string
	Metadata       map[string]any
}

type KanbanProjectSlackRouteInput struct {
	ProjectID string
	TeamID    string
	ChannelID string
	ThreadTS  string
	Actor     KanbanActor
}

type KanbanBoardSnapshot struct {
	Project    KanbanProject           `json:"project"`
	Board      KanbanBoard             `json:"board"`
	Tickets    []KanbanTicket          `json:"tickets"`
	Comments   []KanbanTicketComment   `json:"comments,omitempty"`
	Links      []KanbanTicketLink      `json:"links,omitempty"`
	SourceRefs []KanbanTicketSourceRef `json:"source_refs,omitempty"`
	Events     []KanbanTicketEvent     `json:"events,omitempty"`
}

type KanbanStore interface {
	ListKanbanProjects() []KanbanProject
	GetKanbanProject(ref string) (KanbanProject, bool)
	CreateKanbanProject(input KanbanProjectCreateInput, now time.Time) (KanbanProject, error)
	UpdateKanbanProject(projectID string, input KanbanProjectUpdateInput, now time.Time) (KanbanProject, error)
	GetKanbanDefaultBoard(projectID string) (KanbanBoard, bool)
	GetKanbanBoardSnapshot(projectRef string) (KanbanBoardSnapshot, bool)
	ListKanbanTickets(projectID string) []KanbanTicket
	GetKanbanTicket(ticketID string) (KanbanTicket, bool)
	CreateKanbanTicket(input KanbanTicketCreateInput, now time.Time) (KanbanTicket, error)
	UpdateKanbanTicket(ticketID string, input KanbanTicketUpdateInput, now time.Time) (KanbanTicket, error)
	AddKanbanTicketComment(ticketID string, input KanbanTicketCommentInput, now time.Time) (KanbanTicketComment, error)
	AddKanbanTicketLink(input KanbanTicketLinkInput, now time.Time) (KanbanTicketLink, error)
	SetKanbanSlackProjectRoute(input KanbanProjectSlackRouteInput, now time.Time) (KanbanProjectSlackRoute, error)
	ListKanbanSlackProjectRoutes(projectID string) []KanbanProjectSlackRoute
	ListKanbanEvents(projectID string, afterID string, limit int) []KanbanTicketEvent
	KanbanEventExists(projectID string, eventID string) bool
	LatestKanbanEventID(projectID string) string
	ResolveKanbanSlackProject(teamID string, channelID string, threadTS string) (KanbanProject, bool)
}

func NewKanbanProject(input KanbanProjectCreateInput, now time.Time) (KanbanProject, KanbanBoard, error) {
	input.Slug = normalizeKanbanSlug(input.Slug)
	input.Name = strings.TrimSpace(input.Name)
	if input.Slug == "" && input.Name != "" {
		input.Slug = normalizeKanbanSlug(input.Name)
	}
	if input.Slug == "" {
		return KanbanProject{}, KanbanBoard{}, errors.New("kanban project slug is required")
	}
	if input.Name == "" {
		input.Name = input.Slug
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	project := KanbanProject{
		ID:          "kproj_" + uuid.NewString(),
		Slug:        input.Slug,
		Name:        input.Name,
		Description: strings.TrimSpace(input.Description),
		State:       "active",
		Metadata:    CloneJSONMap(input.Metadata),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	board := KanbanBoard{
		ID:        "kboard_" + uuid.NewString(),
		ProjectID: project.ID,
		Slug:      "default",
		Name:      "Default",
		IsDefault: true,
		Metadata:  map[string]any{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	return project, board, nil
}

func NewKanbanTicket(input KanbanTicketCreateInput, project KanbanProject, board KanbanBoard, now time.Time) (KanbanTicket, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return KanbanTicket{}, errors.New("kanban ticket title is required")
	}
	status := input.Status
	if status == "" {
		status = KanbanStatusTriage
	}
	if !validKanbanStatuses[status] {
		return KanbanTicket{}, fmt.Errorf("invalid kanban ticket status %q", status)
	}
	if status != KanbanStatusTriage {
		return KanbanTicket{}, fmt.Errorf("new kanban tickets must start in triage status, got %q", status)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	createdBy := strings.TrimSpace(firstNonEmpty(input.CreatedBy, input.Actor.ID, input.Actor.Display, "unknown"))
	item := KanbanTicket{
		ID:          "ktkt_" + uuid.NewString(),
		ProjectID:   project.ID,
		BoardID:     board.ID,
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Status:      status,
		Priority:    strings.TrimSpace(input.Priority),
		Assignee:    strings.TrimSpace(input.Assignee),
		CreatedBy:   createdBy,
		Metadata:    CloneJSONMap(input.Metadata),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if status == KanbanStatusDone {
		completed := now
		item.CompletedAt = &completed
	}
	if status == KanbanStatusArchived {
		archived := now
		item.ArchivedAt = &archived
	}
	return item, nil
}

func KanbanStatusTransitionAllowed(from KanbanTicketStatus, to KanbanTicketStatus) bool {
	if from == to {
		return true
	}
	if !validKanbanStatuses[from] || !validKanbanStatuses[to] {
		return false
	}
	if from == KanbanStatusArchived {
		return to == KanbanStatusTodo
	}
	switch from {
	case KanbanStatusTriage:
		return to == KanbanStatusTodo
	case KanbanStatusTodo:
		return to == KanbanStatusInProgress
	case KanbanStatusInProgress:
		return to == KanbanStatusTodo || to == KanbanStatusBlocked || to == KanbanStatusDone
	case KanbanStatusBlocked:
		return to == KanbanStatusTodo || to == KanbanStatusInProgress
	case KanbanStatusDone:
		return to == KanbanStatusTodo || to == KanbanStatusArchived
	default:
		return false
	}
}

func KanbanActorOrDefault(actor KanbanActor, fallbackSurface string) KanbanActor {
	actor.Type = strings.TrimSpace(actor.Type)
	actor.ID = strings.TrimSpace(actor.ID)
	actor.Display = strings.TrimSpace(actor.Display)
	actor.Surface = strings.TrimSpace(actor.Surface)
	if actor.Type == "" {
		actor.Type = "system"
	}
	if actor.ID == "" {
		actor.ID = "unknown"
	}
	if actor.Surface == "" {
		actor.Surface = strings.TrimSpace(fallbackSurface)
	}
	return actor
}

func KanbanEvent(projectID string, ticketID string, eventType string, actor KanbanActor, payload map[string]any, now time.Time) KanbanTicketEvent {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	actor = KanbanActorOrDefault(actor, "api")
	return KanbanTicketEvent{
		ID:            "kevt_" + uuid.NewString(),
		ProjectID:     strings.TrimSpace(projectID),
		TicketID:      strings.TrimSpace(ticketID),
		EventType:     strings.TrimSpace(eventType),
		ActorType:     actor.Type,
		ActorID:       actor.ID,
		ActorDisplay:  actor.Display,
		SourceSurface: actor.Surface,
		Payload:       CloneJSONMap(payload),
		CreatedAt:     now,
	}
}

func KanbanSlackSourceRefKey(ref KanbanTicketSourceRefInput) string {
	return strings.Join([]string{
		strings.TrimSpace(ref.TeamID),
		strings.TrimSpace(ref.ChannelID),
		strings.TrimSpace(ref.ThreadTS),
		strings.TrimSpace(ref.MessageTS),
		kanbanSourceRefActionKind(ref.ActionKind),
	}, "\x00")
}

func KanbanSlackSourceRefIdentityKey(ref KanbanTicketSourceRefInput) string {
	return strings.Join([]string{
		strings.TrimSpace(ref.ChannelID),
		strings.TrimSpace(ref.ThreadTS),
		strings.TrimSpace(ref.MessageTS),
		kanbanSourceRefActionKind(ref.ActionKind),
	}, "\x00")
}

func KanbanSlackSourceRefComplete(ref KanbanTicketSourceRefInput) bool {
	return strings.TrimSpace(ref.SourceType) == "slack" &&
		strings.TrimSpace(ref.ChannelID) != "" &&
		strings.TrimSpace(ref.ThreadTS) != "" &&
		strings.TrimSpace(ref.MessageTS) != "" &&
		kanbanSourceRefActionKind(ref.ActionKind) != ""
}

func kanbanSourceRefActionKind(actionKind string) string {
	actionKind = strings.TrimSpace(actionKind)
	if actionKind == "" {
		return "reference"
	}
	return actionKind
}

func KanbanSlackProjectRouteKey(teamID string, channelID string, threadTS string) string {
	return strings.Join([]string{
		strings.TrimSpace(teamID),
		strings.TrimSpace(channelID),
		strings.TrimSpace(threadTS),
	}, "\x00")
}

func cloneKanbanProject(item KanbanProject) KanbanProject {
	item.Metadata = CloneJSONMap(item.Metadata)
	return item
}

func cloneKanbanBoard(item KanbanBoard) KanbanBoard {
	item.Metadata = CloneJSONMap(item.Metadata)
	return item
}

func cloneKanbanTicket(item KanbanTicket) KanbanTicket {
	item.Metadata = CloneJSONMap(item.Metadata)
	if item.CompletedAt != nil {
		completed := *item.CompletedAt
		item.CompletedAt = &completed
	}
	if item.ArchivedAt != nil {
		archived := *item.ArchivedAt
		item.ArchivedAt = &archived
	}
	return item
}

func cloneKanbanComment(item KanbanTicketComment) KanbanTicketComment {
	item.Metadata = CloneJSONMap(item.Metadata)
	return item
}

func cloneKanbanLink(item KanbanTicketLink) KanbanTicketLink {
	item.Metadata = CloneJSONMap(item.Metadata)
	return item
}

func cloneKanbanSourceRef(item KanbanTicketSourceRef) KanbanTicketSourceRef {
	item.Metadata = CloneJSONMap(item.Metadata)
	return item
}

func cloneKanbanEvent(item KanbanTicketEvent) KanbanTicketEvent {
	item.Payload = CloneJSONMap(item.Payload)
	return item
}

func cloneKanbanSlackRoute(item KanbanProjectSlackRoute) KanbanProjectSlackRoute {
	return item
}

func sortKanbanProjects(items []KanbanProject) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].State != items[j].State {
			return items[i].State < items[j].State
		}
		return items[i].Name < items[j].Name
	})
}

func sortKanbanTickets(items []KanbanTicket) {
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		}
		return items[i].ID < items[j].ID
	})
}

func sortKanbanEvents(items []KanbanTicketEvent) {
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.Before(items[j].CreatedAt)
		}
		return items[i].ID < items[j].ID
	})
}

func normalizeKanbanSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
