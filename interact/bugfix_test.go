package interact

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gookit/cliui"
	"github.com/gookit/goutil/x/assert"
)

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
