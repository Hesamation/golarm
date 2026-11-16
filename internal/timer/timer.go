package timer

import (
	"fmt"
	"os"
	"os/exec"
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

// Options configures a countdown run.
type Options struct {
	Total  time.Duration // countdown length
	Label  string        // optional label, shown in the UI and notification
	Quiet  bool          // headless: no TUI, fire the notification, exit 0
	Target *time.Time    // wall-clock fire time (alarm mode)
}

// Run starts the countdown — interactive TUI unless Quiet is set.
func Run(opts Options) {
	if opts.Quiet {
		runQuiet(opts)
		return
	}
	p := tea.NewProgram(model{
		opts:      opts,
		total:     opts.Total,
		remaining: opts.Total,
		lastTick:  time.Now(),
	}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runQuiet blocks until the countdown completes, fires the notification
// and returns so scripts can chain on the exit code.
func runQuiet(opts Options) {
	select {
	case <-time.After(opts.Total):
	}
	fire(opts)
}

// fire sends a critical desktop notification; failures (e.g. no
// notification daemon running) are silently ignored.
func fire(opts Options) {
	body := "Timer done (" + opts.Total.String() + ")"
	if opts.Target != nil {
		body = "It's " + opts.Target.Format("15:04")
	}
	if opts.Label != "" {
		body += " — " + opts.Label
	}
	cmd := exec.Command("notify-send", "-u", "critical", "-a", "golarm", "golarm", body)
	if cmd.Start() == nil {
		go func() { _ = cmd.Wait() }()
	}
}

type model struct {
	opts      Options
	total     time.Duration
	remaining time.Duration
	lastTick  time.Time
	paused    bool
	alarming  bool
	alarmOn   bool
	dismissed bool
	ticks     int
	width     int
	height    int
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
			switch {
			case m.alarming && !m.dismissed:
				m.dismissed = true
			case !m.alarming:
				m.paused = !m.paused
				m.lastTick = time.Now()
			}
		case "r":
			if m.alarming || m.remaining != m.total || m.paused {
				m.alarming = false
				m.alarmOn = false
				m.dismissed = false
				m.paused = false
				m.remaining = m.total
				m.lastTick = time.Now()
			}
		case "+", "=", "right":
			if !m.alarming {
				m.total += time.Minute
				m.remaining += time.Minute
			}
		case "-", "left":
			if !m.alarming && m.total > time.Minute {
				m.total -= time.Minute
				m.remaining -= time.Minute
				if m.remaining <= 0 {
					m.remaining = 0
					m.alarming = true
					m.alarmOn = true
					fire(m.opts)
				}
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
				fire(m.opts)
			}
		}
		m.lastTick = now
		if m.alarming && !m.dismissed {
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
	_, dateColor := ui.Accents(m.lastTick.Hour())
	timeStr := fmt.Sprintf("%02d:%02d:%02d",
		int(m.remaining.Hours()),
		int(m.remaining.Minutes())%60,
		int(m.remaining.Seconds())%60)

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

	if m.opts.Target != nil {
		content.WriteString("\n\n")
		content.WriteString(ui.HintStyle.Render("fires at " + m.opts.Target.Format("15:04")))
	}

	if m.opts.Label != "" {
		content.WriteString("\n")
		content.WriteString(ui.DateStyle(dateColor).Render(m.opts.Label))
	}

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

	helpText := "space pause  r restart  +/− 1m  q quit"
	switch {
	case m.alarming:
		helpText = "space dismiss  r restart  q quit"
	case m.paused:
		helpText = "space resume  r restart  q quit"
	}
	content.WriteString("\n\n")
	content.WriteString(ui.HelpStyle.Render(helpText))

	return ui.Card(content.String(), m.width, m.height)
}
