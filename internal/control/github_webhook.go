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
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

var (
	proposalTitlePattern        = regexp.MustCompile(`(?i)\brsi proposal\s+([a-z0-9._/-]+)\b`)
	proposalAttemptTitlePattern = regexp.MustCompile(`(?i)\brsi proposal\s+([a-z0-9._/-]+)\s+attempt\s+([a-z0-9._-]+)\b`)
)

type gitHubRepository struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
}

type gitHubSender struct {
	Login string `json:"login"`
}

type gitHubPullRequestWebhook struct {
	Action      string           `json:"action"`
	Repository  gitHubRepository `json:"repository"`
	Sender      gitHubSender     `json:"sender"`
	PullRequest struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
		State   string `json:"state"`
		Merged  bool   `json:"merged"`
		Title   string `json:"title"`
		Head    struct {
			Ref string `json:"ref"`
			Sha string `json:"sha"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
}

type gitHubCheckRunWebhook struct {
	Action     string           `json:"action"`
	Repository gitHubRepository `json:"repository"`
	Sender     gitHubSender     `json:"sender"`
	CheckRun   struct {
		Name       string `json:"name"`
		HeadSHA    string `json:"head_sha"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		DetailsURL string `json:"details_url"`
		HTMLURL    string `json:"html_url"`
		CheckSuite struct {
			HeadBranch string `json:"head_branch"`
		} `json:"check_suite"`
	} `json:"check_run"`
}

type gitHubCheckSuiteWebhook struct {
	Action     string           `json:"action"`
	Repository gitHubRepository `json:"repository"`
	Sender     gitHubSender     `json:"sender"`
	CheckSuite struct {
		HeadBranch string `json:"head_branch"`
		HeadSHA    string `json:"head_sha"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		HTMLURL    string `json:"html_url"`
	} `json:"check_suite"`
}

type gitHubWorkflowRunWebhook struct {
	Action      string           `json:"action"`
	Repository  gitHubRepository `json:"repository"`
	Sender      gitHubSender     `json:"sender"`
	WorkflowRun struct {
		Name       string `json:"name"`
		HeadBranch string `json:"head_branch"`
		HeadSHA    string `json:"head_sha"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		HTMLURL    string `json:"html_url"`
	} `json:"workflow_run"`
}

