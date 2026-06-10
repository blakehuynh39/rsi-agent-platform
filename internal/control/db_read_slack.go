package control

import (
	"context"
	"fmt"
	"sort"
	"strings"
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
	if request.SlackMessageChannelID != "" && request.SlackMessageTS != "" {
		return nil
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

// postDBReadAuditCard posts the same audit card as the approval flow but
// without approve/deny buttons, for requests that are auto-approved after
// read-only validation. It records the Slack message coordinates so later
// status updates land on this card.
func postDBReadAuditCard(ctx context.Context, store storepkg.Store, api slackMessagePoster, request storepkg.DBReadRequest, attempt storepkg.DBReadValidationAttempt, preview string) error {
	if api == nil {
		return fmt.Errorf("slackAPI is required to post DB read audit cards")
	}
	if strings.TrimSpace(request.ChannelID) == "" {
		return fmt.Errorf("db read request has no Slack channel")
	}
	if request.SlackMessageChannelID != "" && request.SlackMessageTS != "" {
		return nil
	}
	text := dbReadAuditText(request)
	blocks := dbReadAuditBlocks(request, attempt, preview)
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

func dbReadAuditBlocks(request storepkg.DBReadRequest, attempt storepkg.DBReadValidationAttempt, preview string) []slack.Block {
	sqlPreview := truncateSlackText(firstNonEmpty(request.SQL, preview), 2200)
	return []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, dbReadAuditText(request), false, false), nil, nil),
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Exact SQL to run:*\n```"+escapeSlackCode(sqlPreview)+"```", false, false), nil, nil),
		slack.NewContextBlock("rsi_db_read_footer", slack.NewTextBlockObject(slack.MarkdownType, dbReadRequestFooter(request, attempt), false, false)),
	}
}

func dbReadAuditText(request storepkg.DBReadRequest) string {
	return fmt.Sprintf(
		"*DB read (auto-approved)*\nTarget: `%s`  Requester: %s\nPurpose: `%s`\nCaps: max_rows=%d max_bytes=%d timeout=%ds\nValidated read-only; executing without manual approval.",
		request.Target,
		dbReadRequesterLabel(request.Requester),
		firstNonEmpty(request.Purpose, "query"),
		request.Caps.MaxRows,
		request.Caps.MaxBytes,
		request.Caps.TimeoutSeconds,
	)
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
	sqlPreview := truncateSlackText(firstNonEmpty(request.SQL, preview), 2200)
	return []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, dbReadApprovalText(request, attempt), false, false), nil, nil),
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Exact SQL to run:*\n```"+escapeSlackCode(sqlPreview)+"```", false, false), nil, nil),
		slack.NewContextBlock("rsi_db_read_footer", slack.NewTextBlockObject(slack.MarkdownType, dbReadRequestFooter(request, attempt), false, false)),
		slack.NewActionBlock("rsi_db_read_actions", approve, deny),
	}
}

