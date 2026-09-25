package plain

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/color"
	"github.com/gookit/goutil/x/assert"
)

func TestSession_RenderColorTags(t *testing.T) {
	is := assert.New(t)
	color.ForceColor()
	defer color.RevertColorLevel()

	out := new(bytes.Buffer)
	sess, err := New().NewSession(strings.NewReader(""), out)
	is.Nil(err)

	err = sess.Render(backend.View{Lines: []string{"<green>Ready</>"}})
	is.Nil(err)
	is.Contains(out.String(), "\x1b[")
	is.NotContains(out.String(), "<green>")
}

// BCE-1: a session must not read past the requested line, otherwise the next
// prompt on the same stream only sees EOF.
func TestSession_ReadEventKeepsFollowingLines(t *testing.T) {
	is := assert.New(t)

	in := strings.NewReader("tom\njerry\n")
	out := new(bytes.Buffer)
	be := New()

	first, err := be.NewSession(in, out)
	is.NoErr(err)

	ev, err := first.ReadEvent(context.Background())
	is.NoErr(err)
	is.Eq(backend.KeyEnter, ev.Key)
	is.Eq("tom", ev.Text)

	second, err := be.NewSession(in, out)
	is.NoErr(err)

	ev, err = second.ReadEvent(context.Background())
	is.NoErr(err)
	is.Eq("jerry", ev.Text)
}

// BCE-1: the stream error is reported on every read after the input ends,
// instead of blocking a later caller forever.
func TestSession_ReadEventReportsEOFRepeatedly(t *testing.T) {
	is := assert.New(t)

	sess, err := New().NewSession(strings.NewReader("tom\n"), new(bytes.Buffer))
	is.NoErr(err)

	ev, err := sess.ReadEvent(context.Background())
	is.NoErr(err)
	is.Eq("tom", ev.Text)

	for i := range 2 {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err = sess.ReadEvent(ctx)
		cancel()

		is.True(errors.Is(err, io.EOF), "read %d: want io.EOF, got %v", i, err)
	}
}

// BCE-1: a read cancelled by ctx is resumed by the next call, so the line the
// caller was waiting for is neither dropped nor read twice.
func TestSession_CancelledReadKeepsPendingLine(t *testing.T) {
	is := assert.New(t)

	pr, pw := io.Pipe()
	defer func() { _ = pw.Close() }()

	sess, err := New().NewSession(pr, new(bytes.Buffer))
	is.NoErr(err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = sess.ReadEvent(ctx)
	is.True(errors.Is(err, context.Canceled), "want context.Canceled, got %v", err)

	if _, err = pw.Write([]byte("late\n")); err != nil {
		t.Fatalf("write input: %v", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ev, err := sess.ReadEvent(ctx)
	is.NoErr(err)
	is.Eq("late", ev.Text)
}
