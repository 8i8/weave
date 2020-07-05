package main

// func TestGenerateEvents(t *testing.T) {

// 	var fn weave.CompFuncs
// 	fn = append(fn, before)
// 	fn = append(fn, equal)
// 	fn = append(fn, after)

// 	loc, err := time.LoadLocation("Europe/London")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	d1 := time.Date(2000, 1, 1, 0, 0, 0, 0, loc)
// 	d2 := time.Date(2021, 1, 1, 0, 0, 0, 0, loc)
// 	min := weave.Stitch{Data: d2}
// 	max := weave.Stitch{Data: d1}
// 	c1 := weave.Stitch{Data: d1}
// 	c2 := weave.Stitch{Data: d2}
// 	for i := 0; i < 100; i++ {
// 		ev := generateEvents(fn, 5, d1, d2)
// 		for _, e := range ev {
// 			if e.Before(min) {
// 				min = e
// 			}
// 			if e.After(max) {
// 				max = e
// 			}
// 		}
// 		for _, d := range ev {
// 			if d.Before(c1) || d.After(c2) {
// 				t.Error("error: generateDate: expected date between", d1, "and", d2, "received", d)
// 				break
// 			}
// 		}
// 	}

// 	before := d1.Add(time.Duration(24 * 365 * time.Hour))
// 	after := d2.Add(time.Duration(-24*365) * time.Hour)
// 	c1.Data = before
// 	c2.Data = after
// 	if !min.Before(c1) {
// 		t.Error("warning: generateEvents: no events were generated that were within",
// 			"one year of the earliest date, expected earlier than", before, "recieved", min)
// 	}

// 	if !max.After(c2) {
// 		t.Error("warning: generateEvents: no events were generated that were within",
// 			"one year of the latest date, expected later than", after, "recieved", max)
// 	}
// }
