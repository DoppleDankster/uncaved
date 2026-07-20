package server

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DoppleDankster/uncaved/internal/store"
)

// eventResponse is the API shape of an event — snake_case JSON, decoupled from
// the store row so persistence changes don't leak to clients.
type eventResponse struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	StartsAt      *time.Time `json:"starts_at"`
	Lat           float64    `json:"lat"`
	Lon           float64    `json:"lon"`
	LocationLabel string     `json:"location_label"`
	CreatedBy     uuid.UUID  `json:"created_by"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func toEventResponse(e store.Event) eventResponse {
	return eventResponse{
		ID:            e.ID,
		Name:          e.Name,
		Type:          e.Type,
		StartsAt:      e.StartsAt,
		Lat:           e.Lat,
		Lon:           e.Lon,
		LocationLabel: e.LocationLabel,
		CreatedBy:     e.CreatedBy,
		Status:        e.Status,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

// listEvents: GET /events
func (h *handlers) listEvents(c *gin.Context) {
	ctx := c.Request.Context()

	events, err := store.NewEventRepo(h.store.DB()).List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "list events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	resp := make([]eventResponse, len(events))
	for i, e := range events {
		resp[i] = toEventResponse(e)
	}
	c.JSON(http.StatusOK, resp)
}

// getEvent: GET /events/:id
func (h *handlers) getEvent(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	event, err := store.NewEventRepo(h.store.DB()).ByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		slog.ErrorContext(ctx, "get event", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, toEventResponse(event))
}
