package lists_test

import (
	"strings"
	"testing"

	"github.com/gookit/cliui/show/lists"
	"github.com/gookit/color"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/assert"
)

// D2: CJK keys must be padded by display width so values stay aligned.
func TestList_CJKKeyAlignment(t *testing.T) {
	is := assert.New(t)

	l := lists.NewList("", map[string]string{
		"name": "v1",
		"名称": "v2",
	})

	out := color.ClearTag(l.String())
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	is.Len(lines, 2)

	w1 := strutil.TextWidth(lines[0][:strings.Index(lines[0], "v1")])
	w2 := strutil.TextWidth(lines[1][:strings.Index(lines[1], "v2")])
	is.Eq(w1, w2)
}
