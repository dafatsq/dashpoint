DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN (
        'can_create_manager_users', 'can_create_cashier_users',
        'can_edit_manager_users', 'can_edit_cashier_users',
        'can_delete_manager_users', 'can_delete_cashier_users'
    )
);

DELETE FROM permissions WHERE key IN (
    'can_create_manager_users', 'can_create_cashier_users',
    'can_edit_manager_users', 'can_edit_cashier_users',
    'can_delete_manager_users', 'can_delete_cashier_users'
);
