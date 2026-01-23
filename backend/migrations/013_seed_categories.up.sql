-- 013_seed_categories.up.sql
-- Seed initial product categories

INSERT INTO categories (id, name, description, sort_order) VALUES
    ('10000000-0000-0000-0000-000000000001', 'Makanan', 'Produk makanan dan snack', 1),
    ('10000000-0000-0000-0000-000000000002', 'Minuman', 'Minuman botol, kaleng, dan sachet', 2),
    ('10000000-0000-0000-0000-000000000003', 'Rokok', 'Rokok dan produk tembakau', 3),
    ('10000000-0000-0000-0000-000000000004', 'Sembako', 'Bahan pokok rumah tangga', 4),
    ('10000000-0000-0000-0000-000000000005', 'Kebersihan', 'Produk kebersihan dan toiletries', 5),
    ('10000000-0000-0000-0000-000000000006', 'Obat-obatan', 'Obat bebas dan vitamin', 6),
    ('10000000-0000-0000-0000-000000000007', 'Lainnya', 'Produk lain-lain', 99);
