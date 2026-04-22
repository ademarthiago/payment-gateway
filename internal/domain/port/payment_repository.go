package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
)

// PaymentRepository defines the contract for payment persistence
// This is a port — the adapter (postgres) implements this interface
type PaymentRepository interface {
	Save(ctx context.Context, payment *entity.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
	FindByExternalID(ctx context.Context, externalID string) (*entity.Payment, error)
	Update(ctx context.Context, payment *entity.Payment) error
	ExistsByExternalID(ctx context.Context, externalID string) (bool, error)
}

// TransactionRepository defines the contract for transaction persistence
type TransactionRepository interface {
	Save(ctx context.Context, transaction *entity.Transaction) error
	FindByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entity.Transaction, error)
}
