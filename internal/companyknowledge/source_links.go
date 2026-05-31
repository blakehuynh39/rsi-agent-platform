package companyknowledge

import (
	"net/url"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/store"
)

func markdownSourceLink(rawURL string, fallbackLabel string) string {
	rawURL = strings.TrimSpace(rawURL)
	if !isHTTPSourceURL(rawURL) {
		return ""
	}
	label := strings.TrimSpace(fallbackLabel)
	if label == "" {
		label = sourceLinkLabel(rawURL)
	}
	return "[" + escapeMarkdownLinkText(label) + "](" + escapeMarkdownLinkURL(rawURL) + ")"
}

func citationMarkdownSourceLink(citation store.CompanyWikiCitationInput, fallbackURL string) string {
	if link := markdownSourceLink(citation.NativeLocator, ""); link != "" {
		return link
	}
	return markdownSourceLink(fallbackURL, "")
}

func isHTTPSourceURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" || parsed.Scheme == "http"
}

func sourceLinkLabel(rawURL string) string {
	host := ""
	if parsed, err := url.Parse(strings.TrimSpace(rawURL)); err == nil {
		host = strings.ToLower(parsed.Hostname())
	}
	switch {
	case strings.Contains(host, "slack.com"):
		return "Slack source"
	case strings.Contains(host, "notion.so"), strings.Contains(host, "notion.site"):
		return "Notion source"
	default:
		return "source"
	}
}

func escapeMarkdownLinkURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	rawURL = strings.ReplaceAll(rawURL, " ", "%20")
	rawURL = strings.ReplaceAll(rawURL, ")", "%29")
	return rawURL
}
