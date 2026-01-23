-- 004_create_users.down.sql
-- Rollback users table

DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_role_id;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
