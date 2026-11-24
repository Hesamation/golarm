package timer

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

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
	Quiet  bool          // headless countdown, with a best-effort notification
	Target time.Time     // wall-clock fire time, zero for a countdown
}

// Run starts the countdown — interactive TUI unless Quiet is set.
func Run(opts Options) {
	if opts.Quiet {
		time.Sleep(opts.Total)
		fire(opts.Total, time.Time{}, opts.Label)
		return
	}
	p := tea.NewProgram(newModel(opts, time.Now()), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// NextAlarm returns the next occurrence of a validated local time of day.
// Walking actual minutes handles skipped and repeated times at DST transitions.
func NextAlarm(now time.Time, hour, minute int) time.Time {
	untilNextMinute := time.Minute - time.Duration(now.Second())*time.Second - time.Duration(now.Nanosecond())
	for candidate := now.Round(0).Add(untilNextMinute); ; candidate = candidate.Add(time.Minute) {
		h, m, _ := candidate.Clock()
		if h == hour && m == minute {
			return candidate
		}
	}
}

// fire sends a critical desktop notification; failures (e.g. no
// notification daemon running) are silently ignored.
func fire(total time.Duration, target time.Time, label string) {
	body := "Timer done (" + total.String() + ")"
	if !target.IsZero() {
		body = "It's " + target.Format("15:04")
	}
	if label != "" {
		body += " — " + label
	}
	cmd := exec.Command("notify-send", "-u", "critical", "-a", "golarm", "golarm", body)
	if cmd.Start() == nil {
		go func() { _ = cmd.Wait() }()
	}
}

type model struct {
	alarm     bool
	label     string
	deadline  time.Time
	total     time.Duration
	remaining time.Duration
	now       time.Time
	paused    bool
	alarming  bool
	alarmOn   bool
	dismissed bool
	ticks     int
	width     int
	height    int
}

func newModel(opts Options, now time.Time) model {
	alarm := !opts.Target.IsZero()
	total, deadline := opts.Total, now.Add(opts.Total)
	if alarm {
		// Alarm deadlines follow the system clock, including after suspend.
		deadline = opts.Target.Round(0)
		total = max(0, deadline.Sub(now))
	}
	return model{
		alarm: alarm, label: opts.Label, deadline: deadline, total: total,
		remaining: total, now: now,
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

	case tea.KeyMsg:
		cmd := m.key(msg.String(), time.Now())
		return m, cmd

	case ui.TickMsg:
		cmd := m.tick(time.Now())
		return m, tea.Batch(cmd, ui.Tick(tickInterval))
	}
	return m, nil
}

func (m *model) key(key string, now time.Time) tea.Cmd {
	m.now = now
	switch key {
	case "q", "esc", "ctrl+c":
		return tea.Quit
	case " ":
		switch {
		case m.alarming:
			m.dismissed, m.alarmOn = true, true
		case m.alarm:
			return nil
		case m.paused:
			m.deadline = now.Add(m.remaining)
			m.paused = false
		default:
			if cmd := m.updateRemaining(now); cmd != nil {
				return cmd
			}
			m.paused = true
		}
	case "r":
		if m.alarm {
			m.deadline = NextAlarm(now, m.deadline.Hour(), m.deadline.Minute())
			m.total = m.deadline.Sub(now)
		} else {
			m.deadline = now.Add(m.total)
		}
		m.remaining = m.total
		m.alarming, m.alarmOn, m.dismissed, m.paused = false, false, false, false
	case "+", "=", "right":
		return m.adjust(time.Minute, now)
	case "-", "left":
		return m.adjust(-time.Minute, now)
	}
	return nil
}

func (m *model) adjust(delta time.Duration, now time.Time) tea.Cmd {
	if m.alarming || (delta > 0 && m.total > math.MaxInt64-delta) {
		return nil
	}
	deadline, total := m.deadline.Add(delta), m.total+delta
	if m.alarm {
		if !deadline.After(now) {
			return nil
		}
		total = max(total, deadline.Sub(now))
	} else if total <= 0 {
		return nil
	}
	m.deadline, m.total = deadline, total
	if m.paused {
		m.remaining += delta
	}
	return m.updateRemaining(now)
}

func (m *model) tick(now time.Time) tea.Cmd {
	m.now = now
	if m.alarming {
		if !m.dismissed {
			m.ticks++
			if m.ticks%blinkEvery == 0 {
				m.alarmOn = !m.alarmOn
			}
		}
		return nil
	}
	return m.updateRemaining(now)
}

func (m *model) updateRemaining(now time.Time) tea.Cmd {
	if !m.paused {
		m.remaining = m.deadline.Sub(now)
	}
	if m.remaining > 0 {
		return nil
	}
	m.remaining = 0
	m.alarming, m.alarmOn, m.paused = true, true, false
	m.ticks = 0
	var target time.Time
	if m.alarm {
		target = m.deadline
	}
	total, label := m.total, m.label
	return func() tea.Msg {
		fire(total, target, label)
		return nil
	}
}

func (m model) View() string {
	accent, dateColor := ui.Accents(m.now.Hour())
	seconds := m.remaining / time.Second
	if m.remaining%time.Second != 0 {
		seconds++
	}
	timeStr := fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds/60%60, seconds%60)

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
		accStyle = ui.AccentStyle(accent)
	}

	var content strings.Builder
	content.WriteString(strings.Join(ui.Fig(timeStr, accStyle, ui.ClockFigW), "\n"))

	var targetText, labelText, statusLine string
	if m.alarm {
		targetText = ui.HintStyle.Render("fires at " + m.deadline.Format("15:04"))
		content.WriteString("\n\n")
		content.WriteString(targetText)
	}

	if m.label != "" {
		labelText = ui.DateStyle(dateColor).Render(m.label)
		content.WriteString("\n")
		content.WriteString(labelText)
	}

	if statusText != "" {
		statusLine = accStyle.Render(statusText)
		content.WriteString("\n\n")
		content.WriteString(statusLine)
	}

	if !m.alarming && m.total > 0 {
		pct := 1.0 - float64(m.remaining)/float64(m.total)
		filled := min(max(int(pct*float64(barWidth)), 0), barWidth)
		bar := accStyle.Render(strings.Repeat("█", filled)) +
			ui.DimStyle.Render(strings.Repeat("░", barWidth-filled))

		content.WriteString("\n\n")
		content.WriteString(ui.SepStyle.Render(strings.Repeat("─", barWidth)))
		content.WriteString("\n\n")
		content.WriteString(bar)
	}

	restart := "r restart"
	if m.alarm {
		restart = "r re-arm"
	}
	helpText := restart + "  +/− 1m  q quit"
	switch {
	case m.alarming:
		helpText = restart + "  q quit"
		if !m.dismissed {
			helpText = "space dismiss  " + helpText
		}
	case m.paused:
		helpText = "space resume  " + restart + "  q quit"
	case !m.alarm:
		helpText = "space pause  " + helpText
	}
	content.WriteString("\n\n")
	content.WriteString(ui.HelpStyle.Render(helpText))

	card := ui.BorderStyle.Render(content.String())
	w, h := lipgloss.Size(card)
	if (m.width > 0 && w > m.width) || (m.height > 0 && h > m.height) {
		if m.width > 0 && m.width < len(timeStr) {
			timeStr = strings.TrimPrefix(timeStr, "00:")
		}
		if m.width > 0 && m.width < lipgloss.Width(helpText) {
			helpText = "q quit"
		}
		rows := []string{accStyle.Render(timeStr)}
		compactLabel := labelText
		if strings.ContainsRune(m.label, '\n') {
			compactLabel = ui.DateStyle(dateColor).Render(strings.Join(strings.Fields(m.label), " "))
		}
		for _, line := range []string{statusLine, targetText, compactLabel, ui.HelpStyle.Render(helpText)} {
			if line != "" {
				rows = append(rows, line)
			}
		}
		if m.height > 0 {
			rows = rows[:min(m.height, len(rows))]
		}
		if m.width > 0 {
			for i, line := range rows {
				rows[i] = ansi.Truncate(line, m.width, "")
			}
		}
		card = lipgloss.JoinVertical(lipgloss.Center, rows...)
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}
