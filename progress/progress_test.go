package progress

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/gookit/cliui"
	"github.com/gookit/goutil/testutil/assert"
)

func TestProgress_Display(t *testing.T) {
	is := assert.New(t)
	ss := widgetMatch.FindAllString(TxtFormat, -1)
	is.Len(ss, 4)

	is.Contains(ss, "{@message}")
}

func TestSpinner(t *testing.T) {
	chars := []rune(`你\|/`)
	str := `你\|/`

	fmt.Println(chars, string(chars[0]), string(str[0]))
}

func TestLoading(t *testing.T) {
	chars := []rune("◐◑◒◓")
	str := "◐◑◒◓"

	fmt.Println(chars, string(chars[0]), str, string(str[0]))
}

func TestProgressUsesCustomOutput(t *testing.T) {
	is := assert.New(t)

	out := new(bytes.Buffer)
	cliui.SetOutput(out)
	defer cliui.ResetOutput()

	p := Txt(2)
	p.Start()
	p.AdvanceTo(2)
	p.Finish()

	is.Contains(out.String(), "100")
}

func TestProgressUsesInstanceOutput(t *testing.T) {
	is := assert.New(t)

	globalOut := new(bytes.Buffer)
	cliui.SetOutput(globalOut)
	defer cliui.ResetOutput()

	out := new(bytes.Buffer)
	p := Txt(2)
	p.Out = out
	p.Start()
	p.AdvanceTo(2)
	p.Finish()

	is.Contains(out.String(), "100")
	is.Eq("", globalOut.String())
}

func ExampleBar() {
	maxStep := 105
	p := CustomBar(60, BarStyles[0], maxStep)
	p.MaxSteps = uint(maxStep)
	p.Format = FullBarFormat

	p.Start()
	for i := 0; i < maxStep; i++ {
		time.Sleep(80 * time.Millisecond)
		p.Advance()
	}
	p.Finish()
}

func ExampleDynamicText() {
	messages := map[int]string{
		// key is percent, range is 0 - 100.
		20:  " Prepare ...",
		40:  " Request ...",
		65:  " Transport ...",
		95:  " Saving ...",
		100: " Handle Complete.",
	}

	maxStep := 105
	p := DynamicText(messages, maxStep)

	p.Start()

	for i := 0; i < maxStep; i++ {
		time.Sleep(80 * time.Millisecond)
		p.Advance()
	}

	p.Finish()
}

func ExampleNewMulti() {
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	build := mp.New(3)
	build.AddMessage("message", " build")

	test := mp.New(2)
	test.AddMessage("message", " test")

	mp.Start()
	build.Advance()
	test.Advance()
	build.Finish()
	test.Finish()
	mp.Finish()
}
