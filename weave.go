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
type FnAdd func(Stitch, interface{}) Stitch

// FnShuttle are the principle user defined functions that are an integral
// part of a loom; They are provided by the user when the loom is created
// and can be updated at any time, by calling the appropriate functions.
// The five FnShuttle functions are found at key points inside the main
// loop of the loom function they are run by the loom whilst it is
// generating its output to order stitches as the shuttle advances with
// each pick of the loom.
type FnShuttle func(Loom, Stitch) (Loom, error)

// State records the state of a thread within a Stitch, used within the
// loom to mark stitches for writing, or other.
type State int

const (
	// Stale indicates that the data is not new, there is no state
	// other than that of being stale.
	Stale State = 0
	// Fresh indicates that there is new data and contains a bitfield
	// for holding state.
	Fresh State = 1 << iota
)

// Stitch is the primary object with which the we interacts with our data
// from the code that calls the package, accessible as it passes through
// the loom by way of the above function types; The Weave functions sort
// our streams of data, encapsulating them within channels of stitches so that
// the weave functions can pass them through the algorithm, where it is
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

// Loom is essentially an abstract data type, its function is to receive
// and sort multiple incoming streams of data by a criteria defined by
// user function, each stitch in the stream of data must arrive in an
// incremental order, the job of the loom is to interleave multiple
// consecutive streams of data, here known as threads, ordering them
// sequentially by way of the Shuttle and its user provided functions.
type Loom struct {
	// Start indicates the first stitch of a weave.
	Start Stitch
	// End indicates the last stitch of a weave.
	End Stitch

	// User data accessible by the shuttle functions during the weave.
	UserData interface{}

	// User Functions.

	// PreWeave is called before the algorithm starts.
	PreWeave FnShuttle
	// PreStitch is called before the stitch function.
	PreStitch FnShuttle
	// Stitch is the central function of the algorithm.
	Stitch FnShuttle
	// PostStitch is alled after the stitch function.
	PostStitch FnShuttle
	// PostWeave is called after the weave algorithm.
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
	Threshold interface{}

	// Output is an array into which the shuttle places the thread
	// stitches that are ready for output.
	Output []Stitch
	// stage is a holding space for delaying or offsetting the output
	// by one iteration.
	stage []Stitch
	// shuttle contains the looms threads.
	shuttle
	// After is a function used to compare two stitches.
	Before FnComp
	// After is a function used to compare two stitches.
	Equal FnComp
	// After is a function used to compare two stitches.
	After FnComp
	// Add is the function used to addition two stitches.
	Add FnAdd
}

// shuttle contains all of the current working threads in the loom.
type shuttle []thread

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
	current Stitch
	next    Stitch
}

