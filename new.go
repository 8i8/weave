package main

//import (
//	"fmt"
//	"log"
//	"math/rand"
//	"sort"
//	"time"
//)

//// Start function.
//func fn1(in ...interface{}) {
//	w := in[0].(weaver)
//	fmt.Println("time grid:", w.current)
//	for i, r := range w.r {
//		if r.current.Before(w.next.Time) {
//			fmt.Println("receiver:", i, r.current)
//		}
//	}
//}

//// Pre event.
//func fn2(in ...interface{}) {
//}

//// Event function.
//func fn3(in ...interface{}) {
//	ev := in[0].(event)
//	if !ev.valid {
//		return
//	}
//	fmt.Println("event:", ev.Time)
//}

//// Post algorithm.
//func fn4(in ...interface{}) {
//}

//func main() {
//	loc, err := time.LoadLocation("Europe/London")
//	if err != nil {
//		log.Fatal(err)
//	}
//	d1 := time.Date(2000, 1, 1, 0, 0, 0, 0, loc)
//	d2 := time.Date(2021, 1, 1, 0, 0, 0, 0, loc)
//	chans := make([]chan event, 4)
//	for i := range chans {
//		chans[i] = make(chan event)
//	}
//	go flow(d1, d2, chans[0])
//	go flow1(d1, d2, chans[1])
//	go flow2(d1, d2, chans[2])
//	go flow3(d1, d2, chans[3])

//	ev := generateEvents(5, d1, d2)

//	Weave(ev, d1, d2, chans[0], chans[1], chans[2], chans[3])
//}

//// Flow

//const year = 365
//const (
//	sTload = 1 << iota
//	sTevent
//)

//func init() {
//	rand.Seed(time.Now().UnixNano())
//}

//type weaver struct {
//	start   time.Time
//	end     time.Time
//	current event
//	next    event
//	r       []receiver
//	fn      []func(...interface{})
//	debug   int
//}

//type receiver struct {
//	ch      chan event
//	prev    event
//	current event
//	next    event
//}

//func (w weaver) load() weaver {
//	for i := range w.r {
//		// Load yantra.
//		w.r[i].next = <-w.r[i].ch
//		// Skip over closed channels.
//		if !w.r[i].next.valid {
//			continue
//		}
//		// Remove all yantra in between required dates.
//		for w.r[i].next.Before(w.start) && w.r[i].next.valid {
//			w.r[i].prev = w.r[i].current
//			w.r[i].current = w.r[i].next
//			w.r[i].next = <-w.r[i].ch
//		}
//	}
//	for i := range w.r {
//		w.r[i].prev = w.r[i].current
//		w.r[i].current = w.r[i].next
//		w.r[i].next = <-w.r[i].ch
//	}

//	if w.debug&sTload > 0 {
//		for _, r := range w.r {
//			fmt.Println("r.prev:", r.prev)
//			fmt.Println("r.current:", r.current)
//			fmt.Println("r.next:", r.next)
//		}
//	}

//	return w
//}

//func (w weaver) advance() weaver {
//	for i := range w.r {
//		if !w.r[i].current.valid {
//			continue
//		}
//		for w.r[i].next.Before(w.next.Time) || w.r[i].next.Equal(w.next.Time) {
//			w.r[i].prev = w.r[i].current
//			w.r[i].current = w.r[i].next
//			w.r[i].next = <-w.r[i].ch
//			if !w.r[i].next.valid {
//				break
//			}
//		}
//	}
//	return w
//}

//// Weave builds and runs a waver.
//func Weave(ev events, start, end time.Time, chans ...chan event) {
//	w := weaver{
//		start: start,
//		end:   end,
//		//debug: sTload | sTevent,
//	}
//	w.r = make([]receiver, len(chans))
//	w.fn = make([]func(...interface{}), 4)
//	w.fn[0] = fn1
//	w.fn[1] = fn2
//	w.fn[2] = fn3
//	w.fn[3] = fn4
//	for i := range chans {
//		w.r[i].ch = chans[i]
//	}
//	weave(w, ev)
//}

