
DELETE FROM assigned_roles WHERE id IN ('ar-t1a', 'ar-t1b', 'ar-t2a', 'ar-t2b');
DELETE FROM issues WHERE id IN (
    'iss-t1a-1', 'iss-t1a-2', 'iss-t1a-3',
    'iss-t1b-1', 'iss-t1b-2', 'iss-t1b-3',
    'iss-t2a-1', 'iss-t2a-2', 'iss-t2a-3',
    'iss-t2b-1', 'iss-t2b-2', 'iss-t2b-3'
);
UPDATE projects SET default_visibility_group_id = NULL
 WHERE id IN ('proj-t1-a', 'proj-t1-b', 'proj-t2-a', 'proj-t2-b');
DELETE FROM user_group_members WHERE user_id IN ('test1', 'test2');
DELETE FROM user_groups WHERE id IN ('grp-t1-a', 'grp-t1-b', 'grp-t2-a', 'grp-t2-b');
DELETE FROM project_team_members WHERE user_id IN ('test1', 'test2');
UPDATE project_teams SET project_id = NULL
 WHERE id IN ('team-t1-a', 'team-t1-b', 'team-t2-a', 'team-t2-b');
DELETE FROM projects WHERE id IN ('proj-t1-a', 'proj-t1-b', 'proj-t2-a', 'proj-t2-b');
DELETE FROM project_teams WHERE id IN ('team-t1-a', 'team-t1-b', 'team-t2-a', 'team-t2-b');
DELETE FROM user_profiles WHERE user_id IN ('test1', 'test2', 'test3');
DELETE FROM users WHERE id IN ('test1', 'test2', 'test3');
DELETE FROM roles WHERE id = 'project_member';
