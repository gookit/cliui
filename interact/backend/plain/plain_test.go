package plain

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gookit/cliui/interact/backend"
	"github.com/gookit/color"
	"github.com/gookit/goutil/testutil/assert"
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
