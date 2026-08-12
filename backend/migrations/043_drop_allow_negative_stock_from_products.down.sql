ALTER TABLE products
ADD COLUMN IF NOT EXISTS allow_negative_stock boolean NOT NULL DEFAULT false;
