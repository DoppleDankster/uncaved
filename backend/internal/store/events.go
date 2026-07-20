package store

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
)

// eventColumns is the full projection, shared by every read so the SELECT list
// can't drift from the struct.
const eventColumns = "id, name, type, starts_at, lat, lon, location_label, created_by, status, created_at, updated_at"

type Event struct {
	ID            uuid.UUID  `db:"id"`
	Name          string     `db:"name"`
	Type          string     `db:"type"`
	StartsAt      *time.Time `db:"starts_at"`
	Lat           float64    `db:"lat"`
	Lon           float64    `db:"lon"`
	LocationLabel string     `db:"location_label"`
	CreatedBy     uuid.UUID  `db:"created_by"`
	Status        string     `db:"status"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

type EventRepo struct {
	db DBTX
}

func NewEventRepo(db DBTX) *EventRepo {
	return &EventRepo{db}
}

func (r *EventRepo) ByID(ctx context.Context, id uuid.UUID) (Event, error) {
	var e Event

	query, args, err := psql.Select(eventColumns).
		From("events").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return Event{}, fmt.Errorf("store: build event by id %v: %w", id, err)
	}

	err = pgxscan.Get(ctx, r.db, &e, query, args...)
	if err != nil {
		if pgxscan.NotFound(err) {
			return Event{}, ErrNotFound
		}
		return Event{}, fmt.Errorf("store: select event by id %v: %w", id, err)
	}
	return e, nil
}

func (r *EventRepo) List(ctx context.Context) ([]Event, error) {
	var events []Event

	query, args, err := psql.Select(eventColumns).
		From("events").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("store: build event list: %w", err)
	}

	err = pgxscan.Select(ctx, r.db, &events, query, args...)
	if err != nil {
		return nil, fmt.Errorf("store: list events: %w", err)
	}
	return events, nil
}

func (r *EventRepo) Create(ctx context.Context, event Event) (Event, error) {
	query, args, err := psql.Insert("events").
		Columns("id", "name", "type", "starts_at", "lat", "lon",
			"location_label", "created_by", "status").
		Values(event.ID, event.Name, event.Type, event.StartsAt, event.Lat, event.Lon,
			event.LocationLabel, event.CreatedBy, event.Status).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return Event{}, fmt.Errorf("store: build insert event: %w", err)
	}

	// created_at and updated_at are DB-owned (DEFAULT now()); read them back.
	err = r.db.QueryRow(ctx, query, args...).Scan(&event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		return Event{}, fmt.Errorf("store: insert event: %w", err)
	}
	return event, nil
}

func (r *EventRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query, args, err := psql.Delete("events").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("store: build delete event by id %v: %w", id, err)
	}
	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("store: exec delete event by id %v: %w", id, err)
	}
	return nil
}
