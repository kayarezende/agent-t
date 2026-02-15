package main

import (
	"fmt"
	"os"

	"agent-t/internal/config"
	"agent-t/internal/launcher"
	"agent-t/internal/scanner"
	"agent-t/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	projects, err := scanner.Scan(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", cwd, err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Fprintf(os.Stderr, "No project folders found in %s\n", cwd)
		os.Exit(1)
	}

	m := tui.NewModel(projects, cfg, cwd)
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	final := result.(tui.Model)

	if final.Cancelled() {
		os.Exit(0)
	}

	// Save config if presets were added
	if final.ConfigChanged() {
		if err := config.Save(final.Config()); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save config: %v\n", err)
		}
	}

	// Launch terminals
	fmt.Printf("Launching %dx%d terminals in %s...\n",
		final.SelectedLayout().Cols, final.SelectedLayout().Rows,
		final.SelectedProject().Name)

	err = launcher.Launch(launcher.Options{
		ProjectDir: final.SelectedProject().Path,
		Cols:       final.SelectedLayout().Cols,
		Rows:       final.SelectedLayout().Rows,
		Command:    final.SelectedTool().Command,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error launching: %v\n", err)
		os.Exit(1)
	}
}
