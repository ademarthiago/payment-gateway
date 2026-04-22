package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
	"github.com/ademarthiago/payment-gateway/internal/domain/port"
	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
)

type CreatePaymentInput struct {
	ExternalID  string
	Amount      int64
	Currency    string
	Provider    string
	Description string
	Metadata    map[string]any
}

type CreatePaymentOutput struct {
	ID         uuid.UUID
	ExternalID string
	Amount     int64
	Currency   string
	Status     string
	Provider   string
	CreatedAt  time.Time
}

type CreatePaymentUseCase struct {
	paymentRepo      port.PaymentRepository
	outboxRepo       port.OutboxRepository
	idempotencyStore port.IdempotencyStore
	eventPublisher   port.EventPublisher
}

func NewCreatePaymentUseCase(
	paymentRepo port.PaymentRepository,
	outboxRepo port.OutboxRepository,
	idempotencyStore port.IdempotencyStore,
	eventPublisher port.EventPublisher,
) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{
		paymentRepo:      paymentRepo,
		outboxRepo:       outboxRepo,
		idempotencyStore: idempotencyStore,
		eventPublisher:   eventPublisher,
	}
}

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, input CreatePaymentInput) (*CreatePaymentOutput, error) {
	// Step 1: Idempotency check via Redis
	idempotencyKey := fmt.Sprintf("payment:create:%s", input.ExternalID)
	exists, err := uc.idempotencyStore.Exists(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("idempotency check failed: %w", err)
	}
	if exists {
		// Return existing payment
		payment, err := uc.paymentRepo.FindByExternalID(ctx, input.ExternalID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing payment: %w", err)
		}
		return toCreatePaymentOutput(payment), nil
	}

	// Step 2: Build domain objects
	currency := valueobject.Currency(input.Currency)
	money, err := valueobject.NewMoney(input.Amount, currency)
	if err != nil {
		return nil, fmt.Errorf("invalid money: %w", err)
	}

	payment, err := entity.NewPayment(
		input.ExternalID,
		money,
		input.Provider,
		input.Description,
		input.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid payment: %w", err)
	}

	// Step 3: Persist payment
	if err := uc.paymentRepo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Step 4: Save outbox event (Outbox Pattern — event persisted with payment)
	payload, _ := json.Marshal(map[string]any{
		"payment_id":  payment.ID(),
		"external_id": payment.ExternalID(),
		"amount":      payment.Money().Amount(),
		"currency":    payment.Money().Currency().String(),
		"status":      payment.Status().String(),
	})

	outboxMsg := &port.OutboxMessage{
		ID:            uuid.New(),
		AggregateID:   payment.ID(),
		AggregateType: "payment",
		EventType:     "payment.created",
		Payload:       payload,
	}
	if err := uc.outboxRepo.Save(ctx, outboxMsg); err != nil {
		return nil, fmt.Errorf("failed to save outbox message: %w", err)
	}

	// Step 5: Set idempotency key in Redis
	idempotencyValue, _ := json.Marshal(payment.ID())
	if err := uc.idempotencyStore.Set(ctx, idempotencyKey, idempotencyValue, 24*time.Hour); err != nil {
		// Non-fatal: log but don't fail
		_ = err
	}

	// Step 6: Publish event via channel
	if err := uc.eventPublisher.Publish(ctx, port.Event{
		Type:    "payment.created",
		Payload: payload,
	}); err != nil {
		// Non-fatal: outbox guarantees delivery
		_ = err
	}

	return toCreatePaymentOutput(payment), nil
}

func toCreatePaymentOutput(p *entity.Payment) *CreatePaymentOutput {
	return &CreatePaymentOutput{
		ID:         p.ID(),
		ExternalID: p.ExternalID(),
		Amount:     p.Money().Amount(),
		Currency:   p.Money().Currency().String(),
		Status:     p.Status().String(),
		Provider:   p.Provider(),
		CreatedAt:  p.CreatedAt(),
	}
}
