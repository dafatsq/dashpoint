-- 000032_move_create_sale_to_pos.up.sql
-- Move can_create_sale permission from 'sales' category to 'pos' category
-- so it appears in the POS section and is gated by can_view_pos access

UPDATE permissions SET category = 'pos' WHERE key = 'can_create_sale';
