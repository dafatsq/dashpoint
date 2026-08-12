-- 000038_rename_delete_expense_permission.down.sql
-- Restore the previous expense delete permission wording

UPDATE permissions
SET
    name = 'Delete/archive Expenses',
    description = 'Can delete/archive expense records'
WHERE key = 'can_delete_expenses';
