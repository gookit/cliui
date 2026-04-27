package ui

import (
	"fmt"
	"strings"

	"github.com/gookit/cliui/interact/backend"
)

type filterState struct {
	query []rune
}

func (s *filterState) Query() string {
	return string(s.query)
}

func (s *filterState) Handle(ev backend.Event) bool {
	switch ev.Key {
	case backend.KeyBackspace:
		if len(s.query) > 0 {
			s.query = s.query[:len(s.query)-1]
		}
		return true
	case backend.KeyCtrlU:
		s.query = nil
		return true
	}

	if ev.Text == "" {
		return false
	}

	s.query = append(s.query, []rune(ev.Text)...)
	return true
}

func (s *filterState) Indexes(items []Item) []int {
	query := strings.ToLower(strings.TrimSpace(s.Query()))
	indexes := make([]int, 0, len(items))
	for i, item := range items {
		if query == "" || itemMatchesQuery(item, query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func itemMatchesQuery(item Item, query string) bool {
	if strings.Contains(strings.ToLower(item.Key), query) {
		return true
	}
	if strings.Contains(strings.ToLower(item.Label), query) {
		return true
	}
	if item.Value != nil && strings.Contains(strings.ToLower(fmt.Sprint(item.Value)), query) {
		return true
	}
	return false
}

func filterPrompt(label string) string {
	if label == "" {
		return "Filter"
	}
	return label
}
