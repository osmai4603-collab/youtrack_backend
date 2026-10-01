
DO $do$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_issues_subsystem') THEN
        ALTER TABLE issues
            ADD CONSTRAINT fk_issues_subsystem
            FOREIGN KEY (subsystem_id) REFERENCES project_subsystems(id) ON DELETE SET NULL;
    END IF;
END;
$do$;

DO $do$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'projects_starting_number_check') THEN
        ALTER TABLE projects
            ADD CONSTRAINT projects_starting_number_check
            CHECK (starting_number IS NULL OR starting_number >= 1);
    END IF;
END;
$do$;

