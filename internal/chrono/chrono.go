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

const tickInterval = 100 * time.Millisecond

type model struct {
	elapsed  time.Duration
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

	content.WriteString("\n\n")
	content.WriteString(ui.HelpStyle.Render("space pause/resume  q quit"))

	return ui.Card(content.String(), m.width, m.height)
}
