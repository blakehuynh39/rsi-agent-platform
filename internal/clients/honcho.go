package clients

import (
	"net/http"
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
		baseURL:    trimBaseURL(baseURL),
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
