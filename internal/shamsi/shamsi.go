// Package shamsi provides conversion between Gregorian and Shamsi (Solar
// Hijri) dates.
//
// The conversion uses the Borkowski 33-year-cycle algorithm as implemented
// by jalaali-js (validated against the official Iranian calendar, which is
// astronomical rather than arithmetic, for the modern era).
package shamsi

import "fmt"

// Shamsi months in Roman letters.
var Months = [12]string{
	"Farvardin", "Ordibehesht", "Khordad",
	"Tir", "Mordad", "Shahrivar",
	"Mehr", "Aban", "Azar",
	"Dey", "Bahman", "Esfand",
}

// Short month names for compact display.
var ShortMonths = [12]string{
	"Far", "Ord", "Kho",
	"Tor", "Mor", "Sha",
	"Meh", "Aba", "Aza",
	"Dey", "Bah", "Esf",
}

// Days of week in Roman letters. The Shamsi week starts on Shanbe (Saturday).
var WeekDays = [7]string{
	"Shanbe", "Yekshanbe", "Doshanbe",
	"Sehshanbe", "Chaharshanbe", "Panjshanbe", "Jome",
}

// Short day names.
var ShortWeekDays = [7]string{
	"Shan", "Yek", "Dos",
	"Seh", "Cha", "Pan", "Jom",
}

// Date is a Shamsi (Jalaali) calendar date.
type Date struct {
	Year  int
	Month int // 1-12
	Day   int // 1-31
	WDay  int // 0=Shanbe, 6=Jome
}

// jalBreaks holds the Jalaali years that begin a new 33-year leap cycle
// (Borkowski), as used by jalaali-js.
var jalBreaks = [...]int{
	-61, 9, 38, 199, 426, 686, 756, 818, 1111, 1181, 1210,
	1635, 2060, 2097, 2192, 2262, 2324, 2394, 2456, 3178,
}

var (
	minJalaaliYear = jalBreaks[0]
	maxJalaaliYear = jalBreaks[len(jalBreaks)-1] - 1
)

// div divides a by b, truncating toward zero (Go's native integer division).
func div(a, b int) int { return a / b }

// md returns a mod b with Go's truncated-division semantics.
func md(a, b int) int { return a - a/b*b }

// jalCal returns the Gregorian year, the day of March on which Farvardin 1
// of the given Jalaali year falls, and the leap-cycle state (0 means the
// year itself is leap). Panics for years outside the supported range.
func jalCal(jy int) (gy, march, leap int) {
	if jy < minJalaaliYear || jy > maxJalaaliYear {
		panic("shamsi: Jalaali year out of range")
	}
	gy = jy + 621
	leapJ := -14
	jp := jalBreaks[0]
	jump := 0
	for i := 1; i < len(jalBreaks); i++ {
		jm := jalBreaks[i]
		jump = jm - jp
		if jy < jm {
			break
		}
		leapJ += div(jump, 33)*8 + div(md(jump, 33), 4)
		jp = jm
	}
	n := jy - jp
	leapJ += div(n, 33)*8 + div(md(n, 33)+3, 4)
	if md(jump, 33) == 4 && jump-n == 4 {
		leapJ++
	}
	leapG := div(gy, 4) - div((div(gy, 100)+1)*3, 4) - 150
	march = 20 + leapJ - leapG

	// Years since the last leap year (0..4); 0 means jy itself is leap.
	adjusted := n
	if jump-n < 6 {
		adjusted = n - jump + div(jump+4, 33)*33
	}
	leap = md(md(adjusted+1, 33)-1, 4)
	if leap == -1 {
		leap = 4
	}
	return gy, march, leap
}

// g2d converts a Gregorian date to a Julian Day Number.
func g2d(gy, gm, gd int) int {
	d := div((gy+div(gm-8, 6)+100100)*1461, 4) + div(153*md(gm+9, 12)+2, 5) + gd - 34840408
	d = d - div(div(gy+100100+div(gm-8, 6), 100)*3, 4) + 752
	return d
}

// d2g converts a Julian Day Number to a Gregorian date.
func d2g(jdn int) (gy, gm, gd int) {
	j := 4*jdn + 139361631
	j += div(div(4*jdn+183187720, 146097)*3, 4)*4 - 3908
	i := div(md(j, 1461), 4)*5 + 308
	gd = div(md(i, 153), 5) + 1
	gm = md(div(i, 153), 12) + 1
	gy = div(j, 1461) - 100100 + div(8-gm, 6)
	return gy, gm, gd
}

// weekdayFromJDN returns the weekday of a Julian Day Number in the Shamsi
// convention (0=Shanbe/Saturday, 6=Jome/Friday).
func weekdayFromJDN(jdn int) int {
	return md(jdn+2, 7)
}

// FromGregorian converts a Gregorian date to a Shamsi (Jalaali) date.
// Panics for dates whose Jalaali year falls outside [-61, 3177].
func FromGregorian(gy, gm, gd int) Date {
	return d2j(g2d(gy, gm, gd))
}

// d2j converts a Julian Day Number to a Shamsi date.
func d2j(jdn int) Date {
	gyy, _, _ := d2g(jdn)
	jy := gyy - 621
	if jy > maxJalaaliYear {
		jy = maxJalaaliYear
	}
	gy, march, leap := jalCal(jy)
	k := jdn - g2d(gy, 3, march)
	wDay := weekdayFromJDN(jdn)
	if k >= 0 {
		if k <= 185 {
			return Date{Year: jy, Month: 1 + div(k, 31), Day: md(k, 31) + 1, WDay: wDay}
		}
		k -= 186
	} else {
		jy--
		k += 179
		if leap == 1 {
			k++
		}
	}
	return Date{Year: jy, Month: 7 + div(k, 30), Day: md(k, 30) + 1, WDay: wDay}
}

// Format returns the month name, or "" if the month is out of range.
func (d Date) Format() string {
	if d.Month < 1 || d.Month > 12 {
		return ""
	}
	return Months[d.Month-1]
}

// FormatFull returns the full date string like "22 Khordad 1405".
func (d Date) FormatFull() string {
	return fmt.Sprintf("%d %s %d", d.Day, d.Format(), d.Year)
}

// FormatCompact returns a compact date like "22 Khord 1405".
func (d Date) FormatCompact() string {
	return fmt.Sprintf("%d %s %d", d.Day, ShortMonths[d.Month-1], d.Year)
}

// WeekDayName returns the day name in the Shamsi week.
func (d Date) WeekDayName() string {
	if d.WDay < 0 || d.WDay > 6 {
		return ""
	}
	return WeekDays[d.WDay]
}

// ShortWeekDayName returns the short day name.
func (d Date) ShortWeekDayName() string {
	if d.WDay < 0 || d.WDay > 6 {
		return ""
	}
	return ShortWeekDays[d.WDay]
}

// DaysInMonth returns the number of days in a Shamsi month for the given year.
func DaysInMonth(jy, jm int) int {
	if jm < 1 || jm > 12 {
		return 0
	}
	if jm <= 6 {
		return 31
	}
	if jm <= 11 {
		return 30
	}
	if IsLeapYear(jy) {
		return 30
	}
	return 29
}

// IsLeapYear returns true if the given Shamsi (Jalaali) year is a leap year.
func IsLeapYear(jy int) bool {
	_, _, leap := jalCal(jy)
	return leap == 0
}
