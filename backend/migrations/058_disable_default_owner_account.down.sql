-- 058_disable_default_owner_account.down.sql
-- Restore the legacy default owner account to active if this migration had
-- deactivated it (i.e., it still carries the original bootstrap credentials).

UPDATE users
SET is_active = true,
    updated_at = NOW()
WHERE id = '00000000-0000-0000-0000-000000000001'
  AND email = 'owner@dashpoint.local'
  AND password_hash = '$2a$12$iTyvGu2YSGZmv0JRLgxfweDEa97hKZj.Ce0q5REiDSxUxLa5HyjkK'
  AND pin_hash = '$2a$12$UhChmcEfkdoC4f17OQe7puzscONOYD/nM1Tt/ppg6a4yb9A6YGIRi';
