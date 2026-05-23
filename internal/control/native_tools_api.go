package control

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const nativeToolsAudience = "rsi-native-tools"

var nativeToolWriteOps = map[string]map[string]bool{
	"slack": {
		"message_post": true, "message_update": true, "message_delete": true, "report_post": true,
		"reaction_add": true, "reaction_remove": true, "file_upload": true,
		"channel_create": true, "channel_rename": true, "channel_archive": true, "channel_invite": true,
	},
	"notion": {
		"page_create": true, "page_update": true, "page_archive": true,
		"blocks_append": true, "block_update": true, "block_delete": true, "comment_create": true,
	},
	"knowledge": {
		"wiki_edit_propose": true, "wiki_edit_apply": true,
	},
	"kanban": {
		"create_ticket": true, "update_ticket": true, "comment_ticket": true, "link_ticket": true,
	},
}

var nativeToolReadOps = map[string]map[string]bool{
	"slack": {
		"channels_list": true, "channel_info": true, "conversation_read": true, "user_lookup": true,
	},
	"notion": {
		"search": true, "page_get": true, "blocks_children": true, "database_get": true,
		"data_source_get": true, "data_source_query": true,
	},
	"knowledge": {
		"search": true, "document_get": true, "conversation_get": true, "messages_read": true,
		"wiki_search": true, "wiki_page_get": true, "wiki_index_get": true, "wiki_log_get": true,
		"source_status": true,
	},
	"sentry": {
		"projects_list": true, "issues_list": true, "issue_view": true,
		"issue_events": true, "releases_list": true,
	},
	"kanban": {
		"list_tickets": true,
	},
}

var nativeToolDestructiveOps = map[string]map[string]bool{
	"slack": {
		"message_delete": true, "channel_archive": true,
	},
	"notion": {
		"page_archive": true, "block_delete": true,
	},
}

type nativeToolClaims struct {
	Audience       string   `json:"aud"`
	IssuedAt       int64    `json:"iat"`
	ExpiresAt      int64    `json:"exp"`
	ExecutionID    string   `json:"execution_id"`
	OperationID    string   `json:"operation_id"`
	TraceID        string   `json:"trace_id"`
	WorkflowID     string   `json:"workflow_id"`
	ConversationID string   `json:"conversation_id"`
	Actor          string   `json:"actor"`
	Surfaces       []string `json:"surfaces"`
	SlackChannelID string   `json:"slack_channel_id,omitempty"`
	SlackThreadTS  string   `json:"slack_thread_ts,omitempty"`
	SlackScope     string   `json:"slack_delivery_scope,omitempty"`
}

