-- 000004_dashboard_widgets_extended.down.sql

ALTER TABLE dashboard_widgets
    DROP COLUMN IF EXISTS expected_height,
    DROP COLUMN IF EXISTS expected_width;