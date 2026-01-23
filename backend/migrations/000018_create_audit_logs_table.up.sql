-- Audit logs table for tracking all important actions
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- When
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Who
    user_id UUID REFERENCES users(id),
    user_email TEXT,
    user_name TEXT,
    user_role TEXT,
    
    -- What
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    
    -- Details
    description TEXT,
    old_values JSONB,
    new_values JSONB,
    metadata JSONB,
    
    -- Where
    ip_address TEXT,
    user_agent TEXT,
    request_id TEXT,
    
    -- Status
    status TEXT NOT NULL DEFAULT 'success' CHECK (status IN ('success', 'failure', 'warning'))
);

-- Indexes for common queries
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);

-- Composite index for filtering
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_user_action ON audit_logs(user_id, action);
