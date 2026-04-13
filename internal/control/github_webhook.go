package control

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

var proposalTitlePattern = regexp.MustCompile(`(?i)\brsi proposal\s+([a-z0-9._-]+)\b`)

type gitHubPullRequestWebhook struct {
	Action     string `json:"action"`
	Repository struct {
		FullName string `json:"full_name"`
		Name     string `json:"name"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
	PullRequest struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
		State   string `json:"state"`
		Merged  bool   `json:"merged"`
		Title   string `json:"title"`
		Head    struct {
			Ref string `json:"ref"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
}

func handleGitHubWebhook(cfg config.Config, store storepkg.Store, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := verifyGitHubWebhook(cfg, body, r.Header.Get("X-Hub-Signature-256")); err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, errGitHubWebhookUnavailable) {
			status = http.StatusServiceUnavailable
		}
		http.Error(w, err.Error(), status)
		return
	}

	eventType := strings.TrimSpace(r.Header.Get("X-GitHub-Event"))
	deliveryID := strings.TrimSpace(r.Header.Get("X-GitHub-Delivery"))
	if deliveryID == "" {
		http.Error(w, "missing X-GitHub-Delivery", http.StatusBadRequest)
		return
	}
	if eventType != "pull_request" {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"accepted":true,"linked":false,"ignored":true}`))
		return
	}

	var payload gitHubPullRequestWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	actionName := strings.ToLower(strings.TrimSpace(payload.Action))
	switch actionName {
	case "opened", "reopened", "closed":
	default:
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"accepted":true,"linked":false,"ignored":true}`))
		return
	}

	proposalID, linked := resolveProposalIDForWebhook(store, payload.PullRequest.Head.Ref, payload.PullRequest.Title)
	metadata := map[string]any{
		"event_type":                    eventType,
		"action":                        actionName,
		"repository":                    firstNonEmpty(payload.Repository.FullName, payload.Repository.Name),
		"number":                        payload.PullRequest.Number,
		"html_url":                      payload.PullRequest.HTMLURL,
		"pr_url":                        payload.PullRequest.HTMLURL,
		"state":                         payload.PullRequest.State,
		"merged":                        payload.PullRequest.Merged,
		"head_ref":                      payload.PullRequest.Head.Ref,
		"base_ref":                      payload.PullRequest.Base.Ref,
		"sender_login":                  payload.Sender.Login,
		"skip_workflow_materialization": true,
	}
	if linked {
		metadata["proposal_id"] = proposalID
	}

	event := ingestion.EventEnvelope{
		Source:                     ingestion.SourceGitHub,
		SourceEventID:              deliveryID,
		DedupeKey:                  fmt.Sprintf("github:%s", deliveryID),
		Severity:                   ingestion.SeverityInfo,
		NormalizedProblemStatement: githubWebhookSummary(payload),
		OwnershipHint:              payload.Repository.Name,
		RawPayloadRef:              fmt.Sprintf("memory://github/%s.json", deliveryID),
		WorkflowHint:               "",
		Metadata:                   metadata,
		CreatedAt:                  time.Now().UTC(),
	}
	created, err := store.CreateEvent(event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"accepted": true,
		"linked":   linked,
		"event_id": created.ID,
	})
}

var errGitHubWebhookUnavailable = errors.New("github webhook secret not configured")

func verifyGitHubWebhook(cfg config.Config, body []byte, signature string) error {
	secret := strings.TrimSpace(cfg.GitHubWebhookSecret)
	if secret == "" {
		if strings.EqualFold(cfg.Environment, "development") {
			return nil
		}
		return errGitHubWebhookUnavailable
	}
	signature = strings.TrimSpace(signature)
	if !strings.HasPrefix(signature, "sha256=") {
		return errors.New("invalid github webhook signature")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("invalid github webhook signature")
	}
	return nil
}

func resolveProposalIDForWebhook(store storepkg.Store, headRef string, title string) (string, bool) {
	candidate := proposalIDFromBranch(headRef)
	if candidate == "" {
		candidate = proposalIDFromTitle(title)
	}
	if candidate == "" {
		return "", false
	}
	for _, proposal := range store.ListProposals() {
		if proposal.ID == candidate {
			return proposal.ID, true
		}
	}
	return "", false
}

func proposalIDFromBranch(headRef string) string {
	headRef = strings.TrimSpace(headRef)
	if !strings.HasPrefix(headRef, "codex/") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(headRef, "codex/"))
}

func proposalIDFromTitle(title string) string {
	matches := proposalTitlePattern.FindStringSubmatch(title)
	if len(matches) != 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func githubWebhookSummary(payload gitHubPullRequestWebhook) string {
	status := strings.TrimSpace(payload.Action)
	if payload.Action == "closed" && payload.PullRequest.Merged {
		status = "merged"
	}
	return fmt.Sprintf("GitHub pull request %s %s for %s.", payload.PullRequest.HTMLURL, status, firstNonEmpty(payload.Repository.FullName, payload.Repository.Name))
}
