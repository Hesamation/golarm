package ui

import "strings"

// FaDigits converts ASCII digits in s to Persian (Eastern Arabic) digits.
func FaDigits(s string) string {
	return strings.NewReplacer(
		"0", "۰", "1", "۱", "2", "۲", "3", "۳", "4", "۴",
		"5", "۵", "6", "۶", "7", "۷", "8", "۸", "9", "۹",
	).Replace(s)
}
