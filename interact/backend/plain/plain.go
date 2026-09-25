// Package plain provides a line-based backend for interact/ui.
package plain

import (
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
	return &Session{in: in, out: out}, nil
}

// lineResult is one line (or error) read from the input stream.
type lineResult struct {
	text string
	err  error
}

// Session implements backend.Session with line-based input.
type Session struct {
	in  io.Reader
	out io.Writer

	mu      sync.Mutex
	pending chan lineResult
}

// readLine reads one line from in, without the trailing newline.
//
// It reads one byte at a time on purpose: any read-ahead would be buffered
// inside this session, while the caller may need those bytes for its next
// prompt. Bytes following the returned line must stay in the stream.
func readLine(in io.Reader) (string, error) {
	var line []byte
	var b [1]byte

	for {
		n, err := in.Read(b[:])
		if n > 0 {
			if b[0] == '\n' {
				return string(line), nil
			}
			line = append(line, b[0])
		}

		if err != nil {
			if len(line) > 0 {
				// deliver the final line without a trailing newline; the
				// error resurfaces on the next read.
				return string(line), nil
			}
			return "", err
		}
	}
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
//
// The read is started on demand: at most one line read is in flight per
// session, so a cancelled call neither consumes a line a later call needs nor
// leaves a second reader on the same stream. A read that is still waiting for
// input when ctx is cancelled is kept and resumed by the next call.
func (s *Session) ReadEvent(ctx context.Context) (backend.Event, error) {
	ch := s.pendingRead()

	select {
	case r := <-ch:
		s.clearPending(ch)
		if r.err != nil {
			return backend.Event{}, r.err
		}

		return backend.Event{
			Type: backend.EventKey,
			Key:  backend.KeyEnter,
			Text: strings.TrimSpace(r.text),
		}, nil
	case <-ctx.Done():
		return backend.Event{}, ctx.Err()
	}
}

// pendingRead returns the in-flight line read, starting one when idle.
func (s *Session) pendingRead() chan lineResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pending == nil {
		ch := make(chan lineResult, 1)
		s.pending = ch

		go func() {
			text, err := readLine(s.in)
			ch <- lineResult{text: text, err: err}
		}()
	}

	return s.pending
}

// clearPending marks the in-flight read as consumed, unless a newer one
// replaced it.
func (s *Session) clearPending(ch chan lineResult) {
	s.mu.Lock()
	if s.pending == ch {
		s.pending = nil
	}
	s.mu.Unlock()
}

// Size returns zero values because the plain backend does not depend on terminal size.
func (s *Session) Size() (width, height int) {
	return 0, 0
}

// Close closes the session. Plain sessions do not own IO streams.
func (s *Session) Close() error {
	return nil
}
