ALTER TABLE sale_items
	ADD COLUMN is_refunded boolean NOT NULL DEFAULT false,
	ADD COLUMN refunded_quantity numeric NOT NULL DEFAULT 0;