type nativeToolActionRequest struct {
	Surface        string         `json:"surface"`
	Operation      string         `json:"operation"`
	TargetRef      string         `json:"target_ref,omitempty"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	Destructive    bool           `json:"destructive,omitempty"`
	ConfirmDestroy bool           `json:"confirm_destroy,omitempty"`
	Arguments      map[string]any `json:"arguments,omitempty"`
}

type nativeToolActionResponse struct {
	OK       bool                        `json:"ok"`
	Replayed bool                        `json:"replayed,omitempty"`
	Action   storepkg.ExternalToolAction `json:"action"`
	Output   any                         `json:"output,omitempty"`
	Error    string                      `json:"error,omitempty"`
}

func registerNativeToolRoutes(r chi.Router, cfg config.Config, store storepkg.Repository) {
	r.Post("/internal/native-tools/actions", func(w http.ResponseWriter, r *http.Request) {
		claims, err := authorizeNativeToolAction(cfg, r, "")
		if err != nil {
			app.WriteError(w, http.StatusUnauthorized, err)
			return
		}
		var input nativeToolActionRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			app.WriteError(w, http.StatusBadRequest, err)
			return
		}
		out, status, err := handleNativeToolAction(r.Context(), cfg, store, claims, input)
		if err != nil && out.Action.ID == "" {
			app.WriteError(w, status, err)
			return
		}
		app.WriteJSON(w, status, out)
	})

	r.Get("/internal/native-tools/actions/{actionID}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := authorizeNativeToolAction(cfg, r, ""); err != nil {
			app.WriteError(w, http.StatusUnauthorized, err)
			return
		}
		actionID := chi.URLParam(r, "actionID")
		action, ok := store.GetExternalToolAction(actionID)
		if !ok {
			app.WriteError(w, http.StatusNotFound, errors.New("native tool action not found"))
			return
		}
		app.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "action": action})
	})
}

func handleNativeToolAction(ctx context.Context, cfg config.Config, repo storepkg.Repository, claims nativeToolClaims, input nativeToolActionRequest) (nativeToolActionResponse, int, error) {
	input.Surface = strings.TrimSpace(input.Surface)
	input.Operation = strings.TrimSpace(input.Operation)
	input.TargetRef = strings.TrimSpace(firstNonEmpty(input.TargetRef, nativeToolTargetRef(input.Arguments)))
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.Reason = strings.TrimSpace(input.Reason)
	if input.Arguments == nil {
		input.Arguments = map[string]any{}
	}
	input = resolveNativeToolActionArguments(cfg, repo, input)
	if !nativeToolSurfaceAllowed(claims, input.Surface) {
		return nativeToolActionResponse{}, http.StatusForbidden, fmt.Errorf("native tool token is not scoped for surface %q", input.Surface)
	}
	if !nativeToolOperationKnown(input.Surface, input.Operation) {
		return nativeToolActionResponse{}, http.StatusBadRequest, fmt.Errorf("native tool operation %s.%s is not allowed", input.Surface, input.Operation)
	}
	isWrite := nativeToolWriteOps[input.Surface][input.Operation]
	isDestructive := input.Destructive || nativeToolDestructiveOps[input.Surface][input.Operation]
	if isWrite {
		if input.IdempotencyKey == "" {
			return nativeToolActionResponse{}, http.StatusBadRequest, errors.New("native tool writes require idempotency_key")
		}
		if input.Reason == "" {
			return nativeToolActionResponse{}, http.StatusBadRequest, errors.New("native tool writes require reason")
		}
	} else {
		if input.IdempotencyKey == "" {
			input.IdempotencyKey = nativeToolReadIdempotencyKey(claims, input)
		}
		if input.Reason == "" {
			input.Reason = "native read"
		}
	}
	requestHash, err := nativeToolRequestHash(input, isDestructive)
	if err != nil {
		return nativeToolActionResponse{}, http.StatusBadRequest, err
	}
	now := time.Now().UTC()
	action, upsertStatus, err := repo.UpsertExternalToolAction(storepkg.ExternalToolActionCreateInput{
		Surface:        input.Surface,
		Operation:      input.Operation,
		TargetRef:      input.TargetRef,
		IdempotencyKey: input.IdempotencyKey,
		RequestHash:    requestHash,
		Actor:          claims.Actor,
		Reason:         input.Reason,
		Destructive:    isDestructive,
		ExecutionID:    claims.ExecutionID,
		OperationID:    claims.OperationID,
		TraceID:        claims.TraceID,
		WorkflowID:     claims.WorkflowID,
		ConversationID: claims.ConversationID,
	}, now)
	if err != nil {
		return nativeToolActionResponse{}, http.StatusBadRequest, err
	}
	if upsertStatus == storepkg.ExternalToolActionUpsertConflict {
		return nativeToolActionResponse{OK: false, Replayed: true, Action: action, Error: "idempotency key already used with a different request hash"}, http.StatusConflict, nil
	}
	replaying := upsertStatus == storepkg.ExternalToolActionUpsertReplay
	resuming := replaying && nativeToolActionCanResume(input, action)
	if replaying && !resuming {
		return nativeToolActionResponse{OK: action.State == storepkg.ExternalToolActionStateSucceeded, Replayed: true, Action: action, Output: action.ResultPayload, Error: action.ErrorMessage}, http.StatusOK, nil
	}
	if resuming {
		args := storepkg.CloneJSONMap(input.Arguments)
		if args == nil {
			args = map[string]any{}
		}
		args["__previous_result_payload"] = action.ResultPayload
		input.Arguments = args
	}
	if validationErr, status := validateNativeToolActionPolicy(cfg, claims, input, isWrite, isDestructive); validationErr != nil {
		failed, updateErr := repo.UpdateExternalToolActionResult(action.ID, storepkg.ExternalToolActionResultUpdate{
			State:        storepkg.ExternalToolActionStateFailed,
			ErrorMessage: validationErr.Error(),
			MirrorEffect: map[string]any{"status": "not_attempted", "reason": "validation_failed"},
		}, now)
		if updateErr == nil {
			action = failed
		}
		return nativeToolActionResponse{OK: false, Action: action, Error: validationErr.Error()}, status, nil
	}
	output, responseSummary, sourceRef, wikiAuditID, mirrorEffect, execStatus, execErr := executeNativeToolAction(ctx, cfg, repo, claims, input)
	resultState := storepkg.ExternalToolActionStateSucceeded
	errorMessage := ""
	if execErr != nil {
		resultState = storepkg.ExternalToolActionStateFailed
		errorMessage = execErr.Error()
	}
	resultPayload := mapValue(output)
	if resultPayload == nil {
		if resuming {
			resultPayload = action.ResultPayload
		} else {
			resultPayload = map[string]any{}
		}
	}
	action, err = repo.UpdateExternalToolActionResult(action.ID, storepkg.ExternalToolActionResultUpdate{
		State:           resultState,
		ResponseSummary: responseSummary,
		ErrorMessage:    errorMessage,
		SourceRef:       sourceRef,
		WikiAuditID:     wikiAuditID,
		ResultPayload:   resultPayload,
		MirrorEffect:    mirrorEffect,
	}, time.Now().UTC())
	if err != nil {
		return nativeToolActionResponse{}, http.StatusInternalServerError, err
	}
	return nativeToolActionResponse{
		OK:       execErr == nil,
		Replayed: replaying,
		Action:   action,
		Output:   output,
		Error:    errorMessage,
	}, execStatus, execErr
}

func nativeToolActionCanResume(input nativeToolActionRequest, action storepkg.ExternalToolAction) bool {
	return input.Surface == "slack" &&
		input.Operation == "report_post" &&
		slackReportResultHasPostedMain(action.ResultPayload) &&
		(action.State == storepkg.ExternalToolActionStateFailed || slackReportResultHasUploadFailures(action.ResultPayload))
}

func executeNativeToolAction(ctx context.Context, cfg config.Config, repo storepkg.Repository, claims nativeToolClaims, input nativeToolActionRequest) (any, string, string, string, map[string]any, int, error) {
	if input.Surface == "slack" {
		return executeSlackNativeToolAction(ctx, cfg, repo, input)
	}
	if input.Surface == "notion" {
		return executeNotionNativeToolAction(ctx, cfg, repo, input)
	}
	if input.Surface == "sentry" {
		return executeSentryNativeToolAction(ctx, cfg, input)
	}
	if input.Surface == "kanban" {
		return executeKanbanNativeToolAction(ctx, repo, claims, input)
	}
	if input.Surface == "knowledge" {
		switch input.Operation {
		case "search":
			out, status, err := nativeKnowledgeSearch(ctx, cfg, repo, input)
			return out, fmt.Sprintf("knowledge search returned %d wiki result(s), %d Slack result(s), %d document result(s)", len(out.Wiki.Results), len(out.Slack.Results), len(out.Documents.Results)), "company_knowledge", "", map[string]any{"status": "not_applicable"}, status, err
		case "document_get":
			out, status, err := nativeKnowledgeDocumentGet(ctx, cfg, repo, input)
			return out, "loaded mirrored company knowledge document", "company_knowledge:document", "", map[string]any{"status": "not_applicable"}, status, err
		case "conversation_get":
			out, status, err := nativeKnowledgeConversationGet(ctx, cfg, repo, input)
			return out, fmt.Sprintf("loaded %d mirrored Slack message(s)", len(out.Messages)), out.SourceSessionKey, "", map[string]any{"status": "not_applicable"}, status, err
		case "messages_read":
			out, status, err := nativeKnowledgeMessagesRead(ctx, cfg, repo, input)
			return out, fmt.Sprintf("read %d mirrored Slack message(s)", len(out.Messages)), out.SourceSessionKey, "", map[string]any{"status": "not_applicable"}, status, err
		case "wiki_search":
			query := stringArg(input.Arguments, "query")
			limit := intArg(input.Arguments, "limit", 10)
			out, status, err := companyWikiSearch(ctx, repo, query, limit)
			return out, fmt.Sprintf("wiki search returned %d result(s)", len(out.Results)), "company_wiki", "", map[string]any{"status": "not_applicable"}, status, err
		case "wiki_page_get":
			ref := firstNonEmpty(stringArg(input.Arguments, "page_ref"), stringArg(input.Arguments, "slug"), input.TargetRef)
			out, status, err := companyWikiPageGet(ctx, repo, ref)
			return out, "loaded wiki page", "company_wiki:" + ref, "", map[string]any{"status": "not_applicable"}, status, err
		case "wiki_index_get":
			out, status, err := companyWikiIndexGet(ctx, cfg, repo)
			return out, "loaded wiki index", "company_wiki:index", "", map[string]any{"status": "not_applicable"}, status, err
		case "wiki_log_get":
			out, status, err := companyWikiLogGet(ctx, cfg, repo, intArg(input.Arguments, "limit", 0))
			return out, "loaded wiki log", "company_wiki:log", "", map[string]any{"status": "not_applicable"}, status, err
		case "source_status":
			sourceTypes := stringSliceArg(input.Arguments, "source_types")
			if len(sourceTypes) == 0 {
				sourceTypes = stringSliceArg(input.Arguments, "source_type")
			}
			out, status, err := sourceMirrorStatus(cfg, repo, sourceTypes, intArg(input.Arguments, "limit", 500), time.Duration(intArg(input.Arguments, "max_age_seconds", 0))*time.Second)
			return out, "loaded source mirror status", "source_mirror", "", map[string]any{"status": "not_applicable"}, status, err
		case "wiki_edit_propose":
			req := companyWikiEditProposeRequest{
				Actor:          claims.Actor,
				Reason:         input.Reason,
				IdempotencyKey: input.IdempotencyKey,
				Slug:           firstNonEmpty(stringArg(input.Arguments, "slug"), stringArg(input.Arguments, "page_ref"), input.TargetRef),
				Title:          stringArg(input.Arguments, "title"),
				Body:           firstNonEmpty(stringArg(input.Arguments, "body"), stringArg(input.Arguments, "content")),
				Metadata:       mapArg(input.Arguments, "metadata"),
			}
			out, status, err := companyWikiEditPropose(ctx, cfg, repo, req)
			wikiAuditID := out.Audit.ID
			return out, "recorded wiki edit proposal", "company_wiki:" + req.Slug, wikiAuditID, map[string]any{"status": "not_applicable", "wiki_audit_id": wikiAuditID}, status, err
		case "wiki_edit_apply":
			req := companyWikiEditApplyRequest{
				Actor:          claims.Actor,
				Reason:         input.Reason,
				IdempotencyKey: input.IdempotencyKey,
				Slug:           firstNonEmpty(stringArg(input.Arguments, "slug"), stringArg(input.Arguments, "page_ref"), input.TargetRef),
				Title:          stringArg(input.Arguments, "title"),
				Body:           firstNonEmpty(stringArg(input.Arguments, "body"), stringArg(input.Arguments, "content")),
				Metadata:       mapArg(input.Arguments, "metadata"),
			}
			out, status, err := companyWikiEditApply(ctx, cfg, repo, req)
			wikiAuditID := out.Audit.ID
			return out, "applied wiki edit", "company_wiki:" + req.Slug, wikiAuditID, map[string]any{"status": "not_applicable", "wiki_audit_id": wikiAuditID}, status, err
		}
	}
	message := fmt.Sprintf("native tool dispatcher for %s.%s is registered but source implementation is not enabled yet", input.Surface, input.Operation)
	return nil, "", "", "", map[string]any{"status": "not_attempted", "reason": "implementation_pending"}, http.StatusNotImplemented, errors.New(message)
}

type nativeKnowledgeSearchResponse struct {
	OK          bool                      `json:"ok"`
	Query       string                    `json:"query"`
	SourceTypes []string                  `json:"source_types,omitempty"`
	Wiki        companyWikiSearchResponse `json:"wiki,omitempty"`
	Slack       nativeHonchoMessages      `json:"slack,omitempty"`
	Documents   nativeHonchoDocuments     `json:"documents,omitempty"`
	Errors      []string                  `json:"errors,omitempty"`
}

type nativeHonchoDocuments struct {
	Source  string           `json:"source"`
	Results []map[string]any `json:"results,omitempty"`
	Limit   int              `json:"limit"`
}

type nativeHonchoMessages struct {
	Source  string                  `json:"source"`
	Results []clients.HonchoMessage `json:"results,omitempty"`
	Limit   int                     `json:"limit"`
}

type nativeKnowledgeConversationResponse struct {
	Source           string                  `json:"source"`
	WorkspaceID      string                  `json:"workspace_id"`
	ChannelID        string                  `json:"channel_id"`
	ThreadTS         string                  `json:"thread_ts,omitempty"`
	SourceSessionKey string                  `json:"source_session_key"`
	HonchoSessionID  string                  `json:"honcho_session_id"`
	Messages         []clients.HonchoMessage `json:"messages"`
	Page             int                     `json:"page"`
	Pages            int                     `json:"pages"`
	Total            int                     `json:"total"`
	OldestTS         string                  `json:"oldest_ts,omitempty"`
	LatestTS         string                  `json:"latest_ts,omitempty"`
}

func nativeKnowledgeSearch(ctx context.Context, cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest) (nativeKnowledgeSearchResponse, int, error) {
	query := stringArg(input.Arguments, "query")
	if strings.TrimSpace(query) == "" {
		return nativeKnowledgeSearchResponse{}, http.StatusBadRequest, errors.New("knowledge search requires query")
	}
	limit := intArg(input.Arguments, "limit", 10)
	sourceTypes := stringSliceArg(input.Arguments, "source_types")
	if len(sourceTypes) == 0 {
		sourceTypes = []string{"wiki", "slack_message", "notion_document"}
	}
	out := nativeKnowledgeSearchResponse{OK: true, Query: query, SourceTypes: sourceTypes}
	statusCode := http.StatusOK
	for _, sourceType := range sourceTypes {
		switch strings.TrimSpace(sourceType) {
		case "wiki", "company_wiki":
			wiki, status, err := companyWikiSearch(ctx, repo, query, limit)
			out.Wiki = wiki
			if err != nil {
				out.Errors = append(out.Errors, err.Error())
				statusCode = maxHTTPStatus(statusCode, status)
			}
		case "slack", "slack_message":
			results, status, err := nativeHonchoSlackSearch(ctx, cfg, query, stringArg(input.Arguments, "channel_id"), limit)
			out.Slack = results
			if err != nil {
				out.Errors = append(out.Errors, err.Error())
				statusCode = maxHTTPStatus(statusCode, status)
			}
		case "notion", "notion_document", "document":
			results, status, err := nativeHonchoDocumentSearch(ctx, cfg, query, limit)
			out.Documents = results
			if err != nil {
				out.Errors = append(out.Errors, err.Error())
				statusCode = maxHTTPStatus(statusCode, status)
			}
		default:
			out.Errors = append(out.Errors, "unsupported knowledge source_type "+sourceType)
			statusCode = maxHTTPStatus(statusCode, http.StatusBadRequest)
		}
	}
	out.OK = len(out.Errors) == 0
	if !out.OK {
		return out, statusCode, errors.New(strings.Join(out.Errors, "; "))
	}
	return out, http.StatusOK, nil
}

func nativeKnowledgeDocumentGet(ctx context.Context, cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest) (map[string]any, int, error) {
	_ = ctx
	documentID := firstNonEmpty(stringArg(input.Arguments, "document_id"), input.TargetRef)
	sourceRef := stringArg(input.Arguments, "source_ref")
	resolvedID := ""
	resolutionRef := ""
	if sourceRef != "" {
		resolutionRef = sourceRef
	} else {
		resolutionRef = documentID
	}
	if resolutionRef != "" {
		record, found, err := nativeSourceMirrorRecordForRef(repo, companyknowledge.NotionDocumentSourceType, resolutionRef)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		if found {
			resolvedID = strings.TrimSpace(record.HonchoObjectID)
			if resolvedID == "" {
				return nil, http.StatusConflict, fmt.Errorf("mirrored Notion source %s has no knowledge document id yet; retry after the mirror completes", resolutionRef)
			}
		}
	}
	if resolvedID != "" {
		documentID = resolvedID
	} else if sourceRef != "" {
		return nil, http.StatusBadRequest, fmt.Errorf("source_ref %s did not resolve to a mirrored Notion document; use rsi_notion.page_get or rsi_notion.blocks_children for raw Notion reads", sourceRef)
	} else if nativeLooksLikeRawNotionRef(documentID) {
		return nil, http.StatusBadRequest, fmt.Errorf("document_id %s looks like a raw Notion id, not a mirrored knowledge document id; use source_ref after search, or rsi_notion.page_get/rsi_notion.blocks_children for direct Notion reads", documentID)
	}
	if documentID == "" {
		return nil, http.StatusBadRequest, errors.New("document_get requires document_id or source_ref that resolves to a mirrored document")
	}
	honcho, workspaceID, err := nativeHonchoClient(cfg)
	if err != nil {
		return nil, http.StatusFailedDependency, err
	}
	page, err := honcho.ListConclusions(workspaceID, map[string]any{
		"AND": []any{
			map[string]any{"id": documentID},
			nativeNotionDocumentFilters(),
		},
	}, 1, 1)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	if len(page.Items) == 0 {
		return nil, http.StatusNotFound, fmt.Errorf("mirrored document %s was not found", documentID)
	}
	return map[string]any{
		"source":   "honcho_notion_documents",
		"document": page.Items[0],
		"lookup": map[string]any{
			"requested_document_id": firstNonEmpty(stringArg(input.Arguments, "document_id"), input.TargetRef),
			"requested_source_ref":  sourceRef,
			"resolved_document_id":  documentID,
		},
	}, http.StatusOK, nil
}

func nativeKnowledgeConversationGet(ctx context.Context, cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest) (nativeKnowledgeConversationResponse, int, error) {
	channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
	threadTS := stringArg(input.Arguments, "thread_ts")
	sourceSessionKey := stringArg(input.Arguments, "conversation_ref")
	if sourceSessionKey == "" {
		sourceSessionKey = stringArg(input.Arguments, "source_ref")
	}
	if sourceSessionKey != "" && !strings.HasPrefix(sourceSessionKey, "slack:") {
		sourceSessionKey = ""
	}
	if sourceSessionKey == "" {
		var err error
		sourceSessionKey, err = resolveSlackSourceSessionKey(ctx, cfg, repo, channelID, threadTS)
		if err != nil {
			return nativeKnowledgeConversationResponse{}, http.StatusBadRequest, err
		}
	}
	workspaceID, resolvedChannelID, resolvedThreadTS, err := parseSlackSourceSessionKey(sourceSessionKey)
	if err != nil {
		return nativeKnowledgeConversationResponse{}, http.StatusBadRequest, err
	}
	if !slackMirrorChannelAllowedByConfig(cfg, resolvedChannelID) {
		return nativeKnowledgeConversationResponse{}, http.StatusForbidden, fmt.Errorf("slack channel %s is not available in the mirrored Slack corpus", resolvedChannelID)
	}
	honcho, honchoWorkspaceID, err := nativeHonchoClient(cfg)
	if err != nil {
		return nativeKnowledgeConversationResponse{}, http.StatusFailedDependency, err
	}
	limit := intArg(input.Arguments, "limit", 50)
	pageNumber := intArg(input.Arguments, "page", 1)
	honchoSessionID := companyknowledge.HonchoCompatibleName("slack", sourceSessionKey)
	page, err := honcho.ListMessages(honchoWorkspaceID, honchoSessionID, limit, pageNumber, false)
	if err != nil {
		return nativeKnowledgeConversationResponse{}, http.StatusBadGateway, err
	}
	return nativeKnowledgeConversationResponse{
		Source:           "honcho_slack_corpus",
		WorkspaceID:      workspaceID,
		ChannelID:        resolvedChannelID,
		ThreadTS:         resolvedThreadTS,
		SourceSessionKey: sourceSessionKey,
		HonchoSessionID:  honchoSessionID,
		Messages:         page.Items,
		Page:             page.Page,
		Pages:            page.Pages,
		Total:            page.Total,
	}, http.StatusOK, nil
}

func nativeKnowledgeMessagesRead(ctx context.Context, cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest) (nativeKnowledgeConversationResponse, int, error) {
	threadTS := stringArg(input.Arguments, "thread_ts")
	oldestTS := firstNonEmpty(stringArg(input.Arguments, "oldest_ts"), stringArg(input.Arguments, "oldest"))
	latestTS := firstNonEmpty(stringArg(input.Arguments, "latest_ts"), stringArg(input.Arguments, "latest"))
	if threadTS == "" && oldestTS == "" && latestTS == "" {
		return nativeKnowledgeConversationResponse{}, http.StatusBadRequest, errors.New("channel-wide messages_read requires oldest_ts or latest_ts")
	}
	if oldestTS == "" && latestTS == "" {
		conversation, status, err := nativeKnowledgeConversationGet(ctx, cfg, repo, input)
		if err != nil {
			return conversation, status, err
		}
		conversation.OldestTS = oldestTS
		conversation.LatestTS = latestTS
		return conversation, http.StatusOK, nil
	}
	pageInput := input
	pageInput.Arguments = cloneNativeToolArguments(input.Arguments)
	pageSize := intArg(input.Arguments, "limit", 50)
	pageInput.Arguments["limit"] = pageSize
	var allMessages []clients.HonchoMessage
	var lastConversation nativeKnowledgeConversationResponse
	page := 1
	totalPages := 0
	for {
		pageInput.Arguments["page"] = page
		conversation, status, err := nativeKnowledgeConversationGet(ctx, cfg, repo, pageInput)
		if err != nil {
			if page == 1 {
				return conversation, status, err
			}
			break
		}
		lastConversation = conversation
		if page == 1 {
			totalPages = conversation.Pages
		}
		for _, message := range conversation.Messages {
			if slackTSInWindow(honchoMessageSlackTS(message), oldestTS, latestTS) {
				allMessages = append(allMessages, message)
			}
		}
		if page >= totalPages || len(conversation.Messages) == 0 {
			break
		}
		page++
	}
	lastConversation.Messages = allMessages
	lastConversation.Page = 1
	lastConversation.Pages = 1
	lastConversation.Total = len(allMessages)
	lastConversation.OldestTS = oldestTS
	lastConversation.LatestTS = latestTS
	return lastConversation, http.StatusOK, nil
}

func nativeHonchoSlackSearch(ctx context.Context, cfg config.Config, query string, channelID string, limit int) (nativeHonchoMessages, int, error) {
	_ = ctx
	if channelID != "" && !slackMirrorChannelAllowedByConfig(cfg, channelID) {
		return nativeHonchoMessages{}, http.StatusForbidden, fmt.Errorf("slack channel %s is not available in the mirrored Slack corpus", channelID)
	}
	honcho, workspaceID, err := nativeHonchoClient(cfg)
	if err != nil {
		return nativeHonchoMessages{}, http.StatusFailedDependency, err
	}
	results, err := honcho.SearchMessages(workspaceID, query, nativeSlackMessageFilters(cfg, channelID), limit)
	if err != nil {
		return nativeHonchoMessages{}, http.StatusBadGateway, err
	}
	filtered := make([]clients.HonchoMessage, 0, len(results))
	for _, item := range results {
		resultChannel := nativeStringFromAny(item.Metadata["channel_id"])
		if resultChannel == "" || slackMirrorChannelAllowedByConfig(cfg, resultChannel) {
			filtered = append(filtered, item)
		}
	}
	return nativeHonchoMessages{Source: "honcho_slack_corpus", Results: filtered, Limit: limit}, http.StatusOK, nil
}

func nativeHonchoDocumentSearch(ctx context.Context, cfg config.Config, query string, limit int) (nativeHonchoDocuments, int, error) {
	_ = ctx
	honcho, workspaceID, err := nativeHonchoClient(cfg)
	if err != nil {
		return nativeHonchoDocuments{}, http.StatusFailedDependency, err
	}
	results, err := honcho.QueryConclusions(workspaceID, query, nativeNotionDocumentFilters(), limit)
	if err != nil {
		return nativeHonchoDocuments{}, http.StatusBadGateway, err
	}
	return nativeHonchoDocuments{Source: "honcho_notion_documents", Results: results, Limit: limit}, http.StatusOK, nil
}

func nativeHonchoClient(cfg config.Config) (*clients.HonchoClient, string, error) {
	if strings.TrimSpace(cfg.HonchoBaseURL) == "" {
		return nil, "", errors.New("RSI_HONCHO_BASE_URL is required for native knowledge corpus tools")
	}
	workspaceID := companyknowledge.HonchoCompatibleName("workspace", firstNonEmpty(cfg.HonchoWorkspaceID, "rsi_company_knowledge"))
	return clients.NewHonchoClientWithAPIKey(cfg.HonchoBaseURL, cfg.HonchoAPIKey), workspaceID, nil
}

func nativeSourceMirrorRecordForRef(repo storepkg.Repository, sourceType string, ref string) (storepkg.SourceMirrorRecord, bool, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return storepkg.SourceMirrorRecord{}, false, nil
	}
	mirrorStore, ok := repo.(storepkg.SourceMirrorWriteStore)
	if !ok {
		return storepkg.SourceMirrorRecord{}, false, nil
	}
	candidates := []string{ref}
	if strings.HasPrefix(ref, sourceType+":") {
		candidates = append(candidates, strings.TrimPrefix(ref, sourceType+":"))
	}
	normalized := normalizeNotionID(ref)
	if normalized != "" {
		candidates = append(candidates,
			companyknowledge.NotionDocumentSourceKey("notion", normalized),
			companyknowledge.NotionObjectSourceKey("notion", companyknowledge.NotionObjectKindPage, normalized),
			companyknowledge.NotionObjectSourceKey("notion", companyknowledge.NotionObjectKindDatabase, normalized),
			companyknowledge.NotionObjectSourceKey("notion", companyknowledge.NotionObjectKindDataSource, normalized),
		)
	}
	for _, candidate := range uniqueNonEmpty(candidates) {
		record, found, err := mirrorStore.GetSourceMirrorRecord(sourceType, candidate)
		if err != nil {
			return storepkg.SourceMirrorRecord{}, false, err
		}
		if found {
			return record, true, nil
		}
	}
	return storepkg.SourceMirrorRecord{}, false, nil
}

func nativeLooksLikeRawNotionRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return false
	}
	if strings.HasPrefix(ref, "notion:") {
		return true
	}
	return normalizeStrictNotionID(ref) != ""
}

func nativeNotionBlockChildrenOutput(out clients.NotionListResponse[clients.NotionBlock]) map[string]any {
	results := make([]map[string]any, 0, len(out.Results))
	for _, block := range out.Results {
		item := map[string]any{
			"object":           block.Object,
			"id":               block.ID,
			"type":             block.Type,
			"has_children":     block.HasChildren,
			"created_time":     block.CreatedTime,
			"last_edited_time": block.LastEditedTime,
			"archived":         block.Archived,
			"in_trash":         block.InTrash,
		}
		if plainText := nativeNotionBlockPlainText(block); plainText != "" {
			item["plain_text"] = plainText
		}
		if markdown := strings.TrimSpace(notionBlockMarkdown(block, 0)); markdown != "" {
			item["markdown"] = markdown
		}
		if payload, ok := block.Raw[block.Type].(map[string]any); ok && len(payload) > 0 {
			item["type_payload"] = payload
		}
		results = append(results, item)
	}
	return map[string]any{
		"object":      out.Object,
		"results":     results,
		"next_cursor": out.NextCursor,
		"has_more":    out.HasMore,
	}
}

func nativeNotionBlockPlainText(block clients.NotionBlock) string {
	payload, ok := block.Raw[block.Type].(map[string]any)
	if !ok {
		return ""
	}
	if text := richTextPlainTextFromAny(payload["rich_text"]); text != "" {
		return text
	}
	if text := richTextPlainTextFromAny(payload["caption"]); text != "" {
		return text
	}
	for _, key := range []string{"title", "name"} {
		if text := strings.TrimSpace(fmt.Sprint(payload[key])); text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func cloneNativeToolArguments(args map[string]any) map[string]any {
	out := make(map[string]any, len(args))
	for key, value := range args {
		out[key] = value
	}
	return out
}

func nativeNotionDocumentFilters() map[string]any {
	return map[string]any{
		"observer_id": "notion_mirror",
		"observed_id": "story_company",
	}
}

func nativeSlackMessageFilters(cfg config.Config, channelID string) map[string]any {
	if channelID != "" {
		return map[string]any{
			"AND": []any{
				map[string]any{"metadata": map[string]any{"source": "slack"}},
				map[string]any{"metadata": map[string]any{"channel_id": channelID}},
			},
		}
	}
	if slackMirrorChannelDiscoveryMode(cfg) == "explicit" && len(cfg.SlackMirrorChannelAllowlist) > 0 {
		ors := make([]any, 0, len(cfg.SlackMirrorChannelAllowlist))
		for _, allowedChannel := range cfg.SlackMirrorChannelAllowlist {
			allowedChannel = strings.TrimSpace(allowedChannel)
			if allowedChannel != "" && slackMirrorChannelAllowedByConfig(cfg, allowedChannel) {
				ors = append(ors, map[string]any{"metadata": map[string]any{"channel_id": allowedChannel}})
			}
		}
		if len(ors) > 0 {
			return map[string]any{
				"AND": []any{
					map[string]any{"metadata": map[string]any{"source": "slack"}},
					map[string]any{"OR": ors},
				},
			}
		}
	}
	return map[string]any{"metadata": map[string]any{"source": "slack"}}
}

func resolveSlackSourceSessionKey(ctx context.Context, cfg config.Config, repo storepkg.Repository, channelID string, threadTS string) (string, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return "", errors.New("conversation_get requires channel_id or conversation_ref")
	}
	if !slackMirrorChannelAllowedByConfig(cfg, channelID) {
		return "", fmt.Errorf("slack channel %s is not available in the mirrored Slack corpus", channelID)
	}
	if record, found := sourceMirrorRecordForSlackConversation(repo, channelID, threadTS); found {
		return strings.TrimSpace(record.SourceSessionKey), nil
	}
	workspaceID, err := nativeSlackWorkspaceID(ctx, cfg, nil)
	if err != nil {
		return "", err
	}
	return companyknowledge.SlackSessionSourceKey(workspaceID, channelID, threadTS, strings.TrimSpace(threadTS) != ""), nil
}

func sourceMirrorRecordForSlackConversation(repo storepkg.Repository, channelID string, threadTS string) (storepkg.SourceMirrorRecord, bool) {
	statusStore, ok := repo.(storepkg.SourceMirrorStatusStore)
	if !ok {
		return storepkg.SourceMirrorRecord{}, false
	}
	records, err := statusStore.ListSourceMirrorRecords([]string{companyknowledge.SlackMessageSourceType}, 5000)
	if err != nil {
		return storepkg.SourceMirrorRecord{}, false
	}
	channelNeedle := ":" + strings.TrimSpace(channelID) + ":"
	for _, record := range records {
		sessionKey := strings.TrimSpace(record.SourceSessionKey)
		if !strings.Contains(sessionKey, channelNeedle) {
			continue
		}
		if threadTS != "" && !strings.HasSuffix(sessionKey, ":"+strings.TrimSpace(threadTS)) {
			continue
		}
		if threadTS == "" && !strings.HasSuffix(sessionKey, ":channel") {
			continue
		}
		return record, true
	}
	return storepkg.SourceMirrorRecord{}, false
}

func parseSlackSourceSessionKey(sourceSessionKey string) (string, string, string, error) {
	parts := strings.Split(strings.TrimSpace(sourceSessionKey), ":")
	if len(parts) < 4 || parts[0] != "slack" {
		return "", "", "", errors.New("slack source session key must have form slack:<workspace>:<channel>:<thread|channel>")
	}
	tail := strings.Join(parts[3:], ":")
	threadTS := tail
	if tail == "channel" {
		threadTS = ""
	}
	return parts[1], parts[2], threadTS, nil
}

func honchoMessageSlackTS(message clients.HonchoMessage) string {
	if message.Metadata == nil {
		return ""
	}
	return nativeStringFromAny(message.Metadata["slack_ts"])
}

func slackTSInWindow(ts string, oldestTS string, latestTS string) bool {
	ts = strings.TrimSpace(ts)
	if ts == "" && (oldestTS != "" || latestTS != "") {
		return false
	}
	if oldestTS != "" && ts != "" && ts <= strings.TrimSpace(oldestTS) {
		return false
	}
	if latestTS != "" && ts != "" && ts > strings.TrimSpace(latestTS) {
		return false
	}
	return true
}

func maxHTTPStatus(current int, candidate int) int {
	if candidate >= 500 {
		if current >= 500 && current > candidate {
			return current
		}
		return candidate
	}
	if current < 500 && candidate >= 400 {
		return candidate
	}
	return current
}

func executeSlackNativeToolAction(ctx context.Context, cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest) (any, string, string, string, map[string]any, int, error) {
	if strings.TrimSpace(cfg.SlackBotToken) == "" {
		return nil, "", "", "", map[string]any{"status": "not_attempted", "reason": "missing_slack_token"}, http.StatusFailedDependency, errors.New("SLACK_BOT_TOKEN is required for native Slack tools")
	}
	api := slackapi.New(cfg.SlackBotToken)
	switch input.Operation {
	case "channels_list":
		types := stringSliceArg(input.Arguments, "types")
		if len(types) == 0 {
			types = []string{"public_channel", "private_channel"}
		}
		channels, nextCursor, err := api.GetConversationsContext(ctx, &slackapi.GetConversationsParameters{
			Cursor:          stringArg(input.Arguments, "cursor"),
			ExcludeArchived: !boolArg(input.Arguments, "include_archived", false),
			Limit:           intArg(input.Arguments, "limit", 200),
			Types:           types,
		})
		return map[string]any{"channels": channels, "next_cursor": nextCursor}, fmt.Sprintf("listed %d Slack channel(s)", len(channels)), "slack:channels", "", map[string]any{"status": "not_applicable"}, statusFromErr(err), err
	case "channel_info":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		channel, err := api.GetConversationInfoContext(ctx, &slackapi.GetConversationInfoInput{ChannelID: channelID})
		return channel, "loaded Slack channel info", "slack:" + channelID, "", map[string]any{"status": "not_applicable"}, statusFromErr(err), err
	case "conversation_read":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		threadTS := stringArg(input.Arguments, "thread_ts")
		if threadTS != "" || boolArg(input.Arguments, "include_replies", false) {
			messages, hasMore, nextCursor, err := api.GetConversationRepliesContext(ctx, &slackapi.GetConversationRepliesParameters{
				ChannelID: channelID,
				Timestamp: threadTS,
				Cursor:    stringArg(input.Arguments, "cursor"),
				Latest:    stringArg(input.Arguments, "latest"),
				Oldest:    stringArg(input.Arguments, "oldest"),
				Limit:     intArg(input.Arguments, "limit", 100),
			})
			return map[string]any{"messages": messages, "has_more": hasMore, "next_cursor": nextCursor}, fmt.Sprintf("read %d Slack message(s)", len(messages)), "slack:" + channelID + ":" + threadTS, "", map[string]any{"status": "not_applicable"}, statusFromErr(err), err
		}
		out, err := api.GetConversationHistoryContext(ctx, &slackapi.GetConversationHistoryParameters{
			ChannelID: channelID,
			Cursor:    stringArg(input.Arguments, "cursor"),
			Latest:    stringArg(input.Arguments, "latest"),
			Oldest:    stringArg(input.Arguments, "oldest"),
			Limit:     intArg(input.Arguments, "limit", 100),
		})
		count := 0
		if out != nil {
			count = len(out.Messages)
		}
		return out, fmt.Sprintf("read %d Slack message(s)", count), "slack:" + channelID, "", map[string]any{"status": "not_applicable"}, statusFromErr(err), err
	case "user_lookup":
		if email := stringArg(input.Arguments, "email"); email != "" {
			user, err := api.GetUserByEmailContext(ctx, email)
			return user, "loaded Slack user by email", "slack_user:" + email, "", map[string]any{"status": "not_applicable"}, statusFromErr(err), err
		}
		userID := firstNonEmpty(stringArg(input.Arguments, "user_id"), input.TargetRef)
		user, err := api.GetUserInfoContext(ctx, userID)
		return user, "loaded Slack user", "slack_user:" + userID, "", map[string]any{"status": "not_applicable"}, statusFromErr(err), err
	case "message_post":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		options, err := slackMessagePostOptions(input.Arguments)
		if err != nil {
			return nil, "", "", "", slackMirrorEffect("not_attempted", ""), http.StatusBadRequest, err
		}
		channel, ts, err := api.PostMessageContext(ctx, channelID, options...)
		sourceRef := "slack:" + channel + ":" + ts
		mirrorEffect := nativeSlackRefreshMirrorEffect(ctx, cfg, repo, api, channel, ts, firstNonEmpty(stringArg(input.Arguments, "thread_ts"), ts), err)
		return map[string]any{"channel_id": channel, "ts": ts}, "posted Slack message", sourceRef, "", mirrorEffect, statusFromErr(err), err
	case "message_update":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		ts := stringArg(input.Arguments, "ts")
		options, err := slackMessageUpdateOptions(input.Arguments)
		if err != nil {
			return nil, "", "", "", slackMirrorEffect("not_attempted", ""), http.StatusBadRequest, err
		}
		channel, updatedTS, text, err := api.UpdateMessageContext(ctx, channelID, ts, options...)
		sourceRef := "slack:" + channel + ":" + updatedTS
		mirrorEffect := nativeSlackRefreshMirrorEffect(ctx, cfg, repo, api, channel, updatedTS, stringArg(input.Arguments, "thread_ts"), err)
		return map[string]any{"channel_id": channel, "ts": updatedTS, "text": text}, "updated Slack message", sourceRef, "", mirrorEffect, statusFromErr(err), err
	case "report_post":
		return executeSlackReportNativeToolAction(ctx, cfg, repo, api, input)
	case "message_delete":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		ts := stringArg(input.Arguments, "ts")
		channel, deletedTS, err := api.DeleteMessageContext(ctx, channelID, ts)
		sourceRef := "slack:" + channel + ":" + deletedTS
		mirrorEffect := nativeSlackMarkStaleMirrorEffect(ctx, cfg, repo, api, channel, deletedTS, "message_delete", err)
		return map[string]any{"channel_id": channel, "ts": deletedTS}, "deleted Slack message", sourceRef, "", mirrorEffect, statusFromErr(err), err
	case "reaction_add":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		ts := stringArg(input.Arguments, "timestamp")
		name := stringArg(input.Arguments, "name")
		err := api.AddReactionContext(ctx, name, slackapi.NewRefToMessage(channelID, ts))
		return map[string]any{"channel_id": channelID, "timestamp": ts, "name": name}, "added Slack reaction", "slack:" + channelID + ":" + ts, "", slackMirrorEffect("not_applicable", ""), statusFromErr(err), err
	case "reaction_remove":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		ts := stringArg(input.Arguments, "timestamp")
		name := stringArg(input.Arguments, "name")
		err := api.RemoveReactionContext(ctx, name, slackapi.NewRefToMessage(channelID, ts))
		return map[string]any{"channel_id": channelID, "timestamp": ts, "name": name}, "removed Slack reaction", "slack:" + channelID + ":" + ts, "", slackMirrorEffect("not_applicable", ""), statusFromErr(err), err
	case "file_upload":
		params, err := slackUploadParams(input.Arguments, input.TargetRef)
		if err != nil {
			return nil, "", "", "", slackMirrorEffect("not_attempted", ""), http.StatusBadRequest, err
		}
		file, err := api.UploadFileContext(ctx, params)
		sourceRef := ""
		if file != nil {
			sourceRef = "slack_file:" + file.ID
		}
		fileID := ""
		if file != nil {
			fileID = file.ID
		}
		mirrorEffect := nativeSlackFileUploadMirrorEffect(ctx, cfg, repo, api, fileID, err)
		return file, "uploaded Slack file", sourceRef, "", mirrorEffect, statusFromErr(err), err
	case "channel_create":
		channel, err := api.CreateConversationContext(ctx, slackapi.CreateConversationParams{ChannelName: stringArg(input.Arguments, "name"), IsPrivate: boolArg(input.Arguments, "is_private", false)})
		sourceRef := ""
		if channel != nil {
			sourceRef = "slack:" + channel.ID
		}
		return channel, "created Slack channel", sourceRef, "", slackMirrorEffect("not_applicable", ""), statusFromErr(err), err
	case "channel_rename":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		channel, err := api.RenameConversationContext(ctx, channelID, stringArg(input.Arguments, "name"))
		return channel, "renamed Slack channel", "slack:" + channelID, "", slackMirrorEffect("not_applicable", ""), statusFromErr(err), err
	case "channel_archive":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		err := api.ArchiveConversationContext(ctx, channelID)
		mirrorEffect := nativeSlackArchiveMirrorEffect(ctx, cfg, repo, api, channelID, err)
		return map[string]any{"channel_id": channelID}, "archived Slack channel", "slack:" + channelID, "", mirrorEffect, statusFromErr(err), err
	case "channel_invite":
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		channel, err := api.InviteUsersToConversationContext(ctx, channelID, stringSliceArg(input.Arguments, "user_ids")...)
		return channel, "invited Slack channel users", "slack:" + channelID, "", slackMirrorEffect("not_applicable", ""), statusFromErr(err), err
	default:
		message := fmt.Sprintf("native Slack operation %s is registered but not implemented", input.Operation)
		return nil, "", "", "", slackMirrorEffect("not_attempted", ""), http.StatusNotImplemented, errors.New(message)
	}
}

func executeNotionNativeToolAction(ctx context.Context, cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest) (any, string, string, string, map[string]any, int, error) {
	if strings.TrimSpace(cfg.NotionToken) == "" {
		return nil, "", "", "", notionMirrorEffect("not_attempted", "", "missing_notion_token"), http.StatusFailedDependency, errors.New("NOTION_TOKEN is required for native Notion tools")
	}
	api := clients.NewNotionClientWithConfig(clients.NotionClientOptions{
		BaseURL:           cfg.NotionAPIBaseURL,
		Token:             cfg.NotionToken,
		Version:           cfg.NotionAPIVersion,
		RequestsPerSecond: cfg.NotionMirrorRequestsPerSecond,
		MaxRetries:        cfg.NotionMirrorMaxRetries,
		RetryBaseDelay:    cfg.NotionMirrorRetryBaseDelay,
	})
	switch input.Operation {
	case "search":
		out, err := api.Search(ctx, clients.NotionSearchOptions{
			Query:    stringArg(input.Arguments, "query"),
			Filter:   mapArg(input.Arguments, "filter"),
			Sort:     mapArg(input.Arguments, "sort"),
			PageSize: intArg(input.Arguments, "page_size", 100),
			Cursor:   stringArg(input.Arguments, "cursor"),
		})
		return out, "searched Notion", "notion:search", "", notionMirrorEffect("not_applicable", "", ""), statusFromErr(err), err
	case "page_get":
		pageID := firstNonEmpty(stringArg(input.Arguments, "page_id"), input.TargetRef)
		out, err := api.RetrievePage(ctx, pageID)
		return out, "loaded Notion page", "notion:page:" + pageID, "", notionMirrorEffect("not_applicable", "", ""), statusFromErr(err), err
	case "blocks_children":
		blockID := firstNonEmpty(stringArg(input.Arguments, "block_id"), input.TargetRef)
		out, err := api.ListBlockChildren(ctx, blockID, stringArg(input.Arguments, "cursor"), intArg(input.Arguments, "page_size", 100))
		return nativeNotionBlockChildrenOutput(out), "loaded Notion block children", "notion:block:" + blockID, "", notionMirrorEffect("not_applicable", "", ""), statusFromErr(err), err
	case "database_get":
		databaseID := firstNonEmpty(stringArg(input.Arguments, "database_id"), input.TargetRef)
		out, err := api.RetrieveDatabase(ctx, databaseID)
		return out, "loaded Notion database", "notion:database:" + databaseID, "", notionMirrorEffect("not_applicable", "", ""), statusFromErr(err), err
	case "data_source_get":
		dataSourceID := firstNonEmpty(stringArg(input.Arguments, "data_source_id"), input.TargetRef)
		out, err := api.RetrieveDataSource(ctx, dataSourceID)
		return out, "loaded Notion data source", "notion:data_source:" + dataSourceID, "", notionMirrorEffect("not_applicable", "", ""), statusFromErr(err), err
	case "data_source_query":
		dataSourceID := firstNonEmpty(stringArg(input.Arguments, "data_source_id"), input.TargetRef)
		out, err := api.QueryDataSource(ctx, dataSourceID, clients.NotionDataSourceQueryOptions{
			Cursor:   stringArg(input.Arguments, "cursor"),
			PageSize: intArg(input.Arguments, "page_size", 100),
		})
		return out, "queried Notion data source", "notion:data_source:" + dataSourceID, "", notionMirrorEffect("not_applicable", "", ""), statusFromErr(err), err
	case "page_create":
		out, err := api.CreatePage(ctx, notionPayload(input.Arguments, "parent", "properties", "children", "icon", "cover"))
		pageID := nativeStringFromAny(out["id"])
		return out, "created Notion page", "notion:page:" + pageID, "", nativeNotionMirrorEffect(cfg, repo, input, pageID, "page", "page_create", err), statusFromErr(err), err
	case "page_update":
		pageID := firstNonEmpty(stringArg(input.Arguments, "page_id"), input.TargetRef)
		out, err := api.UpdatePage(ctx, pageID, notionPayload(input.Arguments, "properties", "icon", "cover"))
		return out, "updated Notion page", "notion:page:" + pageID, "", nativeNotionMirrorEffect(cfg, repo, input, pageID, "page", "page_update", err), statusFromErr(err), err
	case "page_archive":
		pageID := firstNonEmpty(stringArg(input.Arguments, "page_id"), input.TargetRef)
		out, err := api.UpdatePage(ctx, pageID, map[string]any{"in_trash": true})
		return out, "archived Notion page", "notion:page:" + pageID, "", nativeNotionMirrorEffect(cfg, repo, input, pageID, "page", "page_archive", err), statusFromErr(err), err
	case "blocks_append":
		blockID := firstNonEmpty(stringArg(input.Arguments, "block_id"), input.TargetRef)
		out, err := api.AppendBlockChildren(ctx, blockID, arrayArg(input.Arguments, "children"))
		return out, "appended Notion block children", "notion:block:" + blockID, "", nativeNotionMirrorEffect(cfg, repo, input, firstNonEmpty(stringArg(input.Arguments, "mirror_root_id"), blockID), "page", "blocks_append", err), statusFromErr(err), err
	case "block_update":
		blockID := firstNonEmpty(stringArg(input.Arguments, "block_id"), input.TargetRef)
		out, err := api.UpdateBlock(ctx, blockID, mapArg(input.Arguments, "block"))
		return out, "updated Notion block", "notion:block:" + blockID, "", nativeNotionMirrorEffect(cfg, repo, input, firstNonEmpty(stringArg(input.Arguments, "mirror_root_id"), blockID), "page", "block_update", err), statusFromErr(err), err
	case "block_delete":
		blockID := firstNonEmpty(stringArg(input.Arguments, "block_id"), input.TargetRef)
		out, err := api.DeleteBlock(ctx, blockID)
		return out, "deleted Notion block", "notion:block:" + blockID, "", nativeNotionMirrorEffect(cfg, repo, input, firstNonEmpty(stringArg(input.Arguments, "mirror_root_id"), blockID), "page", "block_delete", err), statusFromErr(err), err
	case "comment_create":
		out, err := api.CreateComment(ctx, notionPayload(input.Arguments, "parent", "rich_text", "discussion_id"))
		commentID := nativeStringFromAny(out["id"])
		return out, "created Notion comment", "notion:comment:" + commentID, "", nativeNotionMirrorEffect(cfg, repo, input, firstNonEmpty(stringArg(input.Arguments, "mirror_root_id"), commentID), "page", "comment_create", err), statusFromErr(err), err
	default:
		message := fmt.Sprintf("native Notion operation %s is registered but not implemented", input.Operation)
		return nil, "", "", "", notionMirrorEffect("not_attempted", "", "implementation_pending"), http.StatusNotImplemented, errors.New(message)
	}
}

func validateNativeToolActionPolicy(cfg config.Config, claims nativeToolClaims, input nativeToolActionRequest, isWrite bool, isDestructive bool) (error, int) {
	if isDestructive && !input.ConfirmDestroy {
		return errors.New("destructive native tool operation requires confirm_destroy=true"), http.StatusBadRequest
	}
	if input.Surface == "slack" {
		channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
		if channelID != "" {
			for _, denied := range cfg.SlackMirrorChannelDenylist {
				if strings.TrimSpace(denied) == channelID {
					return fmt.Errorf("slack channel %s is denied by policy", channelID), http.StatusForbidden
				}
			}
		}
		if isWrite {
			if err := validateNativeSlackWriteScope(claims, input); err != nil {
				return err, http.StatusForbidden
			}
		}
	}
	if input.Surface == "notion" && isWrite && cfg.NotionMirrorEnabled {
		if rootID := strings.TrimSpace(stringArg(input.Arguments, "mirror_root_id")); rootID != "" {
			return nil, http.StatusOK
		}
		return errors.New("notion write requires mirror_root_id or successful mirror root resolution"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func validateNativeSlackWriteScope(claims nativeToolClaims, input nativeToolActionRequest) error {
	if !nativeSlackBoundDeliveryOperation(input.Operation) {
		return fmt.Errorf("native Slack write operation %s is not available to workflow execution tokens", input.Operation)
	}
	if strings.TrimSpace(claims.SlackScope) != "bound_thread" {
		return errors.New("native Slack write requires slack_delivery_scope=bound_thread")
	}
	boundChannelID := strings.TrimSpace(claims.SlackChannelID)
	if boundChannelID == "" {
		return errors.New("native Slack write requires a bound slack_channel_id claim")
	}
	channelID := firstNonEmpty(stringArg(input.Arguments, "channel_id"), input.TargetRef)
	if channelID == "" {
		return errors.New("native Slack write requires channel_id")
	}
	if channelID != boundChannelID {
		return fmt.Errorf("slack channel %s is outside bound Slack delivery scope", channelID)
	}
	boundThreadTS := strings.TrimSpace(claims.SlackThreadTS)
	if boundThreadTS == "" {
		return nil
	}
	if nativeSlackReactionOperation(input.Operation) {
		timestamp := strings.TrimSpace(stringArg(input.Arguments, "timestamp"))
		if timestamp == "" {
			return errors.New("native Slack reaction requires timestamp for bound-thread delivery")
		}
		if timestamp != boundThreadTS {
			return fmt.Errorf("slack reaction timestamp %s is outside bound Slack delivery scope", timestamp)
		}
		return nil
	}
	threadTS := strings.TrimSpace(stringArg(input.Arguments, "thread_ts"))
	if threadTS == "" {
		return errors.New("native Slack write requires thread_ts for bound-thread delivery")
	}
	if threadTS != boundThreadTS {
		return fmt.Errorf("slack thread %s is outside bound Slack delivery scope", threadTS)
	}
	return nil
}

func nativeSlackBoundDeliveryOperation(operation string) bool {
	switch strings.TrimSpace(operation) {
	case "message_post", "report_post", "file_upload", "reaction_add", "reaction_remove":
		return true
	default:
		return false
	}
}

func nativeSlackReactionOperation(operation string) bool {
	switch strings.TrimSpace(operation) {
	case "reaction_add", "reaction_remove":
		return true
	default:
		return false
	}
}

func authorizeNativeToolAction(cfg config.Config, r *http.Request, surface string) (nativeToolClaims, error) {
	secret := strings.TrimSpace(cfg.NativeToolsClientToken)
	if secret == "" {
		return nativeToolClaims{}, errors.New("RSI_NATIVE_TOOLS_CLIENT_TOKEN is required for native tools")
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(auth, "Bearer ") {
		return nativeToolClaims{}, errors.New("missing native tool bearer token")
	}
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	if token == "" {
		return nativeToolClaims{}, errors.New("missing native tool bearer token")
	}
	if hmac.Equal([]byte(token), []byte(secret)) {
		return nativeToolClaims{}, errors.New("static native tools client token is not accepted as an execution token")
	}
	claims, err := verifyNativeToolsExecutionToken(secret, token, time.Now().UTC())
	if err != nil {
		return nativeToolClaims{}, err
	}
	if surface != "" && !nativeToolSurfaceAllowed(claims, surface) {
		return nativeToolClaims{}, fmt.Errorf("native tool token is not scoped for surface %q", surface)
	}
	return claims, nil
}

func verifyNativeToolsExecutionToken(secret string, token string, now time.Time) (nativeToolClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nativeToolClaims{}, errors.New("native tool execution token must be a signed JWT")
	}
	signed := parts[0] + "." + parts[1]
	expected := nativeToolSignature(secret, signed)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nativeToolClaims{}, errors.New("native tool execution token signature is invalid")
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nativeToolClaims{}, errors.New("native tool execution token header is invalid")
	}
	var headerClaims map[string]string
	if err := json.Unmarshal(header, &headerClaims); err != nil {
		return nativeToolClaims{}, errors.New("native tool execution token header is invalid")
	}
	if headerClaims["alg"] != "HS256" || headerClaims["typ"] != "JWT" {
		return nativeToolClaims{}, errors.New("native tool execution token header is unsupported")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nativeToolClaims{}, errors.New("native tool execution token payload is invalid")
	}
	var claims nativeToolClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nativeToolClaims{}, errors.New("native tool execution token claims are invalid")
	}
	if claims.Audience != nativeToolsAudience {
		return nativeToolClaims{}, errors.New("native tool execution token audience is invalid")
	}
	if claims.IssuedAt <= 0 || claims.ExpiresAt <= 0 || claims.ExpiresAt <= claims.IssuedAt {
		return nativeToolClaims{}, errors.New("native tool execution token time claims are invalid")
	}
	const skew = 2 * time.Minute
	if now.Add(skew).Unix() < claims.IssuedAt {
		return nativeToolClaims{}, errors.New("native tool execution token issued-at is in the future")
	}
	if now.Add(-skew).Unix() > claims.ExpiresAt {
		return nativeToolClaims{}, errors.New("native tool execution token has expired")
	}
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Unix(claims.IssuedAt, 0)) > 2*time.Hour {
		return nativeToolClaims{}, errors.New("native tool execution token lifetime exceeds 2h")
	}
	if strings.TrimSpace(claims.ExecutionID) == "" || strings.TrimSpace(claims.OperationID) == "" ||
		strings.TrimSpace(claims.TraceID) == "" || strings.TrimSpace(claims.WorkflowID) == "" ||
		strings.TrimSpace(claims.ConversationID) == "" || strings.TrimSpace(claims.Actor) == "" ||
		len(claims.Surfaces) == 0 {
		return nativeToolClaims{}, errors.New("native tool execution token is missing required claims")
	}
	return claims, nil
}

func mintNativeToolsExecutionToken(secret string, claims nativeToolClaims) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signed := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	return signed + "." + nativeToolSignature(secret, signed), nil
}

func nativeToolSignature(secret string, signed string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func nativeToolOperationKnown(surface string, operation string) bool {
	if nativeToolReadOps[surface][operation] {
		return true
	}
	return nativeToolWriteOps[surface][operation]
}

func nativeToolSurfaceAllowed(claims nativeToolClaims, surface string) bool {
	surface = strings.TrimSpace(surface)
	for _, candidate := range claims.Surfaces {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || candidate == surface {
			return true
		}
	}
	return false
}

func nativeToolRequestHash(input nativeToolActionRequest, destructive bool) (string, error) {
	payload := map[string]any{
		"surface":         input.Surface,
		"operation":       input.Operation,
		"target_ref":      input.TargetRef,
		"reason":          input.Reason,
		"destructive":     destructive,
		"confirm_destroy": input.ConfirmDestroy,
		"arguments":       input.Arguments,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func nativeToolReadIdempotencyKey(claims nativeToolClaims, input nativeToolActionRequest) string {
	hash, _ := nativeToolRequestHash(input, false)
	if len(hash) > 16 {
		hash = hash[:16]
	}
	return "read:" + strings.Join([]string{claims.ExecutionID, claims.OperationID, input.Surface, input.Operation, hash}, ":")
}

func nativeToolTargetRef(args map[string]any) string {
	for _, key := range []string{
		"target_ref",
		"channel_id",
		"page_id",
		"database_id",
		"data_source_id",
		"block_id",
		"source_ref",
		"page_ref",
		"slug",
		"issue",
		"issue_ref",
		"short_id",
		"project_ref",
		"project",
		"org",
		"release",
	} {
		if value := stringArg(args, key); value != "" {
			return value
		}
	}
	return ""
}

func resolveNativeToolActionArguments(cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest) nativeToolActionRequest {
	if input.Surface == "notion" && nativeToolWriteOps[input.Surface][input.Operation] && cfg.NotionMirrorEnabled && stringArg(input.Arguments, "mirror_root_id") == "" {
		if rootID := resolveNotionMirrorRootID(cfg, repo, input); rootID != "" {
			if input.Arguments == nil {
				input.Arguments = map[string]any{}
			}
			input.Arguments["mirror_root_id"] = rootID
		}
	}
	return input
}

func resolveNotionMirrorRootID(cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest) string {
	candidates := []string{
		input.TargetRef,
		stringArg(input.Arguments, "page_id"),
		stringArg(input.Arguments, "database_id"),
		stringArg(input.Arguments, "data_source_id"),
		stringArg(input.Arguments, "parent_id"),
		stringArg(input.Arguments, "block_id"),
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		for _, root := range cfg.NotionMirrorAllowlist {
			if candidate == strings.TrimSpace(root) {
				return candidate
			}
		}
	}
	statusStore, ok := repo.(storepkg.SourceMirrorStatusStore)
	if !ok {
		return ""
	}
	records, err := statusStore.ListSourceMirrorRecords([]string{companyknowledge.NotionDocumentSourceType, companyknowledge.NotionCrawlMissSourceType}, 5000)
	if err != nil {
		return ""
	}
	for _, candidate := range candidates {
		candidate = normalizeNotionID(candidate)
		if candidate == "" {
			continue
		}
		for _, record := range records {
			if rootID := notionRootIDFromMirrorRecord(record, candidate); rootID != "" {
				return rootID
			}
		}
	}
	return ""
}

func notionRootIDFromMirrorRecord(record storepkg.SourceMirrorRecord, candidate string) string {
	metadata := record.Metadata
	recordObjectIDs := []string{
		nativeStringFromAny(metadata["object_id"]),
		nativeStringFromAny(metadata["notion_page_id"]),
		nativeStringFromAny(metadata["notion_database_id"]),
		nativeStringFromAny(metadata["notion_data_source_id"]),
		nativeStringFromAny(metadata["target_id"]),
	}
	if strings.Contains(record.SourceKey, candidate) || strings.Contains(record.SourceSessionKey, candidate) {
		recordObjectIDs = append(recordObjectIDs, candidate)
	}
	for _, objectID := range recordObjectIDs {
		if normalizeNotionID(objectID) != candidate {
			continue
		}
		for _, key := range []string{"notion_root_id", "root_id"} {
			if rootID := normalizeNotionID(nativeStringFromAny(metadata[key])); rootID != "" {
				return rootID
			}
		}
	}
	return ""
}

func stringArg(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	raw, ok := args[key]
	if !ok || raw == nil {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case fmt.Stringer:
		return strings.TrimSpace(value.String())
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func intArg(args map[string]any, key string, fallback int) int {
	if args == nil {
		return fallback
	}
	switch value := args[key].(type) {
	case int:
		if value > 0 {
			return value
		}
	case int64:
		if value > 0 {
			return int(value)
		}
	case float64:
		if value > 0 {
			return int(value)
		}
	case string:
		var parsed int
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func stringSliceArg(args map[string]any, key string) []string {
	if args == nil {
		return nil
	}
	switch value := args[key].(type) {
	case []string:
		return uniqueNonEmpty(value)
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			out = append(out, strings.TrimSpace(fmt.Sprint(item)))
		}
		return uniqueNonEmpty(out)
	case string:
		return parseSourceMirrorStatusQuery([]string{value})
	default:
		return nil
	}
}

func mapArg(args map[string]any, key string) map[string]any {
	if args == nil {
		return nil
	}
	if value, ok := args[key].(map[string]any); ok {
		out := make(map[string]any, len(value))
		for k, v := range value {
			out[k] = v
		}
		return out
	}
	return nil
}

func arrayArg(args map[string]any, key string) []any {
	if args == nil {
		return nil
	}
	switch value := args[key].(type) {
	case []any:
		return append([]any(nil), value...)
	default:
		return nil
	}
}

func boolArg(args map[string]any, key string, fallback bool) bool {
	if args == nil {
		return fallback
	}
	switch value := args[key].(type) {
	case bool:
		return value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	return fallback
}

func statusFromErr(err error) int {
	if err != nil {
		return http.StatusBadGateway
	}
	return http.StatusOK
}

func slackMessagePostOptions(args map[string]any) ([]slackapi.MsgOption, error) {
	options := []slackapi.MsgOption{slackapi.MsgOptionText(stringArg(args, "text"), false)}
	options, err := appendSlackMessageDecorators(options, args, true)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func slackMessageUpdateOptions(args map[string]any) ([]slackapi.MsgOption, error) {
	options := []slackapi.MsgOption{}
	if hasNativeArg(args, "text") {
		options = append(options, slackapi.MsgOptionText(stringArg(args, "text"), false))
	}
	return appendSlackMessageDecorators(options, args, false)
}

func appendSlackMessageDecorators(options []slackapi.MsgOption, args map[string]any, includeThread bool) ([]slackapi.MsgOption, error) {
	if includeThread {
		if threadTS := stringArg(args, "thread_ts"); threadTS != "" {
			options = append(options, slackapi.MsgOptionTS(threadTS))
		}
	}
	if hasNativeArg(args, "blocks") {
		blocks, err := slackBlocksArg(args, "blocks")
		if err != nil {
			return nil, err
		}
		options = append(options, slackapi.MsgOptionBlocks(blocks.BlockSet...))
	}
	if hasNativeArg(args, "attachments") {
		attachments, err := slackAttachmentsArg(args, "attachments")
		if err != nil {
			return nil, err
		}
		options = append(options, slackapi.MsgOptionAttachments(attachments...))
	}
	if boolArg(args, "unfurl_links", false) {
		options = append(options, slackapi.MsgOptionEnableLinkUnfurl())
	}
	if !boolArg(args, "unfurl_media", true) {
		options = append(options, slackapi.MsgOptionDisableMediaUnfurl())
	}
	return options, nil
}

func hasNativeArg(args map[string]any, key string) bool {
	if args == nil {
		return false
	}
	_, ok := args[key]
	return ok
}

func slackBlocksArg(args map[string]any, key string) (slackapi.Blocks, error) {
	raw, ok := args[key]
	if !ok {
		return slackapi.Blocks{}, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return slackapi.Blocks{}, fmt.Errorf("marshal Slack blocks: %w", err)
	}
	var blocks slackapi.Blocks
	if err := json.Unmarshal(data, &blocks); err != nil {
		return slackapi.Blocks{}, fmt.Errorf("decode Slack blocks: %w", err)
	}
	return blocks, nil
}

func slackAttachmentsArg(args map[string]any, key string) ([]slackapi.Attachment, error) {
	raw, ok := args[key]
	if !ok {
		return nil, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal Slack attachments: %w", err)
	}
	var attachments []slackapi.Attachment
	if err := json.Unmarshal(data, &attachments); err != nil {
		return nil, fmt.Errorf("decode Slack attachments: %w", err)
	}
	return attachments, nil
}

func slackUploadParams(args map[string]any, targetRef string) (slackapi.UploadFileParameters, error) {
	channelID := firstNonEmpty(stringArg(args, "channel_id"), targetRef)
	contentBase64 := stringArg(args, "content_base64")
	params := slackapi.UploadFileParameters{
		Channel:         channelID,
		InitialComment:  stringArg(args, "initial_comment"),
		ThreadTimestamp: stringArg(args, "thread_ts"),
		Filename:        stringArg(args, "filename"),
		Title:           stringArg(args, "title"),
		Content:         stringArg(args, "content"),
	}
	if contentBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(contentBase64)
		if err != nil {
			return slackapi.UploadFileParameters{}, fmt.Errorf("decode content_base64: %w", err)
		}
		params.Content = string(decoded)
		params.FileSize = len(decoded)
	}
	path := ""
	if params.Content == "" {
		path = slackUploadLocalPath(firstNonEmpty(stringArg(args, "path"), stringArg(args, "artifact_ref")))
	}
	if path != "" {
		info, err := os.Stat(path)
		if err != nil {
			return slackapi.UploadFileParameters{}, err
		}
		if info.IsDir() {
			return slackapi.UploadFileParameters{}, fmt.Errorf("slack upload path is a directory: %s", path)
		}
		params.File = path
		params.FileSize = int(info.Size())
		if params.Filename == "" {
			params.Filename = info.Name()
		}
	}
	if params.File == "" && params.Content != "" && params.FileSize == 0 {
		params.FileSize = len([]byte(params.Content))
	}
	if params.File == "" && params.Content == "" {
		return slackapi.UploadFileParameters{}, errors.New("slack file_upload requires path, artifact_ref, content, or content_base64")
	}
	if params.File == "" && params.Filename == "" {
		params.Filename = "upload.txt"
	}
	return params, nil
}

func slackUploadLocalPath(value string) string {
	value = strings.TrimSpace(value)
	return strings.TrimPrefix(value, "file://")
}

func nativeSlackRefreshMirrorEffect(ctx context.Context, cfg config.Config, repo storepkg.Repository, api *slackapi.Client, channelID string, ts string, threadTS string, sourceErr error) map[string]any {
	sourceRef := "slack:" + strings.TrimSpace(channelID) + ":" + strings.TrimSpace(ts)
	if sourceErr != nil {
		return slackMirrorEffect("not_attempted", sourceRef)
	}
	if !cfg.SlackMirrorEnabled {
		return slackMirrorEffect("not_applicable", sourceRef)
	}
	mirrorStore, ok := repo.(storepkg.SourceMirrorWriteStore)
	if !ok {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = "configured store does not support source mirror writes"
		return effect
	}
	workspaceID, err := nativeSlackWorkspaceID(ctx, cfg, api)
	if err != nil {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = err.Error()
		return effect
	}
	msg, found, err := nativeSlackFetchMessage(ctx, api, channelID, ts, threadTS)
	if err != nil {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = err.Error()
		return effect
	}
	if !found {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = "slack message was not found after source mutation"
		return effect
	}
	if strings.TrimSpace(msg.Permalink) == "" {
		if permalink, err := api.GetPermalinkContext(ctx, &slackapi.PermalinkParameters{Channel: channelID, Ts: msg.Timestamp}); err == nil {
			msg.Permalink = permalink
		}
	}
	channelMetadata := slackMirrorChannelMetadataForChannel(ctx, cfg, api, channelID)
	input := slackInputFromMessage(workspaceID, channelID, msg, "")
	if strings.TrimSpace(threadTS) != "" {
		input.ThreadTS = strings.TrimSpace(threadTS)
	}
	applySlackMirrorPolicyMetadata(&input, cfg, channelMetadata)
	mirror := companyknowledge.NewSlackMirror(mirrorStore, clients.NewHonchoClientWithAPIKey(cfg.HonchoBaseURL, cfg.HonchoAPIKey), companyknowledge.SlackMirrorOptions{
		Environment:     cfg.Environment,
		HonchoWorkspace: cfg.HonchoWorkspaceID,
	})
	result, err := mirror.IngestMessage(ctx, input)
	if err != nil {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = err.Error()
		return effect
	}
	wikiBatch := newSlackWikiPublishBatch(cfg, mirrorStore)
	if shouldPublishSlackWikiSource(result) {
		if err := wikiBatch.record(ctx, input); err != nil {
			effect := slackMirrorEffect("failed", sourceRef)
			effect["error"] = err.Error()
			effect["source_key"] = result.SourceKey
			return effect
		}
	}
	if err := wikiBatch.publish(ctx); err != nil {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = err.Error()
		effect["source_key"] = result.SourceKey
		return effect
	}
	effect := slackMirrorEffect("refreshed", sourceRef)
	effect["source_key"] = result.SourceKey
	effect["source_session_key"] = result.SourceSessionKey
	effect["honcho_session_id"] = result.HonchoSessionID
	effect["honcho_message_id"] = result.HonchoMessageID
	if result.Skipped {
		effect["status"] = "skipped"
		effect["reason"] = result.SkipReason
	}
	return effect
}

func nativeSlackFileUploadMirrorEffect(ctx context.Context, cfg config.Config, repo storepkg.Repository, api *slackapi.Client, fileID string, sourceErr error) map[string]any {
	if sourceErr != nil {
		return slackMirrorEffect("not_attempted", "")
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		effect := slackMirrorEffect("failed", "")
		effect["error"] = "slack upload returned no file id"
		return effect
	}
	file, _, _, err := api.GetFileInfoContext(ctx, fileID, 1, 1)
	if err != nil {
		effect := slackMirrorEffect("failed", "slack_file:"+fileID)
		effect["error"] = err.Error()
		return effect
	}
	if file == nil {
		effect := slackMirrorEffect("failed", "slack_file:"+fileID)
		effect["error"] = "slack upload returned no file"
		return effect
	}
	for channelID, shares := range file.Shares.Public {
		for _, share := range shares {
			if strings.TrimSpace(share.Ts) != "" {
				return nativeSlackRefreshMirrorEffect(ctx, cfg, repo, api, channelID, share.Ts, share.ThreadTs, nil)
			}
		}
	}
	for channelID, shares := range file.Shares.Private {
		for _, share := range shares {
			if strings.TrimSpace(share.Ts) != "" {
				return nativeSlackRefreshMirrorEffect(ctx, cfg, repo, api, channelID, share.Ts, share.ThreadTs, nil)
			}
		}
	}
	effect := slackMirrorEffect("pending_manual_refresh", "slack_file:"+strings.TrimSpace(file.ID))
	effect["reason"] = "slack file response did not include share message timestamp"
	return effect
}

func nativeSlackMarkStaleMirrorEffect(ctx context.Context, cfg config.Config, repo storepkg.Repository, api *slackapi.Client, channelID string, ts string, reason string, sourceErr error) map[string]any {
	sourceRef := "slack:" + strings.TrimSpace(channelID) + ":" + strings.TrimSpace(ts)
	if sourceErr != nil {
		return slackMirrorEffect("not_attempted", sourceRef)
	}
	if !cfg.SlackMirrorEnabled {
		return slackMirrorEffect("not_applicable", sourceRef)
	}
	mirrorStore, ok := repo.(storepkg.SourceMirrorWriteStore)
	if !ok {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = "configured store does not support source mirror writes"
		return effect
	}
	workspaceID, err := nativeSlackWorkspaceID(ctx, cfg, api)
	if err != nil {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = err.Error()
		return effect
	}
	sourceKey := companyknowledge.SlackMessageSourceKey(workspaceID, channelID, ts)
	sessionKey := companyknowledge.SlackSessionSourceKey(workspaceID, channelID, "", false)
	if existing, found, err := mirrorStore.GetSourceMirrorRecord(companyknowledge.SlackMessageSourceType, sourceKey); err == nil && found {
		sessionKey = strings.TrimSpace(firstNonEmpty(existing.SourceSessionKey, sessionKey))
	}
	record := storepkg.SourceMirrorRecord{
		SourceType:       companyknowledge.SlackMessageSourceType,
		SourceKey:        sourceKey,
		Workspace:        workspaceID,
		Environment:      strings.TrimSpace(cfg.Environment),
		SourceSessionKey: sessionKey,
		HonchoWorkspace:  companyknowledge.HonchoCompatibleName("workspace", firstNonEmpty(cfg.HonchoWorkspaceID, "rsi_company_knowledge")),
		HonchoSessionID:  companyknowledge.HonchoCompatibleName("slack", sessionKey),
		SourceRevision:   "stale:" + strings.TrimSpace(reason),
		Metadata: map[string]any{
			"source":       "slack",
			"channel_id":   strings.TrimSpace(channelID),
			"slack_ts":     strings.TrimSpace(ts),
			"stale_reason": strings.TrimSpace(reason),
		},
	}
	if _, err := mirrorStore.MarkSourceMirrorRecordStale(record, reason, map[string]any{"stale_observed_at": time.Now().UTC().Format(time.RFC3339)}); err != nil {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = err.Error()
		return effect
	}
	effect := slackMirrorEffect("stale_marked", sourceRef)
	effect["source_key"] = sourceKey
	return effect
}

func nativeSlackArchiveMirrorEffect(ctx context.Context, cfg config.Config, repo storepkg.Repository, api *slackapi.Client, channelID string, sourceErr error) map[string]any {
	sourceRef := "slack:" + strings.TrimSpace(channelID)
	if sourceErr != nil {
		return slackMirrorEffect("not_attempted", sourceRef)
	}
	if !cfg.SlackMirrorEnabled {
		return slackMirrorEffect("not_applicable", sourceRef)
	}
	mirrorStore, ok := repo.(storepkg.SourceMirrorWriteStore)
	if !ok {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = "configured store does not support source mirror writes"
		return effect
	}
	statusStore, ok := repo.(storepkg.SourceMirrorStatusStore)
	if !ok {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = "configured store does not support source mirror status reads"
		return effect
	}
	records, err := statusStore.ListSourceMirrorRecords([]string{companyknowledge.SlackMessageSourceType}, 10000)
	if err != nil {
		effect := slackMirrorEffect("failed", sourceRef)
		effect["error"] = err.Error()
		return effect
	}
	marked := 0
	for _, record := range records {
		if !strings.Contains(record.SourceSessionKey, ":"+strings.TrimSpace(channelID)+":") && !strings.Contains(record.SourceKey, ":"+strings.TrimSpace(channelID)+":") {
			continue
		}
		record.SourceRevision = "stale:channel_archive"
		record.Metadata = mergeStringAnyMaps(record.Metadata, map[string]any{"stale_reason": "channel_archive"})
		if _, err := mirrorStore.MarkSourceMirrorRecordStale(record, "channel_archive", map[string]any{"stale_observed_at": time.Now().UTC().Format(time.RFC3339)}); err != nil {
			effect := slackMirrorEffect("failed", sourceRef)
			effect["error"] = err.Error()
			effect["marked_count"] = marked
			return effect
		}
		marked++
	}
	effect := slackMirrorEffect("stale_marked", sourceRef)
	effect["marked_count"] = marked
	return effect
}

func nativeSlackWorkspaceID(ctx context.Context, cfg config.Config, api *slackapi.Client) (string, error) {
	if strings.TrimSpace(cfg.SlackWorkspaceID) != "" {
		return strings.TrimSpace(cfg.SlackWorkspaceID), nil
	}
	if api == nil {
		if strings.TrimSpace(cfg.SlackBotToken) == "" {
			return "", errors.New("RSI_SLACK_WORKSPACE_ID or SLACK_BOT_TOKEN is required to resolve Slack workspace")
		}
		api = slackapi.New(cfg.SlackBotToken)
	}
	auth, err := api.AuthTestContext(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(auth.TeamID) == "" {
		return "", errors.New("Slack auth.test returned empty team_id")
	}
	return strings.TrimSpace(auth.TeamID), nil
}

func nativeSlackFetchMessage(ctx context.Context, api *slackapi.Client, channelID string, ts string, threadTS string) (slackapi.Message, bool, error) {
	if strings.TrimSpace(threadTS) != "" {
		messages, _, _, err := api.GetConversationRepliesContext(ctx, &slackapi.GetConversationRepliesParameters{
			ChannelID:          strings.TrimSpace(channelID),
			Timestamp:          strings.TrimSpace(threadTS),
			Oldest:             strings.TrimSpace(ts),
			Latest:             strings.TrimSpace(ts),
			Inclusive:          true,
			Limit:              100,
			IncludeAllMetadata: true,
		})
		if err != nil {
			return slackapi.Message{}, false, err
		}
		for _, msg := range messages {
			if strings.TrimSpace(msg.Timestamp) == strings.TrimSpace(ts) {
				return msg, true, nil
			}
		}
	}
	resp, err := api.GetConversationHistoryContext(ctx, &slackapi.GetConversationHistoryParameters{
		ChannelID:          strings.TrimSpace(channelID),
		Oldest:             strings.TrimSpace(ts),
		Latest:             strings.TrimSpace(ts),
		Inclusive:          true,
		Limit:              1,
		IncludeAllMetadata: true,
	})
	if err != nil {
		return slackapi.Message{}, false, err
	}
	if resp != nil {
		for _, msg := range resp.Messages {
			if strings.TrimSpace(msg.Timestamp) == strings.TrimSpace(ts) {
				return msg, true, nil
			}
		}
	}
	return slackapi.Message{}, false, nil
}

func notionPayload(args map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for _, key := range keys {
		if args == nil {
			continue
		}
		if value, ok := args[key]; ok && value != nil {
			out[key] = value
		}
	}
	return out
}

func nativeStringFromAny(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func slackMirrorEffect(status string, sourceRef string) map[string]any {
	return map[string]any{
		"status":     firstNonEmpty(status, "not_applicable"),
		"source_ref": sourceRef,
	}
}

func notionMirrorEffect(status string, objectID string, reason string) map[string]any {
	return map[string]any{
		"status":    firstNonEmpty(status, "not_applicable"),
		"object_id": objectID,
		"reason":    reason,
	}
}

func nativeNotionMirrorEffect(cfg config.Config, repo storepkg.Repository, input nativeToolActionRequest, objectID string, objectKind string, eventType string, sourceErr error) map[string]any {
	if sourceErr != nil {
		return notionMirrorEffect("not_attempted", objectID, "source_mutation_failed")
	}
	if !cfg.NotionMirrorEnabled {
		return notionMirrorEffect("not_applicable", objectID, "")
	}
	rootID := firstNonEmpty(stringArg(input.Arguments, "mirror_root_id"), resolveNotionMirrorRootID(cfg, repo, input))
	if rootID == "" {
		return notionMirrorEffect("failed", objectID, "missing_mirror_root_id")
	}
	out, status, err := recordNotionMirrorDirtyObject(cfg, notionMirrorDirtyObjectRequest{
		RootID:         rootID,
		ObjectID:       objectID,
		ObjectKind:     objectKind,
		EventType:      eventType,
		EventTimestamp: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return map[string]any{
			"status":      "failed",
			"object_id":   objectID,
			"root_id":     rootID,
			"http_status": status,
			"error":       err.Error(),
		}
	}
	return map[string]any{
		"status":      "queued_dirty",
		"object_id":   out.ObjectID,
		"object_kind": out.ObjectKind,
		"root_id":     out.RootID,
		"queued":      out.Queued,
	}
}
