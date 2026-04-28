package showcom

import (
	"bytes"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestBasePrintlnUsesConfiguredOutput(t *testing.T) {
	is := assert.New(t)

	out := new(bytes.Buffer)
	base := &Base{
		Out: out,
		FormatFn: func() {
		},
	}
	base.Buffer().WriteString("hello")

	base.Println()

	is.Eq("hello\n", out.String())
}
