package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gookit/cliui/interact/backend"
)

// Select collects one selected item from the user.
type Select struct {
	Prompt       string
	Items        []Item
	DefaultKey   string
	Filterable   bool
	FilterPrompt string
	PageSize     int
}

// NewSelect creates a Select component.
func NewSelect(prompt string, items []Item) *Select {
	return &Select{Prompt: prompt, Items: items}
}

// Run uses the default process stdin/stdout streams.
func (c *Select) Run(ctx context.Context, be backend.Backend) (*Result, error) {
	return c.RunWithIO(ctx, be, defaultInput(), defaultOutput())
}

// RunWithIO runs the component with explicit IO streams.
func (c *Select) RunWithIO(ctx context.Context, be backend.Backend, in io.Reader, out io.Writer) (*Result, error) {
	return runWithSession(ctx, be, in, out, func(session backend.Session) (*Result, error) {
		if err := c.ValidateConfig(); err != nil {
			return nil, err
		}

		size := initialTerminalSize(session)
		filter := filterState{}
		cursor := c.defaultIndex()
		errMsg := ""
		for {
			indexes := c.filteredIndexes(filter)
			cursor = c.normalizeCursor(cursor, indexes)
			if err := session.Render(c.view(cursor, errMsg, filter, indexes, size)); err != nil {
				return nil, err
			}

			ev, err := session.ReadEvent(ctx)
			if err != nil {
				return nil, err
			}

			if ev.Type == backend.EventInterrupt || ev.Key == backend.KeyCtrlC || ev.Key == backend.KeyEsc {
				return nil, ErrAborted
			}
			if ev.Type == backend.EventResize {
				size = terminalSizeFromEvent(session, ev, size)
				continue
			}

			errMsg = ""
			switch ev.Key {
			case backend.KeyUp, backend.KeyShiftTab:
				cursor = c.moveInIndexes(indexes, cursor, -1)
				continue
			case backend.KeyDown, backend.KeyTab:
				cursor = c.moveInIndexes(indexes, cursor, 1)
				continue
			case backend.KeyPageUp:
				cursor = c.firstEnabledIndexIn(indexes)
				continue
			case backend.KeyPageDown:
				cursor = c.lastEnabledIndexIn(indexes)
				continue
			case backend.KeyEnter:
				if strings.TrimSpace(ev.Text) == "" {
					if len(indexes) == 0 {
						errMsg = "no matched option"
						continue
					}
					item := c.Items[cursor]
					if item.Disabled {
						errMsg = "selected option is disabled"
						continue
					}
					return singleResult(item), nil
				}
			}

			if c.Filterable && ev.Key != backend.KeyEnter && filter.Handle(ev) {
				continue
			}

			key := strings.TrimSpace(ev.Text)
			if key == "" {
				key = c.DefaultKey
			}

			item, ok := c.findItem(key)
			if !ok {
				errMsg = "unknown option"
				continue
			}
			if item.Disabled {
				errMsg = "selected option is disabled"
				continue
			}

			return singleResult(item), nil
		}
	})
}

// ValidateConfig checks whether the component has enough data to run.
func (c *Select) ValidateConfig() error {
	if c == nil {
		return fmt.Errorf("%w: select is nil", ErrInvalidState)
	}
	if c.Prompt == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidState)
	}
	if err := validateItems(c.Items); err != nil {
		return err
	}
	return nil
}

