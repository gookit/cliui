package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gookit/cliui/interact/backend"
)

// MultiSelect collects multiple selected items from the user.
type MultiSelect struct {
	Prompt       string
	Items        []Item
	DefaultKeys  []string
	MinSelected  int
	Filterable   bool
	FilterPrompt string
	PageSize     int
}

// NewMultiSelect creates a MultiSelect component.
func NewMultiSelect(prompt string, items []Item) *MultiSelect {
	return &MultiSelect{Prompt: prompt, Items: items}
}

// Run uses the default process stdin/stdout streams.
func (c *MultiSelect) Run(ctx context.Context, be backend.Backend) (*Result, error) {
	return c.RunWithIO(ctx, be, defaultInput(), defaultOutput())
}

// RunWithIO runs the component with explicit IO streams.
func (c *MultiSelect) RunWithIO(ctx context.Context, be backend.Backend, in io.Reader, out io.Writer) (*Result, error) {
	return runWithSession(ctx, be, in, out, func(session backend.Session) (*Result, error) {
		if err := c.ValidateConfig(); err != nil {
			return nil, err
		}

		size := initialTerminalSize(session)
		filter := filterState{}
		cursor := c.defaultIndex()
		selected := c.defaultSelected()
		errMsg := ""
		for {
			indexes := c.filteredIndexes(filter)
			cursor = c.normalizeCursor(cursor, indexes)
			if err := session.Render(c.view(cursor, selected, errMsg, filter, indexes, size)); err != nil {
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
			case backend.KeySpace:
				if len(indexes) == 0 {
					errMsg = "no matched option"
					continue
				}
				item := c.Items[cursor]
				if item.Disabled {
					errMsg = "selected option is disabled"
					continue
				}
				if _, ok := selected[item.Key]; ok {
					delete(selected, item.Key)
				} else {
					selected[item.Key] = item
				}
				continue
			case backend.KeyEnter:
				if strings.TrimSpace(ev.Text) == "" {
					if c.Filterable && len(indexes) == 0 {
						errMsg = "no matched option"
						continue
					}
					result := c.resultFromSelected(selected)
					if len(result.Keys) == 0 {
						errMsg = "at least one option is required"
						continue
					}
					if len(result.Keys) < c.MinSelected {
						errMsg = fmt.Sprintf("select at least %d option(s)", c.MinSelected)
						continue
					}
					return result, nil
				}
			}

			if c.Filterable && ev.Key != backend.KeyEnter && filter.Handle(ev) {
				continue
			}

			raw := strings.TrimSpace(ev.Text)
			keys := c.parseKeys(raw)
			if len(keys) == 0 {
				errMsg = "at least one option is required"
				continue
			}

			result, resolveErr := c.resolve(keys)
			if resolveErr != "" {
				errMsg = resolveErr
				continue
			}

			if len(result.Keys) < c.MinSelected {
				errMsg = fmt.Sprintf("select at least %d option(s)", c.MinSelected)
				continue
			}

			return result, nil
		}
	})
}

// ValidateConfig checks whether the component has enough data to run.
func (c *MultiSelect) ValidateConfig() error {
	if c == nil {
		return fmt.Errorf("%w: multi select is nil", ErrInvalidState)
	}
	if c.Prompt == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidState)
	}
	if c.MinSelected < 0 {
		return fmt.Errorf("%w: min selected must be >= 0", ErrInvalidState)
	}
	if err := validateItems(c.Items); err != nil {
		return err
	}
	return nil
}

func (c *MultiSelect) parseKeys(raw string) []string {
	if raw == "" {
		raw = strings.Join(c.DefaultKeys, ",")
	}

	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	keys := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}

func (c *MultiSelect) resolve(keys []string) (*Result, string) {
	res := &Result{
		Keys:   make([]string, 0, len(keys)),
		Values: make([]any, 0, len(keys)),
	}

	for _, key := range keys {
		item, ok := c.findItem(key)
		if !ok {
			return nil, "unknown option"
		}
		if item.Disabled {
			return nil, "selected option is disabled"
		}

		res.Keys = append(res.Keys, item.Key)
		res.Values = append(res.Values, item.Value)
	}

	if len(res.Keys) > 0 {
		res.Key = res.Keys[0]
		res.Value = res.Values[0]
	}

	return res, ""
}

