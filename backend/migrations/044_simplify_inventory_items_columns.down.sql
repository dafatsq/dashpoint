ALTER TABLE inventory_items
ADD COLUMN IF NOT EXISTS reserved_quantity numeric NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS reorder_quantity numeric NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_counted_at timestamp with time zone,
ADD COLUMN IF NOT EXISTS last_restocked_at timestamp with time zone,
ADD COLUMN IF NOT EXISTS created_at timestamp with time zone NOT NULL DEFAULT now();
