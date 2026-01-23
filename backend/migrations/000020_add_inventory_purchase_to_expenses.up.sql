-- 000020_add_inventory_purchase_to_expenses.up.sql
-- Add inventory purchase support to expenses

-- Add columns for inventory purchases
ALTER TABLE expenses
ADD COLUMN product_id UUID REFERENCES products(id) ON DELETE SET NULL,
ADD COLUMN quantity NUMERIC(15, 3);

-- Add index for product lookups
CREATE INDEX idx_expenses_product_id ON expenses(product_id);

-- Add Inventory Purchase category
INSERT INTO expense_categories (name, description) VALUES
    ('Inventory Purchase', 'Purchase of products for resale - automatically updates inventory');

-- Add comment
COMMENT ON COLUMN expenses.product_id IS 'Product ID for inventory purchases (COGS)';
COMMENT ON COLUMN expenses.quantity IS 'Quantity purchased (for inventory purchases)';
