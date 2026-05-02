package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultNotionVersion = "2022-06-28"

type NotionClient struct {
	baseURL    string
	token      string
	version    string
	httpClient *http.Client
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

type NotionDatabase struct {
	Object         string       `json:"object"`
	ID             string       `json:"id"`
	URL            string       `json:"url"`
	Title          []NotionText `json:"title"`
	LastEditedTime string       `json:"last_edited_time"`
	Archived       bool         `json:"archived"`
	InTrash        bool         `json:"in_trash"`
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

func NewNotionClient(token string) *NotionClient {
	return NewNotionClientWithOptions("https://api.notion.com", token, defaultNotionVersion)
}

func NewNotionClientWithOptions(baseURL string, token string, version string) *NotionClient {
	baseURL = trimBaseURL(baseURL)
	if baseURL == "" {
		baseURL = "https://api.notion.com"
	}
	version = strings.TrimSpace(version)
	if version == "" {
		version = defaultNotionVersion
	}
	return &NotionClient{
		baseURL:    baseURL,
		token:      strings.TrimSpace(token),
		version:    version,
		httpClient: newHTTPClient(30 * time.Second),
	}
}

func (c *NotionClient) RetrievePage(ctx context.Context, pageID string) (NotionPage, error) {
	var out NotionPage
	if err := c.doJSON(ctx, http.MethodGet, "/v1/pages/"+strings.TrimSpace(pageID), nil, &out); err != nil {
		return NotionPage{}, err
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

func (c *NotionClient) doJSON(ctx context.Context, method string, path string, payload any, out any) error {
	if strings.TrimSpace(c.token) == "" {
		return fmt.Errorf("NOTION_TOKEN is required")
	}
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
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
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return NotionAPIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}
	return json.NewDecoder(resp.Body).Decode(out)
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
