package event

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/DoppleDankster/uncaved/internal/store"
	"github.com/DoppleDankster/uncaved/pkg/models"
)

// Handler serves the event HTTP routes. It holds the shared store and builds a
// Repo per request off the pool (and later, a tx via store.WithTx).
type Handler struct {
	store *store.Store
}

func NewHandler(st *store.Store) *Handler {
	return &Handler{store: st}
}

// RegisterRoutes mounts the event routes. The server calls this so the feature
// owns its own URL space.
func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/events", h.list)
	r.GET("/events/:id", h.get)
}

// toEventResponse maps a store row to the public API DTO. The mapping lives here
// so pkg/models stays free of any store dependency.
func toEventResponse(e Event) models.EventResponse {
	return models.EventResponse{
		ID:            e.ID,
		GroupID:       e.GroupID,
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

// list: GET /events
func (h *Handler) list(c *gin.Context) {
	ctx := c.Request.Context()

	events, err := NewRepo(h.store.DB()).List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "list events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	resp := make([]models.EventResponse, len(events))
	for i, e := range events {
		resp[i] = toEventResponse(e)
	}
	c.JSON(http.StatusOK, resp)
}

// get: GET /events/:id
func (h *Handler) get(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	e, err := NewRepo(h.store.DB()).ByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		slog.ErrorContext(ctx, "get event", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, toEventResponse(e))
}
