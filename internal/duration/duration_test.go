package duration

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"90", 90 * time.Second, true},
		{"90s", 90 * time.Second, true},
		{"5m", 5 * time.Minute, true},
		{"1h30m", 90 * time.Minute, true},
		{"1h 30m", 0, false}, // ParseDuration rejects spaces
		{"10:00", 10 * time.Minute, true},
		{"0:90", 90 * time.Second, true},
		{"1:30:00", 90 * time.Minute, true},
		{"2:05:03", 2*time.Hour + 5*time.Minute + 3*time.Second, true},
		{"0", 0, false},
		{"-5", 0, false},
		{"abc", 0, false},
		{"", 0, false},
		{"1:2:3:4", 0, false},
		{"5", 5 * time.Second, true},
		{"0.5", 500 * time.Millisecond, true},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if c.ok && (err != nil || got != c.want) {
			t.Errorf("Parse(%q) = %v, %v; want %v", c.in, got, err, c.want)
		}
		if !c.ok && err == nil {
			t.Errorf("Parse(%q) = %v; want error", c.in, got)
		}
	}
}

func TestParseTimeOfDay(t *testing.T) {
	cases := []struct {
		in   string
		hour int
		min  int
		ok   bool
	}{
		{"07:30", 7, 30, true},
		{"7:30", 7, 30, true},
		{"19:45", 19, 45, true},
		{"0:00", 0, 0, true},
		{"23:59", 23, 59, true},
		{"7:30pm", 19, 30, true},
		{"7:30 pm", 19, 30, true},
		{"12am", 0, 0, false}, // needs minutes
		{"12:00am", 0, 0, true},
		{"12:00pm", 12, 0, true},
		{"7:30PM", 19, 30, true},
		{"24:00", 0, 0, false},
		{"7:60", 0, 0, false},
		{"7:30xm", 0, 0, false},
		{"noon", 0, 0, false},
	}
	for _, c := range cases {
		h, m, err := ParseTimeOfDay(c.in)
		if c.ok && (err != nil || h != c.hour || m != c.min) {
			t.Errorf("ParseTimeOfDay(%q) = %d:%02d, %v; want %d:%02d", c.in, h, m, err, c.hour, c.min)
		}
		if !c.ok && err == nil {
			t.Errorf("ParseTimeOfDay(%q) = %d:%02d; want error", c.in, h, m)
		}
	}
}
