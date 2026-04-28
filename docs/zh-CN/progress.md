# progress

`progress` 提供终端进度和加载状态展示辅助能力。

它支持 `Txt`、`Bar`、`Loading`、`RoundTrip` 和 `DynamicText` 等进度输出模式。

- progress bar
- text progress bar
- pending/loading progress bar
- counter 计数
- dynamic text 动态文本
- spinner: loading, round-trip
- multi progress rendering

## 文档

- GoDoc: https://pkg.go.dev/github.com/gookit/cliui/progress

## 安装

```bash
go get github.com/gookit/cliui/progress
```

## 快速示例

### Bar

`Bar` 是默认的进度条组件，用于展示一个有明确总步数的任务进度，例如下载、构建、批量处理等。

```go
package main

import (
	"time"

	"github.com/gookit/cliui/progress"
)

func main() {
	speed := 100
	maxSteps := 110
	p := progress.Bar(maxSteps)
	p.Start()

	for i := 0; i < maxSteps; i++ {
		time.Sleep(time.Duration(speed) * time.Millisecond)
		p.Advance()
	}

	p.Finish()
}
```

效果示例：

```txt
[==============>-------------]  50%(55/110)
[============================] 100%(110/110)
```

![prog-bar](../images/prog-bar.svg)

### Full

`Full` 会在进度条旁显示更完整的运行信息，包括当前进度、耗时、预计耗时和内存占用。

```go
p := progress.Full(100)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

效果示例：

```txt
100%(100/100)  2.1s/ 2.1s  8.3MB
```

### Text Progress

`Txt` 使用纯文本百分比展示进度，不渲染条形图，适合日志密集或窄终端环境。

```go
p := progress.Txt(100)
p.Start()
p.AdvanceTo(50)
p.Finish()
```

效果示例：

![prog-txt](../images/prog-txt.svg)

### Counter

`Counter` 只展示当前计数，适合总量较小或只关心完成数量的任务。

```go
p := progress.Counter(3)
p.Start()
p.Advance()
p.Advance()
p.Advance()
p.Finish()
```

效果示例：

```txt
1
2
3 Files
```

### Dynamic Text

`DynamicText` 会在进度到达指定节点时切换提示文本，适合展示任务阶段。

```go
p := progress.DynamicText(map[int]string{
	10: " prepare",
	50: " build",
	90: " finish",
}, 100)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

效果示例：

```txt
 10%(10/100) prepare
 50%(50/100) build
 90%(90/100) finish
```

### Loading Bar

`LoadingBar` 使用一组字符循环显示等待状态，适合总进度未知但仍需要持续反馈的任务。

```go
p := progress.LoadingBar([]rune{'-', '\\', '|', '/'}, 20)
p.Start()
p.AdvanceTo(20)
p.Finish()
```

效果示例：

```txt
[-]
[\]
[|]
[/]
```

### Round Trip

`RoundTrip` 会让指定字符在固定宽度区域内往返移动，适合展示未确定结束时间的活动状态。

```go
p := progress.RoundTrip('=', 10, 30)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

效果示例：

```txt
[==========                    ]
[     ==========               ]
[                    ==========]
```

### Tape

`Tape` 使用类似连续滚动的条带效果展示进度，适合希望输出更轻量动画感的场景。

```go
p := progress.Tape(100)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

效果示例：

```txt
[====>                         ]  20%(20/100)
[===============>              ]  60%(60/100)
[==============================] 100%(100/100)
```

### Custom Bar

`CustomBar` 允许指定宽度和字符样式，适合需要和应用视觉风格保持一致的终端输出。

```go
p := progress.CustomBar(40, progress.RandomBarStyle(), 100)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

效果示例：

```txt
	1 [->--------------------------]
	3 [■■■>------------------------]
25/50 [==============>-------------]  50%
```

## Spinner

spinner 简单快速的显示一个加载动画。与 progress 不同的是，它不需要在中途更新进度，而是直接启动直到调用 Stop。

### Loading Spinner

让设置的多个字符不停的变化，显示出旋转的效果。

```go
sp := progress.LoadingSpinner([]rune{'-', '\\', '|', '/'}, 100*time.Millisecond)
sp.Start("loading")
time.Sleep(time.Second) // do someting ...
sp.Stop("done")
```

效果示例：

```txt
- loading
\ loading
| loading
/ loading
done
```

### Round Trip Spinner

设置的字符在盒子里来回移动。 eg: `[ =  ]`

```go
sp := progress.RoundTripSpinner('=', 100*time.Millisecond, 10, 30)
sp.Start("loading")
time.Sleep(time.Second) // do someting ...
sp.Stop("done")
```

效果示例：

```txt
[===       ] loading
[   ===    ] loading
[       ===] loading
done
```

更多用法可以参考包测试和导出的构造函数。

## Multi Progress

当需要在一个终端区域中渲染多个 `Progress` 实例时，可以使用 `MultiProgress`。

`NewMulti()` 默认使用共享的 `cutypes.Output` writer。你可以通过设置 `mp.Writer` 覆盖输出目标。

输出效果会固定在一个终端区域内刷新多行，每个任务保留自己的进度状态。

```go
package main

