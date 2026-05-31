package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *MemoryStore) ListKanbanProjects() []KanbanProject {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]KanbanProject, 0, len(s.kanbanProjects))
	for _, item := range s.kanbanProjects {
		out = append(out, cloneKanbanProject(item))
	}
	sortKanbanProjects(out)
	return out
}

func (s *MemoryStore) GetKanbanProject(ref string) (KanbanProject, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getKanbanProjectLocked(ref)
}

func (s *MemoryStore) getKanbanProjectLocked(ref string) (KanbanProject, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return KanbanProject{}, false
	}
	if item, ok := s.kanbanProjects[ref]; ok {
		return cloneKanbanProject(item), true
	}
	if id, ok := s.kanbanProjectBySlug[normalizeKanbanSlug(ref)]; ok {
		item, ok := s.kanbanProjects[id]
		return cloneKanbanProject(item), ok
	}
	return KanbanProject{}, false
}

func (s *MemoryStore) CreateKanbanProject(input KanbanProjectCreateInput, now time.Time) (KanbanProject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, board, err := NewKanbanProject(input, now)
	if err != nil {
		return KanbanProject{}, err
	}
	if existingID, ok := s.kanbanProjectBySlug[project.Slug]; ok {
		return KanbanProject{}, fmt.Errorf("%w: %s", ErrKanbanProjectSlugExists, s.kanbanProjects[existingID].Slug)
	}
	s.kanbanProjects[project.ID] = project
	s.kanbanProjectBySlug[project.Slug] = project.ID
	s.kanbanBoards[board.ID] = board
	s.kanbanDefaultBoardByProject[project.ID] = board.ID
	s.kanbanEvents = append(s.kanbanEvents, KanbanEvent(project.ID, "", "project.created", input.Actor, map[string]any{"slug": project.Slug, "name": project.Name}, now))
	return cloneKanbanProject(project), nil
}

func (s *MemoryStore) UpdateKanbanProject(projectID string, input KanbanProjectUpdateInput, now time.Time) (KanbanProject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.kanbanProjects[strings.TrimSpace(projectID)]
	if !ok {
		return KanbanProject{}, errors.New("kanban project not found")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if input.Name != nil {
		project.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		project.Description = strings.TrimSpace(*input.Description)
	}
	if input.State != nil {
		state := strings.TrimSpace(*input.State)
		if state != "active" && state != "archived" {
			return KanbanProject{}, fmt.Errorf("invalid kanban project state %q", state)
		}
		project.State = state
	}
	if input.Metadata != nil {
		project.Metadata = CloneJSONMap(input.Metadata)
	}
	project.UpdatedAt = now
	s.kanbanProjects[project.ID] = project
	s.kanbanEvents = append(s.kanbanEvents, KanbanEvent(project.ID, "", "project.updated", input.Actor, map[string]any{"state": project.State}, now))
	return cloneKanbanProject(project), nil
}

func (s *MemoryStore) GetKanbanDefaultBoard(projectID string) (KanbanBoard, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.kanbanDefaultBoardByProject[strings.TrimSpace(projectID)]
	if !ok {
		return KanbanBoard{}, false
	}
	board, ok := s.kanbanBoards[id]
	return cloneKanbanBoard(board), ok
}

func (s *MemoryStore) GetKanbanBoardSnapshot(projectRef string) (KanbanBoardSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.getKanbanProjectLocked(projectRef)
	if !ok {
		return KanbanBoardSnapshot{}, false
	}
	boardID, ok := s.kanbanDefaultBoardByProject[project.ID]
	if !ok {
		return KanbanBoardSnapshot{}, false
	}
	board, ok := s.kanbanBoards[boardID]
	if !ok {
		return KanbanBoardSnapshot{}, false
	}
	snapshot := KanbanBoardSnapshot{Project: project, Board: cloneKanbanBoard(board)}
	for _, ticket := range s.kanbanTickets {
		if ticket.ProjectID == project.ID {
			snapshot.Tickets = append(snapshot.Tickets, cloneKanbanTicket(ticket))
		}
	}
	for _, comment := range s.kanbanComments {
		if comment.ProjectID == project.ID {
			snapshot.Comments = append(snapshot.Comments, cloneKanbanComment(comment))
		}
	}
	for _, link := range s.kanbanLinks {
		if link.ProjectID == project.ID {
			snapshot.Links = append(snapshot.Links, cloneKanbanLink(link))
		}
	}
	for _, ref := range s.kanbanSourceRefs {
		if ref.ProjectID == project.ID {
			snapshot.SourceRefs = append(snapshot.SourceRefs, cloneKanbanSourceRef(ref))
		}
	}
	for _, event := range s.kanbanEvents {
		if event.ProjectID == project.ID {
			snapshot.Events = append(snapshot.Events, cloneKanbanEvent(event))
		}
	}
	sortKanbanTickets(snapshot.Tickets)
	sortKanbanEvents(snapshot.Events)
	return snapshot, true
}

func (s *MemoryStore) ListKanbanTickets(projectID string) []KanbanTicket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []KanbanTicket{}
	for _, item := range s.kanbanTickets {
		if projectID == "" || item.ProjectID == projectID {
			out = append(out, cloneKanbanTicket(item))
		}
	}
	sortKanbanTickets(out)
	return out
}

