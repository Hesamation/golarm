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
