package progress

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gookit/cliui"
	"github.com/gookit/goutil/x/assert"
)

type fdWriter struct{ fd uintptr }

func (w fdWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w fdWriter) Fd() uintptr                 { return w.fd }

// E1: IsTerminal must recognize any writer exposing an Fd() method.
func TestIsTerminalAcceptsFdWriter(t *testing.T) {
	is := assert.New(t)

	is.False(IsTerminal(nil))
	is.False(IsTerminal(new(bytes.Buffer)))
	is.Eq(IsTerminal(os.Stdout), IsTerminal(fdWriter{os.Stdout.Fd()}))
}

// D4: refreshing a shorter block must erase the previous rows.
func TestMultiProgressRefreshClearsShrunkBlock(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(10)
	p2 := mp.New(10)
	mp.Start() // renders a two-row block

	// shrink the visible set without going through Hide/Remove
	mp.mu.Lock()
	p2.hidden = true
	mp.mu.Unlock()

	buf.Reset()
	mp.Refresh()

	is.Eq(1, mp.VisibleLen())
	// 2 clears for the old rows + 1 for the redrawn row
	is.True(strings.Count(buf.String(), "\x1B[2K") >= 3)
	_ = p1
}

// The output stream must be resolved at render time, not at construction.
func TestProgressUsesOutputAtRenderTime(t *testing.T) {
	is := assert.New(t)

	p := Txt(10) // created before the output is replaced

	buf := new(bytes.Buffer)
	cliui.SetOutput(buf)
	defer cliui.ResetOutput()

	p.Start()
	p.AdvanceTo(5)
	p.Finish()

	is.Contains(buf.String(), "50.0%")
}

// F1: random theme helpers must be able to pick the last element.
func TestRandomThemeCoversLastElement(t *testing.T) {
	is := assert.New(t)

	lastChar := CharThemes[len(CharThemes)-1]
	lastChars := string(CharsThemes[len(CharsThemes)-1])
	lastBar := BarStyles[len(BarStyles)-1]

	var hitChar, hitChars, hitBar bool
	for i := 0; i < 5000; i++ {
		if RandomCharTheme() == lastChar {
			hitChar = true
		}
		if string(RandomCharsTheme()) == lastChars {
			hitChars = true
		}
		if RandomBarStyle() == lastBar {
			hitBar = true
		}
		if hitChar && hitChars && hitBar {
			break
		}
	}

	is.True(hitChar, "RandomCharTheme should be able to pick the last theme")
	is.True(hitChars, "RandomCharsTheme should be able to pick the last theme")
	is.True(hitBar, "RandomBarStyle should be able to pick the last style")
}

// F2: charNum greater than boxWidth must not panic (negative repeat).
func TestRoundTripWidgetCharNumGreaterThanBox(t *testing.T) {
	is := assert.New(t)

	w := RoundTripWidget('=', 20, 12)
	for i := 0; i < 30; i++ {
		is.True(w(nil) != "")
	}
}

// F3: Reset before Start must keep the bar startable.
func TestResetBeforeStart(t *testing.T) {
	is := assert.New(t)

	p := Txt(10)
	p.Out = new(bytes.Buffer)

	p.Reset(20)
	is.Eq(int64(20), p.Max())
	is.False(p.Started())

	p.Start() // must not panic with "already started"
	is.True(p.Started())
	p.Finish()
}

// F4: widget map is initialized on demand for zero-value instances.
func TestZeroValueProgressWidgetMap(t *testing.T) {
	is := assert.New(t)

	p := &Progress{}
	p.AddWidget("a", func(*Progress) string { return "a" })
	p.SetWidget("b", func(*Progress) string { return "b" })

	is.Len(p.Widgets, 2)
}

// F5: Destroy on a managed bar must not bypass the manager writer.
func TestDestroyManagedBar(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p := mp.New(10)
	mp.Start()

	p.Destroy() // no standalone output, no panic
	is.Eq(1, mp.Len())

	mp.Finish()
}

// F6: spinner can be stopped, restarted and stopped again.
func TestSpinnerRestart(t *testing.T) {
	is := assert.New(t)

	out := new(bytes.Buffer)
	s := LoadingSpinner([]rune{'-'}, time.Millisecond)
	s.Out = out

	s.Start("a %s")
	is.True(s.Active())
	time.Sleep(5 * time.Millisecond)

	s.Restart()
	is.True(s.Active())
	time.Sleep(5 * time.Millisecond)

	s.Stop("done")
	is.False(s.Active())

	// start again after a full stop
	s.Start("b %s")
	is.True(s.Active())
	s.Stop()
	is.False(s.Active())

	is.Contains(out.String(), "done")
}
