package svg

import "math"

func ipow(base, exp int) int {
	var res = 1
	for exp > 0 {
		if exp&1 > 0 {
			res *= base
		}
		exp >>= 1
		base *= base
	}
	return res
}

func fmtInt(buf []byte, v int) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	}
	for v > 0 {
		w--
		buf[w] = byte(v%10) + '0'
		v /= 10
	}
	return w
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (v viewBox) longest() int {
	n := 0
	if abs(v.minX) > n {
		n = abs(v.minX)
	}
	if abs(v.minY) > n {
		n = abs(v.minY)
	}
	if abs(v.width) > n {
		n = abs(v.width)
	}
	if abs(v.height) > n {
		n = abs(v.height)
	}
	x := math.Floor(math.Log10(float64(n))) + 1
	return int(x)
}

func (v viewBox) total() int {
	x := math.Floor(math.Log10(float64(v.minX))) + 1
	x += math.Floor(math.Log10(float64(v.minY))) + 1
	x += math.Floor(math.Log10(float64(v.width))) + 1
	x += math.Floor(math.Log10(float64(v.height))) + 1
	x += 3 // a space between each of the four values.
	return int(x)
}
