package slack

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"

	slackapi "github.com/slack-go/slack"
)

type EntityResolver interface {
	UserDisplayName(userID string) (string, bool)
	ChannelName(channelID string) (string, bool)
	MessagePermalink(channelID string, messageTS string) (string, bool)
}

type slackAPIEntityResolver struct {
	client *slackapi.Client

	mu           sync.Mutex
	userNames    map[string]string
	channelNames map[string]string
}

var plainPromptEntityPattern = regexp.MustCompile(`(^|[^A-Za-z0-9])([@#])([A-Z0-9]{8,})([^A-Za-z0-9]|$)`)

func NewEntityResolver(token string) EntityResolver {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	return &slackAPIEntityResolver{
		client:       slackapi.New(token),
		userNames:    map[string]string{},
		channelNames: map[string]string{},
	}
}

func (r *slackAPIEntityResolver) UserDisplayName(userID string) (string, bool) {
	userID = strings.ToUpper(strings.TrimSpace(userID))
	if userID == "" {
		return "", false
	}
	r.mu.Lock()
	if value, ok := r.userNames[userID]; ok {
		r.mu.Unlock()
		return value, value != ""
	}
	r.mu.Unlock()

	user, err := r.client.GetUserInfo(userID)
	if err != nil {
		return "", false
	}
	value := slackUserDisplayName(user)

	r.mu.Lock()
	r.userNames[userID] = value
	r.mu.Unlock()
	return value, value != ""
}

func (r *slackAPIEntityResolver) ChannelName(channelID string) (string, bool) {
	channelID = strings.ToUpper(strings.TrimSpace(channelID))
	if channelID == "" {
		return "", false
	}
	r.mu.Lock()
	if value, ok := r.channelNames[channelID]; ok {
		r.mu.Unlock()
		return value, value != ""
	}
	r.mu.Unlock()

	channel, err := r.client.GetConversationInfo(&slackapi.GetConversationInfoInput{ChannelID: channelID})
	if err != nil {
		return "", false
	}
	value := strings.TrimSpace(channel.Name)

	r.mu.Lock()
	r.channelNames[channelID] = value
	r.mu.Unlock()
	return value, value != ""
}

func (r *slackAPIEntityResolver) MessagePermalink(channelID string, messageTS string) (string, bool) {
	channelID = strings.TrimSpace(channelID)
	messageTS = strings.TrimSpace(messageTS)
	if channelID == "" || messageTS == "" {
		return "", false
	}
	value, err := r.client.GetPermalink(&slackapi.PermalinkParameters{
		Channel: channelID,
		Ts:      messageTS,
	})
	if err != nil {
		return "", false
	}
	value = strings.TrimSpace(value)
	return value, value != ""
}

func PromptEnvelopeFromValue(value any) SlackPromptEnvelope {
	if value == nil {
		return SlackPromptEnvelope{}
	}
	switch typed := value.(type) {
	case SlackPromptEnvelope:
		return normalizePromptEnvelope(typed)
	case *SlackPromptEnvelope:
		if typed == nil {
			return SlackPromptEnvelope{}
		}
		return normalizePromptEnvelope(*typed)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return SlackPromptEnvelope{}
	}
	var out SlackPromptEnvelope
	if err := json.Unmarshal(data, &out); err != nil {
		return SlackPromptEnvelope{}
	}
	return normalizePromptEnvelope(out)
}

