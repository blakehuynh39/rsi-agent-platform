package control

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	slackReportSchemaVersion = 1
	slackReportMaxCellLength = 500
	slackReportMaxTableCols  = 20
	slackReportMaxInlineRows = 99
	slackReportMaxMarkdown   = 2800
)

var markdownTableSeparatorRegex = regexp.MustCompile(`^\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?$`)
var slackReportNumericCellRegex = regexp.MustCompile(`^[+-]?(?:(?:\d+)|(?:\d{1,3}(?:,\d{3})+))(?:\.\d+)?%?$`)
var safeFilenameRegex = regexp.MustCompile(`[^a-z0-9._-]+`)

type slackReportPayload struct {
	ReportSchemaVersion int                  `json:"report_schema_version"`
	Summary             string               `json:"summary"`
	Sections            []slackReportSection `json:"sections,omitempty"`
	Tables              []slackReportTable   `json:"tables,omitempty"`
	Files               []slackReportFile    `json:"files,omitempty"`
	Images              []slackReportFile    `json:"images,omitempty"`
}

type slackReportSection struct {
	Title string `json:"title,omitempty"`
	Text  string `json:"text,omitempty"`
}

type slackReportColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Align string `json:"align,omitempty"`
}

type slackReportTable struct {
	Title   string                   `json:"title,omitempty"`
	Caption string                   `json:"caption,omitempty"`
	Columns []slackReportColumn      `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

type slackReportFile struct {
	ArtifactRef string `json:"artifact_ref,omitempty"`
	Path        string `json:"path,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Title       string `json:"title,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
	Content     string `json:"content,omitempty"`
	Base64      string `json:"content_base64,omitempty"`
}

type slackReportPlan struct {
	ChannelID string
	ThreadTS  string
	Text      string
	Blocks    []slackapi.Block
	Uploads   []slackReportUploadPlan
	Manifest  map[string]interface{}
}

type slackReportUploadPlan struct {
	ID     string
	Kind   string
	Params slackapi.UploadFileParameters
}

type slackReportManifestUpload struct {
	ID       string
	Kind     string
	Filename string
}

type slackReportValidationError struct {
	Code    string      `json:"code"`
	Path    string      `json:"path"`
	Limit   interface{} `json:"limit,omitempty"`
	Actual  interface{} `json:"actual,omitempty"`
	Message string      `json:"message"`
}

type slackReportValidationErrors []slackReportValidationError

func (errs slackReportValidationErrors) Error() string {
	data, err := json.Marshal(map[string]interface{}{
		"code":   "slack_report_validation_failed",
		"errors": []slackReportValidationError(errs),
	})
	if err != nil {
		return "slack_report_validation_failed"
	}
	return string(data)
}

