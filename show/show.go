// Package show provides some formatter tools for display data.
package show

import (
	"fmt"
	"reflect"
	"text/tabwriter"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/cliui/show/banner"
	"github.com/gookit/cliui/show/lists"
	"github.com/gookit/cliui/show/showcom"
	"github.com/gookit/cliui/show/title"
)

const (
	// OK success exit code
	OK = 0
	// ERR error exit code
	ERR = 2
)

// var errInvalidType = errors.New("invalid input data type")

// FormatterFace interface. for compatible
type FormatterFace = showcom.Formatter

// ShownFace shown interface. for compatible
type ShownFace = showcom.ShownFace

// AnyData format and render any type data.
func AnyData(title string, v any) {
	if v == nil {
		JSON(v)
		return
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			JSON(nil)
			return
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
		AList(title, rv.Interface())
	default:
		JSON(v)
	}
}

// JSON print pretty JSON data
func JSON(v any, prefixAndIndent ...string) int {
	prefix := ""
	indent := "    "

	l := len(prefixAndIndent)
	if l > 0 {
		prefix = prefixAndIndent[0]
		if l > 1 {
			indent = prefixAndIndent[1]
		}
	}

	NewPrettyJSON(v, prefix, indent).Println()
	return OK
}

// ATitle create a Title instance and print. options see: TitleOption
func ATitle(titleText string, fns ...title.OptionFunc) {
	title.New(titleText).WithOptionFns(fns).Println()
}

// ListOption alias for lists.Options. for compatible
type ListOption = lists.Options
type ListOpFunc = lists.OptionFunc

// NewList create a List instance. options see: Options
func NewList(title string, data any, fns ...ListOpFunc) *lists.List {
	return lists.NewList(title, data).WithOptionFns(fns)
}

// NewLists create a Lists instance and print. options see: Options
func NewLists(listMap any, fns ...ListOpFunc) *lists.Lists {
	return lists.NewLists(listMap).WithOptionFns(fns)
}

// AList create a List instance and print. options see: Options
//
// Usage:
//
//	show.AList("some info", map[string]string{"name": "tom"})
func AList(title string, data any, fns ...ListOpFunc) {
	NewList(title, data).WithOptionFns(fns).Println()
}

// MList show multi list data. options see: Options
//
// Usage:
//
//	show.MList(data)
//	show.MList(data, func(opts *Options) {
//		opts.LeftIndent = "    "
//	})
func MList(listMap any, fns ...ListOpFunc) {
	NewLists(listMap).WithOptionFns(fns).Println()
}

// NewBanner create a Banner instance. options see: banner.Options
func NewBanner(content any, fns ...banner.OptionFunc) *banner.Banner {
	return banner.New(content, fns...)
}

// Banner create a Banner instance and print. options see: banner.Options
func Banner(content any, fns ...banner.OptionFunc) {
	banner.New(content, fns...).Println()
}

// TabWriter create. more please see: package text/tabwriter/example_test.go
//
// Usage:
//
//	w := TabWriter([]string{
//		"a\tb\tc\td\t.",
//		"123\t12345\t1234567\t123456789\t."
//	})
//	w.Flush()
func TabWriter(rows []string) *tabwriter.Writer {
	w := tabwriter.NewWriter(cutypes.Output, 0, 4, 2, ' ', tabwriter.Debug)

	for _, row := range rows {
		if _, err := fmt.Fprintln(w, row); err != nil {
			panic(err)
		}
	}

	return w
}
