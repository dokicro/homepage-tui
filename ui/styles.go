package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorGreen  = lipgloss.Color("42")
	ColorRed    = lipgloss.Color("196")
	ColorYellow = lipgloss.Color("214")
	ColorGray   = lipgloss.Color("244")
	ColorWhite  = lipgloss.Color("255")
	ColorCyan   = lipgloss.Color("86")
	ColorDim    = lipgloss.Color("240")

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Padding(0, 1)

	GroupTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			MarginTop(1)

	ServiceStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true).
			Background(lipgloss.Color("237"))

	StatusUpStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StatusDownStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	StatusSlowStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	StatusUnknownStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	LatencyStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	PanelStyle = lipgloss.NewStyle().
			Padding(0, 1)

	ResourceLabelStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Width(5)

	ResourceBarFull = lipgloss.NewStyle().
			Foreground(ColorGreen)

	ResourceBarEmpty = lipgloss.NewStyle().
			Foreground(ColorDim)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim)

	ResourceTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorCyan).
				MarginBottom(1)
)
