package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateEventRequest is the client payload for POST /events. It carries only
// what the client owns: the server assigns id, status (draft), and timestamps.
type CreateEventRequest struct {
	GroupID       uuid.UUID  `json:"group_id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	StartsAt      *time.Time `json:"starts_at"`
	Lat           float64    `json:"lat"`
	Lon           float64    `json:"lon"`
	LocationLabel string     `json:"location_label"`
}

// Validate checks the request shape before it reaches the store. Domain-state
// invariants (the lifecycle, who may post) belong to the service layer, not
// here; this only guards the fields the client sends.
func (r CreateEventRequest) Validate() error {
	if r.GroupID == uuid.Nil {
		return fmt.Errorf("event: group_id is required")
	}
	if r.Name == "" {
		return fmt.Errorf("event: name is required")
	}
	if r.Type == "" {
		return fmt.Errorf("event: type is required")
	}
	if r.Lat < -90 || r.Lat > 90 {
		return fmt.Errorf("event: lat %v out of range", r.Lat)
	}
	if r.Lon < -180 || r.Lon > 180 {
		return fmt.Errorf("event: lon %v out of range", r.Lon)
	}
	return nil
}

// EventResponse is the API shape of an event: snake_case JSON, decoupled from
// the store row so persistence changes don't leak to clients.
type EventResponse struct {
	ID            uuid.UUID  `json:"id"`
	GroupID       uuid.UUID  `json:"group_id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	StartsAt      *time.Time `json:"starts_at"`
	Lat           float64    `json:"lat"`
	Lon           float64    `json:"lon"`
	LocationLabel string     `json:"location_label"`
	CreatedBy     uuid.UUID  `json:"created_by"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