//func weave(w weaver, ev events) error {

//	w = w.load()

//	// Extract calendar.
//	calendar := w.r[0].ch
//	w.r = w.r[1:]
//	w.current = <-calendar // Set up lookahead.
//	w.next = <-calendar
//	if w.next.Equal(time.Time{}) {
//		return fmt.Errorf("weave: date: invalid")
//	}

//	// Event index.
//	j := ev.AdvanceTo(w.start)
//	if w.debug&sTevent > 0 {
//		ev.Debug()
//	}

//	for w.next = range calendar {
//		// Forward the date to the start date.
//		if w.next.Before(w.start) {
//			continue
//		}

//		// Update all receivers.
//		w = w.advance()

//		// Perform date output function
//		if w.fn[0] != nil {
//			w.fn[0](w)
//		}

//		// Events.
//		for ; j < len(ev) && ev[j].Time.Before(w.next.Time); j++ {

//			// Pre Event.
//			if w.fn[1] != nil {
//				w.fn[1]()
//			}

//			// Omit events that are too recent.
//			if ev[j].After(w.end) {
//				break
//			}

//			// Run event function.
//			if w.fn[2] != nil {
//				w.fn[2](ev[j])
//			}
//		}

//		// Check out.
//		if w.next.After(w.end) {
//			break
//		}
//		// Prepare for next row.
//		w.current = w.next // Current time is offset by one iteration.
//	}
//	// Run post algorithm function.
//	if w.fn[3] != nil {
//		w.fn[3]()
//	}
//	return nil
//}

//func flow(start, end time.Time, ch chan event) {
//	for start.Before(end) {
//		start = start.Add(time.Duration(24*year) * time.Hour)
//		ev := event{Time: start, valid: true}
//		ch <- ev
//	}
//	close(ch)
//}

//func flow1(start, end time.Time, ch chan event) {
//	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
//	for start.Before(end) {
//		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
//		ev := event{Time: start, valid: true}
//		ch <- ev
//	}
//	close(ch)
//}

//func flow2(start, end time.Time, ch chan event) {
//	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
//	for start.Before(end) {
//		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
//		ev := event{Time: start, valid: true}
//		ch <- ev
//	}
//	close(ch)
//}

//func flow3(start, end time.Time, ch chan event) {
//	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
//	for start.Before(end) {
//		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
//		ev := event{Time: start, valid: true}
//		ch <- ev
//	}
//	close(ch)
//}

//// Events

//type events []event

//type event struct {
//	time.Time
//	valid bool
//}

//func generateEvents(num int, start, end time.Time) events {

//	ev := make(events, num)
//	st := start.Unix()
//	ed := end.Unix()
//	rng := ed - st

//	for i := range ev {
//		rng = rand.Int63()%rng + 1
//		date := time.Unix(st+rng, 0)
//		ev[i].Time = date
//		ev[i].valid = true
//	}

//	if num > 1 {
//		sort.Sort(events(ev))
//	}

//	return ev
//}

//func (e events) AdvanceTo(t time.Time) int {
//	for i := range e {
//		if e[i].After(t) || e[i].Equal(t) {
//			return i
//		}
//	}
//	return 0
//}

//func (e events) Debug() {
//	for i := range e {
//		fmt.Println("event:", e[i].Time)
//	}
//}

//// Sort functions.

//// Len returns the length of the events list.
//func (e events) Len() int {
//	return len(e)
//}

//// Less returns a boolean response to the question is e[i] less than e[j].
//func (e events) Less(i, j int) bool {
//	return e[i].Before(e[j].Time)
//}

//// Swap inverses the positions of e[i] and e[j].
//func (e events) Swap(i, j int) {
//	e[i], e[j] = e[j], e[i]
//}