// firstIndex returns the index of the first stitch in the stitches array
// greater than or equal to the given stitch.
func (l Loom) firstIndex(s []Stitch, n Stitch) (int, error) {
	const fname = "Loom.firstIndex"
	for i := range s {
		if l.After(s[i], n) || l.Equal(s[i], n) {
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

// Warp returns the stitch that is currently in the warp of the loom.
func (l Loom) Warp() Stitch {
	return l.warp.current
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
func (l Loom) Weave(s []Stitch, chans ...chan Stitch) error {
	const fname = "Loom.Weave"
	// The shuttle holds all the critical data from the channels
	// whilst the subroutines of the algorithm are functioning.
	l.shuttle = make([]thread, len(chans))
	for i := range chans {
		l.shuttle[i].ch = chans[i]
	}
	switch {
	case l.WarpOn:
		err := l.weaveWarped(s)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	default:
		err := l.weave(s)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	}
	return nil
}

// LoadData loads a user data into the weave for access inside of
// WeaveFunc's.
func (l Loom) LoadData(d interface{}) Loom {
	l.UserData = d
	return l
}

// setShuttle advances all threaded channels to the weave starting point.
func (l Loom) setShuttle() (Loom, error) {
	const fname = "Loom.setShuttle"
	for i := range l.shuttle {
		// Load stitches from channels, skipping over any empty
		// values untill data is reached.
		l.shuttle[i].next = <-l.shuttle[i].ch
		if l.shuttle[i].next.Data == nil {
			continue
		}
		// We need to remove all stitches haveing lower values
		// than our required starting point.
		for l.Before(l.shuttle[i].next, l.Start) {
			l.shuttle[i].current = l.shuttle[i].next
			l.shuttle[i].next = <-l.shuttle[i].ch
		}
	}
	// None of our threads should have nil data at this point.
	for i := range l.shuttle {
		if l.shuttle[i].next.Data == nil {
			return l, fmt.Errorf("%s: index %d: %w",
				fname, i, ErrNilPointer)
		}
	}
	return l, nil
}

// loadWarp advances the warp to the required starting value.
func (l Loom) loadWarp() (Loom, error) {
	const fname = "Loom.loadWarp"
	// Set up lookahead.
	l.warp.next = <-l.shuttle[0].ch
	if l.warp.next.Data == nil {
		return l, fmt.Errorf("%s: %w",
			fname, ErrNilPointer)
	}
	for {
		// Stream through the warp untill its value is greater
		// than the starting value, values should not be nil.
		l.warp.current = l.warp.next
		if l.After(l.warp.next, l.Start) {
			break
		}
		l.warp.next = <-l.shuttle[0].ch
		if l.warp.next.Data == nil {
			return l, fmt.Errorf("%s: %w",
				fname, ErrNilPointer)
		}
	}
	return l, nil
}

// threadShuttelAndWarp pulls the shuttle threads forwards to the next
// warp.
func (l Loom) threadShuttelAndWarp() (Loom, error) {
	const fname = "Loom.threadShuttelAndWarp"
	var err error
	for i := range l.shuttle {
		// For every thread in the shuttle.
		var once bool
		for !l.After(l.shuttle[i].next, l.warp.next) {
			// Whilst the next value is not greater than that
			// of the value of the warp, pass another stitch.
			if !once {
				l.shuttle[i].current = l.shuttle[i].next
				once = true
			}
			l.shuttle[i].next = <-l.shuttle[i].ch
			// If we get a nil value in the data, we have
			// run out of thread, something is wrong.
			if l.shuttle[i].next.Data == nil {
				err = fmt.Errorf(
					"%s: index %d: %w",
					fname, i, ErrNilPointer)
				break
			}
		}
	}
	// Load Output.
	copy(l.Output, l.stage)
	// Load the next stage.
	for i, thrd := range l.shuttle {
		if l.Before(thrd.current, l.warp.next) {
			l.stage[i] = thrd.current
			l.stage[i].State = Fresh
			// w.stage = append(w.stage, thrd.current)
			continue
		}
		l.stage[i].State = Stale
	}
	return l, err
}

// weaveWarped weaves the array of stitches with predefined channels using
// a warp or grid, the first channel is used as the warp.
func (l Loom) weaveWarped(s []Stitch) error {
	const fname = "Loom.weaveWarped"
	// Load all required data from chans.
	j, err := l.firstIndex(s, l.Start)
	if err != nil && !errors.Is(err, ErrOutOfBounds) {
		return fmt.Errorf("%s: %w", fname, err)
	}
	l, err = l.setShuttle()
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	l, err = l.loadWarp()
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	// Extract warp, the first channel that is passed into weave is
	// extracted and used as a guide or the warp for the looms output.
	warp := l.shuttle[0].ch
	l.shuttle = l.shuttle[1:]
	// Set the length of the output array to match that of the
	// Shuttle.
	l.Output = make([]Stitch, len(l.shuttle))
	l.stage = make([]Stitch, len(l.shuttle))
	l, err = l.threadShuttelAndWarp()
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	// User output function.
	if l.PreWeave != nil {
		l, err = l.PreWeave(l, l.warp.current)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	}
	for l.warp.next = range warp {
		// Spool channels into the shuttle.
		l, err = l.threadShuttelAndWarp()
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
		// User output function.
		if l.PreStitch != nil {
			l, err = l.PreStitch(l, l.warp.current)
			if err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
		}
		// Stitches.
		for ; j < len(s) && l.Before(s[j], l.warp.next); j++ {
			// Omit Events that are too recent.
			if l.After(s[j], l.End) {
				break
			}
			if s[j].Data == nil {
				return fmt.Errorf("%s: stitches: %w",
					fname, ErrNilPointer)
			}
			// User output function.
			if l.Stitch != nil {
				l, err = l.Stitch(l, s[j])
				if err != nil {
					return fmt.Errorf("%s: %w", fname, err)
				}
			}
		}
		// User output function.
		if l.PostStitch != nil {
			l, err = l.PostStitch(l, l.warp.current)
			if err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
		}
		// Check out if required.
		if l.After(l.warp.current, l.End) {
			break
		}
		// Prepare for the next row.
		l.warp.current = l.warp.next
	}
	// User output function.
	if l.PostWeave != nil {
		l, err = l.PostWeave(l, l.warp.current)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	}
	return nil
}

// threadShuttle finds the next value to output and then all values that
// are within the threshold from that and selects them for output,
// loading in new stitches into the shuttle on so doing.
func (l Loom) threadShuttle() (Loom, error) {

	const fname = "Loom.threadShuttle"
	least := l.shuttle[0].next
	l.warp.next = l.shuttle[0].next

	// Find the stitch with the lowest value we will use this to set
	// the next warp.
	n := 0
	for i := range l.shuttle {
		if l.Before(l.shuttle[i].next, least) {
			least = l.shuttle[i].next
			n = i
		}
	}
	// Set the warp next value for use in output and eventually
	// keeping track of the current values required for output to used
	// functions.
	l.warp.next = l.shuttle[n].next

	// Advance least to generate a threshold, if required.
	if l.Add != nil {
		least = l.Add(least, l.Threshold)
	}

	// Transfer over the previous itterations data, this offset of one
	// itteration is required to offset the output of the bhukti to
	// the correct date.
	copy(l.Output, l.stage)

	// Set threads that have values lower or equal to that of the
	// threshold in the output array and then load the next stitch
	// into the shuttle; Set the state of the newly output stitches to
	// Fresh and those that have not been updated to Stale; This may
	// be required in the user functions.
	for i := range l.shuttle {
		if l.shuttle[i].next.Data == nil {
			return l, fmt.Errorf("%s: nil pointer", fname)
		}
		if !l.After(l.shuttle[i].next, least) {
			l.stage[i] = l.shuttle[i].next
			l.stage[i].State = Fresh
			l.shuttle[i].current = l.shuttle[i].next
			l.shuttle[i].next = <-l.shuttle[i].ch
			continue
		}
		l.stage[i] = l.shuttle[i].current
		l.stage[i].State = Stale
	}
	return l, nil
}

// weave interleaves an array of stitches along with the weaving of
// predefined channels,
func (l Loom) weave(s []Stitch) error {
	const fname = "Loom.weave"
	// Load all required data from chans.
	j, err := l.firstIndex(s, l.Start)
	if err != nil && !errors.Is(err, ErrOutOfBounds) {
		return fmt.Errorf("%s: %w", fname, err)
	}
	l, err = l.setShuttle()
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	// Set the length of the output array to match that of the
	// Shuttle.
	l.Output = make([]Stitch, len(l.shuttle))
	l.stage = make([]Stitch, len(l.shuttle))

	//firstRun := true   // Inhibit reading on the first pass.
	breakNext := false // Allow the last line to be displayed.

	l, err = l.threadShuttle()
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	l.warp.current = l.warp.next

	// User output function.
	if l.PreWeave != nil {
		l, err = l.PreWeave(l, l.warp.current)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	}
	for {
		l, err = l.threadShuttle()
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
		// User output function.
		if l.PreStitch != nil {
			l, err = l.PreStitch(l, l.warp.current)
			if err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
		}
		// Stitches.
		for ; j < len(s) && l.Before(s[j], l.warp.next); j++ {
			// Omit Events that are too recent.
			if l.After(s[j], l.End) {
				break
			}
			if s[j].Data == nil {
				return fmt.Errorf("%s: stitches: %w",
					fname, ErrNilPointer)
			}
			// User output function.
			if l.Stitch != nil {
				l, err = l.Stitch(l, s[j])
				if err != nil {
					return fmt.Errorf("%s: %w", fname, err)
				}
			}
		}
		// User output function.
		if l.PostStitch != nil {
			l, err = l.PostStitch(l, l.warp.current)
			if err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
		}
		// Check out if required.
		if breakNext {
			break
		}
		// Prepare for the next row.
		l.warp.current = l.warp.next
		if l.After(l.warp.current, l.End) {
			breakNext = true
		}
	}
	// User output function.
	if l.PostWeave != nil {
		l, err = l.PostWeave(l, l.warp.current)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	}
	return nil
}
