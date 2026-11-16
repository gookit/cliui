package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gookit/gopty/ptytest"
)

var (
	fixtureOnce sync.Once
	fixtureBin  string
	fixtureErr  error
)

func buildFilteredSelect(t *testing.T) string {
	t.Helper()
	fixtureOnce.Do(func() {
		_, sourceFile, _, ok := runtime.Caller(0)
		if !ok {
			fixtureErr = fmt.Errorf("cannot locate e2e test")
			return
		}

		root := filepath.Dir(sourceFile)
		outDir := filepath.Join(root, "tmp", "e2e")
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			fixtureErr = err
			return
		}

		name := "filtered-select"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		fixtureBin = filepath.Join(outDir, name)

		cmd := exec.Command("go", "build", "-o", fixtureBin, "./cmd/filtered-select")
		cmd.Dir = root
		if output, err := cmd.CombinedOutput(); err != nil {
			fixtureErr = fmt.Errorf("go build filtered-select: %v\n%s", err, output)
		}
	})
	if fixtureErr != nil {
		t.Fatalf("build filtered-select fixture: %v", fixtureErr)
	}
	return fixtureBin
}

func TestFilteredSelectShrinksTerminalBlock(t *testing.T) {
	if testing.Short() {
		t.Skip("pty e2e cases are skipped with -short")
	}

	s := ptytest.Start(t, buildFilteredSelect(t))
	mark := s.Mark()
	if !s.Expect(`Select asset`, mark) {
		t.Fatalf("initial select view did not appear:\n%s", s.Text())
	}

	mark = s.Mark()
	s.Send("1")
	if !s.Expect(`Filter assets: 1`, mark, 20*time.Second) {
		t.Fatalf("filtered select view did not appear:\n%s", s.Since(mark))
	}

	s.Send(ptytest.Enter)
	if !s.Expect(`AFTER_SELECTION key=a`, mark, 20*time.Second) {
		t.Fatalf("selection result did not appear:\n%s", s.Since(mark))
	}

	code, err := s.WaitExit()
	if err != nil || code != 0 {
		t.Fatalf("exit = (%d, %v), output:\n%s", code, err, s.Text())
	}

	raw := s.RawText()
	resultAt := strings.Index(raw, "AFTER_SELECTION key=a")
	if resultAt < 0 {
		t.Fatalf("raw output has no selection result:\n%s", raw)
	}
	beforeResult := raw[:resultAt]
	if regexp.MustCompile(`\x1b\[[0-9]+B`).MatchString(beforeResult) {
		t.Fatalf("cursor moved down before the result, leaving a visual gap; raw output:\n%s", raw)
	}
}
