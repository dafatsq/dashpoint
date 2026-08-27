-- Restore request-source metadata on audit entries (undoes part of 052).
-- Needed so cash-handling actions and auth failures can be tied to a terminal.
ALTER TABLE audit_logs
	ADD COLUMN IF NOT EXISTS ip_address TEXT,
	ADD COLUMN IF NOT EXISTS user_agent TEXT,
	ADD COLUMN IF NOT EXISTS request_id TEXT;
