package debuglog

import (
	"encoding/json"
	"log"
	"strings"
)

const (
	DefaultPayloadLimit = 100000
	redactedValue       = "[redacted]"
)

var sensitiveKeyFragments = []string{
	"authorization",
	"api_key",
	"apikey",
	"token",
	"secret",
	"private_key",
	"password",
}

func Logf(enabled bool, format string, args ...any) {
	if !enabled {
		return
	}
	log.Printf(format, args...)
}

func JSON(value any, limit int) string {
	sanitized := sanitize(value, "")
	encoded, err := json.Marshal(sanitized)
	if err != nil {
		return Truncate(err.Error(), limit)
	}
	return Truncate(string(encoded), limit)
}

func Truncate(value string, limit int) string {
	text := strings.TrimSpace(value)
	if limit <= 0 {
		limit = DefaultPayloadLimit
	}
	if len(text) <= limit {
		return text
	}
	if limit <= 32 {
		return text[:limit]
	}
	return text[:limit-20] + "...[truncated]"
}

func sanitize(value any, key string) any {
	if isSensitiveKey(key) {
		return redactedValue
	}
	switch item := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(item))
		for childKey, childValue := range item {
			out[childKey] = sanitize(childValue, childKey)
		}
		return out
	case []any:
		out := make([]any, 0, len(item))
		for _, child := range item {
			out = append(out, sanitize(child, ""))
		}
		return out
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	if lower == "" {
		return false
	}
	for _, fragment := range sensitiveKeyFragments {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}
