package interact

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/gookit/cliui/cutypes"
)

type result struct {
	answer string
	err    error
}

// promptReaders keeps one reader goroutine per input stream. Reusing it means
// a line that arrives after a cancelled Prompt call is not lost, and reader
// goroutines do not accumulate across calls.
var promptReaders struct {
	sync.Mutex
	in io.Reader
	ch chan result
}

func sameReader(a, b io.Reader) (same bool) {
	defer func() { _ = recover() }() // guard against uncomparable readers
	return a == b
}

func promptLines(in io.Reader) chan result {
	promptReaders.Lock()
	defer promptReaders.Unlock()

	if promptReaders.ch != nil && sameReader(promptReaders.in, in) {
		return promptReaders.ch
	}

	ch := make(chan result, 1)
	promptReaders.in = in
	promptReaders.ch = ch

	go func(r io.Reader) {
		s := bufio.NewScanner(r)
		for s.Scan() {
			ch <- result{answer: strings.TrimSpace(s.Text())}
		}
		ch <- result{err: s.Err()}
	}(in)

	return ch
}

// Prompt query and read user answer.
//
// Usage:
//
//	answer, err := Prompt(context.Background(), "your name?", "")
//
// from package golang.org/x/tools/cmd/getgo
func Prompt(ctx context.Context, query, defaultAnswer string) (string, error) {
	_, _ = fmt.Fprintf(cutypes.Output, "%s [%s]: ", query, defaultAnswer)

	select {
	case r := <-promptLines(cutypes.Input):
		if r.err != nil {
			return "", r.err
		}

		if r.answer == "" {
			return defaultAnswer, nil
		}
		return r.answer, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
