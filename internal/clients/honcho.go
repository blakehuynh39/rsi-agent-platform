package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/harness"
)

type HonchoRuntimeComponent = harness.RuntimeComponent
type HonchoDialecticLevel = harness.DialecticLevel
type HonchoRuntimeResponse = harness.HonchoRuntimeResponse

type HonchoClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHonchoClient(baseURL string) *HonchoClient {
	return &HonchoClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *HonchoClient) Runtime() (HonchoRuntimeResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/runtimez", nil)
	if err != nil {
		return HonchoRuntimeResponse{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return HonchoRuntimeResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return HonchoRuntimeResponse{}, fmt.Errorf("honcho returned %d", resp.StatusCode)
	}
	var out HonchoRuntimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return HonchoRuntimeResponse{}, err
	}
	if out.DialecticLevels == nil {
		out.DialecticLevels = map[string]HonchoDialecticLevel{}
	}
	return out, nil
}
