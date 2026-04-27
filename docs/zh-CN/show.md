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

需要表格布局时，可以使用 `show/table`：

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
