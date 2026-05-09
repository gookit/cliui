package progress

import (
	"io"
	"sync"
	"time"
)

// DefaultByteTrackerInterval is used when a ByteTracker interval is not set.
const DefaultByteTrackerInterval = 100 * time.Millisecond

// ByteTracker aggregates byte progress updates and periodically flushes them
// to a Progress instance.
type ByteTracker struct {
	progress *Progress
	interval time.Duration

	mu      sync.Mutex
	pending int64
	closed  bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// NewByteTracker creates a byte tracker using DefaultByteTrackerInterval.
func NewByteTracker(p *Progress) *ByteTracker {
	return NewByteTrackerWithInterval(p, DefaultByteTrackerInterval)
}

// NewByteTrackerWithInterval creates a byte tracker with a custom flush interval.
func NewByteTrackerWithInterval(p *Progress, interval time.Duration) *ByteTracker {
	if interval <= 0 {
		interval = DefaultByteTrackerInterval
	}

	t := &ByteTracker{
		progress: p,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
	go t.loop()
	return t
}

// Add records a byte delta. Non-positive deltas are ignored.
func (t *ByteTracker) Add(n int64) {
	if t == nil || n <= 0 {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return
	}
	t.pending += n
}

// Close stops the tracker and flushes pending bytes.
func (t *ByteTracker) Close() {
	if t == nil {
		return
	}

	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	close(t.stopCh)
	t.mu.Unlock()

	<-t.doneCh
	t.flush()
}

func (t *ByteTracker) loop() {
	defer close(t.doneCh)

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.flush()
		case <-t.stopCh:
			return
		}
	}
}

func (t *ByteTracker) flush() {
	t.mu.Lock()
	n := t.pending
	t.pending = 0
	p := t.progress
	t.mu.Unlock()

	if n > 0 && p != nil {
		p.Advance(n)
	}
}

// NewConcurrentWriter creates a writer that tracks written byte counts.
func NewConcurrentWriter(p *Progress) io.Writer {
	return byteTrackerWriter{tracker: NewByteTracker(p)}
}

type byteTrackerWriter struct {
	tracker *ByteTracker
}

func (w byteTrackerWriter) Write(bs []byte) (int, error) {
	n := len(bs)
	if n > 0 {
		w.tracker.Add(int64(n))
	}
	return n, nil
}
