package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type slackMirrorCheckpoint struct {
	ChannelID         string    `json:"channel_id"`
	WorkspaceID       string    `json:"workspace_id"`
	LastMirroredTS    string    `json:"last_mirrored_ts"`
	LastCompletedAt   time.Time `json:"last_completed_at"`
	LastMessageCount  int       `json:"last_message_count"`
	LastThreadCount   int       `json:"last_thread_count"`
	LastHonchoSession string    `json:"last_honcho_session,omitempty"`
}

func RunSlackMirror(ctx context.Context, cfg config.Config, state store.Store) error {
	if !cfg.SlackMirrorEnabled {
		return errors.New("slack mirror is disabled")
	}
	mirrorStore, ok := state.(store.SourceMirrorWriteStore)
	if !ok {
		return errors.New("configured store does not support source mirror idempotency")
	}
	api := slackapi.New(cfg.SlackBotToken)
	auth, err := api.AuthTestContext(ctx)
	if err != nil {
		return fmt.Errorf("slack auth test failed: %w", err)
	}
	workspaceID := strings.TrimSpace(auth.TeamID)
	if workspaceID == "" {
		return errors.New("slack auth test returned empty team_id")
	}
	honcho := clients.NewHonchoClientWithAPIKey(cfg.HonchoBaseURL, cfg.HonchoAPIKey)
	mirror := companyknowledge.NewSlackMirror(mirrorStore, honcho, companyknowledge.SlackMirrorOptions{
		Environment:     cfg.Environment,
		HonchoWorkspace: cfg.HonchoWorkspaceID,
	})
	channels := uniqueNonEmpty(cfg.SlackMirrorChannelAllowlist)
	if len(channels) == 0 {
		return errors.New("RSI_SLACK_MIRROR_CHANNEL_ALLOWLIST is empty")
	}
	sort.Strings(channels)
	for _, channelID := range channels {
		if err := mirrorSlackChannel(ctx, cfg, api, mirror, workspaceID, channelID); err != nil {
			return err
		}
	}
	return nil
}

func mirrorSlackChannel(ctx context.Context, cfg config.Config, api *slackapi.Client, mirror *companyknowledge.SlackMirror, workspaceID string, channelID string) error {
	checkpoint, err := readSlackMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, channelID)
	if err != nil {
		return err
	}
	oldest := strings.TrimSpace(checkpoint.LastMirroredTS)
	cursor := ""
	latestSeen := oldest
	messageCount := 0
	threadCount := 0
	for {
		resp, err := api.GetConversationHistoryContext(ctx, &slackapi.GetConversationHistoryParameters{
			ChannelID:          channelID,
			Cursor:             cursor,
			Limit:              200,
			Oldest:             oldest,
			Inclusive:          false,
			IncludeAllMetadata: true,
		})
		if err != nil {
			return fmt.Errorf("read slack history channel=%s: %w", channelID, err)
		}
		for _, msg := range reverseSlackMessages(resp.Messages) {
			if strings.TrimSpace(msg.Timestamp) == "" || shouldSkipSlackMirrorMessage(msg) {
				continue
			}
			input := slackInputFromMessage(workspaceID, channelID, msg, "")
			if msg.ReplyCount > 0 && strings.TrimSpace(input.ThreadTS) == "" {
				input.ThreadTS = msg.Timestamp
			}
			result, err := mirror.IngestMessage(ctx, input)
			if err != nil {
				return fmt.Errorf("mirror slack message channel=%s ts=%s: %w", channelID, msg.Timestamp, err)
			}
			messageCount++
			if result.HonchoSessionID != "" {
				checkpoint.LastHonchoSession = result.HonchoSessionID
			}
			if compareSlackTS(msg.Timestamp, latestSeen) > 0 {
				latestSeen = msg.Timestamp
			}
			if msg.ReplyCount > 0 {
				seenReplies, err := mirrorSlackThread(ctx, api, mirror, workspaceID, channelID, msg.Timestamp)
				if err != nil {
					return err
				}
				threadCount++
				messageCount += seenReplies
			}
		}
		cursor = strings.TrimSpace(resp.ResponseMetaData.NextCursor)
		if !resp.HasMore || cursor == "" {
			break
		}
	}
	checkpoint.ChannelID = channelID
	checkpoint.WorkspaceID = workspaceID
	checkpoint.LastMirroredTS = latestSeen
	checkpoint.LastCompletedAt = time.Now().UTC()
	checkpoint.LastMessageCount = messageCount
	checkpoint.LastThreadCount = threadCount
	if err := writeSlackMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, checkpoint); err != nil {
		return err
	}
	log.Printf("slack mirror channel=%s messages=%d threads=%d last_ts=%s", channelID, messageCount, threadCount, latestSeen)
	return nil
}

