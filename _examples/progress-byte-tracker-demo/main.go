package main

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gookit/cliui/progress"
)

func main() {
	runChunkWorkers()
	runConcurrentWriter()
}

func runChunkWorkers() {
	mp := progress.NewMulti()
	mp.Writer = os.Stderr
	mp.UseAutoRenderMode()
	mp.AutoRefresh = true
	mp.RefreshInterval = 80 * time.Millisecond

	bar := mp.New(240)
	bar.SetFormat("{@name:-14s} [{@bar}] {@percent:5s}% {@curSize}/{@maxSize} {@phase}")
	bar.AddWidget("bar", progress.BarWidget(24, progress.BarStyles[0]))
	bar.SetMessages(map[string]string{
		"name":  "chunked-file",
		"phase": "chunks",
	})

	mp.Start()
	tracker := progress.NewByteTrackerWithInterval(bar, 60*time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 12; j++ {
				time.Sleep(time.Duration(20+workerID*8) * time.Millisecond)
				tracker.Add(5)
			}
		}(i)
	}

	wg.Wait()
	tracker.Close()
	bar.Done("done")
	mp.Println("log: chunk workers finished")
	mp.Finish()
}

func runConcurrentWriter() {
	mp := progress.NewMulti()
	mp.Writer = os.Stderr
	mp.UseAutoRenderMode()
	mp.AutoRefresh = true
	mp.RefreshInterval = 80 * time.Millisecond

	const data = "abcdefghijklmnopqrstuvwxyz0123456789"
	payload := strings.NewReader(strings.Repeat(data, 8))

	bar := mp.New(int64(payload.Len()))
	bar.SetFormat("{@name:-14s} [{@bar}] {@percent:5s}% {@curSize}/{@maxSize} {@phase}")
	bar.AddWidget("bar", progress.BarWidget(24, progress.BarStyles[0]))
	bar.SetMessages(map[string]string{
		"name":  "writer-file",
		"phase": "copy",
	})

	mp.Start()
	writer := progress.NewConcurrentWriterWithInterval(bar, 50*time.Millisecond)
	_, err := io.Copy(writer, slowReader{reader: payload, delay: 15 * time.Millisecond})
	if closeErr := writer.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		bar.Fail("failed")
		mp.Printf("copy failed: %v\n", err)
	} else {
		bar.Done("done")
		mp.Println("log: writer copy finished")
	}
	mp.Finish()
}

type slowReader struct {
	reader io.Reader
	delay  time.Duration
}

func (r slowReader) Read(bs []byte) (int, error) {
	if len(bs) > 16 {
		bs = bs[:16]
	}
	time.Sleep(r.delay)
	return r.reader.Read(bs)
}
