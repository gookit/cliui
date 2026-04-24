package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

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
		if err := c.ValidateConfig(); err != nil {
			return nil, err
		}

		for {
			if err := session.Render(c.view("")); err != nil {
				return nil, err
			}

			ev, err := session.ReadEvent(ctx)
			if err != nil {
				return nil, err
			}

			if ev.Type == backend.EventInterrupt || ev.Key == backend.KeyCtrlC || ev.Key == backend.KeyEsc {
				return nil, ErrAborted
			}

			key := strings.TrimSpace(ev.Text)
			if key == "" {
				key = c.DefaultKey
			}

			item, ok := c.findItem(key)
			if !ok {
				if err := session.Render(c.view("unknown option")); err != nil {
					return nil, err
				}
				continue
			}
			if item.Disabled {
				if err := session.Render(c.view("selected option is disabled")); err != nil {
					return nil, err
				}
				continue
			}

			return &Result{
				Key:   item.Key,
				Keys:  []string{item.Key},
				Value: item.Value,
				Values: []any{
					item.Value,
				},
			}, nil
		}
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

func (c *Select) findItem(key string) (Item, bool) {
	for _, item := range c.Items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func (c *Select) view(errMsg string) backend.View {
	lines := []string{c.Prompt}
	for _, item := range c.Items {
		line := fmt.Sprintf("  %s) %s", item.Key, item.Label)
		if item.Disabled {
			line += " [disabled]"
		}
		lines = append(lines, line)
	}

	prompt := "Your choice"
	if c.DefaultKey != "" {
		prompt += fmt.Sprintf(" [default:%s]", c.DefaultKey)
	}
	prompt += ":"
	lines = append(lines, prompt)

	if errMsg != "" {
		lines = append(lines, "Error: "+errMsg)
	}

	return backend.View{Lines: lines}
}
