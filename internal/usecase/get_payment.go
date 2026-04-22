package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
	"github.com/ademarthiago/payment-gateway/internal/domain/port"
)

var ErrPaymentNotFound = errors.New("payment not found")

type GetPaymentOutput struct {
	ID           uuid.UUID
	ExternalID   string
	Amount       int64
	Currency     string
	Status       string
	Provider     string
	Description  string
	Metadata     map[string]any
	Transactions []TransactionOutput
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TransactionOutput struct {
	ID          uuid.UUID
	Type        string
	Amount      int64
	Currency    string
	Status      string
	ProviderRef string
	CreatedAt   time.Time
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