func (c *MultiSelect) findItem(key string) (Item, bool) {
	for _, item := range c.Items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func (c *MultiSelect) defaultIndex() int {
	return c.firstEnabledIndex()
}

func (c *MultiSelect) firstEnabledIndex() int {
	for i, item := range c.Items {
		if !item.Disabled {
			return i
		}
	}
	return 0
}

func (c *MultiSelect) lastEnabledIndex() int {
	for i := len(c.Items) - 1; i >= 0; i-- {
		if !c.Items[i].Disabled {
			return i
		}
	}
	return 0
}

func (c *MultiSelect) defaultSelected() map[string]Item {
	selected := make(map[string]Item, len(c.DefaultKeys))
	for _, key := range c.DefaultKeys {
		item, ok := c.findItem(key)
		if ok && !item.Disabled {
			selected[key] = item
		}
	}
	return selected
}

func (c *MultiSelect) move(cursor, delta int) int {
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

func (c *MultiSelect) filteredIndexes(filter filterState) []int {
	if c.Filterable {
		return filter.Indexes(c.Items)
	}

	indexes := make([]int, len(c.Items))
	for i := range c.Items {
		indexes[i] = i
	}
	return indexes
}

func (c *MultiSelect) normalizeCursor(cursor int, indexes []int) int {
	if len(indexes) == 0 {
		return cursor
	}
	if indexPosition(indexes, cursor) >= 0 {
		return cursor
	}
	return firstEnabledFilteredIndex(c.Items, indexes)
}

func (c *MultiSelect) moveInIndexes(indexes []int, cursor, delta int) int {
	if !c.Filterable && c.PageSize == 0 {
		return c.move(cursor, delta)
	}
	return moveFilteredCursor(c.Items, indexes, cursor, delta)
}

func (c *MultiSelect) firstEnabledIndexIn(indexes []int) int {
	if !c.Filterable && c.PageSize == 0 {
		return c.firstEnabledIndex()
	}
	return firstEnabledFilteredIndex(c.Items, indexes)
}

func (c *MultiSelect) lastEnabledIndexIn(indexes []int) int {
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

func (c *MultiSelect) resultFromSelected(selected map[string]Item) *Result {
	res := &Result{}
	for _, item := range c.Items {
		if _, ok := selected[item.Key]; !ok {
			continue
		}

		res.Keys = append(res.Keys, item.Key)
		res.Values = append(res.Values, item.Value)
	}

	if len(res.Keys) > 0 {
		res.Key = res.Keys[0]
		res.Value = res.Values[0]
	}

	return res
}

func (c *MultiSelect) view(cursor int, selected map[string]Item, errMsg string, filter filterState, indexes []int, size terminalSize) backend.View {
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
		cursorPrefix := " "
		if i == cursor {
			cursorPrefix = ">"
		}

		check := "[ ]"
		if _, ok := selected[item.Key]; ok {
			check = "[x]"
		}

		line := fmt.Sprintf("%s %s %s) %s", cursorPrefix, check, item.Key, item.Label)
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

	var selectedKeys []string
	for _, item := range c.Items {
		if _, ok := selected[item.Key]; ok {
			selectedKeys = append(selectedKeys, item.Key)
		}
	}
	selectedLine := fmt.Sprintf("Selected(%d): %s", len(selectedKeys), strings.Join(selectedKeys, ", "))
	if len(selectedKeys) == 0 {
		selectedLine = "Selected(0): none"
	}
	lines = append(lines, selectedLine)

	hint := "Use Up/Down to move, Space to toggle, Enter to confirm"
	if c.Filterable {
		hint = "Type to filter, use Up/Down to move, Space to toggle, Enter to confirm"
	}
	lines = append(lines, hint)

	prompt := "Your choices(comma separated)"
	if len(c.DefaultKeys) > 0 {
		prompt += fmt.Sprintf(" [default:%s]", strings.Join(c.DefaultKeys, ","))
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

func (c *MultiSelect) fixedLines() int {
	lines := 5
	if c.Filterable {
		lines++
	}
	return lines
}
