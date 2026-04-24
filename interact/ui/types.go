// Package ui defines interactive component models built on top of a backend.
package ui

import (
	"context"

	"github.com/gookit/cliui/interact/backend"
)

// Runner is implemented by interactive components that return a value.
type Runner[T any] interface {
	Run(ctx context.Context, be backend.Backend) (T, error)
}

// Item is a selectable option.
type Item struct {
	Key      string
	Label    string
	Value    any
	Disabled bool
}

// Result is a common selection result.
type Result struct {
	Key    string
	Keys   []string
	Value  any
	Values []any
}
