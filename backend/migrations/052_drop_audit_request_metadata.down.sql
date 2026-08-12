ALTER TABLE audit_logs
	ADD COLUMN ip_address text,
	ADD COLUMN user_agent text,
	ADD COLUMN request_id text;
