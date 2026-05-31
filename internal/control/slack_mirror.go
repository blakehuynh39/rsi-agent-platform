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
	BackfillComplete  bool      `json:"backfill_complete"`
	BackfillBeforeTS  string    `json:"backfill_before_ts,omitempty"`
	BackfillOldestTS  string    `json:"backfill_oldest_ts,omitempty"`
	LastProgressAt    time.Time `json:"last_progress_at,omitempty"`
	LastCompletedAt   time.Time `json:"last_completed_at"`
	LastMessageCount  int       `json:"last_message_count"`
	LastThreadCount   int       `json:"last_thread_count"`
	LastHonchoSession string    `json:"last_honcho_session,omitempty"`
}

type slackMirrorChannelLister interface {
	GetConversationsContext(ctx context.Context, params *slackapi.GetConversationsParameters) ([]slackapi.Channel, string, error)
}

type slackMirrorChannelMetadata struct {
	ChannelID       string
	ChannelType     string
	ChannelPrivate  bool
	ChannelIM       bool
	InfoChecked     bool
	InfoError       string
	PolicyUntrusted bool
}

func RunSlackMirror(ctx context.Context, cfg config.Config, mirrorStore store.SourceMirrorWriteStore) error {
	if !cfg.SlackMirrorEnabled {
		return errors.New("slack mirror is disabled")
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
	channels, err := slackMirrorChannels(ctx, cfg, api)
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		return errors.New("slack mirror found no channels to mirror")
	}
	for _, channelID := range channels {
		channelMetadata := slackMirrorChannelMetadataForChannel(ctx, cfg, api, channelID)
		if err := mirrorSlackChannel(ctx, cfg, api, mirrorStore, mirror, workspaceID, channelID, channelMetadata); err != nil {
			return err
		}
	}
	return nil
}

func slackMirrorChannels(ctx context.Context, cfg config.Config, api slackMirrorChannelLister) ([]string, error) {
	discovery := slackMirrorChannelDiscoveryMode(cfg)
	denylist := slackMirrorChannelDenylist(cfg)
	switch discovery {
	case "explicit":
		return filterDeniedChannels(uniqueNonEmpty(cfg.SlackMirrorChannelAllowlist), denylist), nil
	case "joined":
		channels, err := discoverJoinedSlackMirrorChannels(ctx, api, true)
		if err != nil {
			return nil, err
		}
		return filterDeniedChannels(uniqueNonEmpty(channels), denylist), nil
	case "joined_public":
		channels, err := discoverJoinedSlackMirrorChannels(ctx, api, false)
		if err != nil {
			return nil, err
		}
		return filterDeniedChannels(uniqueNonEmpty(channels), denylist), nil
	default:
		return nil, fmt.Errorf("unsupported RSI_SLACK_MIRROR_CHANNEL_DISCOVERY %q", cfg.SlackMirrorChannelDiscovery)
	}
}

func discoverJoinedSlackMirrorChannels(ctx context.Context, api slackMirrorChannelLister, includePrivate bool) ([]string, error) {
	if api == nil {
		return nil, errors.New("slack channel discovery requires Slack API client")
	}
	types := []string{"public_channel"}
	if includePrivate {
		types = append(types, "private_channel")
	}
	var out []string
	cursor := ""
	for {
		channels, nextCursor, err := api.GetConversationsContext(ctx, &slackapi.GetConversationsParameters{
			Cursor:          cursor,
			ExcludeArchived: true,
			Limit:           200,
			Types:           types,
		})
		if err != nil {
			return nil, fmt.Errorf("discover joined slack mirror channels: %w", err)
		}
		for _, channel := range channels {
			if strings.TrimSpace(channel.ID) == "" || !channel.IsMember || channel.IsArchived {
				continue
			}
			if !includePrivate && channel.IsPrivate {
				continue
			}
			out = append(out, strings.TrimSpace(channel.ID))
		}
		cursor = strings.TrimSpace(nextCursor)
		if cursor == "" {
			break
		}
	}
	return uniqueNonEmpty(out), nil
}

