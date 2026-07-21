package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"

	"github.com/DoppleDankster/uncaved/internal/store"
)

// psql aliases the shared builder so this package's SQL reads locally while the
// placeholder policy stays in internal/store.
var psql = store.Builder

// User is the persistence mapping for a row in users.
type User struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Label     string    `db:"label"`
	CreatedAt time.Time
}

type Repo struct {
	db store.DBTX
}

func NewRepo(db store.DBTX) *Repo {
	return &Repo{db}
}

func (u *Repo) ByID(ctx context.Context, id uuid.UUID) (User, error) {
	var user User

	query, args, err := psql.Select("id, name, label, created_at").
		From("users").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return User{}, fmt.Errorf("store: build user by id %v: %w", id, err)
	}

	err = pgxscan.Get(ctx, u.db, &user, query, args...)
	if err != nil {
		if pgxscan.NotFound(err) {
			return User{}, store.ErrNotFound
		}
		return User{}, fmt.Errorf("store: select user by id %v: %w", id, err)
	}
	return user, nil
}

func (u *Repo) List(ctx context.Context) ([]User, error) {
	var users []User

	query, args, err := psql.Select("id, name, label, created_at").
		From("users").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("store: build users: %w", err)
	}

	err = pgxscan.Select(ctx, u.db, &users, query, args...)
	if err != nil {
		return []User{}, fmt.Errorf("store: select users: %w", err)
	}
	return users, nil
}

func (u *Repo) Create(ctx context.Context, user User) (User, error) {
	query, args, err := psql.Insert("users").
		Columns("id", "name", "label").
		Values(user.ID, user.Name, user.Label).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return User{}, fmt.Errorf("store: build insert user: %w", err)
	}

	err = u.db.QueryRow(ctx, query, args...).Scan(&user.CreatedAt)
	if err != nil {
		return User{}, fmt.Errorf("store: scanning inserted user: %w", err)
	}
	return user, nil
}

func (u *Repo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query, args, err := psql.
		Delete("users").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("store: build delete user by id %v: %w", id, err)
	}
	_, err = u.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("store: exec delete user by id %v: %w", id, err)
	}
	return nil
}
