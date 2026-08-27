ALTER TABLE audit_logs
	DROP COLUMN IF EXISTS ip_address,
	DROP COLUMN IF EXISTS user_agent,
	DROP COLUMN IF EXISTS request_id;
