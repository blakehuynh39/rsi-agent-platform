package conversation

import "strings"

func NormalizeTitle(kind string, body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return strings.TrimSpace(kind)
	}
	runes := []rune(body)
	if len(runes) > 72 {
		body = string(runes[:72]) + "..."
	}
	return body
}
