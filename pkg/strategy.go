package calendar

import (
	"sort"
	"time"
)

// PriorityBasedStrategy resolves overlaps by giving precedence to higher priority events
// and splitting lower priority events around higher priority ones
type PriorityBasedStrategy struct{}

// ResolveOverlaps implements the OverlapStrategy interface
func (s *PriorityBasedStrategy) ResolveOverlaps(events []Event) []Event {
	if len(events) <= 1 {
		return events
	}

	// Sort events by start time first, then by priority (higher first)
	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].Priority > events[j].Priority
		}
		return events[i].StartTime.Before(events[j].StartTime)
	})

	result := make([]Event, 0)

	for _, event := range events {
		// Check for overlaps with already processed events
		conflicts := s.findConflicts(event, result)

		if len(conflicts) == 0 {
			// No conflicts, add the event as-is
			result = append(result, event)
		} else {
			// Resolve conflicts by splitting the current event
			splitEvents := s.splitEventAroundConflicts(event, conflicts)
			result = append(result, splitEvents...)
		}
	}

	// Sort final result by start time
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.Before(result[j].StartTime)
	})

	return result
}

// findConflicts finds all events in the result that conflict with the given event
func (s *PriorityBasedStrategy) findConflicts(event Event, existing []Event) []Event {
	conflicts := make([]Event, 0)

	for _, existingEvent := range existing {
		if s.eventsOverlap(event, existingEvent) {
			// Only consider it a conflict if the existing event has higher or equal priority
			if existingEvent.Priority >= event.Priority {
				conflicts = append(conflicts, existingEvent)
			}
		}
	}

	return conflicts
}

// eventsOverlap checks if two events overlap in time
func (s *PriorityBasedStrategy) eventsOverlap(a, b Event) bool {
	return a.StartTime.Before(b.EndTime) && a.EndTime.After(b.StartTime)
}

// splitEventAroundConflicts splits an event around conflicting higher-priority events
func (s *PriorityBasedStrategy) splitEventAroundConflicts(event Event, conflicts []Event) []Event {
	// Sort conflicts by start time
	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].StartTime.Before(conflicts[j].StartTime)
	})

	result := make([]Event, 0)
	currentStart := event.StartTime

	for _, conflict := range conflicts {
		// If there's a gap before this conflict, create an event for that gap
		if currentStart.Before(conflict.StartTime) {
			gapEnd := conflict.StartTime
			if gapEnd.After(event.EndTime) {
				gapEnd = event.EndTime
			}

			if currentStart.Before(gapEnd) {
				splitEvent := Event{
					ID:        event.ID + "_split_" + currentStart.Format("1504"),
					Title:     event.Title,
					StartTime: currentStart,
					EndTime:   gapEnd,
					Priority:  event.Priority,
					RuleID:    event.RuleID,
				}
				result = append(result, splitEvent)
			}
		}

		// Move current start to after this conflict
		if conflict.EndTime.After(currentStart) {
			currentStart = conflict.EndTime
		}

		// If we've moved past the original event's end time, we're done
		if currentStart.After(event.EndTime) || !currentStart.Before(event.EndTime) {
			break
		}
	}

	// If there's time remaining after all conflicts, create an event for that
	if currentStart.Before(event.EndTime) {
		splitEvent := Event{
			ID:        event.ID + "_split_" + currentStart.Format("1504"),
			Title:     event.Title,
			StartTime: currentStart,
			EndTime:   event.EndTime,
			Priority:  event.Priority,
			RuleID:    event.RuleID,
		}
		result = append(result, splitEvent)
	}

	return result
}

// SimpleStrategy just returns events as-is without resolving overlaps
type SimpleStrategy struct{}

// ResolveOverlaps implements the OverlapStrategy interface
func (s *SimpleStrategy) ResolveOverlaps(events []Event) []Event {
	return events
}