func executeSlackReportNativeToolAction(ctx context.Context, cfg config.Config, repo storepkg.Repository, api *slackapi.Client, input nativeToolActionRequest) (any, string, string, string, map[string]any, int, error) {
	plan, err := buildSlackReportPlan(input.Arguments, input.TargetRef)
	if err != nil {
		return nil, "", "", "", slackMirrorEffect("not_attempted", ""), http.StatusBadRequest, err
	}
	previousResult := mapValue(input.Arguments["__previous_result_payload"])
	previousManifest := mapValue(previousResult["render_manifest"])
	manifest := storepkg.CloneJSONMap(plan.Manifest)
	channel, ts, sourceRef := "", "", ""
	if mainMessage, ok := slackReportPostedMainMessage(previousManifest); ok {
		channel = stringValueFromMap(mainMessage, "channel_id")
		ts = stringValueFromMap(mainMessage, "ts")
		sourceRef = firstNonEmpty(stringValueFromMap(mainMessage, "source_ref"), "slack:"+channel+":"+ts)
		manifest["main_message"] = mainMessage
	} else {
		options := []slackapi.MsgOption{slackapi.MsgOptionText(plan.Text, false)}
		if plan.ThreadTS != "" {
			options = append(options, slackapi.MsgOptionTS(plan.ThreadTS))
		}
		if len(plan.Blocks) > 0 {
			options = append(options, slackapi.MsgOptionBlocks(plan.Blocks...))
		}
		var err error
		channel, ts, err = api.PostMessageContext(ctx, plan.ChannelID, options...)
		sourceRef = "slack:" + channel + ":" + ts
		manifest["main_message"] = map[string]interface{}{
			"status":     map[bool]string{true: "posted", false: "failed"}[err == nil],
			"channel_id": channel,
			"thread_ts":  firstNonEmpty(plan.ThreadTS, ts),
			"ts":         ts,
			"source_ref": sourceRef,
		}
		if err != nil {
			manifest["error"] = err.Error()
			return map[string]interface{}{"render_manifest": manifest}, "", sourceRef, "", nativeSlackRefreshMirrorEffect(ctx, cfg, repo, api, channel, ts, firstNonEmpty(plan.ThreadTS, ts), err), statusFromErr(err), err
		}
	}

	uploadResults := make([]map[string]interface{}, 0, len(plan.Uploads))
	var firstUploadErr error
	previousUploads := slackReportPreviousUploadsByID(previousManifest)
	for _, upload := range plan.Uploads {
		if previousUpload, ok := previousUploads[upload.ID]; ok && slackReportUploadSucceeded(previousUpload) {
			uploadResults = append(uploadResults, previousUpload)
			continue
		}
		params := upload.Params
		params.Channel = firstNonEmpty(params.Channel, channel)
		params.ThreadTimestamp = firstNonEmpty(params.ThreadTimestamp, plan.ThreadTS, ts)
		file, uploadErr := api.UploadFileContext(ctx, params)
		result := map[string]interface{}{
			"id":       upload.ID,
			"kind":     upload.Kind,
			"filename": params.Filename,
			"status":   map[bool]string{true: "uploaded", false: "failed"}[uploadErr == nil],
		}
		if file != nil {
			result["slack_file_id"] = file.ID
			result["source_ref"] = "slack_file:" + file.ID
		}
		if uploadErr != nil {
			result["error"] = uploadErr.Error()
			if firstUploadErr == nil {
				firstUploadErr = uploadErr
			}
		}
		uploadResults = append(uploadResults, result)
	}
	manifest["uploads"] = uploadResults
	uploadFailureCount := 0
	for _, upload := range uploadResults {
		if strings.EqualFold(stringValueFromMap(upload, "status"), "failed") {
			uploadFailureCount++
		}
	}
	if uploadFailureCount > 0 {
		manifest["upload_failure_count"] = uploadFailureCount
	}
	mirrorEffect := nativeSlackRefreshMirrorEffect(ctx, cfg, repo, api, channel, ts, firstNonEmpty(plan.ThreadTS, ts), nil)
	out := map[string]interface{}{
		"channel_id":       channel,
		"ts":               ts,
		"render_manifest":  manifest,
		"uploaded_files":   uploadResults,
		"renderer_version": slackReportSchemaVersion,
	}
	if firstUploadErr != nil {
		warningText := fmt.Sprintf(":warning: %d attachment upload(s) failed. The report summary was posted; see the RSI trace render manifest for details.", uploadFailureCount)
		warningBlocks := append([]slackapi.Block{}, plan.Blocks...)
		warningBlocks = append(warningBlocks, slackapi.NewMarkdownBlock("report-upload-warning", warningText))
		_, _, _, updateErr := api.UpdateMessageContext(ctx, channel, ts,
			slackapi.MsgOptionText(plan.Text+"\n\n"+warningText, false),
			slackapi.MsgOptionBlocks(warningBlocks...),
		)
		updateStatus := "updated"
		if updateErr != nil {
			updateStatus = "failed"
			manifest["upload_warning_update_error"] = updateErr.Error()
		}
		manifest["upload_warning_update"] = map[string]interface{}{
			"status": updateStatus,
			"text":   warningText,
		}
		out["upload_error"] = firstUploadErr.Error()
		out["upload_failure_count"] = uploadFailureCount
		return out, fmt.Sprintf("posted Slack report with %d upload failure(s)", uploadFailureCount), sourceRef, "", mirrorEffect, http.StatusOK, nil
	}
	return out, "posted Slack report", sourceRef, "", mirrorEffect, http.StatusOK, nil
}

