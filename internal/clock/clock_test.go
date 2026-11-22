package clock

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestHandTurns(t *testing.T) {
	now := time.Date(2026, time.November, 22, 15, 15, 30, 500_000_000, time.UTC)
	h, m, s := handTurns(now)
	if math.Abs(h-0.27153935185185185) > 1e-10 ||
		math.Abs(m-0.2584722222222222) > 1e-10 ||
		math.Abs(s-30.5/60) > 1e-10 {
		t.Fatalf("unexpected turns at %v: %v %v %v", now, h, m, s)
	}
	h, m, s = handTurns(time.Date(2026, time.November, 23, 0, 0, 0, 0, time.UTC))
	if h != 0 || m != 0 || s != 0 {
		t.Fatalf("midnight turns: %v %v %v", h, m, s)
	}
}

func TestChooseDial(t *testing.T) {
	for _, tt := range []struct {
		name              string
		width, height     int
		cell              cellSize
		requested, native bool
		profile           termenv.Profile
		want              Face
		wantStack         bool
	}{
		{"standard", 80, 24, cellSize{9, 18}, true, true, termenv.TrueColor, Analog, false},
		{"vertical split", 48, 28, cellSize{9, 18}, true, true, termenv.TrueColor, Analog, true},
		{"no pixel dimensions", 80, 24, cellSize{}, true, true, termenv.TrueColor, AnalogANSI, false},
		{"no graphics", 80, 24, cellSize{8, 16}, true, false, termenv.ANSI256, AnalogANSI, false},
		{"a few rows", 80, 6, cellSize{8, 16}, true, true, termenv.TrueColor, Digital, false},
		{"color disabled", 80, 24, cellSize{8, 16}, true, true, termenv.Ascii, Digital, false},
		{"digital default", 80, 24, cellSize{8, 16}, false, true, termenv.TrueColor, Digital, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			face, l := chooseDial(tt.width, tt.height, 5, tt.cell, tt.requested, tt.native, tt.profile)
			if face != tt.want || l.stacked != tt.wantStack {
				t.Fatalf("got face %d, stacked %t, want %d, stacked %t", face, l.stacked, tt.want, tt.wantStack)
			}
			if face != Digital {
				x, y := l.pixelOrigin()
				if x < 0 || y < 0 || x+l.diameter > tt.width*l.width || y+l.diameter > tt.height*l.height {
					t.Fatalf("dial exceeds viewport: %+v", l)
				}
			}
		})
	}
}

func TestClockDates(t *testing.T) {
	now := time.Date(2026, time.November, 22, 19, 30, 0, 0, time.UTC)
	const gregorian = "Sunday, 22 November 2026"
	const solarHijri = "Yekshanbe 1 Azar 1405"
	for _, tt := range []struct {
		opts Options
		want []string
	}{
		{Options{}, []string{gregorian}},
		{Options{Calendar: Gregorian}, []string{gregorian}},
		{Options{Calendar: Shamsi}, []string{solarHijri}},
		{Options{Calendar: BothCalendars}, []string{solarHijri, gregorian}},
		{Options{FA: true}, []string{"Sunday, ۲۲ November ۲۰۲۶"}},
		{Options{Calendar: Shamsi, FA: true}, []string{"Yekshanbe ۱ Azar ۱۴۰۵"}},
	} {
		r := readClock(now, tt.opts)
		if !slices.Equal(r.dates, tt.want) {
			t.Errorf("readClock with %+v gives dates %q, want %q", tt.opts, r.dates, tt.want)
		}
	}
	for _, face := range []Face{Digital, AnalogANSI} {
		for _, calendar := range []Calendar{"", Shamsi, BothCalendars} {
			m := &model{opts: Options{Calendar: calendar}, analog: face != Digital, profile: termenv.TrueColor}
			m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			m.now = now
			view := m.View()
			gregorianIndex, shamsiIndex := strings.Index(view, gregorian), strings.Index(view, solarHijri)
			if (gregorianIndex >= 0) != (calendar != Shamsi) ||
				(shamsiIndex >= 0) != (calendar != "") ||
				(calendar == BothCalendars && shamsiIndex >= gregorianIndex) {
				t.Errorf("wrong dates in face %d with calendar %q", face, calendar)
			}
		}
	}
}

