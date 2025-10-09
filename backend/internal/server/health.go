package server

import (
	"context"

	"github.com/vanshika/fintrace/backend/internal/graph"
)

// HealthService defines behaviour for readiness probes.
type HealthService interface {
	Probe(ctx context.Context) error
}

// GraphHealthService verifies graph connectivity as part of health checks.
type GraphHealthService struct {
	Client graph.Client
}

// Probe implements the HealthService interface.
func (s GraphHealthService) Probe(ctx context.Context) error {
	if s.Client == nil {
		return nil
	}
	return s.Client.VerifyConnectivity(ctx)
}
