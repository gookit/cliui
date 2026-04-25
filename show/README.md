# show

`show` provides structured terminal output helpers for CLI applications.

It is designed for rendering readable terminal content such as:

- banner
- title
- table
- list
- multi list
- alert message
- JSON
- generic data

## Documentation

- GoDoc: https://pkg.go.dev/github.com/gookit/cliui/show

## Install

```shell
go get github.com/gookit/cliui/show
```

## Usage

### Any Data

`show.AnyData()` renders structured data as a list and falls back to pretty JSON for scalar values.

```go
show.AnyData("user", map[string]any{
	"name": "tom",
	"age":  18,
})
```

### Banner

```go
show.Banner("Update available")
```

For more control, use `show/banner` directly.

### Alert

`show/alert` renders alert messages using the banner component:

```go
alert.Info("ready")
alert.Warning("check config")
alert.Success("created %s", "user")
alert.Error("failed: %s", "network")
```

### JSON

```go
show.JSON(map[string]any{"name": "tom"})
```

See package tests and exported APIs for more usage examples.

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
