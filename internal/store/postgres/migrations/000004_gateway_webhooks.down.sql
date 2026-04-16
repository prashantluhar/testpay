ALTER TABLE workspaces
    ADD COLUMN webhook_url TEXT NOT NULL DEFAULT '';

UPDATE workspaces
    SET webhook_url = COALESCE(webhook_urls->>'agnostic', '')
    WHERE webhook_urls ? 'agnostic';

ALTER TABLE workspaces
    DROP COLUMN webhook_urls;
