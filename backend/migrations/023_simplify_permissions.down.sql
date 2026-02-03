-- 023_simplify_permissions.down.sql
-- Restore removed permissions

-- Restore can_edit_price
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbe', 'can_edit_price', 'Edit Prices', 'Can modify product prices', 'products');

-- Restore can_backup_data
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('ffffffff-ffff-ffff-ffff-fffffffffffc', 'can_backup_data', 'Backup Data', 'Can create and restore backups', 'system');

-- Restore granular sales permissions
INSERT INTO permissions (id, key, name, description, category) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaab', 'can_refund', 'Process Refunds', 'Can process refunds and returns', 'sales'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaac', 'can_apply_discount', 'Apply Discounts', 'Can apply discounts to sales', 'sales'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaad', 'can_void_sale', 'Void Sales', 'Can void/cancel sales transactions', 'sales');

-- Restore original can_create_sale description
UPDATE permissions 
SET name = 'Create Sale', 
    description = 'Can process sales transactions'
WHERE key = 'can_create_sale';

-- Note: Role and user permission assignments are not restored as they would require backup data
