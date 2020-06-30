package weave

import (
	"fmt"
	"time"
)

const (
	biTload = 1 << iota
	biTevent
)

// WeaverFunc are the functions run by the weaver to generate its output,
// facilitaing the use of closures when calling the functions.
type WeaverFunc func(Weaver, Stitch) Weaver

// Weaver interleave multiple streams of dates.
type Weaver struct {
	Start      time.Time
	End        time.Time
	Data       interface{}
	current    Stitch
	next       Stitch
	shuttle    []shuttle
	Output     Threads
	PreWeave   WeaverFunc
	PreStitch  WeaverFunc
	Stitch     WeaverFunc
	PostStitch WeaverFunc
	PostWeave  WeaverFunc
	debug      int
}

type shuttle struct {
	ch      chan Stitch
	prev    thread
	current thread
	next    thread
}

// State contains the state of a thread, accessible and maintained by the user
// to maintain and pass state between stitches.
type State int

// thread contains a stitch and that stitches state.
type thread struct {
	State int
	Stitch
}

// Threads enables the sorting of threads.
type Threads []thread

// Sort functions.

// Len returns the length of the Events list.
func (t Threads) Len() int {
	return len(t)
}

// Less retuns a boolean response to the question is e[i] less than e[j].
func (t Threads) Less(i, j int) bool {
	return t[i].Time.Before(t[j].Time)
}

// Swap inverses the positions of e[i] and e[j].
func (t Threads) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Date builds and runs a weaver, the first chanel that is provided is used as
// a grid or ruler for structuring of all of the folling channel.
func (w Weaver) Date(ev Stitches, chans ...chan Stitch) {
	//w.debug= biTload | biTevent,
	w.shuttle = make([]shuttle, len(chans))
	for i := range chans {
		w.shuttle[i].ch = chans[i]
	}
	w.weavedate(ev)
}

// Load loads a user data into the weave for access inside of WeaveFunc's.
func (w Weaver) Load(d interface{}) Weaver {
	w.Data = d
	return w
}

func (w Weaver) loadReceivers() Weaver {
	for i := range w.shuttle {
		// Load yantra.
		w.shuttle[i].next.Stitch = <-w.shuttle[i].ch
		// Skip over closed channels.
		if !w.shuttle[i].next.Valid {
			continue
		}
		// Remove all yantra in between required dates.
		for w.shuttle[i].next.Time.Before(w.Start) && w.shuttle[i].next.Valid {
			w.shuttle[i].prev = w.shuttle[i].current
			w.shuttle[i].current = w.shuttle[i].next
			w.shuttle[i].next.Stitch = <-w.shuttle[i].ch
		}
	}
	for i := range w.shuttle {
		w.shuttle[i].prev = w.shuttle[i].current
		w.shuttle[i].current = w.shuttle[i].next
		w.shuttle[i].next.Stitch = <-w.shuttle[i].ch
	}

	if w.debug&biTload > 0 {
		for _, r := range w.shuttle {
			fmt.Println("r.prev:", r.prev)
			fmt.Println("r.current:", r.current)
			fmt.Println("r.next:", r.next)
		}
	}

	return w
}

func (w Weaver) updateReceivers() Weaver {
	for i := range w.shuttle {
		if !w.shuttle[i].current.Valid {
			continue
		}
		for w.shuttle[i].next.Time.Before(w.next.Time) || w.shuttle[i].next.Time.Equal(w.next.Time) {
			w.shuttle[i].prev = w.shuttle[i].current
			w.shuttle[i].current = w.shuttle[i].next
			w.shuttle[i].next.Stitch = <-w.shuttle[i].ch
			if !w.shuttle[i].next.Valid {
				break
			}
		}
	}
	return w
}

func (w Weaver) loadCalendar() Weaver {
	for {
		w.current = w.next // Set up lookahead.
		w.next = <-w.shuttle[0].ch
		if w.next.Time.After(w.Start) {
			break
		}
	}
	w.current = w.next
	w.next = <-w.shuttle[0].ch
	return w
}

func (w Weaver) weavedate(ev Stitches) error {
	// Load all required data from chans.
	j := ev.advanceTo(w.Start)
	if w.debug&biTevent > 0 {
		ev.debug()
	}
	w = w.loadReceivers()
	w = w.loadCalendar()
	if w.next.Time.Equal(time.Time{}) {
		return fmt.Errorf("weave: date: inValid")
	}
	// Extract calendar.
	calendar := w.shuttle[0].ch
	w.shuttle = w.shuttle[1:]
	// Output function.
	if w.PreWeave != nil {
		w = w.PreWeave(w, w.current)
	}
	for w.next = range calendar {
		w = w.updateReceivers()
		w.Output = w.Output[:0]
		for _, r := range w.shuttle {
			if r.current.Time.Before(w.next.Time) {
				w.Output = append(w.Output, r.current)
			}
		}
		// Output function.
		if w.PreStitch != nil {
			w = w.PreStitch(w, w.current)
		}
		// Events.
		for ; j < len(ev) && ev[j].Time.Before(w.next.Time); j++ {
			// Omit Events that are too recent.
			if ev[j].Time.After(w.End) {
				break
			}
			// Output function.
			if w.Stitch != nil {
				w = w.Stitch(w, ev[j])
			}
		}
		// Output function.
		if w.PostStitch != nil {
			w = w.PostStitch(w, w.current)
		}
		// Check out when required.
		if w.next.Time.After(w.End) {
			break
		}
		// Prepare for the next row.
		w.current = w.next
	}
	// Output function.
	if w.PostWeave != nil {
		w = w.PostWeave(w, w.current)
	}
	return nil
}

// func weavedasa(w Weaver, ev Events) error {
// 	// Find first event index.
// 	j := ev.advanceTo(w.Start)
// 	if w.debug&biTevent > 0 {
// 		ev.debug()
// 	}

// 	w = w.loadReceivers()

// 	if w.PreFunc != nil {
// 		w.PreFunc()
// 	}

// 	for w.next.Valid {
// 	}
// 	return nil
// }
