package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db    HealthChecker
	redis HealthChecker
}

func NewHealthHandler(db, redis HealthChecker) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

type healthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	services := map[string]string{
		"postgres": "up",
		"redis":    "up",
	}
	status := "up"

	if err := h.db.Ping(ctx); err != nil {
		services["postgres"] = "down"
		status = "degraded"
	}

	if err := h.redis.Ping(ctx); err != nil {
		services["redis"] = "down"
		status = "degraded"
	}

	httpStatus := http.StatusOK
	if status == "degraded" {
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Services:  services,
	})
}
