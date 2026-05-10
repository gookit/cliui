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

`Title` 用于输出一个醒目的章节标题，适合在命令开始、阶段切换或结果汇总前提示当前上下文。

```go
show.ATitle("Deploy")
```

效果示例：

```txt
-------------------- Deploy --------------------
```

### Any Data

`show.AnyData()` 会将结构化数据渲染为列表；对于标量值，会回退为 pretty JSON 输出。

```go
show.AnyData("user", map[string]any{
	"name": "tom",
	"age":  18,
})
```

效果示例：

```txt
user
  name: tom
  age:  18
```

### Banner

`Banner` 用于输出带边框的块级提示，适合展示状态、阶段标题、重要提示或结果摘要。

```go
show.Banner("Update available")
```

效果示例：

```txt
+------------------+
| Update available |
+------------------+
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

效果示例：

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

`List` 用于展示一组键值信息，适合输出配置、用户信息、运行环境或命令结果详情。

```go
show.AList("User", map[string]any{
	"name": "tom",
	"role": "admin",
})
```

效果示例：

```txt
User
  name: tom
  role: admin
```

### Multi List

`Multi List` 用于一次输出多个分组，每个分组可以是 map、slice 或其它可展示数据。

```go
show.MList(map[string]any{
	"App": map[string]string{"name": "cliui"},
	"Env": []string{"dev", "test"},
})
```

效果示例：

```txt
App
  name: cliui

Env
  - dev
  - test
```

### Table

命令输出需要按列对齐时，可以使用 `show/table`，例如用户列表、进程摘要、包处理结果或部署报告。

基础表格：

```go
tb := table.New("Users")
tb.SetHeads("ID", "Name")
tb.AddRow(1, "Tom")
tb.AddRow(2, "Jane")
tb.Println()
```

效果示例：

```txt
Users
+----+------+
| ID | Name |
+----+------+
| 1  | Tom  |
| 2  | Jane |
+----+------+
```

也可以批量加载行数据。`SetRows()` 支持常见数据形态，例如 `[][]any`、`[]map[string]any` 和结构体切片：

```go
type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

rows := []Package{
	{Name: "fd", Version: "10.2.0", Status: "installed"},
	{Name: "bat", Version: "0.25.0", Status: "pending"},
}

tb := table.New("Packages")
tb.SetRows(rows)
tb.Println()
```

对于简单二维数据，可以先设置表头，再传入行切片：

```go
tb := table.New("Jobs")
tb.SetHeads("Name", "Status", "Duration")
tb.SetRows([][]any{
	{"build", "ok", "12s"},
	{"test", "failed", "31s"},
})
tb.Println()
```

表格样式和边框可以配置：

```go
tb := table.New("Release",
	table.WithStyle(table.StyleRounded),
	table.WithBorderFlags(table.BorderAll),
	table.WithShowRowNumber(true),
)
tb.SetHeads("Package", "Status")
tb.AddRow("cliui", "ready")
tb.AddRow("docs", "updated")
tb.Println()
```

内置样式包括 `StyleSimple`、`StyleMySql`、`StyleMarkdown`、`StyleBold`、`StyleBoldBorder`、`StyleRounded`、`StyleDouble` 和 `StyleMinimal`。

长文本可以配置列宽和溢出策略：

```go
tb := table.New("Tasks")
tb.SetHeads("Task", "Description")
tb.AddRow("download", "Fetch archive from mirror and verify checksum")
tb.AddRow("extract", "Unpack files into the selected destination")
tb.WithOptions(
	table.WithColumnWidths(12, 32),
	table.WithColMaxWidth(32),
	table.WithOverflowFlag(table.OverflowWrap),
)
tb.Println()
```

`OverflowCut` 会截断过长内容，`OverflowWrap` 会把内容换行显示。需要按列排序时，可以使用 `WithSortColumn(index, ascending)`：

```go
tb := table.New("Results")
tb.SetHeads("Name", "Status")
tb.AddRow("test", "failed")
tb.AddRow("build", "ok")
tb.WithOptions(table.WithSortColumn(0, true))
tb.Println()
```

### JSON

`JSON` 用于格式化输出结构化对象，适合调试、展示 API 响应或输出机器可读的结果。

```go
show.JSON(map[string]any{"name": "tom"})
```

效果示例：

```json
{
  "name": "tom"
}
```

### Tab Writer

`Tab Writer` 用于对齐包含 tab 分隔符的文本，适合输出简单的两列或多列列表。

```go
w := show.TabWriter([]string{"Name\tRole", "Tom\tAdmin"})
w.Flush()
```

效果示例：

```txt
Name  Role
Tom   Admin
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
