package improvementplane

import (
	"regexp"
	"strings"
	"sync"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
)

const (
	slackUserNamesMetadataKey    = "slack_user_names"
	slackChannelNamesMetadataKey = "slack_channel_names"
)

var (
	slackUserMentionPattern    = regexp.MustCompile(`<@([A-Z0-9]+)(?:\|[^>]+)?>`)
	slackChannelMentionPattern = regexp.MustCompile(`<#([A-Z0-9]+)(?:\|[^>]+)?>`)
)

type slackTranscriptResolver interface {
	UserName(userID string) (string, bool)
	ChannelName(channelID string) (string, bool)
}

type slackAPITranscriptResolver struct {
	client *slackapi.Client

	mu           sync.Mutex
	userNames    map[string]string
	channelNames map[string]string
}

func newSlackTranscriptResolver(botToken string) slackTranscriptResolver {
	if strings.TrimSpace(botToken) == "" {
		return nil
	}
	return &slackAPITranscriptResolver{
		client:       slackapi.New(botToken),
		userNames:    map[string]string{},
		channelNames: map[string]string{},
	}
}

func (r *slackAPITranscriptResolver) UserName(userID string) (string, bool) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", false
	}
	r.mu.Lock()
	if name, ok := r.userNames[userID]; ok {
		r.mu.Unlock()
		return name, name != ""
	}
	r.mu.Unlock()

	user, err := r.client.GetUserInfo(userID)
	if err != nil {
		return "", false
	}

	name := slackUserDisplayName(user)
	r.mu.Lock()
	if name != "" {
		r.userNames[userID] = name
	}
	r.mu.Unlock()
	return name, name != ""
}

func (r *slackAPITranscriptResolver) ChannelName(channelID string) (string, bool) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return "", false
	}
	r.mu.Lock()
	if name, ok := r.channelNames[channelID]; ok {
		r.mu.Unlock()
		return name, name != ""
	}
	r.mu.Unlock()

	channel, err := r.client.GetConversationInfo(&slackapi.GetConversationInfoInput{ChannelID: channelID})
	if err != nil {
		return "", false
	}

	name := strings.TrimSpace(channel.Name)
	r.mu.Lock()
	if name != "" {
		r.channelNames[channelID] = name
	}
	r.mu.Unlock()
	return name, name != ""
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
		if entry.Source != ingestion.SourceSlack || !strings.Contains(entry.Body, "<") {
			continue
		}
		userIDs := slackEntityIDs(entry.Body, slackUserMentionPattern)
		channelIDs := slackEntityIDs(entry.Body, slackChannelMentionPattern)
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

func slackEntityIDs(body string, pattern *regexp.Regexp) []string {
	matches := pattern.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		id := strings.TrimSpace(match[1])
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
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
