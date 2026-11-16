package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Hesamation/golarm/internal/chrono"
	"github.com/Hesamation/golarm/internal/clock"
	"github.com/Hesamation/golarm/internal/timer"
)

var version = "dev"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		clock.Run(false)
		return
	}
	switch args[0] {
	case "-12":
		clock.Run(true)
	case "-v", "--version":
		fmt.Println("golarm " + version)
	case "-h", "--help":
		printUsage(os.Stdout)
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
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "golarm — terminal clock, timer & stopwatch")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  golarm              24h clock")
	fmt.Fprintln(w, "  golarm -12          12h clock")
	fmt.Fprintln(w, "  golarm timer 5m30s  countdown timer")
	fmt.Fprintln(w, "  golarm chrono       stopwatch")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Keys:  space  pause/resume (timer & chrono)    q / Esc  quit")
}
