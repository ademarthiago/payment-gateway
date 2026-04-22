package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	adapterhttp "github.com/ademarthiago/payment-gateway/internal/adapter/http"
	"github.com/ademarthiago/payment-gateway/internal/adapter/http/handler"
	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
	"github.com/ademarthiago/payment-gateway/internal/domain/port/mock"
	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
	"github.com/ademarthiago/payment-gateway/internal/usecase"
	"github.com/google/uuid"
)

func buildRouter(t *testing.T, paymentRepo *mock.PaymentRepositoryMock) http.Handler {
	t.Helper()
	outboxRepo := &mock.OutboxRepositoryMock{}
	idempotency := &mock.IdempotencyStoreMock{}
	publisher := &mock.EventPublisherMock{}
	createUC := usecase.NewCreatePaymentUseCase(paymentRepo, outboxRepo, idempotency, publisher)
	getUC := usecase.NewGetPaymentUseCase(paymentRepo)
	refundUC := usecase.NewProcessRefundUseCase(paymentRepo, outboxRepo, publisher)
	paymentHandler := handler.NewPaymentHandler(createUC, getUC, refundUC)
	healthHandler := handler.NewHealthHandler(&noopHealthChecker{}, &noopHealthChecker{})
	return adapterhttp.NewRouter(paymentHandler, healthHandler)
}

type noopHealthChecker struct{}

func (n *noopHealthChecker) Ping(_ context.Context) error { return nil }

func TestE2E_CreatePayment(t *testing.T) {
	router := buildRouter(t, &mock.PaymentRepositoryMock{})
	body := map[string]any{
		"external_id": "order-e2e-001",
		"amount":      9900,
		"currency":    "BRL",
		"provider":    "mock",
		"description": "E2E test payment",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["external_id"] != "order-e2e-001" {
		t.Errorf("expected external_id order-e2e-001, got %v", resp["external_id"])
	}
	if resp["status"] != "pending" {
		t.Errorf("expected status pending, got %v", resp["status"])
	}
}

func TestE2E_CreatePayment_InvalidBody(t *testing.T) {
	router := buildRouter(t, &mock.PaymentRepositoryMock{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestE2E_GetPayment_NotFound(t *testing.T) {
	router := buildRouter(t, &mock.PaymentRepositoryMock{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestE2E_GetPayment_Success(t *testing.T) {
	money, _ := valueobject.NewMoney(9900, valueobject.CurrencyBRL)
	existing, _ := entity.NewPayment("order-e2e-002", money, "mock", "Test", nil)
	paymentRepo := &mock.PaymentRepositoryMock{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Payment, error) {
			return existing, nil
		},
	}
	router := buildRouter(t, paymentRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/"+existing.ID().String(), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestE2E_HealthCheck(t *testing.T) {
	router := buildRouter(t, &mock.PaymentRepositoryMock{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
