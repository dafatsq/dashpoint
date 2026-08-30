-- Invalidate all previously issued access tokens when credentials change
-- (audit design observation: instant revocation via token_version).
ALTER TABLE users
	ADD COLUMN token_version INTEGER NOT NULL DEFAULT 0;
