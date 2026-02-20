DROP INDEX IF EXISTS idx_shifts_closed_by;
ALTER TABLE shifts DROP COLUMN IF EXISTS closed_by;
