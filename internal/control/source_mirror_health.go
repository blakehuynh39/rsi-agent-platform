package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	sourceMirrorHealthMessageType  = "source_mirror_health_message"
	sourceMirrorHealthDocumentType = "source_mirror_health_document"
)

type sourceMirrorHealthReport struct {
	OK             bool                               `json:"ok"`
	Environment    string                             `json:"environment"`
	CheckpointRoot string                             `json:"checkpoint_root"`
	Checks         []sourceMirrorHealthCheck          `json:"checks"`
	MessageWrite   *sourceMirrorMessageWriteResponse  `json:"message_write,omitempty"`
	DocumentWrite  *sourceMirrorDocumentWriteResponse `json:"document_write,omitempty"`
}

type sourceMirrorHealthCheck struct {
	Name   string         `json:"name"`
	Status string         `json:"status"`
	Detail map[string]any `json:"detail,omitempty"`
	Error  string         `json:"error,omitempty"`
}

func RunSourceMirrorHealth(ctx context.Context, cfg config.Config, mirrorStore storepkg.SourceMirrorWriteStore) error {
	report, err := CheckSourceMirrorHealth(ctx, cfg, mirrorStore)
	raw, marshalErr := json.Marshal(report)
	if marshalErr == nil {
		log.Printf("source mirror health: %s", string(raw))
	}
	if err != nil {
		return err
	}
	return nil
}

func CheckSourceMirrorHealth(ctx context.Context, cfg config.Config, mirrorStore storepkg.SourceMirrorWriteStore) (sourceMirrorHealthReport, error) {
	report := sourceMirrorHealthReport{
		Environment:    strings.TrimSpace(cfg.Environment),
		CheckpointRoot: strings.TrimSpace(cfg.SourceMirrorCheckpointRoot),
	}
	addCheck := func(name string, detail map[string]any, err error) {
		check := sourceMirrorHealthCheck{Name: name, Status: "ok", Detail: detail}
		if err != nil {
			check.Status = "failed"
			check.Error = err.Error()
		}
		report.Checks = append(report.Checks, check)
	}

	addCheck("checkpoint_root", map[string]any{"path": report.CheckpointRoot}, checkSourceMirrorCheckpointRoot(report.CheckpointRoot))
	if cfg.SlackMirrorEnabled {
		addCheck("slack_auth", map[string]any{
			"channel_discovery": slackMirrorChannelDiscoveryMode(cfg),
			"allowlist":         cfg.SlackMirrorChannelAllowlist,
			"denylist":          cfg.SlackMirrorChannelDenylist,
		}, checkSlackMirrorAuth(ctx, cfg))
	}
	if cfg.NotionMirrorEnabled {
		addCheck("notion_roots", map[string]any{"allowlisted_roots": cfg.NotionMirrorAllowlist}, checkNotionMirrorRoots(ctx, cfg))
	}

	messageWrite, messageErr := sourceMirrorHealthMessageWrite(ctx, cfg, mirrorStore)
	if messageErr == nil {
		report.MessageWrite = &messageWrite
	}
	addCheck("honcho_message_write", map[string]any{"source_type": sourceMirrorHealthMessageType}, messageErr)

	documentWrite, documentErr := sourceMirrorHealthDocumentWrite(ctx, cfg, mirrorStore)
	if documentErr == nil {
		report.DocumentWrite = &documentWrite
	}
	addCheck("honcho_document_write", map[string]any{"source_type": sourceMirrorHealthDocumentType}, documentErr)

	var failures []string
	for _, check := range report.Checks {
		if check.Status == "failed" {
			failures = append(failures, check.Name+": "+check.Error)
		}
	}
	report.OK = len(failures) == 0
	if len(failures) > 0 {
		return report, errors.New(strings.Join(failures, "; "))
	}
	return report, nil
}

