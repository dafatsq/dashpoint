-- 033_add_categories_permission.up.sql
-- Add can_manage_categories permission and assign to owner/manager roles

-- Insert the permission
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('cccccccc-cccc-cccc-cccc-cccccccccccf', 'can_manage_categories', 'Manage Categories', 'Can create, edit, archive, and permanently delete product and expense categories', 'categories')
ON CONFLICT (id) DO UPDATE SET
    key = EXCLUDED.key,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category;

-- Assign to owner role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-cccccccccccf')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign to manager role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('22222222-2222-2222-2222-222222222222', 'cccccccc-cccc-cccc-cccc-cccccccccccf')
ON CONFLICT (role_id, permission_id) DO NOTHING;
