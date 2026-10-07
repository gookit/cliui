package interact

import (
	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/color"
)

// ConfirmE a question, returns the answer and any read error.
func ConfirmE(message string, defVal ...bool) (bool, error) {
	color.Fprint(cutypes.Output, message)
	return AnswerIsYesE(defVal...)
}

// Confirm a question, returns bool.
// it is a shortcut of ConfirmE(), and returns false on read error.
func Confirm(message string, defVal ...bool) bool {
	ok, _ := ConfirmE(message, defVal...)
	return ok
}

// UnconfirmedE a question, returns whether the user did NOT confirm, and any read error.
func UnconfirmedE(message string, defVal ...bool) (bool, error) {
	ok, err := ConfirmE(message, defVal...)
	return !ok, err
}

// Unconfirmed a question, returns whether the user did NOT confirm.
// it is a shortcut of UnconfirmedE(), and returns true on read error.
func Unconfirmed(message string, defVal ...bool) bool {
	ok, _ := UnconfirmedE(message, defVal...)
	return ok
}

// AskE a question and return the result of the input, together with any read error.
//
// Usage:
//
//	answer, err := AskE("Your name?", "", nil)
//	answer, err := AskE("Your name?", "tom", nil)
//	answer, err := AskE("Your name?", "", nil, 3)
func AskE(question, defVal string, fn func(ans string) error, maxTimes ...int) (string, error) {
	q := &Question{Q: question, Func: fn, DefVal: defVal}
	if len(maxTimes) > 0 {
		q.MaxTimes = maxTimes[0]
	}

	v, err := q.RunE()
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

// Ask a question and return the result of the input.
// it is a shortcut of AskE(), and returns an empty string on error.
//
// Usage:
//
//	answer := Ask("Your name?", "tom", nil)
func Ask(question, defVal string, fn func(ans string) error, maxTimes ...int) string {
	answer, _ := AskE(question, defVal, fn, maxTimes...)
	return answer
}

// QueryE is alias of method AskE()
func QueryE(question, defVal string, fn func(ans string) error, maxTimes ...int) (string, error) {
	return AskE(question, defVal, fn, maxTimes...)
}

// Query is alias of method Ask()
func Query(question, defVal string, fn func(ans string) error, maxTimes ...int) string {
	return Ask(question, defVal, fn, maxTimes...)
}

// Choice is alias of method SelectOne()
func Choice(title string, options any, defOpt string, allowQuit ...bool) string {
	return SelectOne(title, options, defOpt, allowQuit...)
}

// ChoiceE is alias of method SelectOneE()
func ChoiceE(title string, options any, defOpt string, allowQuit ...bool) (string, error) {
	return SelectOneE(title, options, defOpt, allowQuit...)
}

// SingleSelect is alias of method SelectOne()
func SingleSelect(title string, options any, defOpt string, allowQuit ...bool) string {
	return SelectOne(title, options, defOpt, allowQuit...)
}

// SingleSelectE is alias of method SelectOneE()
func SingleSelectE(title string, options any, defOpt string, allowQuit ...bool) (string, error) {
	return SelectOneE(title, options, defOpt, allowQuit...)
}

// SelectOneE select one of the options, returns selected option value and any error.
//
// Map options:
//
//	{
//	   	// option key => option value
//	   	'a' => 'chengdu',
//	   	'b' => 'beijing'
//	}
//
// Array options:
//
//	{
//	   // only value, key will use index
//	   'chengdu',
//	   'beijing'
//	}
func SelectOneE(title string, options any, defOpt string, allowQuit ...bool) (string, error) {
	s := &Select{Title: title, Options: options, DefOpt: defOpt}

	if len(allowQuit) > 0 {
		s.DisableQuit = !allowQuit[0]
	}

	r, err := s.RunE()
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// SelectOne select one of the options, returns selected option value.
// it is a shortcut of SelectOneE(), and returns an empty string on error.
func SelectOne(title string, options any, defOpt string, allowQuit ...bool) string {
	val, _ := SelectOneE(title, options, defOpt, allowQuit...)
	return val
}

// SelectOneKeyE select one of the options, returns selected option key and any error.
func SelectOneKeyE(title string, options any, defOpt string, opFns ...func(*Select)) (string, error) {
	r, err := NewSelect(title, options, opFns...).With(func(s *Select) {
		s.DefOpt = defOpt
	}).RunE()
	if err != nil {
		return "", err
	}
	return r.KeyString(), nil
}

// SelectOneKey select one of the options, returns selected option key.
// it is a shortcut of SelectOneKeyE(), and returns an empty string on error.
func SelectOneKey(title string, options any, defOpt string, opFns ...func(*Select)) string {
	key, _ := SelectOneKeyE(title, options, defOpt, opFns...)
	return key
}

// Checkbox select multi of the options. is alias of method MultiSelect()
func Checkbox(title string, options any, defOpts []string, allowQuit ...bool) []string {
	return MultiSelect(title, options, defOpts, allowQuit...)
}

// CheckboxE is alias of method MultiSelectE()
func CheckboxE(title string, options any, defOpts []string, allowQuit ...bool) ([]string, error) {
	return MultiSelectE(title, options, defOpts, allowQuit...)
}

// MultiSelectE select multi of the options, returns selected option values and any error.
//
// like SingleSelectE(), but allow select multi option
func MultiSelectE(title string, options any, defOpts []string, allowQuit ...bool) ([]string, error) {
	s := &Select{Title: title, Options: options, DefOpts: defOpts, MultiSelect: true}

	if len(allowQuit) > 0 {
		s.DisableQuit = !allowQuit[0]
	}

	r, err := s.RunE()
	if err != nil {
		return nil, err
	}
	return r.Strings(), nil
}

// MultiSelect select multi of the options, returns selected option values.
// it is a shortcut of MultiSelectE(), and returns nil on error.
//
// like SingleSelect(), but allow select multi option
func MultiSelect(title string, options any, defOpts []string, allowQuit ...bool) []string {
	vals, _ := MultiSelectE(title, options, defOpts, allowQuit...)
	return vals
}
