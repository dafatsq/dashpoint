-- 006_create_refresh_tokens.up.sql
-- Create refresh_tokens table for JWT session management

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_reason VARCHAR(255)
);

-- Create indexes for faster lookups
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Add comment for documentation
COMMENT ON TABLE refresh_tokens IS 'Stores refresh tokens for JWT session management';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'SHA-256 hash of the refresh token';
COMMENT ON COLUMN refresh_tokens.revoked_at IS 'When the token was revoked (null if still valid)';
