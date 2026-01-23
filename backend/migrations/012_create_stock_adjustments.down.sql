-- 012_create_stock_adjustments.down.sql
DROP INDEX IF EXISTS idx_stock_adjustments_reference;
DROP INDEX IF EXISTS idx_stock_adjustments_adjusted_by;
DROP INDEX IF EXISTS idx_stock_adjustments_created_at;
DROP INDEX IF EXISTS idx_stock_adjustments_type;
DROP INDEX IF EXISTS idx_stock_adjustments_product_id;
DROP TABLE IF EXISTS stock_adjustments;
DROP TYPE IF EXISTS adjustment_type;
