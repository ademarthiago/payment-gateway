-- =============================================================================
-- Migration: 000001 - Create payments table
-- =============================================================================

CREATE TABLE IF NOT EXISTS payment.payments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id     VARCHAR(255) NOT NULL UNIQUE,  -- idempotency key
    amount          BIGINT NOT NULL,                -- stored in cents
    currency        VARCHAR(3) NOT NULL,            -- ISO 4217 (USD, BRL)
    status          VARCHAR(50) NOT NULL,
    provider        VARCHAR(50) NOT NULL,
    description     TEXT,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_amount_positive CHECK (amount > 0),
    CONSTRAINT payments_currency_format CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT payments_status_valid CHECK (
        status IN ('pending', 'processing', 'completed', 'failed', 'refunded')
    )
);

CREATE INDEX idx_payments_external_id ON payment.payments(external_id);
CREATE INDEX idx_payments_status ON payment.payments(status);
CREATE INDEX idx_payments_created_at ON payment.payments(created_at DESC);
