-- 022_consolidate_inventory_permissions.up.sql
-- Consolidate can_adjust_stock and can_receive_stock into can_edit_inventory

-- Insert new consolidated permission
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('cccccccc-cccc-cccc-cccc-cccccccccccb', 'can_edit_inventory', 'Edit Inventory', 'Can add, remove, or adjust stock quantities', 'inventory')
ON CONFLICT (id) DO UPDATE SET
    key = EXCLUDED.key,
    name = EXCLUDED.name,
    description = EXCLUDED.description;

-- Transfer role_permissions from old permissions to new one
-- If a role had either adjust or receive, give them edit
INSERT INTO role_permissions (role_id, permission_id)
SELECT DISTINCT rp.role_id, 'cccccccc-cccc-cccc-cccc-cccccccccccb'::uuid
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
WHERE p.key IN ('can_adjust_stock', 'can_receive_stock')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Transfer user_permissions from old permissions to new one
-- If a user had either adjust or receive override, give them edit override
INSERT INTO user_permissions (user_id, permission_id, allowed)
SELECT DISTINCT up.user_id, 'cccccccc-cccc-cccc-cccc-cccccccccccb'::uuid, up.allowed
FROM user_permissions up
JOIN permissions p ON up.permission_id = p.id
WHERE p.key IN ('can_adjust_stock', 'can_receive_stock')
  AND up.allowed = true
ON CONFLICT (user_id, permission_id) DO NOTHING;

-- Remove old permissions from role_permissions
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN ('can_adjust_stock', 'can_receive_stock')
);

-- Remove old permissions from user_permissions
DELETE FROM user_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN ('can_adjust_stock', 'can_receive_stock')
);

-- Delete old permissions
DELETE FROM permissions WHERE key IN ('can_adjust_stock', 'can_receive_stock');
