package weave

import (
	"fmt"
	"time"
)

// Stiches is an array of event structs.
type Stiches []Stitch

// Stitch is a struct containing data for a particular time event.
type Stitch struct {
	Time  time.Time
	Valid bool
	state int
	Data  interface{}
}

func (e Stiches) advanceTo(t time.Time) int {
	for i := range e {
		if e[i].Time.After(t) || e[i].Time.Equal(t) {
			return i
		}
	}
	return 0
}

func (e Stiches) debug() {
	for i := range e {
		fmt.Println("event:", e[i].Time)
	}
}

// Sort functions.

// Len returns the length of the Events list.
func (e Stiches) Len() int {
	return len(e)
}

// Less retuns a boolean response to the question is e[i] less than e[j].
func (e Stiches) Less(i, j int) bool {
	return e[i].Time.Before(e[j].Time)
}

// Swap inverses the positions of e[i] and e[j].
func (e Stiches) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
