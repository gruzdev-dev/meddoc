package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/gruzdev-dev/meddoc/app/handlers"
	"github.com/gruzdev-dev/meddoc/app/server/middleware"
	"github.com/gruzdev-dev/meddoc/config"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
)

type Server struct {
	cfg      *config.Config
	handlers *handlers.Handlers
}

func NewServer(cfg *config.Config, handlers *handlers.Handlers) *Server {
	return &Server{
		cfg:      cfg,
		handlers: handlers,
	}
}

func (s *Server) Start() error {
	router := mux.NewRouter()

	router.Use(middleware.RequestID())
	router.Use(middleware.Logging())
	router.Use(middleware.Recovery())
	router.Use(middleware.Compression())
	router.Use(middleware.SecurityHeaders())
	api := router.PathPrefix("/api/v1").Subrouter()

	s.handlers.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
		Handler: router,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Info("shutting down server...", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}
	}

	logger.Info("server exited")
	return nil
}
