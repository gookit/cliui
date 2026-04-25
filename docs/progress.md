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

Basic usage:

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

See package tests and exported constructors for more usage patterns.

## Multi Progress

Use `MultiProgress` when you need to render several `Progress` instances in one terminal block.

`NewMulti()` uses the shared `cutypes.Output` writer by default. You can override the output by setting `mp.Writer`.

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