func slackMirrorChannelMetadataForChannel(ctx context.Context, cfg config.Config, api *slackapi.Client, channelID string) slackMirrorChannelMetadata {
	metadata := slackMirrorChannelMetadata{ChannelID: strings.TrimSpace(channelID)}
	discovery := slackMirrorChannelDiscoveryMode(cfg)
	if discovery == "joined_public" {
		metadata.ChannelType = "public_channel"
		metadata.ChannelPrivate = false
		metadata.ChannelIM = false
		metadata.InfoChecked = true
		return metadata
	}
	if api == nil {
		metadata.ChannelType = "unknown"
		metadata.ChannelPrivate = true
		metadata.PolicyUntrusted = true
		metadata.InfoError = "slack_api_client_missing"
		return metadata
	}
	info, err := api.GetConversationInfoContext(ctx, &slackapi.GetConversationInfoInput{ChannelID: metadata.ChannelID})
	if err != nil {
		return slackMirrorChannelMetadataFromInfoError(cfg, metadata.ChannelID, err)
	}
	return slackMirrorChannelMetadataFromSlackChannel(metadata.ChannelID, info)
}

func slackMirrorChannelMetadataFromInfoError(cfg config.Config, channelID string, err error) slackMirrorChannelMetadata {
	metadata := slackMirrorChannelMetadata{
		ChannelID:       strings.TrimSpace(channelID),
		ChannelType:     "unknown",
		ChannelPrivate:  true,
		PolicyUntrusted: true,
	}
	if err != nil {
		metadata.InfoError = err.Error()
	}
	if slackMirrorChannelDiscoveryMode(cfg) == "explicit" &&
		slackMirrorChannelAllowedByConfig(cfg, metadata.ChannelID) &&
		slackMirrorLooksLikePublicChannelID(metadata.ChannelID) &&
		slackMirrorChannelInfoMissingScope(err) {
		metadata.ChannelType = "explicit_allowlisted_public"
		metadata.ChannelPrivate = false
		metadata.ChannelIM = false
		metadata.PolicyUntrusted = false
	}
	return metadata
}

func slackMirrorLooksLikePublicChannelID(channelID string) bool {
	channelID = strings.TrimSpace(channelID)
	return strings.HasPrefix(channelID, "C")
}

func slackMirrorChannelInfoMissingScope(err error) bool {
	if err == nil {
		return false
	}
	var slackErr slackapi.SlackErrorResponse
	if errors.As(err, &slackErr) {
		return slackErr.Err == "missing_scope"
	}
	return strings.Contains(err.Error(), "missing_scope")
}

func slackMirrorChannelMetadataFromSlackChannel(channelID string, channel *slackapi.Channel) slackMirrorChannelMetadata {
	metadata := slackMirrorChannelMetadata{ChannelID: strings.TrimSpace(channelID), InfoChecked: true}
	if channel == nil {
		metadata.ChannelType = "unknown"
		metadata.ChannelPrivate = true
		metadata.PolicyUntrusted = true
		metadata.InfoError = "slack_channel_info_missing"
		return metadata
	}
	metadata.ChannelPrivate = channel.IsPrivate
	metadata.ChannelIM = channel.IsIM || channel.IsMpIM
	switch {
	case metadata.ChannelIM:
		metadata.ChannelType = "im"
	case metadata.ChannelPrivate:
		metadata.ChannelType = "private_channel"
	default:
		metadata.ChannelType = "public_channel"
	}
	return metadata
}

func slackMirrorChannelDiscoveryMode(cfg config.Config) string {
	mode := strings.ToLower(strings.TrimSpace(cfg.SlackMirrorChannelDiscovery))
	if mode == "" {
		return "joined"
	}
	return mode
}

func slackMirrorChannelDenylist(cfg config.Config) map[string]struct{} {
	out := map[string]struct{}{}
	for _, channelID := range cfg.SlackMirrorChannelDenylist {
		channelID = strings.TrimSpace(channelID)
		if channelID != "" {
			out[channelID] = struct{}{}
		}
	}
	return out
}

func filterDeniedChannels(channels []string, denylist map[string]struct{}) []string {
	if len(denylist) == 0 {
		sort.Strings(channels)
		return channels
	}
	out := make([]string, 0, len(channels))
	for _, channelID := range channels {
		if _, denied := denylist[channelID]; denied {
			continue
		}
		out = append(out, channelID)
	}
	sort.Strings(out)
	return out
}

