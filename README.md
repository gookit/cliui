# cliui

`cliui` is a terminal UI helper module extracted from `gookit/gcli`.

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

Provides structured terminal output helpers for displaying formatted content in CLI applications. It includes table, title, banner, list, multi-list, banner-based alert messages and JSON output helpers.

Import:

```go
github.com/gookit/cliui/show
```

Details: [docs/show.md](docs/show.md)

### `interact`

Provides interactive input helpers for CLI programs. It supports prompt, confirm, question, select, multi-select, password input and other common terminal interaction patterns.

It also includes a newer `interact/ui` layer for backend-driven components:

- `plain` backend for line-based input, tests and redirected stdin
- `readline` backend for raw terminal interaction
- `Input`, `Confirm`, `Select` and `MultiSelect` components
- UTF-8 input editing, common navigation keys and persistent validation errors

Import:

```go
github.com/gookit/cliui/interact
```

Details: [docs/interact.md](docs/interact.md)

### `progress`

Provides progress and loading display helpers for long-running tasks. It includes progress bars, text bars, spinner/loading indicators, counters, dynamic text output and `MultiProgress` for rendering multiple bars in one terminal block.

Import:

```go
github.com/gookit/cliui/progress
```

Details: [docs/progress.md](docs/progress.md)

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
