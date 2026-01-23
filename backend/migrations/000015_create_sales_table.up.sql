-- Sales table for transactions
CREATE TABLE sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_no TEXT UNIQUE NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Totals
    subtotal NUMERIC(15,2) NOT NULL DEFAULT 0,
    tax_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    
    -- Item counts
    item_count INTEGER NOT NULL DEFAULT 0,
    
    -- Payment info
    payment_status TEXT NOT NULL DEFAULT 'pending' CHECK (payment_status IN ('pending', 'partial', 'paid', 'refunded', 'voided')),
    amount_paid NUMERIC(15,2) NOT NULL DEFAULT 0,
    change_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    
    -- Discount details
    discount_type TEXT CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value NUMERIC(15,2),
    discount_reason TEXT,
    
    -- References
    employee_id UUID NOT NULL REFERENCES users(id),
    shift_id UUID REFERENCES shifts(id),
    customer_name TEXT,
    customer_phone TEXT,
    
    -- Status
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('draft', 'completed', 'voided', 'refunded')),
    voided_at TIMESTAMP WITH TIME ZONE,
    voided_by UUID REFERENCES users(id),
    void_reason TEXT,
    
    notes TEXT
);

-- Indexes
CREATE INDEX idx_sales_employee_id ON sales(employee_id);
CREATE INDEX idx_sales_shift_id ON sales(shift_id);
CREATE INDEX idx_sales_created_at ON sales(created_at);
CREATE INDEX idx_sales_status ON sales(status);
CREATE INDEX idx_sales_payment_status ON sales(payment_status);
CREATE INDEX idx_sales_invoice_no ON sales(invoice_no);
