-- 007_seed_roles_permissions.up.sql
-- Seed initial roles and permissions

-- Insert default roles
INSERT INTO roles (id, name, description) VALUES
    ('11111111-1111-1111-1111-111111111111', 'owner', 'Store owner with full access to all features'),
    ('22222222-2222-2222-2222-222222222222', 'manager', 'Store manager with operational control'),
    ('33333333-3333-3333-3333-333333333333', 'cashier', 'Cashier with point-of-sale access');

-- Insert default permissions
INSERT INTO permissions (id, key, name, description, category) VALUES
    -- Sales permissions
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'can_create_sale', 'Create Sales', 'Can create and process sales transactions', 'sales'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaab', 'can_view_sales', 'View Sales History', 'Can view sales transactions and history', 'sales'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaac', 'can_void_sale', 'Void Sales', 'Can void/cancel sales transactions', 'sales'),
    
    -- Product permissions
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbba', 'can_view_products', 'View Products', 'Can view product catalog', 'products'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'can_create_product', 'Create Products', 'Can add new products', 'products'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbc', 'can_edit_product', 'Edit Products', 'Can modify product details and prices', 'products'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbd', 'can_delete_product', 'Delete Products', 'Can remove products from catalog', 'products'),
    
    -- Inventory permissions
    ('cccccccc-cccc-cccc-cccc-ccccccccccca', 'can_view_inventory', 'View Inventory', 'Can view stock levels', 'inventory'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccb', 'can_edit_inventory', 'Edit Inventory', 'Can add, remove, or adjust stock quantities', 'inventory'),
    
    -- Report permissions
    ('dddddddd-dddd-dddd-dddd-ddddddddddda', 'can_view_reports', 'View Reports', 'Can view sales and inventory reports', 'reports'),
    ('dddddddd-dddd-dddd-dddd-dddddddddddb', 'can_export_data', 'Export Data', 'Can export reports to CSV', 'reports'),
    
    -- User management permissions
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeea', 'can_view_users', 'View Users', 'Can view employee list', 'users'),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeeb', 'can_manage_users', 'Manage Users', 'Can create and edit employees', 'users'),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeec', 'can_manage_permissions', 'Manage Permissions', 'Can modify user permissions', 'users'),
    
    -- System permissions
    ('ffffffff-ffff-ffff-ffff-fffffffffffa', 'can_view_audit_logs', 'View Audit Logs', 'Can view system audit logs', 'system'),
    ('ffffffff-ffff-ffff-ffff-fffffffffffb', 'can_manage_settings', 'Manage Settings', 'Can modify system settings', 'system');

-- Assign all permissions to owner role
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM permissions;

-- Assign manager permissions (most except system management)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '22222222-2222-2222-2222-222222222222', id FROM permissions
WHERE key IN (
    'can_create_sale', 'can_view_sales', 'can_void_sale',
    'can_view_products', 'can_create_product', 'can_edit_product', 'can_delete_product',
    'can_view_inventory', 'can_edit_inventory',
    'can_view_reports', 'can_export_data',
    'can_view_users', 'can_manage_users'
);

-- Assign cashier permissions (basic POS operations)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '33333333-3333-3333-3333-333333333333', id FROM permissions
WHERE key IN (
    'can_create_sale', 'can_view_sales',
    'can_view_products',
    'can_view_inventory'
);
