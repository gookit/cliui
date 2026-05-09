package progress

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gookit/cliui/cutypes"
)

// DefaultRefreshInterval is used when MultiProgress auto refresh is enabled
// and RefreshInterval is not set.
const DefaultRefreshInterval = 100 * time.Millisecond

// RenderMode controls how MultiProgress writes progress output.
type RenderMode int

const (
	// RenderDynamic renders a managed ANSI block.
	RenderDynamic RenderMode = iota
	// RenderPlain renders plain progress lines without ANSI block controls.
	RenderPlain
	// RenderDisabled disables progress rendering.
	RenderDisabled
)

// MultiProgress manages multiple Progress instances and renders them as one block.
type MultiProgress struct {
	Overwrite       bool
	AutoRefresh     bool
	RefreshInterval time.Duration
	RenderMode      RenderMode
	Writer          io.Writer

	mu        sync.Mutex
	bars      []*Progress
	started   bool
	finished  bool
	rendered  bool
	dirty     bool
	stopCh    chan struct{}
	doneCh    chan struct{}
	lastLines int
}

// NewMulti creates a new multi progress manager.
func NewMulti() *MultiProgress {
	return &MultiProgress{
		Overwrite: true,
		Writer:    cutypes.Output,
	}
}

// Add attaches an existing progress bar to the manager.
func (mp *MultiProgress) Add(p *Progress) *Progress {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	p.manager = mp
	p.index = len(mp.bars)
	mp.bars = append(mp.bars, p)
	return p
}

// New creates a managed progress bar.
func (mp *MultiProgress) New(maxSteps ...int64) *Progress {
	return mp.Add(New(maxSteps...))
}

// Start initializes all registered bars and renders the managed block.
func (mp *MultiProgress) Start() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if mp.started {
		panic("MultiProgress already started")
	}

	for _, p := range mp.bars {
		if !p.started {
			p.init()
		}
	}

	mp.started = true
	mp.refreshLocked()
	if mp.AutoRefresh {
		mp.startAutoRefreshLocked()
	}
}

// Refresh re-renders all managed progress bars.
func (mp *MultiProgress) Refresh() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.dirty = false
	mp.refreshLocked()
}

// RunExclusive clears the managed progress block, lets fn write to the
// manager writer, then redraws the block.
func (mp *MultiProgress) RunExclusive(fn func(w io.Writer)) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.clearLocked()
	if fn != nil {
		fn(mp.writer())
	}
	if mp.started && !mp.finished {
		mp.refreshLocked()
	}
}

// Println writes a line without breaking the managed progress block.
func (mp *MultiProgress) Println(args ...any) {
	mp.RunExclusive(func(w io.Writer) {
		fmt.Fprintln(w, args...)
	})
}

// Printf writes formatted text without breaking the managed progress block.
func (mp *MultiProgress) Printf(format string, args ...any) {
	mp.RunExclusive(func(w io.Writer) {
		fmt.Fprintf(w, format, args...)
	})
}

// Finish renders the final state and ends the managed block.
func (mp *MultiProgress) Finish() {
	mp.mu.Lock()
	if mp.finished {
		mp.mu.Unlock()
		return
	}

	if !mp.started {
		for _, p := range mp.bars {
			if !p.started {
				p.init()
			}
		}
		mp.started = true
	}

	stopCh := mp.stopCh
	doneCh := mp.doneCh
	mp.stopCh = nil
	mp.doneCh = nil
	mp.mu.Unlock()

	if stopCh != nil {
		close(stopCh)
		if doneCh != nil {
			<-doneCh
		}
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()
	if mp.finished {
		return
	}

	mp.dirty = false
	mp.refreshLocked()
	if len(mp.bars) > 0 {
		fmt.Fprintln(mp.writer())
	}
	mp.finished = true
}

func (mp *MultiProgress) writer() io.Writer {
	if mp.Writer != nil {
		return mp.Writer
	}
	return cutypes.Output
}

func (mp *MultiProgress) update(fn func() bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	changed := fn()
	if !changed {
		return
	}

	if mp.AutoRefresh {
		mp.dirty = true
		return
	}

	mp.refreshLocked()
}

func (mp *MultiProgress) startProgress(p *Progress, maxSteps ...int64) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if p.started {
		panic("Progress bar already started")
	}

	p.init(maxSteps...)
	if mp.started && !mp.finished {
		mp.refreshLocked()
	}
}

func (mp *MultiProgress) startAutoRefreshLocked() {
	if mp.stopCh != nil {
		return
	}

	interval := mp.RefreshInterval
	if interval <= 0 {
		interval = DefaultRefreshInterval
	}

	mp.stopCh = make(chan struct{})
	mp.doneCh = make(chan struct{})
	stopCh := mp.stopCh
	doneCh := mp.doneCh

	go func() {
		defer close(doneCh)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mp.flushDirty()
			case <-stopCh:
				return
			}
		}
	}()
}

func (mp *MultiProgress) flushDirty() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !mp.dirty {
		return
	}

	mp.dirty = false
	mp.refreshLocked()
}

func (mp *MultiProgress) clearLocked() {
	if !mp.rendered || mp.lastLines == 0 {
		return
	}

	mp.clearRenderedBlockLocked()
}

func (mp *MultiProgress) moveToBlockStartLocked() {
	out := mp.writer()
	fmt.Fprint(out, "\r")
	if mp.lastLines > 1 {
		fmt.Fprintf(out, "\x1B[%dA", mp.lastLines-1)
	}
}

func (mp *MultiProgress) clearRenderedBlockLocked() {
	if !mp.rendered || mp.lastLines == 0 {
		return
	}

	out := mp.writer()
	mp.moveToBlockStartLocked()

	for i := 0; i < mp.lastLines; i++ {
		fmt.Fprint(out, "\x1B[2K")
		if i < mp.lastLines-1 {
			fmt.Fprint(out, "\n")
		}
	}

	mp.moveToBlockStartLocked()

	mp.rendered = false
	mp.lastLines = 0
}

// Started reports whether the multi progress manager has started.
func (mp *MultiProgress) Started() bool {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	return mp.started
}

// Finished reports whether the multi progress manager has finished.
func (mp *MultiProgress) Finished() bool {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	return mp.finished
}

// Len returns the number of managed progress bars.
func (mp *MultiProgress) Len() int {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	return len(mp.bars)
}

// VisibleLen returns the number of visible managed progress bars.
func (mp *MultiProgress) VisibleLen() int {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	return len(mp.bars)
}

func (mp *MultiProgress) refreshLocked() {
	switch mp.RenderMode {
	default:
		mp.refreshDynamicLocked()
	}
}

func (mp *MultiProgress) refreshDynamicLocked() {
	if !mp.started || mp.finished {
		return
	}

	lines := mp.currentLinesLocked()
	if len(lines) == 0 {
		return
	}

	if mp.rendered {
		mp.moveToBlockStartLocked()
	}

	mp.renderDynamicLinesLocked(lines)
}

func (mp *MultiProgress) currentLinesLocked() []string {
	lines := make([]string, 0, len(mp.bars))
	for _, p := range mp.bars {
		lines = append(lines, p.Line())
	}
	return lines
}

func (mp *MultiProgress) renderDynamicLinesLocked(lines []string) {
	out := mp.writer()
	for i, line := range lines {
		fmt.Fprint(out, "\x1B[2K")
		fmt.Fprint(out, line)
		if i < len(lines)-1 {
			fmt.Fprint(out, "\n")
		}
	}

	mp.rendered = true
	mp.lastLines = len(lines)
}
