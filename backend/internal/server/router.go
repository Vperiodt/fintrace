package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// RouterDependencies collects handler dependencies.
type RouterDependencies struct {
	Health           HealthService
	API              *APIHandlers
	AllowedOrigins   []string
	AllowCredentials bool
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
	}

	handler := http.Handler(loggingMiddleware(logger, mux))
	if len(deps.AllowedOrigins) > 0 {
		handler = corsMiddleware(deps.AllowedOrigins, deps.AllowCredentials)(handler)
	}
	return handler
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

func corsMiddleware(allowedOrigins []string, allowCredentials bool) func(http.Handler) http.Handler {
	normalized := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		normalized[origin] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" || (!containsOrigin(normalized, origin) && !containsOrigin(normalized, "*")) {
				if r.Method == http.MethodOptions {
					// Reject bare pre-flight if origin is not whitelisted.
					w.WriteHeader(http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			if allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func containsOrigin(set map[string]struct{}, origin string) bool {
	_, ok := set[origin]
	return ok
}
