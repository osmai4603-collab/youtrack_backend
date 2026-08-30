-- 000014_user_profile_date_field_pattern.up.sql
-- Add date_field_pattern column to store the datetime format (e.g., "d MMM yyyy HH:mm")
-- The existing date_pattern column stores the date-only format (e.g., "d MMM yyyy")

ALTER TABLE user_profiles ADD COLUMN date_field_pattern VARCHAR(50);
