// Package readline provides a minimal raw-terminal event backend for interact/ui.
package readline

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/cliui/interact/backend/plain"
	"golang.org/x/term"
)

// Backend creates raw-terminal sessions backed by golang.org/x/term.
type Backend struct {
	// Fallback is used when input is not a real terminal stream.
	Fallback backend.Backend
	// StrictTTY disables fallback and returns a TTY-related error instead.
	StrictTTY bool
}

// New creates a readline backend.
func New() *Backend {
	return &Backend{Fallback: plain.New()}
}

// NewStrict creates a readline backend that fails when a TTY is unavailable.
func NewStrict() *Backend {
	return &Backend{StrictTTY: true}
}

// NewSession creates a new raw-terminal session.
func (b *Backend) NewSession(in io.Reader, out io.Writer) (backend.Session, error) {
	file, ok := in.(*os.File)
	if !ok {
		if !b.StrictTTY && b.Fallback != nil {
			return b.Fallback.NewSession(in, out)
		}
		return nil, fmt.Errorf("%w: readline backend requires *os.File input", backend.ErrFileRequired)
	}

	if !term.IsTerminal(int(file.Fd())) {
		if !b.StrictTTY && b.Fallback != nil {
			return b.Fallback.NewSession(in, out)
		}
		return nil, fmt.Errorf("%w: readline backend requires a terminal input", backend.ErrTTYRequired)
	}

	state, err := term.MakeRaw(int(file.Fd()))
	if err != nil {
		return nil, err
	}

	return &Session{
		inFile: file,
		in:     bufio.NewReader(file),
		out:    out,
		state:  state,
	}, nil
}

// Session is a raw-terminal session.
type Session struct {
	inFile   *os.File
	in       *bufio.Reader
	out      io.Writer
	state    *term.State
	rendered int
}

// Render redraws the current interaction block.
func (s *Session) Render(view backend.View) error {
	if len(view.Lines) == 0 {
		return nil
	}

	if s.rendered > 0 {
		fmt.Fprint(s.out, "\r")
		if s.rendered > 1 {
			fmt.Fprintf(s.out, "\x1B[%dA", s.rendered-1)
		}
	}

	for i, line := range view.Lines {
		fmt.Fprint(s.out, "\x1B[2K")
		fmt.Fprint(s.out, line)
		if i < len(view.Lines)-1 {
			fmt.Fprint(s.out, "\n")
		}
	}

	if view.CursorRow >= 0 && view.CursorRow < len(view.Lines) {
		moveUp := len(view.Lines) - 1 - view.CursorRow
		if moveUp > 0 {
			fmt.Fprintf(s.out, "\x1B[%dA", moveUp)
		}
		fmt.Fprint(s.out, "\r")
		if view.CursorColumn > 0 {
			fmt.Fprintf(s.out, "\x1B[%dC", view.CursorColumn)
		}
	}

	s.rendered = len(view.Lines)
	return nil
}

// ReadEvent reads and normalizes one terminal event.
func (s *Session) ReadEvent(ctx context.Context) (backend.Event, error) {
	select {
	case <-ctx.Done():
		return backend.Event{}, ctx.Err()
	default:
	}

	b, err := s.in.ReadByte()
	if err != nil {
		return backend.Event{}, err
	}

	switch b {
	case 1:
		return backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlA}, nil
	case 3:
		return backend.Event{Type: backend.EventInterrupt, Key: backend.KeyCtrlC}, nil
	case 5:
		return backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlE}, nil
	case 13:
		return backend.Event{Type: backend.EventKey, Key: backend.KeyEnter}, nil
	case 27:
		next, err := s.in.Peek(3)
		if err == nil && len(next) >= 2 && next[0] == '[' {
			// CSI sequences
			switch next[1] {
			case 'A':
				_, _ = s.in.Discard(2)
				return backend.Event{Type: backend.EventKey, Key: backend.KeyUp}, nil
			case 'B':
				_, _ = s.in.Discard(2)
				return backend.Event{Type: backend.EventKey, Key: backend.KeyDown}, nil
			case 'C':
				_, _ = s.in.Discard(2)
				return backend.Event{Type: backend.EventKey, Key: backend.KeyRight}, nil
			case 'D':
				_, _ = s.in.Discard(2)
				return backend.Event{Type: backend.EventKey, Key: backend.KeyLeft}, nil
			case 'H':
				_, _ = s.in.Discard(2)
				return backend.Event{Type: backend.EventKey, Key: backend.KeyHome}, nil
			case 'F':
				_, _ = s.in.Discard(2)
				return backend.Event{Type: backend.EventKey, Key: backend.KeyEnd}, nil
			case '3':
				if len(next) >= 3 && next[2] == '~' {
					_, _ = s.in.Discard(3)
					return backend.Event{Type: backend.EventKey, Key: backend.KeyDelete}, nil
				}
			case '1':
				if len(next) >= 3 && next[2] == '~' {
					_, _ = s.in.Discard(3)
					return backend.Event{Type: backend.EventKey, Key: backend.KeyHome}, nil
				}
			case '4':
				if len(next) >= 3 && next[2] == '~' {
					_, _ = s.in.Discard(3)
					return backend.Event{Type: backend.EventKey, Key: backend.KeyEnd}, nil
				}
			}
		}
		return backend.Event{Type: backend.EventInterrupt, Key: backend.KeyEsc}, nil
	case 32:
		return backend.Event{Type: backend.EventKey, Key: backend.KeySpace}, nil
	case 127, 8:
		return backend.Event{Type: backend.EventKey, Key: backend.KeyBackspace}, nil
	default:
		text := strings.TrimSpace(string([]byte{b}))
		key := backend.KeyUnknown
		switch strings.ToLower(text) {
		case "y":
			key = backend.KeyY
		case "n":
			key = backend.KeyN
		}

		return backend.Event{Type: backend.EventKey, Key: key, Text: text}, nil
	}
}

// Size returns the current terminal size.
func (s *Session) Size() (width, height int) {
	w, h, err := term.GetSize(int(s.inFile.Fd()))
	if err != nil {
		return 0, 0
	}
	return w, h
}

// Close restores terminal state and ends the rendered block.
func (s *Session) Close() error {
	if s.rendered > 0 {
		fmt.Fprint(s.out, "\r")
		if s.rendered > 1 {
			fmt.Fprintf(s.out, "\x1B[%dB", s.rendered-1)
		}
		fmt.Fprintln(s.out)
	}

	if s.state != nil {
		return term.Restore(int(s.inFile.Fd()), s.state)
	}
	return nil
}
