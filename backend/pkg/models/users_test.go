package models

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validUser() User {
	return User{
		ID:        uuid.New(),
		Name:      "DoppleDankster",
		Label:     "Chief Autism Officer",
		CreatedAt: time.Now(),
	}
}

func TestUserValidate(t *testing.T) {
	longString := "abcdefghijklmnopqrstuvwxyz1234567890"

	tests := []struct {
		name    string
		mutate  func(*User)
		wantErr bool
	}{
		{"valid full user", func(u *User) {}, false},
		{"user without username", func(u *User) { u.Name = "" }, true},
		{
			"username too long",
			func(u *User) { u.Name = longString },
			true,
		},
		{
			"label too long",
			func(u *User) { u.Label = longString },
			true,
		},
		{
			"no label",
			func(u *User) { u.Label = "" },
			false,
		},
		{
			"username at the cap",
			func(u *User) { u.Name = strings.Repeat("a", maxNameLen) },
			false,
		},
		{
			"username over the cap",
			func(u *User) { u.Name = strings.Repeat("a", maxNameLen+1) },
			true,
		},
		{
			"label at the cap",
			func(u *User) { u.Label = strings.Repeat("a", maxUserLabelLen) },
			false,
		},
		{
			"label over the cap",
			func(u *User) { u.Label = strings.Repeat("a", maxUserLabelLen+1) },
			true,
		},
		// Caps count runes, not bytes: an accented label is 2 bytes per char.
		{
			"accented label at the cap",
			func(u *User) { u.Label = strings.Repeat("é", maxUserLabelLen) },
			false,
		},
		{
			"accented label over the cap",
			func(u *User) { u.Label = strings.Repeat("é", maxUserLabelLen+1) },
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := validUser()
			tt.mutate(&u)
			err := u.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil, got error")
			}
		})
	}
}
