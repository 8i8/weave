package weave

import (
	"fmt"
)

// Stitches is an array of event structs.
type Stitches []Stitch

// Stitch is a struct containing data for a particular time event.
type Stitch struct {
	//Valid  bool
	Data   interface{}
	before Comp
	equal  Comp
	after  Comp
}

// LoadFuncs loads the comaparison functions into a stitch.
func (s Stitch) LoadFuncs(c CompFuncs) Stitch {
	s.before, s.equal, s.after = c[0], c[1], c[2]
	return s
}

// Before returns true if s is less than n.
func (s Stitch) Before(n Stitch) bool {
	return s.before(s, n)
}

// Equal returns true if s is less than n.
func (s Stitch) Equal(n Stitch) bool {
	return s.equal(s, n)
}

// After returns true if s is less than n.
func (s Stitch) After(n Stitch) bool {
	return s.after(s, n)
}

func (s Stitches) advanceTo(n Stitch) int {
	for i := range s {
		if s[i].After(n) || s[i].Equal(n) {
			return i
		}
	}
	return 0
}

func (s Stitches) debug() {
	for i := range s {
		fmt.Println("event:", s[i].Data)
	}
}

// Sort functions.

// Len returns the length of the Events list.
func (s Stitches) Len() int {
	return len(s)
}

// Less retuns a boolean response to the question is e[i] less than e[j].
func (s Stitches) Less(i, j int) bool {
	return s[i].Before(s[j])
}

// Swap inverses the positions of e[i] and e[j].
func (s Stitches) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