func dbReadApprovalText(request storepkg.DBReadRequest, attempt storepkg.DBReadValidationAttempt) string {
	expires := request.ExpiresAt.Format("15:04 MST")
	return fmt.Sprintf(
		"*Approve DB read?*\nTarget: `%s`  Requester: %s\nPurpose: `%s`\nCaps: max_rows=%d max_bytes=%d timeout=%ds  Expires: `%s`\nOnly authorized approvers can approve or deny this request.",
		request.Target,
		dbReadRequesterLabel(request.Requester),
		firstNonEmpty(request.Purpose, "query"),
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
	const slackSectionTextLimit = 2900
	header := fmt.Sprintf("*DB read `%s`*: %s\nTarget: `%s`  Requester: %s", request.ID, statusText, request.Target, dbReadRequesterLabel(request.Requester))
	if request.ApprovedBySlackUserID != "" {
		header += fmt.Sprintf("\nApproved by: <@%s>", request.ApprovedBySlackUserID)
	}
	footer := "\n" + dbReadRequestFooter(request, storepkg.DBReadValidationAttempt{ID: request.CurrentValidationAttemptID})
	overhead := len(header) + len(footer) + len("\n*Exact SQL:*\n``````") + len("\nResult: rows= truncated= ref=``")
	remainingBudget := slackSectionTextLimit - overhead
	if remainingBudget < 0 {
		remainingBudget = 0
	}
	sqlLimit := 1000
	if remainingBudget < sqlLimit {
		sqlLimit = remainingBudget
	}
	text := header
	text += "\n*Exact SQL:*\n```" + escapeSlackCode(truncateSlackText(request.SQL, sqlLimit)) + "```"
	if request.RowCount > 0 || request.ResultArtifactRef != "" || request.Truncated {
		resultLabel := "Result"
		if request.Truncated {
			resultLabel = "Result (truncated)"
		}
		text += fmt.Sprintf("\n*%s:* rows=%d truncated=%t ref=`%s`", resultLabel, request.RowCount, request.Truncated, firstNonEmpty(request.ResultArtifactRef, request.ID))
	}
	text += footer
	_, _, _, err := api.UpdateMessageContext(
		ctx,
		request.SlackMessageChannelID,
		request.SlackMessageTS,
		slack.MsgOptionText(text, false),
		slack.MsgOptionBlocks(slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, text, false, false), nil, nil)),
	)
	return err
}

func dbReadRequestFooter(request storepkg.DBReadRequest, attempt storepkg.DBReadValidationAttempt) string {
	validationID := firstNonEmpty(attempt.ID, request.CurrentValidationAttemptID, "n/a")
	return fmt.Sprintf("Request `%s` | Hash `%s` | Validation `%s`", request.ID, request.SQLSHA256, validationID)
}

func dbReadRequesterLabel(requester string) string {
	requester = strings.TrimSpace(requester)
	for _, prefix := range []string{"user:", "operator:"} {
		if strings.HasPrefix(requester, prefix) {
			id := strings.TrimSpace(strings.TrimPrefix(requester, prefix))
			if id != "" {
				return fmt.Sprintf("<@%s> via Hermes", id)
			}
		}
	}
	if looksLikeSlackUserID(requester) {
		return fmt.Sprintf("<@%s> via Hermes", requester)
	}
	if requester == "" {
		return "`hermes`"
	}
	return "`" + escapeSlackCode(requester) + "`"
}

func looksLikeSlackUserID(value string) bool {
	if len(value) < 3 {
		return false
	}
	switch value[0] {
	case 'U', 'W':
	default:
		return false
	}
	for _, ch := range value[1:] {
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			continue
		}
		return false
	}
	return true
}

func formatDBReadSampleTable(rows []map[string]string, max int) string {
	if len(rows) == 0 {
		return "(no sample rows)"
	}
	columnSet := map[string]bool{}
	for _, row := range rows {
		for column := range row {
			columnSet[column] = true
		}
	}
	columns := make([]string, 0, len(columnSet))
	for column := range columnSet {
		columns = append(columns, column)
	}
	sort.Strings(columns)
	widths := make(map[string]int, len(columns))
	for _, column := range columns {
		widths[column] = len(column)
	}
	for _, row := range rows {
		for _, column := range columns {
			if n := len(row[column]); n > widths[column] {
				widths[column] = n
			}
		}
	}
	var b strings.Builder
	writeRow := func(values map[string]string, header bool) {
		for i, column := range columns {
			if i > 0 {
				b.WriteString(" | ")
			}
			value := column
			if !header {
				value = values[column]
			}
			b.WriteString(value)
			for j := len(value); j < widths[column]; j++ {
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	writeRow(nil, true)
	for i, column := range columns {
		if i > 0 {
			b.WriteString("-+-")
		}
		b.WriteString(strings.Repeat("-", widths[column]))
	}
	b.WriteByte('\n')
	for _, row := range rows {
		writeRow(row, false)
	}
	return truncateSlackText(b.String(), max)
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
