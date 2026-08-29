-- 000004_dashboard_widgets_extended.up.sql
-- Add the expectedHeight/expectedWidth columns used by the
-- GET /api/admin/widgets/general endpoint (request6.txt).

ALTER TABLE dashboard_widgets
    ADD COLUMN IF NOT EXISTS expected_height VARCHAR(20),
    ADD COLUMN IF NOT EXISTS expected_width VARCHAR(20);