package weave

import (
	"errors"
	"fmt"
)

// ErrNilPointer points towards nothing.
var ErrNilPointer = errors.New("nil pointer")

// ErrOutOfBounds value out of bounds.
var ErrOutOfBounds = errors.New("index out of bounds")

// FnComp are user defined comparison functions for the stitch data type
// used by the weave function to generate the looms woven output,
// principally called upon by the stitches accessed from within the
// ShuttleFunc function calls.
type FnComp func(Stitch, Stitch) bool

// FnAdd is a user defined function that returns the sum of two stitch
// data values.
type FnAdd func(Stitch, Stitch) Stitch

// FnShuttle are the principle user defined functions that are an integral
// part of a loom; They are provided by the user when the loom is created
// and can be updated at any time, by calling the appropriate functions.
// The five FnShuttle functions are found at key points inside the main
// loop of the loom function they are run by the loom whilst it is
// generating its output to order stitches as the shuttle advances with
// each pick of the loom.
type FnShuttle func(Loom, Stitch) Loom

// State records the state of a thread within the shuttle.
type State int

const (
	Empty   State = 0
	Written State = 1 << iota
)

// Stitch is the primary object with which the we interacts with our data
// from the code that calls the package, accessible as it passes through
// the loom by way of the above function types; The Weave functions sort
// our streams of data, encapsulating them within channels of stitches so that
// the weave functions can pass them though the algorithm, where it is
// then sorted by the provided Comp functions, before passing into the
// ShuttleFunc's from where we can access it by way of those functions
// that are also user defined.
type Stitch struct {
	// State is to be used to define whether the stitch is in use or
	// not or if it has been written too, used user side to
	// ascertain whether a stitch's data is to be written. It may also
	// be used to define the type of the data within the interface.
	State State
	Data  interface{}
}

// Loom interleave multiple streams of data.
type Loom struct {
	// The loom is essentially an abstract data type, its function is
	// to receive and sort multiple incoming streams of data by a
	// criteria defined by user function, each stitch in the stream of
	// data must arrive in an incremental order, the job of the loom
	// is to interleave multiple consecutive streams of data, here
	// known as threads, ordering them sequentially by way of the
	// Shuttle and its user provided functions.
	Start Stitch // Start indicates the first stitch of a weave.
	End   Stitch // End indicates the last stitch of a weave.

	// User data accessible by the shuttle functions during the weave.
	UserData interface{}

	// User Functions.

	// Called before the algorithm starts.
	PreWeave FnShuttle
	// Called before the stitch function.
	PreStitch FnShuttle
	// Performed on the users data that is fed into the loom.
	Stitch FnShuttle
	// Called after the stitch function.
	PostStitch FnShuttle
	// Called after the weave algorithm.
	PostWeave FnShuttle

	// WarpOn signals that the first channel is a warp and that warp
	// mode is to be used, when false the weave uses the lowest value
	// amongst all of the channels as its next regulatory line, see
	// the threshold for more information.
	WarpOn bool
	// The warp contains the regulatory values about which the weave
	// is structured if Warp is true.
	warp warp

	// Threshold defines the increment used to group together values
	// when the warp is not set.
	Threshold Stitch

	// Output is an array into which the shuttle places the thread
	// stitches that are ready for output.
	Output []Stitch
	Shuttle
	Before FnComp
	Equal  FnComp
	After  FnComp
	Add    FnAdd

	Verbose bool
}

// Shuttle contains all of the current working threads in the loom.
type Shuttle []thread

// The current warps value is used in the shuttles output and the next
// value for lookahead.
type warp struct {
	current Stitch
	next    Stitch
}

// thread holds each stitch and its state inside of the shuttle one for
// every channel that is passed into the loom.
type thread struct {
	ch      chan Stitch
	prev    Stitch
	current Stitch
	next    Stitch
}

// firstIndex returns the index of the first stitch in the stitches array
// greater or equal to the given stitch.
func (w Loom) firstIndex(s []Stitch, n Stitch) (int, error) {
	const fname = "firstIndex"
	for i := range s {
		if w.After(s[i], n) || w.Equal(s[i], n) {
			if s[i].Data == nil {
				return 0, fmt.Errorf(
					"%s: at index %d: %w",
					fname, i, ErrNilPointer)
			}
			return i, nil
		}
	}
	return 0, fmt.Errorf("%s: %w", fname, ErrOutOfBounds)
}

