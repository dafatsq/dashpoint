-- 027_seed_products.down.sql

DELETE FROM inventory_items WHERE product_id IN (
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee1',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee2',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee3',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee4',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee5',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee6'
);

DELETE FROM products WHERE id IN (
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee1',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee2',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee3',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee4',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee5',
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeee6'
);
