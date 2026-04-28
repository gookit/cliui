# interact

`interact` 提供命令行交互式输入辅助能力。

它包含常见终端交互方法，例如：

- `ReadInput`
- `ReadLine`
- `ReadFirst`
- `Prompt`
- `Confirm`
- `Query/Question/Ask`
- `Select/Choice`
- `MultiSelect/Checkbox`
- `ReadPassword`
- `Collector` 和 `cparam`

## 文档

- GoDoc: https://pkg.go.dev/github.com/gookit/cliui/interact

## 安装

```shell
go get github.com/gookit/cliui/interact
```

## 新 UI 层

`interact/ui` 是一个新的抽象层，用于构建由 backend 驱动的交互组件。

- package: `github.com/gookit/cliui/interact/ui`
- 当前 backend: `github.com/gookit/cliui/interact/backend/plain`
- 事件驱动 backend: `github.com/gookit/cliui/interact/backend/readline`
- `readline.New()` 在非 TTY 输入下会回退到 `plain`
- `readline.NewStrict()` 在无法使用 TTY 时会直接返回错误，而不是回退
- `Input` 支持 UTF-8 行编辑和常用快捷键
- `Select` 和 `MultiSelect` 支持禁用项、默认值、导航键和可见的选择状态
- 详情：[interact-ui.md](interact-ui.md)

`interact` 包也提供了一组 bridge helper：

- `NewUIInput`
- `NewUIConfirm`
- `NewUISelect`
- `NewUIMultiSelect`
- `NewUIPlainBackend`
- `NewUIReadlineBackend`
- `NewUIStrictReadlineBackend`
- `NewUIFakeBackend`

如果测试或应用需要替换 `interact`、`interact/ui`、`show` 和 `progress` 共享的默认输入/输出流，可以使用根包 `github.com/gookit/cliui` 提供的辅助方法：

```go
cliui.CustomIO(in, out)
defer cliui.ResetIO()
```

## 快速示例

### Read Input

`ReadInput` 用于读取一行普通文本输入，适合简单参数、名称、路径等一次性输入。

```go
name, err := interact.ReadInput("Your name: ")
if err != nil {
	panic(err)
}
fmt.Println("name:", name)
```

效果示例：

```txt
Your name: tom
name: tom
```

### Prompt

`Prompt` 支持上下文控制和默认值，适合需要可取消、可超时或统一管理上下文的输入流程。

```go
answer, err := interact.Prompt(context.Background(), "Environment", "dev")
if err != nil {
	panic(err)
}
fmt.Println("env:", answer)
```

效果示例：

```txt
Environment [dev]: prod
env: prod
```

### Confirm

`Confirm` 用于确认类问题，返回布尔值，适合删除、覆盖、部署等需要用户确认的操作。

```go
if interact.Confirm("Continue? ", true) {
	fmt.Println("confirmed")
}
```

效果示例：

```txt
Continue? [Y/n] y
confirmed
```

### Question

`Question`/`Ask` 用于带默认值和可选校验的问题输入，适合声明式地组织单个问题。

```go
name := interact.Ask("Your name?", "guest", nil)
fmt.Println("name:", name)
```

需要配置或复用问题时，可以使用 `NewQuestion`：

```go
value := interact.NewQuestion("Your name?", "guest").Run()
fmt.Println(value.String())
```

效果示例：

```txt
Your name? [guest]: tom
tom
```

### Select

`Select` 用于从多个候选项中选择一个值，适合环境、区域、模板、操作类型等单选场景。

```go
city := interact.SelectOne(
	"Your city?",
	[]string{"chengdu", "beijing", "shanghai"},
	"",
)
fmt.Println("city:", city)
```

效果示例：

```txt
Your city?
  1) chengdu
  2) beijing
  3) shanghai
Please select: 1
city: chengdu
```

### Multi Select

`Multi Select` 用于选择多个值，适合批量启用模块、选择服务、选择标签等多选场景。

```go
services := interact.MultiSelect(
	"Choose services",
	[]string{"api", "worker", "web"},
	[]string{"api"},
)
fmt.Println("services:", services)
```

效果示例：

```txt
Choose services
  1) api
  2) worker
  3) web
Please select: 1,3
services: [api web]
```

需要同时获取选中项的 key 和 value 时，可以直接使用 `NewSelect`：

```go
s := interact.NewSelect("Choose env", []string{"dev", "prod"})
result := s.Run()
fmt.Println(result.KeyString(), result.String())
```

### Password

`ReadPassword` 用于读取敏感输入，终端中不会回显实际内容。

```go
password := interact.ReadPassword("Password: ")
fmt.Println("password length:", len(password))
```

效果示例：

```txt
Password:
password length: 8
```

### Collector

`Collector` 可以组合多个输入参数并按顺序执行：

```go
c := interact.NewCollector()
err := c.AddParams(
	cparam.NewStringParam("name", "Your name"),
	cparam.NewChoiceParam("env", "Choose env").WithChoices([]string{"dev", "prod"}),
)
if err != nil {
	panic(err)
}
```

效果示例：

```txt
Your name: tom
Choose env
  1) dev
  2) prod
Please select: 2
```

### UI Bridge

如果希望使用新的 `interact/ui` 组件，但不直接引入子包，可以使用 bridge helper：

```go
be := interact.NewUIReadlineBackend()

name, err := interact.NewUIInput("Your name").Run(context.Background(), be)
if err != nil {
	panic(err)
}

fmt.Println("name:", name)
```

效果示例：

```txt
Your name: tom
name: tom
```

### 完整 Select 示例

```go
package main

import (
	"fmt"

	"github.com/gookit/color"
	"github.com/gookit/cliui/interact"
)

func main() {
	color.Green.Println("This's An Select Demo")
	fmt.Println("----------------------------------------------------------")

	ans := interact.SelectOne(
		"Your city name(use string slice/array)?",
		[]string{"chengdu", "beijing", "shanghai"},
		"",
	)
	color.Info.Println("your select is:", ans)
	fmt.Println("----------------------------------------------------------")

	ans1 := interact.Choice(
		"Your age(use int slice/array)?",
		[]int{23, 34, 45},
		"",
	)
	color.Info.Println("your select is:", ans1)

	fmt.Println("----------------------------------------------------------")

	ans2 := interact.SingleSelect(
		"Your city name(use map)?",
		map[string]string{"a": "chengdu", "b": "beijing", "c": "shanghai"},
		"a",
	)
	color.Info.Println("your select is:", ans2)

	s := interact.NewSelect("Your city", []string{"chengdu", "beijing", "shanghai"})
	s.DefOpt = "2"
	r := s.Run()
	color.Info.Println("your select key:", r.K.String())
	color.Info.Println("your select val:", r.String())
}
```

预览：

![select](../images/select.png)

## 相关项目

- https://github.com/manifoldco/promptui
- https://github.com/chzyer/readline
