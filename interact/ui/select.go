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

		cursor := c.defaultIndex()
		for {
			if err := session.Render(c.view(cursor, "")); err != nil {
				return nil, err
			}

			ev, err := session.ReadEvent(ctx)
			if err != nil {
				return nil, err
			}

			if ev.Type == backend.EventInterrupt || ev.Key == backend.KeyCtrlC || ev.Key == backend.KeyEsc {
				return nil, ErrAborted
			}

			switch ev.Key {
			case backend.KeyUp:
				cursor = c.move(cursor, -1)
				continue
			case backend.KeyDown:
				cursor = c.move(cursor, 1)
				continue
			case backend.KeyEnter:
				if strings.TrimSpace(ev.Text) == "" {
					item := c.Items[cursor]
					if item.Disabled {
						if err := session.Render(c.view(cursor, "selected option is disabled")); err != nil {
							return nil, err
						}
						continue
					}
					return singleResult(item), nil
				}
			}

			key := strings.TrimSpace(ev.Text)
			if key == "" {
				key = c.DefaultKey
			}

			item, ok := c.findItem(key)
			if !ok {
				if err := session.Render(c.view(cursor, "unknown option")); err != nil {
					return nil, err
				}
				continue
			}
			if item.Disabled {
				if err := session.Render(c.view(cursor, "selected option is disabled")); err != nil {
					return nil, err
				}
				continue
			}

			return singleResult(item), nil
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

func (c *Select) defaultIndex() int {
	if c.DefaultKey != "" {
		for i, item := range c.Items {
			if item.Key == c.DefaultKey {
				return i
			}
		}
	}

	for i, item := range c.Items {
		if !item.Disabled {
			return i
		}
	}

	return 0
}

func (c *Select) move(cursor, delta int) int {
	if len(c.Items) == 0 {
		return 0
	}

	next := cursor
	for range c.Items {
		next = (next + delta + len(c.Items)) % len(c.Items)
		if !c.Items[next].Disabled {
			return next
		}
	}

	return cursor
}

func (c *Select) view(cursor int, errMsg string) backend.View {
	lines := []string{c.Prompt}
	for i, item := range c.Items {
		prefix := " "
		if i == cursor {
			prefix = ">"
		}

		line := fmt.Sprintf("%s %s) %s", prefix, item.Key, item.Label)
		if item.Disabled {
			line += " [disabled]"
		}
		lines = append(lines, line)
	}

	current := c.Items[cursor]
	currentLine := fmt.Sprintf("Current: %s (%s)", current.Label, current.Key)
	if current.Disabled {
		currentLine += " [disabled]"
	}
	lines = append(lines, currentLine)

	hint := "Use Up/Down to move, Enter to confirm, or input item key"
	lines = append(lines, hint)

	prompt := "Your choice"
	if c.DefaultKey != "" {
		prompt += fmt.Sprintf(" [default:%s]", c.DefaultKey)
	}
	prompt += ":"
	lines = append(lines, prompt)

	if errMsg != "" {
		lines = append(lines, "Error: "+errMsg)
	}

	return backend.View{
		Lines:        lines,
		CursorRow:    len(c.Items) + 3,
		CursorColumn: len(prompt),
	}
}

func singleResult(item Item) *Result {
	return &Result{
		Key:   item.Key,
		Keys:  []string{item.Key},
		Value: item.Value,
		Values: []any{
			item.Value,
		},
	}
}
