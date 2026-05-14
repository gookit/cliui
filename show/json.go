package show

import (
	"encoding/json"
	"fmt"

	"github.com/gookit/cliui/show/showcom"
	"github.com/gookit/color"
)

// PrettyJSON struct
type PrettyJSON struct {
	showcom.Base

	Data        any
	Prefix      string
	Indent      string
	KeyStyle    string
	BoolStyle   string
	NumberStyle string
}

// NewPrettyJSON create an instance.
func NewPrettyJSON(v ...any) *PrettyJSON {
	pj := &PrettyJSON{
		Indent:      "    ",
		KeyStyle:    "info",
		BoolStyle:   "success",
		NumberStyle: "warning",
	}
	if len(v) > 0 {
		pj.Data = v[0]
	}
	if len(v) > 1 {
		pj.Prefix = fmt.Sprint(v[1])
	}
	if len(v) > 2 {
		pj.Indent = fmt.Sprint(v[2])
	}
	pj.FormatFn = pj.Format
	return pj
}

// Format pretty format JSON with colors.
func (pj *PrettyJSON) Format() {
	bs, err := json.MarshalIndent(pj.Data, pj.Prefix, pj.Indent)
	if err != nil {
		panic(err)
	}

	pj.Buffer().WriteString(pj.colorize(string(bs)))
}

func (pj *PrettyJSON) colorize(s string) string {
	buf := make([]byte, 0, len(s))

	for i := 0; i < len(s); {
		switch {
		case s[i] == '"':
			start := i
			i++
			for i < len(s) {
				if s[i] == '\\' {
					i += 2
					continue
				}
				if s[i] == '"' {
					i++
					break
				}
				i++
			}

			token := s[start:i]
			j := skipJSONSpaces(s, i)
			if j < len(s) && s[j] == ':' {
				token = color.WrapTag(token, pj.KeyStyle)
			}
			buf = append(buf, token...)
		case hasJSONWord(s, i, "true"):
			buf = append(buf, color.WrapTag("true", pj.BoolStyle)...)
			i += 4
		case hasJSONWord(s, i, "false"):
			buf = append(buf, color.WrapTag("false", pj.BoolStyle)...)
			i += 5
		case isJSONNumberStart(s[i]):
			start := i
			i = scanJSONNumber(s, i)
			buf = append(buf, color.WrapTag(s[start:i], pj.NumberStyle)...)
		default:
			buf = append(buf, s[i])
			i++
		}
	}

	return string(buf)
}

func skipJSONSpaces(s string, i int) int {
	for i < len(s) {
		switch s[i] {
		case ' ', '\n', '\r', '\t':
			i++
		default:
			return i
		}
	}
	return i
}

func hasJSONWord(s string, i int, word string) bool {
	end := i + len(word)
	return end <= len(s) && s[i:end] == word && (end == len(s) || isJSONDelimiter(s[end]))
}

func isJSONDelimiter(c byte) bool {
	switch c {
	case ' ', '\n', '\r', '\t', ',', '}', ']':
		return true
	default:
		return false
	}
}

func isJSONNumberStart(c byte) bool {
	return c == '-' || (c >= '0' && c <= '9')
}

func scanJSONNumber(s string, i int) int {
	if s[i] == '-' {
		i++
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	return i
}
