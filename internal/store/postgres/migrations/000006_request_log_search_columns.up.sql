-- Surface two fields on request_logs that users want to search and filter by:
--   merchant_order_id — the caller's reference (pulled from payload
--     order_id / merchant_order_id / metadata.order_id / notes.reference)
--   scenario_id       — which scenario drove this request's response
--
-- Both nullable because a request can arrive with no order ref and with no
-- configured scenario (falls back to built-in "always succeed").
ALTER TABLE request_logs
    ADD COLUMN merchant_order_id TEXT,
    ADD COLUMN scenario_id       UUID REFERENCES scenarios(id) ON DELETE SET NULL;

-- Workspace-scoped partial indexes — filtering by merchant_order_id or
-- scenario_id is only meaningful inside one workspace's traffic.
CREATE INDEX idx_request_logs_merchant_order_id
    ON request_logs(workspace_id, merchant_order_id)
    WHERE merchant_order_id IS NOT NULL;

CREATE INDEX idx_request_logs_scenario_id
    ON request_logs(workspace_id, scenario_id)
    WHERE scenario_id IS NOT NULL;
