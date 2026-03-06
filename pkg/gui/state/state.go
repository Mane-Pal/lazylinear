package state

import (
	"time"

	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// AppState holds all shared mutable application state.
// Contexts and controllers receive a pointer and mutate directly.
type AppState struct {
	// Data
	Teams       []*models.Team
	Issues      []*models.Issue
	Projects    []*models.Project
	CurrentUser *models.User
	TeamMembers []*models.User
	TeamStates  []*models.WorkflowState

	// Selection
	SelectedTeam    int
	SelectedIssue   int
	SelectedProject int

	// Middle pane view mode
	MiddlePaneView types.MiddlePaneView

	// Sidebar filter state
	ActiveFilter   string // "all", "my_issues"
	SelectedFilter int    // Which filter option is highlighted
	SidebarSection int    // 0 = teams, 1 = filters, 2 = states, 3 = priority

	// State filter (multi-select)
	SelectedStateIdx   int             // Navigation cursor in states section
	ActiveStateFilters map[string]bool // Set of selected state IDs (multi-select)

	// Priority filter state
	ActivePriorityFilter int // -1 = no filter, 1-4 = priority level
	SelectedPriority     int // For navigation (0-3 maps to priority 1-4)

	// Focus
	FocusedPanel types.ContextKey

	// UI State
	Loading   bool
	Err       error
	ShowHelp  bool
	StatusMsg string

	// Label menu state (multi-select)
	TeamLabels     []*models.Label // Labels available for current team
	SelectedLabels map[string]bool // Currently selected label IDs in menu

	// Cycle menu state
	TeamCycles []*models.Cycle // Cycles available for current team

	// Issue picker state (for parent/relation target)
	AllTeamIssues []*models.Issue // All issues for picker

	// Detail view state
	DetailedIssue   *models.Issue // Issue with full comments
	DetailScroll    int           // Scroll position in detail view
	SelectedComment int           // Selected comment index (-1 = none)

	// Search state
	SearchQuery    string
	FilteredIssues []*models.Issue // nil = show all issues

	// Background sync
	LastSynced time.Time
}

// GetDisplayIssues returns the issues to display, accounting for search filter.
func (s *AppState) GetDisplayIssues() []*models.Issue {
	if s.FilteredIssues != nil {
		return s.FilteredIssues
	}
	return s.Issues
}

// GetSelectedIssue returns the currently selected issue, or nil.
func (s *AppState) GetSelectedIssue() *models.Issue {
	issues := s.GetDisplayIssues()
	if s.SelectedIssue < len(issues) {
		return issues[s.SelectedIssue]
	}
	return nil
}

// CountIssuesByState counts issues matching a state ID.
func (s *AppState) CountIssuesByState(stateID string) int {
	count := 0
	for _, issue := range s.Issues {
		if issue.State != nil && issue.State.ID == stateID {
			count++
		}
	}
	return count
}

// CountIssuesByPriority counts issues with a given priority level.
func (s *AppState) CountIssuesByPriority(priority int) int {
	count := 0
	for _, issue := range s.Issues {
		if issue.Priority == priority {
			count++
		}
	}
	return count
}
