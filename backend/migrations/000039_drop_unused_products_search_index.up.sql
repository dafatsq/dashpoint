-- 000039_drop_unused_products_search_index.up.sql
-- Remove the unused full-text index because product search is substring-based (ILIKE), not tsvector-based.

DROP INDEX IF EXISTS idx_products_search;
