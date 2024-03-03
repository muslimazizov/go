-- +goose Up
-- +goose StatementBegin
CREATE TABLE nats_logs
(
    Id Int32,
    ProjectId Int32,
    Name String,
    Description String,
    Priority Int32,
    Removed UInt8,
    EventTime DateTime
) ENGINE = NATS SETTINGS
-- +goose ENVSUB ON
    nats_url = 'gogogo-nats-1:4222',
	nats_subjects = 'logs',
-- +goose ENVSUB OFF
	nats_format = 'JSONEachRow',
    nats_max_block_size = 5,
    nats_flush_interval_ms = 1000;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS nats_logs;
-- +goose StatementEnd
