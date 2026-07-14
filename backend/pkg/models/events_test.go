package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func validEvent() Event {
	return Event{
		ID:            uuid.New(),
		Name:          "Ramen night",
		Type:          "restaurant",
		Lat:           48.8566,
		Lon:           2.3522,
		LocationLabel: "Kodawari Ramen, Paris",
		CreatedBy:     uuid.New(),
		Status:        EventStatusDraft,
	}
}

func TestEventValidate(t *testing.T) {
	at := time.Date(2026, 8, 1, 19, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		mutate  func(*Event)
		wantErr bool
	}{
		{"draft without date", func(e *Event) {}, false},
		{"empty name", func(e *Event) { e.Name = "" }, true},
		{"custom type is allowed", func(e *Event) { e.Type = "escape room" }, false},
		{"empty type", func(e *Event) { e.Type = "" }, true},
		{"unknown status", func(e *Event) { e.Status = "cancelled" }, true},
		{"lat out of range", func(e *Event) { e.Lat = 91 }, true},
		{"lon out of range", func(e *Event) { e.Lon = -181 }, true},
		{"missing creator", func(e *Event) { e.CreatedBy = uuid.Nil }, true},
		{
			"scheduled with date",
			func(e *Event) { e.Status = EventStatusScheduled; e.StartsAt = &at },
			false,
		},
		{
			"scheduled without date",
			func(e *Event) { e.Status = EventStatusScheduled },
			true,
		},
		{
			"past without date",
			func(e *Event) { e.Status = EventStatusPast },
			true,
		},
		{
			"polling with date",
			func(e *Event) { e.Status = EventStatusPolling; e.StartsAt = &at },
			true,
		},
		{
			"polling without date",
			func(e *Event) { e.Status = EventStatusPolling },
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := validEvent()
			tt.mutate(&e)
			err := e.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestIsPast(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name   string
		mutate func(*Event)
		want   bool
	}{
		{"draft with no date", func(e *Event) {}, false},
		{
			"scheduled in the future",
			func(e *Event) { e.Status = EventStatusScheduled; e.StartsAt = &future },
			false,
		},
		{
			"date passed but status not swept",
			func(e *Event) { e.Status = EventStatusScheduled; e.StartsAt = &past },
			true,
		},
		{
			"status past",
			func(e *Event) { e.Status = EventStatusPast; e.StartsAt = &past },
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := validEvent()
			tt.mutate(&e)
			if got := e.IsPast(now); got != tt.want {
				t.Fatalf("IsPast() = %v, want %v", got, tt.want)
			}
		})
	}
}