func slackReportResultHasPostedMain(result map[string]interface{}) bool {
	manifest := mapValue(result["render_manifest"])
	_, ok := slackReportPostedMainMessage(manifest)
	return ok
}

func slackReportResultHasUploadFailures(result map[string]interface{}) bool {
	manifest := mapValue(result["render_manifest"])
	if intArg(manifest, "upload_failure_count", 0) > 0 || intArg(result, "upload_failure_count", 0) > 0 {
		return true
	}
	for _, raw := range arrayArg(manifest, "uploads") {
		if strings.EqualFold(stringValueFromMap(mapValue(raw), "status"), "failed") {
			return true
		}
	}
	return false
}

func slackReportPostedMainMessage(manifest map[string]interface{}) (map[string]interface{}, bool) {
	mainMessage := mapValue(manifest["main_message"])
	if !strings.EqualFold(stringValueFromMap(mainMessage, "status"), "posted") {
		return nil, false
	}
	if strings.TrimSpace(stringValueFromMap(mainMessage, "channel_id")) == "" || strings.TrimSpace(stringValueFromMap(mainMessage, "ts")) == "" {
		return nil, false
	}
	return storepkg.CloneJSONMap(mainMessage), true
}

func slackReportPreviousUploadsByID(manifest map[string]interface{}) map[string]map[string]interface{} {
	out := map[string]map[string]interface{}{}
	for _, raw := range arrayArg(manifest, "uploads") {
		item := mapValue(raw)
		id := strings.TrimSpace(stringValueFromMap(item, "id"))
		if id == "" {
			continue
		}
		out[id] = storepkg.CloneJSONMap(item)
	}
	return out
}

func slackReportUploadSucceeded(upload map[string]interface{}) bool {
	return strings.EqualFold(stringValueFromMap(upload, "status"), "uploaded") &&
		strings.TrimSpace(stringValueFromMap(upload, "source_ref")) != ""
}

