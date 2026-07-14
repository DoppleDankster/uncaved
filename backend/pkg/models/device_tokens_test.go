package models

import (
	"testing"

	"github.com/google/uuid"
)

func validDeviceToken() DeviceToken {
	return DeviceToken{
		UserID:   uuid.New(),
		Token:    "fcm-registration-token",
		Platform: PlatformAndroid,
	}
}

func TestDeviceTokenValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*DeviceToken)
		wantErr bool
	}{
		{"android", func(d *DeviceToken) {}, false},
		{"web", func(d *DeviceToken) { d.Platform = PlatformWeb }, false},
		{"no user", func(d *DeviceToken) { d.UserID = uuid.Nil }, true},
		{"no token", func(d *DeviceToken) { d.Token = "" }, true},
		{"empty platform", func(d *DeviceToken) { d.Platform = "" }, true},
		{"unknown platform", func(d *DeviceToken) { d.Platform = "ios" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := validDeviceToken()
			tt.mutate(&d)
			err := d.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}
