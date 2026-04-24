// Package fake provides an in-memory backend for interaction tests.
package fake

import (
	"context"
	"io"
	"sync"

	"github.com/gookit/cliui/interact/backend"
)

// Backend is an in-memory backend for tests and event-flow verification.
type Backend struct {
	Events []backend.Event

	mu       sync.Mutex
	sessions []*Session
}

// New creates a fake backend with preset events.
func New(events ...backend.Event) *Backend {
	return &Backend{Events: events}
}

// NewSession creates a new fake session.
func (b *Backend) NewSession(_ io.Reader, _ io.Writer) (backend.Session, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := &Session{
		events: append([]backend.Event(nil), b.Events...),
	}
	b.sessions = append(b.sessions, s)
	return s, nil
}

// LastSession returns the last created session.
func (b *Backend) LastSession() *Session {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.sessions) == 0 {
		return nil
	}
	return b.sessions[len(b.sessions)-1]
}

// Session is an in-memory backend session.
type Session struct {
	mu     sync.Mutex
	events []backend.Event
	views  []backend.View
}

// Render stores the rendered view.
func (s *Session) Render(view backend.View) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.views = append(s.views, view)
	return nil
}

// ReadEvent returns the next queued event.
func (s *Session) ReadEvent(ctx context.Context) (backend.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-ctx.Done():
		return backend.Event{}, ctx.Err()
	default:
	}

	if len(s.events) == 0 {
		return backend.Event{Type: backend.EventInterrupt, Key: backend.KeyCtrlC}, nil
	}

	ev := s.events[0]
	s.events = s.events[1:]
	return ev, nil
}

// Size returns a fixed fake terminal size.
func (s *Session) Size() (width, height int) {
	return 80, 24
}

// Close closes the session.
func (s *Session) Close() error {
	return nil
}

// Views returns all rendered views.
func (s *Session) Views() []backend.View {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]backend.View(nil), s.views...)
}
