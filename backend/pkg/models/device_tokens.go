package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Platform is the FCM delivery target.
type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformWeb     Platform = "web"
)

func (p Platform) Valid() bool {
	switch p {
	case PlatformAndroid, PlatformWeb:
		return true
	}
	return false
}

// DeviceToken is one FCM registration token for one user's device. Pruned when
// FCM answers UNREGISTERED.
type DeviceToken struct {
	UserID   uuid.UUID `json:"user_id"`
	Token    string    `json:"token"`
	Platform Platform  `json:"platform"`

	UpdatedAt time.Time `json:"updated_at"`
}

func (d DeviceToken) Validate() error {
	if d.UserID == uuid.Nil {
		return fmt.Errorf("device token: user_id is required")
	}
	if d.Token == "" {
		return fmt.Errorf("device token: token is required")
	}
	if !d.Platform.Valid() {
		return fmt.Errorf("device token: invalid platform %q", d.Platform)
	}
	return nil
}
