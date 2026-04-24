package ui

import (
	"context"
	"fmt"
	"io"

	"github.com/gookit/cliui/interact/backend"
)

// MultiSelect collects multiple selected items from the user.
type MultiSelect struct {
	Prompt       string
	Items        []Item
	DefaultKeys  []string
	MinSelected  int
	AllowAbort   bool
	EnableFilter bool
}

// NewMultiSelect creates a MultiSelect component.
func NewMultiSelect(prompt string, items []Item) *MultiSelect {
	return &MultiSelect{Prompt: prompt, Items: items}
}

// Run uses the default process stdin/stdout streams.
func (c *MultiSelect) Run(ctx context.Context, be backend.Backend) (*Result, error) {
	return c.RunWithIO(ctx, be, defaultInput(), defaultOutput())
}

// RunWithIO runs the component with explicit IO streams.
func (c *MultiSelect) RunWithIO(ctx context.Context, be backend.Backend, in io.Reader, out io.Writer) (*Result, error) {
	return runWithSession(ctx, be, in, out, func(session backend.Session) (*Result, error) {
		_ = session
		if err := c.ValidateConfig(); err != nil {
			return nil, err
		}

		return nil, ErrNotImplemented
	})
}

// ValidateConfig checks whether the component has enough data to run.
func (c *MultiSelect) ValidateConfig() error {
	if c == nil {
		return fmt.Errorf("%w: multi select is nil", ErrInvalidState)
	}
	if c.Prompt == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidState)
	}
	if c.MinSelected < 0 {
		return fmt.Errorf("%w: min selected must be >= 0", ErrInvalidState)
	}
	if err := validateItems(c.Items); err != nil {
		return err
	}
	return nil
}
