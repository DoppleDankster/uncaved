package models

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxNameLen      = 32
	maxUserLabelLen = 32
)

type User struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Label string    `json:"label"` // Chief Autism Office

	// Avatars are not stored on the user: they live at the deterministic object
	// key avatar/{id} in R2, with the client falling back to avatar/default on
	// 404. Nothing to carry in the model.

	CreatedAt time.Time `json:"created_at"`
}

func (u User) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("user: name is required")
	}
	if utf8.RuneCountInString(u.Name) > maxNameLen {
		return fmt.Errorf("user: name %v too long", u.Name)
	}
	if utf8.RuneCountInString(u.Label) > maxUserLabelLen {
		return fmt.Errorf("user: label %v too long", u.Label)
	}
	return nil
}
