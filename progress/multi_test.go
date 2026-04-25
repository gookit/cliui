package progress

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/goutil/testutil/assert"
)

func TestProgress_Line(t *testing.T) {
	is := assert.New(t)

	p := Txt(10)
	p.init()
	p.applyStep(3)

	line := p.Line()
	is.Contains(line, "30.0%(3/10)")
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
