-- 1. Rename the existing tasks table to habit_tasks
ALTER TABLE tasks RENAME TO habit_tasks;

-- 2. Add the metadata column to habit_tasks (existing columns preserved for v1 endpoint compatibility)
ALTER TABLE habit_tasks ADD COLUMN IF NOT EXISTS metadata JSONB;

-- 3. Create the completion_events table
CREATE TABLE IF NOT EXISTS completion_events (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    habit_task_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    metadata JSONB,
    CONSTRAINT fk_completion_events_habit_task FOREIGN KEY (habit_task_id) REFERENCES habit_tasks (id) ON DELETE CASCADE,
    CONSTRAINT fk_completion_events_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_completion_events_deleted_at ON completion_events (deleted_at);
CREATE INDEX IF NOT EXISTS idx_completion_events_user_id ON completion_events (user_id);
CREATE INDEX IF NOT EXISTS idx_completion_events_habit_task_id ON completion_events (habit_task_id);

-- 4. Create the partner_groups table
CREATE TABLE IF NOT EXISTS partner_groups (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    invite_token TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    grind_id TEXT NOT NULL,
    CONSTRAINT fk_partner_groups_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT fk_partner_groups_grind FOREIGN KEY (grind_id) REFERENCES grinds (id),
    CONSTRAINT uni_partner_groups_invite_token UNIQUE (invite_token),
    CONSTRAINT uni_partner_groups_grind_id UNIQUE (grind_id)
);

CREATE INDEX IF NOT EXISTS idx_partner_groups_deleted_at ON partner_groups (deleted_at);

-- 5. Create the group_members join table
CREATE TABLE IF NOT EXISTS group_members (
    partner_group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (partner_group_id, user_id),
    CONSTRAINT fk_group_members_group FOREIGN KEY (partner_group_id) REFERENCES partner_groups (id) ON DELETE CASCADE,
    CONSTRAINT fk_group_members_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- 6. Add partner_group_id column to grinds
ALTER TABLE grinds ADD COLUMN IF NOT EXISTS partner_group_id TEXT REFERENCES partner_groups (id);
