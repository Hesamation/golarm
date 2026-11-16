// Package duration parses user-facing duration and time-of-day strings.
package duration

import (
	"fmt"
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
		secs, err := strconv.ParseFloat(s, 64)
		if err != nil || secs <= 0 {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		return time.Duration(secs * float64(time.Second)), nil
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	return d, nil
}

func parseColon(parts []string) (time.Duration, error) {
	nums := make([]int64, len(parts))
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || v < 0 {
			return 0, fmt.Errorf("invalid duration %q", strings.Join(parts, ":"))
		}
		nums[i] = int64(v)
	}
	var d time.Duration
	switch len(nums) {
	case 2: // mm:ss
		d = time.Duration(nums[0])*time.Minute + time.Duration(nums[1])*time.Second
	case 3: // hh:mm:ss
		d = time.Duration(nums[0])*time.Hour + time.Duration(nums[1])*time.Minute + time.Duration(nums[2])*time.Second
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	return d, nil
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
	switch meridiem {
	case 1:
		if h < 1 || h > 12 || m > 59 {
			return 0, 0, fmt.Errorf("invalid time %q", orig)
		}
		if h == 12 {
			h = 0
		}
	case 2:
		if h < 1 || h > 12 || m > 59 {
			return 0, 0, fmt.Errorf("invalid time %q", orig)
		}
		if h != 12 {
			h += 12
		}
	default:
		if h > 23 || m > 59 {
			return 0, 0, fmt.Errorf("invalid time %q", orig)
		}
	}
	return h, m, nil
}
