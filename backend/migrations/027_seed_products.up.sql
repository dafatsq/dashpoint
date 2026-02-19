-- 027_seed_products.up.sql
-- Seed initial products and inventory

-- Products
INSERT INTO products (id, sku, barcode, name, description, category_id, price, cost, tax_rate, unit, is_active, track_inventory) VALUES
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee1', 'BEV-001', '8991234567890', 'Kopi Susu Gula Aren', 'Es kopi susu dengan gula aren asli', '10000000-0000-0000-0000-000000000002', 18000, 8000, 11, 'cup', true, true),
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee2', 'BEV-002', '8991234567891', 'Teh Manis Jumbo', 'Teh manis dingin ukuran jumbo', '10000000-0000-0000-0000-000000000002', 5000, 2000, 11, 'cup', true, true),
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee3', 'FOOD-001', '8991234567892', 'Roti Bakar Coklat', 'Roti bakar dengan selai coklat premium', '10000000-0000-0000-0000-000000000001', 15000, 6000, 11, 'porsi', true, true),
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee4', 'FOOD-002', '8991234567893', 'Mie Goreng Spesial', 'Mie goreng dengan telur dan ayam', '10000000-0000-0000-0000-000000000001', 20000, 8000, 11, 'porsi', true, true),
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee5', 'TOB-001', '8991234567894', 'Sampoerna Mild 16', 'Rokok Sampoerna Mild isi 16', '10000000-0000-0000-0000-000000000003', 32000, 29000, 11, 'bungkus', true, true),
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee6', 'BEV-003', '8991234567895', 'Air Mineral 600ml', 'Air mineral botol 600ml', '10000000-0000-0000-0000-000000000002', 4000, 2000, 11, 'botol', true, true);

-- Inventory
INSERT INTO inventory_items (product_id, quantity, low_stock_threshold, reorder_quantity) VALUES
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee1', 100, 10, 50), -- Kopi
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee2', 200, 20, 100), -- Teh
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee3', 50, 5, 20), -- Roti
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee4', 50, 5, 20), -- Mie
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee5', 500, 50, 200), -- Rokok
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee6', 100, 20, 50); -- Air Mineral
