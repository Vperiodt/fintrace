package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/vanshika/fintrace/backend/internal/config"
	"github.com/vanshika/fintrace/backend/internal/graph"
	"github.com/vanshika/fintrace/backend/internal/logging"
	"github.com/vanshika/fintrace/backend/internal/repository"
	"github.com/vanshika/fintrace/backend/internal/server"
	"github.com/vanshika/fintrace/backend/internal/service"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.Logging)

	graphClient, err := buildGraphClient(ctx, logger, cfg)
	if err != nil {
		logger.Error("failed to create graph client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if graphClient != nil {
			if err := graphClient.Close(context.Background()); err != nil {
				logger.Warn("closing graph client failed", "error", err)
			}
		}
	}()

	repo := repository.New(graphClient)
	relationshipService := service.NewRelationshipService(repo, nil)
	apiHandlers := server.NewAPIHandlers(logger, relationshipService)

	router := server.NewRouter(logger, server.RouterDependencies{
		Health:           server.GraphHealthService{Client: graphClient},
		API:              apiHandlers,
		AllowedOrigins:   parseAllowedOrigins(cfg.HTTP.AllowedOriginsCSV),
		AllowCredentials: true,
	})

	srv := server.New(logger, cfg.HTTP, router)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("received shutdown signal", "signal", sig.String())
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("server stopped unexpectedly", "error", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

func buildGraphClient(ctx context.Context, logger *slog.Logger, cfg config.Config) (graph.Client, error) {
	if cfg.Graph.URI == "" {
		return nil, graph.ErrMissingURI
	}

	opts := graph.Options{
		URI:            cfg.Graph.URI,
		Database:       cfg.Graph.Database,
		Username:       cfg.Graph.Username,
		Password:       cfg.Graph.Password,
		MaxConnections: cfg.Graph.MaxConnections,
	}
	return graph.NewNeo4jClient(ctx, opts)
}

func parseAllowedOrigins(csv string) []string {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	var origins []string
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}
	return origins
}
