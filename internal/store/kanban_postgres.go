package store

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

const kanbanProjectColumns = `id, slug, name, description, state, metadata, created_at, updated_at`
const kanbanBoardColumns = `id, project_id, slug, name, is_default, metadata, created_at, updated_at`
const kanbanTicketColumns = `id, project_id, board_id, title, description, status, priority, assignee, created_by, metadata, created_at, updated_at, completed_at, archived_at`
const kanbanCommentColumns = `id, project_id, ticket_id, body, actor_type, actor_id, actor_display, source_surface, metadata, created_at`
const kanbanLinkColumns = `id, project_id, from_ticket_id, to_ticket_id, link_type, created_by, metadata, created_at`
const kanbanSourceRefColumns = `id, project_id, ticket_id, source_type, action_kind, team_id, channel_id, thread_ts, message_ts, permalink, conversation_id, trace_id, workflow_id, proposal_id, metadata, created_at`
const kanbanEventColumns = `id, project_id, ticket_id, event_type, actor_type, actor_id, actor_display, source_surface, payload, created_at`
const kanbanSlackRouteColumns = `id, project_id, team_id, channel_id, thread_ts, created_at, updated_at`

func (p *PostgresStore) ListKanbanProjects() []KanbanProject {
	rows, err := p.db.Query(`select ` + kanbanProjectColumns + ` from kanban_project order by state asc, name asc`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []KanbanProject{}
	for rows.Next() {
		item, err := scanKanbanProject(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) GetKanbanProject(ref string) (KanbanProject, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return KanbanProject{}, false
	}
	row := p.db.QueryRow(`select `+kanbanProjectColumns+` from kanban_project where id = $1 or slug = $2`, ref, normalizeKanbanSlug(ref))
	item, err := scanKanbanProject(row)
	return item, err == nil
}

func (p *PostgresStore) CreateKanbanProject(input KanbanProjectCreateInput, now time.Time) (KanbanProject, error) {
	project, board, err := NewKanbanProject(input, now)
	if err != nil {
		return KanbanProject{}, err
	}
	err = p.withTx(func(tx *sql.Tx) error {
		if existing, ok := selectKanbanProjectTx(tx, project.Slug); ok {
			return fmt.Errorf("%w: %s", ErrKanbanProjectSlugExists, existing.Slug)
		}
		if _, err := tx.Exec(`insert into kanban_project (`+kanbanProjectColumns+`) values ($1,$2,$3,$4,$5,$6::jsonb,$7,$8)`,
			project.ID, project.Slug, project.Name, project.Description, project.State, jsonString(project.Metadata), project.CreatedAt, project.UpdatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(`insert into kanban_board (`+kanbanBoardColumns+`) values ($1,$2,$3,$4,$5,$6::jsonb,$7,$8)`,
			board.ID, board.ProjectID, board.Slug, board.Name, board.IsDefault, jsonString(board.Metadata), board.CreatedAt, board.UpdatedAt); err != nil {
			return err
		}
		return insertKanbanEventTx(tx, KanbanEvent(project.ID, "", "project.created", input.Actor, map[string]any{"slug": project.Slug, "name": project.Name}, now))
	})
	return cloneKanbanProject(project), err
}

func (p *PostgresStore) UpdateKanbanProject(projectID string, input KanbanProjectUpdateInput, now time.Time) (KanbanProject, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	project, ok := p.GetKanbanProject(projectID)
	if !ok {
		return KanbanProject{}, errors.New("kanban project not found")
	}
	if input.Name != nil {
		project.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		project.Description = strings.TrimSpace(*input.Description)
	}
	if input.State != nil {
		project.State = strings.TrimSpace(*input.State)
		if project.State != "active" && project.State != "archived" {
			return KanbanProject{}, errors.New("invalid kanban project state")
		}
	}
	if input.Metadata != nil {
		project.Metadata = CloneJSONMap(input.Metadata)
	}
	project.UpdatedAt = now
	row := p.db.QueryRow(`update kanban_project set name=$2, description=$3, state=$4, metadata=$5::jsonb, updated_at=$6 where id=$1 returning `+kanbanProjectColumns,
		project.ID, project.Name, project.Description, project.State, jsonString(project.Metadata), project.UpdatedAt)
	out, err := scanKanbanProject(row)
	if err != nil {
		return KanbanProject{}, err
	}
	_ = p.insertKanbanEvent(KanbanEvent(project.ID, "", "project.updated", input.Actor, map[string]any{"state": project.State}, now))
	return out, nil
}

func (p *PostgresStore) GetKanbanDefaultBoard(projectID string) (KanbanBoard, bool) {
	row := p.db.QueryRow(`select `+kanbanBoardColumns+` from kanban_board where project_id = $1 and is_default = true`, strings.TrimSpace(projectID))
	item, err := scanKanbanBoard(row)
	return item, err == nil
}

func (p *PostgresStore) GetKanbanBoardSnapshot(projectRef string) (KanbanBoardSnapshot, bool) {
	project, ok := p.GetKanbanProject(projectRef)
	if !ok {
		return KanbanBoardSnapshot{}, false
	}
	board, ok := p.GetKanbanDefaultBoard(project.ID)
	if !ok {
		return KanbanBoardSnapshot{}, false
	}
	latestEventID := p.LatestKanbanEventID(project.ID)
	events := []KanbanTicketEvent{}
	if latestEventID != "" {
		row := p.db.QueryRow(`select `+kanbanEventColumns+` from kanban_ticket_event where id = $1`, latestEventID)
		if event, err := scanKanbanEvent(row); err == nil {
			events = append(events, event)
		}
	}
	return KanbanBoardSnapshot{
		Project:    project,
		Board:      board,
		Tickets:    p.ListKanbanTickets(project.ID),
		Comments:   p.listKanbanComments(project.ID),
		Links:      p.listKanbanLinks(project.ID),
		SourceRefs: p.listKanbanSourceRefs(project.ID),
		Events:     events,
	}, true
}

func (p *PostgresStore) ListKanbanTickets(projectID string) []KanbanTicket {
	query := `select ` + kanbanTicketColumns + ` from kanban_ticket`
	args := []any{}
	if strings.TrimSpace(projectID) != "" {
		query += ` where project_id = $1`
		args = append(args, strings.TrimSpace(projectID))
	}
	query += ` order by updated_at desc, id asc`
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []KanbanTicket{}
	for rows.Next() {
		item, err := scanKanbanTicket(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) GetKanbanTicket(ticketID string) (KanbanTicket, bool) {
	row := p.db.QueryRow(`select `+kanbanTicketColumns+` from kanban_ticket where id = $1`, strings.TrimSpace(ticketID))
	item, err := scanKanbanTicket(row)
	return item, err == nil
}

func (p *PostgresStore) CreateKanbanTicket(input KanbanTicketCreateInput, now time.Time) (KanbanTicket, error) {
	project, ok := p.GetKanbanProject(firstNonEmpty(input.ProjectID, input.ProjectSlug))
	if !ok {
		return KanbanTicket{}, errors.New("kanban project is required")
	}
	if project.State != "active" {
		return KanbanTicket{}, errors.New("kanban project is not active")
	}
	boardID := strings.TrimSpace(input.BoardID)
	if boardID == "" {
		board, ok := p.GetKanbanDefaultBoard(project.ID)
		if !ok {
			return KanbanTicket{}, errors.New("kanban default board not found")
		}
		boardID = board.ID
	}
	board, ok := p.getKanbanBoard(boardID)
	if !ok || board.ProjectID != project.ID {
		return KanbanTicket{}, errors.New("kanban board not found for project")
	}
	ticket, err := NewKanbanTicket(input, project, board, now)
	if err != nil {
		return KanbanTicket{}, err
	}
	var existingTicket *KanbanTicket
	err = p.withTx(func(tx *sql.Tx) error {
		if err := lockKanbanSlackSourceRefIdentitiesTx(tx, input.SourceRefs); err != nil {
			return err
		}
		for _, ref := range input.SourceRefs {
			if !KanbanSlackSourceRefComplete(ref) {
				continue
			}
			existing, ok, err := getKanbanTicketBySlackSourceRefTx(tx, ref)
			if err != nil {
				return err
			}
			if ok {
				existingTicket = &existing
				return nil
			}
		}
		if _, err := tx.Exec(`insert into kanban_ticket (`+kanbanTicketColumns+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11,$12,$13,$14)`,
			ticket.ID, ticket.ProjectID, ticket.BoardID, ticket.Title, ticket.Description, string(ticket.Status), ticket.Priority, ticket.Assignee, ticket.CreatedBy, jsonString(ticket.Metadata), ticket.CreatedAt, ticket.UpdatedAt, nullTime(ticket.CompletedAt), nullTime(ticket.ArchivedAt)); err != nil {
			return err
		}
		if err := insertKanbanEventTx(tx, KanbanEvent(project.ID, ticket.ID, "ticket.created", input.Actor, map[string]any{"title": ticket.Title, "status": string(ticket.Status)}, now)); err != nil {
			return err
		}
		for _, ref := range input.SourceRefs {
			if err := insertKanbanSourceRefTx(tx, ticket.ProjectID, ticket.ID, ref, now); err != nil {
				return err
			}
		}
		return nil
	})
	if existingTicket != nil {
		return *existingTicket, nil
	}
	if err != nil && isKanbanSlackSourceRefUniqueError(err) {
		for _, ref := range input.SourceRefs {
			if !KanbanSlackSourceRefComplete(ref) {
				continue
			}
			existing, ok, lookupErr := p.getKanbanTicketBySlackSourceRef(ref)
			if lookupErr != nil {
				return KanbanTicket{}, lookupErr
			}
			if ok {
				return existing, nil
			}
		}
	}
	return cloneKanbanTicket(ticket), err
}

func (p *PostgresStore) UpdateKanbanTicket(ticketID string, input KanbanTicketUpdateInput, now time.Time) (KanbanTicket, error) {
	ticket, ok := p.GetKanbanTicket(ticketID)
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
			return KanbanTicket{}, errors.New("invalid kanban status transition")
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
	row := p.db.QueryRow(`update kanban_ticket set title=$2, description=$3, status=$4, priority=$5, assignee=$6, metadata=$7::jsonb, updated_at=$8, completed_at=$9, archived_at=$10 where id=$1 returning `+kanbanTicketColumns,
		ticket.ID, ticket.Title, ticket.Description, string(ticket.Status), ticket.Priority, ticket.Assignee, jsonString(ticket.Metadata), ticket.UpdatedAt, nullTime(ticket.CompletedAt), nullTime(ticket.ArchivedAt))
	out, err := scanKanbanTicket(row)
	if err != nil {
		return KanbanTicket{}, err
	}
	_ = p.insertKanbanEvent(KanbanEvent(out.ProjectID, out.ID, "ticket.updated", input.Actor, payload, now))
	return out, nil
}

func (p *PostgresStore) AddKanbanTicketComment(ticketID string, input KanbanTicketCommentInput, now time.Time) (KanbanTicketComment, error) {
	ticket, ok := p.GetKanbanTicket(ticketID)
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
	err := p.withTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`insert into kanban_ticket_comment (`+kanbanCommentColumns+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10)`,
			comment.ID, comment.ProjectID, comment.TicketID, comment.Body, comment.ActorType, comment.ActorID, comment.ActorDisplay, comment.SourceSurface, jsonString(comment.Metadata), comment.CreatedAt); err != nil {
			return err
		}
		return insertKanbanEventTx(tx, KanbanEvent(ticket.ProjectID, ticket.ID, "comment.created", input.Actor, map[string]any{"comment_id": comment.ID}, now))
	})
	return cloneKanbanComment(comment), err
}

func (p *PostgresStore) AddKanbanTicketLink(input KanbanTicketLinkInput, now time.Time) (KanbanTicketLink, error) {
	from, ok := p.GetKanbanTicket(input.FromTicketID)
	if !ok {
		return KanbanTicketLink{}, errors.New("from ticket not found")
	}
	to, ok := p.GetKanbanTicket(input.ToTicketID)
	if !ok {
		return KanbanTicketLink{}, errors.New("to ticket not found")
	}
	if from.ProjectID != to.ProjectID {
		return KanbanTicketLink{}, errors.New("cross-project ticket links are not supported")
	}
	if from.ID == to.ID {
		return KanbanTicketLink{}, errors.New("self ticket links are not supported")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	linkType := strings.TrimSpace(input.LinkType)
	if linkType == "" {
		linkType = "related"
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
	var inserted bool
	err := p.withTx(func(tx *sql.Tx) error {
		result, err := tx.Exec(`insert into kanban_ticket_link (`+kanbanLinkColumns+`) values ($1,$2,$3,$4,$5,$6,$7::jsonb,$8) on conflict do nothing`,
			link.ID, link.ProjectID, link.FromTicketID, link.ToTicketID, link.LinkType, link.CreatedBy, jsonString(link.Metadata), link.CreatedAt)
		if err != nil {
			return err
		}
		rowsAffected, _ := result.RowsAffected()
		inserted = rowsAffected > 0
		if inserted {
			return insertKanbanEventTx(tx, KanbanEvent(from.ProjectID, from.ID, "link.created", input.Actor, map[string]any{"link_id": link.ID, "to_ticket_id": to.ID, "link_type": linkType}, now))
		}
		return nil
	})
	if err != nil {
		return KanbanTicketLink{}, err
	}
	if !inserted {
		row := p.db.QueryRow(`select `+kanbanLinkColumns+` from kanban_ticket_link where project_id = $1 and from_ticket_id = $2 and to_ticket_id = $3 and link_type = $4`,
			from.ProjectID, from.ID, to.ID, linkType)
		existingLink, err := scanKanbanLink(row)
		if err != nil {
			return KanbanTicketLink{}, err
		}
		return existingLink, nil
	}
	return cloneKanbanLink(link), nil
}

func (p *PostgresStore) SetKanbanSlackProjectRoute(input KanbanProjectSlackRouteInput, now time.Time) (KanbanProjectSlackRoute, error) {
	project, ok := p.GetKanbanProject(input.ProjectID)
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
	route := KanbanProjectSlackRoute{
		ID:        "ksroute_" + uuid.NewString(),
		ProjectID: project.ID,
		TeamID:    strings.TrimSpace(input.TeamID),
		ChannelID: channelID,
		ThreadTS:  strings.TrimSpace(input.ThreadTS),
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := p.withTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(`insert into kanban_project_slack_route (`+kanbanSlackRouteColumns+`) values ($1,$2,$3,$4,$5,$6,$7)
on conflict (team_id, channel_id, thread_ts) do update set project_id = excluded.project_id, updated_at = excluded.updated_at
returning `+kanbanSlackRouteColumns,
			route.ID, route.ProjectID, route.TeamID, route.ChannelID, route.ThreadTS, route.CreatedAt, route.UpdatedAt)
		out, err := scanKanbanSlackRoute(row)
		if err != nil {
			return err
		}
		route = out
		return insertKanbanEventTx(tx, KanbanEvent(project.ID, "", "project.slack_route_set", input.Actor, map[string]any{"channel_id": route.ChannelID, "thread_ts": route.ThreadTS}, now))
	})
	return cloneKanbanSlackRoute(route), err
}

func (p *PostgresStore) ListKanbanSlackProjectRoutes(projectID string) []KanbanProjectSlackRoute {
	rows, err := p.db.Query(`select `+kanbanSlackRouteColumns+` from kanban_project_slack_route where ($1 = '' or project_id = $1) order by channel_id asc, thread_ts asc`, strings.TrimSpace(projectID))
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []KanbanProjectSlackRoute{}
	for rows.Next() {
		item, err := scanKanbanSlackRoute(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) ListKanbanEvents(projectID string, afterID string, limit int) []KanbanTicketEvent {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	projectID = strings.TrimSpace(projectID)
	afterID = strings.TrimSpace(afterID)
	args := []any{projectID}
	query := `select ` + kanbanEventColumns + ` from kanban_ticket_event where ($1 = '' or project_id = $1)`
	if afterID != "" {
		var cursorCreatedAt time.Time
		var cursorID string
		err := p.db.QueryRow(`select created_at, id from kanban_ticket_event where id = $1 and ($2 = '' or project_id = $2)`, afterID, projectID).Scan(&cursorCreatedAt, &cursorID)
		if err != nil {
			return []KanbanTicketEvent{}
		}
		query += ` and (created_at, id) > ($2, $3)`
		args = append(args, cursorCreatedAt, cursorID)
	}
	query += ` order by created_at asc, id asc limit $` + strconv.Itoa(len(args)+1)
	args = append(args, limit)
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []KanbanTicketEvent{}
	for rows.Next() {
		item, err := scanKanbanEvent(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) KanbanEventExists(projectID string, eventID string) bool {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return false
	}
	var exists bool
	_ = p.db.QueryRow(`select exists(select 1 from kanban_ticket_event where id = $1 and ($2 = '' or project_id = $2))`, eventID, strings.TrimSpace(projectID)).Scan(&exists)
	return exists
}

func (p *PostgresStore) LatestKanbanEventID(projectID string) string {
	row := p.db.QueryRow(`select id from kanban_ticket_event where ($1 = '' or project_id = $1) order by created_at desc, id desc limit 1`, strings.TrimSpace(projectID))
	var id string
	if err := row.Scan(&id); err != nil {
		return ""
	}
	return id
}

func (p *PostgresStore) ResolveKanbanSlackProject(teamID string, channelID string, threadTS string) (KanbanProject, bool) {
	teamID = strings.TrimSpace(teamID)
	channelID = strings.TrimSpace(channelID)
	threadTS = strings.TrimSpace(threadTS)
	if channelID == "" {
		return KanbanProject{}, false
	}
	if project, ok := p.resolveKanbanSlackProjectByRoute(teamID, channelID, threadTS); ok {
		return project, true
	}
	if teamID == "" && threadTS != "" {
		if project, ok, ambiguous := p.resolveUniqueKanbanSlackProjectRoute(channelID, threadTS); ambiguous {
			return KanbanProject{}, false
		} else if ok {
			return project, true
		}
	}
	if project, ok := p.resolveKanbanSlackProjectByRoute(teamID, channelID, ""); ok {
		return project, true
	}
	if teamID != "" {
		return KanbanProject{}, false
	}
	if project, ok, ambiguous := p.resolveUniqueKanbanSlackProjectRoute(channelID, ""); ambiguous {
		return KanbanProject{}, false
	} else if ok {
		return project, true
	}
	return KanbanProject{}, false
}

func (p *PostgresStore) resolveKanbanSlackProjectByRoute(teamID string, channelID string, threadTS string) (KanbanProject, bool) {
	row := p.db.QueryRow(`select p.`+strings.ReplaceAll(kanbanProjectColumns, ", ", ", p.")+` from kanban_project_slack_route r join kanban_project p on p.id = r.project_id where r.team_id = $1 and r.channel_id = $2 and r.thread_ts = $3 and p.state = 'active'`, strings.TrimSpace(teamID), strings.TrimSpace(channelID), strings.TrimSpace(threadTS))
	item, err := scanKanbanProject(row)
	return item, err == nil
}

func (p *PostgresStore) resolveUniqueKanbanSlackProjectRoute(channelID string, threadTS string) (KanbanProject, bool, bool) {
	rows, err := p.db.Query(`select distinct on (p.id) p.`+strings.ReplaceAll(kanbanProjectColumns, ", ", ", p.")+` from kanban_project_slack_route r join kanban_project p on p.id = r.project_id where r.channel_id = $1 and r.thread_ts = $2 and p.state = 'active' order by p.id limit 2`, strings.TrimSpace(channelID), strings.TrimSpace(threadTS))
	if err != nil {
		return KanbanProject{}, false, false
	}
	defer rows.Close()
	items := []KanbanProject{}
	for rows.Next() {
		item, err := scanKanbanProject(rows)
		if err == nil {
			items = append(items, item)
		}
	}
	if len(items) == 1 {
		return items[0], true, false
	}
	if len(items) > 1 {
		return KanbanProject{}, false, true
	}
	return KanbanProject{}, false, false
}

func (p *PostgresStore) getKanbanBoard(boardID string) (KanbanBoard, bool) {
	row := p.db.QueryRow(`select `+kanbanBoardColumns+` from kanban_board where id = $1`, strings.TrimSpace(boardID))
	item, err := scanKanbanBoard(row)
	return item, err == nil
}

func (p *PostgresStore) getKanbanTicketBySlackSourceRef(ref KanbanTicketSourceRefInput) (KanbanTicket, bool, error) {
	return getKanbanTicketBySlackSourceRefFrom(p.db, ref)
}

func getKanbanTicketBySlackSourceRefTx(tx *sql.Tx, ref KanbanTicketSourceRefInput) (KanbanTicket, bool, error) {
	return getKanbanTicketBySlackSourceRefFrom(tx, ref)
}

func getKanbanTicketBySlackSourceRefFrom(q sqlReader, ref KanbanTicketSourceRefInput) (KanbanTicket, bool, error) {
	if !KanbanSlackSourceRefComplete(ref) {
		return KanbanTicket{}, false, nil
	}
	if item, ok := getKanbanTicketBySlackSourceRefExact(q, strings.TrimSpace(ref.TeamID), ref); ok {
		return item, true, nil
	}
	if strings.TrimSpace(ref.TeamID) != "" {
		if item, ok := getKanbanTicketBySlackSourceRefExact(q, "", ref); ok {
			return item, true, nil
		}
		return KanbanTicket{}, false, nil
	}
	item, ok, ambiguous := getUniqueKanbanTicketBySlackSourceRefIdentity(q, ref)
	if ambiguous {
		return KanbanTicket{}, false, errors.New("ambiguous kanban slack source ref team")
	}
	return item, ok, nil
}

func getKanbanTicketBySlackSourceRefExact(q sqlReader, teamID string, ref KanbanTicketSourceRefInput) (KanbanTicket, bool) {
	row := q.QueryRow(`select t.`+strings.ReplaceAll(kanbanTicketColumns, ", ", ", t.")+` from kanban_ticket_source_ref r join kanban_ticket t on t.id = r.ticket_id where r.source_type = 'slack' and r.team_id = $1 and r.channel_id = $2 and r.thread_ts = $3 and r.message_ts = $4 and r.action_kind = $5`,
		strings.TrimSpace(teamID), strings.TrimSpace(ref.ChannelID), strings.TrimSpace(ref.ThreadTS), strings.TrimSpace(ref.MessageTS), kanbanSourceRefActionKind(ref.ActionKind))
	item, err := scanKanbanTicket(row)
	return item, err == nil
}

func getUniqueKanbanTicketBySlackSourceRefIdentity(q sqlReader, ref KanbanTicketSourceRefInput) (KanbanTicket, bool, bool) {
	rows, err := q.Query(`select distinct on (t.id) t.`+strings.ReplaceAll(kanbanTicketColumns, ", ", ", t.")+` from kanban_ticket_source_ref r join kanban_ticket t on t.id = r.ticket_id where r.source_type = 'slack' and r.channel_id = $1 and r.thread_ts = $2 and r.message_ts = $3 and r.action_kind = $4 order by t.id limit 2`,
		strings.TrimSpace(ref.ChannelID), strings.TrimSpace(ref.ThreadTS), strings.TrimSpace(ref.MessageTS), kanbanSourceRefActionKind(ref.ActionKind))
	if err != nil {
		return KanbanTicket{}, false, false
	}
	defer rows.Close()
	items := []KanbanTicket{}
	for rows.Next() {
		item, err := scanKanbanTicket(rows)
		if err == nil {
			items = append(items, item)
		}
	}
	if len(items) == 1 {
		return items[0], true, false
	}
	if len(items) > 1 {
		return KanbanTicket{}, false, true
	}
	return KanbanTicket{}, false, false
}

func lockKanbanSlackSourceRefIdentitiesTx(tx *sql.Tx, refs []KanbanTicketSourceRefInput) error {
	keys := []string{}
	seen := map[string]bool{}
	for _, ref := range refs {
		if !KanbanSlackSourceRefComplete(ref) {
			continue
		}
		key := KanbanSlackSourceRefIdentityKey(ref)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := tx.Exec(`select pg_advisory_xact_lock(hashtext($1)::bigint)`, key); err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStore) listKanbanComments(projectID string) []KanbanTicketComment {
	rows, err := p.db.Query(`select `+kanbanCommentColumns+` from kanban_ticket_comment where project_id = $1 order by created_at asc`, strings.TrimSpace(projectID))
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []KanbanTicketComment{}
	for rows.Next() {
		item, err := scanKanbanComment(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) listKanbanLinks(projectID string) []KanbanTicketLink {
	rows, err := p.db.Query(`select `+kanbanLinkColumns+` from kanban_ticket_link where project_id = $1 order by created_at asc`, strings.TrimSpace(projectID))
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []KanbanTicketLink{}
	for rows.Next() {
		item, err := scanKanbanLink(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) listKanbanSourceRefs(projectID string) []KanbanTicketSourceRef {
	rows, err := p.db.Query(`select `+kanbanSourceRefColumns+` from kanban_ticket_source_ref where project_id = $1 order by created_at asc`, strings.TrimSpace(projectID))
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []KanbanTicketSourceRef{}
	for rows.Next() {
		item, err := scanKanbanSourceRef(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out
}

func (p *PostgresStore) insertKanbanEvent(event KanbanTicketEvent) error {
	return p.withTx(func(tx *sql.Tx) error { return insertKanbanEventTx(tx, event) })
}

func selectKanbanProjectTx(tx *sql.Tx, ref string) (KanbanProject, bool) {
	row := tx.QueryRow(`select `+kanbanProjectColumns+` from kanban_project where id = $1 or slug = $2`, strings.TrimSpace(ref), normalizeKanbanSlug(ref))
	item, err := scanKanbanProject(row)
	return item, err == nil
}

func insertKanbanEventTx(tx *sql.Tx, item KanbanTicketEvent) error {
	_, err := tx.Exec(`insert into kanban_ticket_event (`+kanbanEventColumns+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10)`,
		item.ID, item.ProjectID, nullString(item.TicketID), item.EventType, item.ActorType, item.ActorID, item.ActorDisplay, item.SourceSurface, jsonString(item.Payload), item.CreatedAt)
	return err
}

func insertKanbanSourceRefTx(tx *sql.Tx, projectID string, ticketID string, input KanbanTicketSourceRefInput, now time.Time) error {
	sourceType := strings.TrimSpace(input.SourceType)
	if sourceType == "" {
		return nil
	}
	actionKind := kanbanSourceRefActionKind(input.ActionKind)
	_, err := tx.Exec(`insert into kanban_ticket_source_ref (`+kanbanSourceRefColumns+`) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::jsonb,$16)`,
		"ksrc_"+uuid.NewString(), projectID, ticketID, sourceType, actionKind, strings.TrimSpace(input.TeamID), strings.TrimSpace(input.ChannelID), strings.TrimSpace(input.ThreadTS), strings.TrimSpace(input.MessageTS), strings.TrimSpace(input.Permalink), strings.TrimSpace(input.ConversationID), strings.TrimSpace(input.TraceID), strings.TrimSpace(input.WorkflowID), strings.TrimSpace(input.ProposalID), jsonString(input.Metadata), now)
	return err
}

func isKanbanSlackSourceRefUniqueError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.HasPrefix(pgErr.ConstraintName, "kanban_ticket_source_ref_slack_idempotency")
	}
	return strings.Contains(err.Error(), "kanban_ticket_source_ref_slack_idempotency")
}

type kanbanScanner interface{ Scan(dest ...any) error }

func scanKanbanProject(row kanbanScanner) (KanbanProject, error) {
	var item KanbanProject
	var raw []byte
	if err := row.Scan(&item.ID, &item.Slug, &item.Name, &item.Description, &item.State, &raw, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return KanbanProject{}, err
	}
	item.Metadata = decodeJSON(raw, map[string]any{})
	return cloneKanbanProject(item), nil
}

func scanKanbanBoard(row kanbanScanner) (KanbanBoard, error) {
	var item KanbanBoard
	var raw []byte
	if err := row.Scan(&item.ID, &item.ProjectID, &item.Slug, &item.Name, &item.IsDefault, &raw, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return KanbanBoard{}, err
	}
	item.Metadata = decodeJSON(raw, map[string]any{})
	return cloneKanbanBoard(item), nil
}

func scanKanbanTicket(row kanbanScanner) (KanbanTicket, error) {
	var item KanbanTicket
	var raw []byte
	var completedAt, archivedAt sql.NullTime
	if err := row.Scan(&item.ID, &item.ProjectID, &item.BoardID, &item.Title, &item.Description, &item.Status, &item.Priority, &item.Assignee, &item.CreatedBy, &raw, &item.CreatedAt, &item.UpdatedAt, &completedAt, &archivedAt); err != nil {
		return KanbanTicket{}, err
	}
	item.Metadata = decodeJSON(raw, map[string]any{})
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}
	if archivedAt.Valid {
		item.ArchivedAt = &archivedAt.Time
	}
	return cloneKanbanTicket(item), nil
}

func scanKanbanComment(row kanbanScanner) (KanbanTicketComment, error) {
	var item KanbanTicketComment
	var raw []byte
	if err := row.Scan(&item.ID, &item.ProjectID, &item.TicketID, &item.Body, &item.ActorType, &item.ActorID, &item.ActorDisplay, &item.SourceSurface, &raw, &item.CreatedAt); err != nil {
		return KanbanTicketComment{}, err
	}
	item.Metadata = decodeJSON(raw, map[string]any{})
	return cloneKanbanComment(item), nil
}

func scanKanbanLink(row kanbanScanner) (KanbanTicketLink, error) {
	var item KanbanTicketLink
	var raw []byte
	if err := row.Scan(&item.ID, &item.ProjectID, &item.FromTicketID, &item.ToTicketID, &item.LinkType, &item.CreatedBy, &raw, &item.CreatedAt); err != nil {
		return KanbanTicketLink{}, err
	}
	item.Metadata = decodeJSON(raw, map[string]any{})
	return cloneKanbanLink(item), nil
}

func scanKanbanSourceRef(row kanbanScanner) (KanbanTicketSourceRef, error) {
	var item KanbanTicketSourceRef
	var raw []byte
	if err := row.Scan(&item.ID, &item.ProjectID, &item.TicketID, &item.SourceType, &item.ActionKind, &item.TeamID, &item.ChannelID, &item.ThreadTS, &item.MessageTS, &item.Permalink, &item.ConversationID, &item.TraceID, &item.WorkflowID, &item.ProposalID, &raw, &item.CreatedAt); err != nil {
		return KanbanTicketSourceRef{}, err
	}
	item.Metadata = decodeJSON(raw, map[string]any{})
	return cloneKanbanSourceRef(item), nil
}

func scanKanbanEvent(row kanbanScanner) (KanbanTicketEvent, error) {
	var item KanbanTicketEvent
	var ticketID sql.NullString
	var raw []byte
	if err := row.Scan(&item.ID, &item.ProjectID, &ticketID, &item.EventType, &item.ActorType, &item.ActorID, &item.ActorDisplay, &item.SourceSurface, &raw, &item.CreatedAt); err != nil {
		return KanbanTicketEvent{}, err
	}
	item.TicketID = ticketID.String
	item.Payload = decodeJSON(raw, map[string]any{})
	return cloneKanbanEvent(item), nil
}

func scanKanbanSlackRoute(row kanbanScanner) (KanbanProjectSlackRoute, error) {
	var item KanbanProjectSlackRoute
	if err := row.Scan(&item.ID, &item.ProjectID, &item.TeamID, &item.ChannelID, &item.ThreadTS, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return KanbanProjectSlackRoute{}, err
	}
	return cloneKanbanSlackRoute(item), nil
}
