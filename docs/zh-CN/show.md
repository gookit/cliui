# show

`show` 为 CLI 应用提供结构化终端输出辅助能力。

它适合渲染可读性更好的终端内容，例如：

- banner
- title
- table
- list
- multi list
- alert message
- JSON
- generic data

## 文档

- GoDoc: https://pkg.go.dev/github.com/gookit/cliui/show

## 安装

```shell
go get github.com/gookit/cliui/show
```

## 使用

### Title

```go
show.ATitle("Deploy")
```

### Any Data

`show.AnyData()` 会将结构化数据渲染为列表；对于标量值，会回退为 pretty JSON 输出。

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

需要更多控制时，可以直接使用 `show/banner`。

```go
b := banner.New("Update available", banner.WithBannerCenter())
fmt.Print(b.Render())
```

### Alert

`show/alert` 使用 banner 组件渲染提示消息：

```go
alert.Info("ready")
alert.Warning("check config")
alert.Success("created %s", "user")
alert.Error("failed: %s", "network")
```

构建一个可复用的 alert 消息：

```go
msg := alert.New("INFO", "ready", 0)
msg.Println()
```

### List

```go
show.AList("User", map[string]any{
	"name": "tom",
	"role": "admin",
})
```

### Multi List

```go
show.MList(map[string]any{
	"App": map[string]string{"name": "cliui"},
	"Env": []string{"dev", "test"},
})
```

### Table

需要表格布局时，可以使用 `show/table`：

```go
tb := table.New("Users")
tb.SetHeads("ID", "Name")
tb.AddRow(1, "Tom")
tb.AddRow(2, "Jane")
tb.Println()
```

### JSON

```go
show.JSON(map[string]any{"name": "tom"})
```

### Tab Writer

```go
w := show.TabWriter([]string{"Name\tRole", "Tom\tAdmin"})
w.Flush()
```

更多用法可以参考包测试和导出的 API。

## 开发

测试：

```bash
go test -v -run ^TestTable_MultiLineContent$ ./show/table/...
```

## 相关项目

- https://github.com/jedib0t/go-pretty
- https://github.com/alexeyco/simpletable
- https://github.com/InVisionApp/tabular
- https://github.com/gosuri/uitable
- https://github.com/rodaine/table
- https://github.com/tomlazar/table
- https://github.com/nwidger/jsoncolor
