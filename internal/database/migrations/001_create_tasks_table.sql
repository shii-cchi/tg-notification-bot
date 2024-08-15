-- +goose Up

CREATE TYPE task_status AS ENUM ('in_progress', 'completed', 'deleted');

CREATE TABLE tasks (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    task TEXT NOT NULL,
    task_time TEXT NOT NULL,
    chat_id bigint NOT NULL,
    status task_status NOT NULL DEFAULT 'in_progress',
    created_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

-- +goose Down

DROP TABLE tasks;
DROP TYPE task_status;