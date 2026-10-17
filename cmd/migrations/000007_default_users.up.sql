
INSERT INTO users (id, login, email, full_name, name, avatar_url, ring_id, type,
                   guest, online, banned, hash_password)
VALUES
    ('test1', 'test1', 'test1@youtrack.local', 'Test One',   'test1',
     '/hub/api/rest/avatar/t1a00000-0000-0000-0000-000000000001?s=48', '', 'STANDARD_USER',
     FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC'),
    ('test2', 'test2', 'test2@youtrack.local', 'Test Two',   'test2',
     '/hub/api/rest/avatar/t2b00000-0000-0000-0000-000000000002?s=48', '', 'STANDARD_USER',
     FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC'),
    ('test3', 'test3', 'test3@youtrack.local', 'Test Three', 'test3',
     '/hub/api/rest/avatar/t3c00000-0000-0000-0000-000000000003?s=48', '', 'STANDARD_USER',
     FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC')
ON CONFLICT (id) DO NOTHING;


INSERT INTO projects (id, name, short_name, created_by, leader_id)
VALUES
    ('proj-t1-a', 'Test1 Project Alpha', 'T1A', 'test1', 'test1'),
    ('proj-t1-b', 'Test1 Project Beta',  'T1B', 'test1', 'test1'),
    ('proj-t2-a', 'Test2 Project Alpha', 'T2A', 'test2', 'test2'),
    ('proj-t2-b', 'Test2 Project Beta',  'T2B', 'test2', 'test2')
ON CONFLICT (id) DO NOTHING;




