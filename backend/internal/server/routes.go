package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DoppleDankster/uncaved/internal/store"
)

// registerRoutes mounts every HTTP route on the engine. Handlers are grouped by
// aggregate on the handlers struct, which holds the store they query.
func registerRoutes(r *gin.Engine, st *store.Store) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	h := &handlers{store: st}
	r.GET("/events", h.listEvents)
	r.GET("/events/:id", h.getEvent)
}

// handlers carries the dependencies every HTTP handler needs. Repos are built
// per request from st.DB() so each runs on the pool (and later, a tx via WithTx).
type handlers struct {
	store *store.Store
}
