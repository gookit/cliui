// Package cutypes provides common types, definitions for cliui
package cutypes

import (
	"io"
	"os"

	"github.com/gookit/color"
)

// the global input output stream
var (
	Input  io.Reader = os.Stdin
	Output io.Writer = os.Stdout
)

// SetInput stream
func SetInput(in io.Reader) { Input = in }

// SetOutput stream
func SetOutput(out io.Writer) {
	Output = out
	color.SetOutput(out)
}

// CustomIO stream
func CustomIO(in io.Reader, out io.Writer) {
	Input = in
	SetOutput(out)
}

// ResetInput stream
func ResetInput() { Input = os.Stdin }

// ResetOutput stream
func ResetOutput() {
	Output = os.Stdout
	color.ResetOutput()
}

// ResetIO stream
func ResetIO() {
	Input = os.Stdin
	ResetOutput()
}
