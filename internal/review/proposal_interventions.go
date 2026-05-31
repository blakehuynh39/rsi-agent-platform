package review

import (
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
)

func RecommendProposalIntervention(candidate improvement.Candidate) ProposalInterventionKind {
	if candidate.Status == improvement.CandidateNeedsEvidence || candidate.LineStatus == improvement.LineNeedsEvidence {
		return InterventionNeedsEvidence
	}
	if candidate.LineStatus == improvement.LineClosed || candidate.Status == improvement.CandidateDormant {
		return InterventionCloseLine
	}
	if candidate.TargetLayer == harness.TargetLayerHarnessOverlay {
		return InterventionHarnessOverlay
	}
	return InterventionRepoChange
}

func ProposalExecutableIntervention(kind ProposalInterventionKind) bool {
	switch kind {
	case InterventionRepoChange, InterventionHarnessOverlay:
		return true
	default:
		return false
	}
}

func ProposalDispositionForIntervention(kind ProposalInterventionKind) string {
	switch kind {
	case InterventionNeedsEvidence:
		return "collect_evidence"
	case InterventionCloseLine:
		return "close_line"
	default:
		return "approve_intervention"
	}
}

func ProposalTargetSurfaceFromCandidate(candidate improvement.Candidate) string {
	switch {
	case strings.TrimSpace(candidate.ProposedScope) != "":
		return strings.TrimSpace(candidate.ProposedScope)
	case strings.TrimSpace(candidate.TargetKind) != "" && strings.TrimSpace(candidate.TargetRef) != "":
		return fmt.Sprintf("%s:%s", strings.TrimSpace(candidate.TargetKind), strings.TrimSpace(candidate.TargetRef))
	case strings.TrimSpace(candidate.TargetRef) != "":
		return strings.TrimSpace(candidate.TargetRef)
	case strings.TrimSpace(candidate.TargetKind) != "":
		return strings.TrimSpace(candidate.TargetKind)
	default:
		return "unspecified_target_surface"
	}
}

func ProposalTargetSurfaceFromProposal(item Proposal) string {
	switch {
	case strings.TrimSpace(item.TargetSurface) != "":
		return strings.TrimSpace(item.TargetSurface)
	case strings.TrimSpace(item.ProposedScope) != "":
		return strings.TrimSpace(item.ProposedScope)
	case strings.TrimSpace(item.TargetKind) != "" && strings.TrimSpace(item.TargetRef) != "":
		return fmt.Sprintf("%s:%s", strings.TrimSpace(item.TargetKind), strings.TrimSpace(item.TargetRef))
	case strings.TrimSpace(item.TargetRef) != "":
		return strings.TrimSpace(item.TargetRef)
	case strings.TrimSpace(item.TargetKind) != "":
		return strings.TrimSpace(item.TargetKind)
	default:
		return "unspecified_target_surface"
	}
}

func ProposalValidationPlan(kind ProposalInterventionKind, targetSurface string) string {
	switch kind {
	case InterventionHarnessOverlay:
		return fmt.Sprintf("Generate a bounded runtime overlay for %s, validate behavior in the targeted role, and activate only if the overlay remains inside the approved scope.", targetSurface)
	case InterventionNeedsEvidence:
		return fmt.Sprintf("Collect more trace, tool, and outcome evidence for %s before attempting a code or overlay change.", targetSurface)
	case InterventionCloseLine:
		return fmt.Sprintf("Preserve the problem line for memory, but do not execute a remediation attempt against %s.", targetSurface)
	default:
		return fmt.Sprintf("Generate a bounded change for %s, validate it in sandbox, and only then open a draft PR.", targetSurface)
	}
}

func ProposalRiskSummary(riskTier string, targetSurface string, kind ProposalInterventionKind) string {
	risk := strings.TrimSpace(riskTier)
	if risk == "" {
		risk = "medium"
	}
	switch kind {
	case InterventionHarnessOverlay:
		return fmt.Sprintf("%s risk intervention on runtime harness surface %s.", risk, targetSurface)
	case InterventionNeedsEvidence:
		return fmt.Sprintf("%s risk evidence gap on %s; additional grounding is required before intervention.", risk, targetSurface)
	case InterventionCloseLine:
		return fmt.Sprintf("%s risk line closure for %s; current evidence does not justify another remediation attempt.", risk, targetSurface)
	default:
		return fmt.Sprintf("%s risk repo-change intervention on %s.", risk, targetSurface)
	}
}

func ProposalInterventionRationale(candidate improvement.Candidate, kind ProposalInterventionKind, targetSurface string) string {
	hypothesis := strings.TrimSpace(candidate.Hypothesis)
	if hypothesis == "" {
		hypothesis = fmt.Sprintf("Recurring issue %s in %s.", candidate.FailureMode, candidate.Subsystem)
	}
	switch kind {
	case InterventionHarnessOverlay:
		return fmt.Sprintf("%s The evidence points to role behavior or memory policy, so the recommended intervention is a harness overlay on %s rather than a repo PR.", hypothesis, targetSurface)
	case InterventionNeedsEvidence:
		return fmt.Sprintf("%s The current evidence is not strong enough to justify a code or overlay change, so the recommended intervention is to gather more evidence for %s.", hypothesis, targetSurface)
	case InterventionCloseLine:
		return fmt.Sprintf("%s The current line should be closed without another remediation attempt because the approved target surface %s no longer appears justified.", hypothesis, targetSurface)
	default:
		return fmt.Sprintf("%s The evidence points to a concrete repo-change intervention on %s, so opening a draft PR after human approval is the recommended next action.", hypothesis, targetSurface)
	}
}

func NormalizeProposalInterventionFields(item Proposal) Proposal {
	if item.RecommendedInterventionKind == "" {
		if item.TargetLayer == harness.TargetLayerHarnessOverlay {
			item.RecommendedInterventionKind = InterventionHarnessOverlay
		} else {
			item.RecommendedInterventionKind = InterventionRepoChange
		}
	}
	item.TargetSurface = ProposalTargetSurfaceFromProposal(item)
	if item.ValidationPlan == "" {
		item.ValidationPlan = ProposalValidationPlan(item.RecommendedInterventionKind, item.TargetSurface)
	}
	if item.MaterialRiskSummary == "" {
		item.MaterialRiskSummary = ProposalRiskSummary(item.RiskTier, item.TargetSurface, item.RecommendedInterventionKind)
	}
	if item.RecommendedInterventionRationale == "" {
		item.RecommendedInterventionRationale = item.Summary
	}
	if item.RecommendedDisposition == "" {
		item.RecommendedDisposition = ProposalDispositionForIntervention(item.RecommendedInterventionKind)
	}
	if item.TouchedFiles == nil {
		item.TouchedFiles = []string{}
	}
	return item
}
