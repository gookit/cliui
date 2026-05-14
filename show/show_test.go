package show_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/cliui/show"
	"github.com/gookit/cliui/show/lists"
	"github.com/gookit/color"
	"github.com/gookit/goutil/testutil/assert"
)

func TestList(t *testing.T) {
	// is := assert.New(t)
	l := show.NewList("test list", []string{
		"list item 0",
		"list item 1",
		"list item 2",
	})
	l.Println()

	l = show.NewList("test list1", map[string]string{
		"key0":     "list item 0",
		"the key1": "list item 1",
		"key2":     "list item 2",
		"key3":     "", // empty value
	})
	l.Opts.SepChar = " | "
	l.Println()
}

func TestList_mlevel(t *testing.T) {
	d := map[string]any{
		"key0":     "list item 0",
		"key2":     []string{"abc", "def"},
		"key4":     map[string]int{"abc": 23, "def": 45},
		"the key1": "list item 1",
		"key3":     "", // empty value
	}

	l := show.NewList("test list", d)
	l.Println()

	l = show.NewList("test list2", d).WithOptions(func(opts *lists.Options) {
		opts.SepChar = " | "
	})
	l.Println()
}

func TestLists(t *testing.T) {
	ls := show.NewLists(map[string]any{
		"test list": []string{
			"list item 0",
			"list item 1",
			"list item 2",
		},
		"test list1": map[string]string{
			"key0":     "list item 0",
			"the key1": "list item 1",
			"key2":     "list item 2",
			"key3":     "", // empty value
		},
	})
	ls.Opts.SepChar = " : "
	ls.Println()
}

func TestTabWriter(t *testing.T) {
	is := assert.New(t)
	ss := []string{
		"a\tb\taligned\t",
		"aa\tbb\taligned\t",
		"aaa\tbbb\tunaligned",
		"aaaa\tbbbb\taligned\t",
	}

	err := show.TabWriter(ss).Flush()
	is.NoErr(err)
}

func TestSome(t *testing.T) {
	fmt.Printf("|%8s|\n", "text")
	fmt.Printf("|%-8s|\n", "text")
	fmt.Printf("|%8s|\n", "text")
}

func TestAnyData(t *testing.T) {
	is := assert.New(t)

	cases := []struct {
		name string
		run  func()
	}{
		{
			name: "map",
			run: func() {
				show.AnyData("user", map[string]any{
					"name": "tom",
					"age":  18,
				})
			},
		},
		{
			name: "struct",
			run: func() {
				show.AnyData("user", struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{
					Name: "tom",
					Age:  18,
				})
			},
		},
		{
			name: "scalar",
			run: func() {
				show.AnyData("answer", 42)
			},
		},
		{
			name: "nil",
			run: func() {
				show.AnyData("empty", nil)
			},
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cutypes.SetOutput(buf)
			defer cutypes.ResetOutput()

			is.NotPanics(cs.run)
		})
	}
}

func TestPrettyJSON_String(t *testing.T) {
	is := assert.New(t)

	data := map[string]any{
		"ok":    true,
		"age":   int64(18),
		"count": uint(9),
		"score": 99.5,
		"name":  "tom",
	}

	pj := show.NewPrettyJSON(data)
	out := pj.String()

	is.Contains(out, `<info>"ok"</>`)
	is.Contains(out, `<success>true</>`)
	is.Contains(out, `<warning>18</>`)
	is.Contains(out, `<warning>9</>`)
	is.Contains(out, `<warning>99.5</>`)
	is.NotContains(out, `<warning>"tom"</>`)

	bs, err := json.MarshalIndent(data, "", "    ")
	is.NoErr(err)
	is.Eq(string(bs), color.ClearTag(out))
}

func TestJSONPrintsPrettyJSON(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	cutypes.SetOutput(buf)
	defer cutypes.ResetOutput()

	code := show.JSON(map[string]any{"ok": true})

	is.Eq(show.OK, code)
	is.Contains(buf.String(), "\x1b[")
	is.Contains(color.ClearCode(buf.String()), `"ok": true`)
}
