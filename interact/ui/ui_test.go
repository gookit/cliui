package ui

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/goutil/testutil/assert"
)

func TestValidateItems(t *testing.T) {
	is := assert.New(t)

	err := validateItems([]Item{{Key: "1", Label: "one"}, {Key: "2", Label: "two"}})
	is.Nil(err)
}

func TestValidateItems_DuplicateKey(t *testing.T) {
	is := assert.New(t)

	err := validateItems([]Item{{Key: "1", Label: "one"}, {Key: "1", Label: "again"}})
	is.True(errors.Is(err, ErrInvalidState))
}

func TestInput_ValidateConfig(t *testing.T) {
	is := assert.New(t)

	in := NewInput("Your name")
	is.Nil(in.ValidateConfig())

	in.Prompt = ""
	err := in.ValidateConfig()
	is.True(errors.Is(err, ErrInvalidState))
}

func TestSelect_ValidateConfig(t *testing.T) {
	is := assert.New(t)

	sel := NewSelect("Choose", []Item{{Key: "a", Label: "A"}})
	is.Nil(sel.ValidateConfig())

	sel.Items = nil
	err := sel.ValidateConfig()
	is.True(errors.Is(err, ErrInvalidState))
}

func TestMultiSelect_ValidateConfig(t *testing.T) {
	is := assert.New(t)

	sel := NewMultiSelect("Choose", []Item{{Key: "a", Label: "A"}})
	sel.MinSelected = -1

	err := sel.ValidateConfig()
	is.True(errors.Is(err, ErrInvalidState))
}

func TestInput_RunWithIO(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("tom\n")
	out := new(bytes.Buffer)
	be := plain.New()

	ipt := NewInput("Your name")
	got, err := ipt.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.Eq("tom", got)
	is.Contains(out.String(), "Your name:")
}

func TestInput_RunWithDefault(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("\n")
	out := new(bytes.Buffer)
	be := plain.New()

	ipt := NewInput("Your name")
	ipt.Default = "guest"
	got, err := ipt.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.Eq("guest", got)
}

func TestConfirm_RunWithIO(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("yes\n")
	out := new(bytes.Buffer)
	be := plain.New()

	cfm := NewConfirm("Continue", false)
	got, err := cfm.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.True(got)
	is.Contains(out.String(), "Continue [yes/no]")
}

func TestConfirm_RunWithDefault(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("\n")
	out := new(bytes.Buffer)
	be := plain.New()

	cfm := NewConfirm("Continue", true)
	got, err := cfm.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.True(got)
}
