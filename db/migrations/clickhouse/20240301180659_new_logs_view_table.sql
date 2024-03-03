-- +goose Up
-- +goose StatementBegin
CREATE MATERIALIZED VIEW logs_to_main TO logs AS SELECT * FROM nats_logs;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS logs_to_main;
-- +goose StatementEnd
