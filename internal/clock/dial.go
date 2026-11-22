package clock

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type dialPalette struct {
	enamel, bezel, metal, lume color.RGBA
}

func paletteFor(dim, border, date lipgloss.Color) dialPalette {
	return dialPalette{
		enamel: mix(rgb(0x080C12), themeRGB(dim), .58),
		bezel:  mix(rgb(0x3E454C), themeRGB(border), .42),
		metal:  mix(rgb(0xE1E4E7), themeRGB(date), .20),
		lume:   mix(rgb(0xF4EEE1), themeRGB(date), .24),
	}
}

func rgb(hex uint32) color.RGBA {
	return color.RGBA{R: byte(hex >> 16), G: byte(hex >> 8), B: byte(hex), A: 255}
}

func themeRGB(c lipgloss.Color) color.RGBA {
	n, _ := strconv.ParseUint(strings.TrimPrefix(string(c), "#"), 16, 24)
	return rgb(uint32(n))
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	channel := func(a, b byte) byte { return byte(float64(a) + (float64(b)-float64(a))*t + .5) }
	return color.RGBA{channel(a.R, b.R), channel(a.G, b.G), channel(a.B, b.B), 255}
}

type point struct{ x, y float64 }

func (p point) add(q point) point     { return point{p.x + q.x, p.y + q.y} }
func (p point) sub(q point) point     { return point{p.x - q.x, p.y - q.y} }
func (p point) scale(s float64) point { return point{p.x * s, p.y * s} }
func (p point) dot(q point) float64   { return p.x*q.x + p.y*q.y }

func direction(turn float64) point {
	sin, cos := math.Sincos(turn * 2 * math.Pi)
	return point{sin, -cos}
}

func handTurns(now time.Time) (hour, minute, second float64) {
	second = (float64(now.Second()) + float64(now.Nanosecond())/1e9) / 60
	minute = (float64(now.Minute()) + second) / 60
	hour = (float64(now.Hour()%12) + minute) / 12
	return
}

// dialCanvas uses a dial radius of one and antialiases against the pixel grid.
// The transparent margin leaves room for the bezel shadow.
type dialCanvas struct {
	*image.RGBA
	center, radius float64
}

func newDialCanvas(size int) dialCanvas {
	return dialCanvas{image.NewRGBA(image.Rect(0, 0, size, size)), float64(size) / 2, float64(size) * .461}
}

func clamp(v float64) float64 { return min(max(v, 0), 1) }

func (d dialCanvas) paint(bounds image.Rectangle, c color.RGBA, distance func(point) float64) {
	bounds = bounds.Intersect(d.Bounds())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := point{(float64(x) + .5 - d.center) / d.radius, (float64(y) + .5 - d.center) / d.radius}
			coverage := clamp(.5 - distance(p)*d.radius)
			if coverage > 0 {
				d.blend(x, y, c, coverage)
			}
		}
	}
}

func (d dialCanvas) blend(x, y int, c color.RGBA, opacity float64) {
	i := d.PixOffset(x, y)
	a := byte(float64(c.A)*opacity + .5)
	inverse := 255 - uint16(a)
	for ch, value := range [...]byte{c.R, c.G, c.B, 255} {
		d.Pix[i+ch] = byte((uint16(value)*uint16(a) + uint16(d.Pix[i+ch])*inverse + 127) / 255)
	}
}

func (d dialCanvas) bounds(a, b point, pad float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(d.center+(min(a.x, b.x)-pad)*d.radius))-1,
		int(math.Floor(d.center+(min(a.y, b.y)-pad)*d.radius))-1,
		int(math.Ceil(d.center+(max(a.x, b.x)+pad)*d.radius))+1,
		int(math.Ceil(d.center+(max(a.y, b.y)+pad)*d.radius))+1,
	)
}

func (d dialCanvas) stroke(start, end point, width float64, c color.RGBA) {
	delta := end.sub(start)
	lengthSquared := delta.dot(delta)
	d.paint(d.bounds(start, end, width), c, func(p point) float64 {
		p = p.sub(start)
		if lengthSquared > 0 {
			p = p.sub(delta.scale(clamp(p.dot(delta) / lengthSquared)))
		}
		return math.Sqrt(p.dot(p)) - width
	})
}

