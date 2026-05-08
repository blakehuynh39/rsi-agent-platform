package clients

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/harness"
)

type HonchoRuntimeComponent = harness.RuntimeComponent
type HonchoDialecticLevel = harness.DialecticLevel
type HonchoRuntimeResponse = harness.HonchoRuntimeResponse

type HonchoClient struct {
	baseURL    string
	apiBaseURL string
	apiKey     string
	httpClient *http.Client
}

func NewHonchoClient(baseURL string) *HonchoClient {
	return NewHonchoClientWithAPIKey(baseURL, "")
}

func NewHonchoClientWithAPIKey(baseURL string, apiKey string) *HonchoClient {
	baseURL = trimBaseURL(baseURL)
	return &HonchoClient{
		baseURL:    baseURL,
		apiBaseURL: honchoAPIBaseURL(baseURL),
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: newHTTPClient(15 * time.Second),
	}
}

func (c *HonchoClient) Runtime() (HonchoRuntimeResponse, error) {
	var out HonchoRuntimeResponse
	if err := doJSON(c.httpClient, http.MethodGet, c.baseURL+"/runtimez", nil, &out, "honcho"); err != nil {
		return HonchoRuntimeResponse{}, err
	}
	if out.DialecticLevels == nil {
		out.DialecticLevels = map[string]HonchoDialecticLevel{}
	}
	return out, nil
}

type HonchoWorkspace struct {
	ID        string         `json:"id"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
}

type HonchoSession struct {
	ID          string         `json:"id"`
	WorkspaceID string         `json:"workspace_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at,omitempty"`
}

type HonchoPeer struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	WorkspaceID string         `json:"workspace_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at,omitempty"`
}

type HonchoMessage struct {
	ID          string         `json:"id"`
	Content     string         `json:"content"`
	PeerID      string         `json:"peer_id"`
	SessionID   string         `json:"session_id,omitempty"`
	WorkspaceID string         `json:"workspace_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at,omitempty"`
	TokenCount  int            `json:"token_count,omitempty"`
}

type HonchoPage[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Size  int `json:"size"`
	Pages int `json:"pages"`
	Total int `json:"total"`
}

type HonchoMessageCreate struct {
	Content   string         `json:"content"`
	PeerID    string         `json:"peer_id"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
}

