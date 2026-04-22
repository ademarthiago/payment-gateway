package mock

import (
	"context"

	"github.com/google/uuid"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
)

type PaymentRepositoryMock struct {
	SaveFn               func(ctx context.Context, payment *entity.Payment) error
	FindByIDFn           func(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
	FindByExternalIDFn   func(ctx context.Context, externalID string) (*entity.Payment, error)
	UpdateFn             func(ctx context.Context, payment *entity.Payment) error
	ExistsByExternalIDFn func(ctx context.Context, externalID string) (bool, error)
}

func (m *PaymentRepositoryMock) Save(ctx context.Context, payment *entity.Payment) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, payment)
	}
	return nil
}

func (m *PaymentRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *PaymentRepositoryMock) FindByExternalID(ctx context.Context, externalID string) (*entity.Payment, error) {
	if m.FindByExternalIDFn != nil {
		return m.FindByExternalIDFn(ctx, externalID)
	}
	return nil, nil
}

func (m *PaymentRepositoryMock) Update(ctx context.Context, payment *entity.Payment) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, payment)
	}
	return nil
}

func (m *PaymentRepositoryMock) ExistsByExternalID(ctx context.Context, externalID string) (bool, error) {
	if m.ExistsByExternalIDFn != nil {
		return m.ExistsByExternalIDFn(ctx, externalID)
	}
	return false, nil
}
