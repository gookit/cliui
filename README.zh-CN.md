# CliUI

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/cliui?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/cliui)](https://github.com/gookit/cliui)
[![Unit-Tests](https://github.com/gookit/cliui/actions/workflows/go.yml/badge.svg)](https://github.com/gookit/cliui)

> [English](./README.md) | [简体中文](./README.zh-CN.md)

`gookit/cliui` 是一个终端 UI 辅助模块，为 CLI 应用提供常用的展示和交互组件。

它主要聚焦 CLI 应用中的三个核心场景：

- `show`：结构化终端内容展示
- `interact`：交互式输入
- `progress`：进度条和加载状态展示

根包还提供了一组共享输入/输出辅助方法，供各子包统一使用：

- `SetInput(in io.Reader)`
- `SetOutput(out io.Writer)`
- `CustomIO(in io.Reader, out io.Writer)`
- `ResetInput()`
- `ResetOutput()`
- `ResetIO()`

## 安装

```bash
go get github.com/gookit/cliui
```

也可以只安装某个子包：

```bash
go get github.com/gookit/cliui/show
go get github.com/gookit/cliui/progress
go get github.com/gookit/cliui/interact
```

## 包概览

### `show`

`show` 为 CLI 应用提供结构化终端输出能力，用于展示格式化内容。它包含 `table`、`title`、`banner`、`list`、`multi-list`、`alert` 消息和 `JSON` 输出等辅助组件。

导入：

```go
github.com/gookit/cliui/show
```

> [!NOTE]
> 更多详情：[docs/zh-CN/show.md](docs/zh-CN/show.md)

**快速使用**：

```go
package main

import "github.com/gookit/cliui/show"

func main() {
	show.Banner("Deploy started")
	show.AList("App info", map[string]any{
		"name": "cliui",
		"env":  "dev",
	})
	show.JSON(map[string]any{"ok": true})
}
```

### `interact`

`interact` 为 CLI 程序提供交互式输入能力。它支持 `prompt`、`confirm`、`question`、`select`、`multi-select`、`password` 输入以及其它常见终端交互模式。

它还包含新版 `interact/ui` 抽象层，用于基于 backend 驱动交互组件：

- `plain` backend：适用于行输入、测试和重定向 stdin
- `readline` backend：适用于真实终端中的 raw 模式交互
- `Input`、`Confirm`、`Select` 和 `MultiSelect` 组件
- 支持 UTF-8 输入编辑、常用导航按键和持续显示的校验错误

导入：

```go
github.com/gookit/cliui/interact
```

> [!NOTE]
> 更多详情：[docs/zh-CN/interact.md](docs/zh-CN/interact.md)

**快速使用**：

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact"
)

func main() {
	be := interact.NewUIReadlineBackend()

	name, err := interact.NewUIInput("Your name").Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	env, err := interact.NewUISelect("Choose env", []interact.UIItem{
		{Key: "dev", Label: "Development", Value: "dev"},
		{Key: "prod", Label: "Production", Value: "prod"},
	}).Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("name:", name)
	fmt.Println("env:", env.Value)
}
```

### `progress`

`progress` 为长时间运行的任务提供进度和加载状态展示能力。它包含 `progress bars`、`text bars`、`spinner/loading` 指示器、`counters`、`dynamic text` 输出，以及用于在一个终端区域中渲染多个进度条的 `MultiProgress`。

导入：

```go
github.com/gookit/cliui/progress
```

> [!NOTE]
> 更多详情：[docs/zh-CN/progress.md](docs/zh-CN/progress.md)

**快速使用**：

```go
package main

import (
	"time"

	"github.com/gookit/cliui/progress"
)

func main() {
	p := progress.Bar(100)
	p.Start()

	for i := 0; i < 100; i++ {
		time.Sleep(20 * time.Millisecond)
		p.Advance()
	}

	p.Finish()
}
```

## 迁移

如果你正在从 `gookit/gcli/v3` 迁移，对应的包路径如下：

- `github.com/gookit/gcli/v3/show` -> `github.com/gookit/cliui/show`
- `github.com/gookit/gcli/v3/interact` -> `github.com/gookit/cliui/interact`
- `github.com/gookit/gcli/v3/progress` -> `github.com/gookit/cliui/progress`

## 开发

```bash
go test ./...
```

## License

MIT
