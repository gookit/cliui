// Package interact collect some interactive methods for CLI
package interact

import (
	"errors"

	"github.com/gookit/goutil/structs"
)

const (
	// OK success exit code
	OK = 0
	// ERR error exit code
	ERR = 2
)

var (
	// ErrQuit is returned when the user chooses to quit an interaction.
	ErrQuit = errors.New("interact: user quit")
	// ErrMaxAttempts is returned when the max input attempts is exceeded.
	ErrMaxAttempts = errors.New("interact: max attempts exceeded")
)

// Value alias of structs.Value
type Value = structs.Value

/*************************************************************
 * value for select
 *************************************************************/

// SelectResult data store
type SelectResult struct {
	Value // V the select value(s)
	// K the select key(s)
	K Value
}

// create SelectResult create
func newSelectResult(key, val any) *SelectResult {
	return &SelectResult{
		K:     Value{V: key},
		Value: Value{V: val},
	}
}

// KeyString get
func (sv *SelectResult) KeyString() string {
	return sv.K.String()
}

// KeyStrings get
func (sv *SelectResult) KeyStrings() []string {
	return sv.K.Strings()
}

// Key value get
func (sv *SelectResult) Key() any {
	return sv.K.Val()
}

// WithKey value
func (sv *SelectResult) WithKey(key any) *SelectResult {
	sv.K.Set(key)
	return sv
}

