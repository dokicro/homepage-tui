package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ServiceEntry struct {
	GroupName   string
	Name        string
	Href        string
	HTTPStatus  int
	Latency     float64
	DockerState string
	HasMonitor  bool
	HasDocker   bool
	Container   string
	Server      string
	Error       string
	Loading     bool
}

func RenderServices(entries []ServiceEntry, cursor int, width, height int) string {
	var lines []string
	currentGroup := ""

	for i, entry := range entries {
		if entry.GroupName != currentGroup {
			currentGroup = entry.GroupName
			title := GroupTitleStyle.Render(strings.ToUpper(currentGroup))
			lines = append(lines, title)
		}

		indicator := renderIndicator(entry)
		name := entry.Name
		status := renderStatus(entry)

		line := fmt.Sprintf("  %s %s  %s", indicator, name, status)

		if i == cursor {
			line = SelectedStyle.Width(width).Render(line)
		} else {
			line = ServiceStyle.Width(width).Render(line)
		}

		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return PanelStyle.Width(width).Height(height).Render(content)
}

func renderIndicator(e ServiceEntry) string {
	if e.Loading {
		return StatusUnknownStyle.Render("◌")
	}
	if e.Error != "" {
		return StatusUnknownStyle.Render("?")
	}
	if !e.HasMonitor && !e.HasDocker {
		return StatusUnknownStyle.Render("○")
	}

	if e.HasDocker && e.DockerState != "" {
		if e.DockerState == "running" {
			return StatusUpStyle.Render("●")
		}
		return StatusDownStyle.Render("○")
	}

	if e.HasMonitor {
		if e.HTTPStatus >= 200 && e.HTTPStatus < 300 {
			if e.Latency > 2000 {
				return StatusSlowStyle.Render("◐")
			}
			return StatusUpStyle.Render("●")
		}
		if e.HTTPStatus > 0 {
			return StatusSlowStyle.Render("◐")
		}
		return StatusDownStyle.Render("○")
	}

	return StatusUnknownStyle.Render("○")
}

func renderStatus(e ServiceEntry) string {
	if e.Loading {
		return LatencyStyle.Render("...")
	}
	if e.Error != "" {
		return StatusDownStyle.Render("ERROR")
	}

	var parts []string

	if e.HasMonitor && e.HTTPStatus > 0 {
		parts = append(parts, fmt.Sprintf("%d", e.HTTPStatus))
		parts = append(parts, fmt.Sprintf("%dms", int(e.Latency)))
	} else if e.HasMonitor {
		parts = append(parts, "DOWN")
	}

	if e.HasDocker && e.DockerState != "" {
		parts = append(parts, e.DockerState)
	}

	return LatencyStyle.Render(strings.Join(parts, "  "))
}
