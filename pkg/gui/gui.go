package gui

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/config"
	guicontext "github.com/mane-pal/lazylinear/pkg/gui/context"
	"github.com/mane-pal/lazylinear/pkg/gui/popup"
	"github.com/mane-pal/lazylinear/pkg/gui/state"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// Gui is the main application model
type Gui struct {
	// Dependencies
	config  *config.Config
	client  *linear.Client
	version string

	// Shared state
	State *state.AppState

	// Contexts (panels)
	SidebarCtx *guicontext.SidebarContext
	IssuesCtx  *guicontext.IssuesContext
	DetailCtx  *guicontext.DetailContext

	// Dimensions
	width  int
	height int

	// Popups
	MenuPopup   *popup.MenuPopup
	SearchPopup *popup.SearchPopup

	// Form state (to be extracted to popup/form.go later)
	formMode        types.FormMode
	formField       types.FormField
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
	formListField    types.FormField // Which field the popup is for
	formListSearch   textinput.Model // Search input for filtering
	formListSelected int             // Currently highlighted item in the list

	// Issue picker state (for parent/relation target)
	issuePickerCallback func(issueID string) tea.Cmd // Callback when issue is selected

	// Due date / estimate input state
	dueDateInput   textinput.Model // Input for due date
	estimateInput  textinput.Model // Input for estimate
	inputIssueID   string          // Issue ID being edited

	// Comment editing state
	commentBody      textarea.Model
	editingCommentID string // For edit comment mode

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
	st := &state.AppState{
		FocusedPanel:         types.IssuesContext,
		Loading:              true,
		ActiveFilter:         "all",
		ActiveStateFilters:   make(map[string]bool),
		ActivePriorityFilter: -1, // No priority filter by default
	}

	g := &Gui{
		config:  cfg,
		client:  client,
		version: version,
		State:   st,
		cliOpts: opts,
	}

	// Create popups
	g.MenuPopup = popup.NewMenuPopup(st)
	g.MenuPopup.UpdateIssueState = func(issueID, stateID string) tea.Cmd { return g.updateIssueState(issueID, stateID) }
	g.MenuPopup.UpdateIssueAssignee = func(issueID, assigneeID string) tea.Cmd { return g.updateIssueAssignee(issueID, assigneeID) }
	g.MenuPopup.UpdateIssuePriority = func(issueID, priorityStr string) tea.Cmd { return g.updateIssuePriority(issueID, priorityStr) }
	g.MenuPopup.UpdateIssueCycle = func(issueID string, cycleID *string) tea.Cmd { return g.updateIssueCycle(issueID, cycleID) }
	g.MenuPopup.UpdateIssueLabels = func(issueID string, labelIDs []string) tea.Cmd { return g.updateIssueLabels(issueID, labelIDs) }

	g.SearchPopup = popup.NewSearchPopup(st)
	g.SearchPopup.LoadIssues = func() tea.Cmd { return g.loadIssues() }
	g.SearchPopup.LoadSelectedIssueDetails = func() tea.Cmd { return g.loadSelectedIssueDetails() }

	// Create contexts
	g.SidebarCtx = guicontext.NewSidebarContext(st)
	g.IssuesCtx = guicontext.NewIssuesContext(st)
	g.DetailCtx = guicontext.NewDetailContext(st)

	// Wire sidebar closures
	g.SidebarCtx.LoadIssues = func() tea.Cmd { return g.loadIssues() }
	g.SidebarCtx.LoadTeamStates = func() tea.Cmd { return g.loadTeamStates() }

	// Wire issues closures
	g.IssuesCtx.LoadSelectedIssueDetails = func() tea.Cmd { return g.loadSelectedIssueDetails() }
	g.IssuesCtx.OpenInBrowser = func(url string) tea.Cmd { return g.openInBrowser(url) }
	g.IssuesCtx.CopyToClipboard = func(text string) tea.Cmd { return g.copyToClipboard(text) }
	g.IssuesCtx.OpenStateMenu = func() tea.Cmd { return g.openStateMenu() }
	g.IssuesCtx.OpenAssignMenu = func() tea.Cmd { return g.openAssignMenu() }
	g.IssuesCtx.OpenPriorityMenu = func() { g.openPriorityMenu() }
	g.IssuesCtx.OpenCreateFormCmd = func() tea.Cmd { return g.openCreateFormCmd() }
	g.IssuesCtx.OpenEditForm = func(issueID string) {
		if issue := g.findIssueByID(issueID); issue != nil {
			g.openEditForm(issue)
		}
	}
	g.IssuesCtx.OpenArchiveConfirm = func(issueID string) {
		if issue := g.findIssueByID(issueID); issue != nil {
			g.openArchiveConfirm(issue)
		}
	}
	g.IssuesCtx.OpenLabelsMenu = func() tea.Cmd { return g.openLabelsMenu() }
	g.IssuesCtx.OpenCycleMenu = func() tea.Cmd { return g.openCycleMenu() }

	// Wire detail closures
	g.DetailCtx.OpenInBrowser = func(url string) tea.Cmd { return g.openInBrowser(url) }
	g.DetailCtx.CopyToClipboard = func(text string) tea.Cmd { return g.copyToClipboard(text) }
	g.DetailCtx.OpenStateMenu = func() tea.Cmd { return g.openStateMenu() }
	g.DetailCtx.OpenAssignMenu = func() tea.Cmd { return g.openAssignMenu() }
	g.DetailCtx.OpenPriorityMenu = func() { g.openPriorityMenu() }
	g.DetailCtx.OpenLabelsMenu = func() tea.Cmd { return g.openLabelsMenu() }
	g.DetailCtx.OpenCycleMenu = func() tea.Cmd { return g.openCycleMenu() }
	g.DetailCtx.OpenCommentForm = func() { g.openCommentForm() }
	g.DetailCtx.OpenEditForm = func(issueID string) {
		if issue := g.findIssueByID(issueID); issue != nil {
			g.openEditForm(issue)
		}
	}
	g.DetailCtx.OpenEditCommentForm = func(commentID string) {
		if comment := g.findCommentByID(commentID); comment != nil {
			g.openEditCommentForm(comment)
		}
	}
	g.DetailCtx.OpenDeleteCommentConfirm = func(commentID string) {
		if comment := g.findCommentByID(commentID); comment != nil {
			g.openDeleteCommentConfirm(comment)
		}
	}
	g.DetailCtx.OpenArchiveConfirm = func(issueID string) {
		if issue := g.findIssueByID(issueID); issue != nil {
			g.openArchiveConfirm(issue)
		}
	}
	g.DetailCtx.OpenDueDateInput = func(issueID string) {
		if issue := g.findIssueByID(issueID); issue != nil {
			g.openDueDateInput(issue)
		}
	}
	g.DetailCtx.OpenEstimateInput = func(issueID string) {
		if issue := g.findIssueByID(issueID); issue != nil {
			g.openEstimateInput(issue)
		}
	}
	g.DetailCtx.OpenParentPicker = func(issueID string) tea.Cmd {
		if issue := g.findIssueByID(issueID); issue != nil {
			return g.openParentPicker(issue)
		}
		return nil
	}
	g.DetailCtx.OpenRelationTypeMenu = func() { g.openRelationTypeMenu() }
	g.DetailCtx.OpenCreateSubIssueFormCmd = func(issueID string) tea.Cmd {
		if issue := g.findIssueByID(issueID); issue != nil {
			return g.openCreateSubIssueFormCmd(issue)
		}
		return nil
	}
	g.DetailCtx.LoadSelectedIssueDetails = func() tea.Cmd { return g.loadSelectedIssueDetails() }

	return g
}

// findIssueByID looks up an issue by ID from the current issues or detailed issue.
func (g *Gui) findIssueByID(id string) *models.Issue {
	if g.State.DetailedIssue != nil && g.State.DetailedIssue.ID == id {
		return g.State.DetailedIssue
	}
	for _, issue := range g.State.Issues {
		if issue.ID == id {
			return issue
		}
	}
	return g.State.GetSelectedIssue()
}

// findCommentByID looks up a comment by ID from the detailed issue.
func (g *Gui) findCommentByID(id string) *models.Comment {
	if g.State.DetailedIssue == nil {
		return nil
	}
	for _, comment := range g.State.DetailedIssue.GetComments() {
		if comment.ID == id {
			return comment
		}
	}
	return nil
}

// clearStatusAfter returns a command that clears the status after a delay
func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return types.ClearStatusMsg{}
	})
}

// backgroundSyncTick returns a command that triggers background sync
func backgroundSyncTick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return types.BackgroundSyncMsg{}
	})
}

// Init initializes the model
func (g *Gui) Init() tea.Cmd {
	g.State.LastSynced = time.Now()
	return tea.Batch(
		g.loadTeams(),
		g.loadIssues(),
		g.loadUser(),
		backgroundSyncTick(),
	)
}
