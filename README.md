# CliUI

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/cliui?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/cliui)](https://github.com/gookit/cliui)
[![Unit-Tests](https://github.com/gookit/cliui/actions/workflows/go.yml/badge.svg)](https://github.com/gookit/cliui)

---

`gookit/cliui` is a terminal UI helper module, provides some commonly used display and interactive functional components for CLI.

It focuses on three core areas for CLI applications:

- `show`: structured terminal output helpers
- `interact`: interactive input helpers
- `progress`: progress and loading display helpers

The root package also provides shared input/output helpers used by the subpackages:

- `SetInput(in io.Reader)`
- `SetOutput(out io.Writer)`
- `CustomIO(in io.Reader, out io.Writer)`
- `ResetInput()`
- `ResetOutput()`
- `ResetIO()`

## Install

```bash
go get github.com/gookit/cliui
```

Or install a sub package directly:

```bash
go get github.com/gookit/cliui/show
go get github.com/gookit/cliui/progress
go get github.com/gookit/cliui/interact
```

## Package Overview

### `show`

Provides structured terminal output helpers for displaying formatted content in CLI applications. It includes `table`, `title`, `banner`, `list`, `multi-list`, `alert` messages and `JSON` output helpers.

Import:

```go
github.com/gookit/cliui/show
```

> [!NOTE]
> More Details: [docs/show.md](docs/show.md)

**Quick Usage**:

```go
package main

import "github.com/gookit/cliui/show"

func main() {
	show.Banner("Deploy started")
	show.AList("App info", map[string]any{
		"name": "cliui",
		"env":  "dev",
	})
	show.JSON(map[string]any{"ok": true})
}
```

### `interact`

Provides interactive input helpers for CLI programs. It supports `prompt`, `confirm`, `question`, `select`, `multi-select`, `password` input and other common terminal interaction patterns.

It also includes a newer `interact/ui` layer for backend-driven components:

- `plain` backend for line-based input, tests and redirected stdin
- `readline` backend for raw terminal interaction
- `Input`, `Confirm`, `Select` and `MultiSelect` components
- UTF-8 input editing, common navigation keys and persistent validation errors

Import:

```go
github.com/gookit/cliui/interact
```

> [!NOTE]
> More Details: [docs/interact.md](docs/interact.md)

**Quick Usage**:

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

	name, err := ui.NewInput("Your name").Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	env, err := ui.NewSelect("Choose env", []ui.Item{
		{Key: "dev", Label: "Development", Value: "dev"},
		{Key: "prod", Label: "Production", Value: "prod"},
	}).Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("name:", name)
	fmt.Println("env:", env.Value)
}
```

### `progress`

Provides progress and loading display helpers for long-running tasks. It includes `progress bars`, `text bars`, `spinner/loading` indicators, `counters`, `dynamic text` output and `MultiProgress` for rendering multiple bars in one terminal block.

Import:

```go
github.com/gookit/cliui/progress
```

> [!NOTE]
> More Details: [docs/progress.md](docs/progress.md)

**Quick Usage**:

```go
package main

import (
	"time"

	"github.com/gookit/cliui/progress"
)

func main() {
	p := progress.Bar(100)
	p.Start()

	for i := 0; i < 100; i++ {
		time.Sleep(20 * time.Millisecond)
		p.Advance()
	}

	p.Finish()
}
```

## Migration

If you are migrating from `gookit/gcli/v3`, the corresponding package paths are:

- `github.com/gookit/gcli/v3/show` -> `github.com/gookit/cliui/show`
- `github.com/gookit/gcli/v3/interact` -> `github.com/gookit/cliui/interact`
- `github.com/gookit/gcli/v3/progress` -> `github.com/gookit/cliui/progress`

## Development

```bash
go test ./...
```

## License

MIT
