package control

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

var errNoReadyHermesExecutorEndpoints = errors.New("no ready Hermes executor endpoints")

type hermesExecutorEndpoint struct {
	instanceID string
	baseURL    string
	client     *clients.RunnerClient
	ready      clients.RuntimeResponse
}

type hermesExecutorPool struct {
	cfg            config.Config
	role           string
	fallbackClient *clients.RunnerClient
}

func newHermesExecutorPool(cfg config.Config, role string, fallbackClient *clients.RunnerClient) hermesExecutorPool {
	return hermesExecutorPool{
		cfg:            cfg,
		role:           role,
		fallbackClient: fallbackClient,
	}
}

func (p hermesExecutorPool) fallbackBaseURL() string {
	if p.fallbackClient != nil && strings.TrimSpace(p.fallbackClient.BaseURL()) != "" {
		return p.fallbackClient.BaseURL()
	}
	urls := p.cfg.HermesExecutorURLs()
	if len(urls) == 0 {
		return ""
	}
	return urls[0]
}

func (p hermesExecutorPool) clientForRecord(record storepkg.RunnerExecution) *clients.RunnerClient {
	baseURL := strings.TrimSpace(record.ExecutorBaseURL)
	if baseURL == "" {
		baseURL = p.fallbackBaseURL()
	}
	if baseURL == "" {
		return nil
	}
	if p.fallbackClient != nil && strings.TrimSpace(p.fallbackClient.BaseURL()) == baseURL {
		return p.fallbackClient
	}
	return clients.NewRunnerClientWithTimeout(baseURL, p.cfg.RunnerTimeoutForRole(firstNonEmpty(record.Role, p.role)))
}

func (p hermesExecutorPool) startExecution(task clients.RunnerTask) (clients.HermesExecutionStatus, hermesExecutorEndpoint, error) {
	candidates := p.readyCandidates(task.ExecutionID)
	if len(candidates) == 0 {
		return clients.HermesExecutionStatus{}, hermesExecutorEndpoint{}, errNoReadyHermesExecutorEndpoints
	}
	var lastErr error
	for _, endpoint := range candidates {
		status, err := endpoint.client.StartHermesExecution(task)
		if err == nil {
			return status, endpoint, nil
		}
		lastErr = err
		if !hermesStartRetryableBeforeAccept(err) {
			break
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no Hermes executor endpoints attempted")
	}
	return clients.HermesExecutionStatus{}, hermesExecutorEndpoint{}, lastErr
}

func (p hermesExecutorPool) readyCandidates(executionID string) []hermesExecutorEndpoint {
	urls := p.cfg.HermesExecutorURLs()
	if len(urls) == 0 && p.fallbackClient != nil && strings.TrimSpace(p.fallbackClient.BaseURL()) != "" {
		urls = []string{p.fallbackClient.BaseURL()}
	}
	urls = config.CompactUniqueStrings(urls)
	if len(urls) == 0 {
		return nil
	}
	strictReady := len(p.cfg.HermesExecutorPoolURLs) > 0
	rotated := rotateStrings(urls, stableIndex(executionID, len(urls)))
	out := make([]hermesExecutorEndpoint, 0, len(rotated))
	for _, baseURL := range rotated {
		client := p.fallbackClient
		if client == nil || strings.TrimSpace(client.BaseURL()) != strings.TrimSpace(baseURL) {
			client = clients.NewRunnerClientWithTimeout(baseURL, p.cfg.RunnerTimeoutForRole(p.role))
		}
		if !strictReady {
			out = append(out, hermesExecutorEndpoint{
				instanceID: instanceIDFromBaseURL(baseURL),
				baseURL:    strings.TrimSpace(baseURL),
				client:     client,
			})
			continue
		}
		readinessClient := clients.NewRunnerClientWithTimeout(baseURL, 10*time.Second)
		ready, err := readinessClient.Ready()
		if err != nil || !hermesExecutorReady(ready) {
			continue
		}
		out = append(out, hermesExecutorEndpoint{
			instanceID: firstNonEmpty(ready.ExecutorInstanceID, instanceIDFromBaseURL(baseURL)),
			baseURL:    strings.TrimSpace(baseURL),
			client:     client,
			ready:      ready,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ready.ActiveExecutionCount < out[j].ready.ActiveExecutionCount
	})
	return out
}

func hermesExecutorReady(ready clients.RuntimeResponse) bool {
	drainStatus := strings.ToLower(strings.TrimSpace(ready.DrainStatus))
	status := strings.ToLower(strings.TrimSpace(ready.Status))
	return ready.Available && drainStatus != "draining" && status != "degraded"
}

func hermesStartRetryableBeforeAccept(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "returned 503") ||
		strings.Contains(text, "connection refused") ||
		strings.Contains(text, "no such host")
}

func stableIndex(seed string, length int) int {
	if length <= 1 {
		return 0
	}
	sum := sha1.Sum([]byte(strings.TrimSpace(seed)))
	return int(sum[0]) % length
}

func rotateStrings(values []string, offset int) []string {
	if len(values) == 0 {
		return nil
	}
	offset = offset % len(values)
	if offset == 0 {
		return append([]string(nil), values...)
	}
	out := append([]string(nil), values[offset:]...)
	out = append(out, values[:offset]...)
	return out
}

func instanceIDFromBaseURL(baseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err == nil && strings.TrimSpace(parsed.Hostname()) != "" {
		return strings.TrimSpace(parsed.Hostname())
	}
	sum := sha1.Sum([]byte(strings.TrimSpace(baseURL)))
	return "executor-" + hex.EncodeToString(sum[:])[:8]
}
