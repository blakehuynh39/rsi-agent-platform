package control

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
)

type notionMirrorDirtyObjectRequest struct {
	RootID         string `json:"root_id"`
	ObjectID       string `json:"object_id"`
	ObjectKind     string `json:"object_kind"`
	EventType      string `json:"event_type,omitempty"`
	EventTimestamp string `json:"event_timestamp,omitempty"`
}

type notionMirrorDirtyObjectResponse struct {
	RootID     string `json:"root_id"`
	ObjectID   string `json:"object_id"`
	ObjectKind string `json:"object_kind"`
	Queued     bool   `json:"queued"`
}

func recordNotionMirrorDirtyObject(cfg config.Config, payload notionMirrorDirtyObjectRequest) (notionMirrorDirtyObjectResponse, int, error) {
	rootID := normalizeNotionID(payload.RootID)
	objectID := normalizeNotionID(payload.ObjectID)
	objectKind := strings.ToLower(strings.TrimSpace(payload.ObjectKind))
	if objectKind == "" {
		objectKind = notionObjectKindFromEvent(payload.EventType)
	}
	if rootID == "" {
		return notionMirrorDirtyObjectResponse{}, http.StatusBadRequest, fmt.Errorf("root_id is required")
	}
	if objectID == "" {
		return notionMirrorDirtyObjectResponse{}, http.StatusBadRequest, fmt.Errorf("object_id is required")
	}
	switch objectKind {
	case companyknowledge.NotionObjectKindPage, companyknowledge.NotionObjectKindDatabase, companyknowledge.NotionObjectKindDataSource:
	default:
		return notionMirrorDirtyObjectResponse{}, http.StatusBadRequest, fmt.Errorf("unsupported notion object_kind %q", objectKind)
	}
	checkpoint, err := readNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, rootID)
	if err != nil {
		return notionMirrorDirtyObjectResponse{}, http.StatusInternalServerError, err
	}
	if checkpoint.DirtyObjects == nil {
		checkpoint.DirtyObjects = map[string]notionMirrorDirtyObject{}
	}
	key := objectKind + ":" + objectID
	checkpoint.DirtyObjects[key] = notionMirrorDirtyObject{
		ObjectKind:     objectKind,
		ObjectID:       objectID,
		EventType:      strings.TrimSpace(payload.EventType),
		EventTimestamp: strings.TrimSpace(payload.EventTimestamp),
		RecordedAt:     time.Now().UTC(),
	}
	checkpoint.RootID = rootID
	if err := writeNotionMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, checkpoint); err != nil {
		return notionMirrorDirtyObjectResponse{}, http.StatusInternalServerError, err
	}
	return notionMirrorDirtyObjectResponse{
		RootID:     rootID,
		ObjectID:   objectID,
		ObjectKind: objectKind,
		Queued:     true,
	}, http.StatusAccepted, nil
}

func notionObjectKindFromEvent(eventType string) string {
	eventType = strings.ToLower(strings.TrimSpace(eventType))
	if strings.Contains(eventType, "data_source") || strings.Contains(eventType, "data source") {
		return companyknowledge.NotionObjectKindDataSource
	}
	if strings.Contains(eventType, "database") {
		return companyknowledge.NotionObjectKindDatabase
	}
	return companyknowledge.NotionObjectKindPage
}
