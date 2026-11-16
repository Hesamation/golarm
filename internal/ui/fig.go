package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
)

const figletFont = "standard"

// ClockFigW is the render width of a full "88:88:88" clock face. All
// time-only figlet lines are padded to it so the card never jitters as
// digits change.
var ClockFigW = figletWidth("88:88:88")

// ClockAMPMFigW is the same for 12-hour mode ("88:88:88 PM").
var ClockAMPMFigW = figletWidth("88:88:88 PM")

// RenderFiglet renders text with the standard figlet font as lines.
func RenderFiglet(text string) []string {
	fig := figure.NewFigure(text, figletFont, false)
	s := strings.TrimRight(fig.String(), "\n")
	if s == "" {
		return []string{text}
	}
	return strings.Split(s, "\n")
}

func figletWidth(text string) int {
	maxW := 0
	for _, line := range RenderFiglet(text) {
		if w := lipgloss.Width(line); w > maxW {
			maxW = w
		}
	}
	return maxW
}

// Fig renders text as figlet lines styled with style and padded to width.
// Lines that are already wider (e.g. hour >= 100) are left unpadded.
func Fig(text string, style lipgloss.Style, width int) []string {
	lines := RenderFiglet(text)
	out := make([]string, len(lines))
	for i, line := range lines {
		if pad := width - lipgloss.Width(line); pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		out[i] = style.Render(line)
	}
	return out
}
