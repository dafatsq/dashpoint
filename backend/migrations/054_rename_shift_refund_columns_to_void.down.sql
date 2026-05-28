ALTER TABLE shifts
RENAME COLUMN total_voided TO total_refunds;

ALTER TABLE shifts
RENAME COLUMN void_count TO refund_count;
