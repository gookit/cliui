package title_test

import (
	"strings"
	"testing"

	"github.com/gookit/cliui/show/title"
	"github.com/gookit/color"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/assert"
)

// F10: content line width must equal the configured width for all alignments.
func TestTitle_ContentWidthMatchesWidth(t *testing.T) {
	cases := []struct {
		name string
		fns  []title.OptionFunc
	}{
		{"left", nil},
		{"center", []title.OptionFunc{title.WithAlignCenter()}},
		{"right", []title.OptionFunc{title.WithAlignRight()}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			is := assert.New(t)

			fns := []title.OptionFunc{title.WithWidth(30), title.WithBorderBottom()}
			fns = append(fns, c.fns...)

			s := color.ClearTag(title.New("Deploy", fns...).Render())
			lines := strings.Split(s, "\n")

			is.Eq(30, strutil.TextWidth(lines[0]))
			is.Eq(strutil.TextWidth(lines[0]), strutil.TextWidth(lines[1]))
		})
	}
}
