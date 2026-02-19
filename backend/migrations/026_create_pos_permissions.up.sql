-- 026_create_pos_permissions.up.sql

-- Add new permissions
INSERT INTO permissions (key, name, description, category, created_at) VALUES 
('can_view_pos', 'View POS', 'Can access the Point of Sale interface', 'pos', NOW()),
('can_start_shift', 'Start Shift', 'Can start a new shift', 'pos', NOW()),
('can_end_shift', 'End Shift', 'Can end an active shift', 'pos', NOW());

-- Assign permissions to roles
-- Owner (gets everything by default usually, but explicit logic handles it. 
-- However, role_permissions mapping is good practice if not using superuser flag)

-- Get role IDs (assuming standard naming from earlier migrations)
DO $$
DECLARE
    owner_role_id UUID;
    manager_role_id UUID;
    cashier_role_id UUID;
    perm_view_pos_id UUID;
    perm_start_shift_id UUID;
    perm_end_shift_id UUID;
BEGIN
    SELECT id INTO owner_role_id FROM roles WHERE name = 'owner';
    SELECT id INTO manager_role_id FROM roles WHERE name = 'manager';
    SELECT id INTO cashier_role_id FROM roles WHERE name = 'cashier';

    SELECT id INTO perm_view_pos_id FROM permissions WHERE key = 'can_view_pos';
    SELECT id INTO perm_start_shift_id FROM permissions WHERE key = 'can_start_shift';
    SELECT id INTO perm_end_shift_id FROM permissions WHERE key = 'can_end_shift';

    -- Assign to Manager
    IF manager_role_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id) VALUES 
        (manager_role_id, perm_view_pos_id),
        (manager_role_id, perm_start_shift_id),
        (manager_role_id, perm_end_shift_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- Assign to Cashier
    IF cashier_role_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id) VALUES 
        (cashier_role_id, perm_view_pos_id),
        (cashier_role_id, perm_start_shift_id),
        (cashier_role_id, perm_end_shift_id)
        ON CONFLICT DO NOTHING;
    END IF;

    -- Owner usually handled by code logic (user.role_name == 'owner'), but adding for completeness if desired
    IF owner_role_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id) VALUES 
        (owner_role_id, perm_view_pos_id),
        (owner_role_id, perm_start_shift_id),
        (owner_role_id, perm_end_shift_id)
        ON CONFLICT DO NOTHING;
    END IF;
END $$;
