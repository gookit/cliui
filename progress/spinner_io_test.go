package progress

import (
	"bytes"
	"testing"
	"time"

	"github.com/gookit/cliui"
	"github.com/gookit/goutil/testutil/assert"
)

func TestSpinnerUsesCustomOutput(t *testing.T) {
	is := assert.New(t)

	out := new(bytes.Buffer)
	cliui.SetOutput(out)
	defer cliui.ResetOutput()

	sp := LoadingSpinner([]rune{'-'}, time.Millisecond)
	sp.Start("loading %s")
	time.Sleep(5 * time.Millisecond)
	sp.Stop("done")

	is.Contains(out.String(), "loading")
	is.Contains(out.String(), "done")
}

func TestSpinnerUsesInstanceOutput(t *testing.T) {
	is := assert.New(t)

	globalOut := new(bytes.Buffer)
	cliui.SetOutput(globalOut)
	defer cliui.ResetOutput()

	out := new(bytes.Buffer)
	sp := LoadingSpinner([]rune{'-'}, time.Millisecond)
	sp.Out = out
	sp.Start("loading %s")
	time.Sleep(5 * time.Millisecond)
	sp.Stop("done")

	is.Contains(out.String(), "loading")
	is.Contains(out.String(), "done")
	is.Eq("", globalOut.String())
}
