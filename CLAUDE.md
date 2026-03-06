# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LazyLinear is a terminal UI application for Linear issue tracking, inspired by lazygit. It's a keyboard-driven TUI for managing Linear issues without leaving the terminal.

## Tech Stack

- **Language**: Go
- **TUI Framework**: Bubble Tea + Lip Gloss + Bubbles
- **API**: Linear GraphQL API (`https://api.linear.app/graphql`)
- **Config**: YAML configuration at `~/.config/lazylinear/config.yaml`

## Build Commands

```bash
go mod init github.com/mane-pal/lazylinear
go build ./cmd/lazylinear
go run ./cmd/lazylinear
go test ./...
go test -v ./pkg/linear/...  # Run specific package tests
```

## Project Structure

```
lazylinear/
├── cmd/lazylinear/main.go         # Entry point
├── pkg/
│   ├── app/                       # App lifecycle & initialization
│   ├── gui/
│   │   ├── gui.go                 # Root tea.Model, layout, Init()
│   │   ├── update.go              # Message routing, global keys
│   │   ├── view.go                # View composition, overlays, help, status bar
│   │   ├── actions.go             # Issue CRUD, menu openers
│   │   ├── forms.go               # Form state & logic
│   │   ├── load.go                # Async data loading commands
│   │   ├── utils.go               # Text utilities
│   │   ├── state/state.go         # Shared AppState struct
│   │   ├── types/
│   │   │   ├── types.go           # ContextKey, ContextKind enums
│   │   │   └── messages.go        # All message types
│   │   ├── context/
│   │   │   ├── context.go         # IContext interface + BaseContext
│   │   │   ├── sidebar_context.go # Sidebar panel (teams, filters, states, priority)
│   │   │   ├── issues_context.go  # Issues/projects list panel
│   │   │   └── detail_context.go  # Issue detail + comments panel
│   │   ├── controllers/
│   │   │   ├── common.go          # ControllerCommon (shared deps)
│   │   │   └── helpers/
│   │   │       ├── issues_helper.go  # Issue CRUD helpers
│   │   │       ├── filter_helper.go  # Search filtering
│   │   │       ├── menu_helper.go    # Menu building
│   │   │       └── load_helper.go    # Async data loading
│   │   ├── popup/
│   │   │   ├── menu.go            # Menu + confirm overlay
│   │   │   └── search.go          # Search + command mode
│   │   └── styles/styles.go       # Lip Gloss styles
│   ├── linear/                    # Linear API client
│   │   └── models/                # Domain models
│   └── config/                    # Configuration loading
```

## Architecture Patterns

**Context = Panel + State + Behavior**: Each panel (sidebar, issues, detail) is a "context" with its own `HandleKey()` and `View()` methods. Contexts receive `*state.AppState` and use closures for actions to avoid circular deps.

**Shared AppState**: A single `*state.AppState` struct holds all shared mutable data (teams, issues, selections, filters, UI state). Contexts and popups receive a pointer and mutate directly.

**Popup structs**: `MenuPopup` and `SearchPopup` own their state and provide `HandleKey()`/`View()`/`IsOpen()`. The root `Gui` checks popup state before routing keys to panel contexts.

**Closures for cross-package actions**: Contexts and popups use function closures (set by `Gui` in `New()`) to trigger actions without importing the `gui` package.

**Messages in types/**: All message types are in `types/messages.go` so both `gui` and subpackages can produce/consume them without circular imports.

**Dependency graph (no circular deps)**:
```
types/                → nothing
state/                → models/
context/              → types/, state/, controllers/helpers/
controllers/helpers/  → types/, state/, linear/
popup/                → types/, state/
gui                   → types/, state/, context/, popup/, controllers/helpers/
```

**Facade for Linear API**: `LinearClient` exposes domain-specific command structs, keeping API logic separate from UI.

## Authentication

Browser OAuth flow via `lazylinear auth login` (like `gh auth login`).
Tokens are stored at `~/.config/lazylinear/credentials.json`.

## Key References

- Linear GraphQL API: https://linear.app/developers/graphql
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- lazygit (primary inspiration): https://github.com/jesseduffield/lazygit
