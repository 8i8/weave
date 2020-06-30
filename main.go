package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"scratch/datastructures/weave/svg"
	"scratch/datastructures/weave/weave"
	"sort"
	"time"
)

const year = 365

func startSVG() weave.WeaverFunc {
	return func(w weave.Weaver, e weave.Stitch) weave.Weaver {
		d := w.Data.(data)
		fmt.Println(d.inside)
		d.Image = d.Image.ViewBox(os.Stdout, 0, 0, 100, 100)
		w.Data = d
		return w
	}
}

func writeDasa() weave.WeaverFunc {
	return func(w weave.Weaver, e weave.Stitch) weave.Weaver {
		fmt.Println("time grid:", e.Time)
		for _, m := range w.Output {
			fmt.Println("reciever:", m.Time)
		}
		return w
	}
}

func writeEvent() weave.WeaverFunc {
	return func(w weave.Weaver, e weave.Stitch) weave.Weaver {
		if !e.Valid {
			return w
		}
		fmt.Println("event:", e.Time)
		return w
	}
}

func closeSVG() weave.WeaverFunc {
	return func(w weave.Weaver, e weave.Stitch) weave.Weaver {
		svg := w.Data.(data)
		svg.End()
		return w
	}
}

type data struct {
	inside string
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

	ev := generateEvents(10, d1, d2)

	w := weave.Weaver{
		Start:     d1,
		End:       d2,
		PreFunc:   startSVG(),
		PreEvent:  writeDasa(),
		Event:     writeEvent(),
		PostEvent: nil,
		PostFunc:  closeSVG(),
	}

	d := data{
		inside: "Hello World",
	}
	w = w.Load(d)

	w.Date(ev, chans[0], chans[1], chans[2], chans[3])
}

func goFlow(start, end time.Time, ch chan weave.Stitch) {
	for start.Before(end) {
		start = start.Add(time.Duration(24*year) * time.Hour)
		ev := weave.Stitch{Time: start, Valid: true}
		ch <- ev
	}
	close(ch)
}

func goFlow1(start, end time.Time, ch chan weave.Stitch) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := weave.Stitch{Time: start, Valid: true}
		ch <- ev
	}
	close(ch)
}

func goFlow2(start, end time.Time, ch chan weave.Stitch) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := weave.Stitch{Time: start, Valid: true}
		ch <- ev
	}
	close(ch)
}

func goFlow3(start, end time.Time, ch chan weave.Stitch) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := weave.Stitch{Time: start, Valid: true}
		ch <- ev
	}
	close(ch)
}

func generateEvents(num int, start, end time.Time) weave.Stiches {

	ev := make(weave.Stiches, num)
	s := start.Unix()
	e := end.Unix()
	rng := e - s

	for i := range ev {
		rnd := rand.Int63()%rng + 1
		date := time.Unix(s+rnd, 0)
		ev[i].Time = date
		ev[i].Valid = true
	}

	if num > 1 {
		sort.Sort(weave.Stiches(ev))
	}

	return ev
}
