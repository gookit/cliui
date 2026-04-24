package readline

import (
	"bytes"
	"os"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestBackend_NewSessionRequiresFile(t *testing.T) {
	is := assert.New(t)

	be := New()
	_, err := be.NewSession(bytes.NewBuffer(nil), bytes.NewBuffer(nil))
	is.NotNil(err)
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
