# cliui

`cliui` is a terminal UI helper module extracted from `gookit/gcli`.

It focuses on three areas for CLI applications:

- `show`: structured terminal output helpers such as table, title, banner and json
- `progress`: progress bar and loading display helpers
- `interact`: interactive input, select, confirm and prompt helpers

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

## Packages

- `github.com/gookit/cliui/show`
- `github.com/gookit/cliui/progress`
- `github.com/gookit/cliui/interact`

## Status

This repository was split from `gookit/gcli` and keeps the related package history.
The first step keeps the package layout stable for easier migration from:

- `github.com/gookit/gcli/v3/show`
- `github.com/gookit/gcli/v3/progress`
- `github.com/gookit/gcli/v3/interact`

to:

- `github.com/gookit/cliui/show`
- `github.com/gookit/cliui/progress`
- `github.com/gookit/cliui/interact`

## Development

```bash
go test ./...
```

## License

MIT
