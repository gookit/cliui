package interact

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gookit/cliui"
	"github.com/gookit/goutil/x/assert"
)

// E9: a default value must pass the question validator.
func TestQuestionValidatesDefaultValue(t *testing.T) {
	is := assert.New(t)

	cliui.CustomIO(strings.NewReader("\ntom\n"), new(bytes.Buffer))
	defer cliui.ResetIO()

	q := NewQuestion("Name?", "x")
	q.Func = func(ans string) error {
		if ans == "x" {
			return errors.New("bad default")
		}
		return nil
	}

	got, err := q.Run()
	is.NoErr(err)
	is.Eq("tom", got.String())
}

// E12: StepsRun must stop when the context is cancelled.
func TestStepsRunStopsOnCancelledContext(t *testing.T) {
	is := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ran := 0
	s := &StepsRun{Steps: []StepHandler{
		func(context.Context) error { ran++; return nil },
	}}

	s.RunContext(ctx)
	is.Eq(0, ran)
	is.True(s.Err() != nil)
}

func TestStepsRunRunsAllSteps(t *testing.T) {
	is := assert.New(t)

	ran := 0
	s := &StepsRun{Steps: []StepHandler{
		func(context.Context) error { ran++; return nil },
		func(context.Context) error { ran++; return nil },
	}}

	s.Run()
	is.Eq(2, ran)
	is.Nil(s.Err())
}

func TestPromptReadsLine(t *testing.T) {
	is := assert.New(t)

	cliui.CustomIO(strings.NewReader("prod\n"), new(bytes.Buffer))
	defer cliui.ResetIO()

	got, err := Prompt(context.Background(), "Env", "dev")
	is.NoErr(err)
	is.Eq("prod", got)
}

func TestPromptUsesDefaultOnEmpty(t *testing.T) {
	is := assert.New(t)

	cliui.CustomIO(strings.NewReader("\n"), new(bytes.Buffer))
	defer cliui.ResetIO()

	got, err := Prompt(context.Background(), "Env", "dev")
	is.NoErr(err)
	is.Eq("dev", got)
}

// F12: prepare/render must not mutate the caller's map[string]string options.
func TestSelectDoesNotMutateCallerMap(t *testing.T) {
	is := assert.New(t)

	opts := map[string]string{"a": "chengdu", "b": "beijing"}
	s := NewSelect("Your city", opts)
	s.Out = new(bytes.Buffer)

	keys, err := s.prepare()
	is.NoErr(err)
	s.render(keys)

	_, hasQuit := opts["q"]
	is.False(hasQuit, "caller map must not receive the injected quit entry")
	is.Eq("quit", s.valMap["q"])
}
