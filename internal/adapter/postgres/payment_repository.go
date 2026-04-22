package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ademarthiago/payment-gateway/internal/domain/entity"
	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Save(ctx context.Context, p *entity.Payment) error {
	metadata, err := json.Marshal(p.Metadata())
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO payment.payments
			(id, external_id, amount, currency, status, provider, description, metadata, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID(), p.ExternalID(), p.Money().Amount(), p.Money().Currency().String(),
		p.Status().String(), p.Provider(), p.Description(), metadata,
		p.CreatedAt(), p.UpdatedAt(),
	)
	return err
}

func (r *PaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, external_id, amount, currency, status, provider, description, metadata, created_at, updated_at
		FROM payment.payments WHERE id = $1`, id)
	return scanPayment(row)
}

func (r *PaymentRepository) FindByExternalID(ctx context.Context, externalID string) (*entity.Payment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, external_id, amount, currency, status, provider, description, metadata, created_at, updated_at
		FROM payment.payments WHERE external_id = $1`, externalID)
	return scanPayment(row)
}

func (r *PaymentRepository) Update(ctx context.Context, p *entity.Payment) error {
	metadata, err := json.Marshal(p.Metadata())
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE payment.payments SET status=$1, metadata=$2, updated_at=$3 WHERE id=$4`,
		p.Status().String(), metadata, p.UpdatedAt(), p.ID(),
	)
	return err
}

func (r *PaymentRepository) ExistsByExternalID(ctx context.Context, externalID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM payment.payments WHERE external_id = $1)`, externalID,
	).Scan(&exists)
	return exists, err
}

func scanPayment(row pgx.Row) (*entity.Payment, error) {
	var (
		id          uuid.UUID
		externalID  string
		amount      int64
		currency    string
		status      string
		provider    string
		description string
		metadataRaw []byte
		createdAt   time.Time
		updatedAt   time.Time
	)
	err := row.Scan(&id, &externalID, &amount, &currency, &status,
		&provider, &description, &metadataRaw, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan payment: %w", err)
	}
	var metadata map[string]any
	_ = json.Unmarshal(metadataRaw, &metadata)
	money, err := valueobject.NewMoney(amount, valueobject.Currency(currency))
	if err != nil {
		return nil, fmt.Errorf("failed to build money: %w", err)
	}
	return entity.ReconstitutPayment(
		id, externalID, money,
		valueobject.PaymentStatus(status),
		provider, description, metadata, nil,
		createdAt, updatedAt,
	), nil
}
