package ephemeris

import (
	"slices"
	"testing"
	"testing/quick"
	"time"
)

// Fixed time that can be used to ensure that fractional seconds are not off causing inconsistent test results
var rightNow = time.Now().Truncate(time.Millisecond)

func TestSquashEvents(t *testing.T) {
	testCases := []struct {
		desc           string
		e1             Event
		e2             Event
		expectedResult func(Event, Event) []Event
	}{
		{
			desc: "Matching Events",
			e1:   Event{Name: "one", Start: rightNow, End: rightNow.AddDate(0, 0, 7)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 7)},
			expectedResult: func(e1, e2 Event) []Event {
				return []Event{e2}
			},
		},
		{
			desc: "Same Start Different End",
			e1:   Event{Name: "one", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 7)},
			expectedResult: func(e1, e2 Event) []Event {
				e1.Start = rightNow.AddDate(0, 0, 7)
				return []Event{e1, e2}
			},
		},
		{
			desc: "e2 overwrite",
			e1:   Event{Name: "one", Start: rightNow, End: rightNow.AddDate(0, 0, 7)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			expectedResult: func(e1, e2 Event) []Event {
				return []Event{e2}
			},
		},
		{
			desc: "No Overlap",
			e1:   Event{Name: "one", Start: rightNow.AddDate(0, 0, -5), End: rightNow.AddDate(0, 0, -1)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			expectedResult: func(e1, e2 Event) []Event {
				return []Event{e1, e2}
			},
		},
		{
			desc: "No Overlap Matching Start and End Times",
			e1:   Event{Name: "one", Start: rightNow.AddDate(0, 0, -5), End: rightNow},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			expectedResult: func(e1, e2 Event) []Event {
				e1.End = e2.Start
				return []Event{e1, e2}
			},
		},
		{
			desc: "Middle Overlap",
			e1:   Event{Name: "one", Start: rightNow.AddDate(0, 0, -5), End: rightNow.AddDate(0, 0, 2)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			expectedResult: func(e1, e2 Event) []Event {
				e1.End = e2.Start
				return []Event{e1, e2}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := SquashEvents(tC.e1, tC.e2)
			if !slices.Equal(got, tC.expectedResult(tC.e1, tC.e2)) {
				t.Fail()
			}
		})
	}
}

func TestSquashEvents_Property(t *testing.T) {
	f := func(e1s, e1e, e2s, e2e int64) bool {
		t1 := time.UnixMilli(e1s)
		t2 := time.UnixMilli(e1e)
		t3 := time.UnixMilli(e2s)
		t4 := time.UnixMilli(e2e)

		var e1 Event
		var e2 Event
		// Create events with valid time(start is before end)
		if t1.Before(t2) {
			e1 = Event{Start: t1, End: t2}
		} else {
			e1 = Event{Start: t2, End: t1}
		}

		if t3.Before(t4) {
			e2 = Event{Start: t3, End: t4}
		} else {
			e2 = Event{Start: t4, End: t3}
		}

		got := SquashEvents(e1, e2)

		if len(got) <= 0 {
			t.Log("expected to have at least one event")
			return false
		}

		if len(got) > 3 {
			t.Log("expected to have a max of 3 events")
			return false
		}

		earliest := e1.Start
		latest := e1.End

		if e1.Start.After(e2.Start) {
			earliest = e2.Start
		}
		if e1.End.Before(e2.End) {
			latest = e2.End
		}

		for _, gotEvent := range got {
			if gotEvent.Start.Before(earliest) || gotEvent.End.After(latest) {
				t.Log("expected original start and end times to not be exceeded")
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 1_000_000}); err != nil {
		t.Error(err)
	}
}
