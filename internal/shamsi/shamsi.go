// Package shamsi provides conversion between Gregorian and Shamsi (Solar Hijri) calendars.
// Uses the Jalaali calendar algorithm.
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

// Days of week in Roman letters.
var WeekDays = [7]string{
	"Shanbe", "Yekshanbe", "Doshanbe",
	"Sehshanbe", "Chaharshanbe", "Panjshanbe", "Jome",
}

// Short day names.
var ShortWeekDays = [7]string{
	"Shan", "Yek", "Dos",
	"Seh", "Cha", "Pan", "Jom",
}

type Date struct {
	Year  int
	Month int // 1-12
	Day   int // 1-31
	WDay  int // 0=Shanbe, 6=Jome
}

// isLeapGregorian returns true if the given Gregorian year is a leap year.
func isLeapGregorian(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// daysInGregorianMonth returns the number of days in a Gregorian month.
func daysInGregorianMonth(month, year int) int {
	switch month {
	case 2:
		if isLeapGregorian(year) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

// isJalaaliLeapYear returns true if the given Jalaali year is a leap year.
// The leap year pattern repeats every 28 years.
func isJalaaliLeapYear(jy int) bool {
	yearsInCycle := jy % 28
	if yearsInCycle < 0 {
		yearsInCycle += 28
	}
	
	// Leap years in a 28-year cycle (1-indexed positions)
	leapYears := []int{1, 5, 9, 13, 17, 22, 26}
	for _, ly := range leapYears {
		if yearsInCycle == ly {
			return true
		}
	}
	return false
}

// daysInJalaaliMonth returns the number of days in a Jalaali month.
func daysInJalaaliMonth(jm int) int {
	if jm <= 6 {
		return 31
	}
	if jm <= 11 {
		return 30
	}
	return 29 // Esfand - will be 30 in leap years
}

// FromGregorian converts a Gregorian date to Jalaali (Shamsi).
// Based on the algorithm from the Jalaali calendar.
func FromGregorian(gy, gm, gd int) Date {
	// Calculate the day of the year
	dayOfYear := gd
	for m := 1; m < gm; m++ {
		dayOfYear += daysInGregorianMonth(m, gy)
	}
	
	// The Jalaali year starts around March 21
	// March 21 is day 80 of a non-leap year, 81 of a leap year
	march21 := 80
	if isLeapGregorian(gy) {
		march21 = 81
	}
	
	// Days since the start of the Jalaali year
	var jDay int
	if dayOfYear >= march21 {
		// Date is after March 21, so it's in the current Jalaali year
		jDay = dayOfYear - march21
	} else {
		// Date is before March 21, so it's in the previous Jalaali year
		// Calculate the total days in the previous Jalaali year
		prevJy := gy - 622 - 1
		_ = isJalaaliLeapYear(prevJy) // Check if previous year was leap
		// Days from March 21 of previous year to end of year
		prevMarch21 := 80
		if isLeapGregorian(gy-1) {
			prevMarch21 = 81
		}
		daysFromPrevMarch := 365 - prevMarch21 + 1 // Days from March 21 to Dec 31
		if isLeapGregorian(gy-1) {
			daysFromPrevMarch = 366 - prevMarch21 + 1
		}
		jDay = dayOfYear + daysFromPrevMarch
	}
	
	// Find the Jalaali year
	var jy int
	if dayOfYear >= march21 {
		jy = gy - 621
	} else {
		jy = gy - 622
	}
	
	// Find the month and day
	jm := 1
	remaining := jDay
	for jm <= 12 {
		var dim int
		if jm < 12 {
			dim = daysInJalaaliMonth(jm)
		} else {
			if isJalaaliLeapYear(jy) {
				dim = 30
			} else {
				dim = 29
			}
		}
		if remaining < dim {
			break
		}
		remaining -= dim
		jm++
	}
	
	jd := remaining + 1
	
	// Calculate day of week
	totalDays := 0
	for y := 1; y < gy; y++ {
		if isLeapGregorian(y) {
			totalDays += 366
		} else {
			totalDays += 365
		}
	}
	totalDays += dayOfYear - 1
	wDay := (totalDays + 1) % 7
	
	return Date{
		Year:  jy,
		Month: jm,
		Day:   jd,
		WDay:  wDay,
	}
}

// Format returns the month name.
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

// WeekDayName returns the day name in Jalaali week.
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

// FormatGregorianWeekDay returns the English day name for a Gregorian date.
func FormatGregorianWeekDay(year, month, day int) string {
	totalDays := 0
	for y := 1; y < year; y++ {
		if isLeapGregorian(y) {
			totalDays += 366
		} else {
			totalDays += 365
		}
	}
	for m := 1; m < month; m++ {
		totalDays += daysInGregorianMonth(m, year)
	}
	totalDays += day - 1
	wDay := (totalDays + 1) % 7
	days := [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	return days[wDay]
}

// DaysInMonth returns the number of days in a Jalaali month for the given year.
func DaysInMonth(year, month int) int {
	if month < 1 || month > 12 {
		return 0
	}
	if month <= 6 {
		return 31
	}
	if month <= 11 {
		return 30
	}
	if isJalaaliLeapYear(year) {
		return 30
	}
	return 29
}

// IsLeapYear returns true if the given Jalaali year is a leap year.
func IsLeapYear(year int) bool {
	return isJalaaliLeapYear(year)
}
