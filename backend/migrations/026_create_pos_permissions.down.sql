-- 026_create_pos_permissions.down.sql

DELETE FROM permissions WHERE key IN ('can_view_pos', 'can_start_shift', 'can_end_shift');
