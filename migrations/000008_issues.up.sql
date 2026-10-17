-- Migration: create issues tables and supporting triggers based on YouTrack API entities.
-- Official source: https://www.jetbrains.com/help/youtrack/devportal/api-entity-Issue.html
CREATE TABLE IF NOT EXISTS issues (
  -- The database ID of the issue. Read-only.
  id                      VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  -- The issue ID as seen in the YouTrack interface (for example, TST-5). Read-only.
  id_readable             VARCHAR(64) NOT NULL UNIQUE,
  -- The project where the issue belongs.
  project_id              VARCHAR(64) NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  -- The issue number in the project. Read-only.
  number_in_project       BIGINT NOT NULL,
  -- The issue summary. Can be null.
  summary                 TEXT,
  -- The issue description. Can be null.
  description             TEXT,
  -- The issue description as shown in the UI after processing wiki/Markdown markup. Read-only.
  wikified_description    TEXT,
  -- The user who reported (created) the issue. Read-only. Can be null.
  reporter_id             VARCHAR(64) REFERENCES users (id) ON DELETE SET NULL,
  -- The user who last updated the issue. Read-only. Can be null.
  updater_id              VARCHAR(64) REFERENCES users (id) ON DELETE SET NULL,
  -- The creator of the draft if the issue is a draft; null if the issue is reported. Read-only. Can be null.
  draft_owner_id          VARCHAR(64) REFERENCES users (id) ON DELETE SET NULL,
  -- true if this issue is a draft, false if it is reported. Read-only.
  is_draft                BOOLEAN NOT NULL DEFAULT FALSE,
  -- The moment when the issue was created. YouTrack stores it as a unix timestamp at UTC.
  created                 TIMESTAMPTZ NOT NULL DEFAULT now(),
  -- The moment of the last update of the issue. YouTrack stores it as a unix timestamp at UTC.
  updated                 TIMESTAMPTZ NOT NULL DEFAULT now(),
  -- The moment when the issue was assigned a state that is considered to be resolved;
  -- null if the issue is still in an unresolved state. Read-only. Can be null.
  resolved                TIMESTAMPTZ,
  -- The number of comments in the issue. Read-only.
  comments_count          INTEGER NOT NULL DEFAULT 0,
  -- The sum of votes for this issue and votes for its duplicates. Read-only.
  votes                   INTEGER NOT NULL DEFAULT 0,
  -- The parent issue for the current one. Null if the issue is not a sub-task of any issue. Read-only.
  -- The subtasks list (Issue.subtasks) is derived from this column.
  parent_id               VARCHAR(64) REFERENCES issues (id) ON DELETE SET NULL,
  -- The list of tags that are added to the issue.
  tags                    TEXT[] DEFAULT '{}',
  -- Visibility settings of the issue: who is allowed to see it
  -- (LimitedVisibility: permittedUsers[], permittedGroups[]). Can be null.
  visibility              JSONB,
  -- Reference to the issue in an originating third-party system
  -- (ExternalIssue: id, name, url, key). Read-only. Can be null.
  external_issue          JSONB,
  -- Application audit timestamp when this record was created.
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  -- Application audit timestamp when this record was last updated.
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT issues_unique_number_in_project UNIQUE (project_id, number_in_project)
);

CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues (project_id);
CREATE INDEX IF NOT EXISTS idx_issues_reporter_id ON issues (reporter_id);
CREATE INDEX IF NOT EXISTS idx_issues_updater_id ON issues (updater_id);
CREATE INDEX IF NOT EXISTS idx_issues_parent_id ON issues (parent_id);
CREATE INDEX IF NOT EXISTS idx_issues_tags ON issues USING GIN (tags);

