package usecase_test

import (
	"context"
	"testing"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
	"github.com/ademarthiago/payment-gateway/internal/domain/port/mock"
	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
	"github.com/ademarthiago/payment-gateway/internal/usecase"
)

func buildTestPayment(t *testing.T) *entity.Payment {
	t.Helper()
	money, _ := valueobject.NewMoney(5000, valueobject.CurrencyBRL)
	p, err := entity.NewPayment("order-123", money, "mock", "Test payment", nil)
	if err != nil {
		t.Fatalf("failed to build test payment: %v", err)
	}
	return p
}

func TestCreatePayment_Success(t *testing.T) {
	paymentRepo := &mock.PaymentRepositoryMock{}
	outboxRepo := &mock.OutboxRepositoryMock{}
	idempotency := &mock.IdempotencyStoreMock{}
	publisher := &mock.EventPublisherMock{}

	uc := usecase.NewCreatePaymentUseCase(paymentRepo, outboxRepo, idempotency, publisher)

	output, err := uc.Execute(context.Background(), usecase.CreatePaymentInput{
		ExternalID:  "order-123",
		Amount:      5000,
		Currency:    "BRL",
		Provider:    "mock",
		Description: "Test payment",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if output.ExternalID != "order-123" {
		t.Errorf("expected external_id order-123, got %s", output.ExternalID)
	}
	if output.Status != "pending" {
		t.Errorf("expected status pending, got %s", output.Status)
	}
	if output.Amount != 5000 {
		t.Errorf("expected amount 5000, got %d", output.Amount)
	}
	if len(publisher.Events) != 1 {
		t.Errorf("expected 1 event published, got %d", len(publisher.Events))
	}
	if publisher.Events[0].Type != "payment.created" {
		t.Errorf("expected event type payment.created, got %s", publisher.Events[0].Type)
	}
}

func TestCreatePayment_Idempotency(t *testing.T) {
	existingPayment := buildTestPayment(t)

	paymentRepo := &mock.PaymentRepositoryMock{
		FindByExternalIDFn: func(ctx context.Context, externalID string) (*entity.Payment, error) {
			return existingPayment, nil
		},
	}
	idempotency := &mock.IdempotencyStoreMock{
		ExistsFn: func(ctx context.Context, key string) (bool, error) {
			return true, nil
		},
	}
	publisher := &mock.EventPublisherMock{}

	uc := usecase.NewCreatePaymentUseCase(paymentRepo, &mock.OutboxRepositoryMock{}, idempotency, publisher)

	output, err := uc.Execute(context.Background(), usecase.CreatePaymentInput{
		ExternalID: "order-123",
		Amount:     5000,
		Currency:   "BRL",
		Provider:   "mock",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if output.ExternalID != "order-123" {
		t.Errorf("expected existing payment, got %s", output.ExternalID)
	}
	if len(publisher.Events) != 0 {
		t.Errorf("expected 0 events on idempotent call, got %d", len(publisher.Events))
	}
}

func TestCreatePayment_InvalidAmount(t *testing.T) {
	uc := usecase.NewCreatePaymentUseCase(
		&mock.PaymentRepositoryMock{},
		&mock.OutboxRepositoryMock{},
		&mock.IdempotencyStoreMock{},
		&mock.EventPublisherMock{},
	)
	_, err := uc.Execute(context.Background(), usecase.CreatePaymentInput{
		ExternalID: "order-123",
		Amount:     0,
		Currency:   "BRL",
		Provider:   "mock",
	})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestCreatePayment_InvalidCurrency(t *testing.T) {
	uc := usecase.NewCreatePaymentUseCase(
		&mock.PaymentRepositoryMock{},
		&mock.OutboxRepositoryMock{},
		&mock.IdempotencyStoreMock{},
		&mock.EventPublisherMock{},
	)
	_, err := uc.Execute(context.Background(), usecase.CreatePaymentInput{
		ExternalID: "order-123",
		Amount:     1000,
		Currency:   "XYZ",
		Provider:   "mock",
	})
	if err == nil {
		t.Fatal("expected error for invalid currency")
	}
}
