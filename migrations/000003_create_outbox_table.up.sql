-- =============================================================================
-- Migration: 000003 - Create outbox table (Outbox Pattern)
-- =============================================================================

CREATE TABLE IF NOT EXISTS payment.outbox (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id    UUID NOT NULL,
    aggregate_type  VARCHAR(100) NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    payload         JSONB NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'pending',
    attempts        INTEGER NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    CONSTRAINT outbox_status_valid CHECK (
        status IN ('pending', 'processing', 'processed', 'failed')
    ),
    CONSTRAINT outbox_attempts_positive CHECK (attempts >= 0)
);

CREATE INDEX idx_outbox_status ON payment.outbox(status) WHERE status = 'pending';
CREATE INDEX idx_outbox_aggregate_id ON payment.outbox(aggregate_id);
CREATE INDEX idx_outbox_created_at ON payment.outbox(created_at ASC);
