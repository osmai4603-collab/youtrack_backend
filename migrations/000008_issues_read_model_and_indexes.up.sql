-- =============================================================================
--  000008_issues_read_model_and_indexes.up.sql
--  =============================================================================
--  Purpose: Issue list support indexes and issues read model (columns, constraints, seed issues).
--  Source: init_database.sql sections [8] and [21] (lines 1311-1327, 2049-2221).
--  Dependencies: 000001_core_schema, 000007_identity_bootstrap_and_projects_seed.
-- =============================================================================

-- #############################################################################
--  [8]  فهارس عد المسائل
--  المصدر: migrations/000008_issue_count_support.up.sql
-- #############################################################################

-- ترقيعة: فهرس idx_issues_updated أُنشئ هنا تصاعديًا، ثم حاول 000021 إنشاء
-- فهرس بنفس الاسم تنازليًا فصار no-op صامتًا بسبب IF NOT EXISTS، وبقي
-- الفهرس تصاعديًا مع أن مسار الاستعلام الحار في Issues هو updated DESC.
-- هنا نُنشئه مرة واحدة وبالترتيب التنازلي الصحيح.

-- request17.txt: GET /api/issuesGetter/count
-- فهارس لتسريع استعلامات العد SELECT COUNT(*) على جدول issues.

CREATE INDEX IF NOT EXISTS idx_issues_project_resolved ON issues (project_id, resolved);
CREATE INDEX IF NOT EXISTS idx_issues_updated ON issues (updated DESC);



-- #############################################################################
--  [21] نموذج قراءة المسائل (8 أعمدة + 4 فهارس + 6 مسائل)
--  المصدر: migrations/000021_issues_read_model.up.sql
-- #############################################################################

-- ترقيعة: حُذفت من هذا القسم الفهرس idx_issues_updated (وبسطره التالي) لأن
-- القسم [8] وفّره بترتيب updated DESC. وجوده مرتين يجعل إحداهما عديمة الأثر
-- بصمت، ويجعل تراجع 000021 يُسقط فهرس 000008.

-- 000021_issues_read_model.up.sql
--
-- Columns backing the read endpoints GET /api/issues,
-- GET /api/issues/{issueID} and GET /api/projects/{projectID}/issues.
--
-- The `issues` table inherited from 000001 stores only what the original Hub
-- import needed: identity, summary, timestamps and a vote count. The client
-- contract, however, renders state, priority, type, assignee, due date and the
-- effort figures on every row of the issues list, and the issue-filter UI sorts
-- and filters on state, priority and type. Without columns for them the
-- endpoints would have to invent values, and the state/priority/type filters
-- would silently match everything or nothing.
--
-- In Hub these four attributes are custom fields, not columns. This schema has
-- no working custom-field storage (`issue_custom_field_values` carries no value
-- column), so they are modelled as plain VARCHARs constrained to the exact
-- string set the client's enums accept. If the custom-field tables ever become
-- usable, these columns are the migration path back rather than a parallel
-- source of truth.

ALTER TABLE issues
    ADD COLUMN IF NOT EXISTS state        VARCHAR(32)  NOT NULL DEFAULT 'to-do',
    ADD COLUMN IF NOT EXISTS priority     VARCHAR(32)  NOT NULL DEFAULT 'normal',
    ADD COLUMN IF NOT EXISTS issue_type   VARCHAR(32)  NOT NULL DEFAULT 'task',
    ADD COLUMN IF NOT EXISTS assignee_id  VARCHAR(20)  REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS due_date     BIGINT,
    ADD COLUMN IF NOT EXISTS estimation   INT,
    ADD COLUMN IF NOT EXISTS spent_time   INT,
    ADD COLUMN IF NOT EXISTS subsystem_id VARCHAR(20),
    ADD COLUMN IF NOT EXISTS fix_versions TEXT;

-- The client's enums are sealed classes: IssueStateEnum.of(String) takes a
-- non-nullable String, so a row holding anything outside these sets would make
-- the Dart parser throw rather than fall back. CHECK constraints turn that
-- class of bug into a database error at write time instead.
ALTER TABLE issues
    ADD CONSTRAINT issues_state_check
        CHECK (state IN ('to-do', 'in-progress', 'done'));

ALTER TABLE issues
    ADD CONSTRAINT issues_priority_check
        CHECK (priority IN ('show-stopper', 'critical', 'major', 'normal', 'minor'));

ALTER TABLE issues
    ADD CONSTRAINT issues_type_check
        CHECK (issue_type IN (
            'bug', 'cosmetic', 'exception', 'feature', 'task',
            'usability-problem', 'performance-problem', 'epic'
        ));

-- Effort is stored in minutes because the client models it as Duration and
-- converts with Duration(minutes: ...). Negative values are meaningless for
-- both estimation and spent time.
ALTER TABLE issues
    ADD CONSTRAINT issues_estimation_check
        CHECK (estimation IS NULL OR estimation >= 0),
    ADD CONSTRAINT issues_spent_time_check
        CHECK (spent_time IS NULL OR spent_time >= 0);

