// Package duration parses user-facing duration and time-of-day strings.
package duration

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Parse accepts Go duration syntax (90s, 5m, 1h30m), bare numbers as
// seconds (90), and colon forms (10:00 = mm:ss, 1:30:00 = hh:mm:ss).
func Parse(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if !strings.ContainsAny(s, "hms") {
		parts := strings.Split(s, ":")
		if len(parts) >= 2 && len(parts) <= 3 {
			return parseColon(parts)
		}
		if d, err := time.ParseDuration(s + "s"); err == nil && d > 0 {
			return d, nil
		}
		// Keep numeric forms such as exponent notation that ParseDuration rejects.
		secs, err := strconv.ParseFloat(s, 64)
		nanos := secs * float64(time.Second)
		if err != nil || math.IsNaN(nanos) || nanos < 1 || nanos >= 1<<63 {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		return time.Duration(nanos), nil
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	return d, nil
}

func parseColon(parts []string) (time.Duration, error) {
	const maxSeconds = math.MaxInt64 / int64(time.Second)
	var seconds int64
	for _, part := range parts {
		v, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || v < 0 || v > maxSeconds || seconds > (maxSeconds-v)/60 {
			return 0, fmt.Errorf("invalid duration %q", strings.Join(parts, ":"))
		}
		seconds = seconds*60 + v
	}
	if seconds == 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	return time.Duration(seconds) * time.Second, nil
}

// ParseTimeOfDay parses HH:MM in 24-hour form, or with an am/pm suffix
// (7:30pm, 7:30 pm, 07:30AM). Returns the hour (0-23) and minute (0-59).
func ParseTimeOfDay(s string) (hour, minute int, err error) {
	orig := s
	s = strings.ToLower(strings.TrimSpace(s))
	meridiem := 0 // 0 = none, 1 = am, 2 = pm
	if strings.HasSuffix(s, "am") {
		meridiem = 1
		s = strings.TrimSpace(strings.TrimSuffix(s, "am"))
	} else if strings.HasSuffix(s, "pm") {
		meridiem = 2
		s = strings.TrimSpace(strings.TrimSuffix(s, "pm"))
	}
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time %q — use HH:MM", orig)
	}
	h, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	m, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid time %q — use HH:MM", orig)
	}
	if h < 0 || h > 23 || m < 0 || m > 59 || (meridiem != 0 && (h < 1 || h > 12)) {
		return 0, 0, fmt.Errorf("invalid time %q", orig)
	}
	if meridiem != 0 {
		h %= 12
		if meridiem == 2 {
			h += 12
		}
	}
	return h, m, nil
}
