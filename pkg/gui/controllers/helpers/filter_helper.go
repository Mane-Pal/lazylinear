package helpers

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// FilterHelper provides search filtering and match highlighting.
type FilterHelper struct{}

func NewFilterHelper() *FilterHelper {
	return &FilterHelper{}
}

// ApplySearch filters issues by query and returns the filtered list.
// Returns nil if query is empty (meaning "show all").
func (h *FilterHelper) ApplySearch(issues []*models.Issue, query string) []*models.Issue {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	var filtered []*models.Issue
	for _, issue := range issues {
		if h.MatchesSearch(issue, query) {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// MatchesSearch checks if an issue matches a search query.
func (h *FilterHelper) MatchesSearch(issue *models.Issue, query string) bool {
	if strings.Contains(strings.ToLower(issue.Identifier), query) {
		return true
	}
	if strings.Contains(strings.ToLower(issue.Title), query) {
		return true
	}
	if strings.Contains(strings.ToLower(issue.Description), query) {
		return true
	}
	if issue.State != nil && strings.Contains(strings.ToLower(issue.State.Name), query) {
		return true
	}
	if issue.Assignee != nil && strings.Contains(strings.ToLower(issue.Assignee.DisplayName), query) {
		return true
	}
	return false
}

// HighlightMatch highlights the first occurrence of query in text.
func (h *FilterHelper) HighlightMatch(text, query string) string {
	if query == "" {
		return text
	}
	lower := strings.ToLower(text)
	idx := strings.Index(lower, strings.ToLower(query))
	if idx == -1 {
		return text
	}
	before := text[:idx]
	match := text[idx : idx+len(query)]
	after := text[idx+len(query):]
	highlighted := lipgloss.NewStyle().Foreground(styles.Warning).Bold(true).Render(match)
	return before + highlighted + after
}
