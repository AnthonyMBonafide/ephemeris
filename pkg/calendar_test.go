package ephemeris

import (
	"testing"
	"time"
)

func TestCalendar_DayView(t *testing.T) {
	biweeklyDuration, err := time.ParseDuration("336h") // 2 weeks = 14 days
	if err != nil {
		t.FailNow()
	}
	c := Calendar{
		Name: "simple",
		Events: []Event{{
			Start:  time.Now(),
			End:    time.Now().AddDate(0, 0, 7),
			Name:   "Week long event, repeats every 2 weeks",
			Repeat: biweeklyDuration,
		}},
	}

	c.DayView()
}
