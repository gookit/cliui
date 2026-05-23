package lists

import (
	"testing"

	"github.com/gookit/color"
	"github.com/gookit/goutil/testutil/assert"
)

func TestList_WithStructTagName(t *testing.T) {
	type user struct {
		Name string `json:"json_name" yaml:"yaml_name"`
	}

	t.Run("default uses json tag", func(t *testing.T) {
		l := NewList("", user{Name: "tom"})
		out := color.ClearTag(l.String())

		assert.Contains(t, out, "json_name")
		assert.NotContains(t, out, "yaml_name")
	})

	t.Run("custom tag name", func(t *testing.T) {
		l := NewList("", user{Name: "tom"}, func(opts *Options) {
			opts.TagName = "yaml"
		})

		out := color.ClearTag(l.String())

		assert.Contains(t, out, "yaml_name")
		assert.NotContains(t, out, "json_name")
	})

	t.Run("new items with options", func(t *testing.T) {
		items := NewItemsWithOptions(user{Name: "tom"}, &Options{TagName: "yaml"})

		assert.Eq(t, "yaml_name", items.List[0].Key)
	})
}

func TestList_ArrayFieldIsEmpty(t *testing.T) {
	type data struct {
		Names [0]string `json:"names"`
	}

	l := NewList("", data{})

	assert.NotContains(t, l.String(), "names")
}
