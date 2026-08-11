package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/user/golarm/internal/chrono"
	"github.com/user/golarm/internal/clock"
	"github.com/user/golarm/internal/timer"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		clock.Run(false)
		return
	}

	switch args[0] {
	case "-12":
		clock.Run(true)
	case "-h", "--help":
		printUsage()
	case "timer":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: golarm timer <duration>")
			fmt.Fprintln(os.Stderr, "example: golarm timer 5m30s")
			os.Exit(1)
		}
		dur, err := time.ParseDuration(args[1])
		if err != nil || dur <= 0 {
			fmt.Fprintf(os.Stderr, "bad duration %q — try: 90s  5m  1h30m\n", args[1])
			os.Exit(1)
		}
		timer.Run(dur)
	case "chrono":
		chrono.Run()
	default:
		if strings.HasPrefix(args[0], "-") {
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n\n", args[0])
		} else {
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		}
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "golarm — terminal clock, timer & stopwatch")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  golarm              24h clock")
	fmt.Fprintln(os.Stderr, "  golarm -12          12h clock")
	fmt.Fprintln(os.Stderr, "  golarm timer 5m30s  countdown timer")
	fmt.Fprintln(os.Stderr, "  golarm chrono       stopwatch")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Keys:  space  pause/resume    q / Esc  quit")
}
