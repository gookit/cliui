package cliui

import (
	"io"
	"os"

	"github.com/gookit/cliui/cutypes"
)

// SetOutput stream
func SetOutput(out io.Writer) { cutypes.Output = out }

// ResetOutput stream
func ResetOutput() { cutypes.Output = os.Stdout }
