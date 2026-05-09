# progress

`progress` provides terminal progress and loading display helpers.

It supports progress output patterns such as `Txt`, `Bar`, `Loading`, `RoundTrip` and `DynamicText`.

- progress bar
- text progress bar
- pending/loading progress bar
- counter
- dynamic text
- multi progress rendering

## Documentation

- GoDoc: https://pkg.go.dev/github.com/gookit/cliui/progress

## Install

```bash
go get github.com/gookit/cliui/progress
```

## Quick Example

### Bar

`Bar` is the default progress bar component for tasks with a known total, such as downloads, builds, or batch processing.

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

Output preview:

![prog-bar](images/prog-bar.svg)

### Full

`Full` shows a richer progress line with current progress, elapsed time, estimated time, and memory usage.

```go
p := progress.Full(100)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

Output preview:

```txt
100%(100/100)  2.1s/ 2.1s  8.3MB
```

### Text Progress

`Txt` displays progress as text without a bar. It is useful for narrow terminals or log-heavy output.

```go
p := progress.Txt(100)
p.Start()
p.AdvanceTo(50)
p.Finish()
```

Output preview:

![prog-txt](images/prog-txt.svg)

### Counter

`Counter` displays only the current count. It is useful when the total is small or the completed count matters more than a bar.

```go
p := progress.Counter(3)
p.Start()
p.Advance()
p.Advance()
p.Advance()
p.Finish()
```

Output preview:

```txt
1
2
3
```

### Dynamic Text

`DynamicText` changes the message when progress reaches configured points. It is useful for showing task phases.

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

Output preview:

```txt
 10%(10/100) prepare
 50%(50/100) build
 90%(90/100) finish
```

### Loading Bar

`LoadingBar` cycles through a set of characters to show activity when exact progress is not available.

```go
p := progress.LoadingBar([]rune{'-', '\\', '|', '/'}, 20)
p.Start()
p.AdvanceTo(20)
p.Finish()
```

Output preview:

```txt
[-]
[\]
[|]
[/]
```

### Round Trip

`RoundTrip` moves the given character back and forth inside a fixed-width area. It is useful for indeterminate activity states.

```go
p := progress.RoundTrip('=', 10, 30)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

Output preview:

```txt
[==========                    ]
[     ==========               ]
[                    ==========]
```

### Tape

`Tape` uses a rolling tape-like bar effect to show progress with lightweight animation.

```go
p := progress.Tape(100)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

Output preview:

```txt
[====>                         ]  20%(20/100)
[===============>              ]  60%(60/100)
[==============================] 100%(100/100)
```

### Custom Bar

`CustomBar` lets you choose the width and character style. Use it when terminal output should match your application's visual style.

```go
p := progress.CustomBar(40, progress.RandomBarStyle(), 100)
p.Start()
p.AdvanceTo(100)
p.Finish()
```

## Spinner

### Loading Spinner

```go
sp := progress.LoadingSpinner([]rune{'-', '\\', '|', '/'}, 100*time.Millisecond)
sp.Start("loading")
time.Sleep(time.Second)
sp.Stop("done")
```

Output preview:

```txt
- loading
\ loading
| loading
/ loading
done
```

### Round Trip Spinner

`RoundTripSpinner` shows a moving character group inside a box until `Stop` is called.

```go
sp := progress.RoundTripSpinner('=', 100*time.Millisecond, 10, 30)
sp.Start("loading")
time.Sleep(time.Second)
sp.Stop("done")
```

Output preview:

```txt
[===       ] loading
[   ===    ] loading
[       ===] loading
done
```

See package tests and exported constructors for more usage patterns.

## Multi Progress

Use `MultiProgress` when you need to render several `Progress` instances in one terminal block.

`NewMulti()` uses the shared `cutypes.Output` writer by default. You can override the output by setting `mp.Writer`.

The output is refreshed in a fixed terminal block, while every task keeps its own progress state.

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

Output preview:

```txt
[==============>-------------]  50%(50/100) build
[=================>----------]  62%(50/80)  test
```

### Render Modes And TTY Detection

`MultiProgress` uses `RenderDynamic` by default. It refreshes several lines in one terminal block with ANSI control sequences. For non-interactive terminals, CI logs, or redirected output, use `RenderPlain` to avoid ANSI output.

```go
mp := progress.NewMulti()
mp.Writer = os.Stderr
if !progress.IsTerminal(os.Stderr) {
	mp.RenderMode = progress.RenderPlain
}
```

You can also use the convenience method:

```go
mp.UseAutoRenderMode()
```

Available modes:

- `RenderDynamic`: default mode for interactive terminals, refreshing a multi-line progress block in place.
- `RenderPlain`: writes plain text lines without ANSI controls. `Advance()` does not print on every update; `Start()`, `Refresh()`, `Reset()`, `Progress.Finish()`, and status helpers print key states.
- `RenderDisabled`: disables progress rendering, while `Println()`, `Printf()`, and `RunExclusive()` still write logs.

### Auto Refresh And Throttling

By default, `MultiProgress` refreshes synchronously when a managed `Progress` changes state. For high-frequency updates such as downloads or copies, enable `AutoRefresh` so `Advance()`, `SetMessage()`, `Reset()`, and similar operations only mark the manager dirty. A background ticker refreshes the block at `RefreshInterval`.

```go
mp := progress.NewMulti()
mp.AutoRefresh = true
mp.RefreshInterval = 100 * time.Millisecond
mp.Start()
defer mp.Finish()
```

When `RefreshInterval <= 0`, the default refresh interval is used. `Finish()` stops the background refresh loop and renders one final time, so the manager does not keep writing after it is finished.

Managed progress bars also respect each bar's `RedrawFreq` when `AutoRefresh` is disabled, avoiding a full multi-line redraw on every `Advance()`.

### Reusable Worker Slots

Batch tasks can create a fixed number of progress bars and reuse each line for the next task with `Reset()`.

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

`Reset()` clears the current step, percent, start time, and finish time, but keeps the bar attached to its `MultiProgress`. Use `ResetWith()` when several fields should be changed in one managed update.

### Dynamic Bar Management

You can hide, show, or remove managed progress bars while the manager is running:

```go
mp.Hide(bar)   // stop rendering this bar, but keep it managed
mp.Show(bar)   // render it again
mp.Remove(bar) // remove it from the manager; later bar updates are ignored
```

`Len()` returns the number of bars still held by the manager, and `VisibleLen()` returns the number currently rendered. After `Remove()`, the `Progress` does not fall back to standalone output; late `Advance()`, `SetMessage()`, or `Finish()` calls are ignored.

### Exclusive Logging

Do not write normal logs directly to the same writer while a multi-line progress block is active. Use `RunExclusive()`, `Println()`, or `Printf()` to clear the current progress block, write the log, and redraw the block.

```go
mp.Println("warning: fallback to single connection")
mp.Printf("package %s failed: %v\n", name, err)

