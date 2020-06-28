package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type events []event

type event struct {
	time.Time
	valid bool
}

func generateEvents(num int, start, end time.Time) events {

	ev := make(events, num)
	s := start.Unix()
	e := end.Unix()
	rng := e - s
	fmt.Println("unix:", time.Unix(rng, 0))

	for i := range ev {
		rng = rand.Int63()%rng + 1
		date := time.Unix(s+rng, 0)
		ev[i].Time = date
		ev[i].valid = true
	}

	if num > 1 {
		sort.Sort(events(ev))
	}

	return ev
}

func (e events) AdvanceTo(t time.Time) int {
	for i := range e {
		if e[i].After(t) || e[i].Equal(t) {
			return i
		}
	}
	return 0
}

func (e events) Debug() {
	for i := range e {
		fmt.Println("event:", e[i].Time)
	}
}

// Sort functions.

// Len returns the length of the events list.
func (e events) Len() int {
	return len(e)
}

// Less retuns a boolean response to the question is e[i] less than e[j].
func (e events) Less(i, j int) bool {
	return e[i].Before(e[j].Time)
}

// Swap inverses the positions of e[i] and e[j].
func (e events) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
