package ephemeris

// Esure BruteCondenser implements Condencer at compile time
var _ Condencer = BruteCondenser{}

type Condencer interface {
	Condence([]Event) []Event
}

type BruteCondenser struct{}

// Condence squashes all the events recursively adding squashed events to a new slice and repeating over the new slice until there are no changes left.
// Example:
func (b BruteCondenser) Condence([]Event) []Event {
	return nil
}
