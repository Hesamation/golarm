package timer

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestAlarmTargetLifecycle(t *testing.T) {
	now := time.Date(2026, time.November, 24, 7, 28, 10, 0, time.UTC)
	m := newModel(Options{Target: NextAlarm(now, 7, 30)}, now)
	m.key("+", now.Add(5*time.Second))
	target := time.Date(2026, time.November, 24, 7, 31, 0, 0, time.UTC)
	if !m.deadline.Equal(target) || m.remaining != 2*time.Minute+45*time.Second || !strings.Contains(m.View(), "fires at 07:31") {
		t.Fatalf("adjusted alarm deadline = %v, remaining = %v", m.deadline, m.remaining)
	}

	nearlyDue := target.Add(-20 * time.Second)
	m.key("-", nearlyDue)
	m.key(" ", nearlyDue)
	if !m.deadline.Equal(target) || m.paused {
		t.Fatalf("pending alarm deadline = %v, paused = %t", m.deadline, m.paused)
	}
	if cmd := m.tick(nearlyDue); cmd != nil || m.remaining != 20*time.Second {
		t.Fatalf("alarm fired before its target or counted down to %v", m.remaining)
	}
	if cmd := m.tick(target.Add(time.Minute)); cmd == nil || !m.alarming || m.remaining != 0 {
		t.Fatal("expected the alarm to fire on the first tick after its target")
	}
	if cmd := m.tick(target.Add(2 * time.Minute)); cmd != nil {
		t.Fatal("alarm fired more than once")
	}
	m.key(" ", target.Add(2*time.Minute))
	if !m.dismissed {
		t.Fatal("space did not dismiss a fired alarm")
	}
	m.key("r", target.Add(3*time.Minute))
	if m.alarming || m.dismissed || !m.deadline.Equal(target.AddDate(0, 0, 1)) {
		t.Fatalf("alarm failed to re-arm for tomorrow at the adjusted time: %+v", m)
	}
}

func TestNextAlarmAcrossClockChanges(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name   string
		now    time.Time
		hour   int
		minute int
		want   time.Time
	}{
		{"skipped hour", time.Date(2026, time.March, 8, 6, 59, 0, 0, time.UTC), 2, 30,
			time.Date(2026, time.March, 9, 6, 30, 0, 0, time.UTC)},
		{"repeated hour", time.Date(2026, time.November, 1, 5, 45, 0, 0, time.UTC), 1, 30,
			time.Date(2026, time.November, 1, 6, 30, 0, 0, time.UTC)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := NextAlarm(tt.now.In(location), tt.hour, tt.minute); !got.Equal(tt.want) {
				t.Errorf("NextAlarm = %v, want %v", got.UTC(), tt.want)
			}
		})
	}
}

func TestTimerPauseAndRestart(t *testing.T) {
	now := time.Date(2026, time.November, 24, 7, 28, 10, 0, time.UTC)
	m := newModel(Options{Total: 90 * time.Second}, now)
	m.key(" ", now.Add(10*time.Second))
	resumed := now.Add(time.Hour)
	m.tick(resumed)
	if !m.paused || m.remaining != 80*time.Second {
		t.Fatalf("paused timer remaining = %v", m.remaining)
	}
	m.key("+", resumed)
	m.key(" ", resumed)
	m.tick(resumed.Add(5 * time.Second))
	if m.paused || m.remaining != 135*time.Second {
		t.Fatalf("resumed timer remaining = %v", m.remaining)
	}
	m.key("r", resumed.Add(5*time.Second))
	if m.remaining != 150*time.Second {
		t.Fatalf("restarted timer remaining = %v", m.remaining)
	}
}

func TestTimerFitsNarrowPane(t *testing.T) {
	now := time.Date(2026, time.November, 24, 7, 28, 10, 0, time.UTC)
	m := newModel(Options{Total: 65 * time.Second, Label: "Tea\nsteep"}, now)
	m.width, m.height = 36, 12
	view := m.View()
	if w, h := lipgloss.Size(view); w > m.width || h > m.height {
		t.Fatalf("view size = %d×%d, want at most %d×%d", w, h, m.width, m.height)
	}
	for _, text := range []string{"00:01:05", "Tea steep", "q quit"} {
		if !strings.Contains(view, text) {
			t.Errorf("view %q does not contain %q", view, text)
		}
	}
	m.tick(now.Add(64*time.Second + 500*time.Millisecond))
	if !strings.Contains(m.View(), "00:00:01") {
		t.Fatal("timer displayed zero before it expired")
	}
}
