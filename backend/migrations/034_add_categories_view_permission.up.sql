-- 034_add_categories_view_permission.up.sql
-- Add can_view_categories permission and assign to owner, manager, and cashier roles

-- Insert the permission
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('cccccccc-cccc-cccc-cccc-ccccccccccc0', 'can_view_categories', 'View Categories', 'Can view the list of product and expense categories', 'categories')
ON CONFLICT (id) DO UPDATE SET
    key = EXCLUDED.key,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category;

-- Assign to owner role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-ccccccccccc0')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign to manager role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('22222222-2222-2222-2222-222222222222', 'cccccccc-cccc-cccc-cccc-ccccccccccc0')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign to cashier role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('33333333-3333-3333-3333-333333333333', 'cccccccc-cccc-cccc-cccc-ccccccccccc0')
ON CONFLICT (role_id, permission_id) DO NOTHING;
