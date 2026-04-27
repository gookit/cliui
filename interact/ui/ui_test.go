package ui

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/cliui/interact/backend/fake"
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

func TestInput_RunWithFakeEvents(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "t"},
		backend.Event{Type: backend.EventKey, Text: "o"},
		backend.Event{Type: backend.EventKey, Text: "m"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyBackspace},
		backend.Event{Type: backend.EventKey, Text: "n"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Your name")
	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("ton", got)

	session := be.LastSession()
	is.True(len(session.Views()) > 0)
}

func TestInput_KeepsValidationErrorUntilNextInput(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "x"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
		backend.Event{Type: backend.EventKey, Text: "y"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Name")
	ipt.Validate = func(v string) error {
		if v == "x" {
			return errors.New("too short")
		}
		return nil
	}

	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("xy", got)

	assertErrorDoesNotFlash(t, be.LastSession().Views(), "Error: too short", "Current: x")
}

func TestInput_RunWithFakeCursorMove(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Text: "c"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyLeft},
		backend.Event{Type: backend.EventKey, Text: "b"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Your name")
	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("abc", got)

	session := be.LastSession()
	views := session.Views()
	is.True(len(views) > 0)
	is.Eq(1, views[len(views)-1].CursorRow)
}

func TestInput_RunWithFakeUnicodeEditing(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "你"},
		backend.Event{Type: backend.EventKey, Text: "好"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyLeft},
		backend.Event{Type: backend.EventKey, Text: "很"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyBackspace},
		backend.Event{Type: backend.EventKey, Text: "真"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyDelete},
		backend.Event{Type: backend.EventKey, Text: "棒"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Edit")
	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("你真棒", got)

	session := be.LastSession()
	views := session.Views()
	is.True(len(views) > 0)
	is.Eq(len("Current: ")+6, views[len(views)-1].CursorColumn)
}

func TestInput_RunWithFakeHomeEndDelete(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Text: "b"},
		backend.Event{Type: backend.EventKey, Text: "d"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyLeft},
		backend.Event{Type: backend.EventKey, Text: "c"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyHome},
		backend.Event{Type: backend.EventKey, Text: "X"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyDelete},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnd},
		backend.Event{Type: backend.EventKey, Text: "Z"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Edit")
	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("XbcdZ", got)
}

func TestInput_RunWithFakeCtrlAAndCtrlE(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "b"},
		backend.Event{Type: backend.EventKey, Text: "c"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlA},
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlE},
		backend.Event{Type: backend.EventKey, Text: "d"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Edit")
	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("abcd", got)
}

func TestInput_RunWithFakeCtrlUAndCtrlK(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Text: "b"},
		backend.Event{Type: backend.EventKey, Text: "c"},
		backend.Event{Type: backend.EventKey, Text: "d"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyLeft},
		backend.Event{Type: backend.EventKey, Key: backend.KeyLeft},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlK},
		backend.Event{Type: backend.EventKey, Text: "X"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlE},
		backend.Event{Type: backend.EventKey, Text: "Y"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlA},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlU},
		backend.Event{Type: backend.EventKey, Text: "Z"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Edit")
	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("ZabXY", got)
}

func TestInput_RunWithFakeCtrlW(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Text: "l"},
		backend.Event{Type: backend.EventKey, Text: "p"},
		backend.Event{Type: backend.EventKey, Text: "h"},
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Text: " "},
		backend.Event{Type: backend.EventKey, Text: "b"},
		backend.Event{Type: backend.EventKey, Text: "e"},
		backend.Event{Type: backend.EventKey, Text: "t"},
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlW},
		backend.Event{Type: backend.EventKey, Text: "g"},
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Text: "m"},
		backend.Event{Type: backend.EventKey, Text: "m"},
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	ipt := NewInput("Edit")
	got, err := ipt.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("alpha gamma", got)
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

func TestConfirm_RunWithFakeEvents(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Key: backend.KeyRight},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	cfm := NewConfirm("Continue", true)
	got, err := cfm.Run(context.Background(), be)
	is.Nil(err)
	is.False(got)

	session := be.LastSession()
	is.True(len(session.Views()) > 0)
}

