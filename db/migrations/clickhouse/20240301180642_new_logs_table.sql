-- +goose Up
-- +goose StatementBegin
CREATE TABLE logs
(
    Id Int32,
    ProjectId Int32,
    Name String,
    Description String,
    Priority Int32,
    Removed UInt8,
    EventTime DateTime
) ENGINE = MergeTree()
ORDER BY (Id, ProjectId, Name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logs;
-- +goose StatementEnd
