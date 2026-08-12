ALTER TABLE sale_items
	DROP COLUMN IF EXISTS is_refunded,
	DROP COLUMN IF EXISTS refunded_quantity;
