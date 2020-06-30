package flow

import (
	"fmt"
	"time"
)

// Events is an array of event structs.
type Events []Event

// Event is a struct containing data for a particular time event.
type Event struct {
	time.Time
	Valid bool
	Data  interface{}
}

func (e Events) advanceTo(t time.Time) int {
	for i := range e {
		if e[i].After(t) || e[i].Equal(t) {
			return i
		}
	}
	return 0
}

func (e Events) debug() {
	for i := range e {
		fmt.Println("event:", e[i].Time)
	}
}

// Sort functions.

// Len returns the length of the Events list.
func (e Events) Len() int {
	return len(e)
}

// Less retuns a boolean response to the question is e[i] less than e[j].
func (e Events) Less(i, j int) bool {
	return e[i].Before(e[j].Time)
}

// Swap inverses the positions of e[i] and e[j].
func (e Events) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
