package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/8i8/jyo/weave/svg"
	"github.com/8i8/jyo/weave/weave"
)

const year = 365

// Sort function required by the loom.
func lessThan(a, b weave.Stitch) bool {
	i := a.Data.(time.Time)
	j := b.Data.(time.Time)
	return i.Before(j)
}

func equalTo(a, b weave.Stitch) bool {
	i := a.Data.(time.Time)
	j := b.Data.(time.Time)
	return i.Equal(j)
}

func greaterThan(a, b weave.Stitch) bool {
	i := a.Data.(time.Time)
	j := b.Data.(time.Time)
	return i.After(j)
}

// shuttle functions required by the loom.
func startSVG() weave.ShuttleFunc {
	return func(w weave.Loom, e weave.Stitch) weave.Loom {
		d := w.UserData.(mySvg)
		d.Image = d.Image.ViewBox(os.Stdout, 0, 0, 100, 100)
		w.UserData = d
		return w
	}
}

func writeDasa() weave.ShuttleFunc {
	return func(w weave.Loom, e weave.Stitch) weave.Loom {
		fmt.Println("time grid:", e.Data)
		for _, m := range w.Output {
			fmt.Println("reciever:", m.Data)
		}
		return w
	}
}

func writeEvent() weave.ShuttleFunc {
	return func(w weave.Loom, s weave.Stitch) weave.Loom {
		if s.Data == nil {
			return w
		}
		fmt.Println("event:", s.Data)
		return w
	}
}

func closeSVG() weave.ShuttleFunc {
	return func(w weave.Loom, e weave.Stitch) weave.Loom {
		svg := w.UserData.(mySvg)
		svg.End()
		return w
	}
}

// Channels that stream stitches for the loom.
func goFlow(start, end time.Time, ch chan weave.Stitch) {
	for start.Before(end) {
		start = start.Add(time.Duration(24*year) * time.Hour)
		ev := weave.Stitch{Data: start}
		ev.Data = start
		ch <- ev
	}
	close(ch)
}

func goFlow1(start, end time.Time, ch chan weave.Stitch) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := weave.Stitch{Data: start}
		ev.Data = start
		ch <- ev
	}
	close(ch)
}

func goFlow2(start, end time.Time, ch chan weave.Stitch) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := weave.Stitch{Data: start}
		ev.Data = start
		ch <- ev
	}
	close(ch)
}

func goFlow3(start, end time.Time, ch chan weave.Stitch) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := weave.Stitch{Data: start}
		ev.Data = start
		ch <- ev
	}
	close(ch)
}

// Sort functions required by the stitch generating function for this example.
type cmdStitches []weave.Stitch

// Len returns the length of the Events list.
func (s cmdStitches) Len() int {
	return len(s)
}

// Less returns a boolean response to the question is e[i] less than e[j].
func (s cmdStitches) Less(i, j int) bool {
	//return s[i].Before(s[j])
	t1 := s[i].Data.(time.Time)
	t2 := s[j].Data.(time.Time)
	return t1.Before(t2)
}

// Swap inverses the positions of e[i] and e[j].
func (s cmdStitches) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func generateStitches(num int, start, end time.Time) []weave.Stitch {

	ev := make([]weave.Stitch, num)
	s := start.Unix()
	e := end.Unix()
	rng := e - s

	for i := range ev {
		rnd := rand.Int63()%rng + 1
		date := time.Unix(s+rnd, 0)
		ev[i].Data = date
	}

	if num > 1 {
		sort.Sort(cmdStitches(ev))
	}

	return ev
}

type mySvg struct {
	svg.Image
}

func main() {
	rand.Seed(time.Now().UnixNano())
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatal(err)
	}
	d1 := time.Date(2000, 1, 1, 0, 0, 0, 0, loc)
	d2 := time.Date(2021, 1, 1, 0, 0, 0, 0, loc)
	chans := make([]chan weave.Stitch, 4)
	for i := range chans {
		chans[i] = make(chan weave.Stitch)
	}

	go goFlow(d1, d2, chans[0])
	go goFlow1(d1, d2, chans[1])
	go goFlow2(d1, d2, chans[2])
	go goFlow3(d1, d2, chans[3])

	w := weave.Loom{
		Start:      weave.Stitch{Data: d1},
		End:        weave.Stitch{Data: d2},
		PreWeave:   startSVG(),
		PreStitch:  writeDasa(),
		Stitch:     writeEvent(),
		PostStitch: nil,
		PostWeave:  closeSVG(),
		Before:     lessThan,
		Equal:      equalTo,
		After:      greaterThan,
	}

	d := mySvg{}
	w = w.LoadData(d)
	ev := generateStitches(10, d1, d2)
	w.Weave(ev, chans[0], chans[1], chans[2], chans[3])
}