func (s *MemoryStore) GetKanbanTicket(ticketID string) (KanbanTicket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.kanbanTickets[strings.TrimSpace(ticketID)]
	return cloneKanbanTicket(item), ok
}

func (s *MemoryStore) CreateKanbanTicket(input KanbanTicketCreateInput, now time.Time) (KanbanTicket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.getKanbanProjectLocked(firstNonEmpty(input.ProjectID, input.ProjectSlug))
	if !ok {
		return KanbanTicket{}, errors.New("kanban project is required")
	}
	if project.State != "active" {
		return KanbanTicket{}, errors.New("kanban project is not active")
	}
	boardID := strings.TrimSpace(input.BoardID)
	if boardID == "" {
		boardID = s.kanbanDefaultBoardByProject[project.ID]
	}
	board, ok := s.kanbanBoards[boardID]
	if !ok || board.ProjectID != project.ID {
		return KanbanTicket{}, errors.New("kanban board not found for project")
	}
	for _, ref := range input.SourceRefs {
		if KanbanSlackSourceRefComplete(ref) {
			if existing, ok, err := s.findKanbanTicketBySlackSourceRefLocked(ref); err != nil {
				return KanbanTicket{}, err
			} else if ok {
				return existing, nil
			}
		}
	}
	ticket, err := NewKanbanTicket(input, project, board, now)
	if err != nil {
		return KanbanTicket{}, err
	}
	s.kanbanTickets[ticket.ID] = ticket
	s.kanbanEvents = append(s.kanbanEvents, KanbanEvent(project.ID, ticket.ID, "ticket.created", input.Actor, map[string]any{"title": ticket.Title, "status": string(ticket.Status)}, now))
	for _, refInput := range input.SourceRefs {
		_ = s.addKanbanSourceRefLocked(project.ID, ticket.ID, refInput, now)
	}
	return cloneKanbanTicket(ticket), nil
}

