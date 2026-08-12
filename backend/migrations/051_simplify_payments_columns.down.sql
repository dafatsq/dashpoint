ALTER TABLE payments
	ADD COLUMN card_type text,
	ADD COLUMN card_last_four text,
	ADD COLUMN bank_name text,
	ADD COLUMN account_no text,
	ADD COLUMN voucher_code text,
	ADD COLUMN notes text,
	ADD COLUMN processed_by uuid;
