-- 010_create_products.down.sql
DROP INDEX IF EXISTS idx_products_search;
DROP INDEX IF EXISTS idx_products_is_active;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_name;
DROP INDEX IF EXISTS idx_products_barcode;
DROP INDEX IF EXISTS idx_products_sku;
DROP TABLE IF EXISTS products;
