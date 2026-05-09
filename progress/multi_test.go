package progress

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/goutil/testutil/assert"
)

func TestProgress_Line(t *testing.T) {
	is := assert.New(t)

	p := Txt(10)
	p.init()
	p.applyStep(3)

	line := p.Line()
	is.Contains(line, "30.0%( 3/10)")
}

func TestMultiProgress_Render(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(4)
	p1.AddMessage("message", " first")
	p2 := mp.New(2)
	p2.AddMessage("message", " second")

	mp.Start()
	p1.AdvanceTo(2)
	p2.Finish()
	mp.Finish()

	out := buf.String()
	is.Contains(out, " first")
	is.Contains(out, " second")
	is.Contains(out, "\x1B[2K")
	is.True(strings.HasSuffix(out, "\n"))
}

func TestMultiProgressRenderDynamicKeepsANSIBlock(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderDynamic

	p := mp.New(2)
	p.AddMessage("message", " task")
	mp.Start()
	p.Advance()
	mp.Finish()

	out := buf.String()
	is.Contains(out, "\x1B[2K")
	is.Contains(out, " task")
}

func TestMultiProgress_WriterFallbackUsesGlobalOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	cutypes.SetOutput(buf)
	defer cutypes.ResetOutput()

	mp := &MultiProgress{}
	if mp.writer() != buf {
		t.Fatalf("expected fallback writer to use cutypes.Output")
	}
}

func TestMultiProgress_ConcurrentAdvance(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(5)
	p1.AddMessage("message", " task-1")
	p2 := mp.New(5)
	p2.AddMessage("message", " task-2")

	mp.Start()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			p1.Advance()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			p2.Advance()
		}
	}()

	wg.Wait()
	mp.Finish()

	out := buf.String()
	is.Contains(out, "task-1")
	is.Contains(out, "task-2")
	is.Contains(out, "100.0%(5/5)")
}

func TestMultiProgressConcurrentAdvanceSameBar(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.AutoRefresh = true
	mp.RefreshInterval = time.Hour

	p := mp.New(1000)
	mp.Start()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				p.Advance()
			}
		}()
	}

	wg.Wait()
	mp.Finish()

	is.Eq(int64(1000), p.Step())
}

func TestMultiProgressRedrawFreqInManagedMode(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p := mp.New(10)
	p.RedrawFreq = 3
	mp.Start()

	initial := strings.Count(buf.String(), "\x1B[2K")
	p.Advance()
	p.Advance()
	is.Eq(initial, strings.Count(buf.String(), "\x1B[2K"))

	p.Advance()
	is.True(strings.Count(buf.String(), "\x1B[2K") > initial)
}

func TestMultiProgressAutoRefreshMarksDirty(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.AutoRefresh = true
	mp.RefreshInterval = 20 * time.Millisecond

	p := mp.New(10)
	mp.Start()
	initial := strings.Count(buf.String(), "\x1B[2K")
	p.AdvanceTo(5)
	is.Eq(initial, strings.Count(buf.String(), "\x1B[2K"))

	time.Sleep(60 * time.Millisecond)
	is.Contains(buf.String(), "50.0%")
	mp.Finish()
}

func TestMultiProgressFinishStopsAutoRefresh(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.AutoRefresh = true
	mp.RefreshInterval = 10 * time.Millisecond

	p := mp.New(10)
	mp.Start()
	p.AdvanceTo(3)
	mp.Finish()

	size := buf.Len()
	time.Sleep(40 * time.Millisecond)
	is.Eq(size, buf.Len())
}

func TestMultiProgressRenderPlainDoesNotUseANSIBlock(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(3)
	p.AddMessage("message", " task")
	mp.Start()
	p.Advance()
	mp.Refresh()
	mp.Finish()

	out := buf.String()
	is.NotContains(out, "\x1B[2K")
	is.NotContains(out, "\x1B[")
	is.Contains(out, " task")
}

func TestMultiProgressRenderPlainAdvanceDoesNotPrint(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(10)
	mp.Start()
	size := buf.Len()
	p.Advance()
	p.Advance()

	is.Eq(size, buf.Len())
	mp.Refresh()
	is.True(buf.Len() > size)
}

func TestMultiProgressRenderPlainResetPrintsKeyState(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(10)
	p.SetFormat("{@name} {@percent}%")
	p.SetMessage("name", "fd")
	mp.Start()
	size := buf.Len()

	p.Reset(20)
	p.SetMessage("name", "bat")

	out := buf.String()[size:]
	is.Contains(out, "fd 0.0%")
	is.NotContains(out, "\x1B[")
}

func TestMultiProgressRenderPlainSetMessageDoesNotPrint(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(10)
	p.SetFormat("{@name} {@percent}%")
	p.SetMessage("name", "fd")
	mp.Start()
	size := buf.Len()

	p.SetMessage("name", "bat")
	p.SetFormat("{@name} {@current}/{@max}")

	is.Eq(size, buf.Len())
	mp.Refresh()
	is.Contains(buf.String()[size:], "bat  0/10")
}