func buildSlackReportPlan(args map[string]interface{}, targetRef string) (slackReportPlan, error) {
	payload, err := slackReportPayloadFromArgs(args)
	if err != nil {
		return slackReportPlan{}, err
	}
	if validation := validateSlackReportPayload(payload); len(validation) > 0 {
		return slackReportPlan{}, validation
	}
	channelID := firstNonEmpty(stringArg(args, "channel_id"), targetRef)
	if channelID == "" {
		return slackReportPlan{}, slackReportValidationErrors{{Code: "required", Path: "channel_id", Message: "channel_id is required"}}
	}
	blocks := []slackapi.Block{slackapi.NewMarkdownBlock("report-summary", payload.Summary)}
	for idx, section := range payload.Sections {
		text := strings.TrimSpace(section.Text)
		if strings.TrimSpace(section.Title) != "" {
			text = "*" + strings.TrimSpace(section.Title) + "*\n" + text
		}
		if text != "" {
			blocks = append(blocks, slackapi.NewMarkdownBlock(fmt.Sprintf("report-section-%d", idx), text))
		}
	}
	uploads := []slackReportUploadPlan{}
	manifestUploads := []slackReportManifestUpload{}
	inlineTable := slackReportInlineTable(payload)
	if inlineTable {
		table := payload.Tables[0]
		if strings.TrimSpace(table.Title) != "" {
			blocks = append(blocks, slackapi.NewMarkdownBlock("report-table-title", "*"+strings.TrimSpace(table.Title)+"*"))
		}
		blocks = append(blocks, slackTableBlock("report-table-0", table))
	} else if len(payload.Tables) > 0 {
		blocks = append(blocks, slackapi.NewMarkdownBlock("report-table-fallback", fmt.Sprintf("%d structured table(s) are attached as CSV files.", len(payload.Tables))))
		for idx, table := range payload.Tables {
			content, err := slackReportTableCSV(table)
			if err != nil {
				return slackReportPlan{}, err
			}
			filename := safeSlackReportFilename(firstNonEmpty(table.Title, fmt.Sprintf("table-%d", idx+1)), ".csv")
			uploadID := fmt.Sprintf("table-%d", idx)
			manifestUploads = append(manifestUploads, slackReportManifestUpload{ID: uploadID, Kind: "generated_csv", Filename: filename})
			uploads = append(uploads, slackReportUploadPlan{
				ID:   uploadID,
				Kind: "generated_csv",
				Params: slackapi.UploadFileParameters{
					Channel:         channelID,
					ThreadTimestamp: stringArg(args, "thread_ts"),
					Filename:        filename,
					Title:           firstNonEmpty(table.Title, filename),
					Content:         content,
					FileSize:        len([]byte(content)),
				},
			})
		}
	}
	for idx, item := range append(append([]slackReportFile{}, payload.Files...), payload.Images...) {
		path := firstNonEmpty(item.ArtifactRef, item.Path)
		content := strings.TrimSpace(item.Content)
		contentBase64 := strings.TrimSpace(item.Base64)
		if path == "" && content == "" && contentBase64 == "" {
			fieldPath, fieldIdx := slackReportFileFieldPath(idx, len(payload.Files))
			return slackReportPlan{}, slackReportValidationErrors{{Code: "required", Path: fmt.Sprintf("%s[%d].artifact_ref", fieldPath, fieldIdx), Message: "files and images must provide artifact_ref, path, content, or content_base64"}}
		}
		attachmentFilename := slackReportAttachmentFilename(item, path)
		uploadArgs := map[string]interface{}{
			"channel_id":      channelID,
			"thread_ts":       stringArg(args, "thread_ts"),
			"artifact_ref":    path,
			"content":         item.Content,
			"content_base64":  item.Base64,
			"filename":        attachmentFilename,
			"title":           firstNonEmpty(item.Title, attachmentFilename),
			"initial_comment": "",
		}
		params, err := slackUploadParams(uploadArgs, channelID)
		if err != nil {
			return slackReportPlan{}, err
		}
		uploadID := fmt.Sprintf("artifact-%d", idx)
		kind := "artifact_ref"
		if content != "" || contentBase64 != "" {
			kind = "inline_content"
		}
		manifestUploads = append(manifestUploads, slackReportManifestUpload{ID: uploadID, Kind: kind, Filename: attachmentFilename})
		uploads = append(uploads, slackReportUploadPlan{ID: uploadID, Kind: kind, Params: params})
	}
	manifest := buildSlackReportManifest(payload, args, channelID, inlineTable, manifestUploads)
	return slackReportPlan{
		ChannelID: channelID,
		ThreadTS:  stringArg(args, "thread_ts"),
		Text:      payload.Summary,
		Blocks:    blocks,
		Uploads:   uploads,
		Manifest:  manifest,
	}, nil
}

func buildSlackReportDraftManifest(args map[string]interface{}, targetRef string) (map[string]interface{}, error) {
	payload, err := slackReportPayloadFromArgs(args)
	if err != nil {
		return nil, err
	}
	if validation := validateSlackReportPayload(payload); len(validation) > 0 {
		return nil, validation
	}
	channelID := firstNonEmpty(stringArg(args, "channel_id"), targetRef)
	if channelID == "" {
		return nil, slackReportValidationErrors{{Code: "required", Path: "channel_id", Message: "channel_id is required"}}
	}
	inlineTable := slackReportInlineTable(payload)
	plannedUploads := []slackReportManifestUpload{}
	if !inlineTable {
		for idx, table := range payload.Tables {
			filename := safeSlackReportFilename(firstNonEmpty(table.Title, fmt.Sprintf("table-%d", idx+1)), ".csv")
			plannedUploads = append(plannedUploads, slackReportManifestUpload{ID: fmt.Sprintf("table-%d", idx), Kind: "generated_csv", Filename: filename})
		}
	}
	for idx, item := range append(append([]slackReportFile{}, payload.Files...), payload.Images...) {
		path := firstNonEmpty(item.ArtifactRef, item.Path)
		if path == "" && strings.TrimSpace(item.Content) == "" && strings.TrimSpace(item.Base64) == "" {
			fieldPath, fieldIdx := slackReportFileFieldPath(idx, len(payload.Files))
			return nil, slackReportValidationErrors{{Code: "required", Path: fmt.Sprintf("%s[%d].artifact_ref", fieldPath, fieldIdx), Message: "files and images must provide artifact_ref, path, content, or content_base64"}}
		}
		kind := "artifact_ref"
		if strings.TrimSpace(item.Content) != "" || strings.TrimSpace(item.Base64) != "" {
			kind = "inline_content"
		}
		plannedUploads = append(plannedUploads, slackReportManifestUpload{ID: fmt.Sprintf("artifact-%d", idx), Kind: kind, Filename: slackReportAttachmentFilename(item, path)})
	}
	return buildSlackReportManifest(payload, args, channelID, inlineTable, plannedUploads), nil
}

