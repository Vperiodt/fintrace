package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// RouterDependencies collects handler dependencies.
type RouterDependencies struct {
	Health HealthService
	API    *APIHandlers
}

// NewRouter wires the HTTP routes exposed by the backend API.
func NewRouter(logger *slog.Logger, deps RouterDependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := http.StatusOK
		payload := map[string]any{
			"status": "ok",
		}

		if deps.Health != nil {
			if err := deps.Health.Probe(ctx); err != nil {
				logger.Error("health probe failed", "error", err)
				status = http.StatusServiceUnavailable
				payload["status"] = "degraded"
				payload["error"] = err.Error()
			}
		}

		respondJSON(w, status, payload)
	})

	if deps.API != nil {
		mux.HandleFunc("/users", deps.API.handleUsers)
		mux.HandleFunc("/transactions", deps.API.handleTransactions)
		mux.HandleFunc("/relationships/user/", deps.API.handleUserRelationships)
		mux.HandleFunc("/relationships/transaction/", deps.API.handleTransactionRelationships)
		mux.HandleFunc("/analytics/shortest-path", deps.API.handleShortestPath)
		mux.HandleFunc("/export/users", deps.API.handleExportUsers)
		mux.HandleFunc("/export/transactions", deps.API.handleExportTransactions)
	}

	return loggingMiddleware(logger, mux)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
