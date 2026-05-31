package improvementplane

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
)

const (
	slackUserNamesMetadataKey     = "slack_user_names"
	slackChannelNamesMetadataKey  = "slack_channel_names"
	slackTranscriptResolveBackoff = 5 * time.Minute
	slackTranscriptResolveTimeout = 2 * time.Second
)

type slackTranscriptResolver interface {
	UserName(userID string) (string, bool)
	ChannelName(channelID string) (string, bool)
}

type slackAPITranscriptResolver struct {
	client *slackapi.Client

	mu                   sync.Mutex
	userNames            map[string]string
	channelNames         map[string]string
	userInFlight         map[string]bool
	channelInFlight      map[string]bool
	userLastAttemptAt    map[string]time.Time
	channelLastAttemptAt map[string]time.Time
}

func newSlackTranscriptResolver(botToken string) slackTranscriptResolver {
	if strings.TrimSpace(botToken) == "" {
		return nil
	}
	return &slackAPITranscriptResolver{
		client:               slackapi.New(botToken, slackapi.OptionHTTPClient(&http.Client{Timeout: slackTranscriptResolveTimeout})),
		userNames:            map[string]string{},
		channelNames:         map[string]string{},
		userInFlight:         map[string]bool{},
		channelInFlight:      map[string]bool{},
		userLastAttemptAt:    map[string]time.Time{},
		channelLastAttemptAt: map[string]time.Time{},
	}
}

func (r *slackAPITranscriptResolver) UserName(userID string) (string, bool) {
	userID = strings.ToUpper(strings.TrimSpace(userID))
	if userID == "" {
		return "", false
	}
	r.mu.Lock()
	if name, ok := r.userNames[userID]; ok {
		r.mu.Unlock()
		return name, name != ""
	}
	if r.userInFlight[userID] || time.Since(r.userLastAttemptAt[userID]) < slackTranscriptResolveBackoff {
		r.mu.Unlock()
		return "", false
	}
	r.userInFlight[userID] = true
	r.userLastAttemptAt[userID] = time.Now()
	r.mu.Unlock()

	go r.resolveUserName(userID)
	return "", false
}

func (r *slackAPITranscriptResolver) resolveUserName(userID string) {
	defer func() {
		r.mu.Lock()
		delete(r.userInFlight, userID)
		r.mu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), slackTranscriptResolveTimeout)
	defer cancel()
	user, err := r.client.GetUserInfoContext(ctx, userID)
	if err != nil {
		return
	}
	name := slackUserDisplayName(user)
	r.mu.Lock()
	if name != "" {
		r.userNames[userID] = name
	}
	r.mu.Unlock()
}

func (r *slackAPITranscriptResolver) ChannelName(channelID string) (string, bool) {
	channelID = strings.ToUpper(strings.TrimSpace(channelID))
	if channelID == "" {
		return "", false
	}
	r.mu.Lock()
	if name, ok := r.channelNames[channelID]; ok {
		r.mu.Unlock()
		return name, name != ""
	}
	if r.channelInFlight[channelID] || time.Since(r.channelLastAttemptAt[channelID]) < slackTranscriptResolveBackoff {
		r.mu.Unlock()
		return "", false
	}
	r.channelInFlight[channelID] = true
	r.channelLastAttemptAt[channelID] = time.Now()
	r.mu.Unlock()

	go r.resolveChannelName(channelID)
	return "", false
}

func (r *slackAPITranscriptResolver) resolveChannelName(channelID string) {
	defer func() {
		r.mu.Lock()
		delete(r.channelInFlight, channelID)
		r.mu.Unlock()
	}()
	name, ok := slackpkg.ResolveChannelName(r.client, channelID)
	if !ok {
		return
	}
	r.mu.Lock()
	if name != "" {
		r.channelNames[channelID] = name
	}
	r.mu.Unlock()
}

func slackUserDisplayName(user *slackapi.User) string {
	if user == nil {
		return ""
	}
	for _, candidate := range []string{
		user.Profile.DisplayNameNormalized,
		user.Profile.DisplayName,
		user.Profile.RealNameNormalized,
		user.Profile.RealName,
		user.RealName,
		user.Name,
		user.ID,
	} {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func enrichSlackTranscriptEntries(entries []conversation.Entry, resolver slackTranscriptResolver) []conversation.Entry {
	if resolver == nil || len(entries) == 0 {
		return entries
	}

	out := make([]conversation.Entry, len(entries))
	copy(out, entries)
	changed := false
	for index, entry := range out {
		if entry.Source != ingestion.SourceSlack {
			continue
		}
		userIDs, channelIDs := slackTranscriptEntityIDs(entry)
		if len(userIDs) == 0 && len(channelIDs) == 0 {
			continue
		}

		metadata := cloneConversationMetadata(entry.Metadata)
		userNames := metadataStringMap(metadata[slackUserNamesMetadataKey])
		channelNames := metadataStringMap(metadata[slackChannelNamesMetadataKey])

		for _, userID := range userIDs {
			if _, ok := userNames[userID]; ok {
				continue
			}
			if name, ok := resolver.UserName(userID); ok {
				userNames[userID] = name
			}
		}
		for _, channelID := range channelIDs {
			if _, ok := channelNames[channelID]; ok {
				continue
			}
			if name, ok := resolver.ChannelName(channelID); ok {
				channelNames[channelID] = name
			}
		}

		if len(userNames) == 0 && len(channelNames) == 0 {
			continue
		}
		if len(userNames) > 0 {
			metadata[slackUserNamesMetadataKey] = userNames
		}
		if len(channelNames) > 0 {
			metadata[slackChannelNamesMetadataKey] = channelNames
		}

		entry.Metadata = metadata
		out[index] = entry
		changed = true
	}

	if !changed {
		return entries
	}
	return out
}

func slackTranscriptEntityIDs(entry conversation.Entry) ([]string, []string) {
	seenUsers := map[string]struct{}{}
	seenChannels := map[string]struct{}{}
	userIDs := []string{}
	channelIDs := []string{}
	appendEntity := func(item slackpkg.EntityRef) {
		id := strings.ToUpper(strings.TrimSpace(item.ID))
		if id == "" {
			return
		}
		switch item.Kind {
		case slackpkg.EntityUser:
			if _, ok := seenUsers[id]; ok {
				return
			}
			seenUsers[id] = struct{}{}
			userIDs = append(userIDs, id)
		case slackpkg.EntityChannel:
			if _, ok := seenChannels[id]; ok {
				return
			}
			seenChannels[id] = struct{}{}
			channelIDs = append(channelIDs, id)
		}
	}

	for _, item := range slackpkg.ExtractEntityRefs(entry.Body) {
		appendEntity(item)
	}
	for _, item := range slackpkg.EntityRefsFromValue(entry.Metadata["entity_refs"]) {
		appendEntity(item)
	}
	return userIDs, channelIDs
}

func cloneConversationMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		out[key] = value
	}
	return out
}

func metadataStringMap(value interface{}) map[string]string {
	switch typed := value.(type) {
	case map[string]string:
		out := make(map[string]string, len(typed))
		for key, item := range typed {
			key = strings.TrimSpace(key)
			item = strings.TrimSpace(item)
			if key != "" && item != "" {
				out[key] = item
			}
		}
		return out
	case map[string]interface{}:
		out := make(map[string]string, len(typed))
		for key, item := range typed {
			key = strings.TrimSpace(key)
			text, ok := item.(string)
			text = strings.TrimSpace(text)
			if ok && key != "" && text != "" {
				out[key] = text
			}
		}
		return out
	default:
		return map[string]string{}
	}
}
