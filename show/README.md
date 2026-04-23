# show

`show` provides structured terminal output helpers for CLI applications.

It is designed for rendering readable terminal content such as:

- banner
- title
- table
- panel
- section
- padding
- list
- multi list
- alert block
- markdown
- JSON

## Documentation

- GoDoc: https://pkg.go.dev/github.com/gookit/cliui/show

## Install

```shell
go get github.com/gookit/cliui/show
```

## Usage

See package tests and exported APIs for usage examples.

## Development

Testing:

```bash
go test -v -run ^TestTable_MultiLineContent$ ./show/table/...
```

## Related

- https://github.com/jedib0t/go-pretty
- https://github.com/alexeyco/simpletable
- https://github.com/InVisionApp/tabular
- https://github.com/gosuri/uitable
- https://github.com/rodaine/table
- https://github.com/tomlazar/table
- https://github.com/nwidger/jsoncolor
