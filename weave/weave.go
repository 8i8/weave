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
	Start     time.Time
	End       time.Time
	Data      interface{}
	current   Stitch
	next      Stitch
	receiver  []receiver
	Output    Stiches
	PreFunc   WeaverFunc
	PreEvent  WeaverFunc
	Event     WeaverFunc
	PostEvent WeaverFunc
	PostFunc  WeaverFunc
	debug     int
}

type receiver struct {
	ch      chan Stitch
	prev    Stitch
	current Stitch
	next    Stitch
}

// Date builds and runs a weaver, the first chanel that is provided is used as
// a grid or ruler for structuring of all of the folling channel.
func (w Weaver) Date(ev Stiches, chans ...chan Stitch) {
	//w.debug= biTload | biTevent,
	w.receiver = make([]receiver, len(chans))
	for i := range chans {
		w.receiver[i].ch = chans[i]
	}
	w.weavedate(ev)
}

// Load loads a user data into the weave for access inside of WeaveFunc's.
func (w Weaver) Load(d interface{}) Weaver {
	w.Data = d
	return w
}

func (w Weaver) loadReceivers() Weaver {
	for i := range w.receiver {
		// Load yantra.
		w.receiver[i].next = <-w.receiver[i].ch
		// Skip over closed channels.
		if !w.receiver[i].next.Valid {
			continue
		}
		// Remove all yantra in between required dates.
		for w.receiver[i].next.Time.Before(w.Start) && w.receiver[i].next.Valid {
			w.receiver[i].prev = w.receiver[i].current
			w.receiver[i].current = w.receiver[i].next
			w.receiver[i].next = <-w.receiver[i].ch
		}
	}
	for i := range w.receiver {
		w.receiver[i].prev = w.receiver[i].current
		w.receiver[i].current = w.receiver[i].next
		w.receiver[i].next = <-w.receiver[i].ch
	}

	if w.debug&biTload > 0 {
		for _, r := range w.receiver {
			fmt.Println("r.prev:", r.prev)
			fmt.Println("r.current:", r.current)
			fmt.Println("r.next:", r.next)
		}
	}

	return w
}

func (w Weaver) updateReceivers() Weaver {
	for i := range w.receiver {
		if !w.receiver[i].current.Valid {
			continue
		}
		for w.receiver[i].next.Time.Before(w.next.Time) || w.receiver[i].next.Time.Equal(w.next.Time) {
			w.receiver[i].prev = w.receiver[i].current
			w.receiver[i].current = w.receiver[i].next
			w.receiver[i].next = <-w.receiver[i].ch
			if !w.receiver[i].next.Valid {
				break
			}
		}
	}
	return w
}

func (w Weaver) loadCalendar() Weaver {
	for {
		w.current = w.next // Set up lookahead.
		w.next = <-w.receiver[0].ch
		if w.next.Time.After(w.Start) {
			break
		}
	}
	w.current = w.next
	w.next = <-w.receiver[0].ch
	return w
}

func (w Weaver) weavedate(ev Stiches) error {
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
	calendar := w.receiver[0].ch
	w.receiver = w.receiver[1:]
	// Output function.
	if w.PreFunc != nil {
		w = w.PreFunc(w, w.current)
	}
	for w.next = range calendar {
		w = w.updateReceivers()
		w.Output = w.Output[:0]
		for _, r := range w.receiver {
			if r.current.Time.Before(w.next.Time) {
				w.Output = append(w.Output, r.current)
			}
		}
		// Output function.
		if w.PreEvent != nil {
			w = w.PreEvent(w, w.current)
		}
		// Events.
		for ; j < len(ev) && ev[j].Time.Before(w.next.Time); j++ {
			// Omit Events that are too recent.
			if ev[j].Time.After(w.End) {
				break
			}
			// Output function.
			if w.Event != nil {
				w = w.Event(w, ev[j])
			}
		}
		// Output function.
		if w.PostEvent != nil {
			w = w.PostEvent(w, w.current)
		}
		// Check out when required.
		if w.next.Time.After(w.End) {
			break
		}
		// Prepare for the next row.
		w.current = w.next
	}
	// Output function.
	if w.PostFunc != nil {
		w = w.PostFunc(w, w.current)
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
