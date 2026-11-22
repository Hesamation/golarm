package clock

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const readoutWidth = 32

type cellSize struct{ width, height int }

type dialLayout struct {
	cellSize
	diameter, columns, rows int
	left, top               int
	stacked                 bool
}

func layoutDial(width, height, readoutRows int, cell cellSize, pixels bool) dialLayout {
	const gutter = 4
	availableCols, availableRows := width-4, height-4
	minimum, maximum := 180, 640
	if !pixels {
		minimum = max(24*cell.width, 10*cell.height)
		maximum = min(52*cell.width, 26*cell.height)
	}

	diameter := min((availableCols-readoutWidth-gutter)*cell.width, availableRows*cell.height, maximum)
	stacked := diameter < minimum || availableRows < readoutRows
	if stacked {
		diameter = min(availableCols*cell.width, (availableRows-readoutRows-1)*cell.height, maximum)
		if availableCols < readoutWidth {
			return dialLayout{}
		}
	}
	if diameter < minimum {
		return dialLayout{}
	}
	columns := (diameter + cell.width - 1) / cell.width
	rows := (diameter + cell.height - 1) / cell.height
	blockWidth, bodyHeight := columns+gutter+readoutWidth, max(rows, readoutRows)
	if stacked {
		blockWidth, bodyHeight = max(columns, readoutWidth), rows+1+readoutRows
	}
	left := (width - blockWidth) / 2
	if stacked {
		left += (blockWidth - columns) / 2
	}
	return dialLayout{cell, diameter, columns, rows, left, (height - bodyHeight - 2) / 2, stacked}
}

func (l dialLayout) pixelOrigin() (x, y int) {
	return l.left*l.width + (l.columns*l.width-l.diameter)/2,
		l.top*l.height + (l.rows*l.height-l.diameter)/2
}

func (l dialLayout) lines(width, height int, face, readout []string, footer string) []string {
	rows := make([]string, height)
	for i, line := range face {
		rows[l.top+i] = strings.Repeat(" ", l.left) + line
	}
	var footerRow int
	if l.stacked {
		top := l.top + l.rows + 1
		for i, line := range readout {
			rows[top+i] = centerLine(line, width)
		}
		footerRow = top + len(readout) + 1
	} else {
		left := l.left + l.columns + 4
		top := l.top + (max(l.rows, len(readout))-len(readout))/2
		for i, line := range readout {
			rows[top+i] += strings.Repeat(" ", left-lipgloss.Width(rows[top+i])) + line
		}
		footerRow = l.top + max(l.rows, len(readout)) + 1
	}
	rows[footerRow] = centerLine(footer, width)
	for i, row := range rows {
		rows[i] += strings.Repeat(" ", max(width-lipgloss.Width(row), 0))
	}
	return rows
}

func centerLine(line string, width int) string {
	return strings.Repeat(" ", max((width-lipgloss.Width(line))/2, 0)) + line
}

func (l dialLayout) colorCells(dial, hands *image.RGBA, profile termenv.Profile) []string {
	frame := image.NewRGBA(dial.Bounds())
	draw.Draw(frame, frame.Bounds(), dial, image.Point{}, draw.Src)
	draw.Draw(frame, frame.Bounds(), hands, image.Point{}, draw.Over)
	step := float64(frame.Bounds().Dx()) / float64(l.diameter)
	offsetX := float64((l.columns*l.width - l.diameter) / 2)
	offsetY := float64((l.rows*l.height - l.diameter) / 2)
	lines := make([]string, l.rows)
	for y := range l.rows {
		var line strings.Builder
		fg, bg := "39", "49"
		for x := range l.columns {
			left, top := (float64(x*l.width)-offsetX)*step, (float64(y*l.height)-offsetY)*step
			w, h := float64(l.width)*step, float64(l.height)*step/2
			above := sampleHalfCell(frame, left, top, w, h)
			below := sampleHalfCell(frame, left, top+h, w, h)
			glyph := '▀'
			nextFG, nextBG := "39", "49"
			switch {
			case above.A > 0:
				nextFG = profile.FromColor(above).Sequence(false)
				if below.A > 0 {
					nextBG = profile.FromColor(below).Sequence(true)
				}
			case below.A > 0:
				glyph = '▄'
				nextFG = profile.FromColor(below).Sequence(false)
			default:
				glyph = ' '
			}
			if nextFG != fg || nextBG != bg {
				line.WriteString("\x1b[" + nextFG + ";" + nextBG + "m")
				fg, bg = nextFG, nextBG
			}
			line.WriteRune(glyph)
		}
		line.WriteString("\x1b[0m")
		lines[y] = line.String()
	}
	return lines
}

func sampleHalfCell(frame *image.RGBA, x, y, width, height float64) color.RGBA {
	var r, g, b, a int
	const samples = 4
	for row := range samples {
		for col := range samples {
			c := frame.RGBAAt(
				int(x+(float64(col)+.5)*width/samples),
				int(y+(float64(row)+.5)*height/samples),
			)
			r, g, b, a = r+int(c.R), g+int(c.G), b+int(c.B), a+int(c.A)
		}
	}
	if a < 128*samples*samples {
		return color.RGBA{}
	}
	return color.RGBA{R: byte(r * 255 / a), G: byte(g * 255 / a), B: byte(b * 255 / a), A: 255}
}
