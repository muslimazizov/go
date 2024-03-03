-- +goose Up
-- +goose StatementBegin
CREATE TABLE goods (
    id SERIAL PRIMARY KEY,
    project_id INT REFERENCES projects(id),
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255),
    priority INT DEFAULT 1,
    removed BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_goods_project_id ON goods (project_id);
CREATE INDEX idx_goods_name ON goods (name);

-- CREATE OR REPLACE FUNCTION update_priority()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     SELECT pg_advisory_xact_lock(1);
--     NEW.priority := (SELECT COALESCE(MAX(priority), 0) + 1 FROM goods);
-- RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;
--
-- CREATE TRIGGER goods_priority_trigger
--     BEFORE INSERT ON goods
--     FOR EACH ROW EXECUTE PROCEDURE update_priority();

-- CREATE OR REPLACE FUNCTION update_subsequent_priority()
-- RETURNS TRIGGER AS $$
-- DECLARE
--     cur_priority INT;
-- BEGIN
--     cur_priority := NEW.priority;
--     WITH updated_rows AS (
--         SELECT id, priority
--         FROM goods
--         WHERE id > NEW.id
--         ORDER BY id
--         FOR UPDATE
--     ),
--         new_priorities AS (
--         SELECT id, cur_priority + ROW_NUMBER() OVER (ORDER BY id) as new_priority
--         FROM updated_rows
--         )
--     UPDATE goods
--     SET priority = new_priorities.new_priority
--     FROM new_priorities
--     WHERE goods.id = new_priorities.id;
--
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;
--
-- CREATE TRIGGER update_priority_after_update
--     AFTER UPDATE OF priority ON goods
--     FOR EACH ROW EXECUTE PROCEDURE update_subsequent_priority();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE goods;
-- +goose StatementEnd
