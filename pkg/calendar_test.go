package calendar

import (
	"testing"
	"time"
)

func TestNewCalendar(t *testing.T) {
	strategy := &SimpleStrategy{}
	cal := NewCalendar(strategy)
	if cal == nil {
		t.Error("NewCalendar() returned nil")
	}
}

func TestAddEvent(t *testing.T) {
	strategy := &SimpleStrategy{}
	cal := NewCalendar(strategy)
	event := Event{
		ID:        "1",
		Title:     "Test Event",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		Priority:  1,
	}

	cal.AddEvent(event)
	events := cal.GetAllEvents()

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].ID != "1" {
		t.Errorf("Expected event ID '1', got '%s'", events[0].ID)
	}
}

func TestAddRule(t *testing.T) {
	strategy := &SimpleStrategy{}
	cal := NewCalendar(strategy)

	startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	rule := Rule{
		ID:        "daily-meeting",
		Title:     "Daily Standup",
		StartTime: startTime,
		EndTime:   endTime,
		Priority:  2,
		Recurrence: RecurrencePattern{
			Type:     Daily,
			Interval: 1,
		},
		ValidFrom: startTime,
	}

	cal.AddRule(rule)
	rules := cal.GetAllRules()

	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	if rules[0].ID != "daily-meeting" {
		t.Errorf("Expected rule ID 'daily-meeting', got '%s'", rules[0].ID)
	}
}

func TestGetEventsWithRule(t *testing.T) {
	strategy := &SimpleStrategy{}
	cal := NewCalendar(strategy)

	startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	rule := Rule{
		ID:        "daily-meeting",
		Title:     "Daily Standup",
		StartTime: startTime,
		EndTime:   endTime,
		Priority:  2,
		Recurrence: RecurrencePattern{
			Type:     Daily,
			Interval: 1,
		},
		ValidFrom: startTime,
	}

	cal.AddRule(rule)

	// Get events for a 3-day period
	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	if len(events) != 3 {
		t.Errorf("Expected 3 events from daily rule, got %d", len(events))
	}

	// Check that events are on consecutive days
	expectedDates := []time.Time{
		time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 9, 0, 0, 0, time.UTC),
	}

	for i, event := range events {
		if !event.StartTime.Equal(expectedDates[i]) {
			t.Errorf("Event %d: expected start time %v, got %v", i, expectedDates[i], event.StartTime)
		}
	}
}

func TestPriorityBasedOverlapResolution(t *testing.T) {
	strategy := &PriorityBasedStrategy{}
	cal := NewCalendar(strategy)

	// Event A: 8:00 AM - 8:00 PM, Priority 2 (higher)
	eventA := Event{
		ID:        "A",
		Title:     "High Priority Event",
		StartTime: time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		Priority:  2,
	}

	// Event B: 7:00 AM - 9:00 PM, Priority 1 (lower)
	eventB := Event{
		ID:        "B",
		Title:     "Low Priority Event",
		StartTime: time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 21, 0, 0, 0, time.UTC),
		Priority:  1,
	}

	cal.AddEvent(eventA)
	cal.AddEvent(eventB)

	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	// Should have 3 events: B (7-8), A (8-20), B (20-21)
	if len(events) != 3 {
		t.Errorf("Expected 3 events after overlap resolution, got %d", len(events))
		for i, e := range events {
			t.Logf("Event %d: %s from %v to %v", i, e.Title, e.StartTime, e.EndTime)
		}
		return
	}

	// Check first split of B (7:00 AM - 8:00 AM)
	if !events[0].StartTime.Equal(time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC)) {
		t.Errorf("First event should start at 7:00 AM, got %v", events[0].StartTime)
	}
	if !events[0].EndTime.Equal(time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)) {
		t.Errorf("First event should end at 8:00 AM, got %v", events[0].EndTime)
	}

	// Check A (8:00 AM - 8:00 PM)
	if !events[1].StartTime.Equal(time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)) {
		t.Errorf("Second event should start at 8:00 AM, got %v", events[1].StartTime)
	}
	if !events[1].EndTime.Equal(time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)) {
		t.Errorf("Second event should end at 8:00 PM, got %v", events[1].EndTime)
	}

	// Check second split of B (8:00 PM - 9:00 PM)
	if !events[2].StartTime.Equal(time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)) {
		t.Errorf("Third event should start at 8:00 PM, got %v", events[2].StartTime)
	}
	if !events[2].EndTime.Equal(time.Date(2024, 1, 1, 21, 0, 0, 0, time.UTC)) {
		t.Errorf("Third event should end at 9:00 PM, got %v", events[2].EndTime)
	}
}

