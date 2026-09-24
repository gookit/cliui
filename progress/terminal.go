package progress

import (
	"io"
	"os"

	"golang.org/x/term"
)

// FdWriter is implemented by writers that expose a file descriptor.
type FdWriter interface {
	Fd() uintptr
}

// IsTerminal reports whether w is an interactive terminal.
//
// It accepts *os.File as well as any writer exposing an Fd() method, so a
// wrapped writer (e.g. a color writer) is still recognized.
func IsTerminal(w io.Writer) bool {
	if w == nil {
		return false
	}

	var fd uintptr
	switch v := w.(type) {
	case *os.File:
		fd = v.Fd()
	case FdWriter:
		fd = v.Fd()
	default:
		return false
	}

	return term.IsTerminal(int(fd))
}
