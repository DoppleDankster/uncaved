package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/DoppleDankster/uncaved/internal/store"
)

// serviceName labels the server span otelgin opens for each request.
//
// TODO: plumb this from cfg.Instrumentation.ServiceName instead of the
// placeholder once the service config is threaded into the server package.
const serviceName = "uncaved"

type Webservice struct {
	api  *gin.Engine
	port int
}

func NewWerservice(cfg Config, st *store.Store) Webservice {
	a := gin.New()

	// otelgin is registered first so its span wraps recovery and records the final status.
	a.Use(
		otelgin.Middleware(serviceName),
		gin.Recovery(),
	)

	registerRoutes(a, st)

	return Webservice{
		api:  a,
		port: cfg.Port,
	}
}

func (w Webservice) Run() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", w.port),
		Handler: w.api,
	}

	// Buffered so the goroutine never blocks on send if Run has already
	// returned via the signal path.
	serveErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		return fmt.Errorf("server: listen: %w", err)
	case <-quit:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
