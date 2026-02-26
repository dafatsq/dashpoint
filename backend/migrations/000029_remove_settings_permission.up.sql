-- 000029_remove_settings_permission.up.sql
-- Remove "Manage Settings" permission since settings is now a public page

DELETE FROM role_permissions
WHERE permission_id = 'ffffffff-ffff-ffff-ffff-fffffffffffb';

DELETE FROM permissions
WHERE id = 'ffffffff-ffff-ffff-ffff-fffffffffffb';
