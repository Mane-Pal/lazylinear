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

## Project Structure (Planned)

```
lazylinear/
├── cmd/lazylinear/main.go     # Entry point
├── pkg/
│   ├── app/                   # App lifecycle & initialization
│   ├── gui/                   # TUI layer
│   │   ├── context/           # Panel contexts (state + behavior)
│   │   ├── controllers/       # Event handlers (thin)
│   │   │   └── helpers/       # Domain logic helpers
│   │   ├── popup/             # Modal dialogs
│   │   └── styles/            # Lip Gloss styles
│   ├── linear/                # Linear API client
│   │   └── models/            # Domain models
│   └── config/                # Configuration loading
```

## Architecture Patterns

**Context = Panel + State + Behavior**: Each panel is a "context" with its own state, keybindings, and render logic. Context stack enables modal workflows (push popup, pop to return).

**Controller + Helper Separation**: Controllers are thin event handlers that delegate to helpers. Helpers contain reusable domain logic.

**Dependency Injection**: `ControllerCommon` bundles shared deps (gui, state, client, config) and is passed to controllers via constructors.

**Facade for Linear API**: `LinearClient` exposes domain-specific command structs, keeping API logic separate from UI.

## Authentication

Browser OAuth flow via `lazylinear auth login` (like `gh auth login`).
Tokens are stored at `~/.config/lazylinear/credentials.json`.

## Key References

- Linear GraphQL API: https://linear.app/developers/graphql
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- lazygit (primary inspiration): https://github.com/jesseduffield/lazygit