func slackMirrorChannelAllowedByConfig(cfg config.Config, channelID string) bool {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return false
	}
	if _, denied := slackMirrorChannelDenylist(cfg)[channelID]; denied {
		return false
	}
	if mode := slackMirrorChannelDiscoveryMode(cfg); mode == "joined" || mode == "joined_public" {
		return true
	}
	for _, item := range cfg.SlackMirrorChannelAllowlist {
		if strings.TrimSpace(item) == channelID {
			return true
		}
	}
	return false
}

func shouldPublishSlackWikiSource(result companyknowledge.SlackMirrorResult) bool {
	return !result.Skipped
}

type slackWikiPublishBatch struct {
	cfg     config.Config
	repo    store.SourceMirrorWriteStore
	sources map[string]store.CompanyWikiSourceRevisionResult
}

func newSlackWikiPublishBatch(cfg config.Config, repo store.SourceMirrorWriteStore) *slackWikiPublishBatch {
	return &slackWikiPublishBatch{
		cfg:     cfg,
		repo:    repo,
		sources: map[string]store.CompanyWikiSourceRevisionResult{},
	}
}

func (b *slackWikiPublishBatch) record(ctx context.Context, input companyknowledge.SlackMessageInput) error {
	if b == nil {
		return nil
	}
	result, err := companyknowledge.RecordWikiSourceRevision(ctx, b.cfg, b.repo, companyknowledge.SlackWikiSourceRevisionInput(input))
	if err != nil {
		return err
	}
	if result.Skipped {
		return nil
	}
	if documentID := strings.TrimSpace(result.Source.Document.ID); documentID != "" {
		b.sources[documentID] = result.Source
	}
	return nil
}

func (b *slackWikiPublishBatch) publish(ctx context.Context) error {
	if b == nil || len(b.sources) == 0 {
		return nil
	}
	documentIDs := make([]string, 0, len(b.sources))
	for documentID := range b.sources {
		documentIDs = append(documentIDs, documentID)
	}
	sort.Strings(documentIDs)
	for _, documentID := range documentIDs {
		source := b.sources[documentID]
		if source.Changed {
			if _, _, err := companyknowledge.EnqueueWikiCompileItemForSource(ctx, b.cfg, b.repo, source); err != nil {
				return fmt.Errorf("enqueue slack wiki compile document=%s: %w", documentID, err)
			}
		}
		if strings.EqualFold(strings.TrimSpace(b.cfg.CompanyWikiSourcePageMode), "off") {
			delete(b.sources, documentID)
			continue
		}
		if _, err := companyknowledge.PublishWikiSourceDocument(ctx, b.cfg, b.repo, source); err != nil {
			return fmt.Errorf("publish slack wiki source document=%s: %w", documentID, err)
		}
		delete(b.sources, documentID)
	}
	return nil
}

