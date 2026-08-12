ALTER TABLE categories
ADD COLUMN IF NOT EXISTS parent_id uuid;
