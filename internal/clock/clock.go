package clock

import (
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/Hesamation/golarm/internal/shamsi"
	"github.com/Hesamation/golarm/internal/ui"
)

type Face uint8

const (
	Digital Face = iota
	Analog
	AnalogANSI
)

// Options configures the clock display.
type Options struct {
	AMPM bool
	FA   bool
	Face Face
}

type model struct {
	opts     Options
	analog   bool
	now      time.Time
	width    int
	height   int
	cells    cellSize
	profile  termenv.Profile
	graphics *kittyDisplay
	layout   dialLayout
	visible  Face
	tickID   uint

	dial *image.RGBA
	// The cached face is keyed by its palette, size and rendering resolution.
	palette    dialPalette
	cachedFace Face
	placement  dialLayout
	dialUpload string
}

func Run(opts Options) {
	m := &model{opts: opts, now: time.Now(), analog: opts.Face != Digital, profile: lipgloss.ColorProfile()}
	programOpts := []tea.ProgramOption{tea.WithAltScreen()}
	if opts.Face != AnalogANSI && m.profile != termenv.Ascii {
		supported, keys := probeGraphics()
		if !m.applyStartupKeys(keys) {
			return
		}
		if supported {
			out := &kittyOutput{File: os.Stdout, display: newKittyDisplay()}
			m.graphics = &out.display
			programOpts = append(programOpts, tea.WithOutput(out))
		}
	}
	p := tea.NewProgram(m, programOpts...)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// The probe can read a key alongside the terminal replies. Apply the clock's
// own keys before starting Bubble Tea, without mistaking arrow keys for Esc.
func (m *model) applyStartupKeys(keys []byte) bool {
	for len(keys) > 0 {
		seq, _, n, _ := ansi.DecodeSequence(keys, 0, nil)
		if n == 0 {
			break
		}
		switch string(seq) {
		case "q", "\x1b", "\x03":
			return false
		case "a":
			m.analog = !m.analog
		}
		keys = keys[n:]
	}
	return true
}

type tickMsg uint

func (m *model) tick() tea.Cmd {
	interval := time.Second
	switch m.visible {
	case Analog:
		interval = 50 * time.Millisecond
	case AnalogANSI:
		interval = 200 * time.Millisecond
	}
	now, id := time.Now(), m.tickID
	wait := now.Truncate(interval).Add(interval).Sub(now)
	return tea.Tick(wait, func(time.Time) tea.Msg { return tickMsg(id) })
}

func (m *model) Init() tea.Cmd { return m.tick() }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.cells = terminalCells()
		return m, m.reflow()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "a":
			m.analog = !m.analog
			return m, m.reflow()
		}
	case tickMsg:
		if uint(msg) == m.tickID {
			m.now = time.Now()
			return m, m.tick()
		}
	}
	return m, nil
}

func (m *model) reflow() tea.Cmd {
	visible, layout := chooseDial(m.width, m.height, m.cells, m.analog, m.graphics != nil, m.profile)
	m.layout = layout
	if visible == m.visible {
		return nil
	}
	m.visible = visible
	m.now = time.Now()
	m.tickID++
	return m.tick()
}

func chooseDial(width, height int, cell cellSize, requested, native bool, profile termenv.Profile) (Face, dialLayout) {
	if !requested || profile == termenv.Ascii {
		return Digital, dialLayout{}
	}
	if cell.width > 0 && cell.height > 0 && native {
		if layout := layoutDial(width, height, cell, true); layout.diameter > 0 {
			return Analog, layout
		}
	}
	if cell.width == 0 || cell.height == 0 {
		cell = cellSize{8, 16}
	}
	if layout := layoutDial(width, height, cell, false); layout.diameter > 0 {
		return AnalogANSI, layout
	}
	return Digital, dialLayout{}
}

type clockReadout struct {
	exact, shamsi, gregorian, progress string
	fraction                           float64
	accent, date                       lipgloss.Color
}

func readClock(now time.Time, opts Options) clockReadout {
	format := "15:04:05"
	if opts.AMPM {
		format = "03:04:05 PM"
	}
	sd := shamsi.FromGregorian(now.Year(), int(now.Month()), now.Day())
	seconds := now.Hour()*3600 + now.Minute()*60 + now.Second()
	fraction := float64(seconds) / 86400
	accent, date := ui.Accents(now.Hour())
	readout := clockReadout{
		exact:     now.Format(format),
		shamsi:    sd.WeekDayName() + " " + sd.FormatFull(),
		gregorian: now.Format("Monday, 02 January 2006"),
		progress:  fmt.Sprintf("%.0f%% of day", fraction*100),
		fraction:  fraction,
		accent:    accent,
		date:      date,
	}
	if opts.FA {
		readout.shamsi = ui.FaDigits(readout.shamsi)
		readout.gregorian = ui.FaDigits(readout.gregorian)
		readout.progress = ui.FaDigits(readout.progress)
	}
	return readout
}

