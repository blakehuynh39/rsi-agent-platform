package control

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type sourceMirrorMessageWriteRequest struct {
	Record       storepkg.SourceMirrorRecord `json:"record"`
	Message      sourceMirrorMessagePayload  `json:"message"`
	LeaseSeconds int                         `json:"lease_seconds,omitempty"`
}

type sourceMirrorMessagePayload struct {
	Content   string         `json:"content"`
	PeerID    string         `json:"peer_id"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
}

type sourceMirrorMessageWriteResponse struct {
	Record          storepkg.SourceMirrorRecord `json:"record"`
	ShouldWrite     bool                        `json:"should_write"`
	Reason          string                      `json:"reason"`
	HonchoMessageID string                      `json:"honcho_message_id,omitempty"`
}

func writeSourceMirrorMessage(ctx context.Context, cfg config.Config, repo storepkg.Repository, req sourceMirrorMessageWriteRequest) (sourceMirrorMessageWriteResponse, int, error) {
	_ = ctx
	if strings.TrimSpace(cfg.HonchoBaseURL) == "" {
		return sourceMirrorMessageWriteResponse{}, 500, errors.New("RSI_HONCHO_BASE_URL is required for source mirror message writes")
	}
	mirrorStore, ok := repo.(storepkg.SourceMirrorWriteStore)
	if !ok {
		return sourceMirrorMessageWriteResponse{}, 500, errors.New("configured store does not support source mirror idempotency")
	}
	if strings.TrimSpace(req.Message.PeerID) == "" {
		return sourceMirrorMessageWriteResponse{}, 400, errors.New("message.peer_id is required")
	}
	record := req.Record
	record.Status = storepkg.SourceMirrorStatusPending
	if strings.TrimSpace(record.Environment) == "" {
		record.Environment = cfg.Environment
	}
	lease := time.Duration(req.LeaseSeconds) * time.Second
	claim, err := mirrorStore.ClaimSourceMirrorRecord(record, lease)
	if err != nil {
		return sourceMirrorMessageWriteResponse{}, 400, err
	}
	response := sourceMirrorMessageWriteResponse{
		Record:          claim.Record,
		ShouldWrite:     claim.ShouldWrite,
		Reason:          claim.Reason,
		HonchoMessageID: claim.Record.HonchoMessageID,
	}
	if !claim.ShouldWrite {
		return response, 200, nil
	}

	honcho := clients.NewHonchoClientWithAPIKey(cfg.HonchoBaseURL, cfg.HonchoAPIKey)
	if _, err := honcho.EnsureWorkspace(record.HonchoWorkspace, map[string]any{
		"source":      "rsi_company_knowledge",
		"environment": record.Environment,
	}); err != nil {
		_, _ = mirrorStore.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "ensure_workspace"})
		return sourceMirrorMessageWriteResponse{}, 502, err
	}
	if _, err := honcho.EnsureSession(record.HonchoWorkspace, record.HonchoSessionID, map[string]any{
		"source":             "source_mirror",
		"source_session_key": record.SourceSessionKey,
		"workspace":          record.Workspace,
		"environment":        record.Environment,
	}); err != nil {
		_, _ = mirrorStore.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "ensure_session"})
		return sourceMirrorMessageWriteResponse{}, 502, err
	}
	messages, err := honcho.CreateMessages(record.HonchoWorkspace, record.HonchoSessionID, []clients.HonchoMessageCreate{
		{
			Content:   req.Message.Content,
			PeerID:    req.Message.PeerID,
			Metadata:  mergeSourceMirrorMessageMetadata(record.Metadata, req.Message.Metadata),
			CreatedAt: req.Message.CreatedAt,
		},
	})
	if err != nil {
		_, _ = mirrorStore.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "create_message"})
		return sourceMirrorMessageWriteResponse{}, 502, err
	}
	if len(messages) != 1 || strings.TrimSpace(messages[0].ID) == "" {
		err := errors.New("honcho create message returned no stable message id")
		_, _ = mirrorStore.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "create_message"})
		return sourceMirrorMessageWriteResponse{}, 502, err
	}
	completed, err := mirrorStore.CompleteSourceMirrorRecord(record.SourceType, record.SourceKey, messages[0].ID, map[string]any{
		"honcho_message_id": messages[0].ID,
	})
	if err != nil {
		return sourceMirrorMessageWriteResponse{}, 500, err
	}
	response.Record = completed
	response.HonchoMessageID = completed.HonchoMessageID
	return response, 201, nil
}

func mergeSourceMirrorMessageMetadata(base map[string]any, overlay map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	for key, value := range overlay {
		out[key] = value
	}
	return out
}
