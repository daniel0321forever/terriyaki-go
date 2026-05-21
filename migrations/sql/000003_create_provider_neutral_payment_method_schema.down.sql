-- Roll back by preserving data in legacy Stripe table (if still available) before dropping canonical table.
INSERT INTO stripe_payment_info (
    created_at,
    updated_at,
    deleted_at,
    user_id,
    stripe_customer_id,
    stripe_payment_method_id,
    brand,
    last4,
    exp_month,
    exp_year
)
SELECT
    p.created_at,
    p.updated_at,
    p.deleted_at,
    p.user_id,
    p.provider_customer_id,
    p.provider_payment_method_id,
    p.brand,
    p.last4,
    p.exp_month,
    p.exp_year
FROM payment_method_infos p
WHERE p.provider = 'stripe'
ON CONFLICT (stripe_payment_method_id) DO NOTHING;

DROP TABLE IF EXISTS payment_method_infos;
