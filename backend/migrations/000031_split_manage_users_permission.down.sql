-- 000031_split_manage_users_permission.down.sql

-- 1. Re-insert the old 'can_manage_users' permission
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeeb', 'can_manage_users', 'Manage Users', 'Can create and edit employees', 'users');

-- 2. Migrate roles back (if a role has any of the granular permissions, grant them the combined one)
INSERT INTO role_permissions (role_id, permission_id)
SELECT DISTINCT rp.role_id, p_old.id
FROM role_permissions rp
JOIN permissions p_new ON rp.permission_id = p_new.id
CROSS JOIN permissions p_old
WHERE p_new.key IN ('can_create_user', 'can_edit_user', 'can_delete_user')
  AND p_old.key = 'can_manage_users'
ON CONFLICT DO NOTHING;

-- 3. Migrate user overrides back
INSERT INTO user_permissions (user_id, permission_id, allowed)
SELECT up.user_id, p_old.id, bool_or(up.allowed) -- If any of the granular is true, then true.
FROM user_permissions up
JOIN permissions p_new ON up.permission_id = p_new.id
CROSS JOIN permissions p_old
WHERE p_new.key IN ('can_create_user', 'can_edit_user', 'can_delete_user')
  AND p_old.key = 'can_manage_users'
GROUP BY up.user_id, p_old.id
ON CONFLICT (user_id, permission_id) DO UPDATE SET allowed = EXCLUDED.allowed;

-- 4. Delete the new granular permissions
DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE key IN ('can_create_user', 'can_edit_user', 'can_delete_user'));
DELETE FROM user_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE key IN ('can_create_user', 'can_edit_user', 'can_delete_user'));
DELETE FROM permissions WHERE key IN ('can_create_user', 'can_edit_user', 'can_delete_user');
