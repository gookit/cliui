package ui

import "errors"

var (
	// ErrAborted reports that the user aborted the current interaction.
	ErrAborted = errors.New("interact/ui: aborted")
	// ErrNoTTY reports that the backend requires a TTY but none is available.
	ErrNoTTY = errors.New("interact/ui: tty required")
	// ErrInvalidState reports invalid component configuration or state.
	ErrInvalidState = errors.New("interact/ui: invalid state")
	// ErrNotImplemented reports that a concrete backend loop is not implemented yet.
	ErrNotImplemented = errors.New("interact/ui: not implemented")
)
