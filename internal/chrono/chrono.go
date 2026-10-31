package chrono

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
			m.paused = !m.paused
			m.lastTick = time.Now()
		}
		return m, nil

	case tickMsg:
		now := time.Time(msg)
		if !m.paused {
			delta := now.Sub(m.lastTick)
			m.elapsed += delta
		}
		m.lastTick = now
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

	h := int(m.elapsed.Hours())
	min := int(m.elapsed.Minutes()) % 60
	sec := int(m.elapsed.Seconds()) % 60
	cs := (m.elapsed.Milliseconds() / 10) % 100

	timeStr := fmt.Sprintf("%02d:%02d:%02d", h, min, sec)
	csStr := fmt.Sprintf(".%02d", cs)

	var accStyle lipgloss.Style
	if m.paused {
		accStyle = lipgloss.NewStyle().Foreground(internal.PausedCol).Bold(true)
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

	csPad := maxFigW - 1 // right-align with 1 space gap
	if csPad < 0 {
		csPad = 0
	}
	csLine := strings.Repeat(" ", csPad) + " " + csStr

	var content strings.Builder
	content.WriteString(strings.Join(styledFig, "\n"))
	content.WriteString("\n")
	content.WriteString(internal.HintStyle.Render(csLine))

	if m.paused {
		content.WriteString("\n\n")
		content.WriteString(accStyle.Render("PAUSED"))
	}

	content.WriteString("\n\n")
	content.WriteString(internal.HelpStyle.Render("space pause/resume  q quit"))

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
