package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/cliui/interact/backend/fake"
	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/goutil/x/assert"
)

// F11: space/Tab/resize events must not submit Input; space must be typed.
func TestInput_SpaceTabResizeDoNotSubmit(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "x"},
		backend.Event{Type: backend.EventKey, Key: backend.KeySpace},
		backend.Event{Type: backend.EventResize, Width: 100, Height: 40},
		backend.Event{Type: backend.EventKey, Key: backend.KeyTab},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Name")
	ipt.Default = "guest"

	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("x ", got)
}

// F11: space must not accept the Confirm default; a later y does.
func TestConfirm_SpaceDoesNotAccept(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Key: backend.KeySpace},
		backend.Event{Type: backend.EventResize, Width: 100, Height: 40},
		backend.Event{Type: backend.EventKey, Key: backend.KeyY},
	)

	cfm := NewConfirm("Continue", false)
	got, err := cfm.Run(context.Background(), be)
	is.Nil(err)
	is.True(got)
}

// BCE-1: with a non-TTY stream every prompt must receive its own line; a
// backend that reads ahead leaves the later prompts with EOF.
func TestInput_SequentialPromptsKeepTheirLines(t *testing.T) {
	is := assert.New(t)

	in := strings.NewReader("tom\njerry\n")
	out := new(bytes.Buffer)
	be := plain.New()

	first, err := NewInput("First").RunWithIO(context.Background(), be, in, out)
	is.NoErr(err)

	second, err := NewInput("Second").RunWithIO(context.Background(), be, in, out)
	is.NoErr(err)

	is.Eq("tom", first)
	is.Eq("jerry", second)
}
