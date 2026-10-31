package timer

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Hesamation/golarm/internal"
)

type tickMsg time.Time

type model struct {
	total     time.Duration
	remaining time.Duration
	lastTick  time.Time
	paused    bool
	alarming  bool
	alarmOn   bool
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
	return tick()
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

	case tickMsg:
		now := time.Time(msg)
		if !m.paused && !m.alarming {
			delta := now.Sub(m.lastTick)
			m.remaining -= delta
			if m.remaining <= 0 {
				m.remaining = 0
				m.alarming = true
				m.alarmOn = true
			}
		}
		m.lastTick = now
		if m.alarming {
			m.alarmOn = !m.alarmOn
		}
		return m, tick()
	}
	return m, nil
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) View() string {
	if m.width == 0 {
		m.width = 80
	}
	if m.height == 0 {
		m.height = 24
	}

	h := int(m.remaining.Hours())
	min := int(m.remaining.Minutes()) % 60
	sec := int(m.remaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d:%02d", h, min, sec)

	var accStyle lipgloss.Style
	var statusText string

	if m.alarming {
		if m.alarmOn {
			accStyle = lipgloss.NewStyle().Foreground(internal.AlarmCol).Bold(true)
		} else {
			accStyle = internal.HintStyle
		}
		statusText = "TIME'S UP"
	} else if m.paused {
		accStyle = lipgloss.NewStyle().Foreground(internal.PausedCol).Bold(true)
		statusText = "PAUSED"
	} else {
		accent, _ := internal.TimeColors(time.Now().Hour())
		accStyle = internal.AccentStyle(accent)
	}

	maxFigW := internal.FigletWidth("88:88:88")
	figLines := internal.RenderFiglet(timeStr)
	styledFig := make([]string, len(figLines))
	for i, line := range figLines {
		if pad := maxFigW - len(line); pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		styledFig[i] = accStyle.Render(line)
	}

	var content strings.Builder
	content.WriteString(strings.Join(styledFig, "\n"))

	if statusText != "" {
		content.WriteString("\n\n")
		content.WriteString(accStyle.Render(statusText))
	}

	if !m.alarming && m.total > 0 {
		pct := 1.0 - float64(m.remaining)/float64(m.total)
		if pct < 0 {
			pct = 0
		}
		barW := 30
		filled := int(pct * float64(barW))
		if filled > barW {
			filled = barW
		}
		bar := accStyle.Render(strings.Repeat("█", filled)) +
			internal.DimStyle.Render(strings.Repeat("░", barW-filled))

		content.WriteString("\n\n")
		content.WriteString(internal.SepStyle.Render(strings.Repeat("─", 30)))
		content.WriteString("\n\n")
		content.WriteString(bar)
	}

	helpText := "space pause/resume  q quit"
	if m.alarming {
		helpText = "q quit"
	}
	content.WriteString("\n\n")
	content.WriteString(internal.HelpStyle.Render(helpText))

	card := internal.BorderStyle.Render(content.String())

	cardLines := strings.Split(card, "\n")
	cardH := len(cardLines)
	cardW := 0
	for _, line := range cardLines {
		w := lipgloss.Width(line)
		if w > cardW {
			cardW = w
		}
	}

	topPad := (m.height - cardH) / 2
	if topPad < 0 {
		topPad = 0
	}
	leftPad := (m.width - cardW) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	var b strings.Builder
	b.WriteString(strings.Repeat("\n", topPad))
	for _, line := range cardLines {
		b.WriteString(strings.Repeat(" ", leftPad))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
