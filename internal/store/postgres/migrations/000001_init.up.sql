CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE workspaces (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug       TEXT NOT NULL UNIQUE,
    api_key    TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email        TEXT NOT NULL UNIQUE,
    role         TEXT NOT NULL DEFAULT 'owner',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE scenarios (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id     UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name             TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    gateway          TEXT NOT NULL DEFAULT 'agnostic',
    steps            JSONB NOT NULL DEFAULT '[]',
    webhook_delay_ms INTEGER NOT NULL DEFAULT 0,
    is_default       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE scenario_runs (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scenario_id  UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'running',
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE sessions (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    scenario_id  UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
    ttl_seconds  INTEGER NOT NULL DEFAULT 3600,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE request_logs (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id     UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    scenario_run_id  UUID REFERENCES scenario_runs(id),
    gateway          TEXT NOT NULL,
    method           TEXT NOT NULL,
    path             TEXT NOT NULL,
    request_headers  JSONB NOT NULL DEFAULT '{}',
    request_body     JSONB,
    response_headers JSONB NOT NULL DEFAULT '{}',
    response_body    JSONB,
    response_status  INTEGER NOT NULL,
    duration_ms      INTEGER NOT NULL,
    client_ip        TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_logs_workspace ON request_logs(workspace_id, created_at DESC);

CREATE TABLE webhook_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_log_id  UUID NOT NULL REFERENCES request_logs(id) ON DELETE CASCADE,
    payload         JSONB NOT NULL DEFAULT '{}',
    target_url      TEXT NOT NULL,
    delivery_status TEXT NOT NULL DEFAULT 'pending',
    attempts        INTEGER NOT NULL DEFAULT 0,
    attempt_logs    JSONB NOT NULL DEFAULT '[]',
    delivered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
