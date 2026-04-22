-- =============================================================================
-- Migration: 000002 - Create transactions table
-- =============================================================================

CREATE TABLE IF NOT EXISTS payment.transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id      UUID NOT NULL REFERENCES payment.payments(id),
    type            VARCHAR(50) NOT NULL,
    amount          BIGINT NOT NULL,
    status          VARCHAR(50) NOT NULL,
    provider_ref    VARCHAR(255),               -- provider transaction ID
    error_message   TEXT,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transactions_amount_positive CHECK (amount > 0),
    CONSTRAINT transactions_type_valid CHECK (
        type IN ('charge', 'refund', 'chargeback')
    ),
    CONSTRAINT transactions_status_valid CHECK (
        status IN ('pending', 'processing', 'completed', 'failed')
    )
);

CREATE INDEX idx_transactions_payment_id ON payment.transactions(payment_id);
CREATE INDEX idx_transactions_status ON payment.transactions(status);
CREATE INDEX idx_transactions_created_at ON payment.transactions(created_at DESC);
