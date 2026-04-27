package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const (
	effectQueueWorkflow     = "workflow"
	effectQueueAction       = "action"
	effectTaskClassSimple   = "simple"
	effectTaskClassArtifact = "artifact"
	effectTaskClassImprove  = "improvement"
)

type EffectClaimSelector struct {
	MachineKind   transition.MachineKind
	EffectKind    transition.EffectKind
	PayloadEquals map[string]string
}

func normalizeEffectScheduling(effect transition.EffectExecution) transition.EffectExecution {
	if effect.Payload == nil {
		effect.Payload = map[string]any{}
	}
	if strings.TrimSpace(effect.QueueName) == "" {
		defaultQueue := defaultQueueNameForEffect(effect)
		if effect.MachineKind == transition.MachineAction {
			effect.QueueName = defaultQueue
		} else {
			effect.QueueName = firstNonEmpty(
				stringFromMap(effect.Payload, "queue_name"),
				stringFromMap(effect.Payload, "resume_queue"),
				defaultQueue,
			)
		}
	}
	if strings.TrimSpace(effect.ScopeKey) == "" {
		effect.ScopeKey = firstNonEmpty(
			stringFromMap(effect.Payload, "scope_key"),
			stringFromMap(effect.Payload, "conversation_id"),
			stringFromMap(effect.Payload, "case_id"),
			effect.AggregateID,
		)
	}
	if strings.TrimSpace(effect.TaskClass) == "" {
		effect.TaskClass = taskClassForEffect(effect)
	}
	if effect.Priority == 0 {
		effect.Priority = priorityForTaskClass(effect.TaskClass)
	}
	effect.QueueName = strings.TrimSpace(effect.QueueName)
	effect.ScopeKey = strings.TrimSpace(effect.ScopeKey)
	effect.TaskClass = strings.TrimSpace(effect.TaskClass)
	return effect
}

func defaultQueueNameForEffect(effect transition.EffectExecution) string {
	switch effect.MachineKind {
	case transition.MachineAction:
		return effectQueueAction
	case transition.MachineQuestionRun:
		return effectQueueWorkflow
	default:
		return effectQueueWorkflow
	}
}

func taskClassForEffect(effect transition.EffectExecution) string {
	taskClass := stringFromMap(effect.Payload, "task_class")
	if intFromMap(effect.Payload, "requested_artifact_count") > 0 ||
		strings.EqualFold(stringFromMap(effect.Payload, "requested_artifact"), "true") ||
		strings.EqualFold(taskClass, effectTaskClassArtifact) {
		return effectTaskClassArtifact
	}
	if strings.EqualFold(taskClass, effectTaskClassImprove) {
		return effectTaskClassImprove
	}
	if effect.MachineKind == transition.MachineAttempt || effect.MachineKind == transition.MachineProblemLine || effect.MachineKind == transition.MachineRuntimeDiagnosis {
		return effectTaskClassImprove
	}
	return effectTaskClassSimple
}

func priorityForTaskClass(taskClass string) int {
	switch strings.ToLower(strings.TrimSpace(taskClass)) {
	case effectTaskClassSimple:
		return 100
	case effectTaskClassArtifact:
		return 50
	case effectTaskClassImprove:
		return 10
	default:
		return 25
	}
}

func effectAvailableForClaim(item transition.EffectExecution, now time.Time, lease time.Duration) bool {
	if item.NotBefore != nil && item.NotBefore.After(now) {
		return false
	}
	switch item.Status {
	case transition.EffectQueued, transition.EffectFailed:
		return true
	case transition.EffectRunning:
		return effectRunningClaimable(item, now)
	case transition.EffectCompleted, transition.EffectCanceled, transition.EffectSuperseded:
		return false
	default:
		return false
	}
}

func effectRunningClaimable(item transition.EffectExecution, now time.Time) bool {
	if item.NotBefore != nil && item.NotBefore.After(now) {
		return false
	}
	return item.LeaseExpiresAt != nil && !item.LeaseExpiresAt.After(now)
}

func effectRunningBlocksScope(item transition.EffectExecution, now time.Time) bool {
	return item.Status == transition.EffectRunning && !effectRunningClaimable(item, now)
}

func queueNameAllowed(queueName string, allowed map[string]struct{}) bool {
	if len(allowed) == 0 {
		return true
	}
	_, ok := allowed[strings.TrimSpace(queueName)]
	return ok
}

func effectMatchesClaimSelectors(item transition.EffectExecution, selectors []EffectClaimSelector) bool {
	if len(selectors) == 0 {
		return true
	}
	for _, selector := range selectors {
		if selector.MachineKind != "" && item.MachineKind != selector.MachineKind {
			continue
		}
		if selector.EffectKind != "" && item.EffectKind != selector.EffectKind {
			continue
		}
		matchesPayload := true
		for key, expected := range selector.PayloadEquals {
			if key = strings.TrimSpace(key); key == "" {
				continue
			}
			if !strings.EqualFold(strings.TrimSpace(stringFromMap(item.Payload, key)), strings.TrimSpace(expected)) {
				matchesPayload = false
				break
			}
		}
		if !matchesPayload {
			continue
		}
		return true
	}
	return false
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func intFromMap(values map[string]any, key string) int {
	if values == nil {
		return 0
	}
	switch typed := values[key].(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	case string:
		var out int
		if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%d", &out); err == nil {
			return out
		}
	}
	return 0
}
