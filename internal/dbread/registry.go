package dbread

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
	"time"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type Target struct {
	ID             string                         `json:"id"`
	Placement      string                         `json:"placement,omitempty"`
	DSN            string                         `json:"dsn,omitempty"`
	DSNEnv         string                         `json:"dsn_env,omitempty"`
	AllowedSchemas []string                       `json:"allowed_schemas,omitempty"`
	AllowedTables  []string                       `json:"allowed_tables,omitempty"`
	AllowedColumns map[string][]string            `json:"allowed_columns,omitempty"`
	Caps           storepkg.DBReadCaps            `json:"caps"`
	ApprovalTTL    string                         `json:"approval_ttl,omitempty"`
	Redaction      storepkg.DBReadRedactionPolicy `json:"redaction,omitempty"`
	RelayIdentity  string                         `json:"relay_identity,omitempty"`
}

type Registry struct {
	Targets map[string]Target `json:"targets"`
}

func LoadRegistry(raw string) (Registry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("RSI_DB_READ_TARGETS_JSON"))
	}
	if raw == "" {
		return Registry{Targets: map[string]Target{}}, nil
	}
	var cfg struct {
		Targets []Target `json:"targets"`
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return Registry{}, err
	}
	registry := Registry{Targets: map[string]Target{}}
	for _, target := range cfg.Targets {
		target.ID = strings.TrimSpace(target.ID)
		if target.ID == "" {
			return Registry{}, errors.New("db read target id is required")
		}
		if target.Caps.MaxRows <= 0 {
			target.Caps.MaxRows = 100
		}
		if target.Caps.MaxBytes <= 0 {
			target.Caps.MaxBytes = 64 * 1024
		}
		if target.Caps.TimeoutSeconds <= 0 {
			target.Caps.TimeoutSeconds = 5
		}
		if target.Caps.LockTimeoutMS <= 0 {
			target.Caps.LockTimeoutMS = 250
		}
		target.AllowedSchemas = compactStrings(target.AllowedSchemas)
		target.AllowedTables = compactStrings(target.AllowedTables)
		target.Redaction.DenyColumns = compactStrings(target.Redaction.DenyColumns)
		registry.Targets[target.ID] = target
	}
	return registry, nil
}

func (r Registry) Target(id string) (Target, bool) {
	target, ok := r.Targets[strings.TrimSpace(id)]
	if !ok {
		return Target{}, false
	}
	if target.DSN == "" && target.DSNEnv != "" {
		target.DSN = strings.TrimSpace(os.Getenv(target.DSNEnv))
	}
	return target, true
}

func (r Registry) PublicSources() []map[string]any {
	out := make([]map[string]any, 0, len(r.Targets))
	for _, target := range r.Targets {
		out = append(out, map[string]any{
			"id":              target.ID,
			"placement":       target.Placement,
			"allowed_schemas": target.AllowedSchemas,
			"allowed_tables":  target.AllowedTables,
			"caps":            target.Caps,
			"approval_ttl":    firstNonEmpty(target.ApprovalTTL, "1h"),
			"relay_identity":  target.RelayIdentity,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i]["id"].(string) < out[j]["id"].(string)
	})
	return out
}

func (t Target) TTL() time.Duration {
	if d, err := time.ParseDuration(strings.TrimSpace(t.ApprovalTTL)); err == nil && d > 0 {
		return d
	}
	return time.Hour
}

func (t Target) SchemaView() map[string]any {
	return map[string]any{
		"target":          t.ID,
		"allowed_schemas": t.AllowedSchemas,
		"allowed_tables":  t.AllowedTables,
		"allowed_columns": t.AllowedColumns,
	}
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
