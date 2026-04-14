package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type HonchoRuntimeComponent struct {
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
}

type HonchoDialecticLevel struct {
	Provider             string `json:"provider"`
	Model                string `json:"model"`
	ReasoningEffort      string `json:"reasoning_effort"`
	ThinkingBudgetTokens int    `json:"thinking_budget_tokens"`
}

type HonchoRuntimeResponse struct {
	Status             string                          `json:"status"`
	Namespace          string                          `json:"namespace"`
	DBSchema           string                          `json:"db_schema"`
	CacheEnabled       bool                            `json:"cache_enabled"`
	CacheURLConfigured bool                            `json:"cache_url_configured"`
	Deriver            HonchoRuntimeComponent          `json:"deriver"`
	Summary            HonchoRuntimeComponent          `json:"summary"`
	DialecticLevels    map[string]HonchoDialecticLevel `json:"dialectic_levels"`
}

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
