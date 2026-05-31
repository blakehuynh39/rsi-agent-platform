package control

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func executeKanbanNativeToolAction(ctx context.Context, repo storepkg.Repository, claims nativeToolClaims, input nativeToolActionRequest) (any, string, string, string, map[string]any, int, error) {
	_ = ctx
	kanban := storepkg.KanbanStore(repo)
	switch input.Operation {
	case "list_projects":
		projects := kanban.ListKanbanProjects()
		routes := []storepkg.KanbanProjectSlackRoute{}
		if boolArg(input.Arguments, "include_routes", false) {
			routes = kanban.ListKanbanSlackProjectRoutes("")
		}
		output := map[string]any{"ok": true, "projects": projects}
		if boolArg(input.Arguments, "include_routes", false) {
			output["routes"] = routes
		}
		return output, fmt.Sprintf("listed %d Kanban project(s)", len(projects)), "kanban:projects", "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	case "create_project":
		project, err := kanban.CreateKanbanProject(storepkg.KanbanProjectCreateInput{
			Slug:        stringArg(input.Arguments, "slug"),
			Name:        stringArg(input.Arguments, "name"),
			Description: firstNonEmpty(stringArg(input.Arguments, "description"), stringArg(input.Arguments, "summary")),
			Metadata:    mapArg(input.Arguments, "metadata"),
			Actor:       nativeKanbanActor(claims),
		}, time.Now().UTC())
		alreadyExists := false
		if err != nil {
			if errors.Is(err, storepkg.ErrKanbanProjectSlugExists) {
				if existing, ok := kanban.GetKanbanProject(firstNonEmpty(stringArg(input.Arguments, "slug"), stringArg(input.Arguments, "name"))); ok {
					project = existing
					alreadyExists = true
					err = nil
				}
			}
			if err != nil {
				return map[string]any{"ok": false, "error": err.Error()}, "kanban project creation failed", "", "", map[string]any{"status": "not_attempted", "reason": "validation_failed"}, http.StatusBadRequest, err
			}
		}
		output := map[string]any{"ok": true, "project": project, "already_exists": alreadyExists}
		sourceRef := "kanban:" + project.ID
		summary := "created Kanban project " + project.Slug
		if alreadyExists {
			summary = "loaded existing Kanban project " + project.Slug
		}
		return output, summary, sourceRef, "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	case "list_project_routes":
		projectID := ""
		projectRef := firstNonEmpty(stringArg(input.Arguments, "project_id"), stringArg(input.Arguments, "project_slug"))
		if projectRef != "" {
			project, ok := kanban.GetKanbanProject(projectRef)
			if !ok {
				return map[string]any{"ok": false, "error": "kanban project not found"}, "kanban project route listing failed", "", "", map[string]any{"status": "not_attempted", "reason": "project_not_found"}, http.StatusBadRequest, fmt.Errorf("kanban project %s was not found", projectRef)
			}
			projectID = project.ID
		}
		routes := kanban.ListKanbanSlackProjectRoutes(projectID)
		routes = nativeKanbanFilterProjectRoutes(routes, input, claims, projectRef == "")
		return map[string]any{"ok": true, "routes": routes}, fmt.Sprintf("listed %d Kanban project route(s)", len(routes)), "kanban:routes", "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	case "set_project_slack_route":
		projectRef := firstNonEmpty(stringArg(input.Arguments, "project_id"), stringArg(input.Arguments, "project_slug"), input.TargetRef)
		project, ok := kanban.GetKanbanProject(projectRef)
		if !ok {
			return map[string]any{"ok": false, "error": "kanban project not found"}, "kanban project route failed", "", "", map[string]any{"status": "not_attempted", "reason": "project_not_found"}, http.StatusBadRequest, fmt.Errorf("kanban project %s was not found", projectRef)
		}
		route, status, err := nativeKanbanSetProjectSlackRoute(kanban, claims, input, project.ID)
		if err != nil {
			return map[string]any{"ok": false, "project": project, "error": err.Error()}, "kanban project route failed", "kanban:" + project.ID, "", map[string]any{"status": "not_attempted", "reason": "validation_failed"}, status, err
		}
		return map[string]any{"ok": true, "project": project, "route": route}, "bound Slack route to Kanban project " + project.Slug, "kanban:" + project.ID, "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	case "list_tickets":
		project, status, err := resolveKanbanProjectForNativeTool(kanban, claims, input, true)
		if err != nil {
			return nativeKanbanProjectSelectionOutput(input, claims), "kanban project selection required", "", "", map[string]any{"status": "not_attempted", "reason": "needs_project_selection"}, status, err
		}
		tickets := kanban.ListKanbanTickets(project.ID)
		return map[string]any{"ok": true, "project": project, "tickets": tickets}, fmt.Sprintf("listed %d Kanban ticket(s)", len(tickets)), "kanban:" + project.ID, "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	case "create_ticket":
		project, status, err := resolveKanbanProjectForNativeTool(kanban, claims, input, true)
		if err != nil {
			return nativeKanbanProjectSelectionOutput(input, claims), "kanban project selection required", "", "", map[string]any{"status": "not_attempted", "reason": "needs_project_selection"}, status, err
		}
		ticket, err := kanban.CreateKanbanTicket(storepkg.KanbanTicketCreateInput{
			ProjectID:   project.ID,
			Title:       stringArg(input.Arguments, "title"),
			Description: firstNonEmpty(stringArg(input.Arguments, "description"), stringArg(input.Arguments, "body")),
			Priority:    stringArg(input.Arguments, "priority"),
			Assignee:    stringArg(input.Arguments, "assignee"),
			CreatedBy:   claims.Actor,
			Metadata:    mapArg(input.Arguments, "metadata"),
			Actor:       nativeKanbanActor(claims),
			SourceRefs:  nativeKanbanSourceRefs(claims, input, "create_ticket"),
		}, time.Now().UTC())
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error()}, "kanban ticket creation failed", "", "", map[string]any{"status": "not_attempted", "reason": "validation_failed"}, http.StatusBadRequest, err
		}
		return map[string]any{"ok": true, "project": project, "ticket": ticket}, "created Kanban ticket " + ticket.ID, "kanban:" + ticket.ID, "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	case "update_ticket":
		ticketID := firstNonEmpty(stringArg(input.Arguments, "ticket_id"), input.TargetRef)
		if ticketID == "" {
			return nil, "kanban ticket update failed", "", "", map[string]any{"status": "not_attempted", "reason": "missing_ticket_id"}, http.StatusBadRequest, errors.New("update_ticket requires ticket_id")
		}
		patch := storepkg.KanbanTicketUpdateInput{Actor: nativeKanbanActor(claims), Metadata: mapArg(input.Arguments, "metadata")}
		if value, ok := stringArgPresent(input.Arguments, "title"); ok {
			patch.Title = &value
		}
		if value, ok := stringArgPresent(input.Arguments, "description"); ok {
			patch.Description = &value
		}
		if value, ok := stringArgPresent(input.Arguments, "status"); ok {
			status := storepkg.KanbanTicketStatus(value)
			patch.Status = &status
		}
		if value, ok := stringArgPresent(input.Arguments, "priority"); ok {
			patch.Priority = &value
		}
		if value, ok := stringArgPresent(input.Arguments, "assignee"); ok {
			patch.Assignee = &value
		}
		ticket, err := kanban.UpdateKanbanTicket(ticketID, patch, time.Now().UTC())
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error()}, "kanban ticket update failed", "kanban:" + ticketID, "", map[string]any{"status": "not_attempted", "reason": "validation_failed"}, http.StatusBadRequest, err
		}
		return map[string]any{"ok": true, "ticket": ticket}, "updated Kanban ticket " + ticket.ID, "kanban:" + ticket.ID, "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	case "comment_ticket":
		ticketID := firstNonEmpty(stringArg(input.Arguments, "ticket_id"), input.TargetRef)
		comment, err := kanban.AddKanbanTicketComment(ticketID, storepkg.KanbanTicketCommentInput{
			Body:     firstNonEmpty(stringArg(input.Arguments, "body"), stringArg(input.Arguments, "comment")),
			Metadata: mapArg(input.Arguments, "metadata"),
			Actor:    nativeKanbanActor(claims),
		}, time.Now().UTC())
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error()}, "kanban comment failed", "kanban:" + ticketID, "", map[string]any{"status": "not_attempted", "reason": "validation_failed"}, http.StatusBadRequest, err
		}
		return map[string]any{"ok": true, "comment": comment}, "commented on Kanban ticket " + ticketID, "kanban:" + ticketID, "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	case "link_ticket":
		link, err := kanban.AddKanbanTicketLink(storepkg.KanbanTicketLinkInput{
			FromTicketID: firstNonEmpty(stringArg(input.Arguments, "from_ticket_id"), stringArg(input.Arguments, "ticket_id"), input.TargetRef),
			ToTicketID:   stringArg(input.Arguments, "to_ticket_id"),
			LinkType:     firstNonEmpty(stringArg(input.Arguments, "link_type"), "related"),
			CreatedBy:    claims.Actor,
			Metadata:     mapArg(input.Arguments, "metadata"),
			Actor:        nativeKanbanActor(claims),
		}, time.Now().UTC())
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error()}, "kanban link failed", "", "", map[string]any{"status": "not_attempted", "reason": "validation_failed"}, http.StatusBadRequest, err
		}
		return map[string]any{"ok": true, "link": link}, "linked Kanban tickets", "kanban:" + link.FromTicketID, "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
	default:
		return nil, "", "", "", map[string]any{"status": "not_attempted", "reason": "unknown_operation"}, http.StatusBadRequest, fmt.Errorf("unknown kanban operation %s", input.Operation)
	}
}

