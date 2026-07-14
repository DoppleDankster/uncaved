package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventStatus tracks the event lifecycle:
// draft -> polling -> scheduled -> past.
type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPolling   EventStatus = "polling"
	EventStatusScheduled EventStatus = "scheduled"
	EventStatusPast      EventStatus = "past"
)

func (s EventStatus) Valid() bool {
	switch s {
	case EventStatusDraft, EventStatusPolling, EventStatusScheduled, EventStatusPast:
		return true
	}
	return false
}

// Event is a single meetup: a place, a status, and eventually a date.
type Event struct {
	ID uuid.UUID `json:"id"`

	Name string `json:"name"`

	// Type is free-form string ex "restaurant", "bar", "hike".
	Type string `json:"type"`

	// StartsAt is nil until a date is set directly or confirmed from a poll.
	StartsAt *time.Time `json:"starts_at"`

	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	LocationLabel string  `json:"location_label"`

	CreatedBy uuid.UUID   `json:"created_by"`
	Status    EventStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate reports whether the event is internally consistent, independent of
// any transition it is making.
func (e *Event) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("event: name is required")
	}
	if e.Type == "" {
		return fmt.Errorf("event: type is required")
	}
	if !e.Status.Valid() {
		return fmt.Errorf("event: invalid status %q", e.Status)
	}
	if e.Lat < -90 || e.Lat > 90 {
		return fmt.Errorf("event: lat %v out of range", e.Lat)
	}
	if e.Lon < -180 || e.Lon > 180 {
		return fmt.Errorf("event: lon %v out of range", e.Lon)
	}
	if e.CreatedBy == uuid.Nil {
		return fmt.Errorf("event: created_by is required")
	}
	// A scheduled or past event has a date by definition; a polling one must not,
	// since confirming a poll option is what sets StartsAt.
	switch e.Status {
	case EventStatusScheduled, EventStatusPast:
		if e.StartsAt == nil {
			return fmt.Errorf("event: status %q requires starts_at", e.Status)
		}
	case EventStatusPolling:
		if e.StartsAt != nil {
			return fmt.Errorf("event: status %q must not have starts_at", e.Status)
		}
	}
	return nil
}

// IsPast reports whether the event has happened, covering the case where the
// date has passed but the status has not been swept to past yet.
func (e *Event) IsPast(now time.Time) bool {
	if e.Status == EventStatusPast {
		return true
	}
	return e.StartsAt != nil && e.StartsAt.Before(now)
}
