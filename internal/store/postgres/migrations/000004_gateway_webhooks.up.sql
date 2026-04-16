-- Replace the single workspace-level webhook_url with a JSONB map keyed by gateway.
-- Preserves any previously-configured URL as the "agnostic" default.
ALTER TABLE workspaces
    ADD COLUMN webhook_urls JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE workspaces
    SET webhook_urls = jsonb_build_object('agnostic', webhook_url)
    WHERE webhook_url <> '';

ALTER TABLE workspaces
    DROP COLUMN webhook_url;
