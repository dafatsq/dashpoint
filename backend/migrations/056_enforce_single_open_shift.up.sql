DROP INDEX IF EXISTS idx_shifts_employee_open;

CREATE UNIQUE INDEX idx_shifts_single_open
ON shifts (status)
WHERE status = 'open';
