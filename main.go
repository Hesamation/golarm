package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Hesamation/golarm/internal/chrono"
	"github.com/Hesamation/golarm/internal/clock"
	"github.com/Hesamation/golarm/internal/duration"
	"github.com/Hesamation/golarm/internal/timer"
	"github.com/Hesamation/golarm/internal/ui"
)

var version = "dev"

type options struct {
	ampm     bool
	fa       bool
	face     clock.Face
	calendar clock.Calendar
	dial     clock.Dial
	dialOnly bool
	accent   string
	theme    string
	label    string
	quiet    bool
}

func main() {
	opts := options{
		ampm:  os.Getenv("GOLARM_12H") == "1",
		fa:    os.Getenv("GOLARM_FA") == "1",
		theme: os.Getenv("GOLARM_THEME"),
	}

	cmd, args := parseGlobal(os.Args[1:], &opts)
	rejectClockOptions(cmd, opts)

	if opts.theme != "" {
		t, ok := ui.FromName(opts.theme)
		if !ok {
			names := make([]string, 0, len(ui.Themes))
			for name := range ui.Themes {
				names = append(names, name)
			}
			fmt.Fprintf(os.Stderr, "unknown theme %q — available: %s\n", opts.theme, strings.Join(names, ", "))
			os.Exit(1)
		}
		ui.SetTheme(t)
	}

	switch cmd {
	case "clock":
		parseClockFlags(args, &opts)
		calendar, err := calendarForClock(opts.calendar, os.Getenv("GOLARM_CALENDAR"))
		if err != nil {
			fatal(err.Error())
		}
		face := opts.face
		if face == clock.Digital && (opts.dial != "" || opts.dialOnly) {
			face = clock.Analog
		}
		clock.Run(clock.Options{
			AMPM: opts.ampm, FA: opts.fa, Face: face, Calendar: calendar,
			Dial: opts.dial, DialOnly: opts.dialOnly, Accent: opts.accent,
		})
	case "timer":
		dur, o := parseTimerArgs(args, &opts)
		timer.Run(timer.Options{Total: dur, Label: o.label, Quiet: o.quiet})
	case "chrono":
		parseClockFlags(args, &opts)
		rejectClockOptions(cmd, opts)
		chrono.Run()
	case "alarm":
		o, target := parseAlarmArgs(args, &opts)
		timer.Run(timer.Options{Total: time.Until(target), Label: o.label, Target: &target})
	default:
		if strings.HasPrefix(cmd, "-") {
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n\n", cmd)
		} else {
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		}
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func rejectClockOptions(cmd string, opts options) {
	if cmd == "clock" {
		return
	}
	if opts.face != clock.Digital {
		fatal("--analog is only available for the clock")
	}
	if opts.calendar != "" {
		fatal("--calendar is only available for the clock")
	}
	if opts.dial != "" || opts.dialOnly || opts.accent != "" {
		fatal("--dial, --dial-only, and --accent are only available for the clock")
	}
}

// parseGlobal walks args until the first non-flag token (the subcommand)
// and applies global flags. Returns the subcommand (default "clock") and
// the remaining args for per-command parsing.
func parseGlobal(args []string, opts *options) (cmd string, rest []string) {
	cmd = "clock"
	for i := 0; i < len(args); i++ {
		if used := parseClockOption(args[i:], opts); used != 0 {
			i += used - 1
			continue
		}
		a := args[i]
		switch a {
		case "-v", "--version":
			fmt.Println("golarm " + version)
			os.Exit(0)
		case "-h", "--help":
			printUsage(os.Stdout)
			os.Exit(0)
		case "--theme":
			if i++; i >= len(args) {
				fatal("missing value for --theme")
			}
			opts.theme = args[i]
		default:
			if strings.HasPrefix(a, "-") {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n\n", a)
				printUsage(os.Stderr)
				os.Exit(1)
			}
			return a, args[i+1:]
		}
	}
	return cmd, nil
}

func parseClockFlags(args []string, opts *options) {
	for len(args) > 0 {
		if used := parseClockOption(args, opts); used != 0 {
			args = args[used:]
			continue
		}
		if args[0] == "--theme" {
			// consumed globally only; reject here to avoid surprises
			fatal("--theme must come before the command")
		}
		fmt.Fprintf(os.Stderr, "unknown argument: %s\n\n", args[0])
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

// parseClockOption consumes shared clock options on either side of the command.
func parseClockOption(args []string, opts *options) int {
	switch args[0] {
	case "-12", "--12h", "--ampm":
		opts.ampm = true
	case "-fa", "--fa":
		opts.fa = true
	case "--analog":
		opts.face = clock.Analog
	case "--analog=ansi":
		opts.face = clock.AnalogANSI
	case "--dial-only":
		opts.dialOnly = true
	default:
		name, value, joined := strings.Cut(args[0], "=")
		if name != "--calendar" && name != "--dial" && name != "--accent" {
			return 0
		}
		used := 1
		if !joined {
			if len(args) < 2 || strings.HasPrefix(args[1], "-") {
				if name == "--calendar" {
					fatal("--calendar requires a value (gregorian, shamsi, or both)")
				}
				fatal(name + " requires a value")
			}
			used, value = 2, args[1]
		}
		switch name {
		case "--calendar":
			calendar, err := parseCalendar(value)
			if err != nil {
				fatal(err.Error())
			}
			opts.calendar = calendar
		case "--dial":
			opts.dial = clock.Dial(value)
			switch opts.dial {
			case clock.ClassicDial, clock.SectorDial, clock.MinimalDial:
			default:
				fatal(fmt.Sprintf("invalid dial %q (choose classic, sector, or minimal)", value))
			}
		case "--accent":
			if !validAccent(value) {
				fatal(fmt.Sprintf("invalid accent %q (use an RGB color such as #cdb1ff)", value))
			}
			opts.accent = value
		}
		return used
	}
	return 1
}

func validAccent(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}
	for _, c := range value[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}

func parseCalendar(value string) (clock.Calendar, error) {
	calendar := clock.Calendar(value)
	switch calendar {
	case clock.Gregorian, clock.Shamsi, clock.BothCalendars:
		return calendar, nil
	default:
		return "", fmt.Errorf("invalid calendar %q (choose gregorian, shamsi, or both)", value)
	}
}

func calendarForClock(selected clock.Calendar, env string) (clock.Calendar, error) {
	if selected != "" {
		return selected, nil
	}
	if env == "" {
		return clock.Gregorian, nil
	}
	calendar, err := parseCalendar(env)
	if err != nil {
		return "", fmt.Errorf("GOLARM_CALENDAR: %w", err)
	}
	return calendar, nil
}

func parseTimerArgs(args []string, opts *options) (time.Duration, options) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: golarm timer <duration>")
		fmt.Fprintln(os.Stderr, "example: golarm timer 5m30s")
		os.Exit(1)
	}
	dur, err := duration.Parse(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad duration %q — try: 90  90s  5m  1h30m  10:00\n", args[0])
		os.Exit(1)
	}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-q", "--quiet":
			opts.quiet = true
		case "-l", "--label":
			if i++; i >= len(args) {
				fatal("missing value for --label")
			}
			opts.label = args[i]
		default:
			fmt.Fprintf(os.Stderr, "unknown argument: %s\n\n", args[i])
			printUsage(os.Stderr)
			os.Exit(1)
		}
	}
	return dur, *opts
}

