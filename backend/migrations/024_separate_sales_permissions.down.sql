-- 024_separate_sales_permissions.down.sql
-- Rollback: Merge sales permissions back to single permission

-- Remove new permissions
DELETE FROM permissions WHERE key IN ('can_view_sales', 'can_void_sale');

-- Restore original permission
UPDATE permissions 
SET name = 'Process Sales', 
    description = 'Can create sales, apply discounts, process refunds, and void transactions'
WHERE key = 'can_create_sale';