func nativeKanbanFilterProjectRoutes(routes []storepkg.KanbanProjectSlackRoute, input nativeToolActionRequest, claims nativeToolClaims, includeSlackClaimContext bool) []storepkg.KanbanProjectSlackRoute {
	teamID := firstNonEmpty(stringArg(input.Arguments, "team_id"), stringArg(input.Arguments, "workspace_id"))
	channelID := stringArg(input.Arguments, "channel_id")
	threadTS := stringArg(input.Arguments, "thread_ts")
	if includeSlackClaimContext {
		channelID = firstNonEmpty(channelID, claims.SlackChannelID)
		threadTS = firstNonEmpty(threadTS, claims.SlackThreadTS)
	}
	teamID = strings.TrimSpace(teamID)
	channelID = strings.TrimSpace(channelID)
	threadTS = strings.TrimSpace(threadTS)
	if teamID == "" && channelID == "" && threadTS == "" {
		return routes
	}
	out := make([]storepkg.KanbanProjectSlackRoute, 0, len(routes))
	for _, route := range routes {
		if teamID != "" && route.TeamID != teamID {
			continue
		}
		if channelID != "" && route.ChannelID != channelID {
			continue
		}
		if threadTS != "" && route.ThreadTS != "" && route.ThreadTS != threadTS {
			continue
		}
		out = append(out, route)
	}
	return out
}

