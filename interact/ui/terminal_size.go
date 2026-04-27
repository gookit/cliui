package ui

import "github.com/gookit/cliui/interact/backend"

const defaultListPageSize = 10

type terminalSize struct {
	width  int
	height int
}

func initialTerminalSize(session backend.Session) terminalSize {
	width, height := session.Size()
	return terminalSize{width: width, height: height}
}

func terminalSizeFromEvent(session backend.Session, ev backend.Event, current terminalSize) terminalSize {
	if ev.Type != backend.EventResize {
		return current
	}

	next := current
	if ev.Width > 0 {
		next.width = ev.Width
	}
	if ev.Height > 0 {
		next.height = ev.Height
	}
	if next.width <= 0 || next.height <= 0 {
		width, height := session.Size()
		if next.width <= 0 {
			next.width = width
		}
		if next.height <= 0 {
			next.height = height
		}
	}
	return next
}

func listPageSize(configured, terminalHeight, fixedLines int) int {
	if configured > 0 {
		if terminalHeight > 0 {
			available := terminalHeight - fixedLines
			if available < 1 {
				available = 1
			}
			if configured > available {
				return available
			}
		}
		return configured
	}

	if terminalHeight <= 0 {
		return defaultListPageSize
	}

	available := terminalHeight - fixedLines
	if available < 1 {
		return 1
	}
	return available
}
