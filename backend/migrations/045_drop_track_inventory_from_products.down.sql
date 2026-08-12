ALTER TABLE products
ADD COLUMN IF NOT EXISTS track_inventory boolean NOT NULL DEFAULT true;
