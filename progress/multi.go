package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// MultiProgress manages multiple Progress instances and renders them as one block.
type MultiProgress struct {
	Overwrite       bool
	AutoRefresh     bool
	RefreshInterval time.Duration
	Writer          io.Writer

	mu        sync.Mutex
	bars      []*Progress
	started   bool
	finished  bool
	rendered  bool
	lastLines int
}

// NewMulti creates a new multi progress manager.
func NewMulti() *MultiProgress {
	return &MultiProgress{
		Overwrite: true,
		Writer:    os.Stdout,
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
func (mp *MultiProgress) New(maxSteps ...int) *Progress {
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
}

// Refresh re-renders all managed progress bars.
func (mp *MultiProgress) Refresh() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.refreshLocked()
}

// Finish renders the final state and ends the managed block.
func (mp *MultiProgress) Finish() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if mp.finished {
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
	return os.Stdout
}

func (mp *MultiProgress) update(fn func()) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	fn()
	mp.refreshLocked()
}

func (mp *MultiProgress) startProgress(p *Progress, maxSteps ...int) {
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

func (mp *MultiProgress) refreshLocked() {
	if !mp.started || mp.finished {
		return
	}

	lines := len(mp.bars)
	if lines == 0 {
		return
	}

	out := mp.writer()
	if mp.rendered {
		fmt.Fprint(out, "\r")
		if mp.lastLines > 1 {
			fmt.Fprintf(out, "\x1B[%dA", mp.lastLines-1)
		}
	}

	for i, p := range mp.bars {
		fmt.Fprint(out, "\x1B[2K")
		fmt.Fprint(out, p.Line())
		if i < lines-1 {
			fmt.Fprint(out, "\n")
		}
	}

	mp.rendered = true
	mp.lastLines = lines
}
