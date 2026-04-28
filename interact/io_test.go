package interact_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gookit/cliui"
	"github.com/gookit/cliui/interact"
	"github.com/gookit/goutil/testutil/assert"
)

func TestSelectUsesCustomOutput(t *testing.T) {
	is := assert.New(t)

	in := strings.NewReader("0\n")
	out := new(bytes.Buffer)
	cliui.CustomIO(in, out)
	defer cliui.ResetIO()

	got := interact.SelectOne("Choose", []string{"dev", "prod"}, "")

	is.Eq("dev", got)
	is.Contains(out.String(), "Choose")
	is.Contains(out.String(), "Your choice")
}

func TestSelectUsesInstanceOutput(t *testing.T) {
	is := assert.New(t)

	in := strings.NewReader("0\n")
	globalOut := new(bytes.Buffer)
	cliui.CustomIO(in, globalOut)
	defer cliui.ResetIO()

	out := new(bytes.Buffer)
	sel := interact.NewSelect("Choose", []string{"dev", "prod"})
	sel.Out = out
	got := sel.Run()

	is.Eq("dev", got.String())
	is.Contains(out.String(), "Choose")
	is.Contains(out.String(), "Your choice")
	is.Eq("", globalOut.String())
}

func TestQuestionUsesCustomOutput(t *testing.T) {
	is := assert.New(t)

	in := strings.NewReader("tom\n")
	out := new(bytes.Buffer)
	cliui.CustomIO(in, out)
	defer cliui.ResetIO()

	got := interact.NewQuestion("Your name?").Run()

	is.Eq("tom", got.String())
	is.Contains(out.String(), "Your name?")
	is.Contains(out.String(), "A: ")
}

func TestQuestionUsesInstanceOutput(t *testing.T) {
	is := assert.New(t)

	in := strings.NewReader("tom\n")
	globalOut := new(bytes.Buffer)
	cliui.CustomIO(in, globalOut)
	defer cliui.ResetIO()

	out := new(bytes.Buffer)
	q := interact.NewQuestion("Your name?")
	q.Out = out
	got := q.Run()

	is.Eq("tom", got.String())
	is.Contains(out.String(), "Your name?")
	is.Contains(out.String(), "A: ")
	is.Eq("", globalOut.String())
}
