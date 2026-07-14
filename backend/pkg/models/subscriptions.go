package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus is the three-state RSVP.
type SubscriptionStatus string

const (
	SubscriptionStatusGoing    SubscriptionStatus = "going"
	SubscriptionStatusMaybe    SubscriptionStatus = "maybe"
	SubscriptionStatusDeclined SubscriptionStatus = "declined"
)

func (s SubscriptionStatus) Valid() bool {
	switch s {
	case SubscriptionStatusGoing, SubscriptionStatusMaybe, SubscriptionStatusDeclined:
		return true
	}
	return false
}

// Subscription is one user's RSVP to one event. It doubles as the notification
// targeting list, so every push path reads from here.
//
// Keyed by (UserID, EventID) one per person per event.
type Subscription struct {
	UserID  uuid.UUID          `json:"user_id"`
	EventID uuid.UUID          `json:"event_id"`
	Status  SubscriptionStatus `json:"status"`

	UpdatedAt time.Time `json:"updated_at"`
}

func (s Subscription) Validate() error {
	if s.UserID == uuid.Nil {
		return fmt.Errorf("subscription: user_id is required")
	}
	if s.EventID == uuid.Nil {
		return fmt.Errorf("subscription: event_id is required")
	}
	if !s.Status.Valid() {
		return fmt.Errorf("subscription: invalid status %q", s.Status)
	}
	return nil
}
