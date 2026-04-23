// Package cutypes provides common types, definitions for cliui
package cutypes

import (
	"io"
	"os"
)

// Output the global input out stream
var Output io.Writer = os.Stdout

// SetOutput stream
func SetOutput(out io.Writer) { Output = out }

// ResetOutput stream
func ResetOutput() { Output = os.Stdout }
