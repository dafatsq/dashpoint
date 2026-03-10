-- 029_create_expense_permissions.down.sql
-- Remove granular permissions for expense management

-- Remove from role_permissions first (cascading deletes usually handle this but explicit is better)
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN ('can_view_expenses', 'can_manage_expenses')
);

-- Delete permissions
DELETE FROM permissions
WHERE key IN ('can_view_expenses', 'can_manage_expenses');