func mirrorSlackChannel(ctx context.Context, cfg config.Config, api *slackapi.Client, mirrorStore store.SourceMirrorWriteStore, mirror *companyknowledge.SlackMirror, workspaceID string, channelID string, channelMetadata slackMirrorChannelMetadata) error {
	checkpoint, err := readSlackMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, channelID)
	if err != nil {
		return err
	}
	oldest, latest, mode := slackMirrorHistoryWindow(checkpoint)
	cursor := ""
	latestSeen := strings.TrimSpace(checkpoint.LastMirroredTS)
	messageCount := 0
	threadCount := 0
	log.Printf("slack mirror channel=%s mode=%s oldest=%s latest=%s", channelID, mode, oldest, latest)
	for {
		resp, err := api.GetConversationHistoryContext(ctx, &slackapi.GetConversationHistoryParameters{
			ChannelID:          channelID,
			Cursor:             cursor,
			Limit:              200,
			Latest:             latest,
			Oldest:             oldest,
			Inclusive:          false,
			IncludeAllMetadata: true,
		})
		if err != nil {
			return fmt.Errorf("read slack history channel=%s: %w", channelID, err)
		}
		wikiBatch := newSlackWikiPublishBatch(cfg, mirrorStore)
		pageOldestTS, pageNewestTS := slackMirrorMessageTimestampBounds(resp.Messages)
		for _, msg := range reverseSlackMessages(resp.Messages) {
			if strings.TrimSpace(msg.Timestamp) == "" || shouldSkipSlackMirrorMessage(msg) {
				continue
			}
			input := slackInputFromMessage(workspaceID, channelID, msg, "")
			applySlackMirrorPolicyMetadata(&input, cfg, channelMetadata)
			if msg.ReplyCount > 0 && strings.TrimSpace(input.ThreadTS) == "" {
				input.ThreadTS = msg.Timestamp
			}
			result, err := mirror.IngestMessage(ctx, input)
			if err != nil {
				return fmt.Errorf("mirror slack message channel=%s ts=%s: %w", channelID, msg.Timestamp, err)
			}
			if shouldPublishSlackWikiSource(result) {
				if err := wikiBatch.record(ctx, input); err != nil {
					return fmt.Errorf("record slack wiki source channel=%s ts=%s: %w", channelID, msg.Timestamp, err)
				}
			}
			messageCount++
			if result.HonchoSessionID != "" {
				checkpoint.LastHonchoSession = result.HonchoSessionID
			}
			if msg.ReplyCount > 0 {
				seenReplies, err := mirrorSlackThread(ctx, cfg, api, mirror, wikiBatch, workspaceID, channelID, msg.Timestamp, channelMetadata)
				if err != nil {
					return err
				}
				threadCount++
				messageCount += seenReplies
			}
		}
		if err := wikiBatch.publish(ctx); err != nil {
			return err
		}
		if compareSlackTS(pageNewestTS, latestSeen) > 0 {
			latestSeen = pageNewestTS
		}
		cursor = strings.TrimSpace(resp.ResponseMetaData.NextCursor)
		if pageOldestTS != "" || pageNewestTS != "" {
			updateSlackMirrorCheckpointProgress(&checkpoint, workspaceID, channelID, mode, latestSeen, pageOldestTS, messageCount, threadCount, false)
			if err := writeSlackMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, checkpoint); err != nil {
				return err
			}
			log.Printf(
				"slack mirror progress channel=%s mode=%s page_messages=%d total_messages=%d total_threads=%d page_oldest=%s page_newest=%s next_cursor=%t",
				channelID,
				mode,
				len(resp.Messages),
				messageCount,
				threadCount,
				pageOldestTS,
				pageNewestTS,
				cursor != "",
			)
		}
		if !resp.HasMore || cursor == "" {
			break
		}
	}
	if latestSeen == "" {
		latestSeen = slackMirrorTimestamp(time.Now().UTC())
	}
	updateSlackMirrorCheckpointProgress(&checkpoint, workspaceID, channelID, mode, latestSeen, "", messageCount, threadCount, true)
	if err := writeSlackMirrorCheckpoint(cfg.SourceMirrorCheckpointRoot, checkpoint); err != nil {
		return err
	}
	log.Printf("slack mirror channel=%s mode=%s complete messages=%d threads=%d last_ts=%s backfill_complete=%t", channelID, mode, messageCount, threadCount, latestSeen, checkpoint.BackfillComplete)
	return nil
}

func mirrorSlackThread(ctx context.Context, cfg config.Config, api *slackapi.Client, mirror *companyknowledge.SlackMirror, wikiBatch *slackWikiPublishBatch, workspaceID string, channelID string, threadTS string, channelMetadata slackMirrorChannelMetadata) (int, error) {
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
			applySlackMirrorPolicyMetadata(&input, cfg, channelMetadata)
			input.ThreadTS = threadTS
			result, err := mirror.IngestMessage(ctx, input)
			if err != nil {
				return count, fmt.Errorf("mirror slack reply channel=%s thread=%s ts=%s: %w", channelID, threadTS, msg.Timestamp, err)
			}
			if shouldPublishSlackWikiSource(result) {
				if err := wikiBatch.record(ctx, input); err != nil {
					return count, fmt.Errorf("record slack thread wiki source channel=%s thread=%s ts=%s: %w", channelID, threadTS, msg.Timestamp, err)
				}
			}
			count++
		}
		cursor = strings.TrimSpace(nextCursor)
		if !hasMore || cursor == "" {
			break
		}
	}
	return count, nil
}