mp.RunExclusive(func(w io.Writer) {
	fmt.Fprintf(w, "checksum verified with %s\n", sum)
})
```

### Status Helpers

`Done()`, `Fail()`, and `Skip()` provide a consistent way to finish a progress bar and update `{@status}` and `{@message}`:

```go
bar.SetFormat("{@name:-12s} {@status} {@percent:5s}%")
bar.Done()
bar.Fail("network failed")
bar.Skip()
```

The default messages are `done`, `failed`, and `skipped`. `Done()` advances to the maximum progress, while `Fail()` and `Skip()` keep the current step.

### State Queries

`Progress` provides:

- `Started() bool`
- `Finished() bool`
- `Step() int64`
- `Max() int64`
- `Percent() float32`

`MultiProgress` provides:

- `Started() bool`
- `Finished() bool`
- `Len() int`
- `VisibleLen() int`

## Progress Bar

### Internal Widgets

 Widget Name | Usage example      | Description
-------------|--------------------|-------------------------------------------
 `max`       | `{@max}`           | Display max steps for progress bar
 `current`   | `{@current}`       | Display current steps for progress bar
 `maxSize`   | `{@maxSize}`       | Display max steps as a byte size
 `curSize`   | `{@curSize}`       | Display current steps as a byte size
 `percent`   | `{@percent:4s}`    | Display percent for progress run
 `elapsed`   | `{@elapsed:7s}`    | Display has elapsed time for progress run
 `remaining` | `{@remaining:7s}`  | Display remaining time
 `eta`       | `{@eta:7s}`        | Alias of `remaining`
 `estimated` | `{@estimated:-7s}` | Display estimated time
 `memory`    | `{@memory:6s}`     | Display memory consumption size

`StepWidth` controls the display width of `{@current}`. Leave it as `0` to auto-size from `MaxSteps`; set it only when you need a fixed current-count column.

### Custom Progress Bar

You can customize the progress bar render format. Built-in formats include:

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

Format tokens support Go string width, left alignment, and truncation:

```go
bar.SetFormat("{@slot} {@name:-12s} {@name:.20s} {@percent:5s}%")
```

Unknown tokens are still preserved as-is, which helps catch typos or reserve them for a higher-level renderer.

Example:

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

`Progress` can track byte counts for IO flows. `Write` advances by the number of bytes written to the progress itself, and `WrapReader` / `WrapWriter` advance by the actual bytes read or written on the wrapped object. This is useful for downloads where `http.Response.ContentLength` is an `int64`.

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

Quick create progress bar:

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

`Progress.Line()` returns the current rendered line for managed rendering. It is mainly used by `MultiProgress`, but can also be useful when embedding progress output in another renderer.

## Spinner Bar

### Spinner Functions

Quick create progress spinner:

```text
func LoadingSpinner(chars []rune, speed time.Duration) *SpinnerFactory
func RoundTripLoading(char rune, speed time.Duration, charNumAndBoxWidth ...int) *SpinnerFactory
func RoundTripSpinner(char rune, speed time.Duration, charNumAndBoxWidth ...int) *SpinnerFactory
func Spinner(speed time.Duration) *SpinnerFactory
```

## Related

- https://github.com/vbauerster/mpb
- https://github.com/schollz/progressbar
- https://github.com/gosuri/uiprogress