func nativeKanbanSetProjectSlackRoute(kanban storepkg.KanbanStore, claims nativeToolClaims, input nativeToolActionRequest, projectID string) (storepkg.KanbanProjectSlackRoute, int, error) {
	channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), claims.SlackChannelID)
	threadTS := firstNonEmpty(stringArg(input.Arguments, "thread_ts"), claims.SlackThreadTS)
	route, err := kanban.SetKanbanSlackProjectRoute(storepkg.KanbanProjectSlackRouteInput{
		ProjectID: strings.TrimSpace(projectID),
		TeamID:    firstNonEmpty(stringArg(input.Arguments, "team_id"), stringArg(input.Arguments, "workspace_id")),
		ChannelID: channelID,
		ThreadTS:  threadTS,
		Actor:     nativeKanbanActor(claims),
	}, time.Now().UTC())
	if err != nil {
		return storepkg.KanbanProjectSlackRoute{}, http.StatusBadRequest, err
	}
	return route, http.StatusOK, nil
}

func resolveKanbanProjectForNativeTool(kanban storepkg.KanbanStore, claims nativeToolClaims, input nativeToolActionRequest, allowSlackDefault bool) (storepkg.KanbanProject, int, error) {
	projectRef := firstNonEmpty(stringArg(input.Arguments, "project_id"), stringArg(input.Arguments, "project_slug"))
	if projectRef != "" {
		if project, ok := kanban.GetKanbanProject(projectRef); ok {
			return project, http.StatusOK, nil
		}
		return storepkg.KanbanProject{}, http.StatusBadRequest, fmt.Errorf("kanban project %s was not found", projectRef)
	}
	if allowSlackDefault {
		teamID := firstNonEmpty(stringArg(input.Arguments, "team_id"), stringArg(input.Arguments, "workspace_id"))
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), claims.SlackChannelID)
		threadTS := firstNonEmpty(stringArg(input.Arguments, "thread_ts"), claims.SlackThreadTS)
		if channelID != "" {
			if project, ok := kanban.ResolveKanbanSlackProject(teamID, channelID, threadTS); ok {
				return project, http.StatusOK, nil
			}
		}
	}
	return storepkg.KanbanProject{}, http.StatusUnprocessableEntity, errors.New("needs_project_selection")
}

