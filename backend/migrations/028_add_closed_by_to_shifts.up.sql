ALTER TABLE shifts ADD COLUMN closed_by UUID REFERENCES users(id);

CREATE INDEX idx_shifts_closed_by ON shifts(closed_by);
