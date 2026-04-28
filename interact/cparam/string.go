package cparam

import (
	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/cliui/internal"
	"github.com/gookit/color"
)

// StringParam definition
type StringParam struct {
	InputParam
}

// NewStringParam instance
func NewStringParam(name, desc string) *StringParam {
	return &StringParam{
		InputParam: InputParam{
			typ:  TypeStringParam,
			name: name,
			desc: desc,
		},
	}
}

// Config param
func (p *StringParam) Config(fn func(p *StringParam)) *StringParam {
	fn(p)
	return p
}

// Run param and get user input
func (p *StringParam) Run() (err error) {
	var val string
	if p.runFn != nil {
		val, err = p.runFn()
		if err != nil {
			return err
		}

		return p.Set(val)
	}

	askMsg := color.WrapTag(p.desc+"? ", "yellow")
	val, err = internal.ReadLineWithOutput(askMsg, cutypes.Output)
	if err != nil {
		return err
	}
	return p.Set(val)
}
