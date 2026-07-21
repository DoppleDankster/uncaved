package event

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"

	"github.com/DoppleDankster/uncaved/internal/store"
)

// psql aliases the shared builder so this package's SQL reads locally while the
// placeholder policy stays in internal/store.
var psql = store.Builder

// eventColumns is the full projection, shared by every read so the SELECT list
// can't drift from the struct.
const eventColumns = "id, group_id, name, type, starts_at, lat, lon, location_label, created_by, status, created_at, updated_at"

// Event is the persistence mapping for a row in events. Domain rules do not live
// here — the store is just the SQL boundary.
type Event struct {
	ID            uuid.UUID  `db:"id"`
	GroupID       uuid.UUID  `db:"group_id"`
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

type Repo struct {
	db store.DBTX
}

func NewRepo(db store.DBTX) *Repo {
	return &Repo{db}
}

func (r *Repo) ByID(ctx context.Context, id uuid.UUID) (Event, error) {
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
			return Event{}, store.ErrNotFound
		}
		return Event{}, fmt.Errorf("store: select event by id %v: %w", id, err)
	}
	return e, nil
}

func (r *Repo) List(ctx context.Context) ([]Event, error) {
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

func (r *Repo) Create(ctx context.Context, e Event) (Event, error) {
	query, args, err := psql.Insert("events").
		Columns("id", "group_id", "name", "type", "starts_at", "lat", "lon",
			"location_label", "created_by", "status").
		Values(e.ID, e.GroupID, e.Name, e.Type, e.StartsAt, e.Lat, e.Lon,
			e.LocationLabel, e.CreatedBy, e.Status).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return Event{}, fmt.Errorf("store: build insert event: %w", err)
	}

	// created_at and updated_at are DB-owned (DEFAULT now()); read them back.
	err = r.db.QueryRow(ctx, query, args...).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return Event{}, fmt.Errorf("store: insert event: %w", err)
	}
	return e, nil
}

func (r *Repo) DeleteByID(ctx context.Context, id uuid.UUID) error {
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
