ALTER TABLE expense_categories
ADD COLUMN system_key VARCHAR(100);

UPDATE expense_categories
SET system_key = 'inventory_purchase'
WHERE name = 'Inventory Purchase' AND system_key IS NULL;

CREATE UNIQUE INDEX idx_expense_categories_system_key_unique
ON expense_categories(system_key)
WHERE system_key IS NOT NULL;