import (
	"time"

	"github.com/gookit/cliui/progress"
)

func main() {
	mp := progress.NewMulti()
	// mp.Writer = customWriter

	build := mp.New(100)
	build.AddMessage("message", " build")

	test := mp.New(80)
	test.AddMessage("message", " test")

	mp.Start()

	for i := 0; i < 100; i++ {
		time.Sleep(20 * time.Millisecond)
		build.Advance()

		if i < 80 {
			test.Advance()
		}
	}

	mp.Finish()
}
```

效果示例：

```txt
[==============>-------------]  50%(50/100) build
[=================>----------]  62%(50/80)  test
```

## Progress Bar

### 内置 Widgets

 Widget Name | Usage example      | Description
-------------|--------------------|----------------------------
 `max`       | `{@max}`           | 显示进度条最大步数
 `current`   | `{@current}`       | 显示当前进度步数
 `percent`   | `{@percent:4s}`    | 显示当前进度百分比
 `elapsed`   | `{@elapsed:7s}`    | 显示已耗时
 `remaining` | `{@remaining:7s}`  | 显示剩余时间
 `estimated` | `{@estimated:-7s}` | 显示预计耗时
 `memory`    | `{@memory:6s}`     | 显示内存占用

### 自定义进度条

可以自定义进度条渲染格式。内置格式包括：

```go
// txt bar
MinFormat  = "{@message}{@current}"
TxtFormat  = "{@message}{@percent:4s}%({@current}/{@max})"
DefFormat  = "{@message}{@percent:4s}%({@current}/{@max})"
FullFormat = "{@percent:4s}%({@current}/{@max}) {@elapsed:7s}/{@estimated:-7s} {@memory:6s}"

// bar

DefBarFormat  = "{@bar} {@percent:4s}%({@current}/{@max}){@message}"
FullBarFormat = "{@bar} {@percent:4s}%({@current}/{@max}) {@elapsed:7s}/{@estimated:-7s} {@memory:6s}"
```

示例：

```go
package main

import (
	"time"

	"github.com/gookit/cliui/progress"
)

// CustomBar create custom progress bar
func main() {
	maxSteps := 100
	// use special bar style: [==============>-------------]
	// barStyle := progress.BarStyles[0]
	// use random bar style
	barStyle := progress.RandomBarStyle()

	p := progress.New(maxSteps)
	p.Config(func(p *progress.Progress) {
		p.Format = progress.DefBarFormat
	})
	p.AddWidget("bar", progress.BarWidget(60, barStyle))

	p.Start()

	for i := 0; i < maxSteps; i++ {
		time.Sleep(80 * time.Millisecond)
		p.Advance()
	}

	p.Finish()
}
```

### Progress Functions

快速创建进度条：

```text
func Bar(maxSteps ...int) *Progress
func Counter(maxSteps ...int) *Progress
func CustomBar(width int, cs BarChars, maxSteps ...int) *Progress
func DynamicText(messages map[int]string, maxSteps ...int) *Progress
func Full(maxSteps ...int) *Progress
func LoadBar(chars []rune, maxSteps ...int) *Progress
func LoadingBar(chars []rune, maxSteps ...int) *Progress
func New(maxSteps ...int) *Progress
func NewMulti() *MultiProgress
func NewWithConfig(fn func(p *Progress), maxSteps ...int) *Progress
func RoundTrip(char rune, charNumAndBoxWidth ...int) *Progress
func RoundTripBar(char rune, charNumAndBoxWidth ...int) *Progress
func SpinnerBar(chars []rune, maxSteps ...int) *Progress
func Tape(maxSteps ...int) *Progress
func Txt(maxSteps ...int) *Progress
```

`Progress.Line()` 会返回当前渲染行，主要供 `MultiProgress` 管理渲染使用；在将进度输出嵌入其它 renderer 时也可能有用。

## Spinner Bar

### Spinner Functions

快速创建 spinner：

```text
func LoadingSpinner(chars []rune, speed time.Duration) *SpinnerFactory
func RoundTripLoading(char rune, speed time.Duration, charNumAndBoxWidth ...int) *SpinnerFactory
func RoundTripSpinner(char rune, speed time.Duration, charNumAndBoxWidth ...int) *SpinnerFactory
func Spinner(speed time.Duration) *SpinnerFactory
```

## 相关项目

- https://github.com/vbauerster/mpb
- https://github.com/schollz/progressbar
- https://github.com/gosuri/uiprogress