type gitHubWebhookLinkage struct {
	ProposalID string
	AttemptID  string
	HeadRef    string
	HeadSHA    string
	PRURL      string
	Linked     bool
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

	event, linkage, ignored, err := parseGitHubWebhookEvent(store, eventType, deliveryID, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if ignored {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"accepted":true,"linked":false,"ignored":true}`))
		return
	}

	created, err := store.CreateEvent(event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if linkage.Linked {
		if err := applyGitHubAttemptTransition(store, event.Metadata, linkage); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"accepted": true,
		"linked":   linkage.Linked,
		"event_id": created.ID,
	})
}

func parseGitHubWebhookEvent(store storepkg.Store, eventType string, deliveryID string, body []byte) (ingestion.EventEnvelope, gitHubWebhookLinkage, bool, error) {
	now := time.Now().UTC()
	switch strings.TrimSpace(eventType) {
	case "pull_request":
		var payload gitHubPullRequestWebhook
		if err := json.Unmarshal(body, &payload); err != nil {
			return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, false, err
		}
		actionName := strings.ToLower(strings.TrimSpace(payload.Action))
		if actionName != "opened" && actionName != "reopened" && actionName != "closed" {
			return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, true, nil
		}
		linkage := resolveGitHubLinkage(store, payload.PullRequest.Head.Ref, payload.PullRequest.Title, payload.PullRequest.Head.Sha)
		metadata := baseGitHubMetadata(eventType, actionName, payload.Repository, payload.Sender, deliveryID)
		metadata["number"] = payload.PullRequest.Number
		metadata["html_url"] = payload.PullRequest.HTMLURL
		metadata["pr_url"] = payload.PullRequest.HTMLURL
		metadata["state"] = payload.PullRequest.State
		metadata["merged"] = payload.PullRequest.Merged
		metadata["head_ref"] = payload.PullRequest.Head.Ref
		metadata["head_sha"] = payload.PullRequest.Head.Sha
		metadata["base_ref"] = payload.PullRequest.Base.Ref
		if linkage.Linked {
			metadata["proposal_id"] = linkage.ProposalID
			if linkage.AttemptID != "" {
				metadata["attempt_id"] = linkage.AttemptID
			}
		}
		return ingestion.EventEnvelope{
			Source:                     ingestion.SourceGitHub,
			SourceEventID:              deliveryID,
			DedupeKey:                  fmt.Sprintf("github:%s", deliveryID),
			Severity:                   ingestion.SeverityInfo,
			NormalizedProblemStatement: githubPullRequestSummary(payload),
			OwnershipHint:              payload.Repository.Name,
			RawPayloadRef:              fmt.Sprintf("memory://github/%s.json", deliveryID),
			Metadata:                   metadata,
			CreatedAt:                  now,
		}, linkage, false, nil
	case "check_run":
		var payload gitHubCheckRunWebhook
		if err := json.Unmarshal(body, &payload); err != nil {
			return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, false, err
		}
		if !isGitHubFailureEvent(strings.ToLower(strings.TrimSpace(payload.CheckRun.Conclusion))) {
			return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, true, nil
		}
		linkage := resolveGitHubLinkage(store, payload.CheckRun.CheckSuite.HeadBranch, "", payload.CheckRun.HeadSHA)
		metadata := baseGitHubMetadata(eventType, payload.Action, payload.Repository, payload.Sender, deliveryID)
		metadata["status"] = payload.CheckRun.Status
		metadata["conclusion"] = payload.CheckRun.Conclusion
		metadata["name"] = payload.CheckRun.Name
		metadata["html_url"] = firstNonEmpty(payload.CheckRun.DetailsURL, payload.CheckRun.HTMLURL)
		metadata["head_ref"] = payload.CheckRun.CheckSuite.HeadBranch
		metadata["head_sha"] = payload.CheckRun.HeadSHA
		if linkage.Linked {
			metadata["proposal_id"] = linkage.ProposalID
			if linkage.AttemptID != "" {
				metadata["attempt_id"] = linkage.AttemptID
			}
		}
		return ingestion.EventEnvelope{
			Source:                     ingestion.SourceGitHub,
			SourceEventID:              deliveryID,
			DedupeKey:                  fmt.Sprintf("github:%s", deliveryID),
			Severity:                   ingestion.SeverityWarning,
			NormalizedProblemStatement: githubCheckRunSummary(payload),
			OwnershipHint:              payload.Repository.Name,
			RawPayloadRef:              fmt.Sprintf("memory://github/%s.json", deliveryID),
			Metadata:                   metadata,
			CreatedAt:                  now,
		}, linkage, false, nil
	case "check_suite":
		var payload gitHubCheckSuiteWebhook
		if err := json.Unmarshal(body, &payload); err != nil {
			return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, false, err
		}
		if !isGitHubFailureEvent(strings.ToLower(strings.TrimSpace(payload.CheckSuite.Conclusion))) {
			return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, true, nil
		}
		linkage := resolveGitHubLinkage(store, payload.CheckSuite.HeadBranch, "", payload.CheckSuite.HeadSHA)
		metadata := baseGitHubMetadata(eventType, payload.Action, payload.Repository, payload.Sender, deliveryID)
		metadata["status"] = payload.CheckSuite.Status
		metadata["conclusion"] = payload.CheckSuite.Conclusion
		metadata["html_url"] = payload.CheckSuite.HTMLURL
		metadata["head_ref"] = payload.CheckSuite.HeadBranch
		metadata["head_sha"] = payload.CheckSuite.HeadSHA
		if linkage.Linked {
			metadata["proposal_id"] = linkage.ProposalID
			if linkage.AttemptID != "" {
				metadata["attempt_id"] = linkage.AttemptID
			}
		}
		return ingestion.EventEnvelope{
			Source:                     ingestion.SourceGitHub,
			SourceEventID:              deliveryID,
			DedupeKey:                  fmt.Sprintf("github:%s", deliveryID),
			Severity:                   ingestion.SeverityWarning,
			NormalizedProblemStatement: githubCheckSuiteSummary(payload),
			OwnershipHint:              payload.Repository.Name,
			RawPayloadRef:              fmt.Sprintf("memory://github/%s.json", deliveryID),
			Metadata:                   metadata,
			CreatedAt:                  now,
		}, linkage, false, nil
	case "workflow_run":
		var payload gitHubWorkflowRunWebhook
		if err := json.Unmarshal(body, &payload); err != nil {
			return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, false, err
		}
		if !isGitHubFailureEvent(strings.ToLower(strings.TrimSpace(payload.WorkflowRun.Conclusion))) {
			return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, true, nil
		}
		linkage := resolveGitHubLinkage(store, payload.WorkflowRun.HeadBranch, "", payload.WorkflowRun.HeadSHA)
		metadata := baseGitHubMetadata(eventType, payload.Action, payload.Repository, payload.Sender, deliveryID)
		metadata["status"] = payload.WorkflowRun.Status
		metadata["conclusion"] = payload.WorkflowRun.Conclusion
		metadata["name"] = payload.WorkflowRun.Name
		metadata["html_url"] = payload.WorkflowRun.HTMLURL
		metadata["head_ref"] = payload.WorkflowRun.HeadBranch
		metadata["head_sha"] = payload.WorkflowRun.HeadSHA
		if linkage.Linked {
			metadata["proposal_id"] = linkage.ProposalID
			if linkage.AttemptID != "" {
				metadata["attempt_id"] = linkage.AttemptID
			}
		}
		return ingestion.EventEnvelope{
			Source:                     ingestion.SourceGitHub,
			SourceEventID:              deliveryID,
			DedupeKey:                  fmt.Sprintf("github:%s", deliveryID),
			Severity:                   ingestion.SeverityWarning,
			NormalizedProblemStatement: githubWorkflowRunSummary(payload),
			OwnershipHint:              payload.Repository.Name,
			RawPayloadRef:              fmt.Sprintf("memory://github/%s.json", deliveryID),
			Metadata:                   metadata,
			CreatedAt:                  now,
		}, linkage, false, nil
	default:
		return ingestion.EventEnvelope{}, gitHubWebhookLinkage{}, true, nil
	}
}

func baseGitHubMetadata(eventType string, action string, repository gitHubRepository, sender gitHubSender, deliveryID string) map[string]any {
	return map[string]any{
		"event_type":                    strings.ToLower(strings.TrimSpace(eventType)),
		"action":                        strings.ToLower(strings.TrimSpace(action)),
		"repository":                    firstNonEmpty(repository.FullName, repository.Name),
		"sender_login":                  sender.Login,
		"github_delivery_id":            deliveryID,
		"skip_workflow_materialization": true,
	}
}

func applyGitHubAttemptTransition(store storepkg.Store, metadata map[string]any, linkage gitHubWebhookLinkage) error {
	proposal, ok := findProposal(store, linkage.ProposalID)
	if !ok {
		return nil
	}
	attempt, hasAttempt := resolveAttempt(store, proposal, linkage)
	if !hasAttempt {
		return nil
	}
	eventType := strings.ToLower(strings.TrimSpace(stringFromAny(metadata["event_type"])))
	actionName := strings.ToLower(strings.TrimSpace(stringFromAny(metadata["action"])))
	conclusion := strings.ToLower(strings.TrimSpace(stringFromAny(metadata["conclusion"])))
	merged := strings.ToLower(strings.TrimSpace(stringFromAny(metadata["merged"])))
	prURL := firstNonEmpty(stringFromAny(metadata["pr_url"]), stringFromAny(metadata["html_url"]))
	headSHA := stringFromAny(metadata["head_sha"])

	if proposal.CurrentAttemptID != "" && attempt.ID != proposal.CurrentAttemptID && eventType != "pull_request" {
		return nil
	}

	switch {
	case eventType == "pull_request" && (actionName == "opened" || actionName == "reopened"):
		if proposal.CurrentAttemptID != "" && attempt.ID != proposal.CurrentAttemptID {
			return nil
		}
		attempt.State = improvement.AttemptStateCIObserving
		attempt.PRURL = firstNonEmpty(prURL, attempt.PRURL)
		attempt.HeadSHA = firstNonEmpty(headSHA, attempt.HeadSHA)
		attempt.UpdatedAt = time.Now().UTC()
		if _, err := store.UpsertChangeAttempt(attempt); err != nil {
			return err
		}
		_, err := store.UpdateProposalStatus(proposal.ID, review.ProposalPROpen)
		return err
	case eventType == "pull_request" && actionName == "closed" && (merged == "true" || strings.EqualFold(stringFromAny(metadata["state"]), "merged")):
		if proposal.CurrentAttemptID != "" && attempt.ID != proposal.CurrentAttemptID {
			return nil
		}
		attempt.State = improvement.AttemptStateMerged
		attempt.PRURL = firstNonEmpty(prURL, attempt.PRURL)
		attempt.HeadSHA = firstNonEmpty(headSHA, attempt.HeadSHA)
		attempt.UpdatedAt = time.Now().UTC()
		if _, err := store.UpsertChangeAttempt(attempt); err != nil {
			return err
		}
		_, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
			Decision:   string(review.ProposalMerged),
			Rationale:  "GitHub webhook recorded merged PR for the active attempt.",
			ReviewerID: "github-webhook",
		})
		return err
	case eventType == "pull_request" && actionName == "closed":
		return transitionGitHubAttemptFailure(store, proposal, attempt, "closed_unmerged", "GitHub pull request closed without merge.", improvement.AttemptTriggerPRClosed)
	case eventType == "check_run" && isGitHubFailureEvent(conclusion):
		return transitionGitHubAttemptFailure(store, proposal, attempt, "ci_regression", fmt.Sprintf("GitHub check run failed with conclusion %s.", conclusion), improvement.AttemptTriggerCIFailed)
	case eventType == "check_suite" && isGitHubFailureEvent(conclusion):
		return transitionGitHubAttemptFailure(store, proposal, attempt, "ci_regression", fmt.Sprintf("GitHub check suite failed with conclusion %s.", conclusion), improvement.AttemptTriggerCIFailed)
	case eventType == "workflow_run" && isGitHubFailureEvent(conclusion):
		return transitionGitHubAttemptFailure(store, proposal, attempt, "ci_regression", fmt.Sprintf("GitHub workflow run failed with conclusion %s.", conclusion), improvement.AttemptTriggerCIFailed)
	default:
		return nil
	}
}

func transitionGitHubAttemptFailure(store storepkg.Store, proposal review.Proposal, attempt improvement.ChangeAttempt, failureClass string, failureSummary string, trigger improvement.ChangeAttemptTrigger) error {
	if proposal.CurrentAttemptID != "" && attempt.ID != proposal.CurrentAttemptID {
		return nil
	}
	now := time.Now().UTC()
	switch failureClass {
	case "ci_regression":
		attempt.State = improvement.AttemptStateCIFailed
	case "closed_unmerged":
		attempt.State = improvement.AttemptStateClosedUnmerged
	default:
		attempt.State = improvement.AttemptStateNeedsReview
	}
	attempt.FailureClass = failureClass
	attempt.FailureSummary = failureSummary
	attempt.MaterialHypothesisChange = false
	if shouldAutoRetryAttempt(attempt) {
		attempt.RetryDecision = "auto_retry"
		retryAt := now.Add(time.Minute)
		attempt.RetryAfter = &retryAt
	} else {
		attempt.State = improvement.AttemptStateNeedsReview
		attempt.RetryDecision = "needs_review"
		attempt.RetryAfter = nil
	}
	attempt.UpdatedAt = now
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		return err
	}
	if attempt.RetryDecision == "auto_retry" {
		if _, err := store.UpdateProposalStatus(proposal.ID, review.ProposalApproved); err != nil {
			return err
		}
		nextAttemptNumber := attempt.AttemptNumber + 1
		_, err := enqueueOperationBackedWork(store, operation.Execution{
			ScopeKind:     operation.ScopeProposal,
			ScopeID:       proposal.ID,
			OperationKind: "line_activate",
			OperationKey:  fmt.Sprintf("attempt-%02d", nextAttemptNumber),
			Status:        operation.StatusQueued,
			Queue:         queue.ProposalQueue,
			RequestedBy:   "github-webhook",
			TraceID:       firstNonEmpty(attempt.AttemptTraceID, proposal.TraceID),
			ProposalID:    proposal.ID,
			AttemptID:     attempt.ID,
		}, queue.WorkItem{
			Queue:          queue.ProposalQueue,
			Kind:           "approved_proposal",
			Status:         queue.WorkQueued,
			TraceID:        firstNonEmpty(attempt.AttemptTraceID, proposal.TraceID),
			ConversationID: proposal.ConversationID,
			CaseID:         proposal.CaseID,
			TriggerEventID: firstNonEmpty(proposal.OriginTraceID, proposal.TraceID),
			ProposalID:     proposal.ID,
			RequestedBy:    "github-webhook",
			ApprovalMode:   "human_review",
			CreatedAt:      now,
			UpdatedAt:      now,
			Payload: map[string]any{
				"candidate_key":  proposal.CandidateKey,
				"risk_tier":      proposal.RiskTier,
				"trigger":        string(trigger),
				"parent_attempt": attempt.ID,
				"attempt_number": nextAttemptNumber,
			},
		})
		return err
	}
	_, err := store.UpdateProposalStatus(proposal.ID, review.ProposalPendingReview)
	return err
}

func shouldAutoRetryAttempt(attempt improvement.ChangeAttempt) bool {
	return attempt.AttemptNumber < 3
}

func resolveGitHubLinkage(store storepkg.Store, headRef string, title string, headSHA string) gitHubWebhookLinkage {
	headRef = strings.TrimSpace(headRef)
	headSHA = strings.TrimSpace(headSHA)
	if headSHA != "" {
		for _, item := range store.ListPRAttempts() {
			if strings.TrimSpace(item.HeadSHA) == headSHA {
				return gitHubWebhookLinkage{
					ProposalID: item.ProposalID,
					AttemptID:  item.AttemptID,
					HeadRef:    headRef,
					HeadSHA:    headSHA,
					PRURL:      item.PRURL,
					Linked:     true,
				}
			}
		}
	}
	if headRef != "" {
		for _, item := range store.ListChangeAttempts() {
			if strings.TrimSpace(item.BranchName) == headRef {
				return gitHubWebhookLinkage{
					ProposalID: item.ProposalID,
					AttemptID:  item.ID,
					HeadRef:    headRef,
					HeadSHA:    headSHA,
					PRURL:      item.PRURL,
					Linked:     true,
				}
			}
		}
	}
	if proposalID, attemptID := proposalAttemptFromTitle(title); proposalID != "" {
		if _, ok := findProposal(store, proposalID); ok {
			return gitHubWebhookLinkage{
				ProposalID: proposalID,
				AttemptID:  resolveAttemptID(store, proposalID, attemptID),
				HeadRef:    headRef,
				HeadSHA:    headSHA,
				Linked:     true,
			}
		}
	}
	proposalID := proposalIDFromBranch(headRef)
	if proposalID == "" {
		proposalID = proposalIDFromTitle(title)
	}
	if proposalID == "" {
		return gitHubWebhookLinkage{}
	}
	proposal, ok := findProposal(store, proposalID)
	if !ok {
		return gitHubWebhookLinkage{}
	}
	return gitHubWebhookLinkage{
		ProposalID: proposal.ID,
		AttemptID:  firstNonEmpty(proposal.CurrentAttemptID, latestAttemptIDForProposal(store, proposal.ID)),
		HeadRef:    headRef,
		HeadSHA:    headSHA,
		Linked:     true,
	}
}

func resolveAttempt(store storepkg.Store, proposal review.Proposal, linkage gitHubWebhookLinkage) (improvement.ChangeAttempt, bool) {
	if linkage.AttemptID != "" {
		if item, ok := store.GetChangeAttempt(linkage.AttemptID); ok {
			return item, true
		}
	}
	if proposal.CurrentAttemptID != "" {
		if item, ok := store.GetChangeAttempt(proposal.CurrentAttemptID); ok {
			return item, true
		}
	}
	if latestID := latestAttemptIDForProposal(store, proposal.ID); latestID != "" {
		return store.GetChangeAttempt(latestID)
	}
	return improvement.ChangeAttempt{}, false
}

func latestAttemptIDForProposal(store storepkg.Store, proposalID string) string {
	bestNumber := -1
	bestID := ""
	for _, item := range store.ListChangeAttempts() {
		if item.ProposalID != proposalID {
			continue
		}
		if item.AttemptNumber > bestNumber {
			bestNumber = item.AttemptNumber
			bestID = item.ID
		}
	}
	return bestID
}

func resolveAttemptID(store storepkg.Store, proposalID string, attemptToken string) string {
	for _, item := range store.ListChangeAttempts() {
		if item.ProposalID != proposalID {
			continue
		}
		if strings.EqualFold(item.ID, attemptToken) || strings.HasSuffix(strings.TrimSpace(item.BranchName), strings.TrimSpace(attemptToken)) {
			return item.ID
		}
	}
	return ""
}

func findProposal(store storepkg.Store, proposalID string) (review.Proposal, bool) {
	for _, item := range store.ListProposals() {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}

func proposalAttemptFromTitle(title string) (string, string) {
	matches := proposalAttemptTitlePattern.FindStringSubmatch(strings.TrimSpace(title))
	if len(matches) != 3 {
		return "", ""
	}
	return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
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

func proposalIDFromBranch(headRef string) string {
	headRef = strings.TrimSpace(headRef)
	if !strings.HasPrefix(headRef, "codex/") {
		return ""
	}
	trimmed := strings.TrimPrefix(headRef, "codex/")
	parts := strings.Split(trimmed, "/")
	if len(parts) >= 2 && strings.HasPrefix(parts[len(parts)-1], "attempt-") {
		return strings.Join(parts[:len(parts)-1], "/")
	}
	return strings.TrimSpace(trimmed)
}

func proposalIDFromTitle(title string) string {
	matches := proposalTitlePattern.FindStringSubmatch(title)
	if len(matches) != 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func githubPullRequestSummary(payload gitHubPullRequestWebhook) string {
	status := strings.TrimSpace(payload.Action)
	if payload.Action == "closed" && payload.PullRequest.Merged {
		status = "merged"
	}
	return fmt.Sprintf("GitHub pull request %s %s for %s.", payload.PullRequest.HTMLURL, status, firstNonEmpty(payload.Repository.FullName, payload.Repository.Name))
}

func githubCheckRunSummary(payload gitHubCheckRunWebhook) string {
	return fmt.Sprintf("GitHub check run %s failed with conclusion %s for %s.", firstNonEmpty(payload.CheckRun.Name, "unknown"), payload.CheckRun.Conclusion, firstNonEmpty(payload.Repository.FullName, payload.Repository.Name))
}

func githubCheckSuiteSummary(payload gitHubCheckSuiteWebhook) string {
	return fmt.Sprintf("GitHub check suite failed with conclusion %s for %s.", payload.CheckSuite.Conclusion, firstNonEmpty(payload.Repository.FullName, payload.Repository.Name))
}

func githubWorkflowRunSummary(payload gitHubWorkflowRunWebhook) string {
	return fmt.Sprintf("GitHub workflow run %s failed with conclusion %s for %s.", firstNonEmpty(payload.WorkflowRun.Name, "unknown"), payload.WorkflowRun.Conclusion, firstNonEmpty(payload.Repository.FullName, payload.Repository.Name))
}

func isGitHubFailureEvent(conclusion string) bool {
	switch strings.ToLower(strings.TrimSpace(conclusion)) {
	case "failure", "cancelled", "timed_out", "startup_failure", "action_required":
		return true
	default:
		return false
	}
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return ""
	}
}
