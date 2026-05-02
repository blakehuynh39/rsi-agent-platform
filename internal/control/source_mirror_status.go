package control

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type sourceMirrorStatusResponse struct {
	OK             bool                              `json:"ok"`
	StoreSupported bool                              `json:"store_supported"`
	Environment    string                            `json:"environment"`
	SourceTypes    map[string]sourceMirrorTypeStatus `json:"source_types"`
	Issues         []string                          `json:"issues,omitempty"`
	CheckedAt      time.Time                         `json:"checked_at"`
}

type sourceMirrorTypeStatus struct {
	SourceType      string                       `json:"source_type"`
	Counts          map[string]int               `json:"counts"`
	LatestComplete  *storepkg.SourceMirrorRecord `json:"latest_complete,omitempty"`
	LatestFailed    *storepkg.SourceMirrorRecord `json:"latest_failed,omitempty"`
	LatestPending   *storepkg.SourceMirrorRecord `json:"latest_pending,omitempty"`
	LatestStale     *storepkg.SourceMirrorRecord `json:"latest_stale,omitempty"`
	LatestUpdatedAt *time.Time                   `json:"latest_updated_at,omitempty"`
}

func sourceMirrorStatus(cfg config.Config, repo storepkg.Repository, requiredSourceTypes []string, limit int, maxAge time.Duration) (sourceMirrorStatusResponse, int, error) {
	response := sourceMirrorStatusResponse{
		OK:          true,
		Environment: strings.TrimSpace(cfg.Environment),
		SourceTypes: map[string]sourceMirrorTypeStatus{},
		CheckedAt:   time.Now().UTC(),
	}
	statusStore, ok := repo.(storepkg.SourceMirrorStatusStore)
	if !ok {
		response.OK = false
		response.Issues = append(response.Issues, "configured store does not support source mirror status")
		return response, http.StatusInternalServerError, nil
	}
	response.StoreSupported = true
	if limit <= 0 {
		limit = 500
	}
	records, err := statusStore.ListSourceMirrorRecords(requiredSourceTypes, limit)
	if err != nil {
		response.OK = false
		response.Issues = append(response.Issues, err.Error())
		return response, http.StatusInternalServerError, err
	}
	for _, sourceType := range requiredSourceTypes {
		sourceType = strings.TrimSpace(sourceType)
		if sourceType != "" {
			response.SourceTypes[sourceType] = sourceMirrorTypeStatus{SourceType: sourceType, Counts: map[string]int{}}
		}
	}
	for _, record := range records {
		sourceType := strings.TrimSpace(record.SourceType)
		if sourceType == "" {
			continue
		}
		status := response.SourceTypes[sourceType]
		if status.SourceType == "" {
			status = sourceMirrorTypeStatus{SourceType: sourceType, Counts: map[string]int{}}
		}
		if status.Counts == nil {
			status.Counts = map[string]int{}
		}
		status.Counts[strings.TrimSpace(record.Status)]++
		if status.LatestUpdatedAt == nil || record.UpdatedAt.After(*status.LatestUpdatedAt) {
			updatedAt := record.UpdatedAt
			status.LatestUpdatedAt = &updatedAt
		}
		recordCopy := record
		switch strings.TrimSpace(record.Status) {
		case storepkg.SourceMirrorStatusComplete:
			if status.LatestComplete == nil || record.UpdatedAt.After(status.LatestComplete.UpdatedAt) {
				status.LatestComplete = &recordCopy
			}
		case storepkg.SourceMirrorStatusFailed:
			if status.LatestFailed == nil || record.UpdatedAt.After(status.LatestFailed.UpdatedAt) {
				status.LatestFailed = &recordCopy
			}
		case storepkg.SourceMirrorStatusPending:
			if status.LatestPending == nil || record.UpdatedAt.After(status.LatestPending.UpdatedAt) {
				status.LatestPending = &recordCopy
			}
		case storepkg.SourceMirrorStatusStale:
			if status.LatestStale == nil || record.UpdatedAt.After(status.LatestStale.UpdatedAt) {
				status.LatestStale = &recordCopy
			}
		}
		response.SourceTypes[sourceType] = status
	}
	for _, sourceType := range requiredSourceTypes {
		sourceType = strings.TrimSpace(sourceType)
		if sourceType == "" {
			continue
		}
		status := response.SourceTypes[sourceType]
		if status.LatestComplete == nil {
			response.Issues = append(response.Issues, "source type "+sourceType+" has no completed mirror write")
			continue
		}
		if maxAge > 0 && response.CheckedAt.Sub(status.LatestComplete.UpdatedAt) > maxAge {
			response.Issues = append(response.Issues, "source type "+sourceType+" latest completed mirror write is older than "+maxAge.String())
		}
		if status.LatestFailed != nil && status.LatestFailed.UpdatedAt.After(status.LatestComplete.UpdatedAt) {
			response.Issues = append(response.Issues, "source type "+sourceType+" has a failed mirror write newer than latest completed write")
		}
	}
	response.OK = len(response.Issues) == 0
	statusCode := http.StatusOK
	if !response.OK {
		statusCode = http.StatusServiceUnavailable
	}
	return response, statusCode, nil
}

func parseSourceMirrorStatusQuery(values []string) []string {
	var out []string
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
	}
	return out
}

func parsePositiveIntQuery(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
