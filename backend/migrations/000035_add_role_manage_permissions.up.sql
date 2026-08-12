-- Add new granular permissions for managing specific roles
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('f0000000-0000-0000-0000-000000000003', 'can_manage_manager_permissions', 'Manage Manager Permissions', 'Can modify permissions of manager users', 'users'),
    ('f0000000-0000-0000-0000-000000000004', 'can_manage_cashier_permissions', 'Manage Cashier Permissions', 'Can modify permissions of cashier users', 'users');

-- Assign these permissions to the Owner role (so they naturally have them)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'owner' AND p.key IN ('can_manage_manager_permissions', 'can_manage_cashier_permissions')
ON CONFLICT DO NOTHING;

-- Note: We do NOT auto-assign these to managers. Managers must be explicitly granted them via the UI (manage permissions).
