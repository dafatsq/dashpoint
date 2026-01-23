-- 000020_add_inventory_purchase_to_expenses.down.sql
-- Remove inventory purchase support from expenses

-- Remove Inventory Purchase category
DELETE FROM expense_categories WHERE name = 'Inventory Purchase';

-- Drop columns
ALTER TABLE expenses
DROP COLUMN IF EXISTS quantity,
DROP COLUMN IF EXISTS product_id;

-- Index will be dropped automatically with column
