package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type ToolGatewayClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewToolGatewayClient(baseURL string) *ToolGatewayClient {
	return &ToolGatewayClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ToolGatewayClient) Execute(toolName string, input map[string]any) (storepkg.ToolResult, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return storepkg.ToolResult{}, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/tools/"+toolName+"/execute", bytes.NewReader(body))
	if err != nil {
		return storepkg.ToolResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return storepkg.ToolResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return storepkg.ToolResult{}, fmt.Errorf("tool gateway returned %d", resp.StatusCode)
	}
	var out storepkg.ToolResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return storepkg.ToolResult{}, err
	}
	return out, nil
}
