# cliui

`cliui` is a terminal UI helper module extracted from `gookit/gcli`.

It focuses on three core areas for CLI applications:

- `show`: structured terminal output helpers
- `interact`: interactive input helpers
- `progress`: progress and loading display helpers

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

Provides structured terminal output helpers for displaying formatted content in CLI applications. It includes components such as table, title, banner, list, alert and JSON output.

Import:

```go
github.com/gookit/cliui/show
```

Details: [show/README.md](show/README.md)

### `interact`

Provides interactive input helpers for CLI programs. It supports prompt, confirm, question, select, multi-select, password input and other common terminal interaction patterns.

Import:

```go
github.com/gookit/cliui/interact
```

Details: [interact/README.md](interact/README.md)

### `progress`

Provides progress and loading display helpers for long-running tasks. It includes progress bars, text bars, spinner/loading indicators, counters and dynamic text output.

Import:

```go
github.com/gookit/cliui/progress
```

Details: [progress/README.md](progress/README.md)

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
