// Package backend defines the runtime abstraction used by interact/ui.
package backend

import (
	"context"
	"io"
)

// Backend creates a terminal interaction session.
type Backend interface {
	NewSession(in io.Reader, out io.Writer) (Session, error)
}

// Session represents one terminal interaction session.
type Session interface {
	Render(view View) error
	ReadEvent(ctx context.Context) (Event, error)
	Size() (width, height int)
	Close() error
}

// EventType defines the type of an interaction event.
type EventType int

const (
	EventUnknown EventType = iota
	EventKey
	EventInterrupt
	EventResize
)

// Key defines a normalized key value.
type Key string

const (
	KeyUnknown   Key = ""
	KeyEnter     Key = "enter"
	KeyUp        Key = "up"
	KeyDown      Key = "down"
	KeyLeft      Key = "left"
	KeyRight     Key = "right"
	KeySpace     Key = "space"
	KeyEsc       Key = "esc"
	KeyBackspace Key = "backspace"
	KeyCtrlC     Key = "ctrl+c"
	KeyY         Key = "y"
	KeyN         Key = "n"
)

// Event is a normalized terminal input event.
type Event struct {
	Type EventType
	Key  Key
	Text string
}

// View is a minimal renderable terminal frame.
type View struct {
	Lines        []string
	CursorRow    int
	CursorColumn int
}
