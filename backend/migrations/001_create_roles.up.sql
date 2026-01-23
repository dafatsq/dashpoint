-- 001_create_roles.up.sql
-- Create roles table for RBAC

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on role name for faster lookups
CREATE INDEX idx_roles_name ON roles(name);

-- Add comment for documentation
COMMENT ON TABLE roles IS 'Stores role definitions for RBAC (owner, manager, cashier)';
