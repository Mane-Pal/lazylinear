package gui

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/config"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear"
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

// Gui is the main application model
type Gui struct {
	// Dependencies
	config  *config.Config
	client  *linear.Client
	version string

	// Dimensions
	width  int
	height int

	// Data
	teams       []*models.Team
	issues      []*models.Issue
	projects    []*models.Project
	currentUser *models.User
	teamMembers []*models.User
	teamStates  []*models.WorkflowState

	// Selection
	selectedTeam    int
	selectedIssue   int
	selectedProject int

	// Middle pane view mode
	middlePaneView MiddlePaneView

	// Sidebar filter state
	activeFilter    string // "all", "my_issues"
	selectedFilter  int    // Which filter option is highlighted
	sidebarSection  int    // 0 = teams, 1 = filters, 2 = states, 3 = priority

	// State filter (multi-select)
	selectedStateIdx   int             // Navigation cursor in states section
	activeStateFilters map[string]bool // Set of selected state IDs (multi-select)

	// Priority filter state
	activePriorityFilter int // -1 = no filter, 1-4 = priority level
	selectedPriority     int // For navigation (0-3 maps to priority 1-4)

	// Focus
	focusedPanel types.ContextKey

	// UI State
	loading   bool
	err       error
	showHelp  bool
	statusMsg string

	// Menu state
	menuType     MenuType
	menuItems    []MenuItem
	menuSelected int
	menuTitle    string

	// Form state
	formMode        FormMode
	formField       FormField
	formTitle       textinput.Model
	formDescription textarea.Model
	formPriority    int
	formTeam        int    // Index into teams slice
	formAssignee    int    // Index into teamMembers slice (-1 = unassigned)
	formState       int    // Index into teamStates slice (-1 = default)
	formProject     int    // Index into projects slice (-1 = none)
	formEditIssueID string // For edit mode

	// Form list popup state
	formListOpen     bool            // Is the list popup open?
	formListField    FormField       // Which field the popup is for
	formListSearch   textinput.Model // Search input for filtering
	formListSelected int             // Currently highlighted item in the list

	// Confirm dialog
	confirmTitle   string
	confirmMessage string
	confirmAction  func() tea.Cmd

	// Label menu state (multi-select)
	teamLabels     []*models.Label // Labels available for current team
	selectedLabels map[string]bool // Currently selected label IDs in menu

	// Cycle menu state
	teamCycles []*models.Cycle // Cycles available for current team

	// Issue picker state (for parent/relation target)
	issuePickerCallback func(issueID string) tea.Cmd // Callback when issue is selected
	allTeamIssues       []*models.Issue              // All issues for picker

	// Due date / estimate input state
	dueDateInput   textinput.Model // Input for due date
	estimateInput  textinput.Model // Input for estimate
	inputIssueID   string          // Issue ID being edited

	// Detail view state
	detailedIssue    *models.Issue // Issue with full comments
	detailScroll     int           // Scroll position in detail view
	selectedComment  int           // Selected comment index (-1 = none)
	commentBody      textarea.Model
	editingCommentID string // For edit comment mode

	// Search state
	searchMode     bool
	searchInput    textinput.Model
	searchQuery    string
	filteredIssues []*models.Issue // nil = show all issues

	// Vim command mode
	commandMode  bool
	commandInput textinput.Model

	// Background sync
	lastSynced time.Time

	// CLI options
	cliOpts Options
}

// Options holds CLI options for the GUI
type Options struct {
	CreateIssue bool   // Open directly to create issue form
	TeamName    string // Pre-select team by name
}

// New creates a new GUI instance
func New(cfg *config.Config, client *linear.Client, version string, opts Options) *Gui {
	return &Gui{
		config:               cfg,
		client:               client,
		version:              version,
		focusedPanel:         types.IssuesContext,
		loading:              true,
		activeFilter:         "all",
		activeStateFilters:   make(map[string]bool),
		activePriorityFilter: -1, // No priority filter by default
		cliOpts:              opts,
	}
}

// Messages
type teamsLoadedMsg struct{ teams []*models.Team }
type issuesLoadedMsg struct{ issues []*models.Issue }
type projectsLoadedMsg struct{ projects []*models.Project }
type userLoadedMsg struct{ user *models.User }
type teamStatesLoadedMsg struct{ states []*models.WorkflowState }
type teamMembersLoadedMsg struct{ members []*models.User }
type issueUpdatedMsg struct{ issue *models.Issue }
type detailedIssueLoadedMsg struct{ issue *models.Issue }
type commentCreatedMsg struct{ comment *models.Comment }
type commentDeletedMsg struct{ issueID string }
type errMsg struct{ err error }
type statusMsg string
type clearStatusMsg struct{}
type backgroundSyncMsg struct{}

// clearStatusAfter returns a command that clears the status after a delay
func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// backgroundSyncTick returns a command that triggers background sync
func backgroundSyncTick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return backgroundSyncMsg{}
	})
}

// Init initializes the model
func (g *Gui) Init() tea.Cmd {
	g.lastSynced = time.Now()
	return tea.Batch(
		g.loadTeams(),
		g.loadIssues(),
		g.loadUser(),
		backgroundSyncTick(),
	)
}
