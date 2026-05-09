package progress

import (
	"bytes"
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
