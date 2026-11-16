package clock

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hesamation/golarm/internal/shamsi"
	"github.com/Hesamation/golarm/internal/ui"
)

// Options configures the clock display.
type Options struct {
	AMPM bool // 12-hour clock
	FA   bool // Persian digits on the small text lines
}

type model struct {
	opts    Options
	now     time.Time
	width   int
	height  int
	maxFigW int
}

func Run(opts Options) {
	maxFigW := ui.ClockFigW
	if opts.AMPM {
		maxFigW = ui.ClockAMPMFigW
	}
	p := tea.NewProgram(model{opts: opts, now: time.Now(), maxFigW: maxFigW}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return ui.TickNext(time.Now(), time.Second)
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
	case ui.TickMsg:
		m.now = time.Time(msg)
		return m, ui.TickNext(m.now, time.Second)
	}
	return m, nil
}

func (m model) View() string {
	now := m.now
	if now.IsZero() {
		now = time.Now()
	}

	accent, date := ui.Accents(now.Hour())
	accStyle := ui.AccentStyle(accent)
	dateStyle := ui.DateStyle(date)

	var timeStr string
	if m.opts.AMPM {
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

	sd := shamsi.FromGregorian(now.Year(), int(now.Month()), now.Day())
	shamsiStr := fmt.Sprintf("%s %s", sd.WeekDayName(), sd.FormatFull())
	gregStr := now.Format("Monday, 02 January 2006")

	secs := now.Hour()*3600 + now.Minute()*60 + now.Second()
	pct := float64(secs) / 86400.0
	filled := min(int(pct*float64(barWidth)), barWidth)
	bar := accStyle.Render(strings.Repeat("█", filled)) +
		ui.DimStyle.Render(strings.Repeat("░", barWidth-filled))
	pctStr := fmt.Sprintf("%.0f%% of day", pct*100)

	if m.opts.FA {
		shamsiStr = ui.FaDigits(shamsiStr)
		gregStr = ui.FaDigits(gregStr)
		pctStr = ui.FaDigits(pctStr)
	}

	var content strings.Builder
	content.WriteString(strings.Join(ui.Fig(timeStr, accStyle, m.maxFigW), "\n"))
	content.WriteString("\n\n")
	content.WriteString(dateStyle.Render(shamsiStr))
	content.WriteString("\n\n")
	content.WriteString(ui.HintStyle.Render(gregStr))
	content.WriteString("\n\n")
	content.WriteString(ui.SepStyle.Render(strings.Repeat("─", barWidth)))
	content.WriteString("\n\n")
	content.WriteString(bar)
	content.WriteString("\n\n")
	content.WriteString(ui.HintStyle.Render(pctStr))
	content.WriteString("\n\n")
	content.WriteString(ui.HelpStyle.Render("q  quit"))

	return ui.Card(content.String(), m.width, m.height)
}

const barWidth = 30
