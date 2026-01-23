-- 005_create_user_permissions.down.sql
-- Rollback user_permissions table

DROP INDEX IF EXISTS idx_user_permissions_permission_id;
DROP INDEX IF EXISTS idx_user_permissions_user_id;
DROP TABLE IF EXISTS user_permissions;
