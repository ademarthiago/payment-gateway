// Package entity contains the core domain entities for the payment system.
// Payment is the aggregate root — all state changes go through it, never bypassing the state machine.
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
)

// Payment is the aggregate root. All fields are private on purpose — the only way to change status
// is through Transition(), which enforces the state machine. Direct field access would let callers
// skip validation and corrupt the payment lifecycle.
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
	// ErrExternalIDRequired is returned when external_id is empty.
	// Without it we can't do idempotency checks — every payment needs a client-side reference.
	ErrExternalIDRequired = errors.New("external_id is required")
	// ErrProviderRequired is returned when no payment provider is specified.
	ErrProviderRequired = errors.New("provider is required")
)

// NewPayment creates a new payment and validates required fields.
// Use this for new payments only — for rebuilding from the database, use ReconstitutPayment.
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

	// Copy caller's map to prevent external mutation after creation.
	meta := make(map[string]any, len(metadata))
	for k, v := range metadata {
		meta[k] = v
	}

	now := time.Now().UTC()
	return &Payment{
		id:          uuid.New(),
		externalID:  externalID,
		money:       money,
		status:      valueobject.PaymentStatusPending,
		provider:    provider,
		description: description,
		metadata:    meta,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstitutPayment rebuilds a Payment from a database row without running validations.
// The DB is the source of truth for persisted state — we trust what's stored and skip checks
// that would reject records created under different validation rules in the past.
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

// Transition moves the payment to a new status, enforcing the state machine rules.
// Returns an error if the transition is not allowed — e.g. trying to refund a failed payment.
func (p *Payment) Transition(next valueobject.PaymentStatus) error {
	if err := p.status.CanTransitionTo(next); err != nil {
		return err
	}
	p.status = next
	p.updatedAt = time.Now().UTC()
	return nil
}

// AddTransaction appends a transaction to the payment and bumps updatedAt.
// Transactions are owned by the payment aggregate — never persisted independently.
func (p *Payment) AddTransaction(t *Transaction) {
	if t == nil {
		return
	}
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
func (p *Payment) CreatedAt() time.Time              { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time              { return p.updatedAt }

// Metadata returns a shallow copy of the payment metadata to prevent external mutation.
func (p *Payment) Metadata() map[string]any {
	cp := make(map[string]any, len(p.metadata))
	for k, v := range p.metadata {
		cp[k] = v
	}
	return cp
}

// Transactions returns a shallow copy of the transactions slice to prevent external mutation.
func (p *Payment) Transactions() []*Transaction {
	cp := make([]*Transaction, len(p.transactions))
	copy(cp, p.transactions)
	return cp
}
