package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func RenderHeader(width int, lastRefresh time.Time, warning string) string {
	title := HeaderStyle.Render("Homepage TUI")

	var right string
	if warning != "" {
		right = WarningStyle.Render(warning)
	} else {
		elapsed := time.Since(lastRefresh).Truncate(time.Second)
		right = LatencyStyle.Render(fmt.Sprintf("last refresh: %s ago", elapsed))
	}

	gap := width - lipgloss.Width(title) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	padding := lipgloss.NewStyle().Width(gap).Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top, title, padding, right)
}
