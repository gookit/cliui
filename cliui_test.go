package cliui_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/gookit/cliui"
	"github.com/gookit/cliui/interact"
	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
	"github.com/gookit/goutil/testutil/assert"
)

func TestCustomIOAffectsInteractReadInput(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("tom\n")
	out := new(bytes.Buffer)
	cliui.CustomIO(in, out)
	defer cliui.ResetIO()

	got, err := interact.ReadInput("name:")
	is.Nil(err)
	is.Eq("tom", got)
	is.Contains(out.String(), "name:")
}

func TestCustomIOAffectsUIDefaultIO(t *testing.T) {
	is := assert.New(t)

	in := bytes.NewBufferString("tom\n")
	out := new(bytes.Buffer)
	cliui.CustomIO(in, out)
	defer cliui.ResetIO()

	got, err := ui.NewInput("name").Run(context.Background(), plain.New())
	is.Nil(err)
	is.Eq("tom", got)
	is.Contains(out.String(), "name:")
}
