package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gookit/cliui/interact/backend/readline"
	"github.com/gookit/cliui/interact/ui"
)

func main() {
	items := make([]ui.Item, 20)
	for i := range items {
		items[i] = ui.Item{Key: string(rune('a' + i)), Label: "Other asset"}
	}
	items[0] = ui.Item{Key: "a", Label: "Asset 1"}

	selectUI := ui.NewSelect("Select asset", items)
	selectUI.Filterable = true
	selectUI.FilterPrompt = "Filter assets"
	selectUI.PageSize = 12

	result, err := selectUI.Run(context.Background(), readline.New())
	if err != nil {
		fmt.Fprintf(os.Stderr, "select: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("AFTER_SELECTION key=%s\n", result.Key)
}
