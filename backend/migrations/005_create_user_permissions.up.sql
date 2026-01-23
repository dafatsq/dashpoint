-- 005_create_user_permissions.up.sql
-- Create user_permissions table for per-user permission overrides

CREATE TABLE user_permissions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    allowed BOOLEAN NOT NULL,
    granted_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, permission_id)
);

-- Create indexes for faster lookups
CREATE INDEX idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX idx_user_permissions_permission_id ON user_permissions(permission_id);

-- Add comment for documentation
COMMENT ON TABLE user_permissions IS 'Per-user permission overrides (allows or denies specific permissions)';
COMMENT ON COLUMN user_permissions.allowed IS 'true = explicitly allowed, false = explicitly denied';
COMMENT ON COLUMN user_permissions.granted_by IS 'User who granted/denied this permission';
