package progress

import (
	"bytes"
	"fmt"
	"io"
	"strings"
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

func TestProgressSupportsLargeStepCounts(t *testing.T) {
	is := assert.New(t)

	p := Txt()
	p.Out = new(bytes.Buffer)
	p.MaxSteps = int64(5 << 32)
	p.RedrawFreq = uint(1 << 16)

	p.Start()
	p.AdvanceTo(int64(4 << 32))

	is.Eq(int64(4<<32), p.Step())
	is.Eq(float32(0.8), p.Percent())
}

func TestProgressPadsCurrentByStepWidth(t *testing.T) {
	is := assert.New(t)

	p := Txt(1_000_000_000_000)
	p.Out = new(bytes.Buffer)
	p.Start()
	p.AdvanceTo(1)

	current := p.Handler("current")(p)

	is.Eq(strings.Repeat(" ", 12)+"1", current)
}

func TestProgressRespectsCustomStepWidth(t *testing.T) {
	is := assert.New(t)

	p := Txt(1_000_000_000_000)
	p.Out = new(bytes.Buffer)
	p.StepWidth = 4
	p.Start()
	p.AdvanceTo(1)

	current := p.Handler("current")(p)

	is.Eq("   1", current)
}

func TestProgressAutoWidthFollowsMaxStepsGrowth(t *testing.T) {
	is := assert.New(t)

	p := Txt(9)
	p.Out = new(bytes.Buffer)
	p.Start()
	p.AdvanceTo(1_000)

	is.Eq("1000", p.Handler("current")(p))
}

func TestProgressStartTreatsNegativeMaxStepsAsUnknown(t *testing.T) {
	is := assert.New(t)

	p := Txt(100)
	p.Out = new(bytes.Buffer)
	p.Start(-1)
	p.AdvanceTo(5)
	p.Finish()

	is.Eq(int64(5), p.MaxSteps)
	is.Eq(int64(5), p.Step())
}

func TestProgressSizeWidgets(t *testing.T) {
	is := assert.New(t)

	p := Txt(5 << 30)
	p.Out = new(bytes.Buffer)
	p.Start()
	p.AdvanceTo(1200346778)

	is.Eq("1.12G", p.Handler("curSize")(p))
	is.Eq("5.00G", p.Handler("maxSize")(p))
}

func TestProgressRemainingUsesFractionalRateForByteProgress(t *testing.T) {
	is := assert.New(t)

	p := Txt(10 << 20)
	p.Out = new(bytes.Buffer)
	p.Start()
	p.startedAt = time.Now().Add(-2 * time.Second)
	p.AdvanceTo(2 << 20)

	is.Eq("8 secs", p.Handler("remaining")(p))
}

func TestProgressTreatsNegativeMaxStepsAsUnknown(t *testing.T) {
	is := assert.New(t)

	p := Txt(-1)
	p.Out = new(bytes.Buffer)
	p.Start()

	p.AdvanceTo(5)
	p.Finish()

	is.Eq(int64(5), p.MaxSteps)
	is.Eq(int64(5), p.Step())
}

func TestProgressWriteAdvancesByByteCount(t *testing.T) {
	is := assert.New(t)

	p := Txt(10)
	p.Out = new(bytes.Buffer)
	p.Start()

	n, err := p.Write([]byte("hello"))

	is.NoErr(err)
	is.Eq(5, n)
	is.Eq(int64(5), p.Step())
}

func TestProgressWrapsReaderAndWriter(t *testing.T) {
	is := assert.New(t)

	readProgress := Txt(11)
	readProgress.Out = new(bytes.Buffer)
	readProgress.Start()

	data, err := io.ReadAll(readProgress.WrapReader(strings.NewReader("hello world")))
	is.NoErr(err)
	is.Eq("hello world", string(data))
	is.Eq(int64(11), readProgress.Step())

	writeProgress := Txt(11)
	writeProgress.Out = new(bytes.Buffer)
	writeProgress.Start()

	dst := new(bytes.Buffer)
	n, err := writeProgress.WrapWriter(dst).Write([]byte("hello world"))
	is.NoErr(err)
	is.Eq(11, n)
	is.Eq("hello world", dst.String())
	is.Eq(int64(11), writeProgress.Step())
}

func ExampleBar() {
	maxStep := 105
	p := CustomBar(60, BarStyles[0], int64(maxStep))
	p.MaxSteps = int64(maxStep)
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
	p := DynamicText(messages, int64(maxStep))

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
