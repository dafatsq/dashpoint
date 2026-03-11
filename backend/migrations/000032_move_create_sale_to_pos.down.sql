-- 000032_move_create_sale_to_pos.down.sql
-- Revert can_create_sale permission back to 'sales' category

UPDATE permissions SET category = 'sales' WHERE key = 'can_create_sale';
