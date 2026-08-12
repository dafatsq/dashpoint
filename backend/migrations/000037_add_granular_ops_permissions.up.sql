-- 000037_add_granular_ops_permissions.up.sql
-- Add granular permissions for expenses, categories, and inventory operations

-- Insert granular permissions
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('f1000000-0000-0000-0000-000000000001', 'can_create_expenses', 'Create Expenses', 'Can create expense records', 'expenses'),
    ('f1000000-0000-0000-0000-000000000002', 'can_edit_expenses', 'Edit Expenses', 'Can edit expense records', 'expenses'),
    ('f1000000-0000-0000-0000-000000000003', 'can_delete_expenses', 'Delete/archive Expenses', 'Can delete/archive expense records', 'expenses'),
    ('f1000000-0000-0000-0000-000000000004', 'can_create_categories', 'Create Categories', 'Can create product and expense categories', 'categories'),
    ('f1000000-0000-0000-0000-000000000005', 'can_edit_categories', 'Edit Categories', 'Can edit product and expense categories', 'categories'),
    ('f1000000-0000-0000-0000-000000000006', 'can_delete_categories', 'Delete/archive Categories', 'Can delete/archive product and expense categories', 'categories'),
    ('f1000000-0000-0000-0000-000000000007', 'can_add_stock', 'Add Stock', 'Can increase stock quantities', 'inventory'),
    ('f1000000-0000-0000-0000-000000000008', 'can_remove_stock', 'Remove Stock', 'Can decrease stock quantities', 'inventory'),
    ('f1000000-0000-0000-0000-000000000009', 'can_adjust_stock', 'Adjust Stock', 'Can reconcile stock counts and correction adjustments', 'inventory')
ON CONFLICT (key) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category;

-- Map existing role permission grants to the new expense permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p_new.id
FROM role_permissions rp
JOIN permissions p_old ON rp.permission_id = p_old.id
JOIN permissions p_new ON p_new.key IN ('can_create_expenses', 'can_edit_expenses', 'can_delete_expenses')
WHERE p_old.key = 'can_manage_expenses'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Map existing role permission grants to the new category permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p_new.id
FROM role_permissions rp
JOIN permissions p_old ON rp.permission_id = p_old.id
JOIN permissions p_new ON p_new.key IN ('can_create_categories', 'can_edit_categories', 'can_delete_categories')
WHERE p_old.key = 'can_manage_categories'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Map existing role permission grants to the new inventory permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p_new.id
FROM role_permissions rp
JOIN permissions p_old ON rp.permission_id = p_old.id
JOIN permissions p_new ON p_new.key IN ('can_add_stock', 'can_remove_stock', 'can_adjust_stock')
WHERE p_old.key = 'can_edit_inventory'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Copy existing user overrides from coarse expense permission to granular permissions
INSERT INTO user_permissions (user_id, permission_id, allowed)
SELECT up.user_id, p_new.id, up.allowed
FROM user_permissions up
JOIN permissions p_old ON up.permission_id = p_old.id
JOIN permissions p_new ON p_new.key IN ('can_create_expenses', 'can_edit_expenses', 'can_delete_expenses')
WHERE p_old.key = 'can_manage_expenses'
ON CONFLICT (user_id, permission_id) DO UPDATE SET
    allowed = EXCLUDED.allowed;

-- Copy existing user overrides from coarse category permission to granular permissions
INSERT INTO user_permissions (user_id, permission_id, allowed)
SELECT up.user_id, p_new.id, up.allowed
FROM user_permissions up
JOIN permissions p_old ON up.permission_id = p_old.id
JOIN permissions p_new ON p_new.key IN ('can_create_categories', 'can_edit_categories', 'can_delete_categories')
WHERE p_old.key = 'can_manage_categories'
ON CONFLICT (user_id, permission_id) DO UPDATE SET
    allowed = EXCLUDED.allowed;

-- Copy existing user overrides from coarse inventory permission to granular permissions
INSERT INTO user_permissions (user_id, permission_id, allowed)
SELECT up.user_id, p_new.id, up.allowed
FROM user_permissions up
JOIN permissions p_old ON up.permission_id = p_old.id
JOIN permissions p_new ON p_new.key IN ('can_add_stock', 'can_remove_stock', 'can_adjust_stock')
WHERE p_old.key = 'can_edit_inventory'
ON CONFLICT (user_id, permission_id) DO UPDATE SET
    allowed = EXCLUDED.allowed;
