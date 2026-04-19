package slack

import (
	"errors"
	"testing"

	slackapi "github.com/slack-go/slack"
)

type stubChannelNameResolverClient struct {
	conversationInfo map[string]string
	searchResults    map[string]string
	searchQueries    []string
}

func (s *stubChannelNameResolverClient) GetConversationInfo(input *slackapi.GetConversationInfoInput) (*slackapi.Channel, error) {
	channelID := ""
	if input != nil {
		channelID = input.ChannelID
	}
	if name, ok := s.conversationInfo[channelID]; ok {
		return &slackapi.Channel{GroupConversation: slackapi.GroupConversation{Name: name}}, nil
	}
	return nil, errors.New("missing_scope")
}

func (s *stubChannelNameResolverClient) SearchMessages(query string, params slackapi.SearchParameters) (*slackapi.SearchMessages, error) {
	s.searchQueries = append(s.searchQueries, query)
	channelID := ""
	for id, name := range s.searchResults {
		channelID = id
		return &slackapi.SearchMessages{
			Matches: []slackapi.SearchMessage{
				{
					Channel: slackapi.CtxChannel{
						ID:   id,
						Name: name,
					},
				},
			},
		}, nil
	}
	return &slackapi.SearchMessages{
		Matches: []slackapi.SearchMessage{
			{
				Channel: slackapi.CtxChannel{
					ID: channelID,
				},
			},
		},
	}, nil
}

func TestResolveChannelNamePrefersConversationInfo(t *testing.T) {
	client := &stubChannelNameResolverClient{
		conversationInfo: map[string]string{
			"C0AKH5SNGKH": "team-tiger",
		},
	}

	name, ok := resolveChannelName(client, "c0akh5sngkh")

	if !ok || name != "team-tiger" {
		t.Fatalf("expected direct conversation info resolution, got ok=%v name=%q", ok, name)
	}
	if len(client.searchQueries) != 0 {
		t.Fatalf("expected no search fallback, got %#v", client.searchQueries)
	}
}

func TestResolveChannelNameFallsBackToSearchMessages(t *testing.T) {
	client := &stubChannelNameResolverClient{
		searchResults: map[string]string{
			"C0AL7EKNHDF": "proj-numo-depin-app",
		},
	}

	name, ok := resolveChannelName(client, "C0AL7EKNHDF")

	if !ok || name != "proj-numo-depin-app" {
		t.Fatalf("expected search fallback resolution, got ok=%v name=%q", ok, name)
	}
	if len(client.searchQueries) != 1 || client.searchQueries[0] != "in:<#C0AL7EKNHDF>" {
		t.Fatalf("expected channel-scoped search query, got %#v", client.searchQueries)
	}
}

func TestCanonicalizePromptEnvelopePreservesResolvedChannelMetadata(t *testing.T) {
	envelope := SlackEnvelope{
		Text:      "Check <#C0AKH5SNGKH> for context.",
		ChannelID: "CINGRESS",
		ThreadTS:  "171000001.000100",
		UserID:    "U0ASDQKU3UL",
		EntityRefs: []EntityRef{
			{Kind: EntityChannel, ID: "C0AKH5SNGKH", Source: "mrkdwn"},
		},
	}
	resolver := &stubPromptResolver{
		userNames: map[string]string{
			"U0ASDQKU3UL": "blake",
		},
		channelNames: map[string]string{
			"C0AKH5SNGKH": "team-tiger",
			"CINGRESS":    "bot-questions",
		},
	}

	got := CanonicalizePromptEnvelope(envelope, resolver)

	if got.ChannelName != "bot-questions" {
		t.Fatalf("expected ingress channel label, got %#v", got)
	}
	if got.MentionedChannels[0].Label != "team-tiger" {
		t.Fatalf("expected resolved mentioned channel label, got %#v", got.MentionedChannels)
	}
	if got.RenderedText != "Check #team-tiger for context." {
		t.Fatalf("expected rendered text with resolved channel name, got %q", got.RenderedText)
	}
	channelNames := PromptEnvelopeChannelNames(got)
	if channelNames["C0AKH5SNGKH"] != "team-tiger" {
		t.Fatalf("expected prompt envelope channel metadata, got %#v", channelNames)
	}
}

type stubPromptResolver struct {
	userNames    map[string]string
	channelNames map[string]string
}

func (s *stubPromptResolver) UserDisplayName(userID string) (string, bool) {
	name, ok := s.userNames[userID]
	return name, ok
}

func (s *stubPromptResolver) ChannelName(channelID string) (string, bool) {
	name, ok := s.channelNames[channelID]
	return name, ok
}

func (s *stubPromptResolver) MessagePermalink(channelID string, messageTS string) (string, bool) {
	return "", false
}
