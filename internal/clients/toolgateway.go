package clients

import (
	"net/http"
	"time"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type ToolGatewayClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewToolGatewayClient(baseURL string) *ToolGatewayClient {
	return &ToolGatewayClient{
		baseURL:    trimBaseURL(baseURL),
		httpClient: newHTTPClient(30 * time.Second),
	}
}

func (c *ToolGatewayClient) Execute(toolName string, input map[string]any) (storepkg.ToolResult, error) {
	var out storepkg.ToolResult
	if err := doJSON(c.httpClient, http.MethodPost, c.baseURL+"/api/tools/"+toolName+"/execute", input, &out, "tool gateway"); err != nil {
		return storepkg.ToolResult{}, err
	}
	return out, nil
}
