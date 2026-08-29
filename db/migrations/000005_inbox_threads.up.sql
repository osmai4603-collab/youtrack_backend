-- Migration to add Inbox Threads and related entities

CREATE TABLE inbox_threads (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    read BOOLEAN DEFAULT FALSE,
    muted BOOLEAN DEFAULT FALSE,
    notified BOOLEAN DEFAULT TRUE,
    target_type VARCHAR(100),
    thread_id VARCHAR(100),
    timestamp BIGINT,
    updated BIGINT,
    subject_text TEXT,
    subject_target_id VARCHAR(100)
);

CREATE TABLE inbox_messages (
    id VARCHAR(50) PRIMARY KEY,
    thread_id VARCHAR(50) REFERENCES inbox_threads(id) ON DELETE CASCADE,
    author_id VARCHAR(20) REFERENCES users(id) ON DELETE SET NULL,
    author_group_id VARCHAR(20) REFERENCES user_groups(id) ON DELETE SET NULL,
    timestamp BIGINT,
    text TEXT,
    type VARCHAR(50),
    pseudo BOOLEAN DEFAULT FALSE,
    empty_field_text TEXT,
    target_id VARCHAR(100),
    target_type VARCHAR(100)
);

CREATE TABLE inbox_message_params (
    id SERIAL PRIMARY KEY,
    message_id VARCHAR(50) REFERENCES inbox_messages(id) ON DELETE CASCADE,
    key VARCHAR(255),
    value TEXT
);

CREATE TABLE inbox_activities (
    id VARCHAR(50) PRIMARY KEY,
    message_id VARCHAR(50) REFERENCES inbox_messages(id) ON DELETE CASCADE,
    category_id VARCHAR(100),
    added_id VARCHAR(100), -- Reference to User, Issue, Comment etc depending on type
    removed_id VARCHAR(100),
    target_id VARCHAR(100),
    target_type VARCHAR(100),
    timestamp BIGINT
);

-- Indexing for performance
CREATE INDEX idx_inbox_threads_user ON inbox_threads(user_id);
CREATE INDEX idx_inbox_messages_thread ON inbox_messages(thread_id);
CREATE INDEX idx_inbox_activities_message ON inbox_activities(message_id);
