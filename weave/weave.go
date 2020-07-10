package weave

import (
	"fmt"

	"github.com/8i8/jyo/weave/svg"
)

const (
	biTload = 1 << iota
	biTstitch
)

// Loom interleave multiple streams of data.
type Loom struct {
	// The loom is essentially an abstract data type, its function is to
	// receive and sort multiple incoming streams of data by the criteria
	// given by user defined function, each stitch in the stream of data
	// must arrive in an incremental order, the job of the loom is to
	// interleave multiple consecutive streams of data, here known as
	// threads, ordering them sequentially by way of the Shuttle and its
	// user provided functions.
	Start Stitch // Start indicates the first stitch of a weave.
	End   Stitch // End indicates the last stitch of a weave.

	// User data accessible by the shuttle functions during the weave.
	UserData interface{}

	// User Functions.
	PreWeave   FnShuttle // Called before the algorithm starts.
	PreStitch  FnShuttle // Called before the stitch function.
	Stitch     FnShuttle // Performed on the users data that is fed into the loom.
	PostStitch FnShuttle // Called after the stitch function.
	PostWeave  FnShuttle // Called after the weave algorithm.

	// WarpOn signals that the first channel is a warp and that warp mode is
	// to be used, when false the weave uses the lowest value amongst all
	// af the channels as its next regulatory line, see the threshold for
	// more information.
	WarpOn bool
	// The warp contins the regulatory values about which the weave is
	// structured if Warp is true.
	warp warp

	// Output is an array into which the shuttle places the thread stitches
	// that are ready for output.
	Output []Stitch
	Shuttle
	Before FnComp
	Equal  FnComp
	After  FnComp

	Verbose   bool
	Threshold interface{}
}

// Shuttle contains all of the current working threads in the loom.
type Shuttle []thread

// The current warps value is used in the shuttles output and the next value
// for lookahead.
type warp struct {
	current Stitch
	next    Stitch
}

// thread holds each stitch and its state inside of the shuttle one for every
// channel that is passed into the loom.
type thread struct {
	ch      chan Stitch
	prev    Stitch
	current Stitch
	next    Stitch
}

// Stitch is the primary object with which the user interacts with their data,
// it is accesable as it passes through the loom by way of user provided
// functions; The Weave fuctions sort the streams of stitches arriving from the
// channels inside of the encapsulating shuttle. These functions direct the
// loom after accessing the data within the Data interface.
type Stitch struct {
	State int
	Data  interface{}
}

// Comp are user defined comparison functions for the stitch data type used by
// the weave function to generate the looms woven output, principally called
// upon by the stitches accessed from within the ShuttleFunc function calls.
type Comp func(Stitch, Stitch) bool

// CompFuncs is a user defined array of comparison functions that enable the loom
// to sort threads within the shuttle.
type CompFuncs []Comp

// advanceTo returns the index of the first stitch in the stitches array that
// is greater or equal to the given stitch.
func (w Loom) advanceTo(s []Stitch, n Stitch) int {
	for i := range s {
		if w.After(s[i], n) || w.Equal(s[i], n) {
			return i
		}
	}
	return 0
}

// ShuttleFunc are the principle user defined functions that are an integral
// part of a loom; They are provided by the user when the loom is created and
// can be updated at any time, by calling the appropriate functions. The five
// ShuttleFunc functions are found at key points inside the main loop of the
// loom function they are run by the loom whilst it is generating its output to
// order stitches as the shuttle advances with each pick of the loom.
type ShuttleFunc func(Loom, Stitch) Loom

// State records the state of a thread withing the shuttle.
type State int // The state of a thread withing the shuttle.

// Weave interleaves the objects in the stitch array with the channels that the
// user also provides, if the flag for warp is set then the first of these
// channels is used as the warp, a structure or guide for the output. A channel
// of non regularly spaced events can be interleaved with this channel of
// regularly spaced intervals, be they temporal or other. Without the warp the
// channels and provided objects are interleaved, the coincidence of values can
// be regulated by way of the threshold setting, reducing slightly the exigence
// of the comparison functions, can at times greatly decrease the size of the
// output without harming the legibility of the data, effectively in certain
// cases increasing it.
func (w Loom) Weave(s []Stitch, chans ...chan Stitch) {

	// The shuttle holds all the critical data from the channels whilst the
	// subroutines of the algorithm are functioning.
	w.shuttle = make([]thread, len(chans))
	for i := range chans {
		w.shuttle[i].ch = chans[i]
	}
	w.Warp = true
	if w.Warp {
		w.weaveWarped(s)
	}
}

