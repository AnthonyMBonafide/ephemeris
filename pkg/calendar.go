package ephemeris

import (
	"time"
)

type Calendar struct {
	Name string
	// Perserve order here, the later events take precidence. Like a calendar the later events were made with the ealier ones in mind
	Events []Event
}

func (c Calendar) String() string {
	panic("TODO implement list of events")
}

func (c Calendar) StringForView(viewStart, viewEnd time.Time) {
	panic("TODO implement list of events for given timeframe")
}

func (c Calendar) AsciiForView(viewStart, viewEnd time.Time) {
	panic("TODO implement text user interface for viewing of events for the given timeframe in terminal/text")
}

// TODO: Rename to Rule and create another structure called event which has the
// same data minus the Repeat. Also, add information to link Rule to an Event.
// The Rule can be stored at the Calendar level and Events are derived from
// Rules which are single items on a calendar, each of which can be managed
// independently(canceled, skipped, etc(this will have to be managed at the
// Rule level so that we can always derive the skip/cancel when creating the
// events from Rules))
type Event struct {
	Start  time.Time
	End    time.Time
	Name   string
	Repeat time.Duration
}

// Overlaps determines if two events overlap in the "active" window
//
// Can be used for testing to ensure squashing behaviour is correct as the result should be no overlaps
func (e Event) Overlaps(other Event, startView, endView time.Time) bool {
	// Expand the event out from its original start by doing repeats
	panic("TODO implement")
}

// Expand creates events based on the original event by applying the repeating pattern.
func (e Event) Expand(viewStart, viewEnd time.Time) []Event {
	panic("TODO implement")
}

// 1. Get events that apply to the time view we are interested in
// 1a. expand events(recurring events are expanded to specific events within a time span)
// 2. order them(may need to think about adding priority here or ensuring we perserve order)
// 3. Start with the beginning of the day and generate a consolidated report
func (c *Calendar) View(viewStart, viewEnd time.Time) []Event {
	var results []Event
	for _, evnt := range c.Events {
		expandedEvents := evnt.Expand(viewStart, viewEnd)
		for _, expandedEvent := range expandedEvents {
			if (expandedEvent.Start.After(viewStart) || expandedEvent.Start.Equal(viewStart)) || (expandedEvent.End.Before(viewEnd) || expandedEvent.End.Equal(viewEnd)) {
				// Filter out events that are not active in this view
				continue
			}
		}

		results = append(results, evnt)
	}

	// Remove overlaps favoring later events
	results = CondenseOverlaps(results)

	return results
}

// removes all overlaps from a list of events
func CondenseOverlaps(events []Event) []Event {
	var result []Event
	// TODO: Highly ineffecient but will work, need to do this better
	//
	// Go through all permutations of events
	for i := 0; i < len(events); i++ {
		for j := i + 1; i < len(events); j++ {
			if i == j {
				continue
			}

			evs := SquashEvents(events[i], events[j])
			// TODO: Fix ordering to preserve priority
			result = append(result, evs...)

		}
	}

	return result
}

// Takes 2 events that may or may not overlap and reutrns a list of events with no overlaps favoring the later event
func SquashEvents(e1 Event, e2 Event) []Event { // TODO: Make this return 2 slices of events, the first are the events from e1 and the second from e2 this will allow callers to preserve order when comparing bigger amounts of events
	// No overlap
	if e1.Start.Before(e2.Start) && e1.End.Before(e2.Start) || e2.Start.Before(e1.Start) && e2.End.Before(e1.Start) {
		return []Event{e1, e2}
	}

	// no overlap matching start and end times
	// Higher priority up top
	//               |---------e2------|
	// |------e1-----|
	// Result
	// |-------e1----|--------e2-------|
	if e1.Start.Before(e2.Start) && e2.Start.Equal(e1.End) && e2.End.After(e1.End) {
		return []Event{e1, e2}
	}

	// Same time span
	if e1.Start.Equal(e2.Start) && e1.End.Equal(e2.End) {
		return []Event{e2}
	}

	// Same Start different end
	// |-----e2-----|
	// |------e1---------|
	if e1.Start.Equal(e2.Start) && e1.End.After(e2.End) {
		e1.Start = e2.End
		return []Event{e1, e2}
	}

	// Same Start different end
	// |-------e2----------|
	// |------e1-------|
	if e1.Start.Equal(e2.Start) && e1.End.Before(e2.End) {
		e1.Start = e2.End
		return []Event{e2}
	}

	// e2 is within e1
	// // Higher priority up top
	//        |--e2---|
	// |--------------e1-------------|
	//
	// Result
	// |--e1--|--e2---|-----e1-------|
	//

	if e1.Start.Before(e2.Start) && e1.End.After(e2.End) {
		e1p1 := e1
		e1p1.End = e2.Start
		e1p2 := e1
		e1p2.Start = e2.End
		e1p2.End = e1.End
		return []Event{e1p1, e1p2, e2} // Keep e2 later so it retains its priority over e1
	}

	// e1 is within e2
	// Higher priority up top
	// |--------------e2-------------|
	//        |--e1---|
	//
	// Result
	// |--------------e2-------------|
	if e2.Start.Before(e1.Start) && e2.End.After(e1.End) {
		return []Event{e2}
	}

	// middle overlap
	// Higher priority up top
	//         |---------e2------|
	// |------e1-----|
	// Result
	// |---e1--|--------e2-------|
	if e1.Start.Before(e2.Start) && e2.Start.Before(e1.End) && e2.End.After(e1.End) {
		e1p1 := e1
		e1p1.End = e2.Start
		return []Event{e1p1, e2}
	}

	// middle overlap
	// Higher priority up top
	// |------e2-----|
	//         |---------e1------|
	// Result
	// |---e2--------|--e1-------|
	if e2.Start.Before(e1.Start) && e1.Start.Before(e2.End) && e1.End.After(e2.End) {
		e1p1 := e1
		e1p1.Start = e2.End
		return []Event{e1p1, e2}
	}

	// TODO: Repeat the same situations as above, but where the Start and End times are equal
	panic("missed something here")
}
