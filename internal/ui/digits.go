package ui

import "strings"

// FaDigits converts ASCII digits to Persian digits, optionally separating runs
// for terminals that shape adjacent Arabic-script digits right-to-left.
func FaDigits(s string, spaced bool) string {
	var text strings.Builder
	previousDigit := false
	for _, r := range s {
		digit := r >= '0' && r <= '9'
		if digit {
			if spaced && previousDigit {
				text.WriteRune('\u2009')
			}
			r = '۰' + (r - '0')
		}
		text.WriteRune(r)
		previousDigit = digit
	}
	return text.String()
}
