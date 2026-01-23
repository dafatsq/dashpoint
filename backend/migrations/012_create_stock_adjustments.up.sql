-- 012_create_stock_adjustments.up.sql
-- Create stock_adjustments table for audit trail

CREATE TYPE adjustment_type AS ENUM (
    'initial',
    'purchase',
    'sale',
    'return',
    'adjustment',
    'damage',
    'loss',
    'transfer',
    'count'
);

CREATE TABLE stock_adjustments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    adjustment_type adjustment_type NOT NULL,
    quantity_before NUMERIC(15, 3) NOT NULL,
    quantity_change NUMERIC(15, 3) NOT NULL,
    quantity_after NUMERIC(15, 3) NOT NULL,
    reason TEXT,
    reference_type VARCHAR(50),
    reference_id UUID,
    adjusted_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_stock_adjustments_product_id ON stock_adjustments(product_id);
CREATE INDEX idx_stock_adjustments_type ON stock_adjustments(adjustment_type);
CREATE INDEX idx_stock_adjustments_created_at ON stock_adjustments(created_at);
CREATE INDEX idx_stock_adjustments_adjusted_by ON stock_adjustments(adjusted_by);
CREATE INDEX idx_stock_adjustments_reference ON stock_adjustments(reference_type, reference_id);

-- Add comments
COMMENT ON TABLE stock_adjustments IS 'Audit trail for all stock changes';
COMMENT ON COLUMN stock_adjustments.reference_type IS 'Type of related entity (sale, purchase_order, etc.)';
COMMENT ON COLUMN stock_adjustments.reference_id IS 'ID of the related entity';
