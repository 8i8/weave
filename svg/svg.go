package svg

import (
	"io"
)

// File returns the basis of an svg file.
func File(w io.Writer) {
}

// ViewBox holds the data for the svg viewbox setting.
type viewBox struct {
	minX   int
	minY   int
	width  int
	height int
}

// Image contains the data to construct an svg.
type Image struct {
	w io.Writer
	viewBox
}

func (v viewBox) viewBoxToBytes() []byte {
	b := make([]byte, 20)
	buf := make([]byte, 83) // 4 * 20 + 3
	n := fmtInt(b[:], v.minX)
	buf = append(buf, b[n:]...)
	buf = append(buf, ' ')
	n = fmtInt(b[:], v.minY)
	buf = append(buf, b[n:]...)
	buf = append(buf, ' ')
	n = fmtInt(b[:], v.width)
	buf = append(buf, b[n:]...)
	buf = append(buf, ' ')
	n = fmtInt(b[:], v.height)
	buf = append(buf, b[n:]...)
	return buf
}

// ViewBox starts an svg image, the viewbox setting is an absolute requirement.
func (s Image) ViewBox(w io.Writer, sX, sY, width, height int) Image {
	s.w = w
	s.viewBox = viewBox{minX: sX, minY: sY, width: width, height: height}
	s.w.Write(headOpen)
	s.w.Write(s.viewBoxToBytes())
	s.w.Write(headClose)
	return s
}

// New returns an Svg struct.
func New(w io.Writer) Image {
	s := Image{w: w}
	return s
}

// End writes the svg end tag to the currnt image.
func (s Image) End() {
	s.w.Write(svgClose)
}
