package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
	"github.com/ademarthiago/payment-gateway/internal/domain/port"
	"github.com/google/uuid"
)

var ErrPaymentNotFound = errors.New("payment not found")

type GetPaymentOutput struct {
	ID           uuid.UUID           `json:"id"`
	ExternalID   string              `json:"external_id"`
	Amount       int64               `json:"amount"`
	Currency     string              `json:"currency"`
	Status       string              `json:"status"`
	Provider     string              `json:"provider"`
	Description  string              `json:"description"`
	Metadata     map[string]any      `json:"metadata"`
	Transactions []TransactionOutput `json:"transactions"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

type TransactionOutput struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	ProviderRef string    `json:"provider_ref"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetPaymentUseCase struct {
	paymentRepo port.PaymentRepository
}

func NewGetPaymentUseCase(paymentRepo port.PaymentRepository) *GetPaymentUseCase {
	return &GetPaymentUseCase{paymentRepo: paymentRepo}
}

func (uc *GetPaymentUseCase) ExecuteByID(ctx context.Context, id uuid.UUID) (*GetPaymentOutput, error) {
	payment, err := uc.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	return mapPaymentToOutput(payment), nil
}

func (uc *GetPaymentUseCase) ExecuteByExternalID(ctx context.Context, externalID string) (*GetPaymentOutput, error) {
	payment, err := uc.paymentRepo.FindByExternalID(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	return mapPaymentToOutput(payment), nil
}

func mapPaymentToOutput(p *entity.Payment) *GetPaymentOutput {
	txOutputs := make([]TransactionOutput, 0, len(p.Transactions()))
	for _, tx := range p.Transactions() {
		txOutputs = append(txOutputs, TransactionOutput{
			ID:          tx.ID(),
			Type:        string(tx.Type()),
			Amount:      tx.Amount().Amount(),
			Currency:    tx.Amount().Currency().String(),
			Status:      tx.Status().String(),
			ProviderRef: tx.ProviderRef(),
			CreatedAt:   tx.CreatedAt(),
		})
	}
	return &GetPaymentOutput{
		ID:           p.ID(),
		ExternalID:   p.ExternalID(),
		Amount:       p.Money().Amount(),
		Currency:     p.Money().Currency().String(),
		Status:       p.Status().String(),
		Provider:     p.Provider(),
		Description:  p.Description(),
		Metadata:     p.Metadata(),
		Transactions: txOutputs,
		CreatedAt:    p.CreatedAt(),
		UpdatedAt:    p.UpdatedAt(),
	}
}
