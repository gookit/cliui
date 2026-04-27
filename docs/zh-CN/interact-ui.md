# interact/ui

`interact/ui` 为终端组件提供新的交互抽象层。

它面向 backend 驱动的交互流程，目前包含：

- `Input`
- `Confirm`
- `Select`
- `MultiSelect`

当前状态：

- 组件模型和结果类型已稳定
- 已提供 `plain` backend
- 已提供 raw-terminal `readline` backend
- 现有 `interact` 包 API 保持不变
- `Input` 支持 UTF-8 文本编辑和常用行编辑快捷键
- 选择组件会在下一次用户输入前持续显示校验错误
- 暂未实现过滤和 resize 事件

## 包

- `github.com/gookit/cliui/interact/ui`
- `github.com/gookit/cliui/interact/backend`
- `github.com/gookit/cliui/interact/backend/plain`
- `github.com/gookit/cliui/interact/backend/readline`
- `github.com/gookit/cliui/interact/backend/fake`

## 安装

```bash
go get github.com/gookit/cliui/interact/ui
```

## 快速示例

### Input

`Input` 用于读取一行文本。配合 `readline` backend 时支持光标移动、删除、UTF-8 文本编辑等真实终端体验；配合 `plain` backend 时按行读取输入。

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := plain.New()

	input := ui.NewInput("Your name")
	input.Default = "guest"

	name, err := input.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("name:", name)
}
```

最简形式：

```go
be := plain.New()
name, err := ui.NewInput("Your name").Run(context.Background(), be)
```

效果示例：

```txt
Your name [guest]: tom
name: tom
```

### Confirm

`Confirm` 用于二选一确认。它会显示默认选项，用户可以输入 `y/n`，也可以在 `readline` backend 中用方向键切换。

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := plain.New()

	confirm := ui.NewConfirm("Continue", true)
	ok, err := confirm.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("confirmed:", ok)
}
```

最简形式：

```go
be := plain.New()
ok, err := ui.NewConfirm("Continue", true).Run(context.Background(), be)
```

效果示例：

```txt
Continue [Y/n]: y
confirmed: true
```

### Select

`Select` 用于从 `Item` 列表中选择一个结果。每个选项都可以包含稳定的 `Key`、展示用 `Label` 和业务侧使用的 `Value`。

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := plain.New()

	selectOne := ui.NewSelect("Choose env", []ui.Item{
		{Key: "dev", Label: "Development", Value: "dev"},
		{Key: "prod", Label: "Production", Value: "prod"},
	})
	selectOne.DefaultKey = "dev"

	result, err := selectOne.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("selected:", result.Key, result.Value)
}
```

最简形式：

```go
be := plain.New()
result, err := ui.NewSelect("Choose env", []ui.Item{
	{Key: "dev", Label: "Development", Value: "dev"},
	{Key: "prod", Label: "Production", Value: "prod"},
}).Run(context.Background(), be)
```

效果示例：

```txt
Choose env
> dev   Development
  prod  Production
selected: dev dev
```

### MultiSelect

`MultiSelect` 用于从 `Item` 列表中选择多个结果。它支持默认选中项、最少选择数量和禁用项，适合模块、服务、标签等多选配置。

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/plain"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := plain.New()

	selectMany := ui.NewMultiSelect("Choose services", []ui.Item{
		{Key: "api", Label: "API", Value: "api"},
		{Key: "job", Label: "Job Worker", Value: "job"},
		{Key: "web", Label: "Web", Value: "web"},
	})
	selectMany.DefaultKeys = []string{"api", "web"}
	selectMany.MinSelected = 1

	result, err := selectMany.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("selected keys:", result.Keys)
}
```

最简形式：

```go
be := plain.New()
result, err := ui.NewMultiSelect("Choose services", []ui.Item{
	{Key: "api", Label: "API", Value: "api"},
	{Key: "web", Label: "Web", Value: "web"},
}).Run(context.Background(), be)
```

效果示例：

```txt
Choose services
> [x] api  API
  [ ] job  Job Worker
  [x] web  Web
selected keys: [api web]
```

### Readline Backend

如果需要在真实终端中进行事件驱动交互，可以使用 `readline` backend。

`readline.New()` 会在 stdin 不是 TTY 时自动回退到 `plain` backend。
如果希望严格要求 TTY，可以使用 `readline.NewStrict()`。

```go
package main

import (
	"context"
	"fmt"

	"github.com/gookit/cliui/interact/backend/readline"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	be := readline.New()

	input := ui.NewInput("Your name")
	input.Default = "guest"

	name, err := input.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	selectOne := ui.NewSelect("Choose env", []ui.Item{
		{Key: "dev", Label: "Development", Value: "dev"},
		{Key: "prod", Label: "Production", Value: "prod"},
	})
	selectOne.DefaultKey = "dev"

	env, err := selectOne.Run(context.Background(), be)
	if err != nil {
		panic(err)
	}

	fmt.Println("name:", name)
fmt.Println("env:", env.Key)
}
```