func TestConfirm_KeepsInvalidInputErrorUntilNextInput(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "maybe"},
		backend.Event{Type: backend.EventKey, Text: "y"},
	)

	cfm := NewConfirm("Continue", false)
	got, err := cfm.Run(context.Background(), be)
	is.Nil(err)
	is.True(got)

	assertErrorDoesNotFlash(t, be.LastSession().Views(), "Error: please input yes or no", "Current: no")
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

func TestSelect_RunWithFakeEvents(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Key: backend.KeyDown},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("b", got.Key)

	session := be.LastSession()
	views := session.Views()
	is.True(len(views) > 0)
	last := views[len(views)-1]
	is.Contains(last.Lines[len(last.Lines)-3], "Current: Beta")
	is.Contains(last.Lines[len(last.Lines)-2], "Use Up/Down")
}

func TestSelect_RunWithFakePagingKeys(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Key: backend.KeyPageDown},
		backend.Event{Type: backend.EventKey, Key: backend.KeyShiftTab},
		backend.Event{Type: backend.EventKey, Key: backend.KeyTab},
		backend.Event{Type: backend.EventKey, Key: backend.KeyPageUp},
		backend.Event{Type: backend.EventKey, Key: backend.KeyTab},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
		{Key: "c", Label: "Gamma", Value: "gamma"},
	})

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("b", got.Key)
}

func TestSelect_KeepsUnknownOptionErrorUntilNextInput(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "x"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyDown},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("b", got.Key)

	assertErrorDoesNotFlash(t, be.LastSession().Views(), "Error: unknown option", "Current: Alpha (a)")
}

func TestSelect_RunWithFilter(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "b"},
		backend.Event{Type: backend.EventKey, Text: "e"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
		{Key: "g", Label: "Gamma", Value: "gamma"},
	})
	sel.Filterable = true

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("b", got.Key)

	last := be.LastSession().Views()[len(be.LastSession().Views())-1]
	is.True(viewContainsLine(last, "Filter: be"))
	is.True(viewContainsLine(last, "> b) Beta"))
}

func TestSelect_FilterNoMatchCanRecover(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "z"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
		backend.Event{Type: backend.EventKey, Key: backend.KeyBackspace},
		backend.Event{Type: backend.EventKey, Key: backend.KeyDown},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})
	sel.Filterable = true

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("b", got.Key)

	views := be.LastSession().Views()
	is.True(anyViewContainsLine(views, "No matches"))
	is.True(anyViewContainsLine(views, "Error: no matched option"))
}

func TestSelect_FilteredDisabledItemCannotSubmit(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "b"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlU},
		backend.Event{Type: backend.EventKey, Key: backend.KeyUp},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta", Disabled: true},
	})
	sel.Filterable = true

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("a", got.Key)
	is.True(anyViewContainsLine(be.LastSession().Views(), "Error: selected option is disabled"))
}

func TestSelect_FilterResizeKeepsQuery(t *testing.T) {
	is := assert.New(t)

	be := fake.NewWithOptions(
		[]backend.Event{
			{Type: backend.EventKey, Text: "a"},
			{Type: backend.EventResize, Width: 80, Height: 6},
			{Type: backend.EventKey, Key: backend.KeyEnter},
		},
		fake.WithSize(80, 12),
	)

	sel := NewSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})
	sel.Filterable = true

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("a", got.Key)

	views := be.LastSession().Views()
	is.True(anyViewContainsLine(views, "Filter: a"))
	is.Eq(6, views[len(views)-1].Height)
}

func TestMultiSelect_RunWithFakeEvents(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Key: backend.KeySpace},
		backend.Event{Type: backend.EventKey, Key: backend.KeyDown},
		backend.Event{Type: backend.EventKey, Key: backend.KeySpace},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
		{Key: "c", Label: "Gamma", Value: "gamma"},
	})

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Len(got.Keys, 2)
	is.Eq("a", got.Key)

	session := be.LastSession()
	views := session.Views()
	is.True(len(views) > 0)
	last := views[len(views)-1]
	is.Contains(last.Lines[len(last.Lines)-4], "Current: Beta")
	is.Contains(last.Lines[len(last.Lines)-3], "Selected(2): a, b")
	is.Contains(last.Lines[len(last.Lines)-2], "Space to toggle")
}

