package svg

import "unicode/utf8"

type buffer []byte

func (b *buffer) write(p []byte) {
	*b = append(*b, p...)
}

func (b *buffer) writeString(s string) {
	*b = append(*b, s...)
}

func (b *buffer) writeByte(c byte) {
	*b = append(*b, c)
}

func (b *buffer) writeRune(r rune) {
	if r < utf8.RuneSelf {
		*b = append(*b, byte(r))
		return
	}
	buf := *b
	n := len(buf)
	for n+utf8.UTFMax > cap(buf) {
		buf = append(buf, 0)
	}
	w := utf8.EncodeRune(buf[n:n+utf8.UTFMax], r)
	*b = buf[:n+w]
}
