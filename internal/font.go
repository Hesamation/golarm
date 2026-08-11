package internal

import (
	"strings"
	"unicode"

	"github.com/common-nighthawk/go-figure"
)

const FigletFont = "standard"

func RenderFiglet(text string) []string {
	fig := figure.NewFigure(text, FigletFont, false)
	s := fig.String()
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return []string{text}
	}
	return strings.Split(s, "\n")
}

func FigletWidth(text string) int {
	lines := RenderFiglet(text)
	maxW := 0
	for _, line := range lines {
		w := stringWidth(line)
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

func stringWidth(s string) int {
	w := 0
	for _, r := range s {
		if unicode.IsPrint(r) || r == '\t' {
			w++
		}
	}
	return w
}
