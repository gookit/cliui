package title

import (
	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/goutil/comdef"
)

const DefaultWidth = 80

// Options title options
type Options struct {
	// Color 颜色Tag
	Color string
	// PaddingLR 是否左右填充 PaddingChar
	PaddingLR bool
	// PaddingChar 左右填充字符
	PaddingChar rune

	// 是否显示上下边框
	ShowBorder bool
	BorderChar rune
	// BorderPos 边框位置 0: 无, 1: 上, 2: 下, 4: 上下
	BorderPos cutypes.BorderPos

	// 总的显示宽度
	//  - 0 表示使用默认宽度
	Width int
	// PercentWidth 使用终端宽度的百分比宽度 (1-100)
	//  0 表示不使用百分比宽度
	PercentWidth int

	Indent int
	Align  comdef.Align
}

// OptionFunc definition
type OptionFunc func(t *Title)

// WithWidth 设置固定显示宽度
func WithWidth(width int) OptionFunc {
	return func(t *Title) {
		t.Width = width
		t.widthSet = true
	}
}

// WithPercentWidth 使用终端宽度的百分比宽度
func WithPercentWidth(percent int) OptionFunc {
	return func(t *Title) {
		t.PercentWidth = percent
	}
}

// WithBorderTop setting the title border to top
func WithBorderTop() OptionFunc {
	return func(t *Title) {
		t.ShowBorder = true
		t.BorderPos = cutypes.BorderPosTop
	}
}

// WithBorderBottom setting the title border to bottom
func WithBorderBottom() OptionFunc {
	return func(t *Title) {
		t.ShowBorder = true
		t.BorderPos = cutypes.BorderPosBottom
	}
}

// WithBorderBoth setting the title border to both top and bottom
func WithBorderBoth() OptionFunc {
	return func(t *Title) {
		t.ShowBorder = true
		t.BorderPos = cutypes.BorderPosTB
	}
}

// WithoutBorder setting the title border to none
func WithoutBorder() OptionFunc {
	return func(t *Title) {
		t.ShowBorder = false
	}
}

// WithAlignRight setting the title align to right
func WithAlignRight() OptionFunc {
	return func(t *Title) {
		t.Align = comdef.Right
	}
}

// WithAlignCenter setting the title align to center
func WithAlignCenter() OptionFunc {
	return func(t *Title) {
		t.Align = comdef.Center
	}
}
