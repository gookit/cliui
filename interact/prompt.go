package interact

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/gookit/cliui/cutypes"
)

type result struct {
	answer string
	err    error
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

	ch := make(chan result, 1)
	go func() {
		s := bufio.NewScanner(cutypes.Input)
		if !s.Scan() { // reading
			ch <- result{"", s.Err()}
			return
		}

		answer := strings.TrimSpace(s.Text())
		if answer == "" {
			answer = defaultAnswer
		}
		ch <- result{answer, nil}
	}()

	select {
	case r := <-ch:
		return r.answer, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
