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

func TestSelect_RunWithIO(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("b\n")
	out := new(bytes.Buffer)
	be := plain.New()

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})

	got, err := sel.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.Eq("b", got.Key)
	is.Eq("beta", got.Value)
	is.Contains(out.String(), "Choose")
}

func TestSelect_RunWithDefault(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("\n")
	out := new(bytes.Buffer)
	be := plain.New()

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})
	sel.DefaultKey = "a"

	got, err := sel.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.Eq("a", got.Key)
}

func TestMultiSelect_RunWithIO(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("a,c\n")
	out := new(bytes.Buffer)
	be := plain.New()

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
		{Key: "c", Label: "Gamma", Value: "gamma"},
	})

	got, err := sel.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.Len(got.Keys, 2)
	is.Eq("a", got.Key)
	is.Contains(out.String(), "comma separated")
}

func TestMultiSelect_RunWithDefault(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("\n")
	out := new(bytes.Buffer)
	be := plain.New()

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})
	sel.DefaultKeys = []string{"b"}

	got, err := sel.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.Len(got.Keys, 1)
	is.Eq("b", got.Key)
}

func TestMultiSelect_RunWithMinSelected(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("a\n" + "a,b\n")
	out := new(bytes.Buffer)
	be := plain.New()

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})
	sel.MinSelected = 2

	got, err := sel.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.Len(got.Keys, 2)
	is.Contains(out.String(), "select at least 2 option(s)")
}
