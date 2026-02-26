-- 000029_remove_settings_permission.down.sql
-- Re-add "Manage Settings" permission

INSERT INTO permissions (id, key, name, description, category) VALUES
('ffffffff-ffff-ffff-ffff-fffffffffffb', 'can_manage_settings', 'Manage Settings', 'Can modify system settings', 'system');

-- Reassign to owner role by default (ID: 11111111-1111-1111-1111-111111111111)
INSERT INTO role_permissions (role_id, permission_id)
VALUES ('11111111-1111-1111-1111-111111111111', 'ffffffff-ffff-ffff-ffff-fffffffffffb');
