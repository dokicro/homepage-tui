package ui

func RenderFooter(width int) string {
	keys := "↑↓ navigate  enter open  r refresh  tab switch  q quit  ? help"
	return FooterStyle.Width(width).Render(keys)
}