func (d dialCanvas) polygon(c color.RGBA, vertices ...point) {
	lo, hi := vertices[0], vertices[0]
	area := 0.0
	for i, p := range vertices {
		lo = point{min(lo.x, p.x), min(lo.y, p.y)}
		hi = point{max(hi.x, p.x), max(hi.y, p.y)}
		q := vertices[(i+1)%len(vertices)]
		area += p.x*q.y - p.y*q.x
	}
	normals := make([]point, len(vertices))
	for i, p := range vertices {
		edge := vertices[(i+1)%len(vertices)].sub(p)
		normals[i] = point{edge.y, -edge.x}.scale(math.Copysign(1, area) / math.Sqrt(edge.dot(edge)))
	}
	d.paint(d.bounds(lo, hi, 0), c, func(p point) float64 {
		distance := math.Inf(-1)
		for i, vertex := range vertices {
			distance = max(distance, p.sub(vertex).dot(normals[i]))
		}
		return distance
	})
}

func drawDial(size int, palette dialPalette, cells bool) *image.RGBA {
	d := newDialCanvas(size)
	black, white := rgb(0x06080B), rgb(0xECF0F4)
	faceLight := mix(palette.enamel, palette.bezel, .29)
	for y := range size {
		for x := range size {
			u, v := (float64(x)+.5-d.center)/d.radius, (float64(y)+.5-d.center)/d.radius
			r := math.Hypot(u, v)
			if !cells && r > .98 && r < 1.08 {
				shadow := math.Hypot(u-.013, v-.031)
				d.blend(x, y, black, .34*clamp((1.064-shadow)/.065))
			}
			coverage := clamp(.5 + (1-r)*d.radius)
			if coverage == 0 {
				continue
			}
			light := clamp(.54 - .31*u - .38*v)
			fill := mix(palette.enamel, faceLight, clamp(1-math.Hypot(u+.24, v+.29)))
			// The inner rim catches a softer reflection than the raised bezel.
			edge := func(at float64) float64 { return clamp(.5 + (r-at)*d.radius) }
			satin := .013 * math.Sin(r*420)
			bezel := mix(palette.bezel, white, clamp(.025+light*.30+satin))
			fill = mix(fill, mix(palette.enamel, palette.bezel, .52), edge(.912))
			fill = mix(fill, mix(black, palette.bezel, .18+light*.42), edge(.922))
			fill = mix(fill, bezel, edge(.941))
			fill = mix(fill, mix(bezel, white, .12*light), edge(.975))
			fill = mix(fill, mix(black, palette.bezel, .46+light*.43), edge(.987))
			d.blend(x, y, fill, coverage)
		}
	}

	for tick := range 60 {
		axis := direction(float64(tick) / 60)
		if tick%5 != 0 {
			d.stroke(axis.scale(.853), axis.scale(.879), .005,
				mix(palette.enamel, palette.metal, .48))
			continue
		}
		start, width := .766, .012
		if tick%15 == 0 {
			start, width = .739, .016
		}
		if cells {
			width *= 1.75
			start -= .04
		}
		if tick == 0 {
			for _, offset := range []float64{-width * 1.75, width * 1.75} {
				d.stroke(point{offset, -start}, point{offset, -.875}, width, palette.lume)
			}
			continue
		}
		d.stroke(axis.scale(start), axis.scale(.875), width, palette.metal)
	}
	return d.RGBA
}

func drawHands(size int, now time.Time, palette dialPalette, accent color.RGBA, cells bool) *image.RGBA {
	d := newDialCanvas(size)
	hour, minute, second := handTurns(now)
	widthScale := 1.0
	if cells {
		widthScale = 1.7
	}
	d.hand(direction(hour), .508, .047*widthScale, palette.metal)
	d.hand(direction(minute), .725, .033*widthScale, palette.lume)

	axis := direction(second)
	d.stroke(axis.scale(-.24), axis.scale(.856), .007*widthScale, accent)
	weight := axis.scale(-.225)
	d.stroke(weight, weight, .031*widthScale, accent)
	d.stroke(weight, weight, .014*widthScale, palette.enamel)

	d.stroke(point{}, point{}, .068*widthScale, rgb(0x090D13))
	d.stroke(point{}, point{}, .048*widthScale, palette.metal)
	d.stroke(point{}, point{}, .032*widthScale, palette.bezel)
	d.stroke(point{-.009, -.011}, point{-.009, -.011}, .013*widthScale, palette.lume)
	return d.RGBA
}

func (d dialCanvas) hand(axis point, length, width float64, finish color.RGBA) {
	normal := point{-axis.y, axis.x}
	at := func(along, cross float64) point { return axis.scale(along).add(normal.scale(cross)) }
	heel, left, tip, right := at(-.061, 0), at(.12, -width), at(length, 0), at(.12, width)
	shadow := point{.013, .017}
	d.polygon(color.RGBA{A: 168}, heel.add(shadow), left.add(shadow), tip.add(shadow), right.add(shadow))
	d.polygon(mix(finish, rgb(0x29323D), .31), heel, left, tip, right)
	d.polygon(finish, heel, left, tip)
}
