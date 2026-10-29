package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	replicaID string
	pool      *pgxpool.Pool
	startTime time.Time
}

func NewHealthHandler(replicaID string, pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		replicaID: replicaID,
		pool:      pool,
		startTime: time.Now(),
	}
}

type HealthResponse struct {
	Status    string  `json:"status"`
	Database  string  `json:"database"`
	ReplicaID string  `json:"replica_id"`
	Version   string  `json:"version"`
	UptimeSec float64 `json:"uptime_seconds"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if h.pool != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.pool.Ping(ctx); err != nil {
			dbStatus = "disconnected"
		}
	} else {
		dbStatus = "mock_in_memory"
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if dbStatus == "disconnected" {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := HealthResponse{
		Status:    status,
		Database:  dbStatus,
		ReplicaID: h.replicaID,
		Version:   "2.0.0",
		UptimeSec: time.Since(h.startTime).Seconds(),
	}

	writeJSON(w, httpStatus, resp)
}
