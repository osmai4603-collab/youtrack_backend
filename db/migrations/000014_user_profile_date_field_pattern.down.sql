-- 000014_user_profile_date_field_pattern.down.sql

ALTER TABLE user_profiles DROP COLUMN IF EXISTS date_field_pattern;
