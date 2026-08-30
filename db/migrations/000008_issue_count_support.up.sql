-- request17.txt: GET /api/issuesGetter/count
-- فهارس لتسريع استعلامات العد SELECT COUNT(*) على جدول issues.

CREATE INDEX IF NOT EXISTS idx_issues_project_resolved ON issues (project_id, resolved);
CREATE INDEX IF NOT EXISTS idx_issues_updated ON issues (updated);