效果示例：

```txt
Your name [guest]: tom
Choose env
> dev   Development
  prod  Production
name: tom
env: dev
```

### Fake Backend

测试中可以使用 `fake` backend 注入标准化事件，而无需真实终端：

```go
be := fake.New(
	backend.Event{Type: backend.EventKey, Text: "tom"},
	backend.Event{Type: backend.EventKey, Key: backend.KeyEnter},
)

name, err := ui.NewInput("Your name").Run(context.Background(), be)
```

效果示例：

```txt
Your name: tom
```

## 说明

- `plain` backend 使用基于行的输入，可用于普通 stdin/stdout、测试和重定向输入。
- `plain` backend 不提供逐键导航；select 组件接收 item key，multi-select 接收逗号分隔的 item key。
- `readline` backend 使用 raw terminal 输入，支持 UTF-8 文本、方向键、Home/End、Delete、Backspace、Tab、Shift+Tab、PageUp/PageDown、Space、Enter、Esc 和 Ctrl+C。
- `readline.New()` 在没有真实终端时会回退到 `plain`。
- `readline.NewStrict()` 在 stdin 不是真实终端时会返回错误。
- `fake` backend 用于可确定的组件测试。
- `Select` 使用 item key 进行单选。
- `MultiSelect` 使用逗号分隔的 item key。
- 当前交互被取消时会返回 `ErrAborted`。
- `Select` 和 `MultiSelect` 支持禁用项和默认值。
- `Select` 会在专门的状态行显示当前高亮项。
- `MultiSelect` 会显示当前高亮项和已选 key 摘要。
- 校验和选择错误会持续显示，直到下一次输入事件改变组件状态。
- `Input` 的光标位置会考虑显示宽度，因此在支持的终端中可以正确处理 CJK 文本。
- backend 模型中定义了终端 resize 事件，但当前 backend 尚未发出该事件。

## 按键绑定

当前 `readline` backend 支持：

- `Input`：输入文本进行插入，`Left/Right` 移动，`Home/End` 或 `Ctrl+A/Ctrl+E` 跳转，`Backspace/Delete` 编辑，`Ctrl+U` 删除光标前内容，`Ctrl+K` 删除光标后内容，`Ctrl+W` 删除前一个单词，`Enter` 提交
- `Confirm`：`Left/Right` 切换，`y/n` 选择，`Enter` 提交当前值
- `Select`：`Up/Down` 或 `Tab/Shift+Tab` 移动，`PageUp/PageDown` 跳转，`Enter` 确认，也可以直接输入 item key；视图中也会显示当前 item 摘要
- `MultiSelect`：`Up/Down` 或 `Tab/Shift+Tab` 移动，`PageUp/PageDown` 跳转，`Space` 切换，`Enter` 确认；视图中也会显示当前 item 和已选 key 摘要

## Backend 行为

### plain

`plain` 面向非 TTY 输入、测试、重定向 stdin 和简单命令行流程。它一次读取一行：

- `Input` 将提交的行作为输入值。
- `Confirm` 接收 `yes/no`、`y/n`，或空行表示默认值。
- `Select` 接收 item key，空行表示使用默认 key。
- `MultiSelect` 接收逗号分隔的 item key，空行表示使用默认 key 列表。

### readline

`readline` 面向真实终端交互。它会为 UI 组件标准化终端事件：

- UTF-8 输入按 rune 读取，而不是按原始字节读取。
- 常见 CSI 和 SS3 转义序列会映射为方向键、Home/End、Delete、Shift+Tab 和 PageUp/PageDown。
- `Esc` 和 `Ctrl+C` 会用 `ErrAborted` 取消当前组件。
- `readline.New()` 在非 TTY 环境中会自动回退到 `plain`；如需严格 TTY 行为，使用 `readline.NewStrict()`。

## 当前限制

- `Select` 和 `MultiSelect` 暂未实现过滤/搜索。
- resize 事件已建模，但当前 backend 尚未发出。
- 延迟过长的转义序列可能会被视为单独的 `Esc`；常见终端按键序列会在输入缓冲可用时处理。

## Demo

运行交互式 demo 命令：

```bash
go run ./examples/interact-ui-demo
```

如果在非 TTY 环境中运行，`readline.New()` 会自动回退到 `plain` backend。

## 示例测试

运行包示例，包括 `readline` fallback 示例：

```bash
go test ./interact/ui -run Example -v
```

## 下一步

当前抽象已经可以支持后续更丰富的事件驱动 backend、resize 处理和选择过滤能力，并且不需要改变 `ui` 包的表层 API。
