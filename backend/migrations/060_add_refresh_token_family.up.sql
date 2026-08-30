-- Group refresh tokens minted from the same login (and their rotation
-- descendants) into a family, so reuse of any revoked token lets the server
-- contain the compromise by revoking every live token in that family.
-- Pre-existing rows each become singleton families via the default.
ALTER TABLE refresh_tokens
	ADD COLUMN family_id UUID NOT NULL DEFAULT gen_random_uuid();

CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);