func TestWeeklyRecurrence(t *testing.T) {
	strategy := &SimpleStrategy{}
	cal := NewCalendar(strategy)

	startTime := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC) // Monday
	endTime := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)

	rule := Rule{
		ID:        "weekly-meeting",
		Title:     "Weekly Team Meeting",
		StartTime: startTime,
		EndTime:   endTime,
		Priority:  1,
		Recurrence: RecurrencePattern{
			Type:     Weekly,
			Interval: 1,
		},
		ValidFrom: startTime,
	}

	cal.AddRule(rule)

	// Get events for a 3-week period
	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	if len(events) != 3 {
		t.Errorf("Expected 3 events from weekly rule, got %d", len(events))
	}

	// Check that events are on consecutive Mondays
	expectedDates := []time.Time{
		time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),  // Jan 1 (Monday)
		time.Date(2024, 1, 8, 14, 0, 0, 0, time.UTC),  // Jan 8 (Monday)
		time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC), // Jan 15 (Monday)
	}

	for i, event := range events {
		if !event.StartTime.Equal(expectedDates[i]) {
			t.Errorf("Event %d: expected start time %v, got %v", i, expectedDates[i], event.StartTime)
		}
	}
}

func TestMonthlyRecurrence(t *testing.T) {
	strategy := &SimpleStrategy{}
	cal := NewCalendar(strategy)

	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	rule := Rule{
		ID:        "monthly-review",
		Title:     "Monthly Review",
		StartTime: startTime,
		EndTime:   endTime,
		Priority:  1,
		Recurrence: RecurrencePattern{
			Type:     Monthly,
			Interval: 1,
		},
		ValidFrom: startTime,
	}

	cal.AddRule(rule)

	// Get events for a 4-month period
	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	if len(events) != 4 {
		t.Errorf("Expected 4 events from monthly rule, got %d", len(events))
	}

	// Check that events are on the 15th of consecutive months
	expectedDates := []time.Time{
		time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 15, 10, 0, 0, 0, time.UTC),
	}

	for i, event := range events {
		if !event.StartTime.Equal(expectedDates[i]) {
			t.Errorf("Event %d: expected start time %v, got %v", i, expectedDates[i], event.StartTime)
		}
	}
}

func TestPriorityBasedStrategy_NoOverlaps(t *testing.T) {
	strategy := &PriorityBasedStrategy{}
	cal := NewCalendar(strategy)

	// Two non-overlapping events
	eventA := Event{
		ID:        "A",
		Title:     "Morning Event",
		StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Priority:  1,
	}

	eventB := Event{
		ID:        "B",
		Title:     "Afternoon Event",
		StartTime: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		Priority:  2,
	}

	cal.AddEvent(eventA)
	cal.AddEvent(eventB)

	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	// Should have both events unchanged
	if len(events) != 2 {
		t.Errorf("Expected 2 events with no overlaps, got %d", len(events))
	}
}

func TestPriorityBasedStrategy_HigherPriorityWins(t *testing.T) {
	strategy := &PriorityBasedStrategy{}
	cal := NewCalendar(strategy)

	// Lower priority event (completely overlapped)
	eventLow := Event{
		ID:        "LOW",
		Title:     "Low Priority Event",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Priority:  1,
	}

	// Higher priority event (completely overlaps the low priority one)
	eventHigh := Event{
		ID:        "HIGH",
		Title:     "High Priority Event",
		StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		Priority:  5,
	}

	cal.AddEvent(eventLow)
	cal.AddEvent(eventHigh)

	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	// Should only have the high priority event
	if len(events) != 1 {
		t.Errorf("Expected 1 event after overlap resolution, got %d", len(events))
	}

	if events[0].ID != "HIGH" {
		t.Errorf("Expected high priority event to remain, got event with ID '%s'", events[0].ID)
	}
}

func TestPriorityBasedStrategy_PartialOverlap(t *testing.T) {
	strategy := &PriorityBasedStrategy{}
	cal := NewCalendar(strategy)

	// Event A: 9:00 AM - 11:00 AM, Priority 1
	eventA := Event{
		ID:        "A",
		Title:     "Event A",
		StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Priority:  1,
	}

	// Event B: 10:00 AM - 12:00 PM, Priority 3 (higher)
	eventB := Event{
		ID:        "B",
		Title:     "Event B",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Priority:  3,
	}

	cal.AddEvent(eventA)
	cal.AddEvent(eventB)

	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	// Should have 2 events: A (9-10) and B (10-12)
	if len(events) != 2 {
		t.Errorf("Expected 2 events after partial overlap resolution, got %d", len(events))
	}

	// First event should be A truncated to 9-10
	if !events[0].StartTime.Equal(time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)) {
		t.Errorf("First event should start at 9:00 AM, got %v", events[0].StartTime)
	}
	if !events[0].EndTime.Equal(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("First event should end at 10:00 AM, got %v", events[0].EndTime)
	}

	// Second event should be B (10-12)
	if !events[1].StartTime.Equal(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("Second event should start at 10:00 AM, got %v", events[1].StartTime)
	}
	if !events[1].EndTime.Equal(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("Second event should end at 12:00 PM, got %v", events[1].EndTime)
	}
}

