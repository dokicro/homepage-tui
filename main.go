package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	client := NewClient(cfg.HomepageURL, cfg.Auth)

	if err := client.Healthcheck(); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot reach Homepage at %s: %v\n", cfg.HomepageURL, err)
		os.Exit(1)
	}

	m := newModel(cfg, client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