func TestMultiSelect_RunWithFakePagingKeys(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Key: backend.KeyPageDown},
		backend.Event{Type: backend.EventKey, Key: backend.KeySpace},
		backend.Event{Type: backend.EventKey, Key: backend.KeyPageUp},
		backend.Event{Type: backend.EventKey, Key: backend.KeyTab},
		backend.Event{Type: backend.EventKey, Key: backend.KeyShiftTab},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
		{Key: "c", Label: "Gamma", Value: "gamma"},
	})

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Len(got.Keys, 1)
	is.Eq("c", got.Key)
}

func TestMultiSelect_KeepsMinSelectedErrorUntilNextInput(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
		backend.Event{Type: backend.EventKey, Key: backend.KeySpace},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Len(got.Keys, 1)
	is.Eq("a", got.Key)

	assertErrorDoesNotFlash(t, be.LastSession().Views(), "Error: at least one option is required", "Selected(0): none")
}

func TestMultiSelect_RunWithFilter(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "g"},
		backend.Event{Type: backend.EventKey, Text: "a"},
		backend.Event{Type: backend.EventKey, Key: backend.KeySpace},
		backend.Event{Type: backend.EventKey, Key: backend.KeyCtrlU},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
		{Key: "g", Label: "Gamma", Value: "gamma"},
	})
	sel.Filterable = true

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Len(got.Keys, 1)
	is.Eq("g", got.Key)
	is.True(anyViewContainsLine(be.LastSession().Views(), "Selected(1): g"))
}

func TestMultiSelect_FilterNoMatchShowsError(t *testing.T) {
	is := assert.New(t)

	be := fake.New(
		backend.Event{Type: backend.EventKey, Text: "z"},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
		backend.Event{Type: backend.EventKey, Key: backend.KeyBackspace},
		backend.Event{Type: backend.EventKey, Key: backend.KeySpace},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})
	sel.Filterable = true

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Len(got.Keys, 1)
	is.Eq("a", got.Key)
	is.True(anyViewContainsLine(be.LastSession().Views(), "Error: no matched option"))
}

func TestMultiSelect_FilterResizeKeepsSelected(t *testing.T) {
	is := assert.New(t)

	be := fake.NewWithOptions(
		[]backend.Event{
			{Type: backend.EventKey, Text: "b"},
			{Type: backend.EventKey, Key: backend.KeySpace},
			{Type: backend.EventResize, Width: 80, Height: 7},
			{Type: backend.EventKey, Key: backend.KeyEnter},
		},
		fake.WithSize(80, 12),
	)

	sel := NewMultiSelect("Choose", []Item{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})
	sel.Filterable = true

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Len(got.Keys, 1)
	is.Eq("b", got.Key)

	views := be.LastSession().Views()
	is.True(anyViewContainsLine(views, "Filter: b"))
	is.True(anyViewContainsLine(views, "Selected(1): b"))
	is.Eq(7, views[len(views)-1].Height)
}

func assertErrorDoesNotFlash(t *testing.T, views []backend.View, errLine, stateLine string) {
	t.Helper()

	for i, view := range views {
		if !viewContainsLine(view, errLine) {
			continue
		}

		if i+1 < len(views) && viewContainsLine(views[i+1], stateLine) && !viewContainsLine(views[i+1], errLine) {
			t.Fatalf("error line %q was cleared before state changed", errLine)
		}
		return
	}

	t.Fatalf("missing error line %q", errLine)
}

func anyViewContainsLine(views []backend.View, line string) bool {
	for _, view := range views {
		if viewContainsLine(view, line) {
			return true
		}
	}
	return false
}

func viewContainsLine(view backend.View, line string) bool {
	for _, got := range view.Lines {
		if got == line {
			return true
		}
	}
	return false
}
