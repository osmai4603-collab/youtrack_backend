DELETE FROM issues
WHERE id = md5('default-issue:' || project_id);
