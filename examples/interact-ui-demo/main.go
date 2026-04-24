package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gookit/cliui/interact/backend/readline"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	ctx := context.Background()
	be := readline.New()

	nameInput := ui.NewInput("Your name")
	nameInput.Default = "guest"

	name, err := nameInput.Run(ctx, be)
	if err != nil {
		exitOnErr("input", err)
	}

	confirm := ui.NewConfirm("Continue", true)
	ok, err := confirm.Run(ctx, be)
	if err != nil {
		exitOnErr("confirm", err)
	}

	if !ok {
		fmt.Println("Canceled by user")
		return
	}

	selectOne := ui.NewSelect("Choose env", []ui.Item{
		{Key: "dev", Label: "Development", Value: "dev"},
		{Key: "test", Label: "Testing", Value: "test"},
		{Key: "prod", Label: "Production", Value: "prod"},
	})
	selectOne.DefaultKey = "dev"

	env, err := selectOne.Run(ctx, be)
	if err != nil {
		exitOnErr("select", err)
	}

	selectMany := ui.NewMultiSelect("Choose services", []ui.Item{
		{Key: "api", Label: "API", Value: "api"},
		{Key: "job", Label: "Job Worker", Value: "job"},
		{Key: "web", Label: "Web", Value: "web"},
	})
	selectMany.DefaultKeys = []string{"api", "web"}
	selectMany.MinSelected = 1

	services, err := selectMany.Run(ctx, be)
	if err != nil {
		exitOnErr("multiselect", err)
	}

	fmt.Println("Summary")
	fmt.Println("name:", name)
	fmt.Println("env:", env.Key)
	fmt.Println("services:", services.Keys)
}

func exitOnErr(stage string, err error) {
	fmt.Fprintf(os.Stderr, "%s error: %v\n", stage, err)
	os.Exit(1)
}
