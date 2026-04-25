package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/goutil/strutil"
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

		buf := []rune{}
		cursor := 0
		errMsg := ""
		for {
			current := string(buf)
			if err := session.Render(c.view(current, cursor, errMsg)); err != nil {
				return "", err
			}

			ev, err := session.ReadEvent(ctx)
			if err != nil {
				return "", err
			}

			if ev.Type == backend.EventInterrupt || ev.Key == backend.KeyCtrlC || ev.Key == backend.KeyEsc {
				return "", ErrAborted
			}

			errMsg = ""
			switch ev.Key {
			case backend.KeyLeft:
				if cursor > 0 {
					cursor--
				}
				continue
			case backend.KeyRight:
				if cursor < len(buf) {
					cursor++
				}
				continue
			case backend.KeyCtrlA, backend.KeyHome:
				cursor = 0
				continue
			case backend.KeyCtrlE, backend.KeyEnd:
				cursor = len(buf)
				continue
			case backend.KeyCtrlK:
				if cursor < len(buf) {
					buf = buf[:cursor]
				}
				continue
			case backend.KeyCtrlU:
				if cursor > 0 {
					buf = buf[cursor:]
					cursor = 0
				}
				continue
			case backend.KeyCtrlW:
				if cursor == 0 || len(buf) == 0 {
					continue
				}

				start := cursor
				// trim spaces before the current word
				for start > 0 && buf[start-1] == ' ' {
					start--
				}
				// remove the previous word
				for start > 0 && buf[start-1] != ' ' {
					start--
				}

				buf = append(buf[:start], buf[cursor:]...)
				cursor = start
				continue
			case backend.KeyBackspace:
				if cursor > 0 && len(buf) > 0 {
					buf = append(buf[:cursor-1], buf[cursor:]...)
					cursor--
				}
				continue
			case backend.KeyDelete:
				if cursor < len(buf) && len(buf) > 0 {
					buf = append(buf[:cursor], buf[cursor+1:]...)
				}
				continue
			case backend.KeyEnter:
				val := strings.TrimSpace(ev.Text)
				// event-driven backend submits current buffer on enter.
				if val == "" && len(buf) > 0 {
					val = string(buf)
				}
				if val == "" {
					val = c.Default
				}

				if c.Validate != nil {
					if err := c.Validate(val); err != nil {
						errMsg = err.Error()
						continue
					}
				}

				return val, nil
			}

			if ev.Text != "" {
				text := []rune(ev.Text)
				if cursor >= len(buf) {
					buf = append(buf, text...)
				} else {
					buf = append(buf[:cursor], append(text, buf[cursor:]...)...)
				}
				cursor += len(text)
				continue
			}

			val := strings.TrimSpace(ev.Text)
			if val == "" {
				val = c.Default
			}

			if c.Validate != nil {
				if err := c.Validate(val); err != nil {
					errMsg = err.Error()
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

func (c *Input) view(current string, cursor int, errMsg string) backend.View {
	line := c.Prompt
	if c.Default != "" {
		line += fmt.Sprintf(" [%s]", c.Default)
	}
	line += ":"

	view := backend.View{Lines: []string{line}}
	editLine := "Current: " + current
	view.Lines = append(view.Lines, editLine)
	if errMsg != "" {
		view.Lines = append(view.Lines, "Error: "+errMsg)
	}

	view.CursorRow = 1
	view.CursorColumn = len("Current: ") + strutil.TextWidth(string([]rune(current)[:cursor]))

	return view
}
