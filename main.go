package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func findConfig() string {
	// Explicit argument takes priority
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	// Check current directory
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}

	// Check ~/.config/homepage-tui/
	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, ".config", "homepage-tui", "config.yaml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Fall back to current directory (will produce a clear error)
	return "config.yaml"
}

func main() {
	configPath := findConfig()

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
