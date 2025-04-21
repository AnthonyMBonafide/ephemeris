package ephemeris

import (
	"reflect"
	"slices"
	"testing"
	"testing/quick"
	"time"
)

func TestExpandEvents(t *testing.T) {
	testCases := []struct {
		desc             string
		rule             Rule
		viewStart        time.Time
		viewEnd          time.Time
		verificationFunc func(*testing.T, []Event)
	}{
		{
			desc: "Every 365 days Duration",
			rule: Rule{
				Event: Event{
					Start: time.Date(2020, time.February, 13, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2020, time.February, 14, 0, 0, 0, 0, time.UTC),
					Name:  "Dominico's Birthday",
				},
				RepeatDuration:      24 * time.Hour * 365,
				RepeatForwardUntil:  time.Date(2025, 2, 13, 0, 0, 0, 0, time.UTC),
				RepeatBackwardUntil: time.Time{},
				Skip:                []time.Time{},
				Canceled:            []time.Time{},
			},
			viewStart: time.Date(2020, 2, 13, 0, 0, 0, 0, time.UTC),
			viewEnd:   time.Date(2025, 2, 13, 0, 0, 0, 0, time.UTC),
			verificationFunc: func(t *testing.T, e []Event) {
				if len(e) != 5 {
					t.Logf("Expected 5 events but got %d", len(e))
					t.Fail()
				}

				for i, evnt := range e {
					if evnt.Start.Year() != 2020+i {
						t.Logf("Expected start year to be %d but got %d", 2020+i, evnt.Start.Year())
						t.Fail()
					}
					if evnt.Start.Month() != time.February {
						t.Logf("Expected start month to be %s but got %d", evnt.Start.Month(), evnt.Start.Year())
						t.Fail()
					}
					if evnt.Start.Day() != 13 {
						t.Logf("Expected start day of the month to be %d but got %d", 13, evnt.Start.Day())
						t.Fail()

					}
					if evnt.End.Day() != 14 {
						t.Logf("Expected end day of month to be %d but got %d", 14, evnt.End.Day())
						t.Fail()

					}
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := tC.rule.Expand(tC.viewStart, tC.viewEnd)
			tC.verificationFunc(t, got)
		})
	}
}

// Fixed time that can be used to ensure that fractional seconds are not off causing inconsistent test results
var rightNow = time.Now().Truncate(time.Millisecond)

func TestSquashEvents(t *testing.T) {
	testCases := []struct {
		desc           string
		e1             Event
		e2             Event
		expectedResult func(Event, Event) ([]Event, []Event)
	}{
		{
			desc: "Matching Events",
			e1:   Event{Name: "one", Start: rightNow, End: rightNow.AddDate(0, 0, 7)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 7)},
			expectedResult: func(e1, e2 Event) ([]Event, []Event) {
				return []Event{}, []Event{e2}
			},
		},
		{
			desc: "Same Start Different End",
			e1:   Event{Name: "one", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 7)},
			expectedResult: func(e1, e2 Event) ([]Event, []Event) {
				e1.Start = rightNow.AddDate(0, 0, 7)
				return []Event{e1}, []Event{e2}
			},
		},
		{
			desc: "e2 overwrite",
			e1:   Event{Name: "one", Start: rightNow, End: rightNow.AddDate(0, 0, 7)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			expectedResult: func(e1, e2 Event) ([]Event, []Event) {
				return []Event{}, []Event{e2}
			},
		},
		{
			desc: "No Overlap",
			e1:   Event{Name: "one", Start: rightNow.AddDate(0, 0, -5), End: rightNow.AddDate(0, 0, -1)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			expectedResult: func(e1, e2 Event) ([]Event, []Event) {
				return []Event{e1}, []Event{e2}
			},
		},
		{
			desc: "No Overlap Matching Start and End Times",
			e1:   Event{Name: "one", Start: rightNow.AddDate(0, 0, -5), End: rightNow},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			expectedResult: func(e1, e2 Event) ([]Event, []Event) {
				e1.End = e2.Start
				return []Event{e1}, []Event{e2}
			},
		},
		{
			desc: "Middle Overlap",
			e1:   Event{Name: "one", Start: rightNow.AddDate(0, 0, -5), End: rightNow.AddDate(0, 0, 2)},
			e2:   Event{Name: "two", Start: rightNow, End: rightNow.AddDate(0, 0, 8)},
			expectedResult: func(e1, e2 Event) ([]Event, []Event) {
				e1.End = e2.Start
				return []Event{e1}, []Event{e2}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got1, got2 := reduceEvents(tC.e1, tC.e2)
			expected1, expected2 := tC.expectedResult(tC.e1, tC.e2)
			if !slices.Equal(got1, expected1) {
				t.Log("expected event 1 times to match")
				t.Fail()
			}
			if !slices.Equal(got2, expected2) {
				t.Log("expected event 2 times to match")
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

		got1, got2 := reduceEvents(e1, e2)

		if len(got1)+len(got2) <= 0 {
			t.Log("expected to have at least one event")
			return false
		}

		if len(got1)+len(got2) > 3 {
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

		for _, gotEvent := range got1 {
			if gotEvent.Start.Before(earliest) || gotEvent.End.After(latest) {
				t.Log("expected original start and end times to not be exceeded")
				return false
			}
		}

		for _, gotEvent := range got2 {
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

func TestReduceAllEvents(t *testing.T) {
	testCases := []struct {
		desc     string
		events   []Event
		expected func([]Event) []Event
	}{
		{
			desc:   "Empty Events",
			events: []Event{},
			expected: func(e []Event) []Event {
				return e
			},
		},
		{
			desc:   "Single Event",
			events: []Event{{Start: rightNow, End: rightNow.AddDate(0, 0, 1)}},
			expected: func(e []Event) []Event {
				return e
			},
		},
		{
			desc: "Multiple Non-overlapping Events",
			events: []Event{
				{Start: rightNow, End: rightNow.AddDate(0, 0, 1)},
				{Start: rightNow.AddDate(0, 0, 2), End: rightNow.AddDate(0, 0, 3)},
				{Start: rightNow.AddDate(0, 0, 4), End: rightNow.AddDate(0, 0, 5)},
			},
			expected: func(e []Event) []Event {
				return e
			},
		},
		{
			desc: "Multiple Non-overlapping Events Same Start and End Times",
			events: []Event{
				{Start: rightNow, End: rightNow.AddDate(0, 0, 1)},
				{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 2)},
				{Start: rightNow.AddDate(0, 0, 2), End: rightNow.AddDate(0, 0, 3)},
			},
			expected: func(e []Event) []Event {
				return e
			},
		},
		{
			desc: "Simple Overlap",
			events: []Event{
				{Start: rightNow, End: rightNow.AddDate(0, 0, 2)},
				{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 3)},
			},
			expected: func(e []Event) []Event {
				return []Event{
					{Start: rightNow, End: rightNow.AddDate(0, 0, 1)},
					{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 3)},
				}
			},
		},
		{
			desc: "Multiple Simple Overlaps",
			events: []Event{
				{Start: rightNow, End: rightNow.AddDate(0, 0, 2)},
				{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 3)},
				{Start: rightNow.AddDate(0, 0, 2), End: rightNow.AddDate(0, 0, 4)},
			},
			expected: func(e []Event) []Event {
				return []Event{
					{Start: rightNow, End: rightNow.AddDate(0, 0, 1)},
					{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 2)},
					{Start: rightNow.AddDate(0, 0, 2), End: rightNow.AddDate(0, 0, 4)},
				}
			},
		},
		{
			desc: "One Overlap Replacing Many",
			events: []Event{
				{Start: rightNow, End: rightNow.AddDate(0, 0, 5)},
				{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 2)},
				{Start: rightNow.AddDate(0, 0, 2), End: rightNow.AddDate(0, 0, 4)},
			},
			expected: func(e []Event) []Event {
				return []Event{
					{Start: rightNow, End: rightNow.AddDate(0, 0, 1)},
					{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 2)},
					{Start: rightNow.AddDate(0, 0, 2), End: rightNow.AddDate(0, 0, 4)},
				}
			},
		},
		{
			desc: "One Overlap Replacing Many Nested Overlaps",
			events: []Event{
				{Start: rightNow, End: rightNow.AddDate(0, 0, 5)},
				{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 3)},
				{Start: rightNow.AddDate(0, 0, 2), End: rightNow.AddDate(0, 0, 4)},
			},
			expected: func(e []Event) []Event {
				return []Event{
					{Start: rightNow, End: rightNow.AddDate(0, 0, 1)},
					{Start: rightNow.AddDate(0, 0, 1), End: rightNow.AddDate(0, 0, 2)},
					{Start: rightNow.AddDate(0, 0, 2), End: rightNow.AddDate(0, 0, 4)},
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, _ := ReduceAllEvents(tC.events)
			expected := tC.expected(tC.events)
			if !slices.Equal(expected, got) {
				t.Fail()
			}
		})
	}
}

func TestRepeatEventAnnually(t *testing.T) {
	tests := []struct {
		name          string
		event         Event
		numberOfYears int
		start         time.Time
		end           time.Time
		want          []Event
	}{
		{
			name:          "single event",
			event:         Event{Start: time.Now(), End: time.Now(), Name: "Test Event"},
			numberOfYears: 1,
			start:         time.Now(),
			end:           time.Now().AddDate(0, 0, 5),
			want: []Event{
				{Start: time.Now(), End: time.Now(), Name: "Test Event"},
			},
		},
		{
			name:          "forward events",
			event:         Event{Start: time.Now(), End: time.Now(), Name: "Test Event"},
			numberOfYears: 2,
			start:         time.Now(),
			end:           time.Now().AddDate(0, 0, 5),
			want: []Event{
				{Start: time.Now(), End: time.Now(), Name: "Test Event"},
				{Start: time.Now().AddDate(0, 0, 2), End: time.Now().AddDate(0, 0, 4), Name: "Test Event (forward)"},
			},
		},
		{
			name:          "backward events",
			event:         Event{Start: time.Now(), End: time.Now(), Name: "Test Event"},
			numberOfYears: -2,
			start:         time.Now(),
			end:           time.Now().AddDate(0, 0, 5),
			want: []Event{
				{Start: time.Now(), End: time.Now(), Name: "Test Event"},
				{Start: time.Now().AddDate(0, 0, -2), End: time.Now().AddDate(0, 0, 0), Name: "Test Event (backward)"},
			},
		},
		{
			name:          "multiple forward events",
			event:         Event{Start: time.Now(), End: time.Now(), Name: "Test Event"},
			numberOfYears: 5,
			start:         time.Now(),
			end:           time.Now().AddDate(0, 0, 10),
			want: []Event{
				{Start: time.Now(), End: time.Now(), Name: "Test Event"},
				{Start: time.Now().AddDate(0, 0, 5), End: time.Now().AddDate(0, 0, 9), Name: "Test Event (forward)"},
				{Start: time.Now().AddDate(0, 0, 10), End: time.Now().AddDate(0, 0, 14), Name: "Test Event (forward)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RepeatEventAnnually(tt.event, tt.numberOfYears, tt.start, tt.end)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RepeatEventAnnually() = %v, want %v", got, tt.want)
			}
		})
	}
}
