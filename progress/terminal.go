package progress

import (
	"io"
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether w is an interactive terminal.
func IsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	return term.IsTerminal(int(file.Fd()))
}