func slackReportBaseManifest() map[string]interface{} {
	return map[string]interface{}{
		"report_schema_version": slackReportSchemaVersion,
		"renderer":              "rsi-slack-report",
		"decisions":             []string{},
		"fallback_reasons":      []string{},
	}
}

func slackReportInlineTable(payload slackReportPayload) bool {
	return len(payload.Tables) == 1 && len(payload.Tables[0].Columns) <= slackReportMaxTableCols && len(payload.Tables[0].Rows) <= slackReportMaxInlineRows
}

func slackReportFileFieldPath(idx int, fileCount int) (string, int) {
	if idx >= fileCount {
		return "images", idx - fileCount
	}
	return "files", idx
}

func slackReportAttachmentFilename(item slackReportFile, path string) string {
	if filename := strings.TrimSpace(item.Filename); filename != "" {
		return filename
	}
	if path = strings.TrimSpace(path); path != "" {
		if base := strings.TrimSpace(filepath.Base(path)); base != "" && base != "." {
			return base
		}
	}
	return "report-attachment"
}

func buildSlackReportManifest(payload slackReportPayload, args map[string]interface{}, channelID string, inlineTable bool, uploads []slackReportManifestUpload) map[string]interface{} {
	manifest := slackReportBaseManifest()
	if inlineTable {
		manifest["decisions"] = appendStringAny(manifest["decisions"], "rendered one table inline using Slack table block")
	} else if len(payload.Tables) > 0 {
		manifest["fallback_reasons"] = appendStringAny(manifest["fallback_reasons"], "large, wide, or multiple tables are uploaded as CSV artifacts")
	}
	plannedUploads := make([]map[string]interface{}, 0, len(uploads))
	for _, upload := range uploads {
		plannedUploads = append(plannedUploads, map[string]interface{}{
			"id":       upload.ID,
			"kind":     upload.Kind,
			"filename": upload.Filename,
			"status":   "pending",
		})
	}
	manifest["planned_uploads"] = len(plannedUploads)
	manifest["inline_table"] = inlineTable
	manifest["main_message"] = map[string]interface{}{
		"status":     "pending",
		"channel_id": channelID,
		"thread_ts":  stringArg(args, "thread_ts"),
	}
	manifest["uploads"] = plannedUploads
	return manifest
}

func slackReportPayloadFromArgs(args map[string]interface{}) (slackReportPayload, error) {
	if report := mapArg(args, "report"); report != nil {
		return slackReportPayloadFromMap(report)
	}
	return slackReportPayloadFromMap(args)
}

