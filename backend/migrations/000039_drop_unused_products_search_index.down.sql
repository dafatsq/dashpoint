-- 000039_drop_unused_products_search_index.down.sql
-- Restore the legacy full-text index if this migration is rolled back.

CREATE INDEX IF NOT EXISTS idx_products_search
ON products
USING gin (
    to_tsvector(
        'simple',
        name || ' ' || COALESCE(description, '') || ' ' || COALESCE(sku, '') || ' ' || COALESCE(barcode, '')
    )
);
