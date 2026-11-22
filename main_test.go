package main

import (
	"strings"
	"testing"

	"github.com/Hesamation/golarm/internal/clock"
)

func TestAnalogClockFlag(t *testing.T) {
	for args, want := range map[string]clock.Face{
		"":                    clock.Digital,
		"--analog":            clock.Analog,
		"clock --analog":      clock.Analog,
		"--analog clock":      clock.Analog,
		"clock --analog=ansi": clock.AnalogANSI,
		"--analog=ansi clock": clock.AnalogANSI,
	} {
		t.Run(args, func(t *testing.T) {
			var opts options
			cmd, rest := parseGlobal(strings.Fields(args), &opts)
			if cmd != "clock" {
				t.Fatalf("got command %q", cmd)
			}
			parseClockFlags(rest, &opts)
			if opts.face != want {
				t.Fatalf("got face %d, want %d", opts.face, want)
			}
		})
	}
}

func TestClockCalendarOptions(t *testing.T) {
	for _, tt := range []struct {
		args, env string
		want      clock.Calendar
		wantError string
	}{
		{want: clock.Gregorian},
		{args: "--fa", want: clock.Gregorian},
		{args: "--calendar shamsi clock", want: clock.Shamsi},
		{args: "clock --calendar=both", want: clock.BothCalendars},
		{args: "clock --calendar shamsi --analog", want: clock.Shamsi},
		{args: "--calendar=both", want: clock.BothCalendars},
		{env: "shamsi", want: clock.Shamsi},
		{args: "--calendar gregorian", env: "shamsi", want: clock.Gregorian},
		{args: "--calendar both clock --calendar gregorian", env: "invalid", want: clock.Gregorian},
		{env: "invalid", wantError: `GOLARM_CALENDAR: invalid calendar "invalid"`},
	} {
		t.Run(tt.args+"/"+tt.env, func(t *testing.T) {
			var opts options
			cmd, rest := parseGlobal(strings.Fields(tt.args), &opts)
			if cmd != "clock" {
				t.Fatalf("got command %q", cmd)
			}
			parseClockFlags(rest, &opts)
			got, err := calendarForClock(opts.calendar, tt.env)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("got error %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got calendar %q and error %v, want %q", got, err, tt.want)
			}
		})
	}
}
