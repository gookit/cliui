package fake

import (
	"context"
	"testing"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/goutil/testutil/assert"
)

func TestBackend_WithSize(t *testing.T) {
	is := assert.New(t)

	be := NewWithOptions(nil, WithSize(120, 40))
	session, err := be.NewSession(nil, nil)
	is.Nil(err)

	width, height := session.Size()
	is.Eq(120, width)
	is.Eq(40, height)
}

func TestSession_ReadEventResizeUpdatesSize(t *testing.T) {
	is := assert.New(t)

	be := NewWithOptions(
		[]backend.Event{{Type: backend.EventResize, Width: 90, Height: 20}},
		WithSize(120, 40),
	)
	session, err := be.NewSession(nil, nil)
	is.Nil(err)

	ev, err := session.ReadEvent(context.Background())
	is.Nil(err)
	is.Eq(backend.EventResize, ev.Type)

	width, height := session.Size()
	is.Eq(90, width)
	is.Eq(20, height)
}
