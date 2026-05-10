package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/cliui/progress"
)

func main() {
	mode := "auto"
	if len(os.Args) > 1 {
		mode = strings.ToLower(os.Args[1])
	}

	mp := progress.NewMulti()
	mp.Writer = os.Stderr
	mp.RefreshInterval = 100 * time.Millisecond

	switch mode {
	case "dynamic":
		mp.RenderMode = progress.RenderDynamic
	case "plain":
		mp.RenderMode = progress.RenderPlain
	case "disabled":
		mp.RenderMode = progress.RenderDisabled
	case "auto":
		mp.UseAutoRenderMode()
	default:
		fmt.Fprintf(os.Stderr, "usage: go run ./examples/progress-render-mode-demo [auto|dynamic|plain|disabled]\n")
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "render mode: %s\n", mode)
	runSimulation(mp)
}

func runSimulation(mp *progress.MultiProgress) {
	build := mp.New(40)
	build.SetFormat("{@name:-8s} [{@bar}] {@percent:5s}% {@phase}")
	build.AddWidget("bar", progress.BarWidget(18, progress.BarStyles[0]))
	build.SetMessages(map[string]string{
		"name":  "build",
		"phase": "compile",
	})

	test := mp.New(30)
	test.SetFormat("{@name:-8s} [{@bar}] {@percent:5s}% {@phase}")
	test.AddWidget("bar", progress.BarWidget(18, progress.BarStyles[0]))
	test.SetMessages(map[string]string{
		"name":  "test",
		"phase": "queue",
	})

	mp.Start()
	defer mp.Finish()

	for i := 0; i < 40; i++ {
		time.Sleep(35 * time.Millisecond)
		build.Advance()
		if i == 12 {
			build.SetMessage("phase", "link")
		}
		if i < 30 {
			test.Advance()
		}
		if i == 16 {
			test.SetMessage("phase", "unit")
		}
	}

	build.Done("done")
	test.Done("done")
	mp.Println("log: simulated work finished")
}
