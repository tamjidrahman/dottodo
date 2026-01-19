package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tamjidrahman/dottodo/internal/cli"
	"github.com/tamjidrahman/dottodo/internal/storage"
	"github.com/tamjidrahman/dottodo/internal/tui"
)

func main() {
	// Initialize storage
	store, err := storage.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading todos: %v\n", err)
		os.Exit(1)
	}

	// If there are command line arguments, run CLI mode
	if len(os.Args) > 1 {
		c := cli.New(store)
		if err := c.Run(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run TUI mode
	model := tui.New(store)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
