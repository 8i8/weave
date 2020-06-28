package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Start function.
func fn1(in ...interface{}) {
	w := in[0].(weaver)
	fmt.Println("time grid:", w.current)
	for i, r := range w.r {
		if r.current.Before(w.next.Time) {
			fmt.Println("receiver:", i, r.current)
		}
	}
}

// Pre event.
func fn2(in ...interface{}) {
}

// Event function.
func fn3(in ...interface{}) {
	ev := in[0].(event)
	if !ev.valid {
		return
	}
	fmt.Println("event:", ev.Time)
}

// Post algorithem.
func fn4(in ...interface{}) {
}

func main() {
	rand.Seed(time.Now().UnixNano())
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatal(err)
	}
	d1 := time.Date(2000, 1, 1, 0, 0, 0, 0, loc)
	d2 := time.Date(2021, 1, 1, 0, 0, 0, 0, loc)
	chans := make([]chan event, 4)
	for i := range chans {
		chans[i] = make(chan event)
	}
	go flow(d1, d2, chans[0])
	go flow1(d1, d2, chans[1])
	go flow2(d1, d2, chans[2])
	go flow3(d1, d2, chans[3])

	ev := generateEvents(5, d1, d2)

	Weave(ev, d1, d2, chans[0], chans[1], chans[2], chans[3])
}
