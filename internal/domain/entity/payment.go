package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
)

// Payment is the aggregate root of the payment domain
type Payment struct {
	id           uuid.UUID
	externalID   string // idempotency key
	money        valueobject.Money
	status       valueobject.PaymentStatus
	provider     string
	description  string
	metadata     map[string]any
	transactions []*Transaction
	createdAt    time.Time
	updatedAt    time.Time
}

var (
	ErrExternalIDRequired = errors.New("external_id is required")
	ErrProviderRequired   = errors.New("provider is required")
)

func NewPayment(
	externalID string,
	money valueobject.Money,
	provider string,
	description string,
	metadata map[string]any,
) (*Payment, error) {
	if externalID == "" {
		return nil, ErrExternalIDRequired
	}
	if provider == "" {
		return nil, ErrProviderRequired
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}
	now := time.Now().UTC()
	return &Payment{
		id:          uuid.New(),
		externalID:  externalID,
		money:       money,
		status:      valueobject.PaymentStatusPending,
		provider:    provider,
		description: description,
		metadata:    metadata,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Reconstitute rebuilds a Payment from persistence (no validation)
func ReconstitutPayment(
	id uuid.UUID,
	externalID string,
	money valueobject.Money,
	status valueobject.PaymentStatus,
	provider string,
	description string,
	metadata map[string]any,
	transactions []*Transaction,
	createdAt time.Time,
	updatedAt time.Time,
) *Payment {
	return &Payment{
		id:           id,
		externalID:   externalID,
		money:        money,
		status:       status,
		provider:     provider,
		description:  description,
		metadata:     metadata,
		transactions: transactions,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (p *Payment) Transition(next valueobject.PaymentStatus) error {
	if err := p.status.CanTransitionTo(next); err != nil {
		return err
	}
	p.status = next
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Payment) AddTransaction(t *Transaction) {
	p.transactions = append(p.transactions, t)
	p.updatedAt = time.Now().UTC()
}

// Getters
func (p *Payment) ID() uuid.UUID                     { return p.id }
func (p *Payment) ExternalID() string                { return p.externalID }
func (p *Payment) Money() valueobject.Money          { return p.money }
func (p *Payment) Status() valueobject.PaymentStatus { return p.status }
func (p *Payment) Provider() string                  { return p.provider }
func (p *Payment) Description() string               { return p.description }
func (p *Payment) Metadata() map[string]any          { return p.metadata }
func (p *Payment) Transactions() []*Transaction      { return p.transactions }
func (p *Payment) CreatedAt() time.Time              { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time              { return p.updatedAt }
