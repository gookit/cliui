package internal

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/color"
)

// inputReaders keeps one bufio.Reader per input stream. Reusing it means a
// line already buffered by a previous call is not lost when the next call
// reads from the same stream.
var inputReaders struct {
	sync.Mutex
	in  io.Reader
	buf *bufio.Reader
}

func sameReader(a, b io.Reader) (same bool) {
	defer func() { _ = recover() }() // guard against uncomparable readers
	return a == b
}

func inputReader(in io.Reader) *bufio.Reader {
	inputReaders.Lock()
	defer inputReaders.Unlock()

	if inputReaders.buf == nil || !sameReader(inputReaders.in, in) {
		inputReaders.in = in
		inputReaders.buf = bufio.NewReader(in)
	}
	return inputReaders.buf
}

// ReadLineWithOutput read one line from user input. support gookit/color tag.
//
// Usage:
//
//	in := ReadLineWithOutput("")
//	ans, _ := ReadLineWithOutput("your name?")
func ReadLineWithOutput(question string, out io.Writer) (string, error) {
	if len(question) > 0 {
		if out == nil {
			out = cutypes.Output
		}
		color.Fprint(out, question)
	}

	answer, _, err := inputReader(cutypes.Input).ReadLine()
	return strings.TrimSpace(string(answer)), err
}
