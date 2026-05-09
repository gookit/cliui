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
	maxSteps := int64(110)
	p := progress.Bar(maxSteps)
	p.Start()

	for i := int64(0); i < maxSteps; i++ {
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

### 渲染模式和 TTY 检测

`MultiProgress` 默认使用 `RenderDynamic`，通过 ANSI 控制符在同一个终端 block 中动态刷新多行。非交互终端、CI 日志或重定向输出中可以切换到 `RenderPlain`，避免输出 ANSI 控制符。

```go
mp := progress.NewMulti()
mp.Writer = os.Stderr
if !progress.IsTerminal(os.Stderr) {
	mp.RenderMode = progress.RenderPlain
}
```

也可以直接使用便捷方法：

```go
mp.UseAutoRenderMode()
```

可选模式：

- `RenderDynamic`：默认模式，适合交互终端，多行进度原地刷新。
- `RenderPlain`：输出普通文本行，不使用 ANSI；`Advance()` 不会高频刷屏，`Start()`、`Refresh()`、`Reset()`、`Progress.Finish()` 和状态 helper 会输出关键状态。
- `RenderDisabled`：完全关闭 progress 渲染，但 `Println()`、`Printf()`、`RunExclusive()` 仍会写出日志。

### 自动刷新和节流

`MultiProgress` 默认在托管 `Progress` 状态变化时同步刷新。下载、复制等高频更新场景可以开启 `AutoRefresh`，让 `Advance()`、`SetMessage()`、`Reset()` 等操作只标记 dirty，由后台 ticker 按 `RefreshInterval` 刷新。

```go
mp := progress.NewMulti()
mp.AutoRefresh = true
mp.RefreshInterval = 100 * time.Millisecond
mp.Start()
defer mp.Finish()
```

`RefreshInterval <= 0` 时会使用默认刷新间隔。`Finish()` 会停止后台刷新并最终渲染一次，结束后不会继续写终端。

即使没有开启 `AutoRefresh`，托管模式下也会尊重每个 `Progress` 的 `RedrawFreq`，避免每次 `Advance()` 都重绘整个多行区域。

### 复用固定 worker slot

批量任务可以创建固定数量的 progress bar，并在一个任务完成后用 `Reset()` 复用同一行。

```go
bar := mp.New()
bar.SetFormat("{@slot} {@name} {@percent}% {@phase}")
bar.SetMessage("slot", "#1")

bar.Reset(100)
bar.SetMessages(map[string]string{
	"name":  "fd",
	"phase": "downloading",
})
```

`Reset()` 会重置当前步数、百分比、开始时间和完成时间，但不会让 bar 离开 `MultiProgress`。需要一次性修改多个字段时可以使用 `ResetWith()`。

### 动态管理 progress bar

运行中可以隐藏、恢复或移除托管的 progress bar：

```go
mp.Hide(bar)   // 暂时不渲染，仍保留在 manager 中
mp.Show(bar)   // 恢复渲染
mp.Remove(bar) // 从 manager 列表移除，后续 bar 更新会被忽略
```

`Len()` 返回当前 manager 中仍保留的 bar 数量，`VisibleLen()` 返回当前可见 bar 数量。`Remove()` 后该 `Progress` 不会退化成 standalone 输出，迟到的 `Advance()`、`SetMessage()` 或 `Finish()` 调用会被忽略。

### 安全日志输出

多行 progress 正在渲染时，不要直接向同一个 writer 输出普通日志。使用 `RunExclusive()`、`Println()` 或 `Printf()` 可以先移开当前 progress block，输出日志后再恢复渲染。

```go
mp.Println("warning: fallback to single connection")
mp.Printf("package %s failed: %v\n", name, err)

mp.RunExclusive(func(w io.Writer) {
	fmt.Fprintf(w, "checksum verified with %s\n", sum)
})
```

### 状态 helper

`Done()`、`Fail()`、`Skip()` 可以用统一方式结束 progress，并更新 `{@status}` 和 `{@message}`：

```go
bar.SetFormat("{@name:-12s} {@status} {@percent:5s}%")
bar.Done()
bar.Fail("network failed")
bar.Skip()
```

默认 message 分别是 `done`、`failed`、`skipped`。`Done()` 会推进到最大进度，`Fail()` 和 `Skip()` 保持当前进度。

### 状态查询

`Progress` 提供：

- `Started() bool`
- `Finished() bool`
- `Step() int64`
- `Max() int64`
- `Percent() float32`

`MultiProgress` 提供：

- `Started() bool`
- `Finished() bool`
- `Len() int`
- `VisibleLen() int`

## Progress Bar

### 内置 Widgets

 Widget Name | Usage example      | Description
-------------|--------------------|----------------------------
 `max`       | `{@max}`           | 显示进度条最大步数
 `current`   | `{@current}`       | 显示当前进度步数
 `maxSize`   | `{@maxSize}`       | 将最大步数按字节大小显示
 `curSize`   | `{@curSize}`       | 将当前步数按字节大小显示
 `percent`   | `{@percent:4s}`    | 显示当前进度百分比
 `elapsed`   | `{@elapsed:7s}`    | 显示已耗时
 `remaining` | `{@remaining:7s}`  | 显示剩余时间
 `eta`       | `{@eta:7s}`        | `remaining` 的别名
 `estimated` | `{@estimated:-7s}` | 显示预计耗时
 `memory`    | `{@memory:6s}`     | 显示内存占用

`StepWidth` 控制 `{@current}` 的显示宽度。保持 `0` 时会按 `MaxSteps` 自动计算；只有需要固定当前进度数字列宽时才需要设置它。

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

格式 token 支持 Go 字符串格式宽度、左对齐和截断：

```go
bar.SetFormat("{@slot} {@name:-12s} {@name:.20s} {@percent:5s}%")
```

unknown token 会继续原样保留，便于发现拼写错误或保留给上层 renderer 处理。

示例：

```go
package main

import (
	"time"

	"github.com/gookit/cliui/progress"
)

// CustomBar create custom progress bar
func main() {
	maxSteps := int64(100)
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

	for i := int64(0); i < maxSteps; i++ {
		time.Sleep(80 * time.Millisecond)
		p.Advance()
	}

	p.Finish()
}
```

### IO Progress

`Progress` 可以直接用于按字节数跟踪 IO 进度。`Write` 会按写入到进度对象自身的字节数推进进度，`WrapReader` / `WrapWriter` 会按被包装对象实际读写的字节数推进进度。下载场景里 `http.Response.ContentLength` 是 `int64`，可以直接作为最大进度值传入。

```go
resp, err := http.Get(url)
if err != nil {
	return err
}
defer resp.Body.Close()

p := progress.Bar(resp.ContentLength)
p.Format = "{@bar} {@percent:4s}% {@curSize}/{@maxSize}"
p.Start()
defer p.Finish()

_, err = io.Copy(dst, p.WrapReader(resp.Body))
return err
```

### Progress Functions

快速创建进度条：

```text
func Bar(maxSteps ...int64) *Progress
func Counter(maxSteps ...int64) *Progress
func CustomBar(width int, cs BarChars, maxSteps ...int64) *Progress
func DynamicText(messages map[int]string, maxSteps ...int64) *Progress
func Full(maxSteps ...int64) *Progress
func LoadBar(chars []rune, maxSteps ...int64) *Progress
func LoadingBar(chars []rune, maxSteps ...int64) *Progress
func New(maxSteps ...int64) *Progress
func NewMulti() *MultiProgress
func NewWithConfig(fn func(p *Progress), maxSteps ...int64) *Progress
func IsTerminal(w io.Writer) bool
func RoundTrip(char rune, charNumAndBoxWidth ...int) *Progress
func RoundTripBar(char rune, charNumAndBoxWidth ...int) *Progress
func SpinnerBar(chars []rune, maxSteps ...int64) *Progress
func Tape(maxSteps ...int64) *Progress
func Txt(maxSteps ...int64) *Progress
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
