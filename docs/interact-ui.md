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
- filtering and resize events are not implemented yet

## Packages

- `github.com/gookit/cliui/interact/ui`
- `github.com/gookit/cliui/interact/backend`
- `github.com/gookit/cliui/interact/backend/plain`
- `github.com/gookit/cliui/interact/backend/readline`
- `github.com/gookit/cliui/interact/backend/fake`

## Install

```bash
go get github.com/gookit/cliui/interact/ui
```

## Quick Example

### Input

`Input` reads one text line. With the `readline` backend it supports cursor movement, deletion, UTF-8 editing, and other real-terminal editing behavior; with the `plain` backend it reads one submitted line.

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

Minimal form:

```go
be := plain.New()
name, err := ui.NewInput("Your name").Run(context.Background(), be)
```

Output preview:

```txt
Your name [guest]: tom
name: tom
```

### Confirm

`Confirm` handles a yes/no choice. It shows the default option, accepts `y/n`, and can be toggled with arrow keys when using the `readline` backend.

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

Minimal form:

```go
be := plain.New()
ok, err := ui.NewConfirm("Continue", true).Run(context.Background(), be)
```

Output preview:

```txt
Continue [Y/n]: y
confirmed: true
```

### Select

`Select` chooses one result from an `Item` list. Each item can provide a stable `Key`, display `Label`, and application-facing `Value`.

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
selectOne.Filterable = true
selectOne.PageSize = 10

result, err := selectOne.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("selected:", result.Key, result.Value)
}
```

Minimal form:

```go
be := plain.New()
result, err := ui.NewSelect("Choose env", []ui.Item{
	{Key: "dev", Label: "Development", Value: "dev"},
	{Key: "prod", Label: "Production", Value: "prod"},
}).Run(context.Background(), be)
```

When filtering is enabled with the `readline` backend, type to narrow the options, use `Backspace` to delete filter text, and use `Ctrl+U` to clear the filter.

Output preview:

```txt
Choose env
> dev   Development
  prod  Production
selected: dev dev
```

### MultiSelect

`MultiSelect` chooses multiple results from an `Item` list. It supports defaults, minimum selection count, and disabled items.

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
selectMany.Filterable = true
selectMany.PageSize = 10

result, err := selectMany.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("selected keys:", result.Keys)
}
```

Minimal form:

```go
be := plain.New()
result, err := ui.NewMultiSelect("Choose services", []ui.Item{
	{Key: "api", Label: "API", Value: "api"},
	{Key: "web", Label: "Web", Value: "web"},
}).Run(context.Background(), be)
```

Filtering only changes which options are visible. Already selected items are preserved when the filter changes.

Output preview:

```txt
Choose services
> [x] api  API
  [ ] job  Job Worker
  [x] web  Web
selected keys: [api web]
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

Output preview:

```txt
Your name [guest]: tom
Choose env
> dev   Development
  prod  Production
name: tom
env: dev
```

### Fake Backend

Use `fake` backend in tests to feed normalized events without a real terminal:

```go
be := fake.New(
	backend.Event{Type: backend.EventKey, Text: "tom"},
	backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
)

name, err := ui.NewInput("Your name").Run(context.Background(), be)
```

Output preview:

```txt
Your name: tom
```

## Notes

- `plain` backend uses line-based input and works with ordinary stdin/stdout streams.
- `plain` backend does not provide per-key navigation; select components accept item keys, and multi-select accepts comma-separated item keys.
- `readline` backend uses raw terminal input and supports UTF-8 text, arrow keys, Home/End, Delete, Backspace, Tab, Shift+Tab, PageUp/PageDown, Space, Enter, Esc and Ctrl+C.
- `readline.New()` falls back to `plain` when a real terminal is unavailable.
- `readline.NewStrict()` returns an error when stdin is not a real terminal.
- `fake` backend is intended for deterministic component tests.
- `Select` uses single-key selection by item key.
- `MultiSelect` uses comma-separated item keys.
- `ErrAborted` is returned when the current interaction is canceled.
- `Select` and `MultiSelect` support disabled items and default values.
- `Select` and `MultiSelect` can enable per-key filtering with `Filterable`.
- `PageSize` limits the number of visible option rows; when it is `0`, components calculate it from terminal height.
- `Select` shows the current highlighted item in a dedicated status line.
- `MultiSelect` shows both the current highlighted item and the selected key summary.
- Validation and selection errors stay visible until the next input event changes the component state.
- `Input` cursor placement accounts for display width, so CJK text is handled correctly in supported terminals.
- The `readline` backend emits resize events when terminal size changes. Components recalculate visible rows while preserving the current filter and selection state.

## Key Bindings

For the current `readline` backend:

- `Input`: type to insert, `Left/Right` to move, `Home/End` or `Ctrl+A/Ctrl+E` to jump, `Backspace/Delete` to edit, `Ctrl+U` to delete before cursor, `Ctrl+K` to delete after cursor, `Ctrl+W` to delete previous word, `Enter` to submit
- `Confirm`: `Left/Right` to switch, `y/n` to choose, `Enter` to submit current value
- `Select`: `Up/Down` or `Tab/Shift+Tab` to move, `PageUp/PageDown` to jump, `Enter` to confirm; when filtering is enabled, normal text appends to the filter, `Backspace` deletes filter text, and `Ctrl+U` clears the filter
- `MultiSelect`: `Up/Down` or `Tab/Shift+Tab` to move, `PageUp/PageDown` to jump, `Space` to toggle, `Enter` to confirm; when filtering is enabled, normal text appends to the filter, `Backspace` deletes filter text, and `Ctrl+U` clears the filter

## Backend Behavior

### plain

`plain` is intended for non-TTY input, tests, redirected stdin and simple command-line flows. It reads one line at a time:

- `Input` reads the submitted line as the value.
- `Confirm` accepts `yes/no`, `y/n`, or an empty line for the default.
- `Select` accepts an item key, or the default key on an empty line.
- `MultiSelect` accepts comma-separated item keys, or default keys on an empty line.
- `plain` does not provide per-key filtering. Even when `Filterable` is set, input is still parsed as a submitted line.

### readline

`readline` is intended for real terminal interaction. It normalizes terminal events for the UI components:

- UTF-8 input is read as runes instead of raw bytes.
- Common CSI and SS3 escape sequences are mapped for arrows, Home/End, Delete, Shift+Tab and PageUp/PageDown.
- `Esc` and `Ctrl+C` cancel the current component with `ErrAborted`.
- Filterable `Select` and `MultiSelect` update the filter on each text key event.
- Resize events trigger a redraw while keeping the current filter, highlighted item, and selected multi-select items.
- `readline.New()` falls back to `plain` automatically outside a TTY; use `readline.NewStrict()` for TTY-only behavior.

## Current Limits

- The `plain` backend does not support live filtering or resize events.
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

The current abstraction is ready for richer event-driven backends, resize handling and future selection filtering without changing the `ui` package surface.