func (r clockReadout) bar(width int) string {
	filled := int(r.fraction * float64(width))
	return ui.AccentStyle(r.accent).Render(strings.Repeat("█", filled)) +
		ui.DimStyle.Render(strings.Repeat("░", width-filled))
}

func (m *model) exactTime(r clockReadout) string {
	if m.opts.FA {
		return ui.FaDigits(r.exact)
	}
	return r.exact
}

func (m *model) View() string {
	readout := readClock(m.now, m.opts)
	if m.visible != Digital {
		return m.analogView(readout)
	}
	view := m.digitalView(readout)
	m.dialUpload = ""
	// Keep the delete on the digital frame so a rapid series of key presses
	// cannot replace the cleanup frame before Bubble Tea flushes it.
	if m.graphics != nil && m.placement.diameter > 0 {
		view += m.graphics.clear()
	}
	return view
}

func (m *model) digitalView(r clockReadout) string {
	maxFigWidth := ui.ClockFigW
	if m.opts.AMPM {
		maxFigWidth = ui.ClockAMPMFigW
	}
	const barWidth = 30
	help := "a  analog    q  quit"
	if m.analog {
		help = "a  cancel analog    q  quit"
	}
	sections := []string{
		strings.Join(ui.Fig(r.exact, ui.AccentStyle(r.accent), maxFigWidth), "\n"),
		ui.DateStyle(r.date).Render(r.shamsi),
		ui.HintStyle.Render(r.gregorian),
		ui.SepStyle.Render(strings.Repeat("─", barWidth)),
		r.bar(barWidth),
		ui.HintStyle.Render(r.progress),
		ui.HelpStyle.Render(help),
	}
	content := strings.Join(sections, "\n\n")
	w, h := lipgloss.Size(ui.BorderStyle.Render(content))
	if (m.width > 0 && w > m.width) || (m.height > 0 && h > m.height) {
		compact := ansi.Truncate(ui.AccentStyle(r.accent).Render(m.exactTime(r)), m.width, "")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, compact)
	}
	return ui.Card(content, m.width, m.height)
}

func (m *model) analogView(r clockReadout) string {
	native := m.visible == Analog
	size := m.layout.diameter
	if !native {
		size = min(size, 192)
	}
	theme := ui.ActiveTheme()
	palette := paletteFor(theme.Dim, theme.Border, r.date)
	if m.dial == nil || m.dial.Bounds().Dx() != size || m.palette != palette || m.cachedFace != m.visible {
		m.dial = drawDial(size, palette, !native)
		m.palette, m.cachedFace, m.dialUpload = palette, m.visible, ""
	}
	hands := drawHands(size, m.now, palette, themeRGB(r.accent), !native)
	readoutStyle := func(lit lipgloss.Color, lift float64) lipgloss.Style {
		c := mix(themeRGB(lit), palette.lume, lift)
		return lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)))
	}
	details := readoutStyle(theme.Hint, .30)
	graduation := readoutStyle(theme.Border, .08)
	const progressWidth = 16
	progress := int(r.fraction * progressWidth)
	meter := ui.AccentStyle(r.accent).Render(strings.Repeat("━", progress)) +
		graduation.Render(strings.Repeat("─", progressWidth-progress))
	data := []string{
		readoutStyle(r.date, .74).Bold(true).Render(m.exactTime(r)),
		"",
		ui.DateStyle(r.date).Render(r.shamsi),
		details.Render(r.gregorian),
		"",
		meter + "  " + details.Render(r.progress),
	}
	footer := readoutStyle(theme.Hint, .18).Render("a  digital    q  quit")
	if !native {
		rows := m.layout.lines(m.width, m.height, m.layout.colorCells(m.dial, hands, m.profile), data, footer)
		if m.graphics != nil && m.placement.diameter > 0 {
			rows[len(rows)-1] += m.graphics.clear()
		}
		return strings.Join(rows, "\n")
	}
	if m.dialUpload == "" || m.placement != m.layout {
		m.dialUpload = kittyLayer(m.dial, m.layout, m.graphics.dialID, 1)
		m.placement = m.layout
	}
	rows := m.layout.lines(m.width, m.height, nil, data, footer)
	// The two uploads ride on separate rows so Bubble Tea's row diff can
	// skip the static dial while the hands are moving.
	rows[len(rows)-2] += m.dialUpload
	rows[len(rows)-1] += kittyLayer(hands, m.layout, m.graphics.handsID, 2)
	return strings.Join(rows, "\n")
}
