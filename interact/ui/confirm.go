package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gookit/cliui/interact/backend"
)

// Confirm collects a boolean choice from the user.
type Confirm struct {
	Prompt  string
	Default bool
}

// NewConfirm creates a Confirm component.
func NewConfirm(prompt string, def bool) *Confirm {
	return &Confirm{Prompt: prompt, Default: def}
}

// Run uses the default process stdin/stdout streams.
func (c *Confirm) Run(ctx context.Context, be backend.Backend) (bool, error) {
	return c.RunWithIO(ctx, be, defaultInput(), defaultOutput())
}

// RunWithIO runs the component with explicit IO streams.
func (c *Confirm) RunWithIO(ctx context.Context, be backend.Backend, in io.Reader, out io.Writer) (bool, error) {
	return runWithSession(ctx, be, in, out, func(session backend.Session) (bool, error) {
		if err := c.ValidateConfig(); err != nil {
			return false, err
		}

		current := c.Default
		errMsg := ""
		for {
			if err := session.Render(c.view(current, errMsg)); err != nil {
				return false, err
			}

			ev, err := session.ReadEvent(ctx)
			if err != nil {
				return false, err
			}

			if ev.Type == backend.EventInterrupt || ev.Key == backend.KeyCtrlC || ev.Key == backend.KeyEsc {
				return false, ErrAborted
			}

			errMsg = ""
			switch ev.Key {
			case backend.KeyLeft:
				current = true
				continue
			case backend.KeyRight:
				current = false
				continue
			case backend.KeyY:
				return true, nil
			case backend.KeyN:
				return false, nil
			case backend.KeyEnter:
				if strings.TrimSpace(ev.Text) == "" {
					return current, nil
				}
			}

			text := strings.ToLower(strings.TrimSpace(ev.Text))
			switch text {
			case "":
				return current, nil
			case "y", "yes":
				return true, nil
			case "n", "no":
				return false, nil
			default:
				errMsg = "please input yes or no"
			}
		}
	})
}

// ValidateConfig checks whether the component has enough data to run.
func (c *Confirm) ValidateConfig() error {
	if c == nil {
		return fmt.Errorf("%w: confirm is nil", ErrInvalidState)
	}
	if c.Prompt == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidState)
	}
	return nil
}

func (c *Confirm) view(current bool, errMsg string) backend.View {
	defVal := "no"
	if c.Default {
		defVal = "yes"
	}

	line := fmt.Sprintf("%s [yes/no] (default: %s):", c.Prompt, defVal)
	curr := "Current: no"
	if current {
		curr = "Current: yes"
	}

	view := backend.View{Lines: []string{line, curr}}
	if errMsg != "" {
		view.Lines = append(view.Lines, "Error: "+errMsg)
	}

	return view
}
