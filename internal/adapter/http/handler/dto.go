package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// --- Request DTOs ---

type createPaymentRequest struct {
	ExternalID  string         `json:"external_id"`
	Amount      int64          `json:"amount"`
	Currency    string         `json:"currency"`
	Provider    string         `json:"provider"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata"`
}

func (r createPaymentRequest) validate() error {
	if r.ExternalID == "" {
		return errors.New("external_id is required")
	}
	if r.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if r.Currency == "" {
		return errors.New("currency is required")
	}
	if r.Provider == "" {
		return errors.New("provider is required")
	}
	return nil
}

type refundRequest struct {
	Amount int64  `json:"amount"`
	Reason string `json:"reason"`
}

// --- Response DTOs ---

type createPaymentResponse struct {
	ID         uuid.UUID `json:"id"`
	ExternalID string    `json:"external_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	Provider   string    `json:"provider"`
	CreatedAt  time.Time `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
