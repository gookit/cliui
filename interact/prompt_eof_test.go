package interact_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gookit/cliui"
	"github.com/gookit/cliui/interact"
	"github.com/gookit/goutil/x/assert"
)

func TestPromptReportsStreamEOF(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		answers     []string
	}{
		{name: "empty stream"},
		{name: "one line", input: "Ada\n", answers: []string{"Ada"}},
		{name: "final unterminated line", input: "Ada", answers: []string{"Ada"}},
		{name: "blank line uses default", input: "\n", answers: []string{"fallback"}},
		{name: "several lines", input: "Ada\nGrace\n", answers: []string{"Ada", "Grace"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cliui.CustomIO(strings.NewReader(tc.input), io.Discard)
			defer cliui.ResetIO()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			for _, want := range tc.answers {
				got, err := interact.Prompt(ctx, "Name", "fallback")
				assert.NoErr(t, err)
				assert.Eq(t, want, got)
			}
			for i := 0; i < 3; i++ {
				got, err := interact.Prompt(ctx, "Name", "fallback")
				if !errors.Is(err, io.EOF) || got != "" {
					t.Fatalf("Prompt after EOF #%d = %q,%v; want empty answer and EOF", i+1, got, err)
				}
			}
		})
	}
}
