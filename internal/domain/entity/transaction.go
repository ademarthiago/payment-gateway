// Package entity contains the core domain entities for the payment system.
// Transaction records a single financial operation tied to a payment.
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
)

// TransactionType classifies what kind of money movement a transaction represents.
type TransactionType string

const (
	// TransactionTypeCharge is created when we attempt to capture money from the customer.
	TransactionTypeCharge TransactionType = "charge"
	// TransactionTypeRefund is created when we return money to the customer after a completed charge.
	TransactionTypeRefund TransactionType = "refund"
	// TransactionTypeChargeback is created when the customer disputes a charge with their bank —
	// this usually comes from a webhook from the provider, not from a direct API call.
	TransactionTypeChargeback TransactionType = "chargeback"
)

// Transaction represents a financial operation linked to a payment.
// It's immutable after creation except for status — only Fail() and Complete() mutate it.
// Transactions live inside the Payment aggregate and are never fetched independently.
type Transaction struct {
	id          uuid.UUID
	paymentID   uuid.UUID
	txType      TransactionType
	amount      valueobject.Money
	status      valueobject.PaymentStatus
	providerRef string
	errorMsg    string
	metadata    map[string]any
	createdAt   time.Time
}

// ErrInvalidTransactionType is returned when an unknown type is passed to NewTransaction.
var ErrInvalidTransactionType = errors.New("invalid transaction type")

// NewTransaction creates a transaction and validates the type.
// All transactions start as pending — use Complete() or Fail() after the provider responds.
func NewTransaction(
	paymentID uuid.UUID,
	txType TransactionType,
	amount valueobject.Money,
	providerRef string,
	metadata map[string]any,
) (*Transaction, error) {
	if txType != TransactionTypeCharge &&
		txType != TransactionTypeRefund &&
		txType != TransactionTypeChargeback {
		return nil, ErrInvalidTransactionType
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}
	return &Transaction{
		id:          uuid.New(),
		paymentID:   paymentID,
		txType:      txType,
		amount:      amount,
		status:      valueobject.PaymentStatusPending,
		providerRef: providerRef,
		metadata:    metadata,
		createdAt:   time.Now().UTC(),
	}, nil
}

// Fail marks the transaction as failed and stores the provider error for debugging.
func (t *Transaction) Fail(errMsg string) {
	t.status = valueobject.PaymentStatusFailed
	t.errorMsg = errMsg
}

// Complete marks the transaction as successfully processed by the provider.
func (t *Transaction) Complete() {
	t.status = valueobject.PaymentStatusCompleted
}

// Getters
func (t *Transaction) ID() uuid.UUID                     { return t.id }
func (t *Transaction) PaymentID() uuid.UUID              { return t.paymentID }
func (t *Transaction) Type() TransactionType             { return t.txType }
func (t *Transaction) Amount() valueobject.Money         { return t.amount }
func (t *Transaction) Status() valueobject.PaymentStatus { return t.status }
func (t *Transaction) ProviderRef() string               { return t.providerRef }
func (t *Transaction) ErrorMsg() string                  { return t.errorMsg }
func (t *Transaction) Metadata() map[string]any          { return t.metadata }
func (t *Transaction) CreatedAt() time.Time              { return t.createdAt }
