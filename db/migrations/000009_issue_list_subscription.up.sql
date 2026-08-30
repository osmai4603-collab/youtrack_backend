-- 000008_issue_list_subscription.up.sql
-- New independent schema for the issue list subscription endpoint
-- (request18.txt: /api/issueListSubscription?fields=ticket).
-- No existing table is modified.

CREATE TABLE issue_list_subscriptions (
    id SERIAL PRIMARY KEY,
    ticket VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    query TEXT NOT NULL DEFAULT '',
    subscribe BOOLEAN NOT NULL DEFAULT TRUE,
    context_type VARCHAR(100) NOT NULL DEFAULT 'Project',
    context_id VARCHAR(50) NOT NULL DEFAULT '0-0',
    folder_id VARCHAR(50),
    type VARCHAR(100) NOT NULL DEFAULT 'IssueListSubscriptionBean',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE issue_list_subscription_issues (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL REFERENCES issue_list_subscriptions(id) ON DELETE CASCADE,
    issue_id VARCHAR(50) NOT NULL,
    matches BOOLEAN NOT NULL DEFAULT TRUE,
    ordinal INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_issue_list_sub_issues ON issue_list_subscription_issues(subscription_id);

-- Seed a default ticket matching the request18.txt sample response.
INSERT INTO issue_list_subscriptions (ticket, query, subscribe, context_type, context_id)
VALUES ('g9qk4fe77jefpavefi1qacdhbefa3l', 'issue id: DEMO-4', TRUE, 'Project', '0-0');
