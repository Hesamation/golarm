package shamsi

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// mustOpenGolden opens the golden fixture generated from the jalaali-js
// reference implementation (every day 1970-2100).
func mustOpenGolden(t *testing.T) *bufio.Scanner {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "golden_jalaali.csv.gz"))
	if err != nil {
		t.Fatalf("open golden fixture: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	return bufio.NewScanner(gz)
}

// TestGoldenJalaaliJS cross-checks every day of 1970-2100 against values
// generated from the jalaali-js reference implementation, which itself was
// verified against the official Iranian calendar.
func TestGoldenJalaaliJS(t *testing.T) {
	sc := mustOpenGolden(t)
	sc.Scan() // header
	bad := 0
	for sc.Scan() {
		parts := strings.Split(sc.Text(), ",")
		if len(parts) != 6 {
			t.Fatalf("malformed golden row %q", sc.Text())
		}
		gy, _ := strconv.Atoi(parts[0])
		gm, _ := strconv.Atoi(parts[1])
		gd, _ := strconv.Atoi(parts[2])
		jy, _ := strconv.Atoi(parts[3])
		jm, _ := strconv.Atoi(parts[4])
		jd, _ := strconv.Atoi(parts[5])
		got := FromGregorian(gy, gm, gd)
		if got.Year != jy || got.Month != jm || got.Day != jd {
			bad++
			if bad <= 5 {
				t.Errorf("FromGregorian(%d,%d,%d) = %d/%d/%d, want %d/%d/%d",
					gy, gm, gd, got.Year, got.Month, got.Day, jy, jm, jd)
			}
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan golden: %v", err)
	}
	if bad > 0 {
		t.Fatalf("%d of 47847 golden days mismatched", bad)
	}
}

// TestGoldenWeekday verifies the Shamsi weekday convention (0=Shanbe) for
// every day of the golden range using the stdlib as ground truth.
func TestGoldenWeekday(t *testing.T) {
	for gy := 1970; gy <= 2100; gy++ {
		for _, date := range []struct{ m, d int }{{1, 1}, {3, 21}, {6, 15}, {12, 31}} {
			tt := time.Date(gy, time.Month(date.m), date.d, 0, 0, 0, 0, time.UTC)
			want := (int(tt.Weekday()) + 1) % 7 // Go Sun=0 → Shanbe=Saturday
			got := FromGregorian(gy, date.m, date.d).WDay
			if got != want {
				t.Errorf("weekday(%d-%02d-%02d) = %d, want %d", gy, date.m, date.d, got, want)
			}
		}
	}
}

// TestOfficialFarvardinFirst checks Farvardin 1 dates against the official
// Iranian calendar (astronomical, per the equinox-before-noon Tehran rule),
// as tabulated on Wikipedia for the 33-year cycle correspondence.
func TestOfficialFarvardinFirst(t *testing.T) {
	cases := []struct {
		jy   int
		leap bool
		gy   int
		gm   int
		gd   int
	}{
		{1395, true, 2016, 3, 20},
		{1396, false, 2017, 3, 21},
		{1397, false, 2018, 3, 21},
		{1398, false, 2019, 3, 21},
		{1399, true, 2020, 3, 20},
		{1400, false, 2021, 3, 21},
		{1401, false, 2022, 3, 21},
		{1402, false, 2023, 3, 21},
		{1403, true, 2024, 3, 20},
		{1404, false, 2025, 3, 21},
		{1405, false, 2026, 3, 21},
		{1406, false, 2027, 3, 21},
		{1407, false, 2028, 3, 20},
		{1408, true, 2029, 3, 20},
		{1409, false, 2030, 3, 21},
		{1410, false, 2031, 3, 21},
	}
	for _, c := range cases {
		gy, march, _ := jalCal(c.jy)
		if gy != c.gy || march != c.gd {
			t.Errorf("jalCal(%d): Farvardin 1 = %d March %d, want %d March %d", c.jy, gy, march, c.gy, c.gd)
		}
		if got := IsLeapYear(c.jy); got != c.leap {
			t.Errorf("IsLeapYear(%d) = %v, want %v", c.jy, got, c.leap)
		}
	}
}

// TestOfficialLeapYears checks the official modern-era leap year list.
func TestOfficialLeapYears(t *testing.T) {
	leap := []int{1354, 1358, 1362, 1366, 1370, 1375, 1379, 1383, 1387, 1391, 1395, 1399, 1403, 1408, 1412, 1416, 1420}
	for _, jy := range leap {
		if !IsLeapYear(jy) {
			t.Errorf("IsLeapYear(%d) = false, want true (official leap year)", jy)
		}
	}
	for _, jy := range []int{1394, 1396, 1397, 1398, 1400, 1401, 1402, 1404, 1405, 1406, 1407, 1409} {
		if IsLeapYear(jy) {
			t.Errorf("IsLeapYear(%d) = true, want false", jy)
		}
	}
}

// TestPreviouslyBrokenDates regresses the concrete failures found during the
// review of the old hand-rolled implementation.
func TestPreviouslyBrokenDates(t *testing.T) {
	cases := []struct {
		gy, gm, gd int
		want       string
	}{
		{2024, 3, 20, "1 Farvardin 1403"}, // old: "2  1402" (month index 13)
		{2025, 3, 20, "30 Esfand 1403"},   // 1403 is leap; old: "1  1403"
		{2025, 3, 21, "1 Farvardin 1404"}, // Nowruz 1404
		{2027, 3, 20, "29 Esfand 1405"},   // 1405 is not leap; old: "30 Esfand 1405"
		{2028, 3, 20, "1 Farvardin 1407"}, // Nowruz 1407 on March 20; old: "2  1406"
		{1971, 1, 1, "11 Dey 1349"},       // old was off by one day (Dey 12)
		{2026, 11, 16, "25 Aban 1405"},    // today at review time
	}
	for _, c := range cases {
		got := FromGregorian(c.gy, c.gm, c.gd).FormatFull()
		if got != c.want {
			t.Errorf("FromGregorian(%d,%d,%d).FormatFull() = %q, want %q", c.gy, c.gm, c.gd, got, c.want)
		}
	}
}

// TestDaysInMonthAndFormatters covers month lengths and the display helpers.
func TestDaysInMonthAndFormatters(t *testing.T) {
	for m := 1; m <= 6; m++ {
		if got := DaysInMonth(1405, m); got != 31 {
			t.Errorf("DaysInMonth(1405,%d) = %d, want 31", m, got)
		}
	}
	for m := 7; m <= 11; m++ {
		if got := DaysInMonth(1405, m); got != 30 {
			t.Errorf("DaysInMonth(1405,%d) = %d, want 30", m, got)
		}
	}
	if got := DaysInMonth(1403, 12); got != 30 {
		t.Errorf("DaysInMonth(1403,12) = %d, want 30 (leap Esfand)", got)
	}
	if got := DaysInMonth(1405, 12); got != 29 {
		t.Errorf("DaysInMonth(1405,12) = %d, want 29", got)
	}
	if got := DaysInMonth(1405, 0); got != 0 {
		t.Errorf("DaysInMonth(1405,0) = %d, want 0", got)
	}

	d := FromGregorian(2026, 11, 16)
	if d.FormatCompact() != "25 Aba 1405" {
		t.Errorf("FormatCompact = %q, want %q", d.FormatCompact(), "25 Aba 1405")
	}
	if d.WeekDayName() != "Doshanbe" {
		t.Errorf("WeekDayName = %q, want Doshanbe (Monday)", d.WeekDayName())
	}
	if fmt.Sprint(d.WDay) != "2" {
		t.Errorf("WDay = %d, want 2", d.WDay)
	}
}
