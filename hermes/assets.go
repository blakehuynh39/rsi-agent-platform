package hermesassets

import (
	"embed"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed exported-skills/stage/metadata.json exported-skills/stage/skills/*/SKILL.md exported-skills/stage/skills/*/*/SKILL.md exported-skills/stage/skills/*/*/*/SKILL.md
var exportedSkillFS embed.FS

type ExportedSkill struct {
	Name        string
	Description string
	Category    string
	Path        string
}

func ExportedSkills() []ExportedSkill {
	const root = "exported-skills/stage/skills"
	out := []ExportedSkill{}
	_ = fs.WalkDir(exportedSkillFS, root, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Base(filePath) != "SKILL.md" {
			return nil
		}
		raw, readErr := exportedSkillFS.ReadFile(filePath)
		if readErr != nil {
			return nil
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(filePath, root), "/")
		meta := parseFrontmatter(string(raw))
		skillDir := path.Dir(rel)
		name := firstNonEmpty(meta["name"], path.Base(skillDir))
		if name == "." || name == "" {
			return nil
		}
		category := skillDir
		if strings.Contains(skillDir, "/") {
			category = path.Dir(skillDir)
		}
		out = append(out, ExportedSkill{
			Name:        name,
			Description: meta["description"],
			Category:    category,
			Path:        rel,
		})
		return nil
	})
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category == out[j].Category {
			return out[i].Name < out[j].Name
		}
		return out[i].Category < out[j].Category
	})
	return out
}

func parseFrontmatter(text string) map[string]string {
	out := map[string]string{}
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return out
	}
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = cleanScalar(value)
	}
	return out
}

func cleanScalar(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return strings.TrimSpace(value[1 : len(value)-1])
		}
	}
	if before, _, ok := strings.Cut(value, " #"); ok {
		value = before
	}
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
