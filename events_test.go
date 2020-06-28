package main

import (
	"log"
	"testing"
	"time"
)

func TestGenerateEvents(t *testing.T) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatal(err)
	}
	d1 := time.Date(2000, 1, 1, 0, 0, 0, 0, loc)
	d2 := time.Date(2021, 1, 1, 0, 0, 0, 0, loc)
	min, max := d2, d1
	for i := 0; i < 100; i++ {
		ev := generateEvents(5, d1, d2)
		for _, e := range ev {
			if e.Before(min) {
				min = e.Time
			}
			if e.After(max) {
				max = e.Time
			}
		}
		for _, d := range ev {
			if d.Before(d1) || d.After(d2) {
				t.Error("error: generateDate: expected date between", d1, "and", d2, "received", d)
				break
			}
		}
	}

	before := d1.Add(time.Duration(24 * 365 * time.Hour))
	after := d2.Add(time.Duration(-24*365) * time.Hour)
	if !min.Before(before) {
		t.Error("warning: generateEvents: no events were generated that were within",
			"one year of the earliest date, expected earlier than", before, "recieved", min)
	}

	if !max.After(after) {
		t.Error("warning: generateEvents: no events were generated that were within",
			"one year of the latest date, expected later than", after, "recieved", max)
	}
}
