package readline

import (
	"bytes"
	"errors"
	"os"
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

func TestSession_SizeWithoutFile(t *testing.T) {
	is := assert.New(t)

	s := &Session{inFile: os.Stdin}
	_, _ = s.Size()
	is.True(true)
}
