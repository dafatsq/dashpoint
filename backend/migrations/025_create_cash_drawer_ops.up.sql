-- Cash drawer operations for mid-shift pay-in/pay-out tracking
CREATE TABLE cash_drawer_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shift_id UUID NOT NULL REFERENCES shifts(id),
    type TEXT NOT NULL CHECK (type IN ('pay_in', 'pay_out')),
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    reason TEXT NOT NULL,
    performed_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_cash_drawer_ops_shift_id ON cash_drawer_operations(shift_id);
CREATE INDEX idx_cash_drawer_ops_performed_by ON cash_drawer_operations(performed_by);
