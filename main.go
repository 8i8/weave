package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"scratch/datastructures/funnel/flow"
	"scratch/datastructures/funnel/svg"
	"sort"
	"time"
)

const year = 365

func startSVG() flow.WeaverFunc {
	return func(w flow.Weaver, e flow.Event) flow.Weaver {
		d := w.Data.(data)
		fmt.Println(d.inside)
		d.Image = d.Image.ViewBox(os.Stdout, 0, 0, 100, 100)
		w.Data = d
		return w
	}
}

func writeDasa() flow.WeaverFunc {
	return func(w flow.Weaver, e flow.Event) flow.Weaver {
		fmt.Println("time grid:", e.Time)
		for _, e := range w.Output {
			fmt.Println("reciever:", e)
		}
		return w
	}
}

func writeEvent() flow.WeaverFunc {
	return func(w flow.Weaver, e flow.Event) flow.Weaver {
		if !e.Valid {
			return w
		}
		fmt.Println("event:", e.Time)
		return w
	}
}

func closeSVG() flow.WeaverFunc {
	return func(w flow.Weaver, e flow.Event) flow.Weaver {
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
	chans := make([]chan flow.Event, 4)
	for i := range chans {
		chans[i] = make(chan flow.Event)
	}
	go goFlow(d1, d2, chans[0])
	go goFlow1(d1, d2, chans[1])
	go goFlow2(d1, d2, chans[2])
	go goFlow3(d1, d2, chans[3])

	ev := generateEvents(10, d1, d2)

	w := flow.Weaver{
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

func goFlow(start, end time.Time, ch chan flow.Event) {
	for start.Before(end) {
		start = start.Add(time.Duration(24*year) * time.Hour)
		ev := flow.Event{Time: start, Valid: true}
		ch <- ev
	}
	close(ch)
}

func goFlow1(start, end time.Time, ch chan flow.Event) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := flow.Event{Time: start, Valid: true}
		ch <- ev
	}
	close(ch)
}

func goFlow2(start, end time.Time, ch chan flow.Event) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := flow.Event{Time: start, Valid: true}
		ch <- ev
	}
	close(ch)
}

func goFlow3(start, end time.Time, ch chan flow.Event) {
	start = start.Add(time.Duration(-(rand.Int63()%50)*24) * time.Hour)
	for start.Before(end) {
		start = start.Add(time.Duration(rand.Int63()%year*24) * time.Hour)
		ev := flow.Event{Time: start, Valid: true}
		ch <- ev
	}
	close(ch)
}

func generateEvents(num int, start, end time.Time) flow.Events {

	ev := make(flow.Events, num)
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
		sort.Sort(flow.Events(ev))
	}

	return ev
}
