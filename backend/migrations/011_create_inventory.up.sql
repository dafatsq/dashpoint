-- 011_create_inventory.up.sql
-- Create inventory_items table for stock tracking

CREATE TABLE inventory_items (
    product_id UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    quantity NUMERIC(15, 3) DEFAULT 0,
    reserved_quantity NUMERIC(15, 3) DEFAULT 0,
    low_stock_threshold NUMERIC(15, 3) DEFAULT 0,
    reorder_quantity NUMERIC(15, 3) DEFAULT 0,
    last_counted_at TIMESTAMP WITH TIME ZONE,
    last_restocked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for low stock alerts
CREATE INDEX idx_inventory_low_stock ON inventory_items(quantity, low_stock_threshold) 
    WHERE quantity <= low_stock_threshold;

-- Add comments
COMMENT ON TABLE inventory_items IS 'Inventory levels for products';
COMMENT ON COLUMN inventory_items.quantity IS 'Current quantity in stock';
COMMENT ON COLUMN inventory_items.reserved_quantity IS 'Quantity reserved for pending orders';
COMMENT ON COLUMN inventory_items.low_stock_threshold IS 'Alert when quantity falls below this level';
COMMENT ON COLUMN inventory_items.reorder_quantity IS 'Suggested quantity to reorder';
