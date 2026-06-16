-- Reverse migration 000005 in dependency order

-- 1. Remove partner_group_id from grinds
ALTER TABLE grinds DROP COLUMN IF EXISTS partner_group_id;

-- 2. Drop group_members join table
DROP TABLE IF EXISTS group_members;

-- 3. Drop partner_groups table
DROP TABLE IF EXISTS partner_groups;

-- 4. Drop completion_events table
DROP TABLE IF EXISTS completion_events;

-- 5. Remove metadata column from habit_tasks
ALTER TABLE habit_tasks DROP COLUMN IF EXISTS metadata;

-- 6. Rename habit_tasks back to tasks
ALTER TABLE habit_tasks RENAME TO tasks;