func applySlackMirrorPolicyMetadata(input *companyknowledge.SlackMessageInput, cfg config.Config, channelMetadata slackMirrorChannelMetadata) {
	if input == nil {
		return
	}
	discovery := slackMirrorChannelDiscoveryMode(cfg)
	input.MirrorDiscovery = discovery
	input.MirrorDenied = !slackMirrorChannelAllowedByConfig(cfg, input.ChannelID)
	input.MirrorAllowed = !input.MirrorDenied
	if strings.TrimSpace(channelMetadata.ChannelID) != "" && strings.TrimSpace(channelMetadata.ChannelID) == strings.TrimSpace(input.ChannelID) {
		input.ChannelType = channelMetadata.ChannelType
		input.ChannelPrivate = channelMetadata.ChannelPrivate
		input.ChannelIM = channelMetadata.ChannelIM
		if channelMetadata.PolicyUntrusted || channelMetadata.ChannelPrivate || channelMetadata.ChannelIM {
			input.MirrorDenied = true
			input.MirrorAllowed = false
		}
		if channelMetadata.InfoChecked || strings.TrimSpace(channelMetadata.InfoError) != "" {
			if input.Raw == nil {
				input.Raw = map[string]any{}
			}
			input.Raw["channel_info_checked"] = channelMetadata.InfoChecked
			if strings.TrimSpace(channelMetadata.InfoError) != "" {
				input.Raw["channel_info_error"] = channelMetadata.InfoError
			}
		}
	}
	if strings.TrimSpace(input.ChannelType) == "" {
		if discovery == "joined_public" {
			input.ChannelType = "public_channel"
			input.ChannelPrivate = false
			input.ChannelIM = false
		} else {
			input.ChannelType = "unknown"
			input.ChannelPrivate = true
			input.MirrorDenied = true
			input.MirrorAllowed = false
			if input.Raw == nil {
				input.Raw = map[string]any{}
			}
			input.Raw["channel_info_error"] = "missing_channel_metadata"
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

func slackMirrorHistoryWindow(checkpoint slackMirrorCheckpoint) (oldest string, latest string, mode string) {
	if slackMirrorCheckpointBackfillComplete(checkpoint) {
		return strings.TrimSpace(checkpoint.LastMirroredTS), "", "incremental"
	}
	return "", strings.TrimSpace(checkpoint.BackfillBeforeTS), "backfill"
}

func slackMirrorCheckpointBackfillComplete(checkpoint slackMirrorCheckpoint) bool {
	if checkpoint.BackfillComplete {
		return true
	}
	return strings.TrimSpace(checkpoint.LastMirroredTS) != "" && strings.TrimSpace(checkpoint.BackfillBeforeTS) == "" && !checkpoint.LastCompletedAt.IsZero()
}

func updateSlackMirrorCheckpointProgress(checkpoint *slackMirrorCheckpoint, workspaceID string, channelID string, mode string, latestSeen string, pageOldestTS string, messageCount int, threadCount int, completed bool) {
	now := time.Now().UTC()
	checkpoint.ChannelID = channelID
	checkpoint.WorkspaceID = workspaceID
	checkpoint.LastMessageCount = messageCount
	checkpoint.LastThreadCount = threadCount
	checkpoint.LastProgressAt = now
	if mode == "backfill" && !completed {
		if strings.TrimSpace(latestSeen) != "" {
			checkpoint.LastMirroredTS = latestSeen
		}
		if strings.TrimSpace(pageOldestTS) != "" {
			checkpoint.BackfillBeforeTS = pageOldestTS
			checkpoint.BackfillOldestTS = pageOldestTS
		}
		checkpoint.BackfillComplete = false
		return
	}
	if strings.TrimSpace(latestSeen) != "" && (mode == "backfill" || completed) {
		checkpoint.LastMirroredTS = latestSeen
	}
	checkpoint.BackfillComplete = true
	checkpoint.BackfillBeforeTS = ""
	if completed {
		checkpoint.LastCompletedAt = now
	}
}

func slackMirrorMessageTimestampBounds(messages []slackapi.Message) (oldest string, newest string) {
	for _, msg := range messages {
		ts := strings.TrimSpace(msg.Timestamp)
		if ts == "" {
			continue
		}
		if oldest == "" || compareSlackTS(ts, oldest) < 0 {
			oldest = ts
		}
		if newest == "" || compareSlackTS(ts, newest) > 0 {
			newest = ts
		}
	}
	return oldest, newest
}

func slackMirrorTimestamp(t time.Time) string {
	if t.IsZero() {
		t = time.Now().UTC()
	}
	unix := t.Unix()
	micros := t.Nanosecond() / 1000
	return fmt.Sprintf("%d.%06d", unix, micros)
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