func (c *Select) findItem(key string) (Item, bool) {
	for _, item := range c.Items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func (c *Select) defaultIndex() int {
	if c.DefaultKey != "" {
		for i, item := range c.Items {
			if item.Key == c.DefaultKey {
				return i
			}
		}
	}

	for i, item := range c.Items {
		if !item.Disabled {
			return i
		}
	}

	return 0
}

func (c *Select) firstEnabledIndex() int {
	for i, item := range c.Items {
		if !item.Disabled {
			return i
		}
	}
	return 0
}

func (c *Select) lastEnabledIndex() int {
	for i := len(c.Items) - 1; i >= 0; i-- {
		if !c.Items[i].Disabled {
			return i
		}
	}
	return 0
}

func (c *Select) move(cursor, delta int) int {
	if len(c.Items) == 0 {
		return 0
	}

	next := cursor
	for range c.Items {
		next = (next + delta + len(c.Items)) % len(c.Items)
		if !c.Items[next].Disabled {
			return next
		}
	}

	return cursor
}

func (c *Select) filteredIndexes(filter filterState) []int {
	if c.Filterable {
		return filter.Indexes(c.Items)
	}

	indexes := make([]int, len(c.Items))
	for i := range c.Items {
		indexes[i] = i
	}
	return indexes
}

func (c *Select) normalizeCursor(cursor int, indexes []int) int {
	if len(indexes) == 0 {
		return cursor
	}
	if indexPosition(indexes, cursor) >= 0 {
		return cursor
	}
	return firstEnabledFilteredIndex(c.Items, indexes)
}

func (c *Select) moveInIndexes(indexes []int, cursor, delta int) int {
	if !c.Filterable && c.PageSize == 0 {
		return c.move(cursor, delta)
	}
	return moveFilteredCursor(c.Items, indexes, cursor, delta)
}

func (c *Select) firstEnabledIndexIn(indexes []int) int {
	if !c.Filterable && c.PageSize == 0 {
		return c.firstEnabledIndex()
	}
	return firstEnabledFilteredIndex(c.Items, indexes)
}

func (c *Select) lastEnabledIndexIn(indexes []int) int {
	if !c.Filterable && c.PageSize == 0 {
		return c.lastEnabledIndex()
	}
	for i := len(indexes) - 1; i >= 0; i-- {
		idx := indexes[i]
		if !c.Items[idx].Disabled {
			return idx
		}
	}
	if len(indexes) > 0 {
		return indexes[len(indexes)-1]
	}
	return 0
}

func (c *Select) view(cursor int, errMsg string, filter filterState, indexes []int, size terminalSize) backend.View {
	lines := []string{c.Prompt}
	if c.Filterable {
		lines = append(lines, fmt.Sprintf("%s: %s", filterPrompt(c.FilterPrompt), filter.Query()))
	}

	visible := indexes
	if c.Filterable || c.PageSize > 0 {
		cursorPos := indexPosition(indexes, cursor)
		pageSize := listPageSize(c.PageSize, size.height, c.fixedLines())
		win := visibleWindow(len(indexes), cursorPos, pageSize)
		visible = indexes[win.start:win.end]
	}

	if len(visible) == 0 {
		lines = append(lines, "No matches")
	}

	for _, i := range visible {
		item := c.Items[i]
		prefix := " "
		if i == cursor {
			prefix = ">"
		}

		line := fmt.Sprintf("%s %s) %s", prefix, item.Key, item.Label)
		if item.Disabled {
			line += " [disabled]"
		}
		lines = append(lines, line)
	}

	if len(indexes) > 0 {
		current := c.Items[cursor]
		currentLine := fmt.Sprintf("Current: %s (%s)", current.Label, current.Key)
		if current.Disabled {
			currentLine += " [disabled]"
		}
		lines = append(lines, currentLine)
	} else {
		lines = append(lines, "Current: none")
	}

	hint := "Use Up/Down to move, Enter to confirm, or input item key"
	if c.Filterable {
		hint = "Type to filter, use Up/Down to move, Enter to confirm"
	}
	lines = append(lines, hint)

	prompt := "Your choice"
	if c.DefaultKey != "" {
		prompt += fmt.Sprintf(" [default:%s]", c.DefaultKey)
	}
	prompt += ":"
	lines = append(lines, prompt)

	if errMsg != "" {
		lines = append(lines, "Error: "+errMsg)
	}

	return backend.View{
		Lines:        lines,
		CursorRow:    len(lines) - 1,
		CursorColumn: len(prompt),
		Width:        size.width,
		Height:       size.height,
		HideCursor:   c.Filterable,
	}
}

func (c *Select) fixedLines() int {
	lines := 4
	if c.Filterable {
		lines++
	}
	return lines
}

func singleResult(item Item) *Result {
	return &Result{
		Key:   item.Key,
		Keys:  []string{item.Key},
		Value: item.Value,
		Values: []any{
			item.Value,
		},
	}
}
