-- 007_seed_roles_permissions.down.sql
-- Remove seeded data

DELETE FROM role_permissions;
DELETE FROM permissions;
DELETE FROM roles;
