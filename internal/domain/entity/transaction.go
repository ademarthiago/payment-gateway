package entity

import (
	"errors"
	"time"

	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
	"github.com/google/uuid"
)

// TransactionType represents the type of a transaction
type TransactionType string

const (
	TransactionTypeCharge     TransactionType = "charge"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeChargeback TransactionType = "chargeback"
)

// Transaction represents a financial transaction linked to a payment
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

var ErrInvalidTransactionType = errors.New("invalid transaction type")

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

func (t *Transaction) Fail(errMsg string) {
	t.status = valueobject.PaymentStatusFailed
	t.errorMsg = errMsg
}

func (t *Transaction) Complete() {
	t.status = valueobject.PaymentStatusCompleted
}

// Getters
func (t *Transaction) ID() uuid.UUID                    { return t.id }
func (t *Transaction) PaymentID() uuid.UUID             { return t.paymentID }
func (t *Transaction) Type() TransactionType            { return t.txType }
func (t *Transaction) Amount() valueobject.Money        { return t.amount }
func (t *Transaction) Status() valueobject.PaymentStatus { return t.status }
func (t *Transaction) ProviderRef() string              { return t.providerRef }
func (t *Transaction) ErrorMsg() string                 { return t.errorMsg }
func (t *Transaction) Metadata() map[string]any         { return t.metadata }
func (t *Transaction) CreatedAt() time.Time             { return t.createdAt }
