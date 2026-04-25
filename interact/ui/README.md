# interact/ui

`interact/ui` provides a new interaction abstraction layer for terminal components.

It is designed for backend-driven interactive flows, and currently includes:

- `Input`
- `Confirm`
- `Select`
- `MultiSelect`

Current status:

- stable component models and result types
- `plain` backend is available
- raw-terminal `readline` backend is available
- existing `interact` package APIs remain unchanged
- `Input` supports UTF-8 text editing and common line-editing shortcuts
- selection components keep validation errors visible until the next user input
- filtering, resize events, tab navigation and page navigation are not implemented yet

## Packages

- `github.com/gookit/cliui/interact/ui`
- `github.com/gookit/cliui/interact/backend`
- `github.com/gookit/cliui/interact/backend/plain`
- `github.com/gookit/cliui/interact/backend/readline`

## Install

```bash
go get github.com/gookit/cliui/interact/ui
```

## Quick Example

### Input

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := plain.New()

	input := ui.NewInput("Your name")
	input.Default = "guest"

	name, err := input.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("name:", name)
}
```

### Confirm

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := plain.New()

	confirm := ui.NewConfirm("Continue", true)
	ok, err := confirm.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("confirmed:", ok)
}
```

### Select

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := plain.New()

	selectOne := ui.NewSelect("Choose env", []ui.Item{
		{Key: "dev", Label: "Development", Value: "dev"},
		{Key: "prod", Label: "Production", Value: "prod"},
	})
	selectOne.DefaultKey = "dev"

	result, err := selectOne.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("selected:", result.Key, result.Value)
}
```

### MultiSelect

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := plain.New()

	selectMany := ui.NewMultiSelect("Choose services", []ui.Item{
		{Key: "api", Label: "API", Value: "api"},
		{Key: "job", Label: "Job Worker", Value: "job"},
		{Key: "web", Label: "Web", Value: "web"},
	})
	selectMany.DefaultKeys = []string{"api", "web"}
	selectMany.MinSelected = 1

	result, err := selectMany.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("selected keys:", result.Keys)
}
```

### Readline Backend

Use `readline` backend when you want event-driven interaction on a real terminal.

`readline.New()` will fall back to the `plain` backend automatically when stdin is not a TTY.
If you want strict TTY-only behavior, use `readline.NewStrict()`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/readline"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := readline.New()

	input := ui.NewInput("Your name")
	input.Default = "guest"

	name, err := input.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	selectOne := ui.NewSelect("Choose env", []ui.Item{
		{Key: "dev", Label: "Development", Value: "dev"},
		{Key: "prod", Label: "Production", Value: "prod"},
	})
	selectOne.DefaultKey = "dev"

	env, err := selectOne.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("name:", name)
	fmt.Println("env:", env.Key)
}
```

## Notes

- `plain` backend uses line-based input and works with ordinary stdin/stdout streams.
- `plain` backend does not provide per-key navigation; select components accept item keys, and multi-select accepts comma-separated item keys.
- `readline` backend uses raw terminal input and supports UTF-8 text, arrow keys, Home/End, Delete, Backspace, Space, Enter, Esc and Ctrl+C.
- `readline.New()` falls back to `plain` when a real terminal is unavailable.
- `readline.NewStrict()` returns an error when stdin is not a real terminal.
- `Select` uses single-key selection by item key.
- `MultiSelect` uses comma-separated item keys.
- `ErrAborted` is returned when the current interaction is canceled.
- `Select` and `MultiSelect` support disabled items and default values.
- `Select` shows the current highlighted item in a dedicated status line.
- `MultiSelect` shows both the current highlighted item and the selected key summary.
- Validation and selection errors stay visible until the next input event changes the component state.
- `Input` cursor placement accounts for display width, so CJK text is handled correctly in supported terminals.
- Terminal resize events are defined in the backend model but are not emitted by the current backends.

## Key Bindings

For the current `readline` backend:

- `Input`: type to insert, `Left/Right` to move, `Home/End` or `Ctrl+A/Ctrl+E` to jump, `Backspace/Delete` to edit, `Ctrl+U` to delete before cursor, `Ctrl+K` to delete after cursor, `Ctrl+W` to delete previous word, `Enter` to submit
- `Confirm`: `Left/Right` to switch, `y/n` to choose, `Enter` to submit current value
- `Select`: `Up/Down` to move, `Enter` to confirm, or type item key directly; the view also shows the current item summary
- `MultiSelect`: `Up/Down` to move, `Space` to toggle, `Enter` to confirm; the view also shows current item and selected key summary

## Backend Behavior

### plain

`plain` is intended for non-TTY input, tests, redirected stdin and simple command-line flows. It reads one line at a time:

- `Input` reads the submitted line as the value.
- `Confirm` accepts `yes/no`, `y/n`, or an empty line for the default.
- `Select` accepts an item key, or the default key on an empty line.
- `MultiSelect` accepts comma-separated item keys, or default keys on an empty line.

### readline

`readline` is intended for real terminal interaction. It normalizes terminal events for the UI components:

- UTF-8 input is read as runes instead of raw bytes.
- Common CSI and SS3 escape sequences are mapped for arrows, Home/End and Delete.
- `Esc` and `Ctrl+C` cancel the current component with `ErrAborted`.
- `readline.New()` falls back to `plain` automatically outside a TTY; use `readline.NewStrict()` for TTY-only behavior.

## Current Limits

- Filtering/search inside `Select` and `MultiSelect` is not implemented.
- `Tab`, `Shift+Tab`, `PageUp` and `PageDown` are not currently mapped.
- Resize events are modeled but not emitted by current backends.
- Very delayed escape sequences may be treated as a standalone `Esc`; common terminal key sequences are handled when available in the input buffer.

## Demo

Run the interactive demo command:

```bash
go run ./examples/interact-ui-demo
```

If you run it in a non-TTY environment, `readline.New()` will fall back to the `plain` backend automatically.

## Example Test

Run the package examples, including the `readline` fallback example:

```bash
go test ./interact/ui -run Example -v
```

## Next Step

The current abstraction is ready for richer event-driven backends and more advanced line editing behavior without changing the `ui` package surface.