func nativeKanbanProjectSelectionOutput(input nativeToolActionRequest, claims nativeToolClaims) map[string]any {
	return map[string]any{
		"ok":      false,
		"reason":  "needs_project_selection",
		"message": "A Kanban project is required before creating or listing tickets.",
		"slack_context": map[string]any{
			"channel_id": firstNonEmpty(stringArg(input.Arguments, "channel_id"), claims.SlackChannelID),
			"thread_ts":  firstNonEmpty(stringArg(input.Arguments, "thread_ts"), claims.SlackThreadTS),
		},
	}
}

func nativeKanbanActor(claims nativeToolClaims) storepkg.KanbanActor {
	return storepkg.KanbanActor{Type: "agent", ID: firstNonEmpty(claims.Actor, "rsi"), Display: firstNonEmpty(claims.Actor, "RSI"), Surface: "native_tool"}
}

func nativeKanbanSourceRefs(claims nativeToolClaims, input nativeToolActionRequest, actionKind string) []storepkg.KanbanTicketSourceRefInput {
	channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), claims.SlackChannelID)
	messageTS := firstNonEmpty(stringArg(input.Arguments, "message_ts"), stringArg(input.Arguments, "ts"))
	threadTS := firstNonEmpty(stringArg(input.Arguments, "thread_ts"), claims.SlackThreadTS, messageTS)
	if channelID == "" || messageTS == "" {
		return nil
	}
	return []storepkg.KanbanTicketSourceRefInput{
		{
			SourceType:     "slack",
			ActionKind:     actionKind,
			TeamID:         firstNonEmpty(stringArg(input.Arguments, "team_id"), stringArg(input.Arguments, "workspace_id")),
			ChannelID:      channelID,
			ThreadTS:       threadTS,
			MessageTS:      messageTS,
			Permalink:      stringArg(input.Arguments, "permalink"),
			ConversationID: claims.ConversationID,
			TraceID:        claims.TraceID,
			WorkflowID:     claims.WorkflowID,
			Metadata:       mapArg(input.Arguments, "source_metadata"),
		},
	}
}

func stringArgPresent(args map[string]any, key string) (string, bool) {
	if args == nil {
		return "", false
	}
	value, ok := args[key]
	if !ok || value == nil {
		return "", false
	}
	trimmed := strings.TrimSpace(fmt.Sprint(value))
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}
