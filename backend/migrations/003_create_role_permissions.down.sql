-- 003_create_role_permissions.down.sql
-- Rollback role_permissions table

DROP INDEX IF EXISTS idx_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_role_permissions_role_id;
DROP TABLE IF EXISTS role_permissions;
