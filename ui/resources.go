package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Resources struct {
	CPUUsage  float64
	MemTotal  int64
	MemActive int64
	DiskSize  int64
	DiskUsed  int64
	DiskMount string
	NetRxSec  float64
	NetTxSec  float64
}

func RenderResources(res Resources, width, height int) string {
	title := ResourceTitleStyle.Render("RESOURCES")

	barWidth := width - 14
	if barWidth < 5 {
		barWidth = 5
	}

	cpuBar := renderBar("CPU", res.CPUUsage, barWidth)

	var memPct float64
	if res.MemTotal > 0 {
		memPct = float64(res.MemActive) / float64(res.MemTotal) * 100
	}
	memBar := renderBar("MEM", memPct, barWidth)

	var diskPct float64
	if res.DiskSize > 0 {
		diskPct = float64(res.DiskUsed) / float64(res.DiskSize) * 100
	}
	diskBar := renderBar("DISK", diskPct, barWidth)

	netLine := fmt.Sprintf("  NET:  ↓%s  ↑%s",
		formatBytes(res.NetRxSec),
		formatBytes(res.NetTxSec))
	netLine = LatencyStyle.Render(netLine)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, cpuBar, memBar, diskBar, netLine,
	)

	return PanelStyle.Width(width).Height(height).Render(content)
}

func renderBar(label string, pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filled := int(pct / 100 * float64(width))
	empty := width - filled

	bar := ResourceBarFull.Render(strings.Repeat("█", filled)) +
		ResourceBarEmpty.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("  %s %s %3.0f%%",
		ResourceLabelStyle.Render(label+":"),
		bar,
		pct,
	)
}

func formatBytes(bytesPerSec float64) string {
	switch {
	case bytesPerSec >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB/s", bytesPerSec/(1024*1024*1024))
	case bytesPerSec >= 1024*1024:
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/(1024*1024))
	case bytesPerSec >= 1024:
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	default:
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	}
}
