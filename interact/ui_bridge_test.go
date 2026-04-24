package interact

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/goutil/testutil/assert"
)

func TestNewUIInput(t *testing.T) {
	is := assert.New(t)

	be := NewUIPlainBackend()
	in := bytes.NewBufferString("tom\n")
	out := new(bytes.Buffer)

	ipt := NewUIInput("Your name")
	got, err := ipt.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.Eq("tom", got)
}

func TestNewUIReadlineBackendFallback(t *testing.T) {
	is := assert.New(t)

	be := NewUIReadlineBackend()
	in := bytes.NewBufferString("yes\n")
	out := new(bytes.Buffer)

	cfm := NewUIConfirm("Continue", false)
	got, err := cfm.RunWithIO(context.Background(), be, in, out)
	is.Nil(err)
	is.True(got)
}

func TestNewUIFakeBackend(t *testing.T) {
	is := assert.New(t)

	be := NewUIFakeBackend(
		backend.Event{Type: backend.EventKey, Key: backend.KeyDown},
		backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
	)

	sel := NewUISelect("Choose", []UIItem{
		{Key: "a", Label: "Alpha", Value: "alpha"},
		{Key: "b", Label: "Beta", Value: "beta"},
	})

	got, err := sel.Run(context.Background(), be)
	is.Nil(err)
	is.Eq("b", got.Key)
}

func TestUIErrorAlias(t *testing.T) {
	is := assert.New(t)
	is.True(errors.Is(ErrUIAborted, ErrUIAborted))
}
