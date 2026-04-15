CREATE TABLE IF NOT EXISTS payment_idempotency_keys (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    operation TEXT NOT NULL,
    key TEXT NOT NULL,
    response TEXT,
    CONSTRAINT uni_payment_idempotency_op_key UNIQUE (operation, key)
);

CREATE INDEX IF NOT EXISTS idx_payment_idempotency_keys_deleted_at ON payment_idempotency_keys (deleted_at);

CREATE TABLE IF NOT EXISTS payment_settlements (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    provider TEXT NOT NULL,
    payment_method_id TEXT NOT NULL,
    status TEXT NOT NULL,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    provider_reference TEXT,
    network TEXT,
    tx_hash TEXT,
    contract_address TEXT,
    settlement_proof TEXT,
    finalized_at_unix BIGINT,
    CONSTRAINT uni_payment_settlements_op_key UNIQUE (operation, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_payment_settlements_deleted_at ON payment_settlements (deleted_at);
CREATE INDEX IF NOT EXISTS idx_payment_settlements_status ON payment_settlements (status);
CREATE INDEX IF NOT EXISTS idx_payment_settlements_user_id ON payment_settlements (user_id);
