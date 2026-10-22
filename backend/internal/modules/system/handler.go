// Package system exposes the service's own status (health check).
package system

import (
	"context"
	"time"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
)

// Pinger checks that a dependency is reachable. The database pool satisfies it.
type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	db      Pinger
	version string
}

func NewHandler(db Pinger, version string) *Handler {
	return &Handler{db: db, version: version}
}

func (h *Handler) GetHealth(ctx context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.db.Ping(pingCtx); err != nil {
		return nil, apperror.Unavailable(err, "database unreachable")
	}
	return api.GetHealth200JSONResponse{
		Status:  api.HealthStatusOk,
		Version: h.version,
		Checks:  map[string]api.HealthChecks{"database": api.HealthChecksOk},
	}, nil
}
