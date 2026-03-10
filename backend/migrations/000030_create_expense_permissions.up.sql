-- 029_create_expense_permissions.up.sql
-- Add granular permissions for expense management

-- Insert new permissions
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('dddddddd-dddd-dddd-dddd-ddddddddddde', 'can_view_expenses', 'View Expenses', 'Can view expenses and expense reports', 'expenses'),
    ('dddddddd-dddd-dddd-dddd-dddddddddddf', 'can_manage_expenses', 'Manage Expenses', 'Can create, edit, and delete expenses and categories', 'expenses');

-- Assign to owner role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM permissions
WHERE key IN ('can_view_expenses', 'can_manage_expenses');

-- Assign to manager role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '22222222-2222-2222-2222-222222222222', id FROM permissions
WHERE key IN ('can_view_expenses', 'can_manage_expenses');
