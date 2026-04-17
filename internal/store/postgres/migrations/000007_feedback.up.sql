-- Feedback submissions from the in-app "💬 Feedback" button. Public endpoint
-- so visitors on the /docs pages (not just authenticated dashboard users)
-- can submit. workspace_id + user_id are nullable for anonymous submissions.
CREATE TABLE feedback_submissions (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    user_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    what_tried   TEXT NOT NULL DEFAULT '',
    worked       TEXT NOT NULL DEFAULT '',
    missing      TEXT NOT NULL DEFAULT '',
    email        TEXT NOT NULL DEFAULT '',
    user_agent   TEXT NOT NULL DEFAULT '',
    page_url     TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_feedback_created_at ON feedback_submissions(created_at DESC);
