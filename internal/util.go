package internal

import (
	"bufio"
	"io"
	"strings"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/color"
)

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

	reader := bufio.NewReader(cutypes.Input)
	answer, _, err := reader.ReadLine()
	return strings.TrimSpace(string(answer)), err
}
