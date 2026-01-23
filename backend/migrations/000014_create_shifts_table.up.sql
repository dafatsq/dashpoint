-- Shifts table for cashier shift management
CREATE TABLE shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES users(id),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    
    -- Cash drawer tracking
    opening_cash NUMERIC(15,2) NOT NULL DEFAULT 0,
    closing_cash NUMERIC(15,2),
    expected_cash NUMERIC(15,2),
    cash_difference NUMERIC(15,2),
    
    -- Shift summary
    total_sales NUMERIC(15,2) DEFAULT 0,
    total_refunds NUMERIC(15,2) DEFAULT 0,
    transaction_count INTEGER DEFAULT 0,
    refund_count INTEGER DEFAULT 0,
    
    -- Status
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'suspended')),
    notes TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_shifts_employee_id ON shifts(employee_id);
CREATE INDEX idx_shifts_status ON shifts(status);
CREATE INDEX idx_shifts_started_at ON shifts(started_at);

-- Only one open shift per employee at a time
CREATE UNIQUE INDEX idx_shifts_employee_open ON shifts(employee_id) WHERE status = 'open';
