-- 033_add_categories_permission.down.sql
-- Remove can_manage_categories permission

DELETE FROM role_permissions WHERE permission_id = 'cccccccc-cccc-cccc-cccc-cccccccccccf';
DELETE FROM user_permissions WHERE permission_id = 'cccccccc-cccc-cccc-cccc-cccccccccccf';
DELETE FROM permissions WHERE id = 'cccccccc-cccc-cccc-cccc-cccccccccccf';
