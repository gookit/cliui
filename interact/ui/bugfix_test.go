package ui

import (
	"context"
	"testing"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/cliui/interact/backend/fake"
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