func (s *MemoryStore) UpdateKanbanTicket(ticketID string, input KanbanTicketUpdateInput, now time.Time) (KanbanTicket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket, ok := s.kanbanTickets[strings.TrimSpace(ticketID)]
	if !ok {
		return KanbanTicket{}, errors.New("kanban ticket not found")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	payload := map[string]any{}
	if input.Title != nil {
		ticket.Title = strings.TrimSpace(*input.Title)
		if ticket.Title == "" {
			return KanbanTicket{}, errors.New("kanban ticket title is required")
		}
		payload["title"] = ticket.Title
	}
	if input.Description != nil {
		ticket.Description = strings.TrimSpace(*input.Description)
		payload["description_changed"] = true
	}
	if input.Priority != nil {
		ticket.Priority = strings.TrimSpace(*input.Priority)
		payload["priority"] = ticket.Priority
	}
	if input.Assignee != nil {
		ticket.Assignee = strings.TrimSpace(*input.Assignee)
		payload["assignee"] = ticket.Assignee
	}
	if input.Metadata != nil {
		ticket.Metadata = CloneJSONMap(input.Metadata)
		payload["metadata_changed"] = true
	}
	if input.Status != nil {
		next := *input.Status
		if !KanbanStatusTransitionAllowed(ticket.Status, next) {
			return KanbanTicket{}, fmt.Errorf("invalid kanban status transition %s -> %s", ticket.Status, next)
		}
		payload["old_status"] = string(ticket.Status)
		payload["status"] = string(next)
		if next != ticket.Status {
			ticket.Status = next
			ticket.CompletedAt = nil
			ticket.ArchivedAt = nil
			if next == KanbanStatusDone {
				completed := now
				ticket.CompletedAt = &completed
			}
			if next == KanbanStatusArchived {
				archived := now
				ticket.ArchivedAt = &archived
			}
		}
	}
	ticket.UpdatedAt = now
	s.kanbanTickets[ticket.ID] = ticket
	s.kanbanEvents = append(s.kanbanEvents, KanbanEvent(ticket.ProjectID, ticket.ID, "ticket.updated", input.Actor, payload, now))
	return cloneKanbanTicket(ticket), nil
}

func (s *MemoryStore) AddKanbanTicketComment(ticketID string, input KanbanTicketCommentInput, now time.Time) (KanbanTicketComment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket, ok := s.kanbanTickets[strings.TrimSpace(ticketID)]
	if !ok {
		return KanbanTicketComment{}, errors.New("kanban ticket not found")
	}
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return KanbanTicketComment{}, errors.New("kanban comment body is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	actor := KanbanActorOrDefault(input.Actor, "api")
	comment := KanbanTicketComment{
		ID:            "kcmt_" + uuid.NewString(),
		ProjectID:     ticket.ProjectID,
		TicketID:      ticket.ID,
		Body:          body,
		ActorType:     actor.Type,
		ActorID:       actor.ID,
		ActorDisplay:  actor.Display,
		SourceSurface: actor.Surface,
		Metadata:      CloneJSONMap(input.Metadata),
		CreatedAt:     now,
	}
	s.kanbanComments[comment.ID] = comment
	s.kanbanEvents = append(s.kanbanEvents, KanbanEvent(ticket.ProjectID, ticket.ID, "comment.created", input.Actor, map[string]any{"comment_id": comment.ID}, now))
	return cloneKanbanComment(comment), nil
}

func (s *MemoryStore) AddKanbanTicketLink(input KanbanTicketLinkInput, now time.Time) (KanbanTicketLink, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	from, ok := s.kanbanTickets[strings.TrimSpace(input.FromTicketID)]
	if !ok {
		return KanbanTicketLink{}, errors.New("from ticket not found")
	}
	to, ok := s.kanbanTickets[strings.TrimSpace(input.ToTicketID)]
	if !ok {
		return KanbanTicketLink{}, errors.New("to ticket not found")
	}
	if from.ProjectID != to.ProjectID {
		return KanbanTicketLink{}, errors.New("cross-project ticket links are not supported")
	}
	if from.ID == to.ID {
		return KanbanTicketLink{}, errors.New("self ticket links are not supported")
	}
	linkType := strings.TrimSpace(input.LinkType)
	if linkType == "" {
		linkType = "related"
	}
	for _, existing := range s.kanbanLinks {
		if existing.ProjectID == from.ProjectID && existing.FromTicketID == from.ID && existing.ToTicketID == to.ID && existing.LinkType == linkType {
			return cloneKanbanLink(existing), nil
		}
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	link := KanbanTicketLink{
		ID:           "klnk_" + uuid.NewString(),
		ProjectID:    from.ProjectID,
		FromTicketID: from.ID,
		ToTicketID:   to.ID,
		LinkType:     linkType,
		CreatedBy:    firstNonEmpty(input.CreatedBy, input.Actor.ID, "unknown"),
		Metadata:     CloneJSONMap(input.Metadata),
		CreatedAt:    now,
	}
	s.kanbanLinks[link.ID] = link
	s.kanbanEvents = append(s.kanbanEvents, KanbanEvent(from.ProjectID, from.ID, "link.created", input.Actor, map[string]any{"link_id": link.ID, "to_ticket_id": to.ID, "link_type": linkType}, now))
	return cloneKanbanLink(link), nil
}

func (s *MemoryStore) SetKanbanSlackProjectRoute(input KanbanProjectSlackRouteInput, now time.Time) (KanbanProjectSlackRoute, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.kanbanProjects[strings.TrimSpace(input.ProjectID)]
	if !ok {
		return KanbanProjectSlackRoute{}, errors.New("kanban project not found")
	}
	channelID := strings.TrimSpace(input.ChannelID)
	if channelID == "" {
		return KanbanProjectSlackRoute{}, errors.New("slack channel_id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	key := KanbanSlackProjectRouteKey(input.TeamID, channelID, input.ThreadTS)
	route := KanbanProjectSlackRoute{
		ID:        "ksroute_" + uuid.NewString(),
		ProjectID: project.ID,
		TeamID:    strings.TrimSpace(input.TeamID),
		ChannelID: channelID,
		ThreadTS:  strings.TrimSpace(input.ThreadTS),
		CreatedAt: now,
		UpdatedAt: now,
	}
	for id, existing := range s.kanbanSlackRoutes {
		if KanbanSlackProjectRouteKey(existing.TeamID, existing.ChannelID, existing.ThreadTS) == key {
			route.ID = id
			route.CreatedAt = existing.CreatedAt
			break
		}
	}
	s.kanbanSlackRoutes[route.ID] = route
	s.kanbanSlackProjectRoutes[key] = project.ID
	s.kanbanEvents = append(s.kanbanEvents, KanbanEvent(project.ID, "", "project.slack_route_set", input.Actor, map[string]any{"channel_id": route.ChannelID, "thread_ts": route.ThreadTS}, now))
	return cloneKanbanSlackRoute(route), nil
}

func (s *MemoryStore) ListKanbanSlackProjectRoutes(projectID string) []KanbanProjectSlackRoute {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []KanbanProjectSlackRoute{}
	for _, item := range s.kanbanSlackRoutes {
		if strings.TrimSpace(projectID) == "" || item.ProjectID == strings.TrimSpace(projectID) {
			out = append(out, cloneKanbanSlackRoute(item))
		}
	}
	return out
}

func (s *MemoryStore) ListKanbanEvents(projectID string, afterID string, limit int) []KanbanTicketEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	started := afterID == ""
	found := afterID == ""
	out := []KanbanTicketEvent{}
	for _, event := range s.kanbanEvents {
		if strings.TrimSpace(projectID) != "" && event.ProjectID != projectID {
			continue
		}
		if !started {
			if event.ID == afterID {
				started = true
				found = true
			}
			continue
		}
		out = append(out, cloneKanbanEvent(event))
		if len(out) >= limit {
			break
		}
	}
	if !found && afterID != "" {
		return []KanbanTicketEvent{}
	}
	return out
}

func (s *MemoryStore) KanbanEventExists(projectID string, eventID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return false
	}
	projectID = strings.TrimSpace(projectID)
	for _, event := range s.kanbanEvents {
		if event.ID == eventID && (projectID == "" || event.ProjectID == projectID) {
			return true
		}
	}
	return false
}

func (s *MemoryStore) LatestKanbanEventID(projectID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	projectID = strings.TrimSpace(projectID)
	for i := len(s.kanbanEvents) - 1; i >= 0; i-- {
		event := s.kanbanEvents[i]
		if projectID == "" || event.ProjectID == projectID {
			return event.ID
		}
	}
	return ""
}

func (s *MemoryStore) ResolveKanbanSlackProject(teamID string, channelID string, threadTS string) (KanbanProject, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	teamID = strings.TrimSpace(teamID)
	channelID = strings.TrimSpace(channelID)
	threadTS = strings.TrimSpace(threadTS)
	if channelID == "" {
		return KanbanProject{}, false
	}
	if project, ok := s.resolveKanbanSlackProjectByKeyLocked(KanbanSlackProjectRouteKey(teamID, channelID, threadTS)); ok {
		return project, true
	}
	if teamID == "" && threadTS != "" {
		if project, ok, ambiguous := s.resolveUniqueKanbanSlackProjectRouteLocked(channelID, threadTS); ambiguous {
			return KanbanProject{}, false
		} else if ok {
			return project, true
		}
	}
	if project, ok := s.resolveKanbanSlackProjectByKeyLocked(KanbanSlackProjectRouteKey(teamID, channelID, "")); ok {
		return project, true
	}
	if teamID != "" {
		return KanbanProject{}, false
	}
	if project, ok, ambiguous := s.resolveUniqueKanbanSlackProjectRouteLocked(channelID, ""); ambiguous {
		return KanbanProject{}, false
	} else if ok {
		return project, true
	}
	return KanbanProject{}, false
}

func (s *MemoryStore) resolveKanbanSlackProjectByKeyLocked(key string) (KanbanProject, bool) {
	if projectID, ok := s.kanbanSlackProjectRoutes[key]; ok {
		project, ok := s.kanbanProjects[projectID]
		if ok && project.State == "active" {
			return cloneKanbanProject(project), true
		}
	}
	return KanbanProject{}, false
}

func (s *MemoryStore) resolveUniqueKanbanSlackProjectRouteLocked(channelID string, threadTS string) (KanbanProject, bool, bool) {
	channelID = strings.TrimSpace(channelID)
	threadTS = strings.TrimSpace(threadTS)
	var found KanbanProject
	for _, route := range s.kanbanSlackRoutes {
		if route.ChannelID != channelID || route.ThreadTS != threadTS {
			continue
		}
		project, ok := s.kanbanProjects[route.ProjectID]
		if !ok || project.State != "active" {
			continue
		}
		if found.ID != "" && found.ID != project.ID {
			return KanbanProject{}, false, true
		}
		found = project
	}
	if found.ID == "" {
		return KanbanProject{}, false, false
	}
	return cloneKanbanProject(found), true, false
}

func (s *MemoryStore) findKanbanTicketBySlackSourceRefLocked(ref KanbanTicketSourceRefInput) (KanbanTicket, bool, error) {
	if !KanbanSlackSourceRefComplete(ref) {
		return KanbanTicket{}, false, nil
	}
	if ticket, ok, err := s.findKanbanTicketBySlackSourceRefKeyLocked(KanbanSlackSourceRefKey(ref)); err != nil || ok {
		return ticket, ok, err
	}
	if strings.TrimSpace(ref.TeamID) != "" {
		withoutTeam := ref
		withoutTeam.TeamID = ""
		return s.findKanbanTicketBySlackSourceRefKeyLocked(KanbanSlackSourceRefKey(withoutTeam))
	}
	var found KanbanTicket
	identityKey := KanbanSlackSourceRefIdentityKey(ref)
	for _, existingRef := range s.kanbanSourceRefs {
		if strings.TrimSpace(existingRef.SourceType) != "slack" {
			continue
		}
		candidate := KanbanTicketSourceRefInput{
			SourceType: "slack",
			ActionKind: existingRef.ActionKind,
			ChannelID:  existingRef.ChannelID,
			ThreadTS:   existingRef.ThreadTS,
			MessageTS:  existingRef.MessageTS,
		}
		if KanbanSlackSourceRefIdentityKey(candidate) != identityKey {
			continue
		}
		ticket, ok := s.kanbanTickets[existingRef.TicketID]
		if !ok {
			return KanbanTicket{}, false, errors.New("kanban slack source ref ticket is missing")
		}
		if found.ID != "" && found.ID != ticket.ID {
			return KanbanTicket{}, false, errors.New("ambiguous kanban slack source ref team")
		}
		found = ticket
	}
	if found.ID == "" {
		return KanbanTicket{}, false, nil
	}
	return cloneKanbanTicket(found), true, nil
}

func (s *MemoryStore) findKanbanTicketBySlackSourceRefKeyLocked(key string) (KanbanTicket, bool, error) {
	refID, ok := s.kanbanSlackSourceRefByKey[key]
	if !ok {
		return KanbanTicket{}, false, nil
	}
	existingRef, ok := s.kanbanSourceRefs[refID]
	if !ok {
		return KanbanTicket{}, false, errors.New("kanban slack source ref index is stale")
	}
	existing, ok := s.kanbanTickets[existingRef.TicketID]
	if !ok {
		return KanbanTicket{}, false, errors.New("kanban slack source ref ticket is missing")
	}
	return cloneKanbanTicket(existing), true, nil
}

func (s *MemoryStore) addKanbanSourceRefLocked(projectID string, ticketID string, input KanbanTicketSourceRefInput, now time.Time) error {
	sourceType := strings.TrimSpace(input.SourceType)
	if sourceType == "" {
		return nil
	}
	actionKind := kanbanSourceRefActionKind(input.ActionKind)
	ref := KanbanTicketSourceRef{
		ID:             "ksrc_" + uuid.NewString(),
		ProjectID:      projectID,
		TicketID:       ticketID,
		SourceType:     sourceType,
		ActionKind:     actionKind,
		TeamID:         strings.TrimSpace(input.TeamID),
		ChannelID:      strings.TrimSpace(input.ChannelID),
		ThreadTS:       strings.TrimSpace(input.ThreadTS),
		MessageTS:      strings.TrimSpace(input.MessageTS),
		Permalink:      strings.TrimSpace(input.Permalink),
		ConversationID: strings.TrimSpace(input.ConversationID),
		TraceID:        strings.TrimSpace(input.TraceID),
		WorkflowID:     strings.TrimSpace(input.WorkflowID),
		ProposalID:     strings.TrimSpace(input.ProposalID),
		Metadata:       CloneJSONMap(input.Metadata),
		CreatedAt:      now,
	}
	s.kanbanSourceRefs[ref.ID] = ref
	if KanbanSlackSourceRefComplete(input) {
		key := KanbanSlackSourceRefKey(input)
		s.kanbanSlackSourceRefByKey[key] = ref.ID
	}
	return nil
}
