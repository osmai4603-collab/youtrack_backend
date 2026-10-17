INSERT INTO issues (
    id,
    id_readable,
    project_id,
    number_in_project,
    summary,
    description,
    reporter_id,
    updater_id,
    state,
    priority,
    issue_type
)
SELECT
    md5('default-issue:' || p.id),
    left(p.short_name, 64 - length(issue_number.next_number::text) - 1)
        || '-' || issue_number.next_number,
    p.id,
    issue_number.next_number,
    'Review project setup',
    'Default starter issue for ' || p.name || '.',
    COALESCE(p.leader_id, p.created_by),
    COALESCE(p.leader_id, p.created_by),
    'to-do',
    'normal',
    'task'
FROM projects AS p
CROSS JOIN LATERAL (
    SELECT GREATEST(
        p.starting_number,
        COALESCE(MAX(i.number_in_project) + 1, 1)
    ) AS next_number
    FROM issues AS i
    WHERE i.project_id = p.id
) AS issue_number
ON CONFLICT DO NOTHING;
