package interact_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gookit/cliui"
	"github.com/gookit/cliui/interact"
	"github.com/gookit/goutil/x/assert"
)

func TestReadLineReturnsCompletePhysicalLine(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"long ASCII", strings.Repeat("a", 12000) + "\nnext\n", strings.Repeat("a", 12000)},
		{"long Unicode", strings.Repeat("αβ.", 3000) + "\nnext\n", strings.Repeat("αβ.", 3000)},
		{"buffer boundary", strings.Repeat("b", 4096) + "\nnext\n", strings.Repeat("b", 4096)},
		{"CRLF and trim", "  " + strings.Repeat("c", 9000) + "  \r\nnext\n", strings.Repeat("c", 9000)},
		{"short line", " first \nnext\n", "first"},
		{"empty line", "\nnext\n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cliui.CustomIO(strings.NewReader(tc.input), io.Discard)
			defer cliui.ResetIO()
			got, err := interact.ReadLine("")
			assert.NoErr(t, err)
			assert.Eq(t, tc.want, got)
			got, err = interact.ReadLine("")
			assert.NoErr(t, err)
			assert.Eq(t, "next", got)
			_, err = interact.ReadLine("")
			if !errors.Is(err, io.EOF) {
				t.Fatalf("final error=%v,want EOF", err)
			}
		})
	}
	t.Run("unterminated final line", func(t *testing.T) {
		want := strings.Repeat("d", 12000)
		cliui.CustomIO(strings.NewReader(want), io.Discard)
		defer cliui.ResetIO()
		got, err := interact.ReadLine("")
		assert.NoErr(t, err)
		assert.Eq(t, want, got)
		_, err = interact.ReadLine("")
		if !errors.Is(err, io.EOF) {
			t.Fatalf("final error=%v,want EOF", err)
		}
	})
}

func TestReadLineExactBufferFinalLine(t *testing.T) {
	for _, tc := range []struct{ name, line string }{
		{"4096 ASCII", strings.Repeat("a", 4096)},
		{"8192 ASCII", strings.Repeat("b", 8192)},
		{"4096 UTF8", strings.Repeat("α", 2048)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cliui.CustomIO(strings.NewReader(tc.line), io.Discard)
			defer cliui.ResetIO()
			got, err := interact.ReadLine("")
			assert.NoErr(t, err)
			assert.Eq(t, tc.line, got)
			got, err = interact.ReadLine("")
			if got != "" || !errors.Is(err, io.EOF) {
				t.Fatalf("next read=%q,%v;want empty and EOF", got, err)
			}
		})
	}
}

func TestReadLinePreservesNonEOFErrorAfterFragments(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	failure := errors.New("input stream failed")
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.WriteString(writer, strings.Repeat("x", 4096))
		_ = writer.CloseWithError(failure)
	}()
	cliui.CustomIO(reader, io.Discard)
	defer cliui.ResetIO()
	got, err := interact.ReadLine("")
	<-done
	if len(got) != 4096 || !errors.Is(err, failure) {
		t.Fatalf("read bytes=%d,error=%v;want4096bytes and actual stream failure", len(got), err)
	}
}
