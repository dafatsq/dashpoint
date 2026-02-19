-- 024_separate_sales_permissions.up.sql
-- Separate sales permissions into create, view history, and void

-- Update existing permission
UPDATE permissions 
SET name = 'Create Sales', 
    description = 'Can create and process sales transactions'
WHERE key = 'can_create_sale';

-- Add new permissions
INSERT INTO permissions (key, name, description, category) VALUES
    ('can_view_sales', 'View Sales History', 'Can view sales transactions and history', 'sales'),
    ('can_void_sale', 'Void Sales', 'Can void/cancel sales transactions', 'sales')
ON CONFLICT (key) DO UPDATE 
SET name = EXCLUDED.name, 
    description = EXCLUDED.description, 
    category = EXCLUDED.category;

-- Assign new permissions to owner role (already has all)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM permissions
WHERE key IN ('can_view_sales', 'can_void_sale')
ON CONFLICT DO NOTHING;

-- Assign new permissions to manager role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '22222222-2222-2222-2222-222222222222', id FROM permissions
WHERE key IN ('can_view_sales', 'can_void_sale')
ON CONFLICT DO NOTHING;

-- Assign can_view_sales to cashier role (but not void)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '33333333-3333-3333-3333-333333333333', id FROM permissions
WHERE key = 'can_view_sales'
ON CONFLICT DO NOTHING;