func TestPriorityBasedStrategy_MultipleConflicts(t *testing.T) {
	strategy := &PriorityBasedStrategy{}
	cal := NewCalendar(strategy)

	// Low priority event that spans the entire day
	eventLow := Event{
		ID:        "LOW",
		Title:     "All Day Event",
		StartTime: time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
		Priority:  1,
	}

	// High priority meeting 1
	meeting1 := Event{
		ID:        "MEET1",
		Title:     "Important Meeting 1",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Priority:  5,
	}

	// High priority meeting 2
	meeting2 := Event{
		ID:        "MEET2",
		Title:     "Important Meeting 2",
		StartTime: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		Priority:  5,
	}

	cal.AddEvent(eventLow)
	cal.AddEvent(meeting1)
	cal.AddEvent(meeting2)

	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	// Should have 5 events: LOW(8-10), MEET1(10-11), LOW(11-14), MEET2(14-15), LOW(15-18)
	if len(events) != 5 {
		t.Errorf("Expected 5 events after multiple conflict resolution, got %d", len(events))
		for i, e := range events {
			t.Logf("Event %d: %s from %v to %v", i, e.Title, e.StartTime, e.EndTime)
		}
		return
	}

	// Verify the sequence
	expectedSequence := []struct {
		id    string
		start time.Time
		end   time.Time
	}{
		{"LOW", time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)},
		{"MEET1", time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)},
		{"LOW", time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)},
		{"MEET2", time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)},
		{"LOW", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC)},
	}

	for i, expected := range expectedSequence {
		if events[i].ID != expected.id {
			t.Errorf("Event %d: expected ID '%s', got '%s'", i, expected.id, events[i].ID)
		}
		if !events[i].StartTime.Equal(expected.start) {
			t.Errorf("Event %d: expected start time %v, got %v", i, expected.start, events[i].StartTime)
		}
		if !events[i].EndTime.Equal(expected.end) {
			t.Errorf("Event %d: expected end time %v, got %v", i, expected.end, events[i].EndTime)
		}
	}
}

func TestPriorityBasedStrategy_EqualPriority(t *testing.T) {
	strategy := &PriorityBasedStrategy{}
	cal := NewCalendar(strategy)

	// Two events with equal priority - first one should win
	eventA := Event{
		ID:        "A",
		Title:     "Event A",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Priority:  2,
	}

	eventB := Event{
		ID:        "B",
		Title:     "Event B",
		StartTime: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		Priority:  2,
	}

	cal.AddEvent(eventA)
	cal.AddEvent(eventB)

	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	// Should have 2 events: A (10-11) and B (12-13)
	if len(events) != 2 {
		t.Errorf("Expected 2 events with equal priority resolution, got %d", len(events))
	}

	// First event should be A truncated
	if events[0].ID != "A" {
		t.Errorf("First event should be A, got %s", events[0].ID)
	}
	if !events[0].EndTime.Equal(time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)) {
		t.Errorf("Event A should be truncated to end at 11:00 AM, got %v", events[0].EndTime)
	}

	// Second event should be B truncated
	if events[1].ID != "B" {
		t.Errorf("Second event should be B, got %s", events[1].ID)
	}
	if !events[1].StartTime.Equal(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("Event B should be truncated to start at 12:00 PM, got %v", events[1].StartTime)
	}
}

func TestPriorityBasedStrategy_WithRules(t *testing.T) {
	strategy := &PriorityBasedStrategy{}
	cal := NewCalendar(strategy)

	// Add a daily recurring rule with low priority
	startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)

	rule := Rule{
		ID:        "work-day",
		Title:     "Work Day",
		StartTime: startTime,
		EndTime:   endTime,
		Priority:  1,
		Recurrence: RecurrencePattern{
			Type:     Daily,
			Interval: 1,
		},
		ValidFrom: startTime,
	}

	cal.AddRule(rule)

	// Add a high priority meeting that conflicts
	meeting := Event{
		ID:        "MEETING",
		Title:     "Important Meeting",
		StartTime: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
		Priority:  5,
	}

	cal.AddEvent(meeting)

	rangeStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	events := cal.GetEvents(rangeStart, rangeEnd)

	// Should have 3 events: work(9-14), meeting(14-16), work(16-17)
	if len(events) != 3 {
		t.Errorf("Expected 3 events with rule and event conflict, got %d", len(events))
		for i, e := range events {
			t.Logf("Event %d: %s from %v to %v", i, e.Title, e.StartTime, e.EndTime)
		}
		return
	}

	// Verify the meeting is preserved
	foundMeeting := false
	for _, event := range events {
		if event.ID == "MEETING" {
			foundMeeting = true
			if !event.StartTime.Equal(time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)) {
				t.Errorf("Meeting should start at 2:00 PM, got %v", event.StartTime)
			}
			if !event.EndTime.Equal(time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)) {
				t.Errorf("Meeting should end at 4:00 PM, got %v", event.EndTime)
			}
		}
	}

	if !foundMeeting {
		t.Error("High priority meeting should be preserved")
	}
}
