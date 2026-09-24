// Package plain provides a line-based backend for interact/ui.
package plain

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/color"
)

// Backend is a simple line-based backend that works with ordinary IO streams.
type Backend struct{}

// New creates a plain backend.
func New() *Backend {
	return &Backend{}
}

// NewSession creates a new plain backend session.
func (b *Backend) NewSession(in io.Reader, out io.Writer) (backend.Session, error) {
	return &Session{
		in:  bufio.NewReader(in),
		out: out,
	}, nil
}

// lineResult is one line (or error) read from the input stream.
type lineResult struct {
	text string
	err  error
}

// Session implements backend.Session with line-based input.
type Session struct {
	in    *bufio.Reader
	out   io.Writer
	once  sync.Once
	lines chan lineResult
}

// startReader starts a single reader goroutine for the session. Using one
// goroutine instead of one per ReadEvent avoids two readers racing on the
// same bufio.Reader when a call is cancelled by ctx.
func (s *Session) startReader() {
	s.once.Do(func() {
		s.lines = make(chan lineResult, 1)
		go func() {
			for {
				line, err := s.in.ReadString('\n')
				if err != nil && len(line) == 0 {
					s.lines <- lineResult{err: err}
					return
				}

				s.lines <- lineResult{text: strings.TrimSpace(line)}
			}
		}()
	})
}

// Render writes the current view as plain text.
func (s *Session) Render(view backend.View) error {
	if len(view.Lines) == 0 {
		return nil
	}

	_, err := fmt.Fprintln(s.out, color.Render(strings.Join(view.Lines, "\n")))
	return err
}

// ReadEvent reads one line and normalizes it as an enter event.
func (s *Session) ReadEvent(ctx context.Context) (backend.Event, error) {
	s.startReader()

	select {
	case r := <-s.lines:
		if r.err != nil {
			return backend.Event{}, r.err
		}

		return backend.Event{
			Type: backend.EventKey,
			Key:  backend.KeyEnter,
			Text: r.text,
		}, nil
	case <-ctx.Done():
		return backend.Event{}, ctx.Err()
	}
}

// Size returns zero values because the plain backend does not depend on terminal size.
func (s *Session) Size() (width, height int) {
	return 0, 0
}

// Close closes the session. Plain sessions do not own IO streams.
func (s *Session) Close() error {
	return nil
}
