-- 001_create_roles.down.sql
-- Rollback roles table

DROP INDEX IF EXISTS idx_roles_name;
DROP TABLE IF EXISTS roles;
