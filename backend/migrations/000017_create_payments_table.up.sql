-- Payments table for payment records
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    
    -- Payment method
    payment_method TEXT NOT NULL CHECK (payment_method IN ('cash', 'card', 'transfer', 'qris', 'credit', 'voucher', 'other')),
    
    -- Amounts
    amount NUMERIC(15,2) NOT NULL,
    
    -- Cash specific
    amount_tendered NUMERIC(15,2),
    change_given NUMERIC(15,2),
    
    -- Card specific
    card_type TEXT,
    card_last_four TEXT,
    reference_no TEXT,
    
    -- Transfer/QRIS specific
    bank_name TEXT,
    account_no TEXT,
    
    -- Voucher specific
    voucher_code TEXT,
    
    -- Status
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    
    -- Metadata
    notes TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_by UUID REFERENCES users(id)
);

-- Indexes
CREATE INDEX idx_payments_sale_id ON payments(sale_id);
CREATE INDEX idx_payments_payment_method ON payments(payment_method);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_status ON payments(status);
