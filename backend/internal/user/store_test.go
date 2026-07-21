package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/DoppleDankster/uncaved/internal/store"
	"github.com/DoppleDankster/uncaved/internal/storetest"
)

func TestUserRepo(t *testing.T) {
	st := storetest.NewStore(t)
	repo := NewRepo(st.DB())
	ctx := context.Background()

	// Unique ids per subtest keep them isolated on the shared database.
	t.Run("create and read back", func(t *testing.T) {
		in := User{ID: uuid.New(), Name: "Ada", Label: "organizer"}

		created, err := repo.Create(ctx, in)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.CreatedAt.IsZero() {
			t.Error("created_at not populated from the DB")
		}

		got, err := repo.ByID(ctx, in.ID)
		if err != nil {
			t.Fatalf("by id: %v", err)
		}
		if got.ID != in.ID || got.Name != "Ada" || got.Label != "organizer" {
			t.Errorf("round-trip mismatch: got %+v, want %+v", got, in)
		}
		if !got.CreatedAt.Equal(created.CreatedAt) {
			t.Errorf("created_at mismatch: byID %v, create %v", got.CreatedAt, created.CreatedAt)
		}
	})

	t.Run("missing id is ErrNotFound", func(t *testing.T) {
		_, err := repo.ByID(ctx, uuid.New())
		if !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})

	t.Run("list contains inserted users", func(t *testing.T) {
		a := User{ID: uuid.New(), Name: "List-A"}
		b := User{ID: uuid.New(), Name: "List-B"}
		for _, u := range []User{a, b} {
			if _, err := repo.Create(ctx, u); err != nil {
				t.Fatalf("create %s: %v", u.Name, err)
			}
		}

		users, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		// Subset check (other subtests inserted rows too), not exact count.
		found := make(map[uuid.UUID]bool, len(users))
		for _, u := range users {
			found[u.ID] = true
		}
		if !found[a.ID] || !found[b.ID] {
			t.Errorf("list missing inserted users: have a=%v b=%v", found[a.ID], found[b.ID])
		}
	})
}
