package clock

import (
	"image"
	"image/color"
)

func drawSectorDial(size int, palette dialPalette, cells, numerals bool) *image.RGBA {
	d := newDialCanvas(size)
	chapter := mix(palette.enamel, palette.bezel, .43)
	rim := mix(palette.bezel, palette.metal, .27)
	d.stroke(point{}, point{}, 1, rim)
	d.stroke(point{}, point{}, .971, palette.bezel)
	d.stroke(point{}, point{}, .947, chapter)
	d.stroke(point{}, point{}, .778, rim)
	d.stroke(point{}, point{}, .760, palette.enamel)

	if !cells {
		// A fine minute track between the two concentric bands.
		for tick := range 60 {
			if tick%5 != 0 {
				axis := direction(float64(tick) / 60)
				d.stroke(axis.scale(.854), axis.scale(.915), .005, rim)
			}
		}
	}
	for hour := range 12 {
		axis := direction(float64(hour) / 12)
		start, weight, ink := .831, .012, palette.metal
		if cells {
			start, weight = .825, .032
		}
		if hour%3 == 0 {
			start, weight, ink = start-.023, weight*1.25, palette.lume
		}
		d.stroke(axis.scale(start), axis.scale(.910), weight, ink)
	}
	if numerals {
		for i, label := range []string{"12", "3", "6", "9"} {
			axis := direction(float64(i) / 4)
			if cells {
				d.cellNumber(label, axis.scale(.53), .064, palette.lume)
			} else {
				d.dialNumber(label, axis.scale(.606), palette.lume)
			}
		}
	}
	return d.RGBA
}

func drawMinimalDial(size int, palette dialPalette, cells bool) *image.RGBA {
	d := newDialCanvas(size)
	rim := mix(palette.bezel, palette.lume, .23)
	inner, weight := .988, .007
	if cells {
		inner, weight = .954, .025
	}
	d.stroke(point{}, point{}, 1, rim)
	d.stroke(point{}, point{}, inner, palette.enamel)

	for hour := range 12 {
		axis := direction(float64(hour) / 12)
		if hour%3 == 0 {
			d.stroke(axis.scale(.816), axis.scale(.908), weight*1.65, palette.lume)
		} else {
			mark := axis.scale(.875)
			d.stroke(mark, mark, weight, palette.metal)
		}
	}
	return d.RGBA
}

func (d dialCanvas) dialNumber(label string, center point, ink color.RGBA) {
	// A small set of drawn numerals avoids relying on the terminal font.
	digits := map[byte][][]point{
		'1': {{{.07, .18}, {.50, 0}, {.50, 1}}, {{.10, 1}, {.90, 1}}},
		'2': {{{.02, .18}, {.26, .02}, {.74, .02}, {.98, .18}, {.98, .36}, {0, .97}, {1, .97}}},
		'3': {{{.04, .05}, {.80, .05}, {.98, .22}, {.80, .43}, {.40, .48}, {.79, .50}, {1, .68}, {.96, .86}, {.73, .98}, {.04, .94}}},
		'6': {{{.89, .07}, {.42, .03}, {.14, .22}, {0, .56}, {.07, .84}, {.35, .99}, {.76, .97}, {1, .77}, {.97, .56}, {.72, .45}, {.32, .46}, {0, .60}}},
		'9': {{{1, .43}, {.68, .56}, {.27, .53}, {.02, .37}, {.03, .17}, {.27, .02}, {.70, .02}, {.95, .18}, {1, .42}, {.89, .76}, {.61, .97}, {.10, .94}}},
	}
	const width, height, gap = .106, .165, .041
	left := center.x - (float64(len(label))*(width+gap)-gap)/2
	for i := range len(label) {
		for _, path := range digits[label[i]] {
			for j := 1; j < len(path); j++ {
				a, b := path[j-1], path[j]
				origin := point{left + float64(i)*(width+gap), center.y - height/2}
				d.stroke(origin.add(point{a.x * width, a.y * height}),
					origin.add(point{b.x * width, b.y * height}), .0095, ink)
			}
		}
	}
}

// At cell resolution, a 3×5 grid gives the numerals stronger silhouettes.
func (d dialCanvas) cellNumber(label string, center point, pitch float64, ink color.RGBA) {
	digits := map[byte][5]uint8{
		'1': {2, 6, 2, 2, 7},
		'2': {6, 1, 2, 4, 7},
		'3': {6, 1, 2, 1, 6},
		'6': {3, 4, 6, 5, 7},
		'9': {7, 5, 3, 1, 6},
	}
	left := center.x - float64(len(label)*4-1)*pitch/2
	top := center.y - 2.5*pitch
	for i := range len(label) {
		for row, mask := range digits[label[i]] {
			for col := range 3 {
				if mask&(4>>col) == 0 {
					continue
				}
				a := point{left + float64(i*4+col)*pitch, top + float64(row)*pitch}
				b := a.add(point{pitch, pitch})
				d.paint(d.bounds(a, b, 0), ink, func(p point) float64 {
					return max(a.x-p.x, p.x-b.x, a.y-p.y, p.y-b.y)
				})
			}
		}
	}
}
