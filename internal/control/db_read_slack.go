package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	dbReadSlackApproveAction = "rsi_db_read_approve"
	dbReadSlackDenyAction    = "rsi_db_read_deny"
	dbReadSlackTruncated     = "\n-- truncated for Slack --"
)

func postDBReadApprovalCard(ctx context.Context, cfg config.Config, store storepkg.Store, api slackMessagePoster, request storepkg.DBReadRequest, attempt storepkg.DBReadValidationAttempt, preview string) error {
	if api == nil {
		return fmt.Errorf("slackAPI is required to post DB read approval cards")
	}
	if strings.TrimSpace(request.ChannelID) == "" {
		return fmt.Errorf("db read request has no Slack channel")
	}
	text := dbReadApprovalText(request, attempt)
	blocks := dbReadApprovalBlocks(request, attempt, preview)
	options := []slack.MsgOption{
		slack.MsgOptionText(text, false),
		slack.MsgOptionBlocks(blocks...),
	}
	if strings.TrimSpace(request.ThreadTS) != "" {
		options = append(options, slack.MsgOptionTS(request.ThreadTS))
	}
	channel, ts, err := api.PostMessageContext(ctx, request.ChannelID, options...)
	if err != nil {
		return err
	}
	_, err = store.TransitionDBReadRequest(request.ID, storepkg.DBReadStatePendingApproval, storepkg.DBReadStatePendingApproval, func(item *storepkg.DBReadRequest) error {
		item.SlackMessageChannelID = channel
		item.SlackMessageTS = ts
		return nil
	})
	return err
}

func dbReadApprovalBlocks(request storepkg.DBReadRequest, attempt storepkg.DBReadValidationAttempt, preview string) []slack.Block {
	approve := slack.NewButtonBlockElement(
		dbReadSlackApproveAction,
		request.ID,
		slack.NewTextBlockObject(slack.PlainTextType, "Approve once", false, false),
	)
	approve.Style = slack.StylePrimary
	deny := slack.NewButtonBlockElement(
		dbReadSlackDenyAction,
		request.ID,
		slack.NewTextBlockObject(slack.PlainTextType, "Deny", false, false),
	)
	deny.Style = slack.StyleDanger
	sqlPreview := truncateSlackText(preview, 2200)
	return []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, dbReadApprovalText(request, attempt), false, false), nil, nil),
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "```"+escapeSlackCode(sqlPreview)+"```", false, false), nil, nil),
		slack.NewActionBlock("rsi_db_read_actions", approve, deny),
	}
}

func dbReadApprovalText(request storepkg.DBReadRequest, attempt storepkg.DBReadValidationAttempt) string {
	expires := request.ExpiresAt.Format(time.RFC3339)
	hash := request.SQLSHA256
	if len(hash) > 24 {
		hash = hash[:24] + "..."
	}
	return fmt.Sprintf(
		"*RSI DB read approval requested*\nTarget: `%s`\nRequester: `%s`\nHash: `%s`\nValidation attempt: `%s`\nCaps: max_rows=%d max_bytes=%d timeout=%ds\nExpires: `%s`",
		request.Target,
		firstNonEmpty(request.Requester, "hermes"),
		hash,
		attempt.ID,
		request.Caps.MaxRows,
		request.Caps.MaxBytes,
		request.Caps.TimeoutSeconds,
		expires,
	)
}

func updateDBReadSlackCard(ctx context.Context, api slackMessagePoster, request storepkg.DBReadRequest, statusText string) error {
	if api == nil || request.SlackMessageChannelID == "" || request.SlackMessageTS == "" {
		return nil
	}
	text := fmt.Sprintf("*RSI DB read request `%s`*: %s\nTarget: `%s`\nHash: `%s`", request.ID, statusText, request.Target, request.SQLSHA256)
	if len(request.ResultSample) > 0 {
		raw, _ := json.MarshalIndent(request.ResultSample, "", "  ")
		text += "\nSample:\n```" + escapeSlackCode(truncateSlackText(string(raw), 2200)) + "```"
	}
	_, _, _, err := api.UpdateMessageContext(
		ctx,
		request.SlackMessageChannelID,
		request.SlackMessageTS,
		slack.MsgOptionText(text, false),
		slack.MsgOptionBlocks(slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, text, false, false), nil, nil)),
	)
	return err
}

func truncateSlackText(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max <= len(dbReadSlackTruncated) {
		for i := max; i > 0; i-- {
			if utf8.ValidString(value[:i]) {
				return value[:i]
			}
		}
		return ""
	}
	cutPoint := max - len(dbReadSlackTruncated)
	for i := cutPoint; i >= 0 && i > cutPoint-4; i-- {
		if utf8.ValidString(value[:i]) {
			return value[:i] + dbReadSlackTruncated
		}
	}
	if cutPoint > 0 {
		return value[:cutPoint] + dbReadSlackTruncated
	}
	return ""
}

func escapeSlackCode(value string) string {
	return strings.ReplaceAll(value, "```", "` ` `")
}
