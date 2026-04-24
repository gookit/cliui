package interact

import (
	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/cliui/interact/backend/fake"
	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/backend/readline"
	"github.com/gookit/cliui/interact/ui"
)

// UIBackend aliases the new backend abstraction.
type UIBackend = backend.Backend

// UIItem aliases the new ui item type.
type UIItem = ui.Item

// UIResult aliases the new ui result type.
type UIResult = ui.Result

// UIInput aliases the new input component.
type UIInput = ui.Input

// UIConfirm aliases the new confirm component.
type UIConfirm = ui.Confirm

// UISelect aliases the new select component.
type UISelect = ui.Select

// UIMultiSelect aliases the new multi select component.
type UIMultiSelect = ui.MultiSelect

var (
	// ErrUIAborted reports that the current interaction was aborted.
	ErrUIAborted = ui.ErrAborted
	// ErrUINoTTY reports that a strict terminal backend requires a TTY.
	ErrUINoTTY = ui.ErrNoTTY
	// ErrUIInvalidState reports invalid ui component state.
	ErrUIInvalidState = ui.ErrInvalidState
	// ErrUINotImplemented reports a missing backend implementation.
	ErrUINotImplemented = ui.ErrNotImplemented
)

// NewUIInput creates a ui.Input component.
func NewUIInput(prompt string) *ui.Input {
	return ui.NewInput(prompt)
}

// NewUIConfirm creates a ui.Confirm component.
func NewUIConfirm(prompt string, def bool) *ui.Confirm {
	return ui.NewConfirm(prompt, def)
}

// NewUISelect creates a ui.Select component.
func NewUISelect(prompt string, items []ui.Item) *ui.Select {
	return ui.NewSelect(prompt, items)
}

// NewUIMultiSelect creates a ui.MultiSelect component.
func NewUIMultiSelect(prompt string, items []ui.Item) *ui.MultiSelect {
	return ui.NewMultiSelect(prompt, items)
}

// NewUIPlainBackend creates a line-based backend for ui components.
func NewUIPlainBackend() backend.Backend {
	return plain.New()
}

// NewUIReadlineBackend creates a raw-terminal backend that falls back to plain mode on non-TTY input.
func NewUIReadlineBackend() backend.Backend {
	return readline.New()
}

// NewUIStrictReadlineBackend creates a strict raw-terminal backend that requires a TTY.
func NewUIStrictReadlineBackend() backend.Backend {
	return readline.NewStrict()
}

// NewUIFakeBackend creates an in-memory backend for tests or scripted event flows.
func NewUIFakeBackend(events ...backend.Event) backend.Backend {
	return fake.New(events...)
}
