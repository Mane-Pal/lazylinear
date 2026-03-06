package controllers

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/config"
	"github.com/mane-pal/lazylinear/pkg/gui/state"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear"
)

// ControllerCommon bundles shared dependencies and UI action closures.
// Controllers receive this at construction time; closures avoid circular deps
// (controllers never import the gui package).
type ControllerCommon struct {
	State  *state.AppState
	Client *linear.Client
	Config *config.Config

	// UI action closures — set by Gui at construction time.
	SetStatus    func(string) tea.Cmd
	ReloadIssues func() tea.Cmd
	OpenMenu     func(menuType types.MenuType, title string, items []types.MenuItem, selected int)
	CloseMenu    func()
	LoadIssues   func() tea.Cmd
	LoadProjects func() tea.Cmd
}
