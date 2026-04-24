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

## Notes

- `plain` backend uses line-based input and works with ordinary stdin/stdout streams.
- `readline` backend uses raw terminal input and supports arrow keys, space, enter, esc and ctrl+c.
- `Select` uses single-key selection by item key.
- `MultiSelect` uses comma-separated item keys.
- `ErrAborted` is returned when the current interaction is canceled.
- `Select` and `MultiSelect` support disabled items and default values.

## Next Step

The current abstraction is ready for richer event-driven backends and more advanced line editing behavior without changing the `ui` package surface.