func slackReportPayloadFromMap(raw map[string]interface{}) (slackReportPayload, error) {
	var payload slackReportPayload
	data, err := json.Marshal(raw)
	if err != nil {
		return payload, err
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

func validateSlackReportPayload(payload slackReportPayload) slackReportValidationErrors {
	var errs slackReportValidationErrors
	if payload.ReportSchemaVersion != slackReportSchemaVersion {
		errs = append(errs, slackReportValidationError{Code: "unsupported_version", Path: "report_schema_version", Limit: slackReportSchemaVersion, Actual: payload.ReportSchemaVersion, Message: "report_schema_version must be 1"})
	}
	if strings.TrimSpace(payload.Summary) == "" {
		errs = append(errs, slackReportValidationError{Code: "required", Path: "summary", Message: "summary is required"})
	} else if len(payload.Summary) > slackReportMaxMarkdown {
		errs = append(errs, slackReportValidationError{Code: "too_long", Path: "summary", Limit: slackReportMaxMarkdown, Actual: len(payload.Summary), Message: "summary is too long"})
	}
	if containsMarkdownPipeTableOutsideFence(payload.Summary) {
		errs = append(errs, slackReportValidationError{Code: "unsupported_markdown_table", Path: "summary", Message: "use structured report tables instead of raw pipe tables"})
	}
	for i, section := range payload.Sections {
		if len(section.Text) > slackReportMaxMarkdown {
			errs = append(errs, slackReportValidationError{Code: "too_long", Path: fmt.Sprintf("sections[%d].text", i), Limit: slackReportMaxMarkdown, Actual: len(section.Text), Message: "section text is too long"})
		}
		if containsMarkdownPipeTableOutsideFence(section.Text) {
			errs = append(errs, slackReportValidationError{Code: "unsupported_markdown_table", Path: fmt.Sprintf("sections[%d].text", i), Message: "use structured report tables instead of raw pipe tables"})
		}
	}
	for i, table := range payload.Tables {
		if len(table.Columns) == 0 {
			errs = append(errs, slackReportValidationError{Code: "required", Path: fmt.Sprintf("tables[%d].columns", i), Message: "table columns are required"})
		}
		if len(table.Columns) > slackReportMaxTableCols {
			errs = append(errs, slackReportValidationError{Code: "too_many_columns", Path: fmt.Sprintf("tables[%d].columns", i), Limit: slackReportMaxTableCols, Actual: len(table.Columns), Message: "table has too many columns"})
		}
		seen := map[string]bool{}
		for j, col := range table.Columns {
			key := strings.TrimSpace(col.Key)
			if key == "" {
				errs = append(errs, slackReportValidationError{Code: "required", Path: fmt.Sprintf("tables[%d].columns[%d].key", i, j), Message: "column key is required"})
			}
			if seen[key] {
				errs = append(errs, slackReportValidationError{Code: "duplicate", Path: fmt.Sprintf("tables[%d].columns[%d].key", i, j), Actual: key, Message: "column keys must be unique"})
			}
			seen[key] = true
		}
		for rowIdx, row := range table.Rows {
			for _, col := range table.Columns {
				value := row[col.Key]
				if !isSlackReportScalar(value) {
					errs = append(errs, slackReportValidationError{Code: "invalid_cell_type", Path: fmt.Sprintf("tables[%d].rows[%d].%s", i, rowIdx, col.Key), Message: "table cells must be string, number, bool, or null"})
					continue
				}
				rendered := slackReportCellString(value)
				if len(rendered) > slackReportMaxCellLength {
					errs = append(errs, slackReportValidationError{Code: "cell_too_long", Path: fmt.Sprintf("tables[%d].rows[%d].%s", i, rowIdx, col.Key), Limit: slackReportMaxCellLength, Actual: len(rendered), Message: "table cell is too long"})
				}
			}
		}
	}
	return errs
}

func slackTableBlock(blockID string, table slackReportTable) slackapi.Block {
	block := slackapi.NewTableBlock(blockID)
	header := make([]*slackapi.RichTextBlock, 0, len(table.Columns))
	settings := make([]slackapi.ColumnSetting, 0, len(table.Columns))
	for _, col := range table.Columns {
		header = append(header, slackReportRichText(firstNonEmpty(col.Label, col.Key)))
		settings = append(settings, slackapi.ColumnSetting{Align: slackReportColumnAlignmentForTable(table, col), IsWrapped: true})
	}
	block.AddRow(header...)
	block.WithColumnSettings(settings...)
	for _, row := range table.Rows {
		cells := make([]*slackapi.RichTextBlock, 0, len(table.Columns))
		for _, col := range table.Columns {
			cells = append(cells, slackReportRichText(slackReportCellString(row[col.Key])))
		}
		block.AddRow(cells...)
	}
	return block
}

func slackReportRichText(text string) *slackapi.RichTextBlock {
	return slackapi.NewRichTextBlock("", slackapi.NewRichTextSection(slackapi.NewRichTextSectionTextElement(text, nil)))
}

func slackReportColumnAlignment(value string) slackapi.ColumnAlignment {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "right":
		return slackapi.ColumnAlignmentRight
	case "center", "centre":
		return slackapi.ColumnAlignmentCenter
	default:
		return slackapi.ColumnAlignmentLeft
	}
}

func slackReportColumnAlignmentForTable(table slackReportTable, col slackReportColumn) slackapi.ColumnAlignment {
	if strings.TrimSpace(col.Align) != "" {
		return slackReportColumnAlignment(col.Align)
	}
	if slackReportColumnLooksNumeric(table, col.Key) {
		return slackapi.ColumnAlignmentRight
	}
	return slackapi.ColumnAlignmentLeft
}

func slackReportColumnLooksNumeric(table slackReportTable, key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	seenNumeric := false
	for _, row := range table.Rows {
		value, ok := row[key]
		if !ok || value == nil {
			continue
		}
		rendered := strings.TrimSpace(slackReportCellString(value))
		if rendered == "" {
			continue
		}
		if !slackReportCellLooksNumeric(value) {
			return false
		}
		seenNumeric = true
	}
	return seenNumeric
}

func slackReportCellLooksNumeric(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}
	raw := strings.TrimSpace(slackReportCellString(value))
	if raw == "" {
		return false
	}
	if strings.HasPrefix(raw, "(") && strings.HasSuffix(raw, ")") {
		raw = "-" + strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "("), ")"))
	}
	raw = strings.ReplaceAll(raw, " ", "")
	if !slackReportNumericCellRegex.MatchString(raw) {
		return false
	}
	normalized := strings.TrimSuffix(strings.ReplaceAll(raw, ",", ""), "%")
	_, err := strconv.ParseFloat(normalized, 64)
	return err == nil
}

