ALTER TABLE inventory_items
DROP COLUMN IF EXISTS reserved_quantity,
DROP COLUMN IF EXISTS reorder_quantity,
DROP COLUMN IF EXISTS last_counted_at,
DROP COLUMN IF EXISTS last_restocked_at,
DROP COLUMN IF EXISTS created_at;
