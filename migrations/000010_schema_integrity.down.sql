
ALTER TABLE issues DROP CONSTRAINT IF EXISTS fk_issues_subsystem;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_starting_number_check;
