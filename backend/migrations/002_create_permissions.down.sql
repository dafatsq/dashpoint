-- 002_create_permissions.down.sql
-- Rollback permissions table

DROP INDEX IF EXISTS idx_permissions_category;
DROP INDEX IF EXISTS idx_permissions_key;
DROP TABLE IF EXISTS permissions;
