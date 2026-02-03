-- 023_simplify_permissions.up.sql
-- Remove edit_price and backup_data permissions, simplify sales permissions

-- Remove can_edit_price permission and its associations
DELETE FROM user_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE key = 'can_edit_price'
);

DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE key = 'can_edit_price'
);

DELETE FROM permissions WHERE key = 'can_edit_price';

-- Remove can_backup_data permission and its associations
DELETE FROM user_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE key = 'can_backup_data'
);

DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE key = 'can_backup_data'
);

DELETE FROM permissions WHERE key = 'can_backup_data';

-- Update sales permissions to be more accurate
-- Remove granular sales permissions
DELETE FROM user_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN ('can_apply_discount', 'can_refund', 'can_void_sale')
);

DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN ('can_apply_discount', 'can_refund', 'can_void_sale')
);

DELETE FROM permissions WHERE key IN ('can_apply_discount', 'can_refund', 'can_void_sale');

-- Update can_create_sale to be more general
UPDATE permissions 
SET name = 'Process Sales', 
    description = 'Can create sales, apply discounts, process refunds, and void transactions'
WHERE key = 'can_create_sale';
