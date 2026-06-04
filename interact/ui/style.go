package ui

import (
	"fmt"
	"strings"
)

func uiTag(tag, text string) string {
	return fmt.Sprintf("<%s>%s</>", tag, text)
}

func cursorPrefix(active bool) string {
	if active {
		return uiTag("cyan", ">")
	}
	return " "
}

func disabledSuffix(disabled bool) string {
	if disabled {
		return " " + uiTag("gray", "[disabled]")
	}
	return ""
}

func currentLine(item Item) string {
	line := fmt.Sprintf("%s %s (%s)", uiTag("cyan", "Current:"), item.Label, uiTag("yellow", item.Key))
	return line + disabledSuffix(item.Disabled)
}

func errorLine(msg string) string {
	return uiTag("red", "Error:") + " " + msg
}

func confirmOption(val bool) string {
	if val {
		return uiTag("green", "yes")
	}
	return uiTag("red", "no")
}

func confirmHint() string {
	return fmt.Sprintf("Input %s or %s, %s accepts current",
		uiTag("green", "y/yes"), uiTag("red", "n/no"), uiTag("green", "Enter"))
}

func selectHint(filterable bool) string {
	if filterable {
		return fmt.Sprintf("Type to filter, use %s to move, %s to confirm",
			uiTag("cyan", "Up/Down"), uiTag("green", "Enter"))
	}
	return fmt.Sprintf("Use %s to move, %s to confirm, or input %s",
		uiTag("cyan", "Up/Down"), uiTag("green", "Enter"), uiTag("yellow", "item key"))
}

func multiSelectHint(filterable bool) string {
	if filterable {
		return fmt.Sprintf("Type to filter, use %s to move, %s to toggle, %s to confirm",
			uiTag("cyan", "Up/Down"), uiTag("yellow", "Space"), uiTag("green", "Enter"))
	}
	return fmt.Sprintf("Use %s to move, %s to toggle, %s to confirm",
		uiTag("cyan", "Up/Down"), uiTag("yellow", "Space"), uiTag("green", "Enter"))
}

func checkMark(checked bool) string {
	if checked {
		return uiTag("green", "[x]")
	}
	return uiTag("gray", "[ ]")
}

func selectedLine(keys []string) string {
	if len(keys) == 0 {
		return uiTag("green", "Selected(0):") + " none"
	}

	tagged := make([]string, 0, len(keys))
	for _, key := range keys {
		tagged = append(tagged, uiTag("yellow", key))
	}
	return fmt.Sprintf("%s %s", uiTag("green", fmt.Sprintf("Selected(%d):", len(keys))), strings.Join(tagged, ", "))
}
