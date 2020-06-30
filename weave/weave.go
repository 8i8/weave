package weave

import (
	"fmt"
)

const (
	biTload = 1 << iota
	biTstitch
)

// ShuttleFunc are the user functions run by the loom to generate its output,
// facilitaing the use of closures when calling the functions.
type ShuttleFunc func(Loom, Stitch) Loom

// Loom interleave multiple streams of dates.
type Loom struct {
	Start      Stitch
	End        Stitch
	Data       interface{}
	Output     Threads
	PreWeave   ShuttleFunc
	PreStitch  ShuttleFunc
	Stitch     ShuttleFunc
	PostStitch ShuttleFunc
	PostWeave  ShuttleFunc
	current    Stitch
	next       Stitch
	shuttle    []Shuttle
	comp       []Comp
	debug      int
}

// Shuttle carries a thread for every channel passed into the Loom when it is
// created.
type Shuttle struct {
	ch      chan Stitch
	prev    thread
	current thread
	next    thread
}

// State contains the state of a thread, accessible and maintained by the user
// to maintain and pass state between stitches.
type State int

// Comp is a function type for the comparison functions on the thread type used
// by all weave functions and its subroutines.
type Comp func(Stitch, Stitch) bool

// CompFuncs is a user defined array of comparison functions that the loom
// requires to work.
type CompFuncs []Comp

// thread maintins the state of a channels weaving operation.
type thread struct {
	State int
	Stitch
	before Comp
	equal  Comp
	after  Comp
}

func (t thread) Before(n Stitch) bool {
	return t.before(t.Stitch, n)
}

func (t thread) Equal(n Stitch) bool {
	return t.equal(t.Stitch, n)
}

func (t thread) After(n Stitch) bool {
	return t.after(t.Stitch, n)
}

// Threads enables the sorting of threads inside the shuttle.
type Threads []thread

// Sort functions.

// Len returns the length of the Events list.
func (t Threads) Len() int {
	return len(t)
}

// Less retuns a boolean response to the question is e[i] less than e[j].
func (t Threads) Less(i, j int) bool {
	return t[i].Before(t[j].Stitch)
}

// Swap inverses the positions of e[i] and e[j].
func (t Threads) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Cloth builds and runs a weaver, the first chanel that is provided is used as
// a grid or ruler for structuring of all of the folling channel.
func (w Loom) Cloth(c CompFuncs, s Stitches, chans ...chan Stitch) {
	//w.debug= biTload | biTevent,
	w.shuttle = make([]Shuttle, len(chans))
	for i := range chans {
		w.shuttle[i].ch = chans[i]
	}
	for i, s := range w.shuttle {
		w.Start = w.Start.LoadFuncs(c)
		w.End = w.End.LoadFuncs(c)
		w.shuttle[i].prev = s.prev.loadThreadFn(c)
		w.shuttle[i].current = s.current.loadThreadFn(c)
		w.shuttle[i].next = s.next.loadThreadFn(c)
	}
	w.weaveWarped(s)
}

// LoadData loads a user data into the weave for access inside of WeaveFunc's.
func (w Loom) LoadData(d interface{}) Loom {
	w.Data = d
	return w
}

func (t thread) loadThreadFn(c CompFuncs) thread {
	t.before, t.equal, t.after = c[0], c[1], c[2]
	return t
}

// LoadFnComp loads the users comparison functions into the looms shuttle
// threads.
func (w Loom) LoadFnComp(before, equal, after Comp) CompFuncs {
	var c CompFuncs
	c[0], c[1], c[2] = before, equal, after
	return c
}

func (w Loom) loadShuttle() Loom {
	for i := range w.shuttle {
		// Load yantra.
		w.shuttle[i].next.Stitch = <-w.shuttle[i].ch
		// Skip over closed channels.
		if w.shuttle[i].next.Data == nil {
			continue
		}
		// Remove all yantra in between required dates.
		for w.shuttle[i].next.Before(w.Start) {
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

func (w Loom) passShuttle() Loom {
	for i := range w.shuttle {
		if w.shuttle[i].current.Data == nil {
			continue
		}
		for w.shuttle[i].next.Before(w.next) || w.shuttle[i].next.Equal(w.next) {
			w.shuttle[i].prev = w.shuttle[i].current
			w.shuttle[i].current = w.shuttle[i].next
			w.shuttle[i].next.Stitch = <-w.shuttle[i].ch
			if w.shuttle[i].next.Data == nil {
				break
			}
		}
	}
	return w
}

func (w Loom) loadWarp() Loom {
	for {
		w.current = w.next // Set up lookahead.
		w.next = <-w.shuttle[0].ch
		if w.next.After(w.Start) {
			break
		}
	}
	w.current = w.next
	w.next = <-w.shuttle[0].ch
	return w
}

func (w Loom) weaveWarped(s Stitches) error {
	// Load all required data from chans.
	j := s.advanceTo(w.Start)
	if w.debug&biTstitch > 0 {
		s.debug()
	}
	w = w.loadShuttle()
	w = w.loadWarp()
	// Extract warp, the firts channel that is passed into weave is
	// extracted and used as a guide or the warp for the looms output.
	warp := w.shuttle[0].ch
	w.shuttle = w.shuttle[1:]
	// User output function.
	if w.PreWeave != nil {
		w = w.PreWeave(w, w.current)
	}
	for w.next = range warp {
		w = w.passShuttle()
		w.Output = w.Output[:0]
		for _, t := range w.shuttle {
			if t.current.Before(w.next) {
				w.Output = append(w.Output, t.current)
			}
		}
		// User output function.
		if w.PreStitch != nil {
			w = w.PreStitch(w, w.current)
		}
		// Stitches.
		for ; j < len(s) && s[j].Before(w.next); j++ {
			// Omit Events that are too recent.
			if s[j].After(w.End) {
				break
			}
			// User output function.
			if w.Stitch != nil {
				w = w.Stitch(w, s[j])
			}
		}
		// User output function.
		if w.PostStitch != nil {
			w = w.PostStitch(w, w.current)
		}

		// Prepare for the next row.
		w.current = w.next

		// Check out when required.
		if w.next.After(w.End) {
			break
		}
	}
	// User output function.
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
