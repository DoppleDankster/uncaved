package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventRepo(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	events := NewEventRepo(st.DB())
	users := NewUserRepo(st.DB())

	// events.created_by is a FK to users(id) — seed a creator first.
	creator, err := users.Create(ctx, User{ID: uuid.New(), Name: "Creator"})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	newEvent := func(status string, startsAt *time.Time) Event {
		return Event{
			ID:            uuid.New(),
			Name:          "Hike",
			Type:          "hike",
			StartsAt:      startsAt,
			Lat:           45.76,
			Lon:           4.83,
			LocationLabel: "Lyon",
			CreatedBy:     creator.ID,
			Status:        status,
		}
	}

	t.Run("create draft (no date) and read back", func(t *testing.T) {
		in := newEvent("draft", nil)

		created, err := events.Create(ctx, in)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
			t.Error("created_at/updated_at not populated by DB")
		}

		got, err := events.ByID(ctx, in.ID)
		if err != nil {
			t.Fatalf("by id: %v", err)
		}
		if got.Name != "Hike" || got.Type != "hike" || got.Status != "draft" ||
			got.CreatedBy != creator.ID || got.LocationLabel != "Lyon" {
			t.Errorf("round-trip mismatch: %+v", got)
		}
		if got.StartsAt != nil {
			t.Errorf("draft starts_at should be nil, got %v", *got.StartsAt)
		}
		if got.Lat != 45.76 || got.Lon != 4.83 {
			t.Errorf("coords mismatch: %v, %v", got.Lat, got.Lon)
		}
	})

	t.Run("scheduled event carries a date", func(t *testing.T) {
		when := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Microsecond)
		in := newEvent("scheduled", &when)

		if _, err := events.Create(ctx, in); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := events.ByID(ctx, in.ID)
		if err != nil {
			t.Fatalf("by id: %v", err)
		}
		if got.StartsAt == nil || !got.StartsAt.Equal(when) {
			t.Errorf("starts_at: got %v, want %v", got.StartsAt, when)
		}
	})

	t.Run("lifecycle CHECK rejects polling with a date", func(t *testing.T) {
		when := time.Now().UTC()
		if _, err := events.Create(ctx, newEvent("polling", &when)); err == nil {
			t.Fatal("expected CHECK violation for polling + starts_at, got nil")
		}
	})

	t.Run("missing id is ErrNotFound", func(t *testing.T) {
		if _, err := events.ByID(ctx, uuid.New()); !errors.Is(err, ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})

	t.Run("delete removes the event", func(t *testing.T) {
		in := newEvent("draft", nil)
		if _, err := events.Create(ctx, in); err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := events.DeleteByID(ctx, in.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
		if _, err := events.ByID(ctx, in.ID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("after delete want ErrNotFound, got %v", err)
		}
	})
}
