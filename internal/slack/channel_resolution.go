package slack

import (
	"fmt"
	"strings"

	slackapi "github.com/slack-go/slack"
)

type channelNameResolverClient interface {
	GetConversationInfo(input *slackapi.GetConversationInfoInput) (*slackapi.Channel, error)
	SearchMessages(query string, params slackapi.SearchParameters) (*slackapi.SearchMessages, error)
}

func ResolveChannelName(client *slackapi.Client, channelID string) (string, bool) {
	return resolveChannelName(client, channelID)
}

func resolveChannelName(client channelNameResolverClient, channelID string) (string, bool) {
	channelID = strings.ToUpper(strings.TrimSpace(channelID))
	if client == nil || channelID == "" {
		return "", false
	}
	channel, err := client.GetConversationInfo(&slackapi.GetConversationInfoInput{ChannelID: channelID})
	if err == nil {
		if name := strings.TrimSpace(channel.Name); name != "" {
			return name, true
		}
	}

	params := slackapi.NewSearchParameters()
	params.Count = 1
	matches, err := client.SearchMessages(fmt.Sprintf("in:<#%s>", channelID), params)
	if err != nil {
		return "", false
	}
	for _, match := range matches.Matches {
		matchChannelID := strings.ToUpper(strings.TrimSpace(match.Channel.ID))
		if matchChannelID != "" && matchChannelID != channelID {
			continue
		}
		if name := strings.TrimSpace(match.Channel.Name); name != "" {
			return name, true
		}
	}
	for _, match := range matches.Matches {
		if name := strings.TrimSpace(match.Channel.Name); name != "" {
			return name, true
		}
	}
	return "", false
}
