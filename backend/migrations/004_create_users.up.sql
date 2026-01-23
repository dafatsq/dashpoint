-- 004_create_users.up.sql
-- Create users table for employees and owner

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    pin_hash VARCHAR(255),
    role_id UUID NOT NULL REFERENCES roles(id),
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Add comment for documentation
COMMENT ON TABLE users IS 'Stores user accounts (owner, managers, cashiers)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password for email login';
COMMENT ON COLUMN users.pin_hash IS 'Bcrypt hashed PIN for quick cashier login';
