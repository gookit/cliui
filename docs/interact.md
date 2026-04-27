# interact

`interact` provides command-line interactive input helpers.

It includes common terminal interaction methods such as:

- `ReadInput`
- `ReadLine`
- `ReadFirst`
- `Prompt`
- `Confirm`
- `Query/Question/Ask`
- `Select/Choice`
- `MultiSelect/Checkbox`
- `ReadPassword`
- `Collector` and `cparam`

## Documentation

- GoDoc: https://pkg.go.dev/github.com/gookit/cliui/interact

## Install

```shell
go get github.com/gookit/cliui/interact
```

## New UI Layer

`interact/ui` is a new abstraction layer for backend-driven interaction components.

- package: `github.com/gookit/cliui/interact/ui`
- current backend: `github.com/gookit/cliui/interact/backend/plain`
- event-driven backend: `github.com/gookit/cliui/interact/backend/readline`
- `readline.New()` falls back to `plain` on non-TTY input
- `readline.NewStrict()` returns an error instead of falling back when a TTY is unavailable
- `Input` supports UTF-8 line editing and common shortcuts
- `Select` and `MultiSelect` support disabled items, defaults, navigation keys and visible selection status
- details: [interact-ui.md](interact-ui.md)

Bridge helpers are also available from the `interact` package:

- `NewUIInput`
- `NewUIConfirm`
- `NewUISelect`
- `NewUIMultiSelect`
- `NewUIPlainBackend`
- `NewUIReadlineBackend`
- `NewUIStrictReadlineBackend`
- `NewUIFakeBackend`

Use the root package helpers from `github.com/gookit/cliui` when tests or applications need to replace the default streams shared by `interact`, `interact/ui`, `show` and `progress`:

```go
cliui.CustomIO(in, out)
defer cliui.ResetIO()
```

## Quick Examples

### Read Input

`ReadInput` reads one plain text line. It is suitable for simple values such as names, paths, and short parameters.

```go
name, err := interact.ReadInput("Your name: ")
if err != nil {
	panic(err)
}
fmt.Println("name:", name)
```

Output preview:

```txt
Your name: tom
name: tom
```

### Prompt

`Prompt` supports context control and a default value, making it useful when input needs cancellation, timeout, or shared context handling.

```go
answer, err := interact.Prompt(context.Background(), "Environment", "dev")
if err != nil {
	panic(err)
}
fmt.Println("env:", answer)
```

Output preview:

```txt
Environment [dev]: prod
env: prod
```

### Confirm

`Confirm` asks a yes/no question and returns a boolean. It is useful before delete, overwrite, deploy, and other confirmation-sensitive actions.

```go
if interact.Confirm("Continue? ", true) {
	fmt.Println("confirmed")
}
```

Output preview:

```txt
Continue? [Y/n] y
confirmed
```

### Question

`Question`/`Ask` handles a single question with a default value and optional validation.

```go
name := interact.Ask("Your name?", "guest", nil)
fmt.Println("name:", name)
```

Use `NewQuestion` when you need to configure or reuse a question:

```go
value := interact.NewQuestion("Your name?", "guest").Run()
fmt.Println(value.String())
```

Output preview:

```txt
Your name? [guest]: tom
tom
```

### Select

`Select` chooses one value from a list. It is useful for environments, regions, templates, and operation types.

```go
city := interact.SelectOne(
	"Your city?",
	[]string{"chengdu", "beijing", "shanghai"},
	"",
)
fmt.Println("city:", city)
```

Output preview:

```txt
Your city?
  1) chengdu
  2) beijing
  3) shanghai
Please select: 1
city: chengdu
```

### Multi Select

`Multi Select` chooses multiple values. It is useful for enabling modules, selecting services, or choosing tags.

```go
services := interact.MultiSelect(
	"Choose services",
	[]string{"api", "worker", "web"},
	[]string{"api"},
)
fmt.Println("services:", services)
```

Output preview:

```txt
Choose services
  1) api
  2) worker
  3) web
Please select: 1,3
services: [api web]
```

Use `NewSelect` directly when you need the selected key and value:

```go
s := interact.NewSelect("Choose env", []string{"dev", "prod"})
result := s.Run()
fmt.Println(result.KeyString(), result.String())
```

### Password

`ReadPassword` reads sensitive input without echoing the actual value in the terminal.

```go
password := interact.ReadPassword("Password: ")
fmt.Println("password length:", len(password))
```

Output preview:

```txt
Password:
password length: 8
```

### Collector

`Collector` groups several input parameters and runs them in order:

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

Output preview:

```txt
Your name: tom
Choose env
  1) dev
  2) prod
Please select: 2
```

### UI Bridge

Use the bridge helpers when you want the new `interact/ui` components without importing subpackages directly:

```go
be := interact.NewUIReadlineBackend()

name, err := interact.NewUIInput("Your name").Run(context.Background(), be)
if err != nil {
	panic(err)
}

fmt.Println("name:", name)
```

Output preview:

```txt
Your name: tom
name: tom
```

### Full Select Example

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

Preview:

![](../examples/images/select.png)

## Related

- https://github.com/manifoldco/promptui
- https://github.com/chzyer/readline
