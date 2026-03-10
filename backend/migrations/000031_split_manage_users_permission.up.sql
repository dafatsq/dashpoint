-- 000031_split_manage_users_permission.up.sql

-- 1. Insert the new granular permissions
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeedd', 'can_create_user', 'Create Users', 'Can create new employees', 'users'),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'can_edit_user', 'Edit Users', 'Can modify employee details', 'users'),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeff', 'can_delete_user', 'Delete/Archive Users', 'Can archive or delete employees', 'users');

-- 2. Migrate roles that currently have 'can_manage_users' to have all three new permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p_new.id
FROM role_permissions rp
JOIN permissions p_old ON rp.permission_id = p_old.id
CROSS JOIN permissions p_new
WHERE p_old.key = 'can_manage_users'
  AND p_new.key IN ('can_create_user', 'can_edit_user', 'can_delete_user')
ON CONFLICT DO NOTHING;

-- 3. Migrate user-specific overrides that currently override 'can_manage_users'
INSERT INTO user_permissions (user_id, permission_id, allowed)
SELECT up.user_id, p_new.id, up.allowed
FROM user_permissions up
JOIN permissions p_old ON up.permission_id = p_old.id
CROSS JOIN permissions p_new
WHERE p_old.key = 'can_manage_users'
  AND p_new.key IN ('can_create_user', 'can_edit_user', 'can_delete_user')
ON CONFLICT (user_id, permission_id) DO UPDATE SET allowed = EXCLUDED.allowed;

-- 4. Delete the old 'can_manage_users' permission
DELETE FROM role_permissions WHERE permission_id = (SELECT id FROM permissions WHERE key = 'can_manage_users');
DELETE FROM user_permissions WHERE permission_id = (SELECT id FROM permissions WHERE key = 'can_manage_users');
DELETE FROM permissions WHERE key = 'can_manage_users';
