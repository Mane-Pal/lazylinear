package types

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// MenuType identifies what kind of menu is open
type MenuType int

const (
	MenuNone MenuType = iota
	MenuState
	MenuAssign
	MenuPriority
	MenuConfirm
	MenuLabels       // Multi-select labels
	MenuCycle        // Select cycle/sprint
	MenuRelationType // Select relation type (blocks, related, etc.)
	MenuIssueSelect  // Select an issue (for parent/relation target)
)

// FormMode identifies what form is open
type FormMode int

const (
	FormNone FormMode = iota
	FormCreate
	FormEdit
	FormComment
	FormEditComment
	FormDueDate
	FormEstimate
)

// FormField identifies which field is focused
type FormField int

const (
	FieldTitle FormField = iota
	FieldDescription
	FieldTeam
	FieldAssignee
	FieldState
	FieldPriority
	FieldProject
)

// MenuItem represents an item in a menu
type MenuItem struct {
	ID       string
	Label    string
	Color    lipgloss.Color
	Selected bool // For multi-select menus (e.g., labels)
}

// MiddlePaneView identifies what the middle pane shows
type MiddlePaneView int

const (
	ViewIssues MiddlePaneView = iota
	ViewProjects
)

// Messages

type TeamsLoadedMsg struct{ Teams []*models.Team }
type IssuesLoadedMsg struct{ Issues []*models.Issue }
type ProjectsLoadedMsg struct{ Projects []*models.Project }
type UserLoadedMsg struct{ User *models.User }
type TeamStatesLoadedMsg struct{ States []*models.WorkflowState }
type TeamMembersLoadedMsg struct{ Members []*models.User }
type IssueUpdatedMsg struct{ Issue *models.Issue }
type DetailedIssueLoadedMsg struct{ Issue *models.Issue }
type CommentCreatedMsg struct{ Comment *models.Comment }
type CommentDeletedMsg struct{ IssueID string }
type ErrMsg struct{ Err error }
type StatusMsg string
type ClearStatusMsg struct{}
type BackgroundSyncMsg struct{}

// Menu-specific messages
type StateMenuMsg struct {
	States []*models.WorkflowState
	Issue  *models.Issue
}

type AssignMenuMsg struct {
	Members []*models.User
	Issue   *models.Issue
}

type ReloadIssuesMsg struct{}

type LabelsMenuMsg struct {
	Labels []*models.Label
	Issue  *models.Issue
}

type CyclesMenuMsg struct {
	Cycles []*models.Cycle
	Issue  *models.Issue
}

type ParentPickerMsg struct {
	Issues []*models.Issue
	Issue  *models.Issue
}