func parseAlarmArgs(args []string, opts *options) (options, time.Time) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: golarm alarm <HH:MM>")
		fmt.Fprintln(os.Stderr, "example: golarm alarm 7:30pm")
		os.Exit(1)
	}
	h, m, err := duration.ParseTimeOfDay(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-l", "--label":
			if i++; i >= len(args) {
				fatal("missing value for --label")
			}
			opts.label = args[i]
		default:
			fmt.Fprintf(os.Stderr, "unknown argument: %s\n\n", args[i])
			printUsage(os.Stderr)
			os.Exit(1)
		}
	}
	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
	if !target.After(now) {
		target = target.AddDate(0, 0, 1)
	}
	return *opts, target
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "golarm — terminal clock, timer, stopwatch & alarm")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  golarm [clock] [-12] [--fa] [--analog[=ansi]] [--calendar MODE]")
	fmt.Fprintln(w, "        [--dial classic|sector|minimal] [--dial-only] [--accent '#RRGGBB']")
	fmt.Fprintln(w, "  golarm timer 5m30s [-l L] [-q]  countdown timer")
	fmt.Fprintln(w, "  golarm chrono                   stopwatch with laps")
	fmt.Fprintln(w, "  golarm alarm 7:30 [-l L]        fire at a wall-clock time")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Durations: 90 | 90s | 5m | 1h30m | 10:00 | 1:30:00")
	fmt.Fprintln(w, "Themes:    default, catppuccin, gruvbox, tokyo-night")
	fmt.Fprintln(w, "Calendars: gregorian (default), shamsi, both")
	fmt.Fprintln(w, "Env:       GOLARM_12H=1  GOLARM_FA=1  GOLARM_THEME=<name>")
	fmt.Fprintln(w, "           GOLARM_CALENDAR=<mode>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Keys:  space pause/dismiss    r restart/re-arm    l lap (chrono)")
	fmt.Fprintln(w, "       + / - adjust timer by 1m    a switch analog/digital    q / Esc quit")
	fmt.Fprintln(w, "       f cycle analog dials    d toggle dial-only view")
}
