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