// Warp returns the stitch that is currently the warp in the loom.
func (w Loom) Warp() Stitch {
	return w.warp.current
}

// Weave interleaves the objects in the stitch array with the channels
// that the user also provides, if the flag for warp is set then the first
// of these channels is used as the warp, a structure or guide for the
// output. Channels of irregularly spaced intervals may then be interleaved
// using the warp to provide structure, the channels maybe temporal or
// other. Without the warp the channels and provided objects are simply
// interleaved, the coincidence of simultaneous values is regulated by
// way of the threshold setting, reducing or increasing the exigence of the
// comparison functions effectively defining a resolution of the output.
// This can at times greatly alter the size of the output.
func (w Loom) Weave(s []Stitch, chans ...chan Stitch) error {
	const fname = "Weave"
	// The shuttle holds all the critical data from the channels
	// whilst the subroutines of the algorithm are functioning.
	w.Shuttle = make([]thread, len(chans))
	for i := range chans {
		w.Shuttle[i].ch = chans[i]
	}
	switch {
	case w.WarpOn:
		err := w.weaveWarped(s)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	default:
		err := w.weave(s)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	}
	return nil
}

// LoadData loads a user data into the weave for access inside of
// WeaveFunc's.
func (w Loom) LoadData(d interface{}) Loom {
	w.UserData = d
	return w
}

// loadSuhttle advances all threaded channels to the weave starting point.
func (w Loom) threadShuttle() (Loom, error) {
	const fname = "w.threadShuttle"
	for i := range w.Shuttle {
		// Load stitches from channels, skipping over any empty
		// values untill data is reached.
		w.Shuttle[i].next = <-w.Shuttle[i].ch
		if w.Shuttle[i].next.Data == nil {
			continue
		}
		// We need to remove all stitches haveing lower values
		// than our required starting point.
		for w.Before(w.Shuttle[i].next, w.Start) {
			w.Shuttle[i].prev = w.Shuttle[i].current
			w.Shuttle[i].current = w.Shuttle[i].next
			w.Shuttle[i].next = <-w.Shuttle[i].ch
		}
	}
	// None of our threads should have nil data at this point.
	for i := range w.Shuttle {
		if w.Shuttle[i].next.Data == nil {
			return w, fmt.Errorf("%s: index %d: %w",
				fname, i, ErrNilPointer)
		}
	}
	if w.Verbose {
		for _, r := range w.Shuttle {
			fmt.Println("r.prev:", r.prev)
			fmt.Println("r.current:", r.current)
			fmt.Println("r.next:", r.next)
		}
	}
	return w, nil
}

// loadWarp advances the warp to the required starting value.
func (w Loom) loadWarp() (Loom, error) {
	const fname = "w.loadWarp"
	// Set up lookahead.
	w.warp.next = <-w.Shuttle[0].ch
	if w.warp.next.Data == nil {
		return w, fmt.Errorf("%s: %w",
			fname, ErrNilPointer)
	}
	for {
		// Stream through the warp untill its value is greater
		// than the starting value, values should not be nil.
		w.warp.current = w.warp.next
		if w.After(w.warp.next, w.Start) {
			break
		}
		w.warp.next = <-w.Shuttle[0].ch
		if w.warp.next.Data == nil {
			return w, fmt.Errorf("%s: %w",
				fname, ErrNilPointer)
		}
	}
	return w, nil
}

// advanceToWarp pulls the shuttle threads forwards to the next warp.
func (w Loom) advanceToWarp() (Loom, error) {
	const fname = "advanceShuttle"
	var err error
	for i := range w.Shuttle {
		// For every thread in the shuttle.
		for !w.After(w.Shuttle[i].next, w.warp.next) {

			// Whilst the next value is not greater than that
			// of the value in the warp, pass another stitch.
			w.Shuttle[i].prev = w.Shuttle[i].current
			w.Shuttle[i].current = w.Shuttle[i].next
			w.Shuttle[i].next = <-w.Shuttle[i].ch

			// If we get a nil value in the data, we have
			// run out of thread, something is wrong.
			if w.Shuttle[i].next.Data == nil {
				err = fmt.Errorf(
					"%s: index %d: %w",
					fname, i, ErrNilPointer)
				break
			}
		}
	}
	return w, err
}

