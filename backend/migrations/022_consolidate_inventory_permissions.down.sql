-- 022_consolidate_inventory_permissions.down.sql
-- Revert consolidation of inventory permissions

-- Re-insert the old permissions
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('cccccccc-cccc-cccc-cccc-cccccccccccb', 'can_adjust_stock', 'Adjust Stock', 'Can adjust stock quantities', 'inventory'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'can_receive_stock', 'Receive Stock', 'Can record stock receipts', 'inventory')
ON CONFLICT (id) DO UPDATE SET
    key = EXCLUDED.key,
    name = EXCLUDED.name,
    description = EXCLUDED.description;

-- Transfer role_permissions from edit to both adjust and receive
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p.id
FROM role_permissions rp
CROSS JOIN permissions p
WHERE rp.permission_id = (SELECT id FROM permissions WHERE key = 'can_edit_inventory')
  AND p.key IN ('can_adjust_stock', 'can_receive_stock')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Transfer user_permissions from edit to both adjust and receive
INSERT INTO user_permissions (user_id, permission_id, allowed)
SELECT up.user_id, p.id, up.allowed
FROM user_permissions up
CROSS JOIN permissions p
WHERE up.permission_id = (SELECT id FROM permissions WHERE key = 'can_edit_inventory')
  AND p.key IN ('can_adjust_stock', 'can_receive_stock')
ON CONFLICT (user_id, permission_id) DO NOTHING;

-- Remove consolidated permission from role_permissions
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE key = 'can_edit_inventory'
);

-- Remove consolidated permission from user_permissions
DELETE FROM user_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE key = 'can_edit_inventory'
);

-- Delete consolidated permission
DELETE FROM permissions WHERE key = 'can_edit_inventory';
