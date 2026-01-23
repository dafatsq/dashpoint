-- 011_create_inventory.down.sql
DROP INDEX IF EXISTS idx_inventory_low_stock;
DROP TABLE IF EXISTS inventory_items;
