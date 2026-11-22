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
	ampm  bool
	fa    bool
	face  clock.Face
	theme string
	label string
	quiet bool
}

func main() {
	opts := options{
		ampm:  os.Getenv("GOLARM_12H") == "1",
		fa:    os.Getenv("GOLARM_FA") == "1",
		theme: os.Getenv("GOLARM_THEME"),
	}

	cmd, args := parseGlobal(os.Args[1:], &opts)
	if cmd != "clock" && opts.face != clock.Digital {
		fatal("--analog is only available for the clock")
	}

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
		clock.Run(clock.Options{AMPM: opts.ampm, FA: opts.fa, Face: opts.face})
	case "timer":
		dur, o := parseTimerArgs(args, &opts)
		timer.Run(timer.Options{Total: dur, Label: o.label, Quiet: o.quiet})
	case "chrono":
		parseClockFlags(args, &opts)
		if opts.face != clock.Digital {
			fatal("--analog is only available for the clock")
		}
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

// parseGlobal walks args until the first non-flag token (the subcommand)
// and applies global flags. Returns the subcommand (default "clock") and
// the remaining args for per-command parsing.
func parseGlobal(args []string, opts *options) (cmd string, rest []string) {
	cmd = "clock"
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "-12", "--12h", "--ampm":
			opts.ampm = true
		case "-fa", "--fa":
			opts.fa = true
		case "--analog":
			opts.face = clock.Analog
		case "--analog=ansi":
			opts.face = clock.AnalogANSI
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
	for _, a := range args {
		switch a {
		case "-12", "--12h", "--ampm":
			opts.ampm = true
		case "-fa", "--fa":
			opts.fa = true
		case "--analog":
			opts.face = clock.Analog
		case "--analog=ansi":
			opts.face = clock.AnalogANSI
		case "--theme":
			// consumed globally only; reject here to avoid surprises
			fatal("--theme must come before the command")
		default:
			fmt.Fprintf(os.Stderr, "unknown argument: %s\n\n", a)
			printUsage(os.Stderr)
			os.Exit(1)
		}
	}
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
	fmt.Fprintln(w, "  golarm [clock] [-12] [--fa] [--analog[=ansi]]")
	fmt.Fprintln(w, "                                 live clock with optional analog dial")
	fmt.Fprintln(w, "  golarm timer 5m30s [-l L] [-q]  countdown timer")
	fmt.Fprintln(w, "  golarm chrono                   stopwatch with laps")
	fmt.Fprintln(w, "  golarm alarm 7:30 [-l L]        fire at a wall-clock time")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Durations: 90 | 90s | 5m | 1h30m | 10:00 | 1:30:00")
	fmt.Fprintln(w, "Themes:    default, catppuccin, gruvbox, tokyo-night")
	fmt.Fprintln(w, "Env:       GOLARM_12H=1  GOLARM_FA=1  GOLARM_THEME=<name>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Keys:  space pause/dismiss    r restart/re-arm    l lap (chrono)")
	fmt.Fprintln(w, "       + / - adjust timer by 1m    a switch clock face    q / Esc quit")
}
