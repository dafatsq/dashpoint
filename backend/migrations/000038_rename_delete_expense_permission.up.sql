-- 000038_rename_delete_expense_permission.up.sql
-- Rename the expense delete permission to use delete-only wording

UPDATE permissions
SET
    name = 'Delete Expense',
    description = 'Can permanently delete expense records'
WHERE key = 'can_delete_expenses';