func checkSourceMirrorCheckpointRoot(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return errors.New("RSI_SOURCE_MIRROR_CHECKPOINT_ROOT is required")
	}
	dir := filepath.Join(root, "health")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "write-check.json")
	payload := []byte(`{"ok":true}` + "\n")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func checkSlackMirrorAuth(ctx context.Context, cfg config.Config) error {
	api := slackapi.New(cfg.SlackBotToken)
	auth, err := api.AuthTestContext(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(auth.TeamID) == "" {
		return errors.New("slack auth.test returned empty team_id")
	}
	mode := slackMirrorChannelDiscoveryMode(cfg)
	if mode == "joined" || mode == "joined_public" {
		channels, err := slackMirrorChannels(ctx, cfg, api)
		if err != nil {
			return err
		}
		if len(channels) == 0 {
			return errors.New("slack joined-channel discovery returned no mirrorable channels")
		}
	}
	return nil
}

func checkNotionMirrorRoots(ctx context.Context, cfg config.Config) error {
	api := clients.NewNotionClientWithConfig(clients.NotionClientOptions{
		BaseURL:           cfg.NotionAPIBaseURL,
		Token:             cfg.NotionToken,
		Version:           cfg.NotionAPIVersion,
		RequestsPerSecond: cfg.NotionMirrorRequestsPerSecond,
		MaxRetries:        cfg.NotionMirrorMaxRetries,
		RetryBaseDelay:    cfg.NotionMirrorRetryBaseDelay,
	})
	for _, rootID := range cfg.NotionMirrorAllowlist {
		rootID = normalizeNotionID(rootID)
		if rootID == "" {
			continue
		}
		if page, err := api.RetrievePage(ctx, rootID); err == nil {
			if page.Archived || page.InTrash {
				return fmt.Errorf("notion page root %s is stale: archived=%t in_trash=%t", rootID, page.Archived, page.InTrash)
			}
			continue
		} else if !isNotionNotFound(err) && !isNotionPageEndpointTypeMismatch(err) {
			return fmt.Errorf("retrieve notion page root=%s: %w", rootID, err)
		}
		if database, err := api.RetrieveDatabase(ctx, rootID); err == nil {
			if database.Archived || database.InTrash {
				return fmt.Errorf("notion database root %s is stale: archived=%t in_trash=%t", rootID, database.Archived, database.InTrash)
			}
			continue
		} else if !isNotionNotFound(err) && !isNotionDatabaseEndpointTypeMismatch(err) {
			return fmt.Errorf("retrieve notion database root=%s: %w", rootID, err)
		}
		return fmt.Errorf("notion allowlist root %s is neither a visible page nor a visible database", rootID)
	}
	return nil
}

func sourceMirrorHealthMessageWrite(ctx context.Context, cfg config.Config, mirrorStore storepkg.SourceMirrorWriteStore) (sourceMirrorMessageWriteResponse, error) {
	revision := sourceMirrorHealthRevision()
	sourceKey := sourceMirrorHealthSourceKey(cfg, "message")
	sessionKey := sourceMirrorHealthSessionKey(cfg)
	record := storepkg.SourceMirrorRecord{
		SourceType:       sourceMirrorHealthMessageType,
		SourceKey:        sourceKey,
		Workspace:        "source_mirror_health",
		Environment:      strings.TrimSpace(cfg.Environment),
		SourceSessionKey: sessionKey,
		HonchoWorkspace:  companyknowledge.HonchoCompatibleName("workspace", firstNonEmpty(cfg.HonchoWorkspaceID, "rsi_company_knowledge")),
		HonchoSessionID:  companyknowledge.HonchoCompatibleName("health", sessionKey),
		SourceRevision:   revision,
		Metadata: map[string]any{
			"source":          "source_mirror_health",
			"source_key":      sourceKey,
			"source_revision": revision,
		},
	}
	out, _, err := writeSourceMirrorMessage(ctx, cfg, mirrorStore, sourceMirrorMessageWriteRequest{
		Record: record,
		Message: sourceMirrorMessagePayload{
			Content: "RSI source mirror health message write check.",
			PeerID:  "source_mirror_health",
			Metadata: map[string]any{
				"source":          "source_mirror_health",
				"source_revision": revision,
			},
		},
		LeaseSeconds: 60,
	})
	return out, err
}

func sourceMirrorHealthDocumentWrite(ctx context.Context, cfg config.Config, mirrorStore storepkg.SourceMirrorWriteStore) (sourceMirrorDocumentWriteResponse, error) {
	revision := sourceMirrorHealthRevision()
	sourceKey := sourceMirrorHealthSourceKey(cfg, "document")
	sessionKey := sourceMirrorHealthSessionKey(cfg)
	record := storepkg.SourceMirrorRecord{
		SourceType:       sourceMirrorHealthDocumentType,
		SourceKey:        sourceKey,
		Workspace:        "source_mirror_health",
		Environment:      strings.TrimSpace(cfg.Environment),
		SourceSessionKey: sessionKey,
		HonchoWorkspace:  companyknowledge.HonchoCompatibleName("workspace", firstNonEmpty(cfg.HonchoWorkspaceID, "rsi_company_knowledge")),
		HonchoSessionID:  companyknowledge.HonchoCompatibleName("health", sessionKey),
		SourceRevision:   revision,
		Metadata: map[string]any{
			"source":          "source_mirror_health",
			"source_key":      sourceKey,
			"source_revision": revision,
		},
	}
	out, _, err := writeSourceMirrorDocument(ctx, cfg, mirrorStore, sourceMirrorDocumentWriteRequest{
		Record: record,
		Document: sourceMirrorDocumentPayload{
			Content:    "RSI source mirror health document write check.",
			ObserverID: "source_mirror_health",
			ObservedID: "rsi_company_knowledge",
			Metadata: map[string]any{
				"source":          "source_mirror_health",
				"source_revision": revision,
			},
		},
		LeaseSeconds: 60,
	})
	return out, err
}

func sourceMirrorHealthRevision() string {
	return "health:" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
}

func sourceMirrorHealthSourceKey(cfg config.Config, kind string) string {
	return "source_mirror_health:" + firstNonEmpty(strings.TrimSpace(cfg.Environment), "unknown") + ":" + kind
}

func sourceMirrorHealthSessionKey(cfg config.Config) string {
	return "source_mirror_health:" + firstNonEmpty(strings.TrimSpace(cfg.Environment), "unknown")
}
