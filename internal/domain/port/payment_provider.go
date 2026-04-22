package port

import (
	"context"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
)

// ProviderRequest represents a charge request to a payment provider
type ProviderRequest struct {
	ExternalID  string
	Amount      valueobject.Money
	Description string
	Metadata    map[string]any
}

// ProviderResponse represents a response from a payment provider
type ProviderResponse struct {
	ProviderRef string
	Status      valueobject.PaymentStatus
	RawResponse []byte
}

// PaymentProvider defines the contract for external payment providers
// Adapters: StripeAdapter, PagSeguroAdapter, MockAdapter (for tests)
type PaymentProvider interface {
	Charge(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
	Refund(ctx context.Context, transaction *entity.Transaction) (*ProviderResponse, error)
}
