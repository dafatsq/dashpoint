-- 009_create_categories.down.sql
DROP INDEX IF EXISTS idx_categories_name;
DROP INDEX IF EXISTS idx_categories_is_active;
DROP INDEX IF EXISTS idx_categories_parent_id;
DROP TABLE IF EXISTS categories;
