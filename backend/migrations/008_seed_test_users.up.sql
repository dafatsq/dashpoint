-- 008_seed_test_users.up.sql
-- Seed owner account for starter pack

-- Owner account
-- Email: owner@dashpoint.local
-- Password: owner123
-- PIN: 1234
INSERT INTO users (id, email, name, password_hash, pin_hash, role_id, is_active, created_at, updated_at) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'owner@dashpoint.local',
    'Owner',
    '$2a$12$iTyvGu2YSGZmv0JRLgxfweDEa97hKZj.Ce0q5REiDSxUxLa5HyjkK',
    '$2a$12$UhChmcEfkdoC4f17OQe7puzscONOYD/nM1Tt/ppg6a4yb9A6YGIRi',
    '11111111-1111-1111-1111-111111111111',
    true,
    NOW(),
    NOW()
);