// LoadData loads a user data into the weave for access inside of WeaveFunc's.
func (w Loom) LoadData(d interface{}) Loom {
	w.UserData = d
	return w
}

// loadSuhttle advances all threaded channels to the weave starting point.
func (w Loom) loadShuttle() Loom {
	for i := range w.shuttle {
		// Load stitches from channels.
		w.shuttle[i].next = <-w.shuttle[i].ch
		// Skip over closed channels.
		if w.shuttle[i].next.Data == nil {
			continue
		}
		// Remove all yantra in between required dates.
		for w.Before(w.shuttle[i].next, w.Start) {
			w.shuttle[i].prev = w.shuttle[i].current
			w.shuttle[i].current = w.shuttle[i].next
			w.shuttle[i].next = <-w.shuttle[i].ch
		}
	}
	for i := range w.shuttle {
		w.shuttle[i].prev = w.shuttle[i].current
		w.shuttle[i].current = w.shuttle[i].next
		w.shuttle[i].next = <-w.shuttle[i].ch
	}

	if w.Verbose {
		for _, r := range w.shuttle {
			fmt.Println("r.prev:", r.prev)
			fmt.Println("r.current:", r.current)
			fmt.Println("r.next:", r.next)
		}
	}

	return w
}

func (w Loom) threadShuttle() Loom {
	for i := range w.shuttle {
		if w.shuttle[i].current.Data == nil {
			continue
		}
		for w.Before(w.shuttle[i].next, w.warp.next) || w.Equal(w.shuttle[i].next, w.warp.next) {
			w.shuttle[i].prev = w.shuttle[i].current
			w.shuttle[i].current = w.shuttle[i].next
			w.shuttle[i].next = <-w.shuttle[i].ch
			if w.shuttle[i].next.Data == nil {
				break
			}
		}
	}
	return w
}

func (w Loom) loadWarp() Loom {
	for {
		w.warp.current = w.warp.next // Set up lookahead.
		w.warp.next = <-w.shuttle[0].ch
		if w.After(w.warp.next, w.Start) {
			break
		}
	}
	w.warp.current = w.warp.next
	w.warp.next = <-w.shuttle[0].ch
	return w
}

func (w Loom) weaveWarped(s []Stitch) error {
	// Load all required data from chans.
	j := w.advanceTo(s, w.Start)
	w = w.loadShuttle()
	w = w.loadWarp()
	// Extract warp, the first channel that is passed into weave is
	// extracted and used as a guide or the warp for the looms output.
	warp := w.shuttle[0].ch
	w.shuttle = w.shuttle[1:]
	// User output function.
	if w.PreWeave != nil {
		type mySvg struct {
			svg.Image
		}

		w = w.PreWeave(w, w.warp.current)
	}
	for w.warp.next = range warp {

		// Spool channels into the shuttle.
		w = w.threadShuttle()

		// Clear then fill the shuttle output.
		w.Output = w.Output[:0]
		for _, t := range w.shuttle {
			if w.Before(t.current, w.warp.next) {
				w.Output = append(w.Output, t.current)
			}
		}
		// User output function.
		if w.PreStitch != nil {
			w = w.PreStitch(w, w.warp.current)
		}
		// Stitches.
		for ; j < len(s) && w.Before(s[j], w.warp.next); j++ {
			// Omit Events that are too recent.
			if w.After(s[j], w.End) {
				break
			}
			// User output function.
			if w.Stitch != nil {
				w = w.Stitch(w, s[j])
			}
		}
		// User output function.
		if w.PostStitch != nil {
			w = w.PostStitch(w, w.warp.current)
		}

		// Prepare for the next row.
		w.warp.current = w.warp.next

		// Check out when required.
		if w.After(w.warp.next, w.End) {
			break
		}
	}
	// User output function.
	if w.PostWeave != nil {
		w = w.PostWeave(w, w.warp.current)
	}
	return nil
}
