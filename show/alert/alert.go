package alert

import (
	"fmt"

	"github.com/gookit/cliui/show/banner"
	"github.com/gookit/cliui/show/showcom"
)

/*
提示消息: Error, Success, Warning, Info, Debug, Notice, Question, Alert, Fatal, Panic

   ╭──────────────────────────────────────────────────────────────────╮
   │   Error : Update available! 3.21.0 → 3.27.0  	 	 	 	 	  │
   ╰──────────────────────────────────────────────────────────────────╯

*/

type MsgBox struct {
	TypeText  string
	TypeColor string
	Content   string
	Code      int
}

var (
	ErrorMsg = MsgBox{
		TypeText:  "ERROR",
		TypeColor: "red1",
		Code:      1,
	}
	SuccessMsg = MsgBox{
		TypeText:  "SUCCESS",
		TypeColor: "green",
		Code:      0,
	}
	InfoMsg = MsgBox{
		TypeText:  "INFO",
		TypeColor: "cyan",
		Code:      0,
	}
	WarningMsg = MsgBox{
		TypeText:  "WARNING",
		TypeColor: "yellow",
		Code:      0,
	}
)

// New creates a message box.
func New(typeText, content string, code int) *MsgBox {
	return &MsgBox{TypeText: typeText, Content: content, Code: code}
}

// WithContent returns a copy of the message box with content.
func (m MsgBox) WithContent(format string, v ...any) MsgBox {
	m.Content = fmt.Sprintf(format, v...)
	return m
}

// Render formats the message box as a banner string.
func (m MsgBox) Render() string {
	return banner.New(
		m.line(),
		banner.WithMinWidth(30),
		banner.WithOverflowFlag(showcom.OverflowWrap),
		banner.WithTextColor(m.TypeColor),
	).Render()
}

// Println prints the message box and returns its code.
func (m MsgBox) Println() int {
	banner.New(
		m.line(),
		banner.WithMinWidth(30),
		banner.WithOverflowFlag(showcom.OverflowWrap),
		banner.WithTextColor(m.TypeColor),
	).Println()
	return m.Code
}

func (m MsgBox) line() string {
	if m.TypeText == "" {
		return m.Content
	}
	return fmt.Sprintf("%s: %s", m.TypeText, m.Content)
}

// Error tips message print
func Error(format string, v ...any) int {
	return ErrorMsg.WithContent(format, v...).Println()
}

// Success tips message print
func Success(format string, v ...any) int {
	return SuccessMsg.WithContent(format, v...).Println()
}

// Info tips message print
func Info(format string, v ...any) int {
	return InfoMsg.WithContent(format, v...).Println()
}

// Warning tips message print
func Warning(format string, v ...any) int {
	return WarningMsg.WithContent(format, v...).Println()
}
