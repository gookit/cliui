package ui

import (
	"context"
	"fmt"
	"io"

	"github.com/gookit/cliui/interact/backend"
)

// Select collects one selected item from the user.
type Select struct {
	Prompt       string
	Items        []Item
	DefaultKey   string
	AllowAbort   bool
	EnableFilter bool
}

// NewSelect creates a Select component.
func NewSelect(prompt string, items []Item) *Select {
	return &Select{Prompt: prompt, Items: items}
}

// Run uses the default process stdin/stdout streams.
func (c *Select) Run(ctx context.Context, be backend.Backend) (*Result, error) {
	return c.RunWithIO(ctx, be, defaultInput(), defaultOutput())
}

// RunWithIO runs the component with explicit IO streams.
func (c *Select) RunWithIO(ctx context.Context, be backend.Backend, in io.Reader, out io.Writer) (*Result, error) {
	return runWithSession(ctx, be, in, out, func(session backend.Session) (*Result, error) {
		_ = session
		if err := c.ValidateConfig(); err != nil {
			return nil, err
		}

		return nil, ErrNotImplemented
	})
}

// ValidateConfig checks whether the component has enough data to run.
func (c *Select) ValidateConfig() error {
	if c == nil {
		return fmt.Errorf("%w: select is nil", ErrInvalidState)
	}
	if c.Prompt == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidState)
	}
	if err := validateItems(c.Items); err != nil {
		return err
	}
	return nil
}