-- Create issue_comments table based on YouTrack API (IssueComment entity)
-- The pinned comments of the issue (Issue.pinnedComments) are the comments with pinned = true.
CREATE TABLE IF NOT EXISTS issue_comments (
  -- The ID of the comment. Read-only.
  id                      VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  -- The issue the comment belongs to. Read-only.
  issue_id                VARCHAR(64) NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
  -- The user who created the comment. Read-only. Can be null.
  author_id               VARCHAR(64) REFERENCES users (id) ON DELETE SET NULL,
  -- The text of the comment. Can be null.
  text                    TEXT,
  -- The comment text as it is shown in UI after processing wiki/markdown markup. Read-only.
  text_preview            TEXT,
  -- When true, the comment is considered to be deleted, otherwise false.
  deleted                 BOOLEAN NOT NULL DEFAULT FALSE,
  -- Determines whether the comment is pinned in the issue.
  pinned                  BOOLEAN NOT NULL DEFAULT FALSE,
  -- The list of reactions that users added to this comment.
  reactions               JSONB DEFAULT '[]',
  -- Visibility settings of the comment: who is allowed to see the comment. Can be null.
  visibility              JSONB,
  -- The moment when the comment was posted. YouTrack stores it as a unix timestamp at UTC. Read-only.
  created                 TIMESTAMPTZ,
  -- The moment of the last update of the comment. YouTrack stores it as a unix timestamp at UTC. Can be null.
  updated                 TIMESTAMPTZ,
  -- Application audit timestamp when this record was created.
  created_at              TIMESTAMPTZ,
  -- Application audit timestamp when this record was last updated.
  updated_at              TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_issue_comments_issue_id ON issue_comments (issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_comments_author_id ON issue_comments (author_id);

-- Create issue_custom_fields table based on YouTrack API (IssueCustomField entity)
-- Stores the custom field values that are present in a particular issue (Issue.customFields).
CREATE TABLE IF NOT EXISTS issue_custom_fields (
  -- The ID of the custom field in the issue. Read-only.
  id                      VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  -- The issue the custom field value belongs to.
  issue_id                VARCHAR(64) NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
  -- The entity ID of the custom field. Read-only.
  custom_field_id         VARCHAR(64) NOT NULL,
  -- The name of the custom field. Read-only.
  name                    VARCHAR(255) NOT NULL,
  -- The unique type of the issue custom field ($type), for example SingleEnumIssueCustomField.
  field_type              VARCHAR(64),
  -- Reference to the settings of the custom field in the project (ProjectCustomField). Read-only.
  project_custom_field_id VARCHAR(64),
  -- The value assigned to the custom field in the issue. Depending on the type of the field,
  -- this attribute can store a primitive value, a single entity value or an array of entity values.
  -- Read-only. Can be null.
  value                   JSONB,
  -- Application audit timestamp when this record was created.
  created_at              TIMESTAMPTZ,
  -- Application audit timestamp when this record was last updated.
  updated_at              TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_issue_custom_fields_issue_id ON issue_custom_fields (issue_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_issue_custom_fields_issue_field ON issue_custom_fields (issue_id, custom_field_id);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'issue_link_direction') THEN
    CREATE TYPE issue_link_direction AS ENUM ('OUTWARD', 'INWARD', 'BOTH');
  END IF;
END$$;

-- Create issue_links table based on YouTrack API (IssueLink entity)
-- Represents issue links of a particular link type (for example, 'relates to').
CREATE TABLE IF NOT EXISTS issue_links (
  -- The ID of the issue link. Read-only.
  id                      VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  -- The issue that owns this link.
  issue_id                VARCHAR(64) NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
  -- The issue that is linked with this link type (IssueLink.issues).
  linked_issue_id         VARCHAR(64) NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
  -- Link direction: OUTWARD, INWARD or BOTH. Read-only.
  direction               issue_link_direction NOT NULL DEFAULT 'OUTWARD',
  -- The name of the link type, for example 'relates to', 'parent for'. Read-only. Can be null.
  link_type               VARCHAR(255),
  -- The ID of the link type (IssueLinkType). Read-only. Can be null.
  link_type_id            VARCHAR(64),
  -- Application audit timestamp when this record was created.
  created_at              TIMESTAMPTZ,

  CONSTRAINT issue_links_no_self_link CHECK (issue_id <> linked_issue_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_links_issue_id ON issue_links (issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_links_linked_issue_id ON issue_links (linked_issue_id);

-- Create issue_attachments table based on YouTrack API (IssueAttachment entity)
CREATE TABLE IF NOT EXISTS issue_attachments (
  -- The ID of the attachment. Read-only.
  id                      VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  -- The issue that the file is attached to. Read-only.
  issue_id                VARCHAR(64) NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
  -- The comment that the file is attached to; null when attached directly to the issue. Read-only. Can be null.
  comment_id              VARCHAR(64) REFERENCES issue_comments (id) ON DELETE CASCADE,
  -- The name of the file.
  name                    VARCHAR(255),
  -- The user who attached the file to the issue. Read-only. Can be null.
  author_id               VARCHAR(64) REFERENCES users (id) ON DELETE SET NULL,
  -- The size of the attached file in bytes. Read-only.
  size                    BIGINT,
  -- The extension that defines the file type. Read-only. Can be null.
  extension               VARCHAR(64),
  -- Charset of the file. Read-only. Can be null.
  charset                 VARCHAR(64),
  -- Mime type of the file. Read-only. Can be null.
  mime_type               VARCHAR(255),
  -- The dimensions of an image file (rw=&rh=). Empty for a non-image file. Read-only. Can be null.
  meta_data               TEXT,
  -- If true, attachment is not yet published, otherwise false. Read-only.
  draft                   BOOLEAN NOT NULL DEFAULT FALSE,
  -- If true, then attachment is considered to be removed. Read-only.
  removed                 BOOLEAN NOT NULL DEFAULT FALSE,
  -- URL of the file. Read-only. Can be null. The file content itself (base64Content) is not stored here.
  url                     TEXT,
  -- URL of the attachment thumbnail. Read-only. Can be null.
  thumbnail_url           TEXT,
  -- Access setting of the attachment. Can be null.
  visibility              JSONB,
  -- The moment when the attachment was created. YouTrack stores it as a unix timestamp at UTC. Read-only.
  created                 TIMESTAMPTZ,
  -- The moment of the last update of the attachment. YouTrack stores it as a unix timestamp at UTC. Read-only.
  updated                 TIMESTAMPTZ,
  -- Application audit timestamp when this record was created.
  created_at              TIMESTAMPTZ,
  -- Application audit timestamp when this record was last updated.
  updated_at              TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_issue_attachments_issue_id ON issue_attachments (issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_attachments_comment_id ON issue_attachments (comment_id);

-- Create issue_watchers table based on YouTrack API (IssueWatchers entity)
-- Stores the users that are subscribed to notifications about the issue (Issue.watchers).
CREATE TABLE IF NOT EXISTS issue_watchers (
  -- The issue the watcher is subscribed to.
  issue_id                VARCHAR(64) NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
  -- The user who subscribed for notifications about the issue.
  user_id                 VARCHAR(64) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  -- true if the current user added the "Star" tag to this issue. Otherwise, false.
  has_star                BOOLEAN NOT NULL DEFAULT FALSE,
  -- true if the user also receives notifications about duplicates of this issue.
  duplicate_watcher       BOOLEAN NOT NULL DEFAULT FALSE,
  -- Application audit timestamp when this record was created.
  created_at              TIMESTAMPTZ,

  PRIMARY KEY (issue_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_watchers_user_id ON issue_watchers (user_id);

-- Create issue_voters table based on YouTrack API (IssueVoters entity)
-- Stores the users that have voted for the issue or its duplicates (Issue.voters).
CREATE TABLE IF NOT EXISTS issue_voters (
  -- The issue the vote is cast for.
  issue_id                VARCHAR(64) NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
  -- The user who voted for the issue.
  user_id                 VARCHAR(64) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  -- true if the vote was cast for a duplicate of this issue and counts towards it.
  duplicate_vote          BOOLEAN NOT NULL DEFAULT FALSE,
  -- Application audit timestamp when this record was created.
  created_at              TIMESTAMPTZ,

  PRIMARY KEY (issue_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_voters_user_id ON issue_voters (user_id);

-- Keep the read-only counters of the Issue entity in sync (commentsCount, votes).
CREATE OR REPLACE FUNCTION recountIssueCommentsAfterChange()
RETURNS TRIGGER AS $$
DECLARE
  target_issue_id VARCHAR(64);
BEGIN
  IF TG_OP = 'DELETE' THEN
    target_issue_id := OLD.issue_id;
  ELSE
    target_issue_id := NEW.issue_id;
  END IF;

  UPDATE issues
  SET comments_count = (SELECT COUNT(*) FROM issue_comments WHERE issue_id = target_issue_id AND NOT deleted)
  WHERE id = target_issue_id;

  IF TG_OP = 'UPDATE' AND OLD.issue_id IS DISTINCT FROM NEW.issue_id THEN
    UPDATE issues
    SET comments_count = (SELECT COUNT(*) FROM issue_comments WHERE issue_id = OLD.issue_id AND NOT deleted)
    WHERE id = OLD.issue_id;
  END IF;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS triggerAfterChangeIssueComments ON issue_comments;

CREATE TRIGGER triggerAfterChangeIssueComments
  AFTER INSERT OR UPDATE OR DELETE ON issue_comments
  FOR EACH ROW EXECUTE FUNCTION recountIssueCommentsAfterChange();

CREATE OR REPLACE FUNCTION recountIssueVotesAfterChange()
RETURNS TRIGGER AS $$
DECLARE
  target_issue_id VARCHAR(64);
BEGIN
  IF TG_OP = 'DELETE' THEN
    target_issue_id := OLD.issue_id;
  ELSE
    target_issue_id := NEW.issue_id;
  END IF;

  UPDATE issues
  SET votes = (SELECT COUNT(*) FROM issue_voters WHERE issue_id = target_issue_id)
  WHERE id = target_issue_id;

  IF TG_OP = 'UPDATE' AND OLD.issue_id IS DISTINCT FROM NEW.issue_id THEN
    UPDATE issues
    SET votes = (SELECT COUNT(*) FROM issue_voters WHERE issue_id = OLD.issue_id)
    WHERE id = OLD.issue_id;
  END IF;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS triggerAfterChangeIssueVoters ON issue_voters;

CREATE TRIGGER triggerAfterChangeIssueVoters
  AFTER INSERT OR UPDATE OR DELETE ON issue_voters
  FOR EACH ROW EXECUTE FUNCTION recountIssueVotesAfterChange();
