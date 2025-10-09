package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vanshika/fintrace/backend/internal/config"
)

// Server represents the HTTP server lifecycle.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	cfg        config.HTTPConfig
}

// New constructs a Server instance using the provided router.
func New(logger *slog.Logger, cfg config.HTTPConfig, handler http.Handler) *Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		cfg:        cfg,
	}
}

// Start begins listening for HTTP traffic.
func (s *Server) Start() error {
	s.logger.Info("starting http server", "addr", s.httpServer.Addr)
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully terminates all active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.httpServer.Shutdown(ctx)
}
