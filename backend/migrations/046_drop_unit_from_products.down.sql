ALTER TABLE products
ADD COLUMN IF NOT EXISTS unit character varying NOT NULL DEFAULT 'pcs';
