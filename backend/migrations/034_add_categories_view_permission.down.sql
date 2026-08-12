-- 034_add_categories_view_permission.down.sql
-- Remove can_view_categories permission and its associations

DELETE FROM role_permissions WHERE permission_id = 'cccccccc-cccc-cccc-cccc-ccccccccccc0';
DELETE FROM user_permissions WHERE permission_id = 'cccccccc-cccc-cccc-cccc-ccccccccccc0';
DELETE FROM permissions WHERE id = 'cccccccc-cccc-cccc-cccc-ccccccccccc0';
