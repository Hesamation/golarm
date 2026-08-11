package clock

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/user/golarm/internal"
	"github.com/user/golarm/internal/shamsi"
)

type tickMsg time.Time

type model struct {
	now    time.Time
	width  int
	height int
	ampm   bool
}

func Run(ampm bool) {
	p := tea.NewProgram(model{ampm: ampm, now: time.Now()}, tea.WithAltScreen())
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
		}
		return m, nil
	case tickMsg:
		m.now = time.Time(msg)
		return m, tick()
	}
	return m, nil
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
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

	now := m.now
	if now.IsZero() {
		now = time.Now()
	}

	accent, date := internal.TimeColors(now.Hour())
	accStyle := internal.AccentStyle(accent)
	dateStyle := internal.DateStyle(date)

	var timeStr string
	if m.ampm {
		h := now.Hour()
		suffix := "AM"
		if h >= 12 {
			suffix = "PM"
		}
		if h == 0 {
			h = 12
		} else if h > 12 {
			h -= 12
		}
		timeStr = fmt.Sprintf("%02d:%02d:%02d %s", h, now.Minute(), now.Second(), suffix)
	} else {
		timeStr = fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	}

	maxTimeStr := "88:88:88"
	if m.ampm {
		maxTimeStr = "88:88:88 PM"
	}
	maxFigW := internal.FigletWidth(maxTimeStr)

	figLines := internal.RenderFiglet(timeStr)
	styledFig := make([]string, len(figLines))
	for i, line := range figLines {
		if pad := maxFigW - len(line); pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		styledFig[i] = accStyle.Render(line)
	}

	sd := shamsi.FromGregorian(now.Year(), int(now.Month()), now.Day())
	shamsiStr := sd.FormatFull()
	gregDay := shamsi.FormatGregorianWeekDay(now.Year(), int(now.Month()), now.Day())
	gregStr := fmt.Sprintf("%s, %s", gregDay, now.Format("02 January 2006"))

	secs := now.Hour()*3600 + now.Minute()*60 + now.Second()
	pct := float64(secs) / 86400.0
	barW := 30
	filled := int(pct * float64(barW))
	if filled > barW {
		filled = barW
	}
	bar := accStyle.Render(strings.Repeat("█", filled)) +
		internal.DimStyle.Render(strings.Repeat("░", barW-filled))
	pctStr := fmt.Sprintf("%.0f%% of day", pct*100)

	var content strings.Builder
	content.WriteString(strings.Join(styledFig, "\n"))
	content.WriteString("\n\n")
	content.WriteString(dateStyle.Render(shamsiStr))
	content.WriteString("\n\n")
	content.WriteString(internal.HintStyle.Render(gregStr))
	content.WriteString("\n\n")
	content.WriteString(internal.SepStyle.Render(strings.Repeat("─", 30)))
	content.WriteString("\n\n")
	content.WriteString(bar)
	content.WriteString("\n\n")
	content.WriteString(internal.HintStyle.Render(pctStr))
	content.WriteString("\n\n")
	content.WriteString(internal.HelpStyle.Render("q  quit"))

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
