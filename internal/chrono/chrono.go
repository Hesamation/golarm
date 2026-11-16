package chrono

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
	visibleLaps  = 3 // laps shown under the clock
)

type model struct {
	elapsed  time.Duration
	laps     []time.Duration // split totals at each lap press
	lastTick time.Time
	paused   bool
	width    int
	height   int
}

func Run() {
	p := tea.NewProgram(model{lastTick: time.Now()}, tea.WithAltScreen())
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
			m.paused = !m.paused
			m.lastTick = time.Now()
		case "l":
			if !m.paused {
				m.laps = append(m.laps, m.elapsed)
			}
		case "r":
			m.elapsed = 0
			m.laps = nil
			m.paused = false
			m.lastTick = time.Now()
		}
		return m, nil

	case ui.TickMsg:
		now := time.Time(msg)
		if !m.paused {
			m.elapsed += now.Sub(m.lastTick)
		}
		m.lastTick = now
		return m, ui.Tick(tickInterval)
	}
	return m, nil
}

func (m model) View() string {
	_, dateColor := ui.Accents(m.lastTick.Hour())
	timeStr := fmt.Sprintf("%02d:%02d:%02d",
		int(m.elapsed.Hours()),
		int(m.elapsed.Minutes())%60,
		int(m.elapsed.Seconds())%60)
	csStr := fmt.Sprintf(".%02d", (m.elapsed.Milliseconds()/10)%100)

	var accStyle lipgloss.Style
	if m.paused {
		accStyle = ui.PausedStyle
	} else {
		accent, _ := ui.Accents(m.lastTick.Hour())
		accStyle = ui.AccentStyle(accent)
	}

	var content strings.Builder
	content.WriteString(strings.Join(ui.Fig(timeStr, accStyle, ui.ClockFigW), "\n"))

	// Hundredths, right-aligned flush with the clock face's right edge.
	csLine := strings.Repeat(" ", max(ui.ClockFigW-len(csStr), 0)) + csStr
	content.WriteString("\n")
	content.WriteString(ui.HintStyle.Render(csLine))

	if m.paused {
		content.WriteString("\n\n")
		content.WriteString(accStyle.Render("PAUSED"))
	}

	if len(m.laps) > 0 {
		content.WriteString("\n\n")
		first := max(len(m.laps)-visibleLaps, 0)
		for i := first; i < len(m.laps); i++ {
			split := m.laps[i]
			if i > 0 {
				split -= m.laps[i-1]
			}
			row := fmt.Sprintf("L%d  +%s  Σ %s", i+1, lapFmt(split), lapFmt(m.laps[i]))
			if i == len(m.laps)-1 {
				row = ui.DateStyle(dateColor).Render(row)
			} else {
				row = ui.HintStyle.Render(row)
			}
			if i > first {
				content.WriteString("\n")
			}
			content.WriteString(row)
		}
	}

	content.WriteString("\n\n")
	content.WriteString(ui.HelpStyle.Render("space pause  l lap  r reset  q quit"))

	return ui.Card(content.String(), m.width, m.height)
}

// lapFmt renders a duration as m:ss.cc (or h:mm:ss.cc past an hour).
func lapFmt(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	cs := (d.Milliseconds() / 10) % 100
	if h := int(d.Hours()); h > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%02d", h, int(d.Minutes())%60, int(d.Seconds())%60, cs)
	}
	return fmt.Sprintf("%02d:%02d.%02d", int(d.Minutes())%60, int(d.Seconds())%60, cs)
}
