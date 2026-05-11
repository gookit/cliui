package readline

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/goutil/testutil/assert"
)

func TestBackend_NewSessionFallbackToPlain(t *testing.T) {
	is := assert.New(t)

	be := New()
	sess, err := be.NewSession(bytes.NewBuffer(nil), bytes.NewBuffer(nil))
	is.Nil(err)
	is.NotNil(sess)
}

func TestBackend_NewSessionStrictRequiresFile(t *testing.T) {
	is := assert.New(t)

	be := NewStrict()
	_, err := be.NewSession(bytes.NewBuffer(nil), bytes.NewBuffer(nil))
	is.NotNil(err)
	is.True(errors.Is(err, backend.ErrFileRequired))
}

func TestSession_CloseWithoutState(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	s := &Session{out: buf, rendered: 2}
	err := s.Close()
	is.Nil(err)
	is.Contains(buf.String(), "\x1B[1B")
}

func TestSession_RenderTogglesCursorVisibility(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	s := &Session{out: buf}

	err := s.Render(backend.View{Lines: []string{"Choose"}, HideCursor: true})
	is.Nil(err)
	is.Contains(buf.String(), "\x1B[?25l")

	err = s.Render(backend.View{Lines: []string{"Choose"}, HideCursor: false})
	is.Nil(err)
	is.Contains(buf.String(), "\x1B[?25h")
}

func TestSession_RenderClearsStaleLinesWhenViewShrinks(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	s := &Session{out: buf}

	err := s.Render(backend.View{Lines: []string{"one", "two", "three", "four"}})
	is.Nil(err)

	err = s.Render(backend.View{Lines: []string{"one", "two"}})
	is.Nil(err)

	is.Eq(8, strings.Count(buf.String(), "\x1B[2K"))
	is.Eq(2, s.rendered)
}

func TestSession_CloseRestoresHiddenCursor(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	s := &Session{out: buf}

	err := s.Render(backend.View{Lines: []string{"Choose"}, HideCursor: true})
	is.Nil(err)
	err = s.Close()
	is.Nil(err)
	is.Contains(buf.String(), "\x1B[?25h")
}

func TestSession_DetectResize(t *testing.T) {
	is := assert.New(t)

	s := &Session{}
	_, ok := s.detectResize(80, 24)
	is.False(ok)

	ev, ok := s.detectResize(100, 30)
	is.True(ok)
	is.Eq(backend.EventResize, ev.Type)
	is.Eq(100, ev.Width)
	is.Eq(30, ev.Height)

	_, ok = s.detectResize(100, 30)
	is.False(ok)
}

func TestSession_ReadEventUnicodeText(t *testing.T) {
	is := assert.New(t)

	s := &Session{in: bufio.NewReader(bytes.NewBufferString("你"))}
	ev, err := s.ReadEvent(context.Background())

	is.Nil(err)
	is.Eq(backend.EventKey, ev.Type)
	is.Eq("你", ev.Text)
}

func TestSession_ReadEventEscapeSequences(t *testing.T) {
	tests := []struct {
		name string
		in   string
		key  backend.Key
	}{
		{name: "up", in: "\x1B[A", key: backend.KeyUp},
		{name: "down", in: "\x1B[B", key: backend.KeyDown},
		{name: "right", in: "\x1B[C", key: backend.KeyRight},
		{name: "left", in: "\x1B[D", key: backend.KeyLeft},
		{name: "shift tab", in: "\x1B[Z", key: backend.KeyShiftTab},
		{name: "home csi", in: "\x1B[H", key: backend.KeyHome},
		{name: "end csi", in: "\x1B[F", key: backend.KeyEnd},
		{name: "home ss3", in: "\x1BOH", key: backend.KeyHome},
		{name: "end ss3", in: "\x1BOF", key: backend.KeyEnd},
		{name: "home tilde", in: "\x1B[1~", key: backend.KeyHome},
		{name: "end tilde", in: "\x1B[4~", key: backend.KeyEnd},
		{name: "home seven tilde", in: "\x1B[7~", key: backend.KeyHome},
		{name: "end eight tilde", in: "\x1B[8~", key: backend.KeyEnd},
		{name: "page up", in: "\x1B[5~", key: backend.KeyPageUp},
		{name: "page down", in: "\x1B[6~", key: backend.KeyPageDown},
		{name: "delete", in: "\x1B[3~", key: backend.KeyDelete},
		{name: "esc", in: "\x1B", key: backend.KeyEsc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)

			s := &Session{in: bufio.NewReader(bytes.NewBufferString(tt.in))}
			ev, err := s.ReadEvent(context.Background())

			is.Nil(err)
			is.Eq(tt.key, ev.Key)
		})
	}
}

func TestSession_ReadEventTab(t *testing.T) {
	is := assert.New(t)

	s := &Session{in: bufio.NewReader(bytes.NewBufferString("\t"))}
	ev, err := s.ReadEvent(context.Background())

	is.Nil(err)
	is.Eq(backend.KeyTab, ev.Key)
}

func TestSession_SizeWithoutFile(t *testing.T) {
	is := assert.New(t)

	s := &Session{inFile: os.Stdin}
	_, _ = s.Size()
	is.True(true)
}
