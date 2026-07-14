package models

import (
	"testing"

	"github.com/google/uuid"
)

func validSubscription() Subscription {
	return Subscription{
		UserID:  uuid.New(),
		EventID: uuid.New(),
		Status:  SubscriptionStatusGoing,
	}
}

func TestSubscriptionValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Subscription)
		wantErr bool
	}{
		{"going", func(s *Subscription) {}, false},
		{"maybe", func(s *Subscription) { s.Status = SubscriptionStatusMaybe }, false},
		{"declined", func(s *Subscription) { s.Status = SubscriptionStatusDeclined }, false},
		{"no user", func(s *Subscription) { s.UserID = uuid.Nil }, true},
		{"no event", func(s *Subscription) { s.EventID = uuid.Nil }, true},
		{"empty status", func(s *Subscription) { s.Status = "" }, true},
		{"unknown status", func(s *Subscription) { s.Status = "interested" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := validSubscription()
			tt.mutate(&s)
			err := s.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}
