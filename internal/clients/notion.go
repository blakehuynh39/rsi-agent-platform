package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultNotionVersion = "2026-03-11"

type NotionClient struct {
	baseURL           string
	token             string
	version           string
	httpClient        *http.Client
	requestsPerSecond float64
	maxRetries        int
	retryBaseDelay    time.Duration
	rateMu            sync.Mutex
	lastRequest       time.Time
}

type NotionClientOptions struct {
	BaseURL           string
	Token             string
	Version           string
	RequestsPerSecond float64
	MaxRetries        int
	RetryBaseDelay    time.Duration
}

type NotionAPIError struct {
	StatusCode int
	Body       string
}

func (e NotionAPIError) Error() string {
	return fmt.Sprintf("notion returned %d: %s", e.StatusCode, strings.TrimSpace(e.Body))
}

type NotionPage struct {
	Object         string         `json:"object"`
	ID             string         `json:"id"`
	URL            string         `json:"url"`
	Parent         map[string]any `json:"parent"`
	Properties     map[string]any `json:"properties"`
	LastEditedTime string         `json:"last_edited_time"`
	CreatedTime    string         `json:"created_time"`
	Archived       bool           `json:"archived"`
	InTrash        bool           `json:"in_trash"`
}

type NotionDataSourceReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NotionDatabase struct {
	Object         string                      `json:"object"`
	ID             string                      `json:"id"`
	URL            string                      `json:"url"`
	Title          []NotionText                `json:"title"`
	Parent         map[string]any              `json:"parent"`
	DataSources    []NotionDataSourceReference `json:"data_sources"`
	Properties     map[string]any              `json:"properties"`
	LastEditedTime string                      `json:"last_edited_time"`
	CreatedTime    string                      `json:"created_time"`
	Archived       bool                        `json:"archived"`
	InTrash        bool                        `json:"in_trash"`
	Raw            map[string]any              `json:"-"`
}

type NotionDataSource struct {
	Object         string         `json:"object"`
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	URL            string         `json:"url"`
	Title          []NotionText   `json:"title"`
	Parent         map[string]any `json:"parent"`
	DatabaseParent map[string]any `json:"database_parent"`
	Properties     map[string]any `json:"properties"`
	LastEditedTime string         `json:"last_edited_time"`
	CreatedTime    string         `json:"created_time"`
	Archived       bool           `json:"archived"`
	InTrash        bool           `json:"in_trash"`
	Raw            map[string]any `json:"-"`
}

type NotionBlock struct {
	Object         string         `json:"object"`
	ID             string         `json:"id"`
	Type           string         `json:"type"`
	HasChildren    bool           `json:"has_children"`
	CreatedTime    string         `json:"created_time"`
	LastEditedTime string         `json:"last_edited_time"`
	Archived       bool           `json:"archived"`
	InTrash        bool           `json:"in_trash"`
	Raw            map[string]any `json:"-"`
}

type NotionText struct {
	PlainText string `json:"plain_text"`
	Href      string `json:"href"`
}