func mirrorSlackThread(ctx context.Context, api *slackapi.Client, mirror *companyknowledge.SlackMirror, workspaceID string, channelID string, threadTS string) (int, error) {
	cursor := ""
	count := 0
	for {
		messages, hasMore, nextCursor, err := api.GetConversationRepliesContext(ctx, &slackapi.GetConversationRepliesParameters{
			ChannelID:          channelID,
			Timestamp:          threadTS,
			Cursor:             cursor,
			Limit:              200,
			Inclusive:          true,
			IncludeAllMetadata: true,
		})
		if err != nil {
			return count, fmt.Errorf("read slack replies channel=%s thread=%s: %w", channelID, threadTS, err)
		}
		for _, msg := range messages {
			if strings.TrimSpace(msg.Timestamp) == "" || shouldSkipSlackMirrorMessage(msg) {
				continue
			}
			input := slackInputFromMessage(workspaceID, channelID, msg, "")
			input.ThreadTS = threadTS
			if _, err := mirror.IngestMessage(ctx, input); err != nil {
				return count, fmt.Errorf("mirror slack reply channel=%s thread=%s ts=%s: %w", channelID, threadTS, msg.Timestamp, err)
			}
			count++
		}
		cursor = strings.TrimSpace(nextCursor)
		if !hasMore || cursor == "" {
			return count, nil
		}
	}
}

func slackInputFromMessage(workspaceID string, channelID string, msg slackapi.Message, eventID string) companyknowledge.SlackMessageInput {
	files := make([]companyknowledge.SlackFileMetadata, 0, len(msg.Files))
	for _, file := range msg.Files {
		files = append(files, companyknowledge.SlackFileMetadata{
			ID:        file.ID,
			Name:      file.Name,
			Title:     file.Title,
			MimeType:  file.Mimetype,
			FileType:  file.Filetype,
			Size:      file.Size,
			Permalink: file.Permalink,
		})
	}
	editedTS := ""
	if msg.Edited != nil {
		editedTS = msg.Edited.Timestamp
	}
	return companyknowledge.SlackMessageInput{
		WorkspaceID: workspaceID,
		ChannelID:   channelID,
		TS:          msg.Timestamp,
		ThreadTS:    msg.ThreadTimestamp,
		UserID:      msg.User,
		BotID:       msg.BotID,
		Username:    msg.Username,
		Text:        msg.Text,
		EditedTS:    editedTS,
		EventID:     eventID,
		Permalink:   msg.Permalink,
		ReplyCount:  msg.ReplyCount,
		Files:       files,
		CreatedAt:   companyknowledge.SlackTimestampToTime(msg.Timestamp),
	}
}

func shouldSkipSlackMirrorMessage(msg slackapi.Message) bool {
	if msg.Hidden || strings.TrimSpace(msg.DeletedTimestamp) != "" {
		return true
	}
	switch strings.TrimSpace(msg.SubType) {
	case "", slackapi.MsgSubTypeBotMessage, slackapi.MsgSubTypeFileShare, slackapi.MsgSubTypeMeMessage, slackapi.MsgSubTypeThreadBroadcast:
		return false
	default:
		return true
	}
}

func reverseSlackMessages(input []slackapi.Message) []slackapi.Message {
	out := make([]slackapi.Message, len(input))
	for i := range input {
		out[i] = input[len(input)-1-i]
	}
	return out
}

func readSlackMirrorCheckpoint(root string, channelID string) (slackMirrorCheckpoint, error) {
	path := slackMirrorCheckpointPath(root, channelID)
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return slackMirrorCheckpoint{ChannelID: channelID}, nil
	}
	if err != nil {
		return slackMirrorCheckpoint{}, err
	}
	var checkpoint slackMirrorCheckpoint
	if err := json.Unmarshal(raw, &checkpoint); err != nil {
		return slackMirrorCheckpoint{}, fmt.Errorf("decode slack mirror checkpoint %s: %w", path, err)
	}
	return checkpoint, nil
}

func writeSlackMirrorCheckpoint(root string, checkpoint slackMirrorCheckpoint) error {
	path := slackMirrorCheckpointPath(root, checkpoint.ChannelID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func slackMirrorCheckpointPath(root string, channelID string) string {
	return filepath.Join(strings.TrimSpace(root), "slack", sanitizePathPart(channelID)+".json")
}

func sanitizePathPart(value string) string {
	return companyknowledge.HonchoCompatibleName("path", value)
}

func compareSlackTS(a string, b string) int {
	at := companyknowledge.SlackTimestampToTime(a)
	bt := companyknowledge.SlackTimestampToTime(b)
	if at.IsZero() && bt.IsZero() {
		return strings.Compare(a, b)
	}
	if at.After(bt) {
		return 1
	}
	if at.Before(bt) {
		return -1
	}
	return 0
}

func uniqueNonEmpty(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
