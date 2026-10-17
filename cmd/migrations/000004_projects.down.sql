DROP TRIGGER IF EXISTS triggerAfterInsertProject ON projects;
DROP FUNCTION IF EXISTS createTeamAfterInsertProject();
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