func slackReportTableCSV(table slackReportTable) (string, error) {
	var b strings.Builder
	writer := csv.NewWriter(&b)
	header := make([]string, len(table.Columns))
	for i, col := range table.Columns {
		header[i] = firstNonEmpty(col.Label, col.Key)
	}
	if err := writer.Write(header); err != nil {
		return "", err
	}
	for _, row := range table.Rows {
		record := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			record[i] = slackReportCellString(row[col.Key])
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return b.String(), nil
}

func isSlackReportScalar(value interface{}) bool {
	switch value.(type) {
	case nil, string, bool, float64, float32, int, int64, int32, uint, uint64, uint32, json.Number:
		return true
	default:
		return false
	}
}

func slackReportCellString(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func safeSlackReportFilename(value string, ext string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = safeFilenameRegex.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-_.")
	if value == "" {
		value = "report"
	}
	if !strings.HasSuffix(value, ext) {
		value += ext
	}
	return value
}

func appendStringAny(raw interface{}, value string) []string {
	var out []string
	switch items := raw.(type) {
	case []string:
		out = append(out, items...)
	case []interface{}:
		for _, item := range items {
			if text := strings.TrimSpace(fmt.Sprintf("%v", item)); text != "" {
				out = append(out, text)
			}
		}
	}
	return append(out, value)
}

func containsMarkdownPipeTableOutsideFence(text string) bool {
	inFence := false
	var previousPipe bool
	var previousSeparator bool
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			previousPipe = false
			previousSeparator = false
			continue
		}
		if inFence {
			continue
		}
		hasPipe := strings.Count(trimmed, "|") >= 2
		isSeparator := hasPipe && markdownTableSeparatorRegex.MatchString(trimmed)
		if (isSeparator && previousPipe) || (previousSeparator && hasPipe) {
			return true
		}
		previousPipe = hasPipe && !isSeparator
		previousSeparator = isSeparator
	}
	return false
}

func slackReportSummaryFromPayload(payload map[string]interface{}) string {
	if payload == nil {
		return ""
	}
	if summary := strings.TrimSpace(stringValueFromMap(payload, "summary")); summary != "" {
		return summary
	}
	if report := mapValue(payload["report"]); len(report) > 0 {
		return strings.TrimSpace(stringValueFromMap(report, "summary"))
	}
	return ""
}
