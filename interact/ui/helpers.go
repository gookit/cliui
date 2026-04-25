package ui

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/cliui/interact/backend"
)

func newSession(be backend.Backend, in io.Reader, out io.Writer) (backend.Session, error) {
	if be == nil {
		return nil, fmt.Errorf("%w: nil backend", ErrInvalidState)
	}

	session, err := be.NewSession(in, out)
	if err != nil {
		if errors.Is(err, backend.ErrTTYRequired) {
			return nil, fmt.Errorf("%w: %v", ErrNoTTY, err)
		}
		return nil, err
	}

	return session, nil
}

func runWithSession[T any](
	ctx context.Context,
	be backend.Backend,
	in io.Reader,
	out io.Writer,
	fn func(session backend.Session) (T, error),
) (T, error) {
	var zero T

	session, err := newSession(be, in, out)
	if err != nil {
		return zero, err
	}
	defer session.Close()

	return fn(session)
}

func defaultInput() io.Reader  { return cutypes.Input }
func defaultOutput() io.Writer { return cutypes.Output }

func validateItems(items []Item) error {
	if len(items) == 0 {
		return fmt.Errorf("%w: items is required", ErrInvalidState)
	}

	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		if item.Key == "" {
			return fmt.Errorf("%w: item[%d] key is required", ErrInvalidState, i)
		}
		if item.Label == "" {
			return fmt.Errorf("%w: item[%d] label is required", ErrInvalidState, i)
		}
		if _, ok := seen[item.Key]; ok {
			return fmt.Errorf("%w: duplicate item key %q", ErrInvalidState, item.Key)
		}

		seen[item.Key] = struct{}{}
	}

	return nil
}
