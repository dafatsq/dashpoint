-- 013_seed_categories.up.sql
-- Seed initial product categories

INSERT INTO categories (id, name, description, sort_order) VALUES
    ('10000000-0000-0000-0000-000000000001', 'Food', 'Food products and snacks', 1),
    ('10000000-0000-0000-0000-000000000002', 'Beverages', 'Bottled, canned, and sachet drinks', 2),
    ('10000000-0000-0000-0000-000000000003', 'Tobacco', 'Cigarettes and tobacco products', 3),
    ('10000000-0000-0000-0000-000000000004', 'Groceries', 'Household staple foods', 4),
    ('10000000-0000-0000-0000-000000000005', 'Hygiene', 'Cleaning products and toiletries', 5),
    ('10000000-0000-0000-0000-000000000006', 'Medicine', 'Over-the-counter medicine and vitamins', 6),
    ('10000000-0000-0000-0000-000000000007', 'Others', 'Miscellaneous products', 99);
