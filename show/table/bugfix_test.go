package table_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gookit/cliui/show/table"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/goutil/x/ccolor"
)

// F9: rendering the same table twice must be idempotent (no extra "#" column).
func TestTable_RenderTwiceIsStable(t *testing.T) {
	is := assert.New(t)

	tb := table.New("Stable",
		table.WithShowRowNumber(true),
		table.WithBorderFlags(table.BorderAll),
	)
	tb.SetHeads("Name", "Age").AddRow("Tom", 25).AddRow("Jerry", 30)

	first := tb.Render()
	second := tb.Render()

	is.Eq(first, second)
	is.Eq(tb.String(), tb.Render())

	// only one injected "#" head column
	is.Eq(1, strings.Count(ccolor.ClearCode(first), "#"))
}

// F8: multi-byte headers must stay valid UTF-8 and keep column alignment.
func TestTable_ChineseHeaderAlignment(t *testing.T) {
	is := assert.New(t)

	tb := table.New("", table.WithBorderFlags(table.BorderAll))
	tb.SetHeads("名称", "描述").AddRow("小明", "这是描述")

	out := ccolor.ClearCode(tb.String())
	is.True(utf8.ValidString(out))
	is.Contains(out, "名称")

	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	width := strutil.TextWidth(lines[0])
	for _, line := range lines {
		is.Eq(width, strutil.TextWidth(line))
	}
}

// OverflowWrap must actually wrap a too-long single line across rows.
func TestTable_OverflowWrapWrapsLongLine(t *testing.T) {
	is := assert.New(t)

	tb := table.New("", table.WithBorderFlags(table.BorderAll))
	tb.SetHeads("K").AddRow("abcdefghijklmnop")
	tb.WithOptions(
		table.WithOverflowFlag(table.OverflowWrap),
		table.WithColumnWidths(8),
	)

	out := ccolor.ClearCode(tb.String())
	is.Contains(out, "abcdefgh")
	is.Contains(out, "ijklmnop")
	is.NotContains(out, "abcdefghijklmnop")
}

// D1: []map[string]string rows go through reflection and must not panic.
func TestTable_SetRowsMapStringValues(t *testing.T) {
	is := assert.New(t)

	tb := table.New("", table.WithBorderFlags(table.BorderAll))
	tb.SetHeads("Name", "Role")
	tb.SetRows([]map[string]string{
		{"Name": "tom", "Role": "admin"},
		{"Name": "jane", "Role": "user"},
	})

	out := ccolor.ClearCode(tb.String())
	is.Contains(out, "tom")
	is.Contains(out, "admin")
	is.Contains(out, "jane")
}

// D3: mutating the table must invalidate the cached formatted output.
func TestTable_StringReflectsLaterMutation(t *testing.T) {
	is := assert.New(t)

	tb := table.New("", table.WithBorderFlags(table.BorderAll))
	tb.SetHeads("Name").AddRow("first")
	is.Contains(tb.String(), "first")

	tb.AddRow("second")
	second := tb.String()
	is.Contains(second, "second")
	is.Contains(second, "first")
}