func CanonicalizePromptEnvelope(envelope SlackEnvelope, resolver EntityResolver) SlackPromptEnvelope {
	rawText := strings.TrimSpace(envelope.Text)
	channelID := strings.ToUpper(strings.TrimSpace(envelope.ChannelID))
	threadTS := strings.TrimSpace(envelope.ThreadTS)
	senderUserID := strings.ToUpper(strings.TrimSpace(envelope.UserID))
	entityRefs := normalizeEntityRefs(append([]EntityRef(nil), envelope.EntityRefs...))

	userNames := map[string]string{}
	channelNames := map[string]string{}
	mentionedUsers := []PromptEntity{}
	mentionedChannels := []PromptEntity{}

	appendMentionedUser := func(userID string, label string) {
		userID = strings.ToUpper(strings.TrimSpace(userID))
		label = strings.TrimSpace(label)
		if userID == "" {
			return
		}
		if label == "" && resolver != nil {
			if resolved, ok := resolver.UserDisplayName(userID); ok {
				label = resolved
			}
		}
		if label != "" {
			userNames[userID] = label
		}
		for _, item := range mentionedUsers {
			if item.ID == userID {
				return
			}
		}
		mentionedUsers = append(mentionedUsers, PromptEntity{ID: userID, Label: label})
	}
	appendMentionedChannel := func(channelID string, label string) {
		channelID = strings.ToUpper(strings.TrimSpace(channelID))
		label = strings.TrimSpace(label)
		if channelID == "" {
			return
		}
		if label == "" && resolver != nil {
			if resolved, ok := resolver.ChannelName(channelID); ok {
				label = resolved
			}
		}
		if label != "" {
			channelNames[channelID] = label
		}
		for _, item := range mentionedChannels {
			if item.ID == channelID {
				return
			}
		}
		mentionedChannels = append(mentionedChannels, PromptEntity{ID: channelID, Label: label})
	}

	for _, item := range entityRefs {
		switch item.Kind {
		case EntityUser:
			appendMentionedUser(item.ID, item.Label)
		case EntityChannel:
			appendMentionedChannel(item.ID, item.Label)
		}
	}

	senderDisplayName := ""
	if senderUserID != "" && resolver != nil {
		if resolved, ok := resolver.UserDisplayName(senderUserID); ok {
			senderDisplayName = resolved
			userNames[senderUserID] = resolved
		}
	}
	channelName := ""
	if channelID != "" && resolver != nil {
		if resolved, ok := resolver.ChannelName(channelID); ok {
			channelName = resolved
			channelNames[channelID] = resolved
		}
	}

	renderedText := RenderSlackPromptText(rawText, userNames, channelNames)
	if renderedText == "" {
		renderedText = rawText
	}
	permalink := ""
	if resolver != nil {
		if value, ok := resolver.MessagePermalink(channelID, firstNonEmpty(threadTS, strings.TrimSpace(envelope.TS))); ok {
			permalink = value
		}
	}
	return normalizePromptEnvelope(SlackPromptEnvelope{
		ChannelID:         channelID,
		ChannelName:       channelName,
		ThreadTS:          threadTS,
		SenderUserID:      senderUserID,
		SenderDisplayName: senderDisplayName,
		RawText:           rawText,
		RenderedText:      renderedText,
		MentionedChannels: mentionedChannels,
		MentionedUsers:    mentionedUsers,
		Permalink:         permalink,
	})
}

func RenderSlackPromptText(text string, userNames map[string]string, channelNames map[string]string) string {
	decoded := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">").Replace(strings.TrimSpace(text))
	if decoded == "" {
		return ""
	}
	matcher := regexp.MustCompile(`<([^>\n]+)>`)
	var out strings.Builder
	lastIndex := 0
	matches := matcher.FindAllStringSubmatchIndex(decoded, -1)
	for _, match := range matches {
		start, end := match[0], match[1]
		if start > lastIndex {
			out.WriteString(decoded[lastIndex:start])
		}
		token := decoded[match[2]:match[3]]
		out.WriteString(renderSlackToken(token, userNames, channelNames))
		lastIndex = end
	}
	if lastIndex < len(decoded) {
		out.WriteString(decoded[lastIndex:])
	}
	rendered := out.String()
	if rendered == "" {
		rendered = decoded
	}
	return plainPromptEntityPattern.ReplaceAllStringFunc(rendered, func(match string) string {
		submatch := plainPromptEntityPattern.FindStringSubmatch(match)
		if len(submatch) < 5 {
			return match
		}
		boundary := submatch[1]
		prefix := submatch[2]
		id := strings.ToUpper(strings.TrimSpace(submatch[3]))
		suffix := submatch[4]
		if prefix == "@" {
			if label := strings.TrimSpace(userNames[id]); label != "" {
				return boundary + "@" + label + suffix
			}
			return boundary + "@" + id + suffix
		}
		if label := strings.TrimSpace(channelNames[id]); label != "" {
			return boundary + "#" + label + suffix
		}
		return boundary + "#" + id + suffix
	})
}

