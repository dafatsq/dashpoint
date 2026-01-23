-- Sale items table for line items
CREATE TABLE sale_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    
    -- Product snapshot at time of sale
    product_name TEXT NOT NULL,
    product_sku TEXT,
    product_barcode TEXT,
    
    -- Quantities and pricing
    quantity NUMERIC(12,3) NOT NULL,
    unit_price NUMERIC(15,2) NOT NULL,
    cost_price NUMERIC(15,2) NOT NULL DEFAULT 0,
    
    -- Discounts
    discount_type TEXT CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value NUMERIC(15,2) DEFAULT 0,
    discount_amount NUMERIC(15,2) DEFAULT 0,
    
    -- Tax
    tax_rate NUMERIC(5,2) DEFAULT 0,
    tax_amount NUMERIC(15,2) DEFAULT 0,
    
    -- Totals
    subtotal NUMERIC(15,2) NOT NULL,
    total NUMERIC(15,2) NOT NULL,
    
    -- Status
    is_refunded BOOLEAN DEFAULT FALSE,
    refunded_quantity NUMERIC(12,3) DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_sale_items_sale_id ON sale_items(sale_id);
CREATE INDEX idx_sale_items_product_id ON sale_items(product_id);
