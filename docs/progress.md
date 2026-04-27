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

Output preview:

```txt
[==============>-------------]  50%(55/110)
[============================] 100%(110/110)
```

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

```txt
 50%(50/100)
100%(100/100)
```

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

## Progress Bar

### Internal Widgets

 Widget Name | Usage example      | Description                               
-------------|--------------------|-------------------------------------------
 `max`       | `{@max}`           | Display max steps for progress bar        
 `current`   | `{@current}`       | Display current steps for progress bar    
 `percent`   | `{@percent:4s}`    | Display percent for progress run          
 `elapsed`   | `{@elapsed:7s}`    | Display has elapsed time for progress run 
 `remaining` | `{@remaining:7s}`  | Display remaining time                    
 `estimated` | `{@estimated:-7s}` | Display estimated time                    
 `memory`    | `{@memory:6s}`     | Display memory consumption size           

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

Example:

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

Quick create progress bar:

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
