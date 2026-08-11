package internal

import "github.com/charmbracelet/lipgloss"

var (
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#475569")).
			Padding(1, 3)

	HintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	SepStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#334155"))
	DimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E293B"))
	HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#475569"))
	AlarmCol  = lipgloss.Color("#EF4444")
	PausedCol = lipgloss.Color("#FBBF24")
)

func TimeColors(hour int) (accent, date lipgloss.Color) {
	switch {
	case hour >= 6 && hour < 7:
		return lipgloss.Color("#F472B6"), lipgloss.Color("#FB92A0")
	case hour >= 7 && hour < 17:
		return lipgloss.Color("#FB923C"), lipgloss.Color("#FDBA74")
	case hour >= 17 && hour < 19:
		return lipgloss.Color("#F472B6"), lipgloss.Color("#FB92A0")
	default:
		return lipgloss.Color("#818CF8"), lipgloss.Color("#A5B4FC")
	}
}

func AccentStyle(accent lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(accent).Bold(true)
}

func DateStyle(date lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(date).Bold(true)
}