type NotionListResponse[T any] struct {
	Object     string `json:"object"`
	Results    []T    `json:"results"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

type NotionDataSourceQueryOptions struct {
	Cursor                  string
	PageSize                int
	LastEditedTimeOnOrAfter string
	SortTimestamp           string
	SortDirection           string
	FilterProperties        []string
}

type NotionSearchOptions struct {
	Query    string
	Filter   map[string]any
	Sort     map[string]any
	PageSize int
	Cursor   string
}

type NotionPageMarkdown struct {
	Object          string   `json:"object"`
	ID              string   `json:"id"`
	Markdown        string   `json:"markdown"`
	Truncated       bool     `json:"truncated"`
	UnknownBlockIDs []string `json:"unknown_block_ids"`
}

func NewNotionClient(token string) *NotionClient {
	return NewNotionClientWithOptions("https://api.notion.com", token, defaultNotionVersion)
}

func NewNotionClientWithOptions(baseURL string, token string, version string) *NotionClient {
	return NewNotionClientWithConfig(NotionClientOptions{
		BaseURL: baseURL,
		Token:   token,
		Version: version,
	})
}

func NewNotionClientWithConfig(options NotionClientOptions) *NotionClient {
	baseURL := trimBaseURL(options.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.notion.com"
	}
	version := strings.TrimSpace(options.Version)
	if version == "" {
		version = defaultNotionVersion
	}
	maxRetries := options.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	retryBaseDelay := options.RetryBaseDelay
	if retryBaseDelay <= 0 {
		retryBaseDelay = 500 * time.Millisecond
	}
	return &NotionClient{
		baseURL:           baseURL,
		token:             strings.TrimSpace(options.Token),
		version:           version,
		httpClient:        newHTTPClient(30 * time.Second),
		requestsPerSecond: options.RequestsPerSecond,
		maxRetries:        maxRetries,
		retryBaseDelay:    retryBaseDelay,
	}
}

func (c *NotionClient) RetrievePage(ctx context.Context, pageID string) (NotionPage, error) {
	var out NotionPage
	if err := c.doJSON(ctx, http.MethodGet, "/v1/pages/"+strings.TrimSpace(pageID), nil, &out); err != nil {
		return NotionPage{}, err
	}
	return out, nil
}

func (c *NotionClient) Search(ctx context.Context, opts NotionSearchOptions) (map[string]any, error) {
	payload := map[string]any{}
	if query := strings.TrimSpace(opts.Query); query != "" {
		payload["query"] = query
	}
	if len(opts.Filter) > 0 {
		payload["filter"] = opts.Filter
	}
	if len(opts.Sort) > 0 {
		payload["sort"] = opts.Sort
	}
	if opts.PageSize > 0 {
		payload["page_size"] = opts.PageSize
	}
	if cursor := strings.TrimSpace(opts.Cursor); cursor != "" {
		payload["start_cursor"] = cursor
	}
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodPost, "/v1/search", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *NotionClient) RetrieveDatabase(ctx context.Context, databaseID string) (NotionDatabase, error) {
	var out NotionDatabase
	if err := c.doJSON(ctx, http.MethodGet, "/v1/databases/"+strings.TrimSpace(databaseID), nil, &out); err != nil {
		return NotionDatabase{}, err
	}
	return out, nil
}

func (c *NotionClient) RetrieveDataSource(ctx context.Context, dataSourceID string) (NotionDataSource, error) {
	var out NotionDataSource
	if err := c.doJSON(ctx, http.MethodGet, "/v1/data_sources/"+strings.TrimSpace(dataSourceID), nil, &out); err != nil {
		return NotionDataSource{}, err
	}
	return out, nil
}

func (c *NotionClient) RetrievePageMarkdown(ctx context.Context, pageID string, includeTranscript bool) (NotionPageMarkdown, error) {
	path := "/v1/pages/" + strings.TrimSpace(pageID) + "/markdown"
	if includeTranscript {
		path += "?include_transcript=true"
	}
	var out NotionPageMarkdown
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return NotionPageMarkdown{}, err
	}
	return out, nil
}

func (c *NotionClient) ListBlockChildren(ctx context.Context, blockID string, cursor string, pageSize int) (NotionListResponse[NotionBlock], error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100
	}
	path := fmt.Sprintf("/v1/blocks/%s/children?page_size=%d", strings.TrimSpace(blockID), pageSize)
	if strings.TrimSpace(cursor) != "" {
		path += "&start_cursor=" + strings.TrimSpace(cursor)
	}
	var out NotionListResponse[NotionBlock]
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return NotionListResponse[NotionBlock]{}, err
	}
	return out, nil
}

func (c *NotionClient) QueryDataSource(ctx context.Context, dataSourceID string, opts NotionDataSourceQueryOptions) (NotionListResponse[NotionPage], error) {
	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100
	}
	payload := map[string]any{"page_size": pageSize}
	if strings.TrimSpace(opts.Cursor) != "" {
		payload["start_cursor"] = strings.TrimSpace(opts.Cursor)
	}
	if strings.TrimSpace(opts.LastEditedTimeOnOrAfter) != "" {
		payload["filter"] = map[string]any{
			"timestamp": "last_edited_time",
			"last_edited_time": map[string]any{
				"on_or_after": strings.TrimSpace(opts.LastEditedTimeOnOrAfter),
			},
		}
	}
	sortTimestamp := strings.TrimSpace(opts.SortTimestamp)
	if sortTimestamp != "" {
		sortDirection := strings.TrimSpace(opts.SortDirection)
		if sortDirection == "" {
			sortDirection = "ascending"
		}
		payload["sorts"] = []map[string]any{{
			"timestamp": sortTimestamp,
			"direction": sortDirection,
		}}
	}
	path := "/v1/data_sources/" + strings.TrimSpace(dataSourceID) + "/query"
	if len(opts.FilterProperties) > 0 {
		values := url.Values{}
		for _, property := range opts.FilterProperties {
			if property = strings.TrimSpace(property); property != "" {
				values.Add("filter_properties[]", property)
			}
		}
		if encoded := values.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}
	var out NotionListResponse[NotionPage]
	if err := c.doJSON(ctx, http.MethodPost, path, payload, &out); err != nil {
		return NotionListResponse[NotionPage]{}, err
	}
	return out, nil
}

func (c *NotionClient) QueryDatabase(ctx context.Context, databaseID string, cursor string, pageSize int) (NotionListResponse[NotionPage], error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100
	}
	payload := map[string]any{"page_size": pageSize}
	if strings.TrimSpace(cursor) != "" {
		payload["start_cursor"] = strings.TrimSpace(cursor)
	}
	var out NotionListResponse[NotionPage]
	if err := c.doJSON(ctx, http.MethodPost, "/v1/databases/"+strings.TrimSpace(databaseID)+"/query", payload, &out); err != nil {
		return NotionListResponse[NotionPage]{}, err
	}
	return out, nil
}

func (c *NotionClient) CreatePage(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodPost, "/v1/pages", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *NotionClient) UpdatePage(ctx context.Context, pageID string, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodPatch, "/v1/pages/"+strings.TrimSpace(pageID), payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *NotionClient) AppendBlockChildren(ctx context.Context, blockID string, children []any) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodPatch, "/v1/blocks/"+strings.TrimSpace(blockID)+"/children", map[string]any{"children": children}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *NotionClient) UpdateBlock(ctx context.Context, blockID string, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodPatch, "/v1/blocks/"+strings.TrimSpace(blockID), payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *NotionClient) DeleteBlock(ctx context.Context, blockID string) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodDelete, "/v1/blocks/"+strings.TrimSpace(blockID), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *NotionClient) CreateComment(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodPost, "/v1/comments", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *NotionClient) doJSON(ctx context.Context, method string, path string, payload any, out any) error {
	if strings.TrimSpace(c.token) == "" {
		return fmt.Errorf("NOTION_TOKEN is required")
	}
	var rawPayload []byte
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		rawPayload = encoded
	}
	attempts := c.maxRetries + 1
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if err := c.waitForRateLimit(ctx); err != nil {
			return err
		}
		var body io.Reader
		if rawPayload != nil {
			body = bytes.NewReader(rawPayload)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Notion-Version", c.version)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "rsi-company-knowledge-notion-mirror/1.0")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt+1 < attempts {
				if waitErr := sleepWithContext(ctx, c.retryDelay(attempt, "")); waitErr != nil {
					return waitErr
				}
				continue
			}
			return err
		}
		if resp.StatusCode >= 300 {
			raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			apiErr := NotionAPIError{StatusCode: resp.StatusCode, Body: string(raw)}
			retryAfter := resp.Header.Get("Retry-After")
			_ = resp.Body.Close()
			lastErr = apiErr
			if notionRetryableStatus(resp.StatusCode) && attempt+1 < attempts {
				if waitErr := sleepWithContext(ctx, c.retryDelay(attempt, retryAfter)); waitErr != nil {
					return waitErr
				}
				continue
			}
			return apiErr
		}
		err = json.NewDecoder(resp.Body).Decode(out)
		_ = resp.Body.Close()
		return err
	}
	return lastErr
}

func (c *NotionClient) waitForRateLimit(ctx context.Context) error {
	if c.requestsPerSecond <= 0 {
		return nil
	}
	interval := time.Duration(float64(time.Second) / c.requestsPerSecond)
	c.rateMu.Lock()
	wait := time.Duration(0)
	now := time.Now()
	if !c.lastRequest.IsZero() {
		next := c.lastRequest.Add(interval)
		if now.Before(next) {
			wait = next.Sub(now)
		}
	}
	c.rateMu.Unlock()
	if wait > 0 {
		if err := sleepWithContext(ctx, wait); err != nil {
			return err
		}
	}
	c.rateMu.Lock()
	c.lastRequest = time.Now()
	c.rateMu.Unlock()
	return nil
}

func (c *NotionClient) retryDelay(attempt int, retryAfter string) time.Duration {
	if delay := parseRetryAfter(retryAfter); delay > 0 {
		return delay
	}
	delay := c.retryBaseDelay
	if attempt > 0 {
		delay *= time.Duration(1 << attempt)
	}
	return delay
}

func notionRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

func parseRetryAfter(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(raw); err == nil {
		delay := time.Until(when)
		if delay > 0 {
			return delay
		}
	}
	return 0
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (b *NotionBlock) UnmarshalJSON(raw []byte) error {
	type blockAlias NotionBlock
	var alias blockAlias
	if err := json.Unmarshal(raw, &alias); err != nil {
		return err
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return err
	}
	*b = NotionBlock(alias)
	b.Raw = object
	return nil
}

func (d *NotionDatabase) UnmarshalJSON(raw []byte) error {
	type databaseAlias NotionDatabase
	var alias databaseAlias
	if err := json.Unmarshal(raw, &alias); err != nil {
		return err
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return err
	}
	*d = NotionDatabase(alias)
	d.Raw = object
	return nil
}

func (d *NotionDataSource) UnmarshalJSON(raw []byte) error {
	type dataSourceAlias NotionDataSource
	var alias dataSourceAlias
	if err := json.Unmarshal(raw, &alias); err != nil {
		return err
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return err
	}
	*d = NotionDataSource(alias)
	d.Raw = object
	return nil
}
