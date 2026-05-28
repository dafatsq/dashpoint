ALTER TABLE shifts
RENAME COLUMN total_refunds TO total_voided;

ALTER TABLE shifts
RENAME COLUMN refund_count TO void_count;
