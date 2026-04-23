// Package port defines the outbound interfaces (ports) that the domain needs from infrastructure.
// Adapters in internal/adapter/ implement these interfaces — the domain never imports them directly.
package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
)

// PaymentRepository is the persistence port for the Payment aggregate.
// The concrete implementation lives in adapter/postgres/payment_repository.go.
type PaymentRepository interface {
	// Save persists a new payment. Fails if a payment with the same ID already exists.
	Save(ctx context.Context, payment *entity.Payment) error
	// FindByID returns the payment with its transaction history, or nil if not found.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
	// FindByExternalID looks up a payment by the client-provided idempotency key.
	// Returns nil if no payment exists with that external_id.
	FindByExternalID(ctx context.Context, externalID string) (*entity.Payment, error)
	// Update persists status changes and metadata updates to an existing payment.
	Update(ctx context.Context, payment *entity.Payment) error
	// ExistsByExternalID is a cheap existence check used before full idempotency lookup.
	ExistsByExternalID(ctx context.Context, externalID string) (bool, error)
}

// TransactionRepository is the persistence port for Transaction records.
// The concrete implementation lives in adapter/postgres/payment_repository.go.
type TransactionRepository interface {
	// Save persists a new transaction linked to a payment.
	Save(ctx context.Context, transaction *entity.Transaction) error
	// FindByPaymentID returns all transactions for a given payment, ordered by created_at.
	FindByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entity.Transaction, error)
}
