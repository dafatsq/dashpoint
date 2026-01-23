-- 002_create_permissions.up.sql
-- Create permissions table for granular access control

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on permission key for faster lookups
CREATE INDEX idx_permissions_key ON permissions(key);
CREATE INDEX idx_permissions_category ON permissions(category);

-- Add comment for documentation
COMMENT ON TABLE permissions IS 'Stores permission definitions (can_refund, can_edit_price, etc.)';