func renderSlackToken(token string, userNames map[string]string, channelNames map[string]string) string {
	value := strings.TrimSpace(token)
	label := ""
	if separator := strings.Index(value, "|"); separator >= 0 {
		label = strings.TrimSpace(value[separator+1:])
		value = strings.TrimSpace(value[:separator])
	}
	switch {
	case strings.HasPrefix(value, "@"):
		id := strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(value, "@")))
		return "@" + firstNonEmpty(label, userNames[id], id)
	case strings.HasPrefix(value, "#"):
		id := strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(value, "#")))
		return "#" + firstNonEmpty(label, channelNames[id], id)
	case strings.HasPrefix(value, "!subteam^"):
		return firstNonEmpty(label, "@group")
	case strings.HasPrefix(value, "!date^"):
		return firstNonEmpty(label, strings.TrimPrefix(value, "!date^"))
	case strings.HasPrefix(value, "!"):
		return "@" + firstNonEmpty(label, strings.TrimPrefix(value, "!"))
	case strings.HasPrefix(strings.ToLower(value), "http://"), strings.HasPrefix(strings.ToLower(value), "https://"), strings.HasPrefix(strings.ToLower(value), "mailto:"):
		return firstNonEmpty(label, strings.TrimPrefix(value, "mailto:"), value)
	default:
		return firstNonEmpty(label, value)
	}
}

func PromptEnvelopeUserNames(envelope SlackPromptEnvelope) map[string]string {
	out := map[string]string{}
	for _, item := range envelope.MentionedUsers {
		if id, label := strings.ToUpper(strings.TrimSpace(item.ID)), strings.TrimSpace(item.Label); id != "" && label != "" {
			out[id] = label
		}
	}
	if id, label := strings.ToUpper(strings.TrimSpace(envelope.SenderUserID)), strings.TrimSpace(envelope.SenderDisplayName); id != "" && label != "" {
		out[id] = label
	}
	return out
}

func PromptEnvelopeChannelNames(envelope SlackPromptEnvelope) map[string]string {
	out := map[string]string{}
	for _, item := range envelope.MentionedChannels {
		if id, label := strings.ToUpper(strings.TrimSpace(item.ID)), strings.TrimSpace(item.Label); id != "" && label != "" {
			out[id] = label
		}
	}
	if id, label := strings.ToUpper(strings.TrimSpace(envelope.ChannelID)), strings.TrimSpace(envelope.ChannelName); id != "" && label != "" {
		out[id] = label
	}
	return out
}

func normalizePromptEnvelope(envelope SlackPromptEnvelope) SlackPromptEnvelope {
	envelope.ChannelID = strings.ToUpper(strings.TrimSpace(envelope.ChannelID))
	envelope.ChannelName = strings.TrimSpace(envelope.ChannelName)
	envelope.ThreadTS = strings.TrimSpace(envelope.ThreadTS)
	envelope.SenderUserID = strings.ToUpper(strings.TrimSpace(envelope.SenderUserID))
	envelope.SenderDisplayName = strings.TrimSpace(envelope.SenderDisplayName)
	envelope.RawText = strings.TrimSpace(envelope.RawText)
	envelope.RenderedText = strings.TrimSpace(envelope.RenderedText)
	envelope.Permalink = strings.TrimSpace(envelope.Permalink)
	envelope.MentionedChannels = normalizePromptEntities(envelope.MentionedChannels)
	envelope.MentionedUsers = normalizePromptEntities(envelope.MentionedUsers)
	return envelope
}

func normalizePromptEntities(items []PromptEntity) []PromptEntity {
	out := make([]PromptEntity, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item.ID = strings.ToUpper(strings.TrimSpace(item.ID))
		item.Label = strings.TrimSpace(item.Label)
		if item.ID == "" {
			continue
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		out = append(out, item)
	}
	return out
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