// weaveWarped weaves the array of stitches with predefined channels using
// a warp or grid, the first channel is used as the warp.
func (w Loom) weaveWarped(s []Stitch) error {
	const fname = "weaveWarped"
	// Load all required data from chans.
	j, err := w.firstIndex(s, w.Start)
	if err != nil && !errors.Is(err, ErrOutOfBounds) {
		return fmt.Errorf("%s: %w", fname, err)
	}
	w, err = w.threadShuttle()
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	w, err = w.loadWarp()
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	// Extract warp, the first channel that is passed into weave is
	// extracted and used as a guide or the warp for the looms output.
	warp := w.Shuttle[0].ch
	w.Shuttle = w.Shuttle[1:]
	// User output function.
	if w.PreWeave != nil {
		w = w.PreWeave(w, w.warp.current)
	}
	for w.warp.next = range warp {
		// Spool channels into the shuttle.
		w, err = w.advanceToWarp()
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
		// Clear then fill the shuttle output.
		w.Output = w.Output[:0]
		for _, thrd := range w.Shuttle {
			if w.Before(thrd.current, w.warp.next) {
				w.Output = append(w.Output, thrd.current)
			}
		}
		// User output function.
		if w.PreStitch != nil {
			w = w.PreStitch(w, w.warp.next)
		}
		// Stitches.
		for ; j < len(s) && w.Before(s[j], w.warp.next); j++ {
			// Omit Events that are too recent.
			if w.After(s[j], w.End) {
				break
			}
			if s[j].Data == nil {
				return fmt.Errorf("%s: stitches: %w",
					fname, ErrNilPointer)
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
		// Check out if required.
		if w.After(w.warp.next, w.End) {
			break
		}
		// Prepare for the next row.
		w.warp.current = w.warp.next
	}
	// User output function.
	if w.PostWeave != nil {
		w = w.PostWeave(w, w.warp.current)
	}
	return nil
}

// advanceShuttle finds the next value to output and then all values that
// are within the threshold from that and selects them for output,
// loading in new stitches into the shuttle on so doing.
func (w Loom) advanceShuttle() (Loom, error) {

	const fname = "w.advanceShuttle"
	w.warp.current = w.Shuttle[0].next

	// Find the next stitch, that with the lowest value and then set
	// the threshold from that.
	for i := range w.Shuttle {
		if w.Before(w.Shuttle[i].next, w.warp.current) {
			w.warp.current = w.Shuttle[i].next
		}
	}
	w.warp.next = w.Add(w.warp.current, w.Threshold)

	// Set the threads that have values beneath or equal to the
	// threshold into the output array and load the next stitch into
	// the shuttle, set the state on the output stitches so that they
	// can be dealt with appropriately in the stitch user function.
	for i := range w.Shuttle {
		if !w.After(w.Shuttle[i].next, w.warp.next) {
			w.Output[i] = w.Shuttle[i].next
			w.Output[i].State = Written
			w.Shuttle[i].next = <-w.Shuttle[i].ch
			if w.Shuttle[i].next.Data == nil {
				return w, fmt.Errorf("%s: nil pointer", fname)
			}
			continue
		}
		w.Output[i].State = Empty
	}
	return w, nil
}

// weave interleaves an array of stitches along with the weaving of
// predefined channels,
func (w Loom) weave(s []Stitch) error {
	const fname = "weave"
	// Load all required data from chans.
	j, err := w.firstIndex(s, w.Start)
	if err != nil && !errors.Is(err, ErrOutOfBounds) {
		return fmt.Errorf("%s: %w", fname, err)
	}
	w, err = w.threadShuttle()
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	// Set the length of the output array to match that of the
	// Shuttle.
	w.Output = w.Output[:len(w.Shuttle)]

	for {
		// User output function.
		if w.PreStitch != nil {
			w = w.PreStitch(w, w.warp.next)
		}
		// Load w.Output.
		w, err = w.advanceShuttle()
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
		// Stitches.
		for ; j < len(s) && w.Before(s[j], w.warp.next); j++ {
			// Omit Events that are too recent.
			if w.After(s[j], w.End) {
				break
			}
			if s[j].Data == nil {
				return fmt.Errorf("%s: stitches: %w",
					fname, ErrNilPointer)
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
		// Check out if required.
		if w.After(w.warp.next, w.End) {
			break
		}
		// Prepare for the next row.
		w.warp.current = w.warp.next
	}
	// User output function.
	if w.PostWeave != nil {
		w = w.PostWeave(w, w.warp.current)
	}
	return nil
}
