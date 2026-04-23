// Package usecase contains the application use cases.
// Each use case orchestrates domain objects and ports to fulfill a single business operation.
// No infrastructure details (SQL, Redis commands, HTTP) belong here.
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

// ProcessRefundInput carries the refund request data from the HTTP layer.
type ProcessRefundInput struct {
	PaymentID uuid.UUID
	Amount    int64
	Reason    string
}

// ProcessRefundOutput is what the HTTP handler serializes back after a successful refund.
// The transaction starts as pending — actual money movement depends on the provider adapter.
type ProcessRefundOutput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	PaymentID     uuid.UUID `json:"payment_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// ProcessRefundUseCase handles refund requests by validating the payment state,
// recording the refund transaction, and emitting a "payment.refunded" event.
type ProcessRefundUseCase struct {
	paymentRepo    port.PaymentRepository
	outboxRepo     port.OutboxRepository
	eventPublisher port.EventPublisher
}

// NewProcessRefundUseCase wires the use case with its required ports.
func NewProcessRefundUseCase(
	paymentRepo port.PaymentRepository,
	outboxRepo port.OutboxRepository,
	eventPublisher port.EventPublisher,
) *ProcessRefundUseCase {
	return &ProcessRefundUseCase{
		paymentRepo:    paymentRepo,
		outboxRepo:     outboxRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute validates the payment is in a refundable state, records the refund transaction,
// and writes the outbox event. Returns an error if the payment is not completed or already refunded.
func (uc *ProcessRefundUseCase) Execute(ctx context.Context, input ProcessRefundInput) (*ProcessRefundOutput, error) {
	// Step 1: Fetch payment
	payment, err := uc.paymentRepo.FindByID(ctx, input.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	// Step 2: Validate state transition — only completed payments can be refunded
	if err := payment.Transition(valueobject.PaymentStatusRefunded); err != nil {
		return nil, fmt.Errorf("invalid refund: %w", err)
	}

	// Step 3: Build refund transaction using the same currency as the original payment
	refundMoney, err := valueobject.NewMoney(input.Amount, payment.Money().Currency())
	if err != nil {
		return nil, fmt.Errorf("invalid refund amount: %w", err)
	}

	tx, err := entity.NewTransaction(
		payment.ID(),
		entity.TransactionTypeRefund,
		refundMoney,
		"",
		map[string]any{"reason": input.Reason},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}

	payment.AddTransaction(tx)

	// Step 4: Persist updated payment
	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Step 5: Save outbox event
	payload, _ := json.Marshal(map[string]any{
		"payment_id":     payment.ID(),
		"transaction_id": tx.ID(),
		"amount":         tx.Amount().Amount(),
		"currency":       tx.Amount().Currency().String(),
		"reason":         input.Reason,
	})

	if err := uc.outboxRepo.Save(ctx, &port.OutboxMessage{
		ID:            uuid.New(),
		AggregateID:   payment.ID(),
		AggregateType: "payment",
		EventType:     "payment.refunded",
		Payload:       payload,
	}); err != nil {
		return nil, fmt.Errorf("failed to save outbox message: %w", err)
	}

	// Step 6: Publish event via channel (non-fatal)
	_ = uc.eventPublisher.Publish(ctx, port.Event{
		Type:    "payment.refunded",
		Payload: payload,
	})

	return &ProcessRefundOutput{
		TransactionID: tx.ID(),
		PaymentID:     payment.ID(),
		Amount:        tx.Amount().Amount(),
		Currency:      tx.Amount().Currency().String(),
		Status:        tx.Status().String(),
		CreatedAt:     tx.CreatedAt(),
	}, nil
}
