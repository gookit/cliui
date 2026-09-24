package internal

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gookit/cliui"
	"github.com/gookit/goutil/x/assert"
)

// E10: buffered input from one call must survive to the next call on the
// same stream.
func TestReadLineWithOutputKeepsBufferedInput(t *testing.T) {
	is := assert.New(t)

	cliui.SetInput(strings.NewReader("a\nb\n"))
	defer cliui.ResetInput()

	out := new(bytes.Buffer)

	first, err := ReadLineWithOutput("", out)
	is.NoErr(err)

	second, err := ReadLineWithOutput("", out)
	is.NoErr(err)

	is.Eq("a", first)
	is.Eq("b", second)
}
