package interact

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/cliui/internal"
	"github.com/gookit/color"
)

// Question definition
type Question struct {
	// Out output writer. default is cutypes.Output
	Out io.Writer
	// Q the question message
	Q string
	// Func validate user input answer is right.
	// if not set, will only check the answer is empty.
	Func func(ans string) error
	// DefVal default value
	DefVal string
	// MaxTimes maximum allowed number of errors, 0 is don't limited
	MaxTimes int
	errTimes int
}

// NewQuestion instance.
//
// Usage:
//
//	q := NewQuestion("Please input your name?")
//	ans := q.Run().String()
func NewQuestion(q string, defVal ...string) *Question {
	if len(defVal) > 0 {
		return &Question{Out: cutypes.Output, Q: q, DefVal: defVal[0]}
	}
	return &Question{Out: cutypes.Output, Q: q}
}

// Run and returns value
func (q *Question) Run() (*Value, error) {
	if err := q.render(); err != nil {
		return nil, err
	}

DoASK:
	ans, err := internal.ReadLineWithOutput("A: ", q.out())
	if err != nil {
		return nil, fmt.Errorf("(interact.Question) %w", err)
	}

	// don't input
	if ans == "" {
		if q.DefVal != "" { // has default value
			// the default value must pass the validator too
			if q.Func != nil {
				if err := q.Func(q.DefVal); err != nil {
					if err := q.checkErrTimes(); err != nil {
						return nil, err
					}
					fmt.Fprintln(q.out(), color.Error.Render(err.Error()))
					goto DoASK
				}
			}
			return &Value{V: q.DefVal}, nil
		}

		if err := q.checkErrTimes(); err != nil {
			return nil, err
		}
		fmt.Fprintln(q.out(), color.Error.Render("A value is required."))
		goto DoASK
	}

	// has validator func
	if q.Func != nil {
		if err := q.Func(ans); err != nil {
			if err := q.checkErrTimes(); err != nil {
				return nil, err
			}
			fmt.Fprintln(q.out(), color.Error.Render(err.Error()))
			goto DoASK
		}
	}

	return &Value{V: ans}, nil
}

func (q *Question) render() error {
	q.Q = strings.TrimSpace(q.Q)
	if q.Q == "" {
		return errors.New("(interact.Question) must provide question message")
	}

	var defMsg string

	q.DefVal = strings.TrimSpace(q.DefVal)
	if q.DefVal != "" {
		defMsg = fmt.Sprintf(" [default:%s]", color.Green.Render(q.DefVal))
	}

	// print question
	fmt.Fprintf(q.out(), "%s%s\n", color.Comment.Render(q.Q), defMsg)
	return nil
}

func (q *Question) checkErrTimes() error {
	if q.MaxTimes <= 0 {
		return nil
	}

	// limit error times
	if q.MaxTimes == q.errTimes {
		return fmt.Errorf("%w: entered incorrectly %d times", ErrMaxAttempts, q.MaxTimes)
	}

	q.errTimes++
	return nil
}

func (q *Question) out() io.Writer {
	if q.Out != nil {
		return q.Out
	}
	return cutypes.Output
}
