package cliui

import (
	"io"

	"github.com/gookit/cliui/cutypes"
)

// SetInput stream
func SetInput(in io.Reader) { cutypes.SetInput(in) }

// SetOutput stream
func SetOutput(out io.Writer) { cutypes.SetOutput(out) }

// CustomIO stream
func CustomIO(in io.Reader, out io.Writer) { cutypes.CustomIO(in, out) }

// ResetInput stream
func ResetInput() { cutypes.ResetInput() }

// ResetOutput stream
func ResetOutput() { cutypes.ResetOutput() }

// ResetIO stream
func ResetIO() { cutypes.ResetIO() }
