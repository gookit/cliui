package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gookit/cliui/progress"
)

type job struct {
	name  string
	size  int64
	fail  bool
	delay time.Duration
}

func main() {
	jobs := []job{
		{name: "fd", size: 80, delay: 18 * time.Millisecond},
		{name: "ripgrep", size: 120, delay: 12 * time.Millisecond},
		{name: "bat", size: 95, delay: 16 * time.Millisecond},
		{name: "delta", size: 70, fail: true, delay: 20 * time.Millisecond},
		{name: "eza", size: 105, delay: 14 * time.Millisecond},
	}

	mp := progress.NewMulti()
	mp.Writer = os.Stderr
	mp.UseAutoRenderMode()
	mp.AutoRefresh = true
	mp.RefreshInterval = 80 * time.Millisecond

	slots := make([]*progress.Progress, 3)
	for i := range slots {
		slot := mp.New()
		slot.SetFormat("{@slot} {@name:-10s} [{@bar}] {@percent:5s}% {@phase}")
		slot.AddWidget("bar", progress.BarWidget(24, progress.BarStyles[0]))
		slot.SetMessage("slot", fmt.Sprintf("#%d", i+1))
		slot.SetMessage("name", "idle")
		slot.SetMessage("phase", "waiting")
		slots[i] = slot
	}

	mp.Start()
	defer mp.Finish()

	workCh := make(chan job)
	var wg sync.WaitGroup
	for i, slot := range slots {
		wg.Add(1)
		go func(workerID int, bar *progress.Progress) {
			defer wg.Done()
			for item := range workCh {
				runJob(mp, bar, workerID, item)
			}
			bar.ResetWith(func(p *progress.Progress) {
				p.MaxSteps = 0
				p.Messages["slot"] = fmt.Sprintf("#%d", workerID+1)
				p.Messages["name"] = "idle"
				p.Messages["phase"] = "idle"
			})
		}(i, slot)
	}

	for _, item := range jobs {
		workCh <- item
	}
	close(workCh)
	wg.Wait()

	mp.Println("summary: simulated batch complete")
}

func runJob(mp *progress.MultiProgress, bar *progress.Progress, workerID int, item job) {
	bar.ResetWith(func(p *progress.Progress) {
		p.MaxSteps = item.size
		p.Messages["slot"] = fmt.Sprintf("#%d", workerID+1)
		p.Messages["name"] = item.name
		p.Messages["phase"] = "downloading"
	})

	step := int64(5)
	for bar.Step() < item.size {
		time.Sleep(item.delay)
		if item.fail && bar.Step() >= item.size/2 {
			bar.SetMessage("phase", "failed")
			bar.Fail("failed")
			mp.Printf("package %s failed: simulated checksum mismatch\n", item.name)
			return
		}
		bar.Advance(step)
	}

	bar.SetMessage("phase", "verifying")
	time.Sleep(120 * time.Millisecond)
	bar.SetMessage("phase", "done")
	bar.Done("done")
}
