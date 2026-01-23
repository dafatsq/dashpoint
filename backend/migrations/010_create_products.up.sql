-- 010_create_products.up.sql
-- Create products table

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(100) UNIQUE,
    barcode VARCHAR(100) UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    cost NUMERIC(15, 2) DEFAULT 0,
    tax_rate NUMERIC(5, 2) DEFAULT 0,
    unit VARCHAR(50) DEFAULT 'pcs',
    is_active BOOLEAN DEFAULT true,
    track_inventory BOOLEAN DEFAULT true,
    allow_negative_stock BOOLEAN DEFAULT false,
    image_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for fast lookups
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_barcode ON products(barcode);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_is_active ON products(is_active);

-- Full text search index for product search
CREATE INDEX idx_products_search ON products USING gin(to_tsvector('simple', name || ' ' || COALESCE(description, '') || ' ' || COALESCE(sku, '') || ' ' || COALESCE(barcode, '')));

-- Add comments
COMMENT ON TABLE products IS 'Product catalog';
COMMENT ON COLUMN products.sku IS 'Stock Keeping Unit - internal product code';
COMMENT ON COLUMN products.barcode IS 'Barcode for scanning (EAN, UPC, etc.)';
COMMENT ON COLUMN products.price IS 'Selling price in IDR';
COMMENT ON COLUMN products.cost IS 'Cost/purchase price in IDR';
COMMENT ON COLUMN products.tax_rate IS 'Tax rate percentage (e.g., 11 for 11%)';
COMMENT ON COLUMN products.unit IS 'Unit of measure (pcs, kg, liter, etc.)';
