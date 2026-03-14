-- 008_seed_test_users.down.sql
-- Remove test user accounts

DELETE FROM refresh_tokens WHERE user_id IN (
    '00000000-0000-0000-0000-000000000001'
);

DELETE FROM user_permissions WHERE user_id IN (
    '00000000-0000-0000-0000-000000000001'
);

DELETE FROM users WHERE id IN (
    '00000000-0000-0000-0000-000000000001'
);
