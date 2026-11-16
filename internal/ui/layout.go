package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TickMsg carries the time a scheduled tick fired.
type TickMsg time.Time

// Tick schedules the next tick after d.
func Tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return TickMsg(t) })
}

// TickNext schedules a tick at the next boundary of step after now, which
// keeps second-resolution displays from drifting (tea.Tick measured from
// processing time slowly skips seconds otherwise).
func TickNext(now time.Time, step time.Duration) tea.Cmd {
	return tea.Tick(step-time.Duration(now.UnixNano()%int64(step)), func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Card wraps content in the themed rounded border and centers the card
// within w×h cells.
func Card(content string, w, h int) string {
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, BorderStyle.Render(content))
}
