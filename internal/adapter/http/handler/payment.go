package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/usecase"
)

type PaymentHandler struct {
	createUC *usecase.CreatePaymentUseCase
	getUC    *usecase.GetPaymentUseCase
	refundUC *usecase.ProcessRefundUseCase
}

func NewPaymentHandler(
	createUC *usecase.CreatePaymentUseCase,
	getUC *usecase.GetPaymentUseCase,
	refundUC *usecase.ProcessRefundUseCase,
) *PaymentHandler {
	return &PaymentHandler{
		createUC: createUC,
		getUC:    getUC,
		refundUC: refundUC,
	}
}

// CreatePayment godoc
// @Summary      Create a payment
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        request body createPaymentRequest true "Payment request"
// @Success      201 {object} createPaymentResponse
// @Failure      400 {object} errorResponse
// @Failure      422 {object} errorResponse
// @Router       /payments [post]
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	output, err := h.createUC.Execute(r.Context(), usecase.CreatePaymentInput{
		ExternalID:  req.ExternalID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Provider:    req.Provider,
		Description: req.Description,
		Metadata:    req.Metadata,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, createPaymentResponse{
		ID:         output.ID,
		ExternalID: output.ExternalID,
		Amount:     output.Amount,
		Currency:   output.Currency,
		Status:     output.Status,
		Provider:   output.Provider,
		CreatedAt:  output.CreatedAt,
	})
}

// GetPayment godoc
// @Summary      Get a payment by ID
// @Tags         payments
// @Produce      json
// @Param        id path string true "Payment ID"
// @Success      200 {object} usecase.GetPaymentOutput
// @Failure      404 {object} errorResponse
// @Router       /payments/{id} [get]
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment id")
		return
	}

	output, err := h.getUC.ExecuteByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrPaymentNotFound) {
			writeError(w, http.StatusNotFound, "payment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, output)
}

// ProcessRefund godoc
// @Summary      Process a refund
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        id path string true "Payment ID"
// @Param        request body refundRequest true "Refund request"
// @Success      200 {object} usecase.ProcessRefundOutput
// @Failure      400 {object} errorResponse
// @Failure      404 {object} errorResponse
// @Router       /payments/{id}/refund [post]
func (h *PaymentHandler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment id")
		return
	}

	var req refundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	output, err := h.refundUC.Execute(r.Context(), usecase.ProcessRefundInput{
		PaymentID: id,
		Amount:    req.Amount,
		Reason:    req.Reason,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrPaymentNotFound) {
			writeError(w, http.StatusNotFound, "payment not found")
			return
		}
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, output)
}
