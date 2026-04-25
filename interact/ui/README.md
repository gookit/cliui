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
- minimal raw-terminal `readline` backend is available
- existing `interact` package APIs remain unchanged
- advanced line editing and filtering are not implemented yet

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
- `readline` backend uses raw terminal input and supports arrow keys, space, enter, esc and ctrl+c.
- `readline.New()` falls back to `plain` when a real terminal is unavailable.
- `Select` uses single-key selection by item key.
- `MultiSelect` uses comma-separated item keys.
- `ErrAborted` is returned when the current interaction is canceled.
- `Select` and `MultiSelect` support disabled items and default values.
- `Select` shows the current highlighted item in a dedicated status line.
- `MultiSelect` shows both the current highlighted item and the selected key summary.

## Key Bindings

For the current `readline` backend:

- `Input`: type to insert, `Left/Right` to move, `Home/End` or `Ctrl+A/Ctrl+E` to jump, `Backspace/Delete` to edit, `Ctrl+U/Ctrl+K` to clear around cursor, `Ctrl+W` to delete previous word, `Enter` to submit
- `Confirm`: `Left/Right` to switch, `y/n` to choose, `Enter` to submit current value
- `Select`: `Up/Down` to move, `Enter` to confirm, or type item key directly; the view also shows the current item summary
- `MultiSelect`: `Up/Down` to move, `Space` to toggle, `Enter` to confirm; the view also shows current item and selected key summary

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
