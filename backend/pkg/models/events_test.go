package models

import (
	"testing"

	"github.com/google/uuid"
)

func validCreateEventRequest() CreateEventRequest {
	return CreateEventRequest{
		GroupID:       uuid.New(),
		Name:          "Ramen night",
		Type:          "restaurant",
		Lat:           48.8566,
		Lon:           2.3522,
		LocationLabel: "Kodawari Ramen, Paris",
	}
}

func TestCreateEventRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*CreateEventRequest)
		wantErr bool
	}{
		{"valid", func(r *CreateEventRequest) {}, false},
		{"missing group", func(r *CreateEventRequest) { r.GroupID = uuid.Nil }, true},
		{"empty name", func(r *CreateEventRequest) { r.Name = "" }, true},
		{"custom type is allowed", func(r *CreateEventRequest) { r.Type = "escape room" }, false},
		{"empty type", func(r *CreateEventRequest) { r.Type = "" }, true},
		{"lat out of range", func(r *CreateEventRequest) { r.Lat = 91 }, true},
		{"lon out of range", func(r *CreateEventRequest) { r.Lon = -181 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := validCreateEventRequest()
			tt.mutate(&r)
			err := r.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