-- The project-scoped list is the hot path: it filters on project_id and orders
-- by the sort field the client chose. 000008 already indexes
-- (project_id, resolved); these cover the remaining sort orders.
CREATE INDEX IF NOT EXISTS idx_issues_project_updated
    ON issues (project_id, updated DESC);


CREATE INDEX IF NOT EXISTS idx_issues_state
    ON issues (state);

-- The per-issue child counts (comments, attachments, watchers) are correlated
-- subqueries in the repository. Their child tables are keyed on issue_id.
CREATE INDEX IF NOT EXISTS idx_comments_issue_id
    ON comments (issue_id);

CREATE INDEX IF NOT EXISTS idx_attachments_issue_id
    ON attachments (issue_id);

-- Seed data. Every read endpoint here would otherwise answer an empty list,
-- which is indistinguishable from a broken filter, and the Demo project the
-- client links to from GET /api/projects/me would have no issues to show.
--
-- IDs follow the project's existing convention: readable keys are "<project>-<n>"
-- and internal ids are "<n>-0" style sequences already used by 000019. resolved
-- is set for the "done" rows so the state column and the legacy resolved column
-- do not contradict each other.
INSERT INTO issues (
    id, id_readable, number_in_project, summary, description,
    project_id, reporter_id, creator_id, updater_id,
    created, updated, resolved, votes,
    state, priority, issue_type, assignee_id,
    due_date, estimation, spent_time
) VALUES
    ('3-0', 'DEMO-1', 1,
     'Sign in screen ignores keyboard submit on mobile web',
     'Pressing Enter in the credentials field does not submit the form on iOS Safari. '
     'Reproduced on 0-0 with a hardware keyboard attached.',
     '0-0', 'admin', 'admin', 'admin',
     1782000000000, 1784851998350, NULL, 3,
     'in-progress', 'major', 'bug', 'admin',
     1787000000000, 120, 45),

    ('3-1', 'DEMO-2', 2,
     'Project members list is empty for org-scoped admins',
     'An admin holding only an organization-scoped role sees every project in '
     'GET /api/projects/me but an empty roster on the members endpoint.',
     '0-0', 'admin', 'admin', 'admin',
     1782100000000, 1784851998350, NULL, 1,
     'to-do', 'critical', 'bug', NULL,
     NULL, 240, 0),

    ('3-2', 'DEMO-3', 3,
     'Archived projects should stay reachable by id',
     'Favouriting a project that is hidden from the default listing used to make it '
     'unfavouritable, because the listing the star is derived from was unreachable.',
     '0-0', 'admin', 'admin', 'admin',
     1782200000000, 1784800000000, 1784700000000, 5,
     'done', 'minor', 'task', '2-0',
     NULL, 30, 30),

    ('3-3', 'DEMO-4', 4,
     'Add a per-project notification digest',
     'Would let a team review issue activity once a day instead of per notification.',
     '0-0', '2-0', '2-0', '2-0',
     1783000000000, 1784780000000, NULL, 0,
     'to-do', 'normal', 'feature', NULL,
     NULL, NULL, NULL),

    ('3-4', 'DEMO-5', 5,
     'Issue filters reset when navigating back from detail',
     'Returning to the list from an issue drops the active state filter, so the '
     'user has to rebuild it each time.',
     '0-0', '2-0', '2-0', '2-0',
     1783100000000, 1784770000000, NULL, 2,
     'in-progress', 'major', 'usability-problem', 'admin',
     1786500000000, 60, 20),

    ('3-5', 'DEMO-6', 6,
     'Epic: rework the activity feed',
     'Parent effort for the activity stream rewrite. Not user-facing on its own.',
     '0-0', 'admin', 'admin', 'admin',
     1781500000000, 1784760000000, NULL, 4,
     'in-progress', 'show-stopper', 'epic', 'admin',
     1790000000000, 480, 90)
ON CONFLICT (id) DO NOTHING;

-- Two tags on the Demo project, so the tag filter has something to match. The
-- `tags` table has no created_at column, and the client's TagModel parses that
-- field with DateTime.parse, so tags are deliberately NOT serialized into the
-- issue payload yet — see the note in internal/httphandlers/handles/issues.
INSERT INTO tags (id, name, is_deletable, is_updatable, is_usable) VALUES
    ('7-0', 'regression', TRUE, TRUE, TRUE),
    ('7-1', 'frontend',   TRUE, TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO issue_tags (issue_id, tag_id) VALUES
    ('3-0', '7-0'),
    ('3-1', '7-0'),
    ('3-2', '7-1'),
    ('3-3', '7-1'),
    ('3-4', '7-0'),
    ('3-5', '7-1')
ON CONFLICT DO NOTHING;


