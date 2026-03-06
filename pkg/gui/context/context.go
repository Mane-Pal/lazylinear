package context

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/gui/types"
)

// IContext is the interface every panel context implements.
type IContext interface {
	GetKey() types.ContextKey
	GetKind() types.ContextKind
	HandleKey(msg tea.KeyMsg) tea.Cmd
	View(width, height int) string
}

// BaseContext provides shared fields for all contexts.
type BaseContext struct {
	key  types.ContextKey
	kind types.ContextKind
}

// NewBaseContext creates a new BaseContext.
func NewBaseContext(key types.ContextKey, kind types.ContextKind) BaseContext {
	return BaseContext{key: key, kind: kind}
}

func (c *BaseContext) GetKey() types.ContextKey   { return c.key }
func (c *BaseContext) GetKind() types.ContextKind { return c.kind }
