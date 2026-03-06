package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/auth"
	"github.com/mane-pal/lazylinear/pkg/config"
	"github.com/mane-pal/lazylinear/pkg/gui"
	"github.com/mane-pal/lazylinear/pkg/linear"
)

// Options holds CLI options for the application
type Options struct {
	CreateIssue bool   // Open directly to create issue form
	TeamName    string // Pre-select team by name
}

// Run starts the application
func Run(version string, opts Options) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	// Resolve authentication
	creds, err := auth.Resolve()
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	// Create Linear API client
	client, err := linear.NewClient(creds)
	if err != nil {
		return fmt.Errorf("linear client: %w", err)
	}

	// Create GUI with options
	guiOpts := gui.Options{
		CreateIssue: opts.CreateIssue,
		TeamName:    opts.TeamName,
	}
	g := gui.New(cfg, client, version, guiOpts)

	// Create and run program
	p := tea.NewProgram(g, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("program: %w", err)
	}

	return nil
}
