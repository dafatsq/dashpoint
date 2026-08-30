-- 058_disable_default_owner_account.up.sql
-- Neutralize the legacy seeded default owner account on databases that
-- received migration 008 before it was removed. The account is only touched
-- when it still carries the published bootstrap credentials
-- (password "owner123" / PIN "1234"); installations whose operator already
-- rotated the credentials into a real account are left untouched.
--
-- Fresh deployments never have this account (008 is now a no-op) and instead
-- create their first owner through POST /api/v1/setup/owner.

UPDATE users
SET is_active = false,
    updated_at = NOW()
WHERE id = '00000000-0000-0000-0000-000000000001'
  AND email = 'owner@dashpoint.local'
  AND password_hash = '$2a$12$iTyvGu2YSGZmv0JRLgxfweDEa97hKZj.Ce0q5REiDSxUxLa5HyjkK'
  AND pin_hash = '$2a$12$UhChmcEfkdoC4f17OQe7puzscONOYD/nM1Tt/ppg6a4yb9A6YGIRi';

DELETE FROM refresh_tokens
WHERE user_id IN (
    SELECT id
    FROM users
    WHERE id = '00000000-0000-0000-0000-000000000001'
      AND email = 'owner@dashpoint.local'
      AND password_hash = '$2a$12$iTyvGu2YSGZmv0JRLgxfweDEa97hKZj.Ce0q5REiDSxUxLa5HyjkK'
      AND pin_hash = '$2a$12$UhChmcEfkdoC4f17OQe7puzscONOYD/nM1Tt/ppg6a4yb9A6YGIRi'
);
