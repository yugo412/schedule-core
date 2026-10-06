package models

import "time"

type Schedule struct {
	ID        int        `db:"id"`
	Slug      string     `db:"slug"`
	Url       string     `db:"url"`
	Title     string     `db:"title"`
	StartedAt *time.Time `db:"started_at"`
}

// HasStarted reports whether the schedule start time is in the past. It
// returns false when the start time is unknown.
func (s *Schedule) HasStarted() bool {
	return s.StartedAt != nil && s.StartedAt.Before(time.Now())
}
