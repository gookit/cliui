package ui_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/gookit/cliui/interact/backend/readline"
	"github.com/gookit/cliui/interact/ui"
)

func Example_readlineFallback() {
	in := bytes.NewBufferString("\n")
	out := new(bytes.Buffer)

	be := readline.New()
	input := ui.NewInput("Your name")
	input.Default = "guest"

	name, err := input.RunWithIO(context.Background(), be, in, out)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(name)
	fmt.Println(strings.Contains(out.String(), "Your name"))
	// Output:
	// guest
	// true
}
