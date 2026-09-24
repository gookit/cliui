package interact

import (
	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/color"
)

// Confirm a question, returns bool
func Confirm(message string, defVal ...bool) (bool, error) {
	color.Fprint(cutypes.Output, message)
	return AnswerIsYes(defVal...)
}

// Unconfirmed a question, returns bool
func Unconfirmed(message string, defVal ...bool) (bool, error) {
	ok, err := Confirm(message, defVal...)
	return !ok, err
}

// Ask a question and return the result of the input.
//
// Usage:
//
//	answer, err := Ask("Your name?", "", nil)
//	answer, err := Ask("Your name?", "tom", nil)
//	answer, err := Ask("Your name?", "", nil, 3)
func Ask(question, defVal string, fn func(ans string) error, maxTimes ...int) (string, error) {
	q := &Question{Q: question, Func: fn, DefVal: defVal}
	if len(maxTimes) > 0 {
		q.MaxTimes = maxTimes[0]
	}

	v, err := q.Run()
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

// Query is alias of method Ask()
func Query(question, defVal string, fn func(ans string) error, maxTimes ...int) (string, error) {
	return Ask(question, defVal, fn, maxTimes...)
}

// Choice is alias of method SelectOne()
func Choice(title string, options any, defOpt string, allowQuit ...bool) (string, error) {
	return SelectOne(title, options, defOpt, allowQuit...)
}

// SingleSelect is alias of method SelectOne()
func SingleSelect(title string, options any, defOpt string, allowQuit ...bool) (string, error) {
	return SelectOne(title, options, defOpt, allowQuit...)
}

// SelectOne select one of the options, returns selected option value
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
func SelectOne(title string, options any, defOpt string, allowQuit ...bool) (string, error) {
	s := &Select{Title: title, Options: options, DefOpt: defOpt}

	if len(allowQuit) > 0 {
		s.DisableQuit = !allowQuit[0]
	}

	r, err := s.Run()
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// SelectOneKey select one of the options, returns selected option key.
func SelectOneKey(title string, options any, defOpt string, opFns ...func(*Select)) (string, error) {
	r, err := NewSelect(title, options, opFns...).With(func(s *Select) {
		s.DefOpt = defOpt
	}).Run()
	if err != nil {
		return "", err
	}
	return r.KeyString(), nil
}

// Checkbox select multi of the options. is alias of method MultiSelect()
func Checkbox(title string, options any, defOpts []string, allowQuit ...bool) ([]string, error) {
	return MultiSelect(title, options, defOpts, allowQuit...)
}

// MultiSelect select multi of the options, returns selected option values.
//
// like SingleSelect(), but allow select multi option
func MultiSelect(title string, options any, defOpts []string, allowQuit ...bool) ([]string, error) {
	s := &Select{Title: title, Options: options, DefOpts: defOpts, MultiSelect: true}

	if len(allowQuit) > 0 {
		s.DisableQuit = !allowQuit[0]
	}

	r, err := s.Run()
	if err != nil {
		return nil, err
	}
	return r.Strings(), nil
}
