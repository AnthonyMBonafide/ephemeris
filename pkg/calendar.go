package ephemeris

import (
	"fmt"
	"slices"
	"time"
)

// Calendar contains a collection of Events which provides a way of organizing information with respect to time.
//
// A Calendar contains a set of Rules which will allow users to request Events within a given timeframe.
// When a Calendar provides a view it will expand all Events by repeating them as configured, then removing
// any overlaps. This is described as condensing. This type of Calendar only allows one event to be present
// at one time. When there are multiple Events at one time the one that has a higher priority will be used
// and others which are conflicting will be removed. This is meant to mimic a persons caledar where a person
// can only be doing one thing at a time. Similar to Google calendar and other digital calendars where Events
// can be created, repeated, cancelled, ect. However only one Event is displayed(with other conflicts being shown under
// the one which takes priority)
type Calendar struct {
	Name string
	// Perserve order here, the later events take precidence. Like a calendar the later events were made with the ealier ones in mind
	Entries []Rule
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

// Event represents an entry on a calendar which can represent the state of something.
type Event struct {
	Start time.Time
	End   time.Time
	Name  string
}

// Rule additional information about an Event which provides functionality for
// repeating, skipping, or canceling Events. Rules should be persisted so that
// Events can always be derived for a given time window.
//
// 0 values for Repeat, Skip, or Canceled result in that feature not being used.
type Rule struct {
	Event

	// RepeatDuration the duration at which to repeat the event from the Start time.
	RepeatDuration time.Duration

	// RepeatDateAnually will repeat an event every x number of years taking leap
	// years into consideration and ensuring that the date falls on the same month,
	// day of the month, and time each event
	RepeatDateAnually int

	// RepeatWeekly will repeat an Event every x week(s) which will ensure that the
	// event has the same day of the week, start and end times of the day.
	RepeatWeekly int

	// RepeatDayOfMonthMonthly will repeat an Event every x months on the same
	// day of the month. each month. The day of the week may differ.
	//
	// NOTE Anything after the 28th will result in odd behavior as the underlying
	// calendar system will roll over into the following month. For example 31st
	// of June will translate to July 1st
	RepeatDayOfMonthMonthly int

	// RepeatDaily will repeat an Event every x number of days. This will result in
	// events with the same Start and End time.
	RepeatDaily int

	// RepeatForwardUntil the time at which the event should last be repeated
	// when repeating for future events(after the original Event.Start).
	// If an Event's Start time is equal to this then the Event will be valid
	RepeatForwardUntil time.Time

	// RepeatBackwardUntil the time at which the event should last be repeated
	// when repeating for past events(after the original Event.Start).
	// If an Event's Start time is equal to this then the Event will be valid
	RepeatBackwardUntil time.Time

	// Skip contains a list of times where the Event will not be repeated.
	// If the time is within the Start and End times of the Event it will be skipped.
	Skip []time.Time
	// Canceled contains a list of times where the Event will be repeated but
	// marked as cancled. If the time is within the Start and End times of the
	// Event it will be skipped.
	Canceled []time.Time
}

// Expand creates events based on the original event by applying the repeating pattern.
func (r Rule) Expand(viewStart, viewEnd time.Time) []Event {
	var expandedEvents []Event
	// go backwards to the viewstart
	evaluatingEndTime := r.End
	for evaluatingStartTime := r.Start; evaluatingStartTime.After(viewStart); evaluatingStartTime = evaluatingStartTime.Add(-r.RepeatDuration) {
		evaluatingEndTime = evaluatingEndTime.Add(-r.RepeatDuration)
		expandedEvents = append(expandedEvents, Event{
			Start: evaluatingStartTime,
			End:   evaluatingEndTime,
			Name:  r.Name,
		})
	}

	// Go forwards to the viewEnd
	evaluatingEndTime = r.End
	for evaluatingStartTime := r.Start; evaluatingStartTime.Before(viewEnd); evaluatingStartTime = evaluatingStartTime.Add(r.RepeatDuration) {
		evaluatingEndTime = evaluatingEndTime.Add(r.RepeatDuration)
		expandedEvents = append(expandedEvents, Event{
			Start: evaluatingStartTime,
			End:   evaluatingEndTime,
			Name:  r.Name,
		})
	}

	return expandedEvents
}

// View returns Events that are within the Calendar for the given timeframe.
// The Rules will be applied to expand repeating Events as well as skipping,
// canceling, etc.
func (c *Calendar) View(viewStart, viewEnd time.Time) ([]Event, error) {
	// 1. Get events that apply to the time view we are interested in
	// 1a. expand events(recurring events are expanded to specific events within a time span)
	// 2. order them(may need to think about adding priority here or ensuring we perserve order)
	// 3. Start with the beginning of the day and generate a consolidated report

	var results []Event
	for _, rule := range c.Entries {
		expandedEvents := rule.Expand(viewStart, viewEnd)
		for _, expandedEvent := range expandedEvents {
			if (expandedEvent.Start.After(viewStart) || expandedEvent.Start.Equal(viewStart)) || (expandedEvent.End.Before(viewEnd) || expandedEvent.End.Equal(viewEnd)) {
				// TODO: Filter out events that are not active in this view
				continue
			}
		}

		results = append(results, expandedEvents...)
	}

	// Remove overlaps favoring later events
	results, err := ReduceAllEvents(results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ReduceAllEvents like reduceEvents but operates on a any number of Events
func ReduceAllEvents(events []Event) ([]Event, error) {
	if len(events) < 2 {
		// 0 or 1 events cannot have any overlaps
		return events, nil
	}

	processedEvents := events

	i := 0
	j := 1
	for j != len(processedEvents) {

		// Ensure the indexes loop and end correctly.
		// We can just increment i for simplicity and it will be adjusted here
		if i == len(processedEvents) {
			j++
			i = 0
		}
		if i == len(processedEvents)-2 && j == len(processedEvents) {
			// We are done
			break
		}

		if !isOverlap(processedEvents[i], processedEvents[j]) {
			i++
			continue
		}

		updatedEvents1, updatedEvents2 := reduceEvents(processedEvents[i], processedEvents[j])

		// Replace original Events with updated versions and rerun processing
		processedEvents = slices.Replace(processedEvents, i, i+1, updatedEvents1...)
		processedEvents = slices.Replace(processedEvents, j+len(updatedEvents1), j+len(updatedEvents1)+1, updatedEvents2...)

		i = 0
		j = 1
		continue
	}

	return processedEvents, nil
}

// reduceEvents takes 2 events that may or may not overlap and reutrns a list
// of events with no overlaps favoring the later events. This results in two
// slices of Events. The first are derived Events from the first Event
// parameter(e1) and the second slice contains Events derived from the second
// Event paramenter(e2)
//
// The function reduces a group of events so that the resulting Events only
// have one event at any given point in time. Events that are later in the
// group are given precendence over earlier ones with the idea that later
// events were created with the previous in mind.
func reduceEvents(e1 Event, e2 Event) ([]Event, []Event) {
	if !isOverlap(e1, e2) {
		return []Event{e1}, []Event{e2}
	}

	// Same time span
	if e1.Start.Equal(e2.Start) && e1.End.Equal(e2.End) {
		return []Event{}, []Event{e2}
	}

	// Same Start different end
	// |-----e2-----|
	// |------e1---------|
	if e1.Start.Equal(e2.Start) && e1.End.After(e2.End) {
		e1.Start = e2.End
		return []Event{e1}, []Event{e2}
	}

	// Same Start different end
	// |-------e2----------|
	// |------e1-------|
	if e1.Start.Equal(e2.Start) && e1.End.Before(e2.End) {
		e1.Start = e2.End
		return []Event{}, []Event{e2}
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
		return []Event{e1p1, e1p2}, []Event{e2} // Keep e2 later so it retains its priority over e1
	}

	// e1 is within e2
	// Higher priority up top
	// |--------------e2-------------|
	//        |--e1---|
	//
	// Result
	// |--------------e2-------------|
	if e2.Start.Before(e1.Start) && e2.End.After(e1.End) {
		return []Event{}, []Event{e2}
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
		return []Event{e1p1}, []Event{e2}
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
		return []Event{e1p1}, []Event{e2}
	}

	panic(fmt.Sprintf("missed something here: %+v, %+v", e1, e2))
}

// isOverlap determines if the specified Events have any point in time where both are "active".
func isOverlap(e1 Event, e2 Event) bool {
	// No overlap
	if e1.Start.Before(e2.Start) && e1.End.Before(e2.Start) || e2.Start.Before(e1.Start) && e2.End.Before(e1.Start) {
		return false
	}

	// no overlap matching start and end times
	// Higher priority up top
	//               |---------e2------|
	// |------e1-----|
	// Result
	// |-------e1----|--------e2-------|
	if e1.Start.Before(e2.Start) && e2.Start.Equal(e1.End) && e2.End.After(e1.End) {
		return false
	}

	return true
}

// RepeatEventAnnually repeats an Event both forward and backward in time,
// creating multiple Events that fall within the specified window of start and end.
//
// The Event will be repeated every numberOfYears years, either forward or backward.
// This allows for events to be created at regular intervals before or after the original event's timestamp.
func RepeatEventAnnually(e Event, numberOfYears int, start, end time.Time) []Event {
	var repeatedEvents []Event

	// Create a copy of the original event for each direction (forward/backward)
	forwardDirection := e.Start.AddDate(0, 0, numberOfYears)
	backwardDirection := e.Start.AddDate(0, 0, -numberOfYears)

	repeatedEvents = append(repeatedEvents, Event{
		Start: e.Start,
		End:   e.End,
		Name:  e.Name,
	})

	// Repeat events forward in time
	for !forwardDirection.After(e.End) && !forwardDirection.Before(start) {

		repeatedEvent := Event{
			Start: forwardDirection,
			End:   forwardDirection.AddDate(0, 0, numberOfYears),
			Name:  e.Name + " (forward)",
		}

		repeatedEvents = append(repeatedEvents, repeatedEvent)

		forwardDirection = forwardDirection.AddDate(0, 0, numberOfYears)
	}

	// Repeat events backward in time
	for !backwardDirection.Before(e.Start) && !backwardDirection.After(e.End) {

		repeatedEvent := Event{
			Start: backwardDirection,
			End:   backwardDirection.AddDate(0, 0, -numberOfYears),
			Name:  e.Name + " (backward)",
		}

		repeatedEvents = append(repeatedEvents, repeatedEvent)

		backwardDirection = backwardDirection.AddDate(0, 0, -numberOfYears)
	}

	return repeatedEvents
}
