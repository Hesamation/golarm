package clock

import (
	"fmt"
	"math"
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
			face, l := chooseDial(tt.width, tt.height, tt.cell, tt.requested, tt.native, tt.profile)
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
