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
	AllowAbort   bool
	EnableFilter bool
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

		cursor := c.defaultIndex()
		selected := c.defaultSelected()
		errMsg := ""
		for {
			if err := session.Render(c.view(cursor, selected, errMsg)); err != nil {
				return nil, err
			}

			ev, err := session.ReadEvent(ctx)
			if err != nil {
				return nil, err
			}

			if ev.Type == backend.EventInterrupt || ev.Key == backend.KeyCtrlC || ev.Key == backend.KeyEsc {
				return nil, ErrAborted
			}

			errMsg = ""
			switch ev.Key {
			case backend.KeyUp:
				cursor = c.move(cursor, -1)
				continue
			case backend.KeyDown:
				cursor = c.move(cursor, 1)
				continue
			case backend.KeySpace:
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
	for i, item := range c.Items {
		if !item.Disabled {
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

func (c *MultiSelect) view(cursor int, selected map[string]Item, errMsg string) backend.View {
	lines := []string{c.Prompt}
	for i, item := range c.Items {
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

	current := c.Items[cursor]
	currentLine := fmt.Sprintf("Current: %s (%s)", current.Label, current.Key)
	if current.Disabled {
		currentLine += " [disabled]"
	}
	lines = append(lines, currentLine)

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
		CursorRow:    len(c.Items) + 4,
		CursorColumn: len(prompt),
	}
}
