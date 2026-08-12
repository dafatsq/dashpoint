DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN ('can_manage_manager_permissions', 'can_manage_cashier_permissions')
);

DELETE FROM user_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN ('can_manage_manager_permissions', 'can_manage_cashier_permissions')
);

DELETE FROM permissions WHERE key IN ('can_manage_manager_permissions', 'can_manage_cashier_permissions');
