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

### Title

`Title` prints a prominent section heading. It is useful before command execution, phase changes, or result summaries.

```go
show.ATitle("Deploy")
```

Output preview:

```txt
-------------------- Deploy --------------------
```

### Any Data

`show.AnyData()` renders structured data as a list and falls back to pretty JSON for scalar values.

```go
show.AnyData("user", map[string]any{
	"name": "tom",
	"age":  18,
})
```

Output preview:

```txt
user
  name: tom
  age:  18
```

### Banner

`Banner` prints a bordered block message. It is useful for status messages, phase titles, important notices, and summaries.

```go
show.Banner("Update available")
```

Output preview:

```txt
+------------------+
| Update available |
+------------------+
```

For more control, use `show/banner` directly.

```go
b := banner.New("Update available", banner.WithBannerCenter())
fmt.Print(b.Render())
```

### Alert

`show/alert` renders alert messages using the banner component:

```go
alert.Info("ready")
alert.Warning("check config")
alert.Success("created %s", "user")
alert.Error("failed: %s", "network")
```

Build a reusable alert message:

```go
msg := alert.New("INFO", "ready", 0)
msg.Println()
```

Output preview:

```txt
+---------+
| INFO    | ready
+---------+
| WARNING | check config
+---------+
| SUCCESS | created user
+---------+
| ERROR   | failed: network
+---------+
```

### List

`List` displays key-value data. It is useful for configuration, user details, runtime information, and command result details.

```go
show.AList("User", map[string]any{
	"name": "tom",
	"role": "admin",
})
```

Output preview:

```txt
User
  name: tom
  role: admin
```

### Multi List

`Multi List` prints several grouped sections at once. Each group can contain maps, slices, or other displayable data.

```go
show.MList(map[string]any{
	"App": map[string]string{"name": "cliui"},
	"Env": []string{"dev", "test"},
})
```

Output preview:

```txt
App
  name: cliui

Env
  - dev
  - test
```

### Table

Use `show/table` when you need a tabular layout:

```go
tb := table.New("Users")
tb.SetHeads("ID", "Name")
tb.AddRow(1, "Tom")
tb.AddRow(2, "Jane")
tb.Println()
```

Output preview:

```txt
Users
+----+------+
| ID | Name |
+----+------+
| 1  | Tom  |
| 2  | Jane |
+----+------+
```

### JSON

`JSON` prints formatted structured objects. It is useful for debugging, showing API responses, or returning machine-readable output.

```go
show.JSON(map[string]any{"name": "tom"})
```

Output preview:

```json
{
  "name": "tom"
}
```

### Tab Writer

`Tab Writer` aligns text that contains tab separators. It is useful for simple two-column or multi-column lists.

```go
w := show.TabWriter([]string{"Name\tRole", "Tom\tAdmin"})
w.Flush()
```

Output preview:

```txt
Name  Role
Tom   Admin
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
