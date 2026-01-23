-- 021_normalize_user_emails.up.sql
-- Normalize all existing user emails to lowercase for case-insensitive login

UPDATE users SET email = LOWER(email) WHERE email IS NOT NULL;
