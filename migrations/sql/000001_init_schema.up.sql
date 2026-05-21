CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    avatar TEXT,
    stripe_customer_id TEXT,
    default_payment_method_id TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

CREATE TABLE IF NOT EXISTS grinds (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    duration INTEGER NOT NULL,
    budget INTEGER NOT NULL,
    start_date TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_grinds_deleted_at ON grinds (deleted_at);

CREATE TABLE IF NOT EXISTS participate_records (
    grind_schema_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (grind_schema_id, user_id),
    CONSTRAINT fk_participate_records_grind_schema FOREIGN KEY (grind_schema_id) REFERENCES grinds (id),
    CONSTRAINT fk_participate_records_user FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    task_type TEXT NOT NULL,
    user_id TEXT NOT NULL,
    grind_id TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    finished_time TIMESTAMPTZ,
    completed BOOLEAN DEFAULT FALSE,
    problem_title TEXT,
    problem_description TEXT,
    problem_url TEXT,
    problem_difficulty TEXT,
    problem_topic_tags JSONB,
    code TEXT,
    code_language TEXT
);

CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks (deleted_at);

CREATE TABLE IF NOT EXISTS participation (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id TEXT NOT NULL,
    grind_id TEXT NOT NULL,
    missed_days BIGINT NOT NULL DEFAULT 0,
    total_penalty BIGINT NOT NULL DEFAULT 0,
    quitted BOOLEAN NOT NULL DEFAULT FALSE,
    quitted_at TIMESTAMPTZ,
    CONSTRAINT fk_participation_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_participation_grind FOREIGN KEY (grind_id) REFERENCES grinds (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_participation_deleted_at ON participation (deleted_at);

CREATE TABLE IF NOT EXISTS message (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    sender_id TEXT NOT NULL,
    receiver_id TEXT NOT NULL,
    content TEXT NOT NULL,
    type TEXT NOT NULL,
    invitation_grind_id TEXT,
    invitation_accepted BOOLEAN,
    invitation_rejected BOOLEAN,
    read BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_message_deleted_at ON message (deleted_at);

CREATE TABLE IF NOT EXISTS interview_sessions (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    status TEXT NOT NULL,
    conversation_history JSONB,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_interview_sessions_deleted_at ON interview_sessions (deleted_at);

CREATE TABLE IF NOT EXISTS stripe_payment_info (
    id BIGSERIAL,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id TEXT NOT NULL,
    stripe_customer_id TEXT NOT NULL,
    stripe_payment_method_id TEXT NOT NULL,
    brand TEXT,
    last4 TEXT,
    exp_month BIGINT,
    exp_year BIGINT,
    PRIMARY KEY (id, stripe_payment_method_id),
    CONSTRAINT uni_stripe_payment_info_stripe_payment_method_id UNIQUE (stripe_payment_method_id)
);

CREATE INDEX IF NOT EXISTS idx_stripe_payment_info_deleted_at ON stripe_payment_info (deleted_at);