func TestMultiProgressRenderPlainProgressFinishPrintsOnce(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(2)
	p.SetFormat("{@message} {@percent}%")
	mp.Start()
	p.Finish("done")
	afterProgressFinish := strings.Count(buf.String(), "done 100.0%")
	mp.Finish()
	afterManagerFinish := strings.Count(buf.String(), "done 100.0%")

	is.Eq(1, afterProgressFinish)
	is.Eq(1, afterManagerFinish)
}

func TestMultiProgressRenderDisabledSuppressesProgress(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderDisabled

	p := mp.New(3)
	p.AddMessage("message", " hidden")
	mp.Start()
	p.Advance()
	mp.Refresh()
	p.Finish()
	mp.Finish()

	is.Eq("", buf.String())
}

func TestMultiProgressRenderDisabledStillPrintsLogs(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderDisabled
	mp.New(1)
	mp.Start()

	mp.Println("warning: fallback")
	mp.Printf("package %s failed\n", "fd")
	mp.Finish()

	out := buf.String()
	is.Contains(out, "warning: fallback")
	is.Contains(out, "package fd failed")
	is.NotContains(out, "\x1B[")
}

func TestIsTerminalReturnsFalseForBuffer(t *testing.T) {
	is := assert.New(t)
	is.False(IsTerminal(new(bytes.Buffer)))
}

func TestIsTerminalReturnsFalseForRegularFile(t *testing.T) {
	is := assert.New(t)
	file, err := os.CreateTemp("", "cliui-progress-terminal-*")
	is.NoErr(err)
	defer os.Remove(file.Name())
	defer file.Close()

	is.False(IsTerminal(file))
}

func TestMultiProgressHideAndShow(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(1)
	p1.SetMessage("message", " one")
	p2 := mp.New(1)
	p2.SetMessage("message", " two")
	mp.Start()

	mp.Hide(p2)
	is.Eq(2, mp.Len())
	is.Eq(1, mp.VisibleLen())
	buf.Reset()
	mp.Refresh()
	is.Contains(buf.String(), " one")
	is.NotContains(buf.String(), " two")

	mp.Show(p2)
	is.Eq(2, mp.Len())
	is.Eq(2, mp.VisibleLen())
	buf.Reset()
	mp.Refresh()
	is.Contains(buf.String(), " two")
}

func TestMultiProgressRemove(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(1)
	p1.SetMessage("message", " one")
	p2 := mp.New(1)
	p2.SetMessage("message", " two")
	mp.Start()

	mp.Remove(p2)
	is.Eq(1, mp.Len())
	is.Eq(1, mp.VisibleLen())
	p2.Advance()
	p2.SetMessage("message", " removed")
	p2.Finish()
	buf.Reset()
	mp.Refresh()

	out := buf.String()
	is.Contains(out, " one")
	is.NotContains(out, " removed")
}

func TestMultiProgressRemoveBeforeStartMakesUpdatesNoop(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p := mp.New(3)
	mp.Remove(p)

	is.Eq(0, mp.Len())
	is.Eq(0, mp.VisibleLen())
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("removed progress update should be no-op, got panic: %v", r)
			}
		}()

		p.Start()
		p.Advance()
		p.AdvanceTo(2)
		p.Reset(5)
		p.SetMessage("message", " removed")
		p.Finish("done")
		p.Done()
		p.Fail()
		p.Skip()
		p.Display()
	}()
	is.Eq("", buf.String())
}

func TestMultiProgressHideClearsStaleDynamicLines(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(1)
	p1.SetMessage("message", " one")
	p2 := mp.New(1)
	p2.SetMessage("message", " two")
	mp.Start()

	buf.Reset()
	mp.Hide(p2)
	out := buf.String()

	is.Contains(out, "\x1B[2K")
	is.Contains(out, " one")
	is.NotContains(out, " two")
}

func TestMultiProgressRunExclusivePrintsBetweenBlocks(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p := mp.New(10)
	p.AddMessage("message", " task")
	mp.Start()
	p.AdvanceTo(5)
	mp.Println("warning: fallback")
	p.AdvanceTo(10)
	mp.Finish()

	out := buf.String()
	is.Contains(out, "warning: fallback")
	is.Contains(out, "100.0%(10/10)")
	is.True(strings.Count(out, "\x1B[2K") >= 3)
}

func TestMultiProgressPrintf(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.New(1)
	mp.Start()

	mp.Printf("package %s failed: %s\n", "fd", "network")
	mp.Finish()

	is.Contains(buf.String(), "package fd failed: network")
}

func TestMultiProgressStateGetters(t *testing.T) {
	is := assert.New(t)
	mp := NewMulti()
	mp.Writer = new(bytes.Buffer)
	mp.New(1)

	is.False(mp.Started())
	is.False(mp.Finished())
	is.Eq(1, mp.Len())
	is.Eq(1, mp.VisibleLen())

	mp.Start()
	is.True(mp.Started())
	mp.Finish()
	is.True(mp.Finished())
}