func TestStackedDialFitsSingleDate(t *testing.T) {
	for _, tt := range []struct {
		calendar Calendar
		want     Face
	}{{Gregorian, AnalogANSI}, {BothCalendars, Digital}} {
		m := &model{opts: Options{Calendar: tt.calendar}, analog: true, width: 44, height: 22,
			cells: cellSize{8, 16}, profile: termenv.TrueColor}
		m.reflow()
		if m.visible != tt.want || (m.visible == AnalogANSI && !m.layout.stacked) {
			t.Errorf("calendar %q gave face %d and stacked %t, want %d", tt.calendar, m.visible, m.layout.stacked, tt.want)
		}
	}
}

func TestCellDialKeepsHandContrast(t *testing.T) {
	face, needle := rgb(0x131722), rgb(0xD2ADFF)
	box := image.Rect(0, 0, 4, 4)
	dial, hands := image.NewRGBA(box), image.NewRGBA(box)
	for y := range box.Dy() {
		for x := range box.Dx() {
			dial.SetRGBA(x, y, face)
		}
		hands.SetRGBA(1, y, needle)
	}
	if got := sampleDialInk(dial, hands, 0, 0, 4, 4); got != needle {
		t.Errorf("thin hand shows as %v, want %v", got, needle)
	}
	hands.SetRGBA(1, 0, color.RGBA{})
	if got := sampleDialInk(dial, hands, 0, 0, 4, 4); got != face {
		t.Errorf("trace of hand shows as %v, want %v", got, face)
	}
}

func TestAnalogResizeAndToggle(t *testing.T) {
	m := &model{opts: Options{AMPM: true, FA: true, Face: AnalogANSI}, analog: true, profile: termenv.TrueColor}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if m.visible != AnalogANSI {
		t.Fatal("expected analog on startup")
	}
	m.now = time.Date(2026, time.November, 22, 19, 30, 0, 0, time.UTC)
	view := m.View()
	if !strings.Contains(view, "۰۷:۳۰:۰۰ PM") || !strings.ContainsAny(view, "▀▄") {
		t.Fatal("expected localized readout next to the cell dial")
	}
	if w, h := lipgloss.Size(view); w != 80 || h != 24 {
		t.Fatalf("analog view size %dx%d", w, h)
	}
	oldTick := m.tickID
	m.Update(tea.WindowSizeMsg{Width: 19, Height: 4})
	if m.visible != Digital || !m.analog || lipgloss.Width(m.View()) > 19 {
		t.Fatal("expected a compact digital view while the analog dial does not fit")
	}
	if _, cmd := m.Update(tickMsg(oldTick)); cmd != nil {
		t.Fatal("stale analog tick started a new chain")
	}
	m.Update(tea.WindowSizeMsg{Width: 47, Height: 29})
	if m.visible != AnalogANSI || !m.layout.stacked {
		t.Fatal("expected the analog face and stacked layout to return")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if m.analog || m.visible != Digital {
		t.Fatal("expected a to switch to digital")
	}
}

func TestProbeReply(t *testing.T) {
	const id = 42
	da := "\x1b[?62;4;22c"
	for _, tt := range []struct {
		response string
		ok, done bool
	}{
		{fmt.Sprintf("a\x1b_Gi=%d;OK\x1b\\%sq", id, da), true, true},
		{fmt.Sprintf("a\x1b_Gi=%d;EINVAL\x1b\\%sq", id, da), false, true},
		{"a" + da + "q", false, true},
		{"aq", false, false},
	} {
		ok, done, keys := probeReply([]byte(tt.response), id)
		if ok != tt.ok || done != tt.done || string(keys) != "aq" {
			t.Fatalf("probe %q got support=%t complete=%t keys=%q", tt.response, ok, done, keys)
		}
	}
}
