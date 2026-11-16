package timer

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Hesamation/golarm/internal/ui"
)

const (
	tickInterval = 100 * time.Millisecond
	blinkEvery   = 5 // alarm blink toggles every 5th tick (~500 ms)
	barWidth     = 30
)

type model struct {
	total     time.Duration
	remaining time.Duration
	lastTick  time.Time
	paused    bool
	alarming  bool
	alarmOn   bool
	ticks     int
	width     int
	height    int
}

func Run(d time.Duration) {
	p := tea.NewProgram(model{
		total:     d,
		remaining: d,
		lastTick:  time.Now(),
	}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return ui.Tick(tickInterval)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case " ":
			if !m.alarming {
				m.paused = !m.paused
				m.lastTick = time.Now()
			}
		}
		return m, nil

	case ui.TickMsg:
		now := time.Time(msg)
		if !m.paused && !m.alarming {
			m.remaining -= now.Sub(m.lastTick)
			if m.remaining <= 0 {
				m.remaining = 0
				m.alarming = true
				m.alarmOn = true
			}
		}
		m.lastTick = now
		if m.alarming {
			m.ticks++
			if m.ticks%blinkEvery == 0 {
				m.alarmOn = !m.alarmOn
			}
		}
		return m, ui.Tick(tickInterval)
	}
	return m, nil
}

func (m model) View() string {
	hours := int(m.remaining.Hours())
	minutes := int(m.remaining.Minutes()) % 60
	secs := int(m.remaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)

	var accStyle lipgloss.Style
	var statusText string

	switch {
	case m.alarming:
		if m.alarmOn {
			accStyle = ui.AlarmStyle
		} else {
			accStyle = ui.HintStyle
		}
		statusText = "TIME'S UP"
	case m.paused:
		accStyle = ui.PausedStyle
		statusText = "PAUSED"
	default:
		accent, _ := ui.Accents(m.lastTick.Hour())
		accStyle = ui.AccentStyle(accent)
	}

	var content strings.Builder
	content.WriteString(strings.Join(ui.Fig(timeStr, accStyle, ui.ClockFigW), "\n"))

	if statusText != "" {
		content.WriteString("\n\n")
		content.WriteString(accStyle.Render(statusText))
	}

	if !m.alarming {
		pct := 1.0 - float64(m.remaining)/float64(m.total)
		if pct < 0 {
			pct = 0
		}
		filled := min(int(pct*float64(barWidth)), barWidth)
		bar := accStyle.Render(strings.Repeat("█", filled)) +
			ui.DimStyle.Render(strings.Repeat("░", barWidth-filled))

		content.WriteString("\n\n")
		content.WriteString(ui.SepStyle.Render(strings.Repeat("─", barWidth)))
		content.WriteString("\n\n")
		content.WriteString(bar)
	}

	helpText := "space pause/resume  q quit"
	if m.alarming {
		helpText = "q quit"
	}
	content.WriteString("\n\n")
	content.WriteString(ui.HelpStyle.Render(helpText))

	return ui.Card(content.String(), m.width, m.height)
}
