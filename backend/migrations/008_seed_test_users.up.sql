-- 008_seed_test_users.up.sql
-- Seed test user accounts for development

-- Test owner account
-- Email: owner@dashpoint.local
-- Password: owner123
INSERT INTO users (id, email, name, password_hash, pin_hash, role_id, is_active, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'owner@dashpoint.local',
    'Test Owner',
    '$2a$12$iTyvGu2YSGZmv0JRLgxfweDEa97hKZj.Ce0q5REiDSxUxLa5HyjkK',
    NULL,
    '11111111-1111-1111-1111-111111111111',
    true,
    NOW(),
    NOW()
);

-- Test manager account
-- Email: manager@dashpoint.local
-- Password: manager123
-- PIN: 1234
INSERT INTO users (id, email, name, password_hash, pin_hash, role_id, is_active, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'manager@dashpoint.local',
    'Test Manager',
    '$2a$12$5TBVvM3lIG9j5QaVqaLdfOQe294lj9pCpK6glAo92lfI4o2XHaGfq',
    '$2a$12$UhChmcEfkdoC4f17OQe7puzscONOYD/nM1Tt/ppg6a4yb9A6YGIRi',
    '22222222-2222-2222-2222-222222222222',
    true,
    NOW(),
    NOW()
);

-- Test cashier account
-- Email: cashier@dashpoint.local
-- Password: cashier123
-- PIN: 1111
INSERT INTO users (id, email, name, password_hash, pin_hash, role_id, is_active, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    'cashier@dashpoint.local',
    'Test Cashier',
    '$2a$12$8yEtk1Xt4lbaDoOLUMbfaOloDYg4p9OGkoSAegqdxElYuqPg7uvSO',
    '$2a$12$Gkp58amBqTZT71yO.wpAeui04ZPGOicsNogWEYvCLpgG54PXiwlMC',
    '33333333-3333-3333-3333-333333333333',
    true,
    NOW(),
    NOW()
);
