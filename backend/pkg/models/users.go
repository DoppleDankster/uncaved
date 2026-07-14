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

	// AvatarKey is an object storage key for an uploaded avatar, nil when the
	// client should fall back to rendering initials.
	AvatarKey *string `json:"avatar_key"`

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
	if u.AvatarKey != nil && *u.AvatarKey == "" {
		return fmt.Errorf("user: avatar_key must not be empty when set")
	}
	return nil
}
