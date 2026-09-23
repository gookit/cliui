package emoji_test

import (
	"testing"

	"github.com/gookit/cliui/show/emoji"
	"github.com/gookit/goutil/x/assert"
)

// F7: ToUnicode must decode the first rune, not the first byte.
func TestToUnicode(t *testing.T) {
	is := assert.New(t)

	is.Eq("1f496", emoji.ToUnicode("💖"))
	is.Eq("1f496", emoji.ToUnicode("💖abc"))
	is.Eq("", emoji.ToUnicode(""))
	is.Eq("U+1f496", emoji.ToUnicode("💖", "U+"))
}

// F7: Render must match emoji names containing '-' and '+'.
func TestRenderEmojiNamesWithSymbols(t *testing.T) {
	is := assert.New(t)

	for _, name := range []string{":+1:", ":-1:", ":e-mail:", ":non-potable_water:"} {
		want := emoji.GetByName(name)
		is.True(name != want, name+" should map to an emoji")
		is.Eq(want, emoji.Render(name))
	}
}

func TestRenderKeepsUnknownName(t *testing.T) {
	is := assert.New(t)

	is.Eq(":not-a-real-emoji-xyz:", emoji.Render(":not-a-real-emoji-xyz:"))
}

// F14: unicode encode/decode round trip (regexes are now precompiled).
func TestUnicodeRoundTrip(t *testing.T) {
	is := assert.New(t)

	encoded := emoji.Encode("💖")
	is.Contains(encoded, "1f496")
	is.Eq("💖", emoji.Decode(encoded))
	is.Eq("💖", emoji.FromUnicode(encoded))
}
