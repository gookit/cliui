package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gookit/cliui/interact/backend"
)

// Input collects a text value from the user.
type Input struct {
	Prompt   string
	Default  string
	Validate func(string) error
}

// NewInput creates an Input component.
func NewInput(prompt string) *Input {
	return &Input{Prompt: prompt}
}

// Run uses the default process stdin/stdout streams.
func (c *Input) Run(ctx context.Context, be backend.Backend) (string, error) {
	return c.RunWithIO(ctx, be, defaultInput(), defaultOutput())
}

// RunWithIO runs the component with explicit IO streams.
func (c *Input) RunWithIO(ctx context.Context, be backend.Backend, in io.Reader, out io.Writer) (string, error) {
	return runWithSession(ctx, be, in, out, func(session backend.Session) (string, error) {
		if err := c.ValidateConfig(); err != nil {
			return "", err
		}

		for {
			if err := session.Render(c.view("")); err != nil {
				return "", err
			}

			ev, err := session.ReadEvent(ctx)
			if err != nil {
				return "", err
			}

			if ev.Type == backend.EventInterrupt || ev.Key == backend.KeyCtrlC || ev.Key == backend.KeyEsc {
				return "", ErrAborted
			}

			val := strings.TrimSpace(ev.Text)
			if val == "" {
				val = c.Default
			}

			if c.Validate != nil {
				if err := c.Validate(val); err != nil {
					if renderErr := session.Render(c.view(err.Error())); renderErr != nil {
						return "", renderErr
					}
					continue
				}
			}

			return val, nil
		}
	})
}

// ValidateConfig checks whether the component has enough data to run.
func (c *Input) ValidateConfig() error {
	if c == nil {
		return fmt.Errorf("%w: input is nil", ErrInvalidState)
	}
	if c.Prompt == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidState)
	}
	return nil
}

func (c *Input) view(errMsg string) backend.View {
	line := c.Prompt
	if c.Default != "" {
		line += fmt.Sprintf(" [%s]", c.Default)
	}
	line += ":"

	view := backend.View{Lines: []string{line}}
	if errMsg != "" {
		view.Lines = append(view.Lines, "Error: "+errMsg)
	}

	return view
}
