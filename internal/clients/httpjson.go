package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxHTTPErrorBodyBytes = 4096

type HTTPStatusError struct {
	Service    string
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Body) != "" {
		return fmt.Sprintf("%s returned %d: %s", e.Service, e.StatusCode, strings.TrimSpace(e.Body))
	}
	return fmt.Sprintf("%s returned %d", e.Service, e.StatusCode)
}

func HTTPStatusCode(err error) int {
	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) && statusErr != nil {
		return statusErr.StatusCode
	}
	return 0
}

func newHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &http.Client{Timeout: timeout}
}

func trimBaseURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/")
}

func doJSON(client *http.Client, method string, url string, payload any, out any, service string) error {
	return doJSONWithHeaders(client, method, url, payload, out, service, nil)
}

func doJSONWithHeaders(client *http.Client, method string, url string, payload any, out any, service string, headers map[string]string) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPErrorBodyBytes))
		return &HTTPStatusError{
			Service:    service,
			StatusCode: resp.StatusCode,
			Body:       string(raw),
		}
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
