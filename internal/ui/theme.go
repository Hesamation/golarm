package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds every color role used by the UI.
type Theme struct {
	Name   string
	Border lipgloss.Color
	Hint   lipgloss.Color
	Sep    lipgloss.Color
	Dim    lipgloss.Color
	Help   lipgloss.Color
	Alarm  lipgloss.Color
	Paused lipgloss.Color

	Dawn  lipgloss.Color // hour 6-7
	Day   lipgloss.Color // hour 7-17
	Dusk  lipgloss.Color // hour 17-19
	Night lipgloss.Color // the rest

	DateDawn  lipgloss.Color
	DateDay   lipgloss.Color
	DateDusk  lipgloss.Color
	DateNight lipgloss.Color
}

// Themes are the built-in themes, selectable via --theme / GOLARM_THEME.
var Themes = map[string]Theme{
	"default": {
		Name:   "default",
		Border: "#475569", Hint: "#64748B", Sep: "#334155", Dim: "#1E293B", Help: "#475569",
		Alarm: "#EF4444", Paused: "#FBBF24",
		Dawn: "#F472B6", Day: "#FB923C", Dusk: "#F472B6", Night: "#818CF8",
		DateDawn: "#FB92A0", DateDay: "#FDBA74", DateDusk: "#FB92A0", DateNight: "#A5B4FC",
	},
	"catppuccin": {
		Name:   "catppuccin",
		Border: "#45475A", Hint: "#6C7086", Sep: "#313244", Dim: "#181825", Help: "#6C7086",
		Alarm: "#F38BA8", Paused: "#F9E2AF",
		Dawn: "#F5C2E7", Day: "#FAB387", Dusk: "#F5C2E7", Night: "#CBA6F7",
		DateDawn: "#FFD0E0", DateDay: "#F9D5B0", DateDusk: "#FFD0E0", DateNight: "#D7C4FA",
	},
	"gruvbox": {
		Name:   "gruvbox",
		Border: "#504945", Hint: "#928374", Sep: "#3C3836", Dim: "#1D2021", Help: "#928374",
		Alarm: "#FB4934", Paused: "#FABD2F",
		Dawn: "#D3869B", Day: "#FE8019", Dusk: "#D3869B", Night: "#83A598",
		DateDawn: "#E2B4C0", DateDay: "#FFAB5E", DateDusk: "#E2B4C0", DateNight: "#A9BFCF",
	},
	"tokyo-night": {
		Name:   "tokyo-night",
		Border: "#414868", Hint: "#565F89", Sep: "#292E42", Dim: "#16161E", Help: "#565F89",
		Alarm: "#F7768E", Paused: "#E0AF68",
		Dawn: "#BB9AF7", Day: "#FF9E64", Dusk: "#BB9AF7", Night: "#7AA2F7",
		DateDawn: "#C8B3F9", DateDay: "#FFB380", DateDusk: "#C8B3F9", DateNight: "#9FB9F9",
	},
}

// FromName returns the named built-in theme.
func FromName(name string) (Theme, bool) {
	t, ok := Themes[name]
	return t, ok
}

var current = Themes["default"]

// Package-level styles derived from the active theme. Set once at startup
// via SetTheme, before any model runs.
var (
	BorderStyle lipgloss.Style
	HintStyle   lipgloss.Style
	SepStyle    lipgloss.Style
	DimStyle    lipgloss.Style
	HelpStyle   lipgloss.Style
	AlarmStyle  lipgloss.Style
	PausedStyle lipgloss.Style
)

// SetTheme switches the active theme and rebuilds the derived styles.
func SetTheme(t Theme) {
	current = t
	BorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 3)
	HintStyle = lipgloss.NewStyle().Foreground(t.Hint)
	SepStyle = lipgloss.NewStyle().Foreground(t.Sep)
	DimStyle = lipgloss.NewStyle().Foreground(t.Dim)
	HelpStyle = lipgloss.NewStyle().Foreground(t.Help)
	AlarmStyle = lipgloss.NewStyle().Foreground(t.Alarm).Bold(true)
	PausedStyle = lipgloss.NewStyle().Foreground(t.Paused).Bold(true)
}

func init() { SetTheme(current) }

// Accents returns the active theme's accent and date colors for the hour of day.
func Accents(hour int) (accent, date lipgloss.Color) {
	return current.Accents(hour)
}

// Accents returns the accent and date colors for the given hour of day.
func (t Theme) Accents(hour int) (accent, date lipgloss.Color) {
	switch {
	case hour >= 6 && hour < 7:
		return t.Dawn, t.DateDawn
	case hour >= 7 && hour < 17:
		return t.Day, t.DateDay
	case hour >= 17 && hour < 19:
		return t.Dusk, t.DateDusk
	default:
		return t.Night, t.DateNight
	}
}

// AccentStyle returns a bold foreground style for the given accent color.
func AccentStyle(c lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c).Bold(true)
}

// DateStyle returns a bold foreground style for the given date color.
func DateStyle(c lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c).Bold(true)
}