type HonchoConclusion struct {
	ID         string    `json:"id"`
	Content    string    `json:"content"`
	ObserverID string    `json:"observer_id"`
	ObservedID string    `json:"observed_id"`
	SessionID  string    `json:"session_id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

type HonchoConclusionCreate struct {
	Content    string `json:"content"`
	ObserverID string `json:"observer_id"`
	ObservedID string `json:"observed_id"`
	SessionID  string `json:"session_id,omitempty"`
}

func (c *HonchoClient) EnsureWorkspace(id string, metadata map[string]any) (HonchoWorkspace, error) {
	var out HonchoWorkspace
	payload := map[string]any{
		"id":       id,
		"metadata": metadata,
	}
	if err := c.doJSON(http.MethodPost, c.apiBaseURL+"/workspaces", payload, &out); err != nil {
		return HonchoWorkspace{}, err
	}
	return out, nil
}

func (c *HonchoClient) EnsureSession(workspaceID string, sessionID string, metadata map[string]any) (HonchoSession, error) {
	var out HonchoSession
	payload := map[string]any{
		"id":       sessionID,
		"metadata": metadata,
	}
	if err := c.doJSON(http.MethodPost, c.apiBaseURL+"/workspaces/"+workspaceID+"/sessions", payload, &out); err != nil {
		return HonchoSession{}, err
	}
	return out, nil
}

func (c *HonchoClient) EnsurePeer(workspaceID string, peerID string, metadata map[string]any) (HonchoPeer, error) {
	var out HonchoPeer
	payload := map[string]any{
		"name":     peerID,
		"metadata": metadata,
	}
	if err := c.doJSON(http.MethodPost, c.apiBaseURL+"/workspaces/"+workspaceID+"/peers", payload, &out); err != nil {
		return HonchoPeer{}, err
	}
	return out, nil
}

func (c *HonchoClient) CreateMessages(workspaceID string, sessionID string, messages []HonchoMessageCreate) ([]HonchoMessage, error) {
	var out []HonchoMessage
	payload := map[string]any{"messages": messages}
	if err := c.doJSON(http.MethodPost, c.apiBaseURL+"/workspaces/"+workspaceID+"/sessions/"+sessionID+"/messages", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *HonchoClient) CreateConclusions(workspaceID string, conclusions []HonchoConclusionCreate) ([]HonchoConclusion, error) {
	var out []HonchoConclusion
	for _, peerID := range honchoConclusionPeerIDs(conclusions) {
		if _, err := c.EnsurePeer(workspaceID, peerID, nil); err != nil {
			return nil, err
		}
	}
	payload := map[string]any{"conclusions": conclusions}
	if err := c.doJSON(http.MethodPost, c.apiBaseURL+"/workspaces/"+workspaceID+"/conclusions", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func honchoConclusionPeerIDs(conclusions []HonchoConclusionCreate) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, conclusion := range conclusions {
		for _, peerID := range []string{conclusion.ObserverID, conclusion.ObservedID} {
			peerID = strings.TrimSpace(peerID)
			if peerID == "" {
				continue
			}
			if _, ok := seen[peerID]; ok {
				continue
			}
			seen[peerID] = struct{}{}
			out = append(out, peerID)
		}
	}
	return out
}

func (c *HonchoClient) ListMessages(workspaceID string, sessionID string, limit int, page int, reverse bool) (HonchoPage[HonchoMessage], error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	url := c.apiBaseURL + "/workspaces/" + workspaceID + "/sessions/" + sessionID + "/messages/list?size=" + intString(limit) + "&page=" + intString(page)
	if reverse {
		url += "&reverse=true"
	}
	var out HonchoPage[HonchoMessage]
	if err := c.doJSON(http.MethodPost, url, map[string]any{}, &out); err != nil {
		return HonchoPage[HonchoMessage]{}, err
	}
	return out, nil
}

func (c *HonchoClient) QueryConclusions(workspaceID string, query string, filters map[string]any, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}
	payload := map[string]any{
		"query":   strings.TrimSpace(query),
		"top_k":   limit,
		"filters": filters,
	}
	var out []map[string]any
	if err := c.doJSON(http.MethodPost, c.apiBaseURL+"/workspaces/"+workspaceID+"/conclusions/query", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *HonchoClient) ListConclusions(workspaceID string, filters map[string]any, limit int, page int) (HonchoPage[map[string]any], error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	url := c.apiBaseURL + "/workspaces/" + workspaceID + "/conclusions/list?size=" + intString(limit) + "&page=" + intString(page)
	var out HonchoPage[map[string]any]
	if err := c.doJSON(http.MethodPost, url, map[string]any{"filters": filters}, &out); err != nil {
		return HonchoPage[map[string]any]{}, err
	}
	return out, nil
}

func (c *HonchoClient) SearchMessages(workspaceID string, query string, filters map[string]any, limit int) ([]HonchoMessage, error) {
	if limit <= 0 {
		limit = 10
	}
	payload := map[string]any{
		"query":   query,
		"filters": filters,
		"limit":   limit,
	}
	var out []HonchoMessage
	if err := c.doJSON(http.MethodPost, c.apiBaseURL+"/workspaces/"+workspaceID+"/search", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *HonchoClient) doJSON(method string, url string, payload any, out any) error {
	headers := map[string]string{}
	if c.apiKey != "" {
		if strings.HasPrefix(strings.ToLower(c.apiKey), "bearer ") {
			headers["Authorization"] = c.apiKey
		} else {
			headers["Authorization"] = "Bearer " + c.apiKey
		}
	}
	return doJSONWithHeaders(c.httpClient, method, url, payload, out, "honcho", headers)
}

func intString(value int) string {
	return strconv.Itoa(value)
}

func honchoAPIBaseURL(baseURL string) string {
	baseURL = trimBaseURL(baseURL)
	if strings.HasSuffix(baseURL, "/v3") {
		return baseURL
	}
	return baseURL + "/v3"
}
