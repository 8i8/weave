package main

import (
	"fmt"
	"math/rand"
	"time"
)

const year = 365
const (
	sTload = 1 << iota
	sTevent
)

func init() {
}

type weaver struct {
	start   time.Time
	end     time.Time
	current event
	next    event
	r       []receiver
	fn      []func(...interface{})
	debug   int
}

type receiver struct {
	ch      chan event
	prev    event
	current event
	next    event
}

func (w weaver) loadReceivers() weaver {
	for i := range w.r {
		// Load yantra.
		w.r[i].next = <-w.r[i].ch
		// Skip over closed channels.
		if !w.r[i].next.valid {
			continue
		}
		// Remove all yantra in between required dates.
		for w.r[i].next.Before(w.start) && w.r[i].next.valid {
			w.r[i].prev = w.r[i].current
			w.r[i].current = w.r[i].next
			w.r[i].next = <-w.r[i].ch
		}
	}
	for i := range w.r {
		w.r[i].prev = w.r[i].current
		w.r[i].current = w.r[i].next
		w.r[i].next = <-w.r[i].ch
	}

	if w.debug&sTload > 0 {
		for _, r := range w.r {
			fmt.Println("r.prev:", r.prev)
			fmt.Println("r.current:", r.current)
			fmt.Println("r.next:", r.next)
		}
	}

	return w
}

func (w weaver) updateReceivers() weaver {
	for i := range w.r {
		if !w.r[i].current.valid {
			continue
		}
		for w.r[i].next.Before(w.next.Time) || w.r[i].next.Equal(w.next.Time) {
			w.r[i].prev = w.r[i].current
			w.r[i].current = w.r[i].next
			w.r[i].next = <-w.r[i].ch
			if !w.r[i].next.valid {
				break
			}
		}
	}
	return w
}

// Weave builds and runs a waver.
func Weave(ev events, start, end time.Time, chans ...chan event) {
	w := weaver{
		start: start,
		end:   end,
		//debug: sTload | sTevent,
	}
	w.r = make([]receiver, len(chans))
	w.fn = make([]func(...interface{}), 4)
	w.fn[0] = fn1
	//	w.fn[1] = fn2
	w.fn[2] = fn3
	//w.fn[3] = fn4
	for i := range chans {
		w.r[i].ch = chans[i]
	}
	weave(w, ev)
}

func (w weaver) loadCalendar() weaver {
	for {
		w.current = w.next // Set up lookahead.
		w.next = <-w.r[0].ch
		if w.next.After(w.start) {
			break
		}
	}
	w.current = w.next
	w.next = <-w.r[0].ch
	return w
}

func weave(w weaver, ev events) error {
	// Find first event index.
	j := ev.AdvanceTo(w.start)
	if w.debug&sTevent > 0 {
		ev.Debug()
	}

	w = w.loadReceivers()
	// Move forwards to start date.
	w = w.loadCalendar()
	if w.next.Equal(time.Time{}) {
		return fmt.Errorf("weave: date: invalid")
	}

	// Extract calendar.
	calendar := w.r[0].ch
	w.r = w.r[1:]

	for w.next = range calendar {

		w = w.updateReceivers()

		// Perform date output function
		if w.fn[0] != nil {
			w.fn[0](w)
		}

		// Events.
		for ; j < len(ev) && ev[j].Time.Before(w.next.Time); j++ {

			// Pre Event.
			if w.fn[1] != nil {
				w.fn[1]()
			}

			// Omit events that are too recent.
			if ev[j].After(w.end) {
				break
			}

			// Run event function.
			if w.fn[2] != nil {
				w.fn[2](ev[j])
			}
		}

		// Check out.
		if w.next.After(w.end) {
			break
		}
		// Prepare for next row.
		w.current = w.next // Current time is offset by one iteration.
	}
	// Run post algorithm function.
	if w.fn[3] != nil {
		w.fn[3]()
	}
	return nil
}

func flow(start, end time.Time, ch chan event) {
	for start.Before(end) {
		start = start.Add(time.Duration(24*year) * time.Hour)
		ev := event{Time: start, valid: true}
		ch <- ev
	}
	close(ch)
}

func flow1(start, end time.Time, ch chan event) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := event{Time: start, valid: true}
		ch <- ev
	}
	close(ch)
}

func flow2(start, end time.Time, ch chan event) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := event{Time: start, valid: true}
		ch <- ev
	}
	close(ch)
}

func flow3(start, end time.Time, ch chan event) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := event{Time: start, valid: true}
		ch <- ev
	}
	close(ch)
}
