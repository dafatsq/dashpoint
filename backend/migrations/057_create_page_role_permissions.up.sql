CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_permissions_key ON permissions(key);
CREATE INDEX IF NOT EXISTS idx_permissions_category ON permissions(category);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

INSERT INTO permissions (key, name, description, category) VALUES
    ('access_pos_page', 'Access POS page', 'Can open the point-of-sale page', 'pos'),
    ('manage_pos_page', 'Manage POS page', 'Can process sales on the point-of-sale page', 'pos'),
    ('access_products_page', 'Access products page', 'Can open the products page', 'products'),
    ('manage_products_page', 'Manage products page', 'Can create, edit, archive, and delete products', 'products'),
    ('access_inventory_page', 'Access inventory page', 'Can open the inventory page', 'inventory'),
    ('manage_inventory_page', 'Manage inventory page', 'Can adjust stock and maintain inventory thresholds', 'inventory'),
    ('access_sales_page', 'Access sales page', 'Can open sales history and sale details', 'sales'),
    ('manage_sales_page', 'Manage sales page', 'Can void sales and manage sales records', 'sales'),
    ('access_reports_page', 'Access reports page', 'Can open reports', 'reports'),
    ('manage_reports_page', 'Manage reports page', 'Can export report data', 'reports'),
    ('access_expenses_page', 'Access expenses page', 'Can open expenses', 'expenses'),
    ('manage_expenses_page', 'Manage expenses page', 'Can create, edit, archive, and delete expenses', 'expenses'),
    ('access_categories_page', 'Access categories page', 'Can open product and expense categories', 'categories'),
    ('manage_categories_page', 'Manage categories page', 'Can create, edit, archive, and delete categories', 'categories'),
    ('access_users_page', 'Access users page', 'Can open users and roles', 'users'),
    ('manage_users_page', 'Manage users page', 'Can create, edit, archive, and delete users', 'users'),
    ('access_shifts_page', 'Access shifts page', 'Can open shift history', 'shifts'),
    ('manage_shifts_page', 'Manage shifts page', 'Can start, close, and operate shifts', 'shifts'),
    ('access_changes_page', 'Access changes page', 'Can open recent changes', 'changes'),
    ('access_audit_page', 'Access audit page', 'Can open audit logs', 'audit')
ON CONFLICT (key) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category;

DELETE FROM permissions WHERE key = 'access_settings_page';

DELETE FROM role_permissions;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'owner';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
    'access_pos_page', 'manage_pos_page',
    'access_products_page', 'manage_products_page',
    'access_inventory_page', 'manage_inventory_page',
    'access_sales_page', 'manage_sales_page',
    'access_reports_page', 'manage_reports_page',
    'access_expenses_page', 'manage_expenses_page',
    'access_categories_page', 'manage_categories_page',
    'access_users_page', 'manage_users_page',
    'access_shifts_page', 'manage_shifts_page',
    'access_changes_page',
    'access_audit_page'
)
WHERE r.name = 'manager'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
    'access_pos_page', 'manage_pos_page',
    'access_sales_page',
    'access_shifts_page', 'manage_shifts_page'
)
WHERE r.name = 'cashier'
ON CONFLICT DO NOTHING;
