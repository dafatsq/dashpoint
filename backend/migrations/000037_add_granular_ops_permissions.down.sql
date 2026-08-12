-- 000037_add_granular_ops_permissions.down.sql
-- Remove granular permissions for expenses, categories, and inventory operations

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE key IN (
        'can_create_expenses',
        'can_edit_expenses',
        'can_delete_expenses',
        'can_create_categories',
        'can_edit_categories',
        'can_delete_categories',
        'can_add_stock',
        'can_remove_stock',
        'can_adjust_stock'
    )
);

DELETE FROM user_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE key IN (
        'can_create_expenses',
        'can_edit_expenses',
        'can_delete_expenses',
        'can_create_categories',
        'can_edit_categories',
        'can_delete_categories',
        'can_add_stock',
        'can_remove_stock',
        'can_adjust_stock'
    )
);

DELETE FROM permissions
WHERE key IN (
    'can_create_expenses',
    'can_edit_expenses',
    'can_delete_expenses',
    'can_create_categories',
    'can_edit_categories',
    'can_delete_categories',
    'can_add_stock',
    'can_remove_stock',
    'can_adjust_stock'
);
