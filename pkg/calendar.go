package calendar

import (
	"sort"
	"time"
)

// Calendar represents a calendar with events and rules
type Calendar struct {
	events   []Event
	rules    []Rule
	strategy OverlapStrategy
}

// Event represents a calendar event with priority and metadata
type Event struct {
	ID        string
	Title     string
	StartTime time.Time
	EndTime   time.Time
	Priority  int     // Higher number = higher priority
	RuleID    *string // Optional reference to a rule that generated this event
}

// Rule represents a recurring event rule
type Rule struct {
	ID          string
	Title       string
	StartTime   time.Time
	EndTime     time.Time
	Priority    int
	Recurrence  RecurrencePattern
	ValidFrom   time.Time
	ValidUntil  *time.Time // nil means no end date
}

// RecurrencePattern defines how an event repeats
type RecurrencePattern struct {
	Type     RecurrenceType
	Interval int // e.g., every 2 weeks, every 3 months
}

// RecurrenceType defines the type of recurrence
type RecurrenceType int

const (
	Daily RecurrenceType = iota
	Weekly
	Monthly
	Yearly
)

// OverlapStrategy defines how to handle overlapping events
type OverlapStrategy interface {
	ResolveOverlaps(events []Event) []Event
}

// NewCalendar creates a new calendar with the specified overlap strategy
func NewCalendar(strategy OverlapStrategy) *Calendar {
	return &Calendar{
		events:   make([]Event, 0),
		rules:    make([]Rule, 0),
		strategy: strategy,
	}
}

// AddEvent adds an event to the calendar
func (c *Calendar) AddEvent(event Event) {
	c.events = append(c.events, event)
}

// AddRule adds a recurring event rule to the calendar
func (c *Calendar) AddRule(rule Rule) {
	c.rules = append(c.rules, rule)
}

// GetEvents returns all events in the calendar for a given time range
// This includes both standalone events and events generated from rules
func (c *Calendar) GetEvents(start, end time.Time) []Event {
	allEvents := make([]Event, 0)

	// Add standalone events that fall within the range
	for _, event := range c.events {
		if c.eventInRange(event, start, end) {
			allEvents = append(allEvents, event)
		}
	}

	// Generate events from rules
	for _, rule := range c.rules {
		ruleEvents := c.generateEventsFromRule(rule, start, end)
		allEvents = append(allEvents, ruleEvents...)
	}

	// Sort events by start time, then by priority (higher priority first)
	sort.Slice(allEvents, func(i, j int) bool {
		if allEvents[i].StartTime.Equal(allEvents[j].StartTime) {
			return allEvents[i].Priority > allEvents[j].Priority
		}
		return allEvents[i].StartTime.Before(allEvents[j].StartTime)
	})

	// Apply overlap resolution strategy
	return c.strategy.ResolveOverlaps(allEvents)
}

// GetAllEvents returns all standalone events (not including rule-generated events)
func (c *Calendar) GetAllEvents() []Event {
	return c.events
}

// GetAllRules returns all rules
func (c *Calendar) GetAllRules() []Rule {
	return c.rules
}

// eventInRange checks if an event overlaps with the given time range
func (c *Calendar) eventInRange(event Event, start, end time.Time) bool {
	return event.StartTime.Before(end) && event.EndTime.After(start)
}

// generateEventsFromRule generates events from a rule within the specified time range
func (c *Calendar) generateEventsFromRule(rule Rule, start, end time.Time) []Event {
	events := make([]Event, 0)

	// Start from the rule's start time or the requested start time, whichever is later
	current := rule.StartTime
	if start.After(current) {
		current = c.findNextOccurrence(rule, start)
	}

	eventID := 0
	for current.Before(end) {
		// Check if we're still within the rule's valid period
		if rule.ValidUntil != nil && current.After(*rule.ValidUntil) {
			break
		}

		// Calculate event end time
		duration := rule.EndTime.Sub(rule.StartTime)
		eventEnd := current.Add(duration)

		// Create event from rule
		event := Event{
			ID:        rule.ID + "_" + current.Format("20060102_1504"),
			Title:     rule.Title,
			StartTime: current,
			EndTime:   eventEnd,
			Priority:  rule.Priority,
			RuleID:    &rule.ID,
		}

		events = append(events, event)

		// Move to next occurrence
		current = c.getNextOccurrence(rule, current)
		eventID++

		// Safety check to prevent infinite loops
		if eventID > 10000 {
			break
		}
	}

	return events
}

// findNextOccurrence finds the next occurrence of a rule on or after the given time
func (c *Calendar) findNextOccurrence(rule Rule, after time.Time) time.Time {
	current := rule.StartTime

	for current.Before(after) {
		current = c.getNextOccurrence(rule, current)
	}

	return current
}

// getNextOccurrence calculates the next occurrence based on the recurrence pattern
func (c *Calendar) getNextOccurrence(rule Rule, current time.Time) time.Time {
	switch rule.Recurrence.Type {
	case Daily:
		return current.AddDate(0, 0, rule.Recurrence.Interval)
	case Weekly:
		return current.AddDate(0, 0, 7*rule.Recurrence.Interval)
	case Monthly:
		return current.AddDate(0, rule.Recurrence.Interval, 0)
	case Yearly:
		return current.AddDate(rule.Recurrence.Interval, 0, 0)
	default:
		return current.AddDate(0, 0, 1) // Default to daily
	}
}
