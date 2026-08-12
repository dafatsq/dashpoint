-- Add new granular permissions for creating, editing, deleting specific roles
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('f0000000-0000-0000-0000-000000000005', 'can_create_manager_users', 'Create Manager Users', 'Can create manager users', 'users'),
    ('f0000000-0000-0000-0000-000000000006', 'can_create_cashier_users', 'Create Cashier Users', 'Can create cashier users', 'users'),
    ('f0000000-0000-0000-0000-000000000007', 'can_edit_manager_users', 'Edit Manager Users', 'Can modify details of manager users', 'users'),
    ('f0000000-0000-0000-0000-000000000008', 'can_edit_cashier_users', 'Edit Cashier Users', 'Can modify details of cashier users', 'users'),
    ('f0000000-0000-0000-0000-000000000009', 'can_delete_manager_users', 'Delete/Archive Manager Users', 'Can archive or delete manager users', 'users'),
    ('f0000000-0000-0000-0000-000000000010', 'can_delete_cashier_users', 'Delete/Archive Cashier Users', 'Can archive or delete cashier users', 'users');

-- Assign these permissions to the Owner role (so they naturally have them)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'owner' AND p.key IN (
    'can_create_manager_users', 'can_create_cashier_users',
    'can_edit_manager_users', 'can_edit_cashier_users',
    'can_delete_manager_users', 'can_delete_cashier_users'
)
ON CONFLICT DO NOTHING;
