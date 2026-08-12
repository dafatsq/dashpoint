DROP INDEX IF EXISTS idx_expense_categories_system_key_unique;

ALTER TABLE expense_categories
DROP COLUMN IF EXISTS system_key;
