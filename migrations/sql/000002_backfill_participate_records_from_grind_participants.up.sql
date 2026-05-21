DO $$
BEGIN
    IF to_regclass('public.grind_participants') IS NOT NULL
       AND to_regclass('public.participate_records') IS NOT NULL THEN
        INSERT INTO participate_records (grind_schema_id, user_id)
        SELECT gp.grind_id, gp.user_id
        FROM grind_participants gp
        WHERE gp.grind_id IS NOT NULL
          AND gp.user_id IS NOT NULL
        ON CONFLICT (grind_schema_id, user_id) DO NOTHING;
    END IF;
END $$;
