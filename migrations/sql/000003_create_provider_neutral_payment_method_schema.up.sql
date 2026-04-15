CREATE TABLE IF NOT EXISTS payment_method_infos (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    provider_customer_id TEXT,
    provider_payment_method_id TEXT NOT NULL,
    method_type TEXT,
    brand TEXT,
    last4 TEXT,
    exp_month BIGINT,
    exp_year BIGINT,
    network TEXT,
    wallet_address TEXT,
    CONSTRAINT uni_payment_method_infos_provider_payment_method_id UNIQUE (provider_payment_method_id)
);

CREATE INDEX IF NOT EXISTS idx_payment_method_infos_deleted_at ON payment_method_infos (deleted_at);
CREATE INDEX IF NOT EXISTS idx_payment_method_infos_user_id ON payment_method_infos (user_id);

-- Backfill from legacy Stripe table to provider-neutral schema.
INSERT INTO payment_method_infos (
    created_at,
    updated_at,
    deleted_at,
    user_id,
    provider,
    provider_customer_id,
    provider_payment_method_id,
    method_type,
    brand,
    last4,
    exp_month,
    exp_year
)
SELECT
    s.created_at,
    s.updated_at,
    s.deleted_at,
    s.user_id,
    'stripe',
    s.stripe_customer_id,
    s.stripe_payment_method_id,
    'card',
    s.brand,
    s.last4,
    s.exp_month,
    s.exp_year
FROM stripe_payment_info s
ON CONFLICT (provider_payment_method_id) DO NOTHING